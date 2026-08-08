package config

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

func GetEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func ConnectDB() (*sql.DB, error) {
	// Prioritas: SEWINGIOT_DB_SERVER (khusus project) > DB_SERVER > default.
	// Hindari User env lama (mis. 10.5.0.106) mengalahkan server produksi project.
	server := GetEnv("SEWINGIOT_DB_SERVER", "")
	if server == "" {
		server = GetEnv("DB_SERVER", "10.5.0.107")
	}
	port := GetEnv("DB_PORT", "1433")
	user := GetEnv("DB_USER", "sa")
	password := GetEnv("DB_PASSWORD", "Satu1234!")

	// Project ini selalu pakai sewingiot.
	// Jangan baca DB_NAME global (bisa "lectra" dari User env project lain).
	// Override khusus project ini hanya lewat SEWINGIOT_DB_NAME.
	database := GetEnv("SEWINGIOT_DB_NAME", "sewingiot")

	q := url.Values{}
	q.Add("database", database)
	q.Add("encrypt", "disable")

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(user, password),
		Host:     fmt.Sprintf("%s:%s", server, port),
		RawQuery: q.Encode(),
	}

	fmt.Printf("Connecting SQL Server %s:%s database=%s user=%s\n", server, port, database, user)

	db, err := sql.Open("sqlserver", u.String())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
