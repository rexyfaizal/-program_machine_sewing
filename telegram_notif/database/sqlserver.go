package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

func ConnectDB() (*sql.DB, error) {
	server := GetEnv("DB_SERVER", "10.5.0.106")
	port := GetEnv("DB_PORT", "1433")
	user := GetEnv("DB_USER", "sa")
	password := strings.TrimSpace(os.Getenv("DB_PASSWORD"))
	databaseName := GetEnv("DB_NAME", "sewingiot")
	encrypt := GetEnv("DB_ENCRYPT", "disable")

	if password == "" {
		return nil, fmt.Errorf("environment variable DB_PASSWORD belum diisi")
	}

	connectionURL := &url.URL{
		Scheme: "sqlserver",
		Host:   net.JoinHostPort(server, port),
		User:   url.UserPassword(user, password),
	}

	query := connectionURL.Query()
	query.Set("database", databaseName)
	query.Set("encrypt", encrypt)
	query.Set("connection timeout", "30")
	connectionURL.RawQuery = query.Encode()

	db, err := sql.Open("sqlserver", connectionURL.String())
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("gagal terhubung ke SQL Server: %w", err)
	}

	return db, nil
}

func GetEnv(key, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	return value
}
