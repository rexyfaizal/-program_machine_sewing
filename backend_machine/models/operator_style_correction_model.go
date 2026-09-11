package models

type OperatorStyleCorrectionImportRow struct {
	WorkDate        string `json:"workDate"`
	UUID            string `json:"uuid"`
	Line            string `json:"line,omitempty"`
	OperatorNIK     string `json:"operatorNik,omitempty"`
	StyleName       string `json:"styleName"`
	ProcessName     string `json:"processName,omitempty"`
	PreviousStyle   string `json:"previousStyle,omitempty"`
	PreviousProcess string `json:"previousProcess,omitempty"`
}

type OperatorStyleCorrectionImportRequest struct {
	Rows []OperatorStyleCorrectionImportRow `json:"rows"`
}

type OperatorStyleCorrectionImportResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Total    int    `json:"total"`
	Updated  int    `json:"updated"`
	Skipped  int    `json:"skipped"`
}
