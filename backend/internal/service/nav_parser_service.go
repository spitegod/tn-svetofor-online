package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"tn/backend/internal/model"
	"tn/backend/internal/repository"

	"golang.org/x/net/html"
)

const navBaseURL = "https://nav.tn.ru"

type NavParserService struct {
	repo           *repository.SystemCatalogRepository
	client         *http.Client
	parseMu        sync.Mutex
	cancelMu       sync.Mutex
	activeCancel   context.CancelFunc
	progressMu     sync.RWMutex
	progress       model.NavParserProgress
	activeSettings model.NavParserSettings
}

var (
	ErrNavParserRunning    = errors.New("NAV parser is already running")
	ErrNavParserNotRunning = errors.New("NAV parser is not running")
)

const maxNavParserLogs = 400

type navSystemLink struct {
	Name       string
	URL        string
	SystemType string
}

type navCategory struct {
	URL              string
	Slug             string
	Name             string
	ImageURL         string
	ImageContentType string
	ImageData        []byte
	Position         int
}

func NewNavParserService(repo *repository.SystemCatalogRepository) *NavParserService {
	return &NavParserService{
		repo:   repo,
		client: &http.Client{Timeout: 35 * time.Second},
		progress: model.NavParserProgress{
			Stage:   "Ожидание",
			Message: "Парсер готов к запуску",
			Logs:    make([]model.NavParserLogEntry, 0),
		},
		activeSettings: model.NavParserSettings{
			WorkerCount: 4, RequestTimeoutSecs: 35, RetryAttempts: 3, RetryDelaySecs: 2, FallbackSearch: true,
		},
	}
}

func (s *NavParserService) Parse(ctx context.Context) (report model.NavParseReport, parseErr error) {
	return s.parse(ctx, "manual")
}

func (s *NavParserService) parse(ctx context.Context, source string) (report model.NavParseReport, parseErr error) {
	if !s.parseMu.TryLock() {
		return model.NavParseReport{}, ErrNavParserRunning
	}
	s.beginProgress(source)
	parseCtx, cancel := context.WithCancel(ctx)
	s.cancelMu.Lock()
	s.activeCancel = cancel
	s.cancelMu.Unlock()
	ctx = parseCtx
	defer func() {
		cancel()
		s.cancelMu.Lock()
		s.activeCancel = nil
		s.cancelMu.Unlock()
		s.parseMu.Unlock()
	}()
	defer func() {
		if parseErr != nil {
			if errors.Is(parseErr, context.Canceled) {
				s.cancelProgress()
			} else {
				s.failProgress(parseErr)
			}
		}
	}()
	settings, err := s.Settings(ctx)
	if err != nil {
		return model.NavParseReport{}, err
	}
	s.activeSettings = settings
	s.client.Timeout = time.Duration(settings.RequestTimeoutSecs) * time.Second

	rows, err := s.repo.ParserRows(ctx)
	if err != nil {
		return model.NavParseReport{}, err
	}
	report = model.NavParseReport{Total: len(rows), NotFound: make([]string, 0), FailedSystems: make([]string, 0)}
	s.setProgressTotal(len(rows))
	if len(rows) == 0 {
		if err := ctx.Err(); err != nil {
			return report, err
		}
		if err := s.repo.MarkNavParserRun(ctx); err != nil {
			return report, err
		}
		s.completeProgress(report)
		return report, nil
	}

	s.setProgressStage("Каталог NAV", "Загружаем типы и ссылки на системы", 3)
	links, err := s.crawlCatalog(ctx)
	if err != nil {
		return model.NavParseReport{}, err
	}
	s.setProgressStage("Сопоставление", "Сопоставляем системы с каталогом NAV", 28)
	linksByName := make(map[string]navSystemLink, len(links))
	for _, link := range links {
		key := normalizeSystemName(link.Name)
		if _, exists := linksByName[key]; !exists {
			linksByName[key] = link
		}
	}

	type parseJob struct {
		row  model.SystemCatalogRow
		link navSystemLink
	}
	jobs := make([]parseJob, 0, len(rows))
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return report, err
		}
		link, ok := linksByName[normalizeSystemName(row.SystemName)]
		if !ok && settings.FallbackSearch {
			link, ok = s.searchSystem(ctx, row.SystemName)
			if ok {
				report.FallbackFound++
			}
		}
		if !ok {
			report.NotFound = append(report.NotFound, row.SystemName)
			s.appendProgressLog("warning", "Не найдена в NAV: "+row.SystemName)
			continue
		}
		report.Found++
		jobs = append(jobs, parseJob{row: row, link: link})
	}
	s.setMatchedProgress(report)

	jobCh := make(chan parseJob)
	var wg sync.WaitGroup
	var mu sync.Mutex
	workerCount := settings.WorkerCount
	if len(jobs) < workerCount {
		workerCount = len(jobs)
	}
	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				if ctx.Err() != nil {
					return
				}
				characteristics, jobErr := s.scrapeCharacteristics(ctx, job.link.URL)
				if jobErr == nil {
					characteristics = append([]model.SystemCharacteristic{{
						Position: 1,
						Name:     "Тип системы",
						Value:    job.link.SystemType,
					}}, characteristics...)
					for index := range characteristics {
						characteristics[index].Position = index + 1
					}
					jobErr = s.repo.SaveParsed(ctx, job.row.SystemName, job.link.URL, characteristics)
				}
				if errors.Is(jobErr, context.Canceled) || ctx.Err() != nil {
					return
				}
				mu.Lock()
				if jobErr != nil {
					report.Failed++
					report.FailedSystems = append(report.FailedSystems, job.row.SystemName)
				} else {
					report.Updated++
				}
				snapshot := report
				mu.Unlock()
				s.recordSystemProgress(job.row.SystemName, jobErr, snapshot)
			}
		}()
	}
	for _, job := range jobs {
		select {
		case jobCh <- job:
		case <-ctx.Done():
			close(jobCh)
			wg.Wait()
			return report, ctx.Err()
		}
	}
	close(jobCh)
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return report, err
	}
	sort.Strings(report.NotFound)
	sort.Strings(report.FailedSystems)
	if report.Failed == 0 {
		if err := s.repo.MarkNavParserRun(ctx); err != nil {
			return report, err
		}
	}

	s.completeProgress(report)
	return report, nil
}

func (s *NavParserService) Cancel() error {
	s.cancelMu.Lock()
	cancel := s.activeCancel
	s.cancelMu.Unlock()
	if cancel == nil {
		return ErrNavParserNotRunning
	}

	s.progressMu.Lock()
	if !s.progress.Running {
		s.progressMu.Unlock()
		return ErrNavParserNotRunning
	}
	s.progress.Stage = "Отмена"
	s.progress.Message = "Останавливаем парсинг…"
	s.appendProgressLogLocked("warning", "Получена команда остановить парсинг")
	s.progressMu.Unlock()
	cancel()
	return nil
}

func (s *NavParserService) Status() model.NavParserProgress {
	s.progressMu.RLock()
	defer s.progressMu.RUnlock()
	result := s.progress
	result.Logs = make([]model.NavParserLogEntry, len(s.progress.Logs))
	copy(result.Logs, s.progress.Logs)
	return result
}

func (s *NavParserService) Runs(ctx context.Context, limit int) ([]model.NavParserRun, error) {
	return s.repo.NavParserRuns(ctx, limit)
}

func (s *NavParserService) beginProgress(source string) {
	now := time.Now()
	s.progressMu.Lock()
	s.progress = model.NavParserProgress{
		Running:   true,
		Source:    source,
		Stage:     "Подготовка",
		Message:   "Получаем список систем для обновления",
		Percent:   1,
		StartedAt: &now,
		Logs:      make([]model.NavParserLogEntry, 0, 64),
	}
	s.appendProgressLogLocked("info", "Запуск парсера NAV")
	s.progressMu.Unlock()
}

func (s *NavParserService) setProgressTotal(total int) {
	s.progressMu.Lock()
	s.progress.Total = total
	s.appendProgressLogLocked("info", fmt.Sprintf("Получено систем для обработки: %d", total))
	s.progressMu.Unlock()
}

func (s *NavParserService) setProgressStage(stage string, message string, percent int) {
	s.progressMu.Lock()
	s.progress.Stage = stage
	s.progress.Message = message
	s.progress.Percent = percent
	s.appendProgressLogLocked("info", message)
	s.progressMu.Unlock()
}

func (s *NavParserService) setMatchedProgress(report model.NavParseReport) {
	s.progressMu.Lock()
	s.progress.Stage = "Характеристики"
	s.progress.Message = "Загружаем характеристики найденных систем"
	s.progress.Found = report.Found
	s.progress.NotFound = len(report.NotFound)
	s.progress.Processed = len(report.NotFound)
	s.progress.Percent = parserSystemPercent(s.progress.Processed, report.Total)
	s.appendProgressLogLocked("info", fmt.Sprintf("Сопоставление завершено: найдено %d, не найдено %d", report.Found, len(report.NotFound)))
	s.progressMu.Unlock()
}

func (s *NavParserService) recordSystemProgress(systemName string, err error, report model.NavParseReport) {
	s.progressMu.Lock()
	s.progress.Processed = len(report.NotFound) + report.Updated + report.Failed
	s.progress.Found = report.Found
	s.progress.Updated = report.Updated
	s.progress.Failed = report.Failed
	s.progress.NotFound = len(report.NotFound)
	s.progress.Percent = parserSystemPercent(s.progress.Processed, report.Total)
	s.progress.Message = fmt.Sprintf("Обработано %d из %d систем", s.progress.Processed, report.Total)
	if err != nil {
		s.appendProgressLogLocked("error", fmt.Sprintf("Ошибка обработки %s: %v", systemName, err))
	} else {
		s.appendProgressLogLocked("success", "Обновлена: "+systemName)
	}
	s.progressMu.Unlock()
}

func (s *NavParserService) completeProgress(report model.NavParseReport) {
	now := time.Now()
	s.progressMu.Lock()
	s.progress.Running = false
	s.progress.Stage = "Завершено"
	s.progress.Message = fmt.Sprintf("Обновлено %d из %d систем", report.Updated, report.Total)
	s.progress.Percent = 100
	s.progress.Processed = report.Total
	s.progress.Total = report.Total
	s.progress.Found = report.Found
	s.progress.Updated = report.Updated
	s.progress.Failed = report.Failed
	s.progress.NotFound = len(report.NotFound)
	s.progress.FinishedAt = &now
	s.appendProgressLogLocked("success", fmt.Sprintf("Парсинг завершён: обновлено %d, ошибок %d, не найдено %d", report.Updated, report.Failed, len(report.NotFound)))
	s.progressMu.Unlock()
	s.persistProgress("completed")
}

func (s *NavParserService) failProgress(err error) {
	now := time.Now()
	s.progressMu.Lock()
	s.progress.Running = false
	s.progress.Stage = "Ошибка"
	s.progress.Message = err.Error()
	s.progress.FinishedAt = &now
	s.appendProgressLogLocked("error", "Парсинг остановлен: "+err.Error())
	s.progressMu.Unlock()
	s.persistProgress("failed")
}

func (s *NavParserService) cancelProgress() {
	now := time.Now()
	s.progressMu.Lock()
	s.progress.Running = false
	s.progress.Stage = "Отменено"
	s.progress.Message = "Парсинг отменён пользователем"
	s.progress.FinishedAt = &now
	s.appendProgressLogLocked("warning", "Парсинг отменён пользователем")
	s.progressMu.Unlock()
	s.persistProgress("canceled")

	s.progressMu.Lock()
	s.progress = model.NavParserProgress{
		Stage:   "Ожидание",
		Message: "Парсер готов к запуску",
		Logs:    make([]model.NavParserLogEntry, 0),
	}
	s.progressMu.Unlock()
}

func (s *NavParserService) persistProgress(status string) {
	progress := s.Status()
	if progress.StartedAt == nil || progress.FinishedAt == nil {
		return
	}
	run := model.NavParserRun{
		Source:     progress.Source,
		Status:     status,
		Message:    progress.Message,
		Total:      progress.Total,
		Found:      progress.Found,
		Updated:    progress.Updated,
		Failed:     progress.Failed,
		NotFound:   progress.NotFound,
		StartedAt:  *progress.StartedAt,
		FinishedAt: *progress.FinishedAt,
		Logs:       progress.Logs,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.repo.SaveNavParserRun(ctx, run); err != nil {
		log.Printf("save NAV parser run history: %v", err)
	}
}

func (s *NavParserService) appendProgressLog(level string, message string) {
	s.progressMu.Lock()
	s.appendProgressLogLocked(level, message)
	s.progressMu.Unlock()
}

func (s *NavParserService) appendProgressLogLocked(level string, message string) {
	s.progress.Logs = append(s.progress.Logs, model.NavParserLogEntry{Time: time.Now(), Level: level, Message: message})
	if len(s.progress.Logs) > maxNavParserLogs {
		s.progress.Logs = append([]model.NavParserLogEntry(nil), s.progress.Logs[len(s.progress.Logs)-maxNavParserLogs:]...)
	}
}

func parserSystemPercent(processed int, total int) int {
	if total <= 0 {
		return 30
	}
	percent := 30 + processed*68/total
	if percent > 98 {
		return 98
	}
	return percent
}

func (s *NavParserService) Settings(ctx context.Context) (model.NavParserSettings, error) {
	settings, err := s.repo.NavParserSettings(ctx)
	if err != nil {
		return model.NavParserSettings{}, err
	}
	return withNextNavRun(settings), nil
}

func (s *NavParserService) UpdateSettings(ctx context.Context, input model.NavParserSettings) (model.NavParserSettings, error) {
	if input.UpdateIntervalDays < 1 || input.UpdateIntervalDays > 365 {
		return model.NavParserSettings{}, fmt.Errorf("update interval must be between 1 and 365 days")
	}
	if input.WorkerCount < 1 || input.WorkerCount > 10 {
		return model.NavParserSettings{}, fmt.Errorf("worker count must be between 1 and 10")
	}
	if input.RequestTimeoutSecs < 5 || input.RequestTimeoutSecs > 120 {
		return model.NavParserSettings{}, fmt.Errorf("request timeout must be between 5 and 120 seconds")
	}
	if input.RetryAttempts < 1 || input.RetryAttempts > 5 {
		return model.NavParserSettings{}, fmt.Errorf("retry attempts must be between 1 and 5")
	}
	if input.RetryDelaySecs < 1 || input.RetryDelaySecs > 30 {
		return model.NavParserSettings{}, fmt.Errorf("retry delay must be between 1 and 30 seconds")
	}
	settings, err := s.repo.UpdateNavParserSettings(ctx, input)
	if err != nil {
		return model.NavParserSettings{}, err
	}
	return withNextNavRun(settings), nil
}

func (s *NavParserService) searchSystem(ctx context.Context, systemName string) (navSystemLink, bool) {
	document, err := s.fetchDocument(ctx, navBaseURL+"/search/?q="+url.QueryEscape(systemName))
	if err != nil {
		return navSystemLink{}, false
	}
	types, _ := s.repo.SystemTypes(ctx)
	return systemLinkFromSearch(document, systemName, types)
}

func systemLinkFromSearch(document *html.Node, systemName string, types []model.SystemTypeOption) (navSystemLink, bool) {
	wantedName := normalizeSystemName(systemName)
	for _, anchor := range findNodes(document, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "a" && hasClass(node, "b-search-teaser__title")
	}) {
		if normalizeSystemName(nodeText(anchor)) != wantedName {
			continue
		}
		parsed, err := url.Parse(attribute(anchor, "href"))
		if err != nil {
			continue
		}
		absolute := (&url.URL{Scheme: "https", Host: "nav.tn.ru"}).ResolveReference(parsed)
		parts := pathParts(absolute.Path)
		if absolute.Hostname() != "nav.tn.ru" || len(parts) != 3 || parts[0] != "systems" {
			continue
		}
		systemType := ""
		for _, item := range types {
			if item.Slug == parts[1] {
				systemType = item.Name
				break
			}
		}
		absolute.RawQuery = ""
		absolute.Fragment = ""
		return navSystemLink{Name: systemName, URL: absolute.String(), SystemType: systemType}, true
	}
	return navSystemLink{}, false
}

func (s *NavParserService) RunScheduler(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	s.runScheduledParseIfDue(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runScheduledParseIfDue(ctx)
		}
	}
}

func (s *NavParserService) runScheduledParseIfDue(ctx context.Context) {
	settings, err := s.Settings(ctx)
	if err != nil {
		log.Printf("load NAV parser schedule: %v", err)
		return
	}
	if settings.LastRunAt == nil || settings.NextRunAt == nil || time.Now().Before(*settings.NextRunAt) {
		return
	}
	if _, err := s.parse(ctx, "scheduled"); err != nil && !errors.Is(err, ErrNavParserRunning) && !errors.Is(err, context.Canceled) {
		log.Printf("scheduled NAV parsing failed: %v", err)
	}
}

func withNextNavRun(settings model.NavParserSettings) model.NavParserSettings {
	if settings.LastRunAt != nil {
		nextRunAt := settings.LastRunAt.Add(time.Duration(settings.UpdateIntervalDays) * 24 * time.Hour)
		settings.NextRunAt = &nextRunAt
	}
	return settings
}

func (s *NavParserService) crawlCatalog(ctx context.Context) ([]navSystemLink, error) {
	document, err := s.fetchDocument(ctx, navBaseURL+"/systems/")
	if err != nil {
		return nil, fmt.Errorf("load nav.tn.ru systems catalog: %w", err)
	}
	categories := collectCategoryURLs(document)
	if len(categories) == 0 {
		return nil, fmt.Errorf("nav.tn.ru systems catalog has no categories")
	}
	s.setCatalogProgress(0, len(categories), "")

	type categoryResult struct {
		category navCategory
		links    []navSystemLink
		err      error
	}
	categoryCh := make(chan navCategory)
	resultCh := make(chan categoryResult, len(categories))
	workerCount := s.activeSettings.WorkerCount
	if len(categories) < workerCount {
		workerCount = len(categories)
	}
	var wg sync.WaitGroup
	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for category := range categoryCh {
				links, categoryName, err := s.crawlCategory(ctx, category.URL)
				category.Name = categoryName
				if err == nil && category.ImageURL != "" {
					contentType, data, imageErr := s.fetchImage(ctx, category.ImageURL)
					if imageErr != nil {
						log.Printf("load NAV system type image %s: %v", category.ImageURL, imageErr)
					} else {
						category.ImageContentType = contentType
						category.ImageData = data
					}
				}
				resultCh <- categoryResult{category: category, links: links, err: err}
			}
		}()
	}
	go func() {
		defer close(categoryCh)
		for _, category := range categories {
			select {
			case categoryCh <- category:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	byURL := make(map[string]navSystemLink)
	parsedCategories := make([]navCategory, 0, len(categories))
	completedCategories := 0
	for result := range resultCh {
		if result.err != nil {
			return nil, result.err
		}
		completedCategories++
		s.setCatalogProgress(completedCategories, len(categories), result.category.Name)
		parsedCategories = append(parsedCategories, result.category)
		for _, link := range result.links {
			byURL[link.URL] = link
		}
	}
	sort.Slice(parsedCategories, func(left, right int) bool {
		return parsedCategories[left].Position < parsedCategories[right].Position
	})
	systemTypes := make([]model.SystemTypeOption, 0, len(parsedCategories))
	for _, category := range parsedCategories {
		if category.Name == "" {
			continue
		}
		systemTypes = append(systemTypes, model.SystemTypeOption{
			Slug:             category.Slug,
			Name:             category.Name,
			ImageURL:         category.ImageURL,
			ImageContentType: category.ImageContentType,
			ImageData:        category.ImageData,
			Position:         len(systemTypes) + 1,
		})
	}
	if err := s.repo.ReplaceSystemTypes(ctx, systemTypes); err != nil {
		return nil, err
	}
	links := make([]navSystemLink, 0, len(byURL))
	for _, link := range byURL {
		links = append(links, link)
	}
	return links, nil
}

func (s *NavParserService) setCatalogProgress(completed int, total int, categoryName string) {
	s.progressMu.Lock()
	s.progress.Stage = "Каталог NAV"
	s.progress.Percent = 5
	if total > 0 {
		s.progress.Percent = 5 + completed*22/total
	}
	s.progress.Message = fmt.Sprintf("Обработано категорий: %d из %d", completed, total)
	if completed == 0 {
		s.appendProgressLogLocked("info", fmt.Sprintf("Найдено категорий систем: %d", total))
	} else if categoryName != "" {
		s.appendProgressLogLocked("info", fmt.Sprintf("Категория %d/%d: %s", completed, total, categoryName))
	}
	s.progressMu.Unlock()
}

func (s *NavParserService) crawlCategory(ctx context.Context, categoryURL string) ([]navSystemLink, string, error) {
	document, err := s.fetchDocument(ctx, categoryURL)
	if err != nil {
		return nil, "", fmt.Errorf("load nav.tn.ru category %s: %w", categoryURL, err)
	}
	title := findNode(document, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "h1" && hasClass(node, "l-h1")
	})
	categoryName := ""
	if title != nil {
		categoryName = normalizeText(nodeText(title))
	}
	links := collectSystemLinks(document, categoryName)
	pageCount := collectPageCount(document)
	for page := 2; page <= pageCount; page++ {
		pageURL := categoryURL + "?PAGEN_1=" + strconv.Itoa(page)
		document, err := s.fetchDocument(ctx, pageURL)
		if err != nil {
			return nil, "", fmt.Errorf("load nav.tn.ru category page %s: %w", pageURL, err)
		}
		links = append(links, collectSystemLinks(document, categoryName)...)
	}
	return links, categoryName, nil
}

func (s *NavParserService) scrapeCharacteristics(ctx context.Context, systemURL string) ([]model.SystemCharacteristic, error) {
	document, err := s.fetchDocument(ctx, systemURL)
	if err != nil {
		return nil, fmt.Errorf("load nav.tn.ru system %s: %w", systemURL, err)
	}
	table := findNode(document, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "div" && hasClass(node, "b-element-detail-chars__table")
	})
	if table == nil {
		return nil, fmt.Errorf("system characteristics table not found")
	}

	characteristics := make([]model.SystemCharacteristic, 0)
	for _, row := range findNodes(table, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "tr"
	}) {
		cells := findNodes(row, func(node *html.Node) bool {
			return node.Type == html.ElementNode && node.Data == "td"
		})
		if len(cells) < 2 {
			continue
		}
		name := normalizeText(nodeText(cells[0]))
		value := normalizeText(nodeText(cells[1]))
		if name == "" || value == "" || strings.EqualFold(name, "Наименование показателя") {
			continue
		}
		characteristics = append(characteristics, model.SystemCharacteristic{
			Position: len(characteristics) + 1,
			Name:     name,
			Value:    value,
		})
	}
	if len(characteristics) == 0 {
		return nil, fmt.Errorf("system characteristics table is empty")
	}
	return characteristics, nil
}

func (s *NavParserService) fetchDocument(ctx context.Context, address string) (*html.Node, error) {
	attempts := max(1, s.activeSettings.RetryAttempts)
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		document, retry, err := s.fetchDocumentOnce(ctx, address)
		if err == nil {
			return document, nil
		}
		lastErr = err
		if !retry || attempt == attempts {
			break
		}
		delay := time.Duration(max(1, s.activeSettings.RetryDelaySecs)*attempt) * time.Second
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, lastErr
}

func (s *NavParserService) fetchDocumentOnce(ctx context.Context, address string) (*html.Node, bool, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, false, err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; TNSvetofor/1.0; +https://nav.tn.ru/)")
	request.Header.Set("Accept-Language", "ru-RU,ru;q=0.9")
	response, err := s.client.Do(request)
	if err != nil {
		return nil, true, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		retry := response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= http.StatusInternalServerError
		return nil, retry, fmt.Errorf("unexpected HTTP status %s", response.Status)
	}
	document, err := html.Parse(io.LimitReader(response.Body, 12<<20))
	return document, false, err
}

func (s *NavParserService) fetchImage(ctx context.Context, address string) (string, []byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return "", nil, err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; TNSvetofor/1.0; +https://nav.tn.ru/)")
	request.Header.Set("Referer", navBaseURL+"/systems/")
	response, err := s.client.Do(request)
	if err != nil {
		return "", nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("unexpected HTTP status %s", response.Status)
	}
	contentType := strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0])
	if !strings.HasPrefix(contentType, "image/") {
		return "", nil, fmt.Errorf("unexpected content type %q", contentType)
	}
	const maxImageSize = 4 << 20
	data, err := io.ReadAll(io.LimitReader(response.Body, maxImageSize+1))
	if err != nil {
		return "", nil, err
	}
	if len(data) == 0 || len(data) > maxImageSize {
		return "", nil, fmt.Errorf("invalid image size %d", len(data))
	}
	return contentType, data, nil
}

func (s *NavParserService) SystemTypeImage(ctx context.Context, slug string) (model.SystemTypeImage, error) {
	if strings.TrimSpace(slug) == "" {
		return model.SystemTypeImage{}, fmt.Errorf("system type slug is required")
	}
	return s.repo.SystemTypeImage(ctx, slug)
}

func collectCategoryURLs(document *html.Node) []navCategory {
	seen := make(map[string]struct{})
	result := make([]navCategory, 0)
	for _, anchor := range findNodes(document, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "a" && hasClass(node, "b-system-page__item")
	}) {
		href := attribute(anchor, "href")
		parsed, err := url.Parse(href)
		if err != nil {
			continue
		}
		parts := pathParts(parsed.Path)
		if len(parts) == 2 && parts[0] == "systems" {
			address := navBaseURL + "/systems/" + parts[1] + "/"
			if _, exists := seen[address]; exists {
				continue
			}
			seen[address] = struct{}{}
			result = append(result, navCategory{
				URL:      address,
				Slug:     parts[1],
				ImageURL: categoryImageURL(anchor),
				Position: len(result) + 1,
			})
		}
	}
	return result
}

func categoryImageURL(anchor *html.Node) string {
	image := findNode(anchor, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "img"
	})
	if image == nil {
		return ""
	}
	for _, name := range []string{"src", "data-src", "data-lazy-src"} {
		if resolved := resolveNavAssetURL(attribute(image, name)); resolved != "" {
			return resolved
		}
	}
	return ""
}

func resolveNavAssetURL(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Path == "" {
		return ""
	}
	absolute := (&url.URL{Scheme: "https", Host: "nav.tn.ru"}).ResolveReference(parsed)
	host := strings.ToLower(absolute.Hostname())
	if absolute.Scheme != "https" || (host != "tn.ru" && !strings.HasSuffix(host, ".tn.ru")) {
		return ""
	}
	return absolute.String()
}

func collectSystemLinks(document *html.Node, systemType string) []navSystemLink {
	links := make([]navSystemLink, 0)
	for _, anchor := range findNodes(document, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "a" && hasClass(node, "b-products__title")
	}) {
		href := attribute(anchor, "href")
		parsed, err := url.Parse(href)
		if err != nil {
			continue
		}
		parts := pathParts(parsed.Path)
		if len(parts) != 3 || parts[0] != "systems" {
			continue
		}
		links = append(links, navSystemLink{
			Name:       normalizeText(nodeText(anchor)),
			URL:        navBaseURL + parsed.Path,
			SystemType: systemType,
		})
	}
	return links
}

func collectPageCount(document *html.Node) int {
	pageCount := 1
	for _, anchor := range findNodes(document, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "a"
	}) {
		parsed, err := url.Parse(attribute(anchor, "href"))
		if err != nil {
			continue
		}
		page, err := strconv.Atoi(parsed.Query().Get("PAGEN_1"))
		if err == nil && page > pageCount {
			pageCount = page
		}
	}
	return pageCount
}

func normalizeSystemName(value string) string {
	value = strings.ReplaceAll(strings.ToLower(normalizeText(value)), "ё", "е")
	return strings.Join(strings.FieldsFunc(value, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}), " ")
}

func normalizeText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func pathParts(path string) []string {
	return strings.FieldsFunc(path, func(r rune) bool { return r == '/' })
}

func attribute(node *html.Node, name string) string {
	for _, attr := range node.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}

func hasClass(node *html.Node, className string) bool {
	for _, value := range strings.Fields(attribute(node, "class")) {
		if value == className {
			return true
		}
	}
	return false
}

func nodeText(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}
	var builder strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		builder.WriteString(" ")
		builder.WriteString(nodeText(child))
	}
	return builder.String()
}

func findNode(node *html.Node, predicate func(*html.Node) bool) *html.Node {
	if predicate(node) {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result := findNode(child, predicate); result != nil {
			return result
		}
	}
	return nil
}

func findNodes(node *html.Node, predicate func(*html.Node) bool) []*html.Node {
	result := make([]*html.Node, 0)
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if predicate(current) {
			result = append(result, current)
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return result
}
