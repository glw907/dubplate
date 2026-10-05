# dubplate-server runbook

This runbook is for the person who installs and operates `dubplate-server`.
It covers the install order, what each `doctor` row proves, how to wire mail,
where the trust boundary sits, a few browser and presentation notes, the
backup prerequisites, and the deploy sequence. Each section opens with a short
brief, so you can read only the one you need.

Every command here was run against a workstation scratch state directory or a
fixture, and its flags match the binary's `--help`. Host names, addresses, and
paths in examples are placeholders: `music.example.org`, `192.0.2.10`, and
`/var/lib/dubplate`.

## Install order

**Brief.** Run the steps in the order below. Each step depends on the one
before it, and `doctor` run too early correctly reports a server that cannot
start yet.

1. **Install the binary and its units.** `scripts/deploy.sh` does this (see
   [Deploy](#deploy)). The binary lands at `/usr/local/bin/dubplate-server`
   and the unit `dubplate-server-serve.service` lands in
   `/etc/systemd/system`, installed and not enabled.
2. **Render the environment.** `deploy.sh` renders `/etc/dubplate/server.env`
   from `env/dubplate-server.env.template`. A bare `NAME=` row resolves from
   the secrets store at render time, and the render stops, naming every name
   it could not resolve. Without the secrets store, write the same names into
   the file by hand. The file's mode is 0600 and its owner is root.
3. **Create the state leaves.** The unit grants write access to exactly two
   leaves, `/var/lib/dubplate/auth` and `/var/lib/dubplate/roster`.
   `deploy.sh` creates both, owned by the service user with mode 0700. A leaf
   missing from the host stops the unit at start with the namespace failure
   `226/NAMESPACE`, because systemd cannot mount a path that does not exist.
4. **Install the transliteration table.** The matcher cannot score without it,
   and `doctor` fails the `unidecode` row until it exists. `deploy.sh` runs
   this for you. By hand:

   ```sh
   dubplate-server unidecode install
   ```

   Offline, pass the wheel you cached with `make unidecode-wheel`:

   ```sh
   dubplate-server unidecode install --from ~/.cache/unidecode/Unidecode-1.4.0-py3-none-any.whl
   ```

5. **Add the first admin.** The roster file does not exist until a member is
   added, and `doctor` fails the `roster` row until the roster names an admin.
   The verb requires `--name` and `--streamer-username` as well as the address:

   ```sh
   dubplate-server member add admin@example.org --admin --name "Ada Admin" --streamer-username ada
   dubplate-server member list
   ```

   Run both as the service user through `systemd-run`, carrying the serve
   unit's own filesystem and privilege properties. The unit's operator rule
   is the same one the shadow unit states: a hand-run verb never runs as the
   bare binary under root, because the unit's guarantees belong to the unit.
   On the box:

   ```sh
   systemd-run --wait --pipe --collect \
     -p User=music -p EnvironmentFile=/etc/dubplate/server.env \
     -p NoNewPrivileges=yes -p PrivateTmp=yes -p ProtectSystem=strict \
     -p ProtectHome=yes -p ProtectKernelTunables=yes \
     -p ProtectKernelModules=yes -p ProtectControlGroups=yes \
     -p RestrictNamespaces=yes -p RestrictRealtime=yes \
     -p RestrictSUIDSGID=yes -p LockPersonality=yes \
     -p MemoryDenyWriteExecute=yes -p SystemCallFilter=@system-service \
     -p SystemCallArchitectures=native -p UMask=0027 \
     -p 'ReadWritePaths=/var/lib/dubplate/auth /var/lib/dubplate/roster' \
     /usr/local/bin/dubplate-server member add admin@example.org --admin \
       --name "Ada Admin" --streamer-username ada
   ```

   Swap the last line for `... dubplate-server member list`, or `... doctor`,
   to run those verbs the same way.
6. **Run `doctor`.** Use the same `systemd-run` command with `doctor` as the
   verb and **without** `RestrictAddressFamilies`. The reason, in one
   sentence: the `bind addresses` row reads the host's interface table over a
   netlink socket, and `RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX`
   denies it. With the restriction in place the row reads `not checked: route
   ip+net: netlinkrib: address family not supported by protocol` and fails
   `doctor`, so the unit itself keeps the restriction and only this check
   leaves it off. `doctor` exits 0 only when every row passes.
7. **Start the unit.**

   ```sh
   systemctl start dubplate-server-serve.service
   ```

   Run `doctor` once more with the unit up. `boundary probe` should now read
   `refused`.

`dubplate-server check` is the short form of the configuration wall. It prints
`configuration: ok` and the store state, and exits 0 when `Load` accepts the
environment. It runs none of the checks that need the filesystem or the
network, so it never replaces step 6.

## The doctor rows

**Brief.** `doctor` prints one line per row and exits 1 when any row fails.
This section lists the rows by what they prove and what they cannot. A row
that prints `not checked` adds no failure of its own, because the row that
owns the cause reports it.

### The wall and the sources

Each owned key, reserved class, integration class, and threshold prints one
row, in that order. An owned key reads `ok`, `missing`, or `invalid: <reason>`.
The identity source reads `active`, `inactive`, or `unsatisfied: <key>
missing` naming every failing key, and never prints a value, since the source's
keys include secrets. The `roster` row prints only when it fails: `missing`
before step 5, `invalid: ...` for a duplicate or unparseable record, and `no
admin; run dubplate-server member add --admin` when nobody holds the flag.

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

- **`client ip`** reads `direct` when the trusted proxy class is off, and
  `trusted proxy: <header> from <n> address(es)` when it is on and its values
  resolve. It reads `invalid: ...` and fails when an address entry is empty, is
  not an IP address, is a network, carries a zone, or is the unspecified
  address, or when the header name is one the strip removes or the shell
  depends on. It names a count, never an address. It reads `not checked` when
  the wall refused the environment.
- **`web`** runs the same web resolution serve runs at startup: the theme
  match and palette, the branding value, and the operator stylesheet. It reads
  `ok` or `invalid: <reason>`. Because it calls the same function, `doctor`
  and `serve` cannot disagree about whether the web surface starts. It writes
  nothing.
- **`bind addresses`** reads `<address> present` for each address in
  `DUBPLATE_SERVE_BIND_ADDRS` (default `127.0.0.1`), or `<address> not on any
  interface`, which fails. The listener binds with `IP_FREEBIND`, so an address
  the host lacks raises no error at start. This row is the only place its
  absence shows. A loopback address always counts present.
- **`boundary probe`** dials each bind address on `DUBPLATE_SERVE_PORT`
  (default 8080), from that same address, and sends one request carrying every
  header the ingress strip exists to remove. The probe route answers with
  counts only. Per address:

  | Reading | Meaning | Fails? |
  | --- | --- | --- |
  | `refused` | Every count is zero: the strip removed each header. | No |
  | `no listener` | The dial was refused or timed out. This is the expected reading before the unit starts. | No |
  | `not probed (not on any interface)` | The host lacks the address. The `bind addresses` row fails for it. | No |
  | `passed N identity header(s)`, `passed N forwarded-address header(s)`, or `passed the Connection-listed sentinel` | A header reached the route that the strip should have removed. | Yes |
  | `unexpected answer: <status>`, `unexpected answer: timeout` | A status other than 200, a body that is not the three counts, or a connection that accepted and never answered. | Yes |
  | `<address> not probed: a trusted proxy` | The address is in the trusted proxy set, and the route refuses a trusted peer by design. | No |

  When any address answered, the row ends with `proves this server's strip,
  not a proxy's`. That limit is the whole point of the row: it dials the
  server directly, so it proves what this server strips and says nothing about
  a proxy in front of it. Each dial and each answer is bounded to two seconds.

### The settings rows

One row per registered setting, after the boundary rows, in declaration order.
A row reads its value, its source (`default`, `env-seed`, or another channel),
and its apply mode, for example `"14" (source default, next operation); store
holds no settings yet`. A key the store holds that no code still declares
reads `no longer registered; ignored`.

While the store is absent or holds no settings yet, a setting you put in the
environment reads `...; env "20" seeds at serve's next start`. **That
statement is not yet true.** `serve` does not apply the seed today, so an
environment value for a setting does not reach the store until a later release
wires the seed into `serve`. Until then, treat such a row as a hand-set value
that has no effect, and clear it from the environment.

## Mail

**Brief.** Sign-in links travel by SMTP submission through five keys:
`DUBPLATE_SMTP_HOST`, `DUBPLATE_SMTP_PORT`, `DUBPLATE_SMTP_USER`,
`DUBPLATE_SMTP_PASSWORD`, and `DUBPLATE_SMTP_FROM`. No default host ships.
The port picks the TLS mode: port 465 dials with TLS from the first byte, and
every other port opens a plain connection and requires STARTTLS before it
sends credentials. `doctor` does not dial the mail host, so send a test link
before you rely on it.

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

**Outbound port 465.** Some VPS providers block outbound 465 by default,
Hetzner among them. The route this project verified on its own box was 587
with STARTTLS, and Cloudflare's endpoint offers 465 only. On a host that
blocks 465, use example one. Check the port from the host before you choose.

### The mail ceiling and the unknown-address answer

`DUBPLATE_UNKNOWN_ADDRESS_ANSWER` is `kind` by default, and the second value is
`generic`. Three facts about the default follow from how the limits compose.

- **The kind answer reveals two things.** For an address that is not on the
  roster, it returns the first admin's name and address. It also tells a
  stranger whether an address is a member: a known address gets a sent answer,
  an unknown one gets the admin contact. Choose `generic` if a member's
  presence on the roster is private.
- **The enumeration bound is per client only.** Each client address may make
  `DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR` requests (default 20) before it
  is refused. A person with many client addresses is not stopped by that bound
  alone.
- **The global mail ceiling is a group-wide lockout lever.**
  `DUBPLATE_LINK_MAIL_CEILING_PER_HOUR` (default 60) bounds all outbound link
  mail together. Whoever spends it stops sign-in mail for every member until
  the hour turns. Under the kind answer, a known address past the ceiling gets
  a rate-limited answer while an unknown address still gets the admin contact,
  so the two stay distinguishable. With 11 or fewer members at the default
  rates, per-address limits (5 an hour each) keep the total under the ceiling
  (11 times 5 is 55), so the lever works only on a larger roster. If you raise
  the per-address limit or grow the roster past 11, raise the ceiling with it.

## The trust boundary

**Brief.** The server holds two separate facts about a request. The peer
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

Running without a proxy is a chosen state, so
`DUBPLATE_TRUSTED_PROXY_ENABLED=false` is written out explicitly, and with it
false the two other keys must stay unset. Each address must be a single IP
address. A network such as `192.0.2.0/24` is refused: trust is pinned to the
addresses you observe, never widened to a range. A zone, an empty entry, and
the unspecified address are refused too. The header name must be an HTTP token
that the strip does not remove and the shell does not depend on. `Forwarded`
is refused, because its `for=` form is not an address list.

This is the field's normal answer, not an invention of this project.
Authelia's proxy documentation requires a reverse proxy to remove and replace
`X-Forwarded-For` when it arrives from an untrusted client, and warns that a
rule keyed on the header can otherwise be hijacked. It adds that a cloud proxy
must be made to do the same. Miniflux pairs its `AUTH_PROXY_HEADER` with an
explicit `TRUSTED_REVERSE_PROXY_NETWORKS` setting, which is the same shape as
this class.

### What the proxy must do

**The proxy must overwrite the configured header, or append to it.** A proxy
that passes the header through unchanged hands the client's own value to the
server, and the server then reads a client-chosen address from a trusted
peer. Check the proxy's documentation for this behavior, since the probe
cannot.

Four more facts apply once a proxy is trusted:

- **Every group 2 name survives from a trusted peer.** The strip's second
  group holds the forwarded-address headers (`X-Forwarded-For`, `X-Real-Ip`,
  `Forwarded`, the `X-Forwarded-Host`, `-Proto`, and `-Prefix` names, and
  the client-address headers that Cloudflare, Fastly, and Azure set). From a
  trusted peer none of them is removed, not only the configured one. Nothing
  reads the others yet, and a later feature that does must not assume they
  came from the proxy.
- **A hop may carry a port.** `203.0.113.9:4711` and `[2001:db8::1]:443` keep
  their address and drop the port, so a proxy that appends the client's port
  does not collapse every client onto one rate-limit key. A hop with a zone, or
  one that parses as neither an address nor an address and port, ends the scan
  and returns the peer address. A bracketed address with no port,
  `[2001:db8::1]`, is such a hop.
- **The probe route has a residual under a proxy you did not configure.** The
  `boundary-probe` route under the owned prefix answers only a peer that is
  loopback or one of the bind addresses, and never a trusted proxy. A proxy on
  the same host that you did not list as trusted connects from loopback, so it
  can relay a stranger's request to the route. The route then answers that
  stranger with three counts. No header value or address appears in the
  answer, and a secret or second gate would cost more than that leak is worth.
  List a same-host proxy as trusted, which closes the route to it.
- **Sign-in requires HTTPS.** The session cookie's `__Host-` prefix requires
  the `Secure` attribute, and a browser refuses a `Secure` cookie over plain
  HTTP on any origin but loopback. An installation reached over HTTP would
  complete the sign-in flow, set no cookie, and sign nobody in. The server
  therefore refuses a `DUBPLATE_PUBLIC_URL` whose scheme is not `https`
  unless its host is a loopback address, and `doctor` shows the source as
  `unsatisfied: DUBPLATE_PUBLIC_URL invalid`. Terminate TLS at the proxy.

### The forward-auth topology

This topology is not built yet. It arrives with the streamer driver (rung 12),
and this section states its boundary in advance so an adopter plans for it.

In a forward-auth deployment the proxy asks a verify endpoint whether to admit
a request, and the server is not in the streamer's request path. Every strip
the server performs is then inert for the streamer, so the boundary is the
adopter's proxy. The driver will publish its strip set. Your proxy must:

1. strip that set from every inbound request before it makes the verify
   subrequest;
2. strip the same set again before it forwards a request to the streamer; and
3. send the streamer only requests that came through the proxy, with the
   streamer configured to trust that proxy and no other source.

`doctor`'s boundary probe will prove the server's half of this, and nothing
will prove the proxy's.

## Browser and presentation notes

**Brief.** Two behaviors differ from what an operator might assume. Neither
is a defect to fix, and both are decided.

- **`Referrer-Policy: no-referrer` and older browsers.** Every response
  carries `Referrer-Policy: no-referrer`. The origin check reads
  `Sec-Fetch-Site` first and falls back to comparing the `Origin` header's
  host with `Host` only when `Sec-Fetch-Site` is absent. A browser that sends
  no `Sec-Fetch-Site` header may serialize `Origin` as `null` on a same-origin
  POST under `no-referrer`, and the check then refuses the request with 403.
  The sign-in form fails for that browser. Current browsers send
  `Sec-Fetch-Site` and are unaffected.
- **A branding color applies in both color schemes.** `DUBPLATE_BRANDING`
  overrides a theme variable in every theme the bundle carries, light and
  dark. A value you set for `--color-neutral` replaces the dark scheme's value
  as well, and the document's `theme-color` meta takes the same value for both.
  Pick colors that work against both palettes, or declare the variables the
  two palettes need separately in the operator stylesheet. A `--color-neutral`
  value other than a three- or six-digit hex color leaves every document with
  no `theme-color` meta and writes one warning at startup.

## Backup

**Brief.** `dubplate-server backup` syncs the library to an rclone remote and
sends a database snapshot after it. These prerequisites sit outside the
binary, so `doctor` and the unit file are where you meet them.

- **`KillMode=mixed` for any unit that runs rclone.** With the default
  `KillMode=control-group`, a stop sends SIGTERM to every process in the
  unit's group at once, so rclone can die before the verb records its own
  cancel. `KillMode=mixed` sends SIGTERM to the main process alone and
  SIGKILL to the rest after the timeout, which lets the verb cancel through
  its job record first (`systemd.kill(5)`).
- **rclone v1.74.3 or later.** The backup's guard relies on v1.74.3's
  behavior, where a move into `--backup-dir` counts against `--max-delete` as
  a delete. Nothing in the binary reads the installed version, so check it
  yourself: `rclone version`.
- **A crypt remote.** Wrap the backup remote in an rclone `crypt` remote. A
  snapshot carries member data, and a plain remote stores it unencrypted at the
  provider. `doctor`'s `backup remote` row appends `the backup remote is not
  wrapped in an rclone crypt remote; a snapshot carries member data
  unencrypted at the provider` to an `ok` for any other type. Declare the
  remote through environment variables, uppercasing its name:

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
- **The exit-7 remedy.** The guard gives rclone `--max-delete` equal to the
  catalog's track count. With zero tracks that bound is 0, and because a move
  into the archive counts as a delete, any deletion from the library tree
  that the remote already holds fails every run with exit code 7 (`failed to
  delete 1 files`), until the catalog holds tracks. A modified file still
  archives, since only a removal reaches the delete count.
  To clear it by hand, archive the file on the remote yourself, using the same
  archive layout the sync writes:

  ```sh
  rclone moveto vault:backups/library/<path> vault:backups/archive/<stamp>/<path>
  ```

  Run it for each file the failed run names, then run the backup again.
- **The exit statuses.** `dubplate-server backup` exits 0 on success, 1 on a
  failure or a refusal before it starts, 69 when `DUBPLATE_BACKUP_ENABLED` is
  off, 75 when a backup is already running, and 130 when canceled by any
  means.

## Deploy

**Brief.** `scripts/deploy.sh` reports by default and writes the box only
under `--apply`. Run it as a check first, read what it prints, then apply.

```sh
bash scripts/deploy.sh --check <host>
bash scripts/deploy.sh --apply <host>
```

With no flag, the script behaves as `--check`, so a dropped flag is safe.
Passing both flags is refused. Every mode first refuses a working tree with a
modified tracked file or an untracked file under `web/`, then builds the
`web/` bundle with `npm ci` and the Vite build, and cross-compiles the
stamped binary, all before the first remote write. A dirty tree or a failed
build stops the run with nothing on the box. The build ships no Node and no
Playwright, so the rendered browser gates run on the workstation and never on
the box.

`--check` stops at the render when a bare `NAME=` row names a secret the store
does not hold. Mint each one first, then re-run it.

### A staged binary swap

The new binary installs its transliteration table **before** it replaces the
live one. A timer that scores, such as the shadow timer, can fire at any
moment, and a swap followed by an install would leave a window in which a
scoring verb finds a new binary with no table and refuses to start. The staged
order closes it:

1. the script copies the new binary to `/usr/local/bin/dubplate-server.new`;
2. it runs `unidecode install` from that file on the box, under `systemd-run`
   carrying the shadow unit's hardening, with `Environment=` setting only
   `DUBPLATE_STATE_DIR` (the env file is not installed until a later step);
3. only a verified install is moved over the live path with `mv`.

A failed install stops the deploy with the old binary still live and still
scoring, and leaves `dubplate-server.new` on the box, which the next deploy's
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
the three files that change together:

```sh
cp -p /usr/local/bin/dubplate-server /usr/local/bin/dubplate-server.prev
cp -p /etc/dubplate/server.env /etc/dubplate/server.env.prev
cp -p /etc/dubplate/env /etc/dubplate/env.prev
```

Each copy keeps its mode and its owner, and the two environment files stay
0600 and root-owned. Record the `.prev` binary's version with
`/usr/local/bin/dubplate-server.prev version`. To roll back, stop the serve
unit if it runs or loops (`systemctl stop dubplate-server-serve.service`,
then `systemctl reset-failed dubplate-server-serve.service`, since
`Restart=on-failure` loops a failing start), copy all three `.prev` files over
the live ones together, run `dubplate-server version` to see the recorded
stamp, and restart the timers you stopped. If the apply replaced
`/etc/dubplate/env`, bring the compose stack up again so it reads the
restored file. Leave the `.prev` copies on the box until the next attended
deploy replaces them.
