# Session: www follow-ups + JSONC

- **Date:** 2026-06-19 (ran from 2026-06-18 evening; tail of the run was unattended `ace-afk`)
- **Branch:** `main` — everything below is **committed but unpushed** (held per push policy)

## What shipped

| Commit    | What                                                                          |
| --------- | ----------------------------------------------------------------------------- |
| `0663c5e` | `scripts/deploy-docs.sh` — force-push `www/dist` to `gh-pages` via a throwaway worktree (dist is gitignored); `--dry-run`, `.nojekyll`. Dry-run verified; actual push human-gated. |
| `e461b71` | Build-time syntax highlighting (markdown-it `highlight` → highlight.js). Pre-coloured `hljs-*` spans, token theme mapped to the site palette. highlight.js is a **devDependency only** — JS bundle unchanged at 647 B. |
| `46c918d` | Advanced guide section (CUE/JSON/JSONL) + **GitHub-compatible heading slugs**.   |
| `0958fbb` | Sidebar: Advanced renders as a section-header with format sub-items.            |
| `eb7f2f4` | **JSONC support** (`.jsonc`).                                                   |
| `215b232` | Housekeeping: docs-site TODOs marked done; ignore `.afk.log`.                   |

All green: `go test ./...`, `go vet`, `./test.sh` (UNCHANGED, exit 0). Docs site browser-verified throughout (highlighting in-theme, nav structure, anchors).

## Design notes worth keeping

- **GitHub-compatible slugs (`render.mjs`).** The guide is canonical markdown that *also*
  renders on GitHub. markdown-it-anchor's default slugger percent-encodes punctuation
  (`advanced%3A-…`), so an intra-doc anchor link could only resolve on one target. Replaced
  with a GitHub-style slugify (strip punctuation). All headings now match on both. Any future
  intra-doc link is safe.
- **Build-time highlighting.** highlight.js runs only in `render.mjs`; the browser gets static
  spans + a small theme. So bundle size is irrelevant and it's a pure devDep. `jsonl`→`json`
  aliased; CUE renders plain (no hljs grammar — honest, not mis-highlighted). `tab-size:2` on
  `pre` so CUE's tab indentation isn't 8-wide.
- **JSONC stripper.** `jsoncLoader` = `stripJSONComments` then the existing `decodeJSON`, so
  `DisallowUnknownFields` fail-closed parity is inherited (no new validation). The stripper is
  a pure string-aware state machine (normal/string/escape/line/block); comments blank to
  **spaces, newlines kept**, so `json.Decoder` error offsets stay accurate; comment markers
  inside string literals are preserved. Chose slurp-and-strip over a literal stateful
  `io.Reader` (offset fidelity + no chunk-boundary state + testable pure fn; spec files are
  small). **Trailing commas deliberately out of scope** — that's structural editing, not
  stripping; revisit if someone trips on it.

## Findings (resolved)

- **Partial-commit of a brand-new test mis-orders the lock — working as intended.**
  `compareTests` is an order-sensitive LCS diff (`gendiff`, keyed on `ID()`); `Merge`
  appends genuinely-new tests to the lock end, and in a partial commit the filtered overlay
  can't know a new test's spec position. Confirmed by the user: **test order is load-bearing**
  — tests are not isolated; setups and teardowns are modelled via sibling ordering — so the
  order-sensitivity is correct by design. Adding a test is a **full commit**; partial commit
  is for re-blessing existing drift. No code change. Captured in
  `docs/spec/architecture.md` §"Compare and commit".

## Open decisions for the human

- **Push** — `main` is **13 commits** ahead of `gh/main`, plus the local `v0.3.1` tag. All held.
- **gh-pages deploy** — script ready + dry-run-tested; the `push --force` to `gh-pages` and the
  one-time Pages enablement (`gh api repos/chakrit/smoke/pages …`) are outward-facing, awaiting go.
- **Cut v0.4.0** — the prior batch plus this session's new `.jsonc` format strengthen the
  minor-bump case; decide whether to push `v0.3.1` first or supersede it.
