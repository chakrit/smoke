# Run Cache and `--commit-last`

- **Date:** 2026-06-18
- **PR:** manual
- **Status:** accepted

## Decision

SMOKE persists each run's results to a per-spec cache so `--commit-last` can bless the
previous run's output without re-running the commands. The cache is bound to the spec it
observed by a content hash; a commit is refused when the spec has changed since.

## Why this is provenance, not just caching

The only way `--commit-last` goes wrong is blessing a golden that no longer matches the
spec — you edit the spec, then commit stale output. So the cached snapshot carries the
provenance needed to detect that, not merely the results:

| Field       | Purpose                                                                 |
| ----------- | ----------------------------------------------------------------------- |
| `spec_hash` | SHA-256 of the spec bytes at run time. Commit refuses on mismatch.      |
| `partial`   | Whether the run was filtered, so commit merges a subset vs. overwrites. |
| `results`   | The observed `[]TestResultSpec` — same shape as the lock.               |

`--commit-last` re-reads and re-hashes the spec, compares against `spec_hash`, and on
mismatch exits `65` (`commit-last: <spec> changed since the last run; re-run before
committing`). No cache for the spec yet → same exit with `run it once first`.

## Cache location and lifecycle

- Path: `os.UserCacheDir()/smoke/<sha256(abs-spec-path)>.yml`. Keyed by absolute spec path
  so distinct specs never collide; lives in the OS cache dir, **not** beside the spec, so
  it never litters the working tree or needs `.gitignore`.
- **Best-effort.** A cache write failure never fails the run that produced it — the cache
  is a convenience, never the source of truth. If the cache is gone, you re-run; you lose
  nothing but the shortcut.
- Written on every run (compare/commit/print) silently — no console output, so it does not
  perturb SMOKE's own observable output (the self-test stays drift-free).

## Mode exclusivity

`--commit-last` is its own mode: rejected (exit `64`) when combined with another output
mode (`--commit`/`--print`/`--list`/`--show-expected`/`--json`) or with
`--include`/`--exclude` — the run's scope is recorded in the snapshot's `partial` flag, not
re-specified at commit time.
