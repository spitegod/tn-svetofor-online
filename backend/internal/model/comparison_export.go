package model

type ComparisonExport struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}
