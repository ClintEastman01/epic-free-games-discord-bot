package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
	level slog.Level
}

// LogLevel represents logging levels
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// New creates a new logger instance
func New(level LogLevel, environment string) *Logger {
	var slogLevel slog.Level
	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: slogLevel,
		AddSource: environment == "development",
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize timestamp format
			if a.Key == slog.TimeKey {
				return slog.String("time", a.Value.Time().Format(time.RFC3339))
			}
			// Shorten source file paths
			if a.Key == slog.SourceKey {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					// Get just the filename and line number
					parts := strings.Split(source.File, "/")
					if len(parts) > 0 {
						source.File = parts[len(parts)-1]
					}
				}
			}
			return a
		},
	}

	if environment == "production" {
		// JSON format for production
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Text format for development
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	
	return &Logger{
		Logger: logger,
		level:  slogLevel,
	}
}

// WithContext adds context to the logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		Logger: l.Logger.With(),
		level:  l.level,
	}
}

// WithFields adds structured fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		Logger: l.Logger.With(args...),
		level:  l.level,
	}
}

// WithComponent adds a component field to the logger
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger: l.Logger.With("component", component),
		level:  l.level,
	}
}

// WithError adds an error field to the logger
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}
	return &Logger{
		Logger: l.Logger.With("error", err.Error()),
		level:  l.level,
	}
}

// Discord-specific logging methods

// LogDiscordEvent logs Discord-related events
func (l *Logger) LogDiscordEvent(event string, guildID string, fields map[string]interface{}) {
	logFields := map[string]interface{}{
		"event":    event,
		"guild_id": guildID,
	}
	for k, v := range fields {
		logFields[k] = v
	}
	l.WithFields(logFields).Info("Discord event")
}

// LogDiscordError logs Discord-related errors
func (l *Logger) LogDiscordError(event string, guildID string, err error, fields map[string]interface{}) {
	logFields := map[string]interface{}{
		"event":    event,
		"guild_id": guildID,
		"error":    err.Error(),
	}
	for k, v := range fields {
		logFields[k] = v
	}
	l.WithFields(logFields).Error("Discord error")
}

// LogCommand logs command usage
func (l *Logger) LogCommand(command string, userID string, guildID string, success bool) {
	l.WithFields(map[string]interface{}{
		"command":  command,
		"user_id":  userID,
		"guild_id": guildID,
		"success":  success,
	}).Info("Command executed")
}

// LogScraping logs scraping activities
func (l *Logger) LogScraping(action string, duration time.Duration, gamesFound int, err error) {
	fields := map[string]interface{}{
		"action":       action,
		"duration_ms":  duration.Milliseconds(),
		"games_found":  gamesFound,
	}
	
	if err != nil {
		fields["error"] = err.Error()
		l.WithFields(fields).Error("Scraping failed")
	} else {
		l.WithFields(fields).Info("Scraping completed")
	}
}

// LogHTTPRequest logs HTTP requests
func (l *Logger) LogHTTPRequest(method string, path string, statusCode int, duration time.Duration, userAgent string) {
	l.WithFields(map[string]interface{}{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
		"user_agent":  userAgent,
	}).Info("HTTP request")
}

// LogDatabaseOperation logs database operations
func (l *Logger) LogDatabaseOperation(operation string, table string, duration time.Duration, rowsAffected int64, err error) {
	fields := map[string]interface{}{
		"operation":     operation,
		"table":         table,
		"duration_ms":   duration.Milliseconds(),
		"rows_affected": rowsAffected,
	}
	
	if err != nil {
		fields["error"] = err.Error()
		l.WithFields(fields).Error("Database operation failed")
	} else {
		l.WithFields(fields).Debug("Database operation completed")
	}
}

// Performance monitoring

// LogPerformance logs performance metrics
func (l *Logger) LogPerformance(operation string, duration time.Duration, metadata map[string]interface{}) {
	fields := map[string]interface{}{
		"operation":   operation,
		"duration_ms": duration.Milliseconds(),
	}
	for k, v := range metadata {
		fields[k] = v
	}
	
	// Log as warning if operation takes too long
	if duration > 5*time.Second {
		l.WithFields(fields).Warn("Slow operation detected")
	} else {
		l.WithFields(fields).Debug("Performance metric")
	}
}

// LogMemoryUsage logs current memory usage
func (l *Logger) LogMemoryUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	l.WithFields(map[string]interface{}{
		"alloc_mb":      bToMb(m.Alloc),
		"total_alloc_mb": bToMb(m.TotalAlloc),
		"sys_mb":        bToMb(m.Sys),
		"num_gc":        m.NumGC,
		"goroutines":    runtime.NumGoroutine(),
	}).Debug("Memory usage")
}

// Security logging

// LogSecurityEvent logs security-related events
func (l *Logger) LogSecurityEvent(event string, severity string, details map[string]interface{}) {
	fields := map[string]interface{}{
		"security_event": event,
		"severity":       severity,
	}
	for k, v := range details {
		fields[k] = v
	}
	
	switch severity {
	case "critical", "high":
		l.WithFields(fields).Error("Security event")
	case "medium":
		l.WithFields(fields).Warn("Security event")
	default:
		l.WithFields(fields).Info("Security event")
	}
}

// LogRateLimit logs rate limiting events
func (l *Logger) LogRateLimit(endpoint string, clientIP string, userAgent string) {
	l.WithFields(map[string]interface{}{
		"endpoint":   endpoint,
		"client_ip":  clientIP,
		"user_agent": userAgent,
	}).Warn("Rate limit exceeded")
}

// Utility functions

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// Fatal logs a fatal error and exits
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.Error(fmt.Sprintf(msg, args...))
	os.Exit(1)
}

// Panic logs a panic and panics
func (l *Logger) Panic(msg string, args ...interface{}) {
	message := fmt.Sprintf(msg, args...)
	l.Error(message)
	panic(message)
}