package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/repository"
)

// MechanicIdentify godoc
// @Summary Verifikasi NIK / RFID mekanik
// @Tags Mechanic
// @Accept json
// @Produce json
// @Param body body models.MechanicIdentifyRequest true "NIK atau RFID"
// @Success 200 {object} models.MechanicIdentifyResponse
// @Router /mechanic/identify [post]
func (h *Handler) MechanicIdentify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 15*time.Second)
	defer cancel()

	var input models.MechanicIdentifyRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	code := strings.TrimSpace(input.Code)
	if code == "" {
		code = strings.TrimSpace(input.NIK)
	}
	if code == "" {
		code = strings.TrimSpace(input.RFID)
	}
	if code == "" {
		http.Error(w, "nik/rfid wajib diisi", http.StatusBadRequest)
		return
	}

	resp, err := h.Repo.IdentifyMechanic(ctx, code)
	if err != nil {
		http.Error(w, "Gagal verifikasi mekanik: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp)
}

// MechanicRFIDRegister godoc
// @Summary Daftarkan kartu RFID mekanik
// @Tags Mechanic
// @Accept json
// @Produce json
// @Param body body models.MechanicRFIDRegisterRequest true "NIK + RFID"
// @Success 200 {object} models.MechanicRFIDRegisterResponse
// @Router /mechanic/rfid/register [post]
func (h *Handler) MechanicRFIDRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 20*time.Second)
	defer cancel()

	var input models.MechanicRFIDRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.Repo.RegisterMechanicRFID(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrMechanicNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, repository.ErrMechanicNotAuthorized):
			http.Error(w, err.Error(), http.StatusForbidden)
		case errors.Is(err, repository.ErrMechanicRFIDTaken):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "Gagal daftar RFID: "+err.Error(), http.StatusBadRequest)
		}
		return
	}

	writeJSON(w, resp)
}

// MechanicBrokenList godoc
// @Summary Daftar mesin rusak aktif
// @Tags Mechanic
// @Produce json
// @Param status query string false "OPEN | IN_PROGRESS | ALL | DONE"
// @Param location query string false "Filter location"
// @Success 200 {object} models.MechanicBrokenListResponse
// @Router /mechanic/broken-machines [get]
func (h *Handler) MechanicBrokenList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	status := strings.TrimSpace(r.URL.Query().Get("status"))
	location := strings.TrimSpace(r.URL.Query().Get("location"))

	resp, err := h.Repo.ListMechanicBrokenMachines(ctx, status, location)
	if err != nil {
		http.Error(w, "Gagal ambil daftar mesin rusak: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp)
}

// MechanicClaim godoc
// @Summary Ambil tiket mesin rusak
// @Tags Mechanic
// @Accept json
// @Produce json
// @Param body body models.MechanicClaimRequest true "Claim request"
// @Success 200 {object} models.MechanicActionResponse
// @Router /mechanic/claim [post]
func (h *Handler) MechanicClaim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	var input models.MechanicClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.Repo.ClaimMechanicBrokenMachine(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrMechanicNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, repository.ErrMechanicNotAuthorized):
			http.Error(w, err.Error(), http.StatusForbidden)
		case errors.Is(err, repository.ErrMechanicTicketNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, repository.ErrMechanicAlreadyClaimed):
			http.Error(w, err.Error(), http.StatusConflict)
		case errors.Is(err, repository.ErrMechanicAlreadyDone):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "Gagal ambil tiket: "+err.Error(), http.StatusBadRequest)
		}
		return
	}

	writeJSON(w, resp)
}

// MechanicDone godoc
// @Summary Selesaikan tiket mesin rusak
// @Tags Mechanic
// @Accept json
// @Produce json
// @Param body body models.MechanicDoneRequest true "Done request"
// @Success 200 {object} models.MechanicActionResponse
// @Router /mechanic/done [post]
func (h *Handler) MechanicDone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	var input models.MechanicDoneRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.Repo.DoneMechanicBrokenMachine(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrMechanicNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, repository.ErrMechanicNotAuthorized):
			http.Error(w, err.Error(), http.StatusForbidden)
		case errors.Is(err, repository.ErrMechanicTicketNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, repository.ErrMechanicAlreadyClaimed):
			http.Error(w, err.Error(), http.StatusConflict)
		case errors.Is(err, repository.ErrMechanicClaimRequired):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, repository.ErrMechanicAlreadyDone):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "Gagal selesaikan tiket: "+err.Error(), http.StatusBadRequest)
		}
		return
	}

	writeJSON(w, resp)
}
