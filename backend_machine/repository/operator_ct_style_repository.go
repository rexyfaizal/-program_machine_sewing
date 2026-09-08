package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"backend_machine/models"
)

func normalizeOperatorCtStylePart(value string) string {
	text := strings.TrimSpace(value)
	text = strings.ReplaceAll(text, "\u00a0", " ")
	return strings.Join(strings.Fields(text), " ")
}

// EnsureOperatorCtStyleSchema membuat tabel master CT Style+Proses (upload IE, khusus GM3).
func (r *Repository) EnsureOperatorCtStyleSchema(ctx context.Context) error {
	query := `
IF OBJECT_ID(N'dbo.operator_ct_style', N'U') IS NULL
BEGIN
	CREATE TABLE dbo.operator_ct_style (
		id BIGINT IDENTITY(1,1) NOT NULL PRIMARY KEY,
		style_name NVARCHAR(100) NOT NULL,
		process_name NVARCHAR(255) NOT NULL,
		ct_total DECIMAL(12, 2) NOT NULL CONSTRAINT DF_operator_ct_style_ct_total DEFAULT (0),
		uploaded_at DATETIME2 NOT NULL CONSTRAINT DF_operator_ct_style_uploaded_at DEFAULT SYSDATETIME()
	);

	CREATE UNIQUE INDEX UX_operator_ct_style_style_process
		ON dbo.operator_ct_style (style_name, process_name);
END
`
	_, err := r.DB.ExecContext(ctx, query)
	return err
}

func (r *Repository) GetOperatorCtStyleMaster(ctx context.Context) ([]models.OperatorCtStyleRecord, error) {
	query := `
SELECT
	id,
	style_name,
	process_name,
	ct_total,
	CONVERT(VARCHAR(19), uploaded_at, 120) AS uploaded_at
FROM dbo.operator_ct_style
ORDER BY style_name, process_name;
`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.OperatorCtStyleRecord, 0)
	for rows.Next() {
		var item models.OperatorCtStyleRecord
		if err := rows.Scan(
			&item.ID,
			&item.StyleName,
			&item.ProcessName,
			&item.CtTotal,
			&item.UploadedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	return out, rows.Err()
}

func (r *Repository) ImportOperatorCtStyle(
	ctx context.Context,
	input models.OperatorCtStyleImportRequest,
) (models.OperatorCtStyleImportResponse, error) {
	response := models.OperatorCtStyleImportResponse{
		Status:  "ok",
		Message: "Import CT Style-Proses berhasil",
		Total:   len(input.Rows),
	}

	if len(input.Rows) == 0 {
		return response, nil
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return models.OperatorCtStyleImportResponse{}, err
	}
	defer tx.Rollback()

	upsertQuery := `
MERGE dbo.operator_ct_style WITH (HOLDLOCK) AS target
USING (
	SELECT
		@style_name AS style_name,
		@process_name AS process_name,
		@ct_total AS ct_total
) AS source
ON LOWER(LTRIM(RTRIM(target.style_name))) = LOWER(LTRIM(RTRIM(source.style_name)))
 AND LOWER(LTRIM(RTRIM(target.process_name))) = LOWER(LTRIM(RTRIM(source.process_name)))
WHEN MATCHED THEN
	UPDATE SET
		ct_total = source.ct_total,
		uploaded_at = SYSDATETIME()
WHEN NOT MATCHED THEN
	INSERT (style_name, process_name, ct_total, uploaded_at)
	VALUES (
		source.style_name,
		source.process_name,
		source.ct_total,
		SYSDATETIME()
	);
`

	seen := make(map[string]bool)

	for i, row := range input.Rows {
		style := normalizeOperatorCtStylePart(row.StyleName)
		process := normalizeOperatorCtStylePart(row.ProcessName)

		if style == "" || process == "" {
			response.Skipped++
			continue
		}

		ctTotal := row.CtTotal
		if ctTotal < 0 {
			ctTotal = 0
		}

		key := strings.ToLower(style) + "|" + strings.ToLower(process)
		if seen[key] {
			response.Skipped++
			continue
		}
		seen[key] = true

		_, err = tx.ExecContext(
			ctx,
			upsertQuery,
			sql.Named("style_name", style),
			sql.Named("process_name", process),
			sql.Named("ct_total", ctTotal),
		)
		if err != nil {
			return models.OperatorCtStyleImportResponse{}, fmt.Errorf(
				"baris ke-%d gagal import: %w",
				i+1,
				err,
			)
		}

		response.Upserted++
	}

	if err := tx.Commit(); err != nil {
		return models.OperatorCtStyleImportResponse{}, err
	}

	return response, nil
}
