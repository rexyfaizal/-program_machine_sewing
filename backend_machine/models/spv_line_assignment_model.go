package models

type SpvLineAssignment struct {
	Location     string `json:"location"`
	SpvNik       string `json:"spvNik"`
	SpvName      string `json:"spvName"`
	BranchDetail string `json:"branchdetail"`
	IsActive     bool   `json:"isActive"`
}

type SpvLineAssignmentPutRequest struct {
	Location string `json:"location"`
	SpvNik   string `json:"spvNik"`
}

type SpvLineAssignmentRenameRequest struct {
	OldLocation string `json:"oldLocation"`
	NewLocation string `json:"newLocation"`
}
