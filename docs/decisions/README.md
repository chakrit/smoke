# Decisions Log

**Point-in-time defenses against future re-litigation** — rulings made on
a specific date for a specific question, recorded so the same argument
doesn't have to be re-fought next quarter. Each entry is frozen at the
moment of decision; if a later ruling reverses it, write a new dated
decision that links back and mark the old one `superseded`.

## When to add an entry

Add a decision when **the answer goes against the obvious default** —
mainstream practice, what the agent's training data would suggest, or the
project's own prior convention. The point of the log is to capture the
*why* so future arguments don't keep re-discovering it. Examples that
warrant an entry:

- We deliberately deviate from a well-known pattern, and a future agent
  reading our code would assume we just didn't know better.
- A reviewer pushed back on a choice that we then defended; the defense
  is worth preserving.
- Two reasonable approaches were debated and one won — without the entry,
  the next debate replays from scratch.

**Don't** add a decision when the answer is already obvious or matches
the prevailing convention. If there's no future confusion to head off,
just document the result in `../spec/` and move on. A decisions log
cluttered with "we chose the obvious thing" entries makes the actual
load-bearing decisions harder to find.

If your artifact is research, a survey, a draft, a transcript, or any
exploratory write-up — that's notes, not a decision. Use `../notes/`. If
it's forward-looking design, use `../spec/`.

## Format

One file per decision: `YYYY-MM-DD-slug.md`

```markdown
# Short Title
- **Date:** YYYY-MM-DD
- **PR:** #N (or "manual")
- **Status:** accepted | superseded | revised

## Decision
One-liner.

## Rationale
Why this, and specifically why *not* the obvious alternative — that's
the part that prevents re-litigation.
```

## Statuses

- **accepted** — active, follow this decision
- **superseded** — replaced by a newer decision (link to it)
- **revised** — updated in-place with new context
