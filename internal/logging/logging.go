package logging

import (
	"comet/internal/config"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

var Logger *slog.Logger

func getSlogLevelFromString() (slog.Level, error) {
	switch viper.GetString("logging.level") {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 1, errors.New("unknown log type")
	}
}

func init() {
	err := config.ReadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to read config")
		os.Exit(1)
	}
	output := viper.GetString("logging.output")
	fmt.Println(output)
	var logOutput io.Writer

	switch output {
	case "stdout":
		logOutput = os.Stdout
	case "stderr":
		logOutput = os.Stderr
	default:
		file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR: Unable to open log file")
			os.Exit(1)
		}

		logOutput = file
	}

	level, err := getSlogLevelFromString()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	Logger = slog.New(slog.NewTextHandler(logOutput, opts))
}

func LogCritical(message string, args ...any) {
	Logger.Error(message, args...)
	os.Exit(1)
}
