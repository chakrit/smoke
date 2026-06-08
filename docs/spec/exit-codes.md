# Exit Codes

- **Status:** draft

> The numeric scheme and its rationale were ruled in
> [`../decisions/2026-06-08-exit-code-contract.md`](../decisions/2026-06-08-exit-code-contract.md).
> This spec is the live contract surface; that decision is the frozen *why*.
> `draft` until the implementation epic lands — today's code still overloads
> `1` across drift, operational, and usage failures.

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

The five codes are wired (`internal/p` constants; compare-mode emits
`0`/`1`/`3`, `p.Exit` emits `2`, usage paths emit `64`). Two parts of this spec
are **not yet implemented** and remain the target, not the present behavior:

- **Timeout** currently routes through the runner-error path and exits `2`
  (operational), not `1` (drift). Reclassifying it requires engine surgery —
  turning a timeout into a recordable check result instead of an aborting error
  — so it is deferred to its own slice.
- **Stream separation** is not done: operational/usage diagnostics still print
  to stdout (`p.Error` → stdout), not stderr. Exit codes already disambiguate;
  the stderr axis is the redundant second one and lands in a follow-up.
