package models

// AiInsightMachineInput snapshot mesin dari dashboard (tanpa output/tiket).
type AiInsightMachineInput struct {
	UUID            string  `json:"uuid"`
	MachineName     string  `json:"machineName"`
	Location        string  `json:"location"`
	MacState        string  `json:"macState"`
	Status          string  `json:"status"`
	ProductivityPct float64 `json:"productivityPct"`
	RuntimeSec      int64   `json:"runtimeSec"`
	ProcSec         int64   `json:"procSec"`
	LossTimeSec     int64   `json:"lossTimeSec"`
	OperatorName    string  `json:"operatorName,omitempty"`
	OperatorNotes   string  `json:"operatorNotes,omitempty"`
}

// AiInsightAnalyzeRequest body POST /api/ai/insight.
type AiInsightAnalyzeRequest struct {
	Date     string                  `json:"date"`
	Shift    string                  `json:"shift"`
	Location string                  `json:"location"`
	Machines []AiInsightMachineInput `json:"machines"`
}

// AiInsightItem satu rekomendasi keputusan.
type AiInsightItem struct {
	Priority        string  `json:"priority"` // HIGH | MEDIUM | LOW
	Scope           string  `json:"scope,omitempty"` // LINE | MACHINE
	Code            string  `json:"code"`
	UUID            string  `json:"uuid,omitempty"`
	MachineName     string  `json:"machineName,omitempty"`
	Location        string  `json:"location,omitempty"`
	Signal          string  `json:"signal"`
	Evidence        string  `json:"evidence,omitempty"`
	Prediction1h    string  `json:"prediction1h"`
	Improvement     string  `json:"improvement"`
	Decision        string  `json:"decision,omitempty"`
	RiskScore       int     `json:"riskScore"`
	UtilPct         float64 `json:"utilPct,omitempty"`
	LossReasonCode  string  `json:"lossReasonCode,omitempty"`
	LossReasonLabel string  `json:"lossReasonLabel,omitempty"`
	OperatorName    string  `json:"operatorName,omitempty"`
	NoteSummary     string  `json:"noteSummary,omitempty"`
	BrokenCount7d   int     `json:"brokenCount7d,omitempty"`
}

// AiInsightAnalyzeResponse hasil analisa AI rule-based.
type AiInsightAnalyzeResponse struct {
	Status       string          `json:"status"`
	GeneratedAt  string          `json:"generatedAt"`
	WorkDate     string          `json:"workDate"`
	Shift        string          `json:"shift,omitempty"`
	Location     string          `json:"location,omitempty"`
	MachineCount int             `json:"machineCount"`
	HighCount    int             `json:"highCount"`
	MediumCount  int             `json:"mediumCount"`
	LowCount     int             `json:"lowCount"`
	Summary      string          `json:"summary"`
	Items        []AiInsightItem `json:"items"`
	LogID        int64           `json:"logId,omitempty"`
}

// AiInsightActiveSession sesi operator aktif untuk pengayaan insight.
type AiInsightActiveSession struct {
	UUID         string
	Location     string
	OperatorNIK  string
	OperatorName string
	StyleName    string
	ProcessName  string
	LoginTime    string
	Status       string
}

// AiInsightActiveLoss loss event aktif untuk pengayaan insight.
type AiInsightActiveLoss struct {
	UUID         string
	OperatorName string
	ReasonCode   string
	ReasonLabel  string
	Note         string
	StartTime    string
	DurationSec  int64
	Status       string
}

// AiInsightRecentNote catatan operator hari ini.
type AiInsightRecentNote struct {
	UUID         string
	OperatorName string
	ReasonCode   string
	ReasonName   string
	Note         string
	CreatedAt    string
}

// AiInsightBrokenFrequency frekuensi MACHINE_BROKEN per mesin dalam jendela hari.
type AiInsightBrokenFrequency struct {
	UUID  string
	Count int
}
