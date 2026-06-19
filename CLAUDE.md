# PRODIGY9 Coding School

This project's AI coding environment is managed by
[ACE](https://github.com/ace-rs/ace). Run `ace` to start a coding session.
Run `ace setup` if not yet configured.

Skills and conventions are provided by the **PRODIGY9 Coding School** school and
are symlinked into `.claude/skills/`. Skill edits go through
symlinks into the school clone — propose changes back to the school repo
when ready. Run `ace config` or `ace paths` to debug configuration issues.

## What this repo is

SMOKE — a Go CLI for snapshot / golden-file smoke testing of arbitrary shell
commands. Capture a command's observable output once, lock it into a
`*.lock.yml`, and later runs report UNCHANGED when output matches and CHANGED when
it drifts. Not a replacement for proper tests; a fast drift detector.

## Repo layout

| Path           | Role                                                                       |
| -------------- | -------------------------------------------------------------------------- |
| `main.go`      | pflag CLI surface (`--init/--list/--print/--commit/--show-expected/--json`).|
| `process.go`   | Per-file orchestration: load → filter → run → (print \| commit \| compare).|
| `reporter.go`  | `reporter`: `status` enum + console impl; JSON in `report_json.go`.         |
| `engine/`      | Runner, `Config`, `Test`/`*Result` types, `RunHooks`.                      |
| `checks/`      | Pluggable observations: `exitcode`, `stdout`, `stderr`, file globs.        |
| `testspecs/`   | Spec loader for `tests.yml` / `.cue` (recursively nested test tree).       |
| `resultspecs/` | Lock-file format + diff/compare engine (name-keyed, order-sensitive).       |
| `internal/p`   | Console printing/coloring.                                                 |
| `test/`        | Real self-test suite (`tests.yml`); `test.sh` globs real suites here.       |
| `test/testdata/`| Fixtures the suite drives (intentionally CHANGED/NEW/malformed) — never run alone.|

Key idioms: the YAML root *is* a "root test"; `tests:` nests arbitrarily and
subtests inherit parent config/checks/commands. Commands are piped to
`interpreter -s` via stdin rather than argv-parsed.

## Durable artifacts

`docs/{notes,decisions,spec}/` — sorted by permanence (impermanent /
point-in-time / current). Default to `notes/`. See `docs/README.md` and
per-dir READMEs for picker details.

`docs/guides/index.md` is the canonical human guide (usage docs, GitHub-rendered).
`www/` is a Parcel project that generates a static site from it (`scripts/build-docs.sh`
→ `www/dist/`, gitignored). Edit the markdown, not the generated HTML.
