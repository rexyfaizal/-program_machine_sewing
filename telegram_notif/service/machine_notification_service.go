package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"telegram_notif/models"
	"telegram_notif/repository"
)

type MachineNotificationSender interface {
	SendMessage(
		ctx context.Context,
		chatID int64,
		message string,
	) error
}

type MachineNotificationService interface {
	ProcessPending(
		ctx context.Context,
	) error
}

type machineNotificationService struct {
	repository repository.MachineNotificationRepository
	sender     MachineNotificationSender
	batchSize  int
}

func NewMachineNotificationService(
	machineRepository repository.MachineNotificationRepository,
	sender MachineNotificationSender,
	batchSize int,
) MachineNotificationService {
	if batchSize <= 0 {
		batchSize = 100
	}

	return &machineNotificationService{
		repository: machineRepository,
		sender:     sender,
		batchSize:  batchSize,
	}
}

func (s *machineNotificationService) ProcessPending(
	ctx context.Context,
) error {
	notifications, err :=
		s.repository.FindPending(
			ctx,
			s.batchSize,
		)
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		if ctx.Err() != nil {
			return nil
		}

		claimed, err := s.repository.TryClaim(
			ctx,
			notification,
		)
		if err != nil {
			log.Printf(
				"Gagal claim notifikasi type=%s note=%d penerima=%s: %v",
				notification.ReasonCode,
				notification.OperatorNoteID,
				notification.RecipientNIK,
				err,
			)
			continue
		}

		if !claimed {
			continue
		}

		message, err := buildOperatorNotificationMessage(
			notification,
		)
		if err != nil {
			_ = s.repository.MarkFailed(
				ctx,
				notification,
				"",
				err.Error(),
			)

			log.Printf(
				"Gagal membuat pesan note=%d: %v",
				notification.OperatorNoteID,
				err,
			)
			continue
		}

		sendCtx, cancel := context.WithTimeout(
			ctx,
			20*time.Second,
		)

		sendErr := s.sender.SendMessage(
			sendCtx,
			notification.TelegramChatID,
			message,
		)

		cancel()

		if sendErr != nil {
			if err := s.repository.MarkFailed(
				ctx,
				notification,
				message,
				sendErr.Error(),
			); err != nil {
				log.Printf(
					"Gagal mencatat FAILED type=%s note=%d penerima=%s: %v",
					notification.ReasonCode,
					notification.OperatorNoteID,
					notification.RecipientNIK,
					err,
				)
			}

			log.Printf(
				"Notifikasi gagal type=%s note=%d penerima=%s telegram=%d: %v",
				notification.ReasonCode,
				notification.OperatorNoteID,
				notification.RecipientName,
				notification.TelegramChatID,
				sendErr,
			)

			continue
		}

		if err := s.repository.MarkSent(
			ctx,
			notification,
			message,
		); err != nil {
			log.Printf(
				"Pesan terkirim tetapi gagal mencatat SENT type=%s note=%d penerima=%s: %v",
				notification.ReasonCode,
				notification.OperatorNoteID,
				notification.RecipientNIK,
				err,
			)

			continue
		}

		log.Printf(
			"Notifikasi terkirim. type=%s note=%d penerima=%s role=%s branch=%s lokasi=%s",
			notification.ReasonCode,
			notification.OperatorNoteID,
			notification.RecipientName,
			notification.RecipientRole,
			notification.RecipientBranch,
			notification.Location,
		)
	}

	return nil
}

func buildOperatorNotificationMessage(
	notification models.MachineNotification,
) (string, error) {
	switch strings.ToUpper(
		strings.TrimSpace(notification.ReasonCode),
	) {
	case models.NotificationTypeMachineBroken:
		return buildMachineBrokenMessage(
			notification,
		), nil

	case models.NotificationTypeWaitHanca:
		return buildWaitHancaMessage(
			notification,
		), nil

	default:
		return "", fmt.Errorf(
			"jenis notifikasi tidak dikenal: %s",
			notification.ReasonCode,
		)
	}
}

func buildMachineBrokenMessage(
	notification models.MachineNotification,
) string {
	return fmt.Sprintf(
		"🔧 LAPORAN MESIN RUSAK\n\n"+
			"ID Laporan : %d\n"+
			"Mesin      : %s\n"+
			"Lokasi     : %s\n"+
			"Operator   : %s - %s\n"+
			"Keluhan    : %s\n"+
			"Catatan    : %s\n"+
			"Waktu      : %s\n\n"+
			"Mohon segera dilakukan pengecekan.",

		notification.OperatorNoteID,

		valueOrDashNotification(
			notification.MachineName,
		),

		valueOrDashNotification(
			notification.Location,
		),

		valueOrDashNotification(
			notification.OperatorNIK,
		),

		valueOrDashNotification(
			notification.OperatorName,
		),

		valueOrDashNotification(
			notification.ReasonName,
		),

		valueOrDashNotification(
			notification.Note,
		),

		notification.EventCreatedAt.Format(
			"02-01-2006 15:04:05",
		),
	)
}

func buildWaitHancaMessage(
	notification models.MachineNotification,
) string {
	return fmt.Sprintf(
		"📦 OPERATOR MENUNGGU HANCA\n\n"+
			"ID Laporan : %d\n"+
			"Mesin      : %s\n"+
			"Lokasi     : %s\n"+
			"Operator   : %s - %s\n"+
			"Keterangan : %s\n"+
			"Catatan    : %s\n"+
			"Waktu      : %s\n\n"+
			"Notifikasi ini dikirim kepada SPV yang bertanggung jawab pada line tersebut.\n"+
			"Mohon segera dilakukan pengecekan supply hanca.",

		notification.OperatorNoteID,

		valueOrDashNotification(
			notification.MachineName,
		),

		valueOrDashNotification(
			notification.Location,
		),

		valueOrDashNotification(
			notification.OperatorNIK,
		),

		valueOrDashNotification(
			notification.OperatorName,
		),

		valueOrDashNotification(
			notification.ReasonName,
		),

		valueOrDashNotification(
			notification.Note,
		),

		notification.EventCreatedAt.Format(
			"02-01-2006 15:04:05",
		),
	)
}

func valueOrDashNotification(
	value string,
) string {
	value = strings.TrimSpace(value)

	if value == "" {
		return "-"
	}

	return value
}
