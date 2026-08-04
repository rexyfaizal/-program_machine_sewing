package repository

import (
	"context"
	"database/sql"
)

func (r *Repository) RunMachineOperatorWorkEndAutoLogout(ctx context.Context) (int64, error) {
	query := `
		DECLARE @closed TABLE (
			id BIGINT,
			session_date DATE,
			uuid NVARCHAR(100),
			operator_nik NVARCHAR(100),
			operator_name NVARCHAR(255),
			logout_time DATETIME2
		);

		UPDATE s
		SET
			s.logout_time = DATEADD(HOUR, 8, s.login_time),
			s.status = 'AUTO_LOGOUT_WORK_END',
			s.updated_at = SYSDATETIME()
		OUTPUT
			INSERTED.id,
			INSERTED.session_date,
			INSERTED.uuid,
			INSERTED.operator_nik,
			INSERTED.operator_name,
			INSERTED.logout_time
		INTO @closed
		FROM dbo.machine_operator_sessions s
		WHERE
			s.status = 'ACTIVE'
			AND s.logout_time IS NULL
			AND s.login_time IS NOT NULL
			AND DATEDIFF(MINUTE, s.login_time, SYSDATETIME()) >= 480;

		UPDATE e
		SET
			e.end_time =
				CASE
					WHEN c.logout_time > e.start_time THEN c.logout_time
					ELSE SYSDATETIME()
				END,
			e.status = 'CLOSED',
			e.updated_at = SYSDATETIME()
		FROM dbo.machine_operator_loss_events e
		INNER JOIN @closed c
			ON c.id = e.session_id
		WHERE
			e.status = 'ACTIVE'
			AND e.end_time IS NULL;

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
			c.id,
			c.session_date,
			c.uuid,
			c.operator_nik,
			c.operator_name,
			'AUTO_LOGOUT_WORK_END',
			'Auto Logout Selesai Jam Kerja',
			'Session ditutup otomatis karena operator sudah login selama 8 jam.',
			SYSDATETIME()
		FROM @closed c
		WHERE NOT EXISTS (
			SELECT 1
			FROM dbo.machine_operator_notes n
			WHERE n.session_id = c.id
			  AND n.reason_code = 'AUTO_LOGOUT_WORK_END'
		);

		SELECT COUNT(*) FROM @closed;
	`

	var totalClosed sql.NullInt64

	err := r.DB.QueryRowContext(ctx, query).Scan(&totalClosed)
	if err != nil {
		return 0, err
	}

	if !totalClosed.Valid || totalClosed.Int64 < 0 {
		return 0, nil
	}

	return totalClosed.Int64, nil
}
