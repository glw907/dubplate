# Configuration homes spec: collection interaction permutations review

Adversarial review record, 2026-09-11, fresh-context `claude-opus-5` reviewer against
`docs/superpowers/specs/2026-09-11-configuration-homes-design.md` revision 2. Findings are folded
into revision 3 of the spec; this file is the citation record. Reviewer output verbatim below.

---

I've read the config spec, the relevant 2.0 sections, and researched Navidrome's multi-library model. No files edited.

---

# Adversarial review: collection interaction model

The spec gives collections an `owner` beside `writer` and three attach verbs, then stops. Everything that happens *between* collections and *between* members is unwritten. Below, ranked by severity.

Headline: **one-writer/no-merge survives.** Only two findings threaten it (F2, F3), and both are fixed by refusing a thing by construction, not by adding a merge rule. Do not abandon the invariant.

---

## (a) Rules I would adopt with confidence

**F1 — CRITICAL. Contribution into a laptop-written collection is a second writer.**
The spec says "other members contribute to a member-owned collection only if the owner allows," and owner/writer are declared independent. Read literally, that authorizes contribution into a laptop-written collection: the server would write rows the laptop also writes. The invariant dies in one sentence of prose.
*Rule:* the contribution toggle exists only on server-written rows. The API returns a typed refusal on a laptop-written row, and the TUI does not render the control. "Who may contribute" is a property of the writer, not the owner.
*Rejected:* a queue on the server that the laptop drains on next connect (that is a merge with extra steps, and it strands contributions for the year-away laptop of F6).

**F2 — CRITICAL. Checkout leaks the writer.** The reverse of hand-over ("checking out a server collection as a local read-only mirror") is described as a mirror, and mirror elsewhere means *a collection on the server written by the laptop*. Two rows would claim the same content with different writers.
*Rule:* a checkout is a local cache, not a collection. It carries the server's collection id, is flagged read-only, never appears in the server's collection list, and can never be mirrored back. Wanting to write it means asking for hand-over in the other direction, which is a deliberate writer transfer.

**F3 — HIGH. Visibility is absent, and the Navidrome driver cannot enforce it.** Collections map to Subsonic music folders. Navidrome does support multi-library with per-user access since v0.58, but libraries and grants are configured **through the admin UI only** — there is no API surface rung 12's driver can provision against, and rung 12's scope is SSO plus roster sync.
*Rule:* add `visibility` to the row at rung 6 (free in the schema) and make per-user folder access a declared driver capability in hook 6's report. A driver that does not declare it **fails closed**: a private collection is not projected to that streamer at all, rather than projected to everyone. Second rule, stated in the spec's own words: private means not-listed and not-streamable to other members; it is never confidentiality from the admin, who holds the disk, the backups, and the logs. Write that sentence down so no one builds a promise the system cannot keep.

**F4 — HIGH. Removal semantics are undefined across three shared surfaces.** Files are content-addressed and back several collections; per-user rows (starred, playlists, history) point at album ids; album identity is *(release MBID, collection)* per the 2.0 spec.
*Rules:* (i) removing an album from a collection deletes that collection's rows and layout paths only; the blob is freed by a refcount sweep, never inline. (ii) A move is remove-plus-add in one transaction, and it **mints a new album id**, so starred rows, playlist entries, and history for the source album are lost. That is a real user-visible loss — the confirm dialog must state it by count ("2 members have this starred; it appears in 1 playlist"). (iii) Delete-collection is retire (unprojected, files kept) by default; hard purge is a separate admin-only act.
*Rejected:* re-pointing per-user rows across the move. The projector already has a merge/split re-point path, but that is for retags inside one collection; extending it across collections makes the album id collection-independent, which contradicts the 2.0 identity rule.

**F5 — HIGH. Member departure is undefined, and rung 5 explicitly does not build removal** (rung 5 spec, "Member removal… out of scope"). An unowned collection is a permanently unadministrable row.
*Rules:* removing a member never deletes audio. Server-written collections they owned revert to the admin. Laptop-written collections freeze at their last mirrored state and show a "writer gone" status. Preferences and their per-user rows are deleted. File this at rung 14, not rung 6 — but the schema must not make it impossible (owner is nullable-to-admin, not a hard FK with cascade delete).

**F6 — HIGH. The year-away laptop.** A mirrored collection whose writer never returns is real files on the server's disk with no one able to change them.
*Rules:* the row carries a last-mirror timestamp; past an admin-set bound it is flagged stale but keeps serving. Never auto-delete. Give the admin one verb, **seize**: a unilateral hand-over that transfers the writer to the server and invalidates the laptop's mirror token. A returning laptop learns on connect that the server now writes, and its local copy demotes to a checkout (F2). This reuses hand-over rather than inventing a recovery path.

**F7 — HIGH. Quotas are per-collection only, which is gameable.** A member with a per-collection cap creates ten collections.
*Rules:* three bounds, all in the settings table — per collection (shape floor), **per member aggregate** across owned collections, and a group total, because the box has one disk. A shared blob is charged to **every** collection referencing it: it over-counts, it is trivially explainable, and it removes any incentive to game which collection "owns" a file.
*Rejected:* splitting a shared file's bytes across referencers (unstable totals that move when an unrelated member deletes something).

**F8 — MEDIUM. The default-collection preference can go stale under the member.** The stored id may be retired, closed to contribution, or handed over between uploads.
*Rule:* the uploader resolves the preference at page load, not at write. If unwritable, it falls back to the admin-owned default collection **with a visible notice**. Silently uploading into an unintended home is the worst outcome; the album is the verdict unit and an album in the wrong collection is a manual extraction.

**F9 — MEDIUM. Partial overlap: same release, different rip.** The overlap report is two buckets (on the server / not), and contribute "refuses duplicates by hash or MBID." So a better rip of an album the server already holds is silently rejected.
*Rule:* three buckets — absent, identical (hash match), and **same release, different bytes** (MBID match, hash differs). The third is never auto-refused and never auto-replaced. It goes to the member as a choice: skip, mirror instead, or offer as a replacement the *target owner* adjudicates. Applying the album-as-verdict-unit rule to bytes: no automatic winner between two rips.

**F10 — MEDIUM. "Copy this album from her collection into mine" has no verb, and the obvious answer is wrong.**
*Rule:* it is a copy at the row level and a reference at the byte level — a new album row in the target with its own layout paths, hardlinked where possible, sharing the blobs. Consequence to state plainly: a retag in the source does **not** propagate, because the target row is its own and the target collection has its own tagging policy. Call the gesture "Copy (a snapshot)" in the UI, never "Add," because the snapshot semantics are the surprising half.
*Rejected:* a true reference (one row, two collections). It breaks album identity as *(MBID, collection)*, breaks per-collection tagging policy, and reintroduces two writers for one row.

**F11 — MEDIUM. Hand-over's layout fork is unstated.** "Applies its own layout if the admin chooses" leaves a server-written collection whose paths follow no live template — a permanent special case for the projector and every later re-layout.
*Rule:* hand-over always re-lays to the target collection's own template. "Keep my layout" is expressed by the collection row carrying that template, which already travels with the collection. There is no exception to the template; there is only which template the row holds.

**F12 — MEDIUM. A second laptop with an overlapping local library.** The writer is a named binary *instance*, not a person, so laptop B cannot write laptop A's mirror. Its options are mirror-as-new, contribute-to-server, or check out A's mirror read-only.
*Rule:* say this in the spec, and have the TUI detect the same-member case and lead with contribute. Two near-identical mirror collections is the worst browsing outcome, and it is the one a naive user will pick.

**F13 — LOW. Owner stays one person; add transfer, not co-owners.** A role list is a permission system, and admins can already edit any collection's shape, which covers "someone else needs to fix this." Add: the owner may offer ownership to another member, who must **accept** (ownership carries the F7 quota liability, so it cannot be pushed onto someone); an admin may reassign unilaterally.

**F14 — LOW. Admission and notification.** The routing rule admits automatically; an owner review queue is opt-in via the settings table's review-queue knob. Rejection after the fact is just the F4 removal verb — no second mechanism. Notify the owner per batch, never per album.

---

## (b) Decisions only Geoff can make

1. **Does "private" ship at 1.0 at all?** *Recommended default: no.* One visibility (group-visible) at launch; the column lands at rung 6, the affordance at 2.0 when the built-in projector can enforce it. Reason: under the Navidrome driver it is unenforceable through any API, and a half-working privacy control is worse than none.
2. **If private ships, what is the default for a member-created collection?** *Recommended: group-visible.* The product's shape is a group of family or friends; private is the deliberate act.
3. **Default contribution setting on a new member-created server collection?** *Recommended: closed.* The quota liability and the curation burden are the owner's.
4. **May members create laptop-written mirrors without approval?** *Recommended: admin approves the first mirror per member.* A mirror consumes disk the admin did not choose, album by album, with no intake verdict gate.
5. **A departed member's playlists that others still use?** *Recommended: transfer to the admin.* Deleting them is a visible loss to people who did not leave.
6. **Is hard-deleting a collection's files ever a member-reachable action?** *Recommended: no — retire is the member verb, purge is admin-only.*
7. **Quota accounting for a shared blob: charge every referencing collection?** *Recommended: yes* (see F7).

---

Sources: [Navidrome Multi-Library Support](https://www.navidrome.org/docs/usage/features/multi-library/), [Custom libraries per user (discussion #2452)](https://github.com/navidrome/navidrome/discussions/2452)
