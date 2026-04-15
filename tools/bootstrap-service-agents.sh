#!/usr/bin/env bash
# Bootstrap a skeleton AGENTS.md into every service module that lacks one.
# Idempotent — files that already exist are skipped.
# Run from repo root: tools/bootstrap-service-agents.sh
set -euo pipefail

TODAY=$(date -u +%Y-%m-%d)
created=0
skipped=0

while IFS= read -r modfile; do
  dir=$(dirname "$modfile")
  # Skip category-root modules (arr/, metadata/, etc.) — only service modules at depth >= 2 get the template.
  # Depth check: count '/' segments in the path stripped of './'.
  rel=${dir#./}
  depth=$(awk -F/ '{print NF}' <<<"$rel")
  (( depth < 2 )) && continue
  # Skip metadata sub-category roots (metadata/video, metadata/anime, …) — they're not services either.
  case "$rel" in
    metadata/video|metadata/anime|metadata/music|metadata/tracking|metadata/book|metadata/game|metadata/adult) continue ;;
  esac
  if [ -f "$dir/AGENTS.md" ]; then
    skipped=$((skipped + 1))
    continue
  fi
  svc=$(basename "$dir")
  cat_path=$(dirname "$rel")
  doc_slug="${cat_path//\//-}-${svc}"
  cat > "$dir/AGENTS.md" <<EOF
# AGENTS — ${cat_path}/${svc}

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <TODO: pinned upstream docs URL>
- Pinned version / commit: <TODO>
- Last verified: ${TODAY}
- Details: [docs/upstream/${doc_slug}.md](../../docs/upstream/${doc_slug}.md)

## Auth model

<TODO>

## Pagination

<TODO>

## Rate limits

<TODO>

## Known quirks

- <TODO>

## Testing notes

<TODO>
EOF
  created=$((created + 1))
  echo "created: $dir/AGENTS.md"
done < <(find . -mindepth 3 -name 'go.mod' -not -path './.workingdir*/*')

echo
echo "bootstrap complete: ${created} created, ${skipped} skipped (already existed)"
