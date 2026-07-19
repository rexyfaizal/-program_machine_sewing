package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"telegram_notif/bot"
	"telegram_notif/config"
	"telegram_notif/database"
	"telegram_notif/handler"
	"telegram_notif/repository"
	"telegram_notif/service"
	"telegram_notif/telegram"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Konfigurasi tidak valid: %v", err)
	}

	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Gagal terhubung ke SQL Server: %v", err)
	}
	defer db.Close()

	employeeRepository, err := repository.NewEmployeeRepository(
		db,
		cfg.EmployeeTable,
	)
	if err != nil {
		log.Fatalf("Gagal membuat employee repository: %v", err)
	}

	registrationService := service.NewRegistrationService(employeeRepository)
	telegramClient := telegram.NewClient(cfg.BotToken)
	telegramHandler := handler.NewTelegramHandler(
		registrationService,
		telegramClient,
	)
	poller := bot.NewPoller(telegramClient, telegramHandler)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	log.Printf(
		"Bot aktif. Table=%s",
		cfg.EmployeeTable,
	)
	log.Println("Tekan Ctrl+C untuk menghentikan program.")

	if err := poller.Run(ctx); err != nil {
		log.Fatalf("Bot berhenti karena error: %v", err)
	}

	log.Println("Bot berhenti dengan aman.")
}
