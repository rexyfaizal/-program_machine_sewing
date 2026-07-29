package models

type MachineOperatorLossEventStartRequest struct {
	UUID        string `json:"uuid"`
	ReasonCode  string `json:"reasonCode"`
	ReasonLabel string `json:"reasonLabel"`
	Note        string `json:"note"`
}

type MachineOperatorLossEventFinishRequest struct {
	UUID string `json:"uuid"`
}

type MachineOperatorLossEvent struct {
	ID              int64  `json:"id"`
	SessionID       int64  `json:"sessionId"`
	SessionDate     string `json:"sessionDate"`
	UUID            string `json:"uuid"`
	MachineName     string `json:"machineName"`
	Location        string `json:"location"`
	OperatorNIK     string `json:"operatorNik"`
	OperatorName    string `json:"operatorName"`
	ReasonCode      string `json:"reasonCode"`
	ReasonLabel     string `json:"reasonLabel"`
	Note            string `json:"note"`
	StartTime       string `json:"startTime"`
	EndTime         string `json:"endTime"`
	DurationSeconds int64  `json:"durationSeconds"`
	Status          string `json:"status"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type MachineOperatorLossEventStartResponse struct {
	Status     string                   `json:"status"`
	Message    string                   `json:"message"`
	IsExisting bool                     `json:"isExisting"`
	Event      MachineOperatorLossEvent `json:"event"`
}

type MachineOperatorLossEventFinishResponse struct {
	Status  string                   `json:"status"`
	Message string                   `json:"message"`
	Event   MachineOperatorLossEvent `json:"event"`
}

type MachineOperatorLossEventActiveResponse struct {
	Status         string                    `json:"status"`
	HasActiveEvent bool                      `json:"hasActiveEvent"`
	Event          *MachineOperatorLossEvent `json:"event"`
}
