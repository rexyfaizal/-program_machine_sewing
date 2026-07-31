package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"backend_machine/models"
)

func (r *Repository) LoginMachineOperator(ctx context.Context, input models.MachineOperatorLoginRequest) (models.MachineOperatorLoginResponse, error) {
	input = normalizeOperatorInput(input)

	if input.UUID == "" {
		return models.MachineOperatorLoginResponse{}, fmt.Errorf("uuid wajib diisi")
	}

	if input.OperatorNIK == "" {
		return models.MachineOperatorLoginResponse{}, fmt.Errorf("operatorNik wajib diisi")
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return models.MachineOperatorLoginResponse{}, err
	}
	defer tx.Rollback()

	if err := r.fillOperatorFromEmployeeTx(ctx, tx, &input); err != nil {
		return models.MachineOperatorLoginResponse{}, err
	}

	activeQuery := `
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

	active, err := scanMachineOperatorSession(tx.QueryRowContext(
		ctx,
		activeQuery,
		sql.Named("uuid", input.UUID),
	))

	if err != nil && err != sql.ErrNoRows {
		return models.MachineOperatorLoginResponse{}, err
	}

	if err == nil {
		today := time.Now().Format("2006-01-02")
		activeSessionDate := strings.TrimSpace(active.SessionDate)

		// =====================================================
		// AUTO_LOGOUT_DAY_CHANGE
		// Jika session aktif berasal dari tanggal sebelumnya,
		// maka session lama ditutup otomatis dan dibuat session baru.
		// =====================================================
		if activeSessionDate != "" && activeSessionDate != today {
			closeDayChangeQuery := `
				UPDATE dbo.machine_operator_sessions
				SET
					logout_time = SYSDATETIME(),
					status = 'AUTO_LOGOUT_DAY_CHANGE',
					updated_at = SYSDATETIME()
				WHERE id = @id
				  AND status = 'ACTIVE'
				  AND logout_time IS NULL;

				UPDATE dbo.machine_operator_loss_events
				SET
					end_time = SYSDATETIME(),
					status = 'CLOSED',
					updated_at = SYSDATETIME()
				WHERE session_id = @id
				  AND status = 'ACTIVE'
				  AND end_time IS NULL;

				INSERT INTO dbo.machine_operator_notes (
					session_id,
					session_date,
					uuid,
					operator_nik,
					operator_name,
					reason_code,
					reason_name,
					note,
					created_at
				)
				SELECT
					@id,
					CAST(@session_date AS DATE),
					@uuid,
					@operator_nik,
					@operator_name,
					'AUTO_LOGOUT_DAY_CHANGE',
					'Auto Logout Ganti Hari',
					CONCAT(
						'Session ditutup otomatis karena sudah berganti hari. ',
						'Session lama tanggal ',
						@session_date,
						', scan baru tanggal ',
						@today,
						'.'
					),
					SYSDATETIME()
				WHERE NOT EXISTS (
					SELECT 1
					FROM dbo.machine_operator_notes n
					WHERE n.session_id = @id
					  AND n.reason_code = 'AUTO_LOGOUT_DAY_CHANGE'
				);
			`

			if _, err := tx.ExecContext(
				ctx,
				closeDayChangeQuery,
				sql.Named("id", active.ID),
				sql.Named("session_date", active.SessionDate),
				sql.Named("uuid", active.UUID),
				sql.Named("operator_nik", active.OperatorNIK),
				sql.Named("operator_name", active.OperatorName),
				sql.Named("today", today),
			); err != nil {
				return models.MachineOperatorLoginResponse{}, err
			}

			// Setelah session lama ditutup, lanjut ke insert session baru di bawah.
		} else if strings.EqualFold(strings.TrimSpace(active.OperatorNIK), input.OperatorNIK) {
			if err := tx.Commit(); err != nil {
				return models.MachineOperatorLoginResponse{}, err
			}

			return models.MachineOperatorLoginResponse{
				Status:     "active",
				Message:    "Operator sudah aktif di mesin ini",
				IsExisting: true,
				Session:    active,
			}, nil
		} else {
			// Operator berbeda scan di mesin yang sama.
			// Session lama ditutup sebagai AUTO_LOGOUT_REPLACED.
			closeOldQuery := `
				UPDATE dbo.machine_operator_sessions
				SET
					logout_time = SYSDATETIME(),
					status = 'AUTO_LOGOUT',
					updated_at = SYSDATETIME()
				WHERE id = @id
				  AND status = 'ACTIVE'
				  AND logout_time IS NULL;

				UPDATE dbo.machine_operator_loss_events
				SET
					end_time = SYSDATETIME(),
					status = 'CLOSED',
					updated_at = SYSDATETIME()
				WHERE session_id = @id
				  AND status = 'ACTIVE'
				  AND end_time IS NULL;

				INSERT INTO dbo.machine_operator_notes (
					session_id,
					session_date,
					uuid,
					operator_nik,
					operator_name,
					reason_code,
					reason_name,
					note,
					created_at
				)
				SELECT
					@id,
					CAST(@session_date AS DATE),
					@uuid,
					@operator_nik,
					@operator_name,
					'AUTO_LOGOUT_REPLACED',
					'Auto Logout Ganti Operator',
					CONCAT(
						'Session ditutup otomatis karena operator baru scan: ',
						@new_operator_nik,
						' - ',
						@new_operator_name
					),
					SYSDATETIME()
				WHERE NOT EXISTS (
					SELECT 1
					FROM dbo.machine_operator_notes n
					WHERE n.session_id = @id
					  AND n.reason_code = 'AUTO_LOGOUT_REPLACED'
				);
			`

			if _, err := tx.ExecContext(
				ctx,
				closeOldQuery,
				sql.Named("id", active.ID),
				sql.Named("session_date", active.SessionDate),
				sql.Named("uuid", active.UUID),
				sql.Named("operator_nik", active.OperatorNIK),
				sql.Named("operator_name", active.OperatorName),
				sql.Named("new_operator_nik", input.OperatorNIK),
				sql.Named("new_operator_name", input.OperatorName),
			); err != nil {
				return models.MachineOperatorLoginResponse{}, err
			}
		}
	}

	insertQuery := `
		INSERT INTO dbo.machine_operator_sessions (
			session_date,
			uuid,
			machine_name,
			location,
			operator_nik,
			operator_name,
			branchdetail,
			process_name,
			style_name,
			login_time,
			logout_time,
			status,
			created_at,
			updated_at
		)
		OUTPUT INSERTED.id
		VALUES (
			CAST(GETDATE() AS DATE),
			@uuid,
			@machine_name,
			@location,
			@operator_nik,
			@operator_name,
			@branchdetail,
			@process_name,
			@style_name,
			SYSDATETIME(),
			NULL,
			'ACTIVE',
			SYSDATETIME(),
			SYSDATETIME()
		)
	`

	var newID int64
	err = tx.QueryRowContext(
		ctx,
		insertQuery,
		sql.Named("uuid", input.UUID),
		sql.Named("machine_name", input.MachineName),
		sql.Named("location", input.Location),
		sql.Named("operator_nik", input.OperatorNIK),
		sql.Named("operator_name", input.OperatorName),
		sql.Named("branchdetail", input.BranchDetail),
		sql.Named("process_name", input.ProcessName),
		sql.Named("style_name", input.StyleName),
	).Scan(&newID)

	if err != nil {
		return models.MachineOperatorLoginResponse{}, err
	}

	session, err := getMachineOperatorSessionByIDTx(ctx, tx, newID)
	if err != nil {
		return models.MachineOperatorLoginResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.MachineOperatorLoginResponse{}, err
	}

	return models.MachineOperatorLoginResponse{
		Status:     "ok",
		Message:    "Operator berhasil login ke mesin",
		IsExisting: false,
		Session:    session,
	}, nil
}
