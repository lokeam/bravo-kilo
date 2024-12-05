package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/lokeam/bravo-kilo/internal/shared/models"
)

type TokenCleanupWorker struct {
	tokenModel models.TokenModel
	logger     *slog.Logger
	ticker     *time.Ticker
	done       chan bool
}

func NewTokenCleanupWorker(
	tokenModel models.TokenModel,
	logger *slog.Logger,
	) *TokenCleanupWorker {
		if tokenModel == nil {
			logger.Error("Token model is nil")
			return nil
		}
		if logger == nil {
			logger.Error("Logger is nil")
			return nil
		}

		return &TokenCleanupWorker{
			tokenModel: tokenModel,
			logger:     logger,
			done:       make(chan bool),
	}
}

func (w *TokenCleanupWorker) Start() {
	w.ticker = time.NewTicker(24 * time.Hour)
	go func() {
			for {
					select {
					case <-w.ticker.C:
							if err := w.tokenModel.DeleteExpiredTokens(context.Background()); err != nil {
									w.logger.Error("Failed to cleanup expired tokens", "error", err)
							}
					case <-w.done:
							w.ticker.Stop()
							return
					}
			}
	}()
}

func (w *TokenCleanupWorker) Stop() {
	w.done <- true
}