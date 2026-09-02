# ROADMAP

> Strategic initiatives. Managed by `/log-project`. Issues tracked in `BACKLOG.md`.
> Product essence: a group of family or friends building a well-organized
> library for a shared music server. Sequencing ruling (Geoff, 2026-09-02,
> supersedes 2026-09-01): the greenfield dubplate-server build is the one
> active initiative; it launches complete (Go matcher included), TUI and
> local binary follow, `packaging` after that. The family system is not
> live until launch.

## Active

### Greenfield dubplate-server `dubplate-organizer`
The universal shared-library server, built fresh and foundation-first on
a 17-rung ladder per the ratified build spec
(docs/superpowers/specs/2026-09-02-greenfield-server-build.md, revision
2.1), which supersedes the organizer spec's own build order while the
organizer spec continues to govern purview, API shape, and the ops seam
(docs/superpowers/specs/2026-09-01-dubplate-organizer-design.md, amended
through 2026-09-02, folded per "Roadmap revisions this spec carries"
below). Two tracks dependency-merge at rung 15: the architecture spine
(foundation, auth, ops seam + store, intake, pipeline, streamer
drivers/SSO, decision queue, versioned API) and an independent evidence
track (corpus/harness, Go matcher, shadow against the live estate as
oracle), closing with the launch cutover (rung 16: beets DB migration,
box swap, interim retirement) and close-out (rung 17). Launch bar: the
complete server, Go matcher past the two-tier conformance bar, a clean
shadow record, layout parity. Old code is salvage tier, frozen at the
`pre-greenfield` tag under `_salvage/`. Server before TUI is a hard
boundary: rung 10 (two-binary split, local TUI) and the remote-TUI half
of rung 11 move past launch. Rung 1 (foundation: the repo pivot, the
config core, `check`/`doctor` scaffolds) landed 2026-09-02. Next: rung
2's plan (corpus and harness).

Folded from the organizer spec's "Roadmap revisions this spec carries"
(2026-09-01-dubplate-organizer-design.md):
- Streamer integration ships with the Navidrome reference driver only
  (rung 12); Jellyfin and Subsonic drivers are supported by design and
  built on demand. Navidrome is the reference driver, not "the bonus
  only dubplate can offer": this page's product-essence line already
  reads generically ("a shared music server") and needs no further
  edit.
- The built-ins roster is set by policy and probe-backed demotions
  (rung 11), superseding any fixed plugin roster or adoption-band
  ship criterion; the survey backing it is
  `docs/superpowers/research/2026-08-31-beets-plugin-usage.md`.
- The API carries no client assumptions by requirement.
- Sync between rungs is publish-only; mirror and two-way sync are out
  of scope.
- Publish (per-album, through intake) and backup/restore (bulk and
  admin) are separate paths; migration between rungs is not
  backup/restore-only.
- Auth generalizes to a verify-identity seam (rung 5); magic-link is
  the reference implementation, Cloudflare Access an optional
  forward-auth front (rung 12).

## Planned

### Production cleanup and polish `prod-polish`
Largely superseded by the greenfield ruling (2026-09-02): family accounts
and live-use intake move to the new server's launch; what survives is the
debris cleanup (stale failed units, doc drift) and the owed rewrite
post-mortem, foldable into the greenfield build's early passes. Intake for
live findings: `/log-issue` → `BACKLOG.md` with `#project:prod-polish`.
Plan (committed 2026-09-01, ahead of need):
`docs/superpowers/plans/2026-09-01-prod-polish.md`.
Related: #1, #2, #3, #4, #5, #6, #7, #8

### (folded into Active) dubplate as a music library organizer
The 2026-09-01 initiative entry is folded into the Active greenfield
entry above; its settled decisions live in the organizer spec. The
two-binary shape, remote-capable TUI, and plugin built-ins all stand,
sequenced behind the server per the 2026-09-02 rulings.

### Packaging `packaging`
Make the estate easy for a stranger to set up from its ingredients (Navidrome,
beets or its harvested successor, the binaries, their Cloudflare account, any
VPS). Opens with a brainstorm; context staged in agent memory (distribution
model: binary-as-installer vs clone floor; cloud-init generation; existing-box
bootstrap; pre-setup-VPS option map). **Runs after `dubplate-organizer`**
(resequenced 2026-09-01, was after `uploader`) so it packages the final
two-binary shape rather than an intermediate. The `oss-pair` write-up below
doubles as its demand test.

## Someday

### Open-source the pair `oss-pair`
Publish the uploader and the pings Worker as documentation-linked siblings —
a validated-intake drop target and a dead-man's-switch monitor, proven here
first. Preconditions: months of family use, the uploader's auth/validator
seams, and a clear-eyed maintainership decision. First deliverable,
independent of any code release: the validated-intake pattern write-up (what
each layer catches that the others cannot; 907.life is a natural home) —
zero maintainership cost, and the packaging demand test.

## Done

### Uploader `uploader`
Pass A+B built and hardened the full uploader (46-finding fold, fidelity
chain passed, simplified, deployed); superseded pre-launch by the
greenfield ruling. The code is salvage tier for the new server (web
components, tus engine, checks, reject contract); the deployment is
retired. Spec/plan archived under docs/superpowers/. Closed: 2026-09-02.

### Go rewrite `go-rewrite`
The box-side shell layer became one `dubplate` binary (import, backup, beet,
ping-fail, checks, doctor, restore), carrying the musicbox → dubplate estate
rename. All four jobs green on supervised runs, overnight soak confirmed,
shell layer and msmtp deleted everywhere.
Completed: 2026-08-31 (post-mortem owed, tracked as #1)
