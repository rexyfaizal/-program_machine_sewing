package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"backend_machine/models"
)

func (h *Handler) OperatorCtMaster(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	if err := h.Repo.EnsureOperatorCtSchema(ctx); err != nil {
		http.Error(w, "Gagal siapkan schema CT: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := h.Repo.GetOperatorCtMaster(ctx)
	if err != nil {
		http.Error(w, "Gagal ambil data CT master: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, data)
}

func (h *Handler) OperatorCtImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var input models.OperatorCtImportRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := contextWithTimeout(r, 60*time.Second)
	defer cancel()

	if err := h.Repo.EnsureOperatorCtSchema(ctx); err != nil {
		http.Error(w, "Gagal siapkan schema CT: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := h.Repo.ImportOperatorCt(ctx, input)
	if err != nil {
		http.Error(w, "Gagal import CT: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, result)
}
