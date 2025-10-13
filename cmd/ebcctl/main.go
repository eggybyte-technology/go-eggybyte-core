package main

import (
	"os"

	"github.com/eggybyte-technology/go-eggybyte-core/cmd/ebcctl/commands"
)

// main is the entry point for the ebcctl CLI tool.
// ebcctl (EggyByte Control) is a code generation tool that scaffolds
// new microservices and generates repository boilerplate following
// EggyByte standards.
//
// Usage:
//
//	ebcctl init <service-name>        # Create new service project
//	ebcctl new repo <model-name>      # Generate repository code
//	ebcctl new service <service-name> # Generate service template
//
// All generated code follows EggyByte quality standards:
//   - English comments with comprehensive documentation
//   - Methods under 50 lines
//   - Auto-registration patterns
//   - Standard three-layer architecture
func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
