# SMOKE — codebase survey

> Point-in-time survey at the start of the ACE session. Reflects the repo as of
> commit `b6d3cf5`. May drift; treat as orientation, not contract.

## What it is

A Go CLI for **snapshot / golden-file smoke testing of arbitrary shell
commands**. Premise from the README: *code that produces the same observably
correct output exhibits the same correct behavior.* You capture a command's
observable output once, lock it into a `.lock.yml`, and later runs go GREEN when
output matches and RED when it drifts.

## Workflow

1. Write `tests.yml`.
2. `smoke tests.yml` → eyeball output.
3. `smoke -c tests.yml` → commit a `*.lock.yml` (checked into VCS).
4. Later runs compare against the lock: stable → exit 0, drift → exit 1.

## Package layout

| Package        | Role                                                                                              |
| -------------- | ------------------------------------------------------------------------------------------------ |
| `main.go`      | pflag CLI: `--init/--list/--print/--commit/--show-expected`, include/exclude filters, verbosity. |
| `process.go`   | Per-file orchestration: load → filter → run → (print \| commit \| compare). `lockFilename`.      |
| `engine/`      | `Config`, `Test`/`CommandResult`/`TestResult`, `DefaultRunner`, `RunHooks`.                       |
| `checks/`      | Pluggable observations: `exitcode`, `stdout`, `stderr`, file-content globs.                       |
| `testspecs/`   | YAML loader for `tests.yml` (recursively nested test tree).                                       |
| `resultspecs/` | Lock-file format + diff/compare engine (uses `chakrit/gendiff`).                                  |
| `internal/p`   | Console printing/coloring.                                                                        |
| `internal/`    | `lists.go` whitelist/blacklist filtering.                                                         |

## Design points worth knowing

- **One recursive schema.** The YAML root *is* a "root test"; `tests:` nests
  arbitrarily. Subtests inherit the parent's config, checks, and command list;
  config may be overridden per subtest. Recent commits hardened this inheritance.
- **`interpreter -s`.** Commands are piped to the interpreter via stdin rather
  than argv-parsed — stays shell-native, avoids writing a parser.
- **Hooks indirection** (`hooks.go` ↔ `engine.RunHooks`) decouples the engine
  from the console printer.
- **Self-hosting.** `test/tests.yml` smoke-tests the `smoke` binary itself
  (build, `--list`, error cases, subtest config inheritance).

## Known rough edges (see TODO.md / task list)

- `main.go --init` writes `smoke-tests.yml`, but the template's `tests.yml`
  references and README examples are inconsistent about the canonical filename.
- README's `goimports` self-test path references a `specs/` dir that no longer
  exists (split into `testspecs/` + `resultspecs/`).
- `engine.DefaultRunner.Command` mutates `t.RunConfig` in place when filling
  interpreter/timeout defaults — worth auditing against the inheritance model.
