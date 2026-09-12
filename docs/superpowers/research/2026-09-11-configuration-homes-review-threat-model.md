# Configuration homes spec: threat model review

Adversarial review record, 2026-09-11, fresh-context `claude-opus-5` reviewer against
`docs/superpowers/specs/2026-09-11-configuration-homes-design.md` revision 2. Findings are folded
into revision 3 of the spec; this file is the citation record. Reviewer output verbatim below.

---

Read the design spec, the rung 5 threat model and token sections, and the 2.0 forward hooks. No implementation exists yet for templates or collections, so every finding below is a design-time change. Citations are `docs/superpowers/specs/2026-09-11-configuration-homes-design.md:line` unless noted.

# (a) Spec changes

**Critical — layout templates are a filesystem-write primitive (`:124`, `:180`, `:206`, `:241`).** The spec grants an owner (any member) a free path template and says nothing about how it renders. Exploit: owner sets `../../../../var/lib/dubplate/$album` or lets a tag carry `../`; the pipeline's atomic writer renames FLAC bytes over `roster.json`, the epoch file, or the SQLite store — all inside the unit's `ReadWritePaths` — and the member is now admin. Fix, in the spec: a template yields a *sequence of components*, never a string; an expanded field can never contribute a separator, `.`, `..`, NUL, or a control character; each component is sanitized by hook 12's portable-layout sanitizer, NFC-normalized and case-folded for uniqueness; the join is checked for containment with `filepath.Rel`; and every library write goes through `os.Root` (Go 1.24 `openat2`-style, refuses escape and symlink traversal by construction) with `O_EXCL|O_NOFOLLOW`. Cite OWASP *File Upload* and *Path Traversal* prevention.

**Critical — the template language must be declared non-evaluating (`:124`, `:180`).** If it is `text/template`, a member-supplied template calls methods on whatever value is passed and reads any exported field the renderer hands it; beets-style `%if{}`/`%aunique{}` add expansion DoS. Fix: a closed grammar of field substitutions plus a fixed whitelist of pure formatters, parsed to an AST and validated at row-write time (unknown field = typed error), with output length and expansion-count caps. State that `text/template`, `html/template`, and any Turing-complete engine are declined by name.

**Critical — a mirror push is untrusted input, and the spec treats it as data (`:187`, `:376`).** A malicious local binary controls every mirrored row: paths, tags, art bytes, and the content hash. Claiming another album's audio MD5 poisons dedup and hardlink sharing (`:266`, `:283`), so a mirror can adopt or free another collection's bytes; crafted tags flow into the 2.0 FTS index, the Subsonic XML/JSON, and `download`'s `.m3u8`; crafted art hits `x/image` decode at cover-art resize as a decompression bomb. Fix: the server recomputes every fact it will trust — decode the FLAC header and compute the audio MD5 itself, never accept a client hash; re-derive the layout path server-side from the collection row's template rather than accepting the client's path; validate tags against the policy and the sanitizer; decode art under pixel and byte caps and re-encode. Add a sentence: *a mirror is a different transport, not a lower trust level.*

**High — writer identity is unbound (`:125`, `:342`).** "A named local binary" names nothing. Fix: the writer is a server-minted device credential, ≥128 bits from `crypto/rand`, stored only as a SHA-256 of the raw bytes, bound to (surrogate user id, device id), presented as a bearer credential; authorization is `collection.writer == token.device AND collection.owner == token.member`, never a client-asserted writer field. Add a per-collection monotonic **mirror generation** so a captured push cannot replay an older state (undoing an owner's or admin's removal), and a **writer epoch** that seize bumps.

**High — seize races an in-flight push (`:342`).** Admin seizes, re-lays files, and the laptop's resumable push — authenticated minutes earlier — lands rows and files over the server's fresh ones. Fix: seize is one transaction bumping the writer epoch; every row commit and every file link-in re-checks the epoch inside that transaction and aborts the whole push with a typed `writer changed` error; staged tus uploads are discarded, never adopted.

**High — collection on-disk roots must be server-allocated and disjointness enforced (`:185`, `:206`).** If a member influences where a collection lives, two collections overlap or nest, one writer overwrites another's files, and a scanning driver maps two music folders onto one tree (defeating private). Fix: the server derives the root from the collection id; a prefix check against all existing roots runs at row write; the driver's subtree constraint is validated, not assumed.

**High — template collisions destroy data (`:180`).** A template omitting disc/track, or two tracks with identical tags, renders two tracks to one path and the atomic rename silently eats the first. In a contribution-open collection this is cross-member destruction. Fix: static injectivity validation on the template's field set; `O_EXCL` at create with a deterministic disambiguation suffix; never rename over a path owned by a different catalog row; a collision rejects the album (album-is-the-verdict-unit).

**High — the overlap report is a membership oracle over private collections (`:362`, `:370`).** Once private exists, a member hashes candidate albums (or submits public MBIDs) and learns exactly what the server holds in collections they cannot see; the intake duplicate refusal is the same oracle. Fix: overlap and duplicate refusal answer only over collections visible to the caller; a cross-visibility duplicate is allowed to exist; bulk hash queries are rate-limited and logged.

**High — settings that are dangerous live (`:51`–`:63`, `:409`).** SMTP host/port/from repointed at an attacker's relay silently steals every future magic link — full takeover of the whole group. Streamer driver URL repointed aims the SSO proxy and roster sync at an attacker. Backup targets exfiltrate store and library. Monitoring ping URLs are a scheduled SSRF at `169.254.169.254` or the LAN. Lifetimes set to a year defeat rung 5's caps; a match cutoff of 0 corrupts the library. Fix: a `sensitive` class in the registry; a sensitive write requires fresh re-authentication (a magic link consumed within N minutes), not the cookie a shared laptop holds; audit record with actor, channel, old and new (redacted for secrets); notification to the admin address on change; outbound-URL settings validated against loopback, link-local, and private ranges. State that a settings key may never hold a credential, and that changing SMTP host invalidates the stored credential.

**High — private must be unsettable, not merely unenforced (`:144`, `:407`).** "Fails closed" at read time is the wrong layer: rung 14's API should refuse to set `visibility=private` while the active driver does not declare per-user folder access, so the UI never offers a promise the 1.0 driver cannot keep.

**Medium — quota does not bound the disk (`:316`–`:319`).** Charge-every-referencer is a fairness counter; with hardlinks it over-counts, and with per-collection tag divergence (`:278`) it under-counts nothing but still lets the group total drift from real bytes. Fix: two counters — per-member fairness (every referencer) and group real bytes (inode once) — plus a hard free-space floor checked before every write, because a full disk wedges the SQLite WAL and takes the server down. Copy (`:292`) bypasses intake, so its quota check belongs inside the copy transaction.

**Medium — the sweep races copy (`:299`).** Link-count check sees 1 while a copy is mid-hardlink; the inode is unlinked under the new row. Fix: the sweep claims the inode in the same store transaction, or re-checks under lock.

**Medium — machine-settings file (`:26`, `:385`).** Fix in spec: non-secrets in `$XDG_CONFIG_HOME/dubplate/`, the token in a separate `credentials.toml` under `$XDG_STATE_HOME`, created `O_CREAT|O_EXCL` at 0600 in a 0700 dir, load refused if group/other bits are set (ssh's rule); OS keyring preferred where present; redacted from `doctor`, logs, and any support bundle. State what a stolen file yields: member access plus writer control of that member's mirror — hence device tokens carry an expiry, are listed with last-seen address in the member UI, and are individually revocable.

**Medium — tagging policy can reach the identity set indirectly (`:200`).** A policy that *renames* tag X to `ALBUMARTIST` writes the identity set without touching it by name. Fix: the invariant is on the output tag set, not the operation; also cap tag count, value length, and picture size (a FLAC metadata bomb).

**Medium — the admin API must not be one mount serving both cookie and bearer (`:75`, `:409`).** Rung 5's CSRF rule exempts bearer paths via `AddInsecureBypassPattern`, and a bypassed path may not read the session cookie. The TUI wants bearer; the 2.0 web page is cookie-authenticated. Say in this spec: two mounts, or the web page acquires a bearer token; never one mount doing both, or the admin surface silently loses origin protection.

**Low — opaque ids (`:248`).** Sequential album/collection ids let a member enumerate across a visibility boundary even behind a uniform error. Prefer random ids.

# (b) Tasks the rung plans must carry

**Rung 6 (schema/store).** Device-credential table (hash only, member, device, created, last-seen address, expiry, revoked-at); `writer_epoch` and `mirror_generation` on the collection row beside `writer`/`owner`/`visibility` (`:404`); server-allocated `root_path` with a uniqueness/disjointness constraint; two quota counters; a `sensitive` flag in the settings registry; settings audit table (actor, channel, key, old/new, redacted); quota check and link-in inside one transaction (hook 9's transaction model).

**Rung 5 (identity).** Extend the revocation seam so a device/writer credential store subscribes to the same epoch read (the spec's S20:245 seam); a removal checklist test asserting *all* of: epoch bump, app passwords, device tokens, writer epoch freeze, in-flight push abort, collection reassignment, streamer deprovision. Add a last-admin lockout guard, atomic against concurrent self-demotion.

**Rung 12 (driver seam).** Capability report carries shape constraints and per-user folder access (`:214`, `:406`); verify the pinned Navidrome's art filenames and read tags against the build, not the spec; a driver without per-user access makes private unsettable.

**Rung 14 (admin spine).** Re-auth gate on sensitive settings; SSRF validation on URL-valued settings; seize as one epoch-bumping transaction with a typed error for the returning laptop; member removal as the checklist above; audit on every write.

**Rung 7/9 (pipeline).** Template parser and injectivity validator; component sanitizer; `os.Root`-confined writes; intake caps (upload size, per-album file count, decompressed total, compression ratio, path depth, metadata/picture size); server-side recomputation of the audio MD5.

**2.0 projector.** One `visibleCollections(userID)` predicate joined in a single accessor no handler may bypass, with a table-driven test asserting identical "not found" for invisible-and-nonexistent across `getAlbum`, `getCoverArt`, `search3`, `getPlaylist` (`:307` makes a playlist a read oracle), `getRandomSongs`, `getAlbumList2`, `download`, `createShare`, and star/scrobble.

# (c) Accepted risks to state in the spec

1. Private is never confidentiality from the admin (already at `:146`) — keep, and extend: nor from anyone with disk, backup, or log access.
2. A cloned laptop disk yields a working writer token; divergence detection (two pushes at one generation → freeze pending owner re-attestation) is the mitigation, not prevention.
3. Mirrors bypass per-album intake verdicts by design (`:376`); admin approval of the first mirror plus quota is the whole bound.
4. Charge-every-referencer over-counts, deliberately (`:318`).
5. A `move` loses stars, playlist entries, and history (`:299`).
6. Hardlink sharing assumes one filesystem and a correct atomic-writer discipline; a future non-renaming writer breaks collection isolation silently.
7. An admin remains fully trusted; re-auth on sensitive settings raises the bar on a stolen session but does not defeat a live-compromised admin browser.

**Verdict:** the configuration model is sound, but as written it hands members two unbounded write primitives — a path template and an unverified mirror push — either of which is a server takeover today; adopt the component-sequence template rule, the server-recomputes-everything mirror rule, the writer-epoch seize transaction, and the sensitive-settings re-auth class before rung 6 freezes the schema.
