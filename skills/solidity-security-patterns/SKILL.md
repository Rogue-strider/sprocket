---
name: solidity-security-patterns
description: Reference for common Solidity/EVM smart contract vulnerabilities and mitigation patterns. Load when writing or reviewing smart contract code.
---

# Solidity Security Patterns Reference

## Reentrancy
- Follow checks-effects-interactions: validate, update state, then make external calls — never update state after an external call.
- Use `ReentrancyGuard` (OpenZeppelin) on any function that transfers value or calls an untrusted external contract.

## Access control
- Privileged functions (mint, pause, upgrade, withdraw-all) need explicit modifiers (`onlyOwner`, role-based via `AccessControl`), never rely on `tx.origin` for auth.
- Multi-sig or timelock for upgrade/admin actions on anything holding real value.

## Arithmetic
- Solidity >=0.8 has built-in overflow/underflow checks — don't wrap arithmetic in `unchecked {}` blocks without a specific, documented reason (usually just gas optimization in a proven-safe context).

## External calls
- Check return values of low-level `call`/`send` — a silently failed transfer is a common bug class.
- Assume any external contract call can reenter or behave adversarially, including "trusted" token contracts.

## Common gas/DoS pitfalls
- Unbounded loops over arrays that can grow from user input (e.g. iterating "all depositors" to pay out) — use pull-payment patterns instead of push.

## Testing expectations
- Any contract handling value transfer should have tests covering: normal flow, reentrancy attempt, access-control bypass attempt, and boundary values (zero, max uint).
