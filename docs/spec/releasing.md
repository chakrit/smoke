# Releasing

- **Status:** accepted

SMOKE releases are **lightweight git tags** — no build artifacts, no changelog
ceremony, no release branch. Versioning is semver `vMAJOR.MINOR.PATCH`
(latest: `v0.2.4`). The tag *is* the release; `go install` resolves it.

## Process

1. **Be on `master`, clean tree, at the commit you want to release.**
2. **Pass the gate** — both must be green:
   ```sh
   go build -o ./bin/smoke .
   go test ./...                      # unit tests
   ./bin/smoke --no-color test/tests.yml   # self-hosting suite; exit 0 = UNCHANGED
   ```
   The smoke suite is the real gate: exit `0` means the binary's own observable
   output still matches its committed golden. Any nonzero — do **not** tag.
3. **Tag** (lightweight, matching every prior tag — do not annotate):
   ```sh
   git tag vX.Y.Z
   ```
4. **Push the tag** to the GitHub remote:
   ```sh
   git push gh vX.Y.Z
   ```

## Notes

- Remote is `gh`, not `origin` (per repo convention).
- Tags are lightweight commit refs — `git for-each-ref refs/tags` shows
  `objecttype = commit`. Keep it that way; an annotated tag would break the
  uniform history.
- No version string lives in the source, so a release is purely the tag. Nothing
  to bump in code before tagging.
