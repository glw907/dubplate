r"""Author the featuring row's parity fixtures, once, from pinned beets.

Run manually, from the repo root, under the pinned venv, as::

    tools/venv/.venv/bin/python3 \\
        internal/engine/matcher/testdata/gen_featuring_golden.py

It writes ``featuring-golden.json`` (curated and drawn inputs, each
paired with beets' own substitution on the featuring row) and
``python-word-ranges.json`` (Python's whole ``\\w`` class, as sorted
inclusive code-point ranges). Neither the repository gate nor
``make derive`` invokes this script: it is the two fixtures'
provenance record, run once against the pinned venv and never again on
its own.

The script refuses to write either file when the installed beets
version, or the featuring row's literal read from
``beets.autotag.distance.SD_PATTERNS``, differs from what this module
states below, so a beets bump cannot silently regenerate a golden
against a different pattern. It also refuses when the draw fires on
fewer than ``MIN_FIRED`` distinct strings, since a draw that weak
cannot tell the Go translation from a defect.

No draw token below is a code point Unicode first assigned after
version 15.1 (CPython 3.13's own ``unicodedata.unidata_version``): a
token from that range would make the golden encode a known upstream
divergence as if it were parity, since the Go translation and beets
disagree by construction on any such code point until a later
toolchain assigns it a definite word-character class.
"""

from __future__ import annotations

import json
import random
import re
import sys
import unicodedata
from pathlib import Path

import beets
from beets.autotag.distance import SD_PATTERNS

EXPECTED_BEETS_VERSION = "2.14.1"
EXPECTED_LITERAL = r"[\[\(]?\b(featuring|feat|ft)\b[\. :].+"

SEED = 907
DRAW_COUNT = 10_000
MIN_FIRED = 150

# In the order random.choice draws from. Every token is assigned in
# Unicode 15.1 or earlier; see the module docstring.
TOKENS = [
    "a",
    "ä",
    "Ω",
    "ǅ",
    "²",
    "Ⅻ",
    "́",
    "_",
    "2",
    " ",
    "(",
    "[",
    ":",
    ".",
    "ft",
    "feat",
    "featuring",
    "e",
    "r",
    "-",
    " ",
    "ü",
    "ß",
]

HERE = Path(__file__).resolve().parent
FEATURING_GOLDEN = HERE / "featuring-golden.json"
WORD_RANGES = HERE / "python-word-ranges.json"

# The brief's table inputs (autotag-unicode-word-boundary brief), each
# curated at its own case and again fully lowercased: the row's
# literal carries no IGNORECASE flag, so the two forms can disagree on
# whether the trigger matches at all, and production always lowercases
# before this row runs.
BRIEF_INPUTS = [
    ("geschaft", "Geschäft:"),
    ("ruft", "Rüft:"),
    ("omegaft", "Ωft"),
    ("draft", "Draft:"),
    ("arft", "Ärft."),
    ("twoft", "2ft"),
    ("cafeft", "Café ft."),
    ("emberwharfft", "Ember Wharf ft."),
]

# One string per preceding-character class the brief names, each
# placing the class's character directly before the trigger.
PRECEDING_CLASS_INPUTS = [
    ("a-umlaut", "äft. x"),
    ("sharp-s", "ßft. x"),
    ("omega", "Ωft. x"),
    ("titlecase-dz-caron", "ǅft. x"),
    ("superscript-two", "²ft. x"),
    ("roman-numeral-twelve", "Ⅻft. x"),
    ("combining-acute", "éft. x"),
    ("underscore", "_ft. x"),
    ("no-break-space", " ft. x"),
    ("paren-open", "(ft. x"),
    ("bracket-open", "[ft. x"),
    ("copyright", "©ft. x"),
    ("trademark", "™ft. x"),
    ("euro", "€ft. x"),
]

# The shape rows: the trigger at the very start of the string, two
# triggers on one line (each swallowed or matched under Python's
# greedy, non-DOTALL .+), and two triggers split by a newline (which
# the rewrite's negated class must also treat as a boundary).
SHAPE_INPUTS = [
    ("start", "ft. x"),
    ("two-triggers", "a ft. b\nc ft. d"),
    ("newline", "x\nft. y\nft. z"),
]


def check_pin() -> re.Pattern[str]:
    """Verify the pinned beets version and literal, or exit without writing."""
    if beets.__version__ != EXPECTED_BEETS_VERSION:
        sys.exit(
            f"refusing to write: beets version is {beets.__version__!r}, "
            f"generator expects {EXPECTED_BEETS_VERSION!r}"
        )
    literal = SD_PATTERNS[2][0]
    if literal != EXPECTED_LITERAL:
        sys.exit(
            f"refusing to write: featuring literal is {literal!r}, "
            f"generator expects {EXPECTED_LITERAL!r}"
        )
    return re.compile(literal)


def curated_rows(pattern: re.Pattern[str]) -> list[list[str]]:
    """Build the labeled curated rows, each substituted through pattern."""
    rows: list[tuple[str, str]] = []
    for label, text in BRIEF_INPUTS:
        rows.append((f"{label}-mixed", text))
        rows.append((f"{label}-lower", text.lower()))
    rows.extend(PRECEDING_CLASS_INPUTS)
    rows.extend(SHAPE_INPUTS)

    labels = [label for label, _ in rows]
    if len(labels) != len(set(labels)):
        sys.exit(f"refusing to write: duplicate curated labels in {labels!r}")

    return [[label, text, pattern.sub("", text)] for label, text in rows]


def drawn_rows(pattern: re.Pattern[str]) -> list[list[str]]:
    """Draw DRAW_COUNT strings under SEED, dedupe, and substitute each."""
    random.seed(SEED)
    drawn: list[str] = []
    for _ in range(DRAW_COUNT):
        length = random.randint(1, 8)
        drawn.append("".join(random.choice(TOKENS) for _ in range(length)))
    distinct = list(dict.fromkeys(drawn))

    rows = [[text, pattern.sub("", text)] for text in distinct]
    fired = sum(1 for text, output in rows if output != text)
    print(
        f"drawn: {DRAW_COUNT}, distinct: {len(distinct)}, fired: {fired}",
        file=sys.stderr,
    )
    if fired < MIN_FIRED:
        sys.exit(
            f"refusing to write: only {fired} drawn rows fired, "
            f"want at least {MIN_FIRED}"
        )
    return rows


def word_ranges() -> list[list[str]]:
    r"""Return Python's whole \w class as sorted, non-overlapping hex ranges.

    Surrogates (U+D800 to U+DFFF) are excluded, matching the other
    fixed-code-point sweep in this package
    (TestStringDistCaseMappingSweep's 1,112,064-point sweep).
    """
    word_char = re.compile(r"\w")
    ranges: list[tuple[int, int]] = []
    start: int | None = None
    for codepoint in range(0x110000):
        if 0xD800 <= codepoint <= 0xDFFF:
            if start is not None:
                ranges.append((start, codepoint - 1))
                start = None
            continue
        if word_char.match(chr(codepoint)) is not None:
            if start is None:
                start = codepoint
        elif start is not None:
            ranges.append((start, codepoint - 1))
            start = None
    if start is not None:
        ranges.append((start, 0x10FFFF))
    return [[f"{lo:04X}", f"{hi:04X}"] for lo, hi in ranges]


def write_ndjson(
    path: Path, header: dict[str, object], sections: dict[str, list]
) -> None:
    """Write header plus named row-array sections, one JSON value per line.

    ensure_ascii=False keeps a curated row's non-ASCII characters and
    its embedded newline readable in the diff; the one-row-per-line
    shape is what keeps the file's size proportional to its row count
    rather than to json.dumps' default indentation.
    """
    lines = ["{", f'  "header": {json.dumps(header, ensure_ascii=False)},']
    items = list(sections.items())
    for i, (name, rows) in enumerate(items):
        lines.append(f'  "{name}": [')
        for j, row in enumerate(rows):
            comma = "," if j < len(rows) - 1 else ""
            lines.append(f"    {json.dumps(row, ensure_ascii=False)}{comma}")
        lines.append("  ]" + ("," if i < len(items) - 1 else ""))
    lines.append("}")
    text = "\n".join(lines) + "\n"
    json.loads(text)  # the hand-assembled file must parse as one JSON value
    path.write_text(text, encoding="utf-8")
    print(f"wrote {path} ({path.stat().st_size} bytes)", file=sys.stderr)


def main() -> int:
    if not FEATURING_GOLDEN.parent.is_dir():
        sys.exit(f"refusing to write: {FEATURING_GOLDEN.parent} is not a directory")

    pattern = check_pin()
    header = {
        "beets_version": EXPECTED_BEETS_VERSION,
        "cpython_version": ".".join(map(str, sys.version_info[:3])),
        "unicodedata_version": unicodedata.unidata_version,
        "literal": EXPECTED_LITERAL,
        "seed": SEED,
        "draw_count": DRAW_COUNT,
    }

    curated = curated_rows(pattern)
    drawn = drawn_rows(pattern)
    print(f"curated rows: {len(curated)}", file=sys.stderr)
    write_ndjson(FEATURING_GOLDEN, header, {"curated": curated, "drawn": drawn})

    ranges = word_ranges()
    print(f"word-class ranges: {len(ranges)}", file=sys.stderr)
    write_ndjson(WORD_RANGES, header, {"ranges": ranges})

    return 0


if __name__ == "__main__":
    sys.exit(main())
