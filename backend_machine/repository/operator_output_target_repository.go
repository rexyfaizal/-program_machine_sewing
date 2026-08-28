package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"backend_machine/models"
)

// EnsureOperatorOutputTargetSchema membuat tabel histori output target harian (upload IE).
func (r *Repository) EnsureOperatorOutputTargetSchema(ctx context.Context) error {
	query := `
IF OBJECT_ID(N'dbo.operator_output_target', N'U') IS NULL
BEGIN
	CREATE TABLE dbo.operator_output_target (
		id BIGINT IDENTITY(1,1) NOT NULL PRIMARY KEY,
		work_date DATE NOT NULL,
		uuid NVARCHAR(100) NOT NULL,
		line_name NVARCHAR(255) NOT NULL,
		area NVARCHAR(50) NULL,
		output_target INT NOT NULL CONSTRAINT DF_operator_output_target_value DEFAULT (0),
		uploaded_at DATETIME2 NOT NULL CONSTRAINT DF_operator_output_target_uploaded_at DEFAULT SYSDATETIME()
	);

	CREATE UNIQUE INDEX UX_operator_output_target_date_uuid_line
		ON dbo.operator_output_target (work_date, uuid, line_name);

	CREATE INDEX IX_operator_output_target_work_date
		ON dbo.operator_output_target (work_date);

	CREATE INDEX IX_operator_output_target_uuid_line
		ON dbo.operator_output_target (uuid, line_name);
END
`
	_, err := r.DB.ExecContext(ctx, query)
	return err
}

func normalizeOperatorOutputTargetDate(value string) (string, error) {
	text := strings.TrimSpace(value)
	if text == "" {
		return "", fmt.Errorf("tanggal kosong")
	}

	layouts := []string{
		"2006-01-02",
		"02/01/2006",
		"01/02/2006",
		"2006/01/02",
		"02-01-2006",
		"01-02-2006",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, text); err == nil {
			return parsed.Format("2006-01-02"), nil
		}
	}

	return "", fmt.Errorf("format tanggal tidak valid: %s", text)
}

func (r *Repository) GetOperatorOutputTargetByDate(
	ctx context.Context,
	workDate string,
) ([]models.OperatorOutputTargetRecord, error) {
	date, err := normalizeOperatorOutputTargetDate(workDate)
	if err != nil {
		return nil, err
	}

	query := `
SELECT
	id,
	CONVERT(VARCHAR(10), work_date, 23) AS work_date,
	uuid,
	line_name,
	ISNULL(area, '') AS area,
	output_target,
	CONVERT(VARCHAR(19), uploaded_at, 120) AS uploaded_at
FROM dbo.operator_output_target
WHERE work_date = @work_date
ORDER BY line_name, uuid;
`

	rows, err := r.DB.QueryContext(ctx, query, sql.Named("work_date", date))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.OperatorOutputTargetRecord, 0)
	for rows.Next() {
		var item models.OperatorOutputTargetRecord
		if err := rows.Scan(
			&item.ID,
			&item.WorkDate,
			&item.UUID,
			&item.Line,
			&item.Area,
			&item.OutputTarget,
			&item.UploadedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	return out, rows.Err()
}

func (r *Repository) ImportOperatorOutputTarget(
	ctx context.Context,
	input models.OperatorOutputTargetImportRequest,
) (models.OperatorOutputTargetImportResponse, error) {
	response := models.OperatorOutputTargetImportResponse{
		Status:  "ok",
		Message: "Import output target berhasil",
		Total:   len(input.Rows),
	}

	if len(input.Rows) == 0 {
		return response, nil
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return models.OperatorOutputTargetImportResponse{}, err
	}
	defer tx.Rollback()

	upsertQuery := `
MERGE dbo.operator_output_target WITH (HOLDLOCK) AS target
USING (
	SELECT
		@work_date AS work_date,
		@uuid AS uuid,
		@line_name AS line_name,
		@area AS area,
		@output_target AS output_target
) AS source
ON target.work_date = source.work_date
 AND LOWER(LTRIM(RTRIM(target.uuid))) = LOWER(LTRIM(RTRIM(source.uuid)))
 AND LOWER(LTRIM(RTRIM(target.line_name))) = LOWER(LTRIM(RTRIM(source.line_name)))
WHEN MATCHED THEN
	UPDATE SET
		area = source.area,
		output_target = source.output_target,
		uploaded_at = SYSDATETIME()
WHEN NOT MATCHED THEN
	INSERT (work_date, uuid, line_name, area, output_target, uploaded_at)
	VALUES (
		source.work_date,
		source.uuid,
		source.line_name,
		source.area,
		source.output_target,
		SYSDATETIME()
	);
`

	seen := make(map[string]bool)

	for i, row := range input.Rows {
		workDate, err := normalizeOperatorOutputTargetDate(row.WorkDate)
		if err != nil {
			response.Skipped++
			continue
		}

		uuid := normalizeOperatorCtUUID(row.UUID)
		line := normalizeOperatorCtLine(row.Line)
		area := strings.TrimSpace(row.Area)

		if uuid == "" || line == "" {
			response.Skipped++
			continue
		}

		outputTarget := row.OutputTarget
		if outputTarget < 0 {
			outputTarget = 0
		}

		key := strings.ToLower(workDate) + "|" + strings.ToLower(uuid) + "|" + strings.ToLower(line)
		if seen[key] {
			response.Skipped++
			continue
		}
		seen[key] = true

		_, err = tx.ExecContext(
			ctx,
			upsertQuery,
			sql.Named("work_date", workDate),
			sql.Named("uuid", uuid),
			sql.Named("line_name", line),
			sql.Named("area", area),
			sql.Named("output_target", outputTarget),
		)
		if err != nil {
			return models.OperatorOutputTargetImportResponse{}, fmt.Errorf(
				"baris ke-%d gagal import: %w",
				i+1,
				err,
			)
		}

		response.Upserted++
	}

	if err := tx.Commit(); err != nil {
		return models.OperatorOutputTargetImportResponse{}, err
	}

	return response, nil
}
