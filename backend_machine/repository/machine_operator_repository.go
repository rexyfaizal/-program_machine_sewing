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
		if strings.EqualFold(strings.TrimSpace(active.OperatorNIK), input.OperatorNIK) {
			if err := tx.Commit(); err != nil {
				return models.MachineOperatorLoginResponse{}, err
			}

			return models.MachineOperatorLoginResponse{
				Status:     "active",
				Message:    "Operator sudah aktif di mesin ini",
				IsExisting: true,
				Session:    active,
			}, nil
		}

		closeOldQuery := `
			UPDATE dbo.machine_operator_sessions
			SET
				logout_time = SYSDATETIME(),
				status = 'LOGOUT',
				updated_at = SYSDATETIME()
			WHERE id = @id
		`

		if _, err := tx.ExecContext(ctx, closeOldQuery, sql.Named("id", active.ID)); err != nil {
			return models.MachineOperatorLoginResponse{}, err
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

	session, err := scanMachineOperatorSession(tx.QueryRowContext(
		ctx,
		activeQuery,
		sql.Named("uuid", input.UUID),
	))

	if err != nil {
		if err == sql.ErrNoRows {
			return models.MachineOperatorNoteResponse{}, ErrMachineOperatorNotFound
		}
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

func (r *Repository) GetMachineOperatorReport(ctx context.Context, date string) ([]models.MachineOperatorReportItem, error) {
	date = strings.TrimSpace(date)
	if date == "" {
		return nil, fmt.Errorf("date wajib diisi")
	}

	sessionQuery := `
		SELECT
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
		WHERE session_date = CAST(@date AS DATE)
		ORDER BY login_time ASC, id ASC
	`

	rows, err := r.DB.QueryContext(ctx, sessionQuery, sql.Named("date", date))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report := make([]models.MachineOperatorReportItem, 0)

	for rows.Next() {
		session, err := scanMachineOperatorSession(rows)
		if err != nil {
			return nil, err
		}

		report = append(report, models.MachineOperatorReportItem{
			MachineOperatorSession: session,
			Notes:                  make([]models.MachineOperatorNote, 0),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	noteQuery := `
		SELECT
			id,
			session_id,
			CONVERT(VARCHAR(10), session_date, 120) AS session_date,
			ISNULL(uuid, '') AS uuid,
			ISNULL(operator_nik, '') AS operator_nik,
			ISNULL(operator_name, '') AS operator_name,
			ISNULL(reason_code, '') AS reason_code,
			ISNULL(reason_name, '') AS reason_name,
			ISNULL(note, '') AS note,
			ISNULL(CONVERT(VARCHAR(19), created_at, 120), '') AS created_at
		FROM dbo.machine_operator_notes
		WHERE session_date = CAST(@date AS DATE)
		ORDER BY created_at ASC, id ASC
	`

	noteRows, err := r.DB.QueryContext(ctx, noteQuery, sql.Named("date", date))
	if err != nil {
		return nil, err
	}
	defer noteRows.Close()

	noteMap := make(map[int64][]models.MachineOperatorNote)

	for noteRows.Next() {
		var note models.MachineOperatorNote

		err := noteRows.Scan(
			&note.ID,
			&note.SessionID,
			&note.SessionDate,
			&note.UUID,
			&note.OperatorNIK,
			&note.OperatorName,
			&note.ReasonCode,
			&note.ReasonName,
			&note.Note,
			&note.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		noteMap[note.SessionID] = append(noteMap[note.SessionID], note)
	}

	if err := noteRows.Err(); err != nil {
		return nil, err
	}

	for i := range report {
		report[i].Notes = noteMap[report[i].ID]
		if report[i].Notes == nil {
			report[i].Notes = make([]models.MachineOperatorNote, 0)
		}
	}

	return report, nil
}
