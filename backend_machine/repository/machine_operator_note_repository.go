package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"backend_machine/models"
)

func shouldCloseMachineOperatorSession(reasonCode string) bool {
	code := strings.ToUpper(strings.TrimSpace(reasonCode))

	switch code {
	case "LOGOUT", "SELESAI", "SELESAI_PAKAI_MESIN", "FINISH", "END":
		return true
	default:
		return false
	}
}

func (r *Repository) CreateMachineOperatorNote(ctx context.Context, input models.MachineOperatorNoteRequest) (models.MachineOperatorNoteResponse, error) {
	input.UUID = strings.TrimSpace(input.UUID)
	input.OperatorNIK = strings.TrimSpace(input.OperatorNIK)
	input.ReasonCode = strings.TrimSpace(input.ReasonCode)
	input.ReasonName = strings.TrimSpace(input.ReasonName)
	input.Note = strings.TrimSpace(input.Note)

	if input.UUID == "" {
		return models.MachineOperatorNoteResponse{}, fmt.Errorf("uuid wajib diisi")
	}

	if input.ReasonCode == "" {
		return models.MachineOperatorNoteResponse{}, fmt.Errorf("reasonCode wajib diisi")
	}

	if input.ReasonName == "" {
		input.ReasonName = input.ReasonCode
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return models.MachineOperatorNoteResponse{}, err
	}
	defer tx.Rollback()

	session, err := GetActiveMachineOperatorByTx(ctx, tx, input.UUID)
	if err != nil {
		return models.MachineOperatorNoteResponse{}, err
	}

	if input.OperatorNIK != "" && !strings.EqualFold(strings.TrimSpace(session.OperatorNIK), input.OperatorNIK) {
		return models.MachineOperatorNoteResponse{}, fmt.Errorf("operatorNik tidak sama dengan session aktif")
	}

	insertNoteQuery := `
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
		OUTPUT INSERTED.id
		VALUES (
			@session_id,
			CAST(@session_date AS DATE),
			@uuid,
			@operator_nik,
			@operator_name,
			@reason_code,
			@reason_name,
			@note,
			SYSDATETIME()
		)
	`

	var noteID int64
	err = tx.QueryRowContext(
		ctx,
		insertNoteQuery,
		sql.Named("session_id", session.ID),
		sql.Named("session_date", session.SessionDate),
		sql.Named("uuid", session.UUID),
		sql.Named("operator_nik", session.OperatorNIK),
		sql.Named("operator_name", session.OperatorName),
		sql.Named("reason_code", input.ReasonCode),
		sql.Named("reason_name", input.ReasonName),
		sql.Named("note", input.Note),
	).Scan(&noteID)

	if err != nil {
		return models.MachineOperatorNoteResponse{}, err
	}

	closed := shouldCloseMachineOperatorSession(input.ReasonCode)
	if closed {
		closeQuery := `
			UPDATE dbo.machine_operator_sessions
			SET
				logout_time = SYSDATETIME(),
				status = 'LOGOUT',
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
		`

		if _, err := tx.ExecContext(ctx, closeQuery, sql.Named("id", session.ID)); err != nil {
			return models.MachineOperatorNoteResponse{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return models.MachineOperatorNoteResponse{}, err
	}

	message := "Catatan operator berhasil disimpan"
	if closed {
		message = "Catatan operator berhasil disimpan dan session ditutup"
	}

	return models.MachineOperatorNoteResponse{
		Status:    "ok",
		Message:   message,
		NoteID:    noteID,
		SessionID: session.ID,
		Closed:    closed,
	}, nil
}
