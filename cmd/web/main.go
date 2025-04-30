package main

import (
	"fmt"
	"html/template"
	"mlue/internal/handlers"
	"mlue/internal/middleware"
	"mlue/internal/models"
	"mlue/internal/repository"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// load env from file
	godotenv.Load()

	// init database
	var db, err = gorm.Open(sqlite.Open("mlue.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to the db")
	}

	db.AutoMigrate(&models.User{}, &models.Color{})

	// init redis
	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379")
		},
	}

	// init scs
	sessionManager := scs.New()
	sessionManager.Store = redisstore.New(pool)
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Secure = false
	// init oauth
	clientid := os.Getenv("GOOGLE_KEY")
	clientSecret := os.Getenv("GOOGLE_SECRET")

	conf := &oauth2.Config{
		ClientID:     clientid,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:3000/auth/callback",
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}

	// tamplates
	templates := template.Must(template.ParseGlob("templates/*.html"))

	// repos
	colorRepo := repository.NewColorRepo(db)
	userRepo := repository.NewUserRepo(db)

	// handlers
	authHandler := handlers.NewAuthHandler(db, sessionManager, conf, templates)
	colorHandler := handlers.NewColorHandler(colorRepo, templates)
	userHandler := handlers.NewUserHandler(userRepo, templates)

	// middlewares
	requireUser := middleware.RequireUser(sessionManager, db)

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.HandleFunc("/auth/oauth", authHandler.OAuth)
	mux.HandleFunc("/auth/callback", authHandler.OAuthCallback)
	mux.HandleFunc("/auth/logout", authHandler.Logout)

	mux.Handle("/colors/new", requireUser(http.HandlerFunc(colorHandler.CreateColorForm)))
	mux.Handle("/colors/", requireUser(http.HandlerFunc(colorHandler.GetColor)))
	mux.Handle("POST /colors/create", requireUser(http.HandlerFunc(colorHandler.CreateColor)))
	mux.Handle("/", requireUser(http.HandlerFunc(colorHandler.ListColors)))

	mux.Handle("/user", requireUser(http.HandlerFunc(userHandler.GetUser)))

	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./templates/about.html")
	})

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("heloo :)")

	http.ListenAndServe(":3000", sessionManager.LoadAndSave(mux))
}
