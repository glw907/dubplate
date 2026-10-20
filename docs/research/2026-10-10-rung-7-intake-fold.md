# Rung 7 intake spec: fold record, revision 1 to revision 2

Target: `docs/superpowers/specs/2026-10-10-rung-7-intake-design.md`, revision 1 at `9fc7c53` (1,093 lines). Reviews at
`01a1b99`: `research/2026-10-10-rung-7-intake-review-contract.md` (C, O), `-mechanics.md` (M, m, O), `-integrity.md`
(B, M, m), and `-consistency.md` (M, m, o). IDs below carry their lens: `con-` contract, `mec-` mechanics, `int-`
integrity, `cst-` consistency. Section names refer to revision 2.

Each finding was checked against the tree before it was acted on. Re-verified by this fold: the `pipeline` class
holding `DUBPLATE_STAGING_DIR` (`internal/config/keys.go:62-67`); `OpenServe` and `OpenWrite` holding the store lock
shared (`internal/engine/store/open.go:117-118`, `:201-203`) and `TakeKindLease`'s exclusive flock (`lease.go:166-200`,
used by `internal/ops/local/local.go:227`); the state-dir leg's leaf rule (`scripts/check-state-dir.sh:56-61`); every
free-space read using `Frsize` (`snapshot.go:198`, `inspect.go:305`, `backup/restore.go:551`); `safeCastInt` returning 0
(`tagio/encode.go:380-390`); `tagio.Parse` allocating each payload (`container.go:199-202`); `Budget`'s lifetime limit
(`limiter.go:122-128`); tusd's `DisableTermination` and its router (`config.go:33-35`, `handler.go:62`); PCF's lines
(`passb-copy-final.md:45`, `:105`, `:108-112`, `:124-126`, `:143-161`); W8.2 (W:1124-1135); `ROADMAP.md:155-159`; VI's
catalog line and "Open for rung 7" (VI:243-244, :350-372); Picard's adoption in rung 8 (`rung-8-tag-io.md:1157`).
Taken from the lenses without re-reading: tus-js-client's `lib/upload.js` lines (the tarball is not in the tree; two
lenses read it independently and agree), the `statx(2)` `stx_mask` sentence, and the mechanics lens's probe results.

## Rulings recorded

Owner, Geoff, 2026-10-10 (written into the spec as settled and removed from "Rulings for Geoff"):

1. The unmatched route's collection is "Unlisted music"; uploader copy says "music" or "recording" where "album" would
   be wrong; the button reads along the lines of "Add as unlisted music". Spec owner ruling 1.
2. The default collection is "Library". Spec owner ruling 5.
3. Loose tracks: a member chooses "album" or "separate tracks"; separate tracks get per-track verdicts and land as
   unlisted music under a non-album path (beets singleton precedent); loose-track matching waits until after 0.8. Spec
   owner ruling 6, "Submissions and verdict units", and the gate's "Separate tracks" paragraph.
4. A rejected submission is final; the member re-sends it whole (`cst-M4`, option a). Spec owner ruling 7.

Conductor, 2026-10-10 (spec conductor rulings 5 to 9): `cst-M3` held state; `con-C11` W8.2 format exception; platform
files Linux plus a `!linux` refusal with AC 34 scoped and the store's port at 0.9; `mec-O2` chunk setting dropped;
`cst-o1` re-route moved to 14b.

## Convergent findings, folded once at the root

| Group | IDs | Root fix in revision 2 |
|---|---|---|
| Stranded finished upload | `con-C1`, `int-B1`, `con-C15` (lost response), `mec-m7`, `int-m2` | Received bytes stay in `tus/` until the submission ends, so tusd's own HEAD answers the full offset and nothing is renamed at receipt. Receipt is one idempotent compare-and-set function, run by the finish callback, the startup pass, and the sweep. The page sets `uploadUrl` from the server's file id, so a lost 201, a reload, or a second device resumes by HEAD. "The callbacks and receipt", "The client", AC 10 to 12. |
| Duplicate refusal | `con-C2`, `int-B2`, `mec-M1` | The authoritative check runs in the `Update` that enters `accepted`, on every path into it, against the target collection. "Duplicates", AC 23. |
| State-machine dead ends | `con-C3`, `con-C4`, `int-M2` | `stalled` becomes an event; a transition table names every exit; `cancel_submission` is valid from `uploading` and every held state; the sweep expires held states past `HOLD`; every transition is a compare-and-set. "Journey states", AC 26, 28. |
| Staging key reserved by `pipeline` | `cst-M1`, `mec-M2` | The registering commit removes the key from `pipeline`; rung 9 admits from this directory. "Location", AC 1, 14A erratum. |
| Store lock premise | `int-M4`, `mec-M3` | `serve` takes the `intake` kind lease before the startup pass. "One process owns intake", AC 27. |
| ZIP re-extraction after a crash | `int-M1`, `mec-M6` | Per-attempt `x/` directory removed at gate start; member rows only in the verdict `Update`; the archive removed after the verdict. AC 19. |
| ZIP quota and memory | `con-C6`, `int-M5`, `mec-M5` | The extracted total becomes the reservation in one `Update` that re-runs limits 3 to 5 before extraction; an end-of-central-directory entry count bounds `NewReader`; `tagio`'s walk bounds each block before allocating it. "Quota", "The gate", AC 16, 18, 20. |
| Platform files | `con-O1`, `mec-O1`, `mec-M4`, `mec-m4` | Conductor ruling 7: `_linux.go` plus `_other.go`; AC 34 scoped to non-`store` packages; "portable SQLite" corrected. |

## Dispositions

### Contract lens

| ID | Disposition | Where, or why |
|---|---|---|
| con-C1 | folded | stranded-upload group |
| con-C2 | folded; its early gate-pass check refused | duplicate group. Refused part: an early check against non-terminal twins would refuse a member for a twin that may still be cancelled, and the accept-time check already closes the race. |
| con-C3 | folded; the automatic re-probe refused | dead-ends group. Refused part: see refusals. |
| con-C4 | folded | dead-ends group (cancel) and stranded group (resume by `uploadUrl`, AC 12) |
| con-C5 | folded | gate steps 5 and 6: values at or below 0, runs from 1, `DISCTOTAL`, disagreeing totals, sentinel mapping; AC 21 |
| con-C6 | folded | ZIP quota group |
| con-C7 | folded | AC 17's table keyed by reject code; AC 8 (401), AC 10 (`Upload-Length`, existing upload); AC 20 (`NSamples`, bytes hashed) |
| con-C8 | folded | limit functions take `*store.Tx`; receipt decides the last file in its `Update`; AC 15 |
| con-C9 | folded | AC 9 mounts through `Shell.Register` with the production constructor; mutation proof on `Unwrap` |
| con-C10 | folded | "Refusals on the tus path" table; create check order and the `rejects` field |
| con-C11 | folded (conductor ruling 6); its one-time-comparison proposal refused | PCF section 4 is the single authority; `bad-extension`'s parameterized wording is new copy. Refused part: see refusals. |
| con-C12 | folded with mec-m6 | config load compares lexically; resolved nesting moves to the startup check; AC 1, 2 |
| con-C13 | folded | pure function over stat results, `stx_mask`, every row in AC 2 |
| con-C14 | folded | AC 5 (literal roots), 15 (NULL-cap fixture), 16 (deflated ZIP), 24 (named two-track album), 26 (state-by-action table), the state-equals-event assertion on every state-moving test, the fixture line corrected |
| con-C15 | folded | `os.Root` claim narrowed in "Layout"; the lost-response state answered by `uploadUrl`, not by the proposed lock-checked re-creation |
| con-C16 | folded | duplicate ZIP names `unsafe-path`; store error during the verdict stays `checking`; sweep idle clock reads data-file mtime; unwritten bytes named; privileged rows for `role`, `container`, `declared_bytes`, `size`; batch held at the session cap. The re-validation AC is moot (moved to 14b). |
| con-C17 | folded | AC 3 is a one-time mutation proof in the task report |
| con-O1 | folded | platform group |
| con-O2 | folded | AC 15: one plain concurrent test per property, no interleaving harness |

### Mechanics lens

| ID | Disposition | Where, or why |
|---|---|---|
| mec-M1 | folded | duplicate group |
| mec-M2 | folded | staging-key group |
| mec-M3 | folded | lock group |
| mec-M4 | folded | platform group |
| mec-M5 | folded | ZIP memory group; picture cap proposed at 8 MiB with a chain total of four times it |
| mec-M6 | folded | re-extraction group |
| mec-m1 | folded | the wrapper refuses `Upload-Concat` with 400; the claim that `DisableConcatenation` refuses anything dropped |
| mec-m2 | folded | step 3 admits depths 16, 20, and 24 only; rationale and citation corrected; AC 20 |
| mec-m3 | owner fork | "Rulings for Geoff" 1 |
| mec-m4 | folded | platform group (moot) |
| mec-m5 | folded with cst-m14 | the path validator runs on manifest paths and ZIP names; backslash refused; AC 6 |
| mec-m6 | folded with con-C12 | plus `deploy.sh` pre-creates the library root |
| mec-m7 | folded | stranded group. Concurrent creates for one id are named and left to the gate's `corrupt`, with no new lock: two tabs on one submission is the only source, and the damage is never silent. |
| mec-m8 | folded | non-retried 400 or 403 for definitive refusals; `handler.NewError` for every callback refusal |
| mec-m9 | folded | one `Limiter` per process, a fresh `Budget` and `Client` per probe run; AC 25 |
| mec-m10 | folded | an `x/exp/slog` JSON handler at warn level; the `os.Root` claim narrowed |
| mec-O1 | folded | platform group |
| mec-O2 | folded | conductor ruling 8 |

### Integrity lens

| ID | Disposition | Where, or why |
|---|---|---|
| int-B1 | folded | stranded group. Its items 4 and 5 (started means `GetUpload` succeeds; `uploadUrl`) are taken. Items 1 to 3 (rename then commit; the wrapper answering HEAD from the store) are replaced by keeping bytes in `tus/`, which removes the rename and lets tusd answer HEAD itself. |
| int-B2 | folded | duplicate group; the reviewer's own partial-unique-index alternative stays declined |
| int-M1 | folded | re-extraction group |
| int-M2 | folded; items 2 and 4 refused | dead-ends group. Refused parts: see refusals. |
| int-M3 | folded | startup pass step 1 (reconcile) and step 2 (orphans); AC 27 |
| int-M4 | folded | lock group |
| int-M5 | folded | ZIP quota group |
| int-M6 | folded | receipt syncs the data file and `tus/`; rung 9 re-verifies `file_hash` during its copy ("Moved out") |
| int-m1 | folded | cleanup is idempotent and runs after the commit, at startup, and in every sweep |
| int-m2 | folded | receipt compares and sets on `uploading`; a `tus/` file with no row is removed only past the resume window |
| int-m3 | folded | unique `(submission_id, rel_path)`; duplicate ZIP name `unsafe-path`; AC 4, 18 |
| int-m4 | folded | "Unwritten bytes" |

### Consistency lens

| ID | Disposition | Where, or why |
|---|---|---|
| cst-M1 | folded | staging-key group |
| cst-M2 | folded | "The unit grant": one second grant class; `deploy.sh` lines; CLAUDE.md errata owed; AC 3 |
| cst-M3 | folded | conductor ruling 5: `check-unfinished`; AC 24 |
| cst-M4 | owner fork, ruled by Geoff 2026-10-10 (option a) | owner ruling 7; PCF erratum owed |
| cst-M5 | folded | "The collection picker" |
| cst-m1 | folded | `Frsize`, citing `snapshot.go:198` |
| cst-m2 | folded | fix-names cancels and returns to selection; `retry_submission` is "try again" |
| cst-m3 | folded | PCF erratum lists all three Navidrome strings, every "Private Recordings" string, and the section 1 hint; "upload" names the multi-submission set |
| cst-m4 | folded | VI erratum owed; the sitting rules two items |
| cst-m5 | folded | `bad-extension` renders PCF byte for byte under FLAC-only; parameterized wording is new copy; citation now `ROADMAP.md:155-159` |
| cst-m6 | folded | new reason `source-damaged` for a header MD5 or sample-count mismatch; `corrupt` keeps transit damage |
| cst-m7 | folded | per-collection quota neutral, routed to 14b |
| cst-m8 | folded | effective allowlist is the collection's set intersected with the global bound |
| cst-m9 | folded | "re-bands rung 7 as L"; BS erratum owed |
| cst-m10 | folded | 14A erratum lists the subtree-adoption clauses |
| cst-m11 | folded | SS erratum: a buy-link hook for rung 9 |
| cst-m12 | folded | chromaprint declined with a reason; Picard consulted through `tagio` |
| cst-m13 | folded | extraction exempt from `atomicfile`: scratch per-attempt directory, read by nothing before the verdict |
| cst-m14 | folded with mec-m5 | path validator |
| cst-m15 | folded with con-C14 | AC 15's NULL-cap fixture |
| cst-o1 | folded | conductor ruling 9: re-validation at write moved to 14b |

## Refusals

Whole findings refused: none. Parts refused, each with its reason:

1. **Automatic re-probe of `match-unavailable`** (`con-C3`, `int-M2` item 4): the member's "Try again" is PCF's
   ratified copy, and the hold expiry already ends the reservation; a background retry loop is machinery no measured
   defect asks for.
2. **A gate deadline derived from size and throughput** (`int-M2` item 2): no hang is measured; mewkiz's decode advances
   on every frame and ends at EOF, the gate checks its context between frames, and the stall event already mails the
   admins.
3. **An early duplicate check at gate pass** (`con-C2`): it would refuse a member for a twin that may still be
   cancelled; the accept-time check closes the race.
4. **A one-time diff-reviewer comparison in place of a permanent string test** (`con-C11`): literals transcribed from
   PCF are a contract golden, not a copy of the implementation, and CLAUDE.md makes reject wording contract.
5. **Lock-checked re-creation after a lost 201** (`con-C15`), and **the wrapper answering HEAD from the store**
   (`int-B1` items 1 to 3): superseded by keeping received bytes in `tus/` and setting `uploadUrl`, which needs neither.

## Owner forks

- `mec-m3`, ID3-wrapped FLAC files: open, in "Rulings for Geoff" 1. The fold recommends (b), a new reason naming the
  fix; the mechanics lens recommended (a). The fold departs because `corrupt`'s "Uploading it again usually fixes this"
  would send a member round the same loop `source-damaged` was added to avoid.
- `cst-M4`: ruled by Geoff, 2026-10-10.

No other genuine product or taste fork was found.

## Owed errata (ratified documents not edited)

- **SS**: hook 12 moves to rung 9; a new forward hook for rung 9's MusicBrainz buy-link capture (PS:762).
- **CH**: the routing row (CH:289) narrowed by `intake_route`; the per-member counter counts what the member added;
  the startup refusal supersedes CH:474's copy fallback for staging and the library root.
- **14A**: the contested-surfaces rows for `admin.go` and `internal/serve` are sequenced after 14a's merge; the
  `keys.go` row gains the `pipeline` edit; the subtree-adoption clauses (14A:36-38, :522, :634) are moot.
- **W**: W8.2 gains the format-name exception for reject reasons (conductor ruling 6, 2026-10-10).
- **VI**: the catalog line's format (VI:243-244) and the list view, filtering, album page, and phone layouts in "Open
  for rung 7" move to the web pass after rung 9; the sitting rules the phone header and error alerts only.
- **BS**: rung 7 re-bands from M (BS:798) to L.
- **PCF**: the retry line (`:124-126`) for owner ruling 7 (Geoff, 2026-10-10); the three Navidrome strings (`:105`,
  `:143-145`, `:160-161`); every "Private Recordings" string in sections 5 and 6, renamed "Unlisted music" (Geoff,
  2026-10-10); the section 1 hint "One album at a time. Albums go in whole." (`:45`).
- **CLAUDE.md**: "The album is the verdict unit" gains the separate-tracks kind (Geoff, 2026-10-10); the Gate
  paragraph's state-dir leg admits the serve unit's music-volume grant; the `/var/lib/dubplate/` fact line gains that
  staging lives on the music volume by design.

## Not settled by this fold

- The tus-js-client line citations, the `stx_mask` sentence, and the probe results rest on the lenses' reads; the plan
  quotes them before writing.
- tusd's response body for a `handler.NewError` refusal is left for the plan to quote, so the page's reading of the
  reject code may need one adjustment.
- Decisions this fold made that the conductor may want to check: rung 9 admits loose files from tusd's `tus/<id>`
  layout; `route_unlisted` is offered from `pushed-back` only, not from `match-unavailable`; separate tracks carry no
  extras; a separate track's duplicate refusal compares audio keys against accepted files in the unlisted collection.
- 14a's open question on whether rung 7 and a running rung 12 slice may both write `internal/serve` stays the
  conductor's, before 7c's `Shell.terminated` task.

## Measures

**New-mechanism findings.** Folded 6, refused 3.

| Folded mechanism | Source that runs it | Defect it answers |
|---|---|---|
| The `intake` kind lease | `store.Handle.TakeKindLease`, run by `internal/ops/local` (`local.go:227`) | `OpenServe` holds the store lock shared (`open.go:201-203`), so a second `serve` re-queues and gates work twice (code-read by two lenses, re-verified) |
| Receipt replay at startup and sweep, with bytes kept in `tus/` and `uploadUrl` | tusd's filestore derives offset from the file (`filestore.go:166`); tus-js-client's documented `uploadUrl` resumption | tusd calls the finish callback only from the last-byte PATCH (`unrouted_handler.go:993-1010`), so a failed receipt strands a file the client reports sent (code-read by two lenses independently) |
| Startup reconciliation of missing staged bytes | the repo's restore reconciler (`internal/backup/reconcile.go`) | backups exclude staging (`internal/backup/doc.go:4-7`), so a restore leaves rows without bytes and an `accepted` one blocks re-upload (code-read) |
| End-of-central-directory entry count before `zip.NewReader` | `archive/zip`'s own `readDirectoryEnd` reads the same record | measured: 77 MiB of heap for a 35 MB ZIP of 400,000 entries (mechanics probe) |
| `tagio`'s block bound before allocation | the FLAC metadata header `tagio`'s walk already reads | `container.go:199-202` allocates every payload before any cap applies (code-read, re-verified) |
| Sync before receipt | `internal/atomicfile` (`atomicfile.go:29`, `:87`) | tusd's filestore never syncs (`filestore.go:220-228`, `:311-313`) while the store commits durably on another device (code-read) |

Refused: the automatic re-probe, the size-derived gate deadline, and the lock-checked re-creation (refusals 1, 2, and
5 above).

**Spec length.** 1,093 lines before, 1,227 after: +134, about 12%, inside the 25% signal. Growth came from separate
tracks, the transition and tus-refusal tables, receipt, and the startup pass; the platform table, the chunk setting,
the accept-time re-route, and the second naming ruling came out.

**Estimate.** 7a 8.5M to 9.0M (flag 7.2M); 7b 6.5M to 6.8M (flag 5.4M); 7c unchanged at 7.5M. The rung goes from about
22.5M to about 23.3M, and execution from about 16 hours to about 17, excluding the sitting and owner pauses.

## Second fold: the verification read

Source: `research/2026-10-10-rung-7-intake-fold-verification.md` (1 blocker, 1 major, 5 minors), on revision 2 at
`5cf5241`. Scope set by the conductor: B1, M1, m1 to m4, and m5 with Geoff's ID3 ruling. The verifier's `Reader.init`
quote was re-read from Go 1.27.2's `archive/zip/reader.go:118-161` and `:356-380`, mewkiz's ID3v2 test from
`flac.go:136` and `:163-173`, and mutagen 1.48.1's two ID3 tests from the repo's beets venv
(`mutagen/id3/_file.py:377`, `mutagen/id3/_id3v1.py`, `find_id3v1`). The `flac -t` behavior and the 196,609-entry
probe are the verifier's.

### Ruling recorded

Geoff, 2026-10-10: the ID3 fork takes option (b). A new reason, `id3-tags`, tells the member the files carry ID3 tags
and how to strip them. Spec owner ruling 8; "Rulings for Geoff" now reads "None open".

### Dispositions

| ID | Disposition | Where |
|---|---|---|
| ver-B1 | folded | "Staged bytes, per state": a ZIP's archive is its staged bytes through `uploading`, `checking`, and `check-unfinished`; from the verdict that leaves `checking`, the kept `x/` members are, and only then is the archive removed. `check-unfinished` removes `x/` and keeps the archive. Reconcile and cleanup test against that set. AC 24 gains the ZIP held and retried; AC 27 gains an accepted ZIP with members and no archive, and a held ZIP with only its archive, both surviving a restart. |
| ver-M1 | folded | the end-record pre-check is replaced by a constant-memory count of central-directory headers from the directory offset, with `Reader.init`'s stop rule quoted (`reader.go:139-161`, `:356-380`; `struct.go:44`), refused at the `MAX_FILES + 1`th header. `readDirectoryEnd` is unexported, so the plan ports its offset arithmetic (`:544-610`). AC 18 gains a ZIP whose zip64 end record claims one entry. |
| ver-m1 | folded | tus refusal row "receipt: a store error, 503, `could-not-check`, retried"; states row; AC 10 scoped to definitive refusals |
| ver-m2 | folded | "The client": `endpoint` is the base path beside `uploadUrl`, with the 4.3.1 failure text quoted |
| ver-m3 | folded | AC 26's worker-after-cancel clause removed, since no member action can change a submission under the worker or probe; the one reachable race, receipt after a cancel, stays AC 11's. The general compare-and-set rule stays. The states row that claimed the race was removed. |
| ver-m4 | folded | an extra in a separate-tracks manifest answers `invalid_request` with detail `extra` (the page leaves extras out, so only a client fault sends one); an extra inside a separate-tracks ZIP is dropped with a non-blocking warning; every passing track a duplicate ends `rejected`. AC 22 gains both. |
| ver-m5 | folded with the ruling | the gate detects an ID3v2 prefix by its first three bytes, `ID3`, before `tagio.Sniff` (mewkiz `flac.go:136`, `:163-173`; mutagen `_file.py:377`), because `flac -t` and mewkiz both accept it; it detects a trailing ID3v1 record as a decode error with the last 128 bytes beginning `TAG` (mutagen `find_id3v1`), which both decoders refuse. Testing ID3v1 only after a decode error keeps a valid FLAC whose last bytes spell `TAG` from being refused. AC 20 gains both files and that valid one. |

The ver-m4 choice for a ZIP's extras (drop with a warning, not refuse) is the fold's. It keeps a member who zips a
folder of singles with a cover image from a refusal they cannot act on in the page. The warning is new copy.

### Owed errata added

- **PCF section 4, the copy contract**: the `id3-tags` string (Geoff, 2026-10-10), which names ID3 tags and how to
  strip them, and the separate-tracks extras warning, after the register read with the other new reasons.
- **W**: the W8.2 exception (conductor ruling 6) also covers "ID3 tags" in `id3-tags`, by owner ruling 8; "tags"
  stays a leak everywhere else.

### Not settled

- The `id3-tags` next step must name a way to strip ID3 data from a FLAC file that a member can run. `metaflac`
  removes Vorbis comments, not ID3. The candidates are mutagen's `mid3v2 --delete-all` and a desktop tagger's "remove
  ID3" command. The copy read picks one; the spec does not.
- `flac -t`'s ID3v2 warning text and the 196,609-entry probe come from the verifier's runs, which were not re-run here.

### Measures

New-mechanism findings: 2 folded, 0 refused.

| Folded mechanism | Source that runs it | Defect it answers |
|---|---|---|
| A constant-memory central-directory header count, replacing the end-record count of the first fold | `archive/zip`'s own `Reader.init` loop and stop rule (`reader.go:139-161`) | measured: a 21 MB ZIP whose zip64 end record claimed 1 entry gave `NewReader` 196,609 entries and 53 MiB of heap (verification probe) |
| ID3 detection at the gate (a three-byte prefix test, and a trailing-record test after a decode error) | mewkiz/flac's `id3Signature` test; mutagen 1.48.1's `startswith(b"ID3")` and `find_id3v1` | measured: `flac -t` and mewkiz accept an ID3v2-prefixed file and refuse an ID3v1-suffixed one, while `tagio.Sniff` misreports the prefix (verification probe and mechanics m3) |

The first fold's "end-of-central-directory entry count" row is superseded by the first row above.

**Spec length.** 1,227 lines before the second fold, 1,275 after: +48. Against revision 1's 1,093 lines the spec has
grown about 17%, still inside the 25% signal.

**Estimate.** Unchanged. The header count replaces the end-record read, and `id3-tags` adds one reason and two
fixtures inside 7a's 9.0M.
