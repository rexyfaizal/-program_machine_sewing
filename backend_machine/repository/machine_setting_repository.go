package repository

import (
	"context"
	"database/sql"
	"strings"

	"backend_machine/models"
)

func (r *Repository) GetMachineSettings(ctx context.Context) (map[string]models.MachineSetting, error) {
	query := `
		SELECT
			uuid,
			ISNULL(custom_name, '') AS custom_name,
			ISNULL(location, '') AS location,
			ISNULL(pic, '') AS pic,
			ISNULL(spv, '') AS spv,
			CONVERT(VARCHAR(19), updated_at, 120) AS updated_at
		FROM dbo.machine_setting_manual
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]models.MachineSetting)

	for rows.Next() {
		var item models.MachineSetting

		err := rows.Scan(
			&item.UUID,
			&item.CustomName,
			&item.Location,
			&item.Pic,
			&item.Spv,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		key := strings.ToLower(strings.TrimSpace(item.UUID))
		result[key] = item
	}

	return result, rows.Err()
}

func (r *Repository) UpsertMachineSetting(ctx context.Context, input models.MachineSetting) error {
	query := `
		MERGE dbo.machine_setting_manual AS target
		USING (
			SELECT
				@uuid AS uuid,
				@custom_name AS custom_name,
				@location AS location,
				@pic AS pic,
				@spv AS spv
		) AS source
		ON LOWER(LTRIM(RTRIM(target.uuid))) = LOWER(LTRIM(RTRIM(source.uuid)))
		WHEN MATCHED THEN
			UPDATE SET
				custom_name = source.custom_name,
				location = source.location,
				pic = source.pic,
				spv = source.spv,
				updated_at = SYSDATETIME()
		WHEN NOT MATCHED THEN
			INSERT (uuid, custom_name, location, pic, spv, updated_at)
			VALUES (source.uuid, source.custom_name, source.location, source.pic, source.spv, SYSDATETIME());
	`

	_, err := r.DB.ExecContext(
		ctx,
		query,
		sql.Named("uuid", strings.TrimSpace(input.UUID)),
		sql.Named("custom_name", strings.TrimSpace(input.CustomName)),
		sql.Named("location", strings.TrimSpace(input.Location)),
		sql.Named("pic", strings.TrimSpace(input.Pic)),
		sql.Named("spv", strings.TrimSpace(input.Spv)),
	)

	return err
}

func (r *Repository) DeleteMachineSetting(ctx context.Context, uuid string) error {
	query := `
		DELETE FROM dbo.machine_setting_manual
		WHERE LOWER(LTRIM(RTRIM(uuid))) = LOWER(LTRIM(RTRIM(@uuid)))
	`

	_, err := r.DB.ExecContext(
		ctx,
		query,
		sql.Named("uuid", strings.TrimSpace(uuid)),
	)

	return err
}
