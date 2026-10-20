# intake-1 design: the intake core (7a)

Approved by Geoff, 2026-10-11, with unattended planning and launch. The detail behind every line is the ratified rung 7 spec
(`docs/specs/2026-10-10-rung-7-intake-design.md`, revision 2), sections "The split" through
"Acceptance criteria". This page adds no new decision to it.

## Outcome

After this pass, a signed-in member can send a FLAC album or a set of separate tracks to the server over tus, and
the server checks it and records a verdict. Each submission lands in staging on the music volume, and the gate runs
over its files: the format and caps, the ZIP walk, decode, the audio key, duplicates, and the reject reasons with
their contract wording. The quota and free-space floor reserve space at creation. The submission then sits in its
journey state, readable through the `/api/v1` submission operations. Nothing is admitted to the library yet; that is
`pipeline`. There is no uploader page yet; that is `intake-2`.

## Files and interfaces

- New packages: `internal/reject` (ported from `_salvage/`, wording verbatim), `internal/intake` (the gate, the
  worker, the sweep, the lease), and the Linux platform leaf with its `!linux` refusal.
- Store: migration `0006` (submissions, files, state events, the default "Library" and unlisted collections), the
  older-store map, and the quota and floor queries.
- Config: rung 7's own key file, `keys.go`'s `intake` and `pipeline` class edits, `DUBPLATE_STAGING_DIR` as a
  required wall row.
- `tagio`: the block bound on the chain walk.
- After a full gate on master (14a has merged, so the spec's pause point is moot): `internal/auth` (`RequireMember`),
  `internal/serve` (the tus mount and its wrapper), `internal/ops` and `opshttp` (the submission operations and their
  contract rows), `cmd/dubplate-server` (`serve` startup device check, `doctor` rows, route inventory rows).
- Ops: the serve unit's single music-volume grant, `deploy.sh` pre-creating staging and the library root, and the
  state-dir leg.
- Dependencies: the three the spec names (tusd at its current release among them), plus the pass-opening sweep.
- Carried from PR #4 (Geoff, 2026-10-11): mark the unknown-address toggle and the backup floors Sensitive, one mark
  per key with its test, as the pass's first task.
- Living references edited in place as tasks: the spec's "Owed records" that 7a's code changes (CH, SS hook 12,
  CLAUDE.md's verdict-unit and state-dir lines), and BACKLOG #55, #56's code half, and #71.

## Out of scope

The match probe, the unlisted route's form, the uploader page, resume, and the landing (`intake-2`). The photo
loader, the styled 404, and the sitting (`intake-3`). Admission and layout (`pipeline`). The PCF copy redrafts that
need the register read are recorded, not written. Nothing deploys to the box.

## One end-to-end check

On the real binary under `env -i`, signed in as a member: a tus upload of a small FLAC album fixture reaches
`checked` with a pass verdict; a second album carrying an ID3v2 prefix reaches a reject with the `id3-tags` reason in
its contract wording; a third upload over the free-space floor is refused at creation; and a restart mid-check
resumes the submission through the startup pass. Each state is read back through `/api/v1`.

## Risk class and budget

`auth-data` (it writes `internal/auth`, `internal/serve`, a migration, and `cmd/dubplate-server`), so the PR waits
for your read. The spec's figure for 7a is 9.0M tokens, flag 7.2M, about 6.5 hours of execution; a plan over twice
that comes back to you before launch.

## Overnight shape

With unattended planning approved, this session (the conductor) does the rest without you:

1. Lands the hand-over: files PR #4's triage per your answers, rewrites STATUS, merges that PR itself.
2. Plans `intake-1` against master at Opus high, runs `spec-plan-review` at the depth it sets, and folds the
   findings. A finding that changes this design stops the run with one question in STATUS.
3. Launches the executor (`claude --bg`, Sonnet medium, name `intake-1`) from the main checkout, then watches it with
   the wake-up loop and the lid hold armed.
4. In the morning the PR is on your review desk with its "Found, not done" list.
