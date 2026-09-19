// Package learn mines a git repository's commit history for recurring
// patterns and proposes which existing skill (if any) each pattern should
// extend. It never writes anything — callers decide what to do with the
// proposals.
package learn

import (
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var conventionalRe = regexp.MustCompile(
	`^(feat|fix|refactor|chore|test|docs|perf|security)(\(([^)]+)\))?:\s*(.+)$`,
)

// signalKeywords are terms worth flagging when they recur across otherwise
// unrelated commits — often a sign of a systemic issue.
var signalKeywords = []string{
	"flaky", "timeout", "race condition", "deadlock", "n+1", "memory leak",
	"reentrancy", "overflow", "idor", "injection", "leak", "retry",
	"rate limit", "null pointer", "nil pointer", "panic", "unwrap",
}

// skillRouting maps a scope/keyword substring to the existing skill a
// recurring pattern involving it most likely extends.
var skillRouting = []struct {
	key   string
	skill string
}{
	{"auth", "web-security-checklist"},
	{"security", "web-security-checklist"},
	{"api", "api-design-checklist"},
	{"contract", "solidity-security-patterns"},
	{"sol", "solidity-security-patterns"},
	{"test", "playwright-patterns"},
	{"e2e", "playwright-patterns"},
	{"go", "go-idioms"},
	{"rust", "rust-ownership-patterns"},
	{"release", "release-checklist"},
	{"deploy", "release-checklist"},
}

const noSkillMatch = "(no existing skill matches — candidate for a new skill)"

// RouteToSkill returns the existing skill a pattern involving the given
// scope/summary text most likely extends, or noSkillMatch if nothing fits.
func RouteToSkill(scope, summary string) string {
	haystack := strings.ToLower(scope + " " + summary)
	for _, r := range skillRouting {
		if strings.Contains(haystack, r.key) {
			return r.skill
		}
	}
	return noSkillMatch
}

// ParseConventional parses a commit subject as a Conventional Commit.
// ok is false if the subject doesn't match the format.
func ParseConventional(subject string) (commitType, scope, summary string, ok bool) {
	m := conventionalRe.FindStringSubmatch(subject)
	if m == nil {
		return "", "", "", false
	}
	return m[1], m[3], m[4], true
}

type Commit struct {
	Hash    string
	Subject string
	Author  string
	Date    string
}

type Proposal struct {
	Type         string   `json:"type"`
	Pattern      string   `json:"pattern"`
	RoutesToSkill string  `json:"routes_to_skill"`
	Examples     []string `json:"examples"`
}

type Result struct {
	Repo             string     `json:"repo"`
	WindowDays       int        `json:"window_days"`
	CommitsAnalyzed  int        `json:"commits_analyzed"`
	DominantFileTypes []string  `json:"dominant_file_types,omitempty"`
	Proposals        []Proposal `json:"proposals"`
	Note             string     `json:"note"`
	Error            string     `json:"error,omitempty"`
}

// GitRunner abstracts running git commands so the analysis logic is
// testable without shelling out in unit tests if desired; the CLI uses
// RealGitRunner, which shells out to the system git binary.
type GitRunner interface {
	Log(repoDir string, sinceDays int) (string, error)
	ChangedFiles(repoDir, commitHash string) (string, error)
}

type RealGitRunner struct{}

func (RealGitRunner) Log(repoDir string, sinceDays int) (string, error) {
	cmd := exec.Command("git", "-C", repoDir, "log",
		fmt.Sprintf("--since=%d.days.ago", sinceDays),
		"--pretty=format:%H\x1f%s\x1f%an\x1f%ad", "--date=short")
	out, err := cmd.Output()
	return string(out), err
}

func (RealGitRunner) ChangedFiles(repoDir, commitHash string) (string, error) {
	cmd := exec.Command("git", "-C", repoDir, "show", "--name-only", "--pretty=format:", commitHash)
	out, err := cmd.Output()
	return string(out), err
}

func parseCommits(log string) []Commit {
	var commits []Commit
	for _, line := range strings.Split(log, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\x1f")
		if len(parts) != 4 {
			continue
		}
		commits = append(commits, Commit{Hash: parts[0], Subject: parts[1], Author: parts[2], Date: parts[3]})
	}
	return commits
}

func fileExt(path string) string {
	i := strings.LastIndex(path, ".")
	if i == -1 || i == len(path)-1 {
		return ""
	}
	// Guard against directories like ".github/workflows/validate.yml" being
	// misread — LastIndex on '.' is sufficient for our purposes since we
	// only care about the final extension of a file name.
	slash := strings.LastIndex(path, "/")
	if slash > i {
		return ""
	}
	return path[i:]
}

// Analyze mines commits in repoDir over the last windowDays days and
// returns pattern proposals for any that occur at least minOccurrences times.
func Analyze(runner GitRunner, repoDir string, windowDays, minOccurrences int) Result {
	log, err := runner.Log(repoDir, windowDays)
	if err != nil {
		return Result{Error: fmt.Sprintf("git log failed: %v", err)}
	}
	commits := parseCommits(log)
	if len(commits) == 0 {
		return Result{Error: fmt.Sprintf("No commits found in the last %d days in %s.", windowDays, repoDir)}
	}

	scopeFixCounts := map[string]int{}
	scopeExamples := map[string][]string{}
	keywordHits := map[string][]string{}
	extCounts := map[string]int{}

	for _, c := range commits {
		if ctype, scope, _, ok := ParseConventional(c.Subject); ok && ctype == "fix" {
			key := scope
			if key == "" {
				key = "(unscoped)"
			}
			scopeFixCounts[key]++
			scopeExamples[key] = append(scopeExamples[key], c.Subject)
		}

		lowered := strings.ToLower(c.Subject)
		for _, kw := range signalKeywords {
			if strings.Contains(lowered, kw) {
				keywordHits[kw] = append(keywordHits[kw], c.Subject)
			}
		}

		if changed, err := runner.ChangedFiles(repoDir, c.Hash); err == nil {
			for _, f := range strings.Split(changed, "\n") {
				if ext := fileExt(strings.TrimSpace(f)); ext != "" {
					extCounts[ext]++
				}
			}
		}
	}

	var proposals []Proposal

	// Sort scopes by count descending, then alphabetically, for stable output.
	type scopeCount struct {
		scope string
		count int
	}
	var scopes []scopeCount
	for s, n := range scopeFixCounts {
		scopes = append(scopes, scopeCount{s, n})
	}
	sort.Slice(scopes, func(i, j int) bool {
		if scopes[i].count != scopes[j].count {
			return scopes[i].count > scopes[j].count
		}
		return scopes[i].scope < scopes[j].scope
	})
	for _, sc := range scopes {
		if sc.count >= minOccurrences {
			examples := scopeExamples[sc.scope]
			if len(examples) > 5 {
				examples = examples[:5]
			}
			proposals = append(proposals, Proposal{
				Type:          "recurring_fix_scope",
				Pattern:       fmt.Sprintf("%d `fix:` commits scoped to '%s' in the last %d days", sc.count, sc.scope, windowDays),
				RoutesToSkill: RouteToSkill(sc.scope, ""),
				Examples:      examples,
			})
		}
	}

	var keywords []string
	for kw := range keywordHits {
		keywords = append(keywords, kw)
	}
	sort.Strings(keywords)
	for _, kw := range keywords {
		hits := keywordHits[kw]
		if len(hits) >= minOccurrences {
			examples := hits
			if len(examples) > 5 {
				examples = examples[:5]
			}
			proposals = append(proposals, Proposal{
				Type:          "recurring_keyword",
				Pattern:       fmt.Sprintf("'%s' mentioned in %d commit messages", kw, len(hits)),
				RoutesToSkill: RouteToSkill("", kw),
				Examples:      examples,
			})
		}
	}

	var dominantExts []string
	type extCount struct {
		ext   string
		count int
	}
	var exts []extCount
	for e, n := range extCounts {
		exts = append(exts, extCount{e, n})
	}
	sort.Slice(exts, func(i, j int) bool {
		if exts[i].count != exts[j].count {
			return exts[i].count > exts[j].count
		}
		return exts[i].ext < exts[j].ext
	})
	for i, e := range exts {
		if i >= 5 {
			break
		}
		if e.count >= minOccurrences {
			dominantExts = append(dominantExts, e.ext)
		}
	}

	return Result{
		Repo:              repoDir,
		WindowDays:        windowDays,
		CommitsAnalyzed:   len(commits),
		DominantFileTypes: dominantExts,
		Proposals:         proposals,
		Note: "These are candidate patterns only — nothing is written automatically. " +
			"Each proposal names the existing skill it would extend, or flags itself " +
			"as a candidate for a brand-new skill if nothing matches.",
	}
}

// FormatInt is a tiny helper kept for CLI flag parsing convenience.
func FormatInt(n int) string {
	return strconv.Itoa(n)
}
