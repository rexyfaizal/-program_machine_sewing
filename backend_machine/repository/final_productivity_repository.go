package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/utils"
)

// QueryFinalProductivityGroups menjalankan rumus FINAL NORMAL/SHIFT.
// selectedShift kosong / ALL → semua shift aktif (satu baris per shift).
func (r *Repository) QueryFinalProductivityGroups(
	ctx context.Context,
	uuid, processTable, workDate, mode, area, lineName, selectedShift string,
) ([]models.FinalProductivityGroup, error) {
	uuid = strings.TrimSpace(uuid)
	processTable = strings.TrimSpace(processTable)
	workDate = strings.TrimSpace(workDate)
	mode = strings.ToUpper(strings.TrimSpace(mode))
	area = strings.ToUpper(strings.TrimSpace(area))
	lineName = strings.ToUpper(strings.TrimSpace(lineName))
	selectedShift = strings.ToUpper(strings.TrimSpace(selectedShift))

	if uuid == "" {
		return nil, fmt.Errorf("uuid kosong")
	}
	if !utils.SafeTableName(processTable) {
		return nil, fmt.Errorf("nama tabel proses tidak aman: %s", processTable)
	}
	if _, err := ResolveAreaWorkDate(workDate); err != nil {
		return nil, err
	}
	if mode != "NORMAL" && mode != "SHIFT" {
		return nil, fmt.Errorf("mode harus NORMAL atau SHIFT")
	}
	if mode == "SHIFT" && area == "" {
		return nil, fmt.Errorf("area wajib untuk mode SHIFT")
	}
	if selectedShift == "ALL" || selectedShift == "CURRENT" {
		selectedShift = ""
	}

	rows, err := r.DB.QueryContext(
		ctx,
		finalProductivitySQL,
		sql.Named("p_work_date", workDate),
		sql.Named("p_uuid", uuid),
		sql.Named("p_area", area),
		sql.Named("p_line", lineName),
		sql.Named("p_mode", mode),
		sql.Named("p_selected_shift", selectedShift),
		sql.Named("p_process_table", processTable),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.FinalProductivityGroup, 0)
	for rows.Next() {
		var g models.FinalProductivityGroup
		var areaNull, breakStart, breakEnd sql.NullString
		var productivity sql.NullFloat64
		if err := rows.Scan(
			&g.Mode,
			&g.WorkDate,
			&areaNull,
			&g.ShiftName,
			&g.PeriodStart,
			&g.PeriodEnd,
			&breakStart,
			&breakEnd,
			&g.PowerSeconds,
			&g.ProcessSeconds,
			&g.LossSeconds,
			&productivity,
		); err != nil {
			return nil, err
		}
		g.Area = areaNull.String
		g.BreakStart = breakStart.String
		g.BreakEnd = breakEnd.String
		if productivity.Valid {
			g.Productivity = productivity.Float64
		}
		g.ShiftName = strings.ToUpper(strings.TrimSpace(g.ShiftName))
		out = append(out, g)
	}
	return out, rows.Err()
}

// AggregateFinalGroups menjumlahkan beberapa shift jadi satu total mesin.
func AggregateFinalGroups(groups []models.FinalProductivityGroup) models.FinalProductivityGroup {
	if len(groups) == 0 {
		return models.FinalProductivityGroup{}
	}
	if len(groups) == 1 {
		return groups[0]
	}

	agg := models.FinalProductivityGroup{
		Mode:        groups[0].Mode,
		WorkDate:    groups[0].WorkDate,
		Area:        groups[0].Area,
		ShiftName:   "ALL",
		PeriodStart: groups[0].PeriodStart,
		PeriodEnd:   groups[0].PeriodEnd,
	}
	for _, g := range groups {
		agg.PowerSeconds += g.PowerSeconds
		agg.ProcessSeconds += g.ProcessSeconds
		if g.PeriodStart != "" && (agg.PeriodStart == "" || g.PeriodStart < agg.PeriodStart) {
			agg.PeriodStart = g.PeriodStart
		}
		if g.PeriodEnd != "" && (agg.PeriodEnd == "" || g.PeriodEnd > agg.PeriodEnd) {
			agg.PeriodEnd = g.PeriodEnd
		}
	}
	agg.LossSeconds = agg.PowerSeconds - agg.ProcessSeconds
	if agg.LossSeconds < 0 {
		agg.LossSeconds = 0
	}
	if agg.PowerSeconds > 0 {
		agg.Productivity = float64(agg.ProcessSeconds) * 100.0 / float64(agg.PowerSeconds)
		if agg.Productivity > 100 {
			agg.Productivity = 100
		}
	}
	return agg
}

// GetMachineProductivityFinal menggabungkan rumus FINAL + output/CT/alarm window.
func (r *Repository) GetMachineProductivityFinal(
	ctx context.Context,
	m models.Machine,
	workDate, mode, area, lineName, selectedShift string,
) (models.ProductivityRow, error) {
	tableName := strings.TrimSpace(m.TableName)
	if tableName == "" {
		tableName = "m" + strings.TrimSpace(m.UUID)
	}

	groups, err := r.QueryFinalProductivityGroups(
		ctx,
		m.UUID,
		tableName,
		workDate,
		mode,
		area,
		lineName,
		selectedShift,
	)
	if err != nil {
		return models.ProductivityRow{}, err
	}
	if len(groups) == 0 {
		return models.ProductivityRow{}, fmt.Errorf("hasil FINAL kosong")
	}

	code := strings.ToUpper(strings.TrimSpace(selectedShift))
	agg := groups[0]
	if mode == "SHIFT" && (code == "" || code == "ALL") {
		agg = AggregateFinalGroups(groups)
		code = utils.ShiftALL
	} else if mode == "NORMAL" {
		code = utils.ShiftNormal
	} else {
		code = strings.ToUpper(strings.TrimSpace(agg.ShiftName))
	}

	periodStart, periodEnd := parseFinalPeriod(agg.PeriodStart, agg.PeriodEnd, workDate)

	ps, err := r.GetProductionStats(ctx, tableName, periodStart, periodEnd)
	if err != nil {
		return models.ProductivityRow{}, err
	}
	// Process time pakai rumus FINAL (intersection), bukan SUM ProcTime mentah.
	ps.ProcSec = agg.ProcessSeconds

	as, err := r.GetAlarmStats(ctx, m.UUID, periodStart, periodEnd)
	if err != nil {
		log.Printf("alarm skip uuid=%s: %v", m.UUID, err)
		as = models.AlarmStats{}
	}

	row := buildProductivityRow(m, workDate, agg.PowerSeconds, ps, as)
	row.RuntimeSec = agg.PowerSeconds
	row.ProcSec = agg.ProcessSeconds
	row.LossTimeSec = agg.LossSeconds
	if row.LossTimeSec < 0 {
		row.LossTimeSec = 0
	}
	row.RuntimeHours = utils.Round2(float64(row.RuntimeSec) / 3600)
	row.ProcHours = utils.Round2(float64(row.ProcSec) / 3600)
	row.LossTimeHours = utils.Round2(float64(row.LossTimeSec) / 3600)

	pct := agg.Productivity
	if pct > 100 {
		pct = 100
	}
	row.ProductivityRaw = utils.Round2(pct)
	row.ProductivityPct = utils.Round2(pct)
	row.Status = utils.StatusFromPct(row.ProductivityPct)
	row.Category = row.Status
	row.ProductiveSeconds = row.ProcSec
	row.ProductiveHours = row.ProcHours

	row.ShiftCode = code
	if mode == "NORMAL" {
		row.ShiftName = utils.ShiftDisplayName(utils.ShiftNormal)
	} else if code == utils.ShiftALL {
		row.ShiftName = utils.ShiftDisplayName(utils.ShiftALL)
	} else {
		row.ShiftName = agg.ShiftName
		if row.ShiftName == "" {
			row.ShiftName = utils.ShiftDisplayName(code)
		}
	}
	row.MainSource = "final_shift_setting"

	return row, nil
}

func parseFinalPeriod(startText, endText, workDate string) (time.Time, time.Time) {
	day, err := time.ParseInLocation("2006-01-02", workDate, time.Local)
	if err != nil {
		now := time.Now()
		return now, now.Add(24 * time.Hour)
	}
	start := day
	end := day.AddDate(0, 0, 1)

	if t, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(startText), time.Local); err == nil {
		start = t
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(endText), time.Local); err == nil {
		end = t
	}
	return start, end
}
