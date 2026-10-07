package utils

import (
	"fmt"
	"math"
	"strings"
	"time"

	"backend_machine/models"
)

type floorSnapshot struct {
	ActiveCount int
	BadCount    int
	AvgUtilPct  float64
	HighCount   int
	MediumCount int
}

func activeMachineRows(machines []models.AiInsightMachineInput) []models.AiInsightMachineInput {
	out := make([]models.AiInsightMachineInput, 0, len(machines))
	for _, m := range machines {
		if m.RuntimeSec <= 0 && m.ProcSec <= 0 && m.ProductivityPct <= 0 {
			continue
		}
		out = append(out, m)
	}
	return out
}

func machineUtilForExecutive(m models.AiInsightMachineInput) float64 {
	return resolveAiUtilPct(m)
}

func buildFloorSnapshot(machines []models.AiInsightMachineInput, insight *models.AiInsightAnalyzeResponse) floorSnapshot {
	rows := activeMachineRows(machines)
	snap := floorSnapshot{ActiveCount: len(rows)}
	if len(rows) == 0 {
		if insight != nil {
			snap.HighCount = insight.HighCount
			snap.MediumCount = insight.MediumCount
		}
		return snap
	}

	sum := 0.0
	for _, m := range rows {
		sum += machineUtilForExecutive(m)
		if strings.EqualFold(strings.TrimSpace(m.Status), "BAD") {
			snap.BadCount++
		}
	}
	snap.AvgUtilPct = math.Round((sum/float64(len(rows)))*10) / 10

	if insight != nil {
		snap.HighCount = insight.HighCount
		snap.MediumCount = insight.MediumCount
	}
	return snap
}

// BuildFloorHealthScore menghitung skor 0-100 untuk utilisasi mesin.
func BuildFloorHealthScore(machines []models.AiInsightMachineInput, insight *models.AiInsightAnalyzeResponse) models.AiFloorHealthScore {
	snap := buildFloorSnapshot(machines, insight)
	score := 100.0

	if snap.ActiveCount == 0 {
		return models.AiFloorHealthScore{
			Score:       0,
			Label:       "Kritis",
			Tone:        "critical",
			AvgUtilPct:  0,
			ActiveCount: 0,
			BadCount:    0,
			HighCount:   snap.HighCount,
			MediumCount: snap.MediumCount,
			Summary:     "Belum ada mesin aktif pada filter saat ini.",
		}
	}

	// Prioritas Insight menekan skor.
	score -= math.Min(float64(snap.HighCount)*4.0, 40)
	score -= math.Min(float64(snap.MediumCount)*2.0, 16)

	// Proporsi BAD.
	badPct := float64(snap.BadCount) * 100.0 / float64(snap.ActiveCount)
	score -= math.Min(badPct*0.35, 25)

	// Util rata-rata rendah.
	if snap.AvgUtilPct < 70 {
		score -= math.Min((70-snap.AvgUtilPct)*0.6, 24)
	} else if snap.AvgUtilPct >= 85 {
		score += 4
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	finalScore := int(math.Round(score))

	label := "Baik"
	tone := "good"
	if finalScore < 60 {
		label = "Kritis"
		tone = "critical"
	} else if finalScore < 80 {
		label = "Waspada"
		tone = "warn"
	}

	summary := fmt.Sprintf(
		"Skor utilitas mesin %d (%s): util rata-rata %.1f%%, %d mesin aktif, %d BAD, %d prioritas segera.",
		finalScore,
		label,
		snap.AvgUtilPct,
		snap.ActiveCount,
		snap.BadCount,
		snap.HighCount,
	)

	return models.AiFloorHealthScore{
		Score:       finalScore,
		Label:       label,
		Tone:        tone,
		AvgUtilPct:  snap.AvgUtilPct,
		ActiveCount: snap.ActiveCount,
		BadCount:    snap.BadCount,
		HighCount:   snap.HighCount,
		MediumCount: snap.MediumCount,
		Summary:     summary,
	}
}

func metricBetter(delta float64, higherIsBetter bool) string {
	if math.Abs(delta) < 0.05 {
		return "same"
	}
	up := delta > 0
	if higherIsBetter {
		if up {
			return "up"
		}
		return "down"
	}
	if up {
		return "down"
	}
	return "up"
}

// BuildExecutiveCompare bandingkan hari ini vs kemarin.
func BuildExecutiveCompare(
	todayDate, yesterdayDate string,
	todayMachines, yesterdayMachines []models.AiInsightMachineInput,
	insight *models.AiInsightAnalyzeResponse,
) models.AiExecutiveCompare {
	today := buildFloorSnapshot(todayMachines, insight)
	yesterday := buildFloorSnapshot(yesterdayMachines, nil)
	hasYesterday := len(activeMachineRows(yesterdayMachines)) > 0

	metrics := []models.AiCompareMetric{
		{
			Key:       "avgUtil",
			Label:     "Util rata-rata",
			Today:     today.AvgUtilPct,
			Yesterday: yesterday.AvgUtilPct,
			Delta:     math.Round((today.AvgUtilPct-yesterday.AvgUtilPct)*10) / 10,
			Unit:      "%",
			Better:    metricBetter(today.AvgUtilPct-yesterday.AvgUtilPct, true),
		},
		{
			Key:       "active",
			Label:     "Mesin aktif",
			Today:     float64(today.ActiveCount),
			Yesterday: float64(yesterday.ActiveCount),
			Delta:     float64(today.ActiveCount - yesterday.ActiveCount),
			Unit:      "",
			Better:    metricBetter(float64(today.ActiveCount-yesterday.ActiveCount), true),
		},
		{
			Key:       "bad",
			Label:     "Status BAD",
			Today:     float64(today.BadCount),
			Yesterday: float64(yesterday.BadCount),
			Delta:     float64(today.BadCount - yesterday.BadCount),
			Unit:      "",
			Better:    metricBetter(float64(today.BadCount-yesterday.BadCount), false),
		},
		{
			Key:       "high",
			Label:     "Prioritas segera (hari ini)",
			Today:     float64(today.HighCount),
			Yesterday: 0,
			Delta:     float64(today.HighCount),
			Unit:      "",
			Better:    metricBetter(float64(today.HighCount), false),
		},
	}

	summary := "Data kemarin belum tersedia untuk dibandingkan."
	if hasYesterday {
		utilDelta := today.AvgUtilPct - yesterday.AvgUtilPct
		badDelta := today.BadCount - yesterday.BadCount
		utilWord := "stabil"
		if utilDelta > 0.5 {
			utilWord = fmt.Sprintf("naik %.1f poin", utilDelta)
		} else if utilDelta < -0.5 {
			utilWord = fmt.Sprintf("turun %.1f poin", -utilDelta)
		}
		badWord := "jumlah BAD sama"
		if badDelta > 0 {
			badWord = fmt.Sprintf("BAD bertambah %d", badDelta)
		} else if badDelta < 0 {
			badWord = fmt.Sprintf("BAD berkurang %d", -badDelta)
		}
		summary = fmt.Sprintf(
			"Vs kemarin (%s): util %s, %s, mesin aktif %d → %d.",
			yesterdayDate,
			utilWord,
			badWord,
			yesterday.ActiveCount,
			today.ActiveCount,
		)
	}

	return models.AiExecutiveCompare{
		TodayDate:     todayDate,
		YesterdayDate: yesterdayDate,
		HasYesterday:  hasYesterday,
		Metrics:       metrics,
		Summary:       summary,
	}
}

func buildTopActions(insight *models.AiInsightAnalyzeResponse, limit int) []string {
	if insight == nil || len(insight.Items) == 0 || limit <= 0 {
		return []string{}
	}
	if limit > len(insight.Items) {
		limit = len(insight.Items)
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		item := insight.Items[i]
		title := strings.TrimSpace(item.Location)
		if title == "" {
			title = strings.TrimSpace(item.MachineName)
		}
		if title == "" {
			title = item.Code
		}
		step := strings.TrimSpace(item.Improvement)
		if step == "" {
			step = strings.TrimSpace(item.Decision)
		}
		if step == "" {
			step = strings.TrimSpace(item.Signal)
		}
		out = append(out, fmt.Sprintf("[%s] %s — %s", item.Priority, title, step))
	}
	return out
}

// BuildExecutiveBriefing teks briefing 1-klik untuk atasan.
func BuildExecutiveBriefing(
	workDate string,
	health models.AiFloorHealthScore,
	compare models.AiExecutiveCompare,
	topActions []string,
	insight *models.AiInsightAnalyzeResponse,
) string {
	now := currentAiClock()
	var b strings.Builder

	b.WriteString("BRIEFING ATASAN — AI Produksi\n")
	b.WriteString(fmt.Sprintf("%s, %s · update %s\n\n", now.DayNameId, workDate, now.DateTime))

	b.WriteString(fmt.Sprintf("1) Skor Utilitas Mesin: %d/100 (%s)\n", health.Score, health.Label))
	b.WriteString(fmt.Sprintf("   %s\n\n", health.Summary))

	b.WriteString("2) Vs Kemarin\n")
	b.WriteString(fmt.Sprintf("   %s\n\n", compare.Summary))

	if insight != nil && strings.TrimSpace(insight.Summary) != "" {
		b.WriteString("3) Situasi Utilitas Mesin\n")
		b.WriteString("   " + strings.TrimSpace(insight.Summary) + "\n\n")
	} else {
		b.WriteString("3) Situasi Utilitas Mesin\n")
		b.WriteString(fmt.Sprintf(
			"   %d mesin aktif dipantau; fokusatkan supervise pada mesin util rendah / BAD.\n\n",
			health.ActiveCount,
		))
	}

	b.WriteString("4) Prioritas 1 Jam ke Depan\n")
	if len(topActions) == 0 {
		b.WriteString("   Tidak ada prioritas mendesak. Pertahankan ritme produksi.\n")
	} else {
		for i, action := range topActions {
			b.WriteString(fmt.Sprintf("   %d. %s\n", i+1, action))
		}
	}

	b.WriteString("\nCatatan: briefing disusun dari data dashboard + AI Insight (tanpa output/tiket).")
	return strings.TrimSpace(b.String())
}

// BuildAiExecutiveResponse paket lengkap untuk Mode Presentasi / Copilot.
func BuildAiExecutiveResponse(req models.AiExecutiveRequest) models.AiExecutiveResponse {
	workDate := strings.TrimSpace(req.Date)
	if workDate == "" {
		workDate = time.Now().In(jakartaLocation()).Format("2006-01-02")
	}
	compareDate := strings.TrimSpace(req.CompareDate)
	if compareDate == "" {
		t, err := time.ParseInLocation("2006-01-02", workDate, jakartaLocation())
		if err != nil {
			t = time.Now().In(jakartaLocation())
		}
		compareDate = t.AddDate(0, 0, -1).Format("2006-01-02")
	}

	health := BuildFloorHealthScore(req.MachinesToday, req.Insight)
	compare := BuildExecutiveCompare(workDate, compareDate, req.MachinesToday, req.MachinesYesterday, req.Insight)
	topActions := buildTopActions(req.Insight, 5)
	briefing := BuildExecutiveBriefing(workDate, health, compare, topActions, req.Insight)

	return models.AiExecutiveResponse{
		Status:      "ok",
		GeneratedAt: time.Now().In(jakartaLocation()).Format("2006-01-02 15:04:05"),
		WorkDate:    workDate,
		Health:      health,
		Briefing:    briefing,
		Compare:     compare,
		TopActions:  topActions,
	}
}
