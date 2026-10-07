package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/utils"
)

// AiInsightAnalyze godoc
// @Summary Analisa AI insight (rule-based)
// @Description Hitung sinyal risiko 1 jam ke depan + rekomendasi improvement. Tidak memakai output/tiket.
// @Tags AI Insight
// @Accept json
// @Produce json
// @Param body body models.AiInsightAnalyzeRequest true "Snapshot mesin dashboard"
// @Success 200 {object} models.AiInsightAnalyzeResponse
// @Router /ai/insight [post]
func (h *Handler) AiInsightAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 45*time.Second)
	defer cancel()

	if err := h.Repo.EnsureAiInsightSchema(ctx); err != nil {
		http.Error(w, "Gagal prepare schema AI insight: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var input models.AiInsightAnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	workDate := strings.TrimSpace(input.Date)
	if workDate == "" {
		workDate = time.Now().Format("2006-01-02")
	}
	shift := strings.TrimSpace(input.Shift)
	location := strings.TrimSpace(input.Location)
	if strings.EqualFold(location, "ALL") {
		location = ""
	}

	// Session/loss aktif hanya relevan untuk tanggal hari ini — hindari campur data live ke filter historis.
	today := time.Now().Format("2006-01-02")
	sessions := []models.AiInsightActiveSession{}
	losses := []models.AiInsightActiveLoss{}
	if workDate == today {
		var err error
		sessions, err = h.Repo.ListAiInsightActiveSessions(ctx)
		if err != nil {
			log.Println("AI insight sessions:", err)
			sessions = []models.AiInsightActiveSession{}
		}

		losses, err = h.Repo.ListAiInsightActiveLosses(ctx)
		if err != nil {
			log.Println("AI insight losses:", err)
			losses = []models.AiInsightActiveLoss{}
		}
	}

	notes, err := h.Repo.ListAiInsightRecentNotes(ctx, workDate)
	if err != nil {
		log.Println("AI insight notes:", err)
		notes = []models.AiInsightRecentNote{}
	}

	brokenFreq, err := h.Repo.ListAiInsightBrokenFrequency(ctx, workDate, utils.AiBrokenWindowDays)
	if err != nil {
		log.Println("AI insight broken frequency:", err)
		brokenFreq = []models.AiInsightBrokenFrequency{}
	}

	items := utils.BuildAiInsightItems(input.Machines, sessions, losses, notes, brokenFreq, location)
	machineCount := countAiMachinesInScope(input.Machines, location)
	summary := utils.BuildAiInsightSummary(items, machineCount)

	high, medium, low := 0, 0, 0
	for _, item := range items {
		switch item.Priority {
		case utils.AiPriorityHigh:
			high++
		case utils.AiPriorityMedium:
			medium++
		default:
			low++
		}
	}

	resp := models.AiInsightAnalyzeResponse{
		Status:       "ok",
		GeneratedAt:  time.Now().Format("2006-01-02 15:04:05"),
		WorkDate:     workDate,
		Shift:        shift,
		Location:     location,
		MachineCount: machineCount,
		HighCount:    high,
		MediumCount:  medium,
		LowCount:     low,
		Summary:      summary,
		Items:        items,
	}

	logID, err := h.Repo.SaveAiInsightLog(ctx, resp)
	if err != nil {
		log.Println("AI insight log gagal disimpan:", err)
	} else {
		resp.LogID = logID
	}

	writeJSON(w, resp)
}

// AiInsightLatest godoc
// @Summary Ambil log AI insight terbaru
// @Tags AI Insight
// @Produce json
// @Param date query string false "Tanggal kerja YYYY-MM-DD"
// @Param location query string false "Filter location/area"
// @Success 200 {object} models.AiInsightAnalyzeResponse
// @Router /ai/insight/latest [get]
func (h *Handler) AiInsightLatest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 20*time.Second)
	defer cancel()

	if err := h.Repo.EnsureAiInsightSchema(ctx); err != nil {
		http.Error(w, "Gagal prepare schema AI insight: "+err.Error(), http.StatusInternalServerError)
		return
	}

	date := strings.TrimSpace(r.URL.Query().Get("date"))
	location := strings.TrimSpace(r.URL.Query().Get("location"))
	if strings.EqualFold(location, "ALL") {
		location = ""
	}

	resp, found, err := h.Repo.GetLatestAiInsightLog(ctx, date, location)
	if err != nil {
		http.Error(w, "Gagal ambil log AI insight: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !found {
		writeJSON(w, models.AiInsightAnalyzeResponse{
			Status:      "empty",
			GeneratedAt: "",
			WorkDate:    date,
			Location:    location,
			Summary:     "Belum ada hasil AI insight untuk filter ini.",
			Items:       []models.AiInsightItem{},
		})
		return
	}

	writeJSON(w, resp)
}

func countAiMachinesInScope(machines []models.AiInsightMachineInput, locationFilter string) int {
	locationFilter = strings.ToUpper(strings.TrimSpace(locationFilter))
	count := 0
	for _, m := range machines {
		if m.RuntimeSec <= 0 && m.ProcSec <= 0 {
			continue
		}
		if locationFilter != "" && locationFilter != "ALL" {
			loc := strings.ToUpper(strings.TrimSpace(m.Location))
			if loc == "" || !strings.Contains(loc, locationFilter) {
				continue
			}
		}
		count++
	}
	return count
}
