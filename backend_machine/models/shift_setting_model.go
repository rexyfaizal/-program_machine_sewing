package models

// ShiftSetting adalah baris konfigurasi jam shift per area (dbo.shift_setting).
// LineName kosong ("") = jadwal DEFAULT area; terisi = override khusus line.
type ShiftSetting struct {
	ID            int64  `json:"id"`
	Area          string `json:"area"`
	LineName      string `json:"lineName"`
	ShiftNo       int    `json:"shiftNo"`
	ShiftName     string `json:"shiftName"`
	StartTime     string `json:"startTime"` // HH:mm atau HH:mm:ss
	EndTime       string `json:"endTime"`
	BreakStart    string `json:"breakStart,omitempty"`
	BreakEnd      string `json:"breakEnd,omitempty"`
	EffectiveFrom string `json:"effectiveFrom,omitempty"`
	EffectiveTo   string `json:"effectiveTo,omitempty"`
	IsActive      bool   `json:"isActive"`
	UpdatedAt     string `json:"updatedAt,omitempty"`
}

// ShiftSettingLineInput adalah konfigurasi 1 line pada PUT /api/shift-settings.
// - enabled=false → line Normal (hari penuh)
// - enabled=true & custom=false → pakai jadwal default area
// - enabled=true & custom=true → pakai shifts sendiri (override)
type ShiftSettingLineInput struct {
	LineName string         `json:"lineName"`
	Enabled  bool           `json:"enabled"`
	Custom   bool           `json:"custom"`
	Shifts   []ShiftSetting `json:"shifts"`
}

// ShiftSettingPutRequest body PUT /api/shift-settings.
// shifts = jadwal default area; lines = mode + override per line.
type ShiftSettingPutRequest struct {
	Area   string                  `json:"area"`
	Shifts []ShiftSetting          `json:"shifts"`
	Lines  []ShiftSettingLineInput `json:"lines"`
}

// FinalProductivityGroup hasil agregat per NORMAL / per shift dari rumus FINAL.
type FinalProductivityGroup struct {
	Mode           string
	WorkDate       string
	Area           string
	ShiftName      string
	PowerSeconds   int64
	ProcessSeconds int64
	LossSeconds    int64
	Productivity   float64
	PeriodStart    string
	PeriodEnd      string
	BreakStart     string
	BreakEnd       string
}
