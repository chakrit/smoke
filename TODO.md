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
* [x] JSONC support. **Done (2026-06-18):** `.jsonc` → `jsoncLoader`, which runs a
  hand-rolled string-aware `stripJSONComments` (no new dependency) then decodes through the
  same `decodeJSON` path, inheriting `DisallowUnknownFields` fail-closed behavior. Comments
  blank to spaces (newlines kept) so `json.Decoder` error offsets stay accurate; markers
  inside string literals are preserved. Comments only — trailing commas deliberately left
  unsupported (structural editing, not stripping; revisit if someone trips on it). See
  `docs/spec/architecture.md` §"Format dispatch" and the guide's Advanced section.

## Docs site (`www/`)

* [~] **Deploy to GitHub Pages via gh-pages branch.** Script **done** (`0663c5e`):
  `scripts/deploy-docs.sh` builds, then force-pushes `www/dist` to `gh-pages` via a
  throwaway `git worktree` (dist is gitignored), `.nojekyll` included, `--dry-run` for local
  testing — dry-run verified. **Remaining (human-gated):** the actual `push --force` to
  `gh-pages` and the one-time Pages enablement (`gh api repos/chakrit/smoke/pages …`) are
  outward-facing, so they wait for a go.
* [x] **Syntax highlighting** for the code blocks. **Done (`e461b71`):** build-time
  highlight.js via markdown-it's `highlight` option, pre-coloured `hljs-*` spans, token theme
  mapped to the site palette in `styles.css`. highlight.js is a devDependency only — JS
  bundle unchanged. `jsonl`→`json` aliased; CUE stays plain (no hljs grammar).
* [x] **Advanced section in the guide** (`docs/guides/index.md`). **Done (`46c918d`, plus
  JSONC in `eb7f2f4`):** CUE / JSON / JSONL / JSONC authoring with worked examples and
  when-to-use, surfaced as an "Advanced" sidebar section (`0958fbb`). Also fixed heading
  slugs to be GitHub-compatible so intra-doc anchors resolve on both GitHub and the site.
