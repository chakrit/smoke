# Include / import of other spec files — design (spec only, pending ruling)

- **Date:** 2026-06-21 (unattended). **Status:** design proposal, **not implemented.**
  Decisions below need a ruling; this note is the legwork so a one-word reply per item
  unblocks implementation. Tracked in `TODO.md`.
- **Update 2026-06-22/23 (interactive 1-by-1 walk): ALL RULED.** D1–D8 plus a workdir-default
  fork all settled (see each below). The walk grew a *parameterized-include* feature
  (env-down + `os.Expand` in names). Consolidated ruling: `docs/decisions/2026-06-23-include-import-design.md`.
  Implementation is **ready for an AFK session** — roadmap at the end of this note. Next `/ace`
  should pick up there.

## Intent

Let one spec pull in another, so a suite can be split across files and shared
config/checks/commands can be defined once and reused. Today a spec is a single file;
large `tests.yml` trees and cross-suite duplication have no escape hatch. This is the
composability counterpart to the just-shipped path-dependence fix — the two share the
"relative to the root spec file" identity rule (decision **D3**).

## How it threads through the existing model

`testspecs.Load(reader, filename)` loads exactly one file → one `*TestSpec` root →
`Resolve` (value inheritance) → `Tests()` (flatten + name composition + uniqueness gate).
Includes add a **resolve-and-splice** step *before* `Resolve`: load the referenced file(s)
through `loaderFor`, splice their tree(s) into the host tree, then the existing pipeline
runs unchanged. Most of the design is deciding the splice's shape.

## Decisions

**D1 — Keyword + placement. RULED 2026-06-22: `include`, singular scalar, any node.**
Keyword is **`include`** (structural splice, not symbol/namespace `import`). The value is a
**single scalar `string`** — *not* a list. `include` is a **directive** ("instantiate this
sub-spec here, parameterized by this node's config/env"), not test-data aggregation like
`commands`/`checks`; directives are singular, and a parameterized include is inherently 1:1
(one invocation, one parameter set, one name). A list was rejected on DX: a list shares one
env across all its files, so it can't parameterize members differently — the moment you
parameterize you must split to separate named nodes anyway, and even the unparameterized
"group several files" case reads better as named sibling nodes (editorial scope names) than
a basename-named list. **Multiplicity lives on the node tree, where it can be named and
parameterized.** Allowed on **any node** (root included); mutually exclusive with `tests:`
(D3). ("House consistency" with the plural `commands`/`checks` keys was explicitly
rejected as a non-argument — those are test-data, `include` is a directive.)

**D2 — Path resolution base. RULED 2026-06-22: relative to the including file's directory,
per hop.** `a.yml`'s `include: sub/b.yml` resolves against `a.yml`'s dir no matter who
invoked `a.yml` or from what cwd — files stay self-contained and movable. **Not cwd**
(reintroduces the path-as-typed dependence v0.4 normalized away). **Not root-relative**
(a file's includes would depend on which root pulled it in). Same string is both the
locator and the D3 name segment.

**Architecture (ruled with D2).** `testspecs` crosses from pure-parsing into file I/O,
because only the resolver knows each hop's directory. Shape:

- Public entry becomes **path-based**: `Load(filename) ([]*engine.Test, error)` (was
  `Load(reader, filename)`). `process.go` loses its `os.ReadFile`/`bytes.NewReader` — the
  package owns I/O now. Stdin is already dropped, so a path-only entry has no casualty.
- A **recursive, dir-aware** core `loadSpec(path, stack)`: read `path` → decode via
  `loaderFor(path)` → walk the file's node tree → for each `include`, recurse with
  `filepath.Join(filepath.Dir(path), node.Include)` and `path` pushed on `stack` →
  splice per D3. The directory context is just `filepath.Dir(path)`, carried implicitly
  through recursion (no new state). Recursion is over the *include graph*; the per-file
  node walk finds the `include` nodes. `stack` is the cycle guard (D7).
- The per-format `loader` interface (`Load(io.Reader) (*TestSpec, error)`) stays **dumb**
  (reader → tree). Include resolution is format-agnostic, so it lives **once** in
  `loadSpec`, above the format loaders — *not* baked into the interface (that would
  duplicate the recursion across all five formats and give each an unwanted I/O dep).
- Splice is a **pre-pass before `Resolve`**, not folded into the merge. `Resolve` stays
  pure in-memory value-merging; splicing first (fallible, I/O) keeps the merge total.

**Open sub-decision (workdir default) — NOT yet settled.** Should an imported test's
command `WorkDir` (where `cmd.Dir` runs) default to *its own file's directory*, so a
movable sub-spec's relative-path commands work regardless of who included it (the same
movability principle as D2, but a runtime-semantics change to the existing `WorkDir`
merge)? Distinct from where we *find* the file. To be decided before close.

**D3 — Name-prefixing of imported tests (the load-bearing one). RULED 2026-06-22:
file-relative, two-node.**

**Ruling.** `include` on a node splices the imported file's root in as a *child* of that
node — **two nodes**, both segment-bearing: `… \ <importing node> \ <imported root> \
<imported tests>`. The imported root's default segment is its file path *as written in the
`include`* (relative to the including file — the same string used to locate it, so D2 and
D3 share one notion of "relative to the including file," composed per hop). `include` and
`tests:` are **mutually exclusive** on one node (both set → load error 65); the importing
node may still carry `name`/`config`/`commands`/`checks`.

**Parameterized include (new — emerged from the walk).** Env flows down through the
**existing** `Resolve` merge (untouched). Names support `os.Expand` (`$VAR`/`${VAR}`)
evaluated against the node's already-resolved `Config.Env`; undefined → empty (stdlib
default), no error, no defaults syntax — **`os.Expand` as-is is the entire name-expansion
contract.** Source is the spec's declared env only (ambient `os.Environ` is not pulled in;
it simply isn't in `Config.Env`). So the same imported file, included under siblings that
set different env, yields distinctly-named copies without editing the shared file:

```yaml
# root.yml                         # db.yml
tests:                             tests:
  - name: postgres                   - name: "connect-${DB}"
    config: { env: ["DB=postgres"] }     commands: ["psql ..."]
    include: db.yml
  - name: mysql
    config: { env: ["DB=mysql"] }
    include: db.yml
# → root.yml \ postgres \ db.yml \ connect-postgres
#   root.yml \ mysql    \ db.yml \ connect-mysql
```

No own-vs-inherited carve-out: a node interpolating its own declared env into its own name
is allowed (just not a promoted convention) — the source is whatever the existing merge
produced, so no special case is added to the resolution logic.

**Deliberately not done.** SMOKE never expands commands — the configured `interpreter`
does, at runtime; we do **not** invoke the interpreter to resolve names (names use
`os.Expand` only, a deliberately narrower dialect than any shell — a homegrown expander
would match no real shell and break the `interpreter`-governs-expansion expectation).
Name determinism is **not** policed: SMOKE is shell-native, CHANGED-on-drift is the
designed normal state, so a non-stable name (`$(date)`, etc.) is the author's call, not a
bug to prevent. (Note: v0.4's path-dependence fix was *static path-as-typed* normalization
to the basename — unrelated to runtime output stability.)

Original two-model framing (kept for rationale):
  - (i) **Structural** — imported file's *children* splice under the import-site node, the
    imported root dropped. **Rejected:** a single-test imported file (root has commands, no
    children) would contribute nothing — its commands silently lost.
  - (ii) **File-relative** — imported file keeps its identity as a path segment.
    **Chosen,** in the two-node form above (the imported root stays a *real* node, so
    nothing is lost; the path segment makes D7 diamonds distinct by parent chain for free).

**D4 — Lock model. RULED 2026-06-22: single root lock.** The root spec's `.lock.yml` holds
every flattened test, imported ones included, keyed by full flattened name. No
per-imported-file locks — after flatten an imported test is indistinguishable from an
inline one (just a longer name path). Natural extension of "one invocation → one lock";
the lock colocates with the root spec, as today. Two consequences, both intended:

- Editing an imported file surfaces as drift in the **root** lock — the imported suite is
  an input to the root run, so its drift is the root's drift (CHANGED → eyeball →
  re-commit, like any change).
- An imported file has no independent lock. The root run is the unit of record. To
  smoke-test an imported file standalone, run it standalone — it's a valid root spec on its
  own and gets its own basename-keyed lock in *that* invocation (with no env passed down,
  so any `${VAR}` in its names expands empty). Two invocations, two locks, no conflict.

Rejected: per-imported-file locks — fights "one invocation → one lock" and only buys
independent sub-suite re-commit, which the standalone-run path already covers.

**D5 — Inheritance. RULED 2026-06-22: full inheritance via the existing `Resolve`.** The
two-node D3 model makes the imported root a *child* of the importing node, so it inherits
through the **existing** `Resolve` parent→child, **identical to an inline child** — no new
code, no per-include carve-out. That uniformity is what powers the parameterized-include
feature (env down). Full inheritance, not just env: `config` merged/overridden as today,
`commands` and `checks` **prepended** (`append(parent.X, child.X...)`), so an importing
node that carries its own commands runs them before each imported test's — same as any
inline parent-with-commands. (The note's earlier "under (ii), no inheritance" assumed the
imported file stays its own root; the two-node form deliberately makes it a child instead.)

**D6 — Cross-format. RULED 2026-06-22: allow.** `loadSpec` dispatches every file through
`loaderFor(path)` by its own extension, so a `.yml` may `include` a `.cue`/`.json`/`.jsonl`
— each decodes via its own loader and the trees merge structurally (all `*TestSpec`). Zero
marginal mechanism; *disallowing* would take extra code for no benefit. Two implementation
must-dos:

- **CUE schema must gain `include`.** `cueLoader` unifies against the embedded `#Test`
  schema and fails closed on unknown fields, so `schema.cue` needs the `include` field
  added or no `.cue` file could use it. (The fail-closed JSON path is struct-tag driven, so
  it follows automatically — only CUE needs the hand-edit.)
- **JSONL imports work for free.** A `.jsonl` import's root is the implicit empty root; D3
  names the imported-root segment by *file path*, not the root's empty name, so it splices
  as `node \ b.jsonl \ <each line>` cleanly.

**D7 — Cycles & diamonds. RULED 2026-06-22: ancestor-stack guard, diamonds allowed.**
`stack` in `loadSpec(path, stack)` holds the **absolute paths of the current recursion's
ancestors** — pushed on descent, popped on return; it is **not** a global visited-set. That
distinction is load-bearing:

- **Cycle → load error 65.** Before loading, if the file's abs path is already on `stack`,
  it's its own ancestor → cycle, error naming the loop. Abs + `filepath.Clean` so
  `./b.yml`, `b.yml`, `../x/b.yml` to the same file collapse to one key. (Symlink aliasing
  is a residual edge — `EvalSymlinks` only if it ever bites.)
- **Diamonds allowed by construction.** `A` includes `B` and `C`, both include `D`:
  `loadSpec(D, [A,B])` and `loadSpec(D, [A,C])` — `D` is on neither ancestor stack, so no
  false cycle; two fresh subtrees. D3 names them `A \ B \ D \ …` vs `A \ C \ D \ …`,
  distinct by parent chain, so `checkUniqueNames` never fires. A *visited-set* would wrongly
  collapse the diamond (load `D` once); the ancestor-stack is the correct structure. Only a
  true cycle fails.

**D8 — Trust boundary. RULED 2026-06-22: no path restriction.** No directory jail, no
blocking `../` escapes or absolute paths — `include: ../../shared/base.yml` and
`include: /etc/smoke/common.yml` both work. A spec author already runs **arbitrary shell
commands** via `commands:`, so reading a file they named is strictly *less* powerful than
what the tool already grants — a path restriction would secure nothing while the real
capability sits one field over. Matches the shell-native, don't-babysit posture. Document
that `include` reads from the filesystem relative to the including file (D2), no sandbox.
Caveat: *if* specs ever become third-party/untrusted input (not the model today — you author
your own), this is revisited.

**D9 (workdir default) — RULED 2026-06-23: keep `WorkDir` unchanged, no per-file default.**
An imported test's command `WorkDir` (`cmd.Dir`) does **not** auto-default to its spec
file's directory. Rejected the movability-parity intuition because it breaks both headline
use cases (split-suite and DRY both want the shared invocation cwd, not each subdir),
creates an inline-vs-included asymmetry for identical commands, and conflates filesystem
layout (where files live — D2) with runtime semantics (where commands run). Authors keep
full control via explicit `workdir:`. The narrow residual gap (a sub-spec with fixtures
relative to *itself* can't say "relative to my file" — even explicit `workdir:` joins from
the invocation cwd) is left open; if it ever bites, fix with an explicit spec-dir reference,
not an implicit per-file cwd.

## AFK implementation roadmap (ruled 2026-06-23 — ready to build)

All decisions are made, so this is AFK-safe: no approval gates remain. TDD per slice (red →
green → refactor), land each slice as its own commit on `main`. If any step surfaces an
*undecided* question (not just a bug), log it and stop per the AFK envelope — but D1–D9 should
cover the design surface.

**S0 — I/O ownership refactor (pure, no behavior change).** Move file-opening into
`testspecs`: public `Load(reader, filename)` → `Load(filename string)`. Move
`os.ReadFile`/`bytes.NewReader` out of `process.go:31` into the package. Introduce internal
`loadSpec(path string, stack []string) (*TestSpec, error)` that today just reads → dispatches
`loaderFor(path)` → decodes → sets `Filename` (the single-file path, no recursion yet). `Load`
calls `loadSpec`, then `Resolve(nil)` → `Tests()` → `checkUniqueNames` unchanged. Acceptance:
every existing test green after the signature ripple; no new behavior. (This is the "extend
the entry first" step — get the recursive seam in place before includes.)

**S1 — `include` field + mutual exclusion.** Add `Include string` to `TestSpec`
(`yaml:"include" json:"include"`). Add `include?: string` to `schema.cue` (CUE fails closed —
without this no `.cue` spec can use it; JSON/JSONC follow from struct tags automatically). A
node with both `include` and `tests:` (or both `include` and `commands`? — no, commands are
allowed) → load error (exit 65). **Red:** node with `include` + `tests:` → 65.

**S2 — resolve-and-splice (the core, D3/D5).** In `loadSpec`, after decode, walk the node tree;
for each node with `include != ""`: resolve `filepath.Join(filepath.Dir(path), node.Include)`,
recurse `loadSpec(childPath, append(stack, abs(path)))`, set the imported root's `Name =
node.Include` (the path as written — the basename default won't fire since it's spliced as a
non-root child), append it as a child of the node, clear the node's `Include`. Splice runs
*before* `Resolve` (which then applies inheritance D5 verbatim). **Red:** `a.yml` includes
`b.yml` → flattened names `a.yml \ <node> \ b.yml \ <b's tests>`; env/commands/checks from the
importing node inherit into b's tests.

**S3 — name interpolation (parameterized include).** At name-mint (in `tests()`, before
`parent.Child(seg)`), run `os.Expand(seg, mapping)` where `mapping` reads the node's resolved
`Config.Env` (parse `[]string` `KEY=value`, last-wins map). Undefined → empty (stdlib default;
**not** an error — ruled, matches shell/`envsubst`). Source is `Config.Env` only, never
`os.Environ`. **Red:** the `pg`/`mysql` parameterized example → `… \ postgres \ db.yml \
connect-postgres` and `… \ mysql \ db.yml \ connect-mysql`.

**S4 — cycles, diamonds, cross-format, missing files (D6/D7).** `stack` is an ancestor stack
(abs+`filepath.Clean`, push on descent / pop on return — NOT a visited-set). File already on
stack → cycle error 65 naming the loop. **Reds:** cycle → 65; diamond (A→B→D, A→C→D) → both
copies present with distinct names, no `checkUniqueNames` fire; `.yml` includes `.cue`/`.jsonl`
→ works; missing referenced file → 65.

**S5 — docs + self-test fixtures.** `docs/spec/architecture.md`: load pipeline gains the
resolve-and-splice pre-pass (decode → splice → `Resolve` → `Tests`); note I/O now lives in
`testspecs`. Guide Advanced section: `include` + parameterized-include worked examples. Add
`test/testdata/` fixtures exercising include (stable: re-commit; keep the intentional-drift
fixtures' pattern in mind per the v0.4 migration lesson). Lock model is single-root (D4) — no
code change; verify a committed include suite round-trips UNCHANGED.

**Touch list:** `testspecs/testspecs.go` (Load + loadSpec + splice), `testspecs/test_spec.go`
(`Include` field), `testspecs/loaders.go` (none — interface unchanged), `testspecs/schema.cue`
(`include?`), `process.go:31` (call-site), `docs/spec/architecture.md`, `docs/guides/index.md`,
`test/testdata/` + `test/tests.yml`. Exit codes: all include load failures → 65 (`p.DataErr`),
consistent with the standing malformed-spec contract.
