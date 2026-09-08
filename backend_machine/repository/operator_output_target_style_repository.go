package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"backend_machine/models"
)

// EnsureOperatorOutputTargetStyleSchema membuat tabel output target harian Style+Proses (GM3).
func (r *Repository) EnsureOperatorOutputTargetStyleSchema(ctx context.Context) error {
	query := `
IF OBJECT_ID(N'dbo.operator_output_target_style', N'U') IS NULL
BEGIN
	CREATE TABLE dbo.operator_output_target_style (
		id BIGINT IDENTITY(1,1) NOT NULL PRIMARY KEY,
		work_date DATE NOT NULL,
		style_name NVARCHAR(100) NOT NULL,
		process_name NVARCHAR(255) NOT NULL,
		output_target INT NOT NULL CONSTRAINT DF_operator_output_target_style_value DEFAULT (0),
		uploaded_at DATETIME2 NOT NULL CONSTRAINT DF_operator_output_target_style_uploaded_at DEFAULT SYSDATETIME()
	);

	CREATE UNIQUE INDEX UX_operator_output_target_style_date_style_process
		ON dbo.operator_output_target_style (work_date, style_name, process_name);

	CREATE INDEX IX_operator_output_target_style_work_date
		ON dbo.operator_output_target_style (work_date);
END
`
	_, err := r.DB.ExecContext(ctx, query)
	return err
}

func (r *Repository) GetOperatorOutputTargetStyleByDate(
	ctx context.Context,
	workDate string,
) ([]models.OperatorOutputTargetStyleRecord, error) {
	date, err := normalizeOperatorOutputTargetDate(workDate)
	if err != nil {
		return nil, err
	}

	query := `
SELECT
	id,
	CONVERT(VARCHAR(10), work_date, 23) AS work_date,
	style_name,
	process_name,
	output_target,
	CONVERT(VARCHAR(19), uploaded_at, 120) AS uploaded_at
FROM dbo.operator_output_target_style
WHERE work_date = @work_date
ORDER BY style_name, process_name;
`

	rows, err := r.DB.QueryContext(ctx, query, sql.Named("work_date", date))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.OperatorOutputTargetStyleRecord, 0)
	for rows.Next() {
		var item models.OperatorOutputTargetStyleRecord
		if err := rows.Scan(
			&item.ID,
			&item.WorkDate,
			&item.StyleName,
			&item.ProcessName,
			&item.OutputTarget,
			&item.UploadedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	return out, rows.Err()
}

func (r *Repository) ImportOperatorOutputTargetStyle(
	ctx context.Context,
	input models.OperatorOutputTargetStyleImportRequest,
) (models.OperatorOutputTargetStyleImportResponse, error) {
	response := models.OperatorOutputTargetStyleImportResponse{
		Status:  "ok",
		Message: "Import output target Style-Proses berhasil",
		Total:   len(input.Rows),
	}

	if len(input.Rows) == 0 {
		return response, nil
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return models.OperatorOutputTargetStyleImportResponse{}, err
	}
	defer tx.Rollback()

	upsertQuery := `
MERGE dbo.operator_output_target_style WITH (HOLDLOCK) AS target
USING (
	SELECT
		@work_date AS work_date,
		@style_name AS style_name,
		@process_name AS process_name,
		@output_target AS output_target
) AS source
ON target.work_date = source.work_date
 AND LOWER(LTRIM(RTRIM(target.style_name))) = LOWER(LTRIM(RTRIM(source.style_name)))
 AND LOWER(LTRIM(RTRIM(target.process_name))) = LOWER(LTRIM(RTRIM(source.process_name)))
WHEN MATCHED THEN
	UPDATE SET
		output_target = source.output_target,
		uploaded_at = SYSDATETIME()
WHEN NOT MATCHED THEN
	INSERT (work_date, style_name, process_name, output_target, uploaded_at)
	VALUES (
		source.work_date,
		source.style_name,
		source.process_name,
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

		style := normalizeOperatorCtStylePart(row.StyleName)
		process := normalizeOperatorCtStylePart(row.ProcessName)

		if style == "" || process == "" {
			response.Skipped++
			continue
		}

		outputTarget := row.OutputTarget
		if outputTarget < 0 {
			outputTarget = 0
		}

		key := strings.ToLower(workDate) + "|" + strings.ToLower(style) + "|" + strings.ToLower(process)
		if seen[key] {
			response.Skipped++
			continue
		}
		seen[key] = true

		_, err = tx.ExecContext(
			ctx,
			upsertQuery,
			sql.Named("work_date", workDate),
			sql.Named("style_name", style),
			sql.Named("process_name", process),
			sql.Named("output_target", outputTarget),
		)
		if err != nil {
			return models.OperatorOutputTargetStyleImportResponse{}, fmt.Errorf(
				"baris ke-%d gagal import: %w",
				i+1,
				err,
			)
		}

		response.Upserted++
	}

	if err := tx.Commit(); err != nil {
		return models.OperatorOutputTargetStyleImportResponse{}, err
	}

	return response, nil
}
