package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"backend_machine/models"
)

const gccSyncInsertBatchSize = 200

// ReplaceProcessStylesFromGcc menghapus semua Master IE + CT style lama,
// lalu insert ulang dari hasil flatten API GCC.
func (r *Repository) ReplaceProcessStylesFromGcc(
	ctx context.Context,
	rows []models.GccSmvSyncRow,
) (stylesInserted int, ctInserted int, skippedEmpty int, err error) {
	if err := r.EnsureOperatorCtStyleSchema(ctx); err != nil {
		return 0, 0, 0, err
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, 0, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM dbo.operator_ct_style;`); err != nil {
		return 0, 0, 0, fmt.Errorf("gagal hapus CT style lama: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM dbo.dt_proses_style;`); err != nil {
		return 0, 0, 0, fmt.Errorf("gagal hapus Master IE lama: %w", err)
	}

	cleanRows := make([]models.GccSmvSyncRow, 0, len(rows))
	seen := make(map[string]bool, len(rows))
	for _, row := range rows {
		style := cleanProcessStyleImportText(row.StyleName)
		process := cleanProcessStyleImportText(row.ProcessName)
		if style == "" || process == "" {
			skippedEmpty++
			continue
		}
		key := strings.ToLower(style) + "|" + strings.ToLower(process)
		if seen[key] {
			skippedEmpty++
			continue
		}
		seen[key] = true

		ct := row.CtTotal
		if ct < 0 {
			ct = 0
		}
		cleanRows = append(cleanRows, models.GccSmvSyncRow{
			StyleName:   style,
			ProcessName: process,
			CtTotal:     ct,
		})
	}

	for i := 0; i < len(cleanRows); i += gccSyncInsertBatchSize {
		end := i + gccSyncInsertBatchSize
		if end > len(cleanRows) {
			end = len(cleanRows)
		}
		batch := cleanRows[i:end]

		n, err := insertProcessStyleBatch(ctx, tx, batch)
		if err != nil {
			return 0, 0, skippedEmpty, err
		}
		stylesInserted += n

		nCt, err := insertOperatorCtStyleBatch(ctx, tx, batch)
		if err != nil {
			return 0, 0, skippedEmpty, err
		}
		ctInserted += nCt
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, skippedEmpty, err
	}
	return stylesInserted, ctInserted, skippedEmpty, nil
}

func insertProcessStyleBatch(ctx context.Context, tx *sql.Tx, batch []models.GccSmvSyncRow) (int, error) {
	if len(batch) == 0 {
		return 0, nil
	}

	var b strings.Builder
	args := make([]any, 0, len(batch)*2)
	b.WriteString(`INSERT INTO dbo.dt_proses_style (proses, style, created_at) VALUES `)

	for i, row := range batch {
		if i > 0 {
			b.WriteString(",")
		}
		styleKey := fmt.Sprintf("style%d", i)
		prosesKey := fmt.Sprintf("proses%d", i)
		b.WriteString(fmt.Sprintf("(@%s, @%s, SYSDATETIME())", prosesKey, styleKey))
		args = append(args,
			sql.Named(prosesKey, row.ProcessName),
			sql.Named(styleKey, row.StyleName),
		)
	}

	if _, err := tx.ExecContext(ctx, b.String(), args...); err != nil {
		return 0, fmt.Errorf("gagal insert batch Master IE: %w", err)
	}
	return len(batch), nil
}

func insertOperatorCtStyleBatch(ctx context.Context, tx *sql.Tx, batch []models.GccSmvSyncRow) (int, error) {
	if len(batch) == 0 {
		return 0, nil
	}

	var b strings.Builder
	args := make([]any, 0, len(batch)*3)
	b.WriteString(`INSERT INTO dbo.operator_ct_style (style_name, process_name, ct_total, uploaded_at) VALUES `)

	for i, row := range batch {
		if i > 0 {
			b.WriteString(",")
		}
		styleKey := fmt.Sprintf("ct_style%d", i)
		prosesKey := fmt.Sprintf("ct_proses%d", i)
		ctKey := fmt.Sprintf("ct_total%d", i)
		b.WriteString(fmt.Sprintf("(@%s, @%s, @%s, SYSDATETIME())", styleKey, prosesKey, ctKey))
		args = append(args,
			sql.Named(styleKey, row.StyleName),
			sql.Named(prosesKey, row.ProcessName),
			sql.Named(ctKey, row.CtTotal),
		)
	}

	if _, err := tx.ExecContext(ctx, b.String(), args...); err != nil {
		return 0, fmt.Errorf("gagal insert batch CT style: %w", err)
	}
	return len(batch), nil
}
