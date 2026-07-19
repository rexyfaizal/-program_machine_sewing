package handlers

import "backend_machine/repository"

type Handler struct {
	Repo *repository.Repository
}

func New(repo *repository.Repository) *Handler {
	return &Handler{Repo: repo}
}
