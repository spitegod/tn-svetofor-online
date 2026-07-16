package model

import "time"

type SystemDocumentRow struct {
	ID                 int64                  `json:"id"`
	OrderID            int64                  `json:"orderId"`
	OrderName          string                 `json:"orderName"`
	SystemCatalogID    int64                  `json:"systemCatalogId"`
	Position           int                    `json:"position"`
	Code               string                 `json:"code"`
	SystemName         string                 `json:"systemName"`
	SystemURL          string                 `json:"systemUrl"`
	SystemClass        string                 `json:"systemClass"`
	Curator            string                 `json:"curator"`
	ComparisonSelected bool                   `json:"comparisonSelected"`
	Comment            string                 `json:"comment"`
	AttachmentName     string                 `json:"attachmentName"`
	AttachmentType     string                 `json:"attachmentType"`
	AttachmentSize     int64                  `json:"attachmentSize"`
	CreatedAt          time.Time              `json:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
	Characteristics    []SystemCharacteristic `json:"characteristics"`
}

type SystemDocumentAttachment struct {
	Name        string
	ContentType string
	Size        int64
	Data        []byte
}

type SystemDocumentList struct {
	Rows           []SystemDocumentRow `json:"rows"`
	Stats          SystemCatalogStats  `json:"stats"`
	ClassOptions   []string            `json:"classOptions"`
	CuratorOptions []string            `json:"curatorOptions"`
}

type SystemDocumentFilter struct {
	OrderID          int64
	Query            string
	SystemClass      string
	Curator          string
	ConstructionType string
	SystemType       string
	ComparisonOnly   bool
}

type SystemDocumentKey struct {
	Code       string `json:"code"`
	SystemName string `json:"systemName"`
}
