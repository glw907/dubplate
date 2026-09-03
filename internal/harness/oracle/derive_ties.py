"""Measure whether one corpus case's optimal track assignment is unique.

Invoked as ``derive_ties.py <rendered-config.yaml> <input.json>`` by
internal/harness/oracle's oracle-tagged tie driver, once per case, under
the one-process-per-case rule derive_oracle.py already documents.

The tie question is asked of the best candidate alone, and its
``TrackInfo`` list is fetched the way beets fetches it (the plugin's own
``album_for_id`` against the replay stub) rather than re-parsed out of a
committed response body, so the tracks scored here are the tracks beets
scored. ``lap.lapjv`` picks one optimal assignment; a Go solver may pick
a different one at equal cost and still produce a different final
distance, because beets recomputes the per-track distances with
``incl_artist=va`` while the assignment objective ignores the track
artist. That is what makes a tied case worth detecting.
"""

from __future__ import annotations

import json
import os
import sys
import tempfile
from typing import Any

import lap
import numpy as np
from derive_oracle import build_items, pin_date, prepare_config, stable_json

# The cost a re-solve gives to the pair it forbids. Every real
# track_distance lands in [0, 1], so this is unreachable by an
# alternative assignment: it is only ever paid when the forbidden pair
# had no alternative at all, which is the shape a huge recorded
# difference reports.
FORBIDDEN_COST = 1e6

# The margin the optimum must clear to count as unique. Mirrors
# analysis.TieMargin, and sits three orders of magnitude above the
# threshold-proximity epsilon so the two regimes cannot overlap.
TIE_MARGIN = 1e-9


def cost_matrix(items: list[Any], tracks: list[Any]) -> list[list[float]]:
    """Build the matrix ``assign_items`` itself solves.

    ``incl_artist`` stays False because that is the objective beets
    assigns under (beets/autotag/match.py, assign_items), even on a VA
    release whose final per-track distances do include the artist.
    """
    from beets.autotag import track_distance

    return [
        [float(track_distance(item, track, incl_artist=False)) for track in tracks]
        for item in items
    ]


def solve(costs: list[list[float]]) -> tuple[float, list[tuple[int, int]]]:
    """Solve the assignment and return its cost and its item/track pairs.

    The cost is summed from the pairs rather than taken from lapjv's own
    return value: ``extend_cost`` pads a non-square matrix with a value
    derived from the matrix maximum, so lapjv's total moves when a
    forbidden pair raises that maximum, and two solves would stop being
    comparable.
    """
    _, _, assigned_item_idxs = lap.lapjv(np.array(costs), extend_cost=True)
    pairs = [
        (item_idx, track_idx)
        for track_idx, item_idx in enumerate(assigned_item_idxs)
        if item_idx != -1
    ]
    return sum(costs[i][j] for i, j in pairs), pairs


def uniqueness(costs: list[list[float]]) -> tuple[float, float]:
    """Return the optimum and the cheapest re-solve that forbids a pair.

    Forbidding each assigned pair in turn and re-solving is a definite,
    terminating test: the optimum is unique exactly when every one of
    those re-solves costs strictly more than it.
    """
    optimal, pairs = solve(costs)
    if not pairs:
        raise SystemExit("derive_ties.py: nothing to assign; no item matched any track")

    best_forbidden = float("inf")
    for i, j in pairs:
        forbidden = [row[:] for row in costs]
        forbidden[i][j] = FORBIDDEN_COST
        cost, _ = solve(forbidden)
        best_forbidden = min(best_forbidden, cost)
    return optimal, best_forbidden


def measure(config_path: str, input_path: str) -> dict[str, Any]:
    with open(input_path, encoding="utf-8") as f:
        case = json.load(f)

    with tempfile.TemporaryDirectory(prefix="dubplate-ties-") as beetsdir:
        prepare_config(config_path, beetsdir)
        pin_date(case["eval_date"])

        from beets import metadata_plugins, plugins

        plugins.load_plugins()

        album = metadata_plugins.album_for_id(case["album_id"], case["source"])
        if album is None:
            raise SystemExit(
                f"derive_ties.py: {case['source']} returned no release for "
                f"{case['album_id']}"
            )

        costs = cost_matrix(build_items(case["items"]), album.tracks)
        optimal, best_forbidden = uniqueness(costs)

    difference = best_forbidden - optimal
    return {
        "case_id": case["case_id"],
        "album_id": case["album_id"],
        "optimal_cost": optimal,
        "best_forbidden_cost": best_forbidden,
        "difference": difference,
        "unique": difference > TIE_MARGIN,
    }


def main() -> None:
    if len(sys.argv) != 3:
        raise SystemExit("usage: derive_ties.py <config.yaml> <input.json>")
    # Mirrors derive_oracle.py's own self-check: the Go adapter sets the
    # seed on every invocation, so an unset one is a broken invocation to
    # catch here rather than a slow drift to discover in a corpus diff
    # (defense in depth for a hand-run against this script directly).
    if os.environ.get("PYTHONHASHSEED") != "0":
        raise SystemExit("derive_ties.py: PYTHONHASHSEED must be 0")

    real_stdout = sys.stdout
    sys.stdout = sys.stderr
    measurement = measure(sys.argv[1], sys.argv[2])
    real_stdout.write(stable_json(measurement) + "\n")


if __name__ == "__main__":
    main()
