package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	factory "github.com/lokeam/bravo-kilo/cmd/factory"
	"github.com/lokeam/bravo-kilo/config"
	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	"github.com/lokeam/bravo-kilo/internal/shared/driver"
	authHandlers "github.com/lokeam/bravo-kilo/internal/shared/handlers/auth"
	"github.com/lokeam/bravo-kilo/internal/shared/logger"
)

type appConfig struct {
	port int
}

type application struct {
	config   appConfig
	logger   *slog.Logger
}

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	handler := slog.NewJSONHandler(os.Stdout, nil)
	if err != nil {
			slog.New(handler).Error("Error loading .env file", "error", err)
	}

	// Initialize logger
	logger.Init()
	log := logger.Log

	// Check if upload directory is set
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		log.Error("Upload dir is not set")
		os.Exit(1)
	}

	var cfg appConfig
	cfg.port = 8081

	// Connect to the database
	dataSrcName := os.Getenv("DSN")
	db, err := driver.ConnectPostgres(dataSrcName, log)
	if err != nil {
			log.Error("Cannot connect to database", "error", err)
			os.Exit(1)
	}
	defer db.SQL.Close()

	// Initialize the config package with the logger
	config.InitConfig(log)

	// Init factory for api
	factory, err := factory.NewFactory(db.SQL, log)
	if err != nil {
		log.Error("Error initializing factory", err)
		return
	}

	// Create the application instance
	app := &application{
		config: cfg,
		logger: log,
	}

	// Start the server with domain-specific handlers
	err = app.serve(factory.BookHandlers, factory.AuthHandlers)
	if err != nil {
		log.Error("Error starting the server", "error", err)
	}
}

func (app *application) serve(bookHandlers *handlers.BookHandlers, authHandlers *authHandlers.AuthHandlers) error {
	app.logger.Info("API listening on port", "port", app.config.port)

	// Set up routes with domain handlers
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(bookHandlers, authHandlers),
	}

	return srv.ListenAndServe()
}
