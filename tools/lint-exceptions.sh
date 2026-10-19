#!/usr/bin/env bash
# Regenerate the dated complexity exceptions in .golangci.yml from a measured lint run (#141).
# Needs golangci-lint, jq and praetorctl (engine with cordanaLLM/praetor#1388) on PATH.
# tools/go/lintexceptions (make lint-exceptions) judges the result.
set -euo pipefail
root=$(git rev-parse --show-toplevel)
cd "$root"
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
# The stripped copy drops the generated block, so every current finding is measured again.
praetorctl linters exceptions strip-config -config .golangci.yml -out "$work/golangci.yml"
n=0
while IFS= read -r mod; do
  n=$((n + 1))
  (cd "$mod" && golangci-lint run --config="$work/golangci.yml" \
    --enable-only=funlen,gocyclo,gocognit,cyclop --issues-exit-code=0 \
    --max-issues-per-linter=0 --max-same-issues=0 --uniq-by-line=false \
    --output.text.path=stderr --output.json.path="$work/report-$n.json" ./...) 2>/dev/null
done < <(find . -name go.mod -not -path './.workingdir*/*' -exec dirname {} \; | sort)
jq -s '{Issues: (map(.Issues // []) | add)}' "$work"/report-*.json > "$work/report.json"
echo "measured $(jq '.Issues | length' "$work/report.json") findings in $n modules"
praetorctl linters exceptions generate -config .golangci.yml -report "$work/report.json" -issue 141 -category C -write
