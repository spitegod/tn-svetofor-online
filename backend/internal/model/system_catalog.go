package model

import "time"

type SystemCatalogRow struct {
	ID          int64     `json:"id"`
	OrderID     int64     `json:"orderId"`
	Position    int       `json:"position"`
	Code        string    `json:"code"`
	SystemName  string    `json:"systemName"`
	SystemURL   string    `json:"systemUrl"`
	SystemClass string    `json:"systemClass"`
	Curator     string    `json:"curator"`
	ImportedAt  time.Time `json:"importedAt"`
}

type SystemCatalogStats struct {
	Total       int `json:"total"`
	Recommended int `json:"recommended"`
	Allowed     int `json:"allowed"`
	Forbidden   int `json:"forbidden"`
	Curators    int `json:"curators"`
}

type SystemCatalogList struct {
	Rows           []SystemCatalogRow `json:"rows"`
	Stats          SystemCatalogStats `json:"stats"`
	ClassOptions   []string           `json:"classOptions"`
	CuratorOptions []string           `json:"curatorOptions"`
}

type SystemCatalogFilter struct {
	OrderID     int64
	Query       string
	SystemClass string
	Curator     string
}
