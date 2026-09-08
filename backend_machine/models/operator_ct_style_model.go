package models

type OperatorCtStyleRecord struct {
	ID         int64   `json:"id"`
	StyleName  string  `json:"styleName"`
	ProcessName string `json:"processName"`
	CtTotal    float64 `json:"ctTotal"`
	UploadedAt string  `json:"uploadedAt,omitempty"`
}

type OperatorCtStyleImportRow struct {
	StyleName   string  `json:"styleName"`
	ProcessName string  `json:"processName"`
	CtTotal     float64 `json:"ctTotal"`
}

type OperatorCtStyleImportRequest struct {
	Rows []OperatorCtStyleImportRow `json:"rows"`
}

type OperatorCtStyleImportResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Total    int    `json:"total"`
	Upserted int    `json:"upserted"`
	Skipped  int    `json:"skipped"`
}
