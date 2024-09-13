package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func Init() {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	Log = slog.New(handler)
}
