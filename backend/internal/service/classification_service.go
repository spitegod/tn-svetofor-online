package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"tn/backend/internal/model"
	"tn/backend/internal/repository"

	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding/charmap"
)

type ClassificationService struct {
	repo       *repository.ClassificationRepository
	httpClient *http.Client
}

func NewClassificationService(repo *repository.ClassificationRepository) *ClassificationService {
	return &ClassificationService{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

func (s *ClassificationService) List(ctx context.Context, filter model.ClassificationFilter) (model.ClassificationList, error) {
	rows, err := s.repo.List(ctx, filter)
	if err != nil {
		return model.ClassificationList{}, err
	}

	stats, err := s.repo.Stats(ctx, filter.OrderID)
	if err != nil {
		return model.ClassificationList{}, err
	}

	beforeOptions, err := s.repo.Options(ctx, "class_before", filter.OrderID)
	if err != nil {
		return model.ClassificationList{}, err
	}

	afterOptions, err := s.repo.Options(ctx, "class_after", filter.OrderID)
	if err != nil {
		return model.ClassificationList{}, err
	}

	return model.ClassificationList{
		Rows:          rows,
		Stats:         stats,
		BeforeOptions: beforeOptions,
		AfterOptions:  afterOptions,
	}, nil
}

func (s *ClassificationService) Import(ctx context.Context, orderID int64, file io.Reader) (model.ClassificationList, error) {
	if orderID <= 0 {
		orderID = 1
	}

	spreadsheet, err := excelize.OpenReader(file)
	if err != nil {
		return model.ClassificationList{}, fmt.Errorf("open excel file: %w", err)
	}
	defer spreadsheet.Close()

	sheetName := spreadsheet.GetSheetName(0)
	if sheetName == "" {
		return model.ClassificationList{}, fmt.Errorf("excel file has no sheets")
	}

	excelRows, err := spreadsheet.GetRows(sheetName)
	if err != nil {
		return model.ClassificationList{}, fmt.Errorf("read excel rows: %w", err)
	}

	rows := make([]model.ClassificationChange, 0, len(excelRows))
	for index, excelRow := range excelRows {
		if index < 2 || len(excelRow) < 3 {
			continue
		}

		systemName := normalizeCell(excelRow[0])
		classBefore := normalizeStatus(excelRow[1])
		classAfter := normalizeStatus(excelRow[2])
		if systemName == "" || classBefore == "" || classAfter == "" {
			continue
		}

		rows = append(rows, model.ClassificationChange{
			Position:    len(rows) + 1,
			SystemName:  systemName,
			ClassBefore: classBefore,
			ClassAfter:  classAfter,
		})
	}

	if len(rows) == 0 {
		return model.ClassificationList{}, fmt.Errorf("excel file has no classification rows")
	}

	s.resolveSystemURLs(ctx, rows)

	if err := s.repo.ReplaceAll(ctx, orderID, rows); err != nil {
		return model.ClassificationList{}, err
	}

	return s.List(ctx, model.ClassificationFilter{OrderID: orderID})
}

func (s *ClassificationService) Update(ctx context.Context, id int64, orderID int64, row model.ClassificationChange) (model.ClassificationChange, error) {
	if id <= 0 || orderID <= 0 {
		return model.ClassificationChange{}, fmt.Errorf("invalid classification change")
	}
	row.SystemName = normalizeCell(row.SystemName)
	row.ClassBefore = normalizeStatus(row.ClassBefore)
	row.ClassAfter = normalizeStatus(row.ClassAfter)
	if row.SystemName == "" {
		return model.ClassificationChange{}, fmt.Errorf("system name cannot be empty")
	}
	if !validStatus(row.ClassBefore, true) || !validStatus(row.ClassAfter, false) {
		return model.ClassificationChange{}, fmt.Errorf("invalid classification status")
	}
	return s.repo.Update(ctx, id, orderID, row)
}

func (s *ClassificationService) Export(ctx context.Context, filter model.ClassificationFilter) ([]byte, error) {
	rows, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()
	defer file.Close()

	sheet := "Таблица 1"
	file.SetSheetName("Sheet1", sheet)
	_ = file.SetCellValue(sheet, "A1", "Название системы")
	_ = file.SetCellValue(sheet, "B1", "Класс")
	_ = file.SetCellValue(sheet, "B2", "было")
	_ = file.SetCellValue(sheet, "C2", "стало")
	_ = file.MergeCell(sheet, "A1", "A2")
	_ = file.MergeCell(sheet, "B1", "C1")

	for index, row := range rows {
		excelRow := index + 3
		_ = file.SetCellValue(sheet, fmt.Sprintf("A%d", excelRow), row.SystemName)
		_ = file.SetCellValue(sheet, fmt.Sprintf("B%d", excelRow), row.ClassBefore)
		_ = file.SetCellValue(sheet, fmt.Sprintf("C%d", excelRow), row.ClassAfter)
	}

	_ = file.SetColWidth(sheet, "A", "A", 46)
	_ = file.SetColWidth(sheet, "B", "C", 24)

	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		return nil, fmt.Errorf("write excel export: %w", err)
	}

	return buffer.Bytes(), nil
}

func normalizeCell(value string) string {
	value = strings.TrimSpace(repairMojibake(value))
	value = strings.Join(strings.Fields(value), " ")
	return value
}

func normalizeStatus(value string) string {
	value = normalizeCell(value)
	value = strings.TrimSuffix(value, ".")

	switch strings.ToLower(value) {
	case "новая система":
		return "Новая система"
	case "рекомендованная", "рекомендованные":
		return "Рекомендованная"
	case "разрешенная", "разрешённая", "разрешенные", "разрешённые":
		return "Разрешенная"
	case "запрещенная", "запрещённая", "запрещенные", "запрещённые":
		return "Запрещенная"
	default:
		return value
	}
}

func validStatus(value string, allowNew bool) bool {
	if allowNew && value == "Новая система" {
		return true
	}
	return value == "Рекомендованная" || value == "Разрешенная" || value == "Запрещенная"
}

func repairMojibake(value string) string {
	if !looksLikeMojibake(value) {
		return value
	}

	encoded, err := charmap.Windows1251.NewEncoder().String(value)
	if err != nil {
		return value
	}

	if utf8.ValidString(encoded) {
		return encoded
	}

	return value
}

func looksLikeMojibake(value string) bool {
	markers := []string{"Рќ", "Рў", "РЎ", "Рџ", "Р ", "Рљ", "Р‘", "Р°", "Рµ", "Рё", "СЃ", "С‚", "СЏ", "СЊ"}
	for _, marker := range markers {
		if strings.Contains(value, marker) {
			return true
		}
	}

	return false
}

func (s *ClassificationService) resolveSystemURLs(ctx context.Context, rows []model.ClassificationChange) {
	var wg sync.WaitGroup
	guard := make(chan struct{}, 8)

	for index := range rows {
		wg.Add(1)
		go func(rowIndex int) {
			defer wg.Done()

			select {
			case guard <- struct{}{}:
				defer func() { <-guard }()
			case <-ctx.Done():
				return
			}

			rows[rowIndex].SystemURL = s.resolveSystemURL(ctx, rows[rowIndex].SystemName)
		}(index)
	}

	wg.Wait()
}

func (s *ClassificationService) resolveSystemURL(ctx context.Context, systemName string) string {
	searchURL := "https://nav.tn.ru/search/?q=" + url.QueryEscape(systemName)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return ""
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	response, err := s.httpClient.Do(request)
	if err != nil {
		return ""
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024))
	if err != nil {
		return ""
	}

	content := string(body)
	if !strings.Contains(strings.ToLower(content), strings.ToLower(systemName)) {
		return ""
	}

	matches := regexp.MustCompile(`href="([^"]*/systems/[^"]+)"`).FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		link := match[1]
		if strings.HasPrefix(link, "/") {
			return "https://nav.tn.ru" + link
		}
		if strings.HasPrefix(link, "https://nav.tn.ru/") {
			return link
		}
	}

	return ""
}
