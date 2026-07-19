package handlers

import (
	"net/http"
	"time"
)

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := contextWithTimeout(r, 5*time.Second)
	defer cancel()

	if err := h.Repo.Ping(ctx); err != nil {
		http.Error(w, "SQL Server tidak terkoneksi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{
		"status": "ok",
	})
}
