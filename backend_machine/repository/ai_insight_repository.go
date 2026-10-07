package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"backend_machine/models"
)

// EnsureAiInsightSchema membuat tabel log hasil AI insight.
func (r *Repository) EnsureAiInsightSchema(ctx context.Context) error {
	_, err := r.DB.ExecContext(ctx, `
IF OBJECT_ID(N'dbo.ai_insight_log', N'U') IS NULL
BEGIN
    CREATE TABLE dbo.ai_insight_log (
        id BIGINT IDENTITY(1,1) NOT NULL CONSTRAINT PK_ai_insight_log PRIMARY KEY,
        work_date DATE NOT NULL,
        shift_code NVARCHAR(50) NULL,
        location_filter NVARCHAR(100) NULL,
        generated_at DATETIME2 NOT NULL CONSTRAINT DF_ai_insight_log_generated_at DEFAULT SYSDATETIME(),
        high_count INT NOT NULL CONSTRAINT DF_ai_insight_log_high DEFAULT 0,
        medium_count INT NOT NULL CONSTRAINT DF_ai_insight_log_medium DEFAULT 0,
        low_count INT NOT NULL CONSTRAINT DF_ai_insight_log_low DEFAULT 0,
        summary_text NVARCHAR(MAX) NULL,
        payload_json NVARCHAR(MAX) NOT NULL
    );

    CREATE INDEX IX_ai_insight_log_work_date
        ON dbo.ai_insight_log (work_date, generated_at DESC);
END;
`)
	return err
}

func (r *Repository) ListAiInsightActiveSessions(
	ctx context.Context,
) ([]models.AiInsightActiveSession, error) {
	query := `
SELECT
    LTRIM(RTRIM(ISNULL(uuid, ''))) AS uuid,
    LTRIM(RTRIM(ISNULL(location, ''))) AS location,
    LTRIM(RTRIM(ISNULL(operator_nik, ''))) AS operator_nik,
    LTRIM(RTRIM(ISNULL(operator_name, ''))) AS operator_name,
    LTRIM(RTRIM(ISNULL(style_name, ''))) AS style_name,
    LTRIM(RTRIM(ISNULL(process_name, ''))) AS process_name,
    ISNULL(CONVERT(VARCHAR(19), login_time, 120), '') AS login_time,
    LTRIM(RTRIM(ISNULL(status, ''))) AS status
FROM dbo.machine_operator_sessions
WHERE UPPER(LTRIM(RTRIM(ISNULL(status, '')))) IN (N'ACTIVE', N'OPEN')
ORDER BY login_time DESC;
`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]models.AiInsightActiveSession, 0)
	for rows.Next() {
		var item models.AiInsightActiveSession
		if err := rows.Scan(
			&item.UUID,
			&item.Location,
			&item.OperatorNIK,
			&item.OperatorName,
			&item.StyleName,
			&item.ProcessName,
			&item.LoginTime,
			&item.Status,
		); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *Repository) ListAiInsightActiveLosses(
	ctx context.Context,
) ([]models.AiInsightActiveLoss, error) {
	query := `
SELECT
    LTRIM(RTRIM(ISNULL(uuid, ''))) AS uuid,
    LTRIM(RTRIM(ISNULL(operator_name, ''))) AS operator_name,
    LTRIM(RTRIM(ISNULL(reason_code, ''))) AS reason_code,
    LTRIM(RTRIM(ISNULL(reason_label, ''))) AS reason_label,
    LTRIM(RTRIM(ISNULL(note, ''))) AS note,
    ISNULL(CONVERT(VARCHAR(19), start_time, 120), '') AS start_time,
    CAST(
        CASE
            WHEN end_time IS NULL
                THEN DATEDIFF_BIG(SECOND, start_time, SYSDATETIME())
            ELSE DATEDIFF_BIG(SECOND, start_time, end_time)
        END AS BIGINT
    ) AS duration_seconds,
    LTRIM(RTRIM(ISNULL(status, ''))) AS status
FROM dbo.machine_operator_loss_events
WHERE end_time IS NULL
   OR UPPER(LTRIM(RTRIM(ISNULL(status, '')))) = N'ACTIVE'
ORDER BY start_time DESC;
`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]models.AiInsightActiveLoss, 0)
	for rows.Next() {
		var item models.AiInsightActiveLoss
		if err := rows.Scan(
			&item.UUID,
			&item.OperatorName,
			&item.ReasonCode,
			&item.ReasonLabel,
			&item.Note,
			&item.StartTime,
			&item.DurationSec,
			&item.Status,
		); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *Repository) ListAiInsightRecentNotes(
	ctx context.Context,
	workDate string,
) ([]models.AiInsightRecentNote, error) {
	workDate = strings.TrimSpace(workDate)
	if workDate == "" {
		workDate = time.Now().Format("2006-01-02")
	}

	// Gabungkan note manual + loss events (termasuk yang sudah selesai),
	// supaya AI Insight melihat keterangan yang sama dengan kolom Operator Note.
	query := `
SELECT TOP 300
    uuid,
    operator_name,
    reason_code,
    reason_name,
    note,
    created_at
FROM (
    SELECT
        LTRIM(RTRIM(ISNULL(n.uuid, ''))) AS uuid,
        LTRIM(RTRIM(ISNULL(n.operator_name, ''))) AS operator_name,
        LTRIM(RTRIM(ISNULL(n.reason_code, ''))) AS reason_code,
        LTRIM(RTRIM(ISNULL(n.reason_name, ''))) AS reason_name,
        LTRIM(RTRIM(ISNULL(n.note, ''))) AS note,
        ISNULL(CONVERT(VARCHAR(19), n.created_at, 120), '') AS created_at
    FROM dbo.machine_operator_notes n
    WHERE n.session_date = TRY_CONVERT(date, @work_date)
      AND UPPER(LTRIM(RTRIM(ISNULL(n.reason_code, '')))) NOT LIKE N'AUTO_LOGOUT%'

    UNION ALL

    SELECT
        LTRIM(RTRIM(ISNULL(e.uuid, ''))) AS uuid,
        LTRIM(RTRIM(ISNULL(e.operator_name, ''))) AS operator_name,
        LTRIM(RTRIM(ISNULL(e.reason_code, ''))) AS reason_code,
        LTRIM(RTRIM(ISNULL(e.reason_label, ''))) AS reason_name,
        LTRIM(RTRIM(ISNULL(e.note, ''))) AS note,
        ISNULL(CONVERT(VARCHAR(19), e.start_time, 120), '') AS created_at
    FROM dbo.machine_operator_loss_events e
    INNER JOIN dbo.machine_operator_sessions s
        ON s.id = e.session_id
    WHERE s.session_date = TRY_CONVERT(date, @work_date)
      AND UPPER(LTRIM(RTRIM(ISNULL(e.reason_code, '')))) NOT LIKE N'AUTO_LOGOUT%'
) x
ORDER BY created_at DESC;
`

	rows, err := r.DB.QueryContext(ctx, query, sql.Named("work_date", workDate))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]models.AiInsightRecentNote, 0)
	for rows.Next() {
		var item models.AiInsightRecentNote
		if err := rows.Scan(
			&item.UUID,
			&item.OperatorName,
			&item.ReasonCode,
			&item.ReasonName,
			&item.Note,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

// ListAiInsightBrokenFrequency menghitung event MACHINE_BROKEN per UUID
// dalam jendela days (inklusif) yang berakhir di endDate (YYYY-MM-DD).
func (r *Repository) ListAiInsightBrokenFrequency(
	ctx context.Context,
	endDate string,
	days int,
) ([]models.AiInsightBrokenFrequency, error) {
	endDate = strings.TrimSpace(endDate)
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}
	if days <= 0 {
		days = 7
	}

	query := `
SELECT
    LTRIM(RTRIM(ISNULL(e.uuid, ''))) AS uuid,
    COUNT_BIG(1) AS broken_count
FROM dbo.machine_operator_loss_events e
INNER JOIN dbo.machine_operator_sessions s
    ON s.id = e.session_id
WHERE s.session_date >= DATEADD(DAY, 1 - @days, TRY_CONVERT(date, @end_date))
  AND s.session_date <= TRY_CONVERT(date, @end_date)
  AND (
        UPPER(LTRIM(RTRIM(ISNULL(e.reason_code, '')))) = N'MACHINE_BROKEN'
        OR UPPER(LTRIM(RTRIM(ISNULL(e.reason_label, '')))) LIKE N'%MESIN RUSAK%'
      )
GROUP BY LTRIM(RTRIM(ISNULL(e.uuid, '')))
HAVING COUNT_BIG(1) >= 1
ORDER BY broken_count DESC;
`

	rows, err := r.DB.QueryContext(
		ctx,
		query,
		sql.Named("end_date", endDate),
		sql.Named("days", days),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]models.AiInsightBrokenFrequency, 0)
	for rows.Next() {
		var item models.AiInsightBrokenFrequency
		var count int64
		if err := rows.Scan(&item.UUID, &count); err != nil {
			return nil, err
		}
		item.Count = int(count)
		if strings.TrimSpace(item.UUID) == "" {
			continue
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *Repository) SaveAiInsightLog(
	ctx context.Context,
	resp models.AiInsightAnalyzeResponse,
) (int64, error) {
	payload, err := json.Marshal(resp)
	if err != nil {
		return 0, err
	}

	query := `
INSERT INTO dbo.ai_insight_log (
    work_date,
    shift_code,
    location_filter,
    generated_at,
    high_count,
    medium_count,
    low_count,
    summary_text,
    payload_json
)
OUTPUT INSERTED.id
VALUES (
    TRY_CONVERT(date, @work_date),
    @shift_code,
    @location_filter,
    TRY_CONVERT(datetime2, @generated_at),
    @high_count,
    @medium_count,
    @low_count,
    @summary_text,
    @payload_json
);
`

	var id int64
	err = r.DB.QueryRowContext(
		ctx,
		query,
		sql.Named("work_date", strings.TrimSpace(resp.WorkDate)),
		sql.Named("shift_code", nullIfBlank(resp.Shift)),
		sql.Named("location_filter", nullIfBlank(resp.Location)),
		sql.Named("generated_at", strings.TrimSpace(resp.GeneratedAt)),
		sql.Named("high_count", resp.HighCount),
		sql.Named("medium_count", resp.MediumCount),
		sql.Named("low_count", resp.LowCount),
		sql.Named("summary_text", nullIfBlank(resp.Summary)),
		sql.Named("payload_json", string(payload)),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) GetLatestAiInsightLog(
	ctx context.Context,
	workDate, location string,
) (models.AiInsightAnalyzeResponse, bool, error) {
	workDate = strings.TrimSpace(workDate)
	location = strings.TrimSpace(location)

	query := `
SELECT TOP 1 payload_json
FROM dbo.ai_insight_log
WHERE (@work_date = '' OR work_date = TRY_CONVERT(date, @work_date))
  AND (
        @location = ''
        OR UPPER(LTRIM(RTRIM(ISNULL(location_filter, '')))) = UPPER(LTRIM(RTRIM(@location)))
      )
ORDER BY generated_at DESC, id DESC;
`

	var payload string
	err := r.DB.QueryRowContext(
		ctx,
		query,
		sql.Named("work_date", workDate),
		sql.Named("location", location),
	).Scan(&payload)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.AiInsightAnalyzeResponse{}, false, nil
		}
		return models.AiInsightAnalyzeResponse{}, false, err
	}

	var resp models.AiInsightAnalyzeResponse
	if err := json.Unmarshal([]byte(payload), &resp); err != nil {
		return models.AiInsightAnalyzeResponse{}, false, fmt.Errorf("payload log AI tidak valid: %w", err)
	}
	return resp, true, nil
}

func nullIfBlank(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

// ParseAiInsightGeneratedAt helper aman untuk timestamp log.
func ParseAiInsightGeneratedAt(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Now()
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t
		}
	}
	return time.Now()
}
