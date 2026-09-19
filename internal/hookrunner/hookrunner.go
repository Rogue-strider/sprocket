// Package hookrunner implements sprocket's PostToolUse hook: after Claude
// writes or edits a file, it runs a fast, language-appropriate check
// (go vet, cargo check, eslint, hardhat compile) and reports the result.
//
// PostToolUse hooks cannot block the tool call that already ran (the file
// is already written by the time this fires), so Run never returns a
// non-nil error for "the lint found problems" — it only returns an error
// for genuine internal failures (e.g. malformed payload), and even then
// the caller should still exit 0: a broken hook must never break the
// user's session.
package hookrunner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourusername/sprocket/internal/pkgmanager"
)

// Payload mirrors the JSON Claude Code sends on stdin for PostToolUse hooks.
// Only the fields sprocket actually reads are declared.
type Payload struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
}

// Runner executes an external command and captures combined output.
// Abstracted so tests can substitute a fake without requiring go/cargo/npx
// to actually be installed in the test environment.
type Runner interface {
	Run(ctx context.Context, dir, name string, args ...string) (output string, err error)
}

// ExecRunner is the real Runner, backed by os/exec.
type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, dir, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// checkerFor maps a file extension to the check it should run. Returning
// ("", false) means "no check defined for this extension" — not an error.
func checkerFor(ext string) (check func(ctx context.Context, filePath string, r Runner) string, ok bool) {
	switch ext {
	case ".go":
		return checkGo, true
	case ".rs":
		return checkRust, true
	case ".ts", ".tsx", ".js", ".jsx":
		return checkJSTS, true
	case ".sol":
		return checkSolidity, true
	default:
		return nil, false
	}
}

func checkGo(ctx context.Context, filePath string, r Runner) string {
	dir := filepath.Dir(filePath)
	out, err := r.Run(ctx, dir, "go", "vet", "./...")
	if err != nil && out == "" {
		return "" // go toolchain not available or nothing to report — stay silent
	}
	return truncate(out, 20)
}

// findCrateRoot walks up from dir looking for the nearest Cargo.toml, so
// `cargo check` runs at the crate root rather than failing because it was
// invoked from a subdirectory.
func findCrateRoot(dir string) (string, bool) {
	for {
		if _, err := os.Stat(filepath.Join(dir, "Cargo.toml")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func checkRust(ctx context.Context, filePath string, r Runner) string {
	crateDir, found := findCrateRoot(filepath.Dir(filePath))
	if !found {
		return ""
	}
	out, err := r.Run(ctx, crateDir, "cargo", "check", "--quiet")
	if err != nil && out == "" {
		return ""
	}
	return truncate(out, 20)
}

func checkJSTS(ctx context.Context, filePath string, r Runner) string {
	dir := filepath.Dir(filePath)
	mgr := pkgmanager.Detect(dir)
	// mgr.Exec is like "npx" or "pnpm dlx" — split into command + leading args.
	parts := strings.Fields(mgr.Exec)
	if len(parts) == 0 {
		return ""
	}
	args := append(append([]string{}, parts[1:]...), "--no-install", "eslint", filePath)
	out, err := r.Run(ctx, dir, parts[0], args...)
	if err != nil && out == "" {
		return ""
	}
	return truncate(out, 20)
}

func checkSolidity(ctx context.Context, filePath string, r Runner) string {
	dir := filepath.Dir(filePath)
	hasHardhat := fileExists(filepath.Join(dir, "hardhat.config.js")) || fileExists(filepath.Join(dir, "hardhat.config.ts"))
	if !hasHardhat {
		return ""
	}
	out, err := r.Run(ctx, dir, "npx", "--no-install", "hardhat", "compile")
	if err != nil && out == "" {
		return ""
	}
	return truncate(out, 20)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func truncate(s string, maxLines int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n")
}

// Run reads a hook payload from r, runs the matching check if the touched
// file has one, and writes any output to w. It always returns nil — a
// hook that fails internally should degrade to silence, never crash the
// session or block Claude's next action.
func Run(ctx context.Context, r io.Reader, w io.Writer, runner Runner) error {
	var payload Payload
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		return nil // malformed/empty payload — nothing to check, stay silent
	}
	if payload.ToolInput.FilePath == "" {
		return nil
	}
	if _, err := os.Stat(payload.ToolInput.FilePath); err != nil {
		return nil // file doesn't exist (already deleted, race, etc.) — skip
	}

	ext := filepath.Ext(payload.ToolInput.FilePath)
	check, ok := checkerFor(ext)
	if !ok {
		return nil
	}

	// Hard timeout so a hung linter can never hang the user's session.
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	output := check(ctx, payload.ToolInput.FilePath, runner)
	if output != "" {
		fmt.Fprintln(w, output)
	}
	return nil
}
