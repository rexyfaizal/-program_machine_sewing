package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"backend_machine/models"
)

// EnsureSpvLineAssignmentSchema memastikan tabel spv_line_assignments ada
// (reuse tabel yang sudah dipakai telegram_notif — jangan duplikat master).
func (r *Repository) EnsureSpvLineAssignmentSchema(ctx context.Context) error {
	_, err := r.DB.ExecContext(ctx, `
IF OBJECT_ID(N'dbo.spv_line_assignments', N'U') IS NULL
BEGIN
	CREATE TABLE dbo.spv_line_assignments (
		id BIGINT IDENTITY(1,1) NOT NULL PRIMARY KEY,
		location NVARCHAR(255) NOT NULL,
		spv_nik NVARCHAR(50) NOT NULL,
		branchdetail NVARCHAR(255) NULL,
		is_active BIT NOT NULL CONSTRAINT DF_spv_line_assignments_is_active DEFAULT (1),
		updated_at DATETIME2 NOT NULL CONSTRAINT DF_spv_line_assignments_updated_at DEFAULT (SYSDATETIME())
	);

	CREATE UNIQUE INDEX UX_spv_line_assignments_location
		ON dbo.spv_line_assignments(location);
END;

IF COL_LENGTH('dbo.spv_line_assignments', 'branchdetail') IS NULL
BEGIN
	ALTER TABLE dbo.spv_line_assignments ADD branchdetail NVARCHAR(255) NULL;
END;

IF COL_LENGTH('dbo.spv_line_assignments', 'is_active') IS NULL
BEGIN
	ALTER TABLE dbo.spv_line_assignments ADD is_active BIT NOT NULL
		CONSTRAINT DF_spv_line_assignments_is_active_2 DEFAULT (1);
END;

IF COL_LENGTH('dbo.spv_line_assignments', 'updated_at') IS NULL
BEGIN
	ALTER TABLE dbo.spv_line_assignments ADD updated_at DATETIME2 NOT NULL
		CONSTRAINT DF_spv_line_assignments_updated_at_2 DEFAULT (SYSDATETIME());
END;

IF NOT EXISTS (
	SELECT 1
	FROM sys.indexes
	WHERE name = 'UX_spv_line_assignments_location'
	  AND object_id = OBJECT_ID(N'dbo.spv_line_assignments')
)
BEGIN
	CREATE UNIQUE INDEX UX_spv_line_assignments_location
		ON dbo.spv_line_assignments(location);
END;
`)
	return err
}

func (r *Repository) GetEmployeeByNIK(ctx context.Context, nik string) (models.Employee, error) {
	nik = strings.TrimSpace(nik)
	if nik == "" {
		return models.Employee{}, fmt.Errorf("nik wajib diisi")
	}

	query := `
		SELECT TOP 1
			ISNULL(LTRIM(RTRIM(nik)), '') AS nik,
			ISNULL(LTRIM(RTRIM(name)), '') AS name,
			ISNULL(LTRIM(RTRIM(branchdetail)), '') AS branchdetail
		FROM dbo.employee
		WHERE LTRIM(RTRIM(CONVERT(varchar(255), nik))) = @nik
		ORDER BY nik ASC
	`

	var item models.Employee
	err := r.DB.QueryRowContext(ctx, query, sql.Named("nik", nik)).Scan(
		&item.NIK,
		&item.Name,
		&item.BranchDetail,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Employee{}, fmt.Errorf("employee tidak ditemukan: NIK %s", nik)
		}
		return models.Employee{}, err
	}

	return item, nil
}

func (r *Repository) ListSpvLineAssignments(ctx context.Context, factory string) ([]models.SpvLineAssignment, error) {
	if err := r.EnsureSpvLineAssignmentSchema(ctx); err != nil {
		return nil, err
	}

	factory = strings.ToUpper(strings.TrimSpace(factory))

	query := `
		SELECT
			ISNULL(LTRIM(RTRIM(a.location)), '') AS location,
			ISNULL(LTRIM(RTRIM(a.spv_nik)), '') AS spv_nik,
			ISNULL(LTRIM(RTRIM(e.name)), '') AS spv_name,
			ISNULL(LTRIM(RTRIM(a.branchdetail)), '') AS branchdetail,
			CAST(ISNULL(a.is_active, 0) AS INT) AS is_active
		FROM dbo.spv_line_assignments a
		LEFT JOIN dbo.employee e
			ON LTRIM(RTRIM(CONVERT(varchar(255), e.nik))) =
			   LTRIM(RTRIM(a.spv_nik))
		WHERE ISNULL(a.is_active, 0) = 1
	`

	args := []any{}
	if factory != "" {
		query += `
		  AND (
			UPPER(LTRIM(RTRIM(a.location))) LIKE @prefix
			OR UPPER(LTRIM(RTRIM(a.location))) LIKE @prefix_dash
		  )
		`
		args = append(args,
			sql.Named("prefix", factory+" %"),
			sql.Named("prefix_dash", factory+" - %"),
		)
	}

	query += `
		ORDER BY a.location ASC
	`

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]models.SpvLineAssignment, 0)
	for rows.Next() {
		var item models.SpvLineAssignment
		var activeInt int
		if err := rows.Scan(
			&item.Location,
			&item.SpvNik,
			&item.SpvName,
			&item.BranchDetail,
			&activeInt,
		); err != nil {
			return nil, err
		}
		item.IsActive = activeInt == 1
		list = append(list, item)
	}

	return list, rows.Err()
}

func (r *Repository) UpsertSpvLineAssignment(ctx context.Context, location, spvNik string) (models.SpvLineAssignment, error) {
	if err := r.EnsureSpvLineAssignmentSchema(ctx); err != nil {
		return models.SpvLineAssignment{}, err
	}

	location = strings.TrimSpace(location)
	spvNik = strings.TrimSpace(spvNik)
	if location == "" {
		return models.SpvLineAssignment{}, fmt.Errorf("location wajib diisi")
	}

	// spvNik kosong = nonaktifkan assignment line.
	if spvNik == "" {
		_, err := r.DB.ExecContext(ctx, `
			UPDATE dbo.spv_line_assignments
			SET is_active = 0,
				updated_at = SYSDATETIME()
			WHERE UPPER(LTRIM(RTRIM(location))) = UPPER(LTRIM(RTRIM(@location)));
		`, sql.Named("location", location))
		if err != nil {
			return models.SpvLineAssignment{}, err
		}

		return models.SpvLineAssignment{
			Location: location,
			IsActive: false,
		}, nil
	}

	emp, err := r.GetEmployeeByNIK(ctx, spvNik)
	if err != nil {
		return models.SpvLineAssignment{}, err
	}

	_, err = r.DB.ExecContext(ctx, `
		MERGE dbo.spv_line_assignments AS target
		USING (SELECT @location AS location) AS source
		ON UPPER(LTRIM(RTRIM(target.location))) = UPPER(LTRIM(RTRIM(source.location)))
		WHEN MATCHED THEN
			UPDATE SET
				spv_nik = @spv_nik,
				branchdetail = @branchdetail,
				is_active = 1,
				updated_at = SYSDATETIME()
		WHEN NOT MATCHED THEN
			INSERT (location, spv_nik, branchdetail, is_active, updated_at)
			VALUES (@location, @spv_nik, @branchdetail, 1, SYSDATETIME());
	`,
		sql.Named("location", location),
		sql.Named("spv_nik", emp.NIK),
		sql.Named("branchdetail", emp.BranchDetail),
	)
	if err != nil {
		return models.SpvLineAssignment{}, err
	}

	return models.SpvLineAssignment{
		Location:     location,
		SpvNik:       emp.NIK,
		SpvName:      emp.Name,
		BranchDetail: emp.BranchDetail,
		IsActive:     true,
	}, nil
}

func (r *Repository) RenameSpvLineAssignmentLocation(ctx context.Context, oldLocation, newLocation string) error {
	if err := r.EnsureSpvLineAssignmentSchema(ctx); err != nil {
		return err
	}

	oldLocation = strings.TrimSpace(oldLocation)
	newLocation = strings.TrimSpace(newLocation)
	if oldLocation == "" || newLocation == "" {
		return fmt.Errorf("oldLocation dan newLocation wajib diisi")
	}
	if strings.EqualFold(oldLocation, newLocation) {
		return nil
	}

	_, err := r.DB.ExecContext(ctx, `
		-- Nonaktifkan target baru jika sudah ada (hindari unique conflict).
		UPDATE dbo.spv_line_assignments
		SET is_active = 0,
			updated_at = SYSDATETIME()
		WHERE UPPER(LTRIM(RTRIM(location))) = UPPER(LTRIM(RTRIM(@new_location)))
		  AND UPPER(LTRIM(RTRIM(location))) <> UPPER(LTRIM(RTRIM(@old_location)));

		UPDATE dbo.spv_line_assignments
		SET location = @new_location,
			updated_at = SYSDATETIME()
		WHERE UPPER(LTRIM(RTRIM(location))) = UPPER(LTRIM(RTRIM(@old_location)));
	`,
		sql.Named("old_location", oldLocation),
		sql.Named("new_location", newLocation),
	)
	return err
}
