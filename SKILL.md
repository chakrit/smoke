---
name: smoke
description: >-
  Golden-file / snapshot smoke testing of shell-command output with the SMOKE CLI.
  Capture a command's observable output once, lock it into a *.lock.yml, and let later
  runs go UNCHANGED when output matches or CHANGED when it drifts. TRIGGER when: setting
  up snapshot or golden-file testing for a CLI tool or shell command, capturing command
  output as a baseline/golden, detecting output drift across runs, writing or editing a
  tests.yml / *.cue smoke spec, or interpreting SMOKE's UNCHANGED / CHANGED / NEW verdicts
  and 0 / 1 / 2 / 3 / 64 / 65 exit codes. Critical framing other tools get wrong: SMOKE
  detects DRIFT, not correctness — CHANGED is the expected "eyeball and re-commit" state,
  NOT a failing test; UNCHANGED means drift-free, NOT verified-correct. An agent must never
  chase output back to green or auto-commit a CHANGED result it did not read. DO NOT TRIGGER
  for assertion-based unit/integration frameworks (go test, pytest, jest), or test work that
  is not snapshot/golden-file based.
---

# SMOKE — golden-file smoke testing

SMOKE captures a command's observable output once and flags when it later drifts. It is a
**drift detector, not an assertion engine**: it answers *"does the output still match the
committed golden?"*, never *"is the behavior correct?"*. Treat it as `diff` against a
locked baseline, not as a pass/fail test runner.

Not a replacement for real tests — a fast drift tripwire for CLI output.

## Setup

Scaffold a spec, then fill it in:

```sh
smoke --init           # writes tests.yml (or: smoke --init=mytests.yml)
```

A spec is a tree of tests. The YAML root *is* the root test; `tests:` nests arbitrarily,
and subtests inherit the parent's `config`, `checks`, and `commands`.

```yaml
config:
  interpreter: /bin/sh   # commands are piped to `interpreter -s` via stdin
  timeout: 5s
checks:                  # what to observe and lock per command
  - exitcode
  - stdout
  - stderr
  - generated/*.go       # file globs are also checks — locks file contents
tests:
  - name: Greeting
    commands:
      - echo hello world
```

`.cue` and `.json` /`.jsonl` specs work too; `.cue` gets schema validation (typo'd fields
fail closed). Lock files are always `<spec>.lock.yml`.

## The loop: run → eyeball → commit

This three-step cadence is the whole workflow. The eyeball step is load-bearing — skipping
it defeats the tool.

1. **Run** — `smoke tests.yml`. First run reports `NEW` (no lock yet).
2. **Eyeball** — read the output. Is it what the command *should* produce?
3. **Commit** — if correct, `smoke -c tests.yml` writes the `.lock.yml` golden.

Check the `.lock.yml` into source control so teammates and CI compare against the same
baseline.

On later runs: `UNCHANGED` (exit 0) means output matched. When you change behavior that
*should* move output, you get `CHANGED` (exit 1) — eyeball the diff, then `smoke -c` to
re-bless the golden. Re-running the run→eyeball→commit loop on an intended change is
correct, not a workaround.

## Exit codes

Branch on the exit code — never parse the human text or colors. Full contract:
[`docs/spec/exit-codes.md`](https://github.com/chakrit/smoke/blob/main/docs/spec/exit-codes.md).

| Code | State       | Meaning                                                        |
| ---- | ----------- | -------------------------------------------------------------- |
| 0    | `UNCHANGED` | Output matched the lock. *Not* "tests passed."                 |
| 1    | `CHANGED`   | Drift — output moved (includes `MISSING`, timeout). Review it. |
| 2    | —           | Operational error: SMOKE itself broke (runner crash, I/O).     |
| 3    | `NEW`       | No lock yet; first run is unreviewed.                          |
| 64   | —           | Usage error: bad flags or argv.                                |
| 65   | —           | Data error: a spec or lock file is malformed.                  |

`0` is the only clean pass for a CI gate; "fail on any nonzero" works. `2` /`64`/`65` mean
*stop* — none is drift.

## The three traps (the framing every test-trained consumer gets wrong)

Full guidance: [`docs/spec/using-smoke-in-tdd.md`](https://github.com/chakrit/smoke/blob/main/docs/spec/using-smoke-in-tdd.md).

- **`CHANGED` is not a failing test.** Exit 1 during an intentional change is *expected* —
  eyeball, then re-commit. Never pattern-match it as a red test and edit code to chase the
  output back to `UNCHANGED`. Re-commit; don't revert.
- **`UNCHANGED` is not "correct".** Exit 0 means the output didn't move, not that the
  behavior is right — the change may be in an unobserved surface, or the golden may have
  locked in a bug. Treat it as "no drift to review," never as verification.
- **Re-committing can hide a regression.** `smoke -c` is right only when you *intended*
  the change and read the new output. Auto-committing every `CHANGED` blesses regressions
  as the new golden — that turns SMOKE off.

## For an agent driving SMOKE

- Branch on `$?`, not output text. `0` = no drift (continue, not a verification pass); `1`
  = drift (surface the diff, re-commit only if intended and checked); `3` = first run
  (review, commit the initial lock); `2` /`64`/`65` = stop.
- Use `--json` (compare mode only) for a machine-readable mirror: top-level `status`
  (`unchanged`/`changed`/`new`) + `exitCode`, plus a per-node `matched`
  /`changed`/`missing` tree. It is the only thing on stdout at default verbosity. Do not
  combine `--json` with `--list` /`--print`/`--commit`/`--show-expected` (exits 64).
- Never auto-chase green. A drift you did not cause is a signal to surface, not to fix.
