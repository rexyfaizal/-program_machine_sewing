package models

type MachineOperatorLoginRequest struct {
	UUID         string `json:"uuid"`
	MachineName  string `json:"machineName"`
	Location     string `json:"location"`
	OperatorNIK  string `json:"operatorNik"`
	OperatorName string `json:"operatorName"`
	BranchDetail string `json:"branchdetail"`
	ProcessName  string `json:"processName"`
	StyleName    string `json:"styleName"`
}

type MachineOperatorSession struct {
	ID           int64  `json:"id"`
	SessionDate  string `json:"sessionDate"`
	UUID         string `json:"uuid"`
	MachineName  string `json:"machineName"`
	Location     string `json:"location"`
	OperatorNIK  string `json:"operatorNik"`
	OperatorName string `json:"operatorName"`
	BranchDetail string `json:"branchdetail"`
	ProcessName  string `json:"processName"`
	StyleName    string `json:"styleName"`
	LoginTime    string `json:"loginTime"`
	LogoutTime   string `json:"logoutTime"`
	Status       string `json:"status"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type MachineOperatorLoginResponse struct {
	Status     string                 `json:"status"`
	Message    string                 `json:"message"`
	IsExisting bool                   `json:"isExisting"`
	Session    MachineOperatorSession `json:"session"`
}

type MachineOperatorNoteRequest struct {
	UUID        string `json:"uuid"`
	OperatorNIK string `json:"operatorNik"`
	ReasonCode  string `json:"reasonCode"`
	ReasonName  string `json:"reasonName"`
	Note        string `json:"note"`
}

type MachineOperatorNote struct {
	ID           int64  `json:"id"`
	SessionID    int64  `json:"sessionId"`
	SessionDate  string `json:"sessionDate"`
	UUID         string `json:"uuid"`
	OperatorNIK  string `json:"operatorNik"`
	OperatorName string `json:"operatorName"`
	ReasonCode   string `json:"reasonCode"`
	ReasonName   string `json:"reasonName"`
	Note         string `json:"note"`
	CreatedAt    string `json:"createdAt"`

	EndTime           string `json:"endTime"`
	DurationSeconds   int64  `json:"durationSeconds"`
	DurationText      string `json:"durationText"`
	Status            string `json:"status"`
	IsActiveLossEvent bool   `json:"isActiveLossEvent"`
}

type MachineOperatorNoteResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	NoteID    int64  `json:"noteId"`
	SessionID int64  `json:"sessionId"`
	Closed    bool   `json:"closed"`
}

type MachineOperatorReportItem struct {
	MachineOperatorSession

	ActiveLossReasonCode      string `json:"activeLossReasonCode,omitempty"`
	ActiveLossReasonLabel     string `json:"activeLossReasonLabel,omitempty"`
	ActiveLossStartTime       string `json:"activeLossStartTime,omitempty"`
	ActiveLossDurationSeconds int64  `json:"activeLossDurationSeconds,omitempty"`
	ActiveLossDurationText    string `json:"activeLossDurationText,omitempty"`
	ActiveLossStatus          string `json:"activeLossStatus,omitempty"`

	// Stats mesin pada jendela login–logout session (untuk export Excel).
	HasSessionStats        bool    `json:"hasSessionStats"`
	RuntimeSec             int64   `json:"runtimeSec"`
	ProcSec                int64   `json:"procSec"`
	LossTimeSec            int64   `json:"lossTimeSec"`
	ProductivityPct        float64 `json:"productivityPct"`
	ProductivityStatus     string  `json:"productivityStatus"`
	Output                 int64   `json:"output"`
	AvgCycle               float64 `json:"avgCycle"`

	Notes []MachineOperatorNote `json:"notes"`
}

type MachineOperatorActiveResponse struct {
	Active  bool                    `json:"active"`
	Message string                  `json:"message"`
	Session *MachineOperatorSession `json:"session"`
}
