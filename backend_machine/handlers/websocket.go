package handlers

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var productivityUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")

		log.Println("WebSocket Origin:", origin)

		allowedOrigins := map[string]bool{
			"http://localhost:5173": true,
			"http://127.0.0.1:5173": true,
			"http://localhost:5174": true,
			"http://127.0.0.1:5174": true,

			"http://10.5.0.8:5173": true,
			"http://10.5.0.8:5174": true,
			"http://10.5.0.8:5000": true,

			"http://localhost:5000": true,
			"http://127.0.0.1:5000": true,
		}

		if origin == "" {
			return true
		}

		return allowedOrigins[origin]
	},
}

func (h *Handler) ProductivityWS(w http.ResponseWriter, r *http.Request) {
	conn, err := productivityUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("websocket upgrade error:", err)
		return
	}
	defer conn.Close()

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	shift := strings.TrimSpace(r.URL.Query().Get("shift"))

	sendSnapshot := func() bool {
		apiURL := "/api/productivity?date=" + url.QueryEscape(date)
		if shift != "" {
			apiURL += "&shift=" + url.QueryEscape(shift)
		}

		req := httptest.NewRequest(http.MethodGet, apiURL, nil)
		rec := httptest.NewRecorder()

		// Pakai handler API yang sudah ada dan sudah terbukti jalan
		h.Productivity(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			writeErr := conn.WriteJSON(map[string]any{
				"error": err.Error(),
			})
			return writeErr == nil
		}

		if resp.StatusCode >= 400 {
			writeErr := conn.WriteJSON(map[string]any{
				"error":  string(body),
				"status": resp.StatusCode,
			})
			return writeErr == nil
		}

		err = conn.WriteMessage(websocket.TextMessage, body)
		if err != nil {
			log.Println("websocket write error:", err)
			return false
		}

		return true
	}

	// Kirim data pertama langsung saat konek
	if !sendSnapshot() {
		return
	}

	// Update otomatis setiap 5 detik
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !sendSnapshot() {
			return
		}
	}
}
