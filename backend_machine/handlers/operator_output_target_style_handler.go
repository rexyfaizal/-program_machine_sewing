package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"backend_machine/models"
)

func (h *Handler) OperatorOutputTargetStyle(w http.ResponseWriter, r *http.Request) {
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

	if err := h.Repo.EnsureOperatorOutputTargetStyleSchema(ctx); err != nil {
		http.Error(w, "Gagal siapkan schema output target Style: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := h.Repo.GetOperatorOutputTargetStyleByDate(ctx, workDate)
	if err != nil {
		http.Error(w, "Gagal ambil data output target Style: "+err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, data)
}

func (h *Handler) OperatorOutputTargetStyleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var input models.OperatorOutputTargetStyleImportRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := contextWithTimeout(r, 60*time.Second)
	defer cancel()

	if err := h.Repo.EnsureOperatorOutputTargetStyleSchema(ctx); err != nil {
		http.Error(w, "Gagal siapkan schema output target Style: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := h.Repo.ImportOperatorOutputTargetStyle(ctx, input)
	if err != nil {
		http.Error(w, "Gagal import output target Style: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, result)
}
