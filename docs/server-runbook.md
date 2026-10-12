# dubplate-server runbook

This runbook is for the person who installs and operates `dubplate-server`.
It covers install, `doctor`, mail, the trust boundary, backup, and deploy.
It assumes a Linux host that runs systemd. A host without systemd, such as
Alpine with OpenRC or a container, has no documented install path.

The `dubplate-server` commands were run against a scratch state directory, and
their flags match `--help`. The `systemd-run`, `deploy.sh`, and rollback
commands, and the proxy configurations, have not yet run on a fresh host. Host
names, addresses, and paths in examples are placeholders: `music.example.org`,
`192.0.2.10`, and `/var/lib/dubplate`.

## Install order

Run the steps in this order. Steps 7 and later need the ones before them.
A `doctor` run too early correctly reports a server that cannot start yet.

1. **Build the binary.** You need Go at the version `go.mod` names and Node 24
   or later; `make web` refuses an older Node. From a clean checkout:

   ```sh
   npm ci --prefix web
   make web
   make build-linux
   ```

   `make web` builds the web bundle the binary embeds and runs the web
   checks. `make build-linux` writes `dubplate-server-linux-amd64`. A plain
   `go build` without `make web` embeds a placeholder, and `doctor`'s `web`
   row then fails.
2. **Create the service user.** The unit runs as `music`, and every
   `systemd-run` command in this runbook names the same user:

   ```sh
   useradd --system --user-group --no-create-home --shell /usr/sbin/nologin music
   ```

   To use another name, change the unit's `User=` line, the owner in step 6,
   and every `-p User=` line together.
3. **Install the binary and its unit.** Copy the binary to
   `/usr/local/bin/dubplate-server` and `systemd/dubplate-server-serve.service`
   to `/etc/systemd/system`. Delete the unit's `OnFailure=` line, or point it
   at a failure hook of your own: the unit it names belongs to the
   maintainers' estate, so on your host a crash loop would alert nobody. Then
   run `systemctl daemon-reload`. Do not enable the unit yet.
   `scripts/deploy.sh` is not a general installer: it deploys the
   maintainers' own estate (see [Deploy](#deploy)).
4. **Write the environment.** Create `/etc/dubplate/server.env` by hand from
   the rows in `env/dubplate-server.env.template`. A `NAME=value` row is a
   literal. A bare `NAME=` row needs a value from your own secrets. Set the
   file's mode to 0600 and its owner to root. The template carries the
   maintainers' estate values, so replace these rows with your own:

   - `DUBPLATE_PUBLIC_URL` is the address members open, as a bare `https://`
     origin with no path, for example `https://music.example.org`. The serve
     prefix, `/dubplate` by default, sets where the server lives under it.
   - `DUBPLATE_SITE_NAME` is the name the sign-in page shows.
   - `DUBPLATE_ADMIN_EMAIL` is the operator's contact address. The server
     sends it in the user agent of its MusicBrainz requests. It also names
     the first admin: step 10 adds that address with `--admin`. After that
     the roster alone decides who is an admin, and the variable grants
     nothing. `doctor` warns when the address no longer names a live admin.
   - `DUBPLATE_LIBRARY_ROOT` is the directory that holds the music library.
     Every verb requires it, and the backup and restore verbs read it.
   - `DUBPLATE_MAGIC_LINK_SECRET` and `DUBPLATE_SESSION_SECRET` each need at
     least 32 random bytes, and the two must differ. This command prints 48
     random bytes as 64 characters; run it once for each:

     ```sh
     openssl rand -base64 48
     ```

   - The SMTP rows: see [Mail](#mail).
   - The trusted proxy rows: see step 8.
5. **Check that the paths and the grant agree.** The unit grants write access
   to exactly two directories, `/var/lib/dubplate/auth` and
   `/var/lib/dubplate/store`, on its `ReadWritePaths=` line. Those are the
   directories the server writes when `DUBPLATE_STATE_DIR` is
   `/var/lib/dubplate`, as the template sets it: the sign-in state under
   `auth` and the member and settings store under `store`. If you change that
   value, the grant must name `$DUBPLATE_STATE_DIR/auth` and
   `$DUBPLATE_STATE_DIR/store`. Change the unit's `ReadWritePaths=` line, the
   two commands in step 6, and the `ReadWritePaths` line of every
   `systemd-run` command in this runbook together. `doctor` checks the `auth`
   row (the store directory's mode and owner, and that the record directories
   are writable by the user running `doctor`). It cannot read the unit's
   `ReadWritePaths=` line, so it cannot see a grant that disagrees with the
   paths. Run from the `systemd-run` command in step 10, it sees only the
   grant you gave that command. A unit whose grant is wrong then fails at
   start or at its first store write.
6. **Create the state leaves.** Create each granted directory, owned by the
   service user with mode 0700:

   ```sh
   install -d -o music -g music -m 0700 /var/lib/dubplate/auth
   install -d -o music -g music -m 0700 /var/lib/dubplate/store
   ```

   A leaf missing from the host stops the unit at start with the namespace
   failure `226/NAMESPACE`, because systemd cannot mount a path that does not
   exist.
7. **Install the transliteration table.** The matcher cannot score without it,
   and `doctor` fails the `unidecode` row until it exists. Create the leaf the
   install writes, owned by the service user with mode 0755 (the verb refuses
   an existing leaf with another mode or owner), then run the install as the
   service user with a grant of that leaf alone:

   ```sh
   install -d -o music -g music -m 0755 /var/lib/dubplate/unidecode
   systemd-run --wait --pipe --collect \
     -p User=music -p Environment=DUBPLATE_STATE_DIR=/var/lib/dubplate \
     -p NoNewPrivileges=yes -p PrivateTmp=yes -p ProtectSystem=strict \
     -p ProtectHome=yes -p ProtectKernelTunables=yes \
     -p ProtectKernelModules=yes -p ProtectControlGroups=yes \
     -p RestrictNamespaces=yes -p RestrictRealtime=yes \
     -p RestrictSUIDSGID=yes -p LockPersonality=yes \
     -p MemoryDenyWriteExecute=yes -p SystemCallFilter=@system-service \
     -p SystemCallArchitectures=native -p AmbientCapabilities= \
     -p CapabilityBoundingSet= \
     -p 'RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX' \
     -p ReadWritePaths=/var/lib/dubplate/unidecode \
     /usr/local/bin/dubplate-server unidecode install
   ```

   The install downloads the pinned wheel, so this run keeps the network
   families. Offline, end the command with the wheel you cached with `make
   unidecode-wheel`:

   ```sh
   /usr/local/bin/dubplate-server unidecode install --from ~/.cache/unidecode/Unidecode-1.4.0-py3-none-any.whl
   ```

   The grant's `ProtectHome=yes` and `PrivateTmp=yes` hide a wheel left in a
   home directory or in `/tmp`, so copy the wheel to a path outside both that
   the service user can read, and pass that path to `--from`.

8. **Put a TLS front in front of the server.** Sign-in requires HTTPS (see
   [Sign-in requires HTTPS](#sign-in-requires-https)), and the server listens
   on `127.0.0.1:8080` by default, so members reach it through a proxy or a
   tunnel that terminates TLS. Configure the front as a trusted proxy in
   `/etc/dubplate/server.env`. A front on the same host connects from
   loopback, and it may dial either loopback address, so list both:

   ```sh
   DUBPLATE_TRUSTED_PROXY_ENABLED=true
   DUBPLATE_TRUSTED_PROXY_ADDRS=127.0.0.1,::1
   DUBPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For
   ```

   Never leave the class off behind a front. With it off, every request
   comes from the front's address, so the server keys every member on that
   one address. At the default limit, twenty sign-in requests from anyone
   spend the per-client allowance for the whole group, and nobody can sign in until the hour turns.
   [Example fronts](#example-fronts) has a Caddy, an nginx, and a cloudflared
   configuration, and [The trust boundary](#the-trust-boundary) says what the
   front must do with the header.
9. **Start the unit.** The start creates the store and applies the settings
   in the environment to it (see [The settings rows](#the-settings-rows)).

   ```sh
   systemctl start dubplate-server-serve.service
   ```

   Until the first admin exists, every sign-in request answers as an unknown
   address. The window lasts until the next step.
10. **Add the first admin.** `member add` needs the store, which the unit's
    start created, and refuses an absent store with "start serve first". Run
    `member list` first. Only when it shows no admin, add one with the
    `DUBPLATE_ADMIN_EMAIL` address from step 4 and `--admin`. On a removed
    member's address the verb reactivates the row with the flag, so a box that
    already has an admin never runs it. The verb requires `--name` and
    `--username` as well as the address:

    ```sh
    dubplate-server member list
    dubplate-server member add admin@example.org --admin --name "Ada Admin" --username ada
    ```

    Run both as the service user through `systemd-run`, carrying the serve
    unit's own filesystem, privilege, and address-family properties. A hand-run
    verb never runs as the bare binary under root, because the unit's guarantees
    belong to the unit. On the host:

    ```sh
    systemd-run --wait --pipe --collect \
      -p User=music -p EnvironmentFile=/etc/dubplate/server.env \
      -p NoNewPrivileges=yes -p PrivateTmp=yes -p ProtectSystem=strict \
      -p ProtectHome=yes -p ProtectKernelTunables=yes \
      -p ProtectKernelModules=yes -p ProtectControlGroups=yes \
      -p RestrictNamespaces=yes -p RestrictRealtime=yes \
      -p RestrictSUIDSGID=yes -p LockPersonality=yes \
      -p MemoryDenyWriteExecute=yes -p SystemCallFilter=@system-service \
      -p SystemCallArchitectures=native -p AmbientCapabilities= \
      -p CapabilityBoundingSet= -p UMask=0027 \
      -p 'RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX' \
      -p 'ReadWritePaths=/var/lib/dubplate/auth /var/lib/dubplate/store' \
      /usr/local/bin/dubplate-server member add admin@example.org --admin \
        --name "Ada Admin" --username ada
    ```

    Replace the `member add ...` arguments with `member list` to run that verb
    the same way. `member remove`, `member demote`, and `settings reset` run
    through this form too. `settings reset` refuses a key the environment
    file sets. To clear a stored value that the file contradicts, point
    `EnvironmentFile` at a copy of `/etc/dubplate/server.env` with that one key
    removed, or edit the key in the file to the stored value and restart
    `serve`. The `ReadWritePaths` line is the grant step 5 describes.
11. **Run `doctor`.** Use the same `systemd-run` command as step 10 with
    `doctor` as the verb, and leave out the `RestrictAddressFamilies` line.
    The `bind addresses` row reads the interface table over a netlink socket,
    which the restriction denies. With the restriction in place the row reads
    `not checked: route ip+net: netlinkrib: address family not supported by
    protocol` and fails `doctor`, so the unit itself keeps the restriction and
    only this check leaves it off. `doctor` exits 0 only when every row
    passes. The unit is up, so `boundary probe` should read `refused`. The
    probe skips an address you listed as a trusted proxy in step 8, and the row
    says so.
12. **Sign in.** Open `https://music.example.org/dubplate/login`, with your own
    host, and request a link for the admin address from step 10. An admin adds
    members at `https://music.example.org/dubplate/members`. That page answers
    `404 page not found` to anyone not signed in as an admin, and so does `/`.

`dubplate-server check` is the short form of the configuration wall. It prints
`configuration: ok` and the store state, and exits 0 when `Load` accepts the
environment. It runs none of the checks that need the filesystem or the
network, so it never replaces step 11.

## Upgrade a box that holds a roster file

The store replaced the roster file. A fresh install never meets this section.
Follow it on a box that ran an earlier build, whose roster file holds the
admin. The store starts empty, so `member add --admin` into the
empty store is the bootstrap, and you re-add every other member afterward. Run
each verb through the `systemd-run` command in install step 10, with the store
leaf granted.

1. Take the rollback copies (see [The rollback copies](#the-rollback-copies)),
   then run `deploy.sh --apply`. The rendered environment file has no roster
   row, the deploy creates the `store` leaf, and it starts nothing.
2. Run one supervised start and stop of the unit. The start creates the store.
   `doctor`'s `members` row then fails and names `member add --admin`, as
   expected at this point.

   ```sh
   systemctl start dubplate-server-serve.service
   systemctl stop dubplate-server-serve.service
   ```

3. Move the roster file aside, inside its leaf. The new name keeps the old file
   for reference and for a rollback, and no build of this release reads it.

   ```sh
   mv /var/lib/dubplate/roster/roster.json \
      /var/lib/dubplate/roster/roster.json.pre-username
   ```

   `mv` keeps the file's mode `0600` and its owner `music`.
4. Add the admin with the `member add` command from install step 10, taking the
   admin's address, display name, and `streamer_username` from the aside file:

   ```sh
   /usr/local/bin/dubplate-server member add ADDRESS --admin \
     --name "NAME" --username USERNAME
   ```

5. Run `member list` the same way. It shows exactly one member, the admin. Then
   run `doctor` as in install step 11, which exits 0.

Nothing in this upgrade enables the unit, and apart from the supervised start
in step 2 nothing starts it. A re-run after a rollback finds the admin already
stored: `member add` refuses it as a duplicate, which is correct, and `member
list` showing that admin is the pass condition. Add the other members
afterward with `member add` (without `--admin`) or on the members page.

### Rolling back past the store

The rollback copies restore the previous binary, environment file, and units.
That binary predates the username flag the upgrade used, so the rollback
re-creates no admin. Step 3 moved the roster file aside unchanged, so move it
back:

```sh
mv /var/lib/dubplate/roster/roster.json.pre-username \
   /var/lib/dubplate/roster/roster.json
```

The previous unit grants the `roster` leaf, which the deploy left on the box.
The `store` leaf stays on disk, unread by the previous binary.

## The doctor rows

`doctor` prints one line per row and exits 1 when any row fails.
This section lists the rows by what they prove and what they cannot. A row
that prints `not checked` adds no failure of its own, because the row that
owns the cause reports it.

### The wall and the sources

Each owned key, reserved class, integration class, and threshold prints one
row, in that order. An owned key reads `ok`, `missing`, or `invalid: <reason>`.
The identity source reads `active`, `inactive`, or `unsatisfied: <key>
missing` naming every failing key, and never prints a value, since the source's
keys include secrets. The `members` row reads `no store yet; start
serve, then member add --admin` before the first start, `members: <n>, admins:
<m>` once the store holds members, and `not read` for a store the `store` row
cannot vouch for. It fails with `no members; run dubplate-server member add
--admin` or `members: <n>, admins: 0; run dubplate-server member add --admin`
when nobody can add a member through the page. The `DUBPLATE_ADMIN_EMAIL` row
reads `ok` when the address names a live admin. It starts `warning:` and does
not fail `doctor` when the address names a removed member, no member, or a
member without the admin flag. An empty roster adds no warning there, since the
`members` row already fails.

### The state rows

- `store` reads `absent`, `current (v<N>)`, `older ...`, or a failing state
  such as `busy: held by pid <n>`, `refused`, or `newer`. A store absent from
  a fresh install is not a failure.
- `backup last success`, `backup wal size`, `backup free space`, and `backup
  local copies` read from the store and print `not checked` when the store
  cannot supply them.
- `backup remote` reads `unavailable: backup integration disabled` while the
  flag is off. See [Backup](#backup).
- `unidecode` reads `ok (1.4.0)`, `missing: run dubplate-server unidecode
  install`, or `invalid: ...`. It proves the table loads and passes its
  checksum. It does not prove the matcher's output.

### The boundary rows

These four rows print after `unidecode` and before the settings rows. Each
prints exactly one line.

#### `client ip`

Reads `direct` when the trusted proxy class is off, and
`trusted proxy: <header> from <n> address(es)` when it is on and its values
resolve. It reads `invalid: ...` and fails when an address entry is empty, is
not an IP address, is a network, carries a zone, or is the unspecified
address, or when the header name is one the strip removes or the shell
depends on. It names a count, never an address. It reads `not checked` when
the wall refused the environment.

#### `web`

Runs the same web resolution serve runs at startup: the theme match and
palette, the branding value, and the operator stylesheet. It reads `ok` or
`invalid: <reason>`. Because it calls the same function, `doctor` and `serve`
cannot disagree about whether the web surface starts. It writes nothing.

#### `bind addresses`

Reads `<address> present` for each address in `DUBPLATE_SERVE_BIND_ADDRS`
(default `127.0.0.1`), or `<address> not on any interface`, which fails. The
listener binds with `IP_FREEBIND`, so an address the host lacks raises no error
at start. This row is the only place its absence shows. A loopback address
always counts present.

#### `boundary probe`

Dials each bind address on `DUBPLATE_SERVE_PORT` (default 8080), from that same
address, and sends one request carrying every header the ingress strip exists
to remove. The probe route answers with counts only. Per address:

| Reading | Meaning | Fails? |
| --- | --- | --- |
| `refused` | Every count is zero: the strip removed each header. | No |
| `no listener` | The dial was refused or timed out. This is the expected reading before the unit starts. | No |
| `not probed (not on any interface)` | The host lacks the address. The `bind addresses` row fails for it. | No |
| `passed N identity header(s)`, `passed N forwarded-address header(s)`, or `passed the Connection-listed sentinel` | A header reached the route that the strip should have removed. | Yes |
| `unexpected answer: <status>`, `unexpected answer: timeout` | A status other than 200, a body that is not the three counts, or a connection that accepted and never answered. | Yes |
| `<address> not probed: a trusted proxy` | The address is in the trusted proxy set, and the route refuses a trusted peer by design. | No |

When any address answered, the row ends with `proves this server's strip,
not a proxy's`. The row dials the server directly, so it proves what this
server strips and says nothing about a proxy in front of it. Each dial and each answer is bounded to two seconds.

### The settings rows

One row per registered setting, after the boundary rows, in declaration order.
A row reads its value, its source (`default`, `env-seed`, or another channel),
and its apply mode, for example `"14" (source default, next operation); store
holds no settings yet`. A key the store holds that no code still declares
reads `no longer registered; ignored`.

`DUBPLATE_REAUTH_WINDOW` is a store-homed setting beside the others: how long
after an admin signs in that a sensitive setting change, or adding, removing,
or demoting a member, stays allowed over the API before it asks for a fresh
sign-in. It defaults to `2h`, accepts `5m` to `24h`, and
must stay below `DUBPLATE_SESSION_ABSOLUTE_CAP`. A window at or above the cap
refuses the start.

`serve` applies the environment to the store at every start, before it
listens. An environment value for a setting the store holds no row for becomes
that setting's row, with the source `env-seed`. An `env-seed` row follows the
environment afterward: a changed value replaces it, and removing the value
returns the setting to its default. A row an administrator set is never
changed by the environment. An environment value outside its bounds, or one
that does not parse, refuses the start, and `serve` binds no listener.

One consequence for the data-loss settings: when an environment line that
seeds one of them is removed, the setting returns to its registry default at
the next start. If that default is lower than the value that was in force, the
next backup prunes the difference. No floor above the registry's own bound
guards against it, so lower a retention setting on purpose, never by deleting
its line.

Before the first start, a setting you put in the environment reads `...; env
"20" seeds at serve's next start`. An environment value outside its bounds
fails `doctor` and `check` at that point, with the refusal `serve` would
raise, so the first start does not surprise you. A row you cannot vouch for
stays `not checked`.

## Mail

Sign-in links travel by SMTP submission through five keys:
`DUBPLATE_SMTP_HOST`, `DUBPLATE_SMTP_PORT`, `DUBPLATE_SMTP_USER`,
`DUBPLATE_SMTP_PASSWORD`, and `DUBPLATE_SMTP_FROM`. No default host ships.
The port picks the TLS mode: port 465 dials with TLS from the first byte, and
every other port opens a plain connection and requires STARTTLS before it
sends credentials. `doctor` does not dial the mail host, so send a test link
before you rely on it.

The server requires authenticated submission over TLS. It always signs in
with the user and password, and it never sends over a connection without
TLS. A relay that accepts mail without credentials, or without TLS, cannot be
configured. That rules out a local mail server on port 25 with no
authentication, so point the keys at a submission service that
authenticates.

### Example one: a mailbox provider, no Cloudflare account

Use the provider's submission port with STARTTLS and an app password. Most
providers refuse the account password and accept only an app password.

```sh
DUBPLATE_SMTP_HOST=smtp.mail.example.net
DUBPLATE_SMTP_PORT=587
DUBPLATE_SMTP_USER=music@example.org
DUBPLATE_SMTP_PASSWORD=<app password>
DUBPLATE_SMTP_FROM=music@example.org
```

The sending domain's SPF and DKIM records already cover the provider if the
domain sends its mail through that provider. Nothing new to publish.

### Example two: Cloudflare Email Service

Cloudflare Email Service accepts SMTP submission on `smtp.mx.cloudflare.net`
over implicit TLS. The user is the literal string `api_token`, and the
password is a Cloudflare API token with the Email Sending edit permission.
The service was in beta when this runbook was written.

```sh
DUBPLATE_SMTP_HOST=smtp.mx.cloudflare.net
DUBPLATE_SMTP_PORT=465
DUBPLATE_SMTP_USER=api_token
DUBPLATE_SMTP_PASSWORD=<Cloudflare API token>
DUBPLATE_SMTP_FROM=music@example.org
```

Onboard the sending domain in Email Service first. Until you do, Cloudflare
limits sends to destination addresses you have verified in your account.

Some VPS providers block outbound 465 by default; check from the host.
Cloudflare's endpoint offers 465 only, so on a host that blocks 465, use
example one.

### The mail ceiling and the unknown-address answer

`DUBPLATE_UNKNOWN_ADDRESS_ANSWER` is `kind` by default, and the second value is
`generic`. The default answers a member and a stranger differently, and `generic`
is the setting that hides who is a member: under it both get the same answer.
Three facts about the default follow from how the limits compose.

#### The kind answer reveals two things

For an address that no member holds, it returns the first admin's name and
address. It also tells a stranger whether an address is a member: a known
address gets a sent answer, an unknown one gets the admin contact. Choose
`generic` if who is among the members is private.

#### The enumeration bound is per client only

Each client address may make `DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR`
requests (default 20) before it is refused. A person with many client addresses
is not stopped by that bound alone.

#### The global mail ceiling is a group-wide lockout lever

`DUBPLATE_LINK_MAIL_CEILING_PER_HOUR` (default 60) bounds all outbound link
mail together. Whoever spends it stops sign-in mail for every member until the
hour turns. Under the kind answer, a known address past the ceiling gets a
rate-limited answer while an unknown address still gets the admin contact, so
the two stay distinguishable. With 11 or fewer members at the default rates,
per-address limits (`DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR`, default 5)
keep the total under the ceiling (11 times 5 is 55), so the lever works only on
a larger group. If you raise the per-address limit or grow past 11 members,
raise the ceiling with it.

## The trust boundary

The server holds two separate facts about a request. The peer
address is the TCP peer, the only input to any trust decision. The reported
client IP is derived, and it feeds the rate limiter and the log only. A
proxy in front of the server changes who the peer is, and this section says
what you must configure so the derived address stays honest.

### The peer address and the reported client IP

The peer address is the request's `RemoteAddr` with its port removed. No
header changes it. The trusted-proxy allowlist compares against it.

The reported client IP equals the peer address unless the trusted proxy class
is on. With the class on, the configured header is read only when the peer is
one of the configured proxy addresses. A header from any other peer is
discarded unread. The server parses a multi-hop value right to left and takes
the first hop that is not itself a configured proxy, because each proxy
appends to the right. It reads every line the header arrives on, since a proxy
that appends its own line is ordinary.

### Configuring a trusted proxy

```sh
DUBPLATE_TRUSTED_PROXY_ENABLED=true
DUBPLATE_TRUSTED_PROXY_ADDRS=192.0.2.10,192.0.2.11
DUBPLATE_TRUSTED_PROXY_HEADER=Cf-Connecting-Ip
```

`DUBPLATE_TRUSTED_PROXY_ENABLED=false` is direct mode, and with it false the
two other keys must stay unset. Direct mode is right only when browsers
connect to the server itself, so that each request's peer is the member's own
address. Behind any proxy or tunnel, direct mode keys every member on the
proxy's address, and one client can spend the per-client sign-in limits for
the whole group. Configure any front as a trusted proxy. A front on the same
host connects from loopback and may dial `127.0.0.1` or `::1`, depending on
how it resolves its upstream, so list both:

```sh
DUBPLATE_TRUSTED_PROXY_ADDRS=127.0.0.1,::1
```

Each address must be a single IP
address. A network such as `192.0.2.0/24` is refused: trust is pinned to the
addresses you observe, never widened to a range. A zone, an empty entry, and
the unspecified address are refused too. The header name must be an HTTP token
that the strip does not remove and the shell does not depend on. `Forwarded`
is refused, because its `for=` form is not an address list.

Other self-hosted servers take the same approach. Authelia's proxy documentation requires a reverse proxy to remove and replace
`X-Forwarded-For` when it arrives from an untrusted client, and warns that a
rule keyed on the header can otherwise be hijacked. It adds that a cloud proxy
must be made to do the same. Miniflux pairs its `AUTH_PROXY_HEADER` with an
explicit `TRUSTED_REVERSE_PROXY_NETWORKS` setting, which is the same shape as
this class.

### What the proxy must do

The proxy must overwrite the configured header, or append to it. A proxy
that passes the header through unchanged hands the client's own value to the
server, and the server then reads a client-chosen address from a trusted
peer. Check the proxy's documentation for this behavior, since the probe
cannot.

Two common defaults pass the client's value through:

- nginx's `proxy_pass` alone forwards a client's `X-Forwarded-For`
  unchanged. Set the header yourself, as the nginx example in
  [Example fronts](#example-fronts) does.
- Fastly sets `Fastly-Client-Ip` only when a request arrives without one, so
  a client can choose its value through Fastly. Name it as the configured
  header only if your Fastly service overwrites it on every request.

#### Every group 2 name survives from a trusted peer

The strip's second group holds the forwarded-address headers
(`X-Forwarded-For`, `X-Real-Ip`, `Forwarded`, the `X-Forwarded-Host`, `-Proto`,
and `-Prefix` names, and the client-address headers that Cloudflare, Fastly,
and Azure set). From a trusted peer none of them is removed, not only the
configured one. Nothing reads the others yet, and a later feature that does
must not assume they came from the proxy.

#### A hop may carry a port

`203.0.113.9:4711` and `[2001:db8::1]:443` keep their address and drop the
port, so a proxy that appends the client's port does not collapse every client
onto one rate-limit key. A hop with a zone, or one that parses as neither an
address nor an address and port, ends the scan and returns the peer address. A
bracketed address with no port, `[2001:db8::1]`, is such a hop.

#### The probe route has a residual under a proxy you did not configure

The `boundary-probe` route under the owned prefix answers only a peer that is
loopback or one of the bind addresses, and never a trusted proxy. A proxy on
the same host that you did not list as trusted connects from loopback, so it
can relay a stranger's request to the route. The route then answers that
stranger with three counts. No header value or address appears in the answer,
and a secret or second gate would cost more than that leak is worth. List a
same-host proxy as trusted, which closes the route to it.

#### Sign-in requires HTTPS

The session cookie's `__Host-` prefix requires the `Secure` attribute, and a
browser refuses a `Secure` cookie over plain HTTP on any origin but loopback.
An installation reached over HTTP would complete the sign-in flow, set no
cookie, and sign nobody in. The server therefore refuses a
`DUBPLATE_PUBLIC_URL` whose scheme is not `https` unless its host is a loopback
address, and `doctor` shows the source as
`unsatisfied: DUBPLATE_PUBLIC_URL invalid`. Terminate TLS at the proxy.

### Example fronts

Each example terminates TLS for `music.example.org`, forwards to the server on
`127.0.0.1:8080`, and pairs with these rows in `/etc/dubplate/server.env`.
Use the header row the example names.

```sh
DUBPLATE_TRUSTED_PROXY_ENABLED=true
DUBPLATE_TRUSTED_PROXY_ADDRS=127.0.0.1,::1
```

#### Caddy on the same host

Caddy obtains and renews the certificate itself. Its `reverse_proxy` sets
`X-Forwarded-For` to the client's address, so the header row is
`DUBPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For`.

```text
music.example.org {
	reverse_proxy 127.0.0.1:8080
}
```

#### nginx on the same host

Set both headers explicitly. `X-Forwarded-For` replaces any value the client
sent with the client's address. `Host` keeps the name the browser used, which
the origin check compares against when a browser sends no `Sec-Fetch-Site`
header. The header row is `DUBPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For`.

```nginx
server {
    listen 443 ssl;
    server_name music.example.org;
    ssl_certificate     /etc/ssl/music.example.org/fullchain.pem;
    ssl_certificate_key /etc/ssl/music.example.org/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $remote_addr;
    }
}
```

#### cloudflared on the same host

A Cloudflare Tunnel needs no open inbound port, and Cloudflare's edge
terminates TLS. The edge sets `Cf-Connecting-Ip` to the client's address, so
the header row is `DUBPLATE_TRUSTED_PROXY_HEADER=Cf-Connecting-Ip`. In the
tunnel's `config.yml`:

```yaml
tunnel: <tunnel id>
credentials-file: /etc/cloudflared/<tunnel id>.json
ingress:
  - hostname: music.example.org
    service: http://127.0.0.1:8080
  - service: http_status:404
```

If `cloudflared` runs in a container instead, its connections come from the
container's address. Give the container a fixed address and list that
address in place of the two loopback addresses.

#### Keep the sign-in link out of the access log

A sign-in link carries its token in the query string, `?t=`. The server never
logs the query, but a front's access log may record the full request line.
Anyone who reads that log before the member clicks could use the link first.
Drop the query from the front's access log. In nginx, define a format that
logs `$uri`, which excludes the query, in place of `$request`, and name it on
the server's `access_log` line:

```nginx
log_format noquery '$remote_addr - $remote_user [$time_local] '
                   '"$request_method $uri $server_protocol" $status '
                   '$body_bytes_sent "$http_referer" "$http_user_agent"';
access_log /var/log/nginx/music.example.org.log noquery;
```

Caddy writes no access log unless a `log` directive asks for one. If you add
one, delete the `t` parameter from the logged URI with the log's `filter`
format.

### A CDN in front of the server

A CDN such as Cloudflare's proxied DNS, Fastly, or Bunny reaches the origin
from many edge addresses, published only as ranges. The trusted proxy class
refuses a range, so the server cannot trust a CDN's edges directly, and a
CDN pointed straight at the server's public address is not supported. Put a
local proxy between the CDN and the server instead:

1. Have the local proxy accept the CDN's client-address header only from the
   CDN's published ranges. In nginx, that is one `set_real_ip_from` line per
   range and `real_ip_header` naming the CDN's header. In Caddy, it is the
   `trusted_proxies` and `client_ip_headers` options of the `servers` global
   option.
2. Have the local proxy set `X-Forwarded-For` to the client address it
   derived. In nginx, the `proxy_set_header X-Forwarded-For $remote_addr;`
   line from the example does this, since `$remote_addr` then holds the
   derived address.
3. Trust only the local proxy, as in [Example fronts](#example-fronts).

The ranges change, so keep the local proxy's list current with the CDN's
published list.

## Browser and presentation notes

Two behaviors differ from what an operator might assume. Both are decided.

### `Referrer-Policy: no-referrer` and older browsers

Every response carries `Referrer-Policy: no-referrer`. The origin check reads
`Sec-Fetch-Site` first and falls back to comparing the `Origin` header's host
with `Host` only when `Sec-Fetch-Site` is absent. A browser that sends no
`Sec-Fetch-Site` header may serialize `Origin` as `null` on a same-origin POST
under `no-referrer`, and the check then refuses the request with 403. The
sign-in form fails for that browser. Current browsers send `Sec-Fetch-Site` and
are unaffected.

### A branding color applies in both color schemes

`DUBPLATE_BRANDING` overrides a theme variable in every theme the bundle
carries, light and dark. A value you set for `--color-neutral` replaces the dark
scheme's value as well, and the document's `theme-color` meta takes the same
value for both. Pick colors that work against both palettes, or declare the
variables the two palettes need separately in the operator stylesheet. A
`--color-neutral` value other than a three- or six-digit hex color leaves every
document with no `theme-color` meta and writes one warning at startup.

## Backup

`dubplate-server backup` syncs the library to an rclone remote and sends a
database snapshot after it. These prerequisites sit outside the binary, so
`doctor` and the unit file are where you meet them.

1. Set `KillMode=mixed` for any unit that runs rclone. With the default
   `KillMode=control-group`, a stop sends SIGTERM to every process in the
   unit's group at once, so rclone can die before the verb records its own
   cancel. `KillMode=mixed` sends SIGTERM to the main process alone and
   SIGKILL to the rest after the timeout, which lets the verb cancel through
   its job record first (`systemd.kill(5)`).
2. Install rclone v1.74.3 or later. The backup's guard relies on v1.74.3's
   behavior, where a move into `--backup-dir` counts against `--max-delete` as
   a delete. Nothing in the binary reads the installed version, so check it
   yourself: `rclone version`.
3. Wrap the backup remote in an rclone `crypt` remote. A snapshot carries
   member data, and a plain remote stores it unencrypted at the provider.
   `doctor`'s `backup remote` row appends `the backup remote is not wrapped in
   an rclone crypt remote; a snapshot carries member data unencrypted at the
   provider` to an `ok` for any other type. Declare the remote through
   environment variables, uppercasing its name:

   ```sh
   DUBPLATE_BACKUP_ENABLED=true
   DUBPLATE_BACKUP_REMOTE=vault:backups
   RCLONE_CONFIG_VAULT_TYPE=crypt
   RCLONE_CONFIG_VAULT_REMOTE=<the wrapped remote>:<path>
   ```

   A remote whose `_TYPE` is unset reads `invalid remote:
   RCLONE_CONFIG_VAULT_TYPE is not set`. A crypt remote with no `_REMOTE`
   reads `invalid remote: RCLONE_CONFIG_VAULT_REMOTE is not set`. Keep one
   remote path per server.
4. Know the exit-7 remedy. The guard gives rclone `--max-delete` equal to the
   catalog's track count. With zero tracks that bound is 0, and because a move
   into the archive counts as a delete, any deletion from the library tree
   that the remote already holds fails every run with exit code 7 (`failed to
   delete 1 files`), until the catalog holds tracks. A modified file still
   archives, since only a removal reaches the delete count. To clear it by
   hand, archive the file on the remote yourself, using the same archive
   layout the sync writes:

   ```sh
   rclone moveto vault:backups/library/<path> vault:backups/archive/<stamp>/<path>
   ```

   Run it for each file the failed run names, then run the backup again.
5. Know the exit statuses. `dubplate-server backup` exits 0 on success, 1 on a
   failure or a refusal before it starts, 69 when `DUBPLATE_BACKUP_ENABLED` is
   off, 75 when a backup is already running, and 130 when canceled by any
   means.

### After a restore

A restore brings back the member and settings tables as they stood when the
snapshot was taken, and nothing else undoes a removal or a demotion made after
it. A member removed since the snapshot comes
back, and an admin flag a demotion cleared is set again. Re-run any removal or
demotion made since the snapshot with `member remove` or `member demote`, and
re-run any `settings reset` the same way. Then check `member list`.

## Deploy

`scripts/deploy.sh` deploys the maintainers' own estate, and a stranger
installs by hand (see [Install order](#install-order)). Its HOST defaults to the
maintainers' server, and it reads the maintainers' workstation age store. The
template it renders carries the maintainers' public URL and site name. This section describes the script for the people who run it.

The script reports by default and writes the box only under `--apply`. Run it
as a check first, read what it prints, then apply.

```sh
bash scripts/deploy.sh --check <host>
bash scripts/deploy.sh --apply <host>
```

With no flag, the script behaves as `--check`, so a dropped flag is safe.
Passing both flags is refused. Every mode first refuses a working tree with a
modified tracked file, or with an untracked or ignored file under `cmd`,
`internal`, `web`, `go.mod`, or `go.sum`. The
build's own output, `web/node_modules`, `web/.svelte-kit`, and `web/build`,
is exempt. Run the script from a fresh worktree of the commit you deploy, so
that no stray file stops it. It then builds the `web/` bundle with
`npm ci --ignore-scripts` and the Vite build, and cross-compiles the stamped
binary, all before the first remote write. Each build runs with only `PATH`
and `HOME` in its environment, so no secret the script loaded reaches a
build tool. A dirty tree or a failed build stops the run with nothing on the
box. The build ships no Node and no
Playwright, so the rendered browser gates run on the workstation and never on
the box.

`--check` stops at the render when a bare `NAME=` row names a secret the store
does not hold. Add each missing name to your secrets store, then run `--check`
again.

### A staged binary swap

The new binary installs its transliteration table before it replaces the
live one. A unit can start the binary at any moment, and a swap followed by an
install would leave a window in which a verb finds a new binary with no table
and refuses to start. The staged
order closes it:

1. the script copies the new binary to `/usr/local/bin/dubplate-server.new`;
2. it runs `unidecode install` from that file on the box, under `systemd-run`
   with the property list `--check` prints (the env file is not installed until
   a later step). The run may write only `/var/lib/dubplate/unidecode`, which
   the script creates first, so the install cannot touch the auth store or the
   store beside it;
3. only a verified install is moved over the live path with `mv`.

A failed install stops the deploy with the old binary still live, and leaves `dubplate-server.new` on the box, which the next deploy's
copy overwrites. Remove it by hand if you do not plan another deploy.

Before the first deploy that carries this swap:

1. **Run `--check` first, and read the staged install's line.** It prints
   `would run: systemd-run ... unidecode install`, and that line is what the
   apply runs. The install downloads the pinned wheel from PyPI on the box, so
   a `--check` run cannot exercise the download or the table build, and the
   apply takes longer than the check suggests. The box's systemd has not been
   shown to accept the whole property list until the first deploy.
2. **Keep the workstation's verified wheel cache.** `make unidecode-wheel`
   fills it in `~/.cache/unidecode` and is a no-op once the cached file
   verifies. It is the offline copy, and the `--from` source if PyPI ever
   removes the 1.4.0 release.
3. **Plan the rollback before the apply, not after it.** Redeploying the
   previous commit is not a rollback. The binary and the environment file move
   together: the old binary refuses the new environment's keys, and the new
   binary refuses the old environment (`Load` requires the selected source's
   keys and rejects every unknown `DUBPLATE_` name). A redeploy of an older
   commit renders the older environment with the older template, which an
   in-between state can leave mismatched. The rollback is a set of copies kept
   from before the apply.

### The rollback copies

Immediately before an apply, on the box, stop any timer that runs a
`dubplate-server` verb, confirm its service is not running, and keep copies of
every file the apply rewrites together:

```sh
cp -p /usr/local/bin/dubplate-server /usr/local/bin/dubplate-server.prev
cp -p /etc/dubplate/server.env /etc/dubplate/server.env.prev
install -d -m 0700 /root/dubplate-units.prev
cp -a /etc/systemd/system/dubplate-* /root/dubplate-units.prev/
```

The apply also rewrites the unit files, so the binary and the environment
file are not the whole set. The unit copies sit outside
`/etc/systemd/system`, where systemd would read them as units.

The copies do not cover the directories the apply creates or re-owns (the
state directory and its leaves), the transliteration table the staged install
writes, or which units are enabled. A rollback leaves those as the apply left
them.

Each copy keeps its mode and its owner, and the environment file stays 0600
and root-owned. Record the `.prev` binary's version with
`/usr/local/bin/dubplate-server.prev version`.

To roll back:

1. Stop the serve unit if it runs or loops
   (`systemctl stop dubplate-server-serve.service`, then
   `systemctl reset-failed dubplate-server-serve.service`, since
   `Restart=on-failure` loops a failing start).
2. Copy the `.prev` binary and environment file over the live ones together
   (`cp -p <file>.prev <file>` for each), and run `dubplate-server version` to
   see the recorded stamp.
3. Restore the units, then reload systemd:

   ```sh
   rsync -a --delete --checksum --include='dubplate-*' --exclude='*' \
     /root/dubplate-units.prev/ /etc/systemd/system/
   systemctl daemon-reload
   ```

   `--checksum` compares contents, so rsync restores a file the apply changed
   even when its size and timestamp match the copy's. The unit restore also
   deletes a `dubplate-*` unit the apply added, because the copy lacks it. The
   `--exclude='*'` filter protects every other file in
   `/etc/systemd/system` from the delete.
4. Start the serve unit if it ran before, and any timer you stopped.

Leave the copies on the box until the next attended deploy replaces them.

Never preview a rollback by running an older commit's `deploy.sh`. A script
from before commit `c179c99` applies when it gets no flag, so a run meant as a
check writes the box.
