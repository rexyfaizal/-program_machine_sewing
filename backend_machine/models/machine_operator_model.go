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
	Notes []MachineOperatorNote `json:"notes"`
}
