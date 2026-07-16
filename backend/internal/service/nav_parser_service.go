package service

import (
	"context"
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
	repo    *repository.SystemCatalogRepository
	client  *http.Client
	parseMu sync.Mutex
}

type navSystemLink struct {
	Name       string
	URL        string
	SystemType string
}

type navCategory struct {
	URL      string
	Slug     string
	Name     string
	Position int
}

func NewNavParserService(repo *repository.SystemCatalogRepository) *NavParserService {
	return &NavParserService{
		repo:   repo,
		client: &http.Client{Timeout: 35 * time.Second},
	}
}

func (s *NavParserService) Parse(ctx context.Context) (model.NavParseReport, error) {
	s.parseMu.Lock()
	defer s.parseMu.Unlock()

	rows, err := s.repo.ParserRows(ctx)
	if err != nil {
		return model.NavParseReport{}, err
	}
	report := model.NavParseReport{Total: len(rows), NotFound: make([]string, 0)}
	if len(rows) == 0 {
		return report, s.repo.MarkNavParserRun(ctx)
	}

	links, err := s.crawlCatalog(ctx)
	if err != nil {
		return model.NavParseReport{}, err
	}
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
		link, ok := linksByName[normalizeSystemName(row.SystemName)]
		if !ok {
			report.NotFound = append(report.NotFound, row.SystemName)
			continue
		}
		report.Found++
		jobs = append(jobs, parseJob{row: row, link: link})
	}

	jobCh := make(chan parseJob)
	var wg sync.WaitGroup
	var mu sync.Mutex
	workerCount := 4
	if len(jobs) < workerCount {
		workerCount = len(jobs)
	}
	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				characteristics, err := s.scrapeCharacteristics(ctx, job.link.URL)
				if err == nil {
					characteristics = append([]model.SystemCharacteristic{{
						Position: 1,
						Name:     "Тип системы",
						Value:    job.link.SystemType,
					}}, characteristics...)
					for index := range characteristics {
						characteristics[index].Position = index + 1
					}
					err = s.repo.SaveParsed(ctx, job.row.SystemName, job.link.URL, characteristics)
				}
				mu.Lock()
				if err != nil {
					report.Failed++
				} else {
					report.Updated++
				}
				mu.Unlock()
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
	sort.Strings(report.NotFound)
	if err := s.repo.MarkNavParserRun(ctx); err != nil {
		return report, err
	}

	return report, nil
}

func (s *NavParserService) Settings(ctx context.Context) (model.NavParserSettings, error) {
	settings, err := s.repo.NavParserSettings(ctx)
	if err != nil {
		return model.NavParserSettings{}, err
	}
	return withNextNavRun(settings), nil
}

func (s *NavParserService) UpdateInterval(ctx context.Context, days int) (model.NavParserSettings, error) {
	if days < 1 || days > 365 {
		return model.NavParserSettings{}, fmt.Errorf("update interval must be between 1 and 365 days")
	}
	settings, err := s.repo.UpdateNavParserInterval(ctx, days)
	if err != nil {
		return model.NavParserSettings{}, err
	}
	return withNextNavRun(settings), nil
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
	if _, err := s.Parse(ctx); err != nil {
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

	type categoryResult struct {
		category navCategory
		links    []navSystemLink
		err      error
	}
	categoryCh := make(chan navCategory)
	resultCh := make(chan categoryResult, len(categories))
	workerCount := 4
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
	for result := range resultCh {
		if result.err != nil {
			return nil, result.err
		}
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
			Slug:     category.Slug,
			Name:     category.Name,
			Position: len(systemTypes) + 1,
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
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; TNSvetofor/1.0; +https://nav.tn.ru/)")
	request.Header.Set("Accept-Language", "ru-RU,ru;q=0.9")
	response, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %s", response.Status)
	}
	return html.Parse(io.LimitReader(response.Body, 12<<20))
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
			result = append(result, navCategory{URL: address, Slug: parts[1], Position: len(result) + 1})
		}
	}
	return result
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
	value = strings.ToLower(strings.ReplaceAll(normalizeText(value), "ё", "е"))
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
