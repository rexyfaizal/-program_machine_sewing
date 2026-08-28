package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"backend_machine/models"
)

func (h *Handler) OperatorOutputTarget(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	workDate := r.URL.Query().Get("date")
	if workDate == "" {
		http.Error(w, "Parameter date wajib diisi (YYYY-MM-DD).", http.StatusBadRequest)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	if err := h.Repo.EnsureOperatorOutputTargetSchema(ctx); err != nil {
		http.Error(w, "Gagal siapkan schema output target: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := h.Repo.GetOperatorOutputTargetByDate(ctx, workDate)
	if err != nil {
		http.Error(w, "Gagal ambil data output target: "+err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, data)
}

func (h *Handler) OperatorOutputTargetImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var input models.OperatorOutputTargetImportRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := contextWithTimeout(r, 60*time.Second)
	defer cancel()

	if err := h.Repo.EnsureOperatorOutputTargetSchema(ctx); err != nil {
		http.Error(w, "Gagal siapkan schema output target: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := h.Repo.ImportOperatorOutputTarget(ctx, input)
	if err != nil {
		http.Error(w, "Gagal import output target: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, result)
}
