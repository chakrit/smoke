# Include / import of other spec files — design (spec only, pending ruling)

- **Date:** 2026-06-21 (unattended). **Status:** design proposal, **not implemented.**
  Decisions below need a ruling; this note is the legwork so a one-word reply per item
  unblocks implementation. Tracked in `TODO.md`.

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

**D1 — Keyword + placement.** Recommend **`include`** (structural splice, like a YAML/`#include`
merge — not symbol/namespace `import`), accepting a string or list. Allowed on **any test
node**, not just the root: a node with `include: other.yml` gains that file's tree as
children, exactly where inline `tests:` children would sit. (Minimal first cut: root-only.
Lean any-node — strictly more general, composes with inheritance.)

**D2 — Path resolution base.** Recommend **relative to the including file's directory**
(each file self-contained and movable; `a.yml`'s `include: sub/b.yml` resolves against
`a.yml`'s dir no matter who invoked `a.yml`). **Not cwd** — that reintroduces exactly the
path-dependence we just killed. Not root-relative — breaks file movability.

**D3 — Name-prefixing of imported tests (the load-bearing one).** Two models:
  - **(i) Structural** — the imported file's *children* splice under the import-site node;
    the imported file's own root name is dropped. `a.yml` test `Foo` with `include: b.yml`
    → `a.yml \ Foo \ <b's tests>`. Consistent with how inline children already name.
  - **(ii) File-relative** — the imported file keeps its identity, prefixed by its path
    **relative to the root spec's directory**: `a.yml \ b.yml \ …` (or `a.yml \ sub/b.yml \ …`).
    This is the literal reading of "relative to the root spec file" from the path-deps note.

  Recommend **(i) structural** — names stay about the logical test hierarchy, not the
  filesystem layout, and it reuses the existing inheritance/flatten rules verbatim. Flag:
  the path-deps framing hints the user may want **(ii)**. **This is the call to make first;
  D4/D5 follow from it.**

**D4 — Lock model.** Recommend a **single root lock** — the root spec's `.lock.yml` holds
every flattened test, imported ones included, keyed by flattened name. No per-imported-file
locks. Natural extension of "one invocation → one lock"; imported tests are indistinguishable
from inline ones after flatten.

**D5 — Inheritance.** Under D3(i), imported children **inherit the import site's
config/checks/commands**, identical to inline children (`Resolve` parent→child). Under
D3(ii) the imported file would more naturally start as its own root (no inheritance). Pick
with D3.

**D6 — Cross-format.** Recommend **allow** — each referenced file resolves through its own
`loaderFor` (extension dispatch already exists), so a `.yml` may `include` a `.cue`/`.json`.
Low marginal cost; trees merge structurally after load.

**D7 — Cycles & diamonds.** Detect cycles via an **absolute-path stack** across the resolve
recursion; re-entry → load error (exit 65), naming the cycle. **Diamonds allowed**: A
includes B and C, both include D → D's tests appear under two distinct name prefixes, so the
uniqueness gate doesn't fire (different parents → different flattened names). Only true
cycles fail.

**D8 — Trust boundary.** No path restriction. A spec author already runs arbitrary commands
via the spec; reading a file they named is strictly less powerful. Document that `include`
reads from the filesystem relative to the including file. (If specs ever become
third-party/untrusted, revisit — but that's not the model today.)

## Implementation sketch (when ruled)

- New resolver in `testspecs`: `resolveIncludes(root *TestSpec, dir string, stack []string)`
  — walks the tree, for each node with `include`, loads each referenced file (relative to
  `dir`), guards `stack` for cycles, splices per D3, recurses with the referenced file's dir.
  Runs in `Load` between decode and `Resolve`.
- `TestSpec` gains an `Include []string` field (`yaml:"include" json:"include"`), stripped
  from the tree by the resolver before `Resolve`/`Tests` see it.
- `loaderFor` already dispatches by extension — reused as-is for referenced files.
- Tests: relative-path resolution from a nested dir; cycle → 65; diamond → both copies
  present, names distinct; cross-format include; missing referenced file → 65.
- Docs: `docs/spec/architecture.md` (the load pipeline) + the guide's Advanced section.

Not trivial — multiple interacting decisions (D3 governs D4/D5) and a real recursive
resolver with cycle handling. Logged and stopped per instruction; awaiting a ruling on
D1–D8 (D3 first).
