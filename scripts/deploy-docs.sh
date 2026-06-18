#!/usr/bin/env bash
# Deploy the built docs site (www/dist/) to the gh-pages branch.
#
# www/dist/ is gitignored, so a plain `git subtree push` can't see it. Instead we
# build, then stage the output as a fresh orphan commit inside a throwaway worktree
# and force-push that to gh-pages — gh-pages carries the artifact only, no history.
#
#   scripts/deploy-docs.sh            build + force-push to gh-pages
#   scripts/deploy-docs.sh --dry-run  build + commit locally, skip the push
#
# One-time, after the first push, enable GitHub Pages on the branch:
#   gh api repos/chakrit/smoke/pages -X POST \
#     -f 'source[branch]=gh-pages' -f 'source[path]=/'

set -euo pipefail

remote=gh
branch=gh-pages

root="$(cd "$(dirname "$0")/.." && pwd)"
dist="$root/www/dist"

dry_run=false
if [ "${1:-}" = "--dry-run" ]; then
  dry_run=true
fi

"$root/scripts/build-docs.sh"

if [ ! -d "$dist" ]; then
  echo "no build output at $dist" >&2
  exit 1
fi

worktree="$(mktemp -d)"
cleanup() {
  git -C "$root" worktree remove --force "$worktree" 2>/dev/null || true
  git -C "$root" branch -D "$branch" 2>/dev/null || true
}
trap cleanup EXIT

# Fresh orphan branch in the throwaway worktree, wiped down to nothing.
git -C "$root" worktree add --force --detach "$worktree" >/dev/null
git -C "$worktree" checkout --orphan "$branch" >/dev/null 2>&1
git -C "$worktree" rm -rf --quiet . >/dev/null 2>&1 || true
find "$worktree" -mindepth 1 -maxdepth 1 ! -name .git -exec rm -rf {} +

# Stage the built site. .nojekyll stops Pages from running the output through Jekyll.
cp -R "$dist"/. "$worktree"/
touch "$worktree/.nojekyll"

git -C "$worktree" add -A
git -C "$worktree" commit --quiet -m "Deploy docs site"

if [ "$dry_run" = true ]; then
  echo
  echo "[dry-run] committed to $branch, push skipped:"
  git -C "$worktree" log --oneline --stat -1
  exit 0
fi

git -C "$worktree" push --force "$remote" "$branch"
echo
echo "deployed -> $remote/$branch"
