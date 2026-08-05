package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/utils"
)

func formatDurationText(seconds int64) string {
	if seconds < 0 {
		seconds = 0
	}

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}

func parseNaiveDateTime(value string) (time.Time, bool) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return time.Time{}, false
	}

	raw = strings.Replace(raw, "T", " ", 1)
	if idx := strings.Index(raw, "."); idx >= 0 {
		raw = raw[:idx]
	}
	if idx := strings.IndexAny(raw, "+Z"); idx >= 0 {
		raw = strings.TrimSpace(raw[:idx])
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

func naiveNowUTC() time.Time {
	now := time.Now()
	return time.Date(
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		0, time.UTC,
	)
}

func (r *Repository) enrichOperatorReportSessionStats(
	ctx context.Context,
	report []models.MachineOperatorReportItem,
) {
	if len(report) == 0 {
		return
	}

	machines, err := r.GetMachines(ctx)
	if err != nil {
		log.Printf("operator report stats: gagal ambil machines: %v", err)
		return
	}

	tableByUUID := make(map[string]string, len(machines))
	for _, m := range machines {
		uuid := strings.ToLower(strings.TrimSpace(m.UUID))
		if uuid == "" {
			continue
		}
		tableByUUID[uuid] = strings.TrimSpace(m.TableName)
	}

	for i := range report {
		item := &report[i]
		uuid := strings.TrimSpace(item.UUID)
		loginAt, okLogin := parseNaiveDateTime(item.LoginTime)
		if !okLogin || uuid == "" {
			item.HasSessionStats = true
			continue
		}

		endAt, okLogout := parseNaiveDateTime(item.LogoutTime)
		statusUpper := strings.ToUpper(strings.TrimSpace(item.Status))
		if !okLogout || statusUpper == "ACTIVE" || statusUpper == "OPEN" {
			endAt = naiveNowUTC()
		}

		if !endAt.After(loginAt) {
			item.HasSessionStats = true
			continue
		}

		runtimeSec, err := r.GetRuntimeSec(ctx, uuid, "", loginAt, endAt)
		if err != nil {
			log.Printf("operator report stats runtime uuid=%s: %v", uuid, err)
			runtimeSec = 0
		}

		procSec := int64(0)
		tableName := tableByUUID[strings.ToLower(uuid)]
		if tableName != "" {
			ps, err := r.GetProductionStats(ctx, tableName, loginAt, endAt)
			if err != nil {
				log.Printf("operator report stats proc uuid=%s table=%s: %v", uuid, tableName, err)
			} else {
				procSec = ps.ProcSec
			}
		}

		lossSec := runtimeSec - procSec
		if lossSec < 0 {
			lossSec = 0
		}

		pct := 0.0
		if runtimeSec > 0 {
			pct = float64(procSec) / float64(runtimeSec) * 100
		}
		if pct > 100 {
			pct = 100
		}
		pct = utils.Round2(pct)

		item.HasSessionStats = true
		item.RuntimeSec = runtimeSec
		item.ProcSec = procSec
		item.LossTimeSec = lossSec
		item.ProductivityPct = pct
		item.ProductivityStatus = utils.StatusFromPct(pct)
	}
}

func (r *Repository) GetMachineOperatorReport(ctx context.Context, date string, withStats bool) ([]models.MachineOperatorReportItem, error) {
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
	sessionIndex := make(map[int64]int)

	for rows.Next() {
		session, err := scanMachineOperatorSession(rows)
		if err != nil {
			return nil, err
		}

		report = append(report, models.MachineOperatorReportItem{
			MachineOperatorSession: session,
			Notes:                  make([]models.MachineOperatorNote, 0),
		})

		sessionIndex[session.ID] = len(report) - 1
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

		note.EndTime = ""
		note.DurationSeconds = 0
		note.DurationText = ""
		note.Status = ""
		note.IsActiveLossEvent = false

		if idx, ok := sessionIndex[note.SessionID]; ok {
			report[idx].Notes = append(report[idx].Notes, note)
		}
	}

	if err := noteRows.Err(); err != nil {
		return nil, err
	}

	lossEventQuery := `
		WITH loss_ranked AS (
			SELECT
				e.id,
				e.session_id,
				ISNULL(e.uuid, '') AS uuid,
				ISNULL(e.operator_nik, '') AS operator_nik,
				ISNULL(e.operator_name, '') AS operator_name,
				ISNULL(e.reason_code, '') AS reason_code,
				ISNULL(e.reason_label, '') AS reason_label,
				ISNULL(e.note, '') AS note,
				ISNULL(CONVERT(VARCHAR(19), e.start_time, 120), '') AS start_time,
				ISNULL(CONVERT(VARCHAR(19), e.end_time, 120), '') AS end_time,
				CAST(
					CASE
						WHEN e.end_time IS NULL
							THEN DATEDIFF_BIG(SECOND, e.start_time, SYSDATETIME())
						ELSE DATEDIFF_BIG(SECOND, e.start_time, e.end_time)
					END AS BIGINT
				) AS duration_seconds,
				ISNULL(e.status, '') AS status,
				ROW_NUMBER() OVER (
					PARTITION BY e.session_id
					ORDER BY e.start_time DESC, e.id DESC
				) AS rn
			FROM dbo.machine_operator_loss_events e
			INNER JOIN dbo.machine_operator_sessions s
				ON s.id = e.session_id
			WHERE
				s.session_date = CAST(@date AS DATE)
		)
		SELECT
			id,
			session_id,
			uuid,
			operator_nik,
			operator_name,
			reason_code,
			reason_label,
			note,
			start_time,
			end_time,
			duration_seconds,
			status
		FROM loss_ranked
		WHERE rn <= 5
		ORDER BY session_id ASC, start_time DESC, id DESC
	`

	lossRows, err := r.DB.QueryContext(ctx, lossEventQuery, sql.Named("date", date))
	if err != nil {
		return nil, err
	}
	defer lossRows.Close()

	for lossRows.Next() {
		var eventID int64
		var sessionID int64
		var uuid string
		var operatorNIK string
		var operatorName string
		var reasonCode string
		var reasonLabel string
		var eventNote string
		var startTime string
		var endTime string
		var durationSeconds int64
		var status string

		err := lossRows.Scan(
			&eventID,
			&sessionID,
			&uuid,
			&operatorNIK,
			&operatorName,
			&reasonCode,
			&reasonLabel,
			&eventNote,
			&startTime,
			&endTime,
			&durationSeconds,
			&status,
		)
		if err != nil {
			return nil, err
		}

		idx, ok := sessionIndex[sessionID]
		if !ok {
			continue
		}

		durationText := formatDurationText(durationSeconds)
		isActive := strings.EqualFold(status, "ACTIVE") && strings.TrimSpace(endTime) == ""

		noteText := strings.TrimSpace(eventNote)
		if noteText == "" {
			if isActive {
				noteText = "Sedang berjalan " + durationText
			} else {
				noteText = "Selesai " + durationText
			}
		} else {
			if isActive {
				noteText = noteText + " - Sedang berjalan " + durationText
			} else {
				noteText = noteText + " - Selesai " + durationText
			}
		}

		lossNote := models.MachineOperatorNote{
			ID:                eventID,
			SessionID:         sessionID,
			SessionDate:       report[idx].SessionDate,
			UUID:              uuid,
			OperatorNIK:       operatorNIK,
			OperatorName:      operatorName,
			ReasonCode:        reasonCode,
			ReasonName:        reasonLabel,
			Note:              noteText,
			CreatedAt:         startTime,
			EndTime:           endTime,
			DurationSeconds:   durationSeconds,
			DurationText:      durationText,
			Status:            status,
			IsActiveLossEvent: isActive,
		}

		report[idx].Notes = append(report[idx].Notes, lossNote)

		if isActive {
			report[idx].ActiveLossReasonCode = reasonCode
			report[idx].ActiveLossReasonLabel = reasonLabel
			report[idx].ActiveLossStartTime = startTime
			report[idx].ActiveLossDurationSeconds = durationSeconds
			report[idx].ActiveLossDurationText = durationText
			report[idx].ActiveLossStatus = status
		}
	}

	if err := lossRows.Err(); err != nil {
		return nil, err
	}

	for i := range report {
		if report[i].Notes == nil {
			report[i].Notes = make([]models.MachineOperatorNote, 0)
		}
	}

	if withStats {
		r.enrichOperatorReportSessionStats(ctx, report)
	}

	return report, nil
}
