package models

type OperatorOutputTargetStyleRecord struct {
	ID           int64  `json:"id"`
	WorkDate     string `json:"workDate"`
	StyleName    string `json:"styleName"`
	ProcessName  string `json:"processName"`
	OutputTarget int    `json:"outputTarget"`
	UploadedAt   string `json:"uploadedAt,omitempty"`
}

type OperatorOutputTargetStyleImportRow struct {
	WorkDate     string `json:"workDate"`
	StyleName    string `json:"styleName"`
	ProcessName  string `json:"processName"`
	OutputTarget int    `json:"outputTarget"`
}

type OperatorOutputTargetStyleImportRequest struct {
	Rows []OperatorOutputTargetStyleImportRow `json:"rows"`
}

type OperatorOutputTargetStyleImportResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Total    int    `json:"total"`
	Upserted int    `json:"upserted"`
	Skipped  int    `json:"skipped"`
}
