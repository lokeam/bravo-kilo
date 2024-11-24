package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	factory "github.com/lokeam/bravo-kilo/cmd/factory"
	"github.com/lokeam/bravo-kilo/cmd/middleware"
	"github.com/lokeam/bravo-kilo/config"
	authHandlers "github.com/lokeam/bravo-kilo/internal/auth/handlers"
	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	"github.com/lokeam/bravo-kilo/internal/shared/driver"
	"github.com/lokeam/bravo-kilo/internal/shared/jwt"
	"github.com/lokeam/bravo-kilo/internal/shared/logger"
	"github.com/lokeam/bravo-kilo/internal/shared/pages/library"
	"github.com/lokeam/bravo-kilo/internal/shared/redis"
	"github.com/lokeam/bravo-kilo/internal/shared/validator"
)

var defaultRedisClient *redis.RedisClient

type appConfig struct {
	port int
}

type application struct {
	config appConfig
	logger *slog.Logger
	compressionMonitor *middleware.CompressionMonitor
}

func main() {
	// Initialize root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize environment and logging
	if err := initializeEnvironment(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize environment: %v\n", err)
		os.Exit(1)
	}

	// Ensure logs are flushed on exit
	defer func() {
		if r := recover(); r != nil {
				// Now we can safely use logger.Log
				logger.Log.Error("Panic occurred",
						"error", r,
						"stack", string(debug.Stack()))

				// Ensure logs are flushed
				os.Stdout.Sync()
		}
	}()

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
		compressionMonitor: middleware.NewCompressionMonitor(ctx, log),
	}

	// Create server with timeouts
	srv := app.serve(
		f.BookHandlers,
		f.SearchHandlers,
		f.AuthHandlers,
		f.LibraryPageHandler,
		f.BaseValidator,
	)

	// Start background workers
	f.DeletionWorker.StartDeletionWorker()
	defer f.DeletionWorker.StopDeletionWorker()
	defer f.CacheWorker.Shutdown()

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
		if err := gracefulShutdown(shutdownCtx, srv, f, app, log); err != nil {
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
	libraryPageHandler *library.LibraryPageHandler,
	baseValidator *validator.BaseValidator,
) *http.Server {
	app.logger.Info("Initializing server", "port", app.config.port)

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(bookHandlers, searchHandlers, authHandlers, libraryPageHandler, baseValidator),
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
	cfg := redis.NewRedisConfig()
	if err := cfg.LoadFromEnv(); err != nil {
		return nil, nil, fmt.Errorf("redis config error: %w", err)
	}

	redisClient, err := redis.NewRedisClient(cfg, log)
	if err != nil {
		return nil, nil, fmt.Errorf("redis client error: %w", err)
	}
	defaultRedisClient = redisClient

	// Connect to Redis
	if err := redisClient.Connect(ctx); err != nil {
		return nil, nil, fmt.Errorf("redis connection error: %w", err)
	}
	defaultRedisClient = redisClient

	// Initialize factory
	f, err := factory.NewFactory(ctx, db.SQL, redisClient, log)
	if err != nil {
			return nil, nil, fmt.Errorf("factory initialization error: %w", err)
	}

	// Initialize prepared statements for book cache
	if err := f.BookHandlers.BookCache.InitPreparedStatements(); err != nil {
			return nil, nil, fmt.Errorf("error initializing prepared statements: %w", err)
	}

	return db, f, nil
}

func gracefulShutdown(ctx context.Context, srv *http.Server, f *factory.Factory, app *application, log *slog.Logger) error {
	// Cleanup order:
	// 1. Server (stop accepting new requests)
	// 2. Application resources (caches, workers)
	// 3. Database connections (handled by defer in main)

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Info("Starting graceful shutdown sequence")

	// Stop accepting new requests, shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	// Cleanup compression monitor
	if app.compressionMonitor != nil {
		log.Info("Stopping compression monitor")

		// Create context for compression monitor shutdown
		monitorCtx, monitorCancel := context.WithTimeout(shutdownCtx, 5 * time.Second)
		defer monitorCancel()

		if err := app.compressionMonitor.Shutdown(monitorCtx); err != nil {
			log.Error("Compression monitor shutdown error", "error", err)
		}

		// Log final compression metrics
		stats := app.compressionMonitor.GetStats()
		log.Info("Final compression metrics",
			"totalRequests", stats.RequestCount,
			"failureCount", stats.FailureCount,
			"averageLatency", stats.AverageLatency,
		)
	}

	// Stop workers in reverse order of importance:
	log.Info("Stopping background workers")

	// Stop account deletion worker
	f.DeletionWorker.StopDeletionWorker()

	// Shutdown cache cleanup worker
	f.CacheWorker.Shutdown()

	// Stop book cache cleanup worker
	f.BookHandlers.BookCache.StopCleanupWorker()

	// Cleanup prepared statements
	if err := f.BookHandlers.BookCache.CleanupPreparedStatements(); err != nil {
		log.Error("Error cleaning prepared statements", "error", err)
	}

	// Library page cleanup
	if f.LibraryPageHandler != nil {
		if err := f.LibraryPageHandler.Cleanup(); err != nil {
			log.Error("Error during library page cleanup", "error", err)
		}
	}

	// Shut down Redis
	if defaultRedisClient != nil {
		if err := defaultRedisClient.Close(); err != nil {
			log.Error("Error closing Redis connection", "error", err)
		}
	}

	// Log final metrics
	metrics := f.CacheManager.GetMetrics()
	log.Info("Final cache metrics",
		"totalOps", metrics.TotalOps,
		"l1Failures", metrics.L1Failures,
		"l2Failures", metrics.L2Failures,
	)

	log.Info("Graceful shutdown completed")
	return nil
}