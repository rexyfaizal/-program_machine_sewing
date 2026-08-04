package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"backend_machine/models"
	"backend_machine/utils"
)

func (r *Repository) EnsureLineShiftConfigSchema(ctx context.Context) error {
	query := `
IF OBJECT_ID(N'dbo.line_shift_config', N'U') IS NULL
BEGIN
	CREATE TABLE dbo.line_shift_config (
		factory NVARCHAR(50) NOT NULL,
		line_name NVARCHAR(255) NOT NULL,
		enabled BIT NOT NULL CONSTRAINT DF_line_shift_config_enabled DEFAULT 0,
		schedule_json NVARCHAR(MAX) NULL,
		updated_at DATETIME2 NOT NULL CONSTRAINT DF_line_shift_config_updated_at DEFAULT SYSDATETIME(),
		CONSTRAINT PK_line_shift_config PRIMARY KEY (factory, line_name)
	);
END
`

	_, err := r.DB.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	log.Println("Schema line_shift_config siap.")
	return nil
}

func (r *Repository) GetLineShiftConfigs(ctx context.Context, factory string) ([]models.LineShiftConfig, error) {
	factory = strings.ToUpper(strings.TrimSpace(factory))

	query := `
		SELECT
			factory,
			line_name,
			enabled,
			ISNULL(schedule_json, '[]') AS schedule_json,
			CONVERT(VARCHAR(19), updated_at, 120) AS updated_at
		FROM dbo.line_shift_config
	`

	args := []any{}
	if factory != "" {
		query += ` WHERE UPPER(LTRIM(RTRIM(factory))) = @factory`
		args = append(args, sql.Named("factory", factory))
	}

	query += ` ORDER BY factory, line_name`

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.LineShiftConfig, 0)

	for rows.Next() {
		var item models.LineShiftConfig
		var enabled bool

		err := rows.Scan(
			&item.Factory,
			&item.LineName,
			&enabled,
			&item.ScheduleJSON,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		item.Enabled = enabled
		item.Factory = strings.ToUpper(strings.TrimSpace(item.Factory))
		item.LineName = strings.TrimSpace(item.LineName)
		item.Schedule = utils.ParseShiftScheduleJSON(item.ScheduleJSON)
		result = append(result, item)
	}

	return result, rows.Err()
}

// GetLineShiftConfigMap key = "FACTORY||LINE_NAME" (upper).
func (r *Repository) GetLineShiftConfigMap(ctx context.Context) (map[string]models.LineShiftConfig, error) {
	list, err := r.GetLineShiftConfigs(ctx, "")
	if err != nil {
		return nil, err
	}

	result := make(map[string]models.LineShiftConfig, len(list))
	for _, item := range list {
		key := utils.LineShiftConfigKey(item.Factory, item.LineName)
		result[key] = item
	}
	return result, nil
}

func (r *Repository) UpsertLineShiftConfigs(ctx context.Context, factory string, lines []models.LineShiftConfig) error {
	factory = strings.ToUpper(strings.TrimSpace(factory))
	if factory == "" {
		return fmt.Errorf("factory kosong")
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Ganti semua config factory ini dengan payload terbaru.
	_, err = tx.ExecContext(
		ctx,
		`DELETE FROM dbo.line_shift_config WHERE UPPER(LTRIM(RTRIM(factory))) = @factory`,
		sql.Named("factory", factory),
	)
	if err != nil {
		return err
	}

	insertQuery := `
		INSERT INTO dbo.line_shift_config (factory, line_name, enabled, schedule_json, updated_at)
		VALUES (@factory, @line_name, @enabled, @schedule_json, SYSDATETIME())
	`

	for _, line := range lines {
		lineName := strings.TrimSpace(line.LineName)
		if lineName == "" {
			continue
		}

		schedule := line.Schedule
		if schedule == nil {
			schedule = []models.ShiftScheduleItem{}
		}

		normalized := make([]models.ShiftScheduleItem, 0, len(schedule))
		for _, item := range schedule {
			code := utils.NormalizeShiftCode(item.Code)
			if code == utils.ShiftALL || code == utils.ShiftCurrent {
				continue
			}
			normalized = append(normalized, models.ShiftScheduleItem{
				Code:       code,
				Start:      strings.TrimSpace(item.Start),
				End:        strings.TrimSpace(item.End),
				BreakStart: strings.TrimSpace(item.BreakStart),
				BreakEnd:   strings.TrimSpace(item.BreakEnd),
			})
		}

		raw, err := json.Marshal(normalized)
		if err != nil {
			return err
		}

		enabled := 0
		if line.Enabled {
			enabled = 1
		}

		_, err = tx.ExecContext(
			ctx,
			insertQuery,
			sql.Named("factory", factory),
			sql.Named("line_name", lineName),
			sql.Named("enabled", enabled),
			sql.Named("schedule_json", string(raw)),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
