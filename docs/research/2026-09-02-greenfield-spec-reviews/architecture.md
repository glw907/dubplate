# Greenfield-spec adversarial review (architecture and evidence), 2026-09-02

Spec: `docs/specs/2026-09-02-greenfield-server-build.md`
Lens: ladder dependency order, launch-bar completeness, rung scoping,
the evidence track against the fresh matcher-mechanics study, shadow
design, cutover and rollback, the repo pivot, process realism.

The rulings (`research/2026-09-02-greenfield-server-rulings.md`) are
owner rulings and out of scope. Findings below argue fidelity to them,
never against them.

---

## Blockers

**[blocker] The pivot list omits `e2e/`, which imports `internal/` directly, so the gate goes red the moment rung 1 moves the tree**
Spec:172-181 enumerates exactly two sets: `cmd/`, `internal/`, `web/`
move to `_salvage/`; `scripts/`, `tests/*.bats`, `systemd/`,
`cloud-init/`, `compose.yaml`, `config/`, `env/` stay live and gated.
`e2e/` is in neither. It is a real Go package in the module
(1,377 lines across six files) and it imports the moving tree:
`e2e/backup_test.go:10` imports `github.com/glw907/dubplate/internal/lock`,
`e2e/import_test.go:18-19` imports `internal/importer` and
`internal/lock`, and `e2e/main_test.go:64` builds `./cmd/dubplate`.
`_salvage/` is invisible to the go tool (underscore-prefixed
directories are excluded from wildcard patterns), so after the move
`go vet ./...`, `go test ./...`, and `golangci-lint run ./...` — every
Go leg of `make check`, which `scripts/check.sh:33` runs — fail to
compile on rung 1's first gate run. The spec's stated property, that
the Go toolchain and the gate exclude `_salvage/` (Spec:173-174), is
true of the toolchain and false of the repo as it stands until `e2e/`
is dispositioned. It also matters beyond the red gate: `e2e/` is the
only end-to-end coverage of the binary the box is still running, and
deleting or freezing it silently removes the interim estate's
integration net during the exact window the spec promises the estate
keeps running (Spec:61-63).

**[blocker] `tests/*.bats` does not exist; the pivot section is written against a repo state one commit stale**
Spec:178-181 lists `tests/*.bats` among the estate operations files
that "stay live and gated, because they run the interim box today."
`tests/` is empty. Commit `04d5ae3` (2026-08-31, "Land T13+T14 repo
side: shell layer deleted") retired the bats suite — its own message
says "bats and msmtprc retire" — and `scripts/check.sh` has no bats
step at all (its legs are `make check`, shellcheck, vale, compose
config, systemd verify). The spec's own gate description in CLAUDE.md
carries the same staleness, and Spec:365-366's "The `web/` build joins
the repo gate here" (rung 7) is the third instance: the web build is
already in the gate today (`Makefile`'s `check: vet lint test
tidy-check web-check web-test web-build`). Concrete failure: rung 1's
plan will be written from this list. It will budget for preserving a
suite that is gone, will not budget for `e2e/`, and will not know that
"rewire the gate" means removing three web legs the spec elsewhere
says arrive at rung 7. The pivot list has to be re-derived from `ls`,
not from the spec's memory of the tree.

**[blocker] Rung 15 retires the artifacts its own rollback depends on, in the same rung**
Spec:423-427 states the rollback as "re-PUT the prior tunnel ingress,
disable externalized auth, restore the beets snapshot," and the same
rung ends with "the interim estate and `_salvage/` retired." Those two
clauses fight. The organizer spec's "Rollback" section defines
reverting as restoring the beets snapshot *and* "re-importing
post-swap sessions from the ledger's session records," with "beets
stays installed through a deprecation window." Greenfield keeps half
of that (the snapshot) and drops the other half (the ledger re-import,
the deprecation window), while adding a retirement step that deletes
the salvage tree holding `internal/ledger` and `internal/importer`,
and rebuilds `systemd/`, `scripts/deploy.sh`, `compose.yaml`, and
`cloud-init/` for the new server in the same pass. After that pass
there is no deployed old binary, no deploy path that can install one,
no unit files for it, and no ledger reader. A restored beets snapshot
with nothing that can drive an import from it is not a rollback. Three
things must be named and are not: what re-installs the pre-cutover
estate (the pre-pivot tag is named as an *emergency guardian fix*
fallback at Spec:176-177, never as the rollback target), how long
beets and the old units stay resident after the swap, and what
restores the on-disk layout if the new pipeline moved a file before
anyone noticed.

**[blocker] The shadow oracle has no material, and the fallback it rests on cannot accept a contribution**
Spec:151-154: "the new matcher runs against the same material the
box's live beets pipeline imports." Spec:155-160 keeps "at least 30
days and at least 20 real albums, whichever completes later" as the
launch-gate minimum, and calls the whole-build window "strictly
stronger." Against the wind-down the rulings record (rulings, "Interim
estate state", executed the same day): FileBrowser is retired
(container, compose service, env var, `inbox.907.life` DNS and Access
app), the uploader's serve unit and nudge units are disabled, the
uploader tunnel path rule is gone, both interim Access apps are gone,
and "the tunnel serves music.907.life → Navidrome only." There is no
path by which a contributor can put an album into the box's import
directory. STATUS confirms the same picture. So the live beets
pipeline imports whatever Geoff copies over ssh and nothing else. Two
consequences. First, "20 real albums" — the organizer's own gloss is
"real contributor albums," and its rung 4 calls shadow "the only
evidence from real contributor material rather than a curated corpus"
— is now either unreachable or satisfiable only by Geoff hand-feeding
20 albums, which makes it a second curated corpus with worse
provenance, not a stronger oracle. The "strictly stronger" claim
(Spec:156) does not survive: the window got longer and the material
got thinner, and the spec only prices the first half. Second,
Spec:61-63's fallback — "the honest outcome is that the interim estate
keeps running and the launch waits" — describes an estate that today
serves playback and accepts no contributions. The real fallback is
"the family stays offline indefinitely," which is a materially
different thing to ratify. Either the shadow rung re-opens a
contribution path (an explicit deliverable it does not have), or the
floor is restated in terms of material that will actually exist.

**[blocker] The evidence track's corpus quadruple omits the two impurities the mechanics study found, and the build spans a year boundary**
Spec:328-331 defines a case as "the quadruple corpus (input, frozen
MusicBrainz responses, beets config, expected verdict)." The mechanics
study's Q1 closes with what the harness must pin or record per case:
"item order, the evaluation year, the full effective config (or a
named profile), the single-source constraint, and the {unidecode,
jellyfish, lap, beets} version quadruple in the corpus manifest."
Two of those are load-bearing and absent from the spec. The wall-clock
year (study Q1.1, `beets/autotag/distance.py:477` and `487-489`) is
the only genuine impurity in the arithmetic: both year-distance
branches divide by a span ending at `datetime.date.today().year`.
Frozen expected verdicts computed in 2026 drift at the calendar
boundary for every case exercising the year penalty with an
`original_year` present. The spec's own design guarantees that
boundary is crossed: the shadow window is "the whole build"
(Spec:156), scheduled at rung 4 of 15, with a 30-day floor after it.
The failure is quiet — the fast gate reddens or, worse, greens on
cases whose expected verdicts silently moved — and it lands on the one
rung the spec calls the initiative's existential risk. The library
version manifest is the same class: `unidecode`'s transliteration
tables sit directly in `_string_dist_basic` (study Q2;
`beets/autotag/distance.py:55-56`), so an unpinned bump changes edit
distances underneath a frozen corpus. Both belong in the quadruple's
definition, and the Go matcher needs the reference date as an explicit
input rather than a call to `time.Now()`.

**[blocker] The largest port surface in the existential-risk rung is unnamed: the WS response to AlbumInfo/TrackInfo layer**
Spec:337-343 (rung 3) enumerates "the distance and penalty model,
album-level matching, stable candidate ordering, numeric distance
agreement..., beets' full tunable surface..., and the shared
MusicBrainz rate limiter." The tree at Spec:194-195 gives
`matcher/` ("distance and penalty model, candidates, verdicts") and
`musicbrainz/` ("WS client, shared rate limiter, cassette replay").
Nothing in either names the translation between them, and the
mechanics study shows that translation is verdict-bearing, not
plumbing. `_normalize_data` rewrites every key and regroups relations
(study Q5; `beetsplug/_utils/musicbrainz.py:724-775`); `album_info` /
`track_info` build the objects the matcher actually sees
(`beetsplug/musicbrainz.py:631-695`, `433-469`), including per-medium
track extraction, pregap inclusion, data- and video-track filtering,
and track-level overrides of recording length, credit, and title
(`555-600`). Above that, `import.languages` and
`import.ignored_alias_types` drive `_preferred_alias`
(`musicbrainz.py:128-155`), which rewrites candidate titles and artist
names *before any distance is computed* (`158-162`, `329-334`, `443`,
`648`), and `preferred.countries` selects which release event supplies
the candidate's country **and date**, feeding the year penalty
(`musicbrainz.py:165-177`) — a second, non-obvious verdict path the
study flags as missed by the believed tunable list. This is a large
body of exact-fidelity porting sitting between the frozen responses
and the distance model, it is where conformance will actually break,
and the L band on rung 3 was set without it on the list. Name it as a
deliverable of rung 3 (or a leaf of its own) and re-argue the band.

---

## Major

**[major] Rung 5 is an L wearing an M, and the spec half-admits it**
Spec:348-355 bands identity and auth M, then immediately hedges: "the
plan may split it." A band that needs a split hedge is the next band
up. Count the deliverables the rung names: tokens; sessions with epoch
and absolute cap; durable nonces; the login and consumption pages
(Go-templated, no JavaScript, shared design language); SMTP send; rate
limiting; the verify-identity seam with provenance; the forward-auth
verify endpoint; a minimal serve shell under the one owned prefix; the
`web-auth-security-reviewer` gate inside the rung. Ten, against
CLAUDE.md's pass-sizing rule that a task passing roughly four
deliverables must be called out. On top of that the rung inherits
Spec:225-230's instruction that its plan "walk the full review texts
... as checklists, minors included" — that is 6 blockers plus 12
majors plus 2 minors in `security.md` and 3 criticals plus 16 further
items in `universality.md`, each requiring a design decision and a
test. The organizer spec bands one M at "roughly the repo's recent
pass ceilings, about 5M tokens." This is not that.

**[major] Rung 15 is an L wearing an M, on a deliverable the organizer spec already funded separately**
Spec:418-431 packs into one M: a migrator against four failure modes
including the no-move dry-run; the beets database archived to R2; the
box swap under a written and *exercised* rollback; `systemd/`,
`deploy.sh`, `compose.yaml`, and `cloud-init/` rebuilt per
vps-conventions; secrets minted through the age-store flow; the
ingress-before-ExtAuth ordering; retirement of the interim estate and
`_salvage/`; the `UIWelcomeMessage` universality breach; and an
attended verification gate. The organizer spec's build order funds two
of those separately: the migrator is a named part of its rung 5 (M, an
M that also carries the store, the schema, and the migrations
mechanism), and the box-facing rebuild is a *funded deliverable of the
split rung* — "The rename is a funded deliverable of the split rung.
It touches the systemd units, `scripts/deploy.sh`, the `/etc` and
`/opt` paths, the pings check slugs, the runbook, and STATUS, and it
ships with a box-write plan" (organizer, "The two binaries"). The
organizer could band its swap rung S because "rungs 1 through 6 built
and proved everything it needs." Greenfield's rung 15 has that
property for nothing: the migrator, the box rebuild, and the rollback
exercise are all first-built here, on the one irreversible rung, under
an attended gate.

**[major] Rung 4's S band prices the runner and not the deployment, and the deployment has no path**
Spec:344-347 bands shadow "S to build, continuous to launch."
Spec:462-464 says rung 4 "deploys the additive shadow runner." What
does deploying it require at that point in the ladder? A second binary
on a box whose only install path is `scripts/deploy.sh`, which at
`scripts/deploy.sh:262-263` runs `make -C "$REPO_ROOT" build-linux` —
a Makefile target that builds `./cmd/dubplate`, the tree that moved to
`_salvage/` at rung 1. So the deploy tool the spec says "stays live"
(Spec:178-181) cannot build anything after rung 1, and its rebuild is
assigned to rung 15. Rung 4 additionally needs a systemd unit under
vps-conventions with its `ReadWritePaths` pre-created, an
`EnvironmentFile` and a config surface the greenfield binary will
accept, a divergence-log format with nowhere to persist it (the store
is rung 6), and the observation mechanism itself — how the runner sees
the import timer's material — which Spec:166-168 explicitly defers to
the rung's own plan. Deferring the mechanism is fair; banding the rung
S while the mechanism is unknown and the deploy path is broken is not.
Either rung 1 keeps a working build-and-deploy path for the greenfield
binary, or rung 4 carries that work and is banded for it.

**[major] Rung 11 is a third M carrying an L's worth of security surface**
Spec:390-398 gives rung 11 the driver seam (scan trigger, scan status,
album deep link, capability reporting), the Navidrome reference driver
owning the bypass path set, the trusted header, externalized-auth
config names verified against the pinned Navidrome version, logout
wiring, provisioning choreography with a first-admin `doctor`
preflight, *and* the ingress reverse proxy built to Spec:302-310's
proxy-hygiene list: `Rewrite` with `SetXForwarded`, injection inside
`Rewrite`, session-cookie and hop-by-hop stripping, path normalization
rejecting raw/decoded mismatches and `..` before prefix routing,
per-route write-timeout handling for SSE and audio, `Cache-Control:
private` on injected responses, origin checking on the proxy path, and
refusal to inject on non-idempotent requests lacking a same-origin
signal. Every one of those is a security-critical behavior with its
own test surface, and three of them (the strip set, the trusted-source
pin, the ingress flip ordering) are blockers carried forward from the
login security review. This is the rung where a band error costs
custody, not schedule.

**[major] The no-move assertion — the launch bar's own existential check — runs once, at rung 15, after fourteen rungs of sunk work**
Spec:57-59 makes it launch-gate part four; Spec:421 schedules it
inside the cutover rung. The organizer spec states the stakes: "A
one-character difference from `config/beets-config.yaml`'s
query-prefixed path formats re-lays-out the entire library on the
first post-swap run." The code that computes those paths is rung 9's
layout and collection routing, an L rung six rungs earlier. If the
diff comes back non-empty at rung 15, the repair lands in rung 9's
code, at the end, on the irreversible rung, with the attended gate
already scheduled. The assertion is cheap and it does not need the
cutover: it is a pure function of config plus item metadata compared
against a directory listing, and the listing is available read-only
(the safety rails permit "read-only ssh only during reviews," and R2
backups hold a copy besides). The spec never names an earlier probe,
and Spec:462-463's "Rungs 1 through 3 and 5 through 13 touch the box
not at all" reads as forbidding one. Say plainly that a read-only
layout diff runs as soon as rung 9's layout code exists, and that a
non-empty diff there is a rung-9 defect rather than a rung-15
surprise.

**[major] Rung 3's tunable surface is the list the mechanics study refuted, and it regresses against the organizer spec**
Spec:340-342: "beets' full tunable surface (thresholds, distance
weights, preferred, ignored, required)." The mechanics study's Q4 is
titled "The tunable surface, verified complete" and opens: "The
believed surface ... is real but incomplete." The verified additions
it enumerates, each with a source citation: `track_length_grace` (10)
and `track_length_max` (30), which directly scale the track_length
penalty (`distance.py:364-373`, `396-399`); `ignore_data_tracks` and
`ignore_video_tracks`, which change the candidate track list and hence
the missing- and unmatched-track counts
(`beetsplug/musicbrainz.py:555-580`); `musicbrainz.search_limit`
(default 5), which changes which candidates exist and therefore the
single-candidate and `rec_gap` branches
(`beets/metadata_plugins.py:199-204`);
`musicbrainz.data_source_mismatch_penalty`;
`import.timid`, which gates the strong-ID early return
(`match.py:432-442`); `import.languages` and
`import.ignored_alias_types`; `musicbrainz.extra_tags`; the `plugins`
list itself; and the verdict-to-action trio `import.quiet`,
`quiet_fallback`, `none_rec_action` — which the study notes are part
of the surface "if 'verdict' means the final action, as it does for
dubplate's apply/skip semantics." Separately, the spec's list is
*shorter* than the organizer spec's, which named `ignored_media`
alongside the rest ("Conformance harness"); greenfield dropped it,
and `ignored_media` filters whole media out of a candidate's track
list (`beetsplug/musicbrainz.py:638-639`). Rung 3's deliverable list
underprices its own scope, and rung 2's config element of the
quadruple has to cover the same enlarged surface or the corpus cannot
express the cases.

**[major] The epsilon wording underprices the float regime the study established, and omits threshold-proximity flagging**
Spec:339-340 requires "numeric distance agreement within the epsilon
rung 2's plan states." The study's threat 3 says something stronger
and something weaker at once. Stronger: bit-identical distances are
attainable, because every accumulation runs in a fixed code-determined
order over IEEE-754 doubles with no transcendentals, so a port that
replicates operation order reproduces the same doubles on any IEEE
platform. Weaker: an epsilon is not a sufficient acceptance criterion
near a threshold, because `strong` is a strict `<` at
`beets/autotag/match.py:285` while `medium` is `<=` at `288`, and "an
epsilon cannot paper over a comparison-direction disagreement." The
study's pragmatic criterion is to compare at ~1e-12 relative, treat
the verdict as the conformance unit, **and flag any case whose best
distance falls within ~1e-6 of 0.04 or 0.25, or whose candidate gap
falls within ~1e-6 of 0.25**. The spec has the epsilon and not the
flag. Without the flag, a corpus can be green while carrying cases
that flip verdict on a compiler's constant-folding choice or a decimal
literal parsed one ULP apart — and the launch bar's tier one admits no
adjudication for exactly that failure.

**[major] Assignment-solver ties are unnamed, and they are the one place a correct Go port still diverges**
Spec:337-338 gives rung 3 "album-level matching, stable candidate
ordering" and nothing about track assignment. beets solves a full cost
matrix with `lap.lapjv` (`beets/autotag/match.py:12-13`, `64`,
`53-81`), which Go has no drop-in for. The study's threat 2 states the
hazard precisely: when several assignments share the minimal total
cost, "a different-but-equally-optimal choice by a Go solver yields
the same assignment cost but NOT necessarily the same final distance,
because the final per-track distances are recomputed with
`incl_artist=va` while the assignment cost ignores the track artist
(`match.py:64` vs `distance.py:529`). The two objectives diverge
exactly on VA albums." Its remedy is corpus work, not matcher work:
detect cases with non-unique optimal assignments (perturb and
re-solve, or check for duplicate row minima) and either exclude them
or pin the expected mapping. That makes it a rung-2 deliverable the
spec does not have, gating a rung-3 property the spec does not name.
Compilations and various-artists releases are core estate material,
so this is not a corner.

**[major] The spec claims one attended gate and schedules at least three streams of owner attention**
Spec:483-484: "The one attended gate of the initiative is rung 15's;
everything else that needs Geoff is batched questions at
checkpoints." Three things contradict it. Spec:152-154 has shadow
"divergences ... adjudicated by the owner into the same
known-acceptable-divergence artifact," arriving asynchronously from
rung 4 through rung 14 — owner judgment, unbatchable, on the critical
path of the launch gate, since an unadjudicated divergence fails
launch-bar part three by definition. The organizer's success bar says
the same of the corpus: divergences "are adjudicated case by case and
recorded as known-acceptable divergences with the reasoning written
down," which lands owner work inside rungs 2 and 3 as well. And
CLAUDE.md's reviewer rule states that "Uploader UI work additionally
gates on the visual-fidelity chain (Navidrome reference screenshots,
fresh-context `visual-verifier`, Geoff's before/after)" — rung 7
builds the uploader reference web client, so Geoff's before/after is a
second attended gate, nine rungs before the one the spec admits to.
Either the sentence is wrong or three real gates are unscheduled;
both readings cost the same to fix and only one of them is honest
about the attended-time budget.

**[major] The `__Host-` cookie mandate forecloses the split-host deployment the universality review required**
Spec:262-264: "**Sessions.** HMAC-signed `__Host-` cookie, `Secure`,
`HttpOnly`, `SameSite=Lax`." The `__Host-` prefix by specification
forbids a `Domain` attribute and pins `Path=/`, which is exactly the
single-host assumption. The universality review's [High] finding
"Single-hostname topology is assumed throughout, and the cookie is
never scoped" asked for the opposite, and the rulings doc carries it
forward in the load-bearing list: "cookie scoping for split-host
deployments." The greenfield spec's birth requirements adopt the
security review's `__Host-` recommendation and drop the universality
review's scoping requirement without noting that the first forecloses
the second. An adopter running dubplate at `music.example.org` and
their streamer at `stream.example.org` — the configuration
universality ruling four is meant to protect — cannot share a session
under an unqualified `__Host-` mandate. The resolution is a decision,
not a contradiction (`__Host-` as the default with a documented
same-host requirement, plus a scoped alternate under an explicit
split-host flag), but the spec has to make it rather than inherit both
requirements silently.

**[major] The ops seam ships half its organizer contract: no operation-availability mechanism and no second implementation for the contract suite to test**
The organizer spec's "Shared code: the ops seam" defines two capability
mechanisms that "never share a term" — streamer driver capabilities
and *ops operation availability*, which "reports which operations an
implementation offers and belongs to the ops seam" — and rules that
"Capability asymmetries are named in this spec or a successor spec,
never declared ad hoc in code." Greenfield rung 11 carries the
streamer half ("capability reporting", Spec:391). Rung 6 (Spec:356-362)
carries the vocabulary, the job model, cancellation, the store, and a
"contract-test suite skeleton," and never mentions operation
availability. So the mechanism whose whole purpose is preventing an
ad-hoc opt-out from parity is unassigned across all fifteen rungs.
Relatedly, Spec:415 says rung 14 runs "the full contract suite," but
the organizer defines that suite as running "against both
implementations, `ops/local` directly and `ops/client` against an
in-test server," and greenfield builds only the server side
(Spec:126-127). The tree at Spec:196 lists `internal/ops/` with no
`local` or `client` leaf. A contract suite with one implementation
enforces nothing; either say plainly that parity enforcement is
post-launch scaffolding, or name what rung 14 actually runs.

---

## Minor

**[minor] "Independent of the spine from day one" is false as written**
Spec:134-136. Rungs 2 through 4 all depend on rung 1: the threshold
registry and universal wall hold the matcher's tunable surface, and
the `cmd/dubplate-server` skeleton hosts the capture tool and harness
subcommands. The claim that survives is narrower and still useful:
the evidence rungs are independent of rungs 5 through 13. Worth
fixing because a plan author reading it literally may try to schedule
rung 2 first.

**[minor] Rung 3 builds a caller-priority scheduler whose callers arrive four to six rungs later**
Spec:342-343: "the shared MusicBrainz rate limiter with per-caller
budgets in which a user-present probe outranks a batch import." The
user-present probe is intake's synchronous match probe (rung 7) and
the batch import is the pipeline (rung 9). At rung 3 neither caller
exists, so the priority behavior can only be tested against
synthetic callers and its real starvation properties go unverified
until rung 9 — which is also where Spec:381-383 defers the
concurrency-model decision that interacts with it. Either state that
the rung ships the budget interface and the priority policy is
re-verified at rung 9, or move the policy there.

**[minor] The journey-vocabulary amendment moves to rung 9, ahead of the seam the organizer assigns it to**
Spec:384-385 lands "imported, no player configured" as journey
vocabulary in the pipeline rung. The organizer spec's "The server's
purview" assigns it to the streamer-seam work: "The streamer-seam work
owns filing that amendment, and it owns the config-wall change." The
greenfield relocation is defensible (rung 9 is where the terminal
state is produced) but it is a supersession of an organizer assignment
and it is not listed in "What this spec supersedes" (Spec:435-449),
which claims the organizer "otherwise governs." Add it, or move it to
rung 11.

**[minor] The capture tool's pseudo-release double fetch is a replay-breaking omission**
Spec:330-333 gives rung 2 "the capture tool" with no fidelity
requirements. The study's Q5 and threat 7: a release with status
`Pseudo-Release` and a backward `transl-tracklisting` relation
triggers a **second** `get_release` of the linked actual release, and
the two AlbumInfos merge (`beetsplug/musicbrainz.py:785-798`,
`_merge_pseudo_and_actual_album` at `203-240`); "a capture tool
recording only the requested MBID will break replay." Non-Latin-script
material is deliberately in the corpus (organizer, "Migrator
criteria"), which is exactly where pseudo-releases live. The study
also hands rung 2 its stub seam for free — `musicbrainz: {host:
"127.0.0.1:PORT", https: no, ratelimit: 100}`, no monkeypatching —
and a `RateLimitAdapter` floor of 0.25 s per request
(`beetsplug/_utils/requests.py:89`, `106-135`) that survives a high
configured ratelimit and must be budgeted into differential-run wall
time. All three belong in rung 2's deliverable list.

**[minor] The live oracle is coarser than "both verdicts are logged" implies**
Spec:151-152. The box imports under `beet import -q` with
`quiet_fallback: skip`; the study's "Estate plugin participation"
section records the mapping: "strong -> apply best candidate,
everything else -> skip (`session.py:337-342`)", and
`none_rec_action` is unreachable under `-q`. So the live oracle emits
a binary strong/not-strong signal, not a recommendation level and not
a distance. For every skipped album beets supplies no ground truth at
all, and launch-bar part three's "zero false confident matches against
live beets decisions" (Spec:54-56) can only be evaluated on the
applied subset. Say what the shadow can and cannot decide, and where
owner adjudication substitutes.

**[minor] The whole-build shadow converts clock time into a budget, against the workstation rule**
Spec:137-139 argues the shadow deploys early because "shadow evidence
accumulates in calendar time rather than executor time," and
Spec:157-160 keeps a 30-day floor as a launch-gate minimum. With one
executor and sequential rungs, calendar time between rung 4 and rung
14 *is* executor time. If the spine runs efficiently the launch blocks
on the floor; CLAUDE.md's conducting rule states "Clock time is not a
budget: prefer the serial, cheaper path." The floor is Geoff's to keep
or drop, but the spec should say which budget it is spending and what
happens if the spine finishes first.

---

## The three I would fix first

1. **Re-derive the pivot section from the tree that exists** (blockers
   1 and 2). `e2e/` gets a disposition, `tests/*.bats` comes off the
   list, and rung 1's gate rewiring is stated as what it actually is:
   remove three web legs, exclude `_salvage/` from
   `scripts/vale-comments.sh`'s git-driven file list, and keep or
   retire the only end-to-end coverage of the binary still running in
   production. This is cheap, and everything downstream is planned
   from it.
2. **Fix the shadow's material problem before anything is banded on
   it** (blocker 4). Decide whether rung 4 re-opens a contribution
   path, or restate the floor in terms of material that will exist, or
   drop the floor and let the corpus carry tier one alone. Until that
   is settled, the launch gate has an unreachable clause, the
   "strictly stronger" claim is unsupported, and the stated fallback
   describes an estate that cannot accept an album.
3. **Rebuild rung 15 as an L and give it a rollback target that
   outlives it** (blocker 3, major 2). Split the migrator and the
   box-facing rebuild out of the cutover, name what re-installs the
   pre-cutover estate, keep beets and the old units resident through a
   stated deprecation window, restore the organizer's ledger re-import
   half, and move `_salvage/` retirement to a pass after the window
   closes. Fold in the earlier read-only no-move diff (major 5) so the
   irreversible rung stops carrying a discovery risk it can no longer
   afford to fail.

**Verdict:** the architecture is sound and the rulings are faithfully
carried, but the ladder is planned against a repo state and an estate
state that both changed on 2026-09-02 — three rungs are mis-banded,
the cutover rung deletes its own rollback, the shadow has no material,
and the evidence track predates the mechanics study it now has to
absorb; fix the six blockers and this is ready to plan from.
