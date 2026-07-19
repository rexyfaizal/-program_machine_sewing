package models

type Employee struct {
	NIK          string
	Name         string
	BranchDetail string
	TelegramID   *int64
}

type RegistrationStatus string

const (
	RegistrationSuccess           RegistrationStatus = "success"
	RegistrationAlreadyRegistered RegistrationStatus = "already_registered"
	RegistrationTelegramUsed      RegistrationStatus = "telegram_used"
	RegistrationNIKUsed           RegistrationStatus = "nik_used"
	RegistrationNIKNotFound       RegistrationStatus = "nik_not_found"
	RegistrationDuplicateNIK      RegistrationStatus = "duplicate_nik"
)

type RegistrationResult struct {
	Status         RegistrationStatus
	Employee       Employee
	ConflictingNIK string
}
