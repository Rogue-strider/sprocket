# Changelog

All notable changes to this project are documented here.
Format loosely follows [Keep a Changelog](https://keepachangelog.com/).

## [0.2.0] — Go rewrite, renamed kaizen → sprocket

### Changed
- Renamed the project from `kaizen` to `sprocket` (the former collided with
  several existing GitHub projects: a manga app, an AI automation agent,
  and a dev-ops agent)
- Rewrote all tooling from Python + Bash to Go: `sprocket-learn`,
  `sprocket-validate`, `sprocket-hook`. No Python or POSIX shell dependency
  left in the install/runtime path — Windows now works natively without
  Git Bash/WSL.
- Prebuilt binaries for 5 targets (linux/darwin × amd64/arm64, windows/amd64)
  ship in `bin/`; installing needs no Go toolchain
- `install.sh`/`install.ps1` now detect OS/arch and stage the matching
  binary as `bin/sprocket-*.exe` (kept `.exe` on every OS on purpose — it
  lets `hooks.json` reference one fixed path regardless of platform)
- Replaced the pytest suite with Go's native `testing` package (27 tests
  across `internal/learn`, `internal/validator`, `internal/hookrunner`,
  `internal/pkgmanager`)
- CI (`.github/workflows/validate.yml`) now runs `go vet` + `go test` +
  `go build`, plus a matrix job cross-compiling release binaries

## [0.1.0] — initial release

### Added
- 8 seed agents: feature-planner, architecture-reviewer, security-reviewer,
  build-doctor, go-code-reviewer, rust-code-reviewer, e2e-playwright-tester,
  web3-contract-reviewer
- 8 matching skills with reference checklists
- 4 slash commands: `/plan`, `/review`, `/fix-build`, `/learn`
- PostToolUse hook for auto lint/build checks on file writes (Go/Rust/TS/Solidity)
- `.mcp.json` wiring for GitHub, Supabase (read-only default), Vercel, Railway
- Continuous-learning script mining git history for recurring fix-scopes
  and risk keywords, proposing skill extensions (never writes automatically)
- Structural validator for the whole plugin
- Cross-platform installers: `install.sh` (macOS/Linux/WSL/Git Bash),
  `install.ps1` (native Windows PowerShell)
