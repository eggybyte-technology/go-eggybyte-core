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
  ebcctl init frontend mobile-app --org com.mycompany`,
		Args: cobra.ExactArgs(1),
		RunE: runInitFrontend,
	}

	// organization is the Flutter app organization
	organization string
)

// init registers flags for the init frontend command
func init() {
	initFrontendCmd.Flags().StringVarP(&organization, "org", "o", "com.eggybyte",
		"Organization identifier for the Flutter app")

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

	// Check if Flutter is installed
	if err := checkFlutterInstalled(); err != nil {
		return err
	}

	// Create Flutter project
	if err := createFlutterProject(appName); err != nil {
		return fmt.Errorf("failed to create Flutter project: %w", err)
	}

	// Customize Flutter project
	if err := customizeFlutterProject(appName); err != nil {
		return fmt.Errorf("failed to customize Flutter project: %w", err)
	}

	// Print success message
	printFlutterSuccessMessage(appName)

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

// createFlutterProject creates a new Flutter project using flutter create.
func createFlutterProject(appName string) error {
	logInfo("Creating Flutter project...")

	cmd := exec.Command("flutter", "create",
		"--org", organization,
		"--platforms", "android,ios,web",
		appName)
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
func printFlutterSuccessMessage(appName string) {
	logSuccess("Flutter project '%s' initialized successfully!", appName)
	fmt.Println("\nNext steps:")
	fmt.Printf("  1. cd %s\n", appName)
	fmt.Println("  2. Copy .env.example to .env and configure")
	fmt.Println("  3. Install dependencies: flutter pub get")
	fmt.Println("  4. Run the app: flutter run")
	fmt.Println("\nFor more information, see README.md")
	fmt.Println()
}
