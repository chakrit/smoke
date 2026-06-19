# TODO

Authoritative live task state. Shipped epics (Fix `--init`, first-class CUE/JSON/JSONL
support, LLM-consumer semantics, exit-code contract, repo-as-skill) closed in `v0.3.0` —
see git history and `docs/notes/` session logs for the detail.

## Backlog (unsorted)

> **Reverted 2026-06-19.** The partial-commit / `--commit-last` / `engine.TestID` /
> "parse-don't-validate" IR work (all 2026-06-18) was ripped out as over-engineering — see
> `docs/notes/2026-06-19-revert-loader-overbuild.md`. The items below are the *reopened*
> goals, to be redone simply on the restored baseline if/when actually needed.

* [ ] **Test identity as a real field (do first).** If identity is needed beyond the display
  name, bake `ID` as a field assigned once where the flattened name is built, carried
  `engine.Test` → `TestResult` → `TestResultSpec` — not a `TestID(name)` getter. Until then,
  compare matches by `Name` (current, restored). This is the foundation for any merge work.
* [ ] **Partial commit** — let a filtered `--commit` merge onto the lock instead of being
  refused. Needs identity-keyed merge that *inserts new tests in spec order* (compare is
  order-sensitive). Blocked on the identity-field work above. Decide if it's worth it first.
* [ ] **Commit last run** — bless the previous run without re-running. Was a whole `runcache`
  package; only build it back if the re-run cost is actually a problem in practice.
* [ ] **All-errors validation** — collect every spec error per load, not just the first.
  Do it as a plain error-accumulating tree walk in `Tests()` (no IR), if anyone asks for it.
* [x] JSONC support. **Done (2026-06-18):** `.jsonc` → `jsoncLoader`, which runs a
  hand-rolled string-aware `stripJSONComments` (no new dependency) then decodes through the
  same `decodeJSON` path, inheriting `DisallowUnknownFields` fail-closed behavior. Comments
  blank to spaces (newlines kept) so `json.Decoder` error offsets stay accurate; markers
  inside string literals are preserved. Comments only — trailing commas deliberately left
  unsupported (structural editing, not stripping; revisit if someone trips on it). See
  `docs/spec/architecture.md` §"Input parsing and the CUE seam" and the guide's Advanced section.

## Docs site (`www/`)

* [x] **Deploy to GitHub Pages via gh-pages branch. Done & live (2026-06-19).**
  `scripts/deploy-docs.sh` builds, then force-pushes `www/dist` to `gh-pages` via a
  throwaway `git worktree` (dist is gitignored), `.nojekyll` included, `--dry-run` for local
  testing. Deployed; Pages serves at **http://gh.chakrit.net/smoke/** (project site under the
  user custom domain — no per-repo CNAME needed). Re-run the script after any guide change to
  refresh.
* [x] **Syntax highlighting** for the code blocks. **Done (`e461b71`):** build-time
  highlight.js via markdown-it's `highlight` option, pre-coloured `hljs-*` spans, token theme
  mapped to the site palette in `styles.css`. highlight.js is a devDependency only — JS
  bundle unchanged. `jsonl`→`json` aliased; CUE stays plain (no hljs grammar).
* [x] **Advanced section in the guide** (`docs/guides/index.md`). **Done (`46c918d`, plus
  JSONC in `eb7f2f4`):** CUE / JSON / JSONL / JSONC authoring with worked examples and
  when-to-use, surfaced as an "Advanced" sidebar section (`0958fbb`). Also fixed heading
  slugs to be GitHub-compatible so intra-doc anchors resolve on both GitHub and the site.
