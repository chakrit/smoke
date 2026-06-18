# Session log — TODO cleanup + spec audit

Point-in-time breadcrumb. Authoritative task state is `TODO.md`. Working tree is **clean**;
`main` is **1 ahead of `gh/main`** (commit `35ce0e1`, unpushed — push awaits the user's
go).

## Done

- **`TODO.md` gutted to backlog-only** (`35ce0e1`). All five epics (`--init`, CUE/JSON/JSONL
  loaders, LLM-consumer semantics, exit-code contract, repo-as-skill) were 100% `[x]` and
  closed in `v0.3.0`; stripped them plus the two done backlog items (pflag reconcile,
  `Sanity \ Loads` fix). Replaced with a one-line pointer to git history + `docs/notes/`.
  Backlog that remains: partial-commit, commit-last-run, all-errors validation, JSON/JSONL
  unknown-field gap, JSONC.

- **Spec audit — no changes needed.** Verified all `docs/spec/` current:
  `releasing.md` already reads `main` + `latest: v0.3.0`; `architecture.md` already
  documents `loaderFor`, all four formats, the `reporter` abstraction, and the known
  JSON unknown-field gap; statuses (`accepted`/`implemented`/`current`) accurate.

- **Notes left intact** by decision. The two stale "Open / next" items flagged in
  `2026-06-17-testdata-split-and-v0.3.0.md` (`Sanity \ Loads`, `releasing.md` master→main)
  are both already resolved, but notes are dated, self-disclaiming snapshots — per
  `docs/notes/README.md` they're past-thinking records, not present-state claims. The lone
  "master" string left in the repo lives inside that note as accurate history.

## Open / next

- **Push.** `main` 1 ahead of `gh/main` (the TODO cleanup). Awaiting the user's go.
- **Backlog only otherwise** — see `TODO.md`. Next natural pick is the JSON/JSONL
  silent-unknown-field gap (CUE schema closed it for `.cue`; JSON/JSONL still drop typos)
  or the partial-commit / commit-last-run pair.
