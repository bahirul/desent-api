package main

import (
	"fmt"
	"net/http"

	"desent-api/configs"
	"desent-api/internal/handlers"
	"desent-api/internal/middlewares"
	"desent-api/internal/utils"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := configs.Load()

	loggers, err := utils.NewLoggers(cfg.Logging)
	if err != nil {
		panic(fmt.Sprintf("init loggers: %v", err))
	}
	defer loggers.Close()

	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middlewares.HTTPLogger(loggers.HTTP))

	r.Get("/ping", handlers.Ping)
	r.Post("/echo", handlers.Echo)

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
