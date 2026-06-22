# Session: include/import design fully ruled (1-by-1 walk) + AFK prep

- **Date:** 2026-06-23 (interactive). **Branch:** `main`. No code written — docs only.
- **Outcome:** the include/import feature is **fully designed and ruled**; next session is
  AFK to implement it.

## What happened

Resumed via `/ace` into the one explicitly-deferred backlog item: include/import of spec
files. Walked the open design decisions **1-by-1** (forcing-function protocol — present one,
collect the ruling, advance only on explicit "next"). Ruled D1–D8 plus a workdir-default fork
(D9). The walk surfaced a feature that wasn't in the original proposal:

- **Parameterized include.** Env flows down through the *existing* `Resolve` merge; names
  interpolate via `os.Expand` against the node's resolved env. So the same shared spec file,
  included under siblings that set different `config.env`, yields distinctly-named copies
  without editing the shared file (the `pg`/`mysql` example in the design note).

A few rulings reversed my initial recommendations after the user pushed back — recorded so the
reasoning sticks:

- **Singular scalar `include`, not a list** — `include` is a *directive*, not test-data; a
  list would force one shared env across members, defeating parameterization.
- **`os.Expand` only, never invoke the interpreter for names** — a homegrown expander matches
  no real shell; `os.Expand` is the `envsubst` subset, in-process.
- **Determinism is not policed** — SMOKE is shell-native; `$(date)` in a name is the author's
  call. (I had wrongly invoked v0.4's path-dependence fix as precedent — that was *static
  path-as-typed* normalization, unrelated to runtime output.)
- **No own-env-in-own-name carve-out; workdir unchanged.**

## Where it's recorded

- **Design + per-decision rationale + AFK roadmap (S0–S5):**
  `docs/notes/2026-06-21-include-import-spec-files.md` (every D# marked RULED).
- **Consolidated durable ruling:** `docs/decisions/2026-06-23-include-import-design.md`.
- **Live task + next-slice pointer:** `TODO.md`.
- **Resume breadcrumb + AFK launch steps:** `.tasks.md`.

## Next session (AFK)

Implement include/import. Fully designed → no approval gates → AFK-safe. TDD per slice, one
commit each on `main`, S0 (I/O refactor) → S5 (docs + fixtures). Backlog after that
(`--commit-last`/runcache, all-errors validation) each needs its own interactive design pass
first — **not** AFK-ready, don't start them blind.

## Housekeeping

- `b99b414` (v0.4.0 session-save doc) is still **unpushed** on `main`; this save's doc changes
  are uncommitted. Commit before the AFK run; push waits on user say-so.
