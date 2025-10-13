package log

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger defines the interface for structured logging operations.
// All logging methods accept a message and optional fields for structured data.
//
// Implementations must be safe for concurrent use across multiple goroutines.
//
// Standard usage pattern:
//   - Debug: Verbose information for development and troubleshooting
//   - Info: General operational messages and milestones
//   - Warn: Warning conditions that don't require immediate action
//   - Error: Error conditions that may require intervention
//   - Fatal: Critical errors that cause service termination
type Logger interface {
	// Debug logs a debug-level message with optional structured fields.
	// Debug messages are typically disabled in production environments.
	Debug(msg string, fields ...Field)

	// Info logs an info-level message with optional structured fields.
	// Info is the standard log level for operational messages.
	Info(msg string, fields ...Field)

	// Warn logs a warning-level message with optional structured fields.
	// Warnings indicate conditions that should be reviewed but don't prevent operation.
	Warn(msg string, fields ...Field)

	// Error logs an error-level message with optional structured fields.
	// Errors indicate problems that require attention and may affect functionality.
	Error(msg string, fields ...Field)

	// Fatal logs a fatal-level message and terminates the program.
	// Should only be used for unrecoverable errors during initialization.
	Fatal(msg string, fields ...Field)

	// With creates a child logger with the given fields pre-attached.
	// Useful for adding request-scoped or operation-scoped context.
	With(fields ...Field) Logger

	// Sync flushes any buffered log entries.
	// Applications should call Sync before exiting to ensure logs are written.
	Sync() error
}

// Field represents a structured logging field with a key-value pair.
// Fields add structured context to log messages without embedding data in strings.
//
// Example usage:
//
//	logger.Info("User logged in",
//	    Field{Key: "user_id", Value: "12345"},
//	    Field{Key: "ip_address", Value: "192.168.1.1"},
//	)
type Field struct {
	Key   string
	Value interface{}
}

// zapLogger is the default Logger implementation using Uber's zap library.
// It provides high-performance structured logging with minimal allocations.
type zapLogger struct {
	logger *zap.Logger
}

// globalLogger holds the default logger instance used throughout the application.
// It is initialized by Init() and accessed via Default().
var globalLogger Logger

// Init initializes the global logger with the specified level and format.
// This function should be called once during application startup.
//
// Parameters:
//   - level: Log level threshold. Valid values: "debug", "info", "warn", "error", "fatal"
//   - format: Log output format. Valid values: "json" (structured), "console" (human-readable)
//
// Returns:
//   - error: Returns error if level or format is invalid
//
// Example:
//
//	if err := log.Init("info", "json"); err != nil {
//	    panic(fmt.Sprintf("Failed to initialize logger: %v", err))
//	}
//	defer log.Default().Sync()
func Init(level, format string) error {
	// Parse log level
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level '%s': %w", level, err)
	}

	// Configure encoder based on format
	var encoder zapcore.Encoder
	switch format {
	case "json":
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		return fmt.Errorf("invalid log format '%s': must be 'json' or 'console'", format)
	}

	// Build zap logger with stdout output
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapLevel)
	zapLog := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	globalLogger = &zapLogger{logger: zapLog}
	return nil
}

// Default returns the global logger instance.
// If Init has not been called, returns a no-op logger that discards all output.
//
// Returns:
//   - Logger: The global logger instance
//
// Example:
//
//	logger := log.Default()
//	logger.Info("Application started")
func Default() Logger {
	if globalLogger == nil {
		// Return no-op logger if not initialized
		return &zapLogger{logger: zap.NewNop()}
	}
	return globalLogger
}

// SetDefault updates the global logger instance.
// Useful for testing or advanced configuration scenarios.
//
// Parameters:
//   - logger: The logger instance to use as the global default
func SetDefault(logger Logger) {
	globalLogger = logger
}

// Debug logs a debug-level message using the global logger.
// Convenience function that delegates to Default().Debug().
func Debug(msg string, fields ...Field) {
	Default().Debug(msg, fields...)
}

// Info logs an info-level message using the global logger.
// Convenience function that delegates to Default().Info().
func Info(msg string, fields ...Field) {
	Default().Info(msg, fields...)
}

// Warn logs a warning-level message using the global logger.
// Convenience function that delegates to Default().Warn().
func Warn(msg string, fields ...Field) {
	Default().Warn(msg, fields...)
}

// Error logs an error-level message using the global logger.
// Convenience function that delegates to Default().Error().
func Error(msg string, fields ...Field) {
	Default().Error(msg, fields...)
}

// Fatal logs a fatal-level message using the global logger and exits.
// Convenience function that delegates to Default().Fatal().
func Fatal(msg string, fields ...Field) {
	Default().Fatal(msg, fields...)
}

// Implementation of Logger interface for zapLogger

func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, l.convertFields(fields)...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, l.convertFields(fields)...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, l.convertFields(fields)...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, l.convertFields(fields)...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, l.convertFields(fields)...)
}

func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{
		logger: l.logger.With(l.convertFields(fields)...),
	}
}

func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

// convertFields converts our generic Field type to zap.Field.
// This internal helper allows us to maintain a simple public API
// while leveraging zap's type-safe field constructors.
func (l *zapLogger) convertFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	return zapFields
}
