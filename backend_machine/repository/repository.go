package repository

import (
	"context"
	"database/sql"
)

type Repository struct {
	DB *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.DB.PingContext(ctx)
}
