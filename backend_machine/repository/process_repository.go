package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"backend_machine/models"
	"backend_machine/utils"
)

func (r *Repository) GetProcessEvents(ctx context.Context, tableName string, start, end time.Time) ([]models.ProcessCycle, []models.ProcessGroup, []models.HourBar, error) {
	if !utils.SafeTableName(tableName) {
		return nil, nil, nil, fmt.Errorf("nama tabel tidak aman: %s", tableName)
	}

	query := fmt.Sprintf(`
SELECT
    ISNULL(CONVERT(nvarchar(500), FileName), '-') AS FileName,
    ISNULL(TRY_CONVERT(bigint, ProcCounts), 0) AS ProcCountsInt,
    ISNULL(TRY_CONVERT(bigint, ProcTime), 0) AS ProcTimeBig,
    ISNULL(TRY_CONVERT(bigint, EndStitch), 0) AS EndStitchInt,
    ISNULL(TRY_CONVERT(bigint, FileStitches), 0) AS FileStitchesInt,
    ISNULL(TRY_CONVERT(bigint, NodeDistance), 0) AS NodeDistanceInt,
    TRY_CONVERT(datetime2, StartTime) AS StartDT
FROM dbo.[%s]
WHERE TRY_CONVERT(datetime2, StartTime) >= @start_time
  AND TRY_CONVERT(datetime2, StartTime) < @end_time
ORDER BY TRY_CONVERT(datetime2, StartTime) ASC;
`, tableName)

	rows, err := r.DB.QueryContext(ctx, query,
		sql.Named("start_time", start),
		sql.Named("end_time", end),
	)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	type groupAcc struct {
		FileName     string
		Output       int64
		Cycles       int64
		Complete     int64
		Incomplete   int64
		ProcSec      int64
		MaxCycle     int64
		FirstProcess time.Time
		LastProcess  time.Time
	}

	type hourAcc struct {
		Hour    string
		ProcSec int64
		Output  int64
		Cycles  int64
	}

	var events []models.ProcessCycle

	groups := map[string]*groupAcc{}
	hours := map[string]*hourAcc{}
	no := 0

	for rows.Next() {
		var fileName string
		var procCounts int64
		var procSec int64
		var endStitch int64
		var fileStitches int64
		var nodeDistance int64
		var startDT sql.NullTime

		if err := rows.Scan(&fileName, &procCounts, &procSec, &endStitch, &fileStitches, &nodeDistance, &startDT); err != nil {
			return nil, nil, nil, err
		}

		if !startDT.Valid {
			continue
		}

		fileName = utils.CleanDisplayText(fileName)
		if fileName == "" {
			fileName = "-"
		}

		no++

		status := "OK"
		if procCounts == 0 || (fileStitches > 0 && endStitch < fileStitches) {
			status = "ABNORMAL"
		}

		spm := 0.0
		if procSec > 0 && endStitch > 0 {
			spm = utils.Round2(float64(endStitch) / float64(procSec) * 60)
		}

		endDT := startDT.Time.Add(time.Duration(procSec) * time.Second)

		event := models.ProcessCycle{
			No:             no,
			FileName:       fileName,
			StartTime:      startDT.Time.Format("2006-01-02 15:04:05"),
			EndTime:        endDT.Format("2006-01-02 15:04:05"),
			ProcSec:        procSec,
			ProcTime:       utils.FormatDuration(procSec),
			ProcCounts:     procCounts,
			EndStitch:      endStitch,
			FileStitches:   fileStitches,
			NodeDistance:   nodeDistance,
			Status:         status,
			AbnormalReason: utils.AbnormalReason(procCounts, endStitch, fileStitches),
			SPM:            spm,
		}

		events = append(events, event)

		g := groups[fileName]
		if g == nil {
			g = &groupAcc{
				FileName:     fileName,
				FirstProcess: startDT.Time,
				LastProcess:  startDT.Time,
			}
			groups[fileName] = g
		}

		g.Output += procCounts
		g.Cycles++
		g.ProcSec += procSec

		if procSec > g.MaxCycle {
			g.MaxCycle = procSec
		}

		if status == "OK" {
			g.Complete++
		} else {
			g.Incomplete++
		}

		if startDT.Time.Before(g.FirstProcess) {
			g.FirstProcess = startDT.Time
		}

		if startDT.Time.After(g.LastProcess) {
			g.LastProcess = startDT.Time
		}

		hourKey := startDT.Time.Format("15:00")

		h := hours[hourKey]
		if h == nil {
			h = &hourAcc{
				Hour: hourKey,
			}
			hours[hourKey] = h
		}

		h.ProcSec += procSec
		h.Output += procCounts
		h.Cycles++
	}

	if err := rows.Err(); err != nil {
		return nil, nil, nil, err
	}

	var groupList []models.ProcessGroup

	for _, g := range groups {
		avg := 0.0
		if g.Cycles > 0 {
			avg = utils.Round2(float64(g.ProcSec) / float64(g.Cycles))
		}

		groupList = append(groupList, models.ProcessGroup{
			FileName:     g.FileName,
			Output:       g.Output,
			Cycles:       g.Cycles,
			Complete:     g.Complete,
			Incomplete:   g.Incomplete,
			ProcSec:      g.ProcSec,
			ProcTime:     utils.FormatDuration(g.ProcSec),
			AvgCycle:     avg,
			MaxCycle:     g.MaxCycle,
			FirstProcess: g.FirstProcess.Format("2006-01-02 15:04:05"),
			LastProcess:  g.LastProcess.Format("2006-01-02 15:04:05"),
		})
	}

	sort.Slice(groupList, func(i, j int) bool {
		if groupList[i].ProcSec == groupList[j].ProcSec {
			return groupList[i].FileName < groupList[j].FileName
		}

		return groupList[i].ProcSec > groupList[j].ProcSec
	})

	var hourList []models.HourBar

	for _, h := range hours {
		hourList = append(hourList, models.HourBar{
			Hour:     h.Hour,
			ProcSec:  h.ProcSec,
			ProcTime: utils.FormatDuration(h.ProcSec),
			Output:   h.Output,
			Cycles:   h.Cycles,
		})
	}

	sort.Slice(hourList, func(i, j int) bool {
		return hourList[i].Hour < hourList[j].Hour
	})

	return events, groupList, hourList, nil
}
