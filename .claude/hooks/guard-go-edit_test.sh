#!/usr/bin/env bash
# Planted cases for guard-go-edit.sh: each payload must give the expected exit status.
set -uo pipefail
hook="$(dirname "$0")/guard-go-edit.sh"
command -v jq >/dev/null 2>&1 || { echo "guard-go-edit test: jq missing" >&2; exit 1; }
fail=0
expect() { # name, want, file, content
  local payload got
  payload=$(jq -n --arg f "$3" --arg c "$4" '{tool_name:"Write",tool_input:{file_path:$f,content:$c}}')
  printf '%s' "$payload" | bash "$hook" >/dev/null 2>&1
  got=$?
  if [ "$got" != "$2" ]; then
    printf 'FAIL %s: exit %s, want %s\n' "$1" "$got" "$2"
    fail=1
  fi
}
expect "stdlib import" 0 a.go $'import (\n\t"net/http"\n)'
expect "x/ import" 0 a.go $'import (\n\t"golang.org/x/net/html"\n)'
expect "own module" 0 a.go $'import (\n\t"github.com/golusoris/goenvoy/arr/v2"\n)'
expect "own module aliased" 0 a.go $'import (\n\tarr "github.com/golusoris/goenvoy/arr/v2"\n)'
expect "third party" 2 a.go $'import (\n\t"github.com/pkg/errors"\n)'
expect "third party aliased" 2 a.go $'import (\n\terrs "github.com/pkg/errors"\n)'
expect "third party blank" 2 a.go $'import (\n\t_ "gopkg.in/yaml.v3"\n)'
expect "lookalike module" 2 a.go $'import (\n\t"github.com/golusoris/goenvoyx/arr"\n)'
expect "own and foreign" 2 a.go $'import (\n\t"github.com/golusoris/goenvoy/arr/v2"\n\t"github.com/pkg/errors"\n)'
expect "not a go file" 0 a.md $'import (\n\t"github.com/pkg/errors"\n)'
[ "$fail" = 0 ] && echo "guard-go-edit: all planted cases pass"
exit "$fail"
