package models

type OperatorCtRecord struct {
	ID         int64   `json:"id"`
	UUID       string  `json:"uuid"`
	Line       string  `json:"line"`
	Area       string  `json:"area,omitempty"`
	CtSum      float64 `json:"ctSum"`
	CtStd      float64 `json:"ctStd"`
	CtValue    float64 `json:"ctValue"`
	UploadedAt string  `json:"uploadedAt,omitempty"`
}

type OperatorCtImportRow struct {
	UUID    string  `json:"uuid"`
	Line    string  `json:"line"`
	Area    string  `json:"area,omitempty"`
	CtSum   float64 `json:"ctSum"`
	CtStd   float64 `json:"ctStd"`
	CtValue float64 `json:"ctValue"`
}

type OperatorCtImportRequest struct {
	Rows []OperatorCtImportRow `json:"rows"`
}

type OperatorCtImportResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Total    int    `json:"total"`
	Upserted int    `json:"upserted"`
	Skipped  int    `json:"skipped"`
}
