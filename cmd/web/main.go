package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aitumik/snippetbox/pkg/models/postgres"

	"github.com/aitumik/snippetbox/pkg"
	"github.com/aitumik/snippetbox/pkg/models"
	"github.com/golangcollege/sessions"
	psql "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type contextKey string

var contextKeyUser = contextKey("user")

type application struct {
	errorLogger *log.Logger
	infoLogger  *log.Logger
	session     *sessions.Session
	cfg         *pkg.Config
	snippet     interface {
		Insert(title, content, expires string) (int, error)
		Get(id int) (*models.Snippet, error)
		Latest() ([]*models.Snippet, error)
	}
	templateCache map[string]*template.Template
	users         interface {
		Insert(name, email, password string) error
		Authenticate(email, password string) (int, error)
		Get(id int) (*models.User, error)
	}
}

func main() {

	infoLogger := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLogger := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile|log.Llongfile)

	cfg, err := pkg.NewConfig()
	if err != nil {
		errorLogger.Fatal(err)
	}

	flag.StringVar(&cfg.Addr, "addr", cfg.Addr, "HTTP Network Address")
	flag.StringVar(&cfg.StaticDir, "static-dir", cfg.StaticDir, "Path to static assets")
	flag.StringVar(&cfg.SecretKey, "secret", cfg.SecretKey, "Secret Key")
	flag.Parse()

	dsn := cfg.DatabaseURI

	conn, err := gorm.Open(psql.Open(dsn), &gorm.Config{})
	if err != nil {
		errorLogger.Fatal(err)
	}

	db, err := conn.DB()
	if err != nil {
		errorLogger.Fatal(err)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			errorLogger.Fatal(err)
		}
	}(db)

	templateCache, err := NewTemplateCache("./ui/html/")
	if err != nil {
		errorLogger.Fatal(err)
	}
	infoLogger.Print("Initializing the template cache")

	session := sessions.New([]byte(cfg.SecretKey))
	session.Lifetime = 12 * time.Hour
	session.Secure = false
	session.SameSite = http.SameSiteStrictMode

	app := &application{
		errorLogger: errorLogger,
		infoLogger:  infoLogger,
		session:     session,
		cfg:         cfg,
		snippet: &postgres.SnippetModel{
			DB: db,
		},
		templateCache: templateCache,
		users: &postgres.UserModel{
			DB: db,
		},
	}

	conn.AutoMigrate(&models.Snippet{})
	conn.AutoMigrate(&models.User{})
	infoLogger.Print("Migrating database models")

	mux := app.routes()

	server := &http.Server{
		Addr:         cfg.Addr,
		ErrorLog:     errorLogger,
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLogger.Printf("Server started at %s", cfg.Addr)
	err = server.ListenAndServe()
	errorLogger.Fatal(err)
}
