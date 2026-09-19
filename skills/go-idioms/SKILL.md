---
name: go-idioms
description: Reference for idiomatic Go patterns — error handling, concurrency, interface design. Load when writing or reviewing Go code.
---

# Go Idioms Reference

## Errors
- Wrap with context: `fmt.Errorf("fetching user %d: %w", id, err)` — never just `return err` across a package boundary if the caller needs to know where it came from.
- Sentinel errors: `errors.Is`/`errors.As`, not string comparison on `err.Error()`.
- Don't log AND return an error — pick one; logging-and-returning causes duplicate log lines up the call stack.

## Concurrency
- Every goroutine needs a way to stop: pass `context.Context`, select on `ctx.Done()`.
- Prefer channels for ownership transfer, mutexes for protecting shared state — mixing both for the same data is a smell.
- `go.mod` minimum Go version 1.22+: loop variables are now per-iteration, so the classic closure-capture bug is gone — don't add `i := i` workarounds for it in new code.

## Interfaces
- Define interfaces at the consumer, not the producer ("accept interfaces, return structs").
- A one-method interface named after the method (`io.Reader`) is usually right; a five-method interface guessing at future needs usually isn't.

## Project layout
- `cmd/` for binaries, `internal/` for code not meant to be imported externally, flat package structure over deep nesting until it hurts.
