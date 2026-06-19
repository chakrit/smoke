# SMOKE — the human guide

SMOKE is a CLI for **snapshot smoke-testing of shell commands**. You capture a command's
observable output once, lock it into a `*.lock.yml`, and every later run reports
**UNCHANGED** when the output still matches and **CHANGED** when it drifts.

It rests on one assumption: *code that produces the same observably-correct output exhibits
the same correct behaviour.* That is not true 100% of the time — but it gets you a long way
for very little effort. SMOKE is a **drift detector, not an assertion engine**: a clean run
means "the output didn't move," never "the behaviour is correct." It is not a replacement
for a proper test suite; it is a fast tripwire.

<!--DIAGRAM:lifecycle-->

## Install

```sh
go install github.com/chakrit/smoke@latest
```

That drops a `smoke` binary on your `GOBIN`. No runtime dependencies — the YAML and CUE
parsers are compiled in.

## Quick start

The first-run workflow is five steps:

```sh
# 1. Write a spec describing what to run and what to observe.
cat > tests.yml <<'EOF'
config:
  interpreter: /bin/bash
checks:
  - exitcode
  - stdout
tests:
  - name: Greeting
    commands:
      - echo "hello world"
EOF

# 2. Run it. No lock exists yet, so SMOKE reports NEW (exit 3).
smoke tests.yml

# 3. Eyeball the output. Does it look right?

# 4. If yes, commit it as the golden.
smoke -c tests.yml          # writes tests.lock.yml

# 5. Check tests.lock.yml into source control.
git add tests.yml tests.lock.yml
```

From then on, `smoke tests.yml` compares fresh output against the locked golden. Your
teammates run the same command and get the same verdict.

## Writing a spec

A spec is YAML (CUE, JSON, JSONL, and JSONC are also accepted — see
[Advanced](#advanced-spec-formats)). The **root of the document is
itself a test** — a top-level `commands:` runs. `tests:` nests arbitrarily, and subtests
**inherit** their parent's config, checks, and commands.

```yaml
name: "Defaults to the filename if omitted"
config:
  interpreter: /bin/bash    # interpreter for commands (default /bin/bash)
  timeout: 5s               # per-command timeout (default 3s)
  workdir: .                # working directory commands start in
  env:                      # extra environment for commands
    - "PATH=./bin:/bin"
checks:                     # what to observe and record per command
  - exitcode                # the command's exit status
  - stdout                  # the whole standard-output stream
  - stderr                  # the whole standard-error stream
  - go.mod                  # a file's contents
  - generated/*.go          # a file glob's contents
commands:
  - go install -v .
tests:
  - name: Subtest
    config:
      workdir: ..           # overrides the parent's workdir
    commands:               # appended after the inherited commands
      - smoke
```

Inheritance rules, briefly:

- **Name** — child names are path-joined: `parent \ child`.
- **Config** — a missing child config inherits the parent's wholesale; otherwise it merges
  field-by-field (workdir path-joined, env appended, interpreter/timeout first-set-wins).
- **Commands and checks** — *appended* parent-first, so a subtest extends rather than
  replaces.

Commands are piped to `interpreter -s` over **stdin** rather than parsed as argv, so quoting
and shell features behave exactly as they would in a real shell.

## The lifecycle: NEW, UNCHANGED, CHANGED

Every run lands in one of three states, each with its own exit code:

| Run says    | Exit | Meaning                                                          |
| ----------- | ---- | --------------------------------------------------------------- |
| `NEW`       | `3`  | No lock yet — the first run is unreviewed.                      |
| `UNCHANGED` | `0`  | Output matched the lock. *Not* "tests passed."                 |
| `CHANGED`   | `1`  | Output drifted from the lock.                                  |

The discipline that makes SMOKE trustworthy: **eyeball before you commit.** `UNCHANGED`
means drift-free, never verified-correct. Re-committing a `CHANGED` result you didn't look
at just locks in the drift as the new "truth."

## Committing goldens

`smoke -c tests.yml` (or `--commit`) runs the spec and writes the result to
`tests.lock.yml`, replacing whatever was there. A commit always overwrites the **whole**
lock, so a test you delete from the spec disappears from the lock too.

Commit mode writes to a temp file and atomically renames it into place — a crash mid-write
never corrupts an existing lock.

## Filtering which tests run

Two flags narrow a run to a subset of tests by name (substring match):

```sh
smoke --include Greeting tests.yml     # only tests whose name contains "Greeting"
smoke --exclude Slow tests.yml         # everything except names containing "Slow"
```

A commit writes the **whole** lock, so committing a filtered run would prune the tests it
never observed. SMOKE refuses that — `--commit` together with `--include`/`--exclude` is a
usage error (exit `64`). Filter to *run* a subset; commit the whole spec.

## Exit codes

SMOKE exits with exactly one of these. Each names a distinct outcome class; the shipped
codes are a frozen contract.

| Code | State       | Meaning                                                       |
| ---- | ----------- | ------------------------------------------------------------- |
| `0`  | `UNCHANGED` | Output matched the lock. *Not* "tests passed."                |
| `1`  | `CHANGED`   | Drift detected — output moved (includes `MISSING`, timeout).  |
| `2`  | —           | Operational error: SMOKE itself broke (runner crash, I/O).    |
| `3`  | `NEW`       | No lock file; first run is unreviewed.                        |
| `64` | —           | Usage error: invalid invocation (bad flags).                  |
| `65` | —           | Data error: a spec or lock file is malformed.                 |

A timed-out command is recorded as drift (`1`), not an operational error — a hang is
observable behaviour change, not a tool failure. A non-zero command exit is likewise normal
(the `exitcode` check captures it); only SMOKE itself failing is `2`.

## Using SMOKE in CI

The rule is simple: **`0` is the only clean pass; fail the build on any non-zero.** That is
the default behaviour of most CI runners, so often you just run it:

```sh
smoke tests.yml    # nonzero exit fails the step
```

Make sure `tests.lock.yml` is **committed**. Without it, CI gets `NEW` (exit `3`) and fails
— which is correct: an unreviewed first run should never pass silently.

To check several specs, pass them all in one invocation:

```sh
smoke --no-color tests/*.yml
```

Each spec is reported separately and the run exits once: `0` only if every spec is clean,
non-zero if **any** spec drifted (a clean spec never hides an earlier drift). A malformed
spec or a runner failure (`65`/`2`) stops the run there — specs run in order and may set up
state the next depends on, so a broken one means the rest can't be trusted. Drift never
stops the run, so you see every spec's verdict. (Prefer a `for` loop only if you need each
spec in its own process.)

`--no-color` keeps CI logs clean. For machine-readable output, `--json` emits one compact
JSON object per spec (a JSONL stream for multiple specs; one object for a single spec),
compare mode only.

Driving SMOKE from an agent or a TDD loop has its own playbook — see
[`using-smoke-in-tdd`](https://github.com/chakrit/smoke/blob/main/docs/spec/using-smoke-in-tdd.md).

## Advanced: spec formats

YAML is the default, but the same spec model loads from four other formats, chosen by file
extension. Dispatch is **default-deny** — an unrecognized extension is rejected outright,
never guessed — and every format resolves to the identical test tree, so inheritance,
checks, and the lifecycle behave exactly as they do in YAML.

One difference matters up front: **YAML silently drops an unknown key**, so a typo'd
`chekcs:` just vanishes. CUE, JSON, and JSONC all **fail closed** — an unknown field is a
load error (exit `65`), not a silent no-op. Reach for them when a silent drop would bite.

### CUE (`.cue`)

A `.cue` spec is unified against an embedded, **closed** `#Test`/`#Config` schema before it
decodes, so typo'd fields and wrong types surface as constraint errors — recursively,
including nested `tests` and `config`. Beyond checking, CUE brings constraints, defaults,
imports, and comprehensions, so you can *generate* a spec instead of hand-copying it:

```cue
config: {
	interpreter: "/bin/bash"
	timeout:     "5s"   // a string — CUE has no duration type
}
checks: ["exitcode", "stdout"]

// One subtest per fixture, no copy-paste.
tests: [
	for name in ["alpha", "beta", "gamma"] {
		"name":   name
		commands: ["./run.sh \(name)"]
	},
]
```

Reach for CUE when specs grow large or repetitive and you'd rather compute them than
maintain them by hand — or when you already author config in CUE.

### JSON (`.json`)

One JSON object is the root test, with the same keys as YAML (`config`, `checks`,
`commands`, `tests`). Decoding **rejects unknown fields**, matching CUE's closed schema:

```json
{
  "config": { "interpreter": "/bin/bash" },
  "checks": ["exitcode", "stdout"],
  "tests": [
    { "name": "Greeting", "commands": ["echo hello"] }
  ]
}
```

Reach for JSON when a tool emits the spec — it's the lowest-friction format to generate from
another program.

### JSONL (`.jsonl`)

One `TestSpec` per non-blank line (blank lines are skipped). The stream is the **children of
an implicit empty root** — equivalent to a YAML `tests: [...]` with no top-level command.
Each line decodes independently and fails closed on unknown fields:

```jsonl
{ "name": "Greeting", "commands": ["echo hello"] }
{ "name": "Farewell", "commands": ["echo bye"] }
```

The catch: the implicit root carries no config or checks, so **lines share no parent
settings** — each stands alone on the defaults. That makes JSONL a fit for a stream of
independent, self-contained tests (say, appended by a generator) and a poor fit when tests
need shared config or checks; use YAML, JSON, or CUE for those.

### JSONC (`.jsonc`)

Plain JSON with comments — the same object as `.json`, but `//` line and `/* */` block
comments are stripped before decoding. Comment markers inside string literals are data, not
comments, so a command like `"echo // here"` is left intact:

```jsonc
{
  // shared settings for the whole spec
  "config": { "interpreter": "/bin/bash" },
  "checks": ["exitcode", "stdout"],
  "tests": [
    { "name": "Greeting", "commands": ["echo hello"] } /* one per behaviour */
  ]
}
```

Comments are stripped to spaces (line numbers preserved), so decode errors still point at
the right line, and unknown fields fail closed exactly as in `.json`. **Trailing commas are
not allowed** — JSONC here means comments only. Reach for it when you want a JSON spec you
can annotate inline.

## Flag reference

| Flag               | Short | Effect                                                     |
| ------------------ | ----- | ---------------------------------------------------------- |
| `--commit`         | `-c`  | Run, then write the whole lock (refused with a filter).    |
| `--print`          | `-p`  | Print result YAML to stdout (for scripting).               |
| `--list`           | `-l`  | List discovered test names (`-vv` adds commands).          |
| `--show-expected`  | `-s`  | Replay the lock without running anything.                  |
| `--json`           |       | Emit the compare verdict as JSON (compare mode only).      |
| `--init[=path]`    |       | Scaffold a starter spec (won't clobber an existing file).  |
| `--include`        | `-i`  | Only run tests whose name contains the pattern.            |
| `--exclude`        | `-x`  | Skip tests whose name contains the pattern.                |
| `--no-color`       |       | Disable console colouring.                                 |
| `--time`           |       | Log timestamps.                                            |
| `--verbose`        | `-v`  | More output (repeatable: `-vv`).                           |
| `--quiet`          | `-q`  | Less output (repeatable).                                  |
| `--help`           | `-h`  | Show usage.                                                |
