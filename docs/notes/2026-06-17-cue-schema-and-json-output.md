# Session log — CUE schema (Slice D) + `--json` output

Point-in-time breadcrumb. Authoritative task state lives in `TODO.md`. This
session landed two slices; both committed on `main` (now 13 ahead of `gh/main`,
unpushed — push awaits the user's say-so).

## Done

- **Slice D — CUE schema validation shipped** (`9616fbb`). Closes the
  First-class CUE epic (every item `[x]` or consciously dropped).
  - `testspecs/schema.cue` (new, `//go:embed`'d) — **closed** `#Test`/`#Config`
    definitions. Closedness is recursive, so typos nested under `tests`/`config`
    are caught too. `cue fmt`-canonical.
  - `testspecs/loaders.go` — `cueLoader.Load` now: compile schema →
    `LookupPath("#Test")` → `value.Unify(testDef)` →
    `unified.Validate(cue.Concrete(false))` → `Decode`. Previously `Decode`
    silently dropped unknown fields (same gap as `json.Decode`); now a typo'd or
    wrong-typed field fails closed as a clean CUE constraint error
    (`chekcs: field not allowed`) → exit 65.
  - Tests: `testspecs/testspecs_test.go` gained unknown-field-rejected (asserts
    the error names the field) + valid-`.cue`-loads. Smoke self-test
    `test/badcuetests.cue` + `tests.yml \ Tests \ States \ Bad CUE` → exit 65.
  - Docs: `architecture.md` CUE-seam section refreshed (loader abstraction +
    shipped validation; replaced the stale `decodeCUE` / "not yet present"
    callout, which was also mislabeled Slice C).

- **`--json` machine-readable compare output shipped** (`2ff034a`). Closes the
  last open surface of the exit-code epic and the semantics epic's `--json` item.
  - `reporter.go` (new) — minimal reporter abstraction (user-requested): a
    `status` enum (`statusUnchanged`/`Changed`/`New`) owning `String()` +
    `ExitCode()`, a `reporter` interface, and `consoleReporter` (the human
    edit-walk extracted **verbatim** from `compareResults`).
  - `report_json.go` (new) — `jsonReporter` + DTOs (`jsonReport`/`jsonTest`/
    `jsonCommand`/`jsonCheck`) + `jsonStatus(Action)`. Struct/slice-only →
    deterministic field order → stable lock.
  - `process.go` — `compareResults` slimmed to: select reporter by `shouldJSON`
    → `report(status, edits)` closure → `os.Exit(status.ExitCode())`. NEW
    (no-lock) path routes through the same seam.
  - `main.go` — `--json` flag; usage guard (exit 64) if combined with
    `--list`/`--print`/`--commit`/`--show-expected` (fail-closed, no ambiguous
    mode mixing).
  - Tests: `report_json_test.go` (status enum, action→status mapping, reporter
    shape to a buffer); self-tests `tests.yml \ Tests \ JSON Output`
    (unchanged/changed/new + bad-combo→64).
  - Docs: `exit-codes.md` + `architecture.md` updated.

## Contract decisions worth remembering (locked by self-tests)

- **JSON status vocab.** Top-level `status`+`exitCode`: `unchanged`/`changed`/
  `new` ↔ `0`/`1`/`3`. Per-node `matched`/`changed`/`missing`: `Removed`→
  `missing`, `Added`/`InnerChanges`→`changed`, `Equal`→`matched`. `new` is
  whole-run-only (no lock), never per-node.
- **Detail localizes to drift.** A `matched` node carries **no children** —
  `gendiff`'s `Compare` collapses equal subtrees (doesn't enumerate their
  descendants). So "full enumeration" is really "enumerate where it drifts." A
  fully-matched test shows `commands: []`. This is the right shape (matched needs
  no localization) but was a mid-slice correction — the first jsonReporter doc
  comment overclaimed "enumerates every node."
- **`--json` is compare-only** and **clean on stdout at default verbosity only**
  (progress chatter is `output(level>=1)`, suppressed at verbosity 1; `--json -v`
  interleaves and is unsupported for machine consumption).
- **CUE validation is CUE-only.** JSON/JSONL share the silent-unknown-field gap
  (`json.Decode` ignores extras) — backlogged in `TODO.md` (options:
  `DisallowUnknownFields`, or route JSON through the CUE schema since CUE ⊇ JSON).

## Next

Remaining open fronts (CUE + exit-code epics now fully closed):

- **Semantics epic** (2 doc/framing items left): one-line "drift ≠ correct" in
  `--help` (note: `usageHeader` in `main.go` *already* carries this framing —
  verify whether the item is effectively done), and the LLM-in-TDD guidance note
  (`docs/` or a skill).
- **Skill epic** — author `SKILL.md` so the repo doubles as an installable
  agent skill (setup + run→eyeball→commit + drift≠pass/fail framing). Reuses
  `docs/spec/using-smoke-in-tdd.md` + `exit-codes.md`.
- **Backlog** — partial commit, commit-last-run, all-errors validation, JSONC,
  JSON/JSONL unknown-field gap.

The `--json` mode is the natural input for the SKILL.md / LLM-guidance work —
the agentic surface those items describe now exists.
