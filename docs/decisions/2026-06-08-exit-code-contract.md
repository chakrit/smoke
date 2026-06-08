# Exit-Code Contract

- **Date:** 2026-06-08
- **PR:** manual
- **Status:** accepted

## Decision

SMOKE exits with one of five stable, documented codes. Each names a distinct
outcome class; no code is shared across classes.

| Code | State       | Trigger                                                  |
| ---- | ----------- | ------------------------------------------------------- |
| 0    | `UNCHANGED` | Observable output matched the lock. *Not* "tests passed."|
| 1    | `CHANGED`   | Drift detected. Includes `MISSING` and command timeout. |
| 2    | Operational | Tool itself failed: bad spec, runner crash, I/O error.  |
| 3    | `NEW`       | No lock file / unreviewed first run.                    |
| 64   | Usage       | Invalid invocation (`EX_USAGE`, sysexits.h).            |

Operational (`2`) and usage (`64`) diagnostics write to **stderr**. The
drift/match report writes to **stdout**. A consumer can therefore separate
"the command's output drifted" from "SMOKE broke" by both exit code *and*
stream.

Codes are a contract: once shipped they are frozen. New outcome classes get a
new code, never a re-used one.

## Rationale

SMOKE is conceptually a `diff` between observed output and a committed golden,
so the numeric scheme anchors on `diff(1)`'s long-established triad — `0` =
same, `1` = differs, `2` = trouble — not on test-runner `0`/`1` pass/fail. That
choice is the whole point: the codes must describe *drift*, not *correctness*.

**Why `64` for usage, not pflag's default `2`.** pflag exits `2` on a bad flag,
and `diff(1)` uses `2` for "trouble" (operational failure). Both cannot own `2`.
We give `2` to operational trouble because that is the load-bearing distinction
for anything wrapping SMOKE — a CI gate or an agent must tell a regression from
a crash, and the `diff` mental model is the one users already carry. Usage
errors are the rarer case and move to `EX_USAGE` (64) from `sysexits.h`. A
future reader seeing `64` should not assume we forgot pflag's convention — we
deliberately reconciled the collision in favor of the diff model.

**Why timeout is drift (`1`), not trouble (`2`).** A command that fails to
produce its expected output within the deadline is the *command* misbehaving —
observable behavior, which is exactly what SMOKE exists to observe. If the
golden captured a prompt response and the command now hangs, that is drift, the
same as if it had printed different bytes. Classifying timeout as operational
trouble would conflate "the thing under test changed" with "the test harness
failed," which is the exact overloading this contract removes. (Implementation
note: today a timeout aborts the whole run via the runner-error path; honoring
this ruling is a behavior change owned by the exit-code *implementation* task,
not this decision.)

**Why `NEW` gets its own code (`3`), not folded into `2`.** The no-lock,
first-run state is semantically neither "matched," "drifted," nor "the tool
broke" — it means "a golden does not exist yet; a human or LLM must eyeball the
output and commit it before it can be trusted." An agent driving a TDD loop has
to distinguish "go review and `--commit` the first lock" from "SMOKE crashed,
stop." Folding `NEW` into operational `2` re-creates the overloading the whole
exit-code epic is killing. (Today this path hard-errors as exit `1`; the
implementation task changes it to run, report `NEW`, and exit `3`.)

**Why `MISSING` folds *into* drift (`1`), not its own code.** A lock entry with
no corresponding result in the current run (a deleted command, a removed check)
is a form of output drift — the observable surface moved. It needs review and
re-commit exactly like a changed line. Minting a separate code would split one
conceptual state across two numbers for no consumer benefit; the code space
stays small and the `CHANGED` report already enumerates *what* is missing.

**Why errors move to stderr.** Today every diagnostic flows through
`p.Error` → `output()` → stdout; only `p.Usage` reaches stderr. An agent cannot
then separate drift output from tool-trouble output by stream. Routing
operational and usage diagnostics to stderr is the Unix norm and gives consumers
a second, redundant separation axis alongside the exit code.

## Scope

This decision owns the **numeric** contract. The human-readable vocabulary
(`UNCHANGED` / `CHANGED` / `NEW` / `MISSING`, and the ban on "pass"/"green"/✓
language) is co-designed with the "Disambiguate semantics for LLM consumers"
epic — same ruling, two faces: this doc fixes the codes, that one fixes the
strings and the `--json` `status` field that mirrors them. They must stay
consistent; changing one without the other reopens the collision.

Out of scope here (owned by the exit-code *implementation* epic): centralizing
codes as named constants, mirroring the code in `--json`, the timeout and
no-lock behavior changes, and documenting the table in `--help` / README.
