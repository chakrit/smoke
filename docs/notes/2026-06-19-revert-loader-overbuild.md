# Session: revert the loader/identity over-build

- **Date:** 2026-06-19
- **Branch:** `main` (committed; push per policy)
- **Safety tag:** `pre-revert-99df6e0` marks the state just before this revert.

## Why

Poking the loader/identity corner kept turning up wrong-feeling things, and they were all
symptoms of over-abstraction added across the 2026-06-18 sessions:

- **`engine.TestID` was a hollow abstraction.** No `ID` field anywhere — just
  `engine.TestID(name)` cast at three call sites. Its own comment promised "centralized so
  the derivation can grow richer without touching call sites," but the design made that
  impossible: the lock persists only `name` and `ID()` casts it back, hard-baking `id ==
  name` forever. Identity was conflated with display name and recomputed, not assigned.
- **The `test_ir.go` IR didn't earn its keep.** `parse` → `[]testIR` → `validate` was a
  single-consumer, immediately-discarded representation — a function split in two with three
  new types (`testIR`, `parsed[T]`, `buildableTest`, `leafError`) to implement all-errors
  reporting that a plain error-accumulating tree walk does in ~30 lines.
- **Partial commit / `--commit-last`** rode on that identity work and added a `runcache`
  package, lock-merge semantics, and a "use a full commit to add a test" sharp edge.

## What was ripped out (back to baseline `83e8256`, post-CUE/JSON/JSONL loader)

- `engine.TestID` type; `resultspecs` compares by `Name` again, no `ID()`, no `Merge`.
- `testspecs/test_ir.go` deleted; `Tests()` is the direct depth-first walk (first-error).
- Partial commit: a filtered `--commit` is **refused** again (exit 64), not merged.
- `--commit-last` and the entire `runcache` package.
- Duplicate test names are no longer rejected at load.

## What was kept

- **CUE / JSON / JSONL / JSONC loaders** — the genuinely useful work.
- **The www docs site** — untouched.
- **This session's multi-spec exit aggregation** — `main` as the single exit authority,
  `status.Merge`, `outcome.go`, JSONL `--json`. It's independent of the ripped-out code;
  `process.go` just dropped the partial-commit/commit-last branches.

## Settled design rulings (carry forward)

- **Compare is order-sensitive, correctly.** Tests aren't isolated — order is load-bearing
  (setups/teardowns are modelled as sibling ordering), so a reorder *must* read as drift
  even when per-test output is unchanged. `compareTests` stays the `gendiff` LCS, keyed on
  `Name`. (An identity-keyed/set diff was considered and rejected — it would hide reorders.)
- **If identity is ever reintroduced, it's a carried field, not a getter** — assigned once
  where the flattened name is built, threaded to the result. That's the prerequisite for any
  future merge/partial-commit work. See `TODO.md`.

## Verification

`go build`/`vet`/`test` and `./test.sh` all green. Behaviors confirmed: dup names → NEW
(not 65), `--commit-last` → 64 (unknown flag), filtered `--commit` → 64 (refused),
`test_ir.go` + `runcache/` gone. Docs (guide, architecture, run-cache decision, TODO) and
the docs site updated to match.
