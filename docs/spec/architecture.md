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
   (read + splice      (flattened       (run cmds,          (raw check output)
    + resolve)          test tree)       collect checks)         │
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

`testspecs.Load` takes a **path**, not a reader: spec file I/O lives in the
package, because resolving `include`s means opening more files relative to each
hop's own directory (see *Include resolution*). `process.go` no longer reads spec
files — it hands `Load` the filename and maps the result.

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
| `testspecs/`   | Spec file I/O, parsing (YAML + CUE), include splice, resolution, flatten|
| `engine/`      | `Runner`, `Test`/`*Result` types, `Config`, `RunHooks`             |
| `checks/`      | Pluggable observations + the string→check parser                   |
| `resultspecs/` | Lock serialization and the structural (name-keyed) diff engine     |
| `internal/p`   | Console printing, coloring, exit-code constants                     |

`engine` and `resultspecs` have no knowledge of input format — the CUE/YAML split
is confined entirely to `testspecs.Load`.

## Input parsing and the CUE seam

`testspecs.Load(filename)` opens the file through the recursive `loadSpec(path,
stack)` seam, then dispatches on file extension via `loaderFor` into a per-format
`loader` (`Load(reader io.Reader, path string) (*TestSpec, error)`),
default-deny. Opening files and resolving includes live *above* the loader in
`loadSpec`, once, format-agnostic — so the byte-stream loaders stay dumb
reader→tree decoders and ignore `path`. The one exception is `cueLoader`, which
loads *by path* (not the reader) so a `.cue` inside a `cue.mod` module can resolve
imports; the reader is still opened in `loadSpec` first, which is what classifies
a missing root (exit `2`) vs a missing include (exit `65`) before any loader runs.

- `.yml` / `.yaml` / *(none)* → `yamlLoader` (`yaml.NewDecoder`)
- `.cue` → `cueLoader` (`cue/load` instance → unify against `#Test` → validate → decode)
- `.json` → `jsonLoader`; `.jsonl` → `jsonlLoader` (one `TestSpec` per line)
- `.jsonc` → `jsoncLoader` (strip `//` `/* */` comments → `jsonLoader` path)
- anything else → rejected (`unsupported spec format`)

All formats target the **same `TestSpec` struct**, which carries dual struct
tags: `yaml:"..."` for the YAML path and `json:"..."` for the CUE/JSON decoders.
The embedded `cuelang.org/go` evaluator is pinned in `go.mod`, keeping `.cue`
eval as hermetic as YAML parsing — no runtime `cue` binary on PATH.

`cueLoader` loads via `cue/load.Instances([]string{abs}, &load.Config{Dir:
filepath.Dir(abs)})` rather than compiling raw bytes, so a spec inside a `cue.mod`
module can `import` shared packages (a `#Case` schema reused across many specs —
the lowfat-pantry DRY case). The path is **absolutized first**: `cue/load`
resolves file args relative to `Config.Dir`, so a relative path plus a `Dir` would
double up (`a/b/a/b/spec.cue`). A lone `.cue` with no `cue.mod` loads exactly as
before, as a single anonymous instance — local-only resolution keeps eval
hermetic.

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

JSONC is JSON with comments only. `jsoncLoader` runs `stripJSONComments` — a
string-aware state machine that blanks `//` line and `/* */` block comments to
spaces (newlines kept, so `json.Decoder` offsets stay accurate) while leaving
comment-like sequences inside string literals untouched — then decodes through
the same `decodeJSON`, inheriting its fail-closed behavior. Trailing commas are
not tolerated; that would be structural editing, not comment stripping.

## Include resolution

A node may carry `include: <path>` to splice another spec file in as a child of
that node, so a suite can span files and shared config/commands be defined once.
`loadSpec` resolves includes as a **pre-pass before `Resolve`**: decode the file,
then for each `include` node, recurse into the referenced file and graft its root
in as a child. `Resolve` and `Tests()` then run over the merged tree unchanged —
an imported test is indistinguishable from an inline one once spliced.

The model, ruled in `docs/decisions/2026-06-23-include-import-design.md`:

- **Two nodes, file-relative.** The imported file's root is kept as a real child
  node (not flattened away), so a single-test imported file never loses its
  commands. The imported root's segment defaults to the include path as written
  (an explicit `name:` in the imported file wins, mirroring the root-basename
  default). Path is resolved against the **including file's** directory, per hop —
  files stay movable.
- **Mutually exclusive with `tests:`** on one node (both → exit 65).
- **Single root lock.** Imported tests flatten into the root spec's one lock,
  keyed by full name; there are no per-imported-file locks. Editing an imported
  file surfaces as drift in the root lock.
- **Inheritance is the existing `Resolve`.** The imported root is just a child, so
  config merges and commands/checks prepend exactly as for an inline child. That
  uniformity powers **parameterized include**: env flows down, and an imported
  test's name interpolates it via `os.Expand` (`$VAR`/`${VAR}`, undefined → empty,
  declared env only — never `os.Environ`). One shared file under siblings that set
  different env yields distinctly-named copies.
- **Cycle guard, diamonds allowed.** `loadSpec` carries an **ancestor stack** of
  abs+clean paths (cloned per descent, not a global visited-set). A file already
  among its own ancestors is a cycle (exit 65); the same file reached down two
  distinct branches (a diamond) loads independently each time, distinct by parent
  chain. Cross-format includes (`.yml` pulling a `.cue`/`.jsonl`) fall out of
  per-file `loaderFor` dispatch for free. A missing *included* file is malformed
  content (exit 65), distinct from a missing *root* spec (operational, exit 2).

## Inheritance resolution

After parsing, `TestSpec.Resolve(parent)` walks the tree applying parent→child
overriding before flattening:

- **Name** — child names are path-joined: `parent \ child`.
- **Config** — nil child config inherits the parent's wholesale; otherwise
  `ConfigSpec.Resolve` merges field-by-field (workdir path-joined, env appended,
  interpreter/timeout first-non-empty).
- **Commands / Checks** — *appended* parent-first, so a subtest inherits and
  extends rather than replaces.

At the root (`parent == nil`), an empty name defaults to the spec's **basename**
(`filepath.Base`), so the same spec keys the lock identically regardless of how its
path was typed (`x.yml`, `./x.yml`, `dir/x.yml`); a nil config is seeded from
`engine.DefaultConfig`. `TestSpec.Tests()` then flattens:
any node with commands becomes one `engine.Test`; children recurse. The YAML root
*is itself* a test — a top-level `commands:` runs.

`Tests()` walks the resolved tree depth-first: a node with commands becomes one
`engine.Test`, a command-less leaf is a malformed-spec error, and each failable
field (an unknown check name, a bad timeout duration) is checked as it's visited.
Identity is composed here, not in `Resolve`: each node's `engine.TestName` is its
parent's name extended by its own segment (`TestName.Child`), so the flattened
name is minted exactly at this gate. The segment is `os.Expand`ed against the
node's resolved env first (see *Include resolution*), so a non-imported name with
no `$` is untouched. After the walk, `testspecs.Load` asserts the
flattened names are **unique** — two tests that flatten to the same name make the
name-keyed lock ambiguous, so a duplicate is a malformed spec. The walk collects
**every** per-node fault — unknown checks, bad timeouts, command-less leaves — and
joins them (`errors.Join`), so one `Load` surfaces all of a spec's tree-walk errors
at once; the uniqueness pass still stops at the first duplicate. Either way the
result flows out through `testspecs.Load`, routing to exit `65` like any other
malformed-spec failure.

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
no-lock (`NEW`) path routes through the same reporter. Commit mode writes the
lock to a temp file and atomically renames into place, so a crash mid-write
never corrupts an existing lock.

The diff is **order-sensitive by design**. Tests are not isolated: order is
load-bearing — setups and teardowns are modelled as the ordering of sibling
tests, and a run executes them top-to-bottom. So `compareTests` is an LCS
sequence diff (`gendiff`) keyed on the test name; a reordered or out-of-place
test surfaces as a real change, not a false match.

`--include` / `--exclude` filter by test name at both ends (the live test list
and the loaded lock), so a partial run compares only the named subset. A
whole-suite **commit** writes the lock outright; a *filtered* commit **merges**
instead (`resultspecs.Merge`): the run's results land in spec order and tests
that weren't run keep their existing lock entries, so a partial commit never
prunes what it didn't observe. The merge walks the spec (not the lock or the
filter), so a test deleted from the spec drops and a newly committed one lands in
its spec position — order-sensitivity preserved. The merge keys on the test name,
which the loader guarantees unique.

## Modes and exit codes

| Mode               | Flag              | Effect                                          |
| ------------------ | ----------------- | ----------------------------------------------- |
| Compare *(default)*| —                 | diff vs lock → `0`/`1`/`3`                       |
| Compare (JSON)     | `--json`          | same diff as a machine-readable document        |
| Commit             | `--commit`/`-c`   | run, then write the lock (merges if filtered)   |
| Print              | `--print`/`-p`    | result YAML to stdout (scripting)               |
| List               | `--list`/`-l`     | discovered test names (`-vv` adds commands)     |
| Show expected      | `--show-expected` | replay the lock without running                 |
| Init               | `--init[=path]`   | scaffold a starter spec (no-clobber)            |

A run takes one or more specs. `main` is the **single exit authority**:
`processFile` returns `(status, error)` per spec, `main` folds the verdicts
(`status.Merge` — `UNCHANGED` is the identity, so a clean spec never masks an
earlier drift) and fail-fasts on a fatal (typed in `outcome.go`: `dataError` →
`65`, else operational → `2`). Each spec reports separately; `--json` is one
compact object per spec (JSONL). See
[`exit-codes.md`](exit-codes.md) §"Multiple specs".

Exit codes (`internal/p` constants `Exit{Unchanged,Changed,Trouble,New,Usage,DataErr}`)
and the output vocabulary are the frozen contract in
[`exit-codes.md`](exit-codes.md).
