package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"desent-api/configs"
	"desent-api/internal/handlers"
	"desent-api/internal/middlewares"
	"desent-api/internal/repositories"
	"desent-api/internal/usecases"
	"desent-api/internal/utils"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"
)

func main() {
	cfg := configs.Load()

	loggers, err := utils.NewLoggers(cfg.Logging)
	if err != nil {
		panic(fmt.Sprintf("init loggers: %v", err))
	}
	defer loggers.Close()

	db, err := openDatabase(cfg.Database)
	if err != nil {
		panic(fmt.Sprintf("open database: %v", err))
	}
	defer db.Close()

	if err := repositories.InitBooksSchema(context.Background(), db); err != nil {
		panic(fmt.Sprintf("init books schema: %v", err))
	}

	bookRepository := repositories.NewSQLiteBookRepository(db)
	bookHandler := handlers.NewBookHandler(
		usecases.NewCreateBookUsecase(bookRepository),
		usecases.NewListBooksUsecase(bookRepository),
		usecases.NewGetBookUsecase(bookRepository),
		usecases.NewUpdateBookUsecase(bookRepository),
		usecases.NewDeleteBookUsecase(bookRepository),
	)
	authHandler := handlers.NewAuthHandler(cfg.Auth.JWTSecret, time.Duration(cfg.Auth.JWTTTLSeconds)*time.Second)

	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)

	r.Get("/ping", handlers.Ping)
	r.Post("/echo", handlers.Echo)
	r.Post("/auth/token", authHandler.CreateToken)
	r.Post("/books", bookHandler.CreateBook)
	r.With(middlewares.RequireBearerAuth(cfg.Auth.JWTSecret)).Get("/books", bookHandler.ListBooks)
	r.Get("/books/{id}", bookHandler.GetBookByID)
	r.Put("/books/{id}", bookHandler.UpdateBook)
	r.Delete("/books/{id}", bookHandler.DeleteBook)

	srv := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           r,
		ReadTimeout:       cfg.Server.ReadTimeout,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
	}

	loggers.HTTP.Info("server listening", "address", cfg.Server.Address)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		loggers.Error.Error("server failed", "error", err.Error())
		panic(fmt.Sprintf("server failed: %v", err))
	}
}

func openDatabase(cfg configs.DatabaseConfig) (*sql.DB, error) {
	if cfg.Driver == "sqlite" {
		if err := ensureSQLiteDir(cfg.DSN); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func ensureSQLiteDir(dsn string) error {
	if !strings.HasPrefix(dsn, "file:") {
		return nil
	}

	filePart := strings.TrimPrefix(dsn, "file:")
	filePath := strings.SplitN(filePart, "?", 2)[0]
	if filePath == "" || filePath == ":memory:" {
		return nil
	}

	dir := filepath.Dir(filePath)
	if dir == "." || dir == "" {
		return nil
	}

	return os.MkdirAll(dir, 0o755)
}
