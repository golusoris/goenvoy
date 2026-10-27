#!/usr/bin/env bash
# Drop nosemgrep-suppressed results and dated per-file exceptions from a Semgrep
# SARIF file in place, then fail when any result is left.
#   usage: tools/semgrep-filter.sh SARIF [TODAY]   (TODAY defaults to the UTC date)
# An exception past its expiry no longer drops its result, so the finding fails again.
set -euo pipefail

sarif=$1
today=${2:-$(date -u +%F)}

# file, full rule id, expiry (inclusive), reason. Category B: 3 days.
exceptions='[]'

jq -r --arg today "$today" --argjson ex "$exceptions" \
  '$ex[] | select(.expires < $today) | "semgrep exception expired \(.expires): \(.file) \(.rule)"' -n >&2

jq --arg today "$today" --argjson ex "$exceptions" '
  ($ex | map(select(.expires >= $today))) as $live
  | .runs[].results |= map(
      select((.suppressions // []) | length == 0)
      | (.locations[0].physicalLocation.artifactLocation.uri // "" | ltrimstr("./")) as $uri
      | .ruleId as $rule
      | select([$live[] | select(.file == $uri and .rule == $rule)] | length == 0)
    )' "$sarif" > "$sarif.tmp"
mv "$sarif.tmp" "$sarif"

left=$(jq '[.runs[].results[]] | length' "$sarif")
if [ "$left" -gt 0 ]; then
  jq -r '.runs[].results[] | "\(.locations[0].physicalLocation.artifactLocation.uri):\(.locations[0].physicalLocation.region.startLine) \(.ruleId)"' "$sarif" >&2
  echo "semgrep: $left unexcepted result(s)" >&2
  exit 1
fi
