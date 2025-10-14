// Package log provides structured logging for EggyByte services.
package log

import (
	"context"

	"github.com/google/uuid"
)

// Context keys for storing logger and request ID in context.Context.
// These are unexported to prevent collisions with other packages.
type contextKey int

const (
	loggerKey contextKey = iota
	requestIDKey
)

// WithContext attaches a logger instance to a context.
// This enables request-scoped logging where each request has its own logger
// with pre-attached contextual fields like request ID.
//
// Parameters:
//   - ctx: The parent context
//   - logger: The logger instance to attach
//
// Returns:
//   - context.Context: New context with logger attached
//
// Example:
//
//	logger := log.Default().With(log.Field{Key: "request_id", Value: "12345"})
//	ctx = log.WithContext(ctx, logger)
//	// Later: logger := log.FromContext(ctx)
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves the logger instance from a context.
// If no logger is attached to the context, returns the global default logger.
//
// Parameters:
//   - ctx: The context to extract logger from
//
// Returns:
//   - Logger: The logger from context, or global default if none attached
//
// Thread Safety: Safe for concurrent use.
//
// Example:
//
//	logger := log.FromContext(ctx)
//	logger.Info("Processing request")
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	return Default()
}

// WithRequestID attaches a request ID to a context.
// Request IDs are used for distributed tracing and log correlation across services.
// If no request ID is provided, a new UUID is generated.
//
// Parameters:
//   - ctx: The parent context
//   - requestID: Optional request ID. If empty, generates a new UUID
//
// Returns:
//   - context.Context: New context with request ID attached
//   - string: The request ID that was attached (either provided or generated)
//
// Example:
//
//	ctx, requestID := log.WithRequestID(ctx, "")
//	logger := log.FromContext(ctx).With(log.Field{Key: "request_id", Value: requestID})
//	ctx = log.WithContext(ctx, logger)
func WithRequestID(ctx context.Context, requestID string) (newCtx context.Context, finalRequestID string) {
	if requestID == "" {
		requestID = uuid.New().String()
	}
	return context.WithValue(ctx, requestIDKey, requestID), requestID
}

// GetRequestID retrieves the request ID from a context.
// If no request ID is attached, returns an empty string.
//
// Parameters:
//   - ctx: The context to extract request ID from
//
// Returns:
//   - string: The request ID, or empty string if not found
//
// Example:
//
//	requestID := log.GetRequestID(ctx)
//	if requestID != "" {
//	    logger.Info("Processing request", log.Field{Key: "request_id", Value: requestID})
//	}
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// WithLogger creates a new logger with request ID and other contextual fields.
// This is a convenience function that combines WithRequestID, logger creation
// with request ID field, and WithContext.
//
// Parameters:
//   - ctx: The parent context
//   - requestID: Optional request ID. If empty, generates a new UUID
//   - additionalFields: Optional fields to attach to the logger
//
// Returns:
//   - context.Context: New context with logger and request ID attached
//   - Logger: The created logger with contextual fields
//
// Example:
//
//	ctx, logger := log.WithLogger(ctx, "",
//	    log.Field{Key: "user_id", Value: "user123"},
//	    log.Field{Key: "endpoint", Value: "/api/users"},
//	)
//	logger.Info("Request started")
func WithLogger(ctx context.Context, requestID string, additionalFields ...Field) (context.Context, Logger) {
	ctx, reqID := WithRequestID(ctx, requestID)

	// Build fields list with request ID
	fields := make([]Field, 0, len(additionalFields)+1)
	fields = append(fields, Field{Key: "request_id", Value: reqID})
	fields = append(fields, additionalFields...)

	// Create logger with fields
	logger := Default().With(fields...)

	// Attach logger to context
	ctx = WithContext(ctx, logger)

	return ctx, logger
}

// DebugContext logs a debug message using the logger from context.
// Convenience function for context-aware logging.
func DebugContext(ctx context.Context, msg string, fields ...Field) {
	FromContext(ctx).Debug(msg, fields...)
}

// InfoContext logs an info message using the logger from context.
// Convenience function for context-aware logging.
func InfoContext(ctx context.Context, msg string, fields ...Field) {
	FromContext(ctx).Info(msg, fields...)
}

// WarnContext logs a warning message using the logger from context.
// Convenience function for context-aware logging.
func WarnContext(ctx context.Context, msg string, fields ...Field) {
	FromContext(ctx).Warn(msg, fields...)
}

// ErrorContext logs an error message using the logger from context.
// Convenience function for context-aware logging.
func ErrorContext(ctx context.Context, msg string, fields ...Field) {
	FromContext(ctx).Error(msg, fields...)
}

// FatalContext logs a fatal message using the logger from context and exits.
// Convenience function for context-aware logging.
func FatalContext(ctx context.Context, msg string, fields ...Field) {
	FromContext(ctx).Fatal(msg, fields...)
}
