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
	"strconv"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/glebarez/sqlite"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
			return redis.Dial("tcp", os.Getenv("REDIS_URL"))
		},
	}

	// init scs
	sessionManager := scs.New()
	sessionManager.Store = redisstore.New(pool)
	sessionManager.Lifetime = 24 * time.Hour
	secureCookieValue, err := strconv.ParseBool(os.Getenv("SECURE_COOKIE"))
	if err != nil {
		panic("no valid value for secure cookie")
	}
	sessionManager.Cookie.Secure = secureCookieValue
	// init oauth
	clientid := os.Getenv("CLIENT_KEY")
	clientSecret := os.Getenv("CLIENT_SECRET")

	conf := &oauth2.Config{
		ClientID:     clientid,
		ClientSecret: clientSecret,
		RedirectURL:  os.Getenv("OAUTH_CALLBACK"),
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

	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./robots.txt")
	})

	mux.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./sitemap.xml")
	})

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("heloo :)")

	http.ListenAndServe(os.Getenv("LISTEN_ADDR"), sessionManager.LoadAndSave(mux))
}
