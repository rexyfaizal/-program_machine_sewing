package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"backend_machine/models"
)

// EnsureShiftSettingSchema memastikan kolom line_name ada (Opsi B: shift per line).
func (r *Repository) EnsureShiftSettingSchema(ctx context.Context) error {
	_, err := r.DB.ExecContext(ctx, `
IF OBJECT_ID(N'dbo.shift_setting', N'U') IS NOT NULL
   AND COL_LENGTH('dbo.shift_setting', 'line_name') IS NULL
BEGIN
    ALTER TABLE dbo.shift_setting ADD line_name NVARCHAR(255) NOT NULL
        CONSTRAINT DF_shift_setting_line_name DEFAULT N'';
END;
`)
	return err
}

// HasActiveShiftSettings true jika area punya jadwal aktif di dbo.shift_setting.
func (r *Repository) HasActiveShiftSettings(ctx context.Context, area, workDate string) (bool, error) {
	area = strings.ToUpper(strings.TrimSpace(area))
	workDate = strings.TrimSpace(workDate)
	if area == "" || workDate == "" {
		return false, nil
	}

	var n int
	err := r.DB.QueryRowContext(ctx, `
SELECT COUNT(1)
FROM dbo.shift_setting s
WHERE UPPER(LTRIM(RTRIM(s.area))) = @area
  AND s.is_active = 1
  AND s.effective_from <= TRY_CONVERT(date, @work_date)
  AND (s.effective_to IS NULL OR s.effective_to >= TRY_CONVERT(date, @work_date));
`,
		sql.Named("area", area),
		sql.Named("work_date", workDate),
	).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// ListActiveShiftSettings membaca jadwal shift aktif per area + tanggal.
func (r *Repository) ListActiveShiftSettings(ctx context.Context, area, workDate string) ([]models.ShiftSetting, error) {
	area = strings.ToUpper(strings.TrimSpace(area))
	workDate = strings.TrimSpace(workDate)
	if area == "" || workDate == "" {
		return nil, nil
	}

	rows, err := r.DB.QueryContext(ctx, `
SELECT
    s.id,
    LTRIM(RTRIM(s.area)) AS area,
    ISNULL(LTRIM(RTRIM(s.line_name)), N'') AS line_name,
    s.shift_no,
    LTRIM(RTRIM(s.shift_name)) AS shift_name,
    CONVERT(varchar(8), s.start_time, 108) AS start_time,
    CONVERT(varchar(8), s.end_time, 108) AS end_time,
    CONVERT(varchar(8), s.break_start, 108) AS break_start,
    CONVERT(varchar(8), s.break_end, 108) AS break_end,
    CONVERT(varchar(10), s.effective_from, 23) AS effective_from,
    CONVERT(varchar(10), s.effective_to, 23) AS effective_to,
    CAST(s.is_active AS int) AS is_active,
    CONVERT(varchar(19), s.updated_at, 120) AS updated_at
FROM dbo.shift_setting s
WHERE UPPER(LTRIM(RTRIM(s.area))) = @area
  AND s.is_active = 1
  AND s.effective_from <= TRY_CONVERT(date, @work_date)
  AND (s.effective_to IS NULL OR s.effective_to >= TRY_CONVERT(date, @work_date))
ORDER BY s.line_name, s.shift_no, s.shift_name;
`,
		sql.Named("area", area),
		sql.Named("work_date", workDate),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanShiftSettingRows(rows)
}

// ListShiftSettingsForArea semua baris aktif terkini untuk UI admin (tanpa filter tanggal ketat).
func (r *Repository) ListShiftSettingsForArea(ctx context.Context, area string) ([]models.ShiftSetting, error) {
	area = strings.ToUpper(strings.TrimSpace(area))
	if area == "" {
		return nil, nil
	}

	rows, err := r.DB.QueryContext(ctx, `
SELECT
    s.id,
    LTRIM(RTRIM(s.area)) AS area,
    ISNULL(LTRIM(RTRIM(s.line_name)), N'') AS line_name,
    s.shift_no,
    LTRIM(RTRIM(s.shift_name)) AS shift_name,
    CONVERT(varchar(8), s.start_time, 108) AS start_time,
    CONVERT(varchar(8), s.end_time, 108) AS end_time,
    CONVERT(varchar(8), s.break_start, 108) AS break_start,
    CONVERT(varchar(8), s.break_end, 108) AS break_end,
    CONVERT(varchar(10), s.effective_from, 23) AS effective_from,
    CONVERT(varchar(10), s.effective_to, 23) AS effective_to,
    CAST(s.is_active AS int) AS is_active,
    CONVERT(varchar(19), s.updated_at, 120) AS updated_at
FROM dbo.shift_setting s
WHERE UPPER(LTRIM(RTRIM(s.area))) = @area
  AND s.is_active = 1
ORDER BY s.line_name, s.shift_no, s.shift_name;
`,
		sql.Named("area", area),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanShiftSettingRows(rows)
}

func scanShiftSettingRows(rows *sql.Rows) ([]models.ShiftSetting, error) {
	out := make([]models.ShiftSetting, 0)
	for rows.Next() {
		var item models.ShiftSetting
		var lineName, breakStart, breakEnd, effectiveTo, updatedAt sql.NullString
		var active int
		if err := rows.Scan(
			&item.ID,
			&item.Area,
			&lineName,
			&item.ShiftNo,
			&item.ShiftName,
			&item.StartTime,
			&item.EndTime,
			&breakStart,
			&breakEnd,
			&item.EffectiveFrom,
			&effectiveTo,
			&active,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		item.LineName = strings.ToUpper(strings.TrimSpace(lineName.String))
		item.BreakStart = breakStart.String
		item.BreakEnd = breakEnd.String
		item.EffectiveTo = effectiveTo.String
		item.UpdatedAt = updatedAt.String
		item.IsActive = active == 1
		item.Area = strings.ToUpper(strings.TrimSpace(item.Area))
		item.ShiftName = strings.ToUpper(strings.TrimSpace(item.ShiftName))
		out = append(out, item)
	}
	return out, rows.Err()
}

func normalizeClockToSQLTime(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("jam kosong")
	}
	if len(value) == 5 {
		value += ":00"
	}
	if _, err := time.Parse("15:04:05", value); err != nil {
		return "", fmt.Errorf("format jam harus HH:mm")
	}
	return value, nil
}

func nullableSQLTime(value string) sql.NullString {
	value = strings.TrimSpace(value)
	if value == "" {
		return sql.NullString{}
	}
	if len(value) == 5 {
		value += ":00"
	}
	return sql.NullString{String: value, Valid: true}
}

// ReplaceAreaShiftSettings menonaktifkan jadwal lama area lalu insert jadwal default
// area baru (line_name = ''). Dipertahankan untuk kompatibilitas pemanggil lama.
func (r *Repository) ReplaceAreaShiftSettings(ctx context.Context, area string, shifts []models.ShiftSetting) error {
	return r.ReplaceShiftSettings(ctx, area, shifts, nil)
}

// ReplaceShiftSettings menonaktifkan seluruh jadwal area, lalu insert:
//   - defaultShifts → line_name = '' (jadwal umum area)
//   - lineOverrides[LINE] → line_name = LINE (khusus line yang custom)
func (r *Repository) ReplaceShiftSettings(
	ctx context.Context,
	area string,
	defaultShifts []models.ShiftSetting,
	lineOverrides map[string][]models.ShiftSetting,
) error {
	area = strings.ToUpper(strings.TrimSpace(area))
	if area == "" {
		return fmt.Errorf("area kosong")
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Purge: hapus permanen seluruh baris area ini (tidak menyimpan history),
	// lalu insert jadwal aktif terbaru. Menjaga tabel tetap ringkas.
	_, err = tx.ExecContext(ctx, `
DELETE FROM dbo.shift_setting
WHERE UPPER(LTRIM(RTRIM(area))) = @area;
`, sql.Named("area", area))
	if err != nil {
		return err
	}

	if err := insertShiftRows(ctx, tx, area, "", defaultShifts); err != nil {
		return err
	}

	for line, shifts := range lineOverrides {
		lineKey := strings.ToUpper(strings.TrimSpace(line))
		if lineKey == "" || len(shifts) == 0 {
			continue
		}
		if err := insertShiftRows(ctx, tx, area, lineKey, shifts); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func insertShiftRows(ctx context.Context, tx *sql.Tx, area, lineName string, shifts []models.ShiftSetting) error {
	insertQuery := `
INSERT INTO dbo.shift_setting
(
    area, line_name, shift_no, shift_name,
    start_time, end_time, break_start, break_end,
    effective_from, effective_to, is_active, updated_at
)
VALUES
(
    @area, @line_name, @shift_no, @shift_name,
    @start_time, @end_time, @break_start, @break_end,
    CAST(SYSDATETIME() AS date), NULL, 1, SYSDATETIME()
);
`

	for i, shift := range shifts {
		name := strings.ToUpper(strings.TrimSpace(shift.ShiftName))
		if name == "" {
			name = fmt.Sprintf("SHIFT_%d", i+1)
		}
		shiftNo := shift.ShiftNo
		if shiftNo <= 0 {
			shiftNo = i + 1
		}

		startTime, err := normalizeClockToSQLTime(shift.StartTime)
		if err != nil {
			return fmt.Errorf("%s start: %w", name, err)
		}
		endTime, err := normalizeClockToSQLTime(shift.EndTime)
		if err != nil {
			return fmt.Errorf("%s end: %w", name, err)
		}

		_, err = tx.ExecContext(
			ctx,
			insertQuery,
			sql.Named("area", area),
			sql.Named("line_name", strings.ToUpper(strings.TrimSpace(lineName))),
			sql.Named("shift_no", shiftNo),
			sql.Named("shift_name", name),
			sql.Named("start_time", startTime),
			sql.Named("end_time", endTime),
			sql.Named("break_start", nullableSQLTime(shift.BreakStart)),
			sql.Named("break_end", nullableSQLTime(shift.BreakEnd)),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// ShiftSettingsToScheduleItems konversi ke format schedule UI/legacy.
func ShiftSettingsToScheduleItems(settings []models.ShiftSetting) []models.ShiftScheduleItem {
	out := make([]models.ShiftScheduleItem, 0, len(settings))
	for _, s := range settings {
		out = append(out, models.ShiftScheduleItem{
			Code:       strings.ToUpper(strings.TrimSpace(s.ShiftName)),
			Start:      trimHHMM(s.StartTime),
			End:        trimHHMM(s.EndTime),
			BreakStart: trimHHMM(s.BreakStart),
			BreakEnd:   trimHHMM(s.BreakEnd),
		})
	}
	return out
}

func trimHHMM(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 5 {
		return value[:5]
	}
	return value
}

// ResolveAreaWorkDate helper parse tanggal.
func ResolveAreaWorkDate(workDate string) (string, error) {
	workDate = strings.TrimSpace(workDate)
	if workDate == "" {
		return "", fmt.Errorf("work_date kosong")
	}
	if _, err := time.Parse("2006-01-02", workDate); err != nil {
		return "", fmt.Errorf("format tanggal harus YYYY-MM-DD")
	}
	return workDate, nil
}
