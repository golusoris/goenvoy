#!/usr/bin/env python3
# SPDX-License-Identifier: EUPL-1.2
"""Agent PreToolUse evasion interceptor (HISS), written by praetorctl adopt.

Wire this script as a PreToolUse hook of the agent harness's shell tool. The harness passes
the pending tool call as one JSON object on stdin with the shell command at
tool_input.command; a command may instead be passed as arguments. Exit code 2 blocks the
call and returns the reason to the agent. Input of any other shape is refused, not allowed.

The rules and their refusals are the engine's built-in command policy (internal/agenthook),
rendered at adoption; operator rules belong in hooks.command_policy.deny of .standards.yaml.

A git hook cannot observe --no-verify because git skips hooks entirely, so this script
is deliberately not part of lefthook.yml.
"""

import json
import os
import re
import sys

MAX_INPUT_BYTES = 1048576
MAX_SCAN_CHARS = 65536
MAX_SCAN_LINE_CHARS = 2048
MAX_REASON_BYTES = 4096
BLOCK_EXIT = 2

NO_INPUT_REFUSAL = (
    "[BLOCKED BY HISS] input: PreToolUse JSON " "on stdin or command arguments required"
)
INVALID_INPUT_REFUSAL = "[BLOCKED BY HISS] Invalid hook input: "
NARROWING_REFUSAL = "[BLOCKED BY HISS] hook exclusions: prohibited."
SCAN_BOUND_REFUSAL = (
    "[BLOCKED BY HISS] command exceeds scan bound: at most 65536 characters, 2048 per "
    "line; split command or write long content to file first"
)
HISS_REFUSAL = (
    "[BLOCKED BY HISS] verification evasion prohibited; commits, pushes and tool calls "
    "pass verification gates; pattern: "
)
DEV_01_REFUSAL = (
    "[BLOCKED BY DEV-01] adoption or needs target: workstation dev root; repositories "
    "live inside organization folders as leaf Git repositories; pattern: "
)

RULES = [
    (r"--no-v(e(r(i(f(y)?)?)?)?)?\b", HISS_REFUSAL),
    (
        r"\bgit([ \t]+-[Cc][ \t]+(\x22[^\x22]*\x22|\x27[^\x27]*\x27|[^ \t\n\x22\x27][^"
        r" \t\n]*)|[ \t]+(--[A-Za-z][-A-Za-z]*|-[ABD-Zabd-z][-A-Za-z]*|-[Cc][A-Za-z]+)"
        r"(=[^ \t\n]+)?)*\s+(commit\b[^\n]*\s-[aeiopqsvz]*|am\b[^\n]*\s-[3cikmqsu]*)n",
        HISS_REFUSAL,
    ),
    (r"LEFTHOOK=[\x22\x27]?(0|false)\b", HISS_REFUSAL),
    (r"SKIP=.*git", HISS_REFUSAL),
    (r"(?i:core\.hookspath)(\s*=|\s+[\x22\x27]?[/~.$A-Za-z_\\])", HISS_REFUSAL),
    (
        r"\b(?i:rm|rmdir|unlink|mv|cp|ln|chmod|chown|chattr|truncate|shred|tee|del|era"
        r"se|rd|ri|remove-item|move|move-item|ren|rename|rename-item|copy|copy-item|se"
        r"t-content|add-content|out-file|icacls|attrib)\b[^\n]*(?i:\.git[/\\]hooks)",
        HISS_REFUSAL,
    ),
    (
        r"(?m)^(?:(?:[^\Wsp]\w*|s(?:[^\We]\w*|e(?:[^\Wd]\w*|d\w+)?)?|p(?:[^\We]\w*|e(?"
        r":[^\Wr]\w*|r(?:[^\Wl]\w*|l\w+)?)?)?)?[^\w\n])*(sed|perl)\b[^\n]*\s(-[A-Za-z]"
        r"*i|--in-place)[^\n]*(?i:\.git[/\\]hooks)",
        HISS_REFUSAL,
    ),
    (
        r"(?m)^(?:(?:[^\Wf]\w*|f(?:[^\Wi]\w*|i(?:[^\Wn]\w*|n(?:[^\Wd]\w*|d\w+)?)?)?)?["
        r"^\w\n])*(find)\b[^\n]*(?i:\.git[/\\]hooks)[^\n]*\s-(delete|exec|execdir|ok)"
        r"\b",
        HISS_REFUSAL,
    ),
    (r">\s*[\x22\x27]?[^ \t\n\x22\x27]*(?i:\.git[/\\]hooks)", HISS_REFUSAL),
    (r"\blefthook\s+uninstall\b", HISS_REFUSAL),
    (
        r"(?m)^(?:(?:[^\Wg]\w*|g(?:[^\Wh]\w*|h\w+)?)?[^\w\n])*(gh)\b[^\n]*(?:\bpr\s+me"
        r"rge\b[^\n]*(?:\s|[\x22\x27])--admin(?:\s|$|[;&|\x22\x27]|=true\b)|(?:\s|[\x2"
        r"2\x27])--admin(?:\s|$|[;&|\x22\x27]|=true\b)[^\n]*\bpr\s+merge\b|\bpr\s+--ad"
        r"min(?:\s|$|[;&|\x22\x27]|=true\b)[^\n]*\bmerge\b)",
        HISS_REFUSAL,
    ),
    (
        r"(?m)^(?:(?:[^\Wg]\w*|g(?:[^\Wh]\w*|h\w+)?)?[^\w\n])*(gh)\b[^\n]*\bapi\b[^\n]"
        r"*\bpulls/[0-9]+/merge/?(?:\?|\s|$|[;&|\x22\x27])",
        HISS_REFUSAL,
    ),
    (
        r"(?m)^(?:(?:[^\Wg]\w*|g(?:[^\Wh]\w*|h\w+)?)?[^\w\n])*(gh)\b[^\n]*\bapi\s+grap"
        r"hql\b[^\n]*\bmergePullRequest\b",
        HISS_REFUSAL,
    ),
    (
        r"(?m)^(?:(?:[^\Wcw]\w*|c(?:[^\Wu]\w*|u(?:[^\Wr]\w*|r(?:[^\Wl]\w*|l\w+)?)?)?|w"
        r"(?:[^\Wg]\w*|g(?:[^\We]\w*|e(?:[^\Wt]\w*|t\w+)?)?)?)?[^\w\n])*(curl|wget)\b["
        r"^\n]*\bapi\.github\.com/[^\n]*\bpulls/[0-9]+/merge/?(?:\?|\s|$|[;&|\x22\x27]"
        r")",
        HISS_REFUSAL,
    ),
    (
        r"\A(?:[^\x22\x27\\]|\\(?s:.)|\x27[^\x27]*\x27|\x22(?:[^\x22\\]|\\(?s:.))*\x22"
        r")*(?:\A|[\n;&|(\x60]|\x22(?:[^\x22\\]|\\(?s:.))*(?:\$\(|\x60))[ \t]*(?:[A-Za"
        r"-z_][A-Za-z0-9_]*=(?:(?:\x22[^\x22\n]*\x22|\x27[^\x27\n]*\x27|[^ \t\n;&|()\x"
        r"60\x22\x27]+))?[ \t]+|(?:[^ \t\n;&|()\x60\x22\x27]*/)?(?:builtin)(?:[ \t]+(?"
        r":-(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)))*[ \t]+|(?:[^ \t"
        r"\n;&|()\x60\x22\x27]*/)?(?:chrt)(?:[ \t]+(?:-(?:[^ \t\n;&|()\x60\x22\x27][^ "
        r"\t\n;&|()\x60\x22\x27]*)))*[ \t]+(?:\x22[^\x22\n]*\x22|\x27[^\x27\n]*\x27|[^"
        r" \t\n;&|()\x60\x22\x27]+)[ \t]+|(?:[^ \t\n;&|()\x60\x22\x27]*/)?(?:command)("
        r"?:[ \t]+(?:-(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)))*[ \t]+"
        r"|(?:[^ \t\n;&|()\x60\x22\x27]*/)?(?:env)(?:[ \t]+(?:(?:-u|--unset|-C|--chdir"
        r")[ \t]+(?:\x22[^\x22\n]*\x22|\x27[^\x27\n]*\x27|[^- \t\n;&|()\x60\x22\x27][^"
        r" \t\n;&|()\x60\x22\x27]*|-[0-9]+)|-(?:\x2d(?:\x63(?:\x68(?:\x64(?:\x69(?:\x7"
        r"2(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x2"
        r"2\x27\x72][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x69][^ \t\n;"
        r"&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x64][^ \t\n;&|()\x60\x22\x27]*"
        r")?|[^ \t\n;&|()\x60\x22\x27\x68][^ \t\n;&|()\x60\x22\x27]*)?|\x75(?:\x6e(?:"
        r"\x73(?:\x65(?:\x74(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^"
        r" \t\n;&|()\x60\x22\x27\x74][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22"
        r"\x27\x65][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x73][^ \t\n;&"
        r"|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6e][^ \t\n;&|()\x60\x22\x27]*)"
        r"?|[^ \t\n;&|()\x60\x22\x27\x63\x75][^ \t\n;&|()\x60\x22\x27]*)?|\x43(?:[^ \t"
        r"\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|\x75(?:[^ \t\n;&|()\x60\x22"
        r"\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x2d\x43\x75][^ \t"
        r"\n;&|()\x60\x22\x27]*)))*[ \t]+|(?:[^ \t\n;&|()\x60\x22\x27]*/)?(?:exec)(?:["
        r" \t]+(?:(?:-a)[ \t]+(?:\x22[^\x22\n]*\x22|\x27[^\x27\n]*\x27|[^- \t\n;&|()\x"
        r"60\x22\x27][^ \t\n;&|()\x60\x22\x27]*|-[0-9]+)|-(?:\x61(?:[^ \t\n;&|()\x60\x"
        r"22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x61][^ \t\n;&|()"
        r"\x60\x22\x27]*)))*[ \t]+|(?:[^ \t\n;&|()\x60\x22\x27]*/)?(?:flock)(?:[ \t]+("
        r"?:(?:-w|--timeout|-E|--conflict-exit-code)[ \t]+(?:\x22[^\x22\n]*\x22|\x27[^"
        r"\x27\n]*\x27|[^- \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*|-[0-9]+)|-"
        r"(?:\x2d(?:\x63(?:\x6f(?:\x6e(?:\x66(?:\x6c(?:\x69(?:\x63(?:\x74(?:\x2d(?:\x6"
        r"5(?:\x78(?:\x69(?:\x74(?:\x2d(?:\x63(?:\x6f(?:\x64(?:\x65(?:[^ \t\n;&|()\x60"
        r"\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x65][^ \t\n;&|"
        r"()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x64][^ \t\n;&|()\x60\x22\x27]*)?"
        r"|[^ \t\n;&|()\x60\x22\x27\x6f][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60"
        r"\x22\x27\x63][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x2d][^ \t"
        r"\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x74][^ \t\n;&|()\x60\x22\x2"
        r"7]*)?|[^ \t\n;&|()\x60\x22\x27\x69][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()"
        r"\x60\x22\x27\x78][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x65]["
        r"^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x2d][^ \t\n;&|()\x60\x2"
        r"2\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x74][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;"
        r"&|()\x60\x22\x27\x63][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x"
        r"69][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6c][^ \t\n;&|()\x6"
        r"0\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x66][^ \t\n;&|()\x60\x22\x27]*)?|[^ "
        r"\t\n;&|()\x60\x22\x27\x6e][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22"
        r"\x27\x6f][^ \t\n;&|()\x60\x22\x27]*)?|\x74(?:\x69(?:\x6d(?:\x65(?:\x6f(?:\x7"
        r"5(?:\x74(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()"
        r"\x60\x22\x27\x74][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x75]["
        r"^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6f][^ \t\n;&|()\x60\x2"
        r"2\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x65][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;"
        r"&|()\x60\x22\x27\x6d][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x"
        r"69][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x63\x74][^ \t\n;&|("
        r")\x60\x22\x27]*)?|\x45(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*"
        r")|\x77(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x"
        r"60\x22\x27\x2d\x45\x77][^ \t\n;&|()\x60\x22\x27]*)))*[ \t]+(?:\x22[^\x22\n]*"
        r"\x22|\x27[^\x27\n]*\x27|[^ \t\n;&|()\x60\x22\x27]+)[ \t]+|(?:[^ \t\n;&|()\x6"
        r"0\x22\x27]*/)?(?:ionice)(?:[ \t]+(?:(?:-c|--class|-n|--classdata|-p|--pid|-P"
        r"|--pgid|-u|--uid)[ \t]+(?:\x22[^\x22\n]*\x22|\x27[^\x27\n]*\x27|[^- \t\n;&|("
        r")\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*|-[0-9]+)|-(?:\x2d(?:\x63(?:\x6c(?:"
        r"\x61(?:\x73(?:\x73(?:\x64(?:\x61(?:\x74(?:\x61(?:[^ \t\n;&|()\x60\x22\x27][^"
        r" \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x61][^ \t\n;&|()\x60\x22"
        r"\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x74][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&"
        r"|()\x60\x22\x27\x61][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6"
        r"4][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x73][^ \t\n;&|()\x60"
        r"\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x73][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t"
        r"\n;&|()\x60\x22\x27\x61][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x2"
        r"7\x6c][^ \t\n;&|()\x60\x22\x27]*)?|\x70(?:\x67(?:\x69(?:\x64(?:[^ \t\n;&|()"
        r"\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x64][^ \t"
        r"\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x69][^ \t\n;&|()\x60\x22\x2"
        r"7]*)?|\x69(?:\x64(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ "
        r"\t\n;&|()\x60\x22\x27\x64][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22"
        r"\x27\x67\x69][^ \t\n;&|()\x60\x22\x27]*)?|\x75(?:\x69(?:\x64(?:[^ \t\n;&|()"
        r"\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x64][^ \t"
        r"\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x69][^ \t\n;&|()\x60\x22\x2"
        r"7]*)?|[^ \t\n;&|()\x60\x22\x27\x63\x70\x75][^ \t\n;&|()\x60\x22\x27]*)?|\x50"
        r"(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|\x63(?:[^ \t\n;&|()"
        r"\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|\x6e(?:[^ \t\n;&|()\x60\x22\x27][^ "
        r"\t\n;&|()\x60\x22\x27]*)|\x70(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x2"
        r"2\x27]*)|\x75(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n"
        r";&|()\x60\x22\x27\x2d\x50\x63\x6e\x70\x75][^ \t\n;&|()\x60\x22\x27]*)))*[ \t"
        r"]+|(?:[^ \t\n;&|()\x60\x22\x27]*/)?(?:nice)(?:[ \t]+(?:(?:-n|--adjustment)[ "
        r"\t]+(?:\x22[^\x22\n]*\x22|\x27[^\x27\n]*\x27|[^- \t\n;&|()\x60\x22\x27][^ \t"
        r"\n;&|()\x60\x22\x27]*|-[0-9]+)|-(?:\x2d(?:\x61(?:\x64(?:\x6a(?:\x75(?:\x73(?"
        r":\x74(?:\x6d(?:\x65(?:\x6e(?:\x74(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x6"
        r"0\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x74][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t"
        r"\n;&|()\x60\x22\x27\x6e][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x2"
        r"7\x65][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6d][^ \t\n;&|()"
        r"\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x74][^ \t\n;&|()\x60\x22\x27]*)?|["
        r"^ \t\n;&|()\x60\x22\x27\x73][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x2"
        r"2\x27\x75][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6a][^ \t\n;"
        r"&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x64][^ \t\n;&|()\x60\x22\x27]*"
        r")?|[^ \t\n;&|()\x60\x22\x27\x61][^ \t\n;&|()\x60\x22\x27]*)?|\x6e(?:[^ \t\n;"
        r"&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x2d\x6"
        r"e][^ \t\n;&|()\x60\x22\x27]*)))*[ \t]+|(?:[^ \t\n;&|()\x60\x22\x27]*/)?(?:no"
        r"hup)(?:[ \t]+(?:-(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)))*["
        r" \t]+|(?:[^ \t\n;&|()\x60\x22\x27]*/)?(?:setsid)(?:[ \t]+(?:-(?:[^ \t\n;&|()"
        r"\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)))*[ \t]+|(?:[^ \t\n;&|()\x60\x22\x2"
        r"7]*/)?(?:stdbuf)(?:[ \t]+(?:(?:-i|-o|-e)[ \t]+(?:\x22[^\x22\n]*\x22|\x27[^\x"
        r"27\n]*\x27|[^- \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*|-[0-9]+)|-(?"
        r":\x65(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|\x69(?:[^ \t\n;"
        r"&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|\x6f(?:[^ \t\n;&|()\x60\x22\x27"
        r"][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x65\x69\x6f][^ \t\n;&|"
        r"()\x60\x22\x27]*)))*[ \t]+|(?:[^ \t\n;&|()\x60\x22\x27]*/)?(?:systemd-run)(?"
        r":[ \t]+(?:(?:-p|--property|-u|--unit|-E|--setenv|--slice|-M|--machine|-H|--h"
        r"ost|--uid|--gid|--nice|--working-directory|--description)[ \t]+(?:\x22[^\x22"
        r"\n]*\x22|\x27[^\x27\n]*\x27|[^- \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x"
        r"27]*|-[0-9]+)|-(?:\x2d(?:\x64(?:\x65(?:\x73(?:\x63(?:\x72(?:\x69(?:\x70(?:\x"
        r"74(?:\x69(?:\x6f(?:\x6e(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]"
        r"*)|[^ \t\n;&|()\x60\x22\x27\x6e][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x6"
        r"0\x22\x27\x6f][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x69][^ "
        r"\t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x74][^ \t\n;&|()\x60\x22"
        r"\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x70][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&"
        r"|()\x60\x22\x27\x69][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x7"
        r"2][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x63][^ \t\n;&|()\x60"
        r"\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x73][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t"
        r"\n;&|()\x60\x22\x27\x65][^ \t\n;&|()\x60\x22\x27]*)?|\x67(?:\x69(?:\x64(?:[^"
        r" \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27"
        r"\x64][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x69][^ \t\n;&|()"
        r"\x60\x22\x27]*)?|\x68(?:\x6f(?:\x73(?:\x74(?:[^ \t\n;&|()\x60\x22\x27][^ \t"
        r"\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x74][^ \t\n;&|()\x60\x22\x27"
        r"]*)?|[^ \t\n;&|()\x60\x22\x27\x73][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()"
        r"\x60\x22\x27\x6f][^ \t\n;&|()\x60\x22\x27]*)?|\x6d(?:\x61(?:\x63(?:\x68(?:\x"
        r"69(?:\x6e(?:\x65(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ "
        r"\t\n;&|()\x60\x22\x27\x65][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22"
        r"\x27\x6e][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x69][^ \t\n;&"
        r"|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x68][^ \t\n;&|()\x60\x22\x27]*)"
        r"?|[^ \t\n;&|()\x60\x22\x27\x63][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60"
        r"\x22\x27\x61][^ \t\n;&|()\x60\x22\x27]*)?|\x6e(?:\x69(?:\x63(?:\x65(?:[^ \t"
        r"\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x65"
        r"][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x63][^ \t\n;&|()\x60"
        r"\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x69][^ \t\n;&|()\x60\x22\x27]*)?|\x70("
        r"?:\x72(?:\x6f(?:\x70(?:\x65(?:\x72(?:\x74(?:\x79(?:[^ \t\n;&|()\x60\x22\x27]"
        r"[^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x79][^ \t\n;&|()\x60\x2"
        r"2\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x74][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;"
        r"&|()\x60\x22\x27\x72][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x"
        r"65][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x70][^ \t\n;&|()\x6"
        r"0\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6f][^ \t\n;&|()\x60\x22\x27]*)?|[^ "
        r"\t\n;&|()\x60\x22\x27\x72][^ \t\n;&|()\x60\x22\x27]*)?|\x73(?:\x65(?:\x74(?:"
        r"\x65(?:\x6e(?:\x76(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^"
        r" \t\n;&|()\x60\x22\x27\x76][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22"
        r"\x27\x6e][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x65][^ \t\n;&"
        r"|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x74][^ \t\n;&|()\x60\x22\x27]*)"
        r"?|\x6c(?:\x69(?:\x63(?:\x65(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22"
        r"\x27]*)|[^ \t\n;&|()\x60\x22\x27\x65][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|"
        r"()\x60\x22\x27\x63][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x69"
        r"][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x65\x6c][^ \t\n;&|()"
        r"\x60\x22\x27]*)?|\x75(?:\x69(?:\x64(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()"
        r"\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x64][^ \t\n;&|()\x60\x22\x27]*)?|\x"
        r"6e(?:\x69(?:\x74(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ "
        r"\t\n;&|()\x60\x22\x27\x74][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22"
        r"\x27\x69][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x69\x6e][^ \t"
        r"\n;&|()\x60\x22\x27]*)?|\x77(?:\x6f(?:\x72(?:\x6b(?:\x69(?:\x6e(?:\x67(?:\x2"
        r"d(?:\x64(?:\x69(?:\x72(?:\x65(?:\x63(?:\x74(?:\x6f(?:\x72(?:\x79(?:[^ \t\n;&"
        r"|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x79][^ "
        r"\t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x72][^ \t\n;&|()\x60\x22"
        r"\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6f][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&"
        r"|()\x60\x22\x27\x74][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6"
        r"3][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x65][^ \t\n;&|()\x60"
        r"\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x72][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t"
        r"\n;&|()\x60\x22\x27\x69][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x2"
        r"7\x64][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x2d][^ \t\n;&|()"
        r"\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x67][^ \t\n;&|()\x60\x22\x27]*)?|["
        r"^ \t\n;&|()\x60\x22\x27\x6e][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x2"
        r"2\x27\x69][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6b][^ \t\n;"
        r"&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x72][^ \t\n;&|()\x60\x22\x27]*"
        r")?|[^ \t\n;&|()\x60\x22\x27\x6f][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x6"
        r"0\x22\x27\x64\x67\x68\x6d\x6e\x70\x73\x75\x77][^ \t\n;&|()\x60\x22\x27]*)?|"
        r"\x45(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|\x48(?:[^ \t\n;&"
        r"|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|\x4d(?:[^ \t\n;&|()\x60\x22\x27]"
        r"[^ \t\n;&|()\x60\x22\x27]*)|\x70(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60"
        r"\x22\x27]*)|\x75(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ "
        r"\t\n;&|()\x60\x22\x27\x2d\x45\x48\x4d\x70\x75][^ \t\n;&|()\x60\x22\x27]*)))*"
        r"[ \t]+|(?:[^ \t\n;&|()\x60\x22\x27]*/)?(?:taskset)(?:[ \t]+(?:-(?:[^ \t\n;&|"
        r"()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)))*[ \t]+(?:\x22[^\x22\n]*\x22|\x2"
        r"7[^\x27\n]*\x27|[^ \t\n;&|()\x60\x22\x27]+)[ \t]+|(?:[^ \t\n;&|()\x60\x22\x2"
        r"7]*/)?(?:time)(?:[ \t]+(?:(?:-f|--format|-o|--output)[ \t]+(?:\x22[^\x22\n]*"
        r"\x22|\x27[^\x27\n]*\x27|[^- \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*"
        r"|-[0-9]+)|-(?:\x2d(?:\x66(?:\x6f(?:\x72(?:\x6d(?:\x61(?:\x74(?:[^ \t\n;&|()"
        r"\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x74][^ \t"
        r"\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x61][^ \t\n;&|()\x60\x22\x2"
        r"7]*)?|[^ \t\n;&|()\x60\x22\x27\x6d][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()"
        r"\x60\x22\x27\x72][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6f]["
        r"^ \t\n;&|()\x60\x22\x27]*)?|\x6f(?:\x75(?:\x74(?:\x70(?:\x75(?:\x74(?:[^ \t"
        r"\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x74"
        r"][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x75][^ \t\n;&|()\x60"
        r"\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x70][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t"
        r"\n;&|()\x60\x22\x27\x74][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x2"
        r"7\x75][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x66\x6f][^ \t\n;"
        r"&|()\x60\x22\x27]*)?|\x66(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x2"
        r"7]*)|\x6f(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|("
        r")\x60\x22\x27\x2d\x66\x6f][^ \t\n;&|()\x60\x22\x27]*)))*[ \t]+|(?:[^ \t\n;&|"
        r"()\x60\x22\x27]*/)?(?:timeout)(?:[ \t]+(?:(?:-k|--kill-after|-s|--signal)[ "
        r"\t]+(?:\x22[^\x22\n]*\x22|\x27[^\x27\n]*\x27|[^- \t\n;&|()\x60\x22\x27][^ \t"
        r"\n;&|()\x60\x22\x27]*|-[0-9]+)|-(?:\x2d(?:\x6b(?:\x69(?:\x6c(?:\x6c(?:\x2d(?"
        r":\x61(?:\x66(?:\x74(?:\x65(?:\x72(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x6"
        r"0\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x72][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t"
        r"\n;&|()\x60\x22\x27\x65][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x2"
        r"7\x74][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x66][^ \t\n;&|()"
        r"\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x61][^ \t\n;&|()\x60\x22\x27]*)?|["
        r"^ \t\n;&|()\x60\x22\x27\x2d][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x2"
        r"2\x27\x6c][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6c][^ \t\n;"
        r"&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x69][^ \t\n;&|()\x60\x22\x27]*"
        r")?|\x73(?:\x69(?:\x67(?:\x6e(?:\x61(?:\x6c(?:[^ \t\n;&|()\x60\x22\x27][^ \t"
        r"\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x6c][^ \t\n;&|()\x60\x22\x27"
        r"]*)?|[^ \t\n;&|()\x60\x22\x27\x61][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()"
        r"\x60\x22\x27\x6e][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x67]["
        r"^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x69][^ \t\n;&|()\x60\x2"
        r"2\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6b\x73][^ \t\n;&|()\x60\x22\x27]*)?|\x6b"
        r"(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|\x73(?:[^ \t\n;&|()"
        r"\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x2d\x6b\x7"
        r"3][^ \t\n;&|()\x60\x22\x27]*)))*[ \t]+(?:\x22[^\x22\n]*\x22|\x27[^\x27\n]*\x"
        r"27|[^ \t\n;&|()\x60\x22\x27]+)[ \t]+|(?:[^ \t\n;&|()\x60\x22\x27]*/)?(?:xarg"
        r"s)(?:[ \t]+(?:(?:-a|--arg-file|-d|--delimiter|-E|-e|-I|-i|-L|-l|-n|--max-arg"
        r"s|-P|--max-procs|-s|--max-chars)[ \t]+(?:\x22[^\x22\n]*\x22|\x27[^\x27\n]*\x"
        r"27|[^- \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*|-[0-9]+)|-(?:\x2d(?:"
        r"\x61(?:\x72(?:\x67(?:\x2d(?:\x66(?:\x69(?:\x6c(?:\x65(?:[^ \t\n;&|()\x60\x22"
        r"\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x65][^ \t\n;&|()\x"
        r"60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6c][^ \t\n;&|()\x60\x22\x27]*)?|[^ "
        r"\t\n;&|()\x60\x22\x27\x69][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22"
        r"\x27\x66][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x2d][^ \t\n;&"
        r"|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x67][^ \t\n;&|()\x60\x22\x27]*)"
        r"?|[^ \t\n;&|()\x60\x22\x27\x72][^ \t\n;&|()\x60\x22\x27]*)?|\x64(?:\x65(?:\x"
        r"6c(?:\x69(?:\x6d(?:\x69(?:\x74(?:\x65(?:\x72(?:[^ \t\n;&|()\x60\x22\x27][^ "
        r"\t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x72][^ \t\n;&|()\x60\x22\x"
        r"27]*)?|[^ \t\n;&|()\x60\x22\x27\x65][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|("
        r")\x60\x22\x27\x74][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x69]"
        r"[^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6d][^ \t\n;&|()\x60\x"
        r"22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x69][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n"
        r";&|()\x60\x22\x27\x6c][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27"
        r"\x65][^ \t\n;&|()\x60\x22\x27]*)?|\x6d(?:\x61(?:\x78(?:\x2d(?:\x61(?:\x72(?:"
        r"\x67(?:\x73(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&"
        r"|()\x60\x22\x27\x73][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x6"
        r"7][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x72][^ \t\n;&|()\x60"
        r"\x22\x27]*)?|\x63(?:\x68(?:\x61(?:\x72(?:\x73(?:[^ \t\n;&|()\x60\x22\x27][^ "
        r"\t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x73][^ \t\n;&|()\x60\x22\x"
        r"27]*)?|[^ \t\n;&|()\x60\x22\x27\x72][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|("
        r")\x60\x22\x27\x61][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x68]"
        r"[^ \t\n;&|()\x60\x22\x27]*)?|\x70(?:\x72(?:\x6f(?:\x63(?:\x73(?:[^ \t\n;&|()"
        r"\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x73][^ \t"
        r"\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x63][^ \t\n;&|()\x60\x22\x2"
        r"7]*)?|[^ \t\n;&|()\x60\x22\x27\x6f][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()"
        r"\x60\x22\x27\x72][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x61\x"
        r"63\x70][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x2d][^ \t\n;&|("
        r")\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x22\x27\x78][^ \t\n;&|()\x60\x22\x27]*)?|"
        r"[^ \t\n;&|()\x60\x22\x27\x61][^ \t\n;&|()\x60\x22\x27]*)?|[^ \t\n;&|()\x60\x"
        r"22\x27\x61\x64\x6d][^ \t\n;&|()\x60\x22\x27]*)?|\x45(?:[^ \t\n;&|()\x60\x22"
        r"\x27][^ \t\n;&|()\x60\x22\x27]*)|\x49(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|("
        r")\x60\x22\x27]*)|\x4c(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)"
        r"|\x50(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|\x61(?:[^ \t\n;"
        r"&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|\x64(?:[^ \t\n;&|()\x60\x22\x27"
        r"][^ \t\n;&|()\x60\x22\x27]*)|\x65(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x6"
        r"0\x22\x27]*)|\x69(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|\x6"
        r"c(?:[^ \t\n;&|()\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|\x6e(?:[^ \t\n;&|()"
        r"\x60\x22\x27][^ \t\n;&|()\x60\x22\x27]*)|\x73(?:[^ \t\n;&|()\x60\x22\x27][^ "
        r"\t\n;&|()\x60\x22\x27]*)|[^ \t\n;&|()\x60\x22\x27\x2d\x45\x49\x4c\x50\x61\x6"
        r"4\x65\x69\x6c\x6e\x73][^ \t\n;&|()\x60\x22\x27]*)))*[ \t]+)*(?:[^ \t\n;&|()"
        r"\x60\x22\x27]*/)?(?:(?:(?:sudo|doas|pkexec|run0)(?:[ \t\r\n;&|()\x60]|$)|su("
        r"?:[ \t]+(?:\x22[^\x22\n]*\x22|\x27[^\x27\n]*\x27|[^ \t\n;&|()\x60\x22\x27]+)"
        r")*[ \t]+(?:-c[^ \t\n;&|()\x60]*|--command(?:=[^ \t\n;&|()\x60]*)?)(?:[ \t\r"
        r"\n;&|()\x60]|$)))",
        HISS_REFUSAL,
    ),
    (
        r"(?i)(standardsctl|praetorctl)\s+(adopt|conform|bootstrap|needs\s+(scan|repor"
        r"t|migrate|epic))\b.*\bdev[/\\]?(\s|$)",
        DEV_01_REFUSAL,
    ),
]

RULE_PREFILTERS = {
    RULES[14][0]: ("sudo", "doas", "pkexec", "run0", "su"),
}


# The RULES judged without the words of read-only commands chained after a commit. Each
# pattern has a name of its own: wrapped inside the list, it would read as two items with
# the comma between them forgotten.
READ_ONLY_EXEMPT_1 = (
    r"\bgit([ \t]+-[Cc][ \t]+(\x22[^\x22]*\x22|\x27[^\x27]*\x27|[^ \t\n\x22\x27][^ \t\n"
    r"]*)|[ \t]+(--[A-Za-z][-A-Za-z]*|-[ABD-Zabd-z][-A-Za-z]*|-[Cc][A-Za-z]+)(=[^ \t\n]"
    r"+)?)*\s+(commit\b[^\n]*\s-[aeiopqsvz]*|am\b[^\n]*\s-[3cikmqsu]*)n"
)
READ_ONLY_EXEMPT_2 = r"SKIP=.*git"
READ_ONLY_EXEMPT = [
    READ_ONLY_EXEMPT_1,
    READ_ONLY_EXEMPT_2,
]
READ_ONLY_WORDS = (
    r"((?:;|&&|\|\|?)[ \t]*(?:git[ \t]+(?:--no-pager[ \t]+)?(?:log|show|status|diff|rev"
    r"-parse)|head|tail|grep|wc))(?:[ \t]+[-A-Za-z0-9_./:=@,+~*?]+)*|((?:;|&&|\|\|?)[ "
    r"\t]*sed)[ \t]+-n\b(?:[ \t]+[-A-Za-z0-9_./:=@,+~*?]+)*"
)
READ_ONLY_VETO = (
    r"[^\t\x20-\x7e]|[\\\x60$(){}<>^!]|--%|%[;&|]|(?i:(?:\A|[^-A-Za-z0-9_])(?:export|de"
    r"clare|typeset|set|setenv|alias|function|sal|nal|set-alias|new-alias|doskey|hash)("
    r"?:[^-A-Za-z0-9_]|$))"
)

LEFTHOOK_DISABLED = [
    ("0", "[BLOCKED BY HISS] LEFTHOOK=0 detected in environment. Evasion prohibited."),
    (
        "false",
        "[BLOCKED BY HISS] LEFTHOOK=false detected in environment. Evasion prohibited.",
    ),
]
LEFTHOOK_NARROWING = ("LEFTHOOK_EXCLUDE", "LEFTHOOK_SKIP")


class Blocked(Exception):
    pass


def read_payload_command(stream):
    raw = stream.read(MAX_INPUT_BYTES + 1)
    if len(raw) > MAX_INPUT_BYTES:
        raise ValueError("hook input exceeds " + str(MAX_INPUT_BYTES) + " bytes")
    payload = json.loads(raw.decode("utf-8"))
    if not isinstance(payload, dict):
        raise ValueError("hook input must be one JSON object")
    tool_input = payload.get("tool_input")
    if not isinstance(tool_input, dict):
        raise ValueError("tool_input must be an object")
    command = tool_input.get("command")
    if not isinstance(command, str) or not command.strip():
        raise ValueError("tool_input.command must be nonempty text")
    return command


def pending_command():
    if len(sys.argv) > 1:
        return " ".join(sys.argv[1:])
    if sys.stdin.isatty():
        raise Blocked(NO_INPUT_REFUSAL)
    try:
        return read_payload_command(sys.stdin.buffer)
    except (ValueError, OSError, RecursionError) as error:
        raise Blocked(INVALID_INPUT_REFUSAL + str(error)) from error


def check_environment(environ):
    value = environ.get("LEFTHOOK")
    for disabled, refusal in LEFTHOOK_DISABLED:
        if value == disabled:
            raise Blocked(refusal)
    for name in LEFTHOOK_NARROWING:
        if environ.get(name):
            raise Blocked(NARROWING_REFUSAL)


def without_read_only_words(command):
    # The words of git log, head, sed -n and the other read-only commands READ_ONLY_WORDS
    # names, chained after ;, &&, || or |, belong to that command and are dropped, its name
    # kept. A command holding a construct READ_ONLY_VETO names keeps every word.
    if re.search(READ_ONLY_VETO, command):
        return command
    return re.sub(READ_ONLY_WORDS, r"\1\2", command)


def join_line_continuations(command):
    # A backslash followed by a newline (LF or CRLF) joins lines without regex backtracking.
    return re.sub(r"\\\r?\n", " ", command)


# The git rule with an option run (the short skip flag rule) is searched once per run (#1283).
# re retries a rule at every git it finds, and a git inside an option token (-git, --x=git,
# -C dir/git) starts at a boundary of the run an earlier git already walked: each retry walked
# the rest of that run again, quadratic in the line. A start at a walked boundary reaches the
# same end, and only a run's end can be followed by the rest of the rule (every option starts
# with a blank and a dash), so skipping it keeps re.search's answer. The adopter interceptor
# carries this code byte for byte (TestEmittedInterceptorSearchesLikeTheEngineGuard).
GIT_OPTION_RUN_HEAD = r"\bgit("
GIT_OPTION_RUN_JOIN = r")*\s+("


def search(pattern, text):
    """re.search(pattern, text) is not None, with the git option-run rule in linear time."""
    literals = RULE_PREFILTERS.get(pattern, ())
    if literals and not any(literal in text for literal in literals):
        return False
    head, join, tail = pattern.partition(GIT_OPTION_RUN_JOIN)
    if not join or not head.startswith(GIT_OPTION_RUN_HEAD):
        return re.search(pattern, text) is not None
    first = len(GIT_OPTION_RUN_HEAD)
    option = re.compile(head[first:])
    rest = re.compile(r"\s+(" + tail)
    walked = set()
    for found in re.finditer(r"\bgit", text):
        position = found.end()
        if position in walked:
            continue
        while position not in walked:
            walked.add(position)
            step = option.match(text, position)
            if step is None or step.end() == position:
                break
            position = step.end()
        if rest.match(text, position):
            return True
    return False


def check_command(command):
    command = join_line_continuations(command)
    # re backtracks, so a longer command or line could stall this script past the harness's
    # hook timeout. Such a command is refused, never truncated.
    if len(command) > MAX_SCAN_CHARS:
        raise Blocked(SCAN_BOUND_REFUSAL)
    if max(len(line) for line in command.split("\n")) > MAX_SCAN_LINE_CHARS:
        raise Blocked(SCAN_BOUND_REFUSAL)
    judged = without_read_only_words(command)
    for pattern, refusal in RULES:
        text = judged if pattern in READ_ONLY_EXEMPT else command
        if search(pattern, text):
            reason = (
                (refusal + pattern).encode()[:MAX_REASON_BYTES].decode(errors="ignore")
            )
            raise Blocked(reason)


def main():
    try:
        check_environment(os.environ)
        check_command(pending_command())
    except Blocked as blocked:
        sys.stderr.write(str(blocked) + "\n")
        sys.exit(BLOCK_EXIT)
    sys.exit(0)


if __name__ == "__main__":
    main()
