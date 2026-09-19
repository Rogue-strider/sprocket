// Package pkgmanager detects which JavaScript package manager a project
// uses, based on lockfile presence, so the hook runner can invoke the
// right one without the user configuring anything per-repo.
package pkgmanager

import (
	"os"
	"path/filepath"
)

// Manager describes a detected package manager and how to run/exec with it.
type Manager struct {
	Name string // "npm", "pnpm", "yarn", "bun"
	Run  string // command used to run a package script, e.g. "npm run"
	Exec string // command used to run a one-off binary, e.g. "npx"
}

var (
	pnpm = Manager{Name: "pnpm", Run: "pnpm", Exec: "pnpm dlx"}
	yarn = Manager{Name: "yarn", Run: "yarn", Exec: "yarn dlx"}
	bun  = Manager{Name: "bun", Run: "bun", Exec: "bunx"}
	npm  = Manager{Name: "npm", Run: "npm", Exec: "npx"}
)

// Detect inspects dir for known lockfiles and returns the matching manager.
// Falls back to npm/npx, which every Node.js install ships with, when no
// lockfile is found.
func Detect(dir string) Manager {
	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(dir, name))
		return err == nil
	}

	switch {
	case exists("pnpm-lock.yaml"):
		return pnpm
	case exists("yarn.lock"):
		return yarn
	case exists("bun.lockb"), exists("bun.lock"):
		return bun
	default:
		return npm
	}
}
