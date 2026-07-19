package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"backend_machine/models"
)

var ErrProcessStyleNotFound = errors.New("process style tidak ditemukan")

func (r *Repository) GetProcessStyleStyles(ctx context.Context, q string) ([]models.ProcessStyleItem, error) {
	q = strings.TrimSpace(q)

	query := `
		SELECT DISTINCT
			CAST([style] AS NVARCHAR(100)) AS styleName
		FROM [sewingiot].[dbo].[dt_proses_style]
		WHERE
			(@q = '' OR CAST([style] AS NVARCHAR(100)) LIKE '%' + @q + '%')
		ORDER BY
			CAST([style] AS NVARCHAR(100));
	`

	rows, err := r.DB.QueryContext(ctx, query, sql.Named("q", q))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]models.ProcessStyleItem, 0)

	for rows.Next() {
		var item models.ProcessStyleItem
		if err := rows.Scan(&item.StyleName); err != nil {
			return nil, err
		}
		data = append(data, item)
	}

	return data, rows.Err()
}

func (r *Repository) GetProcessStyleProcesses(ctx context.Context, style string, q string) ([]models.ProcessStyleProcess, error) {
	style = strings.TrimSpace(style)
	q = strings.TrimSpace(q)

	query := `
		SELECT
			[id],
			[proses] AS processName,
			CAST([style] AS NVARCHAR(100)) AS styleName
		FROM [sewingiot].[dbo].[dt_proses_style]
		WHERE
			CAST([style] AS NVARCHAR(100)) = @style
			AND (@q = '' OR [proses] LIKE '%' + @q + '%')
		ORDER BY
			[id];
	`

	rows, err := r.DB.QueryContext(
		ctx,
		query,
		sql.Named("style", style),
		sql.Named("q", q),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]models.ProcessStyleProcess, 0)

	for rows.Next() {
		var item models.ProcessStyleProcess
		if err := rows.Scan(&item.ID, &item.ProcessName, &item.StyleName); err != nil {
			return nil, err
		}
		data = append(data, item)
	}

	return data, rows.Err()
}

func (r *Repository) ListProcessStyle(ctx context.Context, q string) ([]models.ProcessStyleRecord, error) {
	q = strings.TrimSpace(q)

	query := `
		SELECT
			[id],
			ISNULL([proses], '') AS processName,
			CAST([style] AS NVARCHAR(100)) AS styleName,
			ISNULL(CONVERT(VARCHAR(19), [created_at], 120), '') AS createdAt
		FROM [sewingiot].[dbo].[dt_proses_style]
		WHERE
			@q = ''
			OR CAST([style] AS NVARCHAR(100)) LIKE '%' + @q + '%'
			OR [proses] LIKE '%' + @q + '%'
		ORDER BY
			CAST([style] AS NVARCHAR(100)),
			[id];
	`

	rows, err := r.DB.QueryContext(ctx, query, sql.Named("q", q))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]models.ProcessStyleRecord, 0)

	for rows.Next() {
		var item models.ProcessStyleRecord

		if err := rows.Scan(
			&item.ID,
			&item.ProcessName,
			&item.StyleName,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		data = append(data, item)
	}

	return data, rows.Err()
}

func (r *Repository) GetProcessStyleByID(ctx context.Context, id int64) (models.ProcessStyleRecord, error) {
	query := `
		SELECT TOP 1
			[id],
			ISNULL([proses], '') AS processName,
			CAST([style] AS NVARCHAR(100)) AS styleName,
			ISNULL(CONVERT(VARCHAR(19), [created_at], 120), '') AS createdAt
		FROM [sewingiot].[dbo].[dt_proses_style]
		WHERE [id] = @id;
	`

	var item models.ProcessStyleRecord

	err := r.DB.QueryRowContext(ctx, query, sql.Named("id", id)).Scan(
		&item.ID,
		&item.ProcessName,
		&item.StyleName,
		&item.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return item, ErrProcessStyleNotFound
		}
		return item, err
	}

	return item, nil
}

func (r *Repository) CreateProcessStyle(ctx context.Context, input models.ProcessStyleRequest) (models.ProcessStyleRecord, error) {
	processName := strings.TrimSpace(input.ProcessName)
	styleName := strings.TrimSpace(input.StyleName)

	query := `
		INSERT INTO [sewingiot].[dbo].[dt_proses_style] (
			[proses],
			[style],
			[created_at]
		)
		OUTPUT INSERTED.[id]
		VALUES (
			@proses,
			@style,
			SYSDATETIME()
		);
	`

	var id int64

	err := r.DB.QueryRowContext(
		ctx,
		query,
		sql.Named("proses", processName),
		sql.Named("style", styleName),
	).Scan(&id)

	if err != nil {
		return models.ProcessStyleRecord{}, err
	}

	return r.GetProcessStyleByID(ctx, id)
}

func (r *Repository) UpdateProcessStyle(ctx context.Context, id int64, input models.ProcessStyleRequest) (models.ProcessStyleRecord, error) {
	processName := strings.TrimSpace(input.ProcessName)
	styleName := strings.TrimSpace(input.StyleName)

	query := `
		UPDATE [sewingiot].[dbo].[dt_proses_style]
		SET
			[proses] = @proses,
			[style] = @style
		WHERE [id] = @id;
	`

	result, err := r.DB.ExecContext(
		ctx,
		query,
		sql.Named("id", id),
		sql.Named("proses", processName),
		sql.Named("style", styleName),
	)
	if err != nil {
		return models.ProcessStyleRecord{}, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return models.ProcessStyleRecord{}, err
	}

	if affected == 0 {
		return models.ProcessStyleRecord{}, ErrProcessStyleNotFound
	}

	return r.GetProcessStyleByID(ctx, id)
}

func (r *Repository) DeleteProcessStyle(ctx context.Context, id int64) error {
	query := `
		DELETE FROM [sewingiot].[dbo].[dt_proses_style]
		WHERE [id] = @id;
	`

	result, err := r.DB.ExecContext(ctx, query, sql.Named("id", id))
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrProcessStyleNotFound
	}

	return nil
}
