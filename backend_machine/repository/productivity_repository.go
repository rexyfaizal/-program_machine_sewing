package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"backend_machine/models"
	"backend_machine/utils"
)

func (r *Repository) GetMachineProductivity(ctx context.Context, m models.Machine, date string) (models.ProductivityRow, error) {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return models.ProductivityRow{}, fmt.Errorf("format tanggal harus YYYY-MM-DD")
	}

	start := day
	end := day.AddDate(0, 0, 1)

	ps, err := r.GetProductionStats(ctx, m.TableName, start, end)
	if err != nil {
		return models.ProductivityRow{}, err
	}

	runtimeSec, err := r.GetRuntimeSec(ctx, m.UUID, start, end)
	if err != nil {
		// Jika tabel record_runtime belum ada atau belum terisi,
		// dashboard tetap tampil memakai data ProcTime.
		log.Printf("runtime skip uuid=%s: %v", m.UUID, err)
		runtimeSec = 0
	}

	as, err := r.GetAlarmStats(ctx, m.UUID, start, end)
	if err != nil {
		// Jika tabel alarm belum ada, jangan gagalkan dashboard.
		log.Printf("alarm skip uuid=%s: %v", m.UUID, err)
		as = models.AlarmStats{}
	}

	runtimeHours := utils.Round2(float64(runtimeSec) / 3600)
	procHours := utils.Round2(float64(ps.ProcSec) / 3600)

	// LOSS TIME
	// Rumus:
	// Loss Time = Power On Duration - Running Time
	// Loss Time = Runtime - ProcTime
	lossTimeSec := runtimeSec - ps.ProcSec
	if lossTimeSec < 0 {
		lossTimeSec = 0
	}

	lossTimeHours := utils.Round2(float64(lossTimeSec) / 3600)

	// PRODUKTIVITAS
	// Rumus:
	// Productivity = Running Time / Power On Duration x 100
	// Productivity = ProcTime / Runtime x 100
	productivityRaw := 0.0
	productivityPct := 0.0

	if runtimeSec > 0 {
		productivityRaw = float64(ps.ProcSec) / float64(runtimeSec) * 100
		productivityPct = productivityRaw
	}

	if productivityPct > 100 {
		productivityPct = 100
	}

	status := utils.StatusFromPct(productivityPct)

	productivityBaseSec := ps.ProcSec
	mainSource := "process_time_runtime"

	row := models.ProductivityRow{
		Date:             date,
		UUID:             m.UUID,
		TableName:        m.TableName,
		NickName:         m.NickName,
		OriginalNickName: m.NickName,
		Location:         "",
		Pic:              "",
		Spv:              "",
		IP:               m.IP,
		MacType:          m.MacType,
		MacState:         m.MacState,

		RuntimeSec:   runtimeSec,
		RuntimeHours: runtimeHours,

		LossTimeSec:   lossTimeSec,
		LossTimeHours: lossTimeHours,

		ProductivityRaw: utils.Round2(productivityRaw),
		ProductivityPct: utils.Round2(productivityPct),
		Status:          status,
		MainSource:      mainSource,

		ProcSec:    ps.ProcSec,
		ProcHours:  procHours,
		OkProcSec:  ps.OkProcSec,
		Output:     ps.Output,
		Cycles:     ps.Cycles,
		Complete:   ps.Complete,
		Incomplete: ps.Incomplete,

		AvgCycle:   utils.Round2(ps.AvgCycle),
		MinCycle:   utils.Round2(ps.MinCycle),
		MaxCycle:   utils.Round2(ps.MaxCycle),
		SlowCycles: ps.SlowCycles,

		UniqueFiles: ps.UniqueFiles,
		TopFile:     ps.TopFile,

		AlarmCount: as.AlarmCount,
		AlarmTypes: as.AlarmTypes,

		FirstProcess: ps.FirstProcess,
		LastProcess:  ps.LastProcess,

		// Alias lama agar dashboard versi sebelumnya tetap aman.
		MachineName:       m.NickName,
		ProductiveSeconds: productivityBaseSec,
		ProductiveHours:   procHours,
		Category:          status,
		OutputOK:          ps.Complete,
		TotalLog:          ps.Cycles,
		FirstStart:        ps.FirstProcess,
		LastStart:         ps.LastProcess,
	}

	// =====================================================
	// FINAL SAFETY GUARD
	// Pengaman akhir sebelum response dikirim ke frontend.
	//
	// Tujuan:
	// 1. Power On Duration tidak boleh lebih kecil dari Running Time.
	// 2. Loss Time tidak boleh minus.
	// 3. Productivity tidak boleh lebih dari 100%.
	// 4. Status harus mengikuti ProductivityPct final.
	// =====================================================

	if row.ProcSec > row.RuntimeSec {
		row.RuntimeSec = row.ProcSec
	}

	row.LossTimeSec = row.RuntimeSec - row.ProcSec
	if row.LossTimeSec < 0 {
		row.LossTimeSec = 0
	}

	row.RuntimeHours = utils.Round2(float64(row.RuntimeSec) / 3600)
	row.ProcHours = utils.Round2(float64(row.ProcSec) / 3600)
	row.LossTimeHours = utils.Round2(float64(row.LossTimeSec) / 3600)

	finalProductivityRaw := 0.0
	finalProductivityPct := 0.0

	if row.RuntimeSec > 0 {
		finalProductivityRaw = float64(row.ProcSec) / float64(row.RuntimeSec) * 100
		finalProductivityPct = finalProductivityRaw
	}

	if finalProductivityPct > 100 {
		finalProductivityPct = 100
	}

	row.ProductivityRaw = utils.Round2(finalProductivityRaw)
	row.ProductivityPct = utils.Round2(finalProductivityPct)

	row.Status = utils.StatusFromPct(row.ProductivityPct)
	row.Category = row.Status

	row.ProductiveSeconds = row.ProcSec
	row.ProductiveHours = row.ProcHours

	return row, nil
}
