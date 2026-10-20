# Rung 7 intake spec, revision 1: consistency review

Lens: consistency with ratified documents. Target: `docs/specs/2026-10-10-rung-7-intake-design.md` at
`9fc7c53` (cited below as `spec:N`). Read against CLAUDE.md, BS, the rulings ledger, SS's forward hooks (1 to 24), CH,
PS, UD, VI, W, 14A, PCF, the Melodee study, and the brief, with every code citation the spec makes spot-checked on the
tree at `9fc7c53`. Settled rulings (the renamed unmatched route, the post-rung-9 web pass, no default quota plus the
floor, one sitting, staging on the music volume, hook 12 and the guard sizing to rung 9, 14a merging first) are not
reopened.

Counts: 0 blockers, 5 major, 15 minor, 1 optional over-ceremony item. One OWNER FORK (M4).

Most citations check out: BS:20-25, :323-331, :508-516, :798, :808-814; SS hooks 1, 4, 12, 17, 20, 21, 22, 23, 24 and
the mewkiz row; CH:322-325, :420-435, :472-475, :548-568, :651; R6:181, :185-186, :562-567; PS:346-356, :510-533,
:757-764, A2, A9, A10; UD:131-137, :251-262; 14A:29-31, :81-88, :117-124, :143-175, :280-285, :429-434, :519-528; the
serve, auth, origin, limiter, tag, client, backup, keys, schema, invariant, and older-store lines. The exceptions are
findings below (m1, m5) plus one trivial offset: CH's routing row is at CH:289, not CH:287.

## Major

**M1. `DUBPLATE_STAGING_DIR` is already claimed by rung 9's `pipeline` reserved class, so the wall row panics at
registration.**
- Spec: spec:96-102 and the keys row at spec:869. The seam at spec:875-877 also limits rung 7's `keys.go` edit to
  "the `intake` class".
- Contradicted: `internal/config/keys.go:62-67` lists `DUBPLATE_STAGING_DIR` in the `pipeline` class (rung 9).
  `NewRegistry` registers that class (`registry.go:77-79`), and `claim` panics on a cross-family claim
  (`registry.go:125-128`). A wall row added from rung 7's file therefore panics with "already declared in reserved
  classes, cannot also declare in owned wall". 14A:521 gives rung 7 only the `intake` region (`:50-57`), and says
  "the regions are disjoint".
- The spec also silently reassigns a name rung 9 reserved for its own pipeline staging, which nothing lists.
- Fold:
  - The registering commit also removes `DUBPLATE_STAGING_DIR` from the `pipeline` class, as a retirement on PA:212's
    precedent.
  - Name that edit in the contested-surfaces row and in the 14A erratum.
  - Add one line saying rung 9 admits from this staging directory, so the pipeline class keeps `INBOX_DIR` and
    `QUARANTINE_DIR` only.
  - Extend AC1 to cover it.

**M2. The music-volume grant breaks the state-dir agreement leg as written, and two CLAUDE.md lines need amending.**
- Spec: spec:127-142 and AC3 (spec:975-977). The spec says the leg "gains a clause".
- Contradicted, the leg: `scripts/check-state-dir.sh:56-61` refuses every `ReadWritePaths=` entry that "is not a leaf
  directly under" `DUBPLATE_STATE_DIR`, so `/srv/music` fails the gate before any new clause runs.
- Contradicted, CLAUDE.md: its Gate paragraph (CLAUDE.md:180 onward) defines the leg as the template, every unit, and
  `DUBPLATE_SERVER_LEAVES` naming "the same state directory and leaves", with "every granted leaf is in that array".
  "Facts agents get wrong" (CLAUDE.md:319) says the server's state lives under `/var/lib/dubplate/`. Staging is the
  first server-written data outside it.
- Fold:
  - The spec states that the leg's leaf rule admits exactly one second grant class: the serve unit's music-volume
    entry, which must be the common parent of the template's `DUBPLATE_STAGING_DIR` and `DUBPLATE_LIBRARY_ROOT`.
  - `deploy.sh` pre-creates staging on its own checked line, outside the leaves loop.
  - The Owed records list both CLAUDE.md amendments: the Gate paragraph, and the fact line, which should say staging
    lives on the music volume by design.
  - AC3 adds a fixture where a second non-leaf grant fails the leg.

**M3. A check that cannot run deletes the member's upload, contradicting PCF's ratified "already uploaded" copy.**
- Spec: spec:542-547 and AC23 (spec:1029-1030). `could-not-check` makes the submission `rejected`, and "On `rejected`,
  the gate deletes the submission's staged files".
- Contradicted: PCF §3 (`research/passb-copy-final.md:108-112`, ratified post-build) covers a check that fails to
  finish because of "a network drop or server error". Its copy reads "Your album is already uploaded and waiting; no
  need to send it again" with a "Try again" button.
- Melodee L4 (`research/2026-10-06-slice-1-melodee-intake.md:350-353`) requires only "a rejection, never a pass".
  Holding the bytes satisfies it.
- Consequence: a transient statfs, read, or store error discards a multi-gigabyte residential upload, and the page
  cannot tell the truth PCF promises.
- Fold:
  - After the bytes arrive, `could-not-check` moves the submission to a held, non-terminal state that keeps its bytes
    and reservation and offers a retry. `match-unavailable` already works this way.
  - The sweep expires that state on the pushed-back hold.
  - The create-time `could-not-check` (a statfs failure before any byte moves) stays a plain refusal.
  - AC23 asserts that the bytes remain and that a retry after the fault clears reaches a verdict.
- No fork: PCF already rules the behavior.

**M4. OWNER FORK: a rejected album is terminal and fully deleted, but PCF ratifies "retry re-sends only the failing
files".**
- Spec: spec:262 and :547, AC20 (spec:1022-1023). `rejected` is terminal and deletes every staged file.
- Contradicted: PCF §4 (`passb-copy-final.md:124-126`) says "Multi-file path retry re-sends only the failing files;
  ZIP path retry copy: 'Retrying re-uploads the zip.'" PCF's own reason for `corrupt` is transit damage: "Uploading it
  again usually fixes this" (`_salvage/internal/reject/reject.go:52`).
- The spec adopts PCF "verbatim where PCF covers the surface" (spec:766) and does not list this departure.
- Options:
  - (a) Keep the spec: rejection is terminal, and the member re-sends the whole album. The state machine and
    reservation stay simple. PCF's retry line is redrafted under the register read and listed in the Owed records.
  - (b) Keep PCF: a multi-file rejection holds the passing files and the reservation, and accepts replacement uploads
    for the failing file ids. The gate then re-runs over the whole album, so the album stays the verdict unit. This
    costs a non-terminal rejected state, a replace operation, a sweep rule, and about one task.
- Recommendation: (a) at 0.8. Most reject reasons (format, CD quality, positions) are album-wide, so a per-file
  re-send rarely helps. Transit damage is what tus resumption already guards. Revisit if the sitting shows re-sends
  hurting.

**M5. Rung 7 makes a second collection, so "the picker hides at one collection, which at 0.8 is always" is false.**
- Spec: spec:578-580 creates the unmatched collection. spec:753-756 says the picker "hides at one collection, which at
  0.8 is always". spec:596-597 and :806 defer the preference and override to 14b as "the first verb that can make a
  second collection".
- Contradicted: W7.19 (`specs/2026-09-12-web-surface-standard.md:1104-1112`) hides the picker only "when one
  collection is configured". W8.2 (`:1124-1131`) makes "collection" a product concept "once a deployment has more than
  one". Once any member takes the unmatched route, the box has two collections, so W7.19 shows the picker, and the
  stated reason for the 14b deferral no longer holds.
- Fold:
  - The picker counts only collections a member may pick as a match-route target. The unmatched collection is reached
    through the route choice alone and never appears in the picker, so W7.19's picker stays hidden at 0.8.
  - Reword the 14b deferral to "the first verb that can make a second pickable collection".
  - Note that W8.2 lets the route copy name the collection.
- No fork.

## Minor

**m1. The free-space formula and its citation disagree with the repo.**
- Spec: spec:185 and :426-427 use `Bavail * Bsize`, "as `internal/engine/store/fsprobe.go:43-50` reads it".
- Contradicted: `fsprobe.go:43-50` reads `st.Type` to classify the filesystem and computes no free space. Every
  free-space read in the tree uses `Bavail * Frsize`: `store/snapshot.go:198`, `store/inspect.go:305`, and
  `backup/restore.go:551`. On Linux, `f_bavail` counts `f_frsize` units, and `f_bsize` is only the optimal transfer
  size.
- Fold: the Linux file uses `Frsize` and cites `snapshot.go:198`. Darwin has no `Frsize`, so it keeps `Bsize`.

**m2. "Fix the names and retry" cannot work on staged bytes.**
- Spec: spec:762 offers "fix the names and retry". `retry_match` (spec:630 and :688) re-probes the same staged files.
- Contradicted: UD:243-246 says "fix names/tags and re-upload". PCF §5 (`passb-copy-final.md:146-148`) pairs "check
  the artist and title spelling and try again" with a "Fix the names and retry" button. A member who edits their local
  files and presses retry gets the same push-back.
- Fold: the fix-names path cancels the staged submission and returns to selection, with the source answer kept.
  `retry_match` is the "try again" path, for after a release is added to MusicBrainz or after an outage.

**m3. The list of PCF strings to redraft is incomplete.**
- Spec: spec:766-768 and the Owed records (spec:1074).
- Missing strings:
  - PCF names Navidrome three times, not twice: `passb-copy-final.md:105`, `:143-145`, and `:160-161`.
  - Every "Private Recordings" string in §5 and §6 changes name under owner ruling 1.
  - §1's hint "One album at a time. Albums go in whole." (`:45`) contradicts the multi-album batch the spec adopts
    from PS:520-523.
- Terminology: the spec's "A member's batch is several submissions" (spec:208) uses "batch" for several albums.
  CLAUDE.md:322 and BS:812-813 use "one failing track rejects the batch" with batch meaning the album.
- Fold: list every string. Use "upload" for the multi-album set, as PS:520 does, so the CLAUDE.md fact reads
  unambiguously.

**m4. A VI erratum is owed but not listed, and the sitting cannot rule "VI's three unruled items".**
- Spec: spec:790-795.
- Moved without an erratum: VI:243-244 says "Rung 7 fixes its exact format" for the catalog line. VI:350-372 "Open
  for rung 7" lists the list view, filtering, the album page, and phone layouts. The spec moves all of them to the
  post-rung-9 pass, and owner ruling 2 names only the album page, the grid and list, filtering, and buy links.
- The sitting conflict: VI's three unruled items are the phone header, error alerts, and the list view (VI:354-360).
  The list view has moved, so the sitting rules two items.
- Fold: add the VI erratum to the Owed records, and change spec:795 to "rules the phone header and error alerts".

**m5. The `bad-extension` wording rule cites a stale line and conflicts with AC15.**
- Spec: spec:658-660 and AC15 (spec:1006-1008).
- Stale citation: `ROADMAP:133-137` is now the home copy server paragraph. The format reject shape moved to
  `ROADMAP.md:155-159` with `e833365`, `ff5061f`, and `d9e9697`, which landed after the spec's stated read commit,
  `80f91ca`.
- Conflict with the adopted strings: they state FLAC as a fixed fact ("This isn't a FLAC file, and the library keeps
  lossless FLAC"), and the next step says "Download this album as FLAC". The ROADMAP shape names the configured
  supported set and suggests a lossless conversion where one exists.
- Conflict with AC15: AC15 requires byte-for-byte equality for all nine adopted reasons.
- Fold:
  - Under the FLAC-only default, `bad-extension` renders the authority's strings byte for byte.
  - The parameterized canonical wording (a configured set, or a lossless source that can convert) is new copy under
    the register read.
  - AC15 says both, and the citation is fixed.

**m6. The adopted `corrupt` string is stretched over source-file faults.**
- Spec: spec:509-519. `corrupt` covers a header MD5 mismatch and a sample-count mismatch.
- Contradicted: the adopted text reads "This file didn't survive the upload intact. Uploading it again usually fixes
  this" (`_salvage/internal/reject/reject.go:52`, `:82`). An encoder that wrote a wrong MD5 survives every re-upload,
  so the copy is false for that case, against BS:508-511's register-gated contract.
- Fold: give a header-MD5 or stream-length mismatch its own reason (or a parameter) with a source-side next step, in
  the register read.

**m7. CH's per-collection quota level is dropped without a word.**
- Spec: spec:411-422.
- Contradicted: CH:548-549 says "Quotas bind at three levels" and "Per collection is the shape floor". The spec checks
  the member and group levels only. `collections` has no quota column (`0003_catalog.sql:12-34`), and 14b is its first
  setter.
- Fold: one line saying the collection level is neutral at rung 7 and routed to 14b.

**m8. The format allowlist's precedence inverts CH's bound.**
- Spec: spec:493-495 uses the collection's `admission_format_allowlist` "when set, else" the global setting.
- Contradicted: CH:280-282 says the global settings hold "the bounds an admin sets that no collection may exceed, such
  as the format allowlist floor". CH:164-165 says "an admin bound wins over a member override". At rung 7 the
  collection value is always NULL, but the code path is written now.
- Fold: the effective allowlist is the collection's set intersected with the global bound, within `tagio`'s
  supported formats.

**m9. The split implies a re-band to L but never states it.**
- Spec: spec:63-64.
- Contradicted: BS:682-684 says "M is one pass under one plan" and "L splits into multiple passes". 14A states its
  re-band (14A:29-31). This spec says only "as 14A split rung 14 and re-banded it L".
- Fold: add "This re-bands rung 7 as L."

**m10. The owed 14A erratum misses the subtree-adoption clauses.**
- Spec: the Owed records (spec:1068-1069) name only the `admin.go` and `internal/serve` rows.
- Moot under conductor ruling 3: 14A:522 ("otherwise mount as their own subtree from a rung 7 file, which 14b
  adopts"), 14A:36-38, and 14A:634 (14b "Adopting any JSON subtree rung 7 mounted outside `/api/v1`").
- Fold: list them in the erratum.

**m11. The buy-link capture hook is cited but never routed to SS's hook list.**
- Spec: spec:808 notes that "PS:762 names it a new forward hook there".
- Contradicted: SS's list (hooks 1 to 24) has no such hook, and CLAUDE.md:41 says to "Route newly discovered hooks
  into that list".
- Fold: add an Owed records line for an SS hook naming rung 9's MusicBrainz buy-link capture.

**m12. The prior-art verdict "none bears on intake" is not true of two entries.**
- Spec: spec:847.
- Chromaprint is SS's fingerprinting "for untagged uploads", adopted for 2.1 "as a subprocess only" (SS:1083).
- Picard's tag-mapping tables are the canonical tag names, and they bear on the new `TagSet` to `ItemFields`
  projection (spec:612-614).
- CLAUDE.md "Build for 0.8" asks each entry to be adopted or declined with a reason.
- Fold: decline chromaprint for 2.1, noting that `serve` runs no subprocess. Either consult Picard's tables for the
  projection or state why `tagio`'s names already follow them.

**m13. ZIP extraction bypasses `internal/atomicfile`.**
- Spec: spec:484-485 opens each member "with `O_EXCL`" through the staging root, and AC17 requires "no partial file".
- Contradicted: CLAUDE.md:237 requires "atomic writes via `internal/atomicfile`". SS's prior-art row says to extend it
  "rather than inventing a second temp-file convention" (SS:1081). `atomicfile` is path-based today
  (`WriteStream(path, …)`).
- Fold: either extend `atomicfile` with an `os.Root`-relative streaming form, or state why staging extraction is
  exempt (the files are scratch, deleted on any failure).

**m14. The manifest path rule admits paths the store refuses.**
- Spec: spec:461-463 relies on `filepath.IsLocal`, and AC6 (spec:985-986) makes the Go validator equal the schema
  checks.
- Contradicted: `IsLocal` is lexical after cleaning, so it accepts `a/./b`, `a//b`, and a trailing slash. The
  `rel_path` checks (`0003_catalog.sql:15-22`, R6:560-561) refuse empty and `.` components. A manifest that passes the
  gate's check then fails at insert, and the failure surfaces as `could-not-check` or `internal` instead of
  `unsafe-path`.
- Fold: require each path to equal its own `path.Clean` form and refuse others with `unsafe-path`, or store the
  cleaned form. Apply the same rule to ZIP member names.

**m15. AC13's NULL-cap case cannot pass as written.**
- Spec: AC13 (spec:1001-1002) requires "a create far above any plausible cap succeeds" with the cap NULL.
- Contradicted: the album cap (`too-large`, spec:466) and the floor refuse such a create first.
- Fold: the fixture's declared total stays under `MAX_ALBUM_BYTES` and the floor, and exceeds the other fixture
  member's `quota_bytes`.

## Over-ceremony (optional)

**o1. The accept-time re-route is built and tested at rung 7 although nothing can trigger it before 14b.**
- Spec: spec:592-595 and AC row spec:958. The spec itself says the path is reachable "only by a direct store write".
- Contradicted: CLAUDE.md:52 says a hook is "built by the rung that first consumes it". 14b's retire verb is the
  first channel that can make a target unwritable.
- Cost: about one state transition, the `rerouted` event, and a store-write fixture, roughly 0.2M to 0.4M tokens and
  part of a task.
- Fold, either way:
  - Route it to 14b and keep a one-line note in rung 7.
  - Or keep it, since CH:560-565's "re-validates at write" is ratified and the cost is small. This is the conductor's
    call.
