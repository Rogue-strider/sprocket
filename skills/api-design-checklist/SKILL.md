---
name: api-design-checklist
description: Checklist for designing REST/API contracts that are stable and won't need breaking changes soon. Load when designing new endpoints or reviewing API contracts.
---

# API Design Checklist

- Version the API from the start (`/v1/...` or a header) even if you never expect to need v2 — retrofitting versioning after clients exist is far more expensive than adding it upfront.
- Use plural nouns for collections (`/users`, not `/user`), and nest only one level deep where possible (`/users/:id/orders`, not `/users/:id/orders/:orderId/items/:itemId/details`).
- Return consistent error shapes across every endpoint (`{ "error": { "code": ..., "message": ... } }`) rather than ad-hoc error formats per route.
- Pagination on any list endpoint from day one (cursor-based preferred over offset for anything that mutates frequently) — adding it later breaks existing clients.
- Idempotency keys on POST endpoints that create billable/side-effecting resources (payments, subscriptions) to make retries safe.
- Don't leak internal IDs/implementation details (database autoincrement IDs, internal enum values) in public responses if you might need to change the backing implementation later — use opaque IDs or stable string enums.
