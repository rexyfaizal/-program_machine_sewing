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
