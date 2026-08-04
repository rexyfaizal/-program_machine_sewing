package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/utils"
)

// GetShiftRuntimeSec menghitung runtime mesin per segment shift (exclude istirahat).
func (r *Repository) GetShiftRuntimeSec(
	ctx context.Context,
	uuid string,
	workDate time.Time,
	shiftCode string,
	segments []utils.ShiftSegment,
) (int64, error) {
	uuid = strings.TrimSpace(uuid)
	if uuid == "" {
		return 0, nil
	}

	filtered := utils.FilterSegmentsByShift(segments, shiftCode)
	if len(filtered) == 0 {
		return 0, nil
	}

	valuesSQL := utils.BuildShiftSegmentValuesSQL(filtered)
	if valuesSQL == "" {
		return 0, nil
	}

	workStartMin, workEndMin := utils.SegmentBounds(filtered)

	query := fmt.Sprintf(`
DECLARE @work_date DATE = @p_work_date;
DECLARE @uuid NVARCHAR(100) = @p_uuid;
DECLARE @work_start_min INT = @p_work_start_min;
DECLARE @work_end_min INT = @p_work_end_min;

DECLARE @base DATETIME2 = CAST(@work_date AS DATETIME2);
DECLARE @work_start DATETIME2 = DATEADD(MINUTE, @work_start_min, @base);
DECLARE @work_end DATETIME2 = DATEADD(MINUTE, @work_end_min, @base);
DECLARE @server_now DATETIME2 = SYSDATETIME();

WITH shift_segments AS
(
    SELECT * FROM (VALUES
        %s
    ) AS x(shift_no, shift_name, segment_start, segment_end)
),
raw_runtime AS
(
    SELECT
        TRY_CONVERT(DATETIME2, [StartTime]) AS start_time,
        COALESCE(
            TRY_CONVERT(DATETIME2, [ShutTime]),
            CASE
                WHEN @server_now < @work_end THEN @server_now
                ELSE @work_end
            END
        ) AS end_time
    FROM [dbo].[Record_RunTime]
    WHERE LOWER(LTRIM(RTRIM([UUID]))) = LOWER(LTRIM(RTRIM(@uuid)))
      AND TRY_CONVERT(DATETIME2, [StartTime]) < @work_end
      AND COALESCE(
            TRY_CONVERT(DATETIME2, [ShutTime]),
            @server_now
          ) > @work_start
),
valid_runtime AS
(
    SELECT start_time, end_time
    FROM raw_runtime
    WHERE start_time IS NOT NULL
      AND end_time IS NOT NULL
      AND end_time > start_time
),
ordered_runtime AS
(
    SELECT
        start_time,
        end_time,
        MAX(end_time) OVER (
            ORDER BY start_time, end_time
            ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING
        ) AS previous_max_end
    FROM valid_runtime
),
marked_runtime AS
(
    SELECT
        start_time,
        end_time,
        CASE
            WHEN previous_max_end IS NULL OR start_time > previous_max_end THEN 1
            ELSE 0
        END AS new_group
    FROM ordered_runtime
),
grouped_runtime AS
(
    SELECT
        start_time,
        end_time,
        SUM(new_group) OVER (
            ORDER BY start_time, end_time
            ROWS UNBOUNDED PRECEDING
        ) AS group_id
    FROM marked_runtime
),
merged_runtime AS
(
    SELECT
        group_id,
        MIN(start_time) AS start_time,
        MAX(end_time) AS end_time
    FROM grouped_runtime
    GROUP BY group_id
),
shift_totals AS
(
    SELECT
        SUM(
            CASE
                WHEN r.group_id IS NULL THEN 0
                WHEN
                    CASE
                        WHEN r.start_time > s.segment_start THEN r.start_time
                        ELSE s.segment_start
                    END
                    <
                    CASE
                        WHEN r.end_time < s.segment_end THEN r.end_time
                        ELSE s.segment_end
                    END
                THEN DATEDIFF(
                    SECOND,
                    CASE WHEN r.start_time > s.segment_start THEN r.start_time ELSE s.segment_start END,
                    CASE WHEN r.end_time < s.segment_end THEN r.end_time ELSE s.segment_end END
                )
                ELSE 0
            END
        ) AS runtime_seconds
    FROM shift_segments AS s
    LEFT JOIN merged_runtime AS r
        ON r.start_time < s.segment_end
       AND r.end_time > s.segment_start
)
SELECT CAST(ISNULL(runtime_seconds, 0) AS BIGINT) FROM shift_totals;
`, valuesSQL)

	var runtimeSec sql.NullInt64

	err := r.DB.QueryRowContext(
		ctx,
		query,
		sql.Named("p_work_date", workDate.Format("2006-01-02")),
		sql.Named("p_uuid", uuid),
		sql.Named("p_work_start_min", workStartMin),
		sql.Named("p_work_end_min", workEndMin),
	).Scan(&runtimeSec)

	if err != nil {
		return 0, err
	}

	if !runtimeSec.Valid || runtimeSec.Int64 < 0 {
		return 0, nil
	}

	return runtimeSec.Int64, nil
}

// GetShiftProcessSec menghitung process time mesin per segment shift.
func (r *Repository) GetShiftProcessSec(
	ctx context.Context,
	tableName string,
	workDate time.Time,
	shiftCode string,
	segments []utils.ShiftSegment,
) (int64, error) {
	if !utils.SafeTableName(tableName) {
		return 0, fmt.Errorf("nama tabel tidak aman: %s", tableName)
	}

	filtered := utils.FilterSegmentsByShift(segments, shiftCode)
	if len(filtered) == 0 {
		return 0, nil
	}

	valuesSQL := utils.BuildShiftSegmentValuesSQL(filtered)
	if valuesSQL == "" {
		return 0, nil
	}

	workStartMin, workEndMin := utils.SegmentBounds(filtered)

	query := fmt.Sprintf(`
DECLARE @work_date DATE = @p_work_date;
DECLARE @work_start_min INT = @p_work_start_min;
DECLARE @work_end_min INT = @p_work_end_min;

DECLARE @base DATETIME2 = CAST(@work_date AS DATETIME2);
DECLARE @work_start DATETIME2 = DATEADD(MINUTE, @work_start_min, @base);
DECLARE @work_end DATETIME2 = DATEADD(MINUTE, @work_end_min, @base);
DECLARE @server_now DATETIME2 = SYSDATETIME();

DECLARE @cutoff_time DATETIME2 =
    CASE
        WHEN @server_now <= @work_start THEN @work_start
        WHEN @server_now >= @work_end THEN @work_end
        ELSE @server_now
    END;

WITH shift_segments AS
(
    SELECT * FROM (VALUES
        %s
    ) AS x(shift_no, shift_name, segment_start, segment_end)
),
process_raw AS
(
    SELECT
        TRY_CONVERT(DATETIME2, [StartTime]) AS process_start,
        TRY_CONVERT(INT, [ProcTime]) AS process_seconds
    FROM dbo.[%s]
    WHERE TRY_CONVERT(DATETIME2, [StartTime]) < @cutoff_time
),
process_data AS
(
    SELECT
        process_start,
        CASE
            WHEN DATEADD(SECOND, process_seconds, process_start) > @cutoff_time
            THEN @cutoff_time
            ELSE DATEADD(SECOND, process_seconds, process_start)
        END AS process_end
    FROM process_raw
    WHERE process_start IS NOT NULL
      AND process_seconds IS NOT NULL
      AND process_seconds > 0
      AND process_start < @cutoff_time
      AND DATEADD(SECOND, process_seconds, process_start) > @work_start
),
shift_totals AS
(
    SELECT
        SUM(
            CASE
                WHEN p.process_start IS NULL THEN 0
                WHEN
                    CASE
                        WHEN p.process_start > s.segment_start THEN p.process_start
                        ELSE s.segment_start
                    END
                    <
                    CASE
                        WHEN p.process_end < s.segment_end THEN p.process_end
                        ELSE s.segment_end
                    END
                THEN DATEDIFF(
                    SECOND,
                    CASE WHEN p.process_start > s.segment_start THEN p.process_start ELSE s.segment_start END,
                    CASE WHEN p.process_end < s.segment_end THEN p.process_end ELSE s.segment_end END
                )
                ELSE 0
            END
        ) AS process_seconds
    FROM shift_segments AS s
    LEFT JOIN process_data AS p
        ON p.process_start < s.segment_end
       AND p.process_end > s.segment_start
)
SELECT CAST(ISNULL(process_seconds, 0) AS BIGINT) FROM shift_totals;
`, valuesSQL, tableName)

	var procSec sql.NullInt64

	err := r.DB.QueryRowContext(
		ctx,
		query,
		sql.Named("p_work_date", workDate.Format("2006-01-02")),
		sql.Named("p_work_start_min", workStartMin),
		sql.Named("p_work_end_min", workEndMin),
	).Scan(&procSec)

	if err != nil {
		return 0, err
	}

	if !procSec.Valid || procSec.Int64 < 0 {
		return 0, nil
	}

	return procSec.Int64, nil
}

// GetGM3ShiftRuntimeSec kompatibilitas lama → default GM3 segments.
func (r *Repository) GetGM3ShiftRuntimeSec(
	ctx context.Context,
	uuid string,
	workDate time.Time,
	shiftCode string,
) (int64, error) {
	return r.GetShiftRuntimeSec(ctx, uuid, workDate, shiftCode, utils.GM3ShiftSegments())
}

// GetGM3ShiftProcessSec kompatibilitas lama → default GM3 segments.
func (r *Repository) GetGM3ShiftProcessSec(
	ctx context.Context,
	tableName string,
	workDate time.Time,
	shiftCode string,
) (int64, error) {
	return r.GetShiftProcessSec(ctx, tableName, workDate, shiftCode, utils.GM3ShiftSegments())
}

// GetMachineProductivityByShift menghitung produktivitas berbasis segment shift dinamis.
func (r *Repository) GetMachineProductivityByShift(
	ctx context.Context,
	m models.Machine,
	workDate string,
	shiftCode string,
	segments []utils.ShiftSegment,
	schedule []models.ShiftScheduleItem,
) (models.ProductivityRow, error) {
	day, err := time.ParseInLocation("2006-01-02", workDate, time.Local)
	if err != nil {
		return models.ProductivityRow{}, fmt.Errorf("format tanggal harus YYYY-MM-DD")
	}

	code := utils.NormalizeShiftCode(shiftCode)
	if code == utils.ShiftCurrent {
		code = utils.ShiftALL
	}

	runtimeSec, err := r.GetShiftRuntimeSec(ctx, m.UUID, day, code, segments)
	if err != nil {
		return models.ProductivityRow{}, err
	}

	procSec, err := r.GetShiftProcessSec(ctx, m.TableName, day, code, segments)
	if err != nil {
		return models.ProductivityRow{}, err
	}

	workStartMin, workEndMin := utils.SegmentBounds(segments)
	workStart := day.Add(time.Duration(workStartMin) * time.Minute)
	workEnd := day.Add(time.Duration(workEndMin) * time.Minute)

	ps, err := r.GetProductionStats(ctx, m.TableName, workStart, workEnd)
	if err != nil {
		return models.ProductivityRow{}, err
	}

	ps.ProcSec = procSec

	as, err := r.GetAlarmStats(ctx, m.UUID, workStart, workEnd)
	if err != nil {
		as = models.AlarmStats{}
	}

	row := buildProductivityRow(m, workDate, runtimeSec, ps, as)
	row.ShiftCode = code
	row.ShiftName = utils.ShiftDisplayNameFromSchedule(code, schedule)
	row.MainSource = "line_shift_process_runtime"

	return row, nil
}

// GetMachineProductivityGM3 kompatibilitas lama → default GM3.
func (r *Repository) GetMachineProductivityGM3(
	ctx context.Context,
	m models.Machine,
	workDate string,
	shiftCode string,
) (models.ProductivityRow, error) {
	return r.GetMachineProductivityByShift(
		ctx,
		m,
		workDate,
		shiftCode,
		utils.GM3ShiftSegments(),
		utils.DefaultGM3ScheduleItems(),
	)
}
