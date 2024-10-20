package workers

import (
	"time"

	authhandlers "github.com/lokeam/bravo-kilo/internal/auth/handlers"
)

type DeletionWorker struct {
    interval    time.Duration
    authHandler *authhandlers.AuthHandlers
    stopChan    chan struct{}
}

func NewDeletionWorker(interval time.Duration, authHandler *authhandlers.AuthHandlers) *DeletionWorker {
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
