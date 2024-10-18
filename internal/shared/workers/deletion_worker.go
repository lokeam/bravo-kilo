package workers

import (
	"time"

	auth "github.com/lokeam/bravo-kilo/internal/shared/handlers/auth"
)

type DeletionWorker struct {
    interval    time.Duration
    authHandler *auth.AuthHandlers
    stopChan    chan struct{}
}

func NewDeletionWorker(interval time.Duration, authHandler *auth.AuthHandlers) *DeletionWorker {
    return &DeletionWorker{
        interval:    interval,
        authHandler: authHandler,
        stopChan:    make(chan struct{}),
    }
}

func (w *DeletionWorker) StartDeletionWorker() {
    ticker := time.NewTicker(w.interval)
    go func() {
        for {
            select {
            case <-ticker.C:
                w.authHandler.ProcessDeletionQueue()
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
