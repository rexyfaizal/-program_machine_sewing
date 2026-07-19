package models

type MachineSetting struct {
	UUID       string `json:"uuid"`
	CustomName string `json:"customName"`
	Location   string `json:"location"`
	Pic        string `json:"pic"`
	Spv        string `json:"spv"`
	UpdatedAt  string `json:"updatedAt"`
}
