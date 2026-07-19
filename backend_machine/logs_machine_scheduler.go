package main

import (
	"context"
	"log"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/repository"
)

func startLogsMachineScheduler(repo *repository.Repository, interval time.Duration) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Println("Gagal load timezone Asia/Jakarta, pakai time.Local:", err)
		loc = time.Local
	}

	runJob := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		date := time.Now().In(loc).Format("2006-01-02")

		rows, err := buildProductivityRowsForLogs(ctx, repo, date)
		if err != nil {
			log.Println("Scheduler logs_machine gagal build productivity rows:", err)
			return
		}

		if err := repo.SaveLogsMachine(ctx, date, rows); err != nil {
			log.Println("Scheduler logs_machine gagal simpan:", err)
			return
		}

		log.Printf("Scheduler logs_machine updated. date=%s total=%d\n", date, len(rows))
	}

	go func() {
		// Jalankan sekali saat backend start.
		runJob()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			runJob()
		}
	}()
}

func buildProductivityRowsForLogs(ctx context.Context, repo *repository.Repository, date string) ([]models.ProductivityRow, error) {
	machines, err := repo.GetMachines(ctx)
	if err != nil {
		return nil, err
	}

	rows := make([]models.ProductivityRow, 0)

	for _, m := range machines {
		row, err := repo.GetMachineProductivity(ctx, m, date)
		if err != nil {
			log.Printf("scheduler skip machine table=%s uuid=%s: %v", m.TableName, m.UUID, err)
			continue
		}

		rows = append(rows, row)
	}

	settings, err := repo.GetMachineSettings(ctx)
	if err != nil {
		log.Println("scheduler gagal ambil machine setting manual:", err)
		return rows, nil
	}

	for i := range rows {
		rows[i].OriginalNickName = rows[i].NickName

		key := strings.ToLower(strings.TrimSpace(rows[i].UUID))

		if setting, ok := settings[key]; ok {
			if setting.CustomName != "" {
				rows[i].NickName = setting.CustomName
				rows[i].MachineName = setting.CustomName
			}

			rows[i].Location = setting.Location
			rows[i].Pic = setting.Pic
			rows[i].Spv = setting.Spv
		}
	}

	return rows, nil
}
