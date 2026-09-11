package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"backend_machine/models"
)

func (r *Repository) EnsureProcessStyleExists(ctx context.Context, styleName string) error {
	styleName = strings.TrimSpace(styleName)
	if styleName == "" {
		return fmt.Errorf("style wajib diisi")
	}

	query := `
		SELECT TOP 1 CAST([style] AS NVARCHAR(100))
		FROM [sewingiot].[dbo].[dt_proses_style]
		WHERE LOWER(LTRIM(RTRIM(CAST([style] AS NVARCHAR(100))))) = LOWER(LTRIM(RTRIM(@style)));
	`

	var found string
	err := r.DB.QueryRowContext(ctx, query, sql.Named("style", styleName)).Scan(&found)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("style tidak ditemukan di Master IE: %s", styleName)
		}
		return err
	}

	return nil
}

func (r *Repository) ImportOperatorStyleCorrection(
	ctx context.Context,
	input models.OperatorStyleCorrectionImportRequest,
) (models.OperatorStyleCorrectionImportResponse, error) {
	response := models.OperatorStyleCorrectionImportResponse{
		Status:  "ok",
		Message: "Koreksi style sesi berhasil",
		Total:   len(input.Rows),
	}

	if len(input.Rows) == 0 {
		return response, nil
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return models.OperatorStyleCorrectionImportResponse{}, err
	}
	defer tx.Rollback()

	updateQuery := `
UPDATE dbo.machine_operator_sessions
SET
	style_name = @style_name,
	process_name = CASE
		WHEN @process_name <> '' THEN @process_name
		ELSE process_name
	END,
	updated_at = SYSDATETIME()
WHERE session_date = CAST(@work_date AS DATE)
  AND LOWER(LTRIM(RTRIM(uuid))) = LOWER(LTRIM(RTRIM(@uuid)))
  AND (
		@operator_nik = ''
		OR LOWER(LTRIM(RTRIM(operator_nik))) = LOWER(LTRIM(RTRIM(@operator_nik)))
  );
`

	seen := make(map[string]bool)

	for i, row := range input.Rows {
		workDate, err := normalizeOperatorOutputTargetDate(row.WorkDate)
		if err != nil {
			response.Skipped++
			continue
		}

		uuid := strings.TrimSpace(row.UUID)
		styleName := normalizeOperatorCtStylePart(row.StyleName)
		processName := normalizeOperatorCtStylePart(row.ProcessName)
		operatorNik := strings.TrimSpace(row.OperatorNIK)

		if uuid == "" || styleName == "" {
			response.Skipped++
			continue
		}

		key := strings.ToLower(workDate) + "|" + strings.ToLower(uuid) + "|" + strings.ToLower(operatorNik)
		if seen[key] {
			response.Skipped++
			continue
		}
		seen[key] = true

		if processName != "" {
			if err := r.EnsureProcessStylePairExists(ctx, tx, styleName, processName); err != nil {
				return models.OperatorStyleCorrectionImportResponse{}, fmt.Errorf(
					"baris ke-%d: %w",
					i+1,
					err,
				)
			}
		} else {
			if err := r.EnsureProcessStyleExists(ctx, styleName); err != nil {
				return models.OperatorStyleCorrectionImportResponse{}, fmt.Errorf(
					"baris ke-%d: %w",
					i+1,
					err,
				)
			}
		}

		result, err := tx.ExecContext(
			ctx,
			updateQuery,
			sql.Named("work_date", workDate),
			sql.Named("uuid", uuid),
			sql.Named("operator_nik", operatorNik),
			sql.Named("style_name", styleName),
			sql.Named("process_name", processName),
		)
		if err != nil {
			return models.OperatorStyleCorrectionImportResponse{}, fmt.Errorf(
				"baris ke-%d gagal update sesi: %w",
				i+1,
				err,
			)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return models.OperatorStyleCorrectionImportResponse{}, err
		}

		if affected == 0 {
			response.Skipped++
			continue
		}

		response.Updated += int(affected)
	}

	if err := tx.Commit(); err != nil {
		return models.OperatorStyleCorrectionImportResponse{}, err
	}

	return response, nil
}
