package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"backend_machine/models"
)

var ErrMachineOperatorLossEventActive = errors.New("loss event masih aktif")
var ErrMachineOperatorLossEventNotFound = errors.New("loss event aktif tidak ditemukan")

func scanMachineOperatorLossEvent(s rowScanner) (models.MachineOperatorLossEvent, error) {
	var item models.MachineOperatorLossEvent

	err := s.Scan(
		&item.ID,
		&item.SessionID,
		&item.SessionDate,
		&item.UUID,
		&item.MachineName,
		&item.Location,
		&item.OperatorNIK,
		&item.OperatorName,
		&item.ReasonCode,
		&item.ReasonLabel,
		&item.Note,
		&item.StartTime,
		&item.EndTime,
		&item.DurationSeconds,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	return item, err
}

func (r *Repository) getActiveLossEventBySessionTx(ctx context.Context, tx *sql.Tx, sessionID int64) (*models.MachineOperatorLossEvent, error) {
	query := `
		SELECT TOP 1
			id,
			session_id,

			-- DB tidak punya kolom session_date, jadi ambil dari start_time
			ISNULL(CONVERT(VARCHAR(10), start_time, 120), '') AS session_date,

			ISNULL(uuid, '') AS uuid,

			-- DB tidak punya kolom machine_name dan location
			'' AS machine_name,
			'' AS location,

			ISNULL(operator_nik, '') AS operator_nik,
			ISNULL(operator_name, '') AS operator_name,
			ISNULL(reason_code, '') AS reason_code,
			ISNULL(reason_label, '') AS reason_label,
			ISNULL(note, '') AS note,
			ISNULL(CONVERT(VARCHAR(19), start_time, 120), '') AS start_time,
			ISNULL(CONVERT(VARCHAR(19), end_time, 120), '') AS end_time,

			-- Kalau masih ACTIVE, durasi dihitung realtime.
			CAST(
				CASE
					WHEN status = 'ACTIVE' AND end_time IS NULL
						THEN DATEDIFF_BIG(SECOND, start_time, SYSDATETIME())
					ELSE ISNULL(duration_sec, 0)
				END AS BIGINT
			) AS duration_seconds,

			ISNULL(status, '') AS status,
			ISNULL(CONVERT(VARCHAR(19), created_at, 120), '') AS created_at,
			ISNULL(CONVERT(VARCHAR(19), updated_at, 120), '') AS updated_at
		FROM dbo.machine_operator_loss_events WITH (UPDLOCK, HOLDLOCK)
		WHERE session_id = @session_id
		  AND status = 'ACTIVE'
		  AND end_time IS NULL
		ORDER BY start_time DESC, id DESC
	`

	event, err := scanMachineOperatorLossEvent(tx.QueryRowContext(
		ctx,
		query,
		sql.Named("session_id", sessionID),
	))

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &event, nil
}

func getMachineOperatorLossEventByIDTx(ctx context.Context, tx *sql.Tx, id int64) (models.MachineOperatorLossEvent, error) {
	query := `
		SELECT TOP 1
			id,
			session_id,

			-- DB tidak punya kolom session_date, jadi ambil dari start_time
			ISNULL(CONVERT(VARCHAR(10), start_time, 120), '') AS session_date,

			ISNULL(uuid, '') AS uuid,

			-- DB tidak punya kolom machine_name dan location
			'' AS machine_name,
			'' AS location,

			ISNULL(operator_nik, '') AS operator_nik,
			ISNULL(operator_name, '') AS operator_name,
			ISNULL(reason_code, '') AS reason_code,
			ISNULL(reason_label, '') AS reason_label,
			ISNULL(note, '') AS note,
			ISNULL(CONVERT(VARCHAR(19), start_time, 120), '') AS start_time,
			ISNULL(CONVERT(VARCHAR(19), end_time, 120), '') AS end_time,

			CAST(
				CASE
					WHEN status = 'ACTIVE' AND end_time IS NULL
						THEN DATEDIFF_BIG(SECOND, start_time, SYSDATETIME())
					ELSE ISNULL(duration_sec, 0)
				END AS BIGINT
			) AS duration_seconds,

			ISNULL(status, '') AS status,
			ISNULL(CONVERT(VARCHAR(19), created_at, 120), '') AS created_at,
			ISNULL(CONVERT(VARCHAR(19), updated_at, 120), '') AS updated_at
		FROM dbo.machine_operator_loss_events
		WHERE id = @id
	`

	return scanMachineOperatorLossEvent(tx.QueryRowContext(
		ctx,
		query,
		sql.Named("id", id),
	))
}

func (r *Repository) StartMachineOperatorLossEvent(ctx context.Context, input models.MachineOperatorLossEventStartRequest) (models.MachineOperatorLossEventStartResponse, error) {
	input.UUID = strings.TrimSpace(input.UUID)
	input.ReasonCode = strings.ToUpper(strings.TrimSpace(input.ReasonCode))
	input.ReasonLabel = strings.TrimSpace(input.ReasonLabel)
	input.Note = strings.TrimSpace(input.Note)

	if input.UUID == "" {
		return models.MachineOperatorLossEventStartResponse{}, fmt.Errorf("uuid wajib diisi")
	}

	if input.ReasonCode == "" {
		return models.MachineOperatorLossEventStartResponse{}, fmt.Errorf("reasonCode wajib diisi")
	}

	if input.ReasonLabel == "" {
		input.ReasonLabel = input.ReasonCode
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return models.MachineOperatorLossEventStartResponse{}, err
	}
	defer tx.Rollback()

	session, err := GetActiveMachineOperatorByTx(ctx, tx, input.UUID)
	if err != nil {
		return models.MachineOperatorLossEventStartResponse{}, err
	}

	activeEvent, err := r.getActiveLossEventBySessionTx(ctx, tx, session.ID)
	if err != nil {
		return models.MachineOperatorLossEventStartResponse{}, err
	}

	if activeEvent != nil {
		if err := tx.Commit(); err != nil {
			return models.MachineOperatorLossEventStartResponse{}, err
		}

		return models.MachineOperatorLossEventStartResponse{
			Status:     "active",
			Message:    "Loss event masih aktif, belum bisa membuat event baru",
			IsExisting: true,
			Event:      *activeEvent,
		}, nil
	}

	insertQuery := `
		INSERT INTO dbo.machine_operator_loss_events (
			session_id,
			uuid,
			operator_nik,
			operator_name,
			reason_code,
			reason_label,
			note,
			start_time,
			end_time,
			status,
			created_at,
			updated_at
		)
		OUTPUT INSERTED.id
		VALUES (
			@session_id,
			@uuid,
			@operator_nik,
			@operator_name,
			@reason_code,
			@reason_label,
			@note,
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
		sql.Named("session_id", session.ID),
		sql.Named("uuid", session.UUID),
		sql.Named("operator_nik", session.OperatorNIK),
		sql.Named("operator_name", session.OperatorName),
		sql.Named("reason_code", input.ReasonCode),
		sql.Named("reason_label", input.ReasonLabel),
		sql.Named("note", input.Note),
	).Scan(&newID)

	if err != nil {
		return models.MachineOperatorLossEventStartResponse{}, err
	}

	event, err := getMachineOperatorLossEventByIDTx(ctx, tx, newID)
	if err != nil {
		return models.MachineOperatorLossEventStartResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.MachineOperatorLossEventStartResponse{}, err
	}

	return models.MachineOperatorLossEventStartResponse{
		Status:     "ok",
		Message:    "Loss event berhasil dimulai",
		IsExisting: false,
		Event:      event,
	}, nil
}

func (r *Repository) FinishMachineOperatorLossEvent(ctx context.Context, input models.MachineOperatorLossEventFinishRequest) (models.MachineOperatorLossEventFinishResponse, error) {
	input.UUID = strings.TrimSpace(input.UUID)

	if input.UUID == "" {
		return models.MachineOperatorLossEventFinishResponse{}, fmt.Errorf("uuid wajib diisi")
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return models.MachineOperatorLossEventFinishResponse{}, err
	}
	defer tx.Rollback()

	session, err := GetActiveMachineOperatorByTx(ctx, tx, input.UUID)
	if err != nil {
		return models.MachineOperatorLossEventFinishResponse{}, err
	}

	activeEvent, err := r.getActiveLossEventBySessionTx(ctx, tx, session.ID)
	if err != nil {
		return models.MachineOperatorLossEventFinishResponse{}, err
	}

	if activeEvent == nil {
		return models.MachineOperatorLossEventFinishResponse{}, ErrMachineOperatorLossEventNotFound
	}

	updateQuery := `
		UPDATE dbo.machine_operator_loss_events
		SET
			end_time = SYSDATETIME(),
			status = 'CLOSED',
			updated_at = SYSDATETIME()
		WHERE id = @id
		  AND session_id = @session_id
		  AND status = 'ACTIVE'
		  AND end_time IS NULL
	`

	result, err := tx.ExecContext(
		ctx,
		updateQuery,
		sql.Named("id", activeEvent.ID),
		sql.Named("session_id", session.ID),
	)
	if err != nil {
		return models.MachineOperatorLossEventFinishResponse{}, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return models.MachineOperatorLossEventFinishResponse{}, err
	}

	if affected == 0 {
		return models.MachineOperatorLossEventFinishResponse{}, ErrMachineOperatorLossEventNotFound
	}

	event, err := getMachineOperatorLossEventByIDTx(ctx, tx, activeEvent.ID)
	if err != nil {
		return models.MachineOperatorLossEventFinishResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.MachineOperatorLossEventFinishResponse{}, err
	}

	return models.MachineOperatorLossEventFinishResponse{
		Status:  "ok",
		Message: "Loss event berhasil diselesaikan",
		Event:   event,
	}, nil
}

func (r *Repository) GetActiveMachineOperatorLossEvent(ctx context.Context, uuid string) (models.MachineOperatorLossEventActiveResponse, error) {
	uuid = strings.TrimSpace(uuid)

	if uuid == "" {
		return models.MachineOperatorLossEventActiveResponse{}, fmt.Errorf("uuid wajib diisi")
	}

	query := `
		SELECT TOP 1
			id,
			session_id,

			-- DB tidak punya kolom session_date, jadi ambil dari start_time
			ISNULL(CONVERT(VARCHAR(10), start_time, 120), '') AS session_date,

			ISNULL(uuid, '') AS uuid,

			-- DB tidak punya kolom machine_name dan location
			'' AS machine_name,
			'' AS location,

			ISNULL(operator_nik, '') AS operator_nik,
			ISNULL(operator_name, '') AS operator_name,
			ISNULL(reason_code, '') AS reason_code,
			ISNULL(reason_label, '') AS reason_label,
			ISNULL(note, '') AS note,
			ISNULL(CONVERT(VARCHAR(19), start_time, 120), '') AS start_time,
			ISNULL(CONVERT(VARCHAR(19), end_time, 120), '') AS end_time,

			-- Untuk event ACTIVE, durasi dihitung realtime dari start_time sampai sekarang.
			CAST(
				CASE
					WHEN status = 'ACTIVE' AND end_time IS NULL
						THEN DATEDIFF_BIG(SECOND, start_time, SYSDATETIME())
					ELSE ISNULL(duration_sec, 0)
				END AS BIGINT
			) AS duration_seconds,

			ISNULL(status, '') AS status,
			ISNULL(CONVERT(VARCHAR(19), created_at, 120), '') AS created_at,
			ISNULL(CONVERT(VARCHAR(19), updated_at, 120), '') AS updated_at
		FROM dbo.machine_operator_loss_events
		WHERE LOWER(LTRIM(RTRIM(uuid))) = LOWER(LTRIM(RTRIM(@uuid)))
		  AND status = 'ACTIVE'
		  AND end_time IS NULL
		ORDER BY start_time DESC, id DESC
	`

	event, err := scanMachineOperatorLossEvent(r.DB.QueryRowContext(
		ctx,
		query,
		sql.Named("uuid", uuid),
	))

	if err != nil {
		if err == sql.ErrNoRows {
			return models.MachineOperatorLossEventActiveResponse{
				Status:         "ok",
				HasActiveEvent: false,
				Event:          nil,
			}, nil
		}

		return models.MachineOperatorLossEventActiveResponse{}, err
	}

	return models.MachineOperatorLossEventActiveResponse{
		Status:         "ok",
		HasActiveEvent: true,
		Event:          &event,
	}, nil
}
