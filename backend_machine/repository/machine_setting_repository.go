package repository

import (
	"context"
	"database/sql"
	"log"
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

// EnsureMachineSettingSchema melebarkan kolom teks agar layout line (JSON)
// dan nama mesin panjang tidak kena truncation.
func (r *Repository) EnsureMachineSettingSchema(ctx context.Context) error {
	query := `
IF OBJECT_ID(N'dbo.machine_setting_manual', N'U') IS NULL
BEGIN
	CREATE TABLE dbo.machine_setting_manual (
		uuid NVARCHAR(100) NOT NULL PRIMARY KEY,
		custom_name NVARCHAR(MAX) NULL,
		location NVARCHAR(255) NULL,
		pic NVARCHAR(255) NULL,
		spv NVARCHAR(MAX) NULL,
		updated_at DATETIME2 NOT NULL CONSTRAINT DF_machine_setting_manual_updated_at DEFAULT SYSDATETIME()
	);
END
ELSE
BEGIN
	IF EXISTS (
		SELECT 1
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = 'dbo'
		  AND TABLE_NAME = 'machine_setting_manual'
		  AND COLUMN_NAME = 'custom_name'
		  AND DATA_TYPE IN ('varchar', 'nvarchar', 'char', 'nchar')
		  AND CHARACTER_MAXIMUM_LENGTH > 0
		  AND CHARACTER_MAXIMUM_LENGTH < 2000
	)
	BEGIN
		ALTER TABLE dbo.machine_setting_manual ALTER COLUMN custom_name NVARCHAR(MAX) NULL;
	END

	IF EXISTS (
		SELECT 1
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = 'dbo'
		  AND TABLE_NAME = 'machine_setting_manual'
		  AND COLUMN_NAME = 'spv'
		  AND DATA_TYPE IN ('varchar', 'nvarchar', 'char', 'nchar')
		  AND CHARACTER_MAXIMUM_LENGTH > 0
		  AND CHARACTER_MAXIMUM_LENGTH < 2000
	)
	BEGIN
		ALTER TABLE dbo.machine_setting_manual ALTER COLUMN spv NVARCHAR(MAX) NULL;
	END

	IF EXISTS (
		SELECT 1
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = 'dbo'
		  AND TABLE_NAME = 'machine_setting_manual'
		  AND COLUMN_NAME = 'location'
		  AND DATA_TYPE IN ('varchar', 'nvarchar', 'char', 'nchar')
		  AND CHARACTER_MAXIMUM_LENGTH > 0
		  AND CHARACTER_MAXIMUM_LENGTH <= 100
	)
	BEGIN
		ALTER TABLE dbo.machine_setting_manual ALTER COLUMN location NVARCHAR(255) NULL;
	END
END
`

	_, err := r.DB.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	log.Println("Schema machine_setting_manual siap (custom_name/spv long text).")
	return nil
}
