package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"backend_machine/models"
)

// MachineIDImport godoc
// @Summary Import Machine ID dari Excel (UUID + Machine ID)
// @Tags MachineSettings
// @Accept json
// @Produce json
// @Param body body models.MachineIDImportRequest true "Baris import"
// @Success 200 {object} models.MachineIDImportResponse
// @Router /machine-settings/import-ids [post]
func (h *Handler) MachineIDImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var input models.MachineIDImportRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(input.Rows) == 0 {
		http.Error(w, "rows kosong", http.StatusBadRequest)
		return
	}

	ctx, cancel := contextWithTimeout(r, 60*time.Second)
	defer cancel()

	result, err := h.Repo.ImportMachineIDs(ctx, input)
	if err != nil {
		http.Error(w, "Gagal import Machine ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, result)
}
