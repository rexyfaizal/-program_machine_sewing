package models

type ProcessStyleItem struct {
	StyleName string `json:"styleName"`
}

type ProcessStyleProcess struct {
	ID          int64  `json:"id"`
	ProcessName string `json:"processName"`
	StyleName   string `json:"styleName"`
}

type ProcessStyleRecord struct {
	ID          int64  `json:"id"`
	ProcessName string `json:"processName"`
	StyleName   string `json:"styleName"`
	CreatedAt   string `json:"createdAt"`
}

type ProcessStyleRequest struct {
	ProcessName string `json:"processName"`
	StyleName   string `json:"styleName"`

	// Alias supaya bisa terima body JSON dengan nama kolom asli juga.
	Proses string `json:"proses"`
	Style  string `json:"style"`
}

type ProcessStyleImportRow struct {
	Style       string `json:"style"`
	ProcessName string `json:"processName"`
	Proses      string `json:"proses,omitempty"`
}

type ProcessStyleImportRequest struct {
	Rows []ProcessStyleImportRow `json:"rows"`
}

type ProcessStyleImportResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Total    int    `json:"total"`
	Inserted int    `json:"inserted"`
	Skipped  int    `json:"skipped"`
}
