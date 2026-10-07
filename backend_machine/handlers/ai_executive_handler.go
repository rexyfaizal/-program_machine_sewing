package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/utils"
)

// AiExecutive godoc
// @Summary Paket AI Executive (skor, briefing, vs kemarin)
// @Description Ringkas skor utilitas mesin + briefing atasan + banding vs kemarin.
// @Tags AI Executive
// @Accept json
// @Produce json
// @Param body body models.AiExecutiveRequest true "Snapshot hari ini + kemarin + insight"
// @Success 200 {object} models.AiExecutiveResponse
// @Router /ai/executive [post]
func (h *Handler) AiExecutive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	_, cancel := contextWithTimeout(r, 20*time.Second)
	defer cancel()

	var input models.AiExecutiveRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(input.Date) == "" {
		input.Date = time.Now().Format("2006-01-02")
	}

	resp := utils.BuildAiExecutiveResponse(input)
	writeJSON(w, resp)
}
