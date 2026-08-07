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
	// Hari penuh / Normal: window kalender lokal 00:00 → +1 hari.
	day, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return models.ProductivityRow{}, fmt.Errorf("format tanggal harus YYYY-MM-DD")
	}

	start := day
	end := day.AddDate(0, 0, 1)

	ps, err := r.GetProductionStats(ctx, m.TableName, start, end)
	if err != nil {
		return models.ProductivityRow{}, err
	}

	// Power On = SUM(RunTime) langsung dari DB (apa adanya).
	runtimeSec, err := r.GetRecordRunTimeSecSum(ctx, m.UUID, start, end)
	if err != nil {
		log.Printf("runtime skip uuid=%s: %v", m.UUID, err)
		runtimeSec = 0
	}

	as, err := r.GetAlarmStats(ctx, m.UUID, start, end)
	if err != nil {
		log.Printf("alarm skip uuid=%s: %v", m.UUID, err)
		as = models.AlarmStats{}
	}

	row := buildProductivityRow(m, date, runtimeSec, ps, as)

	// Jangan timpa Power On dengan ProcTime — tetap nilai kolom RunTime.
	row.RuntimeSec = runtimeSec
	if row.RuntimeSec < 0 {
		row.RuntimeSec = 0
	}
	row.LossTimeSec = row.RuntimeSec - row.ProcSec
	if row.LossTimeSec < 0 {
		row.LossTimeSec = 0
	}
	row.RuntimeHours = utils.Round2(float64(row.RuntimeSec) / 3600)
	row.LossTimeHours = utils.Round2(float64(row.LossTimeSec) / 3600)

	productivityRaw := 0.0
	if row.RuntimeSec > 0 {
		productivityRaw = float64(row.ProcSec) / float64(row.RuntimeSec) * 100
	}
	if productivityRaw > 100 {
		productivityRaw = 100
	}
	row.ProductivityRaw = utils.Round2(productivityRaw)
	row.ProductivityPct = utils.Round2(productivityRaw)
	row.Status = utils.StatusFromPct(row.ProductivityPct)
	row.Category = row.Status

	row.MainSource = "record_runtime_column"
	row.ShiftCode = utils.ShiftNormal
	row.ShiftName = utils.ShiftDisplayName(utils.ShiftNormal)

	return row, nil
}

func buildProductivityRow(
	m models.Machine,
	date string,
	runtimeSec int64,
	ps models.ProductionStats,
	as models.AlarmStats,
) models.ProductivityRow {
	runtimeHours := utils.Round2(float64(runtimeSec) / 3600)
	procHours := utils.Round2(float64(ps.ProcSec) / 3600)

	lossTimeSec := runtimeSec - ps.ProcSec
	if lossTimeSec < 0 {
		lossTimeSec = 0
	}

	lossTimeHours := utils.Round2(float64(lossTimeSec) / 3600)

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
		MainSource:      "process_time_runtime",

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

		MachineName:       m.NickName,
		ProductiveSeconds: ps.ProcSec,
		ProductiveHours:   procHours,
		Category:          status,
		OutputOK:          ps.Complete,
		TotalLog:          ps.Cycles,
		FirstStart:        ps.FirstProcess,
		LastStart:         ps.LastProcess,
	}

	// Power On tetap dari Record_RunTime / query shift (jangan timpa dengan Running).
	// Jika Running > Power On: Loss = 0, produktivitas di-cap 100%.
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

	return row
}
