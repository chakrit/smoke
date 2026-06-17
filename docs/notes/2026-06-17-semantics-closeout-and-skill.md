# Session log — semantics epic closeout + SKILL.md

Point-in-time breadcrumb. Authoritative task state is `TODO.md`. This session landed two
commits on `main` (now **16 ahead of `gh/main`, unpushed** — push awaits the user's
say-so).

## Milestone

**All five epics are closed.** Only the backlog remains (partial commit, commit-last-run,
all-errors validation, JSON/JSONL unknown-field gap, JSONC).

## Done

- **Semantics epic closed by verification** (`df7cc96`). The three "open" items were
  already satisfied by earlier slices but never checked off:
  - *Human-output + `--help` framing* — `usageHeader` (`main.go` 36–37) covers `--help`;
    every verdict line in `internal/p/funcs.go` (`Unchanged`/`Changed`/ `New`) carries the
    inline drift-vs-correct framing. Verified, boxed.
  - *LLM-facing TDD guidance* — already lives in `docs/spec/using-smoke-in-tdd.md` ("three
    traps" + "branch on the exit code"). Boxed; it is the reuse source for SKILL.md.
  - *Vocabulary decision record* — promoted from a Scope-section pointer to a first-class
    **"Vocabulary contract"** amendment in
    `docs/decisions/2026-06-08-exit-code-contract.md` (state→code→string table, the
    pass/green/✓ language ban, the `--json` mirror; frozen on the same terms as the
    codes).

- **SKILL.md shipped** (`d6ad1ea`). Closes the "Repo doubles as an installable skill"
  epic. Root-level `SKILL.md` (the repo *is* the skill — loader finds it at the skill-dir
  root).
  - Front-matter `description` front-loads the drift≠correctness framing and an explicit
    **DO-NOT-TRIGGER** on assertion frameworks (go test/pytest/jest) — the exact misread
    the semantics epic exists to prevent. The description is the triggering lever, so the
    effort went there.
  - Body: setup → run→eyeball→commit → exit-code table → the three traps → agent/`--json`
    guidance, in terse imperatives (house-style override of skill-creator's why-clauses).
    Links `docs/spec/{exit-codes,using-smoke-in-tdd}.md` rather than duplicating. **No
    `references/` subtree** — the specs are in-tree.
  - Verified: adding `SKILL.md` drifts no self-test (globs are `*.go` /`*.mod`, no
    root-dir listing); `smoke test/tests.yml` → UNCHANGED; `go test ./...` ok.

## Open / next

- **Push.** 16 commits ahead of `gh/main`, unpushed. Awaiting the user's go.
- **Optional — SKILL.md description optimization.** skill-creator ships a trigger eval +
  `run_loop.py` description optimizer (needs the `claude` CLI). Not run this session.
  Worth a pass if SKILL.md under/over-triggers in practice; the hand-written description
  is a solid first cut, not benchmark-tuned.
- **Backlog only otherwise** — see `TODO.md`. Next natural pick is the JSON/JSONL
  silent-unknown-field gap (the CUE schema closed it for `.cue`; JSON still drops typos)
  or the partial-commit / commit-last-run pair.
