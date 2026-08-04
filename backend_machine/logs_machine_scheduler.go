package main

import (
	"context"
	"log"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/repository"
	"backend_machine/utils"
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

		now := time.Now().In(loc)
		date := now.Format("2006-01-02")
		gm3WorkDate := utils.ResolveGM3WorkDate(now).Format("2006-01-02")

		rows, err := buildProductivityRowsForLogs(ctx, repo, date, gm3WorkDate)
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

func buildProductivityRowsForLogs(
	ctx context.Context,
	repo *repository.Repository,
	date string,
	gm3WorkDate string,
) ([]models.ProductivityRow, error) {
	machines, err := repo.GetMachines(ctx)
	if err != nil {
		return nil, err
	}

	settings, err := repo.GetMachineSettings(ctx)
	if err != nil {
		log.Println("scheduler gagal ambil machine setting manual:", err)
		settings = map[string]models.MachineSetting{}
	}

	shiftConfigMap, err := repo.GetLineShiftConfigMap(ctx)
	if err != nil {
		log.Println("scheduler gagal ambil line shift config:", err)
		shiftConfigMap = map[string]models.LineShiftConfig{}
	}

	rows := make([]models.ProductivityRow, 0)

	for _, m := range machines {
		key := strings.ToLower(strings.TrimSpace(m.UUID))
		location := ""
		if setting, ok := settings[key]; ok {
			location = setting.Location
		}

		var row models.ProductivityRow

		useShift, segments, schedule := utils.ResolveShiftSegmentsForLocation(location, shiftConfigMap)
		if useShift {
			row, err = repo.GetMachineProductivityByShift(ctx, m, gm3WorkDate, utils.ShiftALL, segments, schedule)
		} else {
			row, err = repo.GetMachineProductivity(ctx, m, date)
		}

		if err != nil {
			log.Printf("scheduler skip machine table=%s uuid=%s: %v", m.TableName, m.UUID, err)
			continue
		}

		row.OriginalNickName = row.NickName

		if setting, ok := settings[key]; ok {
			if setting.CustomName != "" {
				row.NickName = setting.CustomName
				row.MachineName = setting.CustomName
			}

			row.Location = setting.Location
			row.Pic = setting.Pic
			row.Spv = setting.Spv
		}

		rows = append(rows, row)
	}

	return rows, nil
}
