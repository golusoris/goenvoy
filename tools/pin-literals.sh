#!/usr/bin/env bash
# Check tracked files for copies of the shared pin values.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT="$(cd "${2:-$REPO_ROOT}" && pwd)"
PIN_FILE="${1:-$ROOT/tools/tool-versions.env}"
PIN_FILE="$(cd "$(dirname "$PIN_FILE")" && pwd)/$(basename "$PIN_FILE")"
LIST="$(mktemp)"
trap 'rm -f "$LIST"' EXIT
git -C "$ROOT" ls-files -z > "$LIST"
pins=0
failed=0

# scan_file reports each line containing a whole version or digest literal.
scan_file() {
  PIN_VALUE="$value" PIN_NAME="$name" PIN_PATH="$file" awk '
    BEGIN { value = ENVIRON["PIN_VALUE"]; name = ENVIRON["PIN_NAME"]; path = ENVIRON["PIN_PATH"] }
    {
      line = $0
      offset = 0
      for (i = 0; i <= length($0); i++) {
        pos = index(line, value)
        if (!pos) break
        start = offset + pos
        before = start > 1 ? substr($0, start - 1, 1) : ""
        after = substr($0, start + length(value), 1)
        if (digest || (before !~ /[0-9.]/ && after !~ /[0-9.]/)) {
          printf "%s:%d: %s literal duplicated outside tools/tool-versions.env\n", path, NR, name
          found = 1
          break
        }
        offset = start
        line = substr($0, offset + 1)
      }
    }
    END { if (found) exit 1 }
  ' digest="$digest" "$ROOT/$file"
}

mapfile -t pin_lines < "$PIN_FILE"
for line in "${pin_lines[@]}"; do
  case "$line" in
    ''|'#'*) continue ;;
    *=*) name="${line%%=*}"; value="${line#*=}" ;;
    *) printf 'invalid pin file line: %s\n' "$line" >&2; exit 1 ;;
  esac
  case "$name" in ''|*[!A-Za-z0-9_]*) printf 'invalid pin name: %s\n' "$name" >&2; exit 1 ;; esac
  if [ -z "$value" ]; then printf 'empty value in pin file: %s\n' "$name" >&2; exit 1; fi
  pins=$((pins + 1))
  digest=0
  case "$value" in sha256:*) value="${value#sha256:}"; digest=1 ;; esac
  if [ -z "$value" ]; then printf 'empty digest in pin file: %s\n' "$name" >&2; exit 1; fi
  # Why: allow-list of value characters; a quoted value or a trailing comment would be searched as written and miss real copies.
  case "$value" in *[!A-Za-z0-9._-]*) printf 'unsupported pin value for %s: only letters, digits, dot, underscore and hyphen\n' "$name" >&2; exit 1 ;; esac
  while IFS= read -r -d '' file; do
    case "$file" in CHANGELOG.md|*/CHANGELOG.md) continue ;; esac
    [ "$ROOT/$file" != "$PIN_FILE" ] || continue
    [ -f "$ROOT/$file" ] || continue
    if scan_file; then :; else
      status=$?
      if [ "$status" -ne 1 ]; then exit "$status"; fi
      failed=1
    fi
  done < "$LIST"
done
if [ "$pins" -eq 0 ]; then echo 'pin file contains no pins' >&2; exit 1; fi
exit "$failed"
