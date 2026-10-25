#!/usr/bin/env python3
"""Read-only periodic checkpoint and draft-publication planner."""

import argparse
import json
import os
import re
import stat
import sys
import tempfile
import threading
import time
from pathlib import Path
from urllib.parse import urlsplit

from common import HookBudget, HookError, prepare_checkpoint_session, run_bounded

MAX_OUTPUT = 1024 * 1024
MAX_PATHS = 10000
MAX_WORKTREES = 512
MAX_UNFINISHED_WORKTREES = 20
MAX_WORKTREE_ACTIONS = 10
OPEN_ITEMS_LIMIT = 1024 * 1024
SESSION_ENV = "PRAETOR_CHECKPOINT_SESSION"
MAX_RECORD_BYTES = 512 * 1024
OWNERSHIP_UNKNOWN_REASON = "worktree ownership unknown: judged by recent activity"
SESSION_HASH_RE = re.compile(r"^[0-9a-f]{16}$")
# The hook has no session-start timestamp; recent Git admin-file activity stands in for
# "created or changed by this session" and deliberately excludes older inventory worktrees.
WORKTREE_ACTIVITY_WINDOW = 24 * 60 * 60
WORKTREE_CHECK_BUDGET = 22
WORKTREE_CHECK_WORKERS = 8
WORKTREE_COMMAND_TIMEOUT = 1
# Seconds per Git call of the landed-by-content check, which diffs and applies a whole branch.
LANDED_CHECK_TIMEOUT = 5
PULL_REQUEST_TIMEOUT = 2
# gh 2.100.0 requests the first 100 contexts for pr list; equality is incomplete.
MAX_CHECKS = 100
# GitHub reports check-run start times in UTC with second precision.
START_TIME = re.compile(r"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z")
TIMEOUT = 5
PUBLIC_ACTIONS = [
    "Review the owned public paths and run the required verification gates.",
    "Create a normal signed-off commit after review: git commit -s.",
]
PUSH_ACTION = "Push the reviewed checkpoint with normal git push after verification."
PR_ACTION = "Create a draft PR with an explicit head and base after review."
# The hosted gates Praetor emits fail on a draft by design: their first step prints an error
# annotation with exactly this title and message and exits 1, and the gate runs once the pull
# request is marked ready (internal/ghworkflow/hostedgate.go; a Go test holds both lines to it).
DRAFT_GATE_TITLE = "Gate not run on a draft"
DRAFT_GATE_MESSAGE = (
    "The gate did not run because the pull request is a draft."
    " It runs when the pull request is marked ready for review."
)
# Bounds of the one annotation read on a draft: check suites, failed runs per suite, and
# annotations per run. The draft annotation is the first one a refused job writes.
MAX_DRAFT_SUITES = 50
MAX_DRAFT_RUNS = 50
MAX_DRAFT_ANNOTATIONS = 10
DRAFT_GATE_QUERY = (
    "query($owner: String!, $name: String!, $oid: GitObjectID!) {"
    " repository(owner: $owner, name: $name) { object(oid: $oid) { ... on Commit {"
    f" checkSuites(first: {MAX_DRAFT_SUITES}) {{ nodes {{"
    f" checkRuns(first: {MAX_DRAFT_RUNS},"
    " filterBy: {checkType: LATEST, conclusions: [FAILURE]}) {"
    f" nodes {{ name annotations(first: {MAX_DRAFT_ANNOTATIONS})"
    " { nodes { title message } } }"
    " } } } } } } }"
)
OID = re.compile(r"^(?:[0-9a-f]{40}|[0-9a-f]{64})$")
REPO_PART = re.compile(r"^[A-Za-z0-9_.-]{1,100}$")
REF_NAME = re.compile(r"^[A-Za-z0-9][A-Za-z0-9._/-]{0,199}$")


class CheckpointError(Exception):
    """A bounded, actionable planner failure."""


def _run(argv, root, *, allowed=(0,), network=False, timeout=None, index=None):
    env = dict(os.environ)
    for key in (
        "GIT_DIR",
        "GIT_WORK_TREE",
        "GIT_INDEX_FILE",
        "GIT_COMMON_DIR",
        "GIT_EXTERNAL_DIFF",
        "GIT_DIFF_OPTS",
    ):
        env.pop(key, None)
    env.update(
        GIT_TERMINAL_PROMPT="0", GH_PROMPT_DISABLED="1", PYTHONDONTWRITEBYTECODE="1"
    )
    if index is not None:
        env["GIT_INDEX_FILE"] = str(index)
    try:
        return run_bounded(
            argv,
            cwd=root,
            timeout=timeout if timeout is not None else (15 if network else TIMEOUT),
            max_output=MAX_OUTPUT,
            allowed=allowed,
            env=env,
        )
    except HookError as error:
        raise CheckpointError(str(error)) from error


def _git(root, *args, **kwargs):
    return _run(["git", *args], root, **kwargs)


# Descriptor-relative opens are how the policy path is confined: each component is opened
# beneath the previous one and never through a link. Windows has no dir_fd support and no
# O_DIRECTORY or O_NOFOLLOW, so there the evaluator raised AttributeError before it read
# anything, and every checkpoint hook in an adopted repository failed.
DESCRIPTOR_RELATIVE = (
    os.open in os.supports_dir_fd
    and hasattr(os, "O_DIRECTORY")
    and hasattr(os, "O_NOFOLLOW")
)
POLICY_PARTS = (".config", "agent")
POLICY_FILE = "checkpoint.json"


def _policy_bytes(root):
    if DESCRIPTOR_RELATIVE:
        return _policy_bytes_relative(root)
    return _policy_bytes_checked(root)


def _policy_bytes_relative(root):
    """Read the policy through descriptor-relative opens, closing every descriptor here.

    One directory is held at a time: a parent is closed once its child is open, and the
    last one once the policy file is open or an open failed. The stream is opened with
    closefd=False and the descriptor closed below, because a wrap that fails (the policy
    path is a directory) leaves a descriptor it was meant to own open.
    """
    directories = os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW
    try:
        directory = os.open(root, directories)
        try:
            for part in POLICY_PARTS:
                parent = directory
                directory = os.open(part, directories, dir_fd=parent)
                os.close(parent)
            descriptor = os.open(
                POLICY_FILE,
                os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK,
                dir_fd=directory,
            )
        finally:
            os.close(directory)
        try:
            with open(descriptor, "rb", closefd=False) as stream:
                return _read_policy(stream)
        finally:
            os.close(descriptor)
    except FileNotFoundError:
        return None


def _redirects(info):
    """Whether an lstat result is a link, including a Windows junction or other reparse point."""
    attributes = getattr(info, "st_file_attributes", 0)
    return stat.S_ISLNK(info.st_mode) or bool(
        attributes & getattr(stat, "FILE_ATTRIBUTE_REPARSE_POINT", 0)
    )


def _policy_bytes_checked(root):
    """Read the policy where descriptor-relative opens are unavailable.

    Every component is inspected without following it and refused if it redirects or is not
    a directory, with the errors the descriptor path raises for the same layouts. The file is
    then opened and must be the file that was inspected. Unlike the descriptor path this
    detects a component replaced between inspection and open rather than preventing it; it
    is the confinement this platform's API allows.
    """
    path = Path(root)
    try:
        for index, part in enumerate(("",) + POLICY_PARTS):
            path = path / part if part else path
            info = os.lstat(path)
            if _redirects(info):
                raise OSError(f"checkpoint policy path component is a link: {index}")
            if not stat.S_ISDIR(info.st_mode):
                raise NotADirectoryError(
                    f"checkpoint policy path component is not a directory: {index}"
                )
        path = path / POLICY_FILE
        inspected = os.lstat(path)
        if _redirects(inspected):
            raise OSError("checkpoint configuration is a link")
        descriptor = os.open(path, os.O_RDONLY | getattr(os, "O_BINARY", 0))
    except FileNotFoundError:
        return None
    try:
        return _read_inspected(descriptor, inspected)
    finally:
        os.close(descriptor)


def _read_inspected(descriptor, inspected):
    """Read the policy from descriptor, which must be the file that was inspected.

    The caller keeps the descriptor and closes it: a wrap that fails (the policy path is a
    directory) leaves a descriptor it was meant to own open.
    """
    with open(descriptor, "rb", closefd=False) as stream:
        if not os.path.samestat(inspected, os.fstat(stream.fileno())):
            raise CheckpointError(
                "checkpoint configuration was replaced while it was read"
            )
        return _read_policy(stream)


def _read_policy(stream):
    info = os.fstat(stream.fileno())
    if not stat.S_ISREG(info.st_mode) or info.st_size > MAX_OUTPUT:
        raise CheckpointError(
            "checkpoint configuration must be a regular file <= 1 MiB"
        )
    raw = stream.read(MAX_OUTPUT + 1)
    if len(raw) > MAX_OUTPUT:
        raise CheckpointError("checkpoint configuration exceeds 1 MiB")
    return raw


def _config(root):
    try:
        raw = _policy_bytes(root)
        if raw is None:
            return {"enabled": False}
        pairs = []
        value = json.loads(
            raw, object_pairs_hook=lambda items: (pairs.append(items), dict(items))[1]
        )
    except (OSError, ValueError, RecursionError) as error:
        raise CheckpointError(
            "invalid or inaccessible checkpoint configuration"
        ) from error
    return _validate_config(value, pairs)


CONFIG_FIELDS = frozenset(
    {
        "version",
        "enabled",
        "commit_after_minutes",
        "commit_after_files",
        "on_stop",
        "publish",
        "remote",
        "base",
        "repository",
        "branch_prefixes",
        "require_pr",
    }
)
OPTIONAL_CONFIG_FIELDS = frozenset(
    {"enforce_batch_scope", "require_checks", "required_checks"}
)


def _validate_config(value, pairs):
    if any(len(dict(items)) != len(items) for items in pairs):
        raise CheckpointError("checkpoint configuration contains duplicate keys")
    if (
        not isinstance(value, dict)
        or type(value.get("version")) is not int
        or value.get("version") != 1
    ):
        raise CheckpointError("checkpoint configuration requires version 1")
    if not CONFIG_FIELDS.issubset(value) or not set(value).issubset(
        CONFIG_FIELDS | OPTIONAL_CONFIG_FIELDS
    ):
        raise CheckpointError("checkpoint configuration has unknown or missing fields")
    if any(
        type(value[name]) is not bool
        for name in ("enabled", "on_stop", "publish", "require_pr")
    ):
        raise CheckpointError("checkpoint boolean fields must be booleans")
    _validate_review_policy(value)
    _validate_thresholds_and_refs(value)
    _validate_repository(value["repository"], value["branch_prefixes"])
    return value


def _validate_thresholds_and_refs(value):
    if (
        type(value["commit_after_minutes"]) is not int
        or not 1 <= value["commit_after_minutes"] <= 1440
        or type(value["commit_after_files"]) is not int
        or not 1 <= value["commit_after_files"] <= 1000
    ):
        raise CheckpointError("checkpoint thresholds are outside their bounds")
    if any(
        not isinstance(value[name], str)
        or not REF_NAME.fullmatch(value[name])
        or ".." in value[name]
        or "//" in value[name]
        for name in ("remote", "base")
    ):
        raise CheckpointError("checkpoint remote and base must be nonempty strings")


def _validate_repository(repo, prefixes):
    if (
        not isinstance(repo, str)
        or repo.count("/") != 1
        or any(
            item in {".", ".."} or not REPO_PART.fullmatch(item)
            for item in repo.split("/")
        )
        or not isinstance(prefixes, list)
        or not 1 <= len(prefixes) <= 32
        or any(
            not isinstance(item, str)
            or not REF_NAME.fullmatch(item)
            or not item.endswith("/")
            or ".." in item
            or "//" in item
            for item in prefixes
        )
    ):
        raise CheckpointError("checkpoint repository or branch_prefixes is malformed")


def _validate_review_policy(value):
    for key in ("enforce_batch_scope", "require_checks"):
        if key in value and type(value[key]) is not bool:
            raise CheckpointError("checkpoint boolean fields must be booleans")
    names = value.get("required_checks", [])
    if (
        not isinstance(names, list)
        or len(names) > 64
        or any(
            not isinstance(name, str)
            or not 1 <= len(name) <= 200
            or name.strip() != name
            or any(ord(char) < 32 for char in name)
            for name in names
        )
    ):
        raise CheckpointError(
            "required_checks must contain at most 64 bounded check names"
        )
    if len(set(names)) != len(names):
        raise CheckpointError("required_checks contains duplicate names")
    if names and not value.get("require_checks", False):
        raise CheckpointError("required_checks requires require_checks")
    if value.get("require_checks", False) and not (
        value["publish"] and value["require_pr"]
    ):
        raise CheckpointError("require_checks requires publication and PR observation")
    if value.get("require_checks", False) and not names:
        raise CheckpointError(
            "require_checks requires a nonempty required_checks selection"
        )


def _paths(raw):
    return {os.fsdecode(item) for item in raw.split(b"\0") if item}


def _public(path):
    return path not in (
        ".standards-receipt.json",
        ".workingdir",
    ) and not path.startswith(".workingdir/")


WORKTREE_DIFF = ("--no-ext-diff", "--no-textconv", "--ignore-submodules=all")
NAME_ONLY = ("--name-only", "-z", "--no-renames")


def _diff_names(root, *options, timeout=None):
    return _paths(_git(root, "diff", *WORKTREE_DIFF, *options, timeout=timeout))


def _worktree(root, timeout=None):
    tracked = _diff_names(root, *NAME_ONLY, "--", timeout=timeout)
    untracked = _paths(
        _git(
            root,
            "ls-files",
            "--others",
            "--exclude-standard",
            "-z",
            "--",
            timeout=timeout,
        )
    )
    staged = _diff_names(root, "--cached", *NAME_ONLY, "--", timeout=timeout)
    private_writes = _diff_names(
        root, "--cached", *NAME_ONLY, "--diff-filter=ACMRTUXB", "--", timeout=timeout
    )
    conflicts = _paths(_git(root, "ls-files", "-u", "-z", timeout=timeout))
    public = {name for name in tracked | staged | untracked if _public(name)}
    if len(public) > MAX_PATHS:
        raise CheckpointError(f"changed public path count exceeds {MAX_PATHS}")
    if any(
        name == ".workingdir" or name.startswith(".workingdir/")
        for name in private_writes
    ):
        raise CheckpointError("private .workingdir content is staged")
    if conflicts:
        raise CheckpointError("unresolved merge conflict in worktree")
    return public


def _listed_worktrees(root):
    """Read a bounded NUL-delimited worktree listing; record fields cannot contain NUL."""
    fields = _git(
        root, "worktree", "list", "--porcelain", "-z", timeout=WORKTREE_COMMAND_TIMEOUT
    ).split(b"\0")
    entries = []
    entry = {}
    for raw in fields:
        field = os.fsdecode(raw)
        if not field:
            if entry:
                entries.append(entry)
                entry = {}
            if len(entries) >= MAX_WORKTREES:
                break
            continue
        key, separator, value = field.partition(" ")
        if key == "worktree" and separator:
            entry["path"] = value
        elif key in {"bare", "prunable", "detached"}:
            entry[key] = True
        elif key in {"HEAD", "branch"} and separator:
            entry[key] = value
    if entry and len(entries) < MAX_WORKTREES:
        entries.append(entry)
    return entries


def _gitdir(path, common_dir):
    """Resolve a worktree's admin directory from its .git directory or pointer file."""
    marker = path / ".git"
    try:
        if marker.is_dir():
            return marker.resolve()
        raw = marker.read_text(encoding="utf-8", errors="strict")
    except (OSError, UnicodeError):
        return common_dir if path.resolve() == Path(common_dir).parent else None
    if len(raw) > 4096 or not raw.startswith("gitdir: "):
        return None
    target = Path(raw[8:].strip())
    return (target if target.is_absolute() else marker.parent / target).resolve()


def _worktree_recent(path, common_dir, now):
    """Use admin-directory and selected state-file mtimes as bounded activity evidence."""
    directory = _gitdir(path, common_dir)
    if directory is None:
        return False
    candidates = (
        directory,
        directory / "HEAD",
        directory / "index",
        directory / "logs" / "HEAD",
    )
    newest = 0
    for candidate in candidates:
        try:
            newest = max(newest, candidate.stat().st_mtime)
        except OSError:
            continue
    age = now - newest
    return newest > 0 and 0 <= age <= WORKTREE_ACTIVITY_WINDOW


# Characters that continue a name, a dot aside: a branch may hold "/", a directory name may not,
# and a whole path is bounded by white space. A dot continues a name only when a name character
# follows it, so a name that ends a sentence still counts. OPEN.md holds a name when nothing on
# either side continues it: "branch=feat/x" and ".../worktrees/x." name the worktree, while
# "feat/x2", "old-x" and "x.old" do not.
BRANCH_NAME_CHARS = "A-Za-z0-9_/-"
DIRECTORY_NAME_CHARS = "A-Za-z0-9_-"


def _name_held(open_text, name, chars):
    """Whether open_text holds name with no name character, or dot and name character, around it."""
    if not name:
        return False
    pattern = f"(?<![.{chars}])" + re.escape(name) + f"(?![{chars}]|\\.[{chars}])"
    return re.search(pattern, open_text) is not None


def _path_held(open_text, path):
    """Whether open_text holds path bounded by white space, a sentence-ending dot allowed."""
    pattern = r"(?<!\S)" + re.escape(path) + r"(?![^\s.]|\.\S)"
    return bool(path) and re.search(pattern, open_text) is not None


def _named_worktree(open_text, entry):
    """Match branch, directory name, or whole path in bounded OPEN.md text."""
    return (
        _name_held(
            open_text,
            entry.get("branch", "").removeprefix("refs/heads/"),
            BRANCH_NAME_CHARS,
        )
        or _name_held(open_text, Path(entry["path"]).name, DIRECTORY_NAME_CHARS)
        or _path_held(open_text, entry["path"])
    )


def _open_worktree_text(root):
    """Read only the bounded OPEN ledger; a missing ledger names no worktrees."""
    try:
        with (Path(root) / ".workingdir" / "OPEN.md").open("rb") as stream:
            raw = stream.read(OPEN_ITEMS_LIMIT + 1)
    except FileNotFoundError:
        return ""
    if len(raw) > OPEN_ITEMS_LIMIT:
        raise CheckpointError(".workingdir/OPEN.md exceeds 1 MiB")
    return raw.decode("utf-8", errors="replace")


def _remote_ref_exists(path, ref):
    """Whether a remote-tracking ref exists locally, with a bounded Git call."""
    raw = _run(
        ["git", "rev-parse", "--verify", "--quiet", ref],
        path,
        allowed=(0, 1),
        timeout=WORKTREE_COMMAND_TIMEOUT,
    )
    return bool(raw.strip())


def _rev_list_count(path, *refs):
    """Count HEAD commits outside refs, with a bounded Git call."""
    raw = _git(
        path,
        "rev-list",
        "--count",
        "HEAD",
        "--not",
        *refs,
        timeout=WORKTREE_COMMAND_TIMEOUT,
    )
    return int(raw.decode().strip())


def _worktree_base_ahead(path, cfg):
    """Return whether this branch is pushed to its configured remote and ahead of base."""
    branch = _branch(path, timeout=WORKTREE_COMMAND_TIMEOUT)
    if not branch:
        return False
    remote = cfg["remote"]
    branch_ref = f"refs/remotes/{remote}/{branch}"
    base_ref = f"refs/remotes/{remote}/{cfg['base']}"
    if not _remote_ref_exists(path, branch_ref) or not _remote_ref_exists(
        path, base_ref
    ):
        return False
    return (
        _rev_list_count(path, branch_ref) == 0 and _rev_list_count(path, base_ref) > 0
    )


def _pull_request_argv(cfg, branch=None):
    """Build the bounded repository listing or one explicit-head PR probe."""
    argv = ["gh", "pr", "list", "--repo", cfg["repository"], "--state", "all"]
    if branch:
        argv.extend(("--head", branch))
    return [*argv, "--limit", "200", "--json", "headRefName,state"]


def _pull_request_rows(raw, kind):
    """Decode a bounded PR listing and reject unknown row shapes or states."""
    rows = json.loads(raw)
    if not isinstance(rows, list) or len(rows) > 200:
        raise CheckpointError(f"pull request {kind} is malformed")
    if any(
        not isinstance(row, dict)
        or not isinstance(row.get("headRefName"), str)
        or row.get("state") not in {"OPEN", "CLOSED", "MERGED"}
        for row in rows
    ):
        raise CheckpointError(f"pull request {kind} is malformed")
    return rows


def _list_pull_requests(root, cfg, deadline):
    """Return validated states from repository PR listing with bounded retry."""
    last_error = ""
    backoffs = (1, 2)
    for attempt in range(3):
        if time.monotonic() >= deadline:
            break
        try:
            raw = _run(
                _pull_request_argv(cfg),
                root,
                network=True,
                allowed=(0,),
                timeout=PULL_REQUEST_TIMEOUT,
            )
            rows = _pull_request_rows(raw, "listing")
            states = {}
            for row in rows:
                states.setdefault(row["headRefName"], set()).add(row["state"])
            return states, ""
        except (CheckpointError, ValueError, TypeError) as error:
            last_error = str(error)
            if attempt < len(backoffs):
                delay = backoffs[attempt]
                if time.monotonic() + delay > deadline:
                    break
                time.sleep(delay)
    return None, last_error or "worktree check budget exhausted"


def _probe_one_pull_request(root, cfg, branch):
    """Return the validated states for one explicit branch-head query."""
    raw = _run(
        _pull_request_argv(cfg, branch),
        root,
        network=True,
        allowed=(0,),
        timeout=PULL_REQUEST_TIMEOUT,
    )
    rows = _pull_request_rows(raw, "probe")
    return {row["state"] for row in rows if row["headRefName"] == branch}


def _probe_pull_requests(root, cfg, branches, states, deadline):
    """Probe at most three absent branches; unavailable state stays explicitly unverified."""
    probes = 0
    for branch in branches:
        if branch in states or probes >= 3 or time.monotonic() >= deadline:
            continue
        probes += 1
        try:
            states[branch] = _probe_one_pull_request(root, cfg, branch)
        except (CheckpointError, ValueError, TypeError, KeyError):
            states[branch] = {"UNVERIFIED"}
    return states


def _pull_request_states(root, cfg, branches, deadline):
    """Read all PR heads once, then probe at most three absent branches."""
    if not cfg.get("enabled") or not cfg.get("repository"):
        return {}, "checkpoint config disabled or repository missing"
    if not branches:
        return {}, ""
    if time.monotonic() >= deadline:
        return None, "worktree check budget exhausted"
    states, error = _list_pull_requests(root, cfg, deadline)
    if error:
        return None, error
    return _probe_pull_requests(root, cfg, branches, states, deadline), ""


GATE_SCRATCH_RE = re.compile(r"gate-([1-9][0-9]{0,9})-[0-9]{1,20}")
MAX_GATE_CMDLINE = 16 * 1024


def _gate_scratch_pid(resolved, worktree_paths):
    """Recognize a gate path beneath any listed checkout of this repository."""
    for candidate in (resolved, *resolved.parents)[:256]:
        match = GATE_SCRATCH_RE.fullmatch(candidate.name)
        if match is None:
            continue
        parent = candidate.parent
        if (
            parent.name == "worktrees"
            and parent.parent.name == ".standards"
            and parent.parent.parent in worktree_paths
        ):
            return int(match[1])
    return None


def _live_gate_pid(pid):
    """Require bounded /proc command identity; a reused or unreadable PID fails closed."""
    try:
        with Path(f"/proc/{pid}/cmdline").open("rb") as stream:
            raw = stream.read(MAX_GATE_CMDLINE + 1)
    except OSError:
        return False
    if not raw or len(raw) > MAX_GATE_CMDLINE:
        return False
    argv = raw.rstrip(b"\0").split(b"\0")
    return (
        len(argv) >= 2
        and Path(os.fsdecode(argv[0])).name == "praetorctl"
        and argv[1] == b"gate"
    )


def _is_gate_scratch(resolved, worktree_paths):
    """Only a scratch checkout whose owning praetorctl gate is live can be skipped."""
    pid = _gate_scratch_pid(resolved, worktree_paths)
    return pid is not None and _live_gate_pid(pid)


def _existing_worktree(entry):
    """Resolve only present checkout directories; broken paths remain unselected."""
    try:
        path = Path(entry.get("path", "")).resolve(strict=True)
        return path if path.is_dir() else None
    except (OSError, RuntimeError):
        return None


def _unfinished_candidate(entry, path, *, common_dir, open_text, now, owned, paths):
    """Keep dead gate owners independently of session ownership and recent activity."""
    if _is_gate_scratch(path, paths):
        return False
    pid = _gate_scratch_pid(path, paths)
    if pid is not None:
        entry["gate_reason"] = f"gate owner pid {pid} is not a live praetorctl gate run"
        return True
    if not _worktree_recent(path, common_dir, now) or (
        owned is not None and str(path) not in owned
    ):
        return False
    return not _named_worktree(open_text, entry)


def _recent_unnamed_worktrees(root, common_dir, open_text, now, owned=None):
    """Select session worktrees, retaining abandoned gate scratch regardless of ownership."""
    root_path = Path(
        _git(root, "rev-parse", "--show-toplevel", timeout=WORKTREE_COMMAND_TIMEOUT)
        .decode()
        .strip()
    ).resolve()
    entries = [(entry, _existing_worktree(entry)) for entry in _listed_worktrees(root)]
    paths = {path for _entry, path in entries if path is not None}
    candidates = []
    for entry, resolved in entries:
        if resolved is None or resolved == root_path:
            continue
        if entry.get("bare") or entry.get("prunable"):
            continue
        entry["path"] = str(resolved)
        entry["branch"] = entry.get("branch", "").removeprefix("refs/heads/")
        if _unfinished_candidate(
            entry,
            resolved,
            common_dir=common_dir,
            open_text=open_text,
            now=now,
            owned=owned,
            paths=paths,
        ):
            candidates.append((entry, resolved))
    return candidates


def _session_owned_worktrees(common_dir):
    """Return set of resolved worktree paths owned by session, or None if unknown."""
    session_hash = os.environ.get(SESSION_ENV, "").strip().lower()
    if not SESSION_HASH_RE.fullmatch(session_hash):
        return None
    record_file = (
        common_dir / "praetor-checkpoint" / "sessions" / f"{session_hash}.json"
    )
    try:
        with record_file.open("rb") as stream:
            raw = stream.read(MAX_RECORD_BYTES + 1)
        if len(raw) > MAX_RECORD_BYTES:
            return None
        data = json.loads(raw)
        if not isinstance(data, list):
            return None
        return {str(Path(p).resolve()) for p in data if isinstance(p, str)}
    except (OSError, json.JSONDecodeError, UnicodeDecodeError):
        return None


def _public_change_reason(path):
    """Return public dirty-path count or the existing worktree reader's failure."""
    try:
        public = _worktree(path, timeout=WORKTREE_COMMAND_TIMEOUT)
        return [f"{len(public)} public change(s)"] if public else []
    except (CheckpointError, OSError) as error:
        return [str(error)]


def _content_in_base(path, cfg):
    """Whether the worktree's net change since its merge base is already in the remote base.

    A branch that landed through a squash or a train has neither a remote copy of its commits
    nor a pull request of its own head, yet nothing of it is unfinished. Its reversed patch then
    applies to the base. Read-only: the index lives in a temporary folder outside the repository.
    """
    remote, base = cfg.get("remote") or "origin", cfg.get("base") or "main"
    base_ref = f"refs/remotes/{remote}/{base}"
    if not _remote_ref_exists(path, base_ref):
        return False
    fork = _git(path, "merge-base", base_ref, "HEAD", timeout=LANDED_CHECK_TIMEOUT)
    diff = ["diff", "--binary", *WORKTREE_DIFF, fork.decode().strip(), "HEAD"]
    patch = _git(path, *diff, timeout=LANDED_CHECK_TIMEOUT)
    return not patch.strip() or _reverse_applies(path, base_ref, patch)


def _reverse_applies(path, base_ref, patch):
    """Whether patch applies in reverse to base_ref, on an index in a temporary folder."""
    with tempfile.TemporaryDirectory(prefix="praetor-landed-") as scratch:
        index = Path(scratch) / "index"
        change = Path(scratch) / "change.patch"
        change.write_bytes(patch)
        _git(path, "read-tree", base_ref, index=index, timeout=LANDED_CHECK_TIMEOUT)
        check = ["apply", "--check", "--reverse", "--cached", str(change)]
        try:
            _git(path, *check, index=index, timeout=LANDED_CHECK_TIMEOUT)
        except CheckpointError:
            return False
    return True


def _landed(path, cfg):
    """_content_in_base, failing closed: a check that cannot run leaves the worktree judged."""
    try:
        return _content_in_base(path, cfg)
    except (CheckpointError, ValueError, OSError):
        return False


def _worktree_unpublished_count(path):
    """Count HEAD commits unreachable from every remote, using the canonical criterion."""
    return _rev_list_count(path, "--remotes")


def _worktree_pr_needed(path, cfg, reasons):
    """Check unpublished commits and base ahead; append reasons and return needs_pr."""
    try:
        count = _worktree_unpublished_count(path)
    except (CheckpointError, ValueError, OSError) as error:
        reasons.append(f"commit state not verified ({error})")
        return False
    if count:
        reasons.append(f"{count} commit(s) not on any remote")
        return False
    if not cfg.get("enabled") or not cfg.get("repository"):
        return False
    try:
        return _worktree_base_ahead(path, cfg)
    except (CheckpointError, ValueError, OSError) as error:
        reasons.append(f"pull request state not verified ({error})")
        return False


def _judge_worktree(root, cfg, entry, path, deadline):
    """Return one worktree's unfinished reasons and whether its PR needs checking."""
    if time.monotonic() >= deadline:
        return (
            entry,
            ["pull request state not verified (worktree check budget exhausted)"],
            False,
        )
    reasons = _public_change_reason(path)
    if entry.get("gate_reason"):
        reasons.append(entry["gate_reason"])
    if not reasons and _landed(path, cfg):
        return entry, [], False
    needs_pr = _worktree_pr_needed(path, cfg, reasons)
    return entry, reasons, needs_pr


def _ordered_map(fn, items):
    """fn over items on at most WORKTREE_CHECK_WORKERS threads, results in item order.

    Plain threads, because the minimal Python of the devcontainer (python3-minimal) has
    neither concurrent.futures nor queue. The first exception a call raised is raised
    again after every thread has finished, as ThreadPoolExecutor.map does.
    """
    results, errors, lock, cursor = [None] * len(items), [], threading.Lock(), [0]

    def work():
        for _ in range(len(items)):
            with lock:
                index, cursor[0] = cursor[0], cursor[0] + 1
            if index >= len(items):
                return
            try:
                results[index] = fn(items[index])
            except Exception as error:  # noqa: BLE001  re-raised below, after the join
                errors.append(error)
                return

    threads = [
        threading.Thread(target=work)
        for _ in range(min(WORKTREE_CHECK_WORKERS, len(items)))
    ]
    for thread in threads:
        thread.start()
    for thread in threads:
        thread.join()
    if errors:
        raise errors[0]
    return results


def _judge_worktree_candidates(root, cfg, candidates, deadline):
    """Separate unfinished records from branches requiring bounded PR state reads."""
    jobs = [(cfg, entry, path, deadline) for entry, path in candidates]
    results = _ordered_map(lambda job: _judge_worktree(root, *job), jobs)
    records = [(entry, reasons) for entry, reasons, _ in results]
    pr_candidates = []
    for entry, _reasons, needs_pr in results:
        if needs_pr:
            pr_candidates.append((entry, _reasons))
    return records, pr_candidates


def _git_common_dir(root):
    """Resolve repository common Git directory."""
    raw = (
        _git(
            root,
            "rev-parse",
            "--path-format=absolute",
            "--git-common-dir",
            timeout=WORKTREE_COMMAND_TIMEOUT,
        )
        .decode()
        .strip()
    )
    return Path(raw).resolve()


def _unfinished_worktree_records(root, cfg):
    """Observe recent worktrees; return records, skip note, unverified count."""
    deadline = time.monotonic() + WORKTREE_CHECK_BUDGET
    common_dir = _git_common_dir(root)
    owned = _session_owned_worktrees(common_dir)
    candidates = _recent_unnamed_worktrees(
        root, common_dir, _open_worktree_text(root), time.time(), owned=owned
    )
    records, pr_candidates = _judge_worktree_candidates(root, cfg, candidates, deadline)
    states, skipped = _pull_request_states(
        root, cfg, [item[0]["branch"] for item in pr_candidates], deadline
    )
    unverified_pushed = 0
    if states is None:
        unverified_pushed = len(pr_candidates)
    else:
        for entry, reasons in pr_candidates:
            state = states.get(entry["branch"])
            if state is None or "UNVERIFIED" in state:
                reasons.append("pull request state not verified")
            elif not state.intersection({"OPEN", "MERGED"}):
                reasons.append(
                    "pushed branch ahead of base has no open or merged pull request"
                )
    unfinished = [(entry, reasons) for entry, reasons in records if reasons]
    if owned is None:
        for _entry, reasons in unfinished:
            reasons.append(OWNERSHIP_UNKNOWN_REASON)
    return unfinished, skipped, unverified_pushed


def _add_unfinished_worktrees(root, cfg, result):
    """Append bounded stop actions for every recent unnamed worktree with unfinished work."""
    if result.get("deferred"):
        return
    records, skipped, unverified_pushed = _unfinished_worktree_records(root, cfg)
    result["unfinished_worktrees"] = [
        {"path": item[0]["path"], "branch": item[0]["branch"], "reasons": item[1]}
        for item in records[:MAX_UNFINISHED_WORKTREES]
    ]
    result["unfinished_worktrees_truncated"] = max(
        0, len(records) - MAX_UNFINISHED_WORKTREES
    )
    if skipped:
        result["pull_request_check_skipped"] = skipped
    for entry, reasons in records[:MAX_WORKTREE_ACTIONS]:
        reason_text = ", ".join(reasons)
        result["actions"].append(
            f"Unfinished worktree {entry['path']} ({entry['branch'] or 'detached'}): "
            f"{reason_text}; finish it (commit, push, PR) or name it in .workingdir/OPEN.md "
            "with id, branch, last step, next step"
        )
    more = max(0, len(records) - MAX_WORKTREE_ACTIONS)
    if more:
        result["actions"].append(
            f"Unfinished worktrees: and {more} more; name them in .workingdir/OPEN.md"
        )
    if unverified_pushed:
        result["actions"].append(
            f"Pull request lookup failed ({skipped}): {unverified_pushed} pushed "
            "branch(es) unverified; retry or name in .workingdir/OPEN.md"
        )


def _branch(root, timeout=None):
    return (
        _run(
            ["git", "symbolic-ref", "--quiet", "--short", "HEAD"],
            root,
            allowed=(0, 1),
            timeout=timeout,
        )
        .decode()
        .strip()
    )


def _remote_url(root, remote):
    return _git(root, "remote", "get-url", remote).decode().strip()


def _ssh_transport_disabled(url):
    if not (url.startswith("git@") or urlsplit(url).scheme == "ssh"):
        return False
    # A plain split, not shlex: the adopted bundle runs under the owner overlay's Python, which has
    # no shlex module, and only the first word (a program named false) matters here.
    words = os.environ.get("GIT_SSH_COMMAND", "").split()
    return bool(words) and Path(words[0]).name == "false"


def _push_transport_disabled(root, cfg):
    try:
        url = _remote_url(root, cfg["remote"])
    except CheckpointError:
        return False
    if (
        url.startswith("git@")
        or url.startswith("ssh://")
        or urlsplit(url).scheme == "ssh"
    ):
        return _ssh_transport_disabled(url)
    if urlsplit(url).scheme != "https" or os.environ.get("GIT_TERMINAL_PROMPT") != "0":
        return False
    helper = (
        _git(root, "config", "--get", "credential.helper", allowed=(0, 1))
        .decode()
        .strip()
    )
    return not helper


def _forge_token_absent():
    if os.environ.get("GH_TOKEN") or os.environ.get("GITHUB_TOKEN"):
        return False
    config = os.environ.get("GH_CONFIG_DIR")
    return not config or not (Path(config) / "hosts.yml").is_file()


def _detached_job_deferral(root, cfg):
    label = os.environ.get("PRAETOR_DETACHED_JOB")
    if (
        not label
        or not _push_transport_disabled(root, cfg)
        or not _forge_token_absent()
    ):
        return None
    return (
        f"detached job {label}: no push transport, no forge token; "
        "the launching session commits"
    )


def _validate_remote(url, repository):
    owner, name = repository.split("/", 1)
    path = f"/{owner}/{name}"
    if url.startswith("git@github.com:"):
        actual = url.removeprefix("git@github.com:").removesuffix(".git")
        if actual != f"{owner}/{name}":
            raise CheckpointError("configured remote does not match repository")
        return
    parsed = urlsplit(url)
    if (
        parsed.scheme not in {"https", "ssh"}
        or parsed.hostname != "github.com"
        or (parsed.scheme == "https" and parsed.username is not None)
        or parsed.password is not None
        or parsed.port is not None
        or parsed.query
        or parsed.fragment
        or parsed.path not in {path, path + ".git"}
        or (parsed.scheme == "ssh" and parsed.username != "git")
    ):
        raise CheckpointError("configured remote does not match repository")


def _remote_refs(root, cfg, branch):
    branch_ref = f"refs/heads/{branch}"
    base_ref = f"refs/heads/{cfg['base']}"
    lines = (
        _git(
            root,
            "ls-remote",
            "--heads",
            cfg["remote"],
            branch_ref,
            base_ref,
            network=True,
        )
        .decode()
        .splitlines()
    )
    refs = {}
    for line in lines:
        fields = line.split()
        if (
            len(fields) != 2
            or not OID.fullmatch(fields[0])
            or fields[1] not in {branch_ref, base_ref}
        ):
            raise CheckpointError("remote branch/base response is malformed")
        if fields[1] in refs:
            raise CheckpointError("remote branch/base response contains duplicate refs")
        refs[fields[1]] = fields[0]
    return refs, branch_ref, base_ref


def _branch_publication(root, head, remote_oid):
    if remote_oid == head:
        return "pushed", []
    if not remote_oid:
        return "due_push", [PUSH_ACTION]
    try:
        _git(root, "cat-file", "-e", f"{remote_oid}^{{commit}}")
    except CheckpointError as error:
        raise CheckpointError(
            "live branch object is unavailable locally; fetch the configured remote"
        ) from error
    if _git(root, "rev-parse", "--is-shallow-repository").decode().strip() != "false":
        raise CheckpointError(
            "branch ancestry is unverified in shallow history; fetch missing history and review"
        )
    counts = (
        _git(root, "rev-list", "--left-right", "--count", f"{head}...{remote_oid}")
        .decode()
        .split()
    )
    if len(counts) != 2 or any(not value.isdecimal() for value in counts):
        raise CheckpointError("branch ancestry count is malformed")
    local_only, remote_only = map(int, counts)
    if remote_only == 0:
        return "due_push", [PUSH_ACTION]
    if local_only == 0:
        return "remote_ahead", [
            "Review and reconcile the live remote branch ahead of local HEAD before publication."
        ]
    return "diverged", [
        "Review and reconcile divergent local and live remote commits before publication."
    ]


def _publication(root, cfg, branch, head, observation=None):
    _validate_remote(_remote_url(root, cfg["remote"]), cfg["repository"])
    refs, branch_ref, base_ref = _remote_refs(root, cfg, branch)
    base_oid = refs.get(base_ref)
    if not base_oid:
        raise CheckpointError("configured remote base branch is missing")
    if observation is not None:
        observation.update(remote_head=refs.get(branch_ref, ""), base_head=base_oid)
    try:
        _git(root, "cat-file", "-e", f"{base_oid}^{{commit}}")
    except CheckpointError as error:
        raise CheckpointError(
            "live base object is unavailable locally; fetch the configured remote"
        ) from error
    status, actions = _branch_publication(root, head, refs.get(branch_ref))
    if status in {"remote_ahead", "diverged"}:
        return status, actions
    ahead = int(
        _git(root, "rev-list", "--count", f"{base_oid}..{head}").decode().strip()
    )
    if ahead <= 0:
        return "not_ahead", []
    if status == "due_push":
        return status, actions
    if not cfg["require_pr"]:
        return "pushed", []
    return _pull_request(root, cfg, branch, head, observation)


def _open_pull_requests(root, cfg, branch):
    fields = "number,url,isDraft,headRefOid,headRefName,baseRefName"
    if cfg.get("require_checks", False):
        fields += ",statusCheckRollup"
    command = ["gh", "pr", "list", "--repo", cfg["repository"], "--base", cfg["base"]]
    command += ["--head", branch, "--state", "open", "--limit", "2", "--json", fields]
    raw = _run(command, root, network=True)
    try:
        return json.loads(raw)
    except json.JSONDecodeError as error:
        raise CheckpointError(
            f"gh returned malformed pull request data: {error}"
        ) from error


def _pull_request(root, cfg, branch, head, observation=None):
    repo = cfg["repository"]
    prs = _open_pull_requests(root, cfg, branch)
    if not isinstance(prs, list) or len(prs) > 1:
        raise CheckpointError("pull request result is ambiguous")
    if not prs:
        return "due_draft_pr", [PR_ACTION]
    pr = prs[0]
    if (
        not isinstance(pr, dict)
        or type(pr.get("number")) is not int
        or pr["number"] <= 0
        or type(pr.get("isDraft")) is not bool
        or pr.get("headRefOid") != head
        or pr.get("headRefName") != branch
        or pr.get("baseRefName") != cfg["base"]
        or pr.get("url") != f"https://github.com/{repo}/pull/{pr['number']}"
    ):
        raise CheckpointError("pull request does not match the exact branch head/base")
    actions = _review_checks(
        root, pr, cfg, observation if observation is not None else {}
    )
    return "present", actions


def _check_run_state(check):
    status, conclusion = check.get("status"), check.get("conclusion")
    if not isinstance(status, str) or (
        conclusion is not None and not isinstance(conclusion, str)
    ):
        raise CheckpointError("hosted check status and conclusion must be strings")
    if status in {"QUEUED", "IN_PROGRESS", "WAITING", "PENDING", "REQUESTED"}:
        if conclusion not in {None, ""}:
            raise CheckpointError("unfinished hosted check has a terminal conclusion")
        return "pending"
    states = {
        "SUCCESS": "passed",
        "FAILURE": "failed",
        "CANCELLED": "failed",
        "TIMED_OUT": "failed",
        "ACTION_REQUIRED": "failed",
        "STARTUP_FAILURE": "failed",
        "STALE": "failed",
        "SKIPPED": "skipped",
        "NEUTRAL": "skipped",
    }
    if status != "COMPLETED" or conclusion not in states:
        raise CheckpointError("hosted check has an unknown status or conclusion")
    return states[conclusion]


def _check_observation(check):
    if not isinstance(check, dict):
        raise CheckpointError("hosted check must be an object")
    if check.get("__typename") == "CheckRun":
        name, state = check.get("name"), _check_run_state(check)
    elif check.get("__typename") == "StatusContext":
        name = check.get("context")
        if not isinstance(check.get("state"), str):
            raise CheckpointError("hosted status context state must be a string")
        state = {
            "SUCCESS": "passed",
            "FAILURE": "failed",
            "ERROR": "failed",
            "PENDING": "pending",
            "EXPECTED": "pending",
        }.get(check.get("state"))
    else:
        raise CheckpointError("hosted check has an unknown type")
    if (
        not isinstance(name, str)
        or not 1 <= len(name) <= 200
        or any(ord(char) < 32 for char in name)
        or state is None
    ):
        raise CheckpointError("hosted check has an invalid name or state")
    return name, state


def _rerun_key(check):
    """Return (workflow, job, start) for a check run a later run of the same job can supersede.

    GitHub judges a job by its latest run on a head, so a re-run in the same workflow replaces
    an earlier run of the same job; two workflows that share a job name never hide each other.
    A status context, or a run without a workflow name or a valid start time, has no key and is
    always judged.
    """
    workflow, started = check.get("workflowName"), check.get("startedAt")
    if (
        check.get("__typename") != "CheckRun"
        or not isinstance(workflow, str)
        or not workflow
        or not isinstance(started, str)
        or not START_TIME.fullmatch(started)
        or started.startswith("0001-")
    ):
        return None
    return workflow, check.get("name"), started


def _latest_runs(checks):
    """Drop each check run that a later run of the same job in the same workflow supersedes."""
    latest = {}
    for check in checks[:MAX_CHECKS]:
        key = _rerun_key(check) if isinstance(check, dict) else None
        if key is not None:
            latest[key[:2]] = max(latest.get(key[:2], key[2]), key[2])
    kept = []
    for check in checks[:MAX_CHECKS]:
        key = _rerun_key(check) if isinstance(check, dict) else None
        if key is None or key[2] == latest[key[:2]]:
            kept.append(check)
    return kept


def _failed_check_runs(checks):
    """Return the names of the latest check runs that concluded FAILURE."""
    if not isinstance(checks, list):
        return set()
    return {
        check.get("name")
        for check in _latest_runs(checks)
        if isinstance(check, dict)
        and check.get("__typename") == "CheckRun"
        and check.get("status") == "COMPLETED"
        and check.get("conclusion") == "FAILURE"
        and isinstance(check.get("name"), str)
    }


def _nodes(value, key, bound):
    """Return the nodes list of the connection value[key], at most bound long."""
    connection = value.get(key) if isinstance(value, dict) else None
    nodes = connection.get("nodes") if isinstance(connection, dict) else None
    if not isinstance(nodes, list) or len(nodes) > bound:
        raise CheckpointError("gh returned malformed check annotation data")
    return nodes


def _draft_marked_names(data):
    """Return the check names whose every failed run carries exactly the draft annotation."""
    try:
        commit = data["data"]["repository"]["object"]
    except (KeyError, TypeError) as error:
        raise CheckpointError("gh returned malformed check annotation data") from error
    marked = {}
    for suite in _nodes(commit, "checkSuites", MAX_DRAFT_SUITES):
        for run in _nodes(suite, "checkRuns", MAX_DRAFT_RUNS):
            name = run.get("name") if isinstance(run, dict) else None
            if not isinstance(name, str):
                raise CheckpointError("gh returned malformed check annotation data")
            carries = _carries_draft_marker(
                _nodes(run, "annotations", MAX_DRAFT_ANNOTATIONS)
            )
            marked[name] = marked.get(name, True) and carries
    return {name for name, carries in marked.items() if carries}


def _carries_draft_marker(notes):
    """Report whether an annotation of notes has exactly the draft title and message."""
    return any(
        isinstance(note, dict)
        and note.get("title") == DRAFT_GATE_TITLE
        and note.get("message") == DRAFT_GATE_MESSAGE
        for note in notes
    )


def _draft_gate_checks(root, cfg, pr, checks):
    """Return the failed checks of a draft that are a hosted gate refusing to run on it.

    Only a draft pull request is read, only check runs that concluded FAILURE, and only a name
    whose every failed run on the head carries exactly the draft annotation (DRAFT_GATE_TITLE,
    DRAFT_GATE_MESSAGE). A ready pull request, any other conclusion, another annotation, or one
    unmarked run of the name keeps the failure.
    """
    failed = _failed_check_runs(checks) if pr.get("isDraft") is True else set()
    if not failed:
        return frozenset()
    owner, name = cfg["repository"].split("/", 1)
    variables = {"owner": owner, "name": name, "oid": pr["headRefOid"]}
    command = ["gh", "api", "graphql", "-f", "query=" + DRAFT_GATE_QUERY]
    for key, value in variables.items():
        command += ["-f", f"{key}={value}"]
    raw = _run(command, root, network=True)
    try:
        data = json.loads(raw)
    except json.JSONDecodeError as error:
        raise CheckpointError(
            f"gh returned malformed check annotation data: {error}"
        ) from error
    return frozenset(_draft_marked_names(data) & failed)


def _review_state(checks, required, draft_gates=frozenset()):
    if not isinstance(checks, list) or len(checks) >= MAX_CHECKS:
        raise CheckpointError(
            "hosted check coverage is invalid or reaches its observation bound"
        )
    states = ("passed", "failed", "pending", "draft_pending", "skipped")
    counts = {name: 0 for name in states}
    passed = set()
    for check in _latest_runs(checks):
        name, state = _check_observation(check)
        if (
            state == "failed"
            and name in draft_gates
            and check.get("__typename") == "CheckRun"
            and check.get("conclusion") == "FAILURE"
        ):
            state = "draft_pending"
        counts[state] += 1
        if state == "passed":
            passed.add(name)
    missing = sorted(set(required) - passed)
    for status in ("failed", "pending", "draft_pending"):
        if counts[status]:
            return status, counts, missing
    if missing or not counts["passed"]:
        return "missing", counts, missing
    return "passed", counts, missing


def _review_checks(root, pr, cfg, result):
    if not cfg.get("require_checks", False):
        return []
    checks = pr.get("statusCheckRollup")
    status, counts, missing = _review_state(
        checks,
        cfg.get("required_checks", []),
        _draft_gate_checks(root, cfg, pr, checks),
    )
    result.update(
        review_status=status,
        check_counts=counts,
        required_checks_unpassed=missing,
        review_url=pr["url"],
        review_head=pr["headRefOid"],
    )
    actions = {
        "failed": (
            "Inspect and repair failed hosted checks; "
            "record blockers before claiming completion."
        ),
        "pending": (
            "Recheck pending hosted checks for this exact head before claiming completion."
        ),
        "draft_pending": (
            "Hosted gates do not run on a draft; mark it ready for review when the work is "
            "complete and recheck its hosted checks for this exact head before claiming completion."
        ),
        "missing": (
            "Verify missing required hosted checks; "
            "absent or skipped checks do not prove acceptance."
        ),
    }
    return [actions[status]] if status in actions else []


def _empty_result(include_paths):
    result = {
        "schema_version": 1,
        "enabled": False,
        "deferred": False,
        "due": False,
        "actions": [],
        "branch": "",
        "head": "",
        "changed_count": 0,
        "commit_due": False,
        "enforce_batch_scope": False,
        "publication_status": "disabled",
        "review_status": "not_requested",
    }
    if include_paths:
        result["public_paths"] = []
    return result


def _observe_worktree(root, cfg, event, include_paths, result, *, require_branch=True):
    result["enabled"] = True
    result["review_status"] = (
        "not_observed" if cfg.get("require_checks", False) else "not_requested"
    )
    result["enforce_batch_scope"] = cfg.get("enforce_batch_scope", False)
    result["branch"] = _branch(root)
    head = _verified_head(root)
    result["head"] = head
    public = _worktree(root)
    result["changed_count"] = len(public)
    if include_paths:
        result["public_paths"] = sorted(public)
    result["commit_due"] = _commit_due(root, cfg, event, head, public)
    result["due"] = result["commit_due"]
    result["publication_status"] = "not_due"
    if result["due"] and require_branch:
        _require_checkpoint_branch(result["branch"], cfg)
    if result["due"]:
        result["actions"] = list(PUBLIC_ACTIONS)
    return head, public


def _verified_head(root):
    head = (
        _git(root, "rev-parse", "--verify", "--quiet", "HEAD", allowed=(0, 1))
        .decode()
        .strip()
    )
    if head and not OID.fullmatch(head):
        raise CheckpointError("HEAD is not a valid Git object ID")
    return head


def _commit_due(root, cfg, event, head, public):
    age = (
        time.time()
        - int(_git(root, "show", "-s", "--format=%ct", "HEAD").decode().strip())
        if head
        else cfg["commit_after_minutes"] * 60
    )
    due = bool(public) and (
        age >= cfg["commit_after_minutes"] * 60
        or len(public) >= cfg["commit_after_files"]
    )
    return due or (event == "stop" and cfg["on_stop"] and bool(public))


def _require_checkpoint_branch(branch, cfg):
    if not branch:
        raise CheckpointError("checkpoint due on detached HEAD")
    allowed = any(branch.startswith(prefix) for prefix in cfg["branch_prefixes"])
    if branch in {"main", "master", cfg["base"]} or not allowed:
        raise CheckpointError("checkpoint due on protected or nonallowed branch")


def _observe_publication(root, cfg, event, result):
    if event != "stop" or not cfg["publish"] or not result["head"]:
        return
    branch = result["branch"]
    allowed = any(branch.startswith(prefix) for prefix in cfg["branch_prefixes"])
    if not branch:
        raise CheckpointError("publication cannot be verified on detached HEAD")
    if branch in {"main", "master", cfg["base"]} or not allowed:
        if result["due"]:
            raise CheckpointError("checkpoint due on protected or nonallowed branch")
        if branch == cfg["base"]:
            _observe_base(root, cfg, result)
        return
    status, actions = _publication(root, cfg, branch, result["head"], result)
    result["publication_status"] = status
    result["actions"].extend(actions)


def _observe_base(root, cfg, result):
    _validate_remote(_remote_url(root, cfg["remote"]), cfg["repository"])
    refs, _, base_ref = _remote_refs(root, cfg, cfg["base"])
    if base_ref not in refs:
        raise CheckpointError("configured remote base branch is missing")
    status, _ = _branch_publication(root, result["head"], refs[base_ref])
    states = {
        "pushed": "base_current",
        "due_push": "base_local_ahead",
        "remote_ahead": "base_remote_ahead",
        "diverged": "base_diverged",
    }
    result.update(
        publication_status=states[status],
        review_status="not_applicable",
        remote_head=refs[base_ref],
        base_head=refs[base_ref],
    )
    if status != "pushed":
        result["actions"].append(
            "Review and reconcile the protected base with its live remote; "
            "use a review branch for local work."
        )


def _validate_stability(root, head, public, result):
    if (
        _branch(root) != result["branch"]
        or _git(root, "rev-parse", "--verify", "--quiet", "HEAD", allowed=(0, 1))
        .decode()
        .strip()
        != head
    ):
        raise CheckpointError(
            "branch or HEAD changed during checkpoint observation; retry"
        )
    if not result["due"] and _worktree(root) != public:
        raise CheckpointError(
            "working tree changed during checkpoint observation; retry"
        )


def _inspect_enabled_checkpoint(root, cfg, event, include_paths, result):
    """Observe enabled checkout state, publication, and stop-only linked worktrees."""
    deferral = _detached_job_deferral(root, cfg)
    head, public = _observe_worktree(
        root, cfg, event, include_paths, result, require_branch=deferral is None
    )
    if deferral is not None:
        result.update(
            deferred=True,
            deferral=deferral,
            due=False,
            actions=[],
            publication_status="deferred_detached_job",
        )
        _validate_stability(root, head, public, result)
        return result
    _observe_publication(root, cfg, event, result)
    if event == "stop":
        _add_unfinished_worktrees(root, cfg, result)
    result["due"] = bool(result["actions"])
    _validate_stability(root, head, public, result)
    return result


def inspect_checkpoint(root: Path, event: str, include_paths: bool = False) -> dict:
    result = _empty_result(include_paths)
    if event not in {"tool", "stop"}:
        result.update(error="event must be tool or stop", publication_status="error")
        return result
    try:
        cfg = _config(root)
        if not cfg.get("enabled", False):
            if event == "stop":
                _add_unfinished_worktrees(root, cfg, result)
                result["due"] = bool(result["actions"])
            return result
        return _inspect_enabled_checkpoint(root, cfg, event, include_paths, result)
    except (CheckpointError, ValueError, OSError) as error:
        result["publication_status"] = "error"
        result["error"] = str(error)
        return result


def _prepare_native_session(args):
    """Bridge Go to the same bounded session implementation used by the Python adapter."""
    try:
        raw = sys.stdin.buffer.read(MAX_OUTPUT + 1)
        if len(raw) > MAX_OUTPUT:
            raise ValueError("oversized checkpoint session input")
        payload = json.loads(raw)
        if not isinstance(payload, dict):
            raise TypeError("checkpoint session input must be an object")
        session_hash = prepare_checkpoint_session(
            payload, args.root.resolve(), HookBudget(5), args.event == "stop"
        )
        print(
            "PRAETOR_CHECKPOINT_SESSION=" + session_hash
        )  # caveman:not-applicable structured-protocol
        return 0
    except (HookError, OSError, TypeError, ValueError) as error:
        print("checkpoint session: " + str(error)[:1000], file=sys.stderr)
        return 1


def main(argv=None):
    parser = argparse.ArgumentParser()
    parser.add_argument("--event", choices=("tool", "stop"), required=True)
    parser.add_argument("--root", type=Path, default=Path("."))
    parser.add_argument("--json", action="store_true")
    parser.add_argument("--marker", action="store_true")
    parser.add_argument("--include-paths", action="store_true")
    parser.add_argument("--prepare-session", action="store_true")
    args = parser.parse_args(argv)
    if args.prepare_session:
        return _prepare_native_session(args)
    result = inspect_checkpoint(args.root.resolve(), args.event, args.include_paths)
    line = json.dumps(result, sort_keys=True)
    print(
        "PRAETOR_CHECKPOINT_RESULT=" + line if args.marker else line
    )  # caveman:not-applicable structured-protocol
    return 0 if "error" not in result else 1


if __name__ == "__main__":
    sys.exit(main())
