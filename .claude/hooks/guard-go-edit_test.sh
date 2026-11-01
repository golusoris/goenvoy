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
expect "single-line import" 2 a.go 'import "github.com/pkg/errors"'
expect "single-line aliased" 2 a.go 'import errs "github.com/pkg/errors"'
expect "single-line blank" 2 a.go 'import _ "gopkg.in/yaml.v3"'
expect "single-line dot" 2 a.go 'import . "github.com/pkg/errors"'
expect "single-line own module" 0 a.go 'import arr "github.com/golusoris/goenvoy/arr/v2"'
expect "unlisted host" 2 a.go $'import (\n\t"example.com/mod"\n)'
expect "own path in trailing comment" 2 a.go $'import (\n\t"github.com/pkg/errors" // like "github.com/golusoris/goenvoy/arr/v2"\n)'
expect "return of a string" 0 a.go $'func f() string {\n\treturn "github.com/pkg/errors"\n}'
expect "slice of strings" 0 a.go $'var s = []string{\n\t"github.com/pkg/errors",\n}'
expect "edit fragment" 2 a.go $'\t"github.com/pkg/errors"'
expect "file with raw-string fixture" 0 a_test.go $'package a\n\nimport (\n\t"testing"\n)\n\nconst fixture = `[\n"example.com"\n]`\n\nfunc TestA(t *testing.T) {}'
expect "file with foreign import" 2 a.go $'package a\n\nimport (\n\t"net/http"\n\t"github.com/pkg/errors"\n)\n\nfunc f() {}'
expect "file with single-line foreign import" 2 a.go $'package a\n\nimport errs "github.com/pkg/errors"\n\nvar _ = errs.New'
expect "dev-tool module" 0 /repo/tools/go/upstreamspec/manifest.go $'package main\n\nimport (\n\t"go.yaml.in/yaml/v3"\n)\n'
expect "library module named tools" 2 /repo/arr/tools.go $'package arr\n\nimport (\n\t"go.yaml.in/yaml/v3"\n)\n'
expect "not a go file" 0 a.md $'import (\n\t"github.com/pkg/errors"\n)'
expect "quoted path in block comment" 0 a.go $'package a\n\nimport ( /* see\n\t"github.com/pkg/errors"\n*/\n\t"net/http"\n)\n'
expect "comment after group opener" 2 a.go $'package a\n\nimport ( // deps\n\t"github.com/pkg/errors"\n)\n'
expect "grouped import without space" 2 a.go 'import("github.com/pkg/errors")'
expect "grouped specs on one line" 2 a.go 'import ("net/http"; "github.com/pkg/errors")'
expect "file with syntax error below imports" 2 a.go $'package a\n\nimport "github.com/pkg/errors"\n\nfunc f( {'
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
printf 'package a\n\nimport (\n\t"net/http"\n)\n' > "$tmp/a.go"
expect_edit() { # name, want, old_string, new_string
  local payload got
  payload=$(jq -n --arg f "$tmp/a.go" --arg o "$3" --arg n "$4" '{tool_name:"Edit",tool_input:{file_path:$f,old_string:$o,new_string:$n}}')
  printf '%s' "$payload" | bash "$hook" >/dev/null 2>&1
  got=$?
  if [ "$got" != "$2" ]; then
    printf 'FAIL %s: exit %s, want %s\n' "$1" "$got" "$2"
    fail=1
  fi
}
expect_edit "edit adds foreign import" 2 $'\t"net/http"' $'\t"net/http"\n\n\t"github.com/pkg/errors"'
expect_edit "edit adds stdlib import" 0 $'\t"net/http"' $'\t"net/http"\n\t"strings"'
expect_edit "edit adds own module" 0 $'\t"net/http"' $'\t"net/http"\n\n\t"github.com/golusoris/goenvoy/arr/v2"'
# Why: hosts without GNU coreutils (macOS) lack sha256sum; the guard must still block there.
nosha="$tmp/bin"
mkdir -p "$nosha"
for tool in bash jq go cksum cut dirname grep cat; do
  ln -s "$(command -v "$tool")" "$nosha/$tool"
done
payload=$(jq -n '{tool_name:"Write",tool_input:{file_path:"a.go",content:"package a\n\nimport \"github.com/pkg/errors\"\n"}}')
got=0
printf '%s' "$payload" | PATH="$nosha" "$nosha/bash" "$hook" >/dev/null 2>&1 || got=$?
if [ "$got" != 2 ]; then
  printf 'FAIL host without sha256sum: exit %s, want 2\n' "$got"
  fail=1
fi
[ "$fail" = 0 ] && echo "guard-go-edit: all planted cases pass"
exit "$fail"
