package repository

import (
	"context"
	"database/sql"
	"strings"

	"backend_machine/models"
)

func (r *Repository) SaveLogsMachine(ctx context.Context, date string, rows []models.ProductivityRow) error {
	query := `
		MERGE dbo.logs_machine WITH (HOLDLOCK) AS target
		USING (
			SELECT
				CAST(@log_date AS DATE) AS log_date,
				LOWER(LTRIM(RTRIM(@uuid))) AS uuid_key,
				@uuid AS uuid,
				@machine_name AS machine_name,
				@ip AS ip,
				@location AS location,
				@pic AS pic,
				@spv AS spv,
				@power_on_seconds AS power_on_seconds,
				@running_seconds AS running_seconds,
				@loss_seconds AS loss_seconds,
				@productivity AS productivity,
				@status AS status
		) AS source
		ON target.log_day = source.log_date
		   AND target.uuid_key = source.uuid_key

		WHEN MATCHED THEN
			UPDATE SET
				recorded_at = SYSDATETIME(),

				-- HANYA UPDATE 5 KOLOM UNTUK LAPORAN HARIAN
				power_on_seconds = source.power_on_seconds,
				running_seconds = source.running_seconds,
				loss_seconds = source.loss_seconds,
				productivity = source.productivity,
				status = source.status

		WHEN NOT MATCHED THEN
			INSERT (
				log_date,
				recorded_at,
				uuid,
				machine_name,
				ip,
				location,
				pic,
				spv,
				power_on_seconds,
				running_seconds,
				loss_seconds,
				productivity,
				status
			)
			VALUES (
				source.log_date,
				SYSDATETIME(),
				source.uuid,
				source.machine_name,
				source.ip,
				source.location,
				source.pic,
				source.spv,
				source.power_on_seconds,
				source.running_seconds,
				source.loss_seconds,
				source.productivity,
				source.status
			);
	`

	for _, row := range rows {
		uuid := strings.TrimSpace(row.UUID)
		if uuid == "" {
			continue
		}

		machineName := strings.TrimSpace(row.NickName)
		ip := strings.TrimSpace(row.IP)

		location := strings.TrimSpace(row.Location)
		if location == "-" {
			location = ""
		}

		pic := strings.TrimSpace(row.Pic)
		if pic == "-" {
			pic = ""
		}

		spv := strings.TrimSpace(row.Spv)
		if spv == "-" {
			spv = ""
		}

		lossSeconds := row.LossTimeSec
		if lossSeconds <= 0 {
			lossSeconds = row.RuntimeSec - row.ProcSec
			if lossSeconds < 0 {
				lossSeconds = 0
			}
		}

		_, err := r.DB.ExecContext(
			ctx,
			query,
			sql.Named("log_date", strings.TrimSpace(date)),
			sql.Named("uuid", uuid),
			sql.Named("machine_name", machineName),
			sql.Named("ip", ip),
			sql.Named("location", location),
			sql.Named("pic", pic),
			sql.Named("spv", spv),
			sql.Named("power_on_seconds", row.RuntimeSec),
			sql.Named("running_seconds", row.ProcSec),
			sql.Named("loss_seconds", lossSeconds),
			sql.Named("productivity", row.ProductivityPct),
			sql.Named("status", strings.TrimSpace(row.Status)),
		)

		if err != nil {
			return err
		}
	}

	return nil
}
