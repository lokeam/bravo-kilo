package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"bravo-kilo/cmd/handlers"
	"bravo-kilo/config"
	"bravo-kilo/internal/data"
	"bravo-kilo/internal/driver"
	"bravo-kilo/internal/logger"

	"github.com/joho/godotenv"
)

type appConfig struct {
	port int
}

type application struct {
	config   appConfig
	logger   *slog.Logger
	models   data.Models
}

func main() {
	err := godotenv.Load(".env")
	handler := slog.NewJSONHandler(os.Stdout, nil)
	if err != nil {
		slog.New(handler).Error("Error loading .env file", "error", err)
	}

	logger.Init()
	log := logger.Log

	// Check if upload dir set
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		log.Error("Upload dir is not set")
		os.Exit(1)
	}

	var cfg appConfig
	cfg.port = 8081

	dataSrcName := os.Getenv("DSN")
	db, err := driver.ConnectPostgres(dataSrcName, log)
	if err != nil {
		log.Error("Cannot connect to database", "error", err)
	}

	defer db.SQL.Close()

	models, err := data.New(db.SQL, log)
	if err != nil {
		log.Error("Error initializing data models", "error", err)
		os.Exit(1)
	}

	app := &application{
		config:   cfg,
		logger:   log,
		models:   models,
	}

	// Initialize the config package with the logger
	config.InitConfig(log)

	// Initialize handlers with the logger
	h := handlers.NewHandlers(log, app.models)

	err = app.serve(h)
	if err != nil {
		log.Error("Error initializing ", "error", err)
	}
}

func (app *application) serve(h *handlers.Handlers) error {
	app.logger.Info("API listening on port", "port", app.config.port)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(h),
	}

	return srv.ListenAndServe()
}
