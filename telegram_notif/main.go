package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"telegram_notif/bot"
	"telegram_notif/config"
	"telegram_notif/database"
	"telegram_notif/handler"
	"telegram_notif/repository"
	"telegram_notif/service"
	"telegram_notif/telegram"
	"telegram_notif/worker"
)

func main() {
	// =====================================================
	// LOAD KONFIGURASI TELEGRAM
	// =====================================================
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf(
			"Konfigurasi tidak valid: %v",
			err,
		)
	}

	// =====================================================
	// KONEKSI DATABASE
	// =====================================================
	//
	// Database ConnectDB membaca:
	// DB_SERVER
	// DB_PORT
	// DB_USER
	// DB_PASSWORD
	// DB_NAME
	//
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf(
			"Gagal terhubung ke SQL Server: %v",
			err,
		)
	}
	defer db.Close()

	log.Println("Berhasil terhubung ke SQL Server.")

	// =====================================================
	// TELEGRAM CLIENT
	// =====================================================
	telegramClient := telegram.NewClient(
		cfg.BotToken,
	)

	// =====================================================
	// REGISTRASI TELEGRAM KARYAWAN
	// =====================================================
	employeeRepository, err :=
		repository.NewEmployeeRepository(
			db,
			cfg.EmployeeTable,
		)
	if err != nil {
		log.Fatalf(
			"Gagal membuat employee repository: %v",
			err,
		)
	}

	registrationService :=
		service.NewRegistrationService(
			employeeRepository,
		)

	telegramHandler :=
		handler.NewTelegramHandler(
			registrationService,
			telegramClient,
		)

	telegramPoller :=
		bot.NewPoller(
			telegramClient,
			telegramHandler,
		)

	// =====================================================
	// NOTIFIKASI MESIN RUSAK
	// =====================================================
	machineNotificationRepository, err :=
		repository.NewMachineNotificationRepository(
			db,
		)
	if err != nil {
		log.Fatalf(
			"Gagal membuat machine notification repository: %v",
			err,
		)
	}

	machineNotificationService :=
		service.NewMachineNotificationService(
			machineNotificationRepository,
			telegramClient,

			// Maksimal notifikasi yang diproses
			// dalam satu pemeriksaan.
			100,
		)

	machineNotificationWorker :=
		worker.NewMachineNotificationWorker(
			machineNotificationService,

			// Pemeriksaan database setiap 5 detik.
			5*time.Second,
		)

	// =====================================================
	// GRACEFUL SHUTDOWN
	// =====================================================
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	// Jalankan worker notifikasi secara bersamaan.
	go machineNotificationWorker.Start(ctx)

	log.Printf(
		"Bot aktif. Employee Table=%s",
		cfg.EmployeeTable,
	)

	log.Println(
		"Worker notifikasi MACHINE_BROKEN dan WAIT_HANCA aktif.",
	)

	log.Println(
		"MACHINE_BROKEN dikirim ke seluruh mekanik aktif.",
	)

	log.Println(
		"WAIT_HANCA dikirim kepada SPV berdasarkan branch dan line.",
	)

	log.Println(
		"Tekan Ctrl+C untuk menghentikan program.",
	)

	// Poller Telegram untuk menerima /start.
	if err := telegramPoller.Run(ctx); err != nil {
		if ctx.Err() != nil {
			log.Println("Bot dihentikan.")
			return
		}

		log.Fatalf(
			"Bot berhenti karena error: %v",
			err,
		)
	}

	log.Println("Program berhenti dengan aman.")
}
