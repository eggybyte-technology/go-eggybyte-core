package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// initFrontendCmd initializes a new Flutter frontend project
	initFrontendCmd = &cobra.Command{
		Use:   "frontend <app-name>",
		Short: "Initialize a new Flutter frontend project",
		Long: `Initialize a new Flutter application with complete structure.

Creates a new Flutter project directory containing:
  - Complete Flutter project structure
  - pubspec.yaml with dependencies
  - Material Design setup
  - HTTP client configuration
  - State management (Provider)
  - Environment configuration
  - README.md with documentation

Example:
  ebcctl init frontend eggybyte-app
  ebcctl init frontend mobile-app --org com.mycompany --platforms android,ios
  ebcctl init frontend web-app --platforms web`,
		Args: cobra.ExactArgs(1),
		RunE: runInitFrontend,
	}

	// organization is the Flutter app organization
	organization string
	// platforms specifies which platforms to include
	platforms string
)

// init registers flags for the init frontend command
func init() {
	initFrontendCmd.Flags().StringVarP(&organization, "org", "o", "com.eggybyte",
		"Organization identifier for the Flutter app")
	initFrontendCmd.Flags().StringVarP(&platforms, "platforms", "p", "android,ios,web",
		"Comma-separated list of platforms to include (android,ios,web,linux,macos,windows)")

	// Add to parent init command
	initCmd.AddCommand(initFrontendCmd)
}

// runInitFrontend executes the init frontend command to create a new Flutter project.
func runInitFrontend(cmd *cobra.Command, args []string) error {
	appName := args[0]

	// Validate app name
	if err := validateFlutterAppName(appName); err != nil {
		return err
	}

	logInfo("Initializing Flutter project: %s", appName)
	logDebug("Organization: %s", organization)
	logDebug("Platforms: %s", platforms)

	// Check if Flutter is installed
	if err := checkFlutterInstalled(); err != nil {
		return err
	}

	// Determine target directory - check if we're in a full-stack project
	targetDir := appName
	if isInFullStackProject() {
		targetDir = filepath.Join("frontend", appName)
		logInfo("Detected full-stack project, creating frontend in: %s", targetDir)
	}

	// Create Flutter project
	if err := createFlutterProjectInDir(appName, targetDir); err != nil {
		return fmt.Errorf("failed to create Flutter project: %w", err)
	}

	// Customize Flutter project
	if err := customizeFlutterProject(targetDir); err != nil {
		return fmt.Errorf("failed to customize Flutter project: %w", err)
	}

	// Print success message
	printFlutterSuccessMessage(targetDir)

	return nil
}

// validateFlutterAppName checks if the Flutter app name is valid.
func validateFlutterAppName(name string) error {
	if name == "" {
		return fmt.Errorf("app name cannot be empty")
	}

	// Check for valid characters (lowercase, numbers, underscores)
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return fmt.Errorf("app name must contain only lowercase letters, numbers, and underscores")
		}
	}

	// Check if directory already exists
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' already exists", name)
	}

	return nil
}

// checkFlutterInstalled checks if Flutter is available.
func checkFlutterInstalled() error {
	logInfo("Checking Flutter installation...")

	cmd := exec.Command("flutter", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logError("Flutter is not installed or not in PATH")
		logInfo("\nTo install Flutter, visit: https://flutter.dev/docs/get-started/install")
		return fmt.Errorf("flutter not found: %w", err)
	}

	logDebug("Flutter version: %s", strings.Split(string(output), "\n")[0])
	return nil
}

// isInFullStackProject checks if we're in a full-stack project directory.
func isInFullStackProject() bool {
	// Check for indicators of a full-stack project
	if _, err := os.Stat("backend"); err == nil {
		return true
	}
	if _, err := os.Stat("frontend"); err == nil {
		return true
	}
	if _, err := os.Stat("Makefile"); err == nil {
		// Check if Makefile contains full-stack project indicators
		content, err := os.ReadFile("Makefile")
		if err == nil && strings.Contains(string(content), "frontend") {
			return true
		}
	}
	return false
}

// createFlutterProjectInDir creates a new Flutter project in the specified directory.
func createFlutterProjectInDir(appName, targetDir string) error {
	logInfo("Creating Flutter project in: %s", targetDir)

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	cmd := exec.Command("flutter", "create",
		"--org", organization,
		"--platforms", platforms,
		"--project-name", appName,
		targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("flutter create failed: %w", err)
	}

	logInfo("Flutter project created successfully")
	return nil
}

// customizeFlutterProject adds EggyByte-specific configurations.
func customizeFlutterProject(appName string) error {
	logInfo("Customizing Flutter project...")

	// Add additional dependencies to pubspec.yaml
	if err := updatePubspec(appName); err != nil {
		return err
	}

	// Create API configuration file
	if err := createAPIConfig(appName); err != nil {
		return err
	}

	// Create environment configuration
	if err := createEnvConfig(appName); err != nil {
		return err
	}

	// Create README
	if err := createFlutterREADME(appName); err != nil {
		return err
	}

	logInfo("Project customization completed")
	return nil
}

// updatePubspec adds common dependencies to pubspec.yaml.
func updatePubspec(appName string) error {
	pubspecPath := filepath.Join(appName, "pubspec.yaml")

	additionalDeps := `
  # HTTP client
  http: ^1.1.0
  
  # State management
  provider: ^6.1.1
  
  # Environment configuration
  flutter_dotenv: ^5.1.0
  
  # JSON serialization
  json_annotation: ^4.8.1

dev_dependencies:
  flutter_test:
    sdk: flutter
  
  # Code generation
  build_runner: ^2.4.6
  json_serializable: ^6.7.1
  
  flutter_lints: ^3.0.1
`

	// Read existing pubspec
	content, err := os.ReadFile(pubspecPath)
	if err != nil {
		return fmt.Errorf("failed to read pubspec.yaml: %w", err)
	}

	// Insert dependencies after the first dependencies: line
	updated := strings.Replace(string(content),
		"dependencies:\n  flutter:\n    sdk: flutter",
		"dependencies:\n  flutter:\n    sdk: flutter"+additionalDeps,
		1)

	// Write back
	if err := os.WriteFile(pubspecPath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write pubspec.yaml: %w", err)
	}

	logDebug("Updated pubspec.yaml with dependencies")
	return nil
}

// createAPIConfig creates API configuration file.
func createAPIConfig(appName string) error {
	configPath := filepath.Join(appName, "lib", "config", "api_config.dart")

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	content := `/// API Configuration for backend communication
class APIConfig {
  // Base URL for API endpoints
  static const String baseURL = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080',
  );

  // API endpoints
  static const String authEndpoint = '/v1/auth';
  static const String userEndpoint = '/v1/user';

  // Timeout configuration
  static const Duration timeout = Duration(seconds: 30);

  // Headers
  static Map<String, String> get defaultHeaders => {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      };
}
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write api_config.dart: %w", err)
	}

	logDebug("Created API configuration file")
	return nil
}

// createEnvConfig creates environment configuration files.
func createEnvConfig(appName string) error {
	// Create .env.example file
	envExamplePath := filepath.Join(appName, ".env.example")
	envContent := `# API Configuration
API_BASE_URL=http://localhost:8080

# App Configuration
APP_NAME=EggyByte App
`

	if err := os.WriteFile(envExamplePath, []byte(envContent), 0644); err != nil {
		return fmt.Errorf("failed to write .env.example: %w", err)
	}

	// Update .gitignore to exclude .env
	gitignorePath := filepath.Join(appName, ".gitignore")
	gitignoreContent := "\n# Environment files\n.env\n"

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(gitignoreContent); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	logDebug("Created environment configuration")
	return nil
}

// createFlutterREADME creates a comprehensive README for the Flutter project.
func createFlutterREADME(appName string) error {
	readmePath := filepath.Join(appName, "README.md")

	content := fmt.Sprintf(`# %s

Flutter application built with EggyByte standards.

## Getting Started

### Prerequisites

- Flutter SDK 3.16.0+
- Dart 3.2.0+
- Android Studio / Xcode (for mobile development)
- Chrome (for web development)

### Installation

1. Install dependencies:
   `+"```bash"+`
   flutter pub get
   `+"```"+`

2. Configure environment:
   `+"```bash"+`
   cp .env.example .env
   # Edit .env with your configuration
   `+"```"+`

3. Run code generation:
   `+"```bash"+`
   flutter pub run build_runner build
   `+"```"+`

### Run

#### Mobile (Android/iOS)
`+"```bash"+`
flutter run
`+"```"+`

#### Web
`+"```bash"+`
flutter run -d chrome
`+"```"+`

### Build

#### Android APK
`+"```bash"+`
flutter build apk --release
`+"```"+`

#### iOS
`+"```bash"+`
flutter build ios --release
`+"```"+`

#### Web
`+"```bash"+`
flutter build web --release
`+"```"+`

## Project Structure

- `+"`lib/`"+` - Application source code
  - `+"`config/`"+` - Configuration files
  - `+"`models/`"+` - Data models
  - `+"`services/`"+` - API services
  - `+"`screens/`"+` - UI screens
  - `+"`widgets/`"+` - Reusable widgets
  - `+"`providers/`"+` - State management
- `+"`assets/`"+` - Images, fonts, and other assets
- `+"`test/`"+` - Unit and widget tests

## Development

### Code Generation

When you modify model files with JSON annotations:

`+"```bash"+`
flutter pub run build_runner build --delete-conflicting-outputs
`+"```"+`

### Testing

Run all tests:

`+"```bash"+`
flutter test
`+"```"+`

Run tests with coverage:

`+"```bash"+`
flutter test --coverage
`+"```"+`

## API Integration

The app communicates with the EggyByte backend through RESTful APIs.

Base URL is configured in `+"`lib/config/api_config.dart`"+` and can be overridden via environment variables.

## License

Copyright © 2025 EggyByte Technology
`, strings.ReplaceAll(strings.Title(strings.ReplaceAll(appName, "_", " ")), " ", ""))

	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	logDebug("Created Flutter README")
	return nil
}

// printFlutterSuccessMessage prints the success message for Flutter project creation.
func printFlutterSuccessMessage(targetDir string) {
	logSuccess("Flutter project initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Printf("  1. cd %s\n", targetDir)
	fmt.Println("  2. Copy .env.example to .env and configure")
	fmt.Println("  3. Install dependencies: flutter pub get")
	fmt.Println("  4. Run the app: flutter run")
	fmt.Println("\nFor more information, see README.md")
	fmt.Println()
}
