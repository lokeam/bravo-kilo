package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func Init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})
	Log = slog.New(handler)

	// Ensure stdout is unbuffered
	os.Stdout.Sync()
}
