package models

type MechanicIdentifyRequest struct {
	NIK  string `json:"nik"`
	RFID string `json:"rfid"`
	Code string `json:"code"` // NIK atau RFID (tap kartu)
}

type MechanicIdentifyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	IsValid bool   `json:"isValid"`
	NIK     string `json:"nik"`
	Name    string `json:"name"`
	Bagian string `json:"bagian"`
	RFID    string `json:"rfid,omitempty"`
}

type MechanicRFIDRegisterRequest struct {
	NIK    string `json:"nik"`
	RFIDNo string `json:"rfidNo"`
}

type MechanicRFIDRegisterResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	NIK     string `json:"nik"`
	Name    string `json:"name"`
	RFIDNo  string `json:"rfidNo"`
}

type MechanicBrokenMachine struct {
	ID              int64  `json:"id"`
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
	TicketStatus    string `json:"ticketStatus"` // OPEN | IN_PROGRESS | DONE

	ClaimedByNIK  string `json:"claimedByNik"`
	ClaimedByName string `json:"claimedByName"`
	ClaimedAt     string `json:"claimedAt"`

	MechanicDoneAt     string `json:"mechanicDoneAt"`
	MechanicDoneByNIK  string `json:"mechanicDoneByNik"`
	MechanicDoneByName string `json:"mechanicDoneByName"`

	// Durasi transparan (detik)
	WaitMechanicSeconds    int64 `json:"waitMechanicSeconds"`
	MechanicWorkSeconds    int64 `json:"mechanicWorkSeconds"`
	OperatorLossSeconds    int64 `json:"operatorLossSeconds"`
	OperatorStillActive    bool  `json:"operatorStillActive"`
	ClosedByMechanic       bool  `json:"closedByMechanic"`
	ClosedByOperator       bool  `json:"closedByOperator"`
}

type MechanicBrokenListResponse struct {
	Status    string                  `json:"status"`
	Total     int                     `json:"total"`
	Open      int                     `json:"open"`
	Busy      int                     `json:"inProgress"`
	DoneToday int                     `json:"doneToday"`
	Rows      []MechanicBrokenMachine `json:"rows"`
}

type MechanicClaimRequest struct {
	ID           int64  `json:"id"`
	MechanicNIK  string `json:"mechanicNik"`
	MechanicName string `json:"mechanicName"`
}

type MechanicDoneRequest struct {
	ID          int64  `json:"id"`
	MechanicNIK string `json:"mechanicNik"`
}

type MechanicActionResponse struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Item    *MechanicBrokenMachine `json:"item,omitempty"`
}
