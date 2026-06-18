# Architecture

- **Status:** implemented

This describes how SMOKE is built today — the layers, the data model, and the
flow from a spec file to an exit code. It is the as-built map; the exit-code and
vocabulary contracts live in [`exit-codes.md`](exit-codes.md) and are referenced,
not restated, here.

## What SMOKE does

SMOKE captures a command's observable output once, locks it into a `*.lock.yml`,
and on later runs diffs fresh output against that golden. Match → exit `0`
(UNCHANGED); drift → exit `1` (CHANGED); no lock yet → exit `3` (NEW). It is a
drift detector, not an assertion engine: a clean run means "output didn't move,"
never "behavior is correct."

## Pipeline

One invocation maps over each spec filename through `processFile`:

```
spec file                lock file (*.lock.yml)
   │                              │
   ▼                              ▼
testspecs.Load ──► []*engine.Test ──► engine.Runner ──► []engine.TestResult
   (parse +            (flattened       (run cmds,          (raw check output)
    resolve)            test tree)       collect checks)         │
                                                                 ▼
                                              resultspecs.FromTestResult
                                                                 │
                              ┌──────────────┬───────────────────┤
                              ▼              ▼                    ▼
                          --print        --commit            (compare)
                         stdout YAML   write .lock.yml   Compare vs lock → edits
                                                          → UNCHANGED/CHANGED exit
```

`--show-expected` short-circuits before running anything: it loads the lock and
replays it through the same edit-printing path with every action set to `NoOp`.

## Two-stage spec model

The central design split is **input specs** vs **result specs** — two distinct
serialized shapes, never conflated:

| Stage  | Package        | Shape                            | Purpose                       |
| ------ | -------------- | -------------------------------- | ----------------------------- |
| Input  | `testspecs`    | `TestSpec` tree (`tests.yml`)    | what to run + how to observe  |
| Result | `resultspecs`  | `TestResultSpec` list (lock)     | what was observed (the golden)|

Input specs nest arbitrarily and carry config/commands/checks; result specs are a
flat list of observed output per command per check. The lock file is *always*
YAML regardless of input format — `.cue` specs lock to `.lock.yml`, since results
are a captured artifact we never need to round-trip back through CUE.

## Layers

| Path           | Responsibility                                                      |
| -------------- | ------------------------------------------------------------------ |
| `main.go`      | pflag surface, mode dispatch, `--init` scaffolder, usage/exit codes|
| `process.go`   | Per-file orchestration: load → filter → run → print/commit/compare |
| `testspecs/`   | Input parsing (YAML + CUE), inheritance resolution, tree flattening|
| `engine/`      | `Runner`, `Test`/`*Result` types, `Config`, `RunHooks`             |
| `checks/`      | Pluggable observations + the string→check parser                   |
| `resultspecs/` | Lock serialization + the structural diff engine                    |
| `internal/p`   | Console printing, coloring, exit-code constants                     |

`engine` and `resultspecs` have no knowledge of input format — the CUE/YAML split
is confined entirely to `testspecs.Load`.

## Input parsing and the CUE seam

`testspecs.Load` dispatches on file extension via `loaderFor` into a per-format
`loader` (`Load(io.Reader) (*TestSpec, error)`), default-deny:

- `.yml` / `.yaml` / *(none)* → `yamlLoader` (`yaml.NewDecoder`)
- `.cue` → `cueLoader` (compile → unify against `#Test` → validate → decode)
- `.json` → `jsonLoader`; `.jsonl` → `jsonlLoader` (one `TestSpec` per line)
- anything else → rejected (`unsupported spec format`)

All formats target the **same `TestSpec` struct**, which carries dual struct
tags: `yaml:"..."` for the YAML path and `json:"..."` for the CUE/JSON decoders.
The embedded `cuelang.org/go` evaluator is pinned in `go.mod`, keeping `.cue`
eval as hermetic as YAML parsing — no runtime `cue` binary on PATH.

One consequence drives a type choice: CUE has no native duration kind, so
`ConfigSpec.Timeout` is a `string` ("5s") parsed in `RunConfig` via
`time.ParseDuration`, shared by both loaders. There is no separate
`*time.Duration` field.

`cueLoader` unifies the user's file against an embedded (`//go:embed schema.cue`)
**closed** `#Test`/`#Config` schema before `Decode`. Closedness is recursive, so
a typo'd or wrong-typed field — even nested under `tests`/`config` — fails as a
clean CUE constraint error (`chekcs: field not allowed`) routed to exit `65`,
rather than being silently dropped at `Decode`. The schema doubles as a
reference for `.cue` spec authors.

JSON and JSONL reach the same fail-closed behavior via
`json.Decoder.DisallowUnknownFields` (shared `decodeJSON` helper): an unknown
key surfaces as `json: unknown field "chekcs"` — also recursive through
`config`/`tests`, also routed to exit `65`. The CUE schema additionally
constrains value types; the JSON decoder catches type mismatches at `Decode`
the same way.

## Inheritance resolution

After parsing, `TestSpec.Resolve(parent)` walks the tree applying parent→child
overriding before flattening:

- **Name** — child names are path-joined: `parent \ child`.
- **Config** — nil child config inherits the parent's wholesale; otherwise
  `ConfigSpec.Resolve` merges field-by-field (workdir path-joined, env appended,
  interpreter/timeout first-non-empty).
- **Commands / Checks** — *appended* parent-first, so a subtest inherits and
  extends rather than replaces.

At the root (`parent == nil`), an empty name defaults to the filename and a nil
config is seeded from `engine.DefaultConfig`. `TestSpec.Tests()` then flattens:
any node with commands becomes one `engine.Test`; children recurse. The YAML root
*is itself* a test — a top-level `commands:` runs.

## Checks

A check is anything implementing `checks.Interface`
(`Spec`/`Prepare`/`Collect`/`Format`). Checks are named by string in the spec and
resolved by `checks.Parse`:

- HTTP verbs (`GET `, `POST `, …) → `httpRequest`
- bare `exitcode` / `stdout` / `stderr` → the corresponding built-in
- any other bare path → `fileContent` (a file-glob observation)
- a URL with an unknown scheme → `nil` (reported as `unknown check`)

`Prepare` runs before the command (e.g. wiring stdout/stderr pipes); `Collect`
gathers raw bytes after; `Format` renders bytes to the lines stored in the lock.
`checks.Timeout` is a synthetic check the runner injects on deadline expiry — it
never appears in a spec.

## Execution

`DefaultRunner` runs each command by piping it to `interpreter -s` over **stdin**
rather than argv-parsing — closest to shell-native behavior, no quoting games.
Default interpreter `/bin/bash`, default timeout 3s.

Timeout handling is deliberate: a timed-out command is killed and recorded as a
`checks.Timeout` result with the *configured deadline* as data (not elapsed time,
so it stays stable across runs). That makes a hang compare as **drift (exit 1)**,
not an operational error (exit 2) — a misbehaving command is observable drift.
A non-zero command exit is likewise normal (the `exitcode` check captures it);
only a genuine `Wait()` error aborts as trouble.

`RunHooks` brackets every test/command/check with before/after callbacks;
`process.go` uses them only to prefix error messages with the test name.

## Compare and commit

`resultspecs.Compare` runs a structural diff (test → command → check → line),
returning a tree of `*Edit` values tagged `Equal`/`Added`/`Removed`/
`InnerChanges` plus a `differs` bool. Compare mode renders that tree through a
`reporter` (`reporter.go`) chosen by output format: `consoleReporter` prints
only non-`Equal` edits, `jsonReporter` (`--json`) emits the full verdict tree as
JSON. Both map the outcome to a `status` enum that owns its exit code; the
no-lock (`NEW`) path routes through the same reporter. Commit mode writes results
to a temp file and atomically renames into place, so a crash mid-write never
corrupts an existing lock.

`--include` / `--exclude` filter by test name at both ends (the live test list
and the loaded lock), so a partial run compares only the named subset. Committing
a partial set is refused — a partial commit would silently drop the unmatched
tests from the lock.

## Modes and exit codes

| Mode               | Flag              | Effect                                          |
| ------------------ | ----------------- | ----------------------------------------------- |
| Compare *(default)*| —                 | diff vs lock → `0`/`1`/`3`                       |
| Compare (JSON)     | `--json`          | same diff as a machine-readable document        |
| Commit             | `--commit`/`-c`   | write/overwrite the lock                        |
| Print              | `--print`/`-p`    | result YAML to stdout (scripting)               |
| List               | `--list`/`-l`     | discovered test names (`-vv` adds commands)     |
| Show expected      | `--show-expected` | replay the lock without running                 |
| Init               | `--init[=path]`   | scaffold a starter spec (no-clobber)            |

Exit codes (`internal/p` constants `Exit{Unchanged,Changed,Trouble,New,Usage,DataErr}`)
and the output vocabulary are the frozen contract in
[`exit-codes.md`](exit-codes.md).
