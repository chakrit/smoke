# TODO

Authoritative live task state. Shipped epics (Fix `--init`, first-class CUE/JSON/JSONL
support, LLM-consumer semantics, exit-code contract, repo-as-skill) closed in `v0.3.0` —
see git history and `docs/notes/` session logs for the detail.

## Backlog (unsorted)

* [ ] Allow partially committing some results but not all.
* [ ] Allow committing last run results (so we don't have to re-run tests to commit again).
* [x] All-errors validation reporting — collect every spec error per load, not just the
  first. **Done (2026-06-18):** "parse don't validate" landed in `testspecs/test_ir.go`.
  `parse` is total (value-or-error `parsed[T]` carriers, command-less leaves become
  `leafError`); `validate` folds the flat IR collecting every error in depth-first spec
  order via `errors.Join`, flowing out through `testspecs.Load` → exit `65`. First-error
  vs all-errors is a one-line `continue`→`break` change in the fold. `Tests()` is now just
  `validate(parse(t))`. See `docs/spec/architecture.md` §"Inheritance resolution".
* [ ] JSONC support — deferred out of the Loader slice. Needs either a new dependency or a
  hand-rolled string-aware comment-stripper (must skip `//` and `/* */` inside string
  literals — correctness risk on untrusted input). Decide dep-vs-stripper on its own
  merits. Weakest-value of the JSON-family formats.
