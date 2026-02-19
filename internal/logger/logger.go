package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"ctweb/internal/config"

	"github.com/natefinch/lumberjack"
)

var (
	errorLogger  *slog.Logger
	accessLogger *slog.Logger
	auditLogger  *slog.Logger
	logLevel     slog.Level
	logFiles     map[string]io.WriteCloser
	fileMutex    sync.RWMutex
)

func init() {
	logFiles = make(map[string]io.WriteCloser)
}

// Init initializes slog-based logging with error and access streams.
func Init() error {
	cfg := config.Get()
	logCfg := cfg.Logging

	logLevel = parseLevel(logCfg.Level)

	errorPath := logCfg.File
	if errorPath == "" {
		errorPath = "./logs/ct-system.log"
	}

	accessPath := logCfg.AccessFile
	if accessPath == "" {
		accessPath = defaultAccessPath(errorPath)
	}

	auditPath := logCfg.AuditFile
	if auditPath == "" {
		auditPath = defaultAuditPath(errorPath)
	}

	errorWriter, err := buildWriter(logCfg.Output, errorPath, logCfg.MaxSize, logCfg.MaxBackups, logCfg.MaxAge, logCfg.Compress)
	if err != nil {
		return fmt.Errorf("init error logger: %w", err)
	}

	accessMaxSize := logCfg.AccessMaxSize
	accessMaxBackups := logCfg.AccessMaxBackups
	accessMaxAge := logCfg.AccessMaxAge
	if accessMaxSize <= 0 {
		accessMaxSize = 50
	}
	if accessMaxBackups <= 0 {
		accessMaxBackups = 10
	}
	if accessMaxAge <= 0 {
		accessMaxAge = 7
	}

	accessWriter, err := buildWriter(logCfg.Output, accessPath, accessMaxSize, accessMaxBackups, accessMaxAge, logCfg.Compress)
	if err != nil {
		return fmt.Errorf("init access logger: %w", err)
	}

	auditMaxSize := logCfg.AuditMaxSize
	auditMaxBackups := logCfg.AuditMaxBackups
	auditMaxAge := logCfg.AuditMaxAge
	if auditMaxSize <= 0 {
		auditMaxSize = 100
	}
	if auditMaxBackups <= 0 {
		auditMaxBackups = 5
	}
	if auditMaxAge <= 0 {
		auditMaxAge = 30
	}

	auditWriter, err := buildWriter(logCfg.Output, auditPath, auditMaxSize, auditMaxBackups, auditMaxAge, logCfg.Compress)
	if err != nil {
		return fmt.Errorf("init audit logger: %w", err)
	}

	errorLogger = slog.New(buildHandler(logCfg.Format, errorWriter))
	accessLogger = slog.New(buildHandler(logCfg.Format, accessWriter))
	auditLogger = slog.New(buildHandler(logCfg.Format, auditWriter))
	if errorLogger != nil {
		slog.SetDefault(errorLogger)
	}

	return nil
}

// Get returns the main error logger.
func Get() *slog.Logger {
	if errorLogger == nil {
		return slog.New(buildHandler("json", os.Stdout))
	}
	return errorLogger
}

// Access returns the access logger.
func Access() *slog.Logger {
	if accessLogger == nil {
		return Get()
	}
	return accessLogger
}

// Audit returns the audit logger.
func Audit() *slog.Logger {
	if auditLogger == nil {
		return Get()
	}
	return auditLogger
}

// Debug returns a log event with debug level.
func Debug() *Event { return newEvent(Get(), slog.LevelDebug) }

// Info returns a log event with info level.
func Info() *Event { return newEvent(Get(), slog.LevelInfo) }

// Warn returns a log event with warn level.
func Warn() *Event { return newEvent(Get(), slog.LevelWarn) }

// Error returns a log event with error level.
func Error() *Event { return newEvent(Get(), slog.LevelError) }

// Fatal logs a message and exits with code 1.
func Fatal() *Event { return newEvent(Get(), slog.LevelError).withExit() }

// Panic logs a message and panics.
func Panic() *Event { return newEvent(Get(), slog.LevelError).withPanic() }

// Close closes open log files.
func Close() error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	var lastErr error
	for name, f := range logFiles {
		if err := f.Close(); err != nil {
			lastErr = err
		}
		delete(logFiles, name)
	}
	return lastErr
}

type Event struct {
	logger     *slog.Logger
	level      slog.Level
	attrs      []slog.Attr
	exitAfter  bool
	panicAfter bool
}

func newEvent(logger *slog.Logger, level slog.Level) *Event {
	if logger == nil {
		logger = slog.New(buildHandler("json", os.Stdout))
	}
	return &Event{logger: logger, level: level}
}

func (e *Event) withExit() *Event {
	e.exitAfter = true
	return e
}

func (e *Event) withPanic() *Event {
	e.panicAfter = true
	return e
}

func (e *Event) Str(key, value string) *Event {
	e.attrs = append(e.attrs, slog.String(key, value))
	return e
}

func (e *Event) Int(key string, value int) *Event {
	e.attrs = append(e.attrs, slog.Int(key, value))
	return e
}

func (e *Event) Int64(key string, value int64) *Event {
	e.attrs = append(e.attrs, slog.Int64(key, value))
	return e
}

func (e *Event) Uint(key string, value uint) *Event {
	e.attrs = append(e.attrs, slog.Uint64(key, uint64(value)))
	return e
}

func (e *Event) Uint64(key string, value uint64) *Event {
	e.attrs = append(e.attrs, slog.Uint64(key, value))
	return e
}

func (e *Event) Float64(key string, value float64) *Event {
	e.attrs = append(e.attrs, slog.Float64(key, value))
	return e
}

func (e *Event) Bool(key string, value bool) *Event {
	e.attrs = append(e.attrs, slog.Bool(key, value))
	return e
}

func (e *Event) Dur(key string, value time.Duration) *Event {
	e.attrs = append(e.attrs, slog.Duration(key, value))
	return e
}

func (e *Event) Time(key string, value time.Time) *Event {
	e.attrs = append(e.attrs, slog.Time(key, value))
	return e
}

func (e *Event) Interface(key string, value any) *Event {
	e.attrs = append(e.attrs, slog.Any(key, value))
	return e
}

func (e *Event) Err(err error) *Event {
	e.attrs = append(e.attrs, slog.Any("error", err))
	return e
}

func (e *Event) Msg(msg string) {
	if !hasAttrKey(e.attrs, "module") {
		e.attrs = append(e.attrs, slog.String("module", inferModule()))
	}
	e.logger.Log(context.Background(), e.level, msg, attrsToArgs(e.attrs)...)
	if e.exitAfter {
		os.Exit(1)
	}
	if e.panicAfter {
		panic(msg)
	}
}

func attrsToArgs(attrs []slog.Attr) []any {
	args := make([]any, 0, len(attrs)*2)
	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value.Any())
	}
	return args
}

func hasAttrKey(attrs []slog.Attr, key string) bool {
	for _, attr := range attrs {
		if attr.Key == key {
			return true
		}
	}
	return false
}

func buildHandler(format string, writer io.Writer) slog.Handler {
	opts := &slog.HandlerOptions{
		Level:       logLevel,
		ReplaceAttr: replaceTimeAttr,
	}

	switch strings.ToLower(format) {
	case "text":
		return slog.NewTextHandler(writer, opts)
	default:
		return slog.NewJSONHandler(writer, opts)
	}
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func buildWriter(output string, filePath string, maxSize int, maxBackups int, maxAge int, compress bool) (io.Writer, error) {
	var writers []io.Writer

	switch strings.ToLower(output) {
	case "stdout", "both":
		writers = append(writers, os.Stdout)
	}

	if strings.ToLower(output) == "file" || strings.ToLower(output) == "both" {
		if filePath == "" {
			return nil, fmt.Errorf("log file path is empty")
		}
		if err := ensureLogDir(filePath); err != nil {
			return nil, err
		}

		fileWriter := &lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   compress,
			LocalTime:  true,
		}
		fileMutex.Lock()
		logFiles[filePath] = fileWriter
		fileMutex.Unlock()
		writers = append(writers, fileWriter)
	}

	if len(writers) == 0 {
		writers = append(writers, os.Stderr)
	}

	return io.MultiWriter(writers...), nil
}

func ensureLogDir(filePath string) error {
	dir := filepath.Dir(filePath)
	if dir == "." || dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create log dir %s: %w", dir, err)
	}
	testPath := filepath.Join(dir, ".write-test")
	f, err := os.OpenFile(testPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("log dir %s is not writable: %w", dir, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close write test for %s: %w", dir, err)
	}
	_ = os.Remove(testPath)
	return nil
}

func defaultAccessPath(errorPath string) string {
	dir := filepath.Dir(errorPath)
	return filepath.Join(dir, "access.log")
}

func defaultAuditPath(errorPath string) string {
	dir := filepath.Dir(errorPath)
	return filepath.Join(dir, "audit.log")
}

func replaceTimeAttr(_ []string, attr slog.Attr) slog.Attr {
	if attr.Key != slog.TimeKey {
		return attr
	}
	if t, ok := attr.Value.Any().(time.Time); ok {
		attr.Value = slog.StringValue(t.UTC().Format("2006-01-02T15:04:05.000000Z"))
	}
	return attr
}

func inferModule() string {
	for skip := 2; skip <= 10; skip++ {
		_, file, _, ok := runtime.Caller(skip)
		if !ok {
			continue
		}
		normalized := filepath.ToSlash(file)
		if strings.Contains(normalized, "/internal/logger/") {
			continue
		}
		if idx := strings.Index(normalized, "/internal/"); idx >= 0 {
			rest := normalized[idx+len("/internal/"):]
			parts := strings.Split(rest, "/")
			if len(parts) > 0 && parts[0] != "" {
				return parts[0]
			}
		}
		if idx := strings.Index(normalized, "/cmd/"); idx >= 0 {
			rest := normalized[idx+len("/cmd/"):]
			parts := strings.Split(rest, "/")
			if len(parts) > 0 && parts[0] != "" {
				return parts[0]
			}
		}
	}

	return "app"
}
