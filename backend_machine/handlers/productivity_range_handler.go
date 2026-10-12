package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"backend_machine/models"
)

const maxProductivityRangeDays = 31

func parseProductivityRange(startText, endText string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02", strings.TrimSpace(startText))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("start_date wajib diisi dengan format YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", strings.TrimSpace(endText))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("end_date wajib diisi dengan format YYYY-MM-DD")
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("end_date harus sama atau setelah start_date")
	}
	if end.Sub(start) >= maxProductivityRangeDays*24*time.Hour {
		return time.Time{}, time.Time{}, fmt.Errorf("rentang tanggal maksimal %d hari (inklusif)", maxProductivityRangeDays)
	}
	return start, end, nil
}

// ProductivityRange returns daily productivity for every calendar day, inclusive.
func (h *Handler) ProductivityRange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method harus GET", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	start, end, err := parseProductivityRange(query.Get("start_date"), query.Get("end_date"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	shift := strings.TrimSpace(query.Get("shift"))
	if shift == "" {
		shift = "ALL"
	}
	ctx, cancel := contextWithTimeout(r, 5*time.Minute)
	defer cancel()
	resp := models.ProductivityRangeResponse{
		StartDate: start.Format("2006-01-02"),
		EndDate:   end.Format("2006-01-02"),
		Days:      make([]models.APIResponse, 0),
	}
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		date := day.Format("2006-01-02")
		daily, err := h.productivityForDate(ctx, date, shift)
		if err != nil {
			http.Error(w, fmt.Sprintf("Gagal mengambil produktivitas tanggal %s: %v", date, err), http.StatusInternalServerError)
			return
		}
		resp.Days = append(resp.Days, daily)
	}
	resp.TotalDays = len(resp.Days)
	writeJSON(w, resp)
}
