# SPDX-License-Identifier: EUPL-1.2
"""MkDocs hook: render each ```figure fence as the figure markup, and serve the figure stylesheet.

Listed under `hooks:` in mkdocs.yml as tools/figures/mkdocs_hook.py
(docs/adr/0015-interactive-figures-from-vendored-interfig.md, section 4). It uses only the Python
standard library and the MkDocs API and imports no checker (docs/adr/0016-figures-for-adopters.md,
section 6): the figure checks run in Node, `node tools/figures/build.mjs sources|site|portable`.

Each fence becomes the markup tools/figures/core.mjs wrote into docs/assets/figures/<slug>.json as
`html`; the hook only fills its `{{base}}` slot and drops its `{{link}}` line (`render_block`). The
fence scanner is the one behaviour this file shares with tools/figures/checks.mjs, because
MkDocs runs hooks in-process: both replay tools/figures/fence-fixtures.json, so a fence nested
inside a longer fence is left alone here exactly as the `site` check expects. A fence naming a
figure without docs/assets/figures/<slug>.json is logged as a warning, which fails
`mkdocs build --strict`.

figures.css sits beside this file, outside docs_dir. In the engine repository the committed player
files under player/ (loader.js, player.js and their THIRD-PARTY-LICENSES.txt, written by
tools/figures/bundle.mjs) sit beside it too; in an adopted repository player.lock.json names the
engine release asset (figures-player.tar.gz) and its SHA-256
(docs/adr/0021-figure-player-as-release-asset.md). `on_files` adds the CSS at CSS_URI and the
player files under PLAYER_URI, and `on_config` links the stylesheet from every page and the loader
as a module script only when player files are available. A site needs no `extra_css` or
`extra_javascript` entry for them and no
bundling step: the loader reads each figure's props from its SVG
(docs/adr/0016-figures-for-adopters.md, section 3).
"""

from __future__ import annotations

import atexit
import hashlib
import io
import json
import logging
import re
import shutil
import tarfile
import tempfile
from pathlib import Path
from typing import NamedTuple
from urllib.error import URLError
from urllib.request import urlopen

log = logging.getLogger("mkdocs.plugins.praetor_figures")

CSS_FILE = Path(__file__).with_name("figures.css")
CSS_URI = "assets/stylesheets/figures.css"
PLAYER_DIR = Path(__file__).with_name("player")
LOCK_FILE = Path(__file__).with_name("player.lock.json")
PLAYER_URI = "assets/javascripts/figures"
LOADER_URI = f"{PLAYER_URI}/loader.js"
CANONICAL_REPOSITORY = "cordanaLLM/praetor"
CANONICAL_REPOSITORY_URL = f"https://github.com/{CANONICAL_REPOSITORY}"
PLAYER_ARCHIVE = "figures-player.tar.gz"
EXPECTED_PLAYER_FILES = ("loader.js", "player.js", "THIRD-PARTY-LICENSES.txt")
DOWNLOAD_TIMEOUT_SECONDS = 30
SVG_FALLBACK_WARNING = (
    "figures: engine pin %s has no release; "
    "figures render as SVG (no interactive player)"
)
# player/ holds three files (DIST_FILES in bundle.mjs); the bound only keeps the listing finite.
MAX_PLAYER_FILES = 16
# on_config resolves once per build; on_files publishes those exact verified paths.
_build_assets: tuple[object, list[tuple[str, Path]]] | None = None
REBUILD = "node tools/figures/build.mjs build"

# Bounded so a pathological page cannot make the hook unbounded (HISS-02).
MAX_FILE_BYTES = 8 * 1024 * 1024
MAX_LINES = 100_000

# A fence opener or closer: three or more backticks or tildes, then the info string.
FENCE_LINE = re.compile(r"^(\s*)(`{3,}|~{3,})\s*([^\s`{]*)(.*)$")
SLUG = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
# The slots of a figure's JSON `html` (SLOTS and markup in tools/figures/core.mjs). A site shows the
# interactive figure itself, so the line holding the link to it is dropped.
BASE_SLOT = "{{base}}"
LINK_SLOT = "{{link}}"


class CheckError(ValueError):
    """A figure or player asset the hook cannot render or verify."""


class Fence(NamedTuple):
    """One top-level fenced block: its line span (end exclusive), indent, info string and body."""

    start: int
    end: int
    indent: str
    info: str
    body: list[str]


def bounded_lines(text: str) -> list[str]:
    """The lines of `text`, refusing more than MAX_LINES.

    Lines end at "\n" only, as Python-Markdown splits them, with a trailing "\r" dropped.
    str.splitlines() also breaks at form feeds, U+2028 and other separators, which would
    shift every index expand() uses to splice a rendered figure into the page.
    """
    lines = [line.removesuffix("\r") for line in text.split("\n")]
    if lines and lines[-1] == "":
        lines.pop()  # a final newline ends the last line; it does not start another
    if len(lines) > MAX_LINES:
        raise CheckError(f"more than {MAX_LINES} lines")
    return lines


def fence_blocks(text: str) -> list[Fence]:
    """Every top-level fenced block in a Markdown page (`fenceBlocks` in checks.mjs).

    A fence nested inside a longer or different fence is literal text, not a block: a closer
    must use the opener's character, be at least as long, and carry no info string. An unclosed
    fence runs to the end of the page.
    """
    lines = bounded_lines(text)
    blocks: list[Fence] = []
    opener: tuple[int, str, str, str] | None = None
    for index, line in enumerate(lines):
        match = FENCE_LINE.match(line)
        if match is None:
            continue
        indent, marker, info, rest = match.groups()
        if opener is None:
            opener = (index, indent, marker, info.lower())
        elif (
            marker[0] == opener[2][0]
            and len(marker) >= len(opener[2])
            and not info
            and not rest.strip()
        ):
            first_body = opener[0] + 1
            body = lines[first_body:index]
            blocks.append(Fence(opener[0], index + 1, opener[1], opener[3], body))
            opener = None
    if opener is not None:
        first_body = opener[0] + 1
        body = lines[first_body:]
        blocks.append(Fence(opener[0], len(lines), opener[1], opener[3], body))
    return blocks


def figure_slug(body: list[str]) -> str:
    """The slug a ```figure fence names: its first non-blank line."""
    return next((line.strip() for line in body if line.strip()), "")


def site_base(page_url: str) -> str:
    """The relative path from a built page back to the site root, for MkDocs' page.url."""
    depth = page_url.count("/")
    return "/".join([".."] * depth) if depth else "."


def figure_meta(figures_dir: Path, slug: str) -> dict:
    """The build's JSON for `slug`, or CheckError when there is none or it cannot be read."""
    if not SLUG.match(slug):
        raise CheckError(f"figure slug {slug!r} is not lowercase kebab-case")
    path = figures_dir / f"{slug}.json"
    if not path.is_file():
        raise CheckError(
            f"figure {slug!r} has no {path.as_posix()}; "
            f"add docs/figures/{slug}.ts and run {REBUILD}"
        )
    try:
        if path.stat().st_size > MAX_FILE_BYTES:
            raise CheckError(f"{path} exceeds {MAX_FILE_BYTES} bytes")
        data = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, UnicodeDecodeError, json.JSONDecodeError) as error:
        raise CheckError(f"cannot read {path}: {error}") from error
    if not isinstance(data, dict):
        raise CheckError(f"{path} does not hold a JSON object")
    return data


def render_block(meta: dict, base: str) -> str:
    """The figure's markup: its JSON `html`, rendered by tools/figures/core.mjs, with `base` in.

    This fills a slot and renders nothing itself; `markup` in core.mjs documents the slots. `base`
    replaces `{{base}}` as given and the line that holds `{{link}}` is dropped, as `fillSlots` in
    checks.mjs does without a link. A base that looks like a slot stays literal.
    """
    template = meta.get("html")
    if not isinstance(template, str) or not template:
        raise CheckError(
            f"figure {meta.get('slug')!r}: its JSON records no html; rebuild with: {REBUILD}"
        )
    lines = [line for line in template.split("\n") if LINK_SLOT not in line]
    return base.join("\n".join(lines).split(BASE_SLOT))


def expand(markdown: str, base: str, figures_dir: Path) -> tuple[str, list[str]]:
    """`markdown` with every top-level ```figure fence replaced by its rendered block.

    A fence whose figure cannot be rendered is left as it is and reported, so a strict MkDocs
    build fails on the hook's warning.
    """
    lines = markdown.split("\n")
    errors: list[str] = []
    for block in reversed([b for b in fence_blocks(markdown) if b.info == "figure"]):
        try:
            meta = figure_meta(figures_dir, figure_slug(block.body))
            rendered = render_block(meta, base)
        except CheckError as error:
            errors.append(str(error))
            continue
        start, end = block.start, block.end
        lines[start:end] = [block.indent + line for line in rendered.split("\n")]
    return "\n".join(lines), list(reversed(errors))


def on_config(config):
    """Resolve this build's files and link the loader only when it will be published."""
    global _build_assets
    published = published_files()
    _build_assets = (config, published)
    if CSS_URI not in config["extra_css"]:
        config["extra_css"].append(CSS_URI)
    if LOADER_URI not in {uri for uri, _ in published}:
        config["extra_javascript"] = [
            script for script in config["extra_javascript"] if str(script) != LOADER_URI
        ]
        return config
    if all(str(script) != LOADER_URI for script in config["extra_javascript"]):
        # Imported here so the hook loads without MkDocs in tests.
        from mkdocs.config.config_options import ExtraScriptValue

        loader = ExtraScriptValue(LOADER_URI)
        loader.type = "module"
        config["extra_javascript"].append(loader)
    return config


def read_player_lock(lock_file: Path) -> dict:
    """Read and parse player.lock.json beside the hook."""
    try:
        if lock_file.stat().st_size > MAX_FILE_BYTES:
            raise CheckError(f"{lock_file} exceeds {MAX_FILE_BYTES} bytes")
        data = json.loads(lock_file.read_text(encoding="utf-8"))
    except (OSError, UnicodeDecodeError, json.JSONDecodeError) as error:
        raise CheckError(f"cannot read {lock_file}: {error}") from error
    if not isinstance(data, dict):
        raise CheckError(f"{lock_file} does not hold a JSON object")
    return data


def release_archive_url(lock: dict, release: str, asset: str) -> str:
    """Build the release download URL for the lock's repository."""
    repository = str(lock.get("repository") or "").strip()
    if not repository or repository == CANONICAL_REPOSITORY:
        base = CANONICAL_REPOSITORY_URL.rstrip("/")
    elif repository.startswith("https://"):
        base = repository.rstrip("/")
    elif "://" in repository:
        raise CheckError(
            f"the player archive is fetched over https only, not {repository!r}"
        )
    else:
        base = f"https://github.com/{repository.strip('/')}"
    return f"{base}/releases/download/{release}/{asset}"


def download_player_archive(url: str) -> bytes:
    """Fetch the release archive with a timeout and size cap."""
    if not url.startswith("https://"):
        raise CheckError(f"the player archive is fetched over https only, not {url!r}")
    # the https guard above is the only way in, so the URL scheme is fixed
    # nosemgrep: python.lang.security.audit.dynamic-urllib-use-detected.dynamic-urllib-use-detected
    with urlopen(url, timeout=DOWNLOAD_TIMEOUT_SECONDS) as response:  # noqa: S310
        raw = response.read(MAX_FILE_BYTES + 1)
    if len(raw) > MAX_FILE_BYTES:
        raise CheckError(f"{url} exceeds {MAX_FILE_BYTES} bytes")
    return raw


def verify_archive_sha256(raw: bytes, expected: str, asset: str) -> None:
    """Verify the archive's SHA-256 digest matches the lock's sha256 value."""
    actual = hashlib.sha256(raw).hexdigest()
    if not expected or actual != expected.lower():
        raise CheckError(
            f"{asset} SHA-256 mismatch: lock expects {expected!r}, got {actual!r}"
        )


def validate_archive_member(member: tarfile.TarInfo, seen: set[str]) -> str:
    """Verify one tar member is an expected top-level regular file with no path traversal."""
    name = member.name
    if name.startswith("/") or ".." in name.split("/") or "/" in name or "\\" in name:
        raise CheckError(f"unsafe archive member path {name!r}")
    if not member.isfile():
        raise CheckError(f"archive member {name!r} is not a regular file")
    if name not in EXPECTED_PLAYER_FILES:
        raise CheckError(f"unexpected archive member {name!r}")
    if name in seen:
        raise CheckError(f"duplicate archive member {name!r}")
    if member.size > MAX_FILE_BYTES:
        raise CheckError(f"archive member {name!r} exceeds {MAX_FILE_BYTES} bytes")
    return name


def extract_player_archive(raw: bytes) -> list[tuple[str, Path]]:
    """Extract the three expected player files from a verified tar.gz archive."""
    extracted: dict[str, bytes] = {}
    try:
        with tarfile.open(fileobj=io.BytesIO(raw), mode="r:gz") as archive:
            for index, member in enumerate(archive):
                if index >= MAX_PLAYER_FILES:
                    raise CheckError(
                        f"archive holds more than {MAX_PLAYER_FILES} members"
                    )
                name = validate_archive_member(member, set(extracted))
                reader = archive.extractfile(member)
                if reader is None:
                    raise CheckError(f"cannot read archive member {name!r}")
                data = reader.read(MAX_FILE_BYTES + 1)
                if len(data) > MAX_FILE_BYTES:
                    raise CheckError(
                        f"archive member {name!r} exceeds {MAX_FILE_BYTES} bytes"
                    )
                extracted[name] = data
    except (tarfile.TarError, OSError, EOFError) as error:
        raise CheckError(f"invalid player archive: {error}") from error
    if set(extracted) != set(EXPECTED_PLAYER_FILES):
        raise CheckError(
            f"archive members {sorted(extracted)!r} do not match {list(EXPECTED_PLAYER_FILES)!r}"
        )
    out_dir = Path(tempfile.mkdtemp(prefix="praetor-figures-player-"))
    atexit.register(shutil.rmtree, out_dir, True)
    for name in sorted(extracted):
        (out_dir / name).write_bytes(extracted[name])
    return [(f"{PLAYER_URI}/{name}", out_dir / name) for name in sorted(extracted)]


def locked_player_files(lock_file: Path) -> list[tuple[str, Path]]:
    """Resolve player files from player.lock.json, or warn and return none for SVG fallback."""
    if not lock_file.is_file():
        missing = f"unknown ({lock_file.name} is missing; praetorctl adopt writes it)"
        log.warning(SVG_FALLBACK_WARNING, missing)
        return []
    lock = read_player_lock(lock_file)
    release = str(lock.get("release") or "").strip()
    if not release:
        # An unreleased pin is the expected state until the engine's first release carries the
        # archive, so the substitution is stated at INFO: a warning would fail every strict build.
        log.info(SVG_FALLBACK_WARNING, "unreleased")
        return []
    asset = str(lock.get("asset") or PLAYER_ARCHIVE).strip()
    url = release_archive_url(lock, release, asset)
    try:
        raw = download_player_archive(url)
    except (URLError, TimeoutError, OSError) as error:
        log.warning(SVG_FALLBACK_WARNING, f"{release} ({error})")
        return []
    verify_archive_sha256(raw, str(lock.get("sha256") or "").strip(), asset)
    return extract_player_archive(raw)


def published_files() -> list[tuple[str, Path]]:
    """Each file the hook publishes: figures.css, then the player files, as (site URI, path).

    In the engine repository player/ sits beside the hook; in an adopted repository
    player.lock.json names the release archive to fetch and verify. When no release is
    available, the site renders every figure as its SVG and logs one line: INFO for a pin
    without a release, a warning (which fails a strict build) for a missing lock or a failed
    download.
    """
    if PLAYER_DIR.is_dir():
        names = sorted(path.name for path in PLAYER_DIR.iterdir() if path.is_file())
        if len(names) > MAX_PLAYER_FILES:
            where = PLAYER_DIR.as_posix()
            raise CheckError(f"{where} holds more than {MAX_PLAYER_FILES} files")
        served = [(f"{PLAYER_URI}/{name}", PLAYER_DIR / name) for name in names]
    else:
        served = locked_player_files(LOCK_FILE)
    return [(CSS_URI, CSS_FILE), *served]


def on_files(files, config):
    """Add figures.css and the player files to the site.

    A docs_dir file at one of their paths is kept and reported.
    """
    published = (
        _build_assets[1]
        if _build_assets is not None and _build_assets[0] is config
        else published_files()
    )
    for uri, source in published:
        if uri in files.src_uris:
            log.warning(
                "%s: docs_dir already holds this path, "
                "so the figures hook does not publish %s over it",
                uri,
                source.name,
            )
    added = [(uri, source) for uri, source in published if uri not in files.src_uris]
    if added:
        # Imported here so the hook loads without MkDocs in tests.
        from mkdocs.structure.files import File

        for uri, source in added:
            files.append(File.generated(config, uri, abs_src_path=str(source)))
    return files


def on_page_markdown(markdown, page, config, files):
    """Replace the page's figure fences before MkDocs converts the Markdown."""
    figures = Path(config["docs_dir"]) / "assets" / "figures"
    try:
        base = site_base(page.url) + "/assets/figures"
        text, errors = expand(markdown, base, figures)
    except CheckError as error:
        text, errors = markdown, [str(error)]
    for error in errors:
        log.warning("%s: %s", page.file.src_uri, error)
    return text
