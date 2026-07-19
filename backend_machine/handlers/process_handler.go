package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"backend_machine/models"
)

func (h *Handler) ProcessDetail(w http.ResponseWriter, r *http.Request) {
	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
	if uuid == "" {
		http.Error(w, "Parameter uuid wajib diisi", http.StatusBadRequest)
		return
	}

	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		http.Error(w, "Format tanggal harus YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	start := day
	end := day.AddDate(0, 0, 1)

	ctx, cancel := contextWithTimeout(r, 60*time.Second)
	defer cancel()

	machine, err := h.Repo.FindMachineByUUID(ctx, uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	machineRow, err := h.Repo.GetMachineProductivity(ctx, machine, date)
	if err != nil {
		http.Error(w, "Gagal ambil summary mesin: "+err.Error(), http.StatusInternalServerError)
		return
	}

	events, groups, hours, err := h.Repo.GetProcessEvents(ctx, machine.TableName, start, end)
	if err != nil {
		http.Error(w, "Gagal ambil detail proses: "+err.Error(), http.StatusInternalServerError)
		return
	}

	alarms, err := h.Repo.GetAlarmGroups(ctx, machine.UUID, start, end)
	if err != nil {
		log.Printf("alarm group skip uuid=%s: %v", machine.UUID, err)
		alarms = []models.AlarmGroup{}
	}

	resp := models.ProcessDetailResponse{
		Date:    date,
		Machine: machineRow,
		Groups:  groups,
		Hours:   hours,
		Alarms:  alarms,
		Events:  events,
	}

	writeJSON(w, resp)
}
