---
name: aws-infra-checklist
description: Reference checklist for AWS infrastructure, IAM, and deployment safety. Load when writing or reviewing Terraform/CDK, Dockerfiles, or CI/CD deployment configs.
---

# AWS Infra Checklist

## IAM
- Least privilege: name specific actions and resource ARNs, not `*`/`*`. Start narrow and add permissions when something actually fails, not the reverse.
- No access keys in code, `.env` files, or CI logs. Use IAM roles (EC2 instance profiles, ECS task roles, Lambda execution roles, OIDC for GitHub Actions) instead of long-lived keys wherever the platform supports it.
- Cross-account roles: trust policy names specific account IDs and, where possible, an external ID — never a wildcard principal.

## Networking
- Security groups: ingress from `0.0.0.0/0` only on a public load balancer's 80/443. Databases, caches, and internal services get security-group-to-security-group rules, not open CIDRs.
- Private subnets for anything that doesn't need a public IP (databases, internal APIs); NAT gateway for their outbound internet access instead of a public subnet.
- VPC flow logs enabled on anything handling sensitive traffic, for incident investigation after the fact.

## Secrets
- Secrets Manager or Parameter Store (SecureString) for credentials, API keys, database passwords — not plain environment variables passed through Terraform variables (those end up in state files in plaintext).
- Terraform/CDK state itself: treat it as sensitive (it often contains secrets in plaintext) — remote backend with encryption at rest, restricted IAM access to the state bucket.

## State management
- Remote backend with locking: S3 + DynamoDB for Terraform, or Terraform Cloud/CDK's built-in state handling. Local state is only acceptable for solo experimentation, never for anything a team touches or that's deployed to a real environment.
- Before applying: always review the plan output for `# forces replacement` — a resource being destroyed and recreated (not updated in place) is often an unintended downtime source, especially for stateful resources like RDS or EBS-backed instances.

## Containers
- No secrets in Docker image layers — check every `ARG`, `ENV`, and `COPY` for credential files; a secret in an intermediate layer persists even if a later layer removes the file.
- Run as non-root where the base image and application allow it.
- Pin base image versions (not `:latest`) for reproducible builds.

## CI/CD deployment safety
- Secrets come from the CI provider's encrypted secret store, injected at runtime — never committed to the workflow file itself.
- Production deploys gated behind either a manual approval step or a required passing test/build stage — never triggered unconditionally on every push to main.
- Rollback path identified before deploying: previous image tag/artifact to redeploy, and whether any accompanying migration is reversible.

## Cost
- Right-size instance types to the actual workload; flag anything that looks like a default/oversized choice with no stated reason.
- Lifecycle policies on S3 buckets and CloudWatch log groups that accumulate data indefinitely (logs, backups) — unbounded storage growth is a common silent cost leak.
- Auto-scaling configured for anything with variable load, rather than a fixed fleet sized for peak.
