# Session log — architecture spec + Loader re-slice

Point-in-time breadcrumb. Authoritative task state lives in `TODO.md`; resume
state in `.tasks.md`. No code changed this session.

## Done

- **`docs/spec/architecture.md`** (new, status `implemented`). As-built map:
  pipeline, the two-stage spec model (input `testspecs` vs result `resultspecs`),
  layer responsibilities, inheritance resolution, the CUE seam, checks registry,
  runner execution (interpreter `-s` over stdin; timeout→drift), compare/commit,
  modes. Written from a full read of the core files, not memory. Links the
  exit-code contract rather than restating it.

- **CUE epic re-sliced in `TODO.md`.** New Loader slice inserted ahead of the
  schema work; old "Slice C" (schema) → D, stdin → E. All-errors validation added
  to backlog, tied to the partial load/commit/run items.

## Why the Loader slice (the architecture review that prompted it)

Writing the spec surfaced three seams the extension-`switch` in `testspecs.Load`
leaves open, all of which a thin Loader closes:

1. **Dual-tag hazard** — `yaml:`+`json:` tags on `TestSpec`/`ConfigSpec` kept in
   sync by hand. (The json tags exist for CUE's decoder — same reason JSON
   support is nearly free.)
2. **`Timeout string` validated late** in `RunConfig` — a typo'd duration in a
   `.cue` surfaces at run, not load.
3. **No `.cue` schema validation** — `value.Decode` errors are opaque Go-decoder
   errors, not CUE constraints.

The user's read of the model drove the JSONL decision: since `TestSpec` is
recursive (spec ≡ test), a JSONL line *is* a `TestSpec` and the stream is just
the implicit root's children — no new concept. That collapsed what looked like a
semantic fork into a non-question.

## Decisions (kept in TODO/notes, no separate decision doc — tool is small, per
standing user direction)

- Validation: **per-format + shared**, not CUE-universal (rejected coupling
  YAML/JSON to CUE as a single validator).
- **First-error** now; all-errors deferred to vNext alongside partial
  load/commit/run.
- JSONL = children of an implicit empty root.

## Carried forward (unexecuted)

- go-coding skill amendment re: Go 1.24+ fatal non-constant printf format under
  `go test` — route via `ace-school`. Pending since the prior session.
