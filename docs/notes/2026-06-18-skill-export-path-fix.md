# Session log — skill export path fix (root → skills/smoke/)

Point-in-time breadcrumb. Authoritative task state is `TODO.md`.

## Problem

`ace school pull` on `chakrit/smoke` ingested **all 79 files** (`main.go`, `engine/`,
`internal/`, `testspecs/`, the 433KB `smoke.jpg`, the compiled `bin/smoke`) into the
school. Cause: `SKILL.md` lived at the **repo root**, and ACE scopes a skill to the
directory containing its `SKILL.md` — so root placement makes the whole repo one skill
dir. The school side had to hand-strip its copy to `SKILL.md` only, and every re-pull
re-polluted it.

## Fix (commit `4817977`)

`git mv SKILL.md skills/smoke/SKILL.md`. ACE then imports only that subtree.

Precedent from chakrit's own repos settled the topology:

- **`chakrit/kien-thai`** — a *source* repo (corpus, evals, tests, `pyproject.toml`) that
  also ships skills. Nests them under `skills/kien-thai/`, `skills/kode-thai/`. Source
  never gets swept. **This is the pattern smoke now follows.**
- **`chakrit/fact-check`** — a *skill-only* repo; root-level `SKILL.md` is fine there
  because there is nothing else to pollute.

Rule: **a skill living in a source repo must nest under `skills/<name>/SKILL.md`.** Root
placement is only safe for skill-only repos.

## Why nothing else changed

- Body doc links are absolute GitHub URLs (`.../blob/main/docs/spec/...`), not relative —
  the move breaks no links.
- Self-test globs are explicit (`*.go`, `*.mod`, `checks/*.go`, …); no `*.md` or `skills/`
  glob. `smoke test/tests.yml` → UNCHANGED after the move, `go build` clean.

## School-side import

`[[imports]] source = "chakrit/smoke", skills = ["smoke"]`. The school agent
(`prod9.school.claude`) was told via the ace-connect bridge to drop its manual strip and
re-pull **after the push lands**.

## Open

- **Push.** Commit is local-only; the school can't re-pull until `chakrit/smoke` is pushed
  to `gh`. Awaiting chakrit's go. Ping `prod9.school.claude` once pushed.
