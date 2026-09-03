"""Derive the component-level distance table from pinned beets.

Invoked as ``derive_component.py`` with no arguments by
internal/harness/oracle's oracle-tagged component driver. It is the
second, narrower beets entry point: no stub, no ``tag_album``, no
MusicBrainz response, and no rendered profile, because a call to a
distance primitive touches none of them. beets' own shipped defaults are
the config, which is also what both corpus profiles leave the ``match:``
surface at.

Every row is a literal argument tuple applied to a real primitive,
mirroring the parametrize tables in beets' test/autotag/test_distance.py
at the pinned version. The table records what pinned beets returned; it
asserts nothing. Its consumer is rung 3's Go matcher, which is what will
finally compare its own primitives against these outputs.
"""

from __future__ import annotations

import os
import re
import sys
import tempfile
from collections.abc import Iterator
from contextlib import contextmanager
from typing import Any

from derive_oracle import stable_json

# Bumped whenever a row's shape changes, so an old table and a new
# emitter cannot be mistaken for each other.
SHIM_VERSION = "1"


class Pattern:
    """A compiled-regex argument, kept serializable for the table.

    ``Distance._eq`` treats a compiled pattern as a match rather than an
    equality, and that branch is one of the rows worth pinning, so the
    argument has to survive both the call and the JSON encoding.
    """

    def __init__(self, pattern: str, ignorecase: bool = False) -> None:
        self.pattern = pattern
        self.ignorecase = ignorecase

    def compiled(self) -> re.Pattern[str]:
        return re.compile(self.pattern, re.I if self.ignorecase else 0)

    def encoded(self) -> dict[str, Any]:
        return {"regex": self.pattern, "ignorecase": self.ignorecase}


# The adder tables, one group per Distance.add_* method, each a sequence
# of calls against one Distance: the penalty list accumulates down the
# group exactly as beets' own test_add_methods drives it.
ADDER_TABLES: list[tuple[str, list[tuple[Any, ...]]]] = [
    (
        "equality",
        [
            ("ghi", ["abc", "def", "ghi"]),
            ("xyz", ["abc", "def", "ghi"]),
            ("abc", Pattern("ABC", ignorecase=True)),
        ],
    ),
    ("expr", [(True,), (False,)]),
    ("number", [(1, 1), (1, 2), (2, 1), (-1, 2)]),
    (
        "priority",
        [
            ("abc", "abc"),
            ("def", ["abc", "def"]),
            ("gh", ["ab", "cd", "ef", Pattern("GH", ignorecase=True)]),
            ("xyz", ["abc", "def"]),
        ],
    ),
    ("ratio", [(25, 100), (10, 5), (-5, 5), (5, 0)]),
    ("string", [("abc", "bcd"), ("abc", None), (None, None)]),
]

# Weights the aggregate rows score under, matching test_distance.py's
# own fixture: the arithmetic is only legible against weights that are
# not all equal.
AGGREGATE_WEIGHTS = {"album": 4.0, "medium": 2.0, "data_source": 2.0}

# String pairs beets asserts score zero: casing, punctuation, article
# rotation, the (EP) strip, "&" to "and", bracket stripping, and
# diacritic transliteration.
STRING_MATCHING_PAIRS = [
    ("Some String", "Some String"),
    ("Some String", "Some.String!"),
    ("Some String", "sOME sTring"),
    ("My Song (EP)", "My Song"),
    ("The Song Title", "Song Title, The"),
    ("A Song Title", "Song Title, A"),
    ("An Album Title", "Album Title, An"),
    ("", ""),
    ("Untitled", "[Untitled]"),
    ("And", "&"),
    ("\xe9\xe1\xf1", "ean"),
]

# The SD_PATTERNS relative-weight rows: each names two candidate strings
# and the reference both are scored against, and beets asserts the
# second scores strictly closer.
STRING_RELATIVE_TRIPLES = [
    ("XXX Band Name", "The Band Name", "Band Name"),
    ("One .Two.", "One (Two)", "One"),
    ("One .Two.", "One [Two]", "One"),
    ("My Song blah Someone", "My Song feat Someone", "My Song"),
]

# Patterns that must not crash on their own, with nothing left after the
# pattern is stripped.
STRING_SOLO_PAIRS = [("The ", ""), ("(EP)", "(EP)"), (", An", "")]

# TestTrackDistance's rows: one TrackInfo, four item spellings, scored
# with the track artist included.
TRACK_DISTANCE_ROWS = [
    ("identical", "title", "artist"),
    ("tolerate-va", "title", "Various Artists"),
    ("different-artist", "title", "different artist"),
    ("different-title", "different title", "artist"),
]

# TestDataSourceDistance's parameter space: the item's data source, the
# candidate's, the mismatch penalty the other plugin reports, the
# data_source weight, and whether more than one metadata source is
# loaded.
DATA_SOURCE_ROWS = [
    ("match", "Original", "Original", 0.5, 1.0, True),
    ("mismatch", "Original", "Other", 0.5, 1.0, True),
    ("mismatch-reversed", "Other", "Original", 0.5, 1.0, True),
    ("mismatch-unknown", "Original", "unknown", 0.5, 1.0, True),
    ("mismatch-no-info", "Original", None, 0.5, 1.0, True),
    ("mismatch-no-original-multiple-sources", None, "Other", 0.5, 1.0, True),
    ("match-no-original-but-single-source", None, "Other", 0.5, 1.0, False),
    ("match-unknown", "unknown", "unknown", 0.5, 1.0, True),
    ("mismatch-max-penalty", "Original", "Other", 1.0, 1.0, True),
    ("mismatch-high-weight", "Original", "Other", 0.5, 5.0, True),
    ("match-no-penalty", "Original", "Other", 0.0, 1.0, True),
    ("match-no-weight", "Original", "Other", 0.5, 0.0, True),
]


def call_arg(value: Any) -> Any:
    """Unwrap a table literal into the value the beets call receives."""
    if isinstance(value, Pattern):
        return value.compiled()
    if isinstance(value, list):
        return [call_arg(v) for v in value]
    return value


def encoded_arg(value: Any) -> Any:
    """Render a table literal as the table's own JSON argument."""
    if isinstance(value, Pattern):
        return value.encoded()
    if isinstance(value, (list, tuple)):
        return [encoded_arg(v) for v in value]
    return value


def row(row_id: str, method: str, args: Any, output: Any) -> dict[str, Any]:
    return {"id": row_id, "method": method, "args": args, "output": output}


@contextmanager
def weights(overrides: dict[str, float]) -> Iterator[None]:
    """Override Distance's class-level weights for the duration.

    ``Distance._weights`` is a cached class property, so a config edit
    after the first access would not reach it; beets' own test patches
    the materialized dict for the same reason. A key beets does not ship
    is removed again rather than restored, so one row cannot leak a
    weight into the next.
    """
    from beets.autotag import Distance

    original = {
        key: Distance._weights[key] for key in overrides if key in Distance._weights
    }
    added = [key for key in overrides if key not in Distance._weights]
    Distance._weights.update(overrides)
    try:
        yield
    finally:
        Distance._weights.update(original)
        for key in added:
            del Distance._weights[key]


@contextmanager
def metadata_sources(penalty: float, multiple: bool) -> Iterator[None]:
    """Install stand-in metadata source plugins named Original and Other.

    ``MetadataSourcePlugin`` derives a plugin's ``data_source`` from its
    class name, which is why the two classes are named the way they are:
    the penalty lookup finds them by that name.
    """
    from beets import metadata_plugins
    from beets.metadata_plugins import MetadataSourcePlugin

    class StubSourcePlugin(MetadataSourcePlugin):
        def album_for_id(self, *args, **kwargs): ...
        def track_for_id(self, *args, **kwargs): ...
        def candidates(self, *args, **kwargs): ...
        def item_candidates(self, *args, **kwargs): ...

    class OriginalPlugin(StubSourcePlugin):
        pass

    class OtherPlugin(StubSourcePlugin):
        @property
        def data_source_mismatch_penalty(self) -> float:
            return penalty

    loaded = [OtherPlugin()] if not multiple else [OriginalPlugin(), OtherPlugin()]
    original = metadata_plugins.find_metadata_source_plugins
    metadata_plugins.find_metadata_source_plugins = lambda: loaded
    metadata_plugins.get_penalty.cache_clear()
    try:
        yield
    finally:
        metadata_plugins.find_metadata_source_plugins = original
        metadata_plugins.get_penalty.cache_clear()


def adder_rows() -> Iterator[dict[str, Any]]:
    from beets.autotag import Distance

    dist = Distance()
    dist.add("add", 1.0)
    yield row(
        "add/basic",
        "add",
        {"key": "add", "calls": [[1.0]]},
        {"penalties": dist._penalties["add"]},
    )

    for key, arg_sets in ADDER_TABLES:
        calls: list[tuple[Any, ...]] = []
        for i, arg_set in enumerate(arg_sets):
            calls.append(arg_set)
            dist = Distance()
            method = getattr(dist, f"add_{key}")
            for call in calls:
                method(key, *(call_arg(a) for a in call))
            yield row(
                f"add_{key}/{i}",
                f"add_{key}",
                {"key": key, "calls": [encoded_arg(c) for c in calls]},
                {"penalties": dist._penalties[key]},
            )


def aggregate_rows() -> Iterator[dict[str, Any]]:
    """Emit the raw, max, and normalized arithmetic over added penalties.

    The zero rows are the denominator rule the mechanics study names: a
    key added at 0.0 contributes nothing to the numerator and still
    enlarges the normalizer, and a set of penalties whose weights are all
    zero normalizes to 0.0 rather than dividing by zero.
    """
    from beets.autotag import Distance

    groups = [
        (
            "aggregate/normalized",
            [("album", 0.5), ("media", 0.25), ("media", 0.75)],
            {},
        ),
        ("aggregate/one-penalty", [("album", 0.5)], {}),
        (
            "aggregate/zero-added-key",
            [("data_source", 0.0), ("album", 0.5), ("medium", 0.25), ("medium", 0.75)],
            {},
        ),
        ("aggregate/sorted-penalties", [("album", 0.1875), ("medium", 0.75)], {}),
        ("aggregate/sorted-penalties-tied", [("album", 0.375), ("medium", 0.75)], {}),
        (
            "aggregate/all-zero-weights",
            [("zero_weighted", 1.0)],
            {"zero_weighted": 0.0},
        ),
        ("aggregate/empty", [], {}),
    ]
    for row_id, calls, extra in groups:
        overrides = AGGREGATE_WEIGHTS | extra
        with weights(overrides):
            dist = Distance()
            for key, value in calls:
                dist.add(key, value)
            yield row(
                row_id,
                "add",
                {"weights": overrides, "calls": [list(c) for c in calls]},
                {
                    "distance": dist.distance,
                    "raw_distance": dist.raw_distance,
                    "max_distance": dist.max_distance,
                    "per_key": {key: dist[key] for key in dist._penalties},
                    "items": [list(pair) for pair in dist.items()],
                },
            )

    with weights(AGGREGATE_WEIGHTS):
        dist = Distance()
        calls = [("album", 0.5), ("label", 0.0), ("medium", 0.25)]
        for key, value in calls:
            dist.add(key, value)
        yield row(
            "generic_penalty_keys/excludes-zero-penalties",
            "generic_penalty_keys",
            {"weights": AGGREGATE_WEIGHTS, "calls": [list(c) for c in calls]},
            {"keys": dist.generic_penalty_keys},
        )


def string_rows() -> Iterator[dict[str, Any]]:
    from beets.autotag import string_dist

    for i, (str1, str2) in enumerate(STRING_MATCHING_PAIRS):
        yield row(
            f"string_dist/matching/{i}",
            "string_dist",
            [str1, str2],
            {"distance": string_dist(str1, str2)},
        )

    yield row(
        "string_dist/different",
        "string_dist",
        ["Some String", "Totally Different"],
        {"distance": string_dist("Some String", "Totally Different")},
    )

    for i, (str1, str2, reference) in enumerate(STRING_RELATIVE_TRIPLES):
        for label, candidate in (("far", str1), ("near", str2)):
            yield row(
                f"string_dist/relative/{i}/{label}",
                "string_dist",
                [candidate, reference],
                {"distance": string_dist(candidate, reference)},
            )

    for i, (str1, str2) in enumerate(STRING_SOLO_PAIRS):
        yield row(
            f"string_dist/solo/{i}",
            "string_dist",
            [str1, str2],
            {"distance": string_dist(str1, str2)},
        )


def track_rows() -> Iterator[dict[str, Any]]:
    from beets.autotag import TrackInfo, track_distance
    from beets.library import Item

    info = {"title": "title", "artist": "artist"}
    for row_id, title, artist in TRACK_DISTANCE_ROWS:
        dist = track_distance(
            Item(artist=artist, title=title), TrackInfo(**info), incl_artist=True
        )
        yield row(
            f"track_distance/{row_id}",
            "track_distance",
            {
                "item": {"title": title, "artist": artist},
                "track_info": info,
                "incl_artist": True,
            },
            {"distance": dist.distance, "penalties": dict(dist._penalties)},
        )

    for row_id, item_source, info_source, penalty, weight, multiple in DATA_SOURCE_ROWS:
        with weights({"data_source": weight}), metadata_sources(penalty, multiple):
            dist = track_distance(
                Item(data_source=item_source),
                TrackInfo(data_source=info_source, title=""),
            )
            yield row(
                f"track_distance/data_source/{row_id}",
                "track_distance",
                {
                    "item": {"data_source": item_source},
                    "track_info": {"data_source": info_source, "title": ""},
                    "incl_artist": False,
                    "data_source_mismatch_penalty": penalty,
                    "weights": {"data_source": weight},
                    "multiple_data_sources": multiple,
                },
                {"distance": dist.distance, "penalties": dict(dist._penalties)},
            )


def data_source_rows() -> Iterator[dict[str, Any]]:
    """Emit Distance.add_data_source on its own, below track_distance.

    The adder carries the whole "before is set, or a second source is
    loaded" condition, and the single-source case is the one shape a
    corpus cassette case cannot reach under the single-metadata-source
    ruling.
    """
    from beets.autotag import Distance

    calls = [
        ("match", "Original", "Original", True),
        ("mismatch", "Original", "Other", True),
        ("no-before-single-source", None, "Other", False),
        ("no-before-multiple-sources", None, "Other", True),
    ]
    for row_id, before, after, multiple in calls:
        with weights({"data_source": 2.0}), metadata_sources(0.5, multiple):
            dist = Distance()
            dist.add_data_source(before, after)
            yield row(
                f"add_data_source/{row_id}",
                "add_data_source",
                {
                    "before": before,
                    "after": after,
                    "multiple_data_sources": multiple,
                    "weights": {"data_source": 2.0},
                    "data_source_mismatch_penalty": 0.5,
                },
                {"penalties": dict(dist._penalties)},
            )


def derive() -> dict[str, Any]:
    import beets

    rows = [
        *adder_rows(),
        *aggregate_rows(),
        *string_rows(),
        *track_rows(),
        *data_source_rows(),
    ]
    return {"beets": beets.__version__, "shim_version": SHIM_VERSION, "rows": rows}


def main() -> None:
    if len(sys.argv) != 1:
        raise SystemExit("usage: derive_component.py")

    # An empty BEETSDIR is what puts the run on beets' shipped defaults:
    # reading the workstation's own ~/.config/beets would silently
    # rescale every weighted output in the table.
    with tempfile.TemporaryDirectory(prefix="dubplate-component-") as beetsdir:
        os.environ["BEETSDIR"] = beetsdir
        real_stdout = sys.stdout
        sys.stdout = sys.stderr
        table = derive()
        real_stdout.write(stable_json(table) + "\n")


if __name__ == "__main__":
    main()
