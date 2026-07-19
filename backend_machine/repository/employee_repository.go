package repository

import (
	"context"
	"database/sql"
	"strings"

	"backend_machine/models"
)

func (r *Repository) SearchEmployees(ctx context.Context, q string, limit int) ([]models.Employee, error) {
	q = strings.TrimSpace(q)

	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT TOP (@limit)
			ISNULL(LTRIM(RTRIM(nik)), '') AS nik,
			ISNULL(LTRIM(RTRIM(name)), '') AS name,
			ISNULL(LTRIM(RTRIM(branchdetail)), '') AS branchdetail
		FROM dbo.employee
		WHERE
			LTRIM(RTRIM(nik)) = @q
			OR LTRIM(RTRIM(nik)) LIKE @prefix
			OR name LIKE @contains
		ORDER BY
			CASE
				WHEN LTRIM(RTRIM(nik)) = @q THEN 0
				WHEN LTRIM(RTRIM(nik)) LIKE @prefix THEN 1
				ELSE 2
			END,
			nik ASC
	`

	rows, err := r.DB.QueryContext(
		ctx,
		query,
		sql.Named("limit", limit),
		sql.Named("q", q),
		sql.Named("prefix", q+"%"),
		sql.Named("contains", "%"+q+"%"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	employees := make([]models.Employee, 0)

	for rows.Next() {
		var item models.Employee

		if err := rows.Scan(
			&item.NIK,
			&item.Name,
			&item.BranchDetail,
		); err != nil {
			return nil, err
		}

		employees = append(employees, item)
	}

	return employees, rows.Err()
}
