package main

import (
	"context"
	"log"
	"time"

	"backend_machine/repository"
)

func startMachineOperatorAutoLogoutScheduler(repo *repository.Repository, interval time.Duration) {
	runJob := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		resp, err := repo.RunMachineOfflineAutoLogout(ctx)
		if err != nil {
			log.Println("Scheduler auto logout offline gagal:", err)
			return
		}

		log.Printf(
			"Scheduler auto logout offline selesai. total=%d online=%d offline=%d autoLogout=%d checkedAt=%s\n",
			resp.TotalMachines,
			resp.OnlineMachines,
			resp.OfflineMachines,
			resp.AutoLogoutSessions,
			resp.CheckedAt,
		)
	}

	go func() {
		runJob()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			runJob()
		}
	}()
}
