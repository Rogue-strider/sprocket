---
description: Mine recent git history for recurring patterns and propose new/updated skills
---

Run `${CLAUDE_PLUGIN_ROOT}/bin/sprocket-learn.exe --repo . --days ${ARGUMENTS:-90}` against this repository's git history and show me the proposed new or updated skills it finds. For each proposal, explain the pattern it detected and which existing skill (if any) it would extend, then ask me before writing any file to `${CLAUDE_PLUGIN_ROOT}/skills/`.
