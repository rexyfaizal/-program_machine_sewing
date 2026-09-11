package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"backend_machine/models"
)

func (h *Handler) OperatorStyleCorrectionImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var input models.OperatorStyleCorrectionImportRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := contextWithTimeout(r, 60*time.Second)
	defer cancel()

	result, err := h.Repo.ImportOperatorStyleCorrection(ctx, input)
	if err != nil {
		http.Error(w, "Gagal koreksi style sesi: "+err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, result)
}
