package models

type MachineSetting struct {
	UUID       string `json:"uuid"`
	CustomName string `json:"customName"`
	Location   string `json:"location"`
	MachineID  string `json:"machineId"`
	Pic        string `json:"pic"`
	Spv        string `json:"spv"`
	UpdatedAt  string `json:"updatedAt"`
}

type MachineIDImportRow struct {
	UUID      string `json:"uuid"`
	MachineID string `json:"machineId"`
	ExcelRow  int    `json:"excelRow,omitempty"`
}

type MachineIDImportRequest struct {
	Rows []MachineIDImportRow `json:"rows"`
}

type MachineIDImportError struct {
	ExcelRow  int    `json:"excelRow,omitempty"`
	UUID      string `json:"uuid,omitempty"`
	MachineID string `json:"machineId,omitempty"`
	Message   string `json:"message"`
}

type MachineIDImportResponse struct {
	Status   string                 `json:"status"`
	Message  string                 `json:"message"`
	Total    int                    `json:"total"`
	Upserted int                    `json:"upserted"`
	Skipped  int                    `json:"skipped"`
	Errors   []MachineIDImportError `json:"errors,omitempty"`
}
