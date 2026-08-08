package models

type ProductionStats struct {
	ProcSec      int64
	OkProcSec    int64
	Output       int64
	Cycles       int64
	Complete     int64
	Incomplete   int64
	AvgCycle     float64
	MinCycle     float64
	MaxCycle     float64
	SlowCycles   int64
	UniqueFiles  int64
	TopFile      string
	FirstProcess string
	LastProcess  string
}

type ProductivityRow struct {
	Date             string `json:"date"`
	UUID             string `json:"uuid"`
	TableName        string `json:"tableName"`
	NickName         string `json:"nickName"`
	OriginalNickName string `json:"originalNickName"`
	Location         string `json:"location"`
	Pic              string `json:"pic"`
	Spv              string `json:"spv"`
	IP               string `json:"ip"`
	MacType          string `json:"macType"`
	MacState         string `json:"macState"`

	ShiftCode string `json:"shiftCode,omitempty"`
	ShiftName string `json:"shiftName,omitempty"`

	RuntimeSec   int64   `json:"runtimeSec"`
	RuntimeHours float64 `json:"runtimeHours"`

	LossTimeSec   int64   `json:"lossTimeSec"`
	LossTimeHours float64 `json:"lossTimeHours"`

	ProductivityRaw float64 `json:"productivityRaw"`
	ProductivityPct float64 `json:"productivityPct"`
	Status          string  `json:"status"`
	MainSource      string  `json:"mainSource"`

	ProcSec    int64   `json:"procSec"`
	ProcHours  float64 `json:"procHours"`
	// ProcActualSec/Hours = process time langsung dari mUUID (tanpa iris runtime).
	ProcActualSec   int64   `json:"procActualSec"`
	ProcActualHours float64 `json:"procActualHours"`
	OkProcSec  int64   `json:"okProcSec"`
	Output     int64   `json:"output"`
	Cycles     int64   `json:"cycles"`
	Complete   int64   `json:"complete"`
	Incomplete int64   `json:"incomplete"`

	AvgCycle   float64 `json:"avgCycle"`
	MinCycle   float64 `json:"minCycle"`
	MaxCycle   float64 `json:"maxCycle"`
	SlowCycles int64   `json:"slowCycles"`

	UniqueFiles int64  `json:"uniqueFiles"`
	TopFile     string `json:"topFile"`

	AlarmCount int64  `json:"alarmCount"`
	AlarmTypes string `json:"alarmTypes"`

	FirstProcess string `json:"firstProcess"`
	LastProcess  string `json:"lastProcess"`

	// Alias lama agar dashboard versi sebelumnya tetap aman.
	MachineName       string  `json:"machine_name"`
	ProductiveSeconds int64   `json:"productive_seconds"`
	ProductiveHours   float64 `json:"productive_hours"`
	Category          string  `json:"category"`
	OutputOK          int64   `json:"output_ok"`
	TotalLog          int64   `json:"total_log"`
	FirstStart        string  `json:"first_start"`
	LastStart         string  `json:"last_start"`
}
