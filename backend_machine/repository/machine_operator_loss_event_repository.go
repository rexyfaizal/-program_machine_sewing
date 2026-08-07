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

func getMachineOperatorLossEventByIDTx(ctx context.Context, tx *sql.Tx, id int64) (models.MachineOperatorLossEvent, error) {
	query := `
		SELECT TOP 1
			id,
			session_id,

			ISNULL(CONVERT(VARCHAR(10), start_time, 120), '') AS session_date,
			ISNULL(uuid, '') AS uuid,

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

func (r *Repository) getActiveLossEventByUUIDTx(ctx context.Context, tx *sql.Tx, uuid string) (*models.MachineOperatorLossEvent, error) {
	uuid = strings.TrimSpace(uuid)

	query := `
		SELECT TOP 1
			id,
			session_id,

			ISNULL(CONVERT(VARCHAR(10), start_time, 120), '') AS session_date,
			ISNULL(uuid, '') AS uuid,

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
		FROM dbo.machine_operator_loss_events WITH (UPDLOCK, HOLDLOCK)
		WHERE LOWER(LTRIM(RTRIM(uuid))) = LOWER(LTRIM(RTRIM(@uuid)))
		  AND status = 'ACTIVE'
		  AND end_time IS NULL
		ORDER BY start_time DESC, id DESC
	`

	event, err := scanMachineOperatorLossEvent(tx.QueryRowContext(
		ctx,
		query,
		sql.Named("uuid", uuid),
	))

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &event, nil
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

	// Cek berdasarkan UUID mesin, bukan hanya session_id.
	// Tujuannya supaya tidak membuat loss event dobel pada mesin yang sama.
	activeEvent, err := r.getActiveLossEventByUUIDTx(ctx, tx, input.UUID)
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

	// =====================================================
	// FINISH BERDASARKAN UUID MESIN
	// Bukan berdasarkan session_id operator aktif terbaru.
	//
	// Ini menyelesaikan kasus:
	// - event ACTIVE masih ada di mesin
	// - tetapi session_id event berbeda dengan session operator terbaru
	// =====================================================
	closeQuery := `
		DECLARE @event_id BIGINT;

		SELECT TOP 1
			@event_id = id
		FROM dbo.machine_operator_loss_events WITH (UPDLOCK, HOLDLOCK)
		WHERE LOWER(LTRIM(RTRIM(uuid))) = LOWER(LTRIM(RTRIM(@uuid)))
		  AND status = 'ACTIVE'
		  AND end_time IS NULL
		ORDER BY start_time DESC, id DESC;

		IF @event_id IS NULL
		BEGIN
			SELECT CAST(0 AS BIGINT) AS closed_id;
			RETURN;
		END;

		UPDATE dbo.machine_operator_loss_events
		SET
			end_time = SYSDATETIME(),
			status = 'CLOSED',
			updated_at = SYSDATETIME()
		WHERE id = @event_id
		  AND status = 'ACTIVE'
		  AND end_time IS NULL;

		SELECT @event_id AS closed_id;
	`

	var closedID int64

	err = tx.QueryRowContext(
		ctx,
		closeQuery,
		sql.Named("uuid", input.UUID),
	).Scan(&closedID)

	if err != nil {
		return models.MachineOperatorLossEventFinishResponse{}, err
	}

	if closedID == 0 {
		// Idempotent: event mungkin sudah ditutup mekanik / proses lain.
		return models.MachineOperatorLossEventFinishResponse{
			Status:  "ok",
			Message: "Loss event sudah tidak aktif / sudah selesai",
			Event: models.MachineOperatorLossEvent{
				UUID:   input.UUID,
				Status: "CLOSED",
			},
		}, nil
	}

	event, err := getMachineOperatorLossEventByIDTx(ctx, tx, closedID)
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

			ISNULL(CONVERT(VARCHAR(10), start_time, 120), '') AS session_date,
			ISNULL(uuid, '') AS uuid,

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
