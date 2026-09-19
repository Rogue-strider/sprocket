---
name: release-checklist
description: Checklist to run through before cutting a release/deploy. Load when preparing a deploy to Vercel/Railway or tagging a release.
---

# Release Checklist

- All CI checks green on the commit being deployed (build, lint, tests) — verify, don't assume from a stale badge.
- Environment variables for the target environment (staging/production) confirmed present via the platform's dashboard/CLI, not just assumed to match `.env.example`.
- Database migrations (if any) are backward-compatible with the currently-running version — never deploy a migration that drops/renames a column the old code still reads, in one step.
- Rollback plan identified before deploying: previous deployment/tag to revert to, and whether any migration in this release is reversible.
- For Vercel: preview deployment checked against the actual PR before promoting to production.
- For Railway: check the target environment's resource limits if this release changes memory/CPU-heavy code paths.
- Tag the release in git matching semantic versioning if the project uses it, so `learn-from-git.py` and future agents can correlate incidents with releases.
