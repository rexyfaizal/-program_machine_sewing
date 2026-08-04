package models

type Summary struct {
	Date        string  `json:"date"`
	Total       int     `json:"total"`
	Good        int     `json:"good"`
	Normal      int     `json:"normal"`
	Bad         int     `json:"bad"`
	AvgPct      float64 `json:"avg_pct"`
	WorkHours   float64 `json:"work_hours"`
	TotalOutput int64   `json:"total_output"`
	TotalAlarm  int64   `json:"total_alarm"`
	TotalProc   int64   `json:"total_proc_sec"`
	TotalRun    int64   `json:"total_runtime_sec"`

	ShiftCode string `json:"shiftCode,omitempty"`
	ShiftName string `json:"shiftName,omitempty"`
}
