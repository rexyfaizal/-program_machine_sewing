package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/repository"
)

// MachineOperatorLossEventStart godoc
// @Summary Mulai loss event operator
// @Description Membuat event loss/aktivitas non-produktif untuk operator yang sedang aktif di mesin
// @Tags Machine Operator
// @Accept json
// @Produce json
// @Param body body models.MachineOperatorLossEventStartRequest true "Body start loss event"
// @Success 200 {object} models.MachineOperatorLossEventStartResponse
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 405 {string} string
// @Failure 500 {string} string
// @Router /machine-operator/loss-event/start [post]
func (h *Handler) MachineOperatorLossEventStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	var input models.MachineOperatorLossEventStartRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	input.UUID = strings.TrimSpace(input.UUID)
	input.ReasonCode = strings.TrimSpace(input.ReasonCode)
	input.ReasonLabel = strings.TrimSpace(input.ReasonLabel)
	input.Note = strings.TrimSpace(input.Note)

	if input.UUID == "" {
		http.Error(w, "uuid wajib diisi", http.StatusBadRequest)
		return
	}

	if input.ReasonCode == "" {
		http.Error(w, "reasonCode wajib diisi", http.StatusBadRequest)
		return
	}

	resp, err := h.Repo.StartMachineOperatorLossEvent(ctx, input)
	if err != nil {
		if errors.Is(err, repository.ErrMachineOperatorNotFound) || errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Tidak ada operator aktif untuk mesin ini", http.StatusNotFound)
			return
		}

		http.Error(w, "Gagal mulai loss event: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp)
}

// MachineOperatorLossEventFinish godoc
// @Summary Selesaikan loss event operator
// @Description Menutup event loss/aktivitas non-produktif yang masih aktif pada operator mesin
// @Tags Machine Operator
// @Accept json
// @Produce json
// @Param body body models.MachineOperatorLossEventFinishRequest true "Body finish loss event"
// @Success 200 {object} models.MachineOperatorLossEventFinishResponse
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 405 {string} string
// @Failure 500 {string} string
// @Router /machine-operator/loss-event/finish [post]
func (h *Handler) MachineOperatorLossEventFinish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	var input models.MachineOperatorLossEventFinishRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	input.UUID = strings.TrimSpace(input.UUID)

	if input.UUID == "" {
		http.Error(w, "uuid wajib diisi", http.StatusBadRequest)
		return
	}

	resp, err := h.Repo.FinishMachineOperatorLossEvent(ctx, input)
	if err != nil {
		if errors.Is(err, repository.ErrMachineOperatorNotFound) || errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Tidak ada operator aktif untuk mesin ini", http.StatusNotFound)
			return
		}

		if errors.Is(err, repository.ErrMachineOperatorLossEventNotFound) {
			http.Error(w, "Tidak ada loss event aktif untuk mesin ini", http.StatusNotFound)
			return
		}

		http.Error(w, "Gagal selesaikan loss event: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp)
}

// MachineOperatorLossEventActive godoc
// @Summary Cek loss event aktif
// @Description Mengecek apakah mesin sedang memiliki loss event aktif
// @Tags Machine Operator
// @Produce json
// @Param uuid query string true "UUID mesin"
// @Success 200 {object} models.MachineOperatorLossEventActiveResponse
// @Failure 400 {string} string
// @Failure 405 {string} string
// @Failure 500 {string} string
// @Router /machine-operator/loss-event/active [get]
func (h *Handler) MachineOperatorLossEventActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
	if uuid == "" {
		http.Error(w, "uuid wajib diisi", http.StatusBadRequest)
		return
	}

	resp, err := h.Repo.GetActiveMachineOperatorLossEvent(ctx, uuid)
	if err != nil {
		http.Error(w, "Gagal ambil loss event aktif: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp)
}
