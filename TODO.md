# TODO

## Epic: Fix `--init`

`main.go`'s `--init` is half-baked. Tighten it into a predictable scaffolder.

* [ ] Honor the positional filename: `smoke --init foo.yml` should write `foo.yml`,
      not always `smoke-tests.yml`. Fall back to a default only when no arg is given.
* [ ] Settle the canonical default filename. README's workflow says `tests.yml`;
      `--init` writes `smoke-tests.yml`; `template.yml` is the embedded source. Pick
      one name and align README, `--init`, and examples.
* [ ] Don't clobber. `os.WriteFile` truncates silently — refuse to overwrite an
      existing file unless `--force` (or similar) is passed.
* [ ] Report the path actually written (the success message currently hard-codes
      `smoke-tests.yml` even when that's not the target).
* [ ] Add a self-test in `test/tests.yml` covering init into a temp name + the
      no-clobber guard.

## Epic: First-class CUE support

Today `testspecs.Load` is YAML-only. Make `.cue` test specs a first-class input so
specs can be generated/validated by CUE tooling. Folds in the old stdin item below.

* [ ] Decide the integration boundary: embed `cuelang.org/go` (no external binary,
      heavier dep) vs shell out to a `cue` binary on PATH. Record the call in
      `docs/decisions/`.
* [ ] Front-end dispatch by extension in `testspecs.Load` (`.cue` → evaluate to the
      test tree; `.yml`/`.yaml` → current path). Keep engine/resultspecs layers
      format-agnostic.
* [ ] Define lock-file semantics for `.cue` inputs. Results are emitted as YAML via
      `resultspecs.Save`, so `lockFilename` should likely map `foo.cue` →
      `foo.lock.yml` rather than `.lock.cue`. Resolve the ambiguity flagged in the
      old stdin note.
* [ ] Read tests from stdin to support piping from other toolings (e.g.
      `cue export | smoke -`). Decide how the lock file is located when input has no
      filename (require an explicit `--lock` path?).
* [ ] Ship a CUE schema (`#Test`/`#Config` definitions) so authors get validation and
      editor support when writing `.cue` specs.
* [ ] Self-tests: a `.cue` spec round-trips (run → commit → stable) under `test/`.

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

* [ ] Redesign the output vocabulary so it describes *drift*, not *correctness*.
      Drop `STABLE`/`Changes Detected` framing that reads as pass/fail; prefer
      neutral states (e.g. `UNCHANGED` / `CHANGED` / `NEW` / `MISSING`). Never
      emit "pass", "green", or a ✓ that implies a passing assertion.
* [ ] Audit exit-code semantics and document the contract explicitly: exit 0 means
      "matches lock", NOT "tests passed". Decide what an agent should infer and
      state it in `--help` and the README.
* [ ] Distinguish the no-lock first-run state (`NEW` / `UNREVIEWED`) from
      `UNCHANGED`, so an agent knows a human/LLM eyeball is still required before
      the golden can be trusted.
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

## Backlog (unsorted)

* [ ] Allow partially committing some results but not all.
* [ ] Allow committing last run results (so we don't have to re-run tests to commit
      again).
