package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"telegram_notif/models"
	"telegram_notif/repository"
)

var nikPattern = regexp.MustCompile(`^[0-9]{3,30}$`)

const (
	BagianOperator = "Operator"
	BagianSPV      = "SPV"
	BagianMekanik  = "Mekanik"
)

type RegistrationService interface {
	Register(
		ctx context.Context,
		nik string,
		bagian string,
		telegramID int64,
	) (
		message string,
		complete bool,
		err error,
	)
}

type registrationService struct {
	repository repository.EmployeeRepository
}

func NewRegistrationService(
	employeeRepository repository.EmployeeRepository,
) RegistrationService {
	return &registrationService{
		repository: employeeRepository,
	}
}

func (s *registrationService) Register(
	ctx context.Context,
	nik string,
	bagian string,
	telegramID int64,
) (string, bool, error) {
	nik = strings.TrimSpace(nik)

	if !nikPattern.MatchString(nik) {
		return "Format NIK tidak sesuai.\n\n" +
			"NIK hanya boleh berisi angka.\n" +
			"Contoh: 100009\n\n" +
			"Silakan kirim ulang NIK Anda.", false, nil
	}

	normalizedBagian, valid := normalizeBagian(bagian)
	if !valid {
		return "Bagian yang dipilih tidak valid.\n\n" +
			"Pilihan yang tersedia:\n" +
			"- Operator\n" +
			"- SPV\n" +
			"- Mekanik\n\n" +
			"Silakan ketik /start untuk memulai kembali.", false, nil
	}

	if telegramID <= 0 {
		return "", false, fmt.Errorf(
			"Telegram ID tidak valid: %d",
			telegramID,
		)
	}

	result, err := s.repository.RegisterTelegramID(
		ctx,
		nik,
		normalizedBagian,
		telegramID,
	)
	if err != nil {
		return "", false, err
	}

	switch result.Status {
	case models.RegistrationSuccess:
		return employeeSuccessMessage(
			"Pendaftaran berhasil.",
			result.Employee,
			telegramID,
		), true, nil

	case models.RegistrationAlreadyRegistered:
		return employeeSuccessMessage(
			"Data akun Telegram Anda berhasil diperbarui.",
			result.Employee,
			telegramID,
		), true, nil

	case models.RegistrationTelegramUsed:
		return fmt.Sprintf(
			"ID Telegram Anda sudah terdaftar pada NIK %s.\n\n"+
				"Satu akun Telegram hanya dapat digunakan untuk satu NIK.\n"+
				"Silakan hubungi administrator.",
			valueOrDash(result.ConflictingNIK),
		), false, nil

	case models.RegistrationNIKUsed:
		return fmt.Sprintf(
			"NIK %s sudah terhubung dengan akun Telegram lain.\n\n"+
				"Silakan hubungi administrator apabila ingin mengganti akun Telegram.",
			valueOrDash(result.Employee.NIK),
		), false, nil

	case models.RegistrationNIKNotFound:
		return fmt.Sprintf(
			"NIK %s tidak ditemukan pada database karyawan.\n\n"+
				"Periksa kembali NIK Anda lalu kirim ulang.",
			nik,
		), false, nil

	case models.RegistrationDuplicateNIK:
		return fmt.Sprintf(
			"NIK %s ditemukan lebih dari satu kali pada database.\n\n"+
				"Silakan hubungi administrator untuk memperbaiki data NIK ganda.",
			nik,
		), false, nil

	default:
		return "", false, fmt.Errorf(
			"status registrasi tidak dikenal: %v",
			result.Status,
		)
	}
}

func employeeSuccessMessage(
	title string,
	employee models.Employee,
	telegramID int64,
) string {
	return fmt.Sprintf(
		"%s\n\n"+
			"NIK: %s\n"+
			"Nama: %s\n"+
			"Branch: %s\n"+
			"Bagian: %s\n"+
			"ID Telegram: %d\n\n"+
			"Akun Telegram Anda sudah terhubung dengan database karyawan.",
		title,
		valueOrDash(employee.NIK),
		valueOrDash(employee.Name),
		valueOrDash(employee.BranchDetail),
		valueOrDash(employee.Bagian),
		telegramID,
	)
}

func normalizeBagian(
	bagian string,
) (string, bool) {
	switch strings.ToLower(
		strings.TrimSpace(bagian),
	) {
	case "operator":
		return BagianOperator, true

	case "spv":
		return BagianSPV, true

	case "mekanik":
		return BagianMekanik, true

	default:
		return "", false
	}
}

func valueOrDash(
	value string,
) string {
	value = strings.TrimSpace(value)

	if value == "" {
		return "-"
	}

	return value
}
