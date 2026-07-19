package models

type AlarmStats struct {
	AlarmCount int64
	AlarmTypes string
}

type AlarmGroup struct {
	Content string `json:"content"`
	Total   int64  `json:"total"`
}
