#!/usr/bin/env bash
# Check documented module counts against the release manifest.
set -euo pipefail

if [ "$#" -eq 0 ]; then
  set -- .release-please-manifest.json README.md docs/index.md
elif [ "$#" -lt 2 ]; then
  printf 'usage: %s [manifest file ...]\n' "$0" >&2
  exit 2
fi

python3 - "$@" <<'PY'
import json
import re
import sys
from pathlib import Path


def unique_object(pairs):
    """Reject duplicate JSON keys."""
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError(f"duplicate key: {key}")
        result[key] = value
    return result


try:
    manifest_path = Path(sys.argv[1])
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"), object_pairs_hook=unique_object)
    if not isinstance(manifest, dict):
        raise ValueError("manifest must be an object")
    count = len(manifest)
    failed = False
    for filename in sys.argv[2:]:
        found = False
        for line_number, line in enumerate(Path(filename).read_text(encoding="utf-8").splitlines(), 1):
            for match in re.finditer(r"\b([0-9]+) modules\b", line):
                found = True
                if int(match.group(1)) != count:
                    print(f"{filename}:{line_number}: says {match.group(1)} modules, manifest has {count}", file=sys.stderr)
                    failed = True
        if not found:
            print(f"{filename}: no module count found", file=sys.stderr)
            failed = True
    sys.exit(1 if failed else 0)
except (OSError, UnicodeError, ValueError) as error:
    print(f"readme-module-count: {error}", file=sys.stderr)
    sys.exit(1)
PY
