package models

type ProcessCycle struct {
	No             int     `json:"no"`
	FileName       string  `json:"fileName"`
	StartTime      string  `json:"startTime"`
	EndTime        string  `json:"endTime"`
	ProcSec        int64   `json:"procSec"`
	ProcTime       string  `json:"procTime"`
	ProcCounts     int64   `json:"procCounts"`
	EndStitch      int64   `json:"endStitch"`
	FileStitches   int64   `json:"fileStitches"`
	NodeDistance   int64   `json:"nodeDistance"`
	Status         string  `json:"status"`
	AbnormalReason string  `json:"abnormalReason"`
	SPM            float64 `json:"spm"`
}

type ProcessGroup struct {
	FileName     string  `json:"fileName"`
	Output       int64   `json:"output"`
	Cycles       int64   `json:"cycles"`
	Complete     int64   `json:"complete"`
	Incomplete   int64   `json:"incomplete"`
	ProcSec      int64   `json:"procSec"`
	ProcTime     string  `json:"procTime"`
	AvgCycle     float64 `json:"avgCycle"`
	MaxCycle     int64   `json:"maxCycle"`
	FirstProcess string  `json:"firstProcess"`
	LastProcess  string  `json:"lastProcess"`
}

type HourBar struct {
	Hour     string `json:"hour"`
	ProcSec  int64  `json:"procSec"`
	ProcTime string `json:"procTime"`
	Output   int64  `json:"output"`
	Cycles   int64  `json:"cycles"`
}
