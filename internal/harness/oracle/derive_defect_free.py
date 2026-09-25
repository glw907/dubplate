"""Derive the defect-free distance for the string-heuristic adjudication rows.

Invoked as ``derive_defect_free.py <input.json>`` by
internal/harness/oracle's oracle-tagged defect-free driver, once for the
whole subject batch: unlike derive_oracle.py's one-process-per-case
rule, nothing here renders a profile or otherwise changes beets' config
between subjects (an empty scratch ``BEETSDIR`` every time, beets' own
shipped defaults), so the five ``@cached_property`` config reads
derive_album_info.py's own docstring warns about (T15) never go stale
across the batch.

It takes derive_component.py's config shape: no stub, no ``tag_album``,
and no MusicBrainz response fetched over the network. It **does** read
the case's own committed release-lookup response, because an album
distance cannot be computed from scalars alone (ruling A / M6, R19 as
folded, B1). Per subject it runs four steps in order: ``pin_date`` and
``MusicBrainzAPI._normalize_data`` over the decoded response body;
``MusicBrainzPlugin.album_info``; ``match.assign_items`` over the
case's own ``input.items``, re-solved under the substituted table
rather than read from the committed assignment, because a changed
``SD_PATTERNS`` changes every ``track_title`` distance and therefore the
cost matrix; and ``distance.distance``.

The ``SD_PATTERNS`` substitution is a runtime monkeypatch of
``beets.autotag.distance.SD_PATTERNS`` (the module object, not the
``beets.autotag`` package's own ``distance`` name, which resolves to a
function), applied and restored per subject, and never touching
``derive_oracle.py`` or the venv (R20). A subject whose ``pattern_table``
is empty leaves the table untouched, which is the both-ways check's
baseline half: the caller runs the same subjects twice, once with no
substitution to reproduce each case's committed distance bit-identically,
and once with the named table to get the defect-free figure.

The table below carries beets' own literals, the same six
``matcher.SDPatternsNoBracketDiscount`` is derived from, because this
script runs inside Python's ``re`` and needs Python's Unicode-aware
word-boundary check. The Go table carries the featuring row translated
to two ASCII-boundary branches instead, since RE2's word boundary is
ASCII-only, so this script's table and the Go one diverge on that one
row's pattern text while still substituting beets' own credit for a
non-ASCII letter or number directly before a featuring trigger. A
divergence in what either table substitutes is visible in review
rather than hidden in a monkeypatch.
"""

from __future__ import annotations

import importlib
import json
import os
import sys
import tempfile
from typing import Any

from derive_oracle import build_items, pin_date, stable_json

# Bumped whenever a row's shape changes, so an old ledger and a new
# shim cannot be mistaken for each other.
SHIM_VERSION = "1"

# matcher.SDPatternsNoBracketDiscount: BeetsSDPatterns' six discounts
# with the two 0.3-weight bracket patterns removed.
SD_PATTERNS_NO_BRACKET_DISCOUNT = [
    (r"^the ", 0.1),
    (r"[\[\(]?(ep|single)[\]\)]?", 0.0),
    (r"[\[\(]?\b(featuring|feat|ft)\b[\. :].+", 0.1),
    (r"(, )?(pt\.|part) .+", 0.2),
]

PATTERN_TABLES: dict[str, list[tuple[str, float]]] = {
    "SDPatternsNoBracketDiscount": SD_PATTERNS_NO_BRACKET_DISCOUNT,
}


def load_plugin() -> Any:
    """Load beets' shipped default metadata source (musicbrainz)."""
    from beets import metadata_plugins, plugins

    plugins.load_plugins()
    plugin = metadata_plugins.get_metadata_source("musicbrainz")
    if plugin is None:
        raise SystemExit(
            "derive_defect_free.py: beets' shipped defaults load no "
            "musicbrainz metadata source"
        )
    return plugin


def normalized_release(response_path: str, release_mbid: str) -> dict[str, Any]:
    """Read response_path and run it through the normalizer T2 names.

    That is the client's own key/relations normalizer, the half of T2
    that album_info requires of its input.
    """
    from beetsplug._utils.musicbrainz import MusicBrainzAPI

    with open(response_path, encoding="utf-8") as f:
        raw = json.load(f)
    release = MusicBrainzAPI._normalize_data(raw)
    if release["id"] != release_mbid:
        raise SystemExit(
            f"derive_defect_free.py: {response_path}: release id "
            f"{release['id']!r} does not match subject's release_mbid "
            f"{release_mbid!r}"
        )
    return release


def derive_row(plugin: Any, subject: dict[str, Any]) -> dict[str, Any]:
    from beets.autotag.match import AlbumMatch, _recommendation, assign_items
    from beets.util import get_most_common_tags
    from derive_oracle import quiet_action

    pin_date(subject["eval_date"])
    release = normalized_release(subject["response_path"], subject["release_mbid"])
    info = plugin.album_info(release)
    items = build_items(subject["items"])

    table_name = subject["pattern_table"]
    distance_module = importlib.import_module("beets.autotag.distance")
    original_patterns = distance_module.SD_PATTERNS
    if table_name:
        distance_module.SD_PATTERNS = PATTERN_TABLES[table_name]
    try:
        pairs, extra_items, extra_tracks = assign_items(items, info.tracks)
        dist = distance_module.distance(
            get_most_common_tags(items), info, pairs, len(extra_items)
        )
    finally:
        distance_module.SD_PATTERNS = original_patterns

    match = AlbumMatch(dist, info, dict(pairs), extra_items, extra_tracks)
    recommendation = _recommendation([match])
    action = quiet_action(recommendation)

    # pairs holds (item, track) object references, not indices, so the
    # mapping is recovered by identity against the input items and
    # info.tracks (the same technique derive_oracle.py's
    # candidate_record uses for its own assignment field).
    item_index = {id(item): i for i, item in enumerate(items)}
    track_index = {id(track): i for i, track in enumerate(info.tracks)}
    mapping = {
        str(item_index[id(item)]): track_index[id(track)] for item, track in pairs
    }

    return {
        "case_id": subject["case_id"],
        "distance": dist.distance,
        "recommendation": recommendation.name,
        "action": action,
        "pattern_table": table_name,
        "mapping": mapping,
    }


def derive(input_path: str) -> dict[str, Any]:
    with open(input_path, encoding="utf-8") as f:
        document = json.load(f)

    # An empty BEETSDIR is what puts the run on beets' shipped defaults,
    # arithmetically equivalent to the estate profile here (fold, m17):
    # no match: key differs, and the action half resolves through
    # quiet_fallback, skip in both.
    with tempfile.TemporaryDirectory(prefix="dubplate-defect-free-") as beetsdir:
        os.environ["BEETSDIR"] = beetsdir

        import beets

        plugin = load_plugin()
        rows = [derive_row(plugin, subject) for subject in document["subjects"]]

    return {
        "shim_version": SHIM_VERSION,
        "beets_version": beets.__version__,
        "rows": rows,
    }


def main() -> None:
    if len(sys.argv) != 2:
        raise SystemExit("usage: derive_defect_free.py <input.json>")
    # The adapter sets the seed on every invocation, so an unset one is a
    # broken invocation to catch here rather than a slow drift to
    # discover in a corpus diff (derive_oracle.py's own self-check).
    if os.environ.get("PYTHONHASHSEED") != "0":
        raise SystemExit("derive_defect_free.py: PYTHONHASHSEED must be 0")

    real_stdout = sys.stdout
    sys.stdout = sys.stderr
    document = derive(sys.argv[1])
    real_stdout.write(stable_json(document) + "\n")


if __name__ == "__main__":
    main()
