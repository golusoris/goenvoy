"""Bounded process execution and isolated Git snapshots for local checks."""

import contextlib
import hashlib
import json
import math
import os
import re
import selectors
import shutil
import signal
import stat
import subprocess
import sys
import tempfile
import threading
import time
from pathlib import Path


class HookError(Exception):
    """An actionable local gate failure."""


KILL_TREE_TIMEOUT = 10
# The praetor CLI runs each git or go command in a process group of its own. Asked to stop by a
# catchable signal, it forwards that signal to each command's group, waits up to its
# util.CommandWaitDelay (5 s) for them, and kills the rest
# (internal/util/command_interrupt_unix.go). STOP_GRACE outlasts that before the whole
# group is killed.
STOP_GRACE = 10
STOP_POLL = 0.05
SNAPSHOT_GIT_CONFIG = ("-c", "core.autocrlf=false")
# A commit snapshot clones without checkout or hardlinks; the source keeps a remote name of
# its own so the adopter's origin can be restored under "origin".
SNAPSHOT_CLONE = (
    "git",
    "clone",
    "--quiet",
    "--no-hardlinks",
    "--no-checkout",
    "--origin",
    "praetor-snapshot",
)
# The operating system already bounds a process environment, but the hook applies a lower
# deterministic ceiling before scanning it. Git propagates command-line `-c` values to hooks
# through GIT_CONFIG_COUNT plus indexed key/value variables or through GIT_CONFIG_PARAMETERS.
# Those settings select the outer push transport; they are not policy for nested fixture
# repositories and must not cross the isolated gate boundary.
MAX_PROCESS_ENV_ENTRIES = 4096
# One `git rev-parse` per checkout identity: two absolute paths, answered from local metadata.
# Both calls run inside a native hook's budget (see HookBudget), so each is capped low and
# stopped with a short grace: nothing they start holds a lock that needs time to release.
SESSION_GIT_TIMEOUT = 2
SESSION_GIT_GRACE = 0.5
SESSION_GIT_OUTPUT = 16 * 1024
# Git exits 128 from die(): no repository it can open from that directory.
GIT_FATAL = 128
TRANSIENT_GIT_CONFIG_KEYS = ("GIT_CONFIG_COUNT", "GIT_CONFIG_PARAMETERS")
TRANSIENT_GIT_CONFIG_PREFIXES = ("GIT_CONFIG_KEY_", "GIT_CONFIG_VALUE_")
MANAGED_PROCESS_ENV = {
    "CI": "true",
    "GOFLAGS": "-mod=readonly",
    "GOWORK": "off",
    "PYTHONDONTWRITEBYTECODE": "1",
}


def _kill_bounded(process, sig=signal.SIGTERM, grace=None):
    """Stop a bounded child and everything it started.

    ``run`` and ``run_bounded`` start children with ``start_new_session=True``, so on POSIX the
    whole process group is stopped -- a bounded command that forks must not leave orphans behind.
    ``stop_process_group`` asks it with ``sig`` first and kills it only after ``STOP_GRACE``.
    That call is POSIX-only: on Windows ``os.killpg`` does not exist, and the timeout path
    raised ``AttributeError`` instead of reporting that a process had exceeded its bound. The
    failure therefore appeared only when a gate was already failing, which is the worst time to
    lose the reason.

    Windows has no process group to signal, and ``Popen.kill`` terminates the direct child
    alone, so a grandchild outlived its bound and the timeout test observed it writing its
    marker after the parent had been killed (#135). ``taskkill /F /T`` walks descendants by
    parent process id, which is the platform's own answer to the same question, and the direct
    kill stays as the floor for the case where it cannot run.

    Both callers share this one function, so neither grows its own copy. ``grace`` bounds the
    wait for the group, or for the tree killer, in place of ``STOP_GRACE`` and
    ``KILL_TREE_TIMEOUT``.
    """
    try:
        if hasattr(os, "killpg"):
            stop_process_group(process, sig, grace)
        else:
            _kill_process_tree(process, KILL_TREE_TIMEOUT if grace is None else grace)
    except (ProcessLookupError, PermissionError):
        pass  # The process or group already exited.


def stop_process_group(process, sig=signal.SIGTERM, grace=None):
    """Ask the process group ``process`` leads to stop with ``sig``; kill it after ``grace``.

    SIGKILL on the group of a praetor CLI ends the CLI, which cannot catch it, and misses the
    commands it runs in process groups of their own: they run on, with git's index lock held. A
    catchable signal lets the CLI forward it, so git removes its locks and the CLI waits for its
    commands before it exits. The group is killed once ``process`` and every other member have
    not exited within ``grace``, or at once when the wait itself is interrupted (a second
    Ctrl-C). ``grace`` defaults to ``STOP_GRACE``. POSIX-only.
    """
    if not _signal_group(process.pid, sig):
        return
    try:
        stopped = _await_group_exit(process, STOP_GRACE if grace is None else grace)
    except BaseException:
        _signal_group(process.pid, signal.SIGKILL)
        raise
    if not stopped:
        _signal_group(process.pid, signal.SIGKILL)


def _signal_group(pgid, sig):
    """Send ``sig`` to group ``pgid``; False once no process of ours is left in it."""
    if (
        not isinstance(pgid, int)
        or isinstance(pgid, bool)
        or pgid <= 1
        or pgid == os.getpgrp()
    ):
        return False
    try:
        os.killpg(pgid, sig)
    except (ProcessLookupError, PermissionError):
        return False
    return True


def _await_group_exit(process, grace):
    """Wait up to ``grace`` seconds for ``process`` and the rest of its group to exit.

    The group outlives ``process`` when ``process`` is ``go run`` and the CLI it started is
    still stopping its commands. ``process`` is reaped here; other members are reaped by the
    process they are reparented to.
    """
    deadline = time.monotonic() + grace
    for _ in range(math.ceil(grace / STOP_POLL) + 1):
        if process.poll() is not None and not _signal_group(process.pid, 0):
            return True
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            return False
        time.sleep(min(STOP_POLL, remaining))
    return False


def _kill_process_tree(process, timeout=KILL_TREE_TIMEOUT):
    """Kill a child and its descendants where no process group exists to signal."""
    try:
        subprocess.run(
            ["taskkill", "/F", "/T", "/PID", str(process.pid)],
            stdin=subprocess.DEVNULL,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            timeout=timeout,
            check=False,
        )
    except (OSError, subprocess.SubprocessError):
        pass  # No tree killer on this host; the direct kill below is what remains.
    process.kill()


def _stop_bounded(process, grace=None):
    """Stop a bounded child and reap it; with ``grace``, each of the two waits takes at most that.

    A stop therefore adds at most ``2 * grace`` seconds to the timeout that triggered it, which
    is what a native hook's budget reserves for it.
    """
    _kill_bounded(process, grace=grace)
    process.wait(timeout=5 if grace is None else grace)


class _SharedBound:
    """Output two reader threads collect under one byte limit."""

    def __init__(self, maximum):
        self.maximum = maximum
        self.output = {"stdout": bytearray(), "stderr": bytearray()}
        self.total = 0
        self.exceeded = False
        self.finished = 0
        self.changed = threading.Condition()

    def add(self, key, chunk):
        """Record a chunk; False once the limit is passed and reading must stop."""
        with self.changed:
            self.total += len(chunk)
            if self.total > self.maximum:
                self.exceeded = True
                self.changed.notify_all()
                return False
            self.output[key].extend(chunk)
            return True

    def finish(self):
        with self.changed:
            self.finished += 1
            self.changed.notify_all()


def _drain(stream, key, bound):
    # Every read consumes at least one byte or observes EOF, so maximum + 2 reads is enough.
    try:
        for _ in range(bound.maximum + 2):
            chunk = os.read(stream.fileno(), min(65536, bound.maximum + 1))
            if not chunk or not bound.add(key, chunk):
                return
    except OSError:
        return  # The pipe closed because the process was stopped.
    finally:
        bound.finish()


def _bounded_output_threaded(process, timeout, maximum):
    """Collect bounded output where select() cannot wait on pipes.

    On Windows ``selectors`` accepts only sockets, so registering a child's pipes raised
    ``OSError`` and every bounded command failed before it produced a byte -- including each git
    call the checkpoint planner makes. Each pipe is drained by its own thread instead, under the
    same shared byte limit and deadline; a reader never blocks the process on a full pipe, and a
    limit or deadline breach is raised for the caller to stop the process, as on POSIX.
    """
    deadline = time.monotonic() + timeout
    bound = _SharedBound(maximum)
    for stream, key in ((process.stdout, "stdout"), (process.stderr, "stderr")):
        threading.Thread(target=_drain, args=(stream, key, bound), daemon=True).start()
    with bound.changed:
        done = bound.changed.wait_for(
            lambda: bound.exceeded or bound.finished == 2,
            timeout=max(0.0, deadline - time.monotonic()),
        )
        if bound.exceeded:
            raise HookError("checkpoint command output exceeded its byte limit")
        if not done:
            raise HookError("checkpoint command timed out")
    process.wait(timeout=max(0.001, deadline - time.monotonic()))
    return bytes(bound.output["stdout"])


def _bounded_output(process, timeout, maximum):
    if os.name == "nt":
        return _bounded_output_threaded(process, timeout, maximum)
    deadline = time.monotonic() + timeout
    output = {"stdout": bytearray(), "stderr": bytearray()}
    total = 0
    with selectors.DefaultSelector() as selector:
        selector.register(process.stdout, selectors.EVENT_READ, "stdout")
        selector.register(process.stderr, selectors.EVENT_READ, "stderr")
        # Every iteration either consumes a byte, observes EOF, or waits to deadline.
        for _ in range(maximum + 3):
            if not selector.get_map():
                process.wait(timeout=max(0.001, deadline - time.monotonic()))
                return bytes(output["stdout"])
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                raise HookError("checkpoint command timed out")
            events = selector.select(remaining)
            if not events:
                raise HookError("checkpoint command timed out")
            for key, _mask in events:
                chunk = os.read(key.fileobj.fileno(), min(65536, maximum - total + 1))
                if not chunk:
                    selector.unregister(key.fileobj)
                    continue
                total += len(chunk)
                if total > maximum:
                    raise HookError("checkpoint command output exceeded its byte limit")
                output[key.data].extend(chunk)
    raise HookError("checkpoint command exceeded its read bound")


def _spawn_bounded(args, cwd, env):
    """Start args in a session of its own, stdin closed and both streams piped."""
    return subprocess.Popen(
        args,
        cwd=cwd,
        env=env,
        start_new_session=True,
        stdin=subprocess.DEVNULL,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )


def run_bounded(
    args,
    cwd=None,
    *,
    timeout=10,
    max_output=1024 * 1024,
    env=None,
    allowed=(0,),
    grace=None,
):
    """Bound both streams during capture; never copy credential-bearing diagnostics.

    ``grace`` bounds the stop of a child that overran (see ``_stop_bounded``); by default the
    child gets ``STOP_GRACE`` to release what it holds.
    """
    if not 0 < timeout <= 60 or not 0 < max_output <= 1024 * 1024:
        raise HookError("invalid checkpoint process bounds")
    try:
        with _spawn_bounded(args, cwd, env) as process:
            try:
                stdout = _bounded_output(process, timeout, max_output)
            except BaseException:
                _stop_bounded(process, grace)
                raise
            if process.returncode not in allowed:
                raise HookError(
                    f"{args[0]} exited {process.returncode}; checkpoint unverified"
                )
            return stdout
    except (OSError, subprocess.TimeoutExpired) as error:
        raise HookError(
            f"{args[0]} checkpoint process failed ({type(error).__name__})"
        ) from error


def run(
    args, cwd=None, *, data=None, timeout=180, capture=True, env=None, allowed=(0,)
):
    """Execute argv without a shell; preserve failures and bound every process."""
    settings = dict(os.environ if env is None else env)
    settings["PYTHONDONTWRITEBYTECODE"] = MANAGED_PROCESS_ENV["PYTHONDONTWRITEBYTECODE"]
    try:
        with subprocess.Popen(
            args,
            cwd=cwd,
            env=settings,
            start_new_session=True,
            stdin=subprocess.PIPE if data is not None else None,
            stdout=subprocess.PIPE if capture else None,
            stderr=subprocess.PIPE if capture else None,
        ) as process:
            try:
                stdout, stderr = process.communicate(input=data, timeout=timeout)
            except (subprocess.TimeoutExpired, KeyboardInterrupt) as error:
                # Ctrl-C reached only this process: the child leads a session of its own. It
                # gets the signal the terminal would have sent it.
                interrupted = isinstance(error, KeyboardInterrupt)
                _kill_bounded(process, signal.SIGINT if interrupted else signal.SIGTERM)
                process.communicate(timeout=5)
                raise
    except (OSError, subprocess.TimeoutExpired) as error:
        raise HookError(f"{args[0]}: {error}") from error
    if process.returncode not in allowed:
        output = (stdout or b"") + (stderr or b"")
        raise HookError(
            f"{' '.join(map(str, args))} exited {process.returncode}\n"
            + output.decode(errors="replace")
        )
    return stdout or b""


def git(*args, cwd=None, env=None):
    return run(["git", *args], cwd=cwd, env=env)


def resolved_relative_to(path, root):
    """Return ``path`` relative to ``root``, resolving both to their real filesystem form first.

    macOS presents ``/var`` and ``/tmp`` as symlinks to ``/private/var`` and ``/private/tmp``;
    Python's own ``tempfile`` module, ``go list``'s reported package directories, and a test
    fixture's own temp root each pick a side of that symlink independently. Comparing one
    resolved operand against one unresolved operand with ``Path.relative_to`` then raises
    ``ValueError`` for two paths that name the same directory. Every subpath/containment check
    in this package shares this one resolution instead of each call site deciding for itself
    whether to call ``Path.resolve()`` -- mirrors the Go-side collapse onto
    ``internal/util.ResolveExistingPath`` (#282, #135); a caller that needs a custom error
    message catches the ``ValueError`` this raises (identical to ``Path.relative_to``).
    """
    return Path(path).resolve().relative_to(Path(root).resolve())


def unresolved_relative_to(path, root):
    """Return absolute ``path`` relative to ``root``, preserving spellings below ``root``.

    ``resolved_relative_to`` follows all symlinks. Callers like checkpoint_scope.py need
    to see un-resolved paths within ``root`` to detect symlink traversals. But spelling
    matches fail above ``root`` off-Linux (macOS /var vs /private/var, Windows 8.3 names).
    Both broke when tests began sending absolute paths during due checkpoints (#461).

    We match ``root`` by filesystem identity (``os.path.samestat``) against the outermost
    ancestor of ``path`` that is that directory, returning the unresolved remainder.
    """
    candidate = Path(path)
    if not candidate.is_absolute():
        raise ValueError(f"{os.fspath(path)!r} is not an absolute path")
    if os.pardir in candidate.parts:
        raise ValueError(f"{os.fspath(path)!r} contains a parent traversal")
    try:
        target = os.stat(root)
    except OSError as err:
        raise ValueError(f"{os.fspath(root)!r} cannot be inspected") from err
    for ancestor in (*reversed(candidate.parents), candidate):
        try:
            found = os.path.samestat(os.stat(ancestor), target)
        except OSError:
            # A component that cannot be inspected is not ``root``: a file not written
            # yet, or one below a directory this user may not search. The walk goes on.
            continue
        if found:
            return candidate.relative_to(ancestor)
    raise ValueError(f"{os.fspath(path)!r} is not below {os.fspath(root)!r}")


class HookBudget:
    """One deadline every process a native hook starts takes its timeout from.

    A client cancels a command hook at its registered timeout and then lets the call through:
    Claude Code ("A timed-out command ... hook doesn't block the tool call", hooks reference,
    "Timeouts"), Gemini CLI (a timed-out hook resolves unsuccessful with no decision,
    ``HookRunner`` in packages/core/src/hooks/hookRunner.ts) and Codex (a timed-out run is an
    error, never ``should_block``, codex-rs/hooks/src/events/pre_tool_use.rs). Timeouts that
    each fit on their own can add up past that, so each step draws from one budget instead,
    and a step that finds it spent refuses rather than starting.
    """

    def __init__(self, seconds):
        self.expires = time.monotonic() + seconds

    def timeout(self, cap):
        """Seconds the next process may run: ``cap``, or what is left of the budget if less."""
        left = self.expires - time.monotonic()
        if left <= 0:
            raise HookError("native hook budget spent before the check could run")
        return min(cap, left)


def refuse_after(seconds, refuse):
    """Start a timer that ends this process with ``refuse()``'s exit code after ``seconds``.

    A budget bounds each process a hook starts, not every wait the hook makes: a child in
    uninterruptible sleep on a stalled file system outlives SIGKILL, and reaping it blocks, as
    can a ``stat`` of the session directory. The timer thread answers for the hook meanwhile
    and exits without unwinding the blocked thread, so a stalled host still fails closed.
    ``refuse`` returns None when the hook has already answered; the hook then exits itself.
    The caller cancels the timer once it has answered.
    """

    def expire():
        code = refuse()
        if code is not None:
            os._exit(code)

    timer = threading.Timer(seconds, expire)
    timer.daemon = True
    timer.start()
    return timer


def _checkout_identity(directory, budget):
    """Return the resolved toplevel and Git common directory of the checkout holding a path.

    None when Git exits fatally there: the directory is in no repository Git can open.
    """
    raw = run_bounded(
        [
            "git",
            "rev-parse",
            "--path-format=absolute",
            "--show-toplevel",
            "--git-common-dir",
        ],
        cwd=directory,
        timeout=budget.timeout(SESSION_GIT_TIMEOUT),
        max_output=SESSION_GIT_OUTPUT,
        env=clean_env(),
        allowed=(0, GIT_FATAL),
        grace=SESSION_GIT_GRACE,
    )
    lines = os.fsdecode(raw).splitlines()
    if not lines:
        return None
    if len(lines) != 2 or not all(lines):
        raise HookError("git reported no checkout identity")
    try:
        return Path(lines[0]).resolve(), Path(lines[1]).resolve()
    except (OSError, RuntimeError) as error:
        raise HookError(
            f"checkout identity does not resolve ({type(error).__name__})"
        ) from error


def _session_directory(root, cwd):
    """The payload cwd as a directory worth asking Git about, or None to keep ``root``."""
    if not isinstance(cwd, str) or not cwd or "\0" in cwd:
        return None
    directory = Path(cwd)
    try:
        if (
            not directory.is_absolute()
            or not directory.is_dir()
            or directory.resolve() == Path(root).resolve()
        ):
            return None
    except (OSError, RuntimeError, ValueError):
        return None
    return directory


def session_root(root, cwd, budget):
    """Return the checkout whose policy, ledger and batch judge a native call made from ``cwd``.

    Claude Code keeps ${CLAUDE_PROJECT_DIR} on the checkout a session started in after the
    session enters a linked worktree, while the payload's ``cwd`` follows it there (hooks
    reference, "Worktrees are different"). A linked worktree of the same repository shares
    ``root``'s Git common directory, and its own toplevel is the checkout the call works in.
    A ``cwd`` Git proves is anything else -- missing, not an absolute directory, outside any
    repository, in a submodule or another repository -- keeps ``root``, so the caller's own
    checks decide it. When Git cannot answer inside ``budget`` (a timeout, a spent budget,
    output that is no checkout identity) this raises ``HookError``, and the caller refuses the
    call instead of guessing which checkout judges it.
    """
    directory = _session_directory(root, cwd)
    if directory is None:
        return root
    own = _checkout_identity(root, budget)
    other = _checkout_identity(directory, budget)
    if own is None or other is None or other[1] != own[1] or other[0] == own[0]:
        return root
    return other[0]


def paths(raw):
    return [os.fsdecode(item) for item in raw.split(b"\0") if item]


def changed(base, head="HEAD"):
    return paths(git("diff", "--name-only", "-z", "--no-renames", base, head, "--"))


def index_env():
    """Environment for reading the index Git hands the hook, without the caller's `-c` config.

    Git names the index a commit will record through GIT_INDEX_FILE: `git commit -a` points it
    at index.lock and `git commit <path>` at a next-index-*.lock, both of which differ from
    .git/index while the hook runs. The pre-commit export must read that file, so only the
    command-line configuration is removed here; GIT_DIR, GIT_WORK_TREE and GIT_INDEX_FILE stay
    exactly as Git set them. clean_env() builds on this and also drops the repository selection.
    """
    env = dict(os.environ)
    if len(env) > MAX_PROCESS_ENV_ENTRIES:
        raise HookError(
            f"process environment exceeds {MAX_PROCESS_ENV_ENTRIES} entries"
        )
    # Scan a statically bounded snapshot. Removing the count and legacy aggregate disables the
    # Git config injection; removing every indexed value also keeps transport URLs and other
    # caller data out of descendant process environments.
    for key in tuple(env)[:MAX_PROCESS_ENV_ENTRIES]:
        if key in TRANSIENT_GIT_CONFIG_KEYS or key.startswith(
            TRANSIENT_GIT_CONFIG_PREFIXES
        ):
            env.pop(key, None)
    return env


def clean_env():
    env = index_env()
    for key in (
        "GIT_DIR",
        "GIT_WORK_TREE",
        "GIT_INDEX_FILE",
        "GIT_COMMON_DIR",
        "GIT_OBJECT_DIRECTORY",
        "GIT_ALTERNATE_OBJECT_DIRECTORIES",
    ):
        env.pop(key, None)
    env.update(MANAGED_PROCESS_ENV)
    if len(env) > MAX_PROCESS_ENV_ENTRIES:
        raise HookError(
            f"process environment exceeds {MAX_PROCESS_ENV_ENTRIES} entries"
        )
    return env


# A snapshot whose hook was killed (a timeout, a cancelled push) keeps its directory: a commit
# snapshot holds a full clone, 0.6-0.7 GB. Before a new snapshot, the directories of this prefix
# that this user owns and nobody touched for SNAPSHOT_STALE_S are removed; no hook runs that long.
SNAPSHOT_PREFIX = "praetor-hook-"
SNAPSHOT_STALE_S = 3 * 60 * 60
MAX_STALE_SNAPSHOTS = 256


def _writable_retry(function, path, _exc):
    """rmtree onexc: make the parent and the entry writable, then retry the removal once."""
    for target in (os.path.dirname(path), path):
        with contextlib.suppress(OSError):
            os.chmod(target, os.stat(target).st_mode | stat.S_IWUSR | stat.S_IXUSR)
    function(path)


def sweep_stale_snapshots(root=None, now=None):
    """Remove leaked snapshot directories in root (the temporary directory); the count removed."""
    root = Path(root or tempfile.gettempdir())
    now = time.time() if now is None else now
    removed = 0
    for entry in sorted(root.glob(SNAPSHOT_PREFIX + "*"))[:MAX_STALE_SNAPSHOTS]:
        with contextlib.suppress(OSError):
            info = entry.lstat()
            if (
                not stat.S_ISDIR(info.st_mode)
                or now - info.st_mtime < SNAPSHOT_STALE_S
                or (hasattr(os, "getuid") and info.st_uid != os.getuid())
            ):
                continue
            if sys.version_info >= (3, 12):
                shutil.rmtree(entry, onexc=_writable_retry)
            else:  # Python 3.10 and 3.11 (the hooks' floor) name the callback onerror
                shutil.rmtree(
                    entry, onerror=_writable_retry
                )  # pylint: disable=deprecated-argument
            removed += 1
    return removed


@contextlib.contextmanager
def snapshot(ref=None):
    """Export the exact index or commit; never stash, stage, or edit the source. Leaked snapshots of
    killed hooks are swept first (sweep_stale_snapshots)."""
    sweep_stale_snapshots()
    with tempfile.TemporaryDirectory(prefix=SNAPSHOT_PREFIX) as directory:
        dest = Path(directory)
        env = clean_env()
        if ref is None:
            _export_index(dest, env)
        else:
            _export_commit(dest, ref, env)
        yield dest


def _export_index(dest, env):
    """Write the index the commit will record into dest, as a repository of its own."""
    # Snapshot checks consume repository bytes, not the operator's checkout
    # preference. Without this pin, Windows' core.autocrlf=true rewrites LF
    # shell/YAML blobs to CRLF and the isolated gate rejects bytes absent from
    # the index it claims to inspect. The export reads the index the commit will
    # record (index_env), never the stale .git/index clean_env would select.
    git(
        *SNAPSHOT_GIT_CONFIG,
        "checkout-index",
        "--all",
        "--force",
        f"--prefix={dest}/",
        env=index_env(),
    )
    # Lefthook's validator requires a repository even though it only
    # validates configuration. This metadata belongs solely to the export.
    run(["git", "init", "--quiet", str(dest)], env=env)


def _export_commit(dest, ref, env):
    """Clone this repository into dest with origin and its refs, detached at ref."""
    source = git("rev-parse", "--show-toplevel", env=env).decode().strip()
    origin = (
        run(
            ["git", "config", "--get", "remote.origin.url"],
            env=env,
            allowed=(0, 1),
        )
        .decode()
        .strip()
    )
    refs = git(
        "for-each-ref",
        "--format=%(objectname) %(refname)",
        "refs/remotes/origin/",
        env=env,
    )
    run([*SNAPSHOT_CLONE, source, str(dest)], env=env)
    if origin:
        run(["git", "remote", "add", "origin", origin], cwd=dest, env=env)
    for line in refs.decode().splitlines():
        oid, name = line.split()
        run(["git", "update-ref", name, oid], cwd=dest, env=env)
    run(
        ["git", *SNAPSHOT_GIT_CONFIG, "checkout", "--quiet", "--detach", ref],
        cwd=dest,
        env=env,
    )


def present_files(directory, names):
    files = []
    for name in names:
        path = directory / name
        if path.is_symlink():
            raise HookError(f"{name}: changed symlinks require explicit review")
        if path.is_file():
            files.append(name)
    return files


# Shared native checkpoint session records: Python adapters and the Go hook caller.
SESSION_ENV = "PRAETOR_CHECKPOINT_SESSION"
SESSION_INPUT_LIMIT = 1024 * 1024
MAX_WORKTREES = 512
MAX_RECORD_BYTES = 512 * 1024


def extract_session_id(payload):
    """Read session ID from agent payload without guessing."""
    if not isinstance(payload, dict):
        return None
    session_id = payload.get("session_id")
    if isinstance(session_id, str) and session_id.strip():
        return session_id.strip()
    conv_id = payload.get("conversationId")
    if isinstance(conv_id, str) and conv_id.strip():
        return conv_id.strip()
    return None


def session_hash_for(session_id):
    """Return first 16 hex chars of SHA256 of session ID. Never store raw ID."""
    return hashlib.sha256(session_id.encode("utf-8")).hexdigest()[:16]


def _git_common_dir(root, budget):
    """Resolve repository common Git directory."""
    raw = run_bounded(
        ["git", "rev-parse", "--path-format=absolute", "--git-common-dir"],
        cwd=root,
        timeout=budget.timeout(SESSION_GIT_TIMEOUT),
        max_output=SESSION_GIT_OUTPUT,
        grace=SESSION_GIT_GRACE,
    )
    return Path(raw.decode().strip()).resolve()


def _listed_worktrees(root, budget):
    """Read bounded NUL-delimited worktree listing."""
    raw = run_bounded(
        ["git", "worktree", "list", "--porcelain", "-z"],
        cwd=root,
        timeout=budget.timeout(SESSION_GIT_TIMEOUT),
        max_output=SESSION_INPUT_LIMIT,
        grace=SESSION_GIT_GRACE,
    )
    entries, entry = [], {}
    for part in raw.split(b"\0"):
        field = os.fsdecode(part)
        if not field:
            if entry and "path" in entry:
                entries.append(entry)
                entry = {}
            if len(entries) >= MAX_WORKTREES:
                break
            continue
        key, sep, val = field.partition(" ")
        if key == "worktree" and sep:
            entry["path"] = val
    if entry and "path" in entry and len(entries) < MAX_WORKTREES:
        entries.append(entry)
    return entries


MAX_EXTRACT_NODES = 256


def _extract_strings(data):
    """Extract string values iteratively from tool input under a scalar bound."""
    result = []
    stack = [data]
    for _ in range(MAX_EXTRACT_NODES):
        if not stack:
            break
        item = stack.pop()
        if isinstance(item, str):
            result.append(item)
        elif isinstance(item, dict):
            stack.extend(item.values())
        elif isinstance(item, (list, tuple)):
            stack.extend(item)
    return result


def _extract_tool_info(payload):
    """Extract tool name and tool input from payload."""
    tool_name = payload.get("tool_name")
    tool_input = payload.get("tool_input")
    if not tool_name and isinstance(payload.get("toolCall"), dict):
        call = payload["toolCall"]
        tool_name = call.get("name")
        tool_input = call.get("args")
    return str(tool_name or ""), tool_input


def _path_matches(resolved_path_str, text):
    """Compare literal paths in tool text after resolving both filesystem spellings."""
    # Keep quoted paths with spaces intact; also find absolute operands embedded
    # in shell commands or scripts. Never execute or expand the command text.
    pattern = r"""(?<![\w./\\-])(?:"[^"]+"|'[^']+'|(?:[A-Za-z]:[/\\]|/|\\\\)[^\s'";&|<>()]+)"""
    candidates = [
        text,
        *(match.group().strip("\"'") for match in re.finditer(pattern, text)),
    ]
    for candidate in candidates:
        if not Path(candidate).is_absolute():
            continue
        try:
            resolved_relative_to(candidate, resolved_path_str)
            return True
        except (OSError, RuntimeError, ValueError):
            continue
    return False


def _is_worktree_touched(resolved, tool_name, tool_input, strings):
    """Whether worktree was touched by tool invocation."""
    if tool_name == "EnterWorktree" and isinstance(tool_input, dict):
        target = tool_input.get("name") or tool_input.get("path")
        if isinstance(target, str) and target.strip() in (resolved.name, str(resolved)):
            return True
    path_str = str(resolved)
    return any(_path_matches(path_str, s) for s in strings)


def _read_session_record(record_file):
    """Read existing worktree paths from session record file."""
    try:
        with record_file.open("rb") as stream:
            raw = stream.read(MAX_RECORD_BYTES + 1)
        if len(raw) > MAX_RECORD_BYTES:
            return set()
        data = json.loads(raw)
        if isinstance(data, list):
            return {str(Path(p).resolve()) for p in data if isinstance(p, str)}
    except (OSError, json.JSONDecodeError, UnicodeDecodeError):
        pass
    return set()


def _write_session_record(sessions_dir, record_file, paths):
    """Atomically write sorted bounded worktree paths to session record."""
    sessions_dir.mkdir(parents=True, exist_ok=True)
    sorted_paths = sorted(paths)[:MAX_WORKTREES]
    payload = json.dumps(sorted_paths).encode("utf-8")
    if len(payload) > MAX_RECORD_BYTES:
        raise ValueError("checkpoint session record exceeds byte bound")
    tmp = None
    try:
        with tempfile.NamedTemporaryFile(dir=sessions_dir, delete=False) as handle:
            tmp = Path(handle.name)
            handle.write(payload)
        os.replace(tmp, record_file)
        tmp = None
    finally:
        if tmp is not None:
            try:
                tmp.unlink(missing_ok=True)
            except OSError:
                pass


def ensure_session_record(root, session_hash, budget):
    """Ensure session record file exists before running stop checkpoint."""
    common_dir = _git_common_dir(root, budget)
    sessions_dir = common_dir / "praetor-checkpoint" / "sessions"
    record_file = sessions_dir / f"{session_hash}.json"
    if not record_file.is_file():
        _write_session_record(sessions_dir, record_file, set())


def record_tool_worktrees(payload, root, budget):
    """Record any worktree paths touched by tool event."""
    session_id = extract_session_id(payload)
    if not session_id:
        return
    session_hash = session_hash_for(session_id)
    tool_name, tool_input = _extract_tool_info(payload)
    if not tool_input:
        return
    strings = _extract_strings(tool_input)
    common_dir = _git_common_dir(root, budget)
    sessions_dir = common_dir / "praetor-checkpoint" / "sessions"
    record_file = sessions_dir / f"{session_hash}.json"

    worktrees = _listed_worktrees(root, budget)
    touched = set()
    for entry in worktrees:
        try:
            resolved = Path(entry["path"]).resolve(strict=True)
        except (OSError, RuntimeError):
            continue
        if resolved == root:
            continue
        if _is_worktree_touched(resolved, tool_name, tool_input, strings):
            touched.add(str(resolved))
    if not touched and record_file.is_file():
        return
    existing = _read_session_record(record_file) if record_file.is_file() else set()
    if touched - existing or not record_file.is_file():
        _write_session_record(sessions_dir, record_file, existing | touched)


def prepare_checkpoint_session(payload, root, budget, stopping):
    """Prepare one shared ownership record and return its hashed identity."""
    session_id = extract_session_id(payload)
    if not session_id:
        return ""
    session_hash = session_hash_for(session_id)
    if stopping:
        ensure_session_record(root, session_hash, budget)
    else:
        record_tool_worktrees(payload, root, budget)
    return session_hash
