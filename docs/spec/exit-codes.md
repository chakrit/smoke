# Exit Codes

- **Status:** accepted

> The numeric scheme and its rationale were ruled in
> [`../decisions/2026-06-08-exit-code-contract.md`](../decisions/2026-06-08-exit-code-contract.md).
> This spec is the live contract surface; that decision is the frozen *why*.

## Contract

SMOKE exits with exactly one of these codes. Each names a distinct outcome
class; no code is shared across classes, and shipped codes are frozen.

| Code | State       | Meaning                                                  |
| ---- | ----------- | ------------------------------------------------------- |
| 0    | `UNCHANGED` | Observable output matched the lock. *Not* "tests passed."|
| 1    | `CHANGED`   | Drift detected — output moved from the golden.          |
| 2    | Operational | SMOKE itself failed: bad spec, runner crash, I/O error. |
| 3    | `NEW`       | No lock file; first run is unreviewed.                  |
| 64   | Usage       | Invalid invocation (`EX_USAGE`).                         |

`1` (`CHANGED`) also covers:

- **`MISSING`** — a lock entry has no corresponding result this run (deleted
  command or check). It is a form of drift; the report enumerates what is
  missing.
- **Timeout** — a command that fails to produce expected output within its
  deadline is misbehaving observably, which is drift, not tool trouble.

## Streams

- Drift / match report → **stdout**.
- Operational (`2`) and usage (`64`) diagnostics → **stderr**.

Consumers can separate "output drifted" from "SMOKE broke" by exit code and,
redundantly, by stream.

## Consumer guidance

- **CI gate:** "fail on any nonzero" still works — `0` is the only clean pass.
  Branch on the specific code only if you want to treat `NEW`/`CHANGED`
  differently from a crash.
- **Agent in a TDD loop:** `1` (`CHANGED`) is the *expected* state during an
  intentional change — eyeball and `--commit`, do not "fix" output back to
  green. `3` (`NEW`) means review and commit the first lock. `2`/`64` mean stop.

## Implementation status

The full contract is implemented. The five codes are wired as `internal/p`
constants (compare-mode emits `0`/`1`/`3`, `p.Exit` emits `2`, usage paths emit
`64`). A timed-out command is recorded as a synthetic `timeout` check
(`checks.Timeout`) and compares as drift (`1`), not a tool error. Operational
and usage diagnostics route to stderr via `p.Error`; the drift/match report
stays on stdout.

Two surfaces from the broader exit-code epic remain open and are tracked
separately: the `--json` `status` field mirroring the code, and documenting the
table in `--help` / README.
