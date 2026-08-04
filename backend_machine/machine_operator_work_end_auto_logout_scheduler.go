package main

import (
	"context"
	"log"
	"time"

	"backend_machine/repository"
)

func startMachineOperatorWorkEndAutoLogoutScheduler(repo *repository.Repository, interval time.Duration) {
	go func() {
		log.Println("Machine operator work end auto logout scheduler aktif")

		runMachineOperatorWorkEndAutoLogout(repo)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			runMachineOperatorWorkEndAutoLogout(repo)
		}
	}()
}

func runMachineOperatorWorkEndAutoLogout(repo *repository.Repository) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	totalClosed, err := repo.RunMachineOperatorWorkEndAutoLogout(ctx)
	if err != nil {
		log.Println("AUTO_LOGOUT_WORK_END error:", err)
		return
	}

	if totalClosed > 0 {
		log.Printf("AUTO_LOGOUT_WORK_END selesai, session ditutup: %d", totalClosed)
	}
}
