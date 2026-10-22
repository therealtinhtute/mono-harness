#!/usr/bin/env bash
# Drive the tldraw offline desktop app through its local HTTP API.
# Usage:
#   tl.sh draw <file.tldraw> <body.js>   create the file if missing, then run lib.js + body.js
#   tl.sh png  <file.tldraw> <out.png> [maxpx]   export the page as PNG (default max side 2000px)
#   tl.sh exec <file.tldraw> <code.js>   run arbitrary editor code
#   tl.sh search <code.js>               query the Editor API spec
set -euo pipefail
DIR=$(cd "$(dirname "$0")" && pwd)
F="$HOME/Library/Application Support/tldraw/server.json"
[ -f "$F" ] || { echo "error: tldraw offline is not running (no $F); open the tldraw app first" >&2; exit 1; }
PORT=$(jq -r .port "$F"); TOKEN=$(jq -r .token "$F"); B="http://localhost:$PORT"
post() { curl -sf "$B$1" -X POST -H 'content-type: application/json' -H "Authorization: Bearer $TOKEN" -d "$2"; }
abs() { (cd "$(dirname "$1")" && printf '%s/%s' "$(pwd -P)" "$(basename "$1")"); }
docid() { printf 'tldr:file:%s' "$(printf %s "$(abs "$1")" | base64 | tr -d '=\n')"; }
ensure() {
  [ -f "$1" ] && return 0
  local name; name=$(basename "$1" .tldraw)
  post /api/docs/create "$(jq -n --arg n "$name" --arg d "$(cd "$(dirname "$1")" && pwd -P)" '{name:$n,directory:$d}')" >/dev/null
}
run() { post "/api/doc/$(docid "$1")/exec" "$(jq -Rs '{code:.}' < "$2")"; }
case "${1:-}" in
  draw)
    ensure "$2"
    post "/api/doc/$(docid "$2")/exec" "$(cat "$DIR/lib.js" "$3" | jq -Rs '{code:.}')" \
      | jq -c '{success, n: .result.n, lints: [.result.lints.new[]?.message]}' ;;
  png)
    run "$2" "$DIR/export-png.js" | jq -er .result | base64 -d > "$3"
    sips -Z "${4:-2000}" "$3" >/dev/null && echo "wrote $3" ;;
  exec) run "$2" "$3" ;;
  search) post /api/search "$(jq -Rs '{code:.}' < "$2")" ;;
  *) sed -n '2,8p' "$0"; exit 2 ;;
esac
