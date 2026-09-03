"""Derive one corpus case's oracle record by driving pinned beets.

Invoked as ``derive_oracle.py <rendered-config.yaml> <input.json>`` by
internal/harness/oracle, once per (case, profile): beets caches its
config per process, so a second case in the same process would score
against the first one's weights.

Everything the record carries is executed, never restated. The
recommendation half comes from ``beets.autotag.tag_album``, the action
from beets' own quiet-mode mapping (``_summary_judgment``), and the
applied fields from ``AlbumMatch.apply_metadata`` on in-memory items.
No file is read or written by beets here: there is no importer, no
library, and no audio.

The document on stdout is the only thing this script prints. Anything
beets writes to stdout is redirected to stderr for the whole run.
"""

from __future__ import annotations

import datetime
import hashlib
import json
import math
import os
import shutil
import sys
import tempfile
from typing import Any

# Bumped whenever the record's shape or derivation changes; pinned in
# corpus/manifest.json, so an old record and a new shim cannot be
# mistaken for each other.
SHIM_VERSION = "1"

# Config paths the effective-config hash excludes. Each is derivation
# plumbing that internal/harness/oracle sets per run (a scratch
# directory, the replay stub's ephemeral port) and none can move a
# verdict, so excluding them is what lets two runs of one case hash
# identically.
PLUMBING_PATHS = (
    ("directory",),
    ("library",),
    ("statefile",),
    ("musicbrainz", "host"),
    ("musicbrainz", "https"),
    ("musicbrainz", "ratelimit"),
    ("musicbrainz", "ratelimit_interval"),
)


def pin_date(iso: str) -> datetime.date:
    """Freeze the calendar date this process observes, and return it.

    beets divides by ``datetime.date.today().year`` in both year-penalty
    branches (beets/autotag/distance.py), so the year is matcher input.
    Only the calendar date is frozen: ``time.time`` and
    ``time.monotonic`` keep running, because the MusicBrainz client
    holds a 0.25s floor between requests that waits on elapsed time and
    would never expire against a stopped clock.
    """
    pinned = datetime.date.fromisoformat(iso)

    class PinnedDate(datetime.date):
        @classmethod
        def today(cls) -> datetime.date:
            return pinned

    datetime.date = PinnedDate
    return datetime.date.today()


def stable_json(value: Any) -> str:
    """Serialize value with sorted keys and one fixed float spelling.

    Written out by hand rather than left to ``json.dumps`` because the
    float form is the part that has to be pinned: ``%.17g`` round-trips
    every IEEE-754 double exactly and reads the same on any CPython,
    while the default encoder's shortest-repr form is a property of the
    interpreter build. Raises TypeError on any other type, so a value
    that is not obviously deterministic cannot reach the corpus.
    """
    if isinstance(value, bool):
        return "true" if value else "false"
    if value is None:
        return "null"
    if isinstance(value, str):
        return json.dumps(value)
    if isinstance(value, int):
        return str(value)
    if isinstance(value, float):
        if not math.isfinite(value):
            raise ValueError(f"non-finite float in record: {value!r}")
        return f"{value:.17g}"
    if isinstance(value, dict):
        pairs = (
            f"{json.dumps(str(k))}:{stable_json(v)}"
            for k, v in sorted(value.items(), key=lambda kv: str(kv[0]))
        )
        return "{" + ",".join(pairs) + "}"
    if isinstance(value, (list, tuple)):
        return "[" + ",".join(stable_json(v) for v in value) + "]"
    raise TypeError(f"unserializable value of type {type(value).__name__}")


def jsonable(value: Any) -> Any:
    """Coerce a beets field value into something ``stable_json`` accepts.

    Item paths are bytes; the corpus is file-free, so the only path seen
    here is the empty one, and a lossy decode of any other is still
    better than a derivation that dies at serialization time.
    """
    if isinstance(value, bytes):
        return value.decode("utf-8", "replace")
    if isinstance(value, (list, tuple)):
        return [jsonable(v) for v in value]
    if isinstance(value, dict):
        return {str(k): jsonable(v) for k, v in value.items()}
    if isinstance(value, (str, int, float, bool)) or value is None:
        return value
    return str(value)


def prepare_config(config_path: str, beetsdir: str) -> None:
    """Point beets at a private BEETSDIR holding the rendered profile.

    Set before beets is imported: confuse resolves BEETSDIR when it
    first materializes the config, and reading the workstation's own
    ~/.config/beets would silently change every verdict.
    """
    shutil.copyfile(config_path, os.path.join(beetsdir, "config.yaml"))
    os.environ["BEETSDIR"] = beetsdir


def effective_config() -> dict[str, Any]:
    """Return the config the run scored under, minus derivation plumbing.

    The round trip through json flattens confuse's own view types into
    plain dicts, so the hash is over values rather than over whatever
    wrapper confuse happened to hand back.
    """
    from beets import config

    flat = json.loads(json.dumps(config.flatten(), default=str))
    for path in PLUMBING_PATHS:
        node = flat
        for part in path[:-1]:
            node = node.get(part) if isinstance(node, dict) else None
            if not isinstance(node, dict):
                break
        else:
            node.pop(path[-1], None)
    return flat


def build_items(item_fields: list[dict[str, Any]]) -> list[Any]:
    from beets.library import Item

    items = []
    for fields in item_fields:
        item = Item()
        for key, value in fields.items():
            item[key] = value
        items.append(item)
    return items


def candidate_record(match: Any, item_index: dict[int, int]) -> dict[str, Any]:
    dist = match.distance
    track_index = {id(t): i for i, t in enumerate(match.info.tracks)}

    # Distance.items() drops every key whose weighted distance is zero,
    # but a key added at 0.0 still enlarges the normalizer's
    # denominator (the mechanics study, Q2), so the record enumerates
    # _penalties instead: it is the only view of which keys were added.
    penalties = {key: dist[key] for key in dist._penalties}

    return {
        "source": match.info.data_source,
        "album_id": match.info.album_id,
        "distance": dist.distance,
        "penalties": penalties,
        "tracks": [float(dist.tracks[track]) for _, track in match.item_info_pairs],
        "assignment": {
            str(item_index[id(item)]): track_index[id(track)]
            for item, track in match.item_info_pairs
        },
        "extra_items": [item_index[id(i)] for i in match.extra_items],
        "extra_tracks": [track_index[id(t)] for t in match.extra_tracks],
    }


def quiet_action(recommendation: Any) -> str:
    """Return the action beets itself would take, unprompted.

    dubplate imports with ``beet import -q``, and quiet is a CLI flag
    rather than a profile key, so the flag is set here before calling
    beets' own mapping. The mapping is executed, never restated: a
    change to it in a future beets shows up as a corpus divergence.
    """
    from beets import config
    from beets.ui.commands.import_.session import _summary_judgment

    config["import"]["quiet"] = True
    action = _summary_judgment(recommendation)
    return "ask" if action is None else action.value.lower()


def applied_fields(match: Any, items: list[Any]) -> dict[str, Any]:
    """Apply the match in memory and return each item's fields by index.

    Keyed by the item's index in the case's pinned input order rather
    than merged, because apply is per item: a merged view could not
    express two items disagreeing on a field. This is the field-level
    record only. What ``item.write()`` would put in a file is pass 2c's
    deliverable.
    """
    match.apply_metadata()
    return {
        str(i): {k: jsonable(v) for k, v in item.items()}
        for i, item in enumerate(items)
    }


def derive(config_path: str, input_path: str) -> dict[str, Any]:
    with open(input_path, encoding="utf-8") as f:
        case = json.load(f)

    with tempfile.TemporaryDirectory(prefix="dubplate-shim-") as beetsdir:
        prepare_config(config_path, beetsdir)
        observed = pin_date(case["eval_date"])

        import beets
        from beets import autotag, config, plugins

        plugins.load_plugins()

        items = build_items(case["items"])
        item_index = {id(item): i for i, item in enumerate(items)}
        _, _, proposal = autotag.tag_album(items)

        candidates = [candidate_record(c, item_index) for c in proposal.candidates]
        best = candidates[0] if candidates else None
        weights = config["match"]["distance_weights"]

        record = {
            "case_id": case["case_id"],
            "shim_version": SHIM_VERSION,
            "beets_version": beets.__version__,
            "echoed_date": observed.isoformat(),
            "recommendation": proposal.recommendation.name,
            "action": quiet_action(proposal.recommendation),
            "weights": {k: weights[k].as_number() for k in weights.keys()},
            "thresholds": {
                "strong_rec_thresh": config["match"]["strong_rec_thresh"].as_number(),
                "medium_rec_thresh": config["match"]["medium_rec_thresh"].as_number(),
                "rec_gap_thresh": config["match"]["rec_gap_thresh"].as_number(),
                "max_rec": config["match"]["max_rec"].get(dict),
            },
            "candidates": candidates,
            "assignment": best["assignment"] if best else {},
            "extra_items": best["extra_items"] if best else [],
            "extra_tracks": best["extra_tracks"] if best else [],
            "applied_fields": (
                applied_fields(proposal.candidates[0], items)
                if proposal.candidates
                else {}
            ),
        }
        # Hashed last: the action mapping sets the quiet flag, and the
        # echo has to cover the config the run actually scored under.
        record["config_hash"] = hashlib.sha256(
            stable_json(effective_config()).encode("utf-8")
        ).hexdigest()

    return record


def main() -> None:
    if len(sys.argv) != 3:
        raise SystemExit("usage: derive_oracle.py <config.yaml> <input.json>")
    # Extras come out of a set difference and are only then sorted
    # (beets/autotag/match.py), so two extras with equal sort keys order
    # by hash. The adapter sets the seed; an unset one is a broken
    # invocation, not a slow drift to discover in a corpus diff.
    if os.environ.get("PYTHONHASHSEED") != "0":
        raise SystemExit("derive_oracle.py: PYTHONHASHSEED must be 0")

    # beets prints to stdout on its own account (the quiet-mode mapping
    # announces "Skipping."), and stdout carries the record.
    real_stdout = sys.stdout
    sys.stdout = sys.stderr
    record = derive(sys.argv[1], sys.argv[2])
    real_stdout.write(stable_json(record) + "\n")


if __name__ == "__main__":
    main()
