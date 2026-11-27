package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"backend_machine/models"
)

var jukiNumberPattern = regexp.MustCompile(`(?i)^JUKI-(\d+)$`)

func (r *Repository) GetMachineSettings(ctx context.Context) (map[string]models.MachineSetting, error) {
	if err := r.EnsureMachineSettingSchema(ctx); err != nil {
		return nil, err
	}

	query := `
		SELECT
			uuid,
			ISNULL(custom_name, '') AS custom_name,
			ISNULL(location, '') AS location,
			ISNULL(machine_id, '') AS machine_id,
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
			&item.MachineID,
			&item.Pic,
			&item.Spv,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		key := normalizeMachineUUIDKey(item.UUID)
		result[key] = item
	}

	return result, rows.Err()
}

func (r *Repository) UpsertMachineSetting(ctx context.Context, input models.MachineSetting) error {
	if err := r.EnsureMachineSettingSchema(ctx); err != nil {
		return err
	}

	query := `
		MERGE dbo.machine_setting_manual AS target
		USING (
			SELECT
				@uuid AS uuid,
				@custom_name AS custom_name,
				@location AS location,
				@machine_id AS machine_id,
				@pic AS pic,
				@spv AS spv
		) AS source
		ON LOWER(LTRIM(RTRIM(target.uuid))) = LOWER(LTRIM(RTRIM(source.uuid)))
		WHEN MATCHED THEN
			UPDATE SET
				custom_name = source.custom_name,
				location = source.location,
				machine_id = CASE
					WHEN LTRIM(RTRIM(source.machine_id)) = '' THEN target.machine_id
					ELSE source.machine_id
				END,
				pic = source.pic,
				spv = source.spv,
				updated_at = SYSDATETIME()
		WHEN NOT MATCHED THEN
			INSERT (uuid, custom_name, location, machine_id, pic, spv, updated_at)
			VALUES (source.uuid, source.custom_name, source.location, NULLIF(source.machine_id, ''), source.pic, source.spv, SYSDATETIME());
	`

	_, err := r.DB.ExecContext(
		ctx,
		query,
		sql.Named("uuid", strings.TrimSpace(input.UUID)),
		sql.Named("custom_name", strings.TrimSpace(input.CustomName)),
		sql.Named("location", strings.TrimSpace(input.Location)),
		sql.Named("machine_id", strings.TrimSpace(input.MachineID)),
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

func (r *Repository) listMachineInfoUUIDs(ctx context.Context) ([]string, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT DISTINCT LTRIM(RTRIM(ISNULL(uuid, ''))) AS uuid
		FROM dbo.machineinfo
		WHERE LTRIM(RTRIM(ISNULL(uuid, ''))) <> ''
		ORDER BY uuid
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]string, 0, 128)
	for rows.Next() {
		var uuid string
		if err := rows.Scan(&uuid); err != nil {
			return nil, err
		}
		uuid = strings.TrimSpace(uuid)
		if uuid != "" {
			out = append(out, uuid)
		}
	}
	return out, rows.Err()
}

func (r *Repository) upsertMachineIDOnly(ctx context.Context, uuid, machineID string) error {
	uuid = strings.TrimSpace(uuid)
	machineID = strings.TrimSpace(machineID)
	if uuid == "" || machineID == "" {
		return nil
	}

	_, err := r.DB.ExecContext(ctx, `
MERGE dbo.machine_setting_manual AS target
USING (SELECT @uuid AS uuid, @machine_id AS machine_id) AS source
ON LOWER(LTRIM(RTRIM(target.uuid))) = LOWER(LTRIM(RTRIM(source.uuid)))
WHEN MATCHED THEN
	UPDATE SET
		machine_id = source.machine_id,
		updated_at = SYSDATETIME()
WHEN NOT MATCHED THEN
	INSERT (uuid, machine_id, updated_at)
	VALUES (source.uuid, source.machine_id, SYSDATETIME());
`, sql.Named("uuid", uuid), sql.Named("machine_id", machineID))
	return err
}

func nextJukiMachineID(used map[string]bool) string {
	for n := 1; n < 100000; n++ {
		id := fmt.Sprintf("JUKI-%04d", n)
		if !used[strings.ToUpper(id)] {
			return id
		}
	}
	return fmt.Sprintf("JUKI-%d", len(used)+1)
}

// ImportMachineIDs upsert machine_id dari Excel (UUID + Machine ID).
// Satu Machine ID tidak boleh dipakai lebih dari satu UUID.
func (r *Repository) ImportMachineIDs(ctx context.Context, input models.MachineIDImportRequest) (models.MachineIDImportResponse, error) {
	if err := r.EnsureMachineSettingSchema(ctx); err != nil {
		return models.MachineIDImportResponse{}, err
	}

	settings, err := r.GetMachineSettings(ctx)
	if err != nil {
		return models.MachineIDImportResponse{}, err
	}

	ownerByMachineID := make(map[string]string)
	for _, item := range settings {
		id := strings.ToUpper(strings.TrimSpace(item.MachineID))
		uuid := strings.TrimSpace(item.UUID)
		if id == "" || uuid == "" {
			continue
		}
		ownerByMachineID[id] = uuid
	}

	resp := models.MachineIDImportResponse{
		Status: "ok",
		Errors: make([]models.MachineIDImportError, 0),
	}

	seenUUID := make(map[string]int)
	pendingOwner := make(map[string]string) // machineID upper -> uuid

	type pendingRow struct {
		uuid      string
		machineID string
		excelRow  int
	}
	pending := make([]pendingRow, 0)

	for _, row := range input.Rows {
		resp.Total++
		uuid := strings.TrimSpace(row.UUID)
		machineID := strings.TrimSpace(row.MachineID)
		excelRow := row.ExcelRow

		if uuid == "" && machineID == "" {
			resp.Skipped++
			continue
		}

		if uuid == "" || machineID == "" {
			resp.Skipped++
			resp.Errors = append(resp.Errors, models.MachineIDImportError{
				ExcelRow:  excelRow,
				UUID:      uuid,
				MachineID: machineID,
				Message:   "UUID dan Machine ID wajib diisi",
			})
			continue
		}

		uuidKey := normalizeMachineUUIDKey(uuid)
		idKey := strings.ToUpper(machineID)

		if prevRow, ok := seenUUID[uuidKey]; ok {
			resp.Skipped++
			resp.Errors = append(resp.Errors, models.MachineIDImportError{
				ExcelRow:  excelRow,
				UUID:      uuid,
				MachineID: machineID,
				Message:   fmt.Sprintf("UUID duplikat di file (baris %d)", prevRow),
			})
			continue
		}
		seenUUID[uuidKey] = excelRow

		if otherUUID, ok := pendingOwner[idKey]; ok && !strings.EqualFold(otherUUID, uuid) {
			resp.Skipped++
			resp.Errors = append(resp.Errors, models.MachineIDImportError{
				ExcelRow:  excelRow,
				UUID:      uuid,
				MachineID: machineID,
				Message:   fmt.Sprintf("Machine ID sudah dipakai UUID lain di file (%s)", otherUUID),
			})
			continue
		}

		if owner, ok := ownerByMachineID[idKey]; ok && !strings.EqualFold(owner, uuid) {
			resp.Skipped++
			resp.Errors = append(resp.Errors, models.MachineIDImportError{
				ExcelRow:  excelRow,
				UUID:      uuid,
				MachineID: machineID,
				Message:   fmt.Sprintf("Machine ID sudah dipakai UUID lain di DB (%s)", owner),
			})
			continue
		}

		pendingOwner[idKey] = uuid
		pending = append(pending, pendingRow{
			uuid:      uuid,
			machineID: machineID,
			excelRow:  excelRow,
		})
	}

	for _, row := range pending {
		if err := r.upsertMachineIDOnly(ctx, row.uuid, row.machineID); err != nil {
			resp.Skipped++
			resp.Errors = append(resp.Errors, models.MachineIDImportError{
				ExcelRow:  row.excelRow,
				UUID:      row.uuid,
				MachineID: row.machineID,
				Message:   err.Error(),
			})
			continue
		}
		resp.Upserted++
		ownerByMachineID[strings.ToUpper(row.machineID)] = row.uuid
	}

	resp.Message = fmt.Sprintf(
		"Import Machine ID selesai. %d diupdate, %d dilewati dari %d baris.",
		resp.Upserted,
		resp.Skipped,
		resp.Total,
	)
	return resp, nil
}

// EnsureMachineIDs mengisi machine_id kosong: seed mapping IE dulu, sisanya JUKI-xxxx baru.
func (r *Repository) EnsureMachineIDs(ctx context.Context) error {
	if err := r.EnsureMachineSettingSchema(ctx); err != nil {
		return err
	}

	uuids, err := r.listMachineInfoUUIDs(ctx)
	if err != nil {
		return err
	}
	settings, err := r.GetMachineSettings(ctx)
	if err != nil {
		return err
	}

	used := make(map[string]bool)
	for _, item := range settings {
		id := strings.ToUpper(strings.TrimSpace(item.MachineID))
		if id != "" {
			used[id] = true
		}
	}

	// 1) Seed mapping yang masih kosong.
	for _, uuid := range uuids {
		key := normalizeMachineUUIDKey(uuid)
		existing := strings.TrimSpace(settings[key].MachineID)
		if existing != "" {
			continue
		}
		seedID := strings.TrimSpace(machineIDSeedByUUID[key])
		if seedID == "" {
			continue
		}
		seedKey := strings.ToUpper(seedID)
		if used[seedKey] {
			continue
		}
		if err := r.upsertMachineIDOnly(ctx, uuid, seedID); err != nil {
			return err
		}
		used[seedKey] = true
		item := settings[key]
		item.UUID = uuid
		item.MachineID = seedID
		settings[key] = item
	}

	// 2) Sisanya dapat JUKI-xxxx berikutnya.
	for _, uuid := range uuids {
		key := normalizeMachineUUIDKey(uuid)
		existing := strings.TrimSpace(settings[key].MachineID)
		if existing != "" {
			continue
		}
		newID := nextJukiMachineID(used)
		if err := r.upsertMachineIDOnly(ctx, uuid, newID); err != nil {
			return err
		}
		used[strings.ToUpper(newID)] = true
		item := settings[key]
		item.UUID = uuid
		item.MachineID = newID
		settings[key] = item
	}

	return nil
}

func ParseJukiNumber(machineID string) (int, bool) {
	match := jukiNumberPattern.FindStringSubmatch(strings.TrimSpace(machineID))
	if len(match) != 2 {
		return 0, false
	}
	n, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, false
	}
	return n, true
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
		machine_id NVARCHAR(50) NULL,
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

	IF COL_LENGTH('dbo.machine_setting_manual', 'machine_id') IS NULL
	BEGIN
		ALTER TABLE dbo.machine_setting_manual ADD machine_id NVARCHAR(50) NULL;
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
