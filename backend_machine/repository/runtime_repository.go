package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

func (r *Repository) GetRuntimeSec(ctx context.Context, uuid string, macState string, start, end time.Time) (int64, error) {
	uuid = strings.TrimSpace(uuid)

	if uuid == "" {
		return 0, nil
	}

	// macState sengaja tidak dipakai dulu.
	// Alasannya: ada kasus macState = 0, tapi runtime aktif tetap harus dihitung
	// dari ShutTime terakhir sampai waktu aplikasi/server.
	_ = macState

	query := `
DECLARE @server_now DATETIME2 = SYSDATETIME();

-- Kalau tanggal yang diminta adalah hari ini, hitung sampai waktu server sekarang.
-- Kalau tanggal lama, batasnya tetap @end_time supaya tidak melewati tanggal tersebut.
DECLARE @waktu_aplikasi DATETIME2 =
    CASE
        WHEN @server_now < @end_time THEN @server_now
        ELSE @end_time
    END;

WITH runtime_data AS (
    SELECT
        ISNULL(SUM(
            CASE
                WHEN ISNULL(TRY_CONVERT(BIGINT, [RunTime]), 0) >= 60
                    THEN TRY_CONVERT(BIGINT, [RunTime])
                ELSE 0
            END
        ), 0) AS runtime_selesai,

        MAX(TRY_CONVERT(DATETIME2, [ShutTime])) AS waktu_terakhir
    FROM [sewingiot].[dbo].[Record_RunTime]
    WHERE LOWER(LTRIM(RTRIM([UUID]))) = LOWER(LTRIM(RTRIM(@uuid)))
      AND TRY_CONVERT(DATETIME2, [StartTime]) >= @start_time
      AND TRY_CONVERT(DATETIME2, [StartTime]) <  @end_time
)
SELECT
    CAST(
        ISNULL(runtime_selesai, 0)
        +
        CASE
            WHEN waktu_terakhir IS NOT NULL
             AND waktu_terakhir < @waktu_aplikasi
                THEN DATEDIFF_BIG(SECOND, waktu_terakhir, @waktu_aplikasi)
            ELSE 0
        END
    AS BIGINT) AS total_runtime_detik
FROM runtime_data;
`

	var runtimeSec sql.NullInt64

	err := r.DB.QueryRowContext(
		ctx,
		query,
		sql.Named("uuid", uuid),
		sql.Named("start_time", start),
		sql.Named("end_time", end),
	).Scan(&runtimeSec)

	if err != nil {
		return 0, err
	}

	if !runtimeSec.Valid || runtimeSec.Int64 < 0 {
		return 0, nil
	}

	return runtimeSec.Int64, nil
}
