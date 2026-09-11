# Configuration homes: where each setting lives

Design spec, 2026-09-11, brainstormed with Geoff in one sitting. Status: ratified in
conversation, awaiting Geoff's read of this file. It settles where a configuration value
lives across the three surfaces dubplate will have: the server's own configuration, the 2.0
web admin page, and the bubbletea TUI a member runs on their own machine. This spec defines no implementation
work. Ruling 6 of the greenfield rulings holds: no TUI work until the server is fully formed. The
spec routes its hooks into the rung 6 and rung 14 plans and into the 2.0 streamer spec's
forward-hooks list.

## The rule

Every knob has exactly one home, and the home is chosen by who owns the truth. A web page or
a TUI is a view over a home. Neither holds a second copy. Because no knob lives in two homes,
there is no synchronization anywhere in the design. Configuration appears to live on both the
client and the server. In fact there are four kinds of knob, and the table below gives each
one a home.

| Home | Owner of the truth | Examples | How the TUI reaches it |
|---|---|---|---|
| Server bootstrap, the env wall | the operator | state dir, library root, bind address, secrets | never; not shown |
| Server settings, a table in the store | the admin | SMTP from-address, match thresholds, backup schedule | the rung 14 admin API, for admins only |
| Member preferences, per-member rows in the store | the member | default collection, notification choices, starred, playlists, queue | the member API, read on connect, written on change |
| Machine settings, a local file | the member, per machine | server URL, saved token, audio device, theme, keybindings, source folders | edited in place; never sent to the server |

A member with two laptops sees the same preferences on both, because they live on the
server, and a different theme on each, because themes live on the machine. Editing a
server-side value from the TUI is an API write, so the web page shows it on its next load
with no reconciliation step. Concurrent edits are last-write-wins on the server, which is
enough for values a person changes by hand. The TUI holds no local copy of a server value
beyond what the current screen paints, so nothing goes stale offline, and a disconnected TUI
can still change its own machine settings.

## The server's own knobs

Today the server reads an env wall and nothing else: no file, no reload, no write-back, and
no metadata on a key beyond its type and default. That stays true for the bootstrap column.
Every other key moves to the store.

A knob is bootstrap if the process needs it before it can open its store, if changing it live
would strand the running process, or if it is a secret. Everything else is a server setting
and lives in a settings table in the rung 6 store, editable live through the rung 14 admin
API and backed up with the store. This follows the Lidarr and Sonarr split: a small config
file for what the process needs before it can start, and everything else in the database
behind the UI. `deploy.sh` keeps rendering the env
from the age store, so a stranger's setup does not change.

| Bootstrap, env only, restart to change | Server settings, in the store, live |
|---|---|
| state dir, library root | site name, admin email |
| bind addresses, port, URL prefix | SMTP host, port, from address |
| trusted-proxy and forward-auth settings | session and token lifetimes |
| every secret: session secret, SMTP password, streamer admin token, backup credentials | streamer driver choice and its URL |
| shadow settings, workstation-only, retired at launch | backup schedule and targets |
| | intake limits: upload size, concurrency, quotas |
| | pipeline defaults: match cutoffs, tag policy, layout templates |
| | collection bounds and new-collection defaults |
| | review-queue and notification settings |
| | monitoring ping URLs |

The left column is about a dozen names and is the whole env template. The right column grows
with every rung and never touches the env. The registry keeps the wall's compat discipline:
an unknown or superseded env key is still a load error.

The registry gains one piece of metadata: a key's home. A key declared as a server setting is
read from the store with the env value, if any, honored only at first boot as a seed. This
keeps one declaration site per key and lets `doctor` say where each value came from.

## Two surfaces, one API

The 2.0 web admin page and the TUI both call the rung 14 admin API. An admin logged into the
TUI gets an admin section that makes the same calls the web page makes. The TUI never edits a
server file, so the two surfaces have nothing to disagree about. The organizer spec's
remote-TUI mode was already this shape.

The ordering follows from this. The admin API lands at rung 14 and the web admin page at
2.0, so the TUI reaches server settings before the web page exists.

A knob is a member preference when the server's behavior changes with its value. Sort
order the server applies to a listing, notification choices, and the default collection
change what the server does, so they are preferences. Colors, keys, and audio output change
only how one machine renders, so they are machine settings. A knob that passes the test in
one respect and fails it in another is two knobs.

## Profiles in the TUI

The TUI is a local organizer first and a client second. Stand-alone, it embeds the engine and
runs the pipeline against a library on the laptop, so it needs thresholds, layout templates,
and collections of its own. Those are the same knobs the server's settings table holds, and
this is the one place two homes hold the same kind of knob.

The TUI switches modes instead of merging the two. It holds profiles. A local profile owns a
local state dir with its own store, and the engine reads settings from there. A server
profile names a URL and a token, the engine runs on the server, and the server's settings
govern while the local engine settings are not consulted. Choosing a profile is how a member
connects. Machine settings sit above the profiles and apply to all of them.

Two things follow. The settings schema is one schema, defined once in the engine, whether the
store is local or on the box. "Take my local settings to the server" is an explicit one-shot
copy that an admin reviews. No background sync runs. The reverse copy is the same action
the other way.

## Collections carry their shape

Collection shape is where the knobs are richest and where a local binary and a server both
hold them. A collection is a row in the store, and its shape is data on that row. No global keys hold
it. The row carries the name, the layout template for paths and filenames,
the tagging policy, the admission rules (format allowlist, match-threshold overrides), and
two attributes of control:

- The **writer**, already fixed as 2.0 forward hook 8: exactly one process writes the rows,
  this server or a named local binary.
- The **owner**, new in this spec: the person who may change the collection's shape and who
  decides what goes in.

The two are independent, which yields three kinds of collection.

| Kind | Writer | Owner | How it comes to be |
|---|---|---|---|
| laptop-written, member-owned | a member's local binary | that member | the mirror verb, below |
| server-written, member-owned | the server | a member | the member creates it on the server and contributes into it through intake |
| server-written, admin-owned | the server | the admin | the group library; the default collection is one of these |

A collection's shape is edited by its owner or an admin. Contributions to it are admitted by
its routing rule, and other members contribute to a member-owned collection only if the owner
allows. The admin can reassign an owner, retire a collection, and switch member creation off.
Any member may create server collections by default, since a group of family or friends is
the product's shape. A member with no laptop binary at all can still own a collection on the
server. The admin's own group library is a collection like the others, so no privileged
"main" layout is hard-wired anywhere.

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
There is still no merge, because a row has one writer.

### Shape bounds come from three sources

Not every shape beets allows is expressible once a streamer serves the library. The bounds
come from three places, and the engine validates a collection's shape against all three
before the row lands.

The admin's bounds are the settings-table floor above. The engine's own invariant is the
album as the verdict unit: the tags that make tracks group into one album, plus the
compilation flag and the album-artist convention for various-artists releases, must be
consistent across the album or a client splits it. A per-collection tagging policy can add,
rename, or strip other tags freely. It cannot touch the identity set.

The third source is the active streamer driver, and it varies by driver. A driver that
rebuilds its catalog by scanning disk reads the library through its own rules:

- A collection must be a disjoint directory subtree, so it can map to a music folder.
- Album art must be a file in the album folder with a name the driver recognizes, or an
  embedded picture.
- A fact that lives only in a database never reaches the streamer unless it is also a tag
  or a folder.

The rung 12 driver seam already reports capabilities (2.0 forward hook 6), and this
spec extends the report with shape constraints: whether collections are subtrees, which art
filenames are read, and which tags are read. The rung 12 plan verifies the Navidrome driver's
exact list against the pinned version rather than this spec.

The 2.0 built-in streamer removes the scanner's constraints, because the catalog is
dubplate's own and the projector derives music folders. One constraint survives every
driver. The Subsonic API exposes folder browsing alongside tag-based browsing, and clients
that browse by folder show the layout template to the user. In those clients, a collection
laid out as `genre/artist/year album` browses differently from one laid out as
`artist/album`. The owner's choice of template decides how the collection browses.

Relative to beets, the loss is small. With beets, a path format can depend on any query, and
a fact can live only in the database. The dubplate model keeps free path templates and
per-collection policies, and gives up query-shaped layouts that cross subtree boundaries and database-only
facts.

## Attaching a local library to a server

When a member whose laptop already holds an organized collection connects to a server that
already holds one, nothing happens beyond showing both. The local collection keeps its writer
and the server's collection keeps its writer. A name clash is not a conflict, because a
collection is identified by its id and writer. The name carries no identity. The TUI shows the server's
collections as browse-and-play views and the local collection as the thing it owns.

Before asking anything, the TUI measures the overlap. Every album carries a content hash
(2.0 hook 1) and usually a release MBID, so the report reads, for example, "of 312 local
albums, 140 are already on the server, 172 are not". The member then picks one of three
verbs. The TUI runs none of the three without confirmation.

- **Contribute**, the default. Push the albums the server lacks into a server collection
  through ordinary intake. The server's shape and verdicts govern, one album at a time, and
  duplicates are refused by hash or MBID. The local collection is untouched.
- **Mirror.** Publish the local collection to the server as its own collection, still
  written by the laptop and owned by the member. Overlapping albums exist in two
  collections, which the 2.0 spec allows by design, and content addressing stores each file
  once.
- **Hand over.** Transfer the writer to the server. The server takes the rows and files,
  applies its own layout if the admin chooses, and the laptop's copy becomes a read-only
  mirror it can refresh or delete. This is the migration path for the founding library at
  launch. The reverse verb, checking out a server collection as a local read-only mirror,
  gives offline play.

The contract has a second half. The server's layout template never rewrites the laptop's
disk, and the laptop's template never changes the server's rows. Attaching
twice is idempotent, because the hashes make every push a no-op for what already landed.

## What this spec routes elsewhere

- Rung 6 store: the settings table, the per-member preferences table, and the `owner`
  attribute beside `writer` on the collection row. Filed as 2.0 forward hook 15.
- Rung 12 driver seam: shape constraints in the capability report. Filed as an extension of
  2.0 forward hook 6.
- Rung 14 admin spine: settings and collection endpoints the web page and the TUI share.
- The registry: a home per key, and env honored only as a first-boot seed for keys whose home
  is the store.
- The post-launch TUI initiative: profiles, the machine-settings file, the overlap report,
  and the three verbs. This spec is its configuration brief.

The TUI's player is one more Subsonic client of the 2.0 streamer. It is not a reason for 2.0
to grow player features, and the 2.0 spec's "no first-party player" ruling stands.
