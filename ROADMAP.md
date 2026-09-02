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
The universal shared-library server, built fresh and foundation-first per
the organizer spec (docs/superpowers/specs/2026-09-01-dubplate-organizer-design.md,
amended through 2026-09-02) and the ratified rulings
(docs/superpowers/research/2026-09-02-greenfield-server-rulings.md).
Approach A: an architecture spine (config core, magic-link auth to the
hardened birth requirements, ops seam + store, intake, pipeline, streamer
drivers/SSO, decision queue, versioned API) plus an independent evidence
track (corpus/harness, Go matcher, shadow against the running estate as
oracle), meeting at integration and closing with the launch cutover
(beets DB migration, box swap, interim retirement). Launch bar: the
complete server, Go matcher past the two-tier conformance bar. Old code
is salvage tier. Server before TUI is a hard boundary. Next: the fresh
session writes the greenfield build spec from the rulings doc.

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
