package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"backend_machine/models"
)

func cleanProcessStyleImportText(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\u00a0", " ")
	value = strings.Join(strings.Fields(value), " ")
	return value
}

func (r *Repository) ImportProcessStyles(ctx context.Context, input models.ProcessStyleImportRequest) (models.ProcessStyleImportResponse, error) {
	response := models.ProcessStyleImportResponse{
		Status:  "ok",
		Message: "Import berhasil",
		Total:   len(input.Rows),
	}

	if len(input.Rows) == 0 {
		return response, nil
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return models.ProcessStyleImportResponse{}, err
	}
	defer tx.Rollback()

	seen := make(map[string]bool)

	query := `
		DECLARE @style_text NVARCHAR(100) = LTRIM(RTRIM(@style));
		DECLARE @proses_text NVARCHAR(500) = LTRIM(RTRIM(@proses));

		IF EXISTS (
			SELECT 1
			FROM dbo.dt_proses_style WITH (UPDLOCK, HOLDLOCK)
			WHERE LOWER(LTRIM(RTRIM(CAST(style AS NVARCHAR(100))))) = LOWER(@style_text)
			  AND LOWER(LTRIM(RTRIM(proses))) = LOWER(@proses_text)
		)
		BEGIN
			SELECT 0 AS inserted;
		END
		ELSE
		BEGIN
			INSERT INTO dbo.dt_proses_style (
				proses,
				style,
				created_at
			)
			VALUES (
				@proses_text,
				@style_text,
				SYSDATETIME()
			);

			SELECT 1 AS inserted;
		END
	`

	for i, row := range input.Rows {
		styleName := cleanProcessStyleImportText(row.Style)
		processName := cleanProcessStyleImportText(row.ProcessName)

		if processName == "" {
			processName = cleanProcessStyleImportText(row.Proses)
		}

		if styleName == "" || processName == "" {
			response.Skipped++
			continue
		}

		key := strings.ToLower(styleName) + "|" + strings.ToLower(processName)
		if seen[key] {
			response.Skipped++
			continue
		}
		seen[key] = true

		var inserted int

		err := tx.QueryRowContext(
			ctx,
			query,
			sql.Named("style", styleName),
			sql.Named("proses", processName),
		).Scan(&inserted)

		if err != nil {
			return models.ProcessStyleImportResponse{}, fmt.Errorf("baris ke-%d gagal import: %w", i+1, err)
		}

		if inserted == 1 {
			response.Inserted++
		} else {
			response.Skipped++
		}
	}

	if err := tx.Commit(); err != nil {
		return models.ProcessStyleImportResponse{}, err
	}

	return response, nil
}
