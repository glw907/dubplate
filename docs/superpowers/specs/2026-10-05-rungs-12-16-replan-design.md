# Rungs 12 and 16 re-plan: the built-in streamer at first launch

Status: approved in conversation 2026-10-05 (Geoff); adversarial review, then the written-spec read.
Supersedes, where they conflict: the build spec's rungs 12, 15 (shadow clause), 16, and 17, and launch bar
part 3 (`2026-09-02-greenfield-server-build.md`); the streamer spec's slice 9, shadow and cutover sections,
and hooks 5 and 6 (`2026-09-07-streamer-2-0-design.md`, rev 2.5); the version paragraph in `CLAUDE.md`.

## Why the ladder changes

The ladder assumed the group keeps listening through Navidrome until 0.8. Rung 12 put dubplate in front of
Navidrome as an SSO proxy with a driver seam, and rung 16 swapped the box under it with a rollback to the
incumbent. On 2026-10-05 Geoff retired the incumbent stack: Navidrome, the Cloudflare tunnel and Access, the
box's beets, and the shadow program's box side. The box now serves only dubplate's development, and the
matcher counts as validated, so the shadow floor no longer gates launch.

Every Navidrome-facing surface is now code built to be deleted. The 2026-10-05 architecture reads found
four such pieces in `internal/serve` with no production caller, and the shadow program observes a beets
that no longer runs anywhere. This spec removes that code and gives rungs 12 and 16 their new shape.

## Rulings (Geoff, 2026-10-05)

1. **First launch is 0.8.0, on the built-in streamer.** The group listens through dubplate's own
   OpenSubsonic streamer from the first day. No 0.x pre-release exists, and 0.9 and 1.0 keep their meaning.
2. **`dubplate import navidrome` stays in 0.8.** It serves a stranger switching from Navidrome. The group's
   first launch runs it once against the Navidrome database in the retirement archive
   (`/root/musicbox-retired-2026-10-05.tar.gz` on the box), so members keep their stars, playlists, and play
   counts.
3. **The field is "Username".** The roster's streamer username becomes `username` everywhere: column, JSON
   key, flag, Go identifiers, and copy. The copy tells a member it is the name they type into a music app.
4. **dubplate integrates with no external streamer.** The driver seam, hook 6, and the Navidrome and Jellyfin
   targets retire. ROADMAP's streamer-integration entry moves to `IDEAS.md` with this reasoning.
5. **Rung 12 is the built-in streamer, and its first two slices start now.** The streamer spec's slices 1
   (study) and 2 (harness) touch no production surface and open as a document-and-corpus lane beside 6d and
   rung 7, under the standing two-lane grant.
6. **The architecture stays clean, tidy, and beautiful through every change.** A removal reshapes what
   remains so it reads as if it had never had the removed part. Trimming a deletion down to the
   compiler's complaints falls short of this bar. Every touched package is graded by a `go-architecture-reader`
   at the pass's merge, as for any pass. The architecture also stays idiomatic: Go to the stdlib's own
   shape (`idiomatic-beautiful-go`), and `web/` to idiomatic Svelte 5 and DaisyUI.
7. **dubplate stays extremely easy to install and configure for a new operator.** This is a primary product
   goal (`CLAUDE.md`), and every change here is weighed against it:
   - No change adds a setup step, a required key, or a site-specific default outside `internal/config`.
   - Removing the driver seam and the Navidrome keys shrinks the configuration surface. A stranger never
     meets a streamer URL, an admin user, or a forward-auth front.
   - First launch on the built-in streamer means one binary serves the web app and the music apps. No second
     service needs installing, wiring, or keeping in sync.
   - `dubplate import navidrome` is optional. A new operator with no Navidrome never sees it.
   - Slice 8's `dubplate init` and autocert listener carry the install path. Rung 16 walks it as a stranger
     would, on the box, and records every step a stranger would find unclear as a defect against slice 8.

## The ladder after this spec

Rungs 6d, 7, 9, 10, 11, 13, and 14 are unchanged.

**Rung 12, Built-in streamer (L, multi-pass).** It runs as the streamer spec's slices 1 through 8, each a pass
with its own plan, ceiling, oracle diff, and roster walk:

1. Study
2. Harness
3. Catalog projection and browsing
4. Art and search
5. Streaming and transcoding
6. Identity, app passwords, and shares
7. Per-user state and group features
8. Hosting and setup

Slices 1 and 2 run now as a document-and-corpus lane. Slices 3 onward start after rung 9 lands the catalog.
The pinned Navidrome and gonic containers stay in the harness as workstation-only behavior oracles, never on
the box, under the streamer spec's GPL behavior-only rule. Slice 9 (shadow period, flip, and close-out) is
deleted, because there is nothing to shadow or flip.

**Rung 15, Integration.** Unchanged except that the shadow record leaves its evaluation. The conformance
gate is evaluated against the launch bar as before.

**Rung 16, First launch (0.8.0) (M).** A first install on the box, with no cutover and no incumbent to roll
back to. It carries the live items routed below, the one-time multi-disc re-layout (erratum 2026-09-24), the
no-move assertion re-run, the group's `dubplate import navidrome` run against the archive, and one attended
gate covering login, upload, and playback in a real music app. The box reaches the internet through slice 8's
TLS listener, since the tunnel is gone. Its rollback is the runbook's restore from backup.

**Rung 17, Launch close-out (S).** `_salvage/` and the interim remnants retire, and ROADMAP and the organizer
spec reconcile. The beets deprecation window is gone. The retirement archive's disposal is decided here,
after the import has read it.

**Launch bar.** Part 3, the clean shadow record, is struck. Parts 1 (completeness), 2 (the matcher bar over
the conformance corpus), and 4 (layout parity) still gate. Completeness now includes the streamer.

## The removal pass

One small code pass, `navidrome-removal`, runs on lane 1 before 6d. 6d's identity and serve work touches the
same files, so the removals land first and 6d's plan rebases onto them. The pass opens with the
`dependency-upgrade` sweep, as every pass does.

### `internal/serve`

The pass removes four pieces, each rebuilt only by the streamer slice that first consumes it:

- `FrontExemption`, `NewFrontExemption`, `NewStrip`'s exemption parameter, and the exemption threading
  through `collapseToLast`, `deleteIdentity`, and `scrubTrailerOnRead`. This removes the dual-stack
  raw-string peer match bug with it.
- `Shell.SetFallthrough` and the session-cookie wrapper around the fallthrough. `Handler` serves its own
  routes and answers everything else with `http.NotFound`.
- `NewStrip`'s driver `header` parameter and the `Strip.header` field.
- `originCheck.addBearerBypass` and the bypass mux. Slice 6 likely needs a `/rest` exemption from the origin
  check, and builds it then, keyed on `netip.Addr` where it matches peers.

Acceptance:

- After the removals, each remaining type, function, and parameter earns its place.
- A helper that existed only to thread a removed parameter is folded into its caller or deleted.
- Every package doc and every type and function doc describes the package as it now stands, with no
  sentence about a front, a driver, or a fallthrough.
- The rendered `go doc ./internal/serve` reads as a coherent package.
- Generic hardening stays and keeps its tests: the trusted-proxy config, the origin check, and group 1 and
  group 2 stripping.
- `doctor`'s web row and probe route behave as before, proved by their existing tests.

### The shadow program

The pass deletes `cmd/dubplate-server/shadow.go`, its test, the `internal/engine/shadow` package, the
`shadow` integration class, and its config keys. `internal/harness/analysis` and `internal/harness/cassette`
stay, because conformance uses them. A package left with no importer outside its tests is either deleted or
recorded with its remaining consumer. Comments that cite the shadow command as precedent
(`unidecode.go:35,77` and `serve.go:71,184`) are rewritten to stand alone. Git history and the
`pre-greenfield` tag keep the deleted code. Corpus and conformance remain the matcher's gate.

### Configuration

The pass retires these rows from `internal/config/keys.go`:

- the reserved `streamer` class's `DUBPLATE_NAVIDROME_URL`, `_ADMIN_USER`, and `_ADMIN_PASSWORD`
- the monitoring class's `DUBPLATE_PING_NAVIDROME_URL`
- the superseded rows `DUBPLATE_PUBLIC_NAVIDROME_URL` and `DUBPLATE_NAVIDROME_DB`

The streamer's own keys arrive with the slice that reads them. The stale Navidrome comments at
`cmd/dubplate-server/serve.go:173` and `store.go:21` are rewritten to state the reason without the retired
system.

### The username rename

`streamer_username` becomes `username`, and `StreamerUsername` becomes `Username`. The rename covers:

- `internal/auth` (`roster.go`, `admin.go`, and `seam.go`)
- `cmd/dubplate-server` (`member.go` and `doctor.go`)
- migration `0003_catalog.sql` and the store fixtures that snapshot it
- `web/src/lib/copy/deck.ts` and `web/src/routes/members/+page.svelte`

The column keeps `NOT NULL UNIQUE COLLATE NOCASE`. The migration is edited in place rather than amended,
because no store database exists on the box. The plan verifies that precondition (`/var/lib/dubplate/` holds
no store file) before the edit, and amends with a new migration if it fails.

The box's roster file holds one admin row under the old JSON key. The loader reads only the new key, with no
compatibility shim. The pass's deploy migrates that one row through an existing verb or the bootstrap path,
never by hand, and the plan names which. Copy follows the visual identity spec's register: the label is
"Username", and its hint says it is the name used in music apps.

### Close

The close follows the repo's pass close:

- `code-simplifier` over the changed Go, TypeScript, and Svelte
- one `go-architecture-reader` per touched Go package
- `svelte-reviewer` and `daisyui-a11y-reviewer` on the members page
- `diff-reviewer` per task

The pass class is set in its plan.

## Document changes

A document pass, which can run as a lane beside the removal pass, makes these edits:

1. **Build spec.** Add an erratum block at rungs 12, 15, 16, and 17 and at launch bar part 3, pointing here.
   The shadow section ("Shadow: the standing oracle") gets a retirement erratum. The rung text stays as
   written, per the repo's practice for superseded documents.
2. **Streamer spec, revision 3:**
   - Remove the shadow period, the flip, the `shareRole` doctor row, and the escape-hatch text.
   - Delete slice 9.
   - Rewrite slice 8's import rationale to the first-launch import.
   - Retire hook 6, and drop capability reporting unless a slice names a consumer.
   - Rename hook 5's field to `username`.
   - Mark the open item on shadow length and Navidrome retention as moot.
   - Record rulings 1 through 7.
3. **`CLAUDE.md`.** Rewrite the versions paragraph (first launch is 0.8.0, the built-in streamer), the
   build-for-0.8 paragraph (the hooks list is now rung 12's own slices), and the owner touchpoints (the
   rung 16 cutover sitting becomes the first-launch sitting).
4. **`ROADMAP.md`.** Fold `streamer-0-8` into the greenfield initiative as rung 12. Move the
   streamer-integration and forward-auth front entries to `IDEAS.md` with ruling 4's reasoning.
5. **STATUS and HISTORY.** STATUS drops the moot carries and the username owner question. HISTORY records the
   re-plan.

## Where the carried items land

| Item | Lands in |
|---|---|
| Enabling the serve unit | Rung 16 |
| Secrets in every `music`-uid verb's environment | Rung 16 |
| SMTP exercised from the box | Rung 16 |
| 6c's Ruling 2 (`deploy.sh` halts on `check`'s nonzero exit before restarting `serve`) | Rung 16 |
| 6b-ii-b's runbook lines | Rung 16 |
| UNI-H2 (`/` to login) | Slice 8, with the TLS listener |
| Group 2 names in trailers | Slice 8 |
| The streamer database's place in a restore | Slice 7, the first slice with per-user state |
| The streamer-admin mapping | Slice 6 |
| Restore-before-12 losing stars and playlists | Moot: no external streamer holds them |
| Stripping group 2 names but the configured header from a trusted peer | Moot: no front sends a header |
| `FrontExemption`'s dual-stack match | Closed by the removal pass |
| Security M2 (roster API sharing an origin with the fallthrough) | Moot: no fallthrough |
| A branding refusal stopping the fallthrough | Moot: no fallthrough |
| Navidrome username matching is case-insensitive | Moot. `username` keeps `COLLATE NOCASE` on its own merits |
| Ruling 7's availability coupling at the seam | Moot: no seam |
| 5c-iii's CSP-scope M2 and shared-origin M2 | Slice 6, re-read against the in-process streamer |

The trusted-proxy configuration stays as generic hardening for an operator who puts their own reverse proxy
in front of the box.

## Out of scope

- 6d's plan and its three rulings, which await approval separately. It rebases onto the removal pass.
- Rungs 7 onward, other than 12, 15, 16, and 17 as described above.
- Fork B (formats).
- The architecture-cleanup inventory, including `doctor`'s twelve-flag OR chain. It stays with the cleanup
  pass.
- The slice 1 and 2 plans, which are authored just in time in their own lane.
