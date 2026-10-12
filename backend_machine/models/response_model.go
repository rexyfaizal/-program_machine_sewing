package models

type APIResponse struct {
	Summary Summary           `json:"summary"`
	Rows    []ProductivityRow `json:"rows"`
}

type ProductivityRangeResponse struct {
	StartDate string        `json:"start_date"`
	EndDate   string        `json:"end_date"`
	TotalDays int           `json:"total_days"`
	Days      []APIResponse `json:"days"`
}

type ProcessDetailResponse struct {
	Date    string          `json:"date"`
	Machine ProductivityRow `json:"machine"`
	Groups  []ProcessGroup  `json:"groups"`
	Hours   []HourBar       `json:"hours"`
	Alarms  []AlarmGroup    `json:"alarms"`
	Events  []ProcessCycle  `json:"events"`
}
