package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	BotToken      string
	EmployeeTable string
}

func Load() (*Config, error) {
	botToken := strings.TrimSpace(
		os.Getenv("BOT_TOKEN"),
	)

	if botToken == "" {
		return nil, fmt.Errorf(
			"environment variable BOT_TOKEN belum diisi",
		)
	}

	employeeTable := strings.TrimSpace(
		os.Getenv("DB_EMPLOYEE_TABLE"),
	)

	if employeeTable == "" {
		employeeTable = "dbo.employee"
	}

	return &Config{
		BotToken:      botToken,
		EmployeeTable: employeeTable,
	}, nil
}
