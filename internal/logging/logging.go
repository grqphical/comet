package logging

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func init() {
	Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func LogCritical(message string, args ...any) {
	Logger.Error(message, args...)
	os.Exit(1)
}
