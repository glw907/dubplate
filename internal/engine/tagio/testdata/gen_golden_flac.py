"""Author the vendored golden FLAC fixture, once.

Run manually, from the repo root, under the pinned venv, as
``tools/venv/.venv/bin/python3
internal/engine/tagio/testdata/gen_golden_flac.py``: it copies
``internal/harness/oracle/testdata/template.flac`` to a scratch path,
writes ``INPUT`` through ``MediaFile.update``, adds one small invented
picture, and (re)writes ``golden.flac`` next to itself. It is never
invoked again on its own, by anything; ``make derive`` and the
repository gate never run it (Resolution R1). This script is the
fixture's provenance record, not a build step, and its own output was
hand-transcribed into ``golden.txt`` field by field. ``golden.txt``
records the pinned beets, mediafile and mutagen versions, ``INPUT``
restated field by field, the golden's byte size, its decoded block
chain in order, and its decoded comment sequence in on-disk order;
nothing here reconstructs ``INPUT`` from ``golden.flac``'s bytes.

``INPUT`` is this script's own declared input, chosen so the emitted
comment set covers ``INV:409-427``'s measured comment keys (``INV``
publishes the output of a round trip, not an input list; no such list
appears there for this script to restate). ``INPUT`` widens that
covering set to reach two further sets: every one of Task 8's sixteen
ledger-silent keys (empty or null on every row of the corpus's own
ledger, so they need a real value here to be exercised at all), and
every corpus-unreachable case a mediafile write can express: the zero
year on one of the two date fields (the other carries INV's own
measured date, so the month/day decomposition rule stays exercised
too), an empty list, and ``catalognum`` and ``barcode`` both empty.
Two further corpus-unreachable cases stay out of ``INPUT`` by
construction:
mediafile writes valid UTF-8 only, and mutagen's ``is_valid_key``
(``mutagen/_vorbis.py:25-41``) refuses a comment name above ``0x7D``,
so an invalid-UTF-8 sequence and a ``U+007E`` field name cannot reach
disk through this path. Both remain Task 8's hand-built unit test
cases, exercised directly against the block encoding rather than
through mediafile.
"""

from __future__ import annotations

import os
import shutil
import tempfile

from mediafile import Image, ImageType, MediaFile

# The sixteen keys the corpus ledger never exercises (INV:195-209):
# every row of the twelve-subject corpus carries these as null or the
# empty string, so a golden fixture needs its own values to reach the
# writer's shapes for them at all.
_LEDGER_SILENT = {
    "acoustid_fingerprint": "invented-fingerprint-value",
    "acoustid_id": "11111111-2222-3333-4444-555555555555",
    "comments": "Invented comment text.",
    "encoder": "Invented Encoder 1.0",
    "genres": ["Ambient", "Drone"],
    "grouping": "Invented Grouping",
    "initial_key": "Am",
    "lyrics": "la la la",
    "r128_album_gain": -2.25,
    "r128_track_gain": -1.5,
    "remixers": ["Invented Remixer"],
    "rg_album_gain": -6.5,
    "rg_album_peak": 0.654321,
    "rg_track_gain": -7.25,
    "rg_track_peak": 0.123456,
    "subtitle": "Invented Subtitle",
}

# The four corpus-unreachable cases a mediafile write can still
# express: a zero year on the primary date field only (the second
# date field keeps INV's own measured 1999-03-04 below, so the
# month/day decomposition rule stays exercised even though this file
# also carries a zero-date case), an empty list on a field the ledger
# never leaves empty, and the two fields C5 exists for (their trailing
# styles are read-only, so an empty write reads back absent even
# though the comment itself is empty on disk).
_CORPUS_UNREACHABLE = {
    "year": 0,
    "original_year": 1999,
    "original_month": 3,
    "original_day": 4,
    "arrangers": [],
    "catalognum": "",
    "barcode": "",
}

# This script's own declared covering shape, chosen to reach every
# comment key INV:409-427 measured: the collision pair (RELEASETYPE /
# MUSICBRAINZ_ALBUMTYPE), the three-style totals, and the two-style
# pairs it measured.
_DECLARED_COVERING_SHAPE = {
    "album": "Al",
    "albumartists": ["X", "Y"],
    "artist": "A",
    "bpm": 120,
    "comp": True,
    "disc": 1,
    "disctotal": 2,
    "title": "T",
    "track": 3,
    "tracktotal": 12,
    "albumtype": "album",
    "albumtypes": ["album", "live"],
}

INPUT: dict[str, object] = {
    **_DECLARED_COVERING_SHAPE,
    **_CORPUS_UNREACHABLE,
    **_LEDGER_SILENT,
}

# PNG's eight-byte signature, so the block is sniffable by magic bytes
# without being a decodable image; 8a asserts nothing past the parse.
_PICTURE_DATA = b"\x89PNG\r\n\x1a\n" + b"invented-picture-payload"


def main() -> None:
    here = os.path.dirname(os.path.abspath(__file__))
    template = os.path.join(
        here, "..", "..", "..", "harness", "oracle", "testdata", "template.flac"
    )
    out = os.path.join(here, "golden.flac")

    with tempfile.TemporaryDirectory() as scratch_dir:
        scratch = os.path.join(scratch_dir, "golden.flac")
        shutil.copyfile(template, scratch)

        mf = MediaFile(scratch)
        mf.update(INPUT)
        mf.images = [Image(data=_PICTURE_DATA, desc="cover", type=ImageType.front)]
        mf.save()

        shutil.copyfile(scratch, out)

    print(f"wrote {out} ({os.path.getsize(out)} bytes)")


if __name__ == "__main__":
    main()
