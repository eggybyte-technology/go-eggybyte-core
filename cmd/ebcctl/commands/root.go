package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// verbose enables detailed output during command execution
	verbose bool
	// useLocalCore specifies whether to use local core dependency
	useLocalCore bool
	// coreVersion specifies the core version to use when not using local
	coreVersion string

	// rootCmd is the base command for ebcctl CLI
	rootCmd = &cobra.Command{
		Use:   "ebcctl",
		Short: "EggyByte Control - Microservice scaffolding tool",
		Long: `ebcctl (EggyByte Control) is a command-line tool for generating
microservice projects and boilerplate code following EggyByte standards.

It provides commands to:
  - Initialize new microservice projects with complete structure
  - Generate repository code with auto-registration
  - Create service and handler templates
  - Scaffold Docker and Kubernetes configurations

All generated code follows EggyByte quality standards:
  - English comments with comprehensive documentation
  - Methods under 50 lines
  - Auto-registration patterns (repositories via init())
  - Standard three-layer architecture
  - Proper error handling and logging`,
		Version: "1.0.1",
	}
)

// Execute runs the root command and handles errors.
// This is the main entry point called from main.go.
//
// Returns:
//   - error: Returns error if command execution fails
func Execute() error {
	return rootCmd.Execute()
}

// init initializes the root command with global flags and subcommands.
func init() {
	// Global flags available to all subcommands
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"Enable verbose output for debugging")
	rootCmd.PersistentFlags().BoolVar(&useLocalCore, "local-core", false,
		"Use local eggybyte-core dependency instead of GitHub version")
	rootCmd.PersistentFlags().StringVar(&coreVersion, "core-version", "latest",
		"Specify eggybyte-core version to use (ignored when --local-core is set)")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(newCmd)
}

// logInfo prints an informational message to stdout.
// Uses color when terminal supports it.
func logInfo(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, "✓ "+format+"\n", args...)
}

// logError prints an error message to stderr.
// Uses color when terminal supports it.
func logError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "✗ Error: "+format+"\n", args...)
}

// logDebug prints a debug message if verbose mode is enabled.
func logDebug(format string, args ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stdout, "  → "+format+"\n", args...)
	}
}

// logSuccess prints a success message with emphasis.
func logSuccess(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, "\n🎉 "+format+"\n", args...)
}

// getCoreDependency returns the core dependency string based on flags.
func getCoreDependency() string {
	if useLocalCore {
		return `github.com/eggybyte-technology/go-eggybyte-core v0.0.0-00010101000000-000000000000`
	}
	return fmt.Sprintf(`github.com/eggybyte-technology/go-eggybyte-core %s`, coreVersion)
}

// getCoreReplace returns the replace directive for local core dependency.
func getCoreReplace() string {
	if useLocalCore {
		return `replace github.com/eggybyte-technology/go-eggybyte-core => ../../go-eggybyte-core`
	}
	return ""
}
