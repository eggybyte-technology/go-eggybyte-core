// Package db provides database integration for EggyByte services.

import (
	"context"
	"fmt"
	"sync"

	"gorm.io/gorm"

	"github.com/eggybyte-technology/go-eggybyte-core/pkg/log"
)

// Repository defines the interface for database repositories that support
// automatic table initialization through the registry pattern.
//
// Repositories implementing this interface can register themselves during
// package initialization using RegisterRepository(), allowing the core
// library to automatically create and migrate their tables.
//
// Usage pattern:
//  1. Implement Repository interface in your repository struct
//  2. Call RegisterRepository(repo) in init() function
//  3. Core library calls InitTable() during database initialization
//
// Example:
//
//	type UserRepository struct {
//	    db *gorm.DB
//	}
//
//	func (r *UserRepository) TableName() string {
//	    return "users"
//	}
//
//	func (r *UserRepository) InitTable(ctx context.Context, db *gorm.DB) error {
//	    return db.WithContext(ctx).AutoMigrate(&User{})
//	}
//
//	func init() {
//	    db.RegisterRepository(&UserRepository{})
//	}
type Repository interface {
	// TableName returns the database table name managed by this repository.
	// Used for logging and debugging during initialization.
	//
	// Returns:
	//   - string: The table name (e.g., "users", "orders", "sessions")
	TableName() string

	// InitTable performs table creation and schema migration.
	// Called automatically during database initialization for all registered repositories.
	//
	// This method should:
	//   - Create the table if it doesn't exist
	//   - Apply schema migrations for existing tables
	//   - Create indexes and constraints
	//   - Avoid destructive operations (dropping columns, etc.)
	//
	// Parameters:
	//   - ctx: Context for timeout control and cancellation
	//   - db: GORM database connection to use for operations
	//
	// Returns:
	//   - error: Returns error if table creation or migration fails
	//
	// Example:
	//   func (r *UserRepository) InitTable(ctx context.Context, db *gorm.DB) error {
	//       return db.WithContext(ctx).AutoMigrate(&User{}, &Profile{})
	//   }
	InitTable(ctx context.Context, db *gorm.DB) error
}

var (
	// registeredRepositories holds all repositories registered via RegisterRepository().
	// Populated during package initialization through init() functions.
	registeredRepositories []Repository

	// registryMutex protects concurrent access to registeredRepositories.
	// Necessary because init() functions may run in parallel.
	registryMutex sync.Mutex
)

// RegisterRepository adds a repository to the global registry for automatic initialization.
// This function should be called from repository package init() functions.
//
// The registered repository will have its InitTable() method called during
// database initialization, ensuring all tables are created and migrated.
//
// Parameters:
//   - repo: Repository implementation to register
//
// Thread Safety: Safe for concurrent calls from multiple init() functions.
//
// Example:
//
//	func init() {
//	    db.RegisterRepository(&UserRepository{})
//	    db.RegisterRepository(&SessionRepository{})
//	}
func RegisterRepository(repo Repository) {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	log.Debug("Registering repository",
		log.Field{Key: "table", Value: repo.TableName()})

	registeredRepositories = append(registeredRepositories, repo)
}

// GetRegisteredRepositories returns a copy of all registered repositories.
// Useful for introspection and testing.
//
// Returns:
//   - []Repository: Slice containing all registered repositories
//
// Thread Safety: Safe for concurrent access.
func GetRegisteredRepositories() []Repository {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	// Return a copy to prevent external modification
	repos := make([]Repository, len(registeredRepositories))
	copy(repos, registeredRepositories)
	return repos
}

// InitializeAllTables iterates through all registered repositories and
// calls their InitTable() method to create and migrate database tables.
//
// Tables are initialized in registration order. If any table initialization
// fails, the process stops immediately and returns the error.
//
// This function is typically called during application startup as part of
// the database initializer.
//
// Parameters:
//   - ctx: Context for timeout control and cancellation
//   - db: GORM database connection to pass to repositories
//
// Returns:
//   - error: First table initialization error encountered, or nil if all succeed
//
// Example:
//
//	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
//	if err != nil {
//	    return err
//	}
//	if err := db.InitializeAllTables(ctx, db); err != nil {
//	    return fmt.Errorf("table initialization failed: %w", err)
//	}
func InitializeAllTables(ctx context.Context, db *gorm.DB) error {
	repos := GetRegisteredRepositories()

	log.Info("Initializing database tables",
		log.Field{Key: "repository_count", Value: len(repos)})

	for i, repo := range repos {
		log.Debug("Initializing table",
			log.Field{Key: "index", Value: i},
			log.Field{Key: "table", Value: repo.TableName()})

		if err := repo.InitTable(ctx, db); err != nil {
			return fmt.Errorf("failed to initialize table %s: %w", repo.TableName(), err)
		}

		log.Info("Table initialized successfully",
			log.Field{Key: "table", Value: repo.TableName()})
	}

	log.Info("All tables initialized successfully")
	return nil
}

// ClearRegistry removes all registered repositories.
// Primarily used for testing to ensure clean state between test runs.
//
// Warning: This function is NOT thread-safe with RegisterRepository calls.
// Only call during test setup when no init() functions are running.
func ClearRegistry() {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	registeredRepositories = nil
}
