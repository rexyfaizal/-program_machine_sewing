package models

type Employee struct {
	NIK          string `json:"nik"`
	Name         string `json:"name"`
	BranchDetail string `json:"branchdetail"`
}
