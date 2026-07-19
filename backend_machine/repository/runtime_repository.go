package repository

import (
	"context"
	"database/sql"
	"time"
)

func (r *Repository) GetRuntimeSec(ctx context.Context, uuid string, start, end time.Time) (int64, error) {
	query := `
WITH raw AS (
    SELECT
        TRY_CONVERT(datetime2, StartTime) AS st,
        CASE
            WHEN TRY_CONVERT(datetime2, ShutTime) IS NOT NULL
             AND TRY_CONVERT(datetime2, ShutTime) > TRY_CONVERT(datetime2, StartTime)
            THEN TRY_CONVERT(datetime2, ShutTime)
            WHEN TRY_CONVERT(int, RunTime) IS NOT NULL
             AND TRY_CONVERT(datetime2, StartTime) IS NOT NULL
            THEN DATEADD(SECOND, TRY_CONVERT(int, RunTime), TRY_CONVERT(datetime2, StartTime))
            ELSE NULL
        END AS et
    FROM dbo.record_runtime
    WHERE UUID = @uuid
      AND TRY_CONVERT(datetime2, StartTime) IS NOT NULL
)
SELECT ISNULL(SUM(
    CASE
        WHEN st < @end_time AND et > @start_time THEN
            CASE
                WHEN DATEDIFF(SECOND,
                    CASE WHEN st < @start_time THEN @start_time ELSE st END,
                    CASE WHEN et > @end_time THEN @end_time ELSE et END
                ) > 0 THEN
                    DATEDIFF(SECOND,
                        CASE WHEN st < @start_time THEN @start_time ELSE st END,
                        CASE WHEN et > @end_time THEN @end_time ELSE et END
                    )
                ELSE 0
            END
        ELSE 0
    END
), 0) AS RuntimeSec
FROM raw
WHERE st IS NOT NULL AND et IS NOT NULL;
`

	var runtimeSec int64

	err := r.DB.QueryRowContext(ctx, query,
		sql.Named("uuid", uuid),
		sql.Named("start_time", start),
		sql.Named("end_time", end),
	).Scan(&runtimeSec)

	return runtimeSec, err
}
