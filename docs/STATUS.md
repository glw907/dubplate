# dubplate status

## Current state
Master's head is the `admin-api-1` merge (`8477287`, PR #4); its close gate (`gate.sh --merge`) was green on
`e6440e9`. Nothing deploys. The landed rungs, the run order, and the next pass are in ROADMAP, "Path to 0.8". Work a
later pass carries sits in BACKLOG under the `#rung:<name>` tag of the rung that takes it. PR #4's "Found, not done"
list is triaged: item 1 rides in `intake-1`, items 2, 3, and 6 to 11 are BACKLOG #75 to #82, and items 4 and 5 are
dropped.

The box is unchanged: the dubplate-server binary `d1e93e9`, the serve unit installed and disabled, the `auth`,
`roster`, and `unidecode` leaves, one admin member, and `/srv/music/library` (6.6 GB).

## Next action
`intake-1` is stopped at T11's security read: session `intake-1`, id `930921a1-a634-461c-91be-416fd394c230`, branch
`worktree-intake-1` (head `a58ed88`, T0 to T13 committed, gate green, nothing left but the Close). The read's re-read
of the one fix commit returned `fix` again, and a second `fix` is a stop under the plan.

What is closed: the zip64 preallocation (M2) and panic containment (m1). What is open, all in the FLAC gate
(`internal/intake/gate/file.go`): (1) the CPU bound grows with the declared format, since nothing caps the sample
rate (a 20-bit field reaches 1,048,575 Hz), so a hostile file can cost hours of MD5 work and stall the single check
worker; (2) mewkiz logs one stderr line per frame for a 176.4 kHz or 88.2 kHz stream inside `ParseNext`, before any
check; (3) a frame whose block size is 0 or smaller than a subframe's order panics the decoder, which is contained now
but records `could-not-check`, so the member is told to retry and every retry panics again.

Question: take one more fix commit that (a) adds fixed constants, not settings, `maxSampleRate` 768 kHz refused at
step 3, plus a decoded-work-per-input-byte bound with a fixed allowance for silence, (b) parses each frame header first
(`frame.New` over the gate's own buffered reader, check, then `fr.Parse()`) so the log line and the `Correlate` panic
never run on a refused frame, and refuses a block size of 0 or below STREAMINFO's minimum on any frame but the last,
and (c) refuses `corrupt` when a subframe's sample count differs from the block size? Recommendation: yes. The
constants add no setting and change no approved design line, the re-read then reads that commit once, and the close
runs as planned. The alternative is to ship `intake-1` as is and file the three items as the first work of `intake-2`,
which does not deploy before the probe lands either way.

## Open decisions
- R-6d21 overrides the settings-home pass's rule that the server never reads its environment. Geoff may reopen it;
  nothing waits on it.
- The 0.8 and 0.9 launches are FLAC-only. Whether the group needs a non-FLAC format sooner is still open.
- About 41 stale `DUBPLATE_*` exports in `~/.local/secrets` make `check` fail outside `env -i`. They need pruning on
  the workstation.
- One Run-D.M.C. frame (`commons:139197098`) and five maybes wait in `photos/.candidates/` for Geoff's call.
- TypeScript 7.0.2 and `@types/node` 26.6.5 are majors held for Geoff (the 2026-10-11 dependency sweep).
