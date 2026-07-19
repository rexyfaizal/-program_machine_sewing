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

// MachineOperatorLogin godoc
// @Summary Login operator ke mesin
// @Description Login operator berdasarkan UUID mesin dan NIK operator
// @Tags Machine Operator
// @Accept json
// @Produce json
// @Param body body models.MachineOperatorLoginRequest true "Body login operator"
// @Success 200 {object} models.MachineOperatorLoginResponse
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /machine-operator/login [post]
func (h *Handler) MachineOperatorLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	var input models.MachineOperatorLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	input.UUID = strings.TrimSpace(input.UUID)
	input.OperatorNIK = strings.TrimSpace(input.OperatorNIK)

	if input.UUID == "" {
		http.Error(w, "uuid wajib diisi", http.StatusBadRequest)
		return
	}

	if input.OperatorNIK == "" {
		http.Error(w, "operatorNik wajib diisi", http.StatusBadRequest)
		return
	}

	resp, err := h.Repo.LoginMachineOperator(ctx, input)
	if err != nil {
		http.Error(w, "Gagal login operator: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp)
}

// MachineOperatorNote godoc
// @Summary Simpan catatan aktivitas operator
// @Description Menyimpan aktivitas/keterangan operator pada session aktif
// @Tags Machine Operator
// @Accept json
// @Produce json
// @Param body body models.MachineOperatorNoteRequest true "Body note operator"
// @Success 200 {object} models.MachineOperatorNoteResponse
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /machine-operator/note [post]
func (h *Handler) MachineOperatorNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	var input models.MachineOperatorNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	input.UUID = strings.TrimSpace(input.UUID)
	input.ReasonCode = strings.TrimSpace(input.ReasonCode)

	if input.UUID == "" {
		http.Error(w, "uuid wajib diisi", http.StatusBadRequest)
		return
	}

	if input.ReasonCode == "" {
		http.Error(w, "reasonCode wajib diisi", http.StatusBadRequest)
		return
	}

	resp, err := h.Repo.CreateMachineOperatorNote(ctx, input)
	if err != nil {
		if errors.Is(err, repository.ErrMachineOperatorNotFound) || errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Tidak ada session operator aktif untuk mesin ini", http.StatusNotFound)
			return
		}

		http.Error(w, "Gagal simpan note operator: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp)
}

// MachineOperatorActive godoc
// @Summary Cek operator aktif di mesin
// @Description Mengambil session operator aktif berdasarkan UUID mesin. Jika tidak ada, tetap return JSON active=false.
// @Tags Machine Operator
// @Produce json
// @Param uuid query string true "UUID mesin"
// @Success 200 {object} models.MachineOperatorActiveResponse
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /machine-operator/active [get]
func (h *Handler) MachineOperatorActive(w http.ResponseWriter, r *http.Request) {
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

	session, err := h.Repo.GetActiveMachineOperator(ctx, uuid)
	if err != nil {
		if errors.Is(err, repository.ErrMachineOperatorNotFound) || errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, models.MachineOperatorActiveResponse{
				Active:  false,
				Message: "Tidak ada operator aktif",
				Session: nil,
			})
			return
		}

		http.Error(w, "Gagal ambil operator aktif: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, models.MachineOperatorActiveResponse{
		Active:  true,
		Message: "Operator aktif",
		Session: session,
	})
}

// MachineOperatorReport godoc
// @Summary Report operator mesin harian
// @Description Mengambil session operator dan note berdasarkan tanggal
// @Tags Machine Operator
// @Produce json
// @Param date query string true "Tanggal format YYYY-MM-DD"
// @Success 200 {array} models.MachineOperatorReportItem
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /machine-operator/report [get]
func (h *Handler) MachineOperatorReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 60*time.Second)
	defer cancel()

	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	data, err := h.Repo.GetMachineOperatorReport(ctx, date)
	if err != nil {
		http.Error(w, "Gagal ambil report operator: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, data)
}
