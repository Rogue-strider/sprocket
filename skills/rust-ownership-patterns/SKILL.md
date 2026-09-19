---
name: rust-ownership-patterns
description: Reference for Rust ownership, borrowing, and lifetime patterns, and when reaching for Rc/Arc/Clone is (and isn't) justified. Load when writing or reviewing Rust code.
---

# Rust Ownership Patterns Reference

## Before reaching for `.clone()`
Ask: can this be restructured so one owner holds the data and others borrow it? Common fixes instead of cloning:
- Pass `&T` instead of `T` if the callee only reads.
- Split a struct so the field that needs independent ownership is separate from the rest.
- Use an index/id into a collection instead of holding a reference or clone of the element itself.

## When `Rc`/`Arc` is the right call
Genuine shared ownership with unclear lifetime (e.g. a graph/tree with shared child nodes, or cross-thread shared state with `Arc<Mutex<T>>`/`Arc<RwLock<T>>`). Not a substitute for fixing a borrow-checker error you don't understand yet.

## `unsafe`
- Every `unsafe` block needs a `// SAFETY:` comment explaining exactly which invariant makes it sound.
- Keep the unsafe block as small as possible — wrap it in a safe function with a checked precondition rather than leaving `unsafe` sprawling through a function.

## Error handling
- Library code: return `Result<T, E>`, never `unwrap()`/`expect()` on inputs you don't control.
- Application code (`main`, CLI entry points): `unwrap()`/`expect()` with a descriptive message is fine at the boundary where you'd exit anyway.
- Prefer `thiserror` for library error enums, `anyhow` for application-level error propagation.

## Iterators
- Prefer iterator adapter chains (`.filter().map().collect()`) over manual index loops when it doesn't hurt readability — but don't force a chain that becomes harder to read than a `for` loop.
