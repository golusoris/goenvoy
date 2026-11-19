#!/usr/bin/env bash
# Detect privileged in-place writes; privileged install must stage to .new
# or create directories. Only allow listed commands without redirects.
set -euo pipefail

if (( $# > 1 )); then
  printf "usage: tools/privileged-write.sh [root]\n" >&2
  exit 2
fi
root="$(cd "${1:-$(dirname "${BASH_SOURCE[0]}")/..}" && pwd)"
list="$(mktemp)"
trap 'rm -f "$list" "$list.hits"' EXIT
git -C "$root" ls-files -z > "$list"
cd "$root"
count=0
failed=0
while IFS= read -r -d '' file; do
  case "$file" in
    *.sh|*.bash|Makefile|*/Makefile|.github/workflows/*.yml|.github/workflows/*.yaml|.github/actions/*.yml|.github/actions/*.yaml) ;;
    *) continue ;;
  esac
  [[ -f "$file" ]] || continue
  count=$((count + 1))
  if awk '
    BEGIN {
      name = "[a-z_][a-z0-9_]*" "(su" "do|privi" "lege)[a-z0-9_]*|" "(su" "do|privi" "lege)[a-z0-9_]*"
      variable = "\\$(" name ")|\\$\\{(" name ")\\}|\\$\\((" name ")\\)"
      prefix = "(^|[;&|([:space:]])(" "su" "do|do" "as|" variable "|\"(" variable ")\")"
      command = prefix "[[:space:]]+[^[:space:];|&]+"
    }
    FNR > 100000 { print FILENAME ":" FNR; failed = 1; nextfile }
    /^[[:space:]]*#/ { next }
    {
      line = $0
      while (match(tolower(line), command)) {
        call = substr(line, RSTART, RLENGTH)
        rest = substr(line, RSTART + RLENGTH)
        match(tolower(call), prefix "[[:space:]]+")
        call = substr(call, RLENGTH + 1)
        split(rest, segments, /[;|&]/)
        install_ok = 0
        if (call == "install") {
          count_args = split(segments[1], args, /[[:space:]]+/)
          last = ""
          for (i = 1; i <= count_args; i++) {
            if (args[i] == "") continue
            if (args[i] == "-d" || args[i] == "--directory") install_ok = 1
            last = args[i]
          }
          if (last ~ /^".*"$/ || last ~ /^\047.*\047$/) last = substr(last, 2, length(last) - 2)
          if (last ~ /\.new$/) install_ok = 1
          if (segments[1] ~ /\\[[:space:]]*$/) install_ok = 0
        }
        if (call !~ /^(install|mv|rm|chown|chmod|test|cat|systemctl)$/ || segments[1] ~ />/ || (call == "install" && !install_ok)) {
          print FILENAME ":" FNR
          failed = 1
        }
        line = rest
      }
    }
    END { exit failed }
  ' "$file" > "$list.hits"; then :; else
    status=$?
    if (( status != 1 )); then rm -f "$list.hits"; exit "$status"; fi
    failed=1
  fi
  while IFS= read -r hit; do
    printf "privileged-write: %s: privileged command not on the allow-list or writing through a redirect; stage the file and rename it into place\n" "$hit" >&2
  done < "$list.hits"
  rm -f "$list.hits"
done < "$list"
if (( failed )); then exit 1; fi
printf "privileged-write: %s files clean\n" "$count"
