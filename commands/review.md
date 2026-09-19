---
description: Run the full review pass (architecture, security, and language-specific reviewer) over the current changes
---

Review the current uncommitted changes (`git diff`, or the files/paths given below if specified):

1. Run architecture-reviewer if the change touches design/schema/API contracts.
2. Run security-reviewer on any change touching auth, user input, secrets, or payments.
3. Run the matching language reviewer (go-code-reviewer for .go files, rust-code-reviewer for .rs files, web3-contract-reviewer for .sol/Anchor files) for every touched file type.
4. Run aws-infra-reviewer for changes touching .tf/.tf.vars files, CDK/CloudFormation templates, Dockerfiles, docker-compose.yml, or deployment steps in .github/workflows/*.yml.

Consolidate all findings into one report, ordered by severity, with duplicate/overlapping findings merged rather than repeated per-agent.

Target: $ARGUMENTS
