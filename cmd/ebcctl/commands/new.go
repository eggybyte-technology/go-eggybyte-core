package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

var (
	// newCmd is the parent command for code generation
	newCmd = &cobra.Command{
		Use:   "new",
		Short: "Generate new code (repo, service, handler)",
		Long: `Generate new code components for your microservice.

Available subcommands:
  repo    - Generate repository with auto-registration
  service - Generate service layer code

All generated code follows EggyByte standards and includes
comprehensive English documentation.`,
	}

	// newRepoCmd generates repository code
	newRepoCmd = &cobra.Command{
		Use:   "repo <model-name>",
		Short: "Generate repository code with auto-registration",
		Long: `Generate repository code with automatic table registration.

Creates a repository file in internal/repositories/ with:
  - Repository struct and interface
  - TableName() implementation
  - InitTable() with AutoMigrate
  - Auto-registration via init()

Example:
  ebcctl new repo user
  ebcctl new repo order`,
		Args: cobra.ExactArgs(1),
		RunE: runNewRepo,
	}
)

func init() {
	newCmd.AddCommand(newRepoCmd)
}

// runNewRepo generates a new repository file.
func runNewRepo(cmd *cobra.Command, args []string) error {
	modelName := args[0]

	// Validate we're in a service project
	if err := validateInProject(); err != nil {
		return err
	}

	// Generate names
	structName := toStructName(modelName)
	tableName := toTableName(modelName)
	fileName := fmt.Sprintf("%s_repository.go", strings.ToLower(modelName))

	logInfo("Generating repository for model: %s", modelName)
	logDebug("Struct name: %s", structName)
	logDebug("Table name: %s", tableName)
	logDebug("File name: %s", fileName)

	// Generate repository file
	repoPath := filepath.Join("internal", "repositories", fileName)
	content := generateRepositoryCode(modelName, structName, tableName)

	if err := os.WriteFile(repoPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write repository file: %w", err)
	}

	logSuccess("Repository generated: %s", repoPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Define your model struct in the repository file")
	fmt.Println("  2. Import the repository package in main.go:")
	fmt.Printf("     import _ \"<module-path>/internal/repositories\"\n")
	fmt.Println("  3. The table will auto-migrate on service startup!")
	fmt.Println()

	return nil
}

// validateInProject checks if we're inside a service project.
func validateInProject() error {
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("not in a Go project directory (go.mod not found)")
	}

	if _, err := os.Stat("internal"); os.IsNotExist(err) {
		return fmt.Errorf("not in a service project (internal/ directory not found)")
	}

	return nil
}

// toStructName converts model name to struct name (PascalCase).
func toStructName(name string) string {
	// Remove hyphens and underscores, capitalize each word
	words := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_'
	})

	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}

	return strings.Join(words, "")
}

// toTableName converts model name to table name (snake_case, plural).
func toTableName(name string) string {
	// Convert to snake_case
	var result strings.Builder
	for i, r := range name {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}

	tableName := strings.ReplaceAll(result.String(), "-", "_")

	// Simple pluralization (add 's')
	if !strings.HasSuffix(tableName, "s") {
		tableName += "s"
	}

	return tableName
}

// generateRepositoryCode generates the repository file content.
func generateRepositoryCode(modelName, structName, tableName string) string {
	return fmt.Sprintf(`package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/eggybyte-technology/go-eggybyte-core/db"
)

// %s represents the %s data model.
// TODO: Add your model fields here
//
// Example:
//
//	type %s struct {
//	    ID        uint      `+"`gorm:\"primaryKey\"`"+`
//	    CreatedAt time.Time
//	    UpdatedAt time.Time
//	    // Add your fields here
//	}
type %s struct {
	ID uint `+"`gorm:\"primaryKey\"`"+`
	// TODO: Add your model fields
}

// %sRepository handles database operations for %s models.
// It implements the Repository interface for automatic table initialization.
type %sRepository struct {
	db *gorm.DB
}

// %sRepositoryInterface defines the contract for %s repository operations.
type %sRepositoryInterface interface {
	// Create inserts a new %s record into the database
	Create(ctx context.Context, model *%s) error

	// FindByID retrieves a %s by its ID
	FindByID(ctx context.Context, id uint) (*%s, error)

	// Update modifies an existing %s record
	Update(ctx context.Context, model *%s) error

	// Delete removes a %s record by ID
	Delete(ctx context.Context, id uint) error
}

// New%sRepository creates a new instance of %sRepository.
// Note: The database connection is injected during InitTable().
//
// Returns:
//   - *%sRepository: Repository instance ready for registration
func New%sRepository() *%sRepository {
	return &%sRepository{}
}

// TableName returns the database table name for this repository.
// Implements the Repository interface.
//
// Returns:
//   - string: The table name used in the database
func (r *%sRepository) TableName() string {
	return "%s"
}

// InitTable performs table creation and schema migration.
// This method is called automatically during database initialization.
// Implements the Repository interface.
//
// Parameters:
//   - ctx: Context for timeout control and cancellation
//   - db: GORM database connection to use for operations
//
// Returns:
//   - error: Returns error if table creation or migration fails
func (r *%sRepository) InitTable(ctx context.Context, database *gorm.DB) error {
	r.db = database
	return r.db.WithContext(ctx).AutoMigrate(&%s{})
}

// Create inserts a new %s record into the database.
//
// Parameters:
//   - ctx: Context for timeout control and cancellation
//   - model: The %s instance to create
//
// Returns:
//   - error: Returns error if creation fails
func (r *%sRepository) Create(ctx context.Context, model *%s) error {
	return r.db.WithContext(ctx).Create(model).Error
}

// FindByID retrieves a %s by its ID.
//
// Parameters:
//   - ctx: Context for timeout control and cancellation
//   - id: The unique identifier of the %s
//
// Returns:
//   - *%s: The found %s instance
//   - error: Returns error if not found or query fails
func (r *%sRepository) FindByID(ctx context.Context, id uint) (*%s, error) {
	var model %s
	err := r.db.WithContext(ctx).First(&model, id).Error
	return &model, err
}

// Update modifies an existing %s record.
//
// Parameters:
//   - ctx: Context for timeout control and cancellation
//   - model: The %s instance with updated fields
//
// Returns:
//   - error: Returns error if update fails
func (r *%sRepository) Update(ctx context.Context, model *%s) error {
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes a %s record by ID.
//
// Parameters:
//   - ctx: Context for timeout control and cancellation
//   - id: The unique identifier of the %s to delete
//
// Returns:
//   - error: Returns error if deletion fails
func (r *%sRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&%s{}, id).Error
}

// init registers this repository for automatic table initialization.
// This function is called automatically when the package is imported.
func init() {
	db.RegisterRepository(New%sRepository())
}
`,
		// Model type definition
		structName, modelName, structName, structName,

		// Repository struct
		structName, modelName, structName,

		// Interface definition
		structName, modelName, structName,
		modelName, structName,
		modelName, structName,
		modelName, structName,
		modelName,

		// Constructor
		structName, structName, structName, structName, structName, structName,

		// TableName
		structName, tableName,

		// InitTable
		structName, structName,

		// Create
		structName, structName, structName, structName,

		// FindByID
		structName, structName, structName, structName, structName, structName, structName,

		// Update
		structName, structName, structName, structName,

		// Delete
		structName, structName, structName, structName,

		// init registration
		structName,
	)
}
