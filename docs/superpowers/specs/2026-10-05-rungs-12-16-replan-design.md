# Rungs 12 and 16 re-plan: the built-in streamer at first launch

Status: revision 2, 2026-10-05. Revision 1 (`88c535e`) was approved in conversation (Geoff). Revision 2 folds the
four-lens adversarial review (fold record: `research/2026-10-05-rungs-12-16-replan-fold.md`). It awaits Geoff's
written-spec read and the three owner rulings below.

Supersedes, where they conflict:

- the build spec (`2026-09-02-greenfield-server-build.md`): rungs 12, 15 (shadow clause), 16, and 17; launch bar
  part 3; and every passage on Navidrome, a driver, SSO, the fallthrough, or the shadow (listed under "Document
  changes"). Ruling 1's adopter clause and the tree shape's "first-class leaf" sentence go only if Owner ruling 1 is
  yes.
- the streamer spec (`2026-09-07-streamer-2-0-design.md`, rev 2.5): slice 9, the shadow and cutover sections, the
  "Relation to the current ladder" rejections, hook 6, hook 15's capability clause, and hook 5's field name.
- the configuration homes spec (`2026-09-11-configuration-homes-design.md`): its rung 12 verifications, its
  shadow-pair erratum, and rung 14's private gate keyed on a driver capability report. Its bearer admin mount
  stands, because the bearer bypass stays.
- the group-scale principles spec (`2026-09-25-group-scale-principles-design.md`): "Before 0.8", row A13, and the
  delegated route row.
- the rung 6 spec (`2026-09-23-rung-6-ops-seam-store.md`): its append-only migration rule, once, for `0003` only.
- the version paragraph in `CLAUDE.md`.

## Why the ladder changes

The ladder assumed the group keeps listening through Navidrome until 0.8. Rung 12 put dubplate in front of
Navidrome as an SSO proxy with a driver seam, and rung 16 swapped the box under it with a rollback to the
incumbent. On 2026-10-05 Geoff retired the incumbent stack: Navidrome, the Cloudflare tunnel and Access, the
box's beets, and the shadow program's box side. The box now serves only dubplate's development, and the
matcher counts as validated, so the shadow floor no longer gates launch.

Every Navidrome-facing surface is now code built to be deleted. The 2026-10-05 architecture reads found
three such pieces in `internal/serve` with no production caller, and the shadow program observes a beets
that no longer runs anywhere. This spec removes that code and gives rungs 12 and 16 their new shape.

## Rulings (Geoff, 2026-10-05)

1. **First launch is 0.8.0, on the built-in streamer.** The group listens through dubplate's own
   OpenSubsonic streamer from the first day. No 0.x pre-release exists, and 0.9 and 1.0 keep their meaning.
2. **`dubplate import navidrome` stays in 0.8.** It serves a stranger switching from Navidrome. The group's
   first launch runs it once against the Navidrome database in the retirement archive
   (`/root/musicbox-retired-2026-10-05.tar.gz` on the box), so members keep their stars, playlists, and play
   counts.
3. **The field is "Username".** The roster's streamer username becomes `username` everywhere: column, JSON
   key, flag, Go identifiers, wire codes, and copy. The copy tells a member it is the name they type into a music app.
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
   - Removing the driver seam and the Navidrome keys removes keys a stranger could mistake for setup. (Every
     retired row was already refused when set, so the required set itself does not move.)
   - First launch on the built-in streamer means one binary serves the web app and the music apps. No second
     service needs installing, wiring, or keeping in sync.
   - `dubplate import navidrome` is optional. A new operator with no Navidrome never sees it.
   - Slice 8's `dubplate init` and autocert listener carry the install path. Rung 16 walks it as a stranger
     would, on the box, and records every step a stranger would find unclear as a defect against slice 8.
     (Where the walk runs and whether its defects block: Owner ruling 3.)

## Owner rulings

Each is a yes/no question. The spec builds the recommended answer unless Geoff rules otherwise.

1. **Delete the shadow program, superseding the build spec's ruling 1 adopter clause ("the local confidence tool an
   adopter runs against their own live beets") and the tree shape's "first-class leaf" sentence?** Recommend
   **yes**: 0.8's switcher path is `dubplate import navidrome`, few strangers run a live beets import timer, and the
   package is the largest cross-package surface ruling 6 asks this pass to tidy.
   - Yes builds "The shadow program" section below, plus an `IDEAS.md` entry, "a beets-migration confidence
     check", that names the `pre-greenfield` tag and git history as its source.
   - No keeps `internal/engine/shadow`, the verb, and the `shadow` class as an adopter tool, and removes only estate
     values. "The shadow program" section shrinks to rewriting estate-specific comments. `tagio`, `ClassBackground`,
     and `ErrUnavailable` keep their caller.
2. **Retire the forward-auth front along with the external streamers, moving its ROADMAP entry to `IDEAS.md`?**
   Ruling 4 covers streamers, and a front is an identity source. Recommend **yes**: music apps authenticate with app
   passwords that a front cannot gate, so a front would guard only the web app, and its promise lives in three places.
   - Yes builds the `IDEAS.md` entry. The `DUBPLATE_ACCESS_*` supersession messages become "retired". The
     `auth.Seam` doc stops naming an alternate. The runbook's forward-auth topology section is deleted. The
     identity-source selector's default and the seam interface's single implementation join the
     architecture-cleanup inventory.
   - No keeps the ROADMAP entry, the `seam.go` doc, and the `ACCESS_*` messages. The runbook topology section is
     rewritten for a front with no streamer behind it.
3. **Does slice 8 close on a walk of the install path on a fresh throwaway host, with blocking defects?** That
   walk would replace ruling 7's walk on the box. The box is not a stranger's host: `deploy.sh` renders its
   configuration from the age store with estate literals, so an `init` walk there either fights that render or
   skips `init`. Recommend **yes**.
   - Yes builds a slice 8 acceptance line. One short-lived VPS, with no age store and no `deploy.sh`, goes from the
     release artifact through `init` to a played track. Each unclear step either blocks slice 8's close or goes to
     `IDEAS.md` by owner ruling. Rung 16 installs the group's estate through `deploy.sh` as today. Cost: one
     throwaway VPS for one walk.
   - No builds ruling 7's text. Rung 16 walks `init` on the box before `deploy.sh` renders `server.env`, and
     defects are filed against slice 8 without blocking launch.

## The ladder after this spec

Rungs 6d, 7, 9, 10, 11, and 13 are unchanged. Rung 14 is unchanged except its private gate (below).

Every other rung plan still weighs its surface against the streamer spec's forward-hooks list. Hooks 1, 12, 17, and
21 name rung 7; hooks 2, 3, 11, and 13 name rung 9; hook 16 names rung 13; hooks 15 and 18 name rung 14.

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

The streamer's configuration has these properties, which hold for every slice:

- The streamer is always on and declares no enable flag and no integration class.
- Its tunables (cache cap, concurrent encodes, the profile set) are registered thresholds with defaults.
- Any streamer secret, such as the app-password key, follows the auth secrets' pattern (`DUBPLATE_SESSION_SECRET`
  and its `_PREVIOUS` rotation row) and is generated by slice 8's `init`, so an operator never types one.
- Each slice plan states its delta to the set of values a default install's operator supplies by hand. The removal
  pass's Task 0 records that set's count at its base commit. A nonzero delta needs an owner ruling.

The per-slice roster walk is conductor-run on the workstation (desktop clients and the Android emulator). It is
never an owner touchpoint. A client the conductor cannot drive is walked at rung 16's attended gate.

Until slice 8, the box opens no public port. Attended web checks, rung 7's uploader before/after among them, run on
the workstation or through `ssh -L` to the box.

**Rung 14.** Private stays unsettable until slice 3's projector enforces visibility on every endpoint. That
replaces the gate on a driver capability report, and private collections ship at first launch, per the
configuration homes spec's existing 2.0 ruling.

**Rung 15, Integration.** Unchanged except that the shadow record leaves its evaluation. The conformance
gate is evaluated against the launch bar as before.

**Rung 16, First launch (0.8.0) (M).** A first install on the box, with no cutover and no incumbent to roll
back to. Every build-spec rung 16 item that the carried-items table below does not mark moot still stands. The box
reaches the internet through slice 8's TLS listener, since the tunnel is gone. Rung 16 runs in this order:

1. Provision the backup crypt remote and its secrets, enable backup, run one backup, and prove a restore into a
   scratch root.
2. Take the deploy `.prev` set.
3. Take a store snapshot.
4. Run the group's `dubplate import navidrome` (dry run first) against the verified off-box copy of the
   archive's database. The import precedes the re-layout, because its audio-key fallback reads files at
   Navidrome's recorded paths.
5. Run the one-time multi-disc re-layout (erratum 2026-09-24) and the per-collection relocation.
6. Re-run the no-move assertion.
7. Run one attended gate: login, upload, and playback of a named album fixture in a named client from slice 1's
   roster.

Rollback has two forms. Before the first backup, it is "disable `serve`; data stays". After it, rollback is the
runbook's restore.

**Rung 17, Launch close-out (S).** `_salvage/` and the interim remnants retire, and ROADMAP and the organizer
spec reconcile. The beets deprecation window is gone. The retirement archive's disposal is decided here, after
the import has read it, and covers every copy of the archive and the shadow directories inside it.

**Launch bar.** Part 3, the clean shadow record, is struck. Parts 1 (completeness), 2 (the matcher bar over
the conformance corpus), and 4 (layout parity) still gate. Completeness now includes the streamer.

## Preserving the archive (now)

Ruling 2's only input is one unverified file on the box, and the import runs at rung 16. A conductor chore,
read-only on the box, runs before any other box work:

1. List the tar and confirm that `navidrome.db` is present, with its `-wal` and `-shm` files if any.
2. Record the archive's sha256.
3. Copy the archive off the box and age-encrypt it on the workstation beside the secrets store. Rung 16's backup
   remote takes a second copy once it exists.
4. Extract the database and run `PRAGMA integrity_check` on it.
5. Record whether the old stack's R2 `navidrome.db` snapshot still exists.
6. Record the copy's location, the sha256, and the check result in STATUS.

The workstation is the only encrypted store that exists before rung 16, and the file holds members' addresses and
listening history.

## The removal pass

One small code pass, `navidrome-removal`, runs on lane 1 before 6d. The pass opens with the
`dependency-upgrade` sweep, as every pass does. It does not deploy. The serve unit stays disabled and no store
exists, so nothing on the box reads its output before the next deploy that needs it.

6d's identity and serve work touches the same files, so the removals land first. 6d moves to revision 2.3 after
the removal pass merges, and Geoff's approval gate is on 2.3. Revision 2.3 carries four things:

- the rename sweep across every renamed symbol;
- a pointer refresh for the `main.go` registration block;
- the roster step described under "The username rename";
- a re-justification, or a change, of `Principal` staying the username, now that no proxy header or Navidrome
  auto-creation gives the reason.

### `internal/serve`

The pass deletes three pieces for good, because nothing will consume them:

- `FrontExemption` and `NewFrontExemption`. `collapseToLast` exists only for the exemption and goes with it.
  The exemption threading also goes, through `deleteIdentity`, `scrubTrailerOnRead`, `stripConnectionListed`,
  and `trailerScrubbingBody.exempt`. This removes the dual-stack raw-string peer match bug.
- `Shell.SetFallthrough` and the session-cookie removal at the fallthrough: `removeCookie`, `withoutCookie`,
  `Strip.sessionCookie`, and `NewStrip`'s `sessionCookie` parameter. `Handler` passes `http.NotFoundHandler()`
  as `Prefix.Route`'s `next`. `Prefix.Route` stays, and its docs state the owned-versus-unowned decision on its own
  terms. A later root route (UNI-H2, `/rest` if slice 3 puts it at the root) arrives as a named route the shell
  owns, never as a reopened general-purpose slot.
- `NewStrip`'s driver `header` parameter and the `Strip.header` field.

After the removals, `NewStrip(proxy TrustedProxyConfig) Strip` takes one argument.

`originCheck.addBearerBypass` stays, as generic hardening, like the trusted-proxy configuration. It has named
consumers: the configuration homes spec's bearer admin mount at rung 14, and the OpenSubsonic API's non-GET
requests, which arrive at slice 5 at the earliest. Its doc names those consumers and keeps its two contract
sentences. A bypassed handler must not read the session cookie, and the pattern registers twice because either
registration alone leaves one form refusing. It is path-keyed.

`statusCapture`'s `ReadFrom`, `Flush`, `Unwrap`, `Hijack`, and group 3's `Upgrade` carve-out stay, with their
rationale restated for in-process streaming: `http.ServeContent` reaches `io.Copy`, which uses `io.ReaderFrom`.

Acceptance:

- `go doc -all ./internal/serve | grep -inE 'front|driver|fallthrough|exemption'` prints nothing.
- The same holds for the prose in `cmd/dubplate-server/serve*.go`, `internal/auth`, and `internal/config`. Rung and
  history citations in tests are exempt. `TestNewServeShell_StripCarriesNoFrontExemption` is renamed to state its
  surviving property, that a trusted proxy is still identity-stripped.
- The `servePrefixThreshold` comment makes no claim about a sibling server.
- The staged exported-unused leg reports no hit in `internal/serve`. The package's `go-architecture-reader`
  verdict carries no unused-surface finding.
- Every test that used the fallthrough as a probe keeps its assertion. Each is re-pointed at a registered spy
  subtree with a "spy reached" check, or at a status a 404 cannot also produce. The full list sits in
  `internal/serve/serve_test.go` and `cmd/dubplate-server/serve_test.go`; mechanics review M4 names each one.
- A test is deleted only when its sole subject was the fallthrough slot or the cookie removal. The plan lists each
  test with its disposition.
- `TestEmptyFallthroughAnswers404` is re-scoped to "a path outside the prefix answers 404". The outward behavior
  is unchanged, since production never set a fallthrough, so no test can fail against today's wiring here.
- The generic hardening keeps its tests: the trusted-proxy config, the origin check including the bearer bypass,
  and group 1 and group 2 stripping.
- The doctor web-row and probe-route tests are named in the plan, and the diff touches them only to rename.

### The shadow program (if Owner ruling 1 is yes)

The pass deletes `cmd/dubplate-server/shadow.go`, its test, the `internal/engine/shadow` package, the `shadow`
integration class, its config keys, and doctor's `registerShadowIntegration` test fixture. It disposes of the
orphans this way:

- `setUnidecodeTransliterator` is deleted. Rung 9's scoring verb rebuilds it. `unidecode install` and doctor's
  `unidecode` row stay, with rung 9 recorded as their consumer.
- `internal/engine/tagio` keeps its package, because its golden generator is a gate leg, with rung 9 recorded as its
  consumer. `Tags`, whose first caller was the shadow reader, becomes unexported or is removed.
- `musicbrainz.ClassBackground` and `ErrUnavailable`'s shadow justification stay only if rung 9's spec names a
  consumer. Otherwise they are pruned.
- `internal/harness/analysis` drops its shadow-row subject shape (the release-MBID subject, its required
  matcher-commit and config-hash fields, and the validator's shadow branch). The package becomes case-only,
  because `corpus/analysis/adjudications.json` holds no such row.
- `matcherCommit` stays as build identity. Its rationale in `main.go` and the `Makefile` is rewritten.
- `readFLAC` (`shadow/libraryread.go`) is recorded in HISTORY as salvage material for rung 9's plan.

`internal/harness/analysis` and `internal/harness/cassette` stay, because conformance uses them. Corpus and
conformance remain the matcher's gate.

Acceptance:

- `git grep -nIi shadow -- cmd internal scripts Makefile` returns only CSS `box-shadow` and the "silently shadow"
  idiom.
- `go list -deps ./cmd/dubplate-server | grep -E 'internal/(harness|engine/shadow)'` is empty.

### Configuration

The pass deletes the rows that never shipped, the reserved `streamer` class (`DUBPLATE_NAVIDROME_URL`,
`_ADMIN_USER`, `_ADMIN_PASSWORD`) and the monitoring class's `DUBPLATE_PING_NAVIDROME_URL`. It restates the
superseded rows `DUBPLATE_PUBLIC_NAVIDROME_URL` and `DUBPLATE_NAVIDROME_DB` as "retired: dubplate integrates with
no external streamer", after the `DUBPLATE_TUNNEL_TOKEN` precedent. The `DUBPLATE_ACCESS_*` messages follow Owner
ruling 2. The streamer's own keys arrive with the slice that reads them, under the properties above. The stale
Navidrome comments at `cmd/dubplate-server/serve.go` (prefix threshold) and `cmd/dubplate-server/store.go` (DSN)
are rewritten to state the reason without the retired system.

### The username rename

`streamer_username` becomes `username`, and `StreamerUsername` becomes `Username`. The rename also covers these:

- the flag `--username`;
- the wire codes `duplicate_username`, `invalid_username`, and `missing_username`, with the `duplicate_principal`
  copy dropping "streamer";
- `member list`'s column head;
- the rendered gate's ids in `web/scripts/rendered/run.ts`;
- `config/roster.example.json`;
- the runbook's bootstrap commands.

The roster stays at `schema_version` 1, with no version gate, because no released v1 file exists outside the box.

Acceptance:

- `git grep -nIiE 'streamer[_ -]?username' -- cmd internal web/src web/scripts config docs/server-runbook.md`
  returns nothing.
- The label is "Username", and its hint says it is the name used in music apps. The copy follows the visual
  identity spec's register. If Microsoft's A-Z entry spells the term "user name", the deck records the departure
  with ruling 3 and the music apps' own label as provenance.

The column keeps `NOT NULL UNIQUE COLLATE NOCASE`. Migration `0003_catalog.sql` and its fixtures (`0003` to
`0005`) are edited in place rather than amended, under a one-time erratum to the rung 6 spec's append-only rule.
Two preconditions make the edit safe. Doctor's `store` row on the box reads `absent`, with read-only ssh at plan
time. Backup was never enabled, so the remote holds no store snapshot. If either fails, the pass amends with a new
migration. The same commit re-pins `migrationHashes`' `0003` entry from an independent `sha256sum` and cites the
precondition evidence. Any old store fails loudly at open on the ledger's sha256 check, so none can be carried
forward silently.

The box's roster file holds one admin row under the old key. The loader rejects unknown members
(`json.RejectUnknownMembers(true)`), every verb loads the file first, and the bootstrap path needs the file absent.
No existing verb can therefore read the old row, and no compatibility shim is written. The first deploy that reads
the roster (6d's `member import`, or rung 16 if 6d has not deployed) carries one named runbook step:

1. Move `roster.json` aside to `roster.json.pre-username` inside the leaf (0600, `music`-owned).
2. Run `member add <address> --admin --name ... --username ...` through `systemd-run`, which takes the bootstrap
   path.
3. Check that doctor's `auth store` row reads ok and that `member list` shows the one admin.

### Close

The close follows the repo's pass close:

- `code-simplifier` over the changed Go, TypeScript, and Svelte
- one `go-architecture-reader` per touched Go package
- `web-auth-security-reviewer` over the `internal/serve` and `cmd/dubplate-server/serve*.go` diff, which asks for
  correctness gaps ranked by consequence, vacuous re-pointed tests first
- `svelte-reviewer` and `daisyui-a11y-reviewer` on one shared brief scoped to the rename diff
- `diff-reviewer` per task

The pass class is set in its plan.

## Document changes

A document pass, which can run as a lane beside the removal pass, makes these edits. The ledgers and
`CLAUDE.md` land after the removal pass merges.

1. **Build spec.**
   - One header-level erratum states that every passage on Navidrome, a driver, SSO, the forward-auth front, the
     fallthrough, or the shadow is superseded here. It lists these by section: "Drivers own driver knowledge", the
     serve-as-ingress and forward-auth trust boundaries, "One owned prefix", the post-login destination, proxy
     hygiene, the tree shape's `internal/streamer/` and shadow leaf, rung 4 "continuous to launch", the two-tracks
     text, the interim-estate matcher fallback, "Box safety", and the 2026-09-25 `/share/*` erratum.
   - Erratum blocks land at rungs 12, 15, 16, and 17, at launch bar part 3, and at "Shadow: the standing oracle".
   - The rung text stays as written, per the repo's practice.
2. **Streamer spec, revision 3:**
   - Remove the shadow period, the flip, the `shareRole` doctor row, the escape-hatch text, and the arm64
     walk "before the flip".
   - Delete slice 9.
   - Rewrite "Relation to the current ladder" and the header's "sequenced after the greenfield ladder's launch".
     The rejected alternatives become superseded, with ruling 1's reason.
   - Rewrite "Cutover and operations" to the configuration properties above, with no driver, no cloudflared, and
     no integration class.
   - Rewrite slice 8's import rationale to the first-launch import, and add these slice 8 criteria:
     - a dry-run report before any write;
     - a member is matched by address, then by username under `COLLATE NOCASE`, and a disagreement refuses with a
       report;
     - per-user writes land in one transaction, or are keyed so that a rerun is a no-op;
     - the import reads an extracted copy.
   - Slice 8 also gains the unit's `CAP_NET_BIND_SERVICE` (TLS mode only), the certificate leaf (unit
     `ReadWritePaths`, the `deploy.sh` pre-create, and the state-dir agreement leg), and the walk if Owner ruling 3
     is yes.
   - Retire hook 6 and hook 15's capability clause. Drop capability reporting unless a slice names a consumer.
   - Re-home hook 19 (the listener as a configured choice) to slice 8.
   - Rename hook 5's field to `username`.
   - Slice 6 records the bearer bypass's two contract sentences.
   - Mark the open items on shadow length, Navidrome retention, and "whether the study rung runs beside the
     current ladder" as answered.
   - Record rulings 1 through 7 and the owner rulings.
3. **Configuration homes spec.** One erratum. The rung 12 Navidrome verifications are moot, and so is the
   shadow-pair erratum. Rung 14's private gate is re-keyed as above.
4. **Group-scale principles spec.** One erratum. "Before 0.8" and row A13 are moot. The delegated route row has no
   subject until slice 3. "Versions this spec speaks to" follows ruling 1.
5. **Rung 6 spec.** The one-time `0003` erratum above.
6. **`CLAUDE.md`.**
   - Rewrite the versions paragraph: first launch is 0.8.0, on the built-in streamer.
   - Rewrite the build-for-0.8 paragraph: rung 12 is the streamer's slices, and every other rung plan still weighs
     the forward-hooks list per task. Hook 6 retires and hook 5 is renamed.
   - The rung 16 cutover sitting becomes the first-launch sitting.
7. **`ROADMAP.md`.**
   - Fold `streamer-0-8` into the greenfield initiative as rung 12.
   - Rewrite the greenfield paragraph (0.x, shadow, box swap) and the `packaging` entry's Navidrome ingredient.
   - Move the streamer-integration entry to `IDEAS.md` with ruling 4's reasoning. The forward-auth front entry
     follows Owner ruling 2.
8. **`docs/server-runbook.md`.** Update the `--username` flag in both bootstrap commands, add the roster step
   above, and delete or rewrite the topology section per Owner ruling 2.
9. **STATUS and HISTORY.** STATUS drops the moot carries and the username owner question, and records the
   archive chore's result. HISTORY records the re-plan and `readFLAC`'s salvage pointer. The conductor points
   the memory files `version-numbering.md` and `build-for-2-0.md` at the new `CLAUDE.md` paragraphs.

## Where the carried items land

| Item | Lands in |
|---|---|
| Enabling the serve unit | Rung 16 |
| Secrets in every `music`-uid verb's environment | Rung 16 |
| SMTP exercised from the box | Rung 16 |
| 6c's Ruling 2 (`deploy.sh` halts on `check`'s nonzero exit before restarting `serve`) | Rung 16 |
| 6b-ii-b's runbook lines | Rung 16 |
| Backup schedule key, `RCLONE_CONFIG_*` rows, crypt remote provisioned and restore proven | Rung 16, step 1 |
| Founding library's single-link migration and the restore-size runbook line | Rung 16 |
| Per-collection relocation under allocated roots | Rung 16, now unconditional (the import matches by MBID and audio key) |
| Store identity (6d Ruling 3) | Rung 16, if 6d's Ruling 3 is yes |
| Firewall 80 and 443, and the hostname's DNS record by API | Rung 16 |
| `CAP_NET_BIND_SERVICE` and the certificate leaf | Slice 8 |
| Command naming (`dubplate` versus `dubplate-server` for `init` and `import`) | Slice 8 |
| UNI-H2 (`/` to login) | Slice 3, with the root-route set |
| API base URL (`/rest` at the root or under the prefix) and the owned-prefix rule | Slice 1 measures clients against a sub-path base URL; slice 3 decides, with a build-spec erratum |
| The `;`-segment and case-variant prefix gaps, and the percent-encoding ban revisit | Slice 3 |
| Group 2 names in trailers | Slice 8 |
| The streamer database's place in a restore | Slice 7, the first slice with per-user state |
| The streamer-admin mapping | Slice 6 |
| Re-validating existing rows' usernames, and 5c-i security m1 (a username equal to another member's address) | 6d (its import requires the field), then slice 6 for principal uniqueness |
| Offboarding's streamer-deprovision step | Slice 6: the epoch revokes app passwords |
| Navidrome-to-member username mapping on import | Slice 8 (the criteria above) |
| 5c-iii's CSP-scope M2 and shared-origin M2 | Slice 6, re-read against the in-process streamer |
| Restore-before-12 losing stars and playlists | Moot: no external streamer holds them |
| Stripping group 2 names but the configured header from a trusted peer | Moot: the removal pass deletes the configured header, so group 2 is stripped from every peer |
| `FrontExemption`'s dual-stack match | Closed by the removal pass |
| Security M2 (roster API sharing an origin with the fallthrough) | Moot: no fallthrough |
| A branding refusal stopping the fallthrough | Moot: no fallthrough |
| Ruling 7's availability coupling at the seam | Moot: no seam |
| The header step's reconcile with Navidrome's header set (5c-ii) | Moot |
| R-5b26 / R33 (driver-declared post-login destination validator) | Moot: no driver |
| Ingress-before-ExtAuth ordering, `UIWelcomeMessage`, compose convergence | Moot |
| Beets database archived to R2 | Moot: the box's beets is retired, and its data sits in the archive (rung 17) |

The trusted-proxy configuration stays as generic hardening for an operator who puts their own reverse proxy
in front of the box.

## Out of scope

- 6d's plan and its three rulings. They await approval at revision 2.3.
- Rungs 7 onward, other than 12, 14's private gate, 15, 16, and 17 as described above.
- Fork B (formats).
- The architecture-cleanup inventory, including `doctor`'s twelve-flag OR chain. It stays with the cleanup
  pass.
- The slice 1 and 2 plans, which are authored just in time in their own lane.
