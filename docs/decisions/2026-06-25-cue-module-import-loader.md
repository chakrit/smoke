# CUE module `import` support — cueLoader → `cue/load`

- **Date:** 2026-06-25
- **PR:** manual (AFK build, committed to `main`)
- **Status:** accepted

## Decision

`cueLoader` loads a `.cue` spec via `cue/load.Instances` + `ctx.BuildInstance`, replacing
`ctx.CompileBytes`, so a spec inside a `cue.mod` module can `import` shared packages. This
is the enabler the `lowfat-pantry` peer asked for (2026-06-24): factor a shared `#Case`
schema + scaffold out of 64 duplicated plugin `tests.cue` behind one `cue.mod` module.
`CompileBytes` resolves only CUE stdlib builtins — never a `cue.mod` module path — so the
import failed with `package … imported but not defined`.

## Why this and not `include`

The v0.4 `include:` feature is a structural tree-splice of independently-loaded specs,
parameterized only by `os.Expand` in names. It can't carry a CUE value (a list, a
definition) across the file boundary, and each included file is loaded standalone. The
pantry's duplication is a CUE comprehension over a structured `_cases` list — that needs
CUE's *native* `import`, not a spec splice. Confirmed by the peer; the two features are
orthogonal.

## Per-decision

- **Loader contract.** The `loader` interface gains a path:
  `Load(reader io.Reader, path string)`. Byte-stream loaders (yaml/json/jsonc/jsonl) ignore
  `path` and stay dumb reader→tree decoders; only `cueLoader` reads `path`, because
  `cue/load` needs the on-disk directory to find `cue.mod`. Chosen over a type-assert
  special-case in `loadSpec` (uglier; hides the one loader that differs behind a cast).
- **Classification seam preserved.** `loadSpec` still `os.Open`s the file first, so a
  missing root (exit `2`) vs missing include (exit `65`) is classified *before* any loader
  runs. `cueLoader` therefore ignores the passed reader — the open is a readability probe,
  not its data source.
- **Absolutize the path.** `cue/load` resolves file args relative to `Config.Dir`, so a
  relative path + `Dir` doubles up (`a/b/a/b/spec.cue`). `cueLoader` calls `filepath.Abs`
  first; the arg is then self-anchored and `Dir = filepath.Dir(abs)` is where `cue.mod`
  resolution begins. This bug is invisible to absolute-temp-path unit tests; the self-test
  (relative paths through the real binary) caught it, and `TestLoadCUERelativePath` guards
  it now.
- **Backward compatibility.** A lone `.cue` with no `cue.mod` loads as a single anonymous
  instance — package-less specs are unaffected. The existing `TestLoadCUEValid` /
  `TestLoadCUERejectsUnknownField` (both package-less) pass unchanged and are the guard.
- **Hermeticity.** Resolution is local-`cue.mod` only in practice; no `cue` binary on PATH.
  `cue/load` *can* fetch remote modules, but our specs don't, so eval stays as hermetic as
  before.

## Cost

`cue/load` pulls the module machinery into the dependency tree (OCI registry, oauth2,
go-digest, image-spec) — `go.mod` gained six indirect deps via `go mod tidy`. Accepted: it
is the only path to module-aware loading, and the schema unify (`#Test`) is untouched.

## Verification

`go test ./...` green; new `TestLoadCUEModuleImport` + `TestLoadCUERelativePath`. Self-test
fixture `test/testdata/cuemod/` (shared `cases` package imported by `tests.cue`) round-trips
UNCHANGED through the real binary; the `Behavior \ CUE Module` self-test lock entry is purely
additive.
