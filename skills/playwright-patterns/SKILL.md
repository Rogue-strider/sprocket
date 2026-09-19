---
name: playwright-patterns
description: Reference for writing stable, non-flaky Playwright end-to-end tests. Load when writing or debugging Playwright tests.
---

# Playwright Patterns Reference

## Locators (in priority order)
1. `getByRole` — matches how assistive tech and users identify elements
2. `getByLabel` / `getByPlaceholder` — for form fields
3. `getByText` — for static content
4. `data-testid` — last resort, only when no accessible role/label exists

## Avoiding flakiness
- Never use `waitForTimeout` as a fix for a race condition — it hides the bug and is still flaky at different speeds.
- Use web-first assertions (`await expect(locator).toBeVisible()`) which auto-retry, instead of a manual assertion after a fixed wait.
- Isolate test state: each test creates its own data (via API setup, not UI clicking) rather than depending on state left by a previous test.
- For flows involving network requests, use `page.waitForResponse()` keyed to the specific request rather than waiting for a generic loading spinner to disappear.

## Structure
- One user-facing behavior per test; multiple loosely related assertions in one test make failures hard to diagnose.
- Use fixtures for authenticated state (`storageState`) instead of logging in through the UI in every test.
- Run tests in parallel by default; anything that can't run in parallel (shared external resource) should be explicitly marked serial, not accidentally relied upon to run in order.
