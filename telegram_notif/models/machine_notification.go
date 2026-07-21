package models

import "time"

const (
	NotificationTypeMachineBroken = "MACHINE_BROKEN"
	NotificationTypeWaitHanca     = "WAIT_HANCA"
)

const (
	NotificationStatusProcessing = "PROCESSING"
	NotificationStatusSent       = "SENT"
	NotificationStatusFailed     = "FAILED"
)

type MachineNotification struct {
	OperatorNoteID int64
	SessionID      int64

	UUID        string
	MachineName string
	Location    string

	OperatorNIK  string
	OperatorName string

	ReasonCode string
	ReasonName string
	Note       string

	EventCreatedAt time.Time

	RecipientNIK    string
	RecipientName   string
	RecipientRole   string
	RecipientBranch string
	TelegramChatID  int64
}
