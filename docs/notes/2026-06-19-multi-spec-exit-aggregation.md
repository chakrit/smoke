# Session: multi-spec exit aggregation (the "unsound exit code" fix)

- **Date:** 2026-06-19
- **Branch:** `main` — committed **and pushed** (`gh/main` synced).
- **Trigger:** peer agent `chakrit.lowfat-pantry.claude` reported (via ace-connect) that
  `smoke <many specs>` exited 0 while silently skipping all but the first spec.

## The bug

Compare mode (the default) called `os.Exit` inside `compareResults`, so the
`for filename := range filenames` loop in `main` exited after the **first** spec.
`smoke A B C` only ever processed A — drift (or a dup-name 65) in specs 2..N was
silently skipped, order-dependent. The exit-code "contract" had only ever been honored
for a single spec. The user called this out as **unsound** — it should have been pinned
down when the codes were first frozen.

## The fix (commits `abc1683`, `f4430dd`)

Refactored to make **`main` the single exit authority** (this was the user's explicit
design directive — "proper abstraction, not minimal edit; we already model results
properly in engine/"):

- `processFile(name) (status, error)` — mirrors `engine.Runner.Test`'s `(result, error)`
  shape. Every helper in `process.go` returns errors instead of calling
  `os.Exit`/`p.DataErr`.
- Fatals are typed (`outcome.go`): `dataError` → 65, anything else → 2; `reported` marks a
  run error already surfaced live by the hooks so `main` doesn't double-print it.
- `status.Merge` (reporter.go) folds verdicts: **UNCHANGED is the identity**, so a clean
  spec can never clear an earlier drift (anti-masking — the whole point); among non-clean
  specs the last one's code wins (user wanted simple/last-write-win; UNCHANGED-as-identity
  is the one non-literal bit, to avoid re-masking).
- `main` loop: fold verdict, **fail-fast** on a fatal (65/2) because specs run in order and
  carry side effects the next depends on (setups/teardowns modelled as ordering — same
  load-bearing-order principle as within a spec). Drift never fail-fasts.

## Reporting (user: "report results for separate files separately")

Per-spec results report at the site, not merged. `--json` is now **one compact object per
spec (JSONL)** — single spec unchanged, multi-spec a stream. (Switched `MarshalIndent` →
`Marshal`; regenerated the JSON Output self-test locks.)

## Verification

- `status.Merge` unit test (anti-masking property) — `reporter_test.go`.
- Self-test `Behavior \ Multi-spec`: all-clean → 0, **drift-in-2nd → 1** (the regression
  guard; was 0 before the fix), malformed-2nd → fail-fast 65.
- `go test ./...`, `go vet`, `./test.sh` all green. Single-spec behavior unchanged.

## Docs

`exit-codes.md` §"Multiple specs"; decision `2026-06-08-exit-code-contract.md` amended
(2026-06-19: "contract is per-run, not per-spec"); `architecture.md` notes the single exit
authority; guide CI section shows `smoke tests/*.yml`.

## Peer thread (ace-connect)

Replied to `chakrit.lowfat-pantry.claude` with the root cause + repro (they reproduced
independently — their `scripts/test.sh` had been validating only spec #1 of 57). They
patched their side to loop per-spec; agreed worst-severity/aggregation is right. Detail
written to `/tmp/smoke-dupname-findings-chakrit.smoke.claude.md`.

## Carryover

- **v0.4.0** still deferred ("not yet"). This fix is a behaviour change (multi-spec exit),
  strengthening the minor-bump case.
- All AFK-run blockers from `.afk.log` were resolved live by the user (push ✓, deploy ✓,
  order-is-load-bearing ✓ recorded, ace-school amendment dropped per user).
