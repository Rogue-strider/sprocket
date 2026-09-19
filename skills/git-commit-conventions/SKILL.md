---
name: git-commit-conventions
description: Conventions for writing commit messages that are useful input to the sprocket learning loop (which mines git history for patterns). Load when committing changes.
---

# Git Commit Conventions

Use Conventional Commits format: `type(scope): summary`

Types: `feat`, `fix`, `refactor`, `chore`, `test`, `docs`, `perf`, `security`

Rules:
- Summary in imperative mood, under 72 chars, no trailing period.
- Scope is the affected module/package, lowercase, kebab-case.
- Body (optional, blank line after summary): explain *why*, not what — the diff already shows what changed.
- Reference the fix's root cause for `fix:` commits specifically (e.g. `fix(auth): validate JWT expiry before role check`) — specific fix commit messages are what the learning loop uses to detect recurring bug classes and turn them into review-agent checklist items.

This convention matters more than usual here: `scripts/learn-from-git.py` parses commit type + scope + summary to cluster recurring patterns (e.g. five `fix(auth):` commits about the same root cause become a proposed addition to `security-reviewer`'s checklist).
