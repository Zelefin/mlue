package middleware

import (
	"context"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"gorm.io/gorm"
	"mlue/internal/models"
)

type contextKey string

const UserContextKey = contextKey("user")

func RequireUser(session *scs.SessionManager, db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := session.Get(r.Context(), "userID")
			if id == 0 {
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}

			var user models.User
			if err := db.First(&user, id).Error; err != nil {
				session.Destroy(r.Context())
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
