package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"tn/backend/internal/model"
	"tn/backend/internal/repository"

	"github.com/xuri/excelize/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/encoding/charmap"
)

type ClassificationService struct {
	repo       *repository.ClassificationRepository
	httpClient *http.Client
}

const unassignedConstructionType = "Тип не присвоен"

type knownNAVSystem struct {
	URL              string
	ConstructionType string
}

var knownNAVSystems = map[string]knownNAVSystem{
	"тн гео полигон фрост": {
		URL:              "https://nav.tn.ru/systems/poligony-ploshchadki-khraneniya-i-pr/tn-geo-poligon-frost/",
		ConstructionType: "Специальные сооружения",
	},
	"тн гео хвостохранилище фрост": {
		URL:              "https://nav.tn.ru/systems/iskusstvennye-vodoemy-prudy-i-pr/tn-geo-khvostokhranilishche-frost/",
		ConstructionType: "Специальные сооружения",
	},
	"тн гео амбар шламовый фрост": {
		URL:              "https://nav.tn.ru/systems/iskusstvennye-vodoemy-prudy-i-pr/tn-geo-ambar-shlamovyy-frost/",
		ConstructionType: "Специальные сооружения",
	},
	"тн авиа впп фрост": {
		URL:              "https://nav.tn.ru/systems/konstruktsiya-letnogo-polya/tn-avia-vpp-frost/",
		ConstructionType: "Транспортное и дорожное строительство",
	},
	"тн кровля солид керамзит": {
		URL:              "https://nav.tn.ru/systems/ploskaya-krysha/tn-krovlya-solid-keramzit/",
		ConstructionType: "Промышленное и гражданское строительство",
	},
	"тн техизоляция камин": {
		ConstructionType: "Индивидуальное жилищное строительство",
	},
}

func NewClassificationService(repo *repository.ClassificationRepository) *ClassificationService {
	return &ClassificationService{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
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
			Position:         len(rows) + 1,
			SystemName:       systemName,
			ConstructionType: unassignedConstructionType,
			ClassBefore:      classBefore,
			ClassAfter:       classAfter,
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

	borders := []excelize.Border{
		{Type: "left", Color: "000000", Style: 1},
		{Type: "top", Color: "000000", Style: 1},
		{Type: "right", Color: "000000", Style: 1},
		{Type: "bottom", Color: "000000", Style: 1},
	}
	headerStyle, err := file.NewStyle(&excelize.Style{
		Border:    borders,
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D9D9D9"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "1F2937"},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create excel header style: %w", err)
	}
	bodyStyle, err := file.NewStyle(&excelize.Style{
		Border:    borders,
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return nil, fmt.Errorf("create excel body style: %w", err)
	}
	if err := file.SetCellStyle(sheet, "A1", "C2", headerStyle); err != nil {
		return nil, fmt.Errorf("style excel header: %w", err)
	}
	_ = file.SetRowHeight(sheet, 1, 24)
	_ = file.SetRowHeight(sheet, 2, 22)

	for index, row := range rows {
		excelRow := index + 3
		_ = file.SetCellValue(sheet, fmt.Sprintf("A%d", excelRow), row.SystemName)
		_ = file.SetCellValue(sheet, fmt.Sprintf("B%d", excelRow), row.ClassBefore)
		_ = file.SetCellValue(sheet, fmt.Sprintf("C%d", excelRow), row.ClassAfter)
	}
	if len(rows) > 0 {
		lastRow := len(rows) + 2
		if err := file.SetCellStyle(sheet, "A3", fmt.Sprintf("C%d", lastRow), bodyStyle); err != nil {
			return nil, fmt.Errorf("style excel rows: %w", err)
		}
		for row := 3; row <= lastRow; row++ {
			_ = file.SetRowHeight(sheet, row, 20)
		}
	}

	_ = file.SetColWidth(sheet, "A", "A", 48)
	_ = file.SetColWidth(sheet, "B", "C", 28)
	_ = file.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      2,
		TopLeftCell: "A3",
		ActivePane:  "bottomLeft",
	})

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

			rows[rowIndex].SystemURL, rows[rowIndex].ConstructionType = s.resolveSystemData(ctx, rows[rowIndex].SystemName)
		}(index)
	}

	wg.Wait()
}

func (s *ClassificationService) resolveSystemData(ctx context.Context, systemName string) (string, string) {
	if systemURL, constructionType, found, err := s.repo.NavSystemData(ctx, systemName); err == nil && found {
		if normalizedType := normalizeConstructionType(constructionType); normalizedType != unassignedConstructionType {
			return systemURL, normalizedType
		}
	}
	if system, found := knownNAVSystemData(systemName); found {
		return system.URL, system.ConstructionType
	}

	searchURL := "https://nav.tn.ru/search/?q=" + url.QueryEscape(systemName)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return "", unassignedConstructionType
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	response, err := s.httpClient.Do(request)
	if err != nil {
		return "", unassignedConstructionType
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", unassignedConstructionType
	}

	return systemDataFromSearch(response.Body, systemName)
}

func systemDataFromSearch(body io.Reader, systemName string) (string, string) {
	document, err := html.Parse(io.LimitReader(body, 2*1024*1024))
	if err != nil {
		return "", unassignedConstructionType
	}

	wantedName := normalizeNAVLookupName(systemName)
	for _, anchor := range findNodes(document, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "a" && hasClass(node, "b-search-teaser__title")
	}) {
		if normalizeNAVLookupName(nodeText(anchor)) != wantedName {
			continue
		}

		link, err := url.Parse(attribute(anchor, "href"))
		if err != nil {
			continue
		}

		absolute := (&url.URL{Scheme: "https", Host: "nav.tn.ru"}).ResolveReference(link)
		if absolute.Hostname() != "nav.tn.ru" || !strings.HasPrefix(absolute.Path, "/systems/") {
			continue
		}

		absolute.RawQuery = ""
		absolute.Fragment = ""

		constructionType := unassignedConstructionType
		for parent := anchor.Parent; parent != nil; parent = parent.Parent {
			if parent.Type != html.ElementNode || !hasClass(parent, "b-search-teaser") {
				continue
			}
			segment := findNode(parent, func(node *html.Node) bool {
				return node.Type == html.ElementNode && hasClass(node, "b-search-teaser__constr_segment")
			})
			if segment != nil {
				constructionType = normalizeConstructionType(nodeText(segment))
			}
			break
		}

		return absolute.String(), constructionType
	}

	return "", unassignedConstructionType
}

func knownNAVSystemData(systemName string) (knownNAVSystem, bool) {
	system, found := knownNAVSystems[normalizeNAVLookupName(systemName)]
	return system, found
}

func normalizeNAVLookupName(value string) string {
	normalized := normalizeSystemName(value)
	return strings.TrimSpace(strings.TrimPrefix(normalized, "система "))
}

func normalizeConstructionType(value string) string {
	normalized := normalizeSystemName(value)
	switch {
	case normalized == "пгс" || strings.Contains(normalized, "промышлен") || strings.Contains(normalized, "гражданск"):
		return "Промышленное и гражданское строительство"
	case normalized == "ижс" || strings.Contains(normalized, "индивидуальн") || strings.Contains(normalized, "жилищн"):
		return "Индивидуальное жилищное строительство"
	case strings.Contains(normalized, "транспорт") || strings.Contains(normalized, "дорож"):
		return "Транспортное и дорожное строительство"
	case normalized == "сс" || strings.Contains(normalized, "специальн") || strings.Contains(normalized, "спецсооруж") || strings.Contains(normalized, "спец сооруж"):
		return "Специальные сооружения"
	default:
		return unassignedConstructionType
	}
}
