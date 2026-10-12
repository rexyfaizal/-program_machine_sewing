package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/utils"
)

// AiChat godoc
// @Summary Copilot chat (Ollama lokal / OpenAI)
// @Description Tanya jawab dashboard berbasis konteks mesin + AI Insight (tanpa output/tiket).
// @Tags AI Copilot
// @Accept json
// @Produce json
// @Param body body models.AiChatRequest true "Pertanyaan + konteks"
// @Success 200 {object} models.AiChatResponse
// @Router /ai/chat [post]
func (h *Handler) AiChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	if !utils.AiChatConfigured() {
		http.Error(
			w,
			"AI Copilot belum siap. Set AI_PROVIDER=ollama (lokal) atau openai + OPENAI_API_KEY di .env",
			http.StatusServiceUnavailable,
		)
		return
	}

	timeout := 130 * time.Second
	if utils.AiChatProvider() == "openai" {
		timeout = 70 * time.Second
	}
	ctx, cancel := contextWithTimeout(r, timeout)
	defer cancel()

	var input models.AiChatRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	input.Question = strings.TrimSpace(input.Question)
	if input.Question == "" {
		http.Error(w, "question wajib diisi", http.StatusBadRequest)
		return
	}
	if input.Page == "" {
		input.Page = "dashboard"
	}
	if input.Date == "" {
		input.Date = time.Now().Format("2006-01-02")
	}
	// Validate dates before querying; database failures must not reuse today's snapshot.
	if _, _, err := utils.ExtractAiQuestionDate(input.Question, time.Now()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dataCtx, dataCancel := contextWithTimeout(r, 60*time.Second)
	err := prepareAiChatDateContext(dataCtx, &input, time.Now(), h.productivityForDate)
	dataCancel()
	if err != nil {
		http.Error(w, "Gagal mengambil data tanggal yang diminta: "+err.Error(), http.StatusBadGateway)
		return
	}

	resp, err := utils.CallOpenAIChat(ctx, input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	writeJSON(w, resp)
}
