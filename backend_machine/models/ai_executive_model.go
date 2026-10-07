package models

// AiExecutiveRequest body POST /api/ai/executive.
type AiExecutiveRequest struct {
	Date             string                   `json:"date"`
	CompareDate      string                   `json:"compareDate"`
	Shift            string                   `json:"shift"`
	Location         string                   `json:"location"`
	MachinesToday    []AiInsightMachineInput  `json:"machinesToday"`
	MachinesYesterday []AiInsightMachineInput `json:"machinesYesterday"`
	Insight          *AiInsightAnalyzeResponse `json:"insight,omitempty"`
}

// AiFloorHealthScore skor utilisasi mesin (0-100).
type AiFloorHealthScore struct {
	Score       int     `json:"score"`
	Label       string  `json:"label"` // Baik | Waspada | Kritis
	Tone        string  `json:"tone"`  // good | warn | critical
	AvgUtilPct  float64 `json:"avgUtilPct"`
	ActiveCount int     `json:"activeCount"`
	BadCount    int     `json:"badCount"`
	HighCount   int     `json:"highCount"`
	MediumCount int     `json:"mediumCount"`
	Summary     string  `json:"summary"`
}

// AiCompareMetric satu metrik banding.
type AiCompareMetric struct {
	Key       string  `json:"key"`
	Label     string  `json:"label"`
	Today     float64 `json:"today"`
	Yesterday float64 `json:"yesterday"`
	Delta     float64 `json:"delta"`
	Unit      string  `json:"unit,omitempty"`
	Better    string  `json:"better,omitempty"` // up | down | same
}

// AiExecutiveCompare hasil banding hari ini vs kemarin.
type AiExecutiveCompare struct {
	TodayDate     string            `json:"todayDate"`
	YesterdayDate string            `json:"yesterdayDate"`
	HasYesterday  bool              `json:"hasYesterday"`
	Metrics       []AiCompareMetric `json:"metrics"`
	Summary       string            `json:"summary"`
}

// AiExecutiveResponse paket executive untuk atasan.
type AiExecutiveResponse struct {
	Status      string              `json:"status"`
	GeneratedAt string              `json:"generatedAt"`
	WorkDate    string              `json:"workDate"`
	Health      AiFloorHealthScore  `json:"health"`
	Briefing    string              `json:"briefing"`
	Compare     AiExecutiveCompare  `json:"compare"`
	TopActions  []string            `json:"topActions"`
}
