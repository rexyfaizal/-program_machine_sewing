package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"telegram_notif/models"
)

var tableNamePattern = regexp.MustCompile(
	`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)?$`,
)

type EmployeeRepository interface {
	RegisterTelegramID(
		ctx context.Context,
		nik string,
		bagian string,
		telegramID int64,
	) (models.RegistrationResult, error)
}

type employeeRepository struct {
	db    *sql.DB
	table string
}

func NewEmployeeRepository(
	db *sql.DB,
	table string,
) (EmployeeRepository, error) {
	if db == nil {
		return nil, errors.New("database tidak boleh nil")
	}

	table = strings.TrimSpace(table)

	if !tableNamePattern.MatchString(table) {
		return nil, fmt.Errorf(
			"nama tabel tidak valid: %s",
			table,
		)
	}

	return &employeeRepository{
		db:    db,
		table: table,
	}, nil
}

func (r *employeeRepository) RegisterTelegramID(
	ctx context.Context,
	nik string,
	bagian string,
	telegramID int64,
) (models.RegistrationResult, error) {
	nik = strings.TrimSpace(nik)
	bagian = strings.TrimSpace(bagian)

	if nik == "" {
		return models.RegistrationResult{},
			errors.New("NIK tidak boleh kosong")
	}

	if bagian == "" {
		return models.RegistrationResult{},
			errors.New("bagian tidak boleh kosong")
	}

	if telegramID <= 0 {
		return models.RegistrationResult{},
			errors.New("Telegram ID tidak valid")
	}

	tx, err := r.db.BeginTx(
		ctx,
		&sql.TxOptions{
			Isolation: sql.LevelSerializable,
		},
	)
	if err != nil {
		return models.RegistrationResult{}, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	/*
		Periksa apakah Telegram ID sudah digunakan
		oleh NIK lain.
	*/
	conflictingNIK, err := r.findNIKByTelegramID(
		ctx,
		tx,
		telegramID,
	)
	if err != nil {
		return models.RegistrationResult{}, err
	}

	if conflictingNIK != "" && conflictingNIK != nik {
		if err := tx.Commit(); err != nil {
			return models.RegistrationResult{}, err
		}

		return models.RegistrationResult{
			Status:         models.RegistrationTelegramUsed,
			ConflictingNIK: conflictingNIK,
		}, nil
	}

	/*
		Cari data karyawan berdasarkan NIK.
	*/
	employees, err := r.findEmployeesByNIK(
		ctx,
		tx,
		nik,
	)
	if err != nil {
		return models.RegistrationResult{}, err
	}

	if len(employees) == 0 {
		if err := tx.Commit(); err != nil {
			return models.RegistrationResult{}, err
		}

		return models.RegistrationResult{
			Status: models.RegistrationNIKNotFound,
		}, nil
	}

	if len(employees) > 1 {
		if err := tx.Commit(); err != nil {
			return models.RegistrationResult{}, err
		}

		return models.RegistrationResult{
			Status:   models.RegistrationDuplicateNIK,
			Employee: employees[0],
		}, nil
	}

	employee := employees[0]

	/*
		Jika NIK sudah memiliki Telegram ID yang berbeda,
		registrasi ditolak.
	*/
	if employee.TelegramID != nil &&
		*employee.TelegramID != telegramID {

		if err := tx.Commit(); err != nil {
			return models.RegistrationResult{}, err
		}

		return models.RegistrationResult{
			Status:   models.RegistrationNIKUsed,
			Employee: employee,
		}, nil
	}

	/*
		Kondisi yang diperbolehkan:

		1. id_telegram masih NULL:
		   simpan Telegram ID dan bagian.

		2. id_telegram sudah sama:
		   tetap update bagian, termasuk jika bagian
		   sebelumnya NULL.
	*/
	if err := r.assignRegistrationData(
		ctx,
		tx,
		nik,
		bagian,
		telegramID,
	); err != nil {
		return models.RegistrationResult{}, err
	}

	employee.TelegramID = &telegramID
	employee.Bagian = bagian

	status := models.RegistrationSuccess

	if conflictingNIK == nik {
		status = models.RegistrationAlreadyRegistered
	}

	if err := tx.Commit(); err != nil {
		return models.RegistrationResult{}, err
	}

	return models.RegistrationResult{
		Status:   status,
		Employee: employee,
	}, nil
}

func (r *employeeRepository) findNIKByTelegramID(
	ctx context.Context,
	tx *sql.Tx,
	telegramID int64,
) (string, error) {
	query := fmt.Sprintf(`
		SELECT TOP (1)
			LTRIM(RTRIM(
				ISNULL(CONVERT(varchar(255), nik), '')
			))
		FROM %s WITH (UPDLOCK, HOLDLOCK)
		WHERE id_telegram = @telegram_id;
	`, r.table)

	var nik string

	err := tx.QueryRowContext(
		ctx,
		query,
		sql.Named("telegram_id", telegramID),
	).Scan(&nik)

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(nik), nil
}

func (r *employeeRepository) findEmployeesByNIK(
	ctx context.Context,
	tx *sql.Tx,
	nik string,
) ([]models.Employee, error) {
	query := fmt.Sprintf(`
		SELECT
			LTRIM(RTRIM(
				ISNULL(CONVERT(varchar(255), nik), '')
			)) AS nik,

			LTRIM(RTRIM(
				ISNULL(CONVERT(varchar(255), name), '')
			)) AS name,

			LTRIM(RTRIM(
				ISNULL(CONVERT(varchar(255), branchdetail), '')
			)) AS branchdetail,

			LTRIM(RTRIM(
				ISNULL(CONVERT(varchar(255), bagian), '')
			)) AS bagian,

			id_telegram
		FROM %s WITH (UPDLOCK, HOLDLOCK)
		WHERE LTRIM(RTRIM(
			CONVERT(varchar(255), nik)
		)) = @nik;
	`, r.table)

	rows, err := tx.QueryContext(
		ctx,
		query,
		sql.Named("nik", nik),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	employees := make(
		[]models.Employee,
		0,
		1,
	)

	for rows.Next() {
		var employee models.Employee
		var telegramID sql.NullInt64

		if err := rows.Scan(
			&employee.NIK,
			&employee.Name,
			&employee.BranchDetail,
			&employee.Bagian,
			&telegramID,
		); err != nil {
			return nil, err
		}

		if telegramID.Valid {
			value := telegramID.Int64
			employee.TelegramID = &value
		}

		employees = append(
			employees,
			employee,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return employees, nil
}

func (r *employeeRepository) assignRegistrationData(
	ctx context.Context,
	tx *sql.Tx,
	nik string,
	bagian string,
	telegramID int64,
) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET
			id_telegram = @telegram_id,
			bagian = @bagian
		WHERE LTRIM(RTRIM(
			CONVERT(varchar(255), nik)
		)) = @nik
		  AND (
				id_telegram IS NULL
				OR id_telegram = @telegram_id
		  );
	`, r.table)

	result, err := tx.ExecContext(
		ctx,
		query,
		sql.Named("telegram_id", telegramID),
		sql.Named("bagian", bagian),
		sql.Named("nik", nik),
	)
	if err != nil {
		return fmt.Errorf(
			"gagal memperbarui data registrasi: %w",
			err,
		)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affectedRows != 1 {
		return fmt.Errorf(
			"jumlah baris yang diperbarui tidak sesuai: %d",
			affectedRows,
		)
	}

	return nil
}
