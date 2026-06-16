# Using SMOKE in a TDD / agent loop

- **Status:** current

> Consumer-facing guidance for anything driving SMOKE programmatically — an
> agent in a code-edit loop, a CI script, a human pairing with an LLM. The
> numeric contract lives in [`exit-codes.md`](exit-codes.md); this doc is how to
> *read* it without importing test-runner assumptions that do not hold.

## SMOKE is a drift detector, not an assertion engine

SMOKE answers one question: *does the command's observable output still match
the committed golden?* It does **not** answer *is the behavior correct?* Every
consumer trained on `green = pass / red = fail` will conflate the two. That
conflation is the single largest failure mode when an LLM drives the loop.

Map the states to what they actually mean, not to test-runner habits:

| State       | Exit | Reads as "pass/fail"? | What it actually means                        |
| ----------- | ---- | --------------------- | --------------------------------------------- |
| `UNCHANGED` | 0    | **not** "passed"      | Output matches the golden. Says nothing about correctness. |
| `CHANGED`   | 1    | **not** "failed"      | Output drifted. Expected during intentional changes — review it. |
| `NEW`       | 3    | —                     | No golden yet. A human/LLM must eyeball and commit the first lock. |
| (operational) | 2  | —                     | SMOKE itself broke (runner crash, I/O). Stop and fix the harness. |
| (usage)     | 64   | —                     | Invalid invocation — bad flags. Stop.         |
| (data)      | 65   | —                     | A spec or lock file is malformed. Stop and fix the file. |

## The three traps

- **`CHANGED` is not a failing test.** Exit `1` during an intentional change is
  the *expected* state: eyeball the diff, and if it is the new correct output,
  `smoke -c` to re-commit the golden. Do **not** pattern-match `CHANGED` as a
  red test and "fix" the code to chase the output back to `UNCHANGED` — that
  defeats the entire workflow. Re-commit; don't revert.

- **`UNCHANGED` is not "correct".** Exit `0` means the output didn't move, not
  that the behavior is right. Two ways it lies: the behavior changed but the
  observable surface didn't (a coverage gap — SMOKE never claimed to cover it),
  or the golden itself locked in a bug and stability now perpetuates it. Green
  forever, wrong forever. Treat `UNCHANGED` as "no drift to review," never as a
  verification result.

- **Re-committing can hide a regression.** `smoke -c` is correct when you
  *intended* the change and verified the new output. It is wrong when you
  re-commit a `CHANGED` result you didn't read — you've just blessed a
  regression as the new golden. The eyeball step is load-bearing; an agent that
  auto-commits on every `CHANGED` has turned SMOKE off.

## For an agent: branch on the exit code

Do not parse human output or pattern-match colors/words. Branch on `$?`:

- `0` — no drift. Continue; do not treat as a verification pass.
- `1` — drift. Surface the diff for review; re-commit only if the change was
  intended and the new output checks out.
- `2` / `64` / `65` — SMOKE broke (`2`), the invocation was bad (`64`), or a
  spec/lock file is malformed (`65`). Stop; none of these is drift.
- `3` — first run, no golden. Review the output and commit the initial lock.

The whole point of the distinct codes is that you never have to guess which of
"the thing under test changed" vs "the harness broke" you are looking at.
