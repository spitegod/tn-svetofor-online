package model

import "time"

type SystemCatalogRow struct {
	ID              int64                  `json:"id"`
	OrderID         int64                  `json:"orderId"`
	Position        int                    `json:"position"`
	Code            string                 `json:"code"`
	SystemName      string                 `json:"systemName"`
	SystemURL       string                 `json:"systemUrl"`
	SystemClass     string                 `json:"systemClass"`
	Curator         string                 `json:"curator"`
	ImportedAt      time.Time              `json:"importedAt"`
	Characteristics []SystemCharacteristic `json:"characteristics"`
}

type SystemCharacteristic struct {
	Position int    `json:"position"`
	Name     string `json:"name"`
	Value    string `json:"value"`
}

type SystemTypeOption struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

type NavParseReport struct {
	Total    int      `json:"total"`
	Found    int      `json:"found"`
	Updated  int      `json:"updated"`
	Failed   int      `json:"failed"`
	NotFound []string `json:"notFound"`
}

type NavParserSettings struct {
	UpdateIntervalDays int        `json:"updateIntervalDays"`
	LastRunAt          *time.Time `json:"lastRunAt"`
	NextRunAt          *time.Time `json:"nextRunAt"`
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
	SystemTypes    []SystemTypeOption `json:"systemTypes"`
}

type SystemCatalogFilter struct {
	OrderID     int64
	Query       string
	SystemClass string
	Curator     string
}
