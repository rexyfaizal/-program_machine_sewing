package models

type OperatorOutputTargetRecord struct {
	ID           int64  `json:"id"`
	WorkDate     string `json:"workDate"`
	UUID         string `json:"uuid"`
	Line         string `json:"line"`
	Area         string `json:"area,omitempty"`
	OutputTarget int    `json:"outputTarget"`
	UploadedAt   string `json:"uploadedAt,omitempty"`
}

type OperatorOutputTargetImportRow struct {
	WorkDate     string `json:"workDate"`
	UUID         string `json:"uuid"`
	Line         string `json:"line"`
	Area         string `json:"area,omitempty"`
	OutputTarget int    `json:"outputTarget"`
}

type OperatorOutputTargetImportRequest struct {
	Rows []OperatorOutputTargetImportRow `json:"rows"`
}

type OperatorOutputTargetImportResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Total    int    `json:"total"`
	Upserted int    `json:"upserted"`
	Skipped  int    `json:"skipped"`
}
