package handlers

import (
	"net/http"
	"time"
)

// MachineOperatorAutoLogoutOffline godoc
// @Summary Auto logout operator jika mesin offline
// @Description Mengecek mesin offline dan menutup session ACTIVE jika offline >= 60 menit
// @Tags Machine Operator
// @Produce json
// @Success 200 {object} models.MachineOfflineAutoLogoutResponse
// @Failure 405 {string} string
// @Failure 500 {string} string
// @Router /machine-operator/auto-logout/offline [post]
func (h *Handler) MachineOperatorAutoLogoutOffline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 2*time.Minute)
	defer cancel()

	resp, err := h.Repo.RunMachineOfflineAutoLogout(ctx)
	if err != nil {
		http.Error(w, "Gagal proses auto logout offline: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp)
}
