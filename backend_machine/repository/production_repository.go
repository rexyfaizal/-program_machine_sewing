package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"backend_machine/models"
	"backend_machine/utils"
)

func (r *Repository) GetProductionStats(ctx context.Context, tableName string, start, end time.Time) (models.ProductionStats, error) {
	if !utils.SafeTableName(tableName) {
		return models.ProductionStats{}, fmt.Errorf("nama tabel tidak aman: %s", tableName)
	}

	query := fmt.Sprintf(`
WITH src AS (
    SELECT
        ISNULL(CONVERT(nvarchar(500), FileName), '-') AS FileName,
        TRY_CONVERT(int, ProcCounts) AS ProcCountsInt,
        TRY_CONVERT(bigint, ProcTime) AS ProcTimeBig,
        TRY_CONVERT(int, EndStitch) AS EndStitchInt,
        TRY_CONVERT(int, FileStitches) AS FileStitchesInt,
        TRY_CONVERT(datetime2, StartTime) AS StartDT
    FROM dbo.[%s]
    WHERE TRY_CONVERT(datetime2, StartTime) >= TRY_CONVERT(datetime2, @start_time)
      AND TRY_CONVERT(datetime2, StartTime) < TRY_CONVERT(datetime2, @end_time)
), agg AS (
    SELECT
        ISNULL(SUM(ISNULL(ProcTimeBig, 0)), 0) AS ProcSec,
        ISNULL(SUM(CASE
            WHEN ProcCountsInt = 1 AND EndStitchInt = FileStitchesInt
            THEN ISNULL(ProcTimeBig, 0) ELSE 0 END), 0) AS OkProcSec,
        ISNULL(SUM(CASE WHEN ProcCountsInt IS NULL THEN 0 ELSE ProcCountsInt END), 0) AS Output,
        COUNT(1) AS Cycles,
        ISNULL(SUM(CASE
            WHEN ProcCountsInt = 1 AND EndStitchInt = FileStitchesInt THEN 1 ELSE 0 END), 0) AS Complete,
        ISNULL(SUM(CASE
            WHEN ISNULL(ProcCountsInt, 0) = 0 OR ISNULL(EndStitchInt, 0) < ISNULL(FileStitchesInt, 0)
            THEN 1 ELSE 0 END), 0) AS Incomplete,
        ISNULL(AVG(CASE WHEN ProcTimeBig IS NOT NULL THEN CONVERT(float, ProcTimeBig) END), 0) AS AvgCycle,
        ISNULL(MIN(CASE WHEN ProcTimeBig IS NOT NULL AND ProcTimeBig > 0 THEN CONVERT(float, ProcTimeBig) END), 0) AS MinCycle,
        ISNULL(MAX(ISNULL(ProcTimeBig, 0)), 0) AS MaxCycle,
        ISNULL(SUM(CASE WHEN ISNULL(ProcTimeBig, 0) >= 300 THEN 1 ELSE 0 END), 0) AS SlowCycles,
        COUNT(DISTINCT NULLIF(FileName, '')) AS UniqueFiles,
        MIN(StartDT) AS FirstProcess,
        MAX(StartDT) AS LastProcess
    FROM src
)
SELECT
    agg.ProcSec,
    agg.OkProcSec,
    agg.Output,
    agg.Cycles,
    agg.Complete,
    agg.Incomplete,
    agg.AvgCycle,
    agg.MinCycle,
    agg.MaxCycle,
    agg.SlowCycles,
    agg.UniqueFiles,
    ISNULL((
        SELECT TOP (1) s.FileName
        FROM src s
        WHERE s.FileName IS NOT NULL AND s.FileName <> ''
        GROUP BY s.FileName
        ORDER BY COUNT(1) DESC, s.FileName ASC
    ), '-') AS TopFile,
    agg.FirstProcess,
    agg.LastProcess
FROM agg;
`, tableName)

	var ps models.ProductionStats
	var first sql.NullTime
	var last sql.NullTime

	err := r.DB.QueryRowContext(ctx, query,
		sql.Named("start_time", formatNaiveDateTime(start)),
		sql.Named("end_time", formatNaiveDateTime(end)),
	).Scan(
		&ps.ProcSec,
		&ps.OkProcSec,
		&ps.Output,
		&ps.Cycles,
		&ps.Complete,
		&ps.Incomplete,
		&ps.AvgCycle,
		&ps.MinCycle,
		&ps.MaxCycle,
		&ps.SlowCycles,
		&ps.UniqueFiles,
		&ps.TopFile,
		&first,
		&last,
	)
	if err != nil {
		return models.ProductionStats{}, err
	}

	ps.TopFile = utils.CleanDisplayText(ps.TopFile)
	if ps.TopFile == "" {
		ps.TopFile = "-"
	}

	if first.Valid {
		ps.FirstProcess = first.Time.Format("2006-01-02 15:04:05")
	}

	if last.Valid {
		ps.LastProcess = last.Time.Format("2006-01-02 15:04:05")
	}

	return ps, nil
}
