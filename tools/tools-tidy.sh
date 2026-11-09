#!/usr/bin/env bash
# Why: fail when a Go module's go.mod or go.sum differ from what go mod tidy writes (#267: a major bump left the imports behind).
set -euo pipefail

dir="${1:?usage: tools-tidy.sh <module dir>}"
cd "$dir"
if ! go mod tidy -diff; then
  echo "tools-tidy: $dir: go.mod or go.sum differ from go mod tidy; run (cd $dir && go mod tidy)" >&2
  exit 1
fi
