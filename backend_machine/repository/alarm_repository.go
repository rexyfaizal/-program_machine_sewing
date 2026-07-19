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

func (r *Repository) GetAlarmStats(ctx context.Context, uuid string, start, end time.Time) (models.AlarmStats, error) {
	var alarmCount int64

	countQuery := `
SELECT COUNT(1)
FROM dbo.record_alarmlog
WHERE UUID = @uuid
  AND TRY_CONVERT(datetime2, AlarmDate) >= @start_time
  AND TRY_CONVERT(datetime2, AlarmDate) < @end_time;
`

	err := r.DB.QueryRowContext(ctx, countQuery,
		sql.Named("uuid", uuid),
		sql.Named("start_time", start),
		sql.Named("end_time", end),
	).Scan(&alarmCount)
	if err != nil {
		return models.AlarmStats{}, err
	}

	typesQuery := `
SELECT TOP (3)
    ISNULL(CONVERT(nvarchar(500), Content), '-') AS Content,
    COUNT(1) AS Total
FROM dbo.record_alarmlog
WHERE UUID = @uuid
  AND TRY_CONVERT(datetime2, AlarmDate) >= @start_time
  AND TRY_CONVERT(datetime2, AlarmDate) < @end_time
GROUP BY ISNULL(CONVERT(nvarchar(500), Content), '-')
ORDER BY COUNT(1) DESC, Content ASC;
`

	rows, err := r.DB.QueryContext(ctx, typesQuery,
		sql.Named("uuid", uuid),
		sql.Named("start_time", start),
		sql.Named("end_time", end),
	)
	if err != nil {
		return models.AlarmStats{}, err
	}
	defer rows.Close()

	var parts []string

	for rows.Next() {
		var content string
		var total int64

		if err := rows.Scan(&content, &total); err != nil {
			return models.AlarmStats{}, err
		}

		content = utils.CleanDisplayText(content)
		if content == "" {
			content = "Alarm"
		}

		parts = append(parts, fmt.Sprintf("%s (%d)", content, total))
	}

	if err := rows.Err(); err != nil {
		return models.AlarmStats{}, err
	}

	return models.AlarmStats{
		AlarmCount: alarmCount,
		AlarmTypes: strings.Join(parts, " | "),
	}, nil
}

func (r *Repository) GetAlarmGroups(ctx context.Context, uuid string, start, end time.Time) ([]models.AlarmGroup, error) {
	query := `
SELECT TOP (10)
    ISNULL(CONVERT(nvarchar(500), Content), '-') AS Content,
    COUNT(1) AS Total
FROM dbo.record_alarmlog
WHERE UUID = @uuid
  AND TRY_CONVERT(datetime2, AlarmDate) >= @start_time
  AND TRY_CONVERT(datetime2, AlarmDate) < @end_time
GROUP BY ISNULL(CONVERT(nvarchar(500), Content), '-')
ORDER BY COUNT(1) DESC, Content ASC;
`

	rows, err := r.DB.QueryContext(ctx, query,
		sql.Named("uuid", uuid),
		sql.Named("start_time", start),
		sql.Named("end_time", end),
	)
	if err != nil {
		return []models.AlarmGroup{}, err
	}
	defer rows.Close()

	var list []models.AlarmGroup

	for rows.Next() {
		var content string
		var total int64

		if err := rows.Scan(&content, &total); err != nil {
			return []models.AlarmGroup{}, err
		}

		content = utils.CleanDisplayText(content)
		if content == "" {
			content = "Alarm"
		}

		list = append(list, models.AlarmGroup{
			Content: content,
			Total:   total,
		})
	}

	return list, rows.Err()
}
