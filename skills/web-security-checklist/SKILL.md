---
name: web-security-checklist
description: Checklist for common web application vulnerabilities (auth, injection, IDOR, CSRF). Load when reviewing security-sensitive code or endpoints.
---

# Web Security Checklist

## Authorization
- Every endpoint that reads/writes a resource by ID must check the requesting user actually owns/can-access that resource (IDOR) — object existence checks alone are not authorization checks.
- Role checks happen server-side; a hidden UI button is not access control.

## Injection
- Parameterized queries only — string-concatenated SQL/NoSQL queries are a finding regardless of "it's escaped."
- Shell commands built from user input: use argument arrays, never string interpolation into a shell string.

## Auth
- Passwords hashed with bcrypt/argon2/scrypt, never sha256/md5 alone.
- JWTs: signature verified server-side on every request, algorithm pinned (reject `alg: none`), expiry enforced.
- Session tokens: `HttpOnly`, `Secure`, `SameSite` cookies for browser sessions.

## CSRF/CORS
- State-changing endpoints (POST/PUT/DELETE) need CSRF protection if using cookie-based auth.
- `Access-Control-Allow-Origin: *` combined with credentialed requests is a real finding, not a nitpick.

## Secrets
- No API keys/tokens in client-side code, committed `.env` files, or logs.
- Rotate anything that was ever committed to git, even if the commit was later removed — assume it's compromised.
