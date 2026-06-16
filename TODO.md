# TODO

## Epic: Fix `--init`

`main.go`'s `--init` is half-baked. Tighten it into a predictable scaffolder.

* [x] Honor a custom filename: `smoke --init=foo.yml` writes `foo.yml`. Bare `smoke
      --init` falls back to the default. (pflag's `NoOptDefVal` can't also consume a
      space-separated arg, so the `=` form is required; a stray positional errors
      rather than silently writing the default.)
* [x] Settle the canonical default filename: `tests.yml` (already what README and the
      self-tests use; `--init` was the lone outlier writing `smoke-tests.yml`).
* [x] Don't clobber. `initSpec` opens with `O_CREATE|O_EXCL` and refuses if the file
      exists. (Hard refuse, no `--force` override — by decision.)
* [x] Report the path actually written (`p.Pass("Wrote " + target)`).
* [x] Add a self-test in `test/tests.yml` (`Tests \ Init`): writes a named file +
      the no-clobber guard.

## Epic: First-class CUE support

Today `testspecs.Load` is YAML-only. Make `.cue` test specs a first-class input so
specs can be generated/validated by CUE tooling. Folds in the old stdin item below.

* [x] Integration boundary: **embed `cuelang.org/go`** (pinned evaluator, hermetic
      like the YAML path; no runtime `cue` dependency). Ruled inline — no separate
      decision doc for a tool this size. Note: the dep forces `go 1.25` in `go.mod`,
      which made `go vet`'s non-constant-format-string check fatal and required a
      cleanup of `internal/p`'s `output`/`outputErr` helpers (now literal writers; the
      one formatted callsite uses a constant format).
* [x] Front-end dispatch by extension in `testspecs.Load` (`.cue` → `cuecontext`
      eval → `TestSpec` via json tags → existing `Resolve`/`Tests`; `.yml`/`.yaml`
      → current path). Unknown extensions now rejected (default-deny). engine /
      resultspecs layers untouched.
* [x] Lock-file semantics: `lockFilename` maps `foo.cue` → `foo.lock.yml` (results
      are always YAML). `.yml`/`.yaml` unchanged.
* [x] **Loader abstraction + leaf validation.** Replace the extension `switch` in
      `testspecs.Load` with an unexported `loader interface { Load(io.Reader)
      (*TestSpec, error) }` + `loaderFor` dispatch on `filepath.Ext` (default-deny).
      Each loader owns format-specific parsing; `Load()` =
      `loaderFor` → `loader.Load` → set `Filename` → `Resolve(nil)` → `Tests()`.
      **Plan finalized (2026-06-16): first-error, minimal — NO standalone
      `validate(*TestSpec)`, NO shadow fields, NO IR.** The one genuinely-new check
      (a leaf — `len(Children)==0` — must have commands, else error not silent skip)
      folds into the existing `Tests()` walk, alongside the check-name and timeout
      parses that already live there as **live** conversion errors. All parse/leaf
      failures exit `65` (`EX_DATAERR`) via `p.DataErr` — path already wired, no new
      wiring. (CORRECTION: the old "takes over timeout parsing so a bad `"5s"` fails
      at load not run" rationale was wrong — `RunConfig` is reached only via `Tests`
      via `Load`, so it already parses at load today. The seam was only
      organizational, and we're consciously not reorganizing it now; the proper
      split into a total parse + fold-based validate is deferred to the all-errors
      pass — see backlog.) Adds JSON (`json:` tags already exist — near-free) and
      JSONL (each line a `TestSpec`; the stream is the children of an implicit empty
      root — equiv. to a YAML `tests: [...]` with no top-level command). **JSONC
      deferred** (needs a dep or a string-aware comment-stripper — its own
      follow-up). `lockFilename` generalizes: any non-YAML ext → `.lock.yml`.
      Self-tests: round-trip per new format + one structure-invalid (bad-leaf) case
      proving the leaf check fires → exit 65. **Slice C DONE (2026-06-16):** loaders
      in `testspecs/loaders.go` (`loader`/`loaderFor` + yaml/cue/json/jsonl), leaf
      check folded into `Tests()`, `lockFilename` YAML-whitelisted. Go unit tests in
      `testspecs/testspecs_test.go` + `lockfile_test.go`; smoke self-tests added
      (JSON/JSONL round-trip, Bad-leaf → 65). Load `general-coding`, `go-coding`,
      `cue-coding`.
* [-] ~~Read tests from stdin (`cue export | smoke -`).~~ **Dropped (2026-06-16).**
      Piping fights the core premise: drift detection needs a *stable spec identity*
      for the lock to compare against over time; a piped spec is identity-less, so
      "where does the lock live?" has no good answer. Slice C made `.cue` a
      first-class *file* input, which obsoletes the motivating example —
      `smoke spec.cue` → stable `spec.lock.yml`. Stdin would only serve lockless
      modes (`--list`/`--print`), i.e. SMOKE as a command runner, not a drift
      detector. _Revisit only if a concrete tool-generated-spec workflow appears that
      genuinely can't write a file first._
* [ ] Ship a CUE schema (`#Test`/`#Config` definitions) so authors get validation and
      editor support when writing `.cue` specs. Lands as the cueLoader's
      format-specific validation step (unify before `Decode`). **Slice D.**
* [x] Self-tests: a `.cue` spec round-trips (run → commit → stable) under `test/`.
      (`test/cuetests.cue` + `test/tests.yml \ Tests \ CUE`.)

## Epic: Disambiguate semantics for LLM consumers

SMOKE's truth claim is *"observable output matches the committed golden"* — but
every consumer trained on test-runner conventions (green/red, exit 0/1) reads it
as *"behavior is correct."* That gap is tolerable for a human doing
eyeball-and-commit; it actively misleads an LLM driving a TDD loop. The command
surface and output vocabulary must stop borrowing pass/fail connotations they
don't earn.

Failure modes to design against:

* **STABLE-but-wrong.** Behavior changed, but the observable output didn't move →
  green → the agent concludes nothing broke. (Coverage gap, not a test pass.)
* **STABLE-encodes-a-bug.** The golden itself locked in incorrect behavior;
  stability perpetuates it. Green forever, wrong forever.
* **RED-misread-as-failure.** "Changes Detected" / exit 1 is the *expected* state
  during an intentional change — it means "eyeball and re-commit," not "your code
  is broken." LLMs pattern-match it to a failing test and try to "fix" the output
  back to green, defeating the workflow.

Tasks:

* [x] Redesign the output vocabulary so it describes *drift*, not *correctness*.
      Drop `STABLE`/`Changes Detected` framing that reads as pass/fail; prefer
      neutral states (e.g. `UNCHANGED` / `CHANGED` / `NEW` / `MISSING`). Never
      emit "pass", "green", or a ✓ that implies a passing assertion. (Compare-mode
      verdicts now `UNCHANGED`/`CHANGED`/`NEW`; ✔/✘ dropped; no green/red on
      verdicts. `MISSING` folds into `CHANGED` per the contract.)
* [x] Exit-code semantics: see the dedicated "Exit-code design" epic below.
      (Contract wired: `0/1/2/3/64` as `internal/p` constants.)
* [x] Distinguish the no-lock first-run state (`NEW` / `UNREVIEWED`) from
      `UNCHANGED`, so an agent knows a human/LLM eyeball is still required before
      the golden can be trusted. (No-lock now reports `NEW` + exit `3` instead of
      hard-erroring.)
* [ ] Add a machine-readable mode (`--json` or similar) reporting per-check status
      with unambiguous fields (`matched`/`changed`/`new`/`missing`) and zero
      pass/fail language — the primary surface for agentic consumers. Dovetails
      with the CUE/stdin work as the output mirror of structured input.
* [ ] Surface a one-line framing in human output and `--help`: SMOKE is a drift
      detector, not an assertion engine; "UNCHANGED does not mean correct."
* [ ] Write LLM-facing guidance (a skill or `docs/` note) for SMOKE-in-TDD loops:
      why drift ≠ failure, why green can't be auto-chased, when re-committing the
      golden is correct vs. when it's hiding a regression.
* [ ] Record the vocabulary + exit-code contract as a `docs/decisions/` entry —
      this is a semantics ruling other tools/agents will depend on.

## Epic: Exit-code design

Today exit `1` is overloaded: it means both "output drifted" (an expected, benign
workflow state needing review) *and* "the tool itself failed" (timeout, missing
lock, unparseable spec, runner error). Nothing wrapping SMOKE — a CI gate, a shell
script, an LLM — can tell a regression from a crash. Give each outcome class a
distinct, documented, stable code.

Anchor on `diff(1)`'s long-established convention, since SMOKE is conceptually a
diff: `0` = no difference, `1` = differences found, `2` = trouble. This keeps "fail
the build on any nonzero" working for CI while letting an agent branch on the
specific code.

Proposed code space (to ratify in the decision doc):

| Code | Meaning                          | Notes                                  |
| ---- | -------------------------------- | -------------------------------------- |
| 0    | All checks matched the lock      | NOT "tests passed" — drift-free only.  |
| 1    | Drift detected (output changed)  | Expected during intentional changes.   |
| 2    | Operational error                | Bad spec, runner failure, timeout.     |
| ?    | No lock / unreviewed first run   | Distinct code, or fold into 2? Decide. |
| 64+  | Usage error                      | `pflag` convention is `2`; reconcile.  |

Tasks:

* [x] Inventory every current exit path (`main.go`, `process.go` compare/commit,
      `runTests` failure, usage errors) and map each to an outcome class. (Captured
      in the decision doc + `docs/spec/exit-codes.md`.)
* [x] Resolve the collision between drift (`diff`-style `2` = trouble) and pflag's
      usage-error `2`. Pick one scheme and apply it consistently. (Trouble keeps `2`;
      usage moves to `64`/`EX_USAGE`.)
* [x] Decide the no-lock / unreviewed-first-run code (today it hard-errors). It is
      semantically distinct from both "matched" and "drift". (Own code `3` = `NEW`.)
* [x] Implement distinct codes; centralize them as named constants rather than
      scattered `os.Exit(1)` literals. (`p.Exit{Unchanged,Changed,Trouble,New,Usage}`
      in `internal/p`. Timeout now drifts as exit `1` via synthetic `checks.Timeout`;
      operational/usage diagnostics route to stderr. Spec status now `accepted`.)
* [ ] Mirror the exit code in the `--json` output (a `status` field) so agents
      don't have to shell-inspect `$?`.
* [x] Document the full table in `--help` and README; freeze it as a contract.
      (`usageExitCodes` in `main.go`; table in `README.md`.)
* [x] Record the chosen scheme in `docs/decisions/` (shared ruling with the
      semantics epic's vocabulary contract). → `2026-06-08-exit-code-contract.md`
      + live contract in `docs/spec/exit-codes.md`.

## Epic: Repo doubles as an installable skill

Make this repo usable as an agent skill: an agent adds the repo as a skill and
gets instructions on correct use of the tool — how to set up golden-file smoke
testing, the workflow (run → eyeball → commit), and the gotchas to watch.

The headline gotcha is the semantics this project already nailed down: red/green
here means *unmatched/diffed*, NOT test pass/fail. `CHANGED` is expected drift to
review, not a failing assertion; `UNCHANGED` is drift-free, not verified-correct.

* [ ] Author a `SKILL.md` (skill-creator conventions) with name/description
      front-matter so the skill triggers on golden-file / snapshot / smoke-test
      intents.
* [ ] Body: setup (`--init`, writing `tests.yml`, checks), the run→eyeball→commit
      workflow, exit-code contract, and the drift≠pass/fail framing. Reuse
      `docs/spec/using-smoke-in-tdd.md` and `docs/spec/exit-codes.md` rather than
      duplicating — link or distill.
* [ ] Decide packaging: where `SKILL.md` lives so `ace`/Claude skill loaders find
      it, and whether supporting `references/` files are needed.

## Backlog (unsorted)

* [ ] Allow partially committing some results but not all.
* [ ] Allow committing last run results (so we don't have to re-run tests to commit
      again).
* [ ] All-errors validation reporting — collect every spec error per load, not just
      the first. Shares the "don't abort on first problem" machinery with partial
      load/commit/run above; design as one pass, not before. (vNext; the CUE Loader
      slice ships first-error only.) **Intended mechanism (decided 2026-06-16):**
      "parse don't validate" — parse becomes *total* (never fails), each failable
      field carries `value-or-Err` as data in a typed IR, and a fold-based
      `validate` walks the IR collecting errors. First-error vs all-errors is then
      a one-line change in the fold. This is also where `Tests()` finally splits
      into a total parse + validate. Don't build the IR before this lands — under
      first-error it's dead weight.
* [ ] JSONC support — deferred out of the Loader slice (Slice C). Needs either a new
      dependency or a hand-rolled string-aware comment-stripper (must skip `//` and
      `/* */` inside string literals — correctness risk on untrusted input). Decide
      dep-vs-stripper on its own merits. Weakest-value of the JSON-family formats.
* [x] Reconcile pflag's bad-flag exit code. `pflag.CommandLine` switched to
      `ContinueOnError`; `main` handles the `Parse` error itself (stderr + usage +
      `ExitUsage`), so a bad flag exits `64` instead of pflag's default `2`.
      Regression-locked by `test/tests.yml \ Tests \ Usage`.
