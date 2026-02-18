package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

// Level represents logging level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel converts a string to Level.
func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Logger provides structured logging with levels.
type Logger struct {
	mu        sync.Mutex
	level     Level
	logger    *log.Logger
	component string
}

var defaultLogger = &Logger{
	level:  LevelInfo,
	logger: log.New(os.Stderr, "", log.LstdFlags),
}

// SetLevel sets the global log level.
func SetLevel(level Level) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.level = level
}

// SetOutput sets the output destination.
func SetOutput(w io.Writer) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.logger = log.New(w, "", log.LstdFlags)
}

// New creates a component-specific logger.
func New(component string) *Logger {
	return &Logger{
		level:     defaultLogger.level,
		logger:    defaultLogger.logger,
		component: component,
	}
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	l.mu.Lock()
	if level < defaultLogger.level {
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	prefix := fmt.Sprintf("[%s]", level)
	if l.component != "" {
		prefix = fmt.Sprintf("[%s][%s]", level, l.component)
	}
	l.logger.Printf("%s %s", prefix, msg)
}

// Debug logs at debug level.
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs at info level.
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn logs at warn level.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error logs at error level.
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Package-level convenience functions
func Debug(format string, args ...interface{}) { defaultLogger.Debug(format, args...) }
func Info(format string, args ...interface{})  { defaultLogger.Info(format, args...) }
func Warn(format string, args ...interface{})  { defaultLogger.Warn(format, args...) }
func Error(format string, args ...interface{}) { defaultLogger.Error(format, args...) }
