package utils

import (
	"context"
	"os"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Logger wraps logrus to provide context-aware logging
type Logger struct {
	*logrus.Logger
}

// ContextKey is a key for storing request ID in context
type ContextKey string

const (
	// RequestIDKey is the key for request ID in context
	RequestIDKey ContextKey = "request_id"
)

// NewLogger creates a new logger
func NewLogger(level string) *Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)

	// Set log level
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	return &Logger{Logger: log}
}

// NewRequestContext creates a new context with request ID
func NewRequestContext() context.Context {
	return context.WithValue(context.Background(), RequestIDKey, uuid.New().String())
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return "unknown"
}

// WithContext adds context fields to entry
func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	return l.WithField("request_id", GetRequestID(ctx))
}

// WithRequestID adds request ID to entry
func (l *Logger) WithRequestID(reqID string) *logrus.Entry {
	return l.WithField("request_id", reqID)
}
