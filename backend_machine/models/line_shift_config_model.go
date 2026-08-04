package models

// ShiftScheduleItem adalah satu shift dalam schedule_json.
type ShiftScheduleItem struct {
	Code       string `json:"code"`
	Start      string `json:"start"`
	End        string `json:"end"`
	BreakStart string `json:"breakStart,omitempty"`
	BreakEnd   string `json:"breakEnd,omitempty"`
}

// LineShiftConfig adalah konfigurasi shift per line.
type LineShiftConfig struct {
	Factory      string              `json:"factory"`
	LineName     string              `json:"lineName"`
	Enabled      bool                `json:"enabled"`
	Schedule     []ShiftScheduleItem `json:"schedule"`
	ScheduleJSON string              `json:"-"`
	UpdatedAt    string              `json:"updatedAt,omitempty"`
}

// LineShiftConfigPutRequest body PUT /api/line-shift-config.
type LineShiftConfigPutRequest struct {
	Factory string            `json:"factory"`
	Lines   []LineShiftConfig `json:"lines"`
}
