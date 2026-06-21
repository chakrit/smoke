# TODO

Authoritative live task state. Shipped epics (Fix `--init`, first-class CUE/JSON/JSONL
support, LLM-consumer semantics, exit-code contract, repo-as-skill) closed in `v0.3.0` —
see git history and `docs/notes/` session logs for the detail.

## Backlog (unsorted)

> **Reverted 2026-06-19.** The partial-commit / `--commit-last` / `engine.TestID` /
> "parse-don't-validate" IR work (all 2026-06-18) was ripped out as over-engineering — see
> `docs/notes/2026-06-19-revert-loader-overbuild.md`. The items below are the *reopened*
> goals, to be redone simply on the restored baseline if/when actually needed.

> **Implemented 2026-06-21** — see `docs/decisions/2026-06-21-test-name-identity-and-partial-commit.md`.
> The `TestName`/dup-name and partial-commit items below landed (`da03525` + the merge slice).

* [x] **`TestName` identity field + dup-name fail.** **Done (`da03525`).** `type TestName
  string`, minted at the flatten gate and carried `engine.Test` → `TestResult` →
  `resultspecs.TestResultSpec`; composition via `TestName.Child`, never re-minted at call
  sites. `testspecs.Load` asserts uniqueness via `map[TestName]struct{}` → duplicate is a
  load error (exit `65`). Also added `engine.Pattern`/`Filter`, retiring `internal.Whitelist`/
  `Blacklist`.
* [x] **Partial commit.** **Done.** Filtered `--commit` merges via `resultspecs.Merge`,
  walking the *spec* in order: fresh result for run tests, carry-forward by `TestName` for the
  rest. Gone-from-spec entries drop; never-committed tests stay absent (`NEW` next compare).
  The exit-64 refusal is gone.
* [ ] **Spec-filename path-dependence (bug).** The flattened root name embeds the spec
  filename, so `smoke ./x.yml` vs `smoke x.yml` yield different `TestName`s and thus different
  lock keys — cross-invocation lock-key instability. Pre-existing; partial-commit makes it more
  load-bearing (incremental lock updates: a differently-pathed partial commit silently drops
  carried entries). **Ruling required — full option analysis + recommendation (basename) in
  `docs/notes/2026-06-21-spec-filename-path-dependence.md`.**
* [ ] **Commit last run** — bless the previous run without re-running. Was a whole `runcache`
  package; only build it back if the re-run cost is actually a problem in practice. vNext;
  its own design pass.
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

* [ ] **Rebuild/redeploy after the 2026-06-21 guide change.** `docs/guides/index.md` and
  `docs/spec/architecture.md` now describe partial-commit merge (was: refusal). Run
  `scripts/build-docs.sh` + `scripts/deploy-docs.sh`. Check whether the partial-commit SVG
  diagram, if any, still depicts the old exit-64 refusal.
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
