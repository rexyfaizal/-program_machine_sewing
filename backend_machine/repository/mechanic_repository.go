package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/utils"
)

var (
	ErrMechanicNotFound       = errors.New("mekanik tidak ditemukan")
	ErrMechanicNotAuthorized  = errors.New("NIK bukan bagian Mekanik")
	ErrMechanicTicketNotFound = errors.New("tiket mesin rusak tidak ditemukan")
	ErrMechanicAlreadyClaimed = errors.New("tiket sudah diambil mekanik lain")
	ErrMechanicNotClaimed     = errors.New("tiket belum diambil")
	ErrMechanicClaimRequired  = errors.New("ambil tiket dulu sebelum selesai")
	ErrMechanicAlreadyDone    = errors.New("tiket mekanik sudah selesai")
	ErrMechanicRFIDTaken      = errors.New("kartu RFID sudah terdaftar ke NIK lain")
)

func (r *Repository) EnsureMechanicClaimSchema(ctx context.Context) error {
	statements := []string{
		`
IF COL_LENGTH('dbo.machine_operator_loss_events', 'claimed_by_nik') IS NULL
BEGIN
	ALTER TABLE dbo.machine_operator_loss_events
	ADD claimed_by_nik NVARCHAR(50) NULL;
END
`,
		`
IF COL_LENGTH('dbo.machine_operator_loss_events', 'claimed_by_name') IS NULL
BEGIN
	ALTER TABLE dbo.machine_operator_loss_events
	ADD claimed_by_name NVARCHAR(255) NULL;
END
`,
		`
IF COL_LENGTH('dbo.machine_operator_loss_events', 'claimed_at') IS NULL
BEGIN
	ALTER TABLE dbo.machine_operator_loss_events
	ADD claimed_at DATETIME2 NULL;
END
`,
		`
IF COL_LENGTH('dbo.machine_operator_loss_events', 'mechanic_done_at') IS NULL
BEGIN
	ALTER TABLE dbo.machine_operator_loss_events
	ADD mechanic_done_at DATETIME2 NULL;
END
`,
		`
IF COL_LENGTH('dbo.machine_operator_loss_events', 'mechanic_done_by_nik') IS NULL
BEGIN
	ALTER TABLE dbo.machine_operator_loss_events
	ADD mechanic_done_by_nik NVARCHAR(50) NULL;
END
`,
		`
IF COL_LENGTH('dbo.machine_operator_loss_events', 'mechanic_done_by_name') IS NULL
BEGIN
	ALTER TABLE dbo.machine_operator_loss_events
	ADD mechanic_done_by_name NVARCHAR(255) NULL;
END
`,
	}

	for _, query := range statements {
		if _, err := r.DB.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	log.Println("Schema mechanic claim/done columns siap.")
	return nil
}

func isMechanicBagian(bagian string) bool {
	text := strings.ToUpper(strings.TrimSpace(bagian))
	text = strings.ReplaceAll(text, "_", " ")
	text = strings.ReplaceAll(text, "-", " ")
	return strings.Contains(text, "MEKANIK")
}

func (r *Repository) IdentifyMechanic(ctx context.Context, code string) (models.MechanicIdentifyResponse, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return models.MechanicIdentifyResponse{}, fmt.Errorf("nik/rfid wajib diisi")
	}

	// 1) Coba cocokkan sebagai NIK
	byNik, err := r.findEmployeeMechanicByField(ctx, "nik", code)
	if err != nil {
		return models.MechanicIdentifyResponse{}, err
	}
	if byNik != nil {
		return *byNik, nil
	}

	// 2) Coba cocokkan sebagai RFID
	byRFID, err := r.findEmployeeMechanicByField(ctx, "rfid_no", code)
	if err != nil {
		return models.MechanicIdentifyResponse{}, err
	}
	if byRFID != nil {
		return *byRFID, nil
	}

	return models.MechanicIdentifyResponse{
		Status:  "not_found",
		Message: "NIK / kartu RFID tidak ditemukan",
		IsValid: false,
	}, nil
}

func (r *Repository) findEmployeeMechanicByField(
	ctx context.Context,
	field string,
	value string,
) (*models.MechanicIdentifyResponse, error) {
	column := "nik"
	if field == "rfid_no" {
		column = "rfid_no"
	}

	query := fmt.Sprintf(`
		SELECT TOP 1
			LTRIM(RTRIM(ISNULL(CONVERT(VARCHAR(255), nik), ''))) AS nik,
			LTRIM(RTRIM(ISNULL(CONVERT(VARCHAR(255), name), ''))) AS name,
			LTRIM(RTRIM(ISNULL(CONVERT(VARCHAR(255), bagian), ''))) AS bagian,
			LTRIM(RTRIM(ISNULL(CONVERT(VARCHAR(255), rfid_no), ''))) AS rfid_no
		FROM dbo.employee
		WHERE LTRIM(RTRIM(CONVERT(VARCHAR(255), %s))) = @value
	`, column)

	var item models.MechanicIdentifyResponse
	err := r.DB.QueryRowContext(ctx, query, sql.Named("value", value)).Scan(
		&item.NIK,
		&item.Name,
		&item.Bagian,
		&item.RFID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if !isMechanicBagian(item.Bagian) {
		return &models.MechanicIdentifyResponse{
			Status:  "forbidden",
			Message: "NIK/kartu ditemukan, tetapi bagian bukan Mekanik",
			IsValid: false,
			NIK:     item.NIK,
			Name:    item.Name,
			Bagian: item.Bagian,
			RFID:    item.RFID,
		}, nil
	}

	item.Status = "ok"
	item.Message = "Mekanik terverifikasi"
	item.IsValid = true
	return &item, nil
}

func (r *Repository) RegisterMechanicRFID(
	ctx context.Context,
	input models.MechanicRFIDRegisterRequest,
) (models.MechanicRFIDRegisterResponse, error) {
	input.NIK = strings.TrimSpace(input.NIK)
	input.RFIDNo = strings.TrimSpace(input.RFIDNo)

	if input.NIK == "" {
		return models.MechanicRFIDRegisterResponse{}, fmt.Errorf("nik wajib diisi")
	}
	if input.RFIDNo == "" {
		return models.MechanicRFIDRegisterResponse{}, fmt.Errorf("rfidNo wajib diisi")
	}

	identity, err := r.findEmployeeMechanicByField(ctx, "nik", input.NIK)
	if err != nil {
		return models.MechanicRFIDRegisterResponse{}, err
	}
	if identity == nil {
		return models.MechanicRFIDRegisterResponse{}, ErrMechanicNotFound
	}
	if !identity.IsValid {
		return models.MechanicRFIDRegisterResponse{}, ErrMechanicNotAuthorized
	}

	// Satu kartu = satu NIK (unik)
	var otherNik string
	err = r.DB.QueryRowContext(
		ctx,
		`
		SELECT TOP 1
			LTRIM(RTRIM(ISNULL(CONVERT(VARCHAR(255), nik), ''))) AS nik
		FROM dbo.employee
		WHERE LTRIM(RTRIM(CONVERT(VARCHAR(255), rfid_no))) = @rfid
		  AND LTRIM(RTRIM(CONVERT(VARCHAR(255), nik))) <> @nik
		`,
		sql.Named("rfid", input.RFIDNo),
		sql.Named("nik", input.NIK),
	).Scan(&otherNik)

	if err == nil && strings.TrimSpace(otherNik) != "" {
		return models.MechanicRFIDRegisterResponse{}, ErrMechanicRFIDTaken
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return models.MechanicRFIDRegisterResponse{}, err
	}

	_, err = r.DB.ExecContext(
		ctx,
		`
		UPDATE dbo.employee
		SET rfid_no = @rfid
		WHERE LTRIM(RTRIM(CONVERT(VARCHAR(255), nik))) = @nik
		`,
		sql.Named("rfid", input.RFIDNo),
		sql.Named("nik", input.NIK),
	)
	if err != nil {
		return models.MechanicRFIDRegisterResponse{}, err
	}

	return models.MechanicRFIDRegisterResponse{
		Status:  "ok",
		Message: "Kartu RFID berhasil didaftarkan",
		NIK:     identity.NIK,
		Name:    identity.Name,
		RFIDNo:  input.RFIDNo,
	}, nil
}

func parseExportLikeTime(value string) *time.Time {
	text := strings.TrimSpace(value)
	if text == "" {
		return nil
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, text, time.Local); err == nil {
			return &t
		}
	}

	return nil
}

func secondsBetween(startText, endText string, fallbackEnd time.Time) int64 {
	start := parseExportLikeTime(startText)
	if start == nil {
		return 0
	}

	end := parseExportLikeTime(endText)
	endTime := fallbackEnd
	if end != nil {
		endTime = *end
	}

	sec := int64(endTime.Sub(*start).Seconds())
	if sec < 0 {
		return 0
	}
	return sec
}

func enrichMechanicTicket(item *models.MechanicBrokenMachine) {
	now := time.Now()

	item.ClosedByMechanic = strings.TrimSpace(item.MechanicDoneAt) != ""
	item.ClosedByOperator = strings.TrimSpace(item.EndTime) != ""
	item.OperatorStillActive = !item.ClosedByOperator

	item.WaitMechanicSeconds = 0
	item.MechanicWorkSeconds = 0
	item.OperatorLossSeconds = secondsBetween(item.StartTime, item.EndTime, now)

	if strings.TrimSpace(item.ClaimedAt) != "" {
		item.WaitMechanicSeconds = secondsBetween(item.StartTime, item.ClaimedAt, now)

		if item.ClosedByMechanic {
			item.MechanicWorkSeconds = secondsBetween(item.ClaimedAt, item.MechanicDoneAt, now)
		} else {
			item.MechanicWorkSeconds = secondsBetween(item.ClaimedAt, "", now)
		}
	}

	// Status tiket mekanik (independen dari close operator).
	if item.ClosedByMechanic {
		item.TicketStatus = "DONE"
	} else if strings.TrimSpace(item.ClaimedByNIK) != "" {
		item.TicketStatus = "IN_PROGRESS"
	} else {
		item.TicketStatus = "OPEN"
	}

	item.DurationSeconds = item.OperatorLossSeconds
}

func scanMechanicBrokenMachine(s rowScanner) (models.MechanicBrokenMachine, error) {
	var item models.MechanicBrokenMachine
	var claimedByNIK, claimedByName, claimedAt sql.NullString
	var mechanicDoneAt, mechanicDoneByNIK, mechanicDoneByName sql.NullString

	err := s.Scan(
		&item.ID,
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
		&claimedByNIK,
		&claimedByName,
		&claimedAt,
		&mechanicDoneAt,
		&mechanicDoneByNIK,
		&mechanicDoneByName,
	)
	if err != nil {
		return item, err
	}

	item.ClaimedByNIK = strings.TrimSpace(claimedByNIK.String)
	item.ClaimedByName = strings.TrimSpace(claimedByName.String)
	item.ClaimedAt = strings.TrimSpace(claimedAt.String)
	item.EndTime = strings.TrimSpace(item.EndTime)
	item.MechanicDoneAt = strings.TrimSpace(mechanicDoneAt.String)
	item.MechanicDoneByNIK = strings.TrimSpace(mechanicDoneByNIK.String)
	item.MechanicDoneByName = strings.TrimSpace(mechanicDoneByName.String)

	enrichMechanicTicket(&item)
	return item, nil
}

func mechanicBrokenSelectSQL() string {
	return `
		SELECT
			n.id,
			LTRIM(RTRIM(ISNULL(n.uuid, ''))) AS uuid,
			LTRIM(RTRIM(ISNULL(ms.custom_name, n.uuid))) AS machine_name,
			LTRIM(RTRIM(ISNULL(ms.location, ''))) AS location,
			LTRIM(RTRIM(ISNULL(CONVERT(VARCHAR(255), n.operator_nik), ''))) AS operator_nik,
			LTRIM(RTRIM(ISNULL(n.operator_name, ''))) AS operator_name,
			LTRIM(RTRIM(ISNULL(n.reason_code, ''))) AS reason_code,
			LTRIM(RTRIM(ISNULL(n.reason_label, 'Mesin Rusak'))) AS reason_label,
			LTRIM(RTRIM(ISNULL(n.note, ''))) AS note,
			ISNULL(CONVERT(VARCHAR(19), n.start_time, 120), '') AS start_time,
			ISNULL(CONVERT(VARCHAR(19), n.end_time, 120), '') AS end_time,
			CAST(
				DATEDIFF_BIG(
					SECOND,
					COALESCE(n.start_time, n.created_at, SYSDATETIME()),
					COALESCE(n.end_time, SYSDATETIME())
				) AS BIGINT
			) AS duration_seconds,
			LTRIM(RTRIM(ISNULL(n.claimed_by_nik, ''))) AS claimed_by_nik,
			LTRIM(RTRIM(ISNULL(n.claimed_by_name, ''))) AS claimed_by_name,
			ISNULL(CONVERT(VARCHAR(19), n.claimed_at, 120), '') AS claimed_at,
			ISNULL(CONVERT(VARCHAR(19), n.mechanic_done_at, 120), '') AS mechanic_done_at,
			LTRIM(RTRIM(ISNULL(n.mechanic_done_by_nik, ''))) AS mechanic_done_by_nik,
			LTRIM(RTRIM(ISNULL(n.mechanic_done_by_name, ''))) AS mechanic_done_by_name
		FROM dbo.machine_operator_loss_events n
		OUTER APPLY (
			SELECT TOP 1
				custom_name,
				location
			FROM dbo.machine_setting_manual ms
			WHERE LOWER(LTRIM(RTRIM(ms.uuid))) = LOWER(LTRIM(RTRIM(n.uuid)))
			ORDER BY ms.updated_at DESC
		) ms
	`
}

func mechanicBrokenReasonWhere() string {
	return `
		(
			UPPER(LTRIM(RTRIM(ISNULL(n.reason_code, '')))) = 'MACHINE_BROKEN'
			OR UPPER(LTRIM(RTRIM(ISNULL(n.reason_label, '')))) LIKE '%MESIN RUSAK%'
		)
	`
}

func (r *Repository) countMechanicDoneToday(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(1)
		FROM dbo.machine_operator_loss_events n
		WHERE ` + mechanicBrokenReasonWhere() + `
		  AND n.mechanic_done_at IS NOT NULL
		  AND CAST(n.mechanic_done_at AS DATE) = CAST(SYSDATETIME() AS DATE)
	`

	var total int
	err := r.DB.QueryRowContext(ctx, query).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func mechanicActiveBoardWhere() string {
	return `
		(
			(
				LTRIM(RTRIM(ISNULL(n.claimed_by_nik, ''))) <> ''
				AND COALESCE(n.claimed_at, n.start_time, n.created_at) >= @windowStart
				AND COALESCE(n.claimed_at, n.start_time, n.created_at) < @windowEnd
			)
			OR (
				n.end_time IS NULL
				AND COALESCE(n.start_time, n.created_at) >= @windowStart
				AND COALESCE(n.start_time, n.created_at) < @windowEnd
			)
			OR (
				n.end_time IS NOT NULL
				AND n.end_time >= @windowStart
				AND n.end_time < @windowEnd
			)
		)
	`
}

func (r *Repository) ListMechanicBrokenMachines(
	ctx context.Context,
	statusFilter string,
	locationFilter string,
) (models.MechanicBrokenListResponse, error) {
	statusFilter = strings.ToUpper(strings.TrimSpace(statusFilter))
	locationFilter = strings.TrimSpace(locationFilter)

	doneToday, err := r.countMechanicDoneToday(ctx)
	if err != nil {
		return models.MechanicBrokenListResponse{}, err
	}

	args := []any{}
	query := mechanicBrokenSelectSQL()

	windowStart, windowEnd := utils.ResolveGM3WorkWindow(time.Now())

	isHistory := statusFilter == "DONE" ||
		statusFilter == "HISTORY" ||
		statusFilter == "SELESAI" ||
		statusFilter == "HISTORI"

	if isHistory {
		// Histori: hanya tiket yang sudah diselesaikan mekanik hari ini.
		query += `
		WHERE ` + mechanicBrokenReasonWhere() + `
		  AND n.mechanic_done_at IS NOT NULL
		  AND CAST(n.mechanic_done_at AS DATE) = CAST(SYSDATETIME() AS DATE)
		`
	} else {
		// Board aktif: hari kerja GM3 berjalan.
		// Tetap tampil meski operator sudah close duluan di hari yang sama.
		query += `
		WHERE ` + mechanicBrokenReasonWhere() + `
		  AND n.mechanic_done_at IS NULL
		  AND ` + mechanicActiveBoardWhere() + `
		`
		args = append(args,
			sql.Named("windowStart", windowStart),
			sql.Named("windowEnd", windowEnd),
		)
	}

	if locationFilter != "" && !strings.EqualFold(locationFilter, "ALL") {
		query += `
		  AND UPPER(LTRIM(RTRIM(ISNULL(ms.location, '')))) LIKE @location
		`
		args = append(args, sql.Named("location", "%"+strings.ToUpper(locationFilter)+"%"))
	}

	if !isHistory {
		switch statusFilter {
		case "OPEN":
			query += `
		  AND (
			n.claimed_by_nik IS NULL
			OR LTRIM(RTRIM(n.claimed_by_nik)) = ''
		  )
		`
		case "IN_PROGRESS", "BUSY":
			query += `
		  AND LTRIM(RTRIM(ISNULL(n.claimed_by_nik, ''))) <> ''
		`
		}
	}

	if isHistory {
		query += `
		ORDER BY
			COALESCE(n.mechanic_done_at, n.end_time) DESC,
			n.id DESC
	`
	} else {
		query += `
		ORDER BY
			CASE
				WHEN n.claimed_by_nik IS NULL OR LTRIM(RTRIM(n.claimed_by_nik)) = '' THEN 0
				ELSE 1
			END,
			n.start_time ASC,
			n.id ASC
	`
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return models.MechanicBrokenListResponse{}, err
	}
	defer rows.Close()

	list := make([]models.MechanicBrokenMachine, 0)
	openCount := 0
	busyCount := 0

	for rows.Next() {
		item, scanErr := scanMechanicBrokenMachine(rows)
		if scanErr != nil {
			return models.MechanicBrokenListResponse{}, scanErr
		}

		switch item.TicketStatus {
		case "OPEN":
			openCount++
		case "IN_PROGRESS":
			busyCount++
		}

		list = append(list, item)
	}

	if err := rows.Err(); err != nil {
		return models.MechanicBrokenListResponse{}, err
	}

	if isHistory {
		return models.MechanicBrokenListResponse{
			Status:    "ok",
			Total:     len(list),
			Open:      0,
			Busy:      0,
			DoneToday: doneToday,
			Rows:      list,
		}, nil
	}

	return models.MechanicBrokenListResponse{
		Status:    "ok",
		Total:     len(list),
		Open:      openCount,
		Busy:      busyCount,
		DoneToday: doneToday,
		Rows:      list,
	}, nil
}

func (r *Repository) getMechanicBrokenByIDTx(
	ctx context.Context,
	tx *sql.Tx,
	id int64,
) (models.MechanicBrokenMachine, error) {
	query := mechanicBrokenSelectSQL() + `
		WHERE n.id = @id
		  AND ` + mechanicBrokenReasonWhere() + `
	`

	item, err := scanMechanicBrokenMachine(tx.QueryRowContext(
		ctx,
		query,
		sql.Named("id", id),
	))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.MechanicBrokenMachine{}, ErrMechanicTicketNotFound
		}
		return models.MechanicBrokenMachine{}, err
	}

	return item, nil
}

func (r *Repository) ClaimMechanicBrokenMachine(
	ctx context.Context,
	input models.MechanicClaimRequest,
) (models.MechanicActionResponse, error) {
	input.MechanicNIK = strings.TrimSpace(input.MechanicNIK)
	input.MechanicName = strings.TrimSpace(input.MechanicName)

	if input.ID <= 0 {
		return models.MechanicActionResponse{}, fmt.Errorf("id wajib diisi")
	}
	if input.MechanicNIK == "" {
		return models.MechanicActionResponse{}, fmt.Errorf("mechanicNik wajib diisi")
	}

	identity, err := r.IdentifyMechanic(ctx, input.MechanicNIK)
	if err != nil {
		return models.MechanicActionResponse{}, err
	}
	if !identity.IsValid {
		if identity.Status == "not_found" {
			return models.MechanicActionResponse{}, ErrMechanicNotFound
		}
		return models.MechanicActionResponse{}, ErrMechanicNotAuthorized
	}

	if input.MechanicName == "" {
		input.MechanicName = identity.Name
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return models.MechanicActionResponse{}, err
	}
	defer tx.Rollback()

	lockQuery := `
		SELECT TOP 1
			id,
			LTRIM(RTRIM(ISNULL(claimed_by_nik, ''))) AS claimed_by_nik,
			CASE WHEN mechanic_done_at IS NULL THEN 0 ELSE 1 END AS is_done
		FROM dbo.machine_operator_loss_events WITH (UPDLOCK, HOLDLOCK)
		WHERE id = @id
		  AND (
				UPPER(LTRIM(RTRIM(ISNULL(reason_code, '')))) = 'MACHINE_BROKEN'
				OR UPPER(LTRIM(RTRIM(ISNULL(reason_label, '')))) LIKE '%MESIN RUSAK%'
			)
		  AND mechanic_done_at IS NULL
	`

	var lockedID int64
	var claimedBy string
	var isDone int
	err = tx.QueryRowContext(ctx, lockQuery, sql.Named("id", input.ID)).Scan(
		&lockedID,
		&claimedBy,
		&isDone,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.MechanicActionResponse{}, ErrMechanicTicketNotFound
		}
		return models.MechanicActionResponse{}, err
	}

	if isDone == 1 {
		return models.MechanicActionResponse{}, ErrMechanicAlreadyDone
	}

	claimedBy = strings.TrimSpace(claimedBy)
	if claimedBy != "" && !strings.EqualFold(claimedBy, input.MechanicNIK) {
		return models.MechanicActionResponse{}, ErrMechanicAlreadyClaimed
	}

	_, err = tx.ExecContext(
		ctx,
		`
		UPDATE dbo.machine_operator_loss_events
		SET
			claimed_by_nik = @nik,
			claimed_by_name = @name,
			claimed_at = COALESCE(claimed_at, SYSDATETIME()),
			updated_at = SYSDATETIME()
		WHERE id = @id
		`,
		sql.Named("id", input.ID),
		sql.Named("nik", input.MechanicNIK),
		sql.Named("name", strings.ToUpper(input.MechanicName)),
	)
	if err != nil {
		return models.MechanicActionResponse{}, err
	}

	item, err := r.getMechanicBrokenByIDTx(ctx, tx, input.ID)
	if err != nil {
		return models.MechanicActionResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.MechanicActionResponse{}, err
	}

	return models.MechanicActionResponse{
		Status:  "ok",
		Message: "Tiket berhasil diambil",
		Item:    &item,
	}, nil
}

func (r *Repository) DoneMechanicBrokenMachine(
	ctx context.Context,
	input models.MechanicDoneRequest,
) (models.MechanicActionResponse, error) {
	input.MechanicNIK = strings.TrimSpace(input.MechanicNIK)

	if input.ID <= 0 {
		return models.MechanicActionResponse{}, fmt.Errorf("id wajib diisi")
	}
	if input.MechanicNIK == "" {
		return models.MechanicActionResponse{}, fmt.Errorf("mechanicNik wajib diisi")
	}

	identity, err := r.IdentifyMechanic(ctx, input.MechanicNIK)
	if err != nil {
		return models.MechanicActionResponse{}, err
	}
	if !identity.IsValid {
		if identity.Status == "not_found" {
			return models.MechanicActionResponse{}, ErrMechanicNotFound
		}
		return models.MechanicActionResponse{}, ErrMechanicNotAuthorized
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return models.MechanicActionResponse{}, err
	}
	defer tx.Rollback()

	// Selesai mekanik TIDAK menutup loss operator.
	// Boleh dilakukan meski operator sudah close terlebih dulu.
	lockQuery := `
		SELECT TOP 1
			id,
			LTRIM(RTRIM(ISNULL(uuid, ''))) AS uuid,
			LTRIM(RTRIM(ISNULL(claimed_by_nik, ''))) AS claimed_by_nik,
			CASE WHEN mechanic_done_at IS NULL THEN 0 ELSE 1 END AS is_done
		FROM dbo.machine_operator_loss_events WITH (UPDLOCK, HOLDLOCK)
		WHERE id = @id
		  AND (
				UPPER(LTRIM(RTRIM(ISNULL(reason_code, '')))) = 'MACHINE_BROKEN'
				OR UPPER(LTRIM(RTRIM(ISNULL(reason_label, '')))) LIKE '%MESIN RUSAK%'
			)
	`

	var lockedID int64
	var uuid string
	var claimedBy string
	var isDone int
	err = tx.QueryRowContext(ctx, lockQuery, sql.Named("id", input.ID)).Scan(
		&lockedID,
		&uuid,
		&claimedBy,
		&isDone,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.MechanicActionResponse{}, ErrMechanicTicketNotFound
		}
		return models.MechanicActionResponse{}, err
	}

	if isDone == 1 {
		return models.MechanicActionResponse{}, ErrMechanicAlreadyDone
	}

	claimedBy = strings.TrimSpace(claimedBy)
	if claimedBy == "" {
		return models.MechanicActionResponse{}, ErrMechanicClaimRequired
	}
	if !strings.EqualFold(claimedBy, input.MechanicNIK) {
		return models.MechanicActionResponse{}, ErrMechanicAlreadyClaimed
	}

	_, err = tx.ExecContext(
		ctx,
		`
		UPDATE dbo.machine_operator_loss_events
		SET
			mechanic_done_at = SYSDATETIME(),
			mechanic_done_by_nik = @nik,
			mechanic_done_by_name = @name,
			updated_at = SYSDATETIME()
		WHERE id = @id
		  AND mechanic_done_at IS NULL
		`,
		sql.Named("id", input.ID),
		sql.Named("nik", input.MechanicNIK),
		sql.Named("name", strings.ToUpper(identity.Name)),
	)
	if err != nil {
		return models.MechanicActionResponse{}, err
	}

	item, err := r.getMechanicBrokenByIDTx(ctx, tx, input.ID)
	if err != nil {
		return models.MechanicActionResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.MechanicActionResponse{}, err
	}

	return models.MechanicActionResponse{
		Status:  "ok",
		Message: "Perbaikan mekanik selesai. Loss di halaman operator tidak ikut ditutup.",
		Item:    &item,
	}, nil
}
