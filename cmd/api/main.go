package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	factory "github.com/lokeam/bravo-kilo/cmd/factory"
	"github.com/lokeam/bravo-kilo/config"
	authHandlers "github.com/lokeam/bravo-kilo/internal/auth/handlers"
	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	"github.com/lokeam/bravo-kilo/internal/shared/driver"
	"github.com/lokeam/bravo-kilo/internal/shared/jwt"
	"github.com/lokeam/bravo-kilo/internal/shared/logger"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
)

type appConfig struct {
	port int
}

type application struct {
	config appConfig
	logger *slog.Logger
}

func main() {
	// Ensure logs are flushed on exit
	defer func() {
		if err := recover(); err != nil {
			logger.Log.Error("Panic occurred", "error", err)
		}
		os.Stdout.Sync()
	}()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Log.Info("Received interrupt, shutting down...")
		os.Stdout.Sync()
		os.Exit(0)
	}()

	// Load environment variables
	err := godotenv.Load(".env")
	handler := slog.NewJSONHandler(os.Stdout, nil)
	if err != nil {
		slog.New(handler).Error("Error loading .env file", "error", err)
	}

	// Initialize logger
	logger.Init()
	log := logger.Log

	// Init jwt logger
	jwt.InitLogger(log)

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

	// Init Redis
	redisClient, err := redis.InitRedis(log)
	if err != nil {
		log.Error("Failed to initialize Redis", "error", err)
		os.Exit(1)
	}
	defer redis.Close(log)

	// Init factory for api
	factory, err := factory.NewFactory(db.SQL, redisClient, log)
	if err != nil {
		log.Error("Error initializing factory", "error", err)
		return
	}

	// Start deletion worker
	factory.DeletionWorker.StartDeletionWorker()

	// Create the application instance
	app := &application{
		config: cfg,
		logger: log,
	}

	// Start the server with domain-specific handlers
	err = app.serve(factory.BookHandlers, factory.SearchHandlers, factory.AuthHandlers)
	if err != nil {
		log.Error("Error starting the server", "error", err)
	}

	// Stop deletion worker when application shuts down
	defer factory.DeletionWorker.StopDeletionWorker()
}

func (app *application) serve(
	bookHandlers *handlers.BookHandlers,
	searchHandlers *handlers.SearchHandlers,
	authHandlers *authHandlers.AuthHandlers,
) error {
	app.logger.Info("API listening on port", "port", app.config.port)

	// Set up routes with domain handlers
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(bookHandlers, searchHandlers, authHandlers),
	}

	return srv.ListenAndServe()
}
