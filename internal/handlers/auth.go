package handlers

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	"mlue/internal/auth"
	"mlue/internal/models"
	"mlue/internal/utils"
)

type AuthHandler struct {
	db             *gorm.DB
	sessionManager *scs.SessionManager
	oauthConfig    *oauth2.Config
	tmpls          *template.Template
}

func NewAuthHandler(db *gorm.DB, sm *scs.SessionManager, cfg *oauth2.Config, tmpls *template.Template) *AuthHandler {
	return &AuthHandler{db: db, sessionManager: sm, oauthConfig: cfg, tmpls: tmpls}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.tmpls.ExecuteTemplate(w, "login.html", nil)
}

func (h *AuthHandler) OAuth(w http.ResponseWriter, r *http.Request) {
	state := utils.GenerateRandomState()
	h.sessionManager.Put(r.Context(), "oauthState", state)
	url := h.oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	// check for state
	state := r.URL.Query().Get("state")
	expectedState := h.sessionManager.GetString(r.Context(), "oauthState")
	if state != expectedState {
		// CSRF vul
		http.Error(w, "Invalid state parameter", http.StatusForbidden)
		return
	}
	// remove as not needed
	h.sessionManager.Remove(r.Context(), "oauthState")

	// Exchanging the code for an access token
	t, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Creating an HTTP client to make authenticated request using the access key.
	// This client method also regenerate the access key using the refresh key.
	client := h.oauthConfig.Client(context.Background(), t)

	// Getting the user public details from google API endpoint
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Closing the request body when this function returns.
	// This is a good practice to avoid memory leak
	defer resp.Body.Close()

	var gp auth.GoogleProfile

	// Reading the JSON body using JSON decoder
	err = json.NewDecoder(resp.Body).Decode(&gp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var user models.User
	result := h.db.Where("google_sub = ?", gp.ID).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		user = models.User{
			GoogleSub:  gp.ID,
			Email:      gp.Email,
			Name:       gp.Name,
			PictureURL: gp.Picture,
		}
		h.db.Create(&user)
	} else {
		h.db.Model(&user).Updates(models.User{
			Email:      gp.Email,
			Name:       gp.Name,
			PictureURL: gp.Picture,
		})
	}

	h.sessionManager.Put(r.Context(), "userID", user.ID)
	h.sessionManager.Put(r.Context(), "userPFP", user.PictureURL)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.sessionManager.Destroy(r.Context())
	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}
