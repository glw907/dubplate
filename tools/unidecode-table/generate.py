r"""Re-derive the Unidecode transliteration table's pin values.

``internal/engine/matcher/unidecode`` builds this table at runtime from
the pinned wheel; this script no longer feeds a committed artifact. Its
only remaining job is re-deriving the pins on a Unidecode version bump:
run it, offline, against the newly bumped release in
``tools/venv/.venv``, to a scratch path that is never committed::

    tools/venv/.venv/bin/python tools/unidecode-table/generate.py \
        /tmp/scratch.tbl

Read the printed sha256 alongside the run's other figures and update
three pins in ``internal/engine/matcher/unidecode/pins.go`` by hand:
``Version`` (from the bumped release), ``WheelSHA256`` (from PyPI's
release metadata for the new wheel), ``tableSHA256`` (the printed
output digest, once the new wheel's Go build is confirmed to match it),
and ``dataModuleCount`` (the run's own "blocks" figure below). The Go
build of the new wheel must then match every updated pin, or the bump
is not done.

Table file format (all multi-byte integers big-endian)::

    magic        4 bytes   b"UNTB"
    version_len  1 byte
    version      version_len bytes, UTF-8 (importlib.metadata's own
                 "Unidecode" version string)
    sweep_lo     4 bytes   uint32, first codepoint the digest covers
    sweep_hi     4 bytes   uint32, last codepoint the digest covers
                 (inclusive)
    digest       32 bytes  sha256 over every (codepoint, replacement)
                 pair in [sweep_lo, sweep_hi]; see sweep_digest
    block_count  4 bytes   uint32
    blocks       block_count repetitions of:
        section      2 bytes   uint16, codepoint >> 8
        entry_count  2 bytes   uint16, entries in this block
        entries      entry_count repetitions of:
            length   2 bytes   uint16, 0xFFFF means "no mapping" (None)
            bytes    length bytes, UTF-8 (omitted when length is 0xFFFF)

A block is emitted only for a section unidecode ships a data module
for. unidecode itself returns None for every codepoint in a section it
carries no module for (its own ImportError-catches-missing-file path
in ``_get_repl_str``), which the Go decoder reproduces by treating an
absent section the same way, so no sentinel block is needed for the
gaps.
"""

from __future__ import annotations

import hashlib
import importlib
import pkgutil
import struct
import sys
import warnings
from importlib import metadata
from pathlib import Path

import unidecode

MAGIC = b"UNTB"

# unidecode's own upper bound (unidecode/__init__.py's ">0xeffff -> no
# data" check) and the lower bound of the full sweep this generator
# proves fidelity over: every valid codepoint, so the >0xeffff cutoff,
# the <0x80 passthrough, and the unmapped-section case all land in one
# digest.
SWEEP_LO = 0x000000
SWEEP_HI = 0x10FFFF

# A length this large never occurs in a real mapping (the longest
# shipped entry is 13 bytes), so it is free to reuse as the "no
# mapping" marker.
NONE_SENTINEL = 0xFFFF


def harvest_blocks() -> list[tuple[int, list[str | None]]]:
    """Return every (section, entries) pair unidecode ships a module for.

    entries[i] is what unidecode maps codepoint (section << 8) + i to,
    None meaning unmapped. Reading each module's own data tuple
    (rather than calling the public API per codepoint) is what lets the
    entry count, None count, and mapped-byte total match unidecode's
    shipped tables exactly.
    """
    blocks = []
    for _finder, name, _ispkg in pkgutil.iter_modules(unidecode.__path__):
        if len(name) != 4 or name[0] != "x":
            continue
        try:
            section = int(name[1:], 16)
        except ValueError:
            continue
        mod = importlib.import_module(f"unidecode.{name}")
        blocks.append((section, list(mod.data)))
    blocks.sort(key=lambda block: block[0])
    return blocks


def sweep_digest(lo: int, hi: int) -> bytes:
    """Hash unidecode's own output for every codepoint in [lo, hi].

    translit_test.go recomputes this digest from the embedded table and
    the Go decoder, so a table edit or a regeneration against a
    different unidecode release fails that test loudly rather than
    drifting silently.
    """
    digest = hashlib.sha256()
    with warnings.catch_warnings():
        # Lone surrogates (0xd800-0xdfff) warn on a narrow build; the
        # sweep still wants their result (always None, since unidecode
        # ships no data module for a surrogate section).
        warnings.simplefilter("ignore", RuntimeWarning)
        for codepoint in range(lo, hi + 1):
            repl = unidecode._get_repl_str(chr(codepoint))
            digest.update(struct.pack(">I", codepoint))
            if repl is None:
                digest.update(b"\x00")
                continue
            encoded = repl.encode("utf-8")
            digest.update(b"\x01")
            digest.update(struct.pack(">I", len(encoded)))
            digest.update(encoded)
    return digest.digest()


def write_table(
    path: Path,
    version: str,
    blocks: list[tuple[int, list[str | None]]],
    digest: bytes,
) -> None:
    """Write the binary table this module's docstring documents."""
    version_bytes = version.encode("utf-8")
    with path.open("wb") as f:
        f.write(MAGIC)
        f.write(struct.pack(">B", len(version_bytes)))
        f.write(version_bytes)
        f.write(struct.pack(">II", SWEEP_LO, SWEEP_HI))
        f.write(digest)
        f.write(struct.pack(">I", len(blocks)))
        for section, entries in blocks:
            f.write(struct.pack(">HH", section, len(entries)))
            for entry in entries:
                if entry is None:
                    f.write(struct.pack(">H", NONE_SENTINEL))
                    continue
                encoded = entry.encode("utf-8")
                f.write(struct.pack(">H", len(encoded)))
                f.write(encoded)


def main(argv: list[str]) -> int:
    """Write the table to argv[1] and report the figures a pin bump needs."""
    if len(argv) != 2:
        print(f"usage: {argv[0]} <output-path>", file=sys.stderr)
        return 2

    version = metadata.version("Unidecode")
    blocks = harvest_blocks()

    total_entries = sum(len(entries) for _section, entries in blocks)
    none_count = sum(
        1 for _section, entries in blocks for entry in entries if entry is None
    )
    empty_count = sum(
        1 for _section, entries in blocks for entry in entries if entry == ""
    )
    nonempty_count = total_entries - none_count - empty_count
    mapped_bytes = sum(
        len(entry.encode("utf-8"))
        for _section, entries in blocks
        for entry in entries
        if entry
    )

    print(f"unidecode version: {version}", file=sys.stderr)
    print(f"blocks: {len(blocks)}", file=sys.stderr)
    print(f"entries: {total_entries}", file=sys.stderr)
    print(f"non-empty: {nonempty_count}", file=sys.stderr)
    print(f"none: {none_count}", file=sys.stderr)
    print(f"empty string: {empty_count}", file=sys.stderr)
    print(f"mapped bytes: {mapped_bytes}", file=sys.stderr)

    digest = sweep_digest(SWEEP_LO, SWEEP_HI)
    sweep_range = f"[{SWEEP_LO:#x}, {SWEEP_HI:#x}]"
    print(f"sweep {sweep_range} digest: {digest.hex()}", file=sys.stderr)

    output_path = Path(argv[1])
    write_table(output_path, version, blocks, digest)
    output_sha256 = hashlib.sha256(output_path.read_bytes()).hexdigest()
    print(f"output sha256: {output_sha256}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
