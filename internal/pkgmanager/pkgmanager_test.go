package pkgmanager

import (
	"os"
	"path/filepath"
	"testing"
)

func touch(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(""), 0644); err != nil {
		t.Fatalf("failed to create %s: %v", name, err)
	}
}

func TestDetect_NoLockfile_DefaultsToNpm(t *testing.T) {
	dir := t.TempDir()
	got := Detect(dir)
	if got.Name != "npm" {
		t.Errorf("Detect() = %q, want npm", got.Name)
	}
}

func TestDetect_Pnpm(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "pnpm-lock.yaml")
	if got := Detect(dir); got.Name != "pnpm" {
		t.Errorf("Detect() = %q, want pnpm", got.Name)
	}
}

func TestDetect_Yarn(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "yarn.lock")
	if got := Detect(dir); got.Name != "yarn" {
		t.Errorf("Detect() = %q, want yarn", got.Name)
	}
}

func TestDetect_Bun(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "bun.lockb")
	if got := Detect(dir); got.Name != "bun" {
		t.Errorf("Detect() = %q, want bun", got.Name)
	}
}

func TestDetect_PrefersPnpmWhenMultipleLockfilesPresent(t *testing.T) {
	// If a repo has stale lockfiles from a prior migration, pnpm (the
	// increasingly common default) wins deterministically rather than
	// whichever branch happens to be checked first by accident.
	dir := t.TempDir()
	touch(t, dir, "yarn.lock")
	touch(t, dir, "pnpm-lock.yaml")
	if got := Detect(dir); got.Name != "pnpm" {
		t.Errorf("Detect() = %q, want pnpm (priority order)", got.Name)
	}
}

func TestManager_ExecCommandsAreNonEmpty(t *testing.T) {
	for _, m := range []Manager{npm, pnpm, yarn, bun} {
		if m.Run == "" || m.Exec == "" {
			t.Errorf("manager %q has an empty Run or Exec field: %+v", m.Name, m)
		}
	}
}
