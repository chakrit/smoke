# Spec-filename path-dependence — analysis + resolution

- **Date:** 2026-06-21 (analysis), resolved same night (unattended).
- **Status:** **Resolved — Option A (basename), framed as "relative to the root spec
  file."** Shipped. See the Resolution section at the bottom; the analysis below is
  retained as the reasoning of record.

## The bug

The flattened root `TestName` is the spec path **as typed on the command line**.
`testspecs/testspecs.go:22` sets `root.Filename = filename` (the raw CLI arg), and
`testspecs/test_spec.go:35` defaults an unnamed root to `t.Name = t.Filename`. The
flatten walk then prefixes every test with that root segment.

The persisted lock key carries it verbatim — from this repo's own
`test/tests.lock.yml`:

    - name: test/tests.yml \ Builds

So the **same spec file** yields different lock keys depending only on how the path
was typed:

| Invocation                  | Root key fragment      |
| --------------------------- | ---------------------- |
| `smoke test/tests.yml`      | `test/tests.yml \ …`   |
| `cd test && smoke tests.yml`| `tests.yml \ …`        |
| `smoke ./test/tests.yml`    | `./test/tests.yml \ …` |

The lock file is colocated with the spec regardless (`lockFilename` only swaps the
extension, keeps the path), so all three invocations read/write the **same**
`test/tests.lock.yml` but under **different keys**.

## Why partial-commit makes it load-bearing

Full commit rewrites the whole lock, so a path change just re-keys everything in one
shot — annoying but self-consistent. Partial commit (`resultspecs.Merge`) carries
unselected entries forward **by name**. If the partial commit is typed with a
different path prefix than the original full commit, the carried names don't match
the fresh names, and `Merge` **silently drops** every carried entry (it only emits
names present in the current `order`). That's quiet data loss in the lock — the
sharpest edge of this bug.

## Options

| Option | Change | Fixes | Migration (existing locks re-key) |
| ------ | ------ | ----- | --------------------------------- |
| **A — basename** | `filepath.Base(filename)` at the gate | cwd, `./`, abs-vs-rel all collapse to `tests.yml` | yes — every committed key re-keys once (`test/tests.yml \ X` → `tests.yml \ X`) |
| **B — clean only** | `filepath.Clean(filename)` | only `./x` → `x`; leaves cwd and abs-vs-rel broken | only `./`-typed invocations |
| **C — sentinel root** | unnamed root → `""` or `"root"` | fully path- **and** filename-independent | yes — keys lose the spec segment entirely |
| **D — accept + document** | none; doc "invoke with a consistent path" | nothing | none |

## Recommendation — A (basename)

Fixes every load-bearing case (different cwd is the realistic one), keeps a
meaningful display/lock name, and renaming a spec changing identity is *correct* — a
renamed spec is a different spec. One-time migration: existing locks re-commit once
(this repo's `test/tests.lock.yml` re-keys cleanly; `./test.sh` would report the
tests as NEW until re-committed).

- **Not B:** half-fix — leaves the realistic different-cwd case broken, so it gives
  false closure.
- **Not C:** drops the spec name from output, and globbing two specs that each have a
  top-level test of the same name would now collide → uniqueness check fires
  (exit 65). A regression for multi-spec runs.
- **Collision note for A:** two specs with the same basename only collide inside a
  **single** globbed invocation (`smoke a/t.yml b/t.yml`); separate invocations keep
  separate colocated locks, no collision. The single-invocation case is already a
  uniqueness-check (65) situation, so A doesn't make it worse.

## If A is chosen — locus + test plan

- **Locus:** `testspecs/testspecs.go:22` — set `root.Filename = filepath.Base(filename)`
  (or normalize at the `t.Name = t.Filename` default in `test_spec.go:35`). One line.
  `filepath` is already imported in `testspecs.go`.
- **Test:** `testspecs` table test — same spec bytes loaded under `x.yml`,
  `./x.yml`, and `dir/x.yml` must yield identical root `TestName`s (red first).
- **Migration:** re-run `./test.sh` with `--commit` once to re-key
  `test/tests.lock.yml`; eyeball the diff (only the root segment of each key should
  change), then commit the re-keyed lock alongside the fix.

> The ruling (A / B / C / D) is the user's. This note is the legwork so a one-word
> reply unblocks it.

## Resolution — A (basename), framed as "relative to the root spec file"

The user ruled: identity should be the spec path **relative to the root spec file**. For
a single root spec that *is* its basename, so the implementation is Option A's one-liner —
but the framing is the general rule, and it's the one the import/include feature will
extend (an imported spec's tests get prefixed by that file's path relative to the **root
spec's directory**; the root itself, relative to its own directory, is the basename).

- **Locus:** `testspecs/test_spec.go` — the unnamed-root name default became
  `t.Name = filepath.Base(t.Filename)`. Normalized the **name**, not `Filename`: the raw
  path stays available for lock colocation (`process.lockFilename`) and for resolving
  relative imports later. Red test `TestLoadRootNameIsBasename` pins `x.yml` / `./x.yml` /
  `dir/x.yml` / `../x.yml` / `/abs/dir/x.yml` to one `x.yml \ …` root.

- **Migration was larger than this note first assumed.** It cascades to **every** lock in
  the repo, not just `test/tests.lock.yml`: the self-test drives fixtures under
  `test/testdata/`, each with its **own** colocated lock keyed under the old typed path
  (`test/testdata/X.yml \ …`). The basename fix re-keys the fixtures too, so every fixture
  lock lookup missed until re-keyed — stable fixtures flipped UNCHANGED→CHANGED and drift
  fixtures reported a name-mismatch shape instead of their intended content drift.
  - **5 stable fixtures** (`stabletests`, `cuetests`, `jsontests`, `jsonltests`,
    `jsonctests`) → re-committed (`smoke --commit`); pure re-key, still UNCHANGED.
  - **2 intentional-drift fixtures** (`badtests`, `timeouttests`) → **key-line-only**
    rename. Re-committing would have erased the baseline drift the self-test asserts.
    A global substitution was unsafe: `badtests.lock.yml` carries `test/testdata/badtests.txt`
    in *command values*, which must stay verbatim — only the `- name:` keys were edited.
  - **`test/tests.lock.yml`** regenerated last, snapshotting the now-consistent fixture
    outputs.
  - **Acceptance gate:** `normalize(old) == normalize(new)` where `normalize` collapses the
    known path prefixes — empty residual proved the *only* change repo-wide was the basename
    collapse (no network/timing/toolchain noise), and `smoke test/tests.yml` exits 0.
