// Command sprocket-learn mines a git repository's commit history for
// recurring patterns and prints proposals for new/updated skills as JSON.
// It never writes to skills/ itself — see commands/learn.md for how the
// /learn slash command uses this output.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourusername/sprocket/internal/learn"
)

func main() {
	days := flag.Int("days", 90, "lookback window in days")
	minOccurrences := flag.Int("min-occurrences", 3, "minimum occurrences to flag a pattern")
	repo := flag.String("repo", ".", "path to the git repository")
	flag.Parse()

	repoAbs, err := filepath.Abs(*repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if _, err := os.Stat(filepath.Join(repoAbs, ".git")); err != nil {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(map[string]string{"error": fmt.Sprintf("%s is not a git repository root.", repoAbs)})
		os.Exit(1)
	}

	result := learn.Analyze(learn.RealGitRunner{}, repoAbs, *days, *minOccurrences)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if result.Error != "" {
		os.Exit(1)
	}
}
