package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"backend_machine/models"
)

// SpvLineAssignments godoc
// @Summary List / set NIK SPV per location (line)
// @Tags SpvLineAssignments
// @Produce json
// @Param factory query string false "Filter factory/area (GM1, GM2, ...)"
// @Success 200 {object} map[string]any
// @Router /spv-line-assignments [get]
func (h *Handler) SpvLineAssignments(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	switch r.Method {
	case http.MethodGet:
		factory := strings.TrimSpace(r.URL.Query().Get("factory"))
		list, err := h.Repo.ListSpvLineAssignments(ctx, factory)
		if err != nil {
			http.Error(w, "Gagal ambil SPV line: "+err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, map[string]any{
			"factory":     strings.ToUpper(factory),
			"assignments": list,
		})

	case http.MethodPut, http.MethodPost:
		var input models.SpvLineAssignmentPutRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
			return
		}

		item, err := h.Repo.UpsertSpvLineAssignment(ctx, input.Location, input.SpvNik)
		if err != nil {
			http.Error(w, "Gagal simpan SPV line: "+err.Error(), http.StatusBadRequest)
			return
		}

		writeJSON(w, item)

	default:
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
	}
}

// SpvLineAssignmentRename godoc
// @Summary Rename location key assignment SPV saat line di-rename
// @Tags SpvLineAssignments
// @Router /spv-line-assignments/rename [put]
func (h *Handler) SpvLineAssignmentRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	var input models.SpvLineAssignmentRenameRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Repo.RenameSpvLineAssignmentLocation(ctx, input.OldLocation, input.NewLocation); err != nil {
		http.Error(w, "Gagal rename SPV line: "+err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]any{
		"ok":          true,
		"oldLocation": strings.TrimSpace(input.OldLocation),
		"newLocation": strings.TrimSpace(input.NewLocation),
	})
}
