# Session log — JSON/JSONL unknown-field gap closed

Point-in-time breadcrumb. Authoritative task state is `TODO.md`. Working tree **clean**;
`main` is **level with `gh/main`** (pushed through `ec6f9b1`).

## Done

- **JSON/JSONL fail closed on unknown fields** (`47cd520`). `json.Decode` silently dropped
  typo'd keys (`chekcs:`) — the same gap the CUE closed schema already sealed for `.cue`.
  Both JSON loaders now route through a shared `decodeJSON` helper
  (`testspecs/loaders.go`) that sets `Decoder.DisallowUnknownFields`. Recurses through
  `config`/`tests[]`, error names the field, routed to exit `65` like every other
  spec-load error (`process.go` → `p.DataErr`). Three reject-tests added (top-level,
  nested-recursion, JSONL); spec parity documented in `architecture.md`.

- **TODO bookkeeping** (`ec6f9b1`). Dropped the closed backlog item.

- **Pushed.** 4 commits (2 prior TODO/note + these 2) → `gh/main`. Remote level.

## Decision (recorded here, not in decisions/)

Chose `Decoder.DisallowUnknownFields` over the TODO's alternative of routing JSON through
the CUE schema (CUE ⊇ JSON). The CUE route would couple two formats across the loader,
force JSON authors to read CUE *constraint* errors, and add eval cost to the
less-authored format — coupling dressed as DRY. Stdlib flag is the minimal fail-closed
fit. Modest enough to live in the spec + this note, not a standalone decision record.

## Open / next

- **Next session's task (user-chosen): the partial-commit pair.**
  - **#1** Allow partially committing some results but not all.
  - **#2** Commit last-run results (no re-run needed to re-commit).
  - These are coupled. **Read `architecture.md` ~line 163 first** — it documents the
    *current deliberate refusal*: "a partial set is refused — a partial commit would
    silently drop the unmatched." #1 reverses that stance, so start from why it was
    refused.
  - **#3 all-errors validation queues behind this pair.** Confirmed this session that the
    "parse don't validate" split is **unbuilt** — `Tests()` (`testspecs/test_spec.go:51`)
    is still a single recursive pass that builds the tree *and* validates inline,
    aborting on first error. The only thing named `Validate` is CUE schema validation in
    `cueLoader` — a different concern. TODO #3 note is accurate: don't build the IR before
    #1/#2 land.

- **Remaining backlog otherwise:** JSONC support (weakest value of the JSON family).
