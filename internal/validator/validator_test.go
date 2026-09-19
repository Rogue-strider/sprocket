package validator

import (
	"os"
	"path/filepath"
	"testing"
)

// writeValidPlugin creates a minimal, fully valid plugin structure in dir.
// Individual tests then corrupt one piece of it to prove validator catches
// exactly that failure — not a blanket "something's wrong somewhere" check.
func writeValidPlugin(t *testing.T, dir string) {
	t.Helper()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}

	must(os.MkdirAll(filepath.Join(dir, ".claude-plugin"), 0755))
	must(os.WriteFile(filepath.Join(dir, ".claude-plugin", "plugin.json"),
		[]byte(`{"name":"test-plugin","license":"MIT"}`), 0644))
	must(os.WriteFile(filepath.Join(dir, ".claude-plugin", "marketplace.json"),
		[]byte(`{"name":"test-marketplace","plugins":[{"name":"test-plugin","source":"./"}]}`), 0644))

	must(os.MkdirAll(filepath.Join(dir, "agents"), 0755))
	must(os.WriteFile(filepath.Join(dir, "agents", "my-agent.md"),
		[]byte("---\nname: my-agent\ndescription: does a thing\n---\nBody.\n"), 0644))

	must(os.MkdirAll(filepath.Join(dir, "commands"), 0755))
	must(os.WriteFile(filepath.Join(dir, "commands", "my-cmd.md"),
		[]byte("---\ndescription: runs a thing\n---\nBody.\n"), 0644))

	must(os.MkdirAll(filepath.Join(dir, "skills", "my-skill"), 0755))
	must(os.WriteFile(filepath.Join(dir, "skills", "my-skill", "SKILL.md"),
		[]byte("---\nname: my-skill\ndescription: a skill\n---\nBody.\n"), 0644))

	must(os.MkdirAll(filepath.Join(dir, "hooks"), 0755))
	must(os.MkdirAll(filepath.Join(dir, "scripts"), 0755))
	scriptPath := filepath.Join(dir, "scripts", "check.sh")
	must(os.WriteFile(scriptPath, []byte("#!/bin/sh\necho ok\n"), 0755))
	hooksJSON := `{"hooks":{"PostToolUse":[{"matcher":"Write","hooks":[{"type":"command","command":"${CLAUDE_PLUGIN_ROOT}/scripts/check.sh"}]}]}}`
	must(os.WriteFile(filepath.Join(dir, "hooks", "hooks.json"), []byte(hooksJSON), 0644))

	must(os.WriteFile(filepath.Join(dir, ".mcp.json"),
		[]byte(`{"mcpServers":{"github":{"url":"https://example.com/mcp"}}}`), 0644))
}

func TestValidate_ValidPlugin_PassesCleanly(t *testing.T) {
	dir := t.TempDir()
	writeValidPlugin(t, dir)

	r := Validate(dir)
	if !r.OK() {
		t.Errorf("expected a valid plugin to pass, got errors: %v", r.Errors)
	}
}

func TestValidate_MissingPluginManifest_Fails(t *testing.T) {
	dir := t.TempDir()
	writeValidPlugin(t, dir)
	os.Remove(filepath.Join(dir, ".claude-plugin", "plugin.json"))

	r := Validate(dir)
	if r.OK() {
		t.Fatal("expected failure for missing plugin.json")
	}
	assertContains(t, r.Errors, "missing required file")
}

func TestValidate_InvalidJSON_Fails(t *testing.T) {
	dir := t.TempDir()
	writeValidPlugin(t, dir)
	os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte("{not valid json"), 0644)

	r := Validate(dir)
	if r.OK() {
		t.Fatal("expected failure for invalid JSON in .mcp.json")
	}
	assertContains(t, r.Errors, "invalid JSON")
}

func TestValidate_AgentFrontmatterNameMismatch_Fails(t *testing.T) {
	dir := t.TempDir()
	writeValidPlugin(t, dir)
	os.WriteFile(filepath.Join(dir, "agents", "my-agent.md"),
		[]byte("---\nname: wrong-name\ndescription: does a thing\n---\nBody.\n"), 0644)

	r := Validate(dir)
	if r.OK() {
		t.Fatal("expected failure for mismatched agent frontmatter name")
	}
	assertContains(t, r.Errors, "doesn't match filename")
}

func TestValidate_MissingFrontmatter_Fails(t *testing.T) {
	dir := t.TempDir()
	writeValidPlugin(t, dir)
	os.WriteFile(filepath.Join(dir, "agents", "my-agent.md"), []byte("no frontmatter here\n"), 0644)

	r := Validate(dir)
	if r.OK() {
		t.Fatal("expected failure for missing frontmatter")
	}
	assertContains(t, r.Errors, "missing YAML frontmatter")
}

func TestValidate_NonExecutableHookScript_Fails(t *testing.T) {
	dir := t.TempDir()
	writeValidPlugin(t, dir)
	os.Chmod(filepath.Join(dir, "scripts", "check.sh"), 0644) // remove exec bit

	r := Validate(dir)
	if r.OK() {
		t.Fatal("expected failure for non-executable hook script")
	}
	assertContains(t, r.Errors, "not executable")
}

func TestValidate_NonKebabCaseName_Fails(t *testing.T) {
	dir := t.TempDir()
	writeValidPlugin(t, dir)
	os.WriteFile(filepath.Join(dir, ".claude-plugin", "plugin.json"),
		[]byte(`{"name":"Not_Kebab_Case"}`), 0644)

	r := Validate(dir)
	if r.OK() {
		t.Fatal("expected failure for non-kebab-case plugin name")
	}
	assertContains(t, r.Errors, "not kebab-case")
}

func TestValidate_MissingSkillFile_Fails(t *testing.T) {
	dir := t.TempDir()
	writeValidPlugin(t, dir)
	os.Remove(filepath.Join(dir, "skills", "my-skill", "SKILL.md"))

	r := Validate(dir)
	if r.OK() {
		t.Fatal("expected failure for a skill directory missing SKILL.md")
	}
	assertContains(t, r.Errors, "missing SKILL.md")
}

func TestValidate_MissingLicense_WarnsButDoesNotFail(t *testing.T) {
	dir := t.TempDir()
	writeValidPlugin(t, dir)
	os.WriteFile(filepath.Join(dir, ".claude-plugin", "plugin.json"), []byte(`{"name":"test-plugin"}`), 0644)

	r := Validate(dir)
	if !r.OK() {
		t.Errorf("missing license should be a warning, not a failure; got errors: %v", r.Errors)
	}
	assertContains(t, r.Warnings, "no 'license' field")
}

func assertContains(t *testing.T, list []string, substr string) {
	t.Helper()
	for _, s := range list {
		if contains(s, substr) {
			return
		}
	}
	t.Errorf("expected one entry containing %q, got: %v", substr, list)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (func() bool {
		for i := 0; i+len(substr) <= len(s); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()
}
