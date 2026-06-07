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

## Backlog (unsorted)

* [ ] Allow partially committing some results but not all.
* [ ] Allow committing last run results (so we don't have to re-run tests to commit
      again).
