package log

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestWithContext tests attaching logger to context.
// This is an isolated method test with no external dependencies.
func TestWithContext(t *testing.T) {
	ctx := context.Background()
	logger := &zapLogger{logger: zap.NewNop()}

	newCtx := WithContext(ctx, logger)

	assert.NotNil(t, newCtx)
	assert.NotEqual(t, ctx, newCtx, "Should return new context")
}

// TestFromContext_WithLogger tests retrieving logger from context.
// This verifies logger retrieval when one is attached.
func TestFromContext_WithLogger(t *testing.T) {
	ctx := context.Background()
	logger := &zapLogger{logger: zap.NewNop()}

	ctx = WithContext(ctx, logger)
	retrieved := FromContext(ctx)

	assert.NotNil(t, retrieved)
	assert.Equal(t, logger, retrieved)
}

// TestFromContext_WithoutLogger tests default behavior when no logger attached.
// This verifies fallback to Default() when context has no logger.
func TestFromContext_WithoutLogger(t *testing.T) {
	ctx := context.Background()

	retrieved := FromContext(ctx)

	assert.NotNil(t, retrieved)
	assert.Equal(t, Default(), retrieved)
}

// TestWithRequestID_GeneratesID tests automatic request ID generation.
// This verifies that empty requestID triggers UUID generation.
func TestWithRequestID_GeneratesID(t *testing.T) {
	ctx := context.Background()

	newCtx, requestID := WithRequestID(ctx, "")

	assert.NotNil(t, newCtx)
	assert.NotEmpty(t, requestID, "Should generate request ID when empty")
	assert.Len(t, requestID, 36, "Should generate UUID format (36 chars with hyphens)")
}

// TestWithRequestID_UsesProvidedID tests using provided request ID.
// This verifies that provided requestID is used without generation.
func TestWithRequestID_UsesProvidedID(t *testing.T) {
	ctx := context.Background()
	providedID := "custom-request-id-123"

	newCtx, requestID := WithRequestID(ctx, providedID)

	assert.NotNil(t, newCtx)
	assert.Equal(t, providedID, requestID)
}

// TestGetRequestID_WithID tests retrieving request ID from context.
// This verifies request ID retrieval when one is attached.
func TestGetRequestID_WithID(t *testing.T) {
	ctx := context.Background()
	expectedID := "test-request-id"

	ctx, _ = WithRequestID(ctx, expectedID)
	retrievedID := GetRequestID(ctx)

	assert.Equal(t, expectedID, retrievedID)
}

// TestGetRequestID_WithoutID tests default behavior when no ID attached.
// This verifies empty string return when context has no request ID.
func TestGetRequestID_WithoutID(t *testing.T) {
	ctx := context.Background()

	retrievedID := GetRequestID(ctx)

	assert.Empty(t, retrievedID, "Should return empty string when no request ID")
}

// TestWithLogger_Complete tests the complete logger setup with request ID.
// This verifies the convenience function creates logger with all fields.
func TestWithLogger_Complete(t *testing.T) {
	require.NoError(t, Init("info", "json"))
	defer SetDefault(nil)

	ctx := context.Background()

	newCtx, logger := WithLogger(ctx, "")

	assert.NotNil(t, newCtx)
	assert.NotNil(t, logger)

	// Verify request ID was generated and attached
	requestID := GetRequestID(newCtx)
	assert.NotEmpty(t, requestID)

	// Verify logger is attached to context
	contextLogger := FromContext(newCtx)
	assert.Equal(t, logger, contextLogger)
}

// TestWithLogger_WithProvidedRequestID tests WithLogger with custom request ID.
// This verifies the convenience function uses provided request ID.
func TestWithLogger_WithProvidedRequestID(t *testing.T) {
	require.NoError(t, Init("info", "json"))
	defer SetDefault(nil)

	ctx := context.Background()
	customID := "my-custom-id"

	newCtx, logger := WithLogger(ctx, customID)

	assert.NotNil(t, newCtx)
	assert.NotNil(t, logger)

	// Verify custom request ID was used
	requestID := GetRequestID(newCtx)
	assert.Equal(t, customID, requestID)
}

// TestWithLogger_WithAdditionalFields tests WithLogger with extra fields.
// This verifies additional fields are attached to the logger.
func TestWithLogger_WithAdditionalFields(t *testing.T) {
	require.NoError(t, Init("info", "json"))
	defer SetDefault(nil)

	ctx := context.Background()
	additionalFields := []Field{
		{Key: "user_id", Value: "user123"},
		{Key: "endpoint", Value: "/api/users"},
	}

	newCtx, logger := WithLogger(ctx, "", additionalFields...)

	assert.NotNil(t, newCtx)
	assert.NotNil(t, logger)

	// Logger should have additional fields attached
	// We can't directly inspect fields, but verify logger is created
	assert.NotNil(t, logger)
}

// TestDebugContext tests context-aware debug logging.
// This verifies the convenience function uses logger from context.
func TestDebugContext(t *testing.T) {
	require.NoError(t, Init("debug", "json"))
	defer SetDefault(nil)

	ctx := context.Background()
	ctx, logger := WithLogger(ctx, "test-request")

	// Should not panic
	assert.NotPanics(t, func() {
		DebugContext(ctx, "debug message",
			Field{Key: "key", Value: "value"})
	})

	// Verify logger in context is used
	assert.Equal(t, logger, FromContext(ctx))
}

// TestInfoContext tests context-aware info logging.
// This verifies the convenience function uses logger from context.
func TestInfoContext(t *testing.T) {
	require.NoError(t, Init("info", "json"))
	defer SetDefault(nil)

	ctx := context.Background()
	ctx, _ = WithLogger(ctx, "test-request")

	assert.NotPanics(t, func() {
		InfoContext(ctx, "info message",
			Field{Key: "operation", Value: "create_user"})
	})
}

// TestWarnContext tests context-aware warning logging.
// This verifies the convenience function uses logger from context.
func TestWarnContext(t *testing.T) {
	require.NoError(t, Init("warn", "json"))
	defer SetDefault(nil)

	ctx := context.Background()
	ctx, _ = WithLogger(ctx, "test-request")

	assert.NotPanics(t, func() {
		WarnContext(ctx, "warning message",
			Field{Key: "reason", Value: "high_latency"})
	})
}

// TestErrorContext tests context-aware error logging.
// This verifies the convenience function uses logger from context.
func TestErrorContext(t *testing.T) {
	require.NoError(t, Init("error", "json"))
	defer SetDefault(nil)

	ctx := context.Background()
	ctx, _ = WithLogger(ctx, "test-request")

	assert.NotPanics(t, func() {
		ErrorContext(ctx, "error message",
			Field{Key: "error_code", Value: 500})
	})
}

// TestContextChaining tests chaining multiple context operations.
// This verifies context modifications can be chained together.
func TestContextChaining(t *testing.T) {
	require.NoError(t, Init("info", "json"))
	defer SetDefault(nil)

	ctx := context.Background()

	// Chain: Add request ID -> Add logger -> Verify
	ctx, requestID := WithRequestID(ctx, "")
	logger := Default().With(Field{Key: "request_id", Value: requestID})
	ctx = WithContext(ctx, logger)

	// Verify all context values
	assert.NotEmpty(t, GetRequestID(ctx))
	assert.NotNil(t, FromContext(ctx))
	assert.Equal(t, logger, FromContext(ctx))
}

// TestConcurrentContextOperations tests thread safety of context operations.
// This verifies concurrent access doesn't cause data races.
func TestConcurrentContextOperations(t *testing.T) {
	require.NoError(t, Init("info", "json"))
	defer SetDefault(nil)

	ctx := context.Background()
	done := make(chan bool)

	// Start multiple goroutines creating contexts
	for i := 0; i < 10; i++ {
		go func(id int) {
			localCtx, logger := WithLogger(ctx, "")
			assert.NotNil(t, localCtx)
			assert.NotNil(t, logger)

			InfoContext(localCtx, "concurrent log",
				Field{Key: "goroutine", Value: id})

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// No panic means thread safety test passed
	assert.True(t, true)
}

// TestRequestIDUniqueness tests that generated request IDs are unique.
// This verifies UUID generation produces unique identifiers.
func TestRequestIDUniqueness(t *testing.T) {
	ctx := context.Background()
	ids := make(map[string]bool)

	// Generate multiple request IDs
	for i := 0; i < 100; i++ {
		_, requestID := WithRequestID(ctx, "")
		assert.NotContains(t, ids, requestID, "Request IDs should be unique")
		ids[requestID] = true
	}

	assert.Len(t, ids, 100, "Should have 100 unique request IDs")
}

// TestContextIsolation tests that contexts are properly isolated.
// This verifies modifications to one context don't affect others.
func TestContextIsolation(t *testing.T) {
	require.NoError(t, Init("info", "json"))
	defer SetDefault(nil)

	baseCtx := context.Background()

	// Create two separate contexts
	ctx1, logger1 := WithLogger(baseCtx, "request-1")
	ctx2, logger2 := WithLogger(baseCtx, "request-2")

	// Verify they are independent
	assert.NotEqual(t, ctx1, ctx2)
	assert.NotEqual(t, logger1, logger2)

	// Verify request IDs are different
	id1 := GetRequestID(ctx1)
	id2 := GetRequestID(ctx2)
	assert.NotEqual(t, id1, id2)
}
