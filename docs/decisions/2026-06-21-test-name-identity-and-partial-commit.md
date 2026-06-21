# Test Identity (`TestName`) and Partial Commit

- **Date:** 2026-06-21
- **PR:** manual
- **Status:** accepted

## Decision

1. **Identity is the display name, modelled as a `TestName` type** — value-equal to the
   flattened name, not a separate id. Carried as a field `engine.Test` → `TestResult` →
   `resultspecs.TestResultSpec`, **minted once** in the flatten walk (`testspecs.Tests()`),
   **received and used — never re-minted — everywhere else.** No `TestName(name)` casts at
   call sites.
2. **Duplicate flattened names are a load error → exit `65`** (`EX_DATAERR`, the
   malformed-spec class), naming both colliding tests by path. Enforced *by construction*:
   the flatten walk indexes tests into `map[TestName]…`; a colliding insert **is** the
   error. No warning, no skip-the-loser, no first-win.
3. **`TestName` stays a lightweight `type TestName string`** — not an opaque struct. It
   serializes as the existing lock `name` key; no schema change, no separate `id` field.
4. **Partial commit merges via a spec-order walk.** A filtered `--commit` is no longer
   refused; it rewrites the lock by walking the *spec* in order, taking the fresh result for
   tests in the run's filter and carrying forward the existing lock entry (by `TestName`) for
   the rest.
5. **Deferred:** `--commit-last` / `runcache` (own pass, vNext); the spec-filename
   path-dependence of the root name (logged as a bug).

## Rationale

**Why a type at all — didn't we just revert `TestID`?** The reverted `engine.TestID` was a
cast (`TestID(name)`) at three call sites with no invariant — mintable from any string,
including a duplicate, so it bought nothing. `TestName` is the opposite: its value only ever
originates at the flatten gate, where uniqueness is already enforced. The load-bearing rule
is **minted only at the gate, received everywhere else** — that, not the type itself, is what
keeps it from decaying back into the hollow cast. A future reader must not "simplify" it into
a getter or a call-site cast.

**Why `TestName`, not `TestID`.** Identity does *not* diverge from the display name and we
chose not to invent a reason for it to. `TestID` would imply a stable-across-rename or
structural key we don't have; `TestName` is honest about what the value is — the display name
carrying a uniqueness invariant.

**Why hard-fail on duplicates, not warn-and-first-win.** A spec with two identically-named
tests is ambiguous *at the source of truth*: a lock entry named `build` cannot be resolved to
one of two tests — same class as duplicate keys in a YAML map. Failing closed is correct for
an ill-formed spec, and it is strictly simpler than first-win: no skip-the-loser path, no
warning emission, no "did it run?" ambiguity. The earlier instinct toward first-win was really
an argument for *surfacing* the collision; a hard `65` surfaces it maximally, and the
uniqueness it guarantees is exactly the merge key partial-commit needs. (Silent first-win was
rejected outright: for a drift detector, silently swallowing a whole test is the worst
outcome — coverage you think you have but don't.)

**Why not an opaque type.** Tempting, but it does not buy the safety it appears to. In Go a
`type TestName string` is freely convertible cross-package, so a plain newtype's protection is
convention only; a true opaque struct blocks conversion but still needs an *exported*
constructor (the type lives in `engine`, the gate lives in `testspecs`), which is itself a
public minting path. Neither form enforces *uniqueness* — that is a set property only the
gate's map-insert can see. So opacity is ergonomics, not the guarantee, and it costs custom
YAML `Marshal`/`Unmarshal` plus a `String()` accessor threaded through every read site, and
possibly an import-graph inversion to confine minting to the gate. Deferred as an additive
hardening *if* stray re-minting ever shows up; it changes nothing about the guarantee.

**Why the merge walks the spec, not the lock or the filter.** Walking the spec gives spec
order for free (compare is order-sensitive, so new tests must land in position, not appended),
and a test deleted from the spec is simply never visited → drops from the rewritten lock,
matching full-commit semantics with no special case. Carry-forward applies only to in-spec,
not-in-this-filter tests that already have a lock entry; a never-committed test has nothing to
carry and stays `NEW`. Read-path compare is unchanged: a lock entry with no matching spec test
still reports `MISSING` (drift, exit `1`) exactly as today.
