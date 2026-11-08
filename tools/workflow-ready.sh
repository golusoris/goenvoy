#!/usr/bin/env bash
# Why: enforce ready_for_review on all GitHub Actions workflows with pull_request triggers.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET_ROOT=""

# Why: parse optional root argument without unbounded loops.
while [ $# -gt 0 ]; do
  case "$1" in
    -*)
      echo "usage: ${0##*/} [root]" >&2
      exit 2
      ;;
    *)
      if [ -n "$TARGET_ROOT" ]; then
        echo "usage: ${0##*/} [root]" >&2
        exit 2
      fi
      TARGET_ROOT="$1"
      shift
      ;;
  esac
done

ROOT="${TARGET_ROOT:-$REPO_ROOT}"
ROOT="${ROOT%/}"

if [ ! -d "$ROOT" ]; then
  echo "error: directory $ROOT does not exist" >&2
  exit 1
fi

shopt -s nullglob
files=("$ROOT"/.github/workflows/*.yml "$ROOT"/.github/workflows/*.yaml)

if [ ${#files[@]} -eq 0 ]; then
  echo "error: no workflow files found in $ROOT/.github/workflows" >&2
  exit 1
fi

misses=0
ok_prs=0
files_checked=0

# Why: bounded loop over discovered workflow files.
for f in "${files[@]}"; do
  if [ ! -f "$f" ]; then
    continue
  fi

  rel="${f#"$ROOT"/}"
  files_checked=$((files_checked + 1))

  readarray -t lines < "$f"
  num_lines=${#lines[@]}

  in_on=0
  on_indent=-1
  in_pr=0
  pr_indent=0
  pr_has_ready=0
  in_types_block=0
  types_indent=0
  file_miss=0
  file_prs=0
  pending_dash_indent=-1

  # Why: bounded loop over lines of the workflow file.
  for ((lnum = 0; lnum < num_lines; lnum++)); do
    line="${lines[lnum]}"
    # Why: CRLF workflows leave \r after readarray -t; strip it.
    line="${line%$'\r'}"

    # Why: strip leading spaces to measure indent and find trimmed content.
    trimmed="${line#"${line%%[! ]*}"}"

    # Why: skip blank lines and pure comment lines.
    if [ -z "$trimmed" ] || [[ "$trimmed" =~ ^# ]]; then
      continue
    fi

    indent=$(( ${#line} - ${#trimmed} ))

    # Why: an indentless sequence item under on: is '-' or starts with '- ' or a dash and a tab.
    indentless_on_item=0
    if [ "$in_on" -eq 1 ] && { [ "$trimmed" = "-" ] || [[ "$trimmed" == "- "* ]] || [[ "$trimmed" == -$'\t'* ]]; } && [ "$on_indent" -le 0 ]; then
      indentless_on_item=1
    fi

    # Why: indentation 0 marks a top-level key that closes any open block.
    if [ "$indent" -eq 0 ] && [ "$indentless_on_item" -eq 0 ]; then
      if [ "$in_pr" -eq 1 ]; then
        if [ "$pr_has_ready" -eq 0 ]; then
          file_miss=1
        else
          file_prs=$((file_prs + 1))
        fi
        in_pr=0
        pr_has_ready=0
        in_types_block=0
      fi
      in_on=0
      pending_dash_indent=-1

      if [[ "$trimmed" != -* ]] && [[ "$trimmed" == *:* ]]; then
        key="${trimmed%%:*}"
        clean_key="${key#[\'\"]}"
        clean_key="${clean_key%[\'\"]}"
        rest="${trimmed#*:}"

        if [ "$clean_key" = "on" ]; then
          in_on=1
          on_indent=-1
          inline_clean="${rest%%#*}"
          inline_clean="${inline_clean%"${inline_clean##*[! ]}"}"
          if [ -n "$inline_clean" ]; then
            if [[ "$inline_clean" =~ (^|[[:space:],\'\"\[])pull_request([[:space:],\'\"\]]|$) ]]; then
              file_miss=1
            fi
          fi
        fi
      fi
      continue
    fi

    # Why: only inspect lines located inside the on: mapping.
    if [ "$in_on" -eq 1 ]; then
      # Why: record the indent of on:'s direct children on first sight.
      if [ "$on_indent" -eq -1 ]; then
        on_indent="$indent"
      fi

      # Why: resolve a pending bare-dash whose scalar appears on this line.
      if [ "$pending_dash_indent" -ge 0 ]; then
        if [ "$indent" -gt "$pending_dash_indent" ]; then
          item="${trimmed%%#*}"
          item="${item%"${item##*[![:blank:]]}"}"
          item="${item#[\'\"]}"
          item="${item%[\'\"]}"
          if [ "$item" = "pull_request" ]; then
            file_miss=1
          fi
        fi
        pending_dash_indent=-1
      fi

      # Why: check if line exits current pull_request mapping.
      if [ "$in_pr" -eq 1 ]; then
        if [ "$indent" -le "$pr_indent" ]; then
          if [ "$pr_has_ready" -eq 0 ]; then
            file_miss=1
          else
            file_prs=$((file_prs + 1))
          fi
          in_pr=0
          pr_has_ready=0
          in_types_block=0
        fi
      fi

      # Why: check if this line starts a pull_request trigger under on:.
      if [ "$in_pr" -eq 0 ]; then
        if [[ "$trimmed" != -* ]] && [[ "$trimmed" == *:* ]]; then
          key="${trimmed%%:*}"
          clean_key="${key#[\'\"]}"
          clean_key="${clean_key%[\'\"]}"
          if [ "$clean_key" = "pull_request" ]; then
            in_pr=1
            pr_indent="$indent"
            pr_has_ready=0
            in_types_block=0
            continue
          fi
        fi
        # Why: detect pull_request as a block-list item (- pull_request) under on:.
        # Block-list events have no mapping body, so types: cannot be specified;
        # ready_for_review is always missing.
        # Only match at on:'s direct child indent to avoid false positives on
        # nested lists (e.g. branches: [- pull_request]).
        if [[ "$trimmed" == -* ]] && [ "$indent" -eq "$on_indent" ]; then
          item="${trimmed#-}"
          item="${item#"${item%%[![:blank:]]*}"}"
          item="${item%%#*}"
          item="${item%"${item##*[![:blank:]]}"}"
          # Why: strip YAML block scalar indicators (>-, >, |, |-).
          item="${item#>-}"
          item="${item#>}"
          item="${item#\|-}"
          item="${item#\|}"
          item="${item#"${item%%[![:blank:]]*}"}"
          item="${item#[\'\"]}"
          item="${item%[\'\"]}"
          if [ "$item" = "pull_request" ]; then
            file_miss=1
          elif [ -z "$item" ]; then
            # Why: bare dash or block scalar indicator with no inline value;
            # the scalar appears on the next indented line.
            pending_dash_indent="$indent"
          fi
        fi
      fi

      # Why: handle keys and list items inside the pull_request block.
      if [ "$in_pr" -eq 1 ]; then
        # Why: check if line is an item inside block-list types:.
        if [ "$in_types_block" -eq 1 ]; then
          if [ "$indent" -lt "$types_indent" ] || { [ "$indent" -eq "$types_indent" ] && [[ ! "$trimmed" =~ ^- ]]; }; then
            in_types_block=0
          else
            item="${trimmed%%#*}"
            item="${item%"${item##*[![:blank:]]}"}"
            item="${item#-}"
            item="${item#"${item%%[![:blank:]]*}"}"
            item="${item#[\'\"]}"
            item="${item%[\'\"]}"
            if [ "$item" = "ready_for_review" ]; then
              pr_has_ready=1
            fi
            continue
          fi
        fi

        # Why: match types: declaration in flow form or beginning of block-list.
        if [[ "$trimmed" != -* ]] && [[ "$trimmed" == *:* ]]; then
          key="${trimmed%%:*}"
          clean_key="${key#[\'\"]}"
          clean_key="${clean_key%[\'\"]}"
          rest="${trimmed#*:}"

          if [ "$clean_key" = "types" ]; then
            types_indent="$indent"
            clean_rest="${rest%%#*}"
            if [[ "$clean_rest" =~ \[([^]]*)\] ]]; then
              flow_items="${BASH_REMATCH[1]}"
              flow_clean="${flow_items//[\',\"\[\]]/ }"
              flow_clean="${flow_clean//,/ }"
              if [[ " $flow_clean " =~ [[:space:]]ready_for_review[[:space:]] ]]; then
                pr_has_ready=1
              fi
            else
              in_types_block=1
            fi
            continue
          fi
        fi
      fi
    fi
  done

  # Why: check any pull_request block still open at end of file.
  if [ "$in_pr" -eq 1 ]; then
    if [ "$pr_has_ready" -eq 0 ]; then
      file_miss=1
    else
      file_prs=$((file_prs + 1))
    fi
  fi

  if [ "$file_miss" -eq 1 ]; then
    printf 'workflow-ready: %s: pull_request trigger lacks ready_for_review\n' "$rel" >&2
    misses=$((misses + 1))
  else
    ok_prs=$((ok_prs + file_prs))
  fi
done

if [ "$misses" -gt 0 ]; then
  exit 1
fi

echo "workflow-ready: $ok_prs pull_request triggers ok ($files_checked workflows checked)"
