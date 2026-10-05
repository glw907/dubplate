# Rungs 12 and 16 re-plan: consistency review

Target: `docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md` at `88c535e` ("the spec").
Lens: consistency against the ratified documents, the 6d plan, the ledgers, and the code the spec cites.
Ask: correctness gaps ranked by consequence, not polish. Findings whose fix costs more than the risk
it removes may be refused at the fold.

Counts: 3 high, 12 medium, 6 low. Two findings are OWNER FORKs (K2, K14).

## Citations spot-checked

Every citation the spec makes says what the spec claims:

- `cmd/dubplate-server/serve.go:71` and `:184` cite `registerShadowThresholds` in `shadow.go`.
- `serve.go:173` names Navidrome's surface.
- `cmd/dubplate-server/store.go:21` cites Navidrome's DSN. The bare `store.go` is ambiguous, because
  `internal/auth/store.go` and `internal/engine/shadow/store.go` also exist.
- `cmd/dubplate-server/unidecode.go:35` and `:77` cite the shadow leaf and `shadow.OpenStore`.
- `internal/config/keys.go:87` holds `DUBPLATE_PING_NAVIDROME_URL`.
- `keys.go:98-103` holds the `streamer` reserved class with the three Navidrome names.
- `keys.go:137` and `:145` hold the two supersession rows.
- The slice numbering 1 to 9 matches the streamer spec's ladder (`streamer-2-0-design.md:651-679`).
- Hooks 5 and 6 are the stable identity and the driver seam (`:701-708`).
- `internal/harness/analysis` and `internal/harness/cassette` have conform and oracle importers, so
  they stay, as the spec says.

The defects lie in what the spec omits. None lies in what it cites.

## High

### K1. HIGH. Rung 16's new scope silently drops ratified rung 16 items

**Where.** Spec `:75-79` (rung 16) and `:191-211` (the carried-items table).

**Defect.** The spec defines rung 16 as a closed list: "the live items routed below", the multi-disc
re-layout, the no-move re-run, the import, and the attended gate. The build spec's rung 16 text stays
as written behind an erratum (spec `:172-174`). A just-in-time rung 16 planner therefore sees two
scopes and no rule for the items that are not in conflict. These ratified rung 16 items are absent
from both the rung 16 paragraph and the table:

- **The backup schedule key and the `RCLONE_CONFIG_*` template rows.** Source: the build spec's
  2026-09-23 erratum (5), `greenfield-server-build.md:63-66`. Scheduled backup is the only rollback
  the spec now leaves (spec `:79`: "Its rollback is the runbook's restore from backup").
- **The founding library's single-link migration and the restore-size runbook line.** Source:
  `configuration-homes-design.md:427-429`, `:648-650`, and `:710-713`. The principles spec's channel
  table also relies on "rung 16's tooling" (`group-scale-principles-design.md:328`).
- **The per-collection relocation under allocated roots.** Source: the launch bar part 4 erratum,
  `greenfield-server-build.md:166-178`, part of a bar the spec keeps (`:85-86`). That erratum is
  "conditional on rung 12 verifying that the pinned Navidrome build keeps per-user stars and
  playlists across a move". The condition is now moot and nothing says so. The relocation is
  therefore neither unconditional nor routed back to rung 10.
- **6d's Ruling 3: store identity routed to rung 16, "where backups first run on a timer".** Source:
  `plans/2026-09-26-rung-6d-identity-serve-store.md:114-128`. This item is conditional on 6d's
  approval.
- **"The beets database archived to R2".** Source: `greenfield-server-build.md:947`. See K13.

Items the spec should mark moot but does not name: the ingress-before-ExtAuth ordering, the
`UIWelcomeMessage` breach (a Navidrome setting), and "compose" convergence (all
`greenfield-server-build.md:948-952`).

**Proposed fold.** Add rows to the carried-items table:

- the backup schedule key and `RCLONE_CONFIG_*` rows, to rung 16;
- the single-link migration and the restore-size line, to rung 16;
- the collection relocation, to rung 16, now unconditional because the import matches by MBID and
  audio key, never by path (`streamer-2-0-design.md:550-551`);
- store identity, to rung 16 if 6d's Ruling 3 is yes;
- the beets database, per K13.

Add moot rows for the three Navidrome-era steps. State in the rung 16 paragraph that every build-spec
rung 16 item not marked moot here still stands.

### K2. HIGH, OWNER FORK. Deleting the shadow program contradicts a ratified owner ruling

**Where.** Spec `:118-125` and `:16-18`.

**Defect.** The spec justifies deletion with "the shadow program observes a beets that no longer runs
anywhere". That is true of the estate. Ratified text says the shadow serves adopters:

- **Build spec ruling 1 (Geoff, 2026-09-02, `greenfield-server-build.md:976-988`).** "Shadow is the
  local confidence tool an adopter runs against their own live beets before trusting the swap."
- **The tree-shape section (`:459-461`).** "The shadow runner is a first-class leaf because it
  outlives its rung: it is the migration-shadowing capability any beets-migrating adopter wants, not
  scaffolding."

The spec's supersedes list (`:4-6`) names neither passage, and rulings 1 to 7 do not rule on the
shadow. STATUS listed "whether the shadow code is deleted" as still open (`docs/STATUS.md:32`).

**Options.**

- (a) **Delete.** Record a ruling that supersedes ruling 1's adopter clause and `:459-461`. The
  stated reason: 0.8's switcher path is `dubplate import navidrome`, and beets migration is the
  separate one-time translation tool (`:780-789`). File "a beets-migration confidence check" to
  `IDEAS.md`, so git history and the `pre-greenfield` tag are named as its source.
- (b) **Keep.** Keep `internal/engine/shadow` and the verb as an adopter tool. Remove only the
  estate coupling.
- (c) **Park.** Delete now and leave the adopter question to `packaging`.

**Recommendation.** (a). The tool observes a running beets import timer, which few strangers have. It
costs a cross-package surface that ruling 6 asks the pass to tidy. The ruling must still be written
down, because it reverses a ratified one.

### K3. HIGH. Retiring hook 6 breaks the configuration homes spec and hook 15, and rung 14 is "unchanged"

**Where.** Spec `:179` ("Retire hook 6, and drop capability reporting unless a slice names a
consumer") and `:53` ("Rungs ... 14 are unchanged").

**Defect.** A ratified consumer already exists:

- Rung 14's API refuses to set a collection private "for as long as the active driver does not
  declare per-user folder access in its capability report" (`configuration-homes-design.md:220-224`).
- Hook 15 repeats this: "Hook 6's capability report also declares ... per-user folder access; private
  is unsettable without it" (`streamer-2-0-design.md:743-745`).

With hook 6 gone and rung 14 unchanged, rung 14's private gate references a mechanism that will never
exist. The configuration homes spec also routes these items to rung 12 (`:352-356`, `:364`, `:550`,
`:702-716`), and all of them are now moot:

- the Navidrome verifications: art filenames, read tags, folder browsing, the per-user library API,
  the linked-duplicate persistent-id reproduction, and stars across a file move;
- the 2026-09-23 erratum on the shadow pair (`:5`, `:78-79`).

The configuration homes spec is not in the supersedes list or the document pass. Because first launch
is now 0.8, the existing ruling that private is enforced by the built-in projector at 2.0
(`configuration-homes-design.md:221-222`) puts private collections in the first launch. That is
consistent with the ruling, but the spec does not say so.

**Proposed fold.**

- Add the configuration homes spec to the supersedes list and the document pass. Write one erratum:
  the rung 12 verifications are moot, and the shadow-pair erratum is moot.
- Re-key rung 14's private gate. Private is unsettable until slice 3's projector enforces visibility.
  This drops hook 15's capability clause.
- State that private collections ship at first launch, per the existing ruling.
- Change spec `:53` to "rung 14 unchanged except its private gate".

## Medium

### K4. MEDIUM. The build-spec errata scope misses every Navidrome-, driver-, and shadow-bearing passage outside the rung entries

**Where.** Spec `:172-174`.

**Defect.** Errata land only at rungs 12, 15, 16, and 17, launch bar part 3, and the shadow section.
These ratified passages still describe the retired design:

- "Drivers own driver knowledge" (`greenfield-server-build.md:244-247`);
- the serve-as-ingress and forward-auth trust boundaries (`:490-514`);
- "One owned prefix ... The streamer owns everything else" (`:516-519`; see K5);
- the post-login destination as a driver capability (`:603-609`);
- proxy hygiene "(streamer rung)" (`:611-630`);
- the tree shape's `internal/streamer/` and `navidrome/` and the shadow leaf (`:440-444`, `:459-461`);
- rung 4 "continuous to launch" (`:717-725`);
- the two-tracks text (`:341-357`);
- "the fallback for a failed matcher is the interim estate continuing to serve playback"
  (`:1007-1009`), which now has no fallback;
- "Box safety" ("production for Geoff's own listening", `:1038`);
- the 2026-09-25 header erratum on `/share/*` (`:68-77`).

**Proposed fold.** Add one header-level erratum to the build spec. It states that every passage on
Navidrome, a driver, SSO, the forward-auth front, the fallthrough, or the shadow is superseded by the
re-plan, and it lists the sections above by line.

### K5. MEDIUM. The one-owned-prefix rule conflicts with an in-process OpenSubsonic server, and nothing routes it

**Where.** Spec `:101-102`. After the pass, `Handler` "answers everything else with `http.NotFound`".

**Defect.** The build spec's birth requirement puts the entire HTTP surface, API included, under one
configured prefix (`greenfield-server-build.md:516-519`). The `servePrefixThreshold` comment
justifies `/dubplate` by Navidrome's surface (`serve.go:168-176`). Subsonic clients call `/rest/*` at
the server root, and `/share/*` is root-relative too. The streamer spec's hook text assumed the
OpenSubsonic surface lives "under the same 2.0 hook" (`serve.go:170-171`). These rung 12 carries are
dropped:

- the 5a carry that the `;`-segment and case-variant prefix fallthrough gaps are "rung 12's to close
  when a real driver gives them a router" (`docs/HISTORY.md:5993-5994`);
- revisiting the blanket percent-encoding ban "at rung 12 against the streamer's real client paths"
  (`HISTORY.md:5967-5970`).

**Proposed fold.** Add table rows:

- the owned-prefix rule against `/rest` and `/share` at the root, to slice 3, the first slice to
  mount `/rest`, with a build-spec erratum;
- the `;`-segment and case-variant gaps and the percent-encoding revisit, to slice 3.

Have the removal pass rewrite the `servePrefixThreshold` comment with no claim about a sibling server.

### K6. MEDIUM. The ratified group-scale principles spec keeps the 0.x-with-Navidrome world, and the document pass never touches it

**Where.** Spec `:168-189`.

**Defect.** These passages in `group-scale-principles-design.md` contradict ruling 1:

- `:34-37`: "The launch (0.x) ... Navidrome still serves the audio".
- `:44-46`: "Until then the server runs Navidrome".
- `:134-138`: the route-inventory table's delegated row is "the shell's fallthrough
  (`Shell.SetFallthrough`), which from rung 12 forwards `/rest/*` to Navidrome". The removal pass
  deletes that row's subject.
- `:401-433`: the whole "Before 0.8" section, including "whenever Navidrome is the escape hatch".
- `:561` and `:607`: decision row and acceptance row A13, the escape-hatch `shareRole` doctor check.
  Row A13 is a test a slice plan would build.

The spec removes the `shareRole` row only from the streamer spec (`:176`).

**Proposed fold.** Add a principles-spec erratum to the document pass:

- "Before 0.8" and A13 are moot;
- the delegated route class has no row until slice 3, so the route inventory carries public,
  session, and admin;
- "Versions this spec speaks to" follows ruling 1.

### K7. MEDIUM. The streamer spec revision 3 edit list misses the passages that argue against this very re-plan

**Where.** Spec `:175-182`.

**Defect.** These passages are not in the list:

- **"Relation to the current ladder" (`streamer-2-0-design.md:185-190`).** It rejects "folding 2.0
  into the ladder before launch" and "replacing rung 12 with the built-in streamer (loses the escape
  hatch)", which is exactly what this spec does. It also says "the streamer arrives as one more
  driver".
- **The header (`:7-8`).** "sequenced after the greenfield ladder's launch (rung 17)".
- **"Cutover and operations" (`:632-645`).** It mentions cloudflared and a driver selected "by one
  configuration value". The integration class is "in rung 4's shape", and rung 4's shape is being
  deleted. That matters to the ease-of-install claim in ruling 7.
- **The arm64 walk "before the flip" (`:557`).**
- **Hook 19, "the listener as a configured choice, rung 12".** It references the "proxy-hygiene
  ingress" (`:759-761`) and should be re-homed to slice 8, which carries the TLS listener.
- **Hook 15's capability clause.** See K3.
- **The open item "Whether the study rung runs beside the current ladder or after launch"
  (`:847-848`).** Ruling 5 answers it.

**Proposed fold.** Add each item to the revision 3 list, with hook 19 re-homed to slice 8 and the
rejected alternatives rewritten as superseded with ruling 1's reason.

### K8. MEDIUM. The proposed `CLAUDE.md` rewrite would drop the hook obligation for rungs 7, 9, 13, and 14

**Where.** Spec `:183-185`: "the build-for-0.8 paragraph (the hooks list is now rung 12's own
slices)".

**Defect.** Most live hooks name rungs other than 12:

- hooks 1, 12, 17, and 21 name rung 7;
- hooks 2, 3, 11, and 13 name rung 9;
- hook 16 names rung 13;
- hooks 15 and 18 name rung 14.

`CLAUDE.md`'s rule obliges each rung plan to "weigh its surface against that spec's forward-hooks
list and state per task whether it lands a hook". Reworded as "the hooks list is now rung 12's own
slices", the rule reads as if non-12 rungs owe nothing. That reintroduces the rework class the rule
exists to prevent: "a rung 6 schema or a rung 12 driver seam built without 0.8 in view is rework".

**Proposed fold.** Reword the instruction. Rung 12 is the streamer's slices, and every other rung plan
still weighs the forward-hooks list per task. Name only the retired hook 6 and the renamed hook 5.

### K9. MEDIUM. Live "to rung 12" carries from HISTORY are dropped from the routing table

**Where.** Spec `:191-211` against `docs/HISTORY.md`.

**Defect.** These carries are neither routed nor marked moot:

- **5c-ii.** "Rung 12 must re-validate every existing roster row's `streamer_username` principal, not
  only new writes" (`HISTORY.md:2528-2530`).
- **5c-i security m1.** It "becomes a rung 12 blocker": a username can equal another member's
  address, giving two members one principal
  (`research/2026-09-24-rung-5c-i-security-review.md:16-28`). The code still has it: the loader keeps
  the field optional (`internal/auth/roster.go:187`, `omitempty`), and the principal falls back to
  the address (`roster.go:816-819`). This is live under the built-in streamer, because `username`
  becomes the music-app login name (ruling 3).
- **R-5b26 / R33.** "The validator lands with rung 12's parameter or not at all"
  (`HISTORY.md:4257-4259`).
- **5c-ii.** "The header step wraps the fallthrough; rung 12 must reconcile the header set with
  Navidrome" (`HISTORY.md:2521-2522`). This item is moot.
- **The build spec's offboarding.** "Streamer credentials survive [the epoch] ... the runbook carries
  an explicit streamer-deprovision step" (`greenfield-server-build.md:551-555`). Under the built-in
  streamer this becomes slice 6's revoke.

**Proposed fold.** Add rows:

- re-validating existing rows and m1, to slice 6, or to 6d if 6d's store import makes `username`
  required (6d already refuses a missing value at import, plan `:1090`);
- R33, moot (no driver-declared destination);
- the header-set reconcile, moot;
- offboarding, to slice 6, where the epoch must revoke app passwords.

### K10. MEDIUM. The 6d plan collides with the removal pass more than "rebases onto them" admits

**Where.** Spec `:90-91` and `:218`.

**Defect.** 6d revision 2.2 (awaiting approval) names every renamed symbol:

- `errMissingStreamerUsername` (plan `:922`);
- `users.streamer_username` in its privileged-fields table (`:344` and `:348`);
- `--streamer-username` (`:221`);
- `main.go`'s registration block as a contested surface (`:1386-1387`), which the shadow deletion
  edits (`main.go:112-113` and `:143`).

Its Task 0 rule makes "a pointer whose content is gone ... a blocker, not a note". Two of its
rationales rest on the retired design:

- "rung 12 injects the username into a proxy header: one field, one shape" (`:543`);
- `Principal` "stays the streamer username and never the id" because of Navidrome's external-auth
  auto-creation (`:1427`, citing rung 6 spec `:843-846`).

If Geoff approves 6d as written, the plan he approved is stale on arrival, and its `Principal` choice
loses its stated reason.

**Proposed fold.** State that 6d moves to revision 2.3 after the removal pass merges. The revision
carries the rename sweep and the pointer refresh, plus a one-line re-justification of `Principal`
under the built-in streamer, or a change to it. State also that 6d's approval waits for that revision,
or that approval is granted on the understanding that Task 0 absorbs it. Name which.

### K11. MEDIUM. Editing migration `0003` in place contradicts the rung 6 spec's append-only rule, which the spec does not name

**Where.** Spec `:148-150`.

**Defect.** The ratified rung 6 spec says "Files are append-only once merged"
(`specs/2026-09-23-rung-6-ops-seam-store.md:424`). The spec's precondition (no store on the box) is
the right evidence, but the supersedes list (`:4-6`) does not name the rule. A diff reviewer holding
the rung 6 spec would flag the edit.

**Proposed fold.** Add a one-time erratum to the rung 6 spec, listed in the supersedes line. The
erratum names `0003`, its fixtures (`testdata/fixtures/0003` to `0005`), and the box precondition.
The in-place edit is the right call under ruling 6.

### K12. MEDIUM. The roster-row migration path the spec requires may not exist

**Where.** Spec `:152-154`: "migrates that one row through an existing verb or the bootstrap path,
never by hand".

**Defect.** The only `member` verbs are `add` and `list` (`cmd/dubplate-server/member.go:68` and
`:144`). The loader rejects unknown JSON members (`json.RejectUnknownMembers(true)`,
`internal/auth/roster.go:340`), so a roster file under the old key fails to open, and `member add`
cannot load it. The bootstrap path refuses an existing file (`ErrRosterExists`, `roster.go:268-270`).
No existing verb moves the row unless the file is first removed, and removing it is the hand edit the
spec forbids. Leaving the choice to "the plan names which" turns an unsatisfiable constraint into a
mid-pass blocker. Doctor's roster row also fails on the box from the deploy until the row moves.

**Proposed fold.** Decide in the spec. Either (a) the deploy step moves the old file aside to a dated
name inside the leaf and re-bootstraps the one admin with `member add --admin`, recorded in the
runbook, or (b) the removal pass does not deploy, and the box's roster is re-bootstrapped at 6d's
Task 8 deployment, which moves members into the store anyway. Recommend (b): one deploy, no interim
broken doctor.

### K13. MEDIUM. Ruling 2's only input lives only on the box, and the import's order against the re-layout is unstated

**Where.** Spec `:24-27` and `:81-83`.

**Defect.** The Navidrome database the group's import needs is inside
`/root/musicbox-retired-2026-10-05.tar.gz` on the box (`HISTORY.md:78-81`). The spec keeps it there
through rung 12's slices 3 to 8 and decides disposal at rung 17, with no off-box copy. The build spec
it supersedes had rung 16 archive the beets database to R2 (`greenfield-server-build.md:947`). Losing
the box before rung 16 loses every member's stars, playlists, and play counts, which ruling 2 exists
to keep.

Separately, the import's audio-key fallback (`streamer-2-0-design.md:550-551`) must read files at
Navidrome's recorded paths. Rung 16's re-layout and relocation move those files. The spec lists
the import and the re-layout with no order.

**Proposed fold.**

- Copy the archive off-box now, encrypted to the backup remote or the workstation. This is a carried
  item to the document pass or the removal pass's box step.
- In rung 16, either order the import before the re-layout, or state that the import matches by
  release MBID only, with the audio-key fallback reading the pre-move tree.

### K14. MEDIUM, OWNER FORK. The forward-auth front moves to `IDEAS.md` under a ruling about streamers

**Where.** Spec `:186-187`: "Move the streamer-integration and forward-auth front entries to
`IDEAS.md` with ruling 4's reasoning".

**Defect.** Ruling 4 says dubplate "integrates with no external streamer". A forward-auth front is an
identity source, not a streamer:

- the build spec's "One identity source" (`greenfield-server-build.md:592-596`);
- the verify-identity seam whose doc names a forward-auth front as its alternate implementation
  (`internal/auth/seam.go:20-22`);
- ROADMAP's "Cloudflare Access an optional forward-auth front (rung 12)".

These code and doc sites keep promising one:

- the `DUBPLATE_ACCESS_TEAM_DOMAIN` and `DUBPLATE_ACCESS_AUD` supersession messages: "the seam admits
  one, and its configuration arrives with the rung that builds it" (`internal/config/keys.go:133-136`);
- the runbook's forward-auth topology section (`docs/server-runbook.md:615-627`).

**Options.**

- (a) **Retire it too.** Record a ruling 4b that identity fronts retire: music apps authenticate with
  app passwords a front cannot gate, so a front would guard only the web app. Then rewrite
  `keys.go:133-136`, `seam.go:20-22`, and the runbook section.
- (b) **Keep it.** Keep the forward-auth front as a future integration on ROADMAP.

**Recommendation.** (a). A front that cannot cover `/rest` buys an operator little, and keeping the
promise in three places contradicts ruling 7's smaller surface.

### K15. MEDIUM. The removal's cleanup criteria cover only `internal/serve`, so shadow and Navidrome residue survives elsewhere

**Where.** Spec `:107-116` and `:123-125`.

**Defect.** The rewrite list names only `unidecode.go:35,77` and `serve.go:71,184`. Ruling 6's bar
("reads as if it had never had the removed part") is not met by these remaining sites:

- **A citation to a deleted file.** `internal/auth/store.go:120` cites
  `internal/engine/shadow/store.go:66-70`.
- **Shadow references.**
  - `internal/auth/store.go:28`;
  - `internal/engine/store/leaf.go:14`;
  - `internal/config/registry.go:213`, `config.go:28`, and `keys.go:197`;
  - `internal/engine/matcher/quiet.go:5`;
  - `internal/engine/tag/tag.go:58` and `:67`;
  - `internal/engine/musicbrainz/limiter.go:16`, `:25`, and `:68`;
  - `cmd/dubplate-server/main.go:11`, `:42`, and `:156`.
- **Fallthrough or streamer references.**
  - `serve.go:168-176`, the whole comment, including "the same 2.0 hook";
  - `serve.go:301`;
  - `internal/config/config.go:19`;
  - `internal/auth/token.go:163`;
  - `internal/auth/handlers.go:106`.

Two structural leftovers fall outside "a package left with no importer":

- `musicbrainz.ClassBackground` loses its only production caller (`internal/engine/shadow/run.go:271`);
- `internal/harness/analysis`'s shadow-row subject shape (`adjudication.go:34-69` and `:216-278`)
  becomes schema with no writer.

**Proposed fold.** Replace the enumerated list with a grep acceptance criterion: no comment in
`cmd/` or `internal/` names the shadow command, `engine/shadow`, Navidrome, a fallthrough, or a driver
seam. Rare historical mentions in tests are exempt. Add two decisions:

- keep `ClassBackground` only if rung 9 names a consumer, or else remove it;
- drop the shadow-row shape from the adjudication schema, or record why the corpus still needs it.

## Low

### K16. LOW. The rename list misses sites, including operator-facing ones

**Where.** Spec `:141-146`.

**Defect.** These sites are not in the list:

- `config/roster.example.json` (guarded by `TestRosterExampleFile`, `roster_test.go:440`);
- `docs/server-runbook.md:138-164`, the `--streamer-username` commands a stranger copies (ruling 7);
- `web/scripts/rendered/run.ts`;
- the `internal/backup` tests;
- the wire outcome codes `duplicate_streamer_username`, `invalid_streamer_username`, and
  `missing_streamer_username` (`internal/auth/admin.go:37-42`, `deck.ts:149-155`, and
  `+page.svelte:46-53`), which are a Go-to-web contract;
- `duplicate_principal`'s copy;
- `member list`'s `STREAMER USERNAME` column head (`member.go:177`).

**Proposed fold.** Make the acceptance a grep: no `streamer_username`, `StreamerUsername`,
`streamer-username`, or "streamer username" outside `_salvage/` and the dated documents. Add the
runbook to the pass.

### K17. LOW. The moot reason for the group 2 strip carry contradicts HISTORY's recorded reasoning

**Where.** Spec `:205`.

**Defect.** "Moot: no front sends a header". HISTORY 5d-ii kept the carry live precisely because "the
trusted-proxy class serves any stranger's reverse proxy, not only cloudflared" (`HISTORY.md:136-138`).
The carry is moot for a different reason: the removal pass deletes the configured header itself
(spec `:103`), so no header is exempt from the strip.

**Proposed fold.** Reword to "Moot: the removal pass deletes the configured header; group 2 is
stripped from every peer."

### K18. LOW. UNI-H2's deferral reason disappears with the fallthrough

**Where.** Spec `:200`.

**Defect.** `/` went to slice 8 "with the TLS listener". The redirect was carried because `/` belonged
to the Navidrome fallthrough. After the removal pass `/` is `Handler`'s own and answers 404 through
every slice before slice 8. The carry has no tie to TLS.

**Proposed fold.** Either land the `/` redirect in the removal pass, a small change in a file it
already rewrites, or record why slice 8 is still the right home.

### K19. LOW. "Username" departs from the copy register's base standard

**Where.** Spec `:28-29` and `:154-155`.

**Defect.** The visual identity spec puts UI copy on the Microsoft standard
(`visual-identity-design.md:316`). Microsoft's guide spells the term "user name", to be verified
against the guide's A-Z entry. Ruling 3 is Geoff's call, and the music apps a member types into
label the field "Username". `CLAUDE.md`'s voice rule wants such a departure recorded with provenance.

**Proposed fold.** Record the departure in the deck's provenance or the register-review note.

### K20. LOW. A new owner touchpoint is implied but not listed

**Where.** Spec `:55-57`: "each a pass with its own ... roster walk".

**Defect.** The streamer spec's roster tier "walks each slice in the group's apps"
(`streamer-2-0-design.md:212`), and Symfonium "is proprietary and paid, so the group must own a copy"
(`:835-837`). `CLAUDE.md`'s "Expected owner touchpoints, and only these" lists none, and the
document pass edits only the rung 16 sitting (spec `:184-185`).

**Proposed fold.** State whether the roster walk is scripted (conductor-run) or attended. If
attended, add it to the touchpoints list.

### K21. LOW. Other ledgers keep the old world

**Where.** Spec `:186-189`.

**Defect.** These sites still describe the retired design:

- ROADMAP's greenfield paragraph (`ROADMAP.md:22-31`): "0.x pre-release", "shadow against the live
  estate", "rung 16: beets DB migration, box swap", and "a clean shadow record";
- ROADMAP's `packaging` entry (`:202`): "ingredients (Navidrome, beets ...)";
- the memory files `version-numbering.md` and `build-for-2-0.md`.

**Proposed fold.** Name these in document-pass item 4. Point the two memory files at the new
`CLAUDE.md` paragraph.

## Checked and consistent

- **The refused-features list.** Nothing in the spec fetches music, federates, opens a public
  library, widens `download`, or loosens a share bound. `dubplate import navidrome` imports
  members' own state, never audio.
- **The version ruling.** Ruling 1 supersedes the 2026-09-25 version paragraph explicitly, and 0.9,
  1.0, and Fork B (FLAC-only through 0.9) keep their meaning.
- **The GPL rule.** The workstation-only Navidrome and gonic oracle containers match the streamer
  spec's GPL behavior-only rule and its slice 2 text.
- **The ledgers.** Every STATUS "Open items" rung 12 and 16 carry is routed or marked moot, apart
  from the HISTORY items in K9 and K1.
