package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"telegram_notif/models"
)

type MachineNotificationRepository interface {
	FindPending(
		ctx context.Context,
		limit int,
	) ([]models.MachineNotification, error)

	TryClaim(
		ctx context.Context,
		notification models.MachineNotification,
	) (bool, error)

	MarkSent(
		ctx context.Context,
		notification models.MachineNotification,
		message string,
	) error

	MarkFailed(
		ctx context.Context,
		notification models.MachineNotification,
		message string,
		errorMessage string,
	) error
}

type machineNotificationRepository struct {
	db *sql.DB
}

func NewMachineNotificationRepository(
	db *sql.DB,
) (MachineNotificationRepository, error) {
	if db == nil {
		return nil, errors.New("database tidak boleh nil")
	}

	return &machineNotificationRepository{
		db: db,
	}, nil
}

func (r *machineNotificationRepository) FindPending(
	ctx context.Context,
	limit int,
) ([]models.MachineNotification, error) {
	if limit <= 0 {
		limit = 100
	}

	if limit > 500 {
		limit = 500
	}

	query := `
		WITH CandidateNotifications AS
		(
			/* =================================================
			   MESIN RUSAK → SELURUH MEKANIK AKTIF
			   ================================================= */
			SELECT
				n.id AS operator_note_id,
				ISNULL(n.session_id, 0) AS session_id,

				LTRIM(RTRIM(
					ISNULL(n.uuid, '')
				)) AS uuid,

				LTRIM(RTRIM(
					ISNULL(ms.custom_name, n.uuid)
				)) AS machine_name,

				LTRIM(RTRIM(
					ISNULL(ms.location, '')
				)) AS location,

				LTRIM(RTRIM(
					ISNULL(
						CONVERT(varchar(255), n.operator_nik),
						''
					)
				)) AS operator_nik,

				LTRIM(RTRIM(
					ISNULL(n.operator_name, '')
				)) AS operator_name,

				'MACHINE_BROKEN' AS reason_code,

				LTRIM(RTRIM(
					ISNULL(n.reason_label, 'Mesin Rusak')
				)) AS reason_name,

				LTRIM(RTRIM(
					ISNULL(n.note, '')
				)) AS note,

				COALESCE(
					n.start_time,
					n.created_at,
					SYSDATETIME()
				) AS event_created_at,

				LTRIM(RTRIM(
					ISNULL(
						CONVERT(varchar(255), e.nik),
						''
					)
				)) AS recipient_nik,

				LTRIM(RTRIM(
					ISNULL(e.name, '')
				)) AS recipient_name,

				LTRIM(RTRIM(
					ISNULL(e.bagian, '')
				)) AS recipient_role,

				LTRIM(RTRIM(
					ISNULL(e.branchdetail, '')
				)) AS recipient_branch,

				e.id_telegram AS telegram_chat_id

			FROM dbo.machine_operator_loss_events n

			OUTER APPLY
			(
				SELECT TOP (1)
					m.custom_name,
					m.location
				FROM dbo.machine_setting_manual m
				WHERE LOWER(LTRIM(RTRIM(m.uuid))) =
					  LOWER(LTRIM(RTRIM(n.uuid)))
				ORDER BY m.updated_at DESC
			) ms

			CROSS JOIN dbo.employee e

			WHERE UPPER(
					LTRIM(RTRIM(
						ISNULL(n.reason_code, '')
					))
				  ) = 'MACHINE_BROKEN'

			  AND COALESCE(n.start_time, n.created_at) >=
				  DATEADD(MINUTE, -30, SYSDATETIME())

			  AND UPPER(
					LTRIM(RTRIM(
						ISNULL(e.bagian, '')
					))
				  ) = 'MEKANIK'

			  AND e.id_telegram IS NOT NULL
			  AND ISNULL(e.notification_active, 0) = 1

			UNION ALL

			/* =================================================
			   TUNGGU HANCA → SPV BERDASARKAN BRANCH DAN LINE
			   ================================================= */
			SELECT
				n.id AS operator_note_id,
				ISNULL(n.session_id, 0) AS session_id,

				LTRIM(RTRIM(
					ISNULL(n.uuid, '')
				)) AS uuid,

				LTRIM(RTRIM(
					ISNULL(ms.custom_name, n.uuid)
				)) AS machine_name,

				LTRIM(RTRIM(
					ISNULL(ms.location, '')
				)) AS location,

				LTRIM(RTRIM(
					ISNULL(
						CONVERT(varchar(255), n.operator_nik),
						''
					)
				)) AS operator_nik,

				LTRIM(RTRIM(
					ISNULL(n.operator_name, '')
				)) AS operator_name,

				'WAIT_HANCA' AS reason_code,

				LTRIM(RTRIM(
					ISNULL(n.reason_label, 'Tunggu Hanca')
				)) AS reason_name,

				LTRIM(RTRIM(
					ISNULL(n.note, '')
				)) AS note,

				COALESCE(
					n.start_time,
					n.created_at,
					SYSDATETIME()
				) AS event_created_at,

				LTRIM(RTRIM(
					ISNULL(
						CONVERT(varchar(255), e.nik),
						''
					)
				)) AS recipient_nik,

				LTRIM(RTRIM(
					ISNULL(e.name, '')
				)) AS recipient_name,

				LTRIM(RTRIM(
					ISNULL(e.bagian, '')
				)) AS recipient_role,

				LTRIM(RTRIM(
					ISNULL(a.branchdetail, '')
				)) AS recipient_branch,

				e.id_telegram AS telegram_chat_id

			FROM dbo.machine_operator_loss_events n

			OUTER APPLY
			(
				SELECT TOP (1)
					m.custom_name,
					m.location
				FROM dbo.machine_setting_manual m
				WHERE LOWER(LTRIM(RTRIM(m.uuid))) =
					  LOWER(LTRIM(RTRIM(n.uuid)))
				ORDER BY m.updated_at DESC
			) ms

			INNER JOIN dbo.spv_line_assignments a
				ON UPPER(LTRIM(RTRIM(a.location))) =
				   UPPER(LTRIM(RTRIM(ms.location)))
			   AND a.is_active = 1

			INNER JOIN dbo.employee e
				ON LTRIM(RTRIM(
					CONVERT(varchar(255), e.nik)
				)) =
				   LTRIM(RTRIM(a.spv_nik))

			WHERE UPPER(
					LTRIM(RTRIM(
						ISNULL(n.reason_code, '')
					))
				  ) = 'WAIT_HANCA'

			  AND COALESCE(n.start_time, n.created_at) >=
				  DATEADD(MINUTE, -30, SYSDATETIME())

			  AND UPPER(
					LTRIM(RTRIM(
						ISNULL(e.bagian, '')
					))
				  ) = 'SPV'

			  AND e.id_telegram IS NOT NULL
			  AND ISNULL(e.notification_active, 0) = 1
		)

		SELECT TOP (@limit)
			c.operator_note_id,
			c.session_id,
			c.uuid,
			c.machine_name,
			c.location,
			c.operator_nik,
			c.operator_name,
			c.reason_code,
			c.reason_name,
			c.note,
			c.event_created_at,
			c.recipient_nik,
			c.recipient_name,
			c.recipient_role,
			c.recipient_branch,
			c.telegram_chat_id

		FROM CandidateNotifications c

		WHERE NOT EXISTS
		(
			SELECT 1
			FROM dbo.telegram_notification_logs l
			WHERE l.operator_note_id =
				  c.operator_note_id

			  AND LTRIM(RTRIM(
					ISNULL(l.recipient_nik, '')
				  )) =
				  LTRIM(RTRIM(c.recipient_nik))

			  AND UPPER(LTRIM(RTRIM(
					ISNULL(l.notification_type, '')
				  ))) =
				  UPPER(LTRIM(RTRIM(c.reason_code)))

			  AND l.sent_status = 'SENT'
		)

		AND NOT EXISTS
		(
			SELECT 1
			FROM dbo.telegram_notification_logs l
			WHERE l.operator_note_id =
				  c.operator_note_id

			  AND LTRIM(RTRIM(
					ISNULL(l.recipient_nik, '')
				  )) =
				  LTRIM(RTRIM(c.recipient_nik))

			  AND UPPER(LTRIM(RTRIM(
					ISNULL(l.notification_type, '')
				  ))) =
				  UPPER(LTRIM(RTRIM(c.reason_code)))

			  AND l.sent_status = 'PROCESSING'

			  AND l.last_attempt_at >
				  DATEADD(
					  MINUTE,
					  -2,
					  SYSDATETIME()
				  )
		)

		AND NOT EXISTS
		(
			SELECT 1
			FROM dbo.telegram_notification_logs l
			WHERE l.operator_note_id =
				  c.operator_note_id

			  AND LTRIM(RTRIM(
					ISNULL(l.recipient_nik, '')
				  )) =
				  LTRIM(RTRIM(c.recipient_nik))

			  AND UPPER(LTRIM(RTRIM(
					ISNULL(l.notification_type, '')
				  ))) =
				  UPPER(LTRIM(RTRIM(c.reason_code)))

			  AND l.sent_status = 'FAILED'

			  AND l.last_attempt_at >
				  DATEADD(
					  MINUTE,
					  -5,
					  SYSDATETIME()
				  )
		)

		ORDER BY
			c.event_created_at ASC,
			c.recipient_name ASC;
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		sql.Named("limit", limit),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"gagal mengambil notifikasi operator: %w",
			err,
		)
	}
	defer rows.Close()

	notifications := make(
		[]models.MachineNotification,
		0,
	)

	for rows.Next() {
		var item models.MachineNotification

		if err := rows.Scan(
			&item.OperatorNoteID,
			&item.SessionID,
			&item.UUID,
			&item.MachineName,
			&item.Location,
			&item.OperatorNIK,
			&item.OperatorName,
			&item.ReasonCode,
			&item.ReasonName,
			&item.Note,
			&item.EventCreatedAt,
			&item.RecipientNIK,
			&item.RecipientName,
			&item.RecipientRole,
			&item.RecipientBranch,
			&item.TelegramChatID,
		); err != nil {
			return nil, fmt.Errorf(
				"gagal membaca notifikasi operator: %w",
				err,
			)
		}

		notifications = append(
			notifications,
			item,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"gagal membaca hasil notifikasi: %w",
			err,
		)
	}

	return notifications, nil
}

func (r *machineNotificationRepository) TryClaim(
	ctx context.Context,
	notification models.MachineNotification,
) (bool, error) {
	notificationType, err :=
		validNotificationType(notification.ReasonCode)
	if err != nil {
		return false, err
	}

	tx, err := r.db.BeginTx(
		ctx,
		&sql.TxOptions{
			Isolation: sql.LevelSerializable,
		},
	)
	if err != nil {
		return false, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	updateQuery := `
		UPDATE dbo.telegram_notification_logs
			WITH (UPDLOCK, HOLDLOCK)

		SET
			sent_status = 'PROCESSING',
			error_message = NULL,
			sent_at = NULL,
			attempt_count =
				ISNULL(attempt_count, 0) + 1,
			last_attempt_at = SYSDATETIME()

		WHERE operator_note_id = @operator_note_id

		  AND LTRIM(RTRIM(
				ISNULL(recipient_nik, '')
			  )) = @recipient_nik

		  AND UPPER(LTRIM(RTRIM(
				ISNULL(notification_type, '')
			  ))) = @notification_type

		  AND
		  (
				(
					sent_status = 'FAILED'
					AND ISNULL(
						last_attempt_at,
						'19000101'
					) <= DATEADD(
						MINUTE,
						-5,
						SYSDATETIME()
					)
				)

				OR

				(
					sent_status = 'PROCESSING'
					AND ISNULL(
						last_attempt_at,
						'19000101'
					) <= DATEADD(
						MINUTE,
						-2,
						SYSDATETIME()
					)
				)

				OR ISNULL(sent_status, '')
				   NOT IN (
						'SENT',
						'FAILED',
						'PROCESSING'
				   )
		  );
	`

	result, err := tx.ExecContext(
		ctx,
		updateQuery,
		sql.Named(
			"operator_note_id",
			notification.OperatorNoteID,
		),
		sql.Named(
			"recipient_nik",
			strings.TrimSpace(
				notification.RecipientNIK,
			),
		),
		sql.Named(
			"notification_type",
			notificationType,
		),
	)
	if err != nil {
		return false, fmt.Errorf(
			"gagal mengunci notifikasi: %w",
			err,
		)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	if affectedRows == 1 {
		if err := tx.Commit(); err != nil {
			return false, err
		}

		return true, nil
	}

	var existingStatus string

	checkQuery := `
		SELECT TOP (1)
			ISNULL(sent_status, '')
		FROM dbo.telegram_notification_logs
			WITH (UPDLOCK, HOLDLOCK)

		WHERE operator_note_id = @operator_note_id

		  AND LTRIM(RTRIM(
				ISNULL(recipient_nik, '')
			  )) = @recipient_nik

		  AND UPPER(LTRIM(RTRIM(
				ISNULL(notification_type, '')
			  ))) = @notification_type;
	`

	err = tx.QueryRowContext(
		ctx,
		checkQuery,
		sql.Named(
			"operator_note_id",
			notification.OperatorNoteID,
		),
		sql.Named(
			"recipient_nik",
			strings.TrimSpace(
				notification.RecipientNIK,
			),
		),
		sql.Named(
			"notification_type",
			notificationType,
		),
	).Scan(&existingStatus)

	if err == nil {
		if err := tx.Commit(); err != nil {
			return false, err
		}

		return false, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf(
			"gagal memeriksa klaim notifikasi: %w",
			err,
		)
	}

	insertQuery := `
		INSERT INTO dbo.telegram_notification_logs
		(
			notification_date,
			notification_type,
			uuid,
			machine_name,
			location,

			session_id,
			operator_nik,
			operator_name,

			session_status,
			productivity,
			machine_status,

			telegram_chat_id,
			telegram_message,

			sent_status,
			error_message,
			sent_at,
			created_at,

			operator_note_id,
			recipient_nik,
			recipient_name,
			recipient_role,

			attempt_count,
			last_attempt_at
		)
		VALUES
		(
			CAST(@event_created_at AS DATE),
			@notification_type,
			@uuid,
			@machine_name,
			@location,

			@session_id,
			@operator_nik,
			@operator_name,

			NULL,
			NULL,
			NULL,

			@telegram_chat_id,
			'',

			'PROCESSING',
			NULL,
			NULL,
			SYSDATETIME(),

			@operator_note_id,
			@recipient_nik,
			@recipient_name,
			@recipient_role,

			1,
			SYSDATETIME()
		);
	`

	_, err = tx.ExecContext(
		ctx,
		insertQuery,
		notificationSQLParams(
			notification,
			notificationType,
		)...,
	)
	if err != nil {
		return false, fmt.Errorf(
			"gagal membuat log PROCESSING: %w",
			err,
		)
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

func (r *machineNotificationRepository) MarkSent(
	ctx context.Context,
	notification models.MachineNotification,
	message string,
) error {
	notificationType, err :=
		validNotificationType(notification.ReasonCode)
	if err != nil {
		return err
	}

	query := `
		UPDATE dbo.telegram_notification_logs
		SET
			telegram_message = @telegram_message,
			sent_status = 'SENT',
			error_message = NULL,
			sent_at = SYSDATETIME()

		WHERE operator_note_id = @operator_note_id

		  AND LTRIM(RTRIM(
				ISNULL(recipient_nik, '')
			  )) = @recipient_nik

		  AND UPPER(LTRIM(RTRIM(
				ISNULL(notification_type, '')
			  ))) = @notification_type;
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		sql.Named(
			"telegram_message",
			message,
		),
		sql.Named(
			"operator_note_id",
			notification.OperatorNoteID,
		),
		sql.Named(
			"recipient_nik",
			strings.TrimSpace(
				notification.RecipientNIK,
			),
		),
		sql.Named(
			"notification_type",
			notificationType,
		),
	)
	if err != nil {
		return fmt.Errorf(
			"gagal memperbarui log menjadi SENT: %w",
			err,
		)
	}

	return ensureOneRowAffected(
		result,
		"SENT",
	)
}

func (r *machineNotificationRepository) MarkFailed(
	ctx context.Context,
	notification models.MachineNotification,
	message string,
	errorMessage string,
) error {
	notificationType, err :=
		validNotificationType(notification.ReasonCode)
	if err != nil {
		return err
	}

	query := `
		UPDATE dbo.telegram_notification_logs
		SET
			telegram_message = @telegram_message,
			sent_status = 'FAILED',
			error_message = @error_message,
			sent_at = NULL

		WHERE operator_note_id = @operator_note_id

		  AND LTRIM(RTRIM(
				ISNULL(recipient_nik, '')
			  )) = @recipient_nik

		  AND UPPER(LTRIM(RTRIM(
				ISNULL(notification_type, '')
			  ))) = @notification_type;
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		sql.Named(
			"telegram_message",
			message,
		),
		sql.Named(
			"error_message",
			truncateString(
				errorMessage,
				2000,
			),
		),
		sql.Named(
			"operator_note_id",
			notification.OperatorNoteID,
		),
		sql.Named(
			"recipient_nik",
			strings.TrimSpace(
				notification.RecipientNIK,
			),
		),
		sql.Named(
			"notification_type",
			notificationType,
		),
	)
	if err != nil {
		return fmt.Errorf(
			"gagal memperbarui log menjadi FAILED: %w",
			err,
		)
	}

	return ensureOneRowAffected(
		result,
		"FAILED",
	)
}

func notificationSQLParams(
	notification models.MachineNotification,
	notificationType string,
) []interface{} {
	return []interface{}{
		sql.Named(
			"notification_type",
			notificationType,
		),
		sql.Named(
			"event_created_at",
			notification.EventCreatedAt,
		),
		sql.Named(
			"uuid",
			notification.UUID,
		),
		sql.Named(
			"machine_name",
			notification.MachineName,
		),
		sql.Named(
			"location",
			notification.Location,
		),
		sql.Named(
			"session_id",
			notification.SessionID,
		),
		sql.Named(
			"operator_nik",
			notification.OperatorNIK,
		),
		sql.Named(
			"operator_name",
			notification.OperatorName,
		),
		sql.Named(
			"telegram_chat_id",
			notification.TelegramChatID,
		),
		sql.Named(
			"operator_note_id",
			notification.OperatorNoteID,
		),
		sql.Named(
			"recipient_nik",
			strings.TrimSpace(
				notification.RecipientNIK,
			),
		),
		sql.Named(
			"recipient_name",
			notification.RecipientName,
		),
		sql.Named(
			"recipient_role",
			notification.RecipientRole,
		),
	}
}

func validNotificationType(
	value string,
) (string, error) {
	value = strings.ToUpper(
		strings.TrimSpace(value),
	)

	switch value {
	case models.NotificationTypeMachineBroken:
		return models.NotificationTypeMachineBroken, nil

	case models.NotificationTypeWaitHanca:
		return models.NotificationTypeWaitHanca, nil

	default:
		return "", fmt.Errorf(
			"jenis notifikasi tidak didukung: %s",
			value,
		)
	}
}

func ensureOneRowAffected(
	result sql.Result,
	status string,
) error {
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affectedRows != 1 {
		return fmt.Errorf(
			"log notifikasi tidak berhasil diubah menjadi %s, jumlah baris: %d",
			status,
			affectedRows,
		)
	}

	return nil
}

func truncateString(
	value string,
	maxLength int,
) string {
	value = strings.TrimSpace(value)

	if len(value) <= maxLength {
		return value
	}

	return value[:maxLength]
}
