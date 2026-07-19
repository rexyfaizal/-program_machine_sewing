package handler

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"telegram_notif/models"
	"telegram_notif/service"
)

type MessageSender interface {
	SendMessage(ctx context.Context, chatID int64, message string) error
}

type TelegramHandler struct {
	registrationService service.RegistrationService
	sender              MessageSender

	mu          sync.RWMutex
	awaitingNIK map[int64]bool
}

func NewTelegramHandler(
	registrationService service.RegistrationService,
	sender MessageSender,
) *TelegramHandler {
	return &TelegramHandler{
		registrationService: registrationService,
		sender:              sender,
		awaitingNIK:         make(map[int64]bool),
	}
}

func (h *TelegramHandler) Handle(
	ctx context.Context,
	update models.TelegramUpdate,
) {
	message := update.Message
	if message == nil {
		return
	}

	text := strings.TrimSpace(message.Text)
	if text == "" {
		return
	}

	chatID := message.Chat.ID

	if message.Chat.Type != "private" {
		h.reply(
			ctx,
			chatID,
			"Silakan lakukan pendaftaran melalui chat pribadi dengan bot.",
		)
		return
	}

	command, argument := parseCommand(text)

	switch command {
	case "start":
		if argument != "" {
			h.handleNIK(ctx, chatID, argument)
			return
		}

		h.setAwaitingNIK(chatID, true)
		h.reply(
			ctx,
			chatID,
			"Selamat datang di Bot Registrasi Karyawan.\n\n"+
				"Silakan kirim NIK Anda.\n\n"+
				"Contoh:\n100009\n\n"+
				"Ketik /cancel untuk membatalkan.",
		)
		return

	case "cancel":
		h.setAwaitingNIK(chatID, false)
		h.reply(
			ctx,
			chatID,
			"Pendaftaran dibatalkan. Ketik /start untuk memulai kembali.",
		)
		return

	case "":
		// Pesan biasa akan diperiksa sebagai NIK.

	default:
		h.reply(
			ctx,
			chatID,
			"Perintah tidak dikenali. Ketik /start untuk melakukan pendaftaran.",
		)
		return
	}

	if !h.isAwaitingNIK(chatID) {
		h.reply(
			ctx,
			chatID,
			"Ketik /start terlebih dahulu untuk melakukan pendaftaran.",
		)
		return
	}

	h.handleNIK(ctx, chatID, text)
}

func (h *TelegramHandler) handleNIK(
	ctx context.Context,
	chatID int64,
	nik string,
) {
	requestCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	message, complete, err := h.registrationService.Register(
		requestCtx,
		nik,
		chatID,
	)
	if err != nil {
		log.Printf(
			"Gagal registrasi NIK=%s TelegramID=%d Error=%v",
			nik,
			chatID,
			err,
		)

		h.setAwaitingNIK(chatID, true)
		h.reply(
			ctx,
			chatID,
			"Terjadi kesalahan saat menghubungi database.\n\n"+
				"Silakan coba kembali beberapa saat lagi.",
		)
		return
	}

	h.setAwaitingNIK(chatID, !complete)
	h.reply(ctx, chatID, message)
}

func (h *TelegramHandler) reply(
	ctx context.Context,
	chatID int64,
	message string,
) {
	requestCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := h.sender.SendMessage(requestCtx, chatID, message); err != nil {
		log.Printf(
			"Gagal mengirim pesan ke chat %d: %v",
			chatID,
			err,
		)
	}
}

func (h *TelegramHandler) setAwaitingNIK(
	chatID int64,
	awaiting bool,
) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if awaiting {
		h.awaitingNIK[chatID] = true
		return
	}

	delete(h.awaitingNIK, chatID)
}

func (h *TelegramHandler) isAwaitingNIK(chatID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.awaitingNIK[chatID]
}

func parseCommand(text string) (command string, argument string) {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 || !strings.HasPrefix(fields[0], "/") {
		return "", ""
	}

	command = strings.TrimPrefix(fields[0], "/")

	if atIndex := strings.Index(command, "@"); atIndex >= 0 {
		command = command[:atIndex]
	}

	command = strings.ToLower(command)

	if len(fields) > 1 {
		argument = strings.Join(fields[1:], " ")
	}

	return command, strings.TrimSpace(argument)
}
