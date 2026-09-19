# Changelog

All notable changes to this project are documented here.
Format loosely follows [Keep a Changelog](https://keepachangelog.com/).

## [0.3.0] — AWS/infra agent

### Added
- 9th agent: `aws-infra-reviewer` — reviews Terraform/CDK/CloudFormation,
  IAM policies, Dockerfiles, and deployment CI/CD for security and
  blast-radius issues (IAM over-permissioning, exposed security groups,
  unsafe state management, secrets in image layers, ungated production
  deploys). Added because AWS/deployment was named as an actual gap in the
  existing 8 — not guessed in advance.
- Matching skill: `aws-infra-checklist`
- `/review` now routes `.tf` files, CDK/CloudFormation templates,
  Dockerfiles, `docker-compose.yml`, and deployment workflow files to the
  new agent
- README notes that `bin/sprocket-*.exe` (the staged, gitignored copies)
  need `./install.sh`/`install.ps1` re-run after any `git pull` that
  touches `bin/` — found from a real confusion during Windows testing
  where `git pull` updated the platform binaries but not the staged ones

## [0.2.3] — Windows CRLF fix

### Fixed
- `sprocket-validate` reported **every single agent/skill/command as
  "missing YAML frontmatter"** on a real Windows machine. Root cause:
  Windows `git clone` with `core.autocrlf` converts LF line endings to
  CRLF on checkout, and the frontmatter-matching regex only matched a bare
  `\n`, so `---\r\n` never matched `---\n`. Fixed by normalizing CRLF → LF
  before parsing. Added a regression test
  (`TestValidate_CRLFLineEndings_StillPasses`) that reproduces this exact
  failure against a fixture with real CRLF line endings — confirmed the
  test fails against the old code and passes against the fix before
  shipping this.
- Defensively hardened `sprocket-learn`'s git-log parser the same way
  (stray `\r` stripped per line) even though no failure was observed there
  — same class of bug, cheap to close preemptively.
- Found via a real end-to-end test on an actual Windows 11 machine (Git
  Bash inside VS Code) — the second real bug this kind of testing has
  caught that the Linux-only build sandbox could not have caught itself.

## [0.2.2] — Git Bash detection fix

### Fixed
- `install.sh` failed on Windows Git Bash: `uname -s` reports
  `MINGW64_NT-...` there, which didn't match the `Linux`/`Darwin` cases and
  fell through to "Unsupported OS". Added `MINGW*|MSYS*|CYGWIN*` detection
  so Git Bash correctly resolves to the `windows-amd64` binaries.
- Fixed the corresponding source filename lookup: prebuilt Windows binaries
  are named with a `.exe` suffix already (`sprocket-learn-windows-amd64.exe`)
  while Linux/macOS ones aren't — the staging step now accounts for that
  per-platform instead of assuming one naming scheme.
- Found via a real test on an actual Windows machine (Git Bash inside
  VS Code's integrated terminal) — first real-world run outside the build
  sandbox, and it caught a genuine bug the sandbox couldn't have caught
  (the sandbox only has Linux's `uname -s` output to test against).

## [0.2.1] — GitHub-ready

### Changed
- Filled in all placeholders: author/owner set to `Rogue-strider`, module
  path is `github.com/Rogue-strider/sprocket`, LICENSE copyright line set
- All 15 binaries rebuilt with the real import path (verified via `strings`
  that the correct module path is baked into the compiled output)
- README's push section now has copy-pasteable `git` commands instead of a
  placeholder checklist

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
