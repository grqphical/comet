package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/viper"
)

type PrettyPrintHandlerOptions struct {
	Level slog.Leveler
}

type PrettyPrintHandler struct {
	mu   *sync.Mutex
	out  io.Writer
	opts PrettyPrintHandlerOptions
}

func NewPrettyPrintHandler(out io.Writer, opts *PrettyPrintHandlerOptions) *PrettyPrintHandler {
	h := &PrettyPrintHandler{
		out: out,
		mu:  &sync.Mutex{},
	}

	if opts != nil {
		h.opts = *opts
	}

	if h.opts.Level == nil {
		h.opts.Level = slog.LevelInfo
	}

	return h
}

func (p *PrettyPrintHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= p.opts.Level.Level()
}

func (p *PrettyPrintHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))

	if !r.Time.IsZero() {
		fmt.Fprintf(buf, "%s ", r.Time.Format(time.RFC822Z))
	}

	var levelPrintFunc func(w io.Writer, format string, a ...interface{})

	switch r.Level {
	case slog.LevelDebug:
		levelPrintFunc = color.New(color.FgCyan).FprintfFunc()
	case slog.LevelWarn:
		levelPrintFunc = color.New(color.FgYellow).FprintfFunc()
	case slog.LevelError:
		levelPrintFunc = color.New(color.FgRed).FprintfFunc()
	default:
		levelPrintFunc = color.New(color.Reset).FprintfFunc()
	}

	levelPrintFunc(buf, "%s", r.Level)
	fmt.Fprint(buf, ": ")

	fmt.Fprintf(buf, "%s ", r.Message)

	attrs := map[string]string{}

	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.String()
		return true
	})

	if len(attrs) != 0 {
		json := json.NewEncoder(buf)

		json.Encode(attrs)
	} else {
		fmt.Fprint(buf, "\n")
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	_, err := p.out.Write(buf.Bytes())
	return err
}

func (p *PrettyPrintHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return p
}

func (p *PrettyPrintHandler) WithGroup(name string) slog.Handler {
	return p
}

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

func InitLogger() {
	output := viper.GetString("logging.output")

	level, err := getSlogLevelFromString()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	var logHandler slog.Handler

	switch output {
	case "stdout":
		logHandler = NewPrettyPrintHandler(os.Stdout, &PrettyPrintHandlerOptions{
			Level: level,
		})
	case "stderr":
		logHandler = NewPrettyPrintHandler(os.Stderr, &PrettyPrintHandlerOptions{
			Level: level,
		})
	default:
		file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR: Unable to open log file")
			os.Exit(1)
		}
		opts := &slog.HandlerOptions{
			Level: level,
		}

		logHandler = slog.NewTextHandler(file, opts)
	}

	Logger = slog.New(logHandler)
}

func LogCritical(message string, args ...any) {
	Logger.Error(message, args...)
	os.Exit(1)
}
