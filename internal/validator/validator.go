// Package validator checks that a sprocket plugin directory is
// structurally correct: manifests are valid JSON with required fields,
// every hook script it references exists and is executable, every
// agent/command/skill has valid frontmatter, and naming follows the
// kebab-case convention Claude Code expects.
package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	kebabRE       = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	frontmatterRE = regexp.MustCompile(`(?s)^---\n(.*?)\n---\n`)
)

// Result holds everything found during validation. Errors mean the plugin
// is broken; Warnings are advisory.
type Result struct {
	Errors   []string
	Warnings []string
}

func (r *Result) fail(format string, args ...any) {
	r.Errors = append(r.Errors, fmt.Sprintf(format, args...))
}

func (r *Result) warn(format string, args ...any) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, args...))
}

func (r *Result) OK() bool { return len(r.Errors) == 0 }

func rel(root, path string) string {
	if r, err := filepath.Rel(root, path); err == nil {
		return r
	}
	return path
}

func loadJSON(root, path string, r *Result) map[string]any {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // caller decides whether the file is required
		}
		r.fail("cannot read %s: %v", rel(root, path), err)
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		r.fail("invalid JSON in %s: %v", rel(root, path), err)
		return nil
	}
	return out
}

func checkKebabCase(name, context string, r *Result) {
	if !kebabRE.MatchString(name) {
		r.fail("%s: %q is not kebab-case (lowercase, hyphen-separated)", context, name)
	}
}

// frontmatter parses the leading "---\n...\n---\n" YAML block of a file
// into a flat string map. It's intentionally not a full YAML parser —
// sprocket's frontmatter is always flat key: value pairs, and a minimal
// parser here avoids pulling in a YAML dependency for something this small.
func parseFrontmatter(path string, r *Result, root string) map[string]string {
	data, err := os.ReadFile(path)
	if err != nil {
		r.fail("cannot read %s: %v", rel(root, path), err)
		return nil
	}
	// Normalize CRLF -> LF before matching. Windows git checkouts commonly
	// convert LF to CRLF on checkout (core.autocrlf), which otherwise makes
	// every file here false-fail as "missing frontmatter" since the regex
	// only matched a bare \n.
	normalized := strings.ReplaceAll(string(data), "\r\n", "\n")
	m := frontmatterRE.FindStringSubmatch(normalized)
	if m == nil {
		r.fail("%s: missing YAML frontmatter (must start with '---')", rel(root, path))
		return nil
	}
	fm := map[string]string{}
	for _, line := range strings.Split(m[1], "\n") {
		if idx := strings.Index(line, ":"); idx != -1 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			fm[key] = value
		}
	}
	return fm
}

func checkPluginManifest(root string, r *Result) {
	path := filepath.Join(root, ".claude-plugin", "plugin.json")
	data := loadJSON(root, path, r)
	if data == nil {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			r.fail("missing required file: %s", rel(root, path))
		}
		return
	}
	name, ok := data["name"].(string)
	if !ok || name == "" {
		r.fail("%s: missing required 'name' field", rel(root, path))
	} else {
		checkKebabCase(name, rel(root, path), r)
	}
	if _, ok := data["license"]; !ok {
		r.warn("%s: no 'license' field set", rel(root, path))
	}
}

func checkMarketplaceManifest(root string, r *Result) {
	path := filepath.Join(root, ".claude-plugin", "marketplace.json")
	data := loadJSON(root, path, r)
	if data == nil {
		return
	}
	plugins, _ := data["plugins"].([]any)
	for _, p := range plugins {
		entry, ok := p.(map[string]any)
		if !ok {
			continue
		}
		_, hasName := entry["name"]
		_, hasSource := entry["source"]
		if !hasName || !hasSource {
			r.fail("%s: plugin entry missing 'name' or 'source': %v", rel(root, path), entry)
		}
	}
}

func checkHooks(root string, r *Result) {
	path := filepath.Join(root, "hooks", "hooks.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return // hooks are optional
	}
	data := loadJSON(root, path, r)
	if data == nil {
		return
	}
	hooks, _ := data["hooks"].(map[string]any)
	for event, matchers := range hooks {
		matcherList, _ := matchers.([]any)
		for _, mb := range matcherList {
			block, _ := mb.(map[string]any)
			hookList, _ := block["hooks"].([]any)
			for _, h := range hookList {
				hook, _ := h.(map[string]any)
				cmdRaw, _ := hook["command"].(string)
				resolved := strings.ReplaceAll(cmdRaw, "${CLAUDE_PLUGIN_ROOT}", root)
				fields := strings.Fields(resolved)
				if len(fields) == 0 {
					continue
				}
				scriptPath := fields[0]
				ext := filepath.Ext(scriptPath)
				if ext == ".sh" || ext == ".py" {
					info, err := os.Stat(scriptPath)
					if err != nil {
						r.fail("hooks.json (%s): referenced script does not exist: %s", event, cmdRaw)
						continue
					}
					if ext == ".sh" && info.Mode()&0111 == 0 {
						r.fail("hooks.json (%s): script is not executable: %s", event, rel(root, scriptPath))
					}
				}
			}
		}
	}
}

func checkMCP(root string, r *Result) {
	path := filepath.Join(root, ".mcp.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return // MCP config is optional
	}
	data := loadJSON(root, path, r)
	if data == nil {
		return
	}
	servers, _ := data["mcpServers"].(map[string]any)
	for name, s := range servers {
		server, _ := s.(map[string]any)
		_, hasURL := server["url"]
		_, hasCommand := server["command"]
		if !hasURL && !hasCommand {
			r.fail(".mcp.json: server %q has neither 'url' nor 'command'", name)
		}
	}
}

func checkAgents(root string, r *Result) {
	dir := filepath.Join(root, "agents")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // agents/ is optional
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		stem := strings.TrimSuffix(e.Name(), ".md")
		checkKebabCase(stem, fmt.Sprintf("agents/%s", e.Name()), r)
		fm := parseFrontmatter(filepath.Join(dir, e.Name()), r, root)
		if fm == nil {
			continue
		}
		for _, key := range []string{"name", "description"} {
			if _, ok := fm[key]; !ok {
				r.fail("agents/%s: frontmatter missing required '%s'", e.Name(), key)
			}
		}
		if name, ok := fm["name"]; ok && name != stem {
			r.fail("agents/%s: frontmatter name %q doesn't match filename", e.Name(), name)
		}
	}
}

func checkCommands(root string, r *Result) {
	dir := filepath.Join(root, "commands")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // commands/ is optional
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		stem := strings.TrimSuffix(e.Name(), ".md")
		checkKebabCase(stem, fmt.Sprintf("commands/%s", e.Name()), r)
		fm := parseFrontmatter(filepath.Join(dir, e.Name()), r, root)
		if fm == nil {
			continue
		}
		if _, ok := fm["description"]; !ok {
			r.warn("commands/%s: frontmatter missing 'description' (recommended)", e.Name())
		}
	}
}

func checkSkills(root string, r *Result) {
	dir := filepath.Join(root, "skills")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // skills/ is optional
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		checkKebabCase(e.Name(), fmt.Sprintf("skills/%s", e.Name()), r)
		skillFile := filepath.Join(dir, e.Name(), "SKILL.md")
		if _, err := os.Stat(skillFile); err != nil {
			r.fail("skills/%s: missing SKILL.md", e.Name())
			continue
		}
		fm := parseFrontmatter(skillFile, r, root)
		if fm == nil {
			continue
		}
		for _, key := range []string{"name", "description"} {
			if _, ok := fm[key]; !ok {
				r.fail("skills/%s/SKILL.md: frontmatter missing required '%s'", e.Name(), key)
			}
		}
		if name, ok := fm["name"]; ok && name != e.Name() {
			r.fail("skills/%s/SKILL.md: frontmatter name %q doesn't match directory name", e.Name(), name)
		}
	}
}

// Validate runs every check against the plugin rooted at root.
func Validate(root string) *Result {
	r := &Result{}
	checkPluginManifest(root, r)
	checkMarketplaceManifest(root, r)
	checkHooks(root, r)
	checkMCP(root, r)
	checkAgents(root, r)
	checkCommands(root, r)
	checkSkills(root, r)
	return r
}
