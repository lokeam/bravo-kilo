package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	// Initialize root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Ensure logs are flushed on exit
	defer func() {
		if err := recover(); err != nil {
			logger.Log.Error("Panic occurred", "error", err)
		}
		os.Stdout.Sync()
	}()

	// Initialize environment and logging
	if err := initializeEnvironment(); err != nil {
		os.Exit(1)
	}

	log := logger.Log

	// Initialize configuration
	cfg := appConfig{port: 8081}
	config.InitConfig(log)

	// Initialize resources
	db, f, err := initializeResources(ctx, log)
	if err != nil {
			log.Error("Failed to initialize resources", "error", err)
			os.Exit(1)
	}
	defer db.SQL.Close()
	defer redis.Close(log)

	// Create the application instance
	app := &application{
		config: cfg,
		logger: log,
	}

	// Create server with timeouts
	srv := app.serve(f.BookHandlers, f.SearchHandlers, f.AuthHandlers)

	// Start background workers
	f.DeletionWorker.StartDeletionWorker()
	defer f.DeletionWorker.StopDeletionWorker()

	// Set up error and shutdown channels
	serverErrors := make(chan error, 1)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Start server in background
	go func() {
		log.Info("API starting", "port", app.config.port)
		serverErrors <- srv.ListenAndServe()
	}()

	// Block until we receive a shutdown signal or server error
	select {
	case err := <-serverErrors:
		log.Error("Server error", "error", err)
	case sig := <-shutdown:
		log.Info("Shutdown signal received", "signal", sig)

		// Create shutdown context with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
		defer shutdownCancel()

		// Perform graceful shutdown
		if err := gracefulShutdown(shutdownCtx, srv, f, log); err != nil {
			log.Error("Shutdown error", "error", err)
			// Force shutdown if graceful shutdown fails
			srv.Close()
		}
	}
}

func initializeEnvironment() error {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		handler := slog.NewJSONHandler(os.Stdout, nil)
		slog.New(handler).Error("Error loading .env file", "error", err)
	}

	// Initialize loggers
	logger.Init()
	jwt.InitLogger(logger.Log)

	// Check if upload directory is set
	if uploadDir := os.Getenv("UPLOAD_DIR"); uploadDir == "" {
		logger.Log.Error("Upload dir is not set")
		return fmt.Errorf("upload directory not set")
	}

	return nil
}

func (app *application) serve(
	bookHandlers *handlers.BookHandlers,
	searchHandlers *handlers.SearchHandlers,
	authHandlers *authHandlers.AuthHandlers,
) *http.Server {
	app.logger.Info("Initializing server", "port", app.config.port)

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(bookHandlers, searchHandlers, authHandlers),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}

func initializeResources(ctx context.Context, log *slog.Logger) (*driver.DB, *factory.Factory, error) {
	// Connect to database
	db, err := driver.ConnectPostgres(ctx, os.Getenv("DSN"), log)
	if err != nil {
			return nil, nil, fmt.Errorf("database connection error: %w", err)
	}

	// Initialize Redis
	redisClient, err := redis.InitRedis(ctx, log)
	if err != nil {
			return nil, nil, fmt.Errorf("redis initialization error: %w", err)
	}

	// Initialize factory
	f, err := factory.NewFactory(db.SQL, redisClient, log)
	if err != nil {
			return nil, nil, fmt.Errorf("factory initialization error: %w", err)
	}

	// Initialize prepared statements for book cache
	if err := f.BookHandlers.BookCache.InitPreparedStatements(); err != nil {
			return nil, nil, fmt.Errorf("error initializing prepared statements: %w", err)
	}

	return db, f, nil
}

func gracefulShutdown(ctx context.Context, srv *http.Server, f *factory.Factory, log *slog.Logger) error {
	// Cleanup order:
	// 1. Server (stop accepting new requests)
	// 2. Application resources (caches, workers)
	// 3. Database connections (handled by defer in main)

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	// Cleanup prepared statements
	if err := f.BookHandlers.BookCache.CleanupPreparedStatements(); err != nil {
		log.Error("Error cleaning prepared statements", "error", err)
	}

	// Stop deletion worker (already handled by defer in main)
	f.DeletionWorker.StopDeletionWorker()

	log.Info("Graceful shutdown completed")
	return nil
}