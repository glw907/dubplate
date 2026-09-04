"""Derive the file-level applied-tag ledger from pinned beets' write path.

Invoked as ``derive_file_tags.py <input.json>`` by
internal/harness/oracle's oracle-tagged file-tags driver. ``input.json``
is a JSON array of subjects, each naming a committed case id, an item
index, and that item's frozen ``Expected.AppliedFields`` record (the
in-memory ``apply_metadata()`` result ``derive_oracle.py`` already
produced for the corpus). This shim is a second, independent derivation
entry point: it starts no stub, issues no request, renders no profile,
and loads no plugin, matching ``derive_component.py``'s minimalism, and
it is never called from ``oracle.Derive``. No item it builds is ever
given a path that reaches the corpus: every write lands in this
process's own scratch copy of the committed template FLAC, never in a
case's own directory.

For each subject, a beets ``Item`` is reconstructed from the recorded
fields, written into a fresh copy of the vendored template FLAC, and
read back through an independent ``mediafile.MediaFile`` parse of the
bytes on disk. ``Item.write()`` returns ``None``, so that read-back is
the only signal there is of what actually landed in the container; the
ledger records that, never what anyone expected either to be.

The value keys split into a writable set, computed from the pinned
venv as the intersection of ``Item._media_fields`` and
``MediaFile.sorted_fields()`` (what ``MediaFile.update()`` actually
iterates and therefore the only fields ``Item.write()`` can touch), and
a read-only remainder outside that intersection: audio properties
``MediaFile`` reports from the container's own ``STREAMINFO`` (bit
depth, bitrate, channels, format, length, sample rate, and similar),
never written and never a property of any subject. Only the writable
set is a per-row value; the read-only remainder is recorded once, in a
``template`` block shared by the whole ledger, since its values come
from the template blob and never from a case.
"""

from __future__ import annotations

import json
import os
import shutil
import sys
import tempfile
from typing import Any

from derive_oracle import build_items, stable_json

# Bumped whenever a row's shape changes, so an old ledger and a new
# shim cannot be mistaken for each other.
SHIM_VERSION = "1"

TEMPLATE_PATH = os.path.join(
    os.path.dirname(os.path.abspath(__file__)), "testdata", "template.flac"
)


def field_lists(beetsdir: str) -> tuple[list[str], list[str]]:
    """Return the (writable, read-only) field-name lists, off the venv.

    ``Item._media_fields`` (``beets/library/models.py:731``) is the
    78-key intersection of every ``MediaFile``-readable field and
    ``Item``'s own column names. ``Item.write()`` filters to that set
    before calling ``MediaFile.update()``, which in turn only ever sets
    a field named in ``MediaFile.sorted_fields()`` (the writable set):
    the nine-key remainder never reaches a file at all. Both lists are
    computed here, never hardcoded, so a future ``mediafile`` release
    that moves a field between the two sets changes this shim's output
    rather than silently drifting from it. This is also where beets is
    first imported, right after ``BEETSDIR`` is set, and the isolation
    that depends on is proven, not assumed, immediately below: an
    import that silently read the runner's own ``~/.config/beets``
    instead would leave ``os.environ["BEETSDIR"]`` disagreeing with the
    scratch directory the caller just created for this run alone.
    """
    from beets.library import Item

    if os.environ.get("BEETSDIR") != beetsdir:
        raise SystemExit(
            "derive_file_tags.py: BEETSDIR isolation broken: expected "
            f"{beetsdir!r}, found {os.environ.get('BEETSDIR')!r} at the "
            "moment beets.library.Item was imported"
        )

    from mediafile import MediaFile

    media_fields = set(Item._media_fields)
    sorted_fields = set(MediaFile.sorted_fields())
    writable = sorted(media_fields & sorted_fields)
    readonly = sorted(media_fields - sorted_fields)
    return writable, readonly


def assert_json_primitive(key: str, value: Any) -> None:
    """Assert value is already a JSON primitive, or a list of them.

    ``AppliedFields``' own encoder falls back to ``str(value)`` for a
    beets field type it cannot otherwise serialize
    (``derive_oracle.py``'s ``jsonable``); measured against the pinned
    venv, no field this shim reads ever needs that fallback, but a
    future field type silently could, turning a value into its
    stringified repr with nothing here to notice. This assertion is
    what keeps that from shipping unnoticed.
    """
    if isinstance(value, list):
        for element in value:
            assert_json_primitive(key, element)
        return
    if isinstance(value, (str, int, float, bool)) or value is None:
        return
    raise TypeError(f"{key}: AppliedFields value is not a JSON primitive: {value!r}")


def template_block(readonly: list[str]) -> dict[str, Any]:
    """Read the template blob's own read-only properties, once.

    A fresh, unwritten copy of the committed template: these nine
    values are a property of the container's STREAMINFO, identical for
    every subject, and never touched by any subject's write.
    """
    from mediafile import MediaFile

    with tempfile.TemporaryDirectory(prefix="dubplate-file-tags-template-") as scratch:
        path = os.path.join(scratch, "template.flac")
        shutil.copyfile(TEMPLATE_PATH, path)
        mf = MediaFile(path)
        return {field: getattr(mf, field) for field in readonly}


def derive_row(
    scratch: str, subject: dict[str, Any], writable: list[str]
) -> dict[str, Any]:
    """Write one subject's item into its own fresh template copy.

    Each subject gets its own path inside the caller's
    ``TemporaryDirectory``, matching ``derive_oracle.py``'s and
    ``derive_component.py``'s own scratch-directory pattern: a shared
    path would let one subject's write leak into the next subject's
    read-back.
    """
    from mediafile import MediaFile

    fields = subject["applied_fields"]
    for key, value in fields.items():
        assert_json_primitive(key, value)

    item = build_items([fields])[0]

    path = os.path.join(scratch, f"{subject['case_id']}-{subject['item_index']}.flac")
    shutil.copyfile(TEMPLATE_PATH, path)
    item.write(path)

    # A fresh MediaFile parse of the bytes on disk, never the object
    # write() already had open: this is the independent read-back
    # Item.read() itself performs (models.py:917), and the ledger wants
    # the raw per-field values mediafile reports rather than a second
    # Item.
    mf = MediaFile(path)
    values = {field: getattr(mf, field) for field in writable}
    return {
        "case_id": subject["case_id"],
        "item_index": subject["item_index"],
        "values": values,
    }


def derive(subjects: list[dict[str, Any]], beetsdir: str) -> dict[str, Any]:
    import beets

    writable, readonly = field_lists(beetsdir)

    with tempfile.TemporaryDirectory(prefix="dubplate-file-tags-subjects-") as scratch:
        rows = [derive_row(scratch, subject, writable) for subject in subjects]

    return {
        "shim_version": SHIM_VERSION,
        "beets_version": beets.__version__,
        "writable_fields": writable,
        "readonly_fields": readonly,
        "template": template_block(readonly),
        "rows": rows,
    }


def main() -> None:
    if len(sys.argv) != 2:
        raise SystemExit("usage: derive_file_tags.py <input.json>")

    with open(sys.argv[1], encoding="utf-8") as f:
        subjects = json.load(f)

    # An empty BEETSDIR is what puts the run on beets' shipped defaults
    # and proves this run never reads the workstation's own
    # ~/.config/beets, which could silently change a written value or
    # raise on a non-bool id3v23 setting the runner happens to have
    # configured (Item.write() resolves beets.config["id3v23"] before it
    # opens the file).
    with tempfile.TemporaryDirectory(prefix="dubplate-file-tags-") as beetsdir:
        os.environ["BEETSDIR"] = beetsdir
        real_stdout = sys.stdout
        sys.stdout = sys.stderr
        document = derive(subjects, beetsdir)
        real_stdout.write(stable_json(document) + "\n")


if __name__ == "__main__":
    main()
