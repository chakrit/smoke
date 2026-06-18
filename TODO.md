# TODO

Authoritative live task state. Shipped epics (Fix `--init`, first-class CUE/JSON/JSONL
support, LLM-consumer semantics, exit-code contract, repo-as-skill) closed in `v0.3.0` —
see git history and `docs/notes/` session logs for the detail.

## Backlog (unsorted)

* [ ] Allow partially committing some results but not all.
* [ ] Allow committing last run results (so we don't have to re-run tests to commit again).
* [ ] All-errors validation reporting — collect every spec error per load, not just the
  first. Shares the "don't abort on first problem" machinery with partial load/commit/run
  above; design as one pass, not before. (vNext; the CUE Loader slice ships first-error
  only.) **Intended mechanism (decided 2026-06-16):** "parse don't validate" — parse
  becomes *total* (never fails), each failable field carries `value-or-Err` as data in a
  typed IR, and a fold-based `validate` walks the IR collecting errors. First-error vs
  all-errors is then a one-line change in the fold. This is also where `Tests()` finally
  splits into a total parse + validate. Don't build the IR before this lands — under
  first-error it's dead weight.
* [ ] JSONC support — deferred out of the Loader slice. Needs either a new dependency or a
  hand-rolled string-aware comment-stripper (must skip `//` and `/* */` inside string
  literals — correctness risk on untrusted input). Decide dep-vs-stripper on its own
  merits. Weakest-value of the JSON-family formats.
