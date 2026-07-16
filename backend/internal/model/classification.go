package model

import "time"

type ClassificationChange struct {
	ID               int64     `json:"id"`
	OrderID          int64     `json:"orderId"`
	Position         int       `json:"position"`
	SystemName       string    `json:"systemName"`
	SystemURL        string    `json:"systemUrl"`
	ConstructionType string    `json:"constructionType"`
	ClassBefore      string    `json:"classBefore"`
	ClassAfter       string    `json:"classAfter"`
	ImportedAt       time.Time `json:"importedAt"`
}

type ClassificationStats struct {
	AddedSystems          int `json:"addedSystems"`
	Recommended           int `json:"recommended"`
	Allowed               int `json:"allowed"`
	ClassificationChanges int `json:"classificationChanges"`
}

type ClassificationList struct {
	Rows          []ClassificationChange `json:"rows"`
	Stats         ClassificationStats    `json:"stats"`
	BeforeOptions []string               `json:"beforeOptions"`
	AfterOptions  []string               `json:"afterOptions"`
}

type ClassificationFilter struct {
	OrderID          int64
	Query            string
	ConstructionType string
	ClassBefore      string
	ClassAfter       string
}
