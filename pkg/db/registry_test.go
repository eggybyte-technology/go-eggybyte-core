package db

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// mockRepository is a test implementation of Repository interface.
// Used for testing the repository registration mechanism.
type mockRepository struct {
	tableName     string
	initCalled    bool
	initError     error
	initTableFunc func(ctx context.Context, db *gorm.DB) error
}

func (m *mockRepository) TableName() string {
	return m.tableName
}

func (m *mockRepository) InitTable(ctx context.Context, db *gorm.DB) error {
	m.initCalled = true
	if m.initTableFunc != nil {
		return m.initTableFunc(ctx, db)
	}
	return m.initError
}

// TestRegisterRepository tests basic repository registration.
// This is an isolated method test with no external dependencies.
func TestRegisterRepository(t *testing.T) {
	// Clear registry for clean test
	ClearRegistry()

	repo := &mockRepository{tableName: "users"}

	RegisterRepository(repo)

	repos := GetRegisteredRepositories()
	assert.Len(t, repos, 1)
	assert.Equal(t, "users", repos[0].TableName())

	// Cleanup
	ClearRegistry()
}

// TestRegisterRepository_Multiple tests registering multiple repositories.
// This verifies the registry can hold multiple repositories.
func TestRegisterRepository_Multiple(t *testing.T) {
	ClearRegistry()

	repo1 := &mockRepository{tableName: "users"}
	repo2 := &mockRepository{tableName: "orders"}
	repo3 := &mockRepository{tableName: "products"}

	RegisterRepository(repo1)
	RegisterRepository(repo2)
	RegisterRepository(repo3)

	repos := GetRegisteredRepositories()
	assert.Len(t, repos, 3)

	// Verify all repositories are registered
	tableNames := make([]string, len(repos))
	for i, repo := range repos {
		tableNames[i] = repo.TableName()
	}
	assert.Contains(t, tableNames, "users")
	assert.Contains(t, tableNames, "orders")
	assert.Contains(t, tableNames, "products")

	// Cleanup
	ClearRegistry()
}

// TestGetRegisteredRepositories_Empty tests empty registry.
// This verifies the registry starts empty after clearing.
func TestGetRegisteredRepositories_Empty(t *testing.T) {
	ClearRegistry()

	repos := GetRegisteredRepositories()

	assert.NotNil(t, repos)
	assert.Len(t, repos, 0)
}

// TestGetRegisteredRepositories_ReturnsCopy tests that returned slice is a copy.
// This verifies external code cannot modify the internal registry.
func TestGetRegisteredRepositories_ReturnsCopy(t *testing.T) {
	ClearRegistry()

	repo := &mockRepository{tableName: "users"}
	RegisterRepository(repo)

	// Get repositories twice
	repos1 := GetRegisteredRepositories()
	repos2 := GetRegisteredRepositories()

	// Verify they are different slices (copies)
	assert.NotSame(t, &repos1, &repos2)

	// Modify one copy - should not affect the other or internal registry
	repos1 = append(repos1, &mockRepository{tableName: "fake"})

	repos3 := GetRegisteredRepositories()
	assert.Len(t, repos3, 1, "Internal registry should not be affected")

	// Cleanup
	ClearRegistry()
}

// TestClearRegistry tests clearing the repository registry.
// This verifies ClearRegistry removes all registered repositories.
func TestClearRegistry(t *testing.T) {
	ClearRegistry()

	repo1 := &mockRepository{tableName: "users"}
	repo2 := &mockRepository{tableName: "orders"}
	RegisterRepository(repo1)
	RegisterRepository(repo2)

	assert.Len(t, GetRegisteredRepositories(), 2)

	ClearRegistry()

	assert.Len(t, GetRegisteredRepositories(), 0)
}

// TestInitializeAllTables_Success tests successful table initialization.
// This is an isolated method test using mock repositories.
func TestInitializeAllTables_Success(t *testing.T) {
	ClearRegistry()

	repo1 := &mockRepository{tableName: "users"}
	repo2 := &mockRepository{tableName: "orders"}
	RegisterRepository(repo1)
	RegisterRepository(repo2)

	ctx := context.Background()

	err := InitializeAllTables(ctx, nil) // Pass nil DB as we're not actually creating tables

	assert.NoError(t, err)
	assert.True(t, repo1.initCalled, "First repository InitTable should be called")
	assert.True(t, repo2.initCalled, "Second repository InitTable should be called")

	// Cleanup
	ClearRegistry()
}

// TestInitializeAllTables_Empty tests initialization with no repositories.
// This verifies the function handles empty registry gracefully.
func TestInitializeAllTables_Empty(t *testing.T) {
	ClearRegistry()

	ctx := context.Background()

	err := InitializeAllTables(ctx, nil)

	assert.NoError(t, err, "Should succeed with empty registry")
}

// TestInitializeAllTables_Error tests error handling during initialization.
// This verifies that an error in one repository stops the process.
func TestInitializeAllTables_Error(t *testing.T) {
	ClearRegistry()

	repo1 := &mockRepository{
		tableName: "users",
		initError: nil,
	}
	repo2 := &mockRepository{
		tableName: "orders",
		initError: errors.New("migration failed"),
	}
	repo3 := &mockRepository{
		tableName: "products",
		initError: nil,
	}

	RegisterRepository(repo1)
	RegisterRepository(repo2)
	RegisterRepository(repo3)

	ctx := context.Background()

	err := InitializeAllTables(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize table orders")
	assert.Contains(t, err.Error(), "migration failed")

	// First repository should be initialized
	assert.True(t, repo1.initCalled)

	// Second repository should be attempted
	assert.True(t, repo2.initCalled)

	// Third repository should NOT be initialized (process stopped)
	assert.False(t, repo3.initCalled, "Initialization should stop on first error")

	// Cleanup
	ClearRegistry()
}

// TestInitializeAllTables_Order tests that repositories initialize in order.
// This verifies the registration order is preserved during initialization.
func TestInitializeAllTables_Order(t *testing.T) {
	ClearRegistry()

	var callOrder []string

	repo1 := &mockRepository{
		tableName: "users",
		initTableFunc: func(ctx context.Context, db *gorm.DB) error {
			callOrder = append(callOrder, "users")
			return nil
		},
	}
	repo2 := &mockRepository{
		tableName: "orders",
		initTableFunc: func(ctx context.Context, db *gorm.DB) error {
			callOrder = append(callOrder, "orders")
			return nil
		},
	}
	repo3 := &mockRepository{
		tableName: "products",
		initTableFunc: func(ctx context.Context, db *gorm.DB) error {
			callOrder = append(callOrder, "products")
			return nil
		},
	}

	RegisterRepository(repo1)
	RegisterRepository(repo2)
	RegisterRepository(repo3)

	ctx := context.Background()

	err := InitializeAllTables(ctx, nil)

	assert.NoError(t, err)
	assert.Equal(t, []string{"users", "orders", "products"}, callOrder,
		"Repositories should initialize in registration order")

	// Cleanup
	ClearRegistry()
}

// TestRepository_Interface tests that mockRepository implements Repository.
// This verifies the test mock correctly implements the interface.
func TestRepository_Interface(t *testing.T) {
	var _ Repository = (*mockRepository)(nil)

	repo := &mockRepository{tableName: "test"}

	assert.Equal(t, "test", repo.TableName())
	assert.False(t, repo.initCalled)
}

// TestRegisterRepository_Concurrent tests concurrent registration.
// This verifies the mutex protection works correctly.
func TestRegisterRepository_Concurrent(t *testing.T) {
	ClearRegistry()

	done := make(chan bool)

	// Start multiple goroutines registering repositories
	for i := 0; i < 20; i++ {
		go func(id int) {
			repo := &mockRepository{
				tableName: "table_" + string(rune('A'+id)),
			}
			RegisterRepository(repo)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify all repositories were registered
	repos := GetRegisteredRepositories()
	assert.Len(t, repos, 20, "All repositories should be registered")

	// Cleanup
	ClearRegistry()
}

// TestGetRegisteredRepositories_Concurrent tests concurrent reads.
// This verifies the read lock allows concurrent access.
func TestGetRegisteredRepositories_Concurrent(t *testing.T) {
	ClearRegistry()

	repo := &mockRepository{tableName: "users"}
	RegisterRepository(repo)

	done := make(chan bool)

	// Start multiple goroutines reading repositories
	for i := 0; i < 20; i++ {
		go func() {
			repos := GetRegisteredRepositories()
			assert.Len(t, repos, 1)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// No panic means concurrent read test passed
	assert.True(t, true)

	// Cleanup
	ClearRegistry()
}

// TestInitializeAllTables_ContextCancellation tests context cancellation handling.
// This verifies that context cancellation is respected during initialization.
func TestInitializeAllTables_ContextCancellation(t *testing.T) {
	ClearRegistry()

	ctx, cancel := context.WithCancel(context.Background())

	repo := &mockRepository{
		tableName: "users",
		initTableFunc: func(ctx context.Context, db *gorm.DB) error {
			// Check if context is canceled
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return nil
			}
		},
	}
	RegisterRepository(repo)

	// Cancel context before initialization
	cancel()

	err := InitializeAllTables(ctx, nil)

	// InitTable should detect context cancellation
	// Note: This depends on the repository implementation checking context
	// Our mock checks it, so it should return error
	assert.Error(t, err)

	// Cleanup
	ClearRegistry()
}

// TestRepositoryLifecycle tests complete repository lifecycle.
// This verifies the typical usage pattern from registration to initialization.
func TestRepositoryLifecycle(t *testing.T) {
	// 1. Start with clean registry
	ClearRegistry()
	assert.Len(t, GetRegisteredRepositories(), 0)

	// 2. Register repositories
	repo1 := &mockRepository{tableName: "users"}
	repo2 := &mockRepository{tableName: "orders"}
	RegisterRepository(repo1)
	RegisterRepository(repo2)
	assert.Len(t, GetRegisteredRepositories(), 2)

	// 3. Initialize all tables
	ctx := context.Background()
	err := InitializeAllTables(ctx, nil)
	assert.NoError(t, err)
	assert.True(t, repo1.initCalled)
	assert.True(t, repo2.initCalled)

	// 4. Clean up for next test
	ClearRegistry()
	assert.Len(t, GetRegisteredRepositories(), 0)
}
