# Rung 5d plan review: domain risk lens (auth trust boundary, first production deploy)

Reviewed: `docs/superpowers/plans/2026-09-25-rung-5d-boundary-deployment-close.md` at `af7d941`
(revision 1). Lens: domain risk only. Weighed against the rung 5 spec
(`docs/specs/2026-09-08-rung-5-identity-auth.md`), the build spec, and the repo
CLAUDE.md safety rails. Box read 2026-09-26 03:02 UTC, read-only (`ssh -o BatchMode=yes`).

**Counts: 1 blocker, 4 major, 12 minor.** Every blocker and major sits in Task 3 (the probe) or
Task 10 (the sitting). Tasks 1, 2, and 5 through 9 carry minors only.

## Evidence gathered (so the fold does not re-derive it)

- Box `/etc/dubplate/server.env` has exactly ten rows (the rung 4 set) and no
  `DUBPLATE_IDENTITY_SOURCE`. The box binary is `pre-greenfield-321-gbe27434` (installed
  2026-09-08). The repo template already carries `DUBPLATE_IDENTITY_SOURCE=magic-link`, and
  `config.Load` requires the selected source's keys (`internal/config/config.go:324-351`). It
  also rejects every unknown `DUBPLATE_*` name (`config.go:245-298`). The drafter's finding
  stands: a deploy of any commit from 5b-i onward, without the whole magic-link set, breaks the
  live shadow.
- `dubplate-shadow.timer` is `OnCalendar=*:0/5`, and its last run returned `Result=success`.
  `dubplate-disk-check` and `dubplate-navidrome-check` also fire about every five minutes, and
  `dubplate-import.timer` fires about every fifteen.
- Nothing listens on 8080. The only TCP listeners are 22 and `127.0.0.1:4533` (docker-proxy).
  SELinux is Enforcing. systemd is 257. `/var/lib/dubplate` is `music:music 0755` and holds only
  `shadow`.
- Repo against box: `compose.yaml`, every unit, `navidrome.toml`, `beets-config.yaml`,
  `tmpfiles-dubplate.conf`, and `cloudflared-tunnel-config.json` match. **`config/contributors.yaml`
  and `config/reject-email.txt` differ** ("family" to "group" copy, `53ecd0a`).
- `deploy.sh --check` is read-only on the box. It runs only `rsync -n`, `ssh cat`, and `test -d`,
  and every `run_remote` call echoes. Apply is the default when no flag is given
  (`scripts/deploy.sh:81-95`).
- Age store, names only: `DUBPLATE_SMTP_HOST/PORT/USER/FROM/PASSWORD` are present. Neither
  magic-link secret is present. `~/.local/secrets` also exports about 40 incumbent `DUBPLATE_*`
  names, `DUBPLATE_SERVE_BIND_ADDRS`, `DUBPLATE_SERVE_PORT`, and `DUBPLATE_STATE_DIR` among them.
- SEC's review text counts 6, 13, and 12 findings, and UNI's counts 19. The total, 50, matches
  spec `:2003-2006`.
- Measured on the workstation: `net.InterfaceAddrs` under
  `RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX` returns
  `route ip+net: netlinkrib: address family not supported by protocol`.
  `textproto.CanonicalMIMEHeaderKey("Remote_User")` returns `Remote_user`.

## Blocker

### B1. Task 10 cannot run as written on this box, and it has no rollback

Location: plan `:928-930`, `:941-949`, `:963-965`; `scripts/deploy.sh:337-343` and `:405-414`.

The defect has four parts, and each one is on the live box.

1. **The quiet-box check cannot pass.** Step 2 requires "the shadow timer's next firing more than
   ten minutes away". The timer fires every five minutes, so the check always fails. At that
   point the conductor either stops forever or improvises on production.
2. **Apply puts a new binary beside an old env.** `deploy.sh` swaps the binary (`:337-343`, or
   translit's staged `mv`), runs the smoke test, and only then installs `server.env`
   (`:405-414`). A shadow run that fires in that window loads the post-5d binary against the old
   ten-row env and fails at `Load`. Worse, an abort between the two steps (an ssh drop, a smoke
   failure after translit's swap) leaves the box running the new binary with the old env. The
   shadow then fails every five minutes, unmonitored ("no monitoring at rung 4",
   `dubplate-shadow.service:47-51`). Translit's plan reasons only about the old binary refusing
   the new env (translit plan `:1263-1270`). The opposite window is 5d's.
3. **The failure rule leaves the shadow broken.** Line 963 says the box stays "as the last
   successful step left it" and forbids "an improvised fix on the box." After step 6, that state
   is a new binary that a later failure, such as step 7's shadow verification, has already shown
   broken. The rule keeps the launch-gating evidence observer down until a follow-up task lands.
4. **No rollback is named, and the obvious one does not work.** "Redeploy the previous commit"
   fails for every commit from 5b-i to 5d, because each one renders
   `DUBPLATE_IDENTITY_SOURCE=magic-link` without secrets. Only `be27434` (the box's current
   stamp) or saved copies restore a working pair, and the pair has to be restored together:
   because of `rejectNamespace`, the old binary refuses the new env's keys.

**Fold.** Rewrite steps 2 and 5 through 7 and the failure clause, in this order:

1. Run the quiet check as written, minus the ten-minute clause.
2. `systemctl stop dubplate-shadow.timer dubplate-import.timer`, then confirm neither service
   is activating.
3. On the box, save `cp -p /usr/local/bin/dubplate-server{,.prev}` and
   `cp -p /etc/dubplate/server.env{,.prev}`. The copy keeps mode 0600.
4. Apply.
5. Run one `systemctl start dubplate-shadow.service` and read its `Result`.
6. `systemctl start` both timers.

The failure clause names the rollback: restore both `.prev` files, delete the new unit file,
`daemon-reload`, and restart the timers. The leaves may stay, because nothing reads them.

Two open items for the conductor:
- Check whether a pause of a few minutes costs the shadow's 720h floor anything; one grep of
  `internal/engine/shadow` answers it. The pause is short and attended.
- Treat the rollback as part of the sitting, never as an improvised fix.

## Major

### M1. The apply changes the whole estate, and step 5's acceptance covers only a slice of it

Location: plan `:941-944`, `:249-252`, R-5d13 `:424-429`; `scripts/deploy.sh:293-303`, `:405`,
`:420-426`, `:506-509`, `:523-527`, `:529-559`.

Beyond the unit, the leaves, and the binary, today's apply does all of the following.

- It runs `rsync --delete` on `config/`. Today that pushes the `contributors.yaml` and
  `reject-email.txt` drift to the live incumbent importer and its reject mail.
- It re-renders the incumbent's `/etc/dubplate/env`. `--check` prints only "would change (diff
  not shown)".
- It runs `docker compose up -d`, `systemctl enable --now` on the four guardian timers, and
  `uv tool install --force beets==2.14.1` as `music`. That last command replaces the live
  importer's venv, and an import run can start at any quarter hour.

Step 5 accepts the apply after checking only that "nothing touches the incumbent binary or its
units." The plan's claim that "`deploy.sh` gains the unit and the two pre-creates, and nothing
else" (spec `:1766`) describes the diff to `deploy.sh`, not the diff to the box.

There are two more ordering gaps:
- R-5d13's dirty-`web/` refusal and its `npm ci` sit before `make build-linux`, which in apply
  mode runs after the config rsync and the unit rsync. A refusal therefore aborts a partial
  apply.
- Nothing requires a clean whole tree. `git describe --dirty` stamps a dirty Go tree, and lanes
  share the main tree.

**Fold.**
- Step 5 lists the expected `--check` outcome for every step. Changes: the new unit, two leaves,
  the binary, and `server.env`. No-ops: compose and the guardian units. The named copy drift is
  accepted by Geoff by name, or the conductor reverts it first.
- The incumbent-env "would change" line is a stop unless it is explained. A names-only diff is
  enough.
- The import timer stays stopped across the apply (B1).
- R-5d13's refusal and both builds move above the first remote write.
- Task 10's precondition adds "`git status --porcelain` empty, or deploy from a fresh worktree
  at Task 9's commit."

### M2. "Local caller only" does not hold for a same-host proxy, and the plan never names which address the check compares

Location: R-5d8 `:361-375`, privileged-fields row `:279`, Task 3 `:645-676`.

The route answers "a peer that is loopback or one of the configured bind addresses." That rule
fails in two ways.

- **A same-host proxy exposes the route.** The most common adopter topology is Caddy or nginx on
  the same host, proxying to `127.0.0.1`, with a trusted proxy of `127.0.0.1`. The same holds for
  cloudflared under `network_mode: host`. Every internet request then arrives from loopback, so
  the route answers the internet.
- **Using the client IP would let a client spoof in.** The plan never says the check reads
  `PeerAddress`. An implementer who writes it on `ClientIP` lets any client behind a configured
  proxy send `X-Forwarded-For: 127.0.0.1` and get an answer. Task 4V lists this exact mutation
  (`:712-714`), and Task 3 does not.

The leak is small, because the answer carries counts only. Still, it makes the pass's one new
anonymous-reachable route (the argument Ruling 2 leans on) a reconnaissance oracle for strip
regressions. It also breaks BS:510-512's "spoofed identity header from a non-trusted source,"
since the dial source is a trusted source in exactly this topology.

**Fold.** R-5d8 answers only a peer, read through `PeerAddress` and compared as an unmapped
`netip.Addr`, that is in the local set and **not** in the trusted-proxy set. Doctor dials from a
local source that is not a trusted proxy. When every local candidate is a trusted proxy, the row
reads `not probed: every local address is a trusted proxy`, which is not a failure. Task 3 adds
two rows and two mutations:
- A request from the configured trusted proxy gets the 404.
- A trusted peer whose header names `127.0.0.1` gets the 404.
- Mutation: "compare via `ClientIP`."
- Mutation: "drop the trusted-set exclusion."

This is not an SSRF. Doctor dials only configured bind addresses, and the route makes no
outbound call.

### M3. The probe can read zero for every underscore spelling, so a zero is not evidence

Location: Task 3 outcomes `:646-648`, acceptance `:654-655`, mutation 4 `:674-675`.

Go's server canonicalizes header names on read. `Remote_User` arrives as `Remote_user`, and
`Remote-User` arrives as `Remote-User`. The client-side exact map keys the plan specifies
therefore make no difference on the wire. Mutation 4's premise (that `Header.Set` changes what
arrives) holds only for the in-package direct-handler call.

Suppose the route counts by indexing the header map with the spellings as written, which is
natural given "the header lists stay unexported" and are shared with the client. Over a real
socket, every underscore spelling then counts zero whether or not the strip removed it, and the
live row reads `refused` for half of group 1 vacuously. The no-strip acceptance criterion asks
only for "every count nonzero", an aggregate that stays nonzero from the dash spellings alone.

This is the planning-miss "proof names the fixture state" item: the proof needs the wire state,
not the in-process state.

**Fold.** The route counts with the canonicalizing accessor (`r.Header.Values(name)`). Quote
`go doc net/http.Header.Values` and `textproto.CanonicalMIMEHeaderKey` in R-5d8. Add a
**wire-level** no-strip row: route mounted without the strip, behind a real listener. That row
asserts the identity count **equals** the number of distinct canonical names sent. Replace
mutation 4 with "count by map index" failing that row.

### M4. "No deploy while unattended" rests on prose, while implementers hold root ssh and apply is the default

Location: Global constraints `:249-252`, Task 5 `:756-759`, Task 6 `:800-801`;
`scripts/deploy.sh:81-95`.

`bash scripts/deploy.sh` with no flag writes production. Implementer subagents can reach the box
over BatchMode ssh; this review did it. This pass is the first to put `deploy.sh --check` in
implementers' hands in two tasks, and dropping one flag gives you M1's estate-wide apply plus
B1's broken shadow, unattended. The lane-2 translit plan forbids even `--check` for its
implementer (translit plan `:1283`). CLAUDE.md's "a rule lives where it executes" asks for the
rule to be enforced by the tool.

**Fold.** Take one of these, in order of preference:
1. Task 5 flips the default: no flag means check, and apply needs `--apply`. That is one `case`
   arm, a usage line, and the runbook's command. Record it as a spec erratum beside R-5d13
   (see m3).
2. Tasks 5 and 6 hand the `--check` run to the conductor. The implementer reports the gate only,
   on translit's precedent.

## Minor

**m1. The bind-addresses row has an unnamed state, and the runbook's operator rule triggers it.**
Location: R-5d7 `:351-359`, Task 7 `:823-825`. `net.InterfaceAddrs` needs `AF_NETLINK` (measured
above), which the unit's hardening denies. The operator rule is "a hand-run verb goes through
`systemd-run` carrying this unit's own properties" (`dubplate-shadow.service:36-40`). Followed
for `doctor`, it turns the row into an error the plan never names. Fold: add the fifth state,
`not checked: <error>`, failing. The runbook runs `doctor` without `RestrictAddressFamilies`,
the way Task 10 step 9 already does, and says why.

**m2. Ruling 2's errata list is incomplete.** Location: `:110-114`, Task 11 `:987-995`. Moving
the verify endpoint also breaks two ratified promises the plan leaves out:
- Build spec ladder entry 5 (`2026-09-02-greenfield-server-build.md:707-714`, "the forward-auth
  verify endpoint with its own trust-boundary documentation").
- BS:510-513's requirement that "`dubplate doctor` proves the boundary" for forward-auth.

Ruling 2 itself is sound. There is no consumer before rung 12, the answer header is driver
knowledge, and deferral removes an identity oracle from the first deploy. Fold: Task 11 pastes
both errata, and STATUS's "To rung 12" line names the doctor proof as well as the endpoint.
Task 8 marks every SEC or UNI finding absorbed only by R21 (the verify oracle) as **carried to
rung 12**, not absorbed.

**m3. R-5d13 widens `deploy.sh` past the spec's closed list.** The spec says "`deploy.sh` gains
the unit and the two pre-creates, and nothing else. Fork 1 keeps that list closed"
(`:1766-1767`). The clean web build is justified, but Task 11's errata list lacks it (and M4's
flag, if taken).

**m4. Task 5's `--check` acceptance cannot be met as worded.** Location: `:756-759`. "It still
aborts at the render on the names Task 6 has not yet added": `render_template` fails only on a
row that is present but unresolved (`deploy.sh:228-241`). At Task 5 the template renders, so
`--check` runs through every step. Fold: delete the clause, and have the report quote the unit,
leaf, and web-build lines.

**m5. Task 10 step 4's local doctor inherits the flat namespace.** Location: `:937-940`. Sourcing
`~/.local/secrets` puts about 40 incumbent `DUBPLATE_*` names into `os.Environ`, including
`DUBPLATE_STATE_DIR=/var/lib/musicbox` and the incumbent's `DUBPLATE_SERVE_BIND_ADDRS`. Doctor's
`rejectNamespace` then fails on them, or the wrong values win. Fold: run
`env -i PATH=... <rendered rows> <scratch overrides> dubplate-server doctor`, with the rendered
rows only.

**m6. The mint mechanics are underspecified.** Location: `:931-936`.
- The floor counts string bytes (`auth.go:71`, `:274`), so "at least the length floor" admits
  32 printable characters. Specify 32 random bytes or more, encoded, for example
  `openssl rand -base64 48`.
- Pipe the value into `secret-set.sh NAME --stdin`. `--value` puts it in argv
  (`secret-set.sh`'s own usage note).
- The store rule the plan picks is right, and the registry note should say why. The CLAUDE.md
  worker-only exception covers cairn per-site keys set by `wrangler secret put`. There is no
  Worker here, and `deploy.sh` renders only from the age store. Keep both names out of
  `sync.sh`'s WORKER_SECRETS.

**m7. The first deploy widens where the secrets live, and the plan does not record it.** Every
`dubplate-server` verb shares `server.env`, and `Load` requires the selected source's keys. The
five-minute shadow, which parses member-uploaded files and calls MusicBrainz, will hold both HMAC
keys and the SMTP password in its environment. Any other process running as `music` can read
them from `/proc/<pid>/environ`; the incumbent's beets runs as `music` and also parses uploads.
Fold: add a threat-model residual in Task 8 or Task 11, carried to rung 16 (`LoadCredential=`
or a dedicated uid for serve).

**m8. The rest of group 2 survives from a trusted peer, and the runbook omits the proxy's duty.**
Location: `strip.go:452-470`, spec `:818`. From a configured proxy, every group 2 name is kept,
not only the configured header. At 5d nothing reads them (grep-confirmed). At the cutover, a
client-forged `X-Real-Ip` or `True-Client-Ip` rides through cloudflared toward the fallthrough.
Fold:
- Carry "strip every group 2 name except the configured header from a trusted peer" to rung 12.
- Task 7 states that the proxy must overwrite or append the configured header. A configured
  header the proxy merely passes through is attacker-chosen.

**m9. R-5d4 does not name hops that carry a port.** Some proxies emit `203.0.113.9:443` or
`[2001:db8::1]:443`. Under the hop rule, each one falls back to the peer, so every client behind
that proxy shares one limiter bucket, and one client can exhaust it for all. Fold: name the
state. Either accept `netip.ParseAddrPort` and keep the address, or refuse and document it in
the runbook.

**m10. The sitting record would put member identities in the ledger.** Location: `:924`. "Each
step's output is recorded in the close" copies `member list` and doctor output, meaning Geoff's
address and Navidrome login, into `HISTORY.md`. Fold: record statuses and exit codes, never
member fields.

**m11. The supervised start has no failure cleanup.** Location: `:956-959`. With
`Restart=on-failure`, `RestartSec=5`, and `StartLimitBurst=30`, a failing start loops. Fold: the
failure clause runs `systemctl stop` and then `reset-failed` on the unit.

**m12. Task 6's template parser is untested for parity with `render_template`.** Location:
`:782-787`. A Go re-implementation can be more lenient than the bash parser, and the proof would
still pass. Fold: table rows mirror `deploy.sh:168-189`'s documented states (a literal
containing `=`, a bare row, a comment, a line with no `=`), each asserted.

A related note for the sitting record, not a finding: the apply also ships every matcher change
since `be27434` into the live shadow (`reevaluate.go` retires the old commit). Name it as a
shadow-evidence event.

## The four planning-miss items, per task

| Task | Behavior quoted | Every state reported | Every channel named | Fixture state named |
| --- | --- | --- | --- | --- |
| 1 | Yes (`ParseAddr`, `Unmap`, `String`, `CanonicalMIMEHeaderKey`) | Mostly. Name "flag false with keys set" (`config.go:257`, "set without") and a non-boolean flag, both `Load`'s | Yes | Yes (limiter test) |
| 2 | The loopback claim (`:356`) and `InterfaceAddrs`'s netlink use are recalled, not quoted | Missing the enumeration-error state (m1) | Yes | Yes |
| 3 | The `Header.Set` behavior is recalled, and the server-side canonicalization is missing (M3) | Missing "dial source is a trusted proxy" (M2) | Missing `PeerAddress` and the trusted-set exclusion for the probe answer (M2) | Missing the wire state for the no-strip proof (M3) |
| 5 | Yes (man pages) | Yes, but the `--check` acceptance is incoherent (m4) | Yes | Yes |
| 6 | Parser parity assumed (m12) | Yes | Yes | Yes |
| 7, 8, 9, 11 | Docs. Task 8's "absorbed" needs a test or "by design" pointer, which the plan has | n/a | n/a | n/a |
| 10 | The shadow timer cadence was not read (B1) | Missing the post-apply failure states and the rollback (B1, m11) | Yes (age store, render, 0600) | The scratch roster must carry an admin, which the plan has |

## Verified sound (for the do-not-re-litigate list)

- **Ruling 2's substance.** No consumer exists before rung 12. The answer header and the bypass
  paths are driver knowledge (spec "Universality"). Deferral removes an identity oracle from the
  first deploy.
- **Ruling 3's diagnosis.** The box env lacks the selector, and `Load` requires the source keys
  and rejects unknown names. The age store is the correct store for the two HMAC keys.
- **The SMTP bare rows.** All five names exist in the age store under the incumbent's exact
  names (`env/dubplate.env.template:41-45`), so the flat namespace resolves both files to one
  value.
- **`DUBPLATE_PUBLIC_URL` as a literal.** The hostname is already public in
  `config/cloudflared-tunnel-config.json`, and the value is not a secret. No secret reaches the
  repo under the plan as written.
- **`--check` is read-only on the box.** The plan's no-deploy constraint holds as long as the
  flag is present (M4).
- **The unit's privilege surface.** It adds no capability and grants exactly the two leaves.
  `auth.Open` and the roster loader enforce the modes. It is installed disabled and absent from
  `TIMER_UNITS`. Port 8080 is free.
- **The 50-row count.** It matches both review texts, so the completeness risk is in "absorbed"
  claims, not the count. Task 8's test-or-pointer rule covers it, plus m2's carried marking.
