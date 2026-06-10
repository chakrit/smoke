# Session log — exit-code reconcile + first-class CUE

Point-in-time breadcrumb. Authoritative task state lives in `TODO.md`.

## Shipped

- **`fb9d2c4` — pflag bad-flag exit 64.** `pflag.CommandLine` now runs in
  `ContinueOnError`; `main` handles the parse error itself (stderr + usage +
  `ExitUsage`) instead of letting pflag `os.Exit(2)`. Closes the last gap in the
  exit-code contract. Regression-locked by `test/tests.yml \ Tests \ Usage`.
- **`d556421` — first-class `.cue` specs.** `testspecs.Load` dispatches on
  extension: `.cue` → embedded `cuelang.org/go` eval → `TestSpec` (via new
  `json` tags) → existing `Resolve`/`Tests`. `.cue` locks to `.lock.yml`.
  Unknown extensions rejected (default-deny). Round-trip locked by
  `test/cuetests.cue` + `Tests \ CUE`.

## Decisions made (no separate decision doc — by user direction, tool is small)

- **Embed `cuelang.org/go`** over shelling out to a PATH `cue` — pins the
  evaluator in `go.mod`, keeping `.cue` eval as hermetic as the YAML parser.
- **`ConfigSpec.Timeout` is now a `string`** ("5s"), parsed in `RunConfig`. CUE
  has no duration kind, so both loaders share one string path; dropped the
  `*time.Duration` field + `resolveDurations` helper.
- Forced cleanup: cuelang requires **`go 1.25`**, under which `go vet`'s
  non-constant-format-string check is fatal in `go test`. `internal/p`'s
  `output`/`outputErr` were latent printf wrappers fed dynamic strings — now
  plain literal writers; the one formatted callsite uses a constant format.

## Next — CUE epic remaining (see TODO.md "First-class CUE support")

- **Slice C (next): ship `#Test`/`#Config` CUE schema + validate `.cue` on
  load.** Embed `schema.cue` via `go:embed`, unify the user's value against
  `#Test` before `Decode` so invalid specs get clean CUE errors. Load
  `cue-coding` skill. Add a self-test for a schema-invalid `.cue` erroring
  cleanly (decide exit code: usage `64` vs trouble `2`).
- **Slice D (deferred, separable): stdin** — `smoke -` + `--lock` for the lock
  path when there's no filename.

## Other open work (unchanged this session)

- `--json` machine-readable mode (mirrors exit code in a `status` field) —
  unblocks the remaining exit-code + LLM-semantics epic items. User wants this
  *after* CUE.
- `SKILL.md` (repo-as-installable-skill epic) — self-contained, reuses
  `docs/spec/using-smoke-in-tdd.md` + `exit-codes.md`.
- Backlog: partial commit; commit last-run results.

## Follow-up to route (not yet done)

- **go-coding skill candidate:** "Go 1.24+ makes non-constant printf format
  strings a fatal `go vet` error under `go test`; don't pass dynamically-built
  strings as the format arg to printf-family wrappers." Generic Go gotcha that
  bit us here — worth an `ace-school` amendment to `go-coding`. Surfaced to user;
  not executed (school PR is a deliberate separate action).
