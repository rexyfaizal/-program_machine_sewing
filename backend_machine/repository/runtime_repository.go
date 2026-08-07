package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

func formatNaiveDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// GetRecordRunTimeSecSum = Power On apa adanya dari kolom RunTime.
// Setara query:
//
//	SELECT SUM(ISNULL(RunTime, 0))
//	FROM dbo.Record_RunTime
//	WHERE UUID = @uuid
//	  AND StartTime >= @start AND StartTime < @end
func (r *Repository) GetRecordRunTimeSecSum(
	ctx context.Context,
	uuid string,
	start, end time.Time,
) (int64, error) {
	uuid = strings.TrimSpace(uuid)
	if uuid == "" {
		return 0, nil
	}

	query := `
DECLARE @start_time DATETIME2 = TRY_CONVERT(DATETIME2, @p_start_time);
DECLARE @end_time DATETIME2 = TRY_CONVERT(DATETIME2, @p_end_time);

SELECT CAST(ISNULL(SUM(ISNULL([RunTime], 0)), 0) AS BIGINT) AS total_runtime_detik
FROM [dbo].[Record_RunTime]
WHERE LOWER(LTRIM(RTRIM([UUID]))) = LOWER(LTRIM(RTRIM(@uuid)))
  AND [StartTime] >= @start_time
  AND [StartTime] < @end_time;
`

	var runtimeSec sql.NullInt64
	err := r.DB.QueryRowContext(
		ctx,
		query,
		sql.Named("uuid", uuid),
		sql.Named("p_start_time", formatNaiveDateTime(start)),
		sql.Named("p_end_time", formatNaiveDateTime(end)),
	).Scan(&runtimeSec)
	if err != nil {
		return 0, err
	}
	if !runtimeSec.Valid || runtimeSec.Int64 < 0 {
		return 0, nil
	}
	return runtimeSec.Int64, nil
}

// GetRuntimeSec menghitung Power On Duration seperti aplikasi bawaan:
// StartTime → ShutTime (atau now jika masih ON / ShutTime kosong).
// Tidak menambah waktu setelah ShutTime.
func (r *Repository) GetRuntimeSec(ctx context.Context, uuid string, macState string, start, end time.Time) (int64, error) {
	uuid = strings.TrimSpace(uuid)

	if uuid == "" {
		return 0, nil
	}

	// macState tidak dipakai: status ON/OFF ditentukan dari ShutTime record.
	_ = macState

	// Kirim sebagai string naive supaya driver tidak geser timezone
	// (time.Time UTC pernah membuat window jadi H-1 17:00 → H 17:00).
	startText := start.UTC().Format("2006-01-02 15:04:05")
	endText := end.UTC().Format("2006-01-02 15:04:05")

	query := `
DECLARE @start_time DATETIME2 = TRY_CONVERT(DATETIME2, @p_start_time);
DECLARE @end_time DATETIME2 = TRY_CONVERT(DATETIME2, @p_end_time);
DECLARE @server_now DATETIME2 = SYSDATETIME();

-- Batas hitung: hari ini sampai now, tanggal lama sampai end of day.
DECLARE @cutoff DATETIME2 =
    CASE
        WHEN @server_now < @end_time THEN @server_now
        ELSE @end_time
    END;

WITH raw_runtime AS
(
    SELECT
        TRY_CONVERT(DATETIME2, [StartTime]) AS start_time,

        -- ShutTime ada = mesin sudah mati → pakai ShutTime.
        -- ShutTime kosong = masih ON → pakai cutoff (now / end of day).
        COALESCE(
            TRY_CONVERT(DATETIME2, [ShutTime]),
            @cutoff
        ) AS end_time
    FROM [sewingiot].[dbo].[Record_RunTime]
    WHERE LOWER(LTRIM(RTRIM([UUID]))) = LOWER(LTRIM(RTRIM(@uuid)))
      AND TRY_CONVERT(DATETIME2, [StartTime]) < @cutoff
      AND COALESCE(
            TRY_CONVERT(DATETIME2, [ShutTime]),
            @cutoff
          ) > @start_time
),
clipped_runtime AS
(
    SELECT
        CASE
            WHEN start_time > @start_time THEN start_time
            ELSE @start_time
        END AS start_time,
        CASE
            WHEN end_time < @cutoff THEN end_time
            ELSE @cutoff
        END AS end_time
    FROM raw_runtime
    WHERE start_time IS NOT NULL
      AND end_time IS NOT NULL
),
valid_runtime AS
(
    SELECT start_time, end_time
    FROM clipped_runtime
    WHERE end_time > start_time
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
)
SELECT
    CAST(ISNULL(SUM(DATEDIFF_BIG(SECOND, start_time, end_time)), 0) AS BIGINT) AS total_runtime_detik
FROM merged_runtime;
`

	var runtimeSec sql.NullInt64

	err := r.DB.QueryRowContext(
		ctx,
		query,
		sql.Named("uuid", uuid),
		sql.Named("p_start_time", startText),
		sql.Named("p_end_time", endText),
	).Scan(&runtimeSec)

	if err != nil {
		return 0, err
	}

	if !runtimeSec.Valid || runtimeSec.Int64 < 0 {
		return 0, nil
	}

	return runtimeSec.Int64, nil
}
