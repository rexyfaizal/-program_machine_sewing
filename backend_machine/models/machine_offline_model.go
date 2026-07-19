package models

type MachineOfflineAutoLogoutResponse struct {
	Status             string `json:"status"`
	Message            string `json:"message"`
	CheckedAt          string `json:"checkedAt"`
	TotalMachines      int    `json:"totalMachines"`
	OnlineMachines     int    `json:"onlineMachines"`
	OfflineMachines    int    `json:"offlineMachines"`
	AutoLogoutSessions int64  `json:"autoLogoutSessions"`
}
