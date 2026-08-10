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
	// Paksa server produksi project ini. Jangan baca DB_SERVER (User env lama = 10.5.0.106).
	const forcedServer = "10.5.0.107"
	server := forcedServer

	port := GetEnv("DB_PORT", "1433")
	user := GetEnv("DB_USER", "sa")
	password := GetEnv("DB_PASSWORD", "Satu1234!")

	// Project ini selalu pakai sewingiot.
	database := "sewingiot"

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
