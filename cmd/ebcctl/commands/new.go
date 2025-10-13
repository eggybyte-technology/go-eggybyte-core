package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var (
	// newCmd is the parent command for code generation
	newCmd = &cobra.Command{
		Use:   "new",
		Short: "Generate new code components",
		Long: `Generate various code components:
  - repo: Create a new repository with auto-registration
  - service: Create a new service template
  - handler: Create a new handler template

Use 'ebcctl new <type> <name>' to generate code.`,
	}

	// newRepoCmd generates a new repository
	newRepoCmd = &cobra.Command{
		Use:   "repo <model-name>",
		Short: "Generate a new repository with auto-registration",
		Long: `Generate a new repository file with:
  - Model definition with GORM tags
  - Repository struct with CRUD methods
  - Auto-registration via init() function
  - Complete documentation and examples

The model name will be converted to proper case:
  - user -> User (model), users (table)
  - order-item -> OrderItem (model), order_items (table)

Examples:
  ebcctl new repo user
  ebcctl new repo order-item
  ebcctl new repo product-category`,
		Args: cobra.ExactArgs(1),
		RunE: runNewRepo,
	}
)

// init registers the new subcommands
func init() {
	newCmd.AddCommand(newRepoCmd)
}

// runNewRepo executes the new repo command to generate repository code.
func runNewRepo(cmd *cobra.Command, args []string) error {
	modelName := args[0]

	// Validate model name
	if err := validateModelName(modelName); err != nil {
		return err
	}

	// Check if we're in a Go project
	if !isGoProject() {
		return fmt.Errorf("not in a Go project directory (go.mod not found)")
	}

	// Check if internal/repositories directory exists
	repoDir := "internal/repositories"
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("internal/repositories directory not found. Run 'ebcctl init backend <name>' first")
	}

	// Generate repository file
	if err := generateRepositoryFile(modelName); err != nil {
		return fmt.Errorf("failed to generate repository: %w", err)
	}

	logSuccess("Repository '%s' generated successfully!", modelName)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Implement your business logic in the repository methods")
	fmt.Println("  2. Import the repository package in your main.go:")
	fmt.Printf("     import _ \"%s/internal/repositories\"\n", getModuleName())
	fmt.Println("  3. Use the repository in your services:")
	fmt.Printf("     repo := repositories.New%sRepository(db.GetDB())\n", toTitleCase(modelName))

	return nil
}

// validateModelName checks if the model name is valid.
func validateModelName(name string) error {
	if name == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	// Check for valid characters (lowercase, numbers, hyphens)
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '-') {
			return fmt.Errorf("model name must contain only lowercase letters, numbers, and hyphens")
		}
	}

	return nil
}

// isGoProject checks if we're in a Go project directory.
func isGoProject() bool {
	_, err := os.Stat("go.mod")
	return !os.IsNotExist(err)
}

// getModuleName extracts the module name from go.mod.
func getModuleName() string {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		return "your-module"
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}

	return "your-module"
}

// generateRepositoryFile generates the repository file from template.
func generateRepositoryFile(modelName string) error {
	fileName := fmt.Sprintf("%s_repository.go", modelName)
	filePath := filepath.Join("internal/repositories", fileName)

	// Check if file already exists
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return fmt.Errorf("repository file '%s' already exists", fileName)
	}

	// Generate content from template
	content := generateRepositoryTemplate(modelName)

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write repository file: %w", err)
	}

	logInfo("Generated repository file: %s", fileName)
	return nil
}

// generateRepositoryTemplate generates the repository code template.
func generateRepositoryTemplate(modelName string) string {
	modelStruct := toTitleCase(modelName)
	tableName := toTableName(modelName)
	packageName := "repositories"

	tmpl := `package {{.PackageName}}

import (
	"context"
	"time"

	"gorm.io/gorm"
	"github.com/eggybyte-technology/go-eggybyte-core/pkg/db"
)

// {{.ModelStruct}} represents the {{.ModelName}} entity in the database.
// This struct defines the data model with GORM tags for database mapping.
type {{.ModelStruct}} struct {
	ID        uint           ` + "`gorm:\"primaryKey\"`" + `
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt ` + "`gorm:\"index\"`" + `
	
	// Add your fields here
	// Example:
	// Name  string ` + "`gorm:\"size:255;not null\"`" + `
	// Email string ` + "`gorm:\"size:255;uniqueIndex;not null\"`" + `
}

// {{.ModelStruct}}Repository provides data access methods for {{.ModelStruct}}.
// This repository implements the db.Repository interface for auto-registration.
type {{.ModelStruct}}Repository struct {
	db *gorm.DB
}

// New{{.ModelStruct}}Repository creates a new repository instance.
func New{{.ModelStruct}}Repository(db *gorm.DB) *{{.ModelStruct}}Repository {
	return &{{.ModelStruct}}Repository{db: db}
}

// TableName returns the database table name for this repository.
// This method is required by the db.Repository interface.
func (r *{{.ModelStruct}}Repository) TableName() string {
	return "{{.TableName}}"
}

// InitTable initializes the database table for this repository.
// This method is called automatically during service startup.
func (r *{{.ModelStruct}}Repository) InitTable(ctx context.Context, db *gorm.DB) error {
	r.db = db
	return db.WithContext(ctx).AutoMigrate(&{{.ModelStruct}}{})
}

// Create creates a new {{.ModelName}} record in the database.
func (r *{{.ModelStruct}}Repository) Create(ctx context.Context, {{.ModelName}} *{{.ModelStruct}}) error {
	return r.db.WithContext(ctx).Create({{.ModelName}}).Error
}

// GetByID retrieves a {{.ModelName}} by its ID.
func (r *{{.ModelStruct}}Repository) GetByID(ctx context.Context, id uint) (*{{.ModelStruct}}, error) {
	var {{.ModelName}} {{.ModelStruct}}
	err := r.db.WithContext(ctx).First(&{{.ModelName}}, id).Error
	if err != nil {
		return nil, err
	}
	return &{{.ModelName}}, nil
}

// Update updates an existing {{.ModelName}} record.
func (r *{{.ModelStruct}}Repository) Update(ctx context.Context, {{.ModelName}} *{{.ModelStruct}}) error {
	return r.db.WithContext(ctx).Save({{.ModelName}}).Error
}

// Delete soft deletes a {{.ModelName}} record.
func (r *{{.ModelStruct}}Repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&{{.ModelStruct}}{}, id).Error
}

// List retrieves all {{.ModelName}} records with pagination.
func (r *{{.ModelStruct}}Repository) List(ctx context.Context, offset, limit int) ([]*{{.ModelStruct}}, error) {
	var {{.ModelName}}s []*{{.ModelStruct}}
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&{{.ModelName}}s).Error
	return {{.ModelName}}s, err
}

// Count returns the total number of {{.ModelName}} records.
func (r *{{.ModelStruct}}Repository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&{{.ModelStruct}}{}).Count(&count).Error
	return count, err
}

// Auto-register this repository with the database registry.
// This init() function is called automatically when the package is imported.
func init() {
	db.RegisterRepository(&{{.ModelStruct}}Repository{})
}
`

	t := template.Must(template.New("repository").Parse(tmpl))
	
	var buf strings.Builder
	err := t.Execute(&buf, struct {
		PackageName string
		ModelName   string
		ModelStruct string
		TableName   string
	}{
		PackageName: packageName,
		ModelName:   modelName,
		ModelStruct: modelStruct,
		TableName:   tableName,
	})
	
	if err != nil {
		panic(err)
	}
	
	return buf.String()
}

// toTitleCase converts a kebab-case string to TitleCase.
func toTitleCase(s string) string {
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(string(part[0])) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// toTableName converts a kebab-case string to snake_case table name.
func toTableName(s string) string {
	// Convert hyphens to underscores and add 's' for plural
	return strings.ReplaceAll(s, "-", "_") + "s"
}