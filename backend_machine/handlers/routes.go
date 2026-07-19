package handlers

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", h.Health)
	mux.HandleFunc("/api/productivity", h.Productivity)
	mux.HandleFunc("/api/process-detail", h.ProcessDetail)
}
