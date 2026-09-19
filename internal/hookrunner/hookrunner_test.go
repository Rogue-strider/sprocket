package hookrunner

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeRunner records every call it receives and returns a scripted
// response, so tests never depend on go/cargo/npx actually being
// installed or behaving a particular way.
type fakeRunner struct {
	calls  []string
	output string
	err    error
}

func (f *fakeRunner) Run(ctx context.Context, dir, name string, args ...string) (string, error) {
	f.calls = append(f.calls, strings.Join(append([]string{name}, args...), " "))
	return f.output, f.err
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func payloadJSON(t *testing.T, toolName, filePath string) []byte {
	t.Helper()
	p := Payload{ToolName: toolName}
	p.ToolInput.FilePath = filePath
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestRun_GoFile_InvokesGoVet(t *testing.T) {
	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")
	writeFile(t, goFile, "package main\nfunc main() {}\n")

	fr := &fakeRunner{output: ""}
	err := Run(context.Background(), bytes.NewReader(payloadJSON(t, "Write", goFile)), &bytes.Buffer{}, fr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(fr.calls) != 1 || fr.calls[0] != "go vet ./..." {
		t.Errorf("expected exactly one 'go vet ./...' call, got %v", fr.calls)
	}
}

func TestRun_RustFile_FindsCrateRootAndRunsCargoCheck(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "Cargo.toml"), "[package]\nname = \"x\"\n")
	rsFile := filepath.Join(dir, "src", "main.rs")
	writeFile(t, rsFile, "fn main() {}\n")

	fr := &fakeRunner{}
	err := Run(context.Background(), bytes.NewReader(payloadJSON(t, "Edit", rsFile)), &bytes.Buffer{}, fr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(fr.calls) != 1 || fr.calls[0] != "cargo check --quiet" {
		t.Errorf("expected 'cargo check --quiet', got %v", fr.calls)
	}
}

func TestRun_RustFile_NoCargoToml_SkipsSilently(t *testing.T) {
	dir := t.TempDir() // no Cargo.toml anywhere up the tree within this temp dir
	rsFile := filepath.Join(dir, "orphan.rs")
	writeFile(t, rsFile, "fn main() {}\n")

	fr := &fakeRunner{}
	var out bytes.Buffer
	err := Run(context.Background(), bytes.NewReader(payloadJSON(t, "Write", rsFile)), &out, fr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(fr.calls) != 0 {
		t.Errorf("expected no command to run without a Cargo.toml, got %v", fr.calls)
	}
}

func TestRun_UnhandledExtension_DoesNothing(t *testing.T) {
	dir := t.TempDir()
	txtFile := filepath.Join(dir, "notes.txt")
	writeFile(t, txtFile, "hello")

	fr := &fakeRunner{}
	err := Run(context.Background(), bytes.NewReader(payloadJSON(t, "Write", txtFile)), &bytes.Buffer{}, fr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(fr.calls) != 0 {
		t.Errorf("expected no checks for .txt files, got %v", fr.calls)
	}
}

func TestRun_NonFileTool_DoesNothing(t *testing.T) {
	fr := &fakeRunner{}
	payload := []byte(`{"tool_name":"Bash","tool_input":{"command":"ls"}}`)
	err := Run(context.Background(), bytes.NewReader(payload), &bytes.Buffer{}, fr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(fr.calls) != 0 {
		t.Errorf("expected no checks for a non-file tool, got %v", fr.calls)
	}
}

func TestRun_MalformedJSON_NeverErrors(t *testing.T) {
	fr := &fakeRunner{}
	err := Run(context.Background(), strings.NewReader("not even json"), &bytes.Buffer{}, fr)
	if err != nil {
		t.Fatalf("Run must degrade to silence on bad input, got error: %v", err)
	}
}

func TestRun_FileNoLongerExists_SkipsSilently(t *testing.T) {
	fr := &fakeRunner{}
	nonexistent := filepath.Join(t.TempDir(), "gone.go")
	err := Run(context.Background(), bytes.NewReader(payloadJSON(t, "Write", nonexistent)), &bytes.Buffer{}, fr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(fr.calls) != 0 {
		t.Errorf("expected no checks for a file that no longer exists, got %v", fr.calls)
	}
}

func TestRun_LinterOutput_IsWrittenAndTruncated(t *testing.T) {
	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")
	writeFile(t, goFile, "package main\n")

	longOutput := strings.Repeat("issue found\n", 30)
	fr := &fakeRunner{output: longOutput}
	var out bytes.Buffer
	err := Run(context.Background(), bytes.NewReader(payloadJSON(t, "Write", goFile)), &out, fr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	lineCount := strings.Count(out.String(), "\n")
	if lineCount > 20 {
		t.Errorf("expected output truncated to 20 lines, got %d lines", lineCount)
	}
	if lineCount == 0 {
		t.Error("expected linter output to be written, got nothing")
	}
}
