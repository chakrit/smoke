# Include / import of spec files — design ruling

- **Date:** 2026-06-23
- **PR:** manual (design ruled in an interactive 1-by-1 walk; not yet implemented)
- **Status:** accepted

## Decision

`include: <path>` (singular scalar) on any node splices another spec file's root in as a
child — a **two-node**, file-relative model — with env flowing down through the existing
`Resolve` to power a minimal **parameterized include** via `os.Expand` in names. Full design
and the AFK build roadmap: `docs/notes/2026-06-21-include-import-spec-files.md`.

Per-decision:

- **D1 — keyword/shape.** `include`, a **singular scalar `string`**, allowed on any node,
  mutually exclusive with `tests:` (both → load error 65).
- **D2 — path base.** Resolved **relative to the including file's directory**, per hop. The
  package gains file I/O: public entry becomes `Load(filename)` (was `Load(reader, filename)`);
  a recursive dir-aware `loadSpec(path, stack)` owns reads; the per-format `loader` interface
  stays a dumb reader→tree decoder; splice is a pre-pass *before* `Resolve`.
- **D3 — naming.** File-relative, two nodes: `… \ <importing node> \ <imported root> \
  <imported tests>`; imported root's segment is the include path as written. Names interpolate
  with `os.Expand` (`$VAR`/`${VAR}`) over the node's resolved `Config.Env`; undefined → empty;
  no extensions. Source is declared env only, never `os.Environ`. SMOKE never expands commands
  — the interpreter does, at runtime.
- **D4 — lock.** Single root lock; imported tests are indistinguishable from inline after
  flatten. No per-imported-file locks.
- **D5 — inheritance.** Imported tree inherits via the existing `Resolve` parent→child,
  identical to an inline child (config merged; commands/checks prepended).
- **D6 — cross-format.** Allowed, free from `loaderFor` dispatch. (CUE schema must add the
  `include` field; JSON/JSONC follow from struct tags; JSONL works as-is.)
- **D7 — cycles/diamonds.** Ancestor-path stack (abs+clean), re-entry → 65; diamonds allowed
  by construction (distinct parent chains).
- **D8 — trust.** No path restriction (`../`, absolute paths allowed).
- **D9 — workdir.** Command `WorkDir` unchanged; no per-file cwd default.

## Rationale

Why these over the obvious alternatives — the parts that would otherwise be re-litigated:

- **Singular, not a list (D1).** A list would force one shared env across all listed files,
  defeating the parameterized-include feature; multiplicity that needs distinct params or
  names must live on the node tree anyway. "House consistency" with the plural
  `commands`/`checks` keys was explicitly rejected: those are *test data* (plural by nature);
  `include` is a *directive* (singular, like `interpreter`).
- **File-relative two-node, not structural (D3).** Structural splice (drop the imported root,
  lift its children) silently loses a single-test imported file's commands. Keeping the
  imported root as a real node loses nothing and makes D7 diamonds distinct by parent chain
  for free.
- **`os.Expand` only; do NOT invoke the interpreter for names.** A real shell can't be
  restricted to variable-expansion-only, and a homegrown expander would match no actual shell
  (bash/ash/zsh/dash), breaking the expectation that `interpreter` governs expansion.
  `os.Expand` *is* the `envsubst` subset — shell-matching for the deterministic part, in
  process, no dependency. Names are a deliberately narrower dialect than commands.
- **Determinism is NOT policed.** SMOKE is shell-native; CHANGED-on-drift is the designed
  normal state, so a non-stable name (`$(date)` etc.) is the author's call, not a bug to
  prevent. (Note: v0.4's path-dependence fix was *static path-as-typed* normalization to the
  basename — unrelated to runtime output stability; conflating the two was the original error.)
- **No own-env-in-own-name carve-out.** Names interpolate against whatever the existing merge
  produced; declining to special-case a node's own env out of its own name keeps the
  resolution logic untouched.
- **Workdir unchanged (D9).** Auto-defaulting an imported test's cwd to its file's dir breaks
  the split-suite and DRY cases (both want the shared invocation cwd) and splits inline-vs-
  included semantics; explicit `workdir:` already covers the narrow movable-fixtures case.
- **No trust boundary (D8).** A spec author already runs arbitrary shell via `commands:`;
  restricting which files `include` may read secures nothing while the real capability sits
  one field over.
