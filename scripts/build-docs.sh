#!/usr/bin/env bash
# Build the SMOKE documentation site from docs/guides/ into www/dist/.
#
# docs/guides/index.md is the canonical content (and renders on GitHub as-is).
# www/ is a Parcel project that frames that markdown into a static site; the
# build output in www/dist/ is the deployable artifact (subtree it to gh-pages).

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root/www"

if [ ! -d node_modules ]; then
  echo "installing www/ dependencies..."
  npm install
fi

npm run build

echo
echo "built -> www/dist/"
echo "preview locally: (cd www/dist && python3 -m http.server 8910) then open http://localhost:8910/"
