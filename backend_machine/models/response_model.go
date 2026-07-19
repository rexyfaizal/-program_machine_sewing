package models

type APIResponse struct {
	Summary Summary           `json:"summary"`
	Rows    []ProductivityRow `json:"rows"`
}

type ProcessDetailResponse struct {
	Date    string          `json:"date"`
	Machine ProductivityRow `json:"machine"`
	Groups  []ProcessGroup  `json:"groups"`
	Hours   []HourBar       `json:"hours"`
	Alarms  []AlarmGroup    `json:"alarms"`
	Events  []ProcessCycle  `json:"events"`
}
