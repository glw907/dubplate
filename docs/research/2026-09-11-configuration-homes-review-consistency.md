# Configuration homes spec: consistency audit against ratified documents review

Adversarial review record, 2026-09-11, fresh-context `claude-opus-5` reviewer against
`docs/specs/2026-09-11-configuration-homes-design.md` revision 2. Findings are folded
into revision 3 of the spec; this file is the citation record. Reviewer output verbatim below.

---

Read the new spec, the five ratified specs/research docs, ROADMAP, STATUS, CLAUDE.md, and the five memory files. Note: `docs/superpowers/specs/2026-09-07-streamer-2-0-design.md` and the new spec both have **uncommitted working-tree edits** right now (hook 1 and hook 15 were rewritten after commit `5e4feef`); my quotes are the working-tree text.

## 1. Contradictions

**C1. SMTP and admin address homes vs rung 5's ruled wall.**
New spec:26 (table) / `:54` — "| SMTP host, port, from address |" in the *store* column; `:52` — "| site name, admin email |" in the store.
Rung 5 spec `:1532-1536` — "`DUBPLATE_SMTP_HOST` | Required. … `DUBPLATE_SMTP_FROM` | Required." (env wall, Geoff's ruling 2, memory `rung-5-owner-rulings`), and `:1553-1556` — "fork 4's admin contact copy derives from the roster's admin entry rather than from a configured address … now no address lives in configuration at all."
**Rung 5 yields nothing.** The new spec must retract "admin email" as a settings key (it lives on the roster record) and either keep the five SMTP values on the wall or file an explicit erratum against rung 5's ruled key table. `DUBPLATE_SMTP_USER` is in neither column of the new spec's table — a dropped key.

**C2. Site name and lifetimes leave the owned wall.**
Build spec `:172-176` — "**Owned, required, no defaults:** the library root, the state directory, `SiteName`, the admin identity, and … an active identity source … for magic-link that is the two secrets, `DUBPLATE_PUBLIC_URL`, and an SMTP transport." Build spec `:160-162` — "Every tunable is a threshold with a sane default: session lifetime, link expiry, rate limits, match thresholds."
New spec `:52,:55` moves site name, admin email, and "session and token lifetimes" into the store.
**The build spec wins** (ratified, revision 2.1, and rung 5 already shipped against it). The new spec needs a dated erratum in the build spec's config-wall section, the way ruling 1 and the 2026-09-08 rung-1b erratum were carried.

**C3. Integration classes split across two homes.**
Build spec `:177-189` — "**Integrations, flag-enabled, required-when-enabled:** the streamer driver …, backup targets (the rclone remote pair), monitoring (pings)."
New spec `:56-57,:61` keeps "streamer admin token, backup credentials" in env while moving "streamer driver choice and its URL", "backup schedule and targets", and "monitoring ping URLs" into the store.
**The build spec wins.** One integration class with half its correlated config live-editable and half on the wall breaks the flag-makes-correlated-config-required rule that `Load` and `doctor` enforce today. Either the whole class moves or none does.

**C4. Playlists as last-write-wins preferences.**
New spec `:25` lists "starred, playlists, queue" under member preferences, and `:32` — "Concurrent edits are last-write-wins on the server, which is enough for values a person changes by hand."
2.0 spec `:23-24` (ruling 3) — "Playlists use `updatePlaylist`'s delta semantics against the stored order, with a version column per playlist for concurrent clients", and `:269-271` — "never against a client-supplied list, so two clients removing one track each do not lose a write."
**The 2.0 spec wins.** Per-user state is four dedicated tables (`:264-267`), not preference rows. Remove starred/playlists/queue from the preferences row.

**C5. "Mirror" is ruled out of scope.**
ROADMAP `:54-55` — "Sync between rungs is publish-only; mirror and two-way sync are out of scope." Organizer spec `:215-216` — "A server-to-local read-only mirror is out of scope and would design its own schema if it were ever scheduled, **so this spec adds no column for it**." Build spec `:914-916` (Out of scope) — "two-way sync and the mirror."
New spec `:376-378` makes **Mirror** one of three attach verbs and `:405` asks rung 6 for "a last-mirror time" column.
The 2.0 spec's Local-mirroring section (`:137-146`) already reopened this, but neither ROADMAP nor the organizer spec was amended. **Both need reconciling in one edit**; as written the new spec adds a mirror column to the rung-6 schema that two ratified docs forbid.

## 2. Silent supersessions

**S1. Forward hook 1's content hash (uncommitted edit).** Hook 1 read "computed at rung 7's gate on the uploaded bytes. The cache key, the alternate key, and the drift check all read it." The working tree now appends the FLAC decoded-audio MD5. Three consequences are unstated: 2.0 `:93` — "the pipeline's content hash makes invalidation a delete on retag rather than a heuristic" becomes false (an audio MD5 survives a retag, so the transcode cache key does not change); 2.0 `:19-20`'s "indexed alternate keys" stop being unique once two collections share the audio, which the new spec `:263-266` *requires*; and the drift walk's "hashes only on suspicion" (`:74`) now means a full decode, not a file hash. Needs an explicit supersedes line plus a second key (file digest) for the cache and drift, or a retraction.

**S2. Hook 15's rewrite.** Same commit boundary: `visibility`, share rows, a last-mirror time, a duplicate-collapse flag, per-member overrides, and "per-user folder access … failing closed on private collections" were added to a ratified hook with no revision bump or erratum. The 2.0 spec's revision history stops at 2.1, 2026-09-07.

**S3. The "2.0 web admin page" does not exist in the 2.0 spec.** New spec `:7,:24,:75,:80-81` asserts it repeatedly. 2.0 `:301-306` (Web surface) names only "the device page and share management". Build spec `:914` and organizer spec `:690-691` both keep "the web admin page … deferred … built only if phone triage proves wanted." The new spec invents a ratified home for settings that was never ratified.

**S4. A TUI that plays.** New spec `:26` (audio device, ReplayGain mode) and `:416-417` assert "the 2.0 spec's 'no first-party player' ruling stands." It addresses 2.0 but not the organizer spec, whose TUI is `:175-179` "browse the library, import with interactive ask-resolution, triage duplicates, run hygiene tools" — no playback anywhere. Playback is new local-binary surface; it needs Geoff's ruling, not a footnote.

**S5. "Checkout" is the out-of-scope server-to-local mirror, renamed.** New spec `:385-388` — "A **checkout** … carries the server's collection id, is read-only … gives offline play." That is exactly organizer `:215` and build spec `:916`. Rename plus supersedes line, or retract.

**S6. Ruling 6 and the registry contract.** New spec `:9` cites ruling 6 correctly and claims no implementation work, yet `:68-71` — "The registry gains one piece of metadata: a key's home" — is a contract change to rung 1b's landed registry, and `:98-114` specifies TUI profile behavior in detail. The registry change belongs in a rung-6 or rung-1b erratum, named as such.

**S7. Layout selector.** Config methodology study `:66` (S4) — "the layout selector should be an owned key with a shipped, research-backed default", assigned as mandatory to rung 9 (`:78`). New spec `:61,:180` moves layout templates to the settings table and the collection row. Defensible, but it silently retires S4's wall-key argument; say so.

## 3. Terminology drift

- **"mirror"** — laptop→server publish (new spec `:376`) vs server→local read-only (ROADMAP `:54`, organizer `:215`, build spec `:916`). Two opposite directions, same word. Highest-harm drift in the set.
- **"checkout"** — new term for the second sense above; unknown to every other doc.
- **"profile"** — TUI connection profile (`:104-108`), ffmpeg transcode profile (2.0 `:217`), Roon user profile (`:87`), DAP export profile (2.0 `:443`). Qualify every use.
- **"library"** — the group's music on disk; Navidrome/Jellyfin's "library = a disk root folder" (`:233-235`); "a local library" meaning one member's collection (`:357`). Three senses inside one section.
- **"family"** — `:146` "the children's music stay out of the **family** collection" uses family as the standing noun, against memory `group-not-family` ("write 'member', 'the group'"). `:157` "a group of family or friends" is the licensed product-essence example and is fine.
- **"music folder"** — Subsonic entity (`:207`) vs "source folders" as a machine setting (`:26`). Different things; rename the latter to watch folders.
- **"blob"** — checked; the new spec never uses it. No drift.

## 4. Wrong citations

- `:5` "Geoff's four rulings folded in" — five are cited in the body (`:139`, `:151`, `:159`, `:354`, `:376`). Fix the count or name which four.
- `:363` "Every album carries a content hash (2.0 hook 1)" — hook 1 puts the hash on the **catalog row**, which is one row per **track** (2.0 `:160`). No album-level hash is defined, yet `:331` folds duplicates "by content hash". Define it or key on the content-hash set the album-identity tuple already uses (2.0 `:173-175`).
- `:336-337` "Rung 5 leaves member removal out of scope" — rung 5 does define offboarding (`:1777` "offboarding stays the epoch bump plus the streamer deprovision"). Narrow the claim to the removal verb and the data cleanup.
- `:413` "The post-launch TUI initiative" — no such ROADMAP entry. The commit filed the pointer under `packaging` (ROADMAP `:143-146`), while the TUI itself sits inside the Active greenfield entry (`:29-32`, "unscheduled on the new ladder"). Name a real entry or open one.
- `:144` "the Navidrome driver has no API through which to provision per-user library access" contradicts `:141`, which cites "Navidrome's per-user library grants" as field evidence. Unverified against the pinned `0.63.2`; rung 12 should verify, and the sentence should say so.
- `:26` vs `:106-107` — "server URL, saved token" are listed as machine settings and again as fields of a server profile. Pick one.
- `:65` "The left column is about a dozen names and is the whole env template" — rung 5's magic-link table alone carries eleven correlated keys plus the selector, before state dir, library root, bind, port, prefix, and the proxy pairs. Recount or drop the number.

No file was edited.
