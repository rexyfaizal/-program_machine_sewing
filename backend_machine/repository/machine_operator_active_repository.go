package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"backend_machine/models"
)

func (r *Repository) GetActiveMachineOperator(ctx context.Context, uuid string) (*models.MachineOperatorSession, error) {
	uuid = strings.TrimSpace(uuid)
	if uuid == "" {
		return nil, fmt.Errorf("uuid wajib diisi")
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
		FROM dbo.machine_operator_sessions
		WHERE LOWER(LTRIM(RTRIM(uuid))) = LOWER(LTRIM(RTRIM(@uuid)))
		  AND status = 'ACTIVE'
		  AND logout_time IS NULL
		ORDER BY login_time DESC, id DESC
	`

	session, err := scanMachineOperatorSession(r.DB.QueryRowContext(
		ctx,
		query,
		sql.Named("uuid", uuid),
	))

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrMachineOperatorNotFound
		}
		return nil, err
	}

	return &session, nil
}

func (r *Repository) GetMachineOperatorActiveStatus(ctx context.Context, uuid string) (models.MachineOperatorActiveResponse, error) {
	session, err := r.GetActiveMachineOperator(ctx, uuid)
	if err != nil {
		if errors.Is(err, ErrMachineOperatorNotFound) || errors.Is(err, sql.ErrNoRows) {
			return models.MachineOperatorActiveResponse{
				Active:  false,
				Message: "Tidak ada operator aktif",
				Session: nil,
			}, nil
		}

		return models.MachineOperatorActiveResponse{}, err
	}

	return models.MachineOperatorActiveResponse{
		Active:  true,
		Message: "Operator aktif",
		Session: session,
	}, nil
}
