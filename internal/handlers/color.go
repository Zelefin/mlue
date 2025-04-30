package handlers

import (
	"html/template"
	"mlue/internal/middleware"
	"mlue/internal/models"
	"mlue/internal/repository"
	"mlue/internal/service"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

type ColorHandler struct {
	repo  repository.ColorRepo
	tmpls *template.Template
}

type ColorsPageData struct {
	User   *models.User
	Colors []models.Color
}

func NewColorHandler(repo repository.ColorRepo, tmlps *template.Template) *ColorHandler {
	return &ColorHandler{
		repo:  repo,
		tmpls: tmlps,
	}
}

func (h *ColorHandler) ListColors(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	var colors []models.Color
	var err error

	// all vs own based on query
	if r.URL.Query().Get("mine") == "1" {
		colors, err = h.repo.GetByUser(user.ID)
	} else {
		colors, err = h.repo.GetAll()
	}

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// revers by pk
	// changes in memory- no new value
	slices.Reverse(colors)
	data := ColorsPageData{
		User:   user,
		Colors: colors,
	}

	h.tmpls.ExecuteTemplate(w, "colors.html", data)
}

func (h *ColorHandler) CreateColorForm(w http.ResponseWriter, r *http.Request) {
	h.tmpls.ExecuteTemplate(w, "form.html", nil)
}

func (h *ColorHandler) CreateColor(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	r.ParseForm()
	hex := r.PostForm.Get("hex")
	name := r.PostForm.Get("name")

	colorResponse := service.CallColorApi(hex)

	c := &models.Color{
		UserID:        user.ID,
		Hex:           hex,
		UserColorName: name,
		RealColorName: colorResponse.Name,
		Match:         colorResponse.Match,
		Palette:       colorResponse.Palette,
	}
	if err := h.repo.Create(c); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *ColorHandler) GetColor(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/colors/")

	id64, err := strconv.ParseUint(path, 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	id := uint(id64)
	color, err := h.repo.Get(uint(id))

	// actually we can check for not found error, but I'm too lazy
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	paletteSlice := []string{}
	if color.Palette != "" {
		paletteSlice = strings.Split(color.Palette, ",")
	}

	data := struct {
		Color        models.Color
		PaletteSlice []string
	}{
		Color:        color,
		PaletteSlice: paletteSlice,
	}
	h.tmpls.ExecuteTemplate(w, "color.html", data)
}
