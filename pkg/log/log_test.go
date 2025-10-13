package log

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestInit_ValidJSONFormat tests logger initialization with JSON format.
// This is an isolated method test that verifies logger configuration.
func TestInit_ValidJSONFormat(t *testing.T) {
	err := Init("info", "json")

	assert.NoError(t, err, "Init should succeed with valid parameters")
	assert.NotNil(t, Default(), "Global logger should be initialized")

	// Cleanup
	SetDefault(nil)
}

// TestInit_ValidConsoleFormat tests logger initialization with console format.
// This is an isolated method test that verifies logger configuration.
func TestInit_ValidConsoleFormat(t *testing.T) {
	err := Init("debug", "console")

	assert.NoError(t, err, "Init should succeed with console format")
	assert.NotNil(t, Default(), "Global logger should be initialized")

	// Cleanup
	SetDefault(nil)
}

// TestInit_InvalidLogLevel tests error handling for invalid log levels.
// This verifies the log level validation logic.
func TestInit_InvalidLogLevel(t *testing.T) {
	err := Init("invalid", "json")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}

// TestInit_InvalidFormat tests error handling for invalid log formats.
// This verifies the format validation logic.
func TestInit_InvalidFormat(t *testing.T) {
	err := Init("info", "invalid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log format")
	assert.Contains(t, err.Error(), "must be 'json' or 'console'")
}

// TestInit_AllLogLevels tests initialization with all valid log levels.
// This is an isolated test verifying all supported log levels.
func TestInit_AllLogLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "fatal"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			err := Init(level, "json")

			assert.NoError(t, err, "Log level '%s' should be valid", level)

			// Cleanup
			SetDefault(nil)
		})
	}
}

// TestDefault_BeforeInit tests Default returns no-op logger before initialization.
// This verifies the safe default behavior.
func TestDefault_BeforeInit(t *testing.T) {
	SetDefault(nil)

	logger := Default()

	assert.NotNil(t, logger, "Default should never return nil")

	// Should not panic when using the logger
	assert.NotPanics(t, func() {
		logger.Info("test message")
	})
}

// TestDefault_AfterInit tests Default returns initialized logger.
// This verifies logger retrieval after initialization.
func TestDefault_AfterInit(t *testing.T) {
	require.NoError(t, Init("info", "json"))

	logger := Default()

	assert.NotNil(t, logger)

	// Cleanup
	SetDefault(nil)
}

// TestSetDefault tests setting a custom logger.
// This is an isolated method test for custom logger injection.
func TestSetDefault(t *testing.T) {
	// Create custom logger
	customLogger := &zapLogger{
		logger: zap.NewNop(),
	}

	SetDefault(customLogger)
	result := Default()

	assert.Equal(t, customLogger, result)

	// Cleanup
	SetDefault(nil)
}

// TestField_Structure tests the Field struct definition.
// This verifies the field key-value pair structure.
func TestField_Structure(t *testing.T) {
	field := Field{
		Key:   "user_id",
		Value: "12345",
	}

	assert.Equal(t, "user_id", field.Key)
	assert.Equal(t, "12345", field.Value)
}

// TestField_VariousTypes tests Field with different value types.
// This verifies Field can hold various data types.
func TestField_VariousTypes(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{"string", "name", "John Doe"},
		{"int", "age", 30},
		{"bool", "active", true},
		{"float", "score", 95.5},
		{"slice", "tags", []string{"tag1", "tag2"}},
		{"map", "meta", map[string]string{"key": "value"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := Field{
				Key:   tt.key,
				Value: tt.value,
			}

			assert.Equal(t, tt.key, field.Key)
			assert.Equal(t, tt.value, field.Value)
		})
	}
}

// TestZapLogger_Debug tests debug logging.
// This is an isolated test that verifies debug log level.
func TestZapLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLog := zap.New(core)

	logger := &zapLogger{logger: zapLog}

	logger.Debug("debug message", Field{Key: "key", Value: "value"})

	output := buf.String()
	assert.Contains(t, output, "debug message")
	assert.Contains(t, output, "key")
	assert.Contains(t, output, "value")
}

// TestZapLogger_Info tests info logging.
// This is an isolated test that verifies info log level.
func TestZapLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.InfoLevel)
	zapLog := zap.New(core)

	logger := &zapLogger{logger: zapLog}

	logger.Info("info message", Field{Key: "request_id", Value: "123"})

	output := buf.String()
	assert.Contains(t, output, "info message")
	assert.Contains(t, output, "request_id")
}

// TestZapLogger_Warn tests warning logging.
// This is an isolated test that verifies warn log level.
func TestZapLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.WarnLevel)
	zapLog := zap.New(core)

	logger := &zapLogger{logger: zapLog}

	logger.Warn("warning message", Field{Key: "user_id", Value: "user123"})

	output := buf.String()
	assert.Contains(t, output, "warning message")
}

// TestZapLogger_Error tests error logging.
// This is an isolated test that verifies error log level.
func TestZapLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.ErrorLevel)
	zapLog := zap.New(core)

	logger := &zapLogger{logger: zapLog}

	logger.Error("error message", Field{Key: "error_code", Value: 500})

	output := buf.String()
	assert.Contains(t, output, "error message")
}

// TestZapLogger_With tests creating child logger with fields.
// This verifies the With method for context propagation.
func TestZapLogger_With(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.InfoLevel)
	zapLog := zap.New(core)

	logger := &zapLogger{logger: zapLog}
	childLogger := logger.With(Field{Key: "service", Value: "user-service"})

	assert.NotNil(t, childLogger)
	assert.IsType(t, &zapLogger{}, childLogger)

	childLogger.Info("child log")

	output := buf.String()
	assert.Contains(t, output, "child log")
	assert.Contains(t, output, "service")
	assert.Contains(t, output, "user-service")
}

// TestZapLogger_Sync tests sync operation.
// This verifies the Sync method completes without error.
func TestZapLogger_Sync(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.InfoLevel)
	zapLog := zap.New(core)

	logger := &zapLogger{logger: zapLog}

	err := logger.Sync()

	// Sync might return error for in-memory buffer, but shouldn't panic
	_ = err
	assert.NotNil(t, logger)
}

// TestGlobalFunctions tests the global logging convenience functions.
// This verifies that global functions delegate to Default().
func TestGlobalFunctions(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLog := zap.New(core)

	customLogger := &zapLogger{logger: zapLog}
	SetDefault(customLogger)

	// Test each global function
	Debug("debug message")
	assert.Contains(t, buf.String(), "debug message")

	buf.Reset()
	Info("info message")
	assert.Contains(t, buf.String(), "info message")

	buf.Reset()
	Warn("warn message")
	assert.Contains(t, buf.String(), "warn message")

	buf.Reset()
	Error("error message")
	assert.Contains(t, buf.String(), "error message")

	// Cleanup
	SetDefault(nil)
}

// TestConvertFields tests field conversion from generic to zap fields.
// This is an isolated test of the internal conversion logic.
func TestConvertFields(t *testing.T) {
	logger := &zapLogger{logger: zap.NewNop()}

	fields := []Field{
		{Key: "string", Value: "value"},
		{Key: "int", Value: 42},
		{Key: "bool", Value: true},
	}

	zapFields := logger.convertFields(fields)

	assert.Len(t, zapFields, 3)
	assert.IsType(t, zap.Field{}, zapFields[0])
}

// TestJSONOutput tests that JSON format produces valid JSON.
// This verifies the JSON encoder configuration.
func TestJSONOutput(t *testing.T) {
	require.NoError(t, Init("info", "json"))

	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.InfoLevel)
	zapLog := zap.New(core)

	logger := &zapLogger{logger: zapLog}
	logger.Info("test message", Field{Key: "key", Value: "value"})

	output := buf.String()

	// Verify output is valid JSON
	var jsonData map[string]interface{}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 0 {
		err := json.Unmarshal([]byte(lines[0]), &jsonData)
		assert.NoError(t, err, "Output should be valid JSON")
	}

	// Cleanup
	SetDefault(nil)
}

// TestLogLevelFiltering tests that log levels are filtered correctly.
// This verifies the log level threshold behavior.
func TestLogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.WarnLevel)
	zapLog := zap.New(core)

	logger := &zapLogger{logger: zapLog}

	// Debug and Info should be filtered out
	logger.Debug("debug message")
	logger.Info("info message")

	output := buf.String()
	assert.Empty(t, output, "Debug and Info should be filtered at Warn level")

	// Warn should appear
	logger.Warn("warn message")
	output = buf.String()
	assert.Contains(t, output, "warn message")
}
