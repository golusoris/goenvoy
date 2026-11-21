#!/usr/bin/env bash
# Detect privileged in-place writes; privileged install must stage to .new
# or create directories. Only allow listed commands without redirects.
# Bash splits words on space and tab only.
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
    /^[ \t]*#/ { next }
    {
      line = $0
      while (match(tolower(line), command)) {
        start = RSTART
        call = substr(line, RSTART, RLENGTH)
        rest = substr(line, RSTART + RLENGTH)
        match(tolower(call), prefix "[[:space:]]+")
        candidate = tolower(substr(call, 1, RLENGTH))
        gsub(/[^a-z0-9_]/, "", candidate)
        if (candidate ~ "^su" "do_(askpass|command|editor|gid|home|prompt|ps1|tty|uid|user)$") {
          line = substr(line, start + RLENGTH - 1)
          continue
        }
        call = substr(call, RLENGTH + 1)
        split(rest, segments, /[;|&]/)
        install_ok = 0
        if (call == "install") {
          delete args
          count_args = 0
          word = ""
          started = 0
          quote = ""
          refused = 0
          segment = segments[1]
          for (pos = 1; pos <= length(segment); pos++) {
            char = substr(segment, pos, 1)
            if (quote == "\047") {
              if (char == "\047") quote = ""
              else word = word char
            } else if (quote == "\"") {
              if (char == "\\" && pos < length(segment)) {
                next_char = substr(segment, ++pos, 1)
                if (index("$`\"\\", next_char)) word = word next_char
                else word = word char next_char
              } else if (char == "\"") quote = ""
              else word = word char
            } else if (char == "\\") {
              if (pos < length(segment)) word = word substr(segment, ++pos, 1)
              else word = word char
              started = 1
            } else if (char == "\047" || char == "\"") {
              quote = char
              started = 1
            } else if (char == " " || char == "\t") {
              if (started) args[++count_args] = word
              word = ""
              started = 0
            } else if (char == "#" && !started) {
              break
            } else {
              if (char !~ /^[[:alnum:]_./:=,@%+-]$/) refused = 1
              word = word char
              started = 1
            }
          }
          if (started) args[++count_args] = word
          last = ""
          if (quote != "") refused = 1
          directory = 0
          for (i = 1; i <= count_args; i++) {
            if (args[i] == "") continue
            if (args[i] !~ /^-/) {
              last = args[i]
            } else if (args[i] == "--directory") {
              directory = 1
            } else if (args[i] ~ /^--(compare|preserve-timestamps|verbose|strip|no-target-directory)$/) {
              continue
            } else if (args[i] ~ /^--(mode|owner|group)=./) {
              continue
            } else if (args[i] ~ /^--(mode|owner|group)$/) {
              i++
            } else if (args[i] ~ /^-[CDTdpsv]+$/) {
              if (args[i] ~ /d/) directory = 1
            } else if (args[i] ~ /^-[CDTdpsv]*[gmo]$/) {
              if (args[i] ~ /d/) directory = 1
              i++
            } else if (args[i] ~ /^-[CDTdpsv]*[gmo].+$/) {
              flags = args[i]
              sub(/[gmo].*$/, "", flags)
              if (flags ~ /d/) directory = 1
            } else {
              refused = 1
              break
            }
          }
          if (!refused && (directory || last ~ /\.new$/)) install_ok = 1
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
