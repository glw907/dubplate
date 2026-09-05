"""Dump the translated AlbumInfo pinned beets builds for one corpus case.

Invoked as ``derive_album_info.py <rendered-config.yaml> <input.json>`` by
internal/harness/oracle's oracle-tagged AlbumInfo driver, once per case,
under the one-process-per-case rule derive_oracle.py already documents.
``input.json`` names the case, its pinned evaluation date, and the
release ids the case's own request log fetches by id.

The releases are fetched the way beets fetches them, through the
plugin's own API against the replay stub, rather than re-parsed out of a
committed response body: key normalization and the ``relations``
regrouping happen inside that client (beetsplug/_utils/musicbrainz.py),
so a shim that read the body itself would grade the Go port against
input beets never saw. Each row carries the pre-merge
``album_info(release)`` value, plus, on a pseudo-release linked backward
by a ``transl-tracklisting`` relation, the merged ``album_for_id``
result, so the merge is gradeable on its own.

The projection is the fourteen matcher.AlbumInfo fields and the eight
matcher.TrackInfo fields, and nothing else: rung 3 computes no applied
field, and the remaining forty-odd AlbumInfo keys are rung 8's surface.
A beets ``None`` is written as JSON null rather than coerced, because
the null and the empty string are different inputs to the distance path.
"""

from __future__ import annotations

import json
import os
import sys
import tempfile
from typing import Any

from derive_oracle import pin_date, prepare_config, stable_json

# Bumped whenever a row's shape changes, so an old ledger and a new
# shim cannot be mistaken for each other.
SHIM_VERSION = "1"

# The fourteen matcher.AlbumInfo fields, by beets' own attribute names.
ALBUM_FIELDS = (
    "artist",
    "album",
    "va",
    "media",
    "mediums",
    "year",
    "original_year",
    "country",
    "label",
    "catalognum",
    "albumdisambig",
    "album_id",
    "data_source",
)

# The eight matcher.TrackInfo fields, by beets' own attribute names.
TRACK_FIELDS = (
    "title",
    "artist",
    "length",
    "index",
    "medium_index",
    "medium",
    "track_id",
    "data_source",
)


def load_plugin(beetsdir: str) -> Any:
    """Load the configured metadata source and prove the run is isolated.

    beets is first imported here, right after ``BEETSDIR`` is set, so the
    isolation the whole derivation rests on is checked rather than
    assumed: an import that silently read the workstation's own
    ~/.config/beets would leave ``os.environ["BEETSDIR"]`` disagreeing
    with the scratch directory the caller just created. The plugin is
    constructed once and its config is never mutated afterwards, because
    five of its config reads are ``@cached_property`` on the instance
    (beetsplug/musicbrainz.py:246-268).
    """
    from beets import plugins

    if os.environ.get("BEETSDIR") != beetsdir:
        raise SystemExit(
            "derive_album_info.py: BEETSDIR isolation broken: expected "
            f"{beetsdir!r}, found {os.environ.get('BEETSDIR')!r} at the "
            "moment beets.plugins was imported"
        )

    plugins.load_plugins()

    from beets import metadata_plugins

    plugin = metadata_plugins.get_metadata_source("musicbrainz")
    if plugin is None:
        raise SystemExit(
            "derive_album_info.py: the rendered profile loads no musicbrainz "
            "metadata source"
        )
    return plugin


def json_primitive(field: str, value: Any) -> Any:
    """Return value unchanged, or exit if it is not a JSON primitive.

    A silent ``str(value)`` fallback is how a ledger ships a wrong type,
    so a beets field that ever stops being a scalar fails the derivation
    here rather than becoming its own repr in the committed record
    (derive_file_tags.py's ``assert_json_primitive`` is the precedent).
    """
    if isinstance(value, (str, int, float, bool)) or value is None:
        return value
    raise SystemExit(
        f"derive_album_info.py: {field}: value is not a JSON primitive: {value!r}"
    )


def project(info: Any) -> dict[str, Any]:
    """Project one beets AlbumInfo onto the ledger's twenty-two fields."""
    row = {field: json_primitive(field, getattr(info, field)) for field in ALBUM_FIELDS}
    row["tracks"] = [
        {
            field: json_primitive(f"tracks[{i}].{field}", getattr(track, field))
            for field in TRACK_FIELDS
        }
        for i, track in enumerate(info.tracks)
    ]
    return row


def is_pseudo_release(release: dict[str, Any]) -> bool:
    """Report whether album_for_id would merge this release with another.

    The trigger is beetsplug/musicbrainz.py:785-798: the exact status
    string plus a backward ``transl-tracklisting`` relation. The
    ``release_relations`` key exists only because the client's own
    ``_normalize_data`` regrouped the generic ``relations`` array by
    target type.
    """
    if release["status"] != "Pseudo-Release":
        return False
    return any(
        rel["type"] == "transl-tracklisting" and rel["direction"] == "backward"
        for rel in release.get("release_relations", [])
    )


def derive_row(plugin: Any, case_id: str, mbid: str) -> dict[str, Any]:
    release = plugin.mb_api.get_release(mbid)
    row = {
        "case_id": case_id,
        "release_mbid": mbid,
        "album": project(plugin.album_info(release)),
    }
    if is_pseudo_release(release):
        merged = plugin.album_for_id(mbid)
        if merged is None:
            raise SystemExit(
                f"derive_album_info.py: {case_id}: album_for_id returned no "
                f"release for pseudo-release {mbid}"
            )
        row["merged"] = project(merged)
    return row


def derive(config_path: str, input_path: str) -> dict[str, Any]:
    with open(input_path, encoding="utf-8") as f:
        case = json.load(f)

    with tempfile.TemporaryDirectory(prefix="dubplate-album-info-") as beetsdir:
        prepare_config(config_path, beetsdir)
        pin_date(case["eval_date"])
        plugin = load_plugin(beetsdir)

        import beets

        rows = [derive_row(plugin, case["case_id"], mbid) for mbid in case["lookups"]]

    return {
        "shim_version": SHIM_VERSION,
        "beets_version": beets.__version__,
        "rows": rows,
    }


def main() -> None:
    if len(sys.argv) != 3:
        raise SystemExit("usage: derive_album_info.py <config.yaml> <input.json>")
    # The ledger records a set-derived value: AlbumInfo.media is
    # ``medias.pop()`` over a set comprehension (musicbrainz.py:654-658),
    # whose iteration order is hash-dependent. An unset seed is a broken
    # invocation to catch here rather than a drift to discover in a diff.
    if os.environ.get("PYTHONHASHSEED") != "0":
        raise SystemExit("derive_album_info.py: PYTHONHASHSEED must be 0")

    real_stdout = sys.stdout
    sys.stdout = sys.stderr
    document = derive(sys.argv[1], sys.argv[2])
    real_stdout.write(stable_json(document) + "\n")


if __name__ == "__main__":
    main()
