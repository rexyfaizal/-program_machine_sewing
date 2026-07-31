package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"backend_machine/models"
)

var ErrMachineOperatorNotFound = errors.New("machine operator session tidak ditemukan")

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanMachineOperatorSession(s rowScanner) (models.MachineOperatorSession, error) {
	var item models.MachineOperatorSession

	err := s.Scan(
		&item.ID,
		&item.SessionDate,
		&item.UUID,
		&item.MachineName,
		&item.Location,
		&item.OperatorNIK,
		&item.OperatorName,
		&item.BranchDetail,
		&item.ProcessName,
		&item.StyleName,
		&item.LoginTime,
		&item.LogoutTime,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	return item, err
}

func normalizeOperatorInput(input models.MachineOperatorLoginRequest) models.MachineOperatorLoginRequest {
	input.UUID = strings.TrimSpace(input.UUID)
	input.MachineName = strings.TrimSpace(input.MachineName)
	input.Location = strings.TrimSpace(input.Location)
	input.OperatorNIK = strings.TrimSpace(input.OperatorNIK)
	input.OperatorName = strings.TrimSpace(input.OperatorName)
	input.BranchDetail = strings.TrimSpace(input.BranchDetail)
	input.ProcessName = strings.TrimSpace(input.ProcessName)
	input.StyleName = strings.TrimSpace(input.StyleName)

	return input
}

func (r *Repository) fillOperatorFromEmployeeTx(ctx context.Context, tx *sql.Tx, input *models.MachineOperatorLoginRequest) error {
	if input.OperatorNIK == "" {
		return fmt.Errorf("operatorNik wajib diisi")
	}

	if input.OperatorName != "" && input.BranchDetail != "" {
		return nil
	}

	query := `
		SELECT TOP 1
			ISNULL(LTRIM(RTRIM(nik)), '') AS nik,
			ISNULL(LTRIM(RTRIM(name)), '') AS name,
			ISNULL(LTRIM(RTRIM(branchdetail)), '') AS branchdetail
		FROM dbo.employee
		WHERE LTRIM(RTRIM(nik)) = @nik
		ORDER BY nik ASC
	`

	var nik string
	var name string
	var branchdetail string

	err := tx.QueryRowContext(
		ctx,
		query,
		sql.Named("nik", input.OperatorNIK),
	).Scan(&nik, &name, &branchdetail)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("operator dengan NIK %s tidak ditemukan di dbo.employee", input.OperatorNIK)
		}
		return err
	}

	if input.OperatorName == "" {
		input.OperatorName = name
	}

	if input.BranchDetail == "" {
		input.BranchDetail = branchdetail
	}

	return nil
}

func getMachineOperatorSessionByIDTx(ctx context.Context, tx *sql.Tx, id int64) (models.MachineOperatorSession, error) {
	query := `
		SELECT TOP 1
			id,
			CONVERT(VARCHAR(10), session_date, 120) AS session_date,
			ISNULL(uuid, '') AS uuid,
			ISNULL(machine_name, '') AS machine_name,
			ISNULL(location, '') AS location,
			ISNULL(operator_nik, '') AS operator_nik,
			ISNULL(operator_name, '') AS operator_name,
			ISNULL(branchdetail, '') AS branchdetail,
			ISNULL(process_name, '') AS process_name,
			ISNULL(style_name, '') AS style_name,
			ISNULL(CONVERT(VARCHAR(19), login_time, 120), '') AS login_time,
			ISNULL(CONVERT(VARCHAR(19), logout_time, 120), '') AS logout_time,
			ISNULL(status, '') AS status,
			ISNULL(CONVERT(VARCHAR(19), created_at, 120), '') AS created_at,
			ISNULL(CONVERT(VARCHAR(19), updated_at, 120), '') AS updated_at
		FROM dbo.machine_operator_sessions
		WHERE id = @id
	`

	return scanMachineOperatorSession(tx.QueryRowContext(ctx, query, sql.Named("id", id)))
}

func GetActiveMachineOperatorByTx(ctx context.Context, tx *sql.Tx, uuid string) (models.MachineOperatorSession, error) {
	uuid = strings.TrimSpace(uuid)
	if uuid == "" {
		return models.MachineOperatorSession{}, fmt.Errorf("uuid wajib diisi")
	}

	query := `
		SELECT TOP 1
			id,
			CONVERT(VARCHAR(10), session_date, 120) AS session_date,
			ISNULL(uuid, '') AS uuid,
			ISNULL(machine_name, '') AS machine_name,
			ISNULL(location, '') AS location,
			ISNULL(operator_nik, '') AS operator_nik,
			ISNULL(operator_name, '') AS operator_name,
			ISNULL(branchdetail, '') AS branchdetail,
			ISNULL(process_name, '') AS process_name,
			ISNULL(style_name, '') AS style_name,
			ISNULL(CONVERT(VARCHAR(19), login_time, 120), '') AS login_time,
			ISNULL(CONVERT(VARCHAR(19), logout_time, 120), '') AS logout_time,
			ISNULL(status, '') AS status,
			ISNULL(CONVERT(VARCHAR(19), created_at, 120), '') AS created_at,
			ISNULL(CONVERT(VARCHAR(19), updated_at, 120), '') AS updated_at
		FROM dbo.machine_operator_sessions WITH (UPDLOCK, HOLDLOCK)
		WHERE LOWER(LTRIM(RTRIM(uuid))) = LOWER(LTRIM(RTRIM(@uuid)))
		  AND status = 'ACTIVE'
		  AND logout_time IS NULL
		ORDER BY login_time DESC, id DESC
	`

	session, err := scanMachineOperatorSession(tx.QueryRowContext(
		ctx,
		query,
		sql.Named("uuid", uuid),
	))

	if err != nil {
		if err == sql.ErrNoRows {
			return models.MachineOperatorSession{}, ErrMachineOperatorNotFound
		}
		return models.MachineOperatorSession{}, err
	}

	return session, nil
}
