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
| 2    | Operational | SMOKE itself broke: runner crash, I/O error.            |
| 3    | `NEW`       | No lock file; first run is unreviewed.                  |
| 64   | Usage       | Invalid invocation — bad flags or argv (`EX_USAGE`).    |
| 65   | Data        | A spec or lock file was read but is malformed (`EX_DATAERR`). |

`1` (`CHANGED`) also covers:

- **`MISSING`** — a lock entry has no corresponding result this run (deleted
  command or check). It is a form of drift; the report enumerates what is
  missing.
- **Timeout** — a command that fails to produce expected output within its
  deadline is misbehaving observably, which is drift, not tool trouble.

`65` (`EX_DATAERR`) covers every malformed-input failure — a spec or lock file
that SMOKE *read* but could not parse or validate: YAML/JSON syntax errors,
structural-validation failures (a leaf with no command, an unresolvable check
name), a bad timeout duration, or a CUE constraint/unification failure. The line
against `2` is whether SMOKE got usable bytes: *read but malformed* → `65`;
*could not read, or broke mid-run* (missing file, I/O error, runner crash) →
`2`. A missing or unreadable file stays `2` — there is no `EX_NOINPUT` (66) in
the scheme; minting one would split file failures for no consumer benefit.

## Streams

- Drift / match report → **stdout**.
- Operational (`2`), usage (`64`), and data-error (`65`) diagnostics → **stderr**.

Consumers can separate "output drifted" from "SMOKE broke" by exit code and,
redundantly, by stream.

## Consumer guidance

- **CI gate:** "fail on any nonzero" still works — `0` is the only clean pass.
  Branch on the specific code only if you want to treat `NEW`/`CHANGED`
  differently from a crash.
- **Agent in a TDD loop:** `1` (`CHANGED`) is the *expected* state during an
  intentional change — eyeball and `--commit`, do not "fix" output back to
  green. `3` (`NEW`) means review and commit the first lock. `2`/`64`/`65` mean
  stop — `65` specifically means fix the malformed spec or lock file.

## Implementation status

The full contract is implemented. The six codes are wired as `internal/p`
constants (compare-mode emits `0`/`1`/`3`, `p.Exit` emits `2`, `p.DataErr` emits
`65`, usage paths emit `64`). Spec and lock-file parse/validation failures route
through `p.DataErr` (`65`); only `os.Open` failures and mid-run trouble keep
`p.Exit` (`2`). A timed-out command is recorded as a synthetic `timeout` check
(`checks.Timeout`) and compares as drift (`1`), not a tool error. Operational,
usage, and data-error diagnostics route to stderr via `p.Error`; the drift/match
report stays on stdout. Bad-flag parse errors exit `64`: `pflag.CommandLine`
runs in `ContinueOnError` mode and `main` handles the error rather than letting
pflag exit `2`, so no usage path collides with operational trouble.

One surface from the broader exit-code epic remains open and is tracked
separately: the `--json` `status` field mirroring the code. The contract table
is documented on both user-facing surfaces (`--help` and README).
