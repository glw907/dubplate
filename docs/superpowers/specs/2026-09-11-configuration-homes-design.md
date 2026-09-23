# Configuration homes: where each setting lives

Design spec. Revision 3, 2026-09-12. Status: ratified in conversation, awaiting Geoff's read
of this file.

Revision 1 is dated 2026-09-11, revisions 2 and 3 fold five adversarial reviews
(competitors and demand; collection interactions; technical design; threat model;
consistency against the ratified documents), and the review records live under
`docs/superpowers/research/2026-09-11-configuration-homes-review-*.md`. Every ruling in this
file is dated, and the rulings dated 2026-09-11 and 2026-09-12 are Geoff's.

The spec settles where a configuration value lives across the three surfaces dubplate will
have: the server's own configuration, the 2.0 web app, and the bubbletea TUI a member runs
on their own machine. It defines no implementation work. Ruling 6 of the greenfield rulings
holds: no TUI work until the server is fully formed. The spec routes its hooks into the
rung 6, 7, 9, 12, 14, and 16 plans and into the 2.0 streamer spec's forward-hooks list, and
it supersedes named passages of four ratified documents, listed under Supersessions.

## The rule

Every knob has exactly one home, and the home is chosen by who owns the truth. A web page
or a TUI is a view over a home. Neither holds a second copy. No knob has two homes, so no
value is reconciled between homes. Where a local and a server store hold the same kind of
knob, the TUI switches between them instead of merging them, and the Profiles section gives
the rule. Configuration appears to live on both the client and the server. In fact there
are four kinds of knob, and the table gives each one a home.

| Home | Owner of the truth | Examples | How the TUI reaches it |
|---|---|---|---|
| Server wall, the env | the operator | state dir, library root, bind address, SMTP, secrets | never; not shown |
| Server settings, a table in the store | the admin | match thresholds, session lifetime, intake limits, collection bounds, per-member overrides | the rung 14 admin API, for admins only |
| Member preferences, per-member rows in the store | the member | default collection, notification choices, theme, language, ReplayGain mode | the member API, read on connect, written on change |
| Machine settings, a local file | the member, per machine | audio device, keybindings, watch folders, the list of connection profiles | edited in place; never sent to the server |

A member with two laptops sees the same preferences and the same theme on both, because
both live on the server, and a different audio device on each, because devices live on the
machine. Editing a server-side value from the TUI is an API write, so the web page shows it
on its next load with no reconciliation step. Concurrent edits to a scalar setting are
last-write-wins on the server. Playlists, starred items, play history, and the saved queue
are not preferences. They are the 2.0 streamer's own per-user tables, with the delta
semantics and the version column that spec gives them. The TUI holds no local copy of a
server value beyond what the current screen paints, so nothing goes stale offline, and a
disconnected TUI can still change its own machine settings.

## The server's own knobs

Today the server reads an env wall and nothing else: no file, no reload, no write-back, and
no metadata on a key beyond its type and default. The wall stays, and its contents stay
what the build spec's config-wall section says (Geoff, 2026-09-12). The store takes what
that section calls tunables.

The wall holds the owned, required, no-default keys (the library root, the state directory,
the site name, the public URL, the identity source and its correlated set, and the roster
file), every secret, and every integration class whole, flag-enabled and
required-when-enabled, exactly as rung 1b's registry enforces today: SMTP, the streamer
driver, backup targets and schedule, monitoring, the trusted-proxy pair, and a forward-auth
front. The integration classes stay whole rather than split by convenience. The registry's
rule that a flag makes its correlated config required
is enforced in `Load` and reported by `doctor`, and a class half on the wall and half in
the store would break it. And two of the classes are read by processes other than the
server: the backup schedule is a systemd timer, and the failure pings are sent by a separate
unit with its own env. A setting whose consumer cannot see the store must not live in the
store, or an admin changes it on a page that confirms the change while the timer keeps its
old hour forever.

The store holds every tunable that ships with a default, and everything created after boot.
Tunables are the thresholds: match cutoffs, session and token lifetimes, link expiry, rate
limits, the intake limits (upload size, concurrency, quota defaults), the tag policy
defaults, and the layout defaults. Created-after-boot state is the collections, the members,
the per-member overrides, and the review-queue and notification settings. This is still the
majority of what an admin touches after setup, and nothing on the wall is set more than
once.

This follows the Lidarr and Sonarr split: a small config file for what the process needs
before it can start, and everything else in the database behind the UI. `deploy.sh` keeps
rendering the env from the age store, so a stranger's setup does not change. The registry
keeps the wall's compat discipline: an unknown or superseded env key is still a load error.

Each settings row records where its value came from. A row's source is one of `default`,
`env-seed`, or `admin`, with the seed value, the actor, and a timestamp. A tunable present
in the env at boot behaves in one of three ways. If the row is at its default, the env value
seeds it and the source becomes `env-seed`. If the row is `env-seed` and the env value has
changed, the new value re-seeds it and the change is logged. If the row is `admin` and the
env value differs, the load fails and names both values and the API call that resolves it.
The wall's loud-failure property survives the move, and an operator who rotates a value in
the age store never has it silently ignored.

Every store-homed key also declares how a change applies: immediately, at the next
operation, or at restart. The admin API and `doctor` show the value with its source and its
apply mode. Consumers read settings through an accessor with version invalidation, never
from a long-lived cached struct. A lifetime change states its rule for tokens already
issued.

The registry gains this metadata, a key's home and its apply mode, as a contract change to
rung 1b's landed registry. The build spec carries it as an erratum dated 2026-09-12.

## Two surfaces, one API

The 2.0 web app gains an admin settings surface (Geoff, 2026-09-12): settings, per-member
overrides, and collection management, over the rung 14 API. The TUI's admin section makes
the same calls. The TUI never edits a server file, so the two surfaces have nothing to
disagree about. The organizer spec's remote-TUI mode was already this shape.

The ordering follows from this. The admin API lands at rung 14 and the web admin surface at
2.0, so the TUI reaches server settings before the web page exists.

The admin API is served on two mounts, one cookie-authenticated for the web app and one
bearer-authenticated for the TUI, because rung 5's CSRF rule exempts bearer paths from
origin checks and a bypassed path may not read the session cookie. One mount serving both
would silently lose origin protection on the admin surface.

A knob is a member preference when the server's behavior changes with its value, or when
the value expresses the person rather than the hardware. Sort order the server applies to a
listing, notification choices, and the default collection change what the server does, so
they are preferences. Theme, language, and ReplayGain mode change nothing on the server,
but they follow the person, and Roon profiles, Navidrome, Plex, and Jellyfin all keep them
per user. Audio device and watch folders describe one machine, so they stay local.
Keybindings stay local too (Geoff, 2026-09-12), for a different reason: they are a TUI-only
setting with no counterpart on the web page or in any other client, and they depend on the
terminal. A knob that passes the test in one respect and fails it in another is two knobs.

Some member-scoped values are the admin's to set: a quota, whether the member may upload at
all, whether they may create collections, and an upload concurrency cap. These are
per-member overrides in the settings table, not preferences. Precedence is fixed: an admin
bound wins over a member override, which wins over the settings default.

Some settings are dangerous to change live even though they are not secrets: a lifetime set
to a year defeats rung 5's caps, a match cutoff of zero corrupts the library. The registry
marks these `sensitive`. A sensitive write requires fresh re-authentication (a magic link
consumed within a short window), not the cookie a shared laptop holds. Every settings write
is audited with the actor, the channel, the key, and the old and new values, and a sensitive
change notifies the admin address. A settings key never holds a credential. Any URL-valued
setting is validated against loopback, link-local, and private ranges before it is stored.

## Profiles in the TUI

The TUI is a local organizer first and a client second. Stand-alone, it embeds the engine
and runs the pipeline against a library on the laptop, so it needs thresholds, layout
templates, and collections of its own. Those are the same knobs the server's settings table
holds, and this is the one place two homes hold the same kind of knob.

The TUI switches modes instead of merging the two. It holds connection profiles. A local
profile owns a local state dir with its own store, and the engine reads settings from
there. A server profile names a URL and a device credential, the engine runs on the server,
and the server's settings govern while the local engine settings are not consulted.
Choosing a profile is how a member connects. The machine file holds the profile list and
the settings that apply across every profile. A profile's URL and credential belong to the
profile, not to the machine file's cross-profile settings.

The machine file lives under `$XDG_CONFIG_HOME/dubplate/`. The device credential lives
apart from it, in a file under `$XDG_STATE_HOME/dubplate/` created exclusively at mode 0600
in a 0700 directory, refused at load if group or other bits are set, and preferred into the
OS keyring where one exists. A local profile's store lives under `$XDG_DATA_HOME/dubplate/`.
The credential is redacted from `doctor`, logs, and any support bundle. A stolen credential
file yields member access plus writer control of that member's mirrored collections, so
device credentials carry an expiry, are listed with their last-seen address in the member's
own page, and are individually revocable.

The settings schema is one schema for the tables both stores share (the
settings table, collections, the catalog), defined once in the engine, whether the store is
local or on the box. Per-member overrides and preference tables are server-only. "Take my
local settings to the server" is an explicit one-shot copy that an admin reviews. No
background sync runs. The reverse copy is the same action the other way.

A local store can be years behind the server. Migrations replay forward. A newer local
store talking to an older server is the untested direction, so connect negotiates a
version and refuses with an explanation when the laptop is ahead. Every push between the
two is a versioned row-level API call, never a schema-shaped dump.

## Collections carry their shape

Collection shape is where the knobs are richest and where a local binary and a server both
hold them. A collection is a row in the store, and its shape is data on that row. No global
keys hold it. The row carries the name, the layout template for paths and filenames, the
tagging policy, the admission rules (format allowlist, match-threshold overrides), a
server-allocated on-disk root, and three attributes of control:

- The **writer**, already fixed as 2.0 forward hook 8: exactly one process writes the rows,
  this server or a named local binary. The row also carries a writer epoch, bumped by every
  writer change, and a mirror generation, bumped by every push.
- The **owner**: the person who may change the collection's shape and who decides what goes
  in. A non-null reference to the surrogate user id of hook 5.
- The **visibility**: who may see and stream it.

Writer and owner are independent, which yields three kinds of collection.

| Kind | Writer | Owner | How it comes to be |
|---|---|---|---|
| laptop-written, member-owned | a member's local binary | that member | the mirror verb, below |
| server-written, member-owned | the server | a member | the member creates it on the server and contributes into it through intake |
| server-written, admin-owned | the server | the admin | the group library; the default collection is one of these |

### Visibility

The column lands on the row at rung 6, and the private setting is not offered until 2.0
(Geoff, 2026-09-11, held 2026-09-12). Navidrome, Jellyfin, and Funkwhale each expose a
reader axis: Navidrome grants library access per user, Jellyfin scopes access per library,
and Funkwhale offers private, instance, and public levels. Navidrome's tracker carries the
demand as well, in discussion 2452, where a parent asks that the children's music stay out
of the group's collection. Retrofitting a reader axis after collections exist is a schema
and projector rewrite, and the column costs nothing now.

At 1.0 every collection is group-visible. The rung 12 plan records what the pinned Navidrome
build can do about per-user library access through its native API. That is a verification
task, and the ruling holds whatever it finds. At 2.0 the built-in projector enforces
visibility on every endpoint. The rung 14 API refuses to set private for as long as the
active driver does not declare per-user folder access in its capability report, so the UI
never offers a promise the streamer cannot keep. Private means not listed and not
streamable to other members. It is never confidentiality from the admin, or from anyone
with disk, backup, or log access.

### Control

A collection's shape is edited by its owner or an admin. Contributions are admitted by the
routing rule. A new member-created collection is closed to contribution, and the owner
opens it (Geoff, 2026-09-11), because the quota liability and the curation burden are the
owner's. Contribution exists only on server-written collections. A laptop-written
collection cannot accept others' uploads, by construction, because the server would then
write rows the laptop also writes. The API refuses with a typed error and the TUI does not
draw the control. The writer decides whether contribution is possible at all, and the owner
decides whether it is open.

Any member may create server collections by default, since a group of family or friends is
the product's shape, and the admin can switch creation off per member or for everyone.
Once private exists, a collection starts group-visible unless its owner says otherwise
(Geoff, 2026-09-11). Contributing to the group is the product's point, and private is the
deliberate act. A collection has exactly one owner. A member-to-member transfer is an offer
the recipient must accept, because ownership carries the quota liability. An admin
reassignment does not wait for acceptance; it is the recovery path when the owner is gone.
The admin can also retire a collection. Co-owner lists are out of scope: an admin can already edit any collection's shape, which covers "someone else
needs to fix this".

A member with no laptop binary at all can still own a collection on the server. The admin's
own group library is a collection like the others, so no privileged "main" layout is
hard-wired anywhere. A member-created collection is for contributed content. Do not solve
curation by minting a collection per member. Starred items and playlists are the personal
shelf over the group's collections, which is Roon's answer to per-user libraries and the
answer given on Navidrome discussion 2452.

The global settings table holds only two things about shape: the defaults a new collection
starts from, and the bounds an admin sets that no collection may exceed, such as the format
allowlist floor and a size quota.

| Knob | Home | Who edits it |
|---|---|---|
| which collections exist, their names | collection rows | the admin, or a member for their own |
| layout template, tagging policy, admission overrides | the collection row | the owner or an admin |
| new-collection defaults and admin bounds | settings table | the admin |
| routing rules: which collection an upload lands in | settings table | the admin |
| a member's default collection | member preferences | the member |
| a per-upload collection override | the uploader, not stored | the member |
| where a member's local collection lives on disk | machine settings | the member, per machine |

Mirroring pushes the rows and the files, so the shape travels with the collection and the
server never re-decides it. A server-owned collection's shape is edited only on the server.
There is still no merge, because a row has one writer. Shape edits carry an `if_version`
check, because a template write triggers a re-layout of files and last-write-wins is wrong
for that. Re-layout runs as one serialized job per collection.

### The layout template is a filesystem-write primitive

Any member may own a collection, so any member may write a layout template, and the
template decides where the pipeline writes bytes inside the unit's writable paths. Treated
as a string, it is a path-traversal primitive: a template of `../../` segments, or a tag
carrying them, renames FLAC bytes over the roster file or the store, and the member is now
admin. The rung 7 and 9 plans carry the rules below as tasks.

A template is a closed grammar of field substitutions plus a fixed allowlist of pure
formatters, parsed to a syntax tree and validated when the row is written. An unknown field
or formatter is a typed error naming the likely intended field, which is the methodology
study's S2 rule applied at row-write time rather than config-load time. `text/template`,
`html/template`, and any evaluating engine are out. Output length and
expansion count are capped.

A template yields a sequence of path components, never a string. An expanded field can
never contribute a separator, a dot segment, a NUL, or a control character. Each component
passes the portable-layout sanitizer of 2.0 hook 12, and components are NFC-normalized and
case-folded for uniqueness. The joined path is checked for containment within the
collection's server-allocated root, and every library write goes through a root-confined
handle that refuses escape and symlink traversal by construction.

A collection's on-disk root is derived from the collection id by the server. No member
influences it. A prefix check against every existing root runs at row write, so no two
collections overlap or nest, and a scanning driver never maps two music folders onto one
tree.

The template's field set is validated for injectivity: a template that omits the disc or
track fields can render two tracks to one path. At create time a file is opened
exclusively, a collision gets a deterministic disambiguation suffix, and a rename never
lands over a path owned by a different catalog row. A collision the suffix cannot resolve
rejects the album, because the album is the verdict unit.

A tagging policy can reach the identity set indirectly, by renaming some other tag to
`ALBUMARTIST`. The invariant is on the output tag set, not on the operation. A policy also
caps tag count, value length, and picture size, because a FLAC metadata block can be a
bomb.

### Shape bounds come from three sources

Not every shape beets allows is expressible once a streamer serves the library. The bounds
come from three places, and the engine validates a collection's shape against all three
before the row lands.

The admin's bounds are the settings-table floor above. The engine's own invariant is the
album as the verdict unit: the tags that make tracks group into one album, plus the
compilation flag and the album-artist convention for various-artists releases, must be
consistent across the album or a client splits it. A per-collection tagging policy can add,
rename, or strip other tags freely. It cannot change the identity set in its output.

The third source is the active streamer driver, and it varies by driver. A driver that
rebuilds its catalog by scanning disk reads the library through its own rules:

- A collection must be a disjoint directory subtree, so it can map to a music folder. The
  subtree isolates files, not entities: an artist can still have albums in several
  collections.
- Album art must be a file in the album folder with a name the driver recognizes, or an
  embedded picture.
- A fact that lives only in a database never reaches the streamer unless it is also a tag
  or a folder.

The rung 12 driver seam already reports capabilities (2.0 forward hook 6), and this spec
extends the report with shape constraints (whether collections are subtrees, which art
filenames are read, which tags are read) and with per-user folder access. The rung 12 plan
verifies the Navidrome driver's exact list against the pinned version rather than this
spec.

The 2.0 built-in streamer removes the scanner's constraints, because the catalog is
dubplate's own and the projector derives music folders. Whether the layout template is
visible to a listener depends on the driver and its version. The Subsonic API exposes folder
browsing alongside tag-based browsing, and a streamer may answer folder calls from the real
disk layout or from a tree it synthesizes from tags. Navidrome can do either, so whether
the template is visible to a folder-mode client under the 1.0 driver depends on that
setting, and the rung 12 plan records the pinned build's default and the key that changes
it. Under a driver that serves the real layout, a
collection laid out as `genre/artist/year album` browses differently from one laid out as
`artist/album`, and the owner's choice of template decides it.

A dubplate collection is a catalog row, and its on-disk location is an implementation
detail of its writer. Navidrome and Jellyfin make a library a disk root folder, which turns
a member's self-service act into an admin and a directory. A driver's subtree constraint
constrains layout, not the concept. The second declined model is container-granted access. Per-member objects
(starred, playlists, queue, history) are keyed to the member, and visibility is computed per
referenced item at read time. Jellyfin models playlists inside its library permission
system, and its tracker carries reports of members losing access to playlists they created
(issue 7071).

Relative to beets, the loss is small. With beets, a path format can depend on any query, and
a fact can live only in the database. The dubplate model keeps free path templates and
per-collection policies, and gives up query-shaped layouts that cross subtree boundaries
and database-only facts. The methodology study's S4, a layout selector as an owned wall key
with a shipped default, is retired by this spec: the selector becomes the layout default in
the settings table, and each collection row carries its own template.

## Collections interacting

Everything that happens between collections and between members follows from one writer
per collection and content-addressed audio. Album identity is a release within a
collection. Per-member rows point at album ids, which are opaque and random, so a member
cannot enumerate across a visibility boundary.

### Audio identity and byte storage

"The same album in two collections" asks about identity and about storage, and the answers
differ.

Identity is keyed on the audio, and the key is computed, never trusted. A FLAC file carries
an MD5 of the decoded audio in its stream header, and that value is whatever some encoder
wrote. Several common encoders leave it zeroed (the flac reference tracker, issue 155,
carries the cases), and a mirror push can claim any value. So the rung 7 gate decodes every
admitted track and computes the audio key itself, recording that it did. An all-zero
header value is absent, never a value. The 2.0 catalog row carries two keys. The **audio
key** identifies the same audio across collections and drives the overlap report, the
duplicate-collapse view, and re-upload refusal. The **file hash**, computed streaming over
the file bytes at the same gate, drives drift detection, the transcode-cache key, and
backup change detection, because it changes on retag and the audio key does not. Forward
hook 1 is amended to name both.

Identical audio across different releases is normal above the track level: a track on a
single and on its album, a bonus track repeated across editions, and every short silence
track collide exactly. Album sameness is therefore the ordered set of a release's track
audio keys with its track count, never a per-track match. Re-upload refusal is album-keyed.
Hardlink sharing, below, is a storage optimization that stays decoupled from identity,
which is a catalog judgment.

Storage is one file per track per collection, laid out at that collection's paths. A
scanning streamer reads real subtrees, and a per-collection tagging policy needs a file it
can write. Sharing bytes is an optimization on top. When two collections hold identical
bytes, the second file is a reflink where the filesystem offers one and a hardlink
otherwise. With a reflink an in-place write by a foreign tool damages only its own copy.
With a hardlink an in-place write reaches every collection sharing the inode, and beets,
mutagen, metaflac, and `rsync --inplace` all write in place. So the engine enforces one rule
before any write to a library path: if the file has more than one link, break the link by
copy then atomic rename. The atomic writer already does every tag write as write-to-temp
then rename, which replaces one directory entry with a fresh inode and leaves the other
collection's link untouched. Copy-on-write follows from the write discipline the repo
already has, plus the break-link rule for anything that arrives with links. The rung 16
migration lands the library into a single-link tree and shares inodes only afterwards, and
`doctor` reports any library file with more than one link whose collection rows disagree on
tags.

The consequences:

- Same audio, same tags: one copy on disk, two directory entries. This is the common case
  for two members mirroring the same rip under default policies.
- Same audio, different tags: two copies. That is the cost of per-collection policy, bounded
  by how many collections diverge, not by library size.
- Removing a track deletes one directory entry. The audio survives while any other entry
  points at the inode.
- Backup does not preserve links. rclone transfers linked files as independent objects
  (rclone issue 6160) and object storage has no link concept, so backup bytes equal the
  charged quota number, not the physical one, and a restored library is larger than the box
  it came from and may not fit. The restore runbook states this, and a post-restore relink
  pass keyed on audio key plus byte equality is optional.
- The 2.0 built-in streamer may keep one canonical file per audio key and project each
  collection's tags at read time. That is a 2.0 decision, and computing the audio key now is
  what keeps it open.

Sharing needs one filesystem. The tus staging directory, the library root, and every
collection root must share a device, and `doctor` checks it. A rename that fails across
devices falls back to a copy, and a hardlink that cannot be made across collections is a
reported condition, never a hard failure.

### The verbs

The catalog is split. The verbs a member or admin runs against collections that already
sit on the server are here. Contribute, mirror, and hand over, the verbs a laptop runs
when it attaches, are defined under Attaching, and hand over is referenced below before
that section.

**Copy** puts an album from another collection into the member's own. It is a copy at the row level: a
new album row in the target with its own layout paths, linked to the source's files while
the bytes agree and split by the target's own tag write when they do not. A retag in the
source does not propagate, because the target row is its own and the target collection has
its own tagging policy. The UI calls the gesture "Copy". The snapshot is the surprising
half, and "Add" would hide it. A true reference, one row in two collections, would break
album identity as a release within a collection, break per-collection tagging policy, and
reintroduce two writers for one row. Copy bypasses intake, so its quota check runs inside
the copy transaction.

**Remove** takes an album out of one collection and deletes that collection's rows and
layout paths only. The audio is freed by the filesystem when the last link goes, and a
sweep confirms it. Nothing deletes inline, because another collection may hold a link. The
sweep considers only paths with no open intent-journal row and older than a grace interval,
takes the same per-collection lock the writer takes, and runs dry-run first, so it never
unlinks a file whose row is committed but whose filesystem work is in flight.

**Move** is remove plus add, and it mints a new album id, so stars, playlist entries, and
history that pointed at the old id are lost. The confirmation names the loss by count: "2
members have this starred; it appears in 1 playlist". Re-pointing per-member rows across a
move would make the album id independent of the collection, which the 2.0 identity rule
forbids. The filesystem is not in the database transaction, so a move is not one
transaction. The catalog change commits with a move-intent row in the same transaction.
Then, per file, the target entry is created first (a rename within one filesystem,
otherwise link then unlink), the directories are synced, and only then is the source entry
removed. The journal replays at startup. Scan triggers are suppressed for the duration and
exactly one scan fires after, so a scanning driver never sees the tracks missing from the
source and purges its per-user rows for them.

**Retire** is the member verb for a whole collection: the collection stops projecting to the
streamer and the files stay. Under the Navidrome driver, un-projecting a collection removes
a Navidrome library, and Navidrome then marks its tracks missing and eventually purges them
with its own per-user stars and playlist entries, which at 1.0 have no other home. The
retire confirmation says so. At 2.0, per-member rows that point into a retired or
invisible collection stay in place and resolve at read time, so a playlist that mixes
collections keeps its entries and shows the reachable ones.

**Purge** deletes a retired collection's files. It is admin-only.

**Check out** takes a read-only offline copy of a server collection. The server keeps the
writer. A checkout carries the source collection's id, stays out of the server's collection
list, and cannot be mirrored back. A member asks for one directly, a hand-over leaves the
laptop's copy as one, and a seize demotes the laptop's copy to one. Writing it means asking
for hand-over in the other direction.

**Seize** is the admin's one recovery verb over a laptop-written collection: a unilateral
hand-over that transfers the writer to the server. It is one transaction that bumps the
writer epoch. Every row commit and every file link-in of a push re-checks the epoch inside
its own transaction and aborts the whole push with a typed `writer changed` error, and
staged uploads are discarded, never adopted. A returning laptop learns on connect that the
server now writes, and its local copy demotes to a checkout.

**Admission** into a collection is automatic under its routing rule. An owner's review queue
is opt-in through the review-queue setting. Rejecting after the fact is the remove verb,
with no second mechanism. The owner is notified once per batch.

### Rules that cut across the verbs

Quotas bind at three levels, all in the settings table. Per collection is the shape floor.
Per member across everything they own is the fairness counter, and a file shared by several
collections is charged to every one of them. The over-count is deliberate. It explains in
one sentence, it matches what backup will cost, and it removes any incentive to game which
collection "owns" a file. A shared file's bytes are not split across referencers, because
a total would then move when an unrelated member deleted something. The third level is the
group total, measured physically, each inode once, because the box has one disk. A hard
free-space floor is checked before every write, because a
full disk wedges the store's write-ahead log and takes the server down. Enforcement is a
reservation taken before bytes move, with a true-up at completion, so two concurrent
uploads cannot each pass the check and together overshoot. A mirror push declares its total
up front and reserves the same way.

A member's default collection can go stale under them: retired, closed to contribution, or
handed over between uploads. The uploader resolves the preference when the page loads and
re-validates at write. If the collection is unwritable at either point, the uploader falls
back to the admin-owned default collection and says so on screen. Silently landing an album
in the wrong home is the worst outcome, since extracting it is a manual move.

Duplicates across collections are allowed by design and visible by default. A member
preference named `collapse-duplicates` folds duplicate albums across visible collections by
album audio identity in the 2.0 projector, and a per-collection flag says whether the
collection participates in that merged view. The keys are named now so the rung 6 schema
carries them. Under the Navidrome driver, two libraries holding linked, identically tagged
files can collide on Navidrome's tag-derived persistent ids and render as one album or flap
between libraries per scan. The rung 12 plan reproduces this against the pinned build and
pins a persistent-id formula that includes the library or the path if one exists.

A member leaving never deletes audio. Server-written collections they owned pass to the
admin. Laptop-written collections freeze at their last mirrored state and show a "writer
gone" status. Their preferences are deleted. Playlists other members still use transfer to
the admin with an explicit owner rewrite and a name-collision rule. The member's user row
is tombstoned with a removal time, never deleted, so collections, history, and the audit
trail keep referential integrity. Rung 5 defines offboarding as the epoch bump plus the
streamer deprovision. The removal verb and this data cleanup file at rung 14, with a
checklist test that asserts all of: epoch bump, app passwords, device credentials, writer
epoch freeze, in-flight push abort, collection reassignment, streamer deprovision. The
store enables foreign keys per connection in the DSN, or every referential rule here is
decorative.

A stale mirror is a laptop-written collection whose writer has not connected for a while.
The row carries a last-mirror time. Past an admin-set bound the collection is flagged stale
and keeps serving. It is never deleted automatically. Seize, above, is the recovery.

Sharing outside the roster is declined (Geoff, 2026-09-11). The roster is the trust boundary
for a group of friends. Plex tiers account classes and Funkwhale approves followers per
library. Both add a second identity kind, and this design has one: one roster, one admin
flag, with capability differences as per-member overrides. Share rows keyed to an item and
an expiry are filed as a schema hook beside visibility, so the two models are designed
together if sharing ever ships.

## The writer is a device credential

"A named local binary" has to name something. The writer is a server-minted device
credential per connection profile, so two profiles on one laptop are two writers. The
server stores only a hash of the credential, bound to the surrogate user id and a device
id, with its creation time, last-seen address, expiry, and revocation time. Authorization
for a push is that the collection's writer equals the credential's device and the
collection's owner equals the credential's member. A client-asserted writer field is never
read. The credential is never derived from a hostname or a hardware address.

A push is untrusted input. A mirror is a different transport, not a lower trust level. The
server recomputes every fact it will trust: it decodes the audio and computes the audio key
rather than accepting one, re-derives the layout path from the collection row's template
rather than accepting the client's, validates tags against the policy and the sanitizer,
and decodes art under pixel and byte caps and re-encodes it. A crafted audio key would
otherwise poison the overlap report and the link sharing, and crafted tags or art would
reach the search index, the Subsonic responses, and the cover-art resizer.

The per-collection mirror generation defeats replay. A captured push cannot land an older
state over an owner's or admin's removal. The writer epoch defeats the seized laptop. A
reinstalled laptop registers as a new writer and needs an explicit admin adopt action, or
its old mirror stays orphaned. A cloned laptop disk yields a working credential. Two pushes
at one generation from two devices is the signal, and the response is to freeze the
collection pending the owner's re-attestation. That is mitigation, not prevention, and it
is an accepted risk.

## Attaching a local library to a server

When a member whose laptop already holds an organized collection connects to a server that
already holds one, nothing happens beyond showing both. The local collection keeps its
writer and the server's collection keeps its writer. A name clash is not a conflict,
because a collection is identified by its id and writer. The name carries no identity. The
TUI shows the server's collections as browse-and-play views and the local collection as
the thing it owns.

Before asking anything, the TUI measures the overlap. Every album carries its audio
identity and usually a release MBID. The report sorts local albums into three buckets: an
album absent from the server, an album identical to one already there (audio match), and
the same release in different bytes (MBID match, audio differs). The third bucket is a
different rip, and it is never refused or replaced automatically. The member chooses per
album to skip it, mirror it instead, or offer it as a replacement that the target
collection's owner adjudicates. The report answers only over collections the member can
see, and so does intake's duplicate refusal, because either would otherwise tell a member
what the server holds in collections they cannot see. Bulk identity queries are rate-limited
and logged. The member then picks one of four verbs. The TUI runs none of them without
confirmation.

- **Contribute**, the default. Push the albums the server lacks into a server collection
  through ordinary intake. The server's shape and verdicts govern, one album at a time, and
  duplicates are refused by album identity. The local collection is untouched.
- **Mirror.** Publish the local collection to the server as its own collection, still
  written by the laptop and owned by the member. Overlapping albums exist in two
  collections, which the 2.0 spec allows by design, and identical bytes share one inode.
  The admin approves each member's first mirror (Geoff, 2026-09-11), because a mirror
  consumes disk with no intake verdict per album. After that the member mirrors freely
  within quota.
- **Hand over.** Transfer the writer to the server. The server takes the rows and files and
  re-lays them to the collection row's own template. "Keep my layout" means the row carries
  that template, which already travels with the collection. No collection ever sits outside
  a live template. The laptop's copy becomes a checkout. This is the migration path for the
  founding library at launch.

  > Erratum, 2026-09-23 (Geoff): for the founding library, rung 16's migration tooling
  > performs the hand-over at launch, landing the library as a single-link tree. The TUI verb
  > arrives with the TUI, after launch.
- **Check out.** Take a read-only offline copy of a server collection, as defined above.

A second laptop of the same member, holding an overlapping local library, cannot write the
first laptop's mirror, because the writer is a credential per profile. Its choices are
contribute, mirror as a new collection, or check out the first mirror. The TUI detects the
same-member case and leads with contribute, because two near-identical mirrors clutter
browsing more than anything else here, and a naive choice produces them.

One guarantee runs in both directions. The server's layout template never rewrites the
laptop's disk, and the laptop's template never changes the server's rows. Attaching twice
is idempotent, because the identity keys make every push a no-op for what already landed.

## Accepted risks

This section is an index of the risks the body accepts, plus one it does not state
elsewhere. An admin is fully trusted. Re-authentication on sensitive settings raises the
bar on a stolen session and does nothing against a live-compromised admin browser.

- Private is not confidentiality from the admin or from disk, backup, or log access
  (Visibility).
- Mirrors bypass per-album intake verdicts; first-mirror approval plus the quota
  reservation is the whole bound (Attaching).
- The per-member quota over-counts shared files, deliberately (Rules that cut across the
  verbs).
- A move loses stars, playlist entries, and history, and says so by count (The verbs).
- Link sharing assumes one filesystem and the break-link rule. A writer that skips the
  rule breaks collection isolation silently, and `doctor` catches only the subset where the
  sharing collections' tags disagree (Audio identity and byte storage).
- A cloned laptop disk holds a working writer credential; divergence detection freezes the
  collection and does not prevent the clone (The writer is a device credential).

## What this spec routes elsewhere

- Rung 6 store, filed as 2.0 forward hook 15:
  - the settings table with source, apply mode, the sensitive flag, and the audit table;
  - the per-member overrides and preferences tables;
  - on the collection row: `owner` (non-null, surrogate id), `visibility`, `root_path`
    (server-allocated, disjoint), writer epoch, mirror generation, last-mirror time, and
    the duplicate-collapse flag;
  - the device-credential table, the two quota counters, and the move-intent journal;
  - foreign keys enabled in the DSN, and share rows reserved beside visibility.
- Rung 7 and 9 pipeline:
  - the template grammar and validator, the component sanitizer, root-confined writes,
    and the injectivity check with exclusive create;
  - the break-link rule;
  - the audio key computed by decode at the gate, with the file hash beside it;
  - intake caps: upload size, per-album file count, decompressed total, path depth,
    metadata and picture size;
  - the same-device check for staging, library, and collection roots.
- Rung 12 driver seam: shape constraints and per-user folder access in the capability
  report; verification of the pinned build's art filenames, read tags, folder-browsing
  behavior, and per-user library API; the linked-duplicate persistent-id reproduction.
  Filed as an extension of 2.0 forward hook 6.
- Rung 14 admin spine: two mounts; settings, per-member overrides, and collection endpoints
  the web app and the TUI share, including retire, purge, reassign, adopt, and seize; the
  re-authentication gate and outbound-URL validation on sensitive settings; the audit on
  every write; member removal as the checklist above; version negotiation on connect.
- Rung 16 migration: land the library as a single-link tree, and state the restore-size
  expectation in the runbook.
- Rung 5's plan: the revocation seam's read is the one a device-credential store subscribes
  to; a last-admin lockout guard, atomic against concurrent self-demotion.
- The 2.0 projector: one `visibleCollections(member)` predicate joined in a single accessor
  no handler bypasses, with a table-driven test asserting an identical not-found for
  invisible and nonexistent across `getMusicFolders`, `getAlbum`, `getCoverArt`, `search3`,
  `getPlaylist`, `getRandomSongs`, `getAlbumList2`, `getArtists`, `getIndexes`,
  `getStarred2`, `getGenres`, `getTopSongs`, `stream`, `download`, `createShare`, and star
  and scrobble; shares re-checked at fetch time.
- The registry: a home and an apply mode per key, and the env seed rule above. An erratum
  in the build spec, dated 2026-09-12, against rung 1b's contract.
- The TUI, unscheduled on the ladder (ROADMAP, the greenfield entry): connection profiles,
  the machine-settings and credential files, the overlap report, and the four verbs. This
  spec is its configuration brief.

The TUI's player is one more Subsonic client of the 2.0 streamer. It is not a reason for 2.0
to grow player features, and the 2.0 spec's "no first-party player" ruling stands.

## Supersessions

Each line names a ratified passage this spec changes, and the erratum that records it.

- 2.0 spec, forward hook 1: one content hash becomes two keys, the audio key and the file
  hash. Recorded in the hook text.
- 2.0 spec, forward hook 15: extended with the rung 6 list above. Recorded in the hook text.
- 2.0 spec, web surface: gains the admin settings surface (Geoff, 2026-09-12). Recorded as
  revision 2.2.
- Build spec, config wall: thresholds move from the wall to the store with an env seed;
  the registry gains a home and an apply mode per key. Erratum 2026-09-12. The build spec's
  out-of-scope line for the web admin page is superseded by the 2.0 web surface change.
- Organizer spec, the "server-to-local read-only mirror is out of scope" sentence: the
  checkout verb is that mirror, scheduled with the TUI. And the TUI's surface list gains
  playback (Geoff, 2026-09-11). Erratum 2026-09-12 in that spec.
- ROADMAP, "mirror and two-way sync are out of scope": the laptop-to-server mirror verb is
  in, one writer per collection; two-way sync stays out. Recorded beside the line.
- Config methodology study, S4: the layout selector wall key is retired in favor of the
  settings-table default and the per-collection template. Recorded beside S4.
- Rung 5 spec: "no address lives in configuration" holds, and the admin email stays on the
  roster; this spec never puts it in settings. No change.

Terms, because three of them drifted across the documents. A **mirror** is a laptop-written
collection published to the server. A **checkout** is a read-only local copy of a server
collection, the thing older documents called a server-to-local mirror. A **connection
profile** is the TUI's named server or local target, unrelated to a transcode profile, a
Roon profile, or a DAP export profile. **Watch folders** are the laptop directories the TUI
imports from; a **music folder** is the Subsonic entity a collection maps to.
