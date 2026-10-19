# Adversarial review: dubplate-server 2.0 design (revision 1)

Fresh-context review of `docs/specs/2026-09-07-streamer-2-0-design.md` (revision 1,
2026-09-07) against the greenfield build spec, CLAUDE.md, and ROADMAP. Protocol claims checked
against opensubsonic.netlify.app. Line citations are to the spec under review unless prefixed `BS:`
(greenfield build spec).

## Verdict

The decisions are sound and the rejected alternatives are recorded well enough that a later reader
will not reopen them. The architecture is right: a projection over the pipeline's catalog, one door
for admission, a content-keyed cache the pipeline can invalidate exactly. Nothing in the Decisions
section needs retaking. The spec fails at the job it names as its own, "the forward hooks the
current build owes," and that is where the damage lands, because those hooks are consumed by rungs
about to be planned. Two hooks name the wrong rung; four things rung 6 cannot infer never reach the
list, and rung 6 is next in the spine phase; one protocol claim is false as written; and one
sentence trims rung 12 in a document that opens by promising it amends nothing. Five blockers,
fifteen majors, fourteen minors, all cheap to fix in this document and none reopening a decision.
Fix them before the study rung opens and before any rung 6 plan is authored.

Ranks: **blocker** — a later rung builds the wrong thing. **major** — 2.0 pays later, or a security
or correctness gap a slice plan cannot be trusted to invent.

## Blockers

**B1. The spec amends the build spec while disclaiming that it does.** *Scope, consistency.* Lines
10-15 ("does not amend the greenfield build spec ... or the 17-rung ladder") against 93-94 ("Rung 12
is trimmed ... its theming polish is dropped"). That amends `BS:733-758`, and the blast radius
reaches further: `BS:640-686` gives rung 7 a two-axis theming model whose structural axis is the
system MATCH (navidrome, jellyfin, or dubplate's own), estate config `match navidrome`, plus a
standing rule that the roster grows a matching theme per shipped streamer and per streamer factory
default. With 2.0 ratified, the Navidrome-matched structural theme is throwaway for this estate, and
rung 7 lands long before 2.0. A planner reading only the build spec builds all of it. **Change:**
land the trim as a dated erratum inside the build spec's rung 12 entry and ROADMAP, or delete the
sentence; either way add a hook telling rung 7 that dubplate's own match becomes the estate default
and streamer-matched themes are driver niceties, and re-price rung 7's theming against that.

**B2. Hooks 2 and 3 name rung 7 as "admit"; admission is rung 9.** *Forward hooks.* Lines 285-289:
folded sort key "computed at rung 7's admit", art blob "rung 6 schema, rung 7 intake". Rung 7 is tus
transport, gate policy, roster loading, journey records, uploader client (`BS:611-620`). Admission,
tagging, layout, and collection routing are rung 9 (`BS:716-729`). The folded sort key derives from
applied tags and the admitted cover is chosen by tagging; neither exists at rung 7. As written, rung
7 looks for a hook it cannot land and rung 9 never weighs it. **Change:** re-key both to rung 9,
keeping the rung 6 schema half, and state separately that hook 1's content hash is computed at rung
7's gate on the uploaded bytes, so it has an author.

**B3. Rung 6's hook list omits four things rung 6 cannot infer.** *Forward hooks, operations, data
model.* Lines 283-298 against 137-138, 213-218. Four requirements sit in the Architecture body and
never reach the hook list, so the rung 6 plan will not see them. (1) **A transaction model across
catalog events.** Line 137-138 requires the projector's delta to apply "in the same transaction that
commits the event", but rung 6's ops seam is an interface over JSON request and response types
(`BS:597-604`) with nothing a later projector can join; retrofitting that is a seam rewrite. (2) **A
stable per-user identity row.** Line 213 keys four tables "by roster user", but the roster is a
validated file loaded at rung 7 (`BS:613`), not a table, and its handle is renameable; a rename
orphans every starred row, playlist, queue, and history event. Rung 6 must mint a durable surrogate
user id and rung 5 must resolve identities to it. (3) **The SQLite concurrency decision** — one
writer, many readers, WAL, `busy_timeout`, serialization on the modernc driver — because 2.0 turns
the DB from near-quiescent into written-on-every-play (line 193) while the reconcile walk reads. (4)
**Backup consistency of a live database:** `BS:607-608` gives rung 6 snapshot backup to rclone
remotes, and copying a live WAL SQLite file byte-for-byte yields a torn snapshot, so it must go
through `VACUUM INTO` or the online backup API. **Change:** add all four against rung 6. Item 4 is
worth fixing there regardless of 2.0.

**B4. "The spec's not-implemented error" does not exist.** *Protocol.* Lines 229-230. The
OpenSubsonic error table (opensubsonic.netlify.app/docs/responses/error/) defines 0, 10, 20, 30, 40,
41, 42, 43, 44, 50, 60, 70 and no not-implemented code; 30 is "Server must upgrade", not
"unimplemented". Navidrome carries only the original eight codes
(`server/subsonic/responses/errors.go`: 0, 10, 20, 30, 40, 50, 60, 70) and simply does not register
handlers for endpoints it lacks, so an unimplemented call gets a bare 404 rather than a
`subsonic-response` at all. The practical answer also differs per endpoint: many clients treat any
error as fatal and disable a whole screen where an empty successful list degrades gracefully.
**Change:** state the convention explicitly — code 0 with a message inside a well-formed
`subsonic-response`, never a bare 404 — and make the set that must instead return an empty success a
study-rung deliverable measured against the client roster. Codes 41 through 44 arrive with the
apiKey extension.

**B5. Entity ids re-key on retag and the per-user rows dangle.** *Data model.* Lines 144-147 ("ids
survive a rebuild and a starred list never dangles") with 218 ("a catalog removal cascades to the
rows that named it"). The claim holds for a rebuild and fails for a retag, which rung 9 ships as a
launch-bar admin verb (`BS:722`). A retag correcting a release MBID changes the album id and every
song id under it; a retag of an untagged track changes its content hash. Neither is a removal, so no
cascade fires and the rows survive pointing at ids nothing answers. Starred, playlists, the saved
queue, and play history all break silently, and play history is append-only so re-starring cannot
repair it. **Change:** entity ids are opaque server-issued surrogates minted once per entity;
release MBID and content hash are indexed alternate keys, never the id; a retag updates the
alternate key on the existing row and per-user rows are untouched. State the one residual case (a
retag moving tracks between albums, merging or splitting an entity) and what happens to the per-user
rows there.

## Majors

**M1. Plaintext at rest is a consequence of a choice, not a fact.** *Security, protocol.* Lines
62-66 and 198-201. The reasoning is correct given the one-secret decision, but that decision is what
forces recoverable secrets, and the alternative is documented: the `apiKeyAuthentication` extension
recommends deprecating salt/token auth, and a server dropping it "must return error 41"
(opensubsonic.netlify.app/docs/extensions/apikeyauth/). Because apiKey verification is a lookup
rather than a salted challenge, a device row can then hold only a hash; whether that is viable is
what the client roster measures (M15 on why the oracle cannot). The encryption is also
underspecified: no AEAD, no per-row nonce, no domain-separation label (mandated for every MAC input
at `BS:453-457`), and no rotation, while both existing secrets carry a `_PREVIOUS` overlap
(`BS:456`) — so rotating this key is an estate-wide re-pair of every device. **Change:** state the
threat model in one sentence (encryption protects the R2 snapshot and a stolen disk, not a box
compromise, since the key sits in the box's env file); name the AEAD, nonce, and label; add a
`_PREVIOUS` overlap with a re-encrypt pass; record the hash-only path as the alternative the roster
decides.

**M2. Offboarding does not revoke app passwords.** *Security.* Lines 204-206 list mint, first use,
and revoke as events but never tie revocation to the roster. `BS:445-449` states today's residual
plainly: the per-email valid-after epoch revokes the dubplate session, "streamer credentials survive
it, because `/rest/*` bypasses serve by design", with a manual runbook step. With an in-house
streamer that residual stops being structural; as written, a removed family member keeps playing
until someone remembers the runbook. **Change:** state that the valid-after epoch invalidates every
app password for that identity, and add a rung 5 hook — the identity layer exposes a revocation seam
a credential store can subscribe to, not a session-only epoch.

**M3. `/rest/*` and `/share/*` collide with the one-owned-prefix rule.** *Security, consistency.*
Lines 224-225 and 207-208. `BS:398-402` binds serve from rung 1: dubplate's entire HTTP surface
lives under exactly one configured prefix and the streamer owns everything else. `BS:359-366` names
`/rest/*` and `/share/*` as bypass paths serve strips headers on but does not authenticate, and
`BS:471-475` records `/rest/*` as an anonymous surface whose auth is the streamer's own. In 2.0 both
are dubplate's own code and that residual becomes dubplate's to answer. **Change:** state that 2.0
moves both inside the owned surface, that the ingress strip set still applies, and that the
one-prefix rule is amended (Subsonic clients cannot be told to use a prefix, so it must be). Say it
here so rung 12's proxy work is not built against an assumption 2.0 reverses.

**M4. The play-history threshold measures bytes served, not listening.** *Data model.* Lines
193-194. On the direct path (175-177) the server serves ranges from a static file and most clients
buffer the whole track ahead of playback, so the threshold trips within seconds whether or not
anyone listened; "whichever comes first" then guarantees the byte heuristic beats the honest signal
and play counts measure prefetch. It also diverges from the oracle by construction: Navidrome "does
not mark songs as played by calls to `stream`, only when `scrobble` is called with
`submission=true`", and its 50%-or-four-minute threshold belongs to the `playbackReport` extension,
not to `scrobble`. OpenSubsonic specifies no threshold at all. **Change:** `scrobble` with
submission=true is the primary record, matching the oracle; the byte heuristic is a fallback used
only when no scrobble arrives within the track's duration, recording its provenance in the table's
"how reported" column. State the dedupe rule for a client that both buffers and scrobbles, and note
`submission=false` as now-playing only.

**M5. Whole-list replace does not match `updatePlaylist`.** *Protocol, data model.* Line
215. Subsonic exposes two verbs: `createPlaylist` with an existing id replaces, while
    `updatePlaylist` is a delta API (`songIdToAdd`, `songIndexToRemove`, name, comment).
    Implementing update by read-modify-replace makes every concurrent client a lost update — two
    phones each remove one track and the second write silently restores the first's. **Change:**
    apply updatePlaylist's deltas inside the transaction against the stored order, never against a
    client-supplied list; carry a monotonic version column on the header so a stale replace is
    detectable; name the conflict behavior.

**M6. No rule that ids are opaque and paths never come from the client.** *Security.* Lines 147-148
and 175-176. Path traversal through `getMusicDirectory` and `stream` is the classic Subsonic-server
bug class and it arrives exactly when a folder view is built from paths; the spec never states the
invariant. **Change:** every id on the wire is an opaque server-issued token, never an encoded path;
folder ids are minted rows in the projection; `stream`, `download`, and `getCoverArt` resolve a path
only by id lookup in the catalog, and the resolved path is verified under the configured library
root, byte-preserving per `BS:604-606`. State that the folder view is a projection table and never a
filesystem walk, or the "catalog is the index" decision leaks.

**M7. The transcode argument templates need a stated safety shape.** *Security.* Lines 180-182. "No
user-supplied argument strings" is good but the request still supplies a format and a bitrate cap
(178-180) that interpolate somewhere, and templates are operator config, so they are an injection
surface even without a client. **Change:** format selects a whole template by name from the
configured set, never a substituted string; bitrate is an integer validated against an allowlist;
templates are argv slices, never shell strings; no template may contain `-i` or an output path (the
server supplies both); seek offsets are validated numerics; config load rejects a template naming
anything but the declared placeholders.

**M8. The transcode path's failure and range behavior is unstated.** *Operations.* Lines 184-192.
Three gaps. A filling entry has no known length, so a Range request on a miss cannot be answered the
way 187 answers a hit, and most clients seek on transcoded streams. An ffmpeg that dies midway
leaves a partial entry that, once the writer closes, is indistinguishable from a complete one and
serves as a truncated track forever. Requests past the concurrency knob "queue in arrival order"
(192) with no wait cap, so on a two-core box the third listener holds a connection until their
client times out. **Change:** state the miss-path range answer; state that a non-zero ffmpeg exit or
a cancelled context invalidates and deletes the entry rather than committing it, with attached
readers erroring; give the queue a bounded wait and a named behavior past it (503 with
`Retry-After`, or fall back to the direct path).

**M9. Shares are designed but absent from the launch surface, and unsecured.** *Security, scope.*
Line 207-208 against 225-229. Shares appear on neither the launch endpoint list nor the 2.1 list, so
the document contradicts itself about whether they ship — and if they ship, the public path is the
estate's only unauthenticated surface with no rules at all. **Change:** put shares on one list. If
launch, name `createShare`/`updateShare`/`deleteShare`/`getShares` and state token entropy (≥128
bits from `crypto/rand`), constant-time comparison, server-side expiry, per-address rate limiting,
resolution scoped to the share's own entity ids, and that a share cannot drive an unbounded
transcode (direct-only, or a per-share encode budget).

**M10. The album identity rule has no answer for real libraries.** *Data model.* Lines 141-144.
Three uncovered cases. An album admitted with no release MBID: the decision queue and custody policy
(`BS:759-770`) exist because not every import matches, and the content-hash fallback (144-146) is
per-track and does not group an album. The same release imported twice into one collection, a re-rip
at different quality being ordinary, collapses into one album with doubled tracks. And the same
release in two collections is two albums by design (accepted), but a client browsing all music
folders sees it twice under different ids. **Change:** name the grouping key for a no-MBID album
(the pipeline's own import-session grouping is the honest answer, since intake already treats the
album as the verdict unit); state what a second import of the same release into one collection does
— reject, replace, or a disambiguated second album — as a pipeline rule rather than a projector
rule; record the cross-collection duplicate as accepted.

**M11. The web surfaces name no stack.** *Web surface.* Lines 62-64 (device page) and 207-208 (share
management). CLAUDE.md's rule is Svelte 5 runes, standalone with no SvelteKit, DaisyUI 5, Vite to
static assets embedded via `go:embed`, with `svelte-conventions` and `ts-conventions` mandatory; the
spec names none of it. A real seam question sits behind that: rung 5's login and consumption pages
are Go-templated with no JavaScript (`BS:583-585`) while `web/` is the Svelte tree rebuilt at rung 7
(`BS:325`, `BS:620`), and a device page reached straight from a magic link sits between them.
**Change:** state the stack and which tree each surface joins. A device page is a form plus a
shown-once secret and fits rung 5's Go-templated set; share management is list-and-revoke and fits
`web/`. State the flow: magic link → session → device page under the owned prefix.

**M12. The driver-seam hook assumes the driver is a separate process.** *Forward hooks.* Lines
293-295 against `BS:733-748`. Rung 12's seam is scan trigger, scan status, album deep link,
capability reporting, and driver-declared post-login destination, plus streamer address semantics
("the proxy target under SSO, the public base URL without it") and a reverse proxy. An in-process
streamer has no address, needs no proxy, has no external scan, and its post-login destination is a
dubplate route; "the seam stays universal" does not tell rung 12 which of those must be optional.
**Change:** restate concretely — address, proxy, and provisioning choreography are per-driver
capabilities a driver may decline to declare, never core assumptions; scan trigger and status must
be satisfiable by a driver that owns the catalog and never scans; capability reporting is the
mechanism that makes both true.

**M13. The ladder carries no magnitude bands and no ceilings.** *Scope.* Lines 254-274. Every other
ladder here bands its rungs S/M/L and every pass header carries a ceiling and a checkpoint interval.
Two rungs are plainly L: rung 3 (projection plus both browsing views plus the id space) and rung 5
(streaming, cache, concurrency, ranges). Eight unbanded rungs read as eight passes and will not be.
**Change:** band each rung, mark 3 and 5 as L splitting at their own planning, per `BS:487-500`.

**M14. ReplayGain appears on neither list.** *Scope.* Lines 328-333 against 225-229. `BS:729-731`
makes ReplayGain a probe-backed rung 11 built-in, so the tags may exist at launch; OpenSubsonic
defines `replayGain` on the song response and clients use it for volume normalization, so a family
that had it under Navidrome loses it silently at the flip. **Change:** one line — if the pipeline
writes ReplayGain tags, song responses carry the `replayGain` object at launch (a projection of tags
already read); if not, say so and put it on the 2.1 list.

**M15. The oracle is blind on two launch surfaces.** *Evidence.* Lines 101-108 ("a differ compares
responses field by field") and 145-148. Navidrome, at current master, advertises `transcodeOffset,
formPost, songLyrics, indexBasedQueue, transcoding, playbackReport, topSongsByArtistId` and
implements neither `apiKeyAuthentication` nor codes 41-44, so the modern auth path this spec builds
on has no oracle at all. Navidrome also does not do folder browsing: it answers `getIndexes` and
`getMusicDirectory` with a tree synthesized from ID3 in the form `/Artist/Album/01 - Song.mp3`, not
from real paths. Since 2.0 derives its folder view from the paths the pipeline chose (147-148),
every folder response diverges by construction and a field-by-field differ buries the ledger in
noise. **Change:** declare both as oracle gaps at the harness rung, adjudicated once as classes
rather than per case — the auth path graded against the specification and the client roster only,
the folder view graded on shape rather than content or dropped (minor 9). Say plainly that Navidrome
is not a compatibility baseline for apiKey.

## Minors

1. **Line 226, the system group.** Name `ping`, `getLicense`, `getStarred2`, and `getScanStatus`;
   clients hard-fail setup on a missing `getLicense`. `getOpenSubsonicExtensions` **must** be
   reachable unauthenticated — the surface's one carve-out, easy to lose behind a blanket auth
   middleware.
2. **Line 202, refusing `p`.** Spec-legal (`p` is optional when `t` and `s` are given) and low-risk:
   its documented dependents are LDAP- and PAM-backed servers, which dubplate is not. Keep the
   refusal, return error 42, say so.
3. **Lines 51-55, the reconcile walk.** State its cost shape: size and mtime first, hash only on
   suspicion, a configured schedule and an I/O bound. Re-hashing the library on a timer is the most
   expensive thing 2.0 would do on this box.
4. **Lines 153-155, taglib as art fallback.** Forbid per-request wazero instantiation: embedded
   pictures are extracted once at admit (hook 3); serve never touches taglib.
5. **Lines 71-76, cache sizing.** A ratio with no denominator here and no statement of the box's
   free space. Name the default cap and the doctor row that warns before the volume fills; art
   variants share the same cache (line 158).
6. **Line 203, rate limits.** Point at the settled rule (`BS:461-470`) — real client IP from the
   trusted-proxy pair when proxy mode is on, `RemoteAddr` when not, `doctor` reporting which is live
   — rather than re-deriving it.
7. **Lines 289-291, hook 4.** Conditional on "if any listening record lands before 2.0", and nothing
   does, so it is a no-op. Delete it, or restate it unconditionally as a rung 6 schema rule for any
   event-shaped table.
8. **Lines 291-293, hook 5.** The build spec already has the field (`BS:477-482`) as an *optional*
   `streamer_username` defaulting to the inbox directory name; 2.0 needs it required, unique,
   stable. Reconcile the two rather than restating one.
9. **Lines 335-351, open items.** Two owner calls the study rung needs are absent: whether folder
   browsing ships at all (it doubles the browse surface and the id space, modern clients use ID3,
   and M15 shows the oracle cannot grade it), and whether salted-token auth is dropped for
   apiKey-only (M1).
10. **Lines 165-171, search.** "Prefix match, then FTS5 rank" is two queries, not one ranking; an
    FTS5 trigram index does not order by prefix, so say the prefix pass is a separate indexed lookup
    unioned ahead. Same section: `search3` **must** accept an empty query and return everything for
    offline sync — trivially broken by naive validation, and paging must survive a whole-library
    result.
11. **Lines 224-231, response envelope.** Every `subsonic-response` must carry `openSubsonic: true`,
    `serverVersion`, and `type` beside `status` and `version` — the fields clients switch
    OpenSubsonic behavior on. Name them with the typed structs.
12. **Lines 156-158, cover art ids.** Under OpenSubsonic `getCoverArt`'s `id` is a dedicated
    coverArt id, not a song, album, or artist id; legacy clients pass the entity id anyway. State
    which dubplate accepts, reconciled with M6.
13. **Line 147, collections as music folders.** `musicFolderId` is an endpoint-by-endpoint optional
    filter, absent from single-entity fetches, not a blanket rule. List the endpoints that honor it.
14. **Line 216, the play queue.** Name the edge cases: a call with no `id` clears the saved queue,
    `current` is required unless `id` is empty, absent `position` means zero, and the response
    carries `changed` and `changedBy`.

## What holds

So the study rung does not re-litigate settled ground: the catalog-as-index decision survives the
folder view (M6 is about ids, not the decision); the content-keyed cache with delete-on-retag
invalidation is the document's strongest idea; the evidence-first build shape is the machinery rungs
2 through 4 already proved here; and deferring scrobbling and ratings to 2.1 is right.
