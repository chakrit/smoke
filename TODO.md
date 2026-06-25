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
* [x] **Spec-filename path-dependence (bug). Done.** Ruling: root identity is the spec's
  basename (`filepath.Base`) at the unnamed-root name default in `testspecs/test_spec.go` —
  framed as "relative to the root spec file" (for a single root spec, that *is* its basename;
  imported specs will extend the rule against the root's directory). `cwd` / `./` / abs-vs-rel
  all collapse to one stable key. Migration cascaded to **every** lock in the repo (each
  fixture has its own colocated lock keyed under the old typed path): 5 stable fixtures
  re-committed, 2 intentional-drift fixtures (`badtests`, `timeouttests`) key-renamed in place
  to preserve their baseline, self-test lock regenerated. Analysis + resolution in
  `docs/notes/2026-06-21-spec-filename-path-dependence.md`.
* [x] **Include / import other spec files. DONE (AFK 2026-06-23, S0–S5).** Singular scalar
  `include: <path>` on any node, two-node file-relative splice, env-down + `os.Expand`
  parameterized names, single root lock, ancestor-stack cycle guard. Shipped exactly per the
  design (`docs/notes/2026-06-21-include-import-spec-files.md`) and ruling
  (`docs/decisions/2026-06-23-include-import-design.md`). Two design refinements made during
  build (both recorded in `docs/notes/2026-06-23-include-import-landed-afk.md`): (1) S0's
  "no behavior change" *required* a `testspecs.SpecError` marker — the frozen exit-code
  contract pins a missing **root** spec at 2, so only a missing **included** file is 65;
  (2) the imported-root segment defaults to the include path only when the imported file
  names no root of its own (D3 "default", not the roadmap's unconditional set). `testspecs`
  now owns spec file I/O. Self-test fixture: `test/testdata/include/`.
* [x] **All-errors validation. Done (2026-06-24).** `TestSpec.tests()` accumulates every
  tree-walk fault — all unknown checks per node, bad timeouts, command-less leaves — into
  `[]error`; `Tests()` joins via `errors.Join`. One `Load` now surfaces all spec faults at
  once. Plain walk, no IR — the cheap version of the 2026-06-18 work reverted on
  2026-06-19. Scope: the flatten walk only; the uniqueness pass stays first-dup (running it
  on a partial flatten would false-positive), and `loadSpec` parse errors abort their file
  by nature.
* [x] **CUE module `import` support (cueLoader → `cue/load`). Done (2026-06-25).**
  `cueLoader` loads via `cue/load.Instances` + `ctx.BuildInstance` (was `ctx.CompileBytes`,
  stdlib-only), so a `.cue` inside a `cue.mod` module can `import` shared packages — the
  `lowfat-pantry` DRY ask. The `loader` interface gained a `path` param
  (`Load(reader, path)`); byte loaders ignore it, `cueLoader` needs the on-disk dir.
  `loadSpec` still opens the file first (exit-2-vs-65 classification preserved). Path is
  absolutized before `load.Instances` — cue/load resolves args relative to `Config.Dir`, so
  a relative path + Dir doubled up (caught by the self-test, not unit tests). Package-less
  `.cue` unchanged (single anonymous instance). Design ruling:
  `docs/decisions/2026-06-25-cue-module-import-loader.md`. Self-test fixture
  `test/testdata/cuemod/`. New deps: cue/load pulls module machinery (OCI, oauth2).
* [x] JSONC support. **Done (2026-06-18):** `.jsonc` → `jsoncLoader`, which runs a
  hand-rolled string-aware `stripJSONComments` (no new dependency) then decodes through the
  same `decodeJSON` path, inheriting `DisallowUnknownFields` fail-closed behavior. Comments
  blank to spaces (newlines kept) so `json.Decoder` error offsets stay accurate; markers
  inside string literals are preserved. Comments only — trailing commas deliberately left
  unsupported (structural editing, not stripping; revisit if someone trips on it). See
  `docs/spec/architecture.md` §"Input parsing and the CUE seam" and the guide's Advanced section.

## Docs site (`www/`)

* [x] **Rebuild/redeploy after the guide changes. Done (2026-06-22).** Deployed via
  `scripts/deploy-docs.sh` (force-pushed `gh-pages` `d6e524f`) after the partial-commit
  merge + basename-identity guide edits. No SVG diagram exists to update. **Note:** the old
  `gh.chakrit.net` custom domain (account-level) no longer points to GitHub; user is
  disabling repo Pages (`gh api -X DELETE repos/chakrit/smoke/pages`). Revisit hosting if
  the docs site is wanted again.
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
