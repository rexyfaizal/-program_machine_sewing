package repository

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"backend_machine/models"
)

func parseMacStateToInt(macState string) int {
	value := strings.ToLower(strings.TrimSpace(macState))

	switch value {
	case "1", "online", "on", "aktif", "active", "running":
		return 1
	default:
		return 0
	}
}

func isMacStateOffline(macState string) bool {
	value := strings.ToLower(strings.TrimSpace(macState))

	switch value {
	case "0", "offline", "off", "inactive", "disconnect", "disconnected", "":
		return true
	default:
		stateInt, err := strconv.Atoi(value)
		if err != nil {
			return false
		}
		return stateInt == 0
	}
}

func (r *Repository) RunMachineOfflineAutoLogout(ctx context.Context) (models.MachineOfflineAutoLogoutResponse, error) {
	machines, err := r.GetMachines(ctx)
	if err != nil {
		return models.MachineOfflineAutoLogoutResponse{}, err
	}

	settings, _ := r.GetMachineSettings(ctx)

	upsertOfflineQuery := `
		MERGE dbo.machine_offline_state WITH (HOLDLOCK) AS target
		USING (
			SELECT
				@uuid AS uuid,
				@machine_name AS machine_name,
				@mac_state AS mac_state
		) AS source
		ON LOWER(LTRIM(RTRIM(target.uuid))) = LOWER(LTRIM(RTRIM(source.uuid)))

		WHEN MATCHED THEN
			UPDATE SET
				machine_name = source.machine_name,
				mac_state = source.mac_state,
				offline_since = CASE
					WHEN target.offline_since IS NULL THEN SYSDATETIME()
					ELSE target.offline_since
				END,
				updated_at = SYSDATETIME()

		WHEN NOT MATCHED THEN
			INSERT (
				uuid,
				machine_name,
				mac_state,
				offline_since,
				last_seen_at,
				updated_at
			)
			VALUES (
				source.uuid,
				source.machine_name,
				source.mac_state,
				SYSDATETIME(),
				SYSDATETIME(),
				SYSDATETIME()
			);
	`

	upsertOnlineQuery := `
		MERGE dbo.machine_offline_state WITH (HOLDLOCK) AS target
		USING (
			SELECT
				@uuid AS uuid,
				@machine_name AS machine_name,
				@mac_state AS mac_state
		) AS source
		ON LOWER(LTRIM(RTRIM(target.uuid))) = LOWER(LTRIM(RTRIM(source.uuid)))

		WHEN MATCHED THEN
			UPDATE SET
				machine_name = source.machine_name,
				mac_state = source.mac_state,
				offline_since = NULL,
				last_seen_at = SYSDATETIME(),
				updated_at = SYSDATETIME()

		WHEN NOT MATCHED THEN
			INSERT (
				uuid,
				machine_name,
				mac_state,
				offline_since,
				last_seen_at,
				updated_at
			)
			VALUES (
				source.uuid,
				source.machine_name,
				source.mac_state,
				NULL,
				SYSDATETIME(),
				SYSDATETIME()
			);
	`

	totalMachines := 0
	onlineMachines := 0
	offlineMachines := 0

	for _, m := range machines {
		uuid := strings.TrimSpace(m.UUID)
		if uuid == "" {
			continue
		}

		totalMachines++

		machineName := strings.TrimSpace(m.NickName)

		key := strings.ToLower(strings.TrimSpace(m.UUID))
		if setting, ok := settings[key]; ok {
			if strings.TrimSpace(setting.CustomName) != "" {
				machineName = strings.TrimSpace(setting.CustomName)
			}
		}

		macStateInt := parseMacStateToInt(m.MacState)

		if isMacStateOffline(m.MacState) {
			offlineMachines++

			_, err := r.DB.ExecContext(
				ctx,
				upsertOfflineQuery,
				sql.Named("uuid", uuid),
				sql.Named("machine_name", machineName),
				sql.Named("mac_state", macStateInt),
			)
			if err != nil {
				return models.MachineOfflineAutoLogoutResponse{}, err
			}
		} else {
			onlineMachines++

			_, err := r.DB.ExecContext(
				ctx,
				upsertOnlineQuery,
				sql.Named("uuid", uuid),
				sql.Named("machine_name", machineName),
				sql.Named("mac_state", macStateInt),
			)
			if err != nil {
				return models.MachineOfflineAutoLogoutResponse{}, err
			}
		}
	}

	autoLogoutQuery := `
		DECLARE @closed TABLE (
			id BIGINT,
			session_date DATE,
			uuid NVARCHAR(100),
			operator_nik NVARCHAR(100),
			operator_name NVARCHAR(255)
		);

		UPDATE s
		SET
			s.logout_time = SYSDATETIME(),
			s.status = 'AUTO_LOGOUT',
			s.updated_at = SYSDATETIME()
		OUTPUT
			INSERTED.id,
			INSERTED.session_date,
			INSERTED.uuid,
			INSERTED.operator_nik,
			INSERTED.operator_name
		INTO @closed
		FROM dbo.machine_operator_sessions s
		INNER JOIN dbo.machine_offline_state o
			ON LOWER(LTRIM(RTRIM(s.uuid))) = LOWER(LTRIM(RTRIM(o.uuid)))
		WHERE
			s.status = 'ACTIVE'
			AND s.logout_time IS NULL
			AND o.offline_since IS NOT NULL
			AND o.offline_since <= DATEADD(MINUTE, -60, SYSDATETIME());

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
			'AUTO_LOGOUT_OFFLINE',
			'Auto Logout Mesin Offline',
			'Session ditutup otomatis karena mesin offline lebih dari 60 menit',
			SYSDATETIME()
		FROM @closed c
		WHERE NOT EXISTS (
			SELECT 1
			FROM dbo.machine_operator_notes n
			WHERE n.session_id = c.id
			  AND n.reason_code = 'AUTO_LOGOUT_OFFLINE'
		);

		SELECT COUNT(*) FROM @closed;
	`

	var autoLogoutSessions int64

	if err := r.DB.QueryRowContext(ctx, autoLogoutQuery).Scan(&autoLogoutSessions); err != nil {
		return models.MachineOfflineAutoLogoutResponse{}, err
	}

	return models.MachineOfflineAutoLogoutResponse{
		Status:             "ok",
		Message:            "Auto logout offline selesai diproses",
		CheckedAt:          time.Now().Format("2006-01-02 15:04:05"),
		TotalMachines:      totalMachines,
		OnlineMachines:     onlineMachines,
		OfflineMachines:    offlineMachines,
		AutoLogoutSessions: autoLogoutSessions,
	}, nil
}
