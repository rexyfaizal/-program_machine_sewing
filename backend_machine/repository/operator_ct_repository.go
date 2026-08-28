package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"backend_machine/models"
)

func normalizeOperatorCtLine(value string) string {
	text := strings.TrimSpace(value)
	text = strings.ReplaceAll(text, "\u00a0", " ")
	if idx := strings.Index(text, " - "); idx > 0 {
		text = strings.TrimSpace(text[:idx])
	}
	return strings.Join(strings.Fields(text), " ")
}

func normalizeOperatorCtUUID(value string) string {
	return strings.TrimSpace(value)
}

// EnsureOperatorCtSchema membuat tabel master CT operator (upload IE).
func (r *Repository) EnsureOperatorCtSchema(ctx context.Context) error {
	query := `
IF OBJECT_ID(N'dbo.operator_ct_master', N'U') IS NULL
BEGIN
	CREATE TABLE dbo.operator_ct_master (
		id BIGINT IDENTITY(1,1) NOT NULL PRIMARY KEY,
		uuid NVARCHAR(100) NOT NULL,
		line_name NVARCHAR(255) NOT NULL,
		area NVARCHAR(50) NULL,
		ct_sum DECIMAL(12, 2) NOT NULL,
		ct_std DECIMAL(12, 2) NULL,
		ct_value DECIMAL(12, 2) NULL,
		uploaded_at DATETIME2 NOT NULL CONSTRAINT DF_operator_ct_master_uploaded_at DEFAULT SYSDATETIME()
	);

	CREATE UNIQUE INDEX UX_operator_ct_master_uuid_line
		ON dbo.operator_ct_master (uuid, line_name);
END
`
	_, err := r.DB.ExecContext(ctx, query)
	return err
}

func (r *Repository) GetOperatorCtMaster(ctx context.Context) ([]models.OperatorCtRecord, error) {
	query := `
SELECT
	id,
	uuid,
	line_name,
	ISNULL(area, '') AS area,
	ct_sum,
	ISNULL(ct_std, 0) AS ct_std,
	ISNULL(ct_value, 0) AS ct_value,
	CONVERT(VARCHAR(19), uploaded_at, 120) AS uploaded_at
FROM dbo.operator_ct_master
ORDER BY line_name, uuid;
`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.OperatorCtRecord, 0)
	for rows.Next() {
		var item models.OperatorCtRecord
		if err := rows.Scan(
			&item.ID,
			&item.UUID,
			&item.Line,
			&item.Area,
			&item.CtSum,
			&item.CtStd,
			&item.CtValue,
			&item.UploadedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	return out, rows.Err()
}

func (r *Repository) ImportOperatorCt(
	ctx context.Context,
	input models.OperatorCtImportRequest,
) (models.OperatorCtImportResponse, error) {
	response := models.OperatorCtImportResponse{
		Status:  "ok",
		Message: "Import CT berhasil",
		Total:   len(input.Rows),
	}

	if len(input.Rows) == 0 {
		return response, nil
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return models.OperatorCtImportResponse{}, err
	}
	defer tx.Rollback()

	upsertQuery := `
MERGE dbo.operator_ct_master WITH (HOLDLOCK) AS target
USING (
	SELECT
		@uuid AS uuid,
		@line_name AS line_name,
		@area AS area,
		@ct_sum AS ct_sum,
		@ct_std AS ct_std,
		@ct_value AS ct_value
) AS source
ON LOWER(LTRIM(RTRIM(target.uuid))) = LOWER(LTRIM(RTRIM(source.uuid)))
 AND LOWER(LTRIM(RTRIM(target.line_name))) = LOWER(LTRIM(RTRIM(source.line_name)))
WHEN MATCHED THEN
	UPDATE SET
		area = source.area,
		ct_sum = source.ct_sum,
		ct_std = source.ct_std,
		ct_value = source.ct_value,
		uploaded_at = SYSDATETIME()
WHEN NOT MATCHED THEN
	INSERT (uuid, line_name, area, ct_sum, ct_std, ct_value, uploaded_at)
	VALUES (
		source.uuid,
		source.line_name,
		source.area,
		source.ct_sum,
		source.ct_std,
		source.ct_value,
		SYSDATETIME()
	);
`

	seen := make(map[string]bool)

	for i, row := range input.Rows {
		uuid := normalizeOperatorCtUUID(row.UUID)
		line := normalizeOperatorCtLine(row.Line)
		area := strings.TrimSpace(row.Area)

		if uuid == "" || line == "" {
			response.Skipped++
			continue
		}

		key := strings.ToLower(uuid) + "|" + strings.ToLower(line)
		if seen[key] {
			response.Skipped++
			continue
		}
		seen[key] = true

		_, err := tx.ExecContext(
			ctx,
			upsertQuery,
			sql.Named("uuid", uuid),
			sql.Named("line_name", line),
			sql.Named("area", area),
			sql.Named("ct_sum", row.CtSum),
			sql.Named("ct_std", row.CtStd),
			sql.Named("ct_value", row.CtValue),
		)
		if err != nil {
			return models.OperatorCtImportResponse{}, fmt.Errorf(
				"baris ke-%d gagal import: %w",
				i+1,
				err,
			)
		}

		response.Upserted++
	}

	if err := tx.Commit(); err != nil {
		return models.OperatorCtImportResponse{}, err
	}

	return response, nil
}
