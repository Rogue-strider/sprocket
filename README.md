# sprocket

![CI](https://img.shields.io/badge/CI-vet%20%2B%20test%20%2B%20cross--compile-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Go](https://img.shields.io/badge/tooling-Go-00ADD8)

A plug-and-play Claude Code plugin: specialized agents, skills, slash commands,
hooks, and pre-wired MCP servers — with a continuous-learning loop that reads
your git history and proposes new skills as recurring patterns emerge.

Everything here is free: Claude Code itself, all four MCP servers below
(GitHub, Supabase, Vercel, Railway all offer free-tier accounts and their
MCP endpoints have no separate cost beyond your own usage), and the plugin
runs entirely on your own machine.

## Why Go for the tooling

The agents/skills/commands are markdown prompt files — no language question
there. But the actual logic (mining git history, validating the plugin,
running post-edit checks) used to be Python + Bash, which meant every
install implicitly depended on the user having `python3` on PATH and, on
Windows, a POSIX shell (Git Bash/WSL). That's a bad fit for "cross-platform,
zero-friction install."

It's rewritten in Go now: three small binaries (`sprocket-learn`,
`sprocket-validate`, `sprocket-hook`), prebuilt for Linux/macOS/Windows ×
amd64/arm64, shipped in `bin/`. Installing needs **no Go toolchain, no
Python, no shell dependency** — just the right binary for your machine,
which `install.sh`/`install.ps1` stages automatically.

## This isn't a mockup — how to verify that yourself

```bash
go vet ./...      # static analysis, clean
go test ./... -v  # 27 tests across 4 packages, real git repos with planted
                   # commit histories and deliberately-broken plugin fixtures
                   # — not mocked output
go build ./...    # builds all three binaries + their internal packages
```
All three run in CI (`.github/workflows/validate.yml`) on every push, plus a
matrix job that cross-compiles release binaries for all 5 targets.

## What's in v0.1

- **8 agents** (`agents/`) — feature-planner, architecture-reviewer,
  security-reviewer, build-doctor, go-code-reviewer, rust-code-reviewer,
  e2e-playwright-tester, web3-contract-reviewer
- **8 skills** (`skills/`) — reference checklists the agents (and Claude
  generally) pull in on demand: Go idioms, Rust ownership patterns, web
  security checklist, Playwright patterns, Solidity security patterns, API
  design checklist, git commit conventions, release checklist
- **4 commands** (`commands/`) — `/plan`, `/review`, `/fix-build`, `/learn`
- **1 hook** (`hooks/hooks.json`) — `sprocket-hook.exe` runs a fast
  language-appropriate check after every file write (`go vet`, `cargo
  check`, eslint, `hardhat compile`), auto-detecting the JS package manager
  per-project via `internal/pkgmanager`
- **MCP config** (`.mcp.json`) — GitHub, Supabase (read-only by default),
  Vercel, Railway, all via their official hosted remote endpoints with
  browser OAuth (no tokens to manage)
- **Learning binary** (`sprocket-learn.exe`) — mines commit history for
  recurring fix-scopes and risk keywords, proposes which skill file each
  pattern should extend, never writes anything without you reviewing it
  first via `/learn`
- **Validator binary** (`sprocket-validate.exe`) — structural linter for the
  whole plugin, run in CI and available to run yourself any time

## Install

**macOS / Linux / Windows (Git Bash or WSL):**
```bash
./install.sh
```

**Windows (native PowerShell):**
```powershell
.\install.ps1
```

Both scripts: detect your OS/arch, stage the matching prebuilt binary as
`bin/sprocket-{learn,validate,hook}.exe` (the `.exe` suffix is kept on every
platform on purpose — Linux/macOS execute it fine, and it lets `hooks.json`
reference one fixed path regardless of OS), then register the local
marketplace and install the plugin:
```
/plugin marketplace add /path/to/sprocket
/plugin install sprocket@sprocket-marketplace
```

No binary for your platform? `go build -o bin/sprocket-learn.exe
./cmd/sprocket-learn` (and same for `-validate`/`-hook`) if you have Go
installed — the installer prints this exact command when it can't find a
match.

## Honest roadmap — this is not 64 agents yet

The inspiration for this (ECC, the agent-harness project you found) got to
64 agents / 262 skills through months of real usage feeding its learning
loop, not by hand-authoring everything upfront. Doing that here would mostly
produce filler nobody uses. The plan:

1. **Now:** 8 agents/skills that are actually good, in the languages you
   know (Go, Rust, Web3) plus the generically useful ones (planning,
   architecture, security, build fixes, e2e testing).
2. **Use it for real work.** Every time you hit a task none of the 8 agents
   fit well, that's the signal for agent #9 — not a guess in advance.
3. **Run `/learn` periodically** (weekly, or after finishing a feature). It
   reads your git history and tells you which skills are getting reinforced
   by real bugs vs. which agents/skills you added but never actually needed.
4. **Grow deliberately.** Each new agent/skill should come from an actual
   gap you hit, which is exactly how the number climbs to 20, then 40, then
   whatever it settles at for how you actually work — not a fixed target.

## Adding a new agent

Copy the shape of an existing one in `agents/`: YAML frontmatter with
`name`, `description` (this is what tells Claude *when* to invoke it — be
specific about trigger phrases), and `tools` (the minimum tool set it
needs), then the system prompt body. Same pattern for `skills/*/SKILL.md`
and `commands/*.md`. `sprocket-validate.exe` checks every one of these
matches its filename/directory and has required frontmatter fields.

## Extending the Go tooling

- `internal/learn` — the pattern-mining logic (`cmd/sprocket-learn` is a
  thin CLI wrapper around it)
- `internal/validator` — the structural checks (`cmd/sprocket-validate`
  wraps it)
- `internal/hookrunner` — the PostToolUse dispatch logic (`cmd/sprocket-hook`
  wraps it); add a new language check by adding a case to `checkerFor` and a
  `check<Lang>` function
- `internal/pkgmanager` — JS package-manager detection, used by
  `hookrunner`'s ESLint check

After changing any of these: `go test ./...`, then re-run the cross-compile
step (see `.github/workflows/validate.yml`'s `cross-compile` job, or copy
the loop from `install.sh`) to refresh `bin/`.

## Extending the MCP config

`.mcp.json` currently points at the four services you use. Supabase is
scoped `read_only=true` by default (recommended — see Supabase's own
security guidance on connecting MCP to a database); remove that query
param if you want write access. Add a `project_ref=<id>` query param to
scope Supabase to one project instead of your whole account.

## Pushing to GitHub

The placeholders are already filled in (author/owner = `Rogue-strider`,
module path = `github.com/Rogue-strider/sprocket`). To publish:

```bash
git init
git add -A
git commit -m "chore(release): initial sprocket v0.2.0"
git branch -M main
git remote add origin https://github.com/Rogue-strider/sprocket.git
git push -u origin main
```

Once pushed, anyone (including you, from any machine) can install it
without the local path:
```
/plugin marketplace add Rogue-strider/sprocket
/plugin install sprocket@sprocket-marketplace
```

CI (`.github/workflows/validate.yml`) runs automatically on the first push
— check the Actions tab for the `go vet` + `go test` + `go build` run and
the cross-compile matrix job.
