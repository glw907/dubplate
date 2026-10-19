# Photo curation spec review: data integrity and failure risk

Target: `docs/superpowers/specs/2026-09-25-photo-curation-design.md` at `67d4adf`.
Lens: data integrity and failure risk only. Findings are ranked by consequence. Line numbers are
the spec's at `67d4adf`.

Summary: 1 blocker, 6 major, 5 minor, 2 owner forks (inside D6 and D8).

The design is sound in its main moves. One file per record avoids a shared manifest. Only picked
images reach git. The artifact database makes a sitting resumable. The gaps are about what holds
state between those moves: where candidates live, which identity keys a record, when the owner's
choices are frozen, and what evidence survives a changed source page.

## Blocker

### D1. Candidate storage is undefined and, if it is session scratch, it is deleted under the owner's picks

- Severity: blocker
- Location: spec:128-131, spec:149-152, spec:171-174
- Defect: candidates "wait in untracked scratch storage until Geoff picks them", and a sitting
  "can stop and resume on any device". A sorting sitting can span days and sessions. The
  session-scoped scratch directory is deleted when the session ends. The spec never names where
  the downloaded files and the status-unset records live, or whether the records sit in `photos/`
  while the files sit in scratch. If scratch goes, the artifact database still holds Geoff's `Y`
  choices, but their images and verification work are gone. The conductor then re-downloads.
  The source may now serve a different file, a different crop, or nothing. Geoff's pick would
  then be applied silently to an image he never saw. That corrupts the one input that costs
  attended time.
- Fold: name one durable, gitignored candidate home that outlives sessions. Two options fit. An
  in-tree `/photos/.candidates/` entry in `.gitignore` works, though lane 1's worktree gitleaks
  leg scans it (see D2). An out-of-tree home such as `~/.cache/dubplate/photo-candidates/` avoids
  that scan. Hold each candidate's file and record there together. Rule that a pick is applied
  only to a file whose hash equals the record's `source_sha256`. If the file is missing, or a
  re-download does not match, the candidate goes back to Geoff unpicked. Never apply a pick to a
  substitute file.

## Major

### D2. Commits share a tree and an index with lane 1

- Severity: major
- Location: spec:173-174
- Defect: "path-scoped commits" is the only rule for a tree where lane 1 has staged files. A
  plain `git add photos/...` followed by `git commit` sweeps lane 1's staged work into a photo
  commit. Concurrent git operations from both lanes race on `.git/index.lock`. A failure halfway
  through a move-then-commit leaves images in `photos/` uncommitted, and lane 1's next
  `git status` or gate then sees them. Lane 1's gate runs gitleaks over the worktree
  (`--no-git`), so candidate images or records placed in-tree also enter lane 1's gate. The
  process section never registers the pass in the lane registry, which CLAUDE.md names as the
  live-executor check.
- Fold: state the exact mechanic. Commit with `git commit --only -- photos/<slug>.webp
  photos/<slug>.json ...`, and never stage a path first with a separate `git add`. Check
  `git show --stat HEAD` afterwards against the intended path list. Register the pass as a
  document lane in the lane registry. Keep candidates out of the tree (D1). Retry on
  `index.lock`, and never delete the lock.

### D3. The owner's choices can change after the conductor reads them, and nothing reconciles the drift

- Severity: major
- Location: spec:171-175
- Defect: the database is live on every device. Geoff can flip `Y` to `N` on his phone after the
  conductor read it. If the image is already committed, the "only picked images are committed"
  property breaks for good, because the blob stays in history. A flip from `N` to `Y` after the
  read leaves a record `rejected` against his intent. Nothing compares the database to the record
  statuses afterwards, so the two stores drift silently. The spec also never says what the
  database keys on. If it keys on the slug and the slug changes (D4), the choices are orphaned.
- Fold: add a round-close step. The conductor writes a `closed` marker for the round into the
  database, and the page disables edits for a closed round. The conductor then reads a snapshot
  and commits against that snapshot. Each round close writes a reconciliation line: database
  choice against record status for every candidate, with zero mismatches required. Key the
  database on an immutable candidate ID (the `source_sha256` works), never the slug. A change
  after close goes through an explicit reopen, which for a committed image means removal (D8).

### D4. Parallel agents can mint the same slug and overwrite each other, and a crash leaves orphans

- Severity: major
- Location: spec:92-104, spec:149-152
- Defect: one file per record prevents manifest collisions, but not slug collisions. Two agents
  that find two different Dischord House photos from 1984 both write
  `dischord-house-1984.{webp,json}`, and the second write silently replaces the first. Its record
  can then describe the other agent's image. An agent killed mid-write leaves a truncated JSON
  file, an image without a record, or a record without an image. The slug is also promised
  stable "for the life of the project" as a website path, yet it is minted by a search agent
  before any fact in it has been checked.
- Fold: candidates carry an immutable ID (`source_sha256`, or an agent prefix plus a counter),
  never a slug. The conductor mints the slug once, at pick time, with a uniqueness check across
  `photos/`. Agents write to a temp name and rename into place, and refuse to overwrite. The
  round close lists orphans in both directions and truncated or unparseable JSON, and none may
  remain before sorting starts. State that a slug is never renamed. A factual correction changes
  the record, and the slug stays.

### D5. Dedupe has no key, so duplicates and license-laundered copies get through

- Severity: major
- Location: spec:129-130 ("so later rounds do not find it again"), spec:149-152
- Defect: the only hashes are exact-file hashes. The same photograph reached through a Flickr
  page, a Commons re-upload, a crop, or a different resolution has a different hash and URL. It
  therefore evades the rejected-record memory, and two copies can both be picked. The worse case
  is licensing. A Commons copy uploaded as "own work" CC BY-SA and a Flickr original marked
  all-rights-reserved are the same image. Without dedupe, the agent that found only the Commons
  copy verifies a license that the true author never granted. That is the laundering case
  spec:143-144 rules out, and here it is invisible. Two agents in one round can also find the
  same image at once, and neither sees the other's record.
- Fold: record a perceptual hash (dHash or pHash) and the full list of known source URLs in each
  record. Each agent checks all existing records, whatever their status, and the in-flight
  candidates before it writes. The round close runs a near-duplicate pass across the round. When
  copies exist, the license is taken from the upstream-most copy that names its author. A
  conflict between copies' licenses rejects the image.

### D6. The legal evidence trail does not survive a changed or deleted source page

- Severity: major
- Location: spec:113-117, spec:158-159, spec:188-189
- Defect: `verified` holds a quoted string and a date. Nothing preserves the page. Flickr users
  change licenses after the fact, and Commons files are deleted for copyright violation. The
  license that governs is the one in force when dubplate received the file. That can be proved
  only by a third-party snapshot, which a quote dubplate typed itself does not provide. The Opus
  spot-check reads the live page. It can then flag a correct quote whose page changed since, or
  pass a page that now differs from what the agent saw. `source_url` is the license page, not
  the file URL downloaded, so the file's origin is also unrecorded. A letter grant has no source
  page at all. "Verified like any other source" has nothing to verify against. The spec also does
  not say where the grant lives, or whether an email with a person's name and address may go
  into a public repository.
- Fold: add these evidence fields and require them for `picked` and `pack` records:
  - `archive_url`: a Wayback Machine Save Page Now snapshot of the source page, taken at
    verification.
  - `file_url`: the exact URL downloaded.
  - `retrieved_at`: a full timestamp, not a date.

  The spot-check compares quotes against `archive_url`, and the live page only as a secondary
  signal. For letter grants, ask the grantor to publish the license where it can be verified.
  Wikimedia VRT confirmation, or a line on the grantor's own site, gives a real source page and
  archive.
- OWNER FORK, for grants that stay private:
  - (a) Commit a grant record: the grant text, its scope (the named images and the license), the
    grantor's name, the date, and the email's `Message-ID`. The full email stays in Geoff's
    mailbox.
  - (b) Keep grant evidence outside the repository entirely, and have the record cite only a
    `Message-ID`.
  - Recommendation: (a). The repository is the durable record, and a grant scoped to named images
    and a license carries no private data beyond a name the grantor wants credited.

### D7. Nothing stops a wrong record or file from being committed before the consumer rung's gate exists

- Severity: major
- Location: spec:133-136, spec:120, spec:140-142
- Defect: validation is a manual `ajv` run "after each round", and the gate leg arrives with the
  consuming rung. Until then, nothing in `scripts/check.sh` rejects the following:
  - a committed `.webp` whose record is `rejected` or `pack`, or has no record;
  - a hash mismatch;
  - a record whose `license` is outside the shipped classes;
  - a `public-domain` record with no US status.

  The schema as described checks enums, not cross-field rules. `status: picked` together with
  `CC-BY-NC-4.0` would pass. The consumer rung then inherits a pool it must audit backwards.
  This conflicts with the repository rule that "a check either gates or does not exist". The "no
  Go" argument is sound, but it does not cover a script leg.
- Fold: land the gate leg in this pass as a shell or Node script in `scripts/check.sh`, not as a
  Go package. Node is already a gate requirement. Make `schema.json` enforce these with
  `if`/`then`:
  - `picked` requires a license in the shipped classes, plus `verified`, `archive_url`, and
    `sha256`.
  - `public-domain` requires the `public_domain` object with a US status of public domain.
  - `pack` and `rejected` forbid an image file.

  The script checks every `.webp` against a `picked` record and its `sha256`. It also checks
  every `picked` record against its `.webp`. Lane 1 runs the same gate, so photo commits must be
  green when they land. D2's path-scoped commit keeps a red candidate out of lane 1's tree.

## Minor

### D8. No procedure for a takedown, a revocation claim, or a Commons deletion after commit

- Severity: minor now, major once a binary ships
- Location: spec:138-145 (absent)
- Defect: CC licenses are irrevocable. A claim that the uploader was never the author is still a
  valid removal reason, and so is a Commons copyright-violation deletion. The spec has no
  `withdrawn` status, no re-check of committed sources, and no rule that a withdrawn image leaves
  the binary. Removal from HEAD leaves the blob in history.
- Fold: add a `withdrawn` status with a reason and date. The D7 gate refuses a withdrawn record
  that still has an image. Add a periodic link and deletion check, for example before each
  release.
- OWNER FORK, on history:
  - (a) Accept that a withdrawn image stays in git history. Removing it from HEAD stops future
    binaries.
  - (b) Store images outside git history (Git LFS or R2), so a takedown is complete.
  - Recommendation: (a). A source repository's history is low exposure. The binary is the
    distribution act. And (b) adds infrastructure for about 300 files.

### D9. Rejected and pack records leak into the repository and possibly the binary

- Severity: minor
- Location: spec:129-131, spec:121
- Defect: rejected and pack records are committed into `photos/` next to picked ones. They carry
  photographer names, source URLs, and `screen` notes. A public note such as "2: muddy, dull" on
  a named photographer's work is a small reputational cost, and it sits permanently in an MIT
  repository. A consumer that embeds `photos/*.json` ships every rejected and NC record in the
  binary. Pack images are not committed and not durably stored (D1), so the optional pack later
  has no files. The consumer rung's check that "every slug [has] both files" (spec:136) also
  fails on every rejected record unless it is status-scoped.
- Fold: move rejected records to `photos/rejected/`, stripped to dedupe keys (source URLs,
  hashes, and the D5 perceptual hash) plus a reason code. Drop `screen` from every committed
  record at commit. Give pack records their own directory, and store their files durably outside
  git. Tell the consumer rung to embed only `picked` records.

### D10. Flyers and zine covers have two licensors but one license field

- Severity: minor (major if flyers form a real share)
- Location: spec:111, spec:114, spec:144-145
- Defect: the rule that the artwork's license must come from its artist has no field. A record
  can carry only the scanner's CC BY and still pass schema.
- Fold: for `format` of flyer or zine cover, the schema requires an `artwork` object. It holds
  the artist, the license, and evidence, with the same fields as D6.

### D11. The stored hash and file content are not pinned to a defined transform

- Severity: minor
- Location: spec:118-119, spec:123-126
- Defect: the spec never says who encodes the WebP, or when. If the conductor re-encodes after an
  agent recorded `sha256`, the record is wrong. WebP output also varies by encoder version and
  settings, so a re-run cannot reproduce the file. Default conversion may keep EXIF GPS in a
  photo of someone's home, or drop embedded rights metadata.
- Fold: `sha256` is written by the step that produces the committed file, after the final
  encode. Record the encoder and its settings once in `vocab.json` or a README. Strip GPS and
  other location EXIF.

### D12. A vocabulary edit can silently re-family or invalidate committed records

- Severity: minor
- Location: spec:48-58, spec:72-73
- Defect: family and region are derived from `vocab.json`. Renaming or moving a term changes the
  counts, or makes old records fail with no migration.
- Fold: a vocabulary edit that removes or moves a term rewrites the affected records in the same
  commit, and the D7 gate proves the whole set still validates.
