package utils

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"desent-api/configs"
)

type Loggers struct {
	HTTP    *slog.Logger
	Error   *slog.Logger
	closers []io.Closer
}

func NewLoggers(cfg configs.LoggingConfig) (*Loggers, error) {
	if err := os.MkdirAll(cfg.LogsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create logs dir: %w", err)
	}

	if err := cleanupOldLogs(cfg.LogsDir, cfg.RetentionDays, []string{cfg.HTTPLogFilePrefix, cfg.ErrorLogFilePrefix}); err != nil {
		return nil, fmt.Errorf("cleanup old logs: %w", err)
	}

	httpWriter, httpClosers, err := buildWriters(cfg, cfg.HTTPLogFilePrefix, os.Stdout)
	if err != nil {
		return nil, err
	}

	errorWriter, errorClosers, err := buildWriters(cfg, cfg.ErrorLogFilePrefix, os.Stderr)
	if err != nil {
		closeAll(httpClosers)
		return nil, err
	}

	loggers := &Loggers{
		HTTP: slog.New(slog.NewJSONHandler(httpWriter, &slog.HandlerOptions{Level: slog.LevelInfo})),
		Error: slog.New(slog.NewJSONHandler(errorWriter, &slog.HandlerOptions{
			Level: slog.LevelError,
		})),
		closers: append(httpClosers, errorClosers...),
	}

	return loggers, nil
}

func (l *Loggers) Close() error {
	if l == nil {
		return nil
	}

	closeAll(l.closers)
	return nil
}

func buildWriters(cfg configs.LoggingConfig, prefix string, console io.Writer) (io.Writer, []io.Closer, error) {
	writers := make([]io.Writer, 0, 2)
	closers := make([]io.Closer, 0, 1)

	if cfg.ConsoleEnabled {
		writers = append(writers, console)
	}

	if cfg.FileEnabled {
		filename := fmt.Sprintf("%s.%s.log", prefix, time.Now().Format("2006-01-02"))
		path := filepath.Join(cfg.LogsDir, filename)

		file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, nil, fmt.Errorf("open log file %s: %w", path, err)
		}

		writers = append(writers, file)
		closers = append(closers, file)
	}

	if len(writers) == 0 {
		writers = append(writers, console)
	}

	return io.MultiWriter(writers...), closers, nil
}

func cleanupOldLogs(logDir string, retentionDays int, prefixes []string) error {
	if retentionDays <= 0 {
		return nil
	}

	entries, err := os.ReadDir(logDir)
	if err != nil {
		return err
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		prefix, datePart, ok := splitLogFilename(entry.Name())
		if !ok || !hasPrefix(prefix, prefixes) {
			continue
		}

		logDate, err := time.Parse("2006-01-02", datePart)
		if err != nil {
			continue
		}

		if logDate.Before(cutoff) {
			if err := os.Remove(filepath.Join(logDir, entry.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}

func splitLogFilename(name string) (prefix, datePart string, ok bool) {
	trimmed := strings.TrimSuffix(name, ".log")
	if trimmed == name {
		return "", "", false
	}

	parts := strings.Split(trimmed, ".")
	if len(parts) < 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

func hasPrefix(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if value == prefix {
			return true
		}
	}

	return false
}

func closeAll(closers []io.Closer) {
	for _, closer := range closers {
		if closer == nil {
			continue
		}

		_ = closer.Close()
	}
}

var ErrNilLogger = errors.New("logger is nil")
