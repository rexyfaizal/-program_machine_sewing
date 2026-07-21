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

const (
	BagianOperator = "Operator"
	BagianSPV      = "SPV"
	BagianMekanik  = "Mekanik"
)

// MessageSender mendefinisikan fungsi Telegram client
// yang digunakan oleh handler.
type MessageSender interface {
	SendMessage(
		ctx context.Context,
		chatID int64,
		message string,
	) error

	SendMessageWithKeyboard(
		ctx context.Context,
		chatID int64,
		message string,
		buttonRows [][]string,
	) error

	SendMessageRemoveKeyboard(
		ctx context.Context,
		chatID int64,
		message string,
	) error
}

// TelegramHandler menangani pesan masuk dari Telegram.
type TelegramHandler struct {
	registrationService service.RegistrationService
	sender              MessageSender

	mu sync.RWMutex

	// User sedang menunggu memilih bagian.
	awaitingBagian map[int64]bool

	// User sedang menunggu memasukkan NIK.
	awaitingNIK map[int64]bool

	// Menyimpan pilihan bagian berdasarkan Telegram chat ID.
	selectedBagian map[int64]string
}

// NewTelegramHandler membuat Telegram handler baru.
func NewTelegramHandler(
	registrationService service.RegistrationService,
	sender MessageSender,
) *TelegramHandler {
	return &TelegramHandler{
		registrationService: registrationService,
		sender:              sender,

		awaitingBagian: make(map[int64]bool),
		awaitingNIK:    make(map[int64]bool),
		selectedBagian: make(map[int64]string),
	}
}

// Handle memproses setiap update atau pesan dari Telegram.
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

	// Registrasi hanya diperbolehkan melalui private chat.
	if message.Chat.Type != "private" {
		h.reply(
			ctx,
			chatID,
			"Silakan lakukan pendaftaran melalui chat pribadi dengan bot.",
		)
		return
	}

	// Mendukung:
	// /start
	// start
	if isStartMessage(text) {
		h.startRegistration(ctx, chatID)
		return
	}

	// Mendukung:
	// /cancel
	// cancel
	if isCancelMessage(text) {
		h.clearRegistrationState(chatID)

		h.replyRemoveKeyboard(
			ctx,
			chatID,
			"Pendaftaran dibatalkan.\n\n"+
				"Ketik /start atau start untuk memulai kembali.",
		)
		return
	}

	// Memeriksa apakah pesan merupakan pilihan bagian.
	if bagian, valid := normalizeBagian(text); valid {
		if !h.isAwaitingBagian(chatID) {
			h.reply(
				ctx,
				chatID,
				"Ketik /start atau start terlebih dahulu untuk memulai pendaftaran.",
			)
			return
		}

		h.selectBagian(chatID, bagian)

		h.replyRemoveKeyboard(
			ctx,
			chatID,
			"Bagian dipilih: "+bagian+"\n\n"+
				"Sekarang silakan kirim NIK Anda.\n\n"+
				"Contoh:\n"+
				"100009\n\n"+
				"Ketik /cancel untuk membatalkan.",
		)
		return
	}

	// Periksa command yang tidak dikenal.
	command, _ := parseCommand(text)

	if command != "" {
		h.reply(
			ctx,
			chatID,
			"Perintah tidak dikenali.\n\n"+
				"Ketik /start atau start untuk melakukan pendaftaran.",
		)
		return
	}

	// User sudah start, tetapi belum memilih bagian.
	if h.isAwaitingBagian(chatID) {
		h.replyWithBagianKeyboard(
			ctx,
			chatID,
			"Silakan pilih bagian Anda terlebih dahulu:",
		)
		return
	}

	// User belum menjalankan start atau tidak sedang menunggu NIK.
	if !h.isAwaitingNIK(chatID) {
		h.reply(
			ctx,
			chatID,
			"Ketik /start atau start terlebih dahulu untuk melakukan pendaftaran.",
		)
		return
	}

	// Pesan dianggap sebagai NIK.
	h.handleNIK(
		ctx,
		chatID,
		text,
	)
}

// startRegistration memulai proses registrasi.
func (h *TelegramHandler) startRegistration(
	ctx context.Context,
	chatID int64,
) {
	h.mu.Lock()

	h.awaitingBagian[chatID] = true

	delete(
		h.awaitingNIK,
		chatID,
	)

	delete(
		h.selectedBagian,
		chatID,
	)

	h.mu.Unlock()

	h.replyWithBagianKeyboard(
		ctx,
		chatID,
		"Selamat datang di Bot Registrasi Karyawan.\n\n"+
			"Silakan pilih bagian Anda:",
	)
}

// handleNIK memproses NIK yang dikirim user.
func (h *TelegramHandler) handleNIK(
	ctx context.Context,
	chatID int64,
	nik string,
) {
	bagian := h.getSelectedBagian(chatID)

	// Bagian wajib dipilih sebelum NIK diproses.
	if strings.TrimSpace(bagian) == "" {
		h.startRegistration(ctx, chatID)
		return
	}

	requestCtx, cancel := context.WithTimeout(
		ctx,
		15*time.Second,
	)
	defer cancel()

	// Mengirimkan:
	// - NIK
	// - Bagian
	// - Telegram ID
	message, complete, err :=
		h.registrationService.Register(
			requestCtx,
			nik,
			bagian,
			chatID,
		)

	if err != nil {
		log.Printf(
			"Gagal registrasi NIK=%s Bagian=%s TelegramID=%d Error=%v",
			nik,
			bagian,
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

	// Jika registrasi selesai, hapus seluruh state.
	if complete {
		h.clearRegistrationState(chatID)
	} else {
		// Jika gagal karena NIK salah atau tidak ditemukan,
		// tetap tunggu input NIK berikutnya.
		h.setAwaitingNIK(chatID, true)
	}

	h.reply(
		ctx,
		chatID,
		message,
	)
}

// replyWithBagianKeyboard mengirim pilihan bagian.
func (h *TelegramHandler) replyWithBagianKeyboard(
	ctx context.Context,
	chatID int64,
	message string,
) {
	requestCtx, cancel := context.WithTimeout(
		ctx,
		15*time.Second,
	)
	defer cancel()

	buttonRows := [][]string{
		{
			BagianOperator,
			BagianSPV,
			BagianMekanik,
		},
	}

	if err := h.sender.SendMessageWithKeyboard(
		requestCtx,
		chatID,
		message,
		buttonRows,
	); err != nil {
		log.Printf(
			"Gagal mengirim pilihan bagian ke chat %d: %v",
			chatID,
			err,
		)
	}
}

// replyRemoveKeyboard mengirim pesan sekaligus menghapus keyboard.
func (h *TelegramHandler) replyRemoveKeyboard(
	ctx context.Context,
	chatID int64,
	message string,
) {
	requestCtx, cancel := context.WithTimeout(
		ctx,
		15*time.Second,
	)
	defer cancel()

	if err := h.sender.SendMessageRemoveKeyboard(
		requestCtx,
		chatID,
		message,
	); err != nil {
		log.Printf(
			"Gagal mengirim pesan dan menghapus keyboard ke chat %d: %v",
			chatID,
			err,
		)
	}
}

// reply mengirim pesan biasa.
func (h *TelegramHandler) reply(
	ctx context.Context,
	chatID int64,
	message string,
) {
	requestCtx, cancel := context.WithTimeout(
		ctx,
		15*time.Second,
	)
	defer cancel()

	if err := h.sender.SendMessage(
		requestCtx,
		chatID,
		message,
	); err != nil {
		log.Printf(
			"Gagal mengirim pesan ke chat %d: %v",
			chatID,
			err,
		)
	}
}

// selectBagian menyimpan bagian yang dipilih user.
func (h *TelegramHandler) selectBagian(
	chatID int64,
	bagian string,
) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(
		h.awaitingBagian,
		chatID,
	)

	h.awaitingNIK[chatID] = true
	h.selectedBagian[chatID] = bagian
}

// getSelectedBagian mengambil bagian yang sudah dipilih user.
func (h *TelegramHandler) getSelectedBagian(
	chatID int64,
) string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.selectedBagian[chatID]
}

// setAwaitingNIK mengatur status menunggu NIK.
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

	delete(
		h.awaitingNIK,
		chatID,
	)
}

// isAwaitingBagian memeriksa apakah user sedang memilih bagian.
func (h *TelegramHandler) isAwaitingBagian(
	chatID int64,
) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.awaitingBagian[chatID]
}

// isAwaitingNIK memeriksa apakah user sedang menunggu input NIK.
func (h *TelegramHandler) isAwaitingNIK(
	chatID int64,
) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.awaitingNIK[chatID]
}

// clearRegistrationState menghapus state registrasi user.
func (h *TelegramHandler) clearRegistrationState(
	chatID int64,
) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(
		h.awaitingBagian,
		chatID,
	)

	delete(
		h.awaitingNIK,
		chatID,
	)

	delete(
		h.selectedBagian,
		chatID,
	)
}

// normalizeBagian memastikan bagian sesuai pilihan yang diperbolehkan.
func normalizeBagian(
	text string,
) (string, bool) {
	switch strings.ToLower(
		strings.TrimSpace(text),
	) {
	case "operator":
		return BagianOperator, true

	case "spv":
		return BagianSPV, true

	case "mekanik":
		return BagianMekanik, true

	default:
		return "", false
	}
}

// isStartMessage memeriksa perintah start.
func isStartMessage(
	text string,
) bool {
	if strings.EqualFold(
		strings.TrimSpace(text),
		"start",
	) {
		return true
	}

	command, _ := parseCommand(text)

	return command == "start"
}

// isCancelMessage memeriksa perintah cancel.
func isCancelMessage(
	text string,
) bool {
	if strings.EqualFold(
		strings.TrimSpace(text),
		"cancel",
	) {
		return true
	}

	command, _ := parseCommand(text)

	return command == "cancel"
}

// parseCommand memisahkan command dan argument Telegram.
//
// Contoh:
// /start              -> command=start
// /start@nama_bot     -> command=start
// /start 100009       -> command=start, argument=100009
func parseCommand(
	text string,
) (
	command string,
	argument string,
) {
	fields := strings.Fields(
		strings.TrimSpace(text),
	)

	if len(fields) == 0 ||
		!strings.HasPrefix(fields[0], "/") {
		return "", ""
	}

	command = strings.TrimPrefix(
		fields[0],
		"/",
	)

	// Mendukung command dari group seperti:
	// /start@nama_bot
	if atIndex := strings.Index(
		command,
		"@",
	); atIndex >= 0 {
		command = command[:atIndex]
	}

	command = strings.ToLower(
		strings.TrimSpace(command),
	)

	if len(fields) > 1 {
		argument = strings.Join(
			fields[1:],
			" ",
		)
	}

	return command, strings.TrimSpace(argument)
}
