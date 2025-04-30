package handlers

import (
	"html/template"
	"mlue/internal/middleware"
	"mlue/internal/models"
	"mlue/internal/repository"
	"net/http"
)

type UserHandler struct {
	repo  repository.UserRepo
	tmpls *template.Template
}

func NewUserHandler(repo repository.UserRepo, tmpls *template.Template) *UserHandler {
	return &UserHandler{
		repo:  repo,
		tmpls: tmpls,
	}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	userColorsCount := h.repo.GetCountUserColors(user.ID)

	data := struct {
		User        *models.User
		ColorsCount int64
	}{
		User:        user,
		ColorsCount: userColorsCount,
	}
	h.tmpls.ExecuteTemplate(w, "user.html", data)
}
