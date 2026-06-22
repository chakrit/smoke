# Session: include/import shipped (AFK, S0–S5)

- **Date:** 2026-06-23 (unattended / AFK). **Branch:** `main`.
- **Outcome:** the include/import feature is **implemented, tested, documented, and
  committed** — the whole S0–S5 roadmap from
  `docs/notes/2026-06-21-include-import-spec-files.md`, one commit per slice. Self-test
  UNCHANGED throughout; the exit-code contract preserved.

## What landed

`include: <path>` (singular scalar) on any node splices another spec file's root in as a
child (two-node, file-relative), env flows down through the existing `Resolve`, and
imported test names interpolate that env via `os.Expand`. Commits (oldest first):

- **S0** `6443859` — own spec file I/O: `Load(reader, filename)` → `Load(filename)`;
  recursive `loadSpec(path, stack)` seam; `testspecs.SpecError` classifier.
- **S1** `c6139cf` — `Include` field + CUE `include?` + include⊕`tests:` exclusion (65).
- **S2** `9b54ba0` — resolve-and-splice core (D2 file-relative, D3 two-node, D5 inherit).
- **S3** `8ad453d` — parameterized names via `os.Expand` over resolved `Config.Env`.
- **S4** `1935ae8` — ancestor-stack cycle guard; diamond / cross-format / missing tests.
- **S5** `3546354` — docs (`architecture.md` Include-resolution section + guide Advanced)
  and a real self-test fixture (`test/testdata/include/`) that round-trips UNCHANGED.
- **audit** `ce531f3` — dropped a stale `loadSpec` comment.

## Two refinements vs the written roadmap

Both were forced by the existing code/contract, not free choices — recorded so they aren't
mistaken for drift from the ruling.

1. **S0 was not a pure no-op; it needed `SpecError`.** The frozen exit-code contract
   (`docs/decisions/2026-06-08-exit-code-contract.md`, 2026-06-16 amendment) pins a
   **missing/unreadable root spec at exit 2** — and the self-test asserts it via
   `nonexistent.yml` (`Exit states \ Operational`). The naive S0 (move `os.ReadFile` in,
   wrap every `Load` error in `dataErr`) would have flipped that to 65. So `testspecs` now
   marks malformed-spec failures (`SpecError` → 65) and returns a *bare* I/O error for a
   root that won't open (→ 2); `loadSpec` uses **stack depth** to classify a file-open
   failure (root vs included). Verified end-to-end: missing root → 2, missing include →
   65, cycle → 65.

2. **Imported-root segment default is conditional.** The roadmap's S2 said set the
   imported root's `Name = node.Include` unconditionally. But an imported file spliced as
   a *child* skips both the root-basename default and the path-name default in `Resolve`
   (parent != nil), so an empty name would yield an empty segment. The fix matches D3's
   actual wording ("**default** segment"):
   `if childRoot.Name == "" { childRoot.Name = node.Include }` — an imported file that
   names its own root keeps that name, like the root-basename rule.

## Design surface held

No *undecided* questions surfaced during the build — D1–D9 covered everything. The only
deviations were the two mechanical refinements above, both consistent with the ruling. The
`slices.Clone` per descent (not a shared append) is what keeps a diamond from a false
cycle.

## Verification

- `go test ./...` green (11 new include tests in `testspecs/include_test.go`).
- `goimports -l` clean; self-test (`test/tests.yml`) UNCHANGED, exit 0.
- New self-test node `Behavior \ Include` exercises it through the real binary; its lock
  entry is purely additive (verified by `git diff` of `test/tests.lock.yml`).

## Backlog after include (unchanged — each still needs its own interactive design pass)

- **Commit last run** (`--commit-last`/runcache) — bless the prior run without re-running.
  vNext.
- **All-errors validation** — collect every spec error per load, not just the first.

Both are undecided in design — **not** AFK-ready; don't start them blind.
