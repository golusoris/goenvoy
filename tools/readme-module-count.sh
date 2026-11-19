#!/usr/bin/env bash
# Check documented module counts against the release manifest.
set -euo pipefail

mode=explicit
if [ "$#" -eq 0 ]; then
  mode=tracked
  set -- .release-please-manifest.json
elif [ "$#" -lt 2 ]; then
  printf 'usage: %s [manifest file ...]\n' "$0" >&2
  exit 2
fi

python3 - "$mode" "$@" <<'PY'
import json
import os
import re
import subprocess
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
    tracked = sys.argv[1] == "tracked"
    manifest_path = Path(sys.argv[2])
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"), object_pairs_hook=unique_object)
    if not isinstance(manifest, dict):
        raise ValueError("manifest must be an object")
    count = len(manifest)
    failed = False
    filenames = sys.argv[3:]
    if tracked:
        filenames = ["README.md", "docs/index.md"]
        paths = subprocess.check_output(["git", "ls-files", "-z"], timeout=60).split(b"\0")
        for raw_path in paths:
            if not raw_path:
                continue
            filename = os.fsdecode(raw_path)
            path = Path(filename)
            # Why: changelogs preserve historical release counts.
            if path.name == "CHANGELOG.md":
                continue
            # Why: tool tests contain intentionally stale fixture counts.
            if path.parent == Path("tools") and path.name.endswith("_test.sh"):
                continue
            # Why: this dated report records a historical count.
            if filename == "GOE-24_GATE_HEALTH_REPORT.md":
                continue
            if filename in filenames:
                continue
            if not path.is_file():
                continue
            data = path.read_bytes()
            if b"\0" in data:
                continue
            try:
                content = data.decode("utf-8")
            except UnicodeDecodeError:
                continue
            if re.search(r"\b[0-9]+ modules\b", content):
                filenames.append(filename)
    for filename in filenames:
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
except (OSError, UnicodeError, ValueError, subprocess.CalledProcessError, subprocess.TimeoutExpired) as error:
    print(f"readme-module-count: {error}", file=sys.stderr)
    sys.exit(1)
PY
