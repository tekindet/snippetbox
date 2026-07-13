package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aitumik/snippetbox/pkg"
	"github.com/aitumik/snippetbox/pkg/models"
	"github.com/aitumik/snippetbox/pkg/models/postgres"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golangcollege/sessions"
)

type contextKey string

var contextKeyUser = contextKey("user")

type application struct {
	logger  *slog.Logger
	session *sessions.Session
	cfg     *pkg.Config
	snippet interface {
		Insert(title, content, expires string, tagIDs []int, userID int) (int, error)
		Get(id int) (*models.Snippet, error)
		Latest() ([]*models.Snippet, error)
		GetByUser(userID int) ([]*models.Snippet, error)
		GetByTag(tagID int) ([]*models.Snippet, error)
	}
	templateCache map[string]*template.Template
	users         interface {
		Insert(name, email, password string) error
		Authenticate(email, password string) (int, error)
		Get(id int) (*models.User, error)
	}
	tags interface {
		Insert(name string) (int, error)
		GetByName(name string) (*models.Tag, error)
		GetForSnippet(snippetID int) ([]*models.Tag, error)
		GetAll() ([]*models.Tag, error)
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := pkg.NewConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	flag.StringVar(&cfg.Addr, "addr", cfg.Addr, "HTTP Network Address")
	flag.StringVar(&cfg.StaticDir, "static-dir", cfg.StaticDir, "Path to static assets")
	flag.StringVar(&cfg.SecretKey, "secret", cfg.SecretKey, "Secret Key")
	flag.Parse()

	db, err := sql.Open("postgres", cfg.DatabaseURI)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	if err := db.Ping(); err != nil {
		logger.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	logger.Info("database connection established")

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	m, err := migrate.New("file://migrations", cfg.DatabaseURI)
	if err != nil {
		logger.Error("failed to create migrate instance", "error", err)
		os.Exit(1)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("database migrations applied")

	templateCache, err := NewTemplateCache("./ui/html/")
	if err != nil {
		logger.Error("failed to initialize template cache", "error", err)
		os.Exit(1)
	}
	logger.Info("template cache initialized")

	session := sessions.New([]byte(cfg.SecretKey))
	session.Lifetime = 12 * time.Hour
	session.Secure = false
	session.SameSite = http.SameSiteStrictMode

	app := &application{
		logger:        logger,
		session:       session,
		cfg:           cfg,
		snippet:       &postgres.SnippetModel{DB: db},
		templateCache: templateCache,
		users:         &postgres.UserModel{DB: db},
		tags:          &postgres.TagModel{DB: db},
	}

	mux := app.routes()

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		logger.Info("caught signal", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownError <- server.Shutdown(ctx)
	}()

	logger.Info("server starting", "addr", cfg.Addr)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}

	if err := <-shutdownError; err != nil {
		logger.Error("error during shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")

	if _, err := m.Close(); err != nil {
		logger.Error("error closing migrations", "error", err)
	}

	if err := db.Close(); err != nil {
		logger.Error("error closing database", "error", err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func printVersion() string {
	return fmt.Sprintf("snippetbox %s", "1.0.0")
}
