package workers

import (
	"context"
	"log/slog"
	"time"

	authservices "github.com/lokeam/bravo-kilo/internal/auth/services"
)

type DeletionWorker struct {
    interval    time.Duration
    authService authservices.AuthService
    logger      *slog.Logger
    stopChan    chan struct{}
}

func NewDeletionWorker(
    interval time.Duration,
    authService authservices.AuthService,
    logger *slog.Logger,
    ) *DeletionWorker {
        if logger == nil {
            panic("logger cannot be nil")
        }
        if authService == nil {
            panic("authService cannot be nil")
        }

    return &DeletionWorker{
        interval:    interval,
        authService: authService,
        logger:      logger.With("component", "deletion_worker"),
        stopChan:    make(chan struct{}),
    }
}

func (w *DeletionWorker) StartDeletionWorker() {
    ticker := time.NewTicker(w.interval)
    go func() {
        for {
            select {
            case <-ticker.C:
                if err := w.authService.ProcessAccountDeletion(context.Background(), 0); err != nil {
                    w.logger.Error("Failed to process account deletion", "error", err)
                }
            case <-w.stopChan:
                ticker.Stop()
                return
            }
        }
    }()
}

func (w *DeletionWorker) StopDeletionWorker() {
    close(w.stopChan)
}

