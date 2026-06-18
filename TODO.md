# TODO

Authoritative live task state. Shipped epics (Fix `--init`, first-class CUE/JSON/JSONL
support, LLM-consumer semantics, exit-code contract, repo-as-skill) closed in `v0.3.0` —
see git history and `docs/notes/` session logs for the detail.

## Backlog (unsorted)

* [x] Allow partially committing some results but not all. **Done (2026-06-18):**
  `engine.TestID` centralizes test identity (duplicates rejected at load); a filtered
  `--commit` merges the observed subset onto the existing lock by identity
  (`resultspecs.Merge`) instead of being refused, preserving unrun tests. Unfiltered
  commit still overwrites wholesale. See `docs/spec/architecture.md` §"Compare and commit".
* [x] Allow committing last run results (so we don't have to re-run tests to commit again).
  **Done (2026-06-18):** each run persists a provenance-stamped snapshot (`runcache`);
  `--commit-last` blesses it without re-running, refusing (exit 65) if the spec changed
  since. See `docs/decisions/2026-06-18-run-cache-and-commit-last.md`.
* [x] All-errors validation reporting — collect every spec error per load, not just the
  first. **Done (2026-06-18):** "parse don't validate" landed in `testspecs/test_ir.go`.
  `parse` is total (value-or-error `parsed[T]` carriers, command-less leaves become
  `leafError`); `validate` folds the flat IR collecting every error in depth-first spec
  order via `errors.Join`, flowing out through `testspecs.Load` → exit `65`. First-error
  vs all-errors is a one-line `continue`→`break` change in the fold. `Tests()` is now just
  `validate(parse(t))`. See `docs/spec/architecture.md` §"Inheritance resolution".
* [ ] JSONC support — deferred out of the Loader slice. Needs either a new dependency or a
  hand-rolled string-aware comment-stripper (must skip `//` and `/* */` inside string
  literals — correctness risk on untrusted input). Decide dep-vs-stripper on its own
  merits. Weakest-value of the JSON-family formats.

## Docs site (`www/`) — next session

* [ ] **Deploy to GitHub Pages via gh-pages branch.** Build output (`www/dist`) is
  gitignored, so a plain `git subtree push --prefix www/dist` won't work (subtree needs the
  prefix tracked on the source branch). Write `scripts/deploy-docs.sh`: build, then push
  `www/dist` to `gh-pages` via a throwaway `git worktree` on that branch (or `git subtree
  split` of a temporarily-committed dist) and force-push. Then enable Pages on `gh-pages`.
  No GitHub Actions. Confirm relative asset paths work on the Pages sub-path (`--public-url
  ./` already set).
* [ ] **Syntax highlighting** for the code blocks. Prefer build-time over a runtime CDN
  script: a markdown-it highlighter (e.g. `markdown-it`'s `highlight` option backed by
  highlight.js or Shiki) so the HTML ships pre-coloured. Add the theme CSS to
  `www/src/styles.css`. Plain JS only, no TypeScript.
* [ ] **Advanced section in the guide** (`docs/guides/index.md`): authoring specs in CUE
  (`.cue` — the embedded `#Test`/`#Config` schema, closedness/fail-closed behaviour) and in
  JSON/JSONL (`.jsonl` = one `TestSpec` per line; `DisallowUnknownFields`). Show a worked
  example of each and when to reach for them over YAML. Rebuild the site after.
