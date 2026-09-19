// Command sprocket-validate checks a sprocket plugin directory is
// structurally correct — see internal/validate for what it checks.
// Exit code 0 = all checks passed. Exit code 1 = at least one failure.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourusername/sprocket/internal/validator"
)

func main() {
	root := flag.String("root", ".", "path to the plugin root")
	flag.Parse()

	rootAbs, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	report := validator.Validate(rootAbs)

	for _, w := range report.Warnings {
		fmt.Printf("WARNING: %s\n", w)
	}
	for _, e := range report.Errors {
		fmt.Printf("ERROR: %s\n", e)
	}

	if !report.OK() {
		fmt.Printf("\n%d error(s), %d warning(s). Validation FAILED.\n", len(report.Errors), len(report.Warnings))
		os.Exit(1)
	}
	fmt.Printf("All checks passed (%d warning(s)).\n", len(report.Warnings))
}
