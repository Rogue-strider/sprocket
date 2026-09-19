package learn

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// makeRepo creates a throwaway git repo with one commit per entry in
// commits (subject -> filename to touch), returning its path.
func makeRepo(t *testing.T, commits []struct {
	subject string
	file    string
}) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")

	for i, c := range commits {
		path := filepath.Join(dir, c.file)
		if err := os.WriteFile(path, []byte(string(rune('a'+i))), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}
		run("add", "-A")
		run("commit", "-q", "-m", c.subject)
	}
	return dir
}

func TestParseConventional(t *testing.T) {
	ctype, scope, summary, ok := ParseConventional("fix(auth): validate JWT expiry")
	if !ok {
		t.Fatal("expected match")
	}
	if ctype != "fix" || scope != "auth" || summary != "validate JWT expiry" {
		t.Errorf("got type=%q scope=%q summary=%q", ctype, scope, summary)
	}
}

func TestParseConventionalNonMatch(t *testing.T) {
	_, _, _, ok := ParseConventional("quick fix for the thing")
	if ok {
		t.Error("expected no match for non-conventional subject")
	}
}

func TestRouteToSkillKnownScope(t *testing.T) {
	if got := RouteToSkill("auth", ""); got != "web-security-checklist" {
		t.Errorf("got %q", got)
	}
	if got := RouteToSkill("", "flaky e2e test"); got != "playwright-patterns" {
		t.Errorf("got %q", got)
	}
}

func TestRouteToSkillFallback(t *testing.T) {
	got := RouteToSkill("unrelated-thing", "nothing matches here")
	if got != noSkillMatch {
		t.Errorf("expected fallback, got %q", got)
	}
}

func TestAnalyzeDetectsRecurringFixScope(t *testing.T) {
	dir := makeRepo(t, []struct {
		subject string
		file    string
	}{
		{"fix(auth): validate JWT expiry before role check", "a.go"},
		{"fix(auth): reject expired refresh tokens", "b.go"},
		{"fix(auth): handle nil pointer on missing session", "c.go"},
		{"feat(ui): add dark mode toggle", "d.tsx"},
	})

	result := Analyze(RealGitRunner{}, dir, 365, 3)
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}

	var found *Proposal
	for i := range result.Proposals {
		if result.Proposals[i].Type == "recurring_fix_scope" {
			found = &result.Proposals[i]
		}
	}
	if found == nil {
		t.Fatal("expected a recurring_fix_scope proposal")
	}
	if found.RoutesToSkill != "web-security-checklist" {
		t.Errorf("got routes_to_skill=%q", found.RoutesToSkill)
	}
	if len(found.Examples) != 3 {
		t.Errorf("expected 3 examples, got %d", len(found.Examples))
	}
}

func TestAnalyzeBelowThresholdNotFlagged(t *testing.T) {
	dir := makeRepo(t, []struct {
		subject string
		file    string
	}{
		{"fix(auth): validate JWT expiry", "a.go"},
		{"fix(auth): reject bad token", "b.go"},
	})

	result := Analyze(RealGitRunner{}, dir, 365, 3)
	for _, p := range result.Proposals {
		if p.Type == "recurring_fix_scope" {
			t.Errorf("did not expect a proposal below threshold, got %+v", p)
		}
	}
}

func TestAnalyzeDetectsRecurringKeyword(t *testing.T) {
	dir := makeRepo(t, []struct {
		subject string
		file    string
	}{
		{"fix(e2e): flaky checkout test", "a.spec.ts"},
		{"fix(e2e): flaky login test", "b.spec.ts"},
		{"fix(e2e): flaky signup test", "c.spec.ts"},
	})

	result := Analyze(RealGitRunner{}, dir, 365, 3)
	found := false
	for _, p := range result.Proposals {
		if p.Type == "recurring_keyword" {
			found = true
		}
	}
	if !found {
		t.Error("expected a recurring_keyword proposal for 'flaky'")
	}
}

func TestAnalyzeEmptyRepoReturnsError(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("git", "-C", dir, "init", "-q")
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}

	result := Analyze(RealGitRunner{}, dir, 90, 3)
	if result.Error == "" {
		t.Error("expected an error for empty repo")
	}
}

func TestAnalyzeCommitCountMatches(t *testing.T) {
	dir := makeRepo(t, []struct {
		subject string
		file    string
	}{
		{"fix(auth): bug one", "a.go"},
		{"fix(auth): bug two", "b.go"},
		{"fix(auth): bug three", "c.go"},
	})

	result := Analyze(RealGitRunner{}, dir, 365, 3)
	if result.CommitsAnalyzed != 3 {
		t.Errorf("expected 3 commits analyzed, got %d", result.CommitsAnalyzed)
	}
	if len(result.Proposals) < 1 {
		t.Error("expected at least one proposal")
	}
}
