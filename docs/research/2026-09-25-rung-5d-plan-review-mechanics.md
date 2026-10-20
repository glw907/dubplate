# Rung 5d plan review: mechanics and feasibility

Lens: mechanics and feasibility, one of three plan lenses. Target: `docs/superpowers/plans/2026-09-25-rung-5d-boundary-deployment-close.md` at `af7d941`. Reviewed 2026-09-25 against HEAD `af7d941`, which descends from the plan's `4c766c3` basis. Every claim below comes from a probe, a man page, or source read on the day. Probes ran in `/tmp/claude-1000/rung5d-probe/` and `/dev/shm`. The box was read once through `ssh -o BatchMode=yes`, read-only.

**Counts: 0 blockers, 6 majors, 12 minors.** No finding is an owner fork. Each has one correct answer, which the finding proposes.

Findings are ranked by consequence, not polish. Pointer drift that Task 0 already catches is listed at the end as notes, not findings.

---

## Majors

### M1. Task 5's namespace evidence cannot tell a missing leaf from a badly placed scratch leaf

- **Location:** plan:738-746 (Task 5 evidence), plan:1015-1016 (exit gate 3).
- **Defect:** the evidence runs a user-level transient unit that carries "the new unit's hardening properties" with `ReadWritePaths=` pointed at "two scratch leaves". The inherited block includes `PrivateTmp=yes` and `ProtectHome=yes` (`dubplate-shadow.service:60-61`). Both break the proof, measured on this workstation (systemd 259):
  - **Scratch under `/tmp` or `/var/tmp`** (the implementer's usual scratch location): a leaf that *exists* fails with the same words the plan expects from a missing one. The journal read `Failed to set up mount namespacing: /tmp/claude-1000/rung5d-probe/present: No such file or directory` and `status=226/NAMESPACE`. The positive half fails, and the negative half proves nothing.
  - **Scratch under the home directory** (`/home` links to `var/home` on this host, and `ProtectHome=` covers it): the positive `touch` exits `status=1/FAILURE`, and the absent-leaf case **succeeds** (`status=0/SUCCESS`), so the named error never appears. A binary built in the repo also sits under the home directory and cannot be executed there.
  - **Scratch under `/dev/shm`**: the positive half succeeded (`status=0/SUCCESS`, file written), and the absent half failed with `Failed to set up mount namespacing: /dev/shm/rung5d-probe-109971/absent: No such file or directory`, `226/NAMESPACE`.
- **Proposed fold:** Task 5 names the location. Put the scratch leaves and a copy of the built binary under `/dev/shm/<unique>`, and drop `User=`, which a user manager cannot switch. Run the positive half first and record its success in that same location, since a namespace failure means something only after the same placement has succeeded. Then remove one leaf and record the 226 text. Add one line saying why `/tmp`, `/var/tmp`, and the home directory are excluded (quote `man systemd.exec`: `PrivateTmp=` gives "/tmp/ and /var/tmp/ ... not shared"; `ProtectHome=` makes "/home/, /root, and /run/user ... inaccessible").

### M2. The probe route's method refusal is a 405 with an `Allow` header, not the terminator's 404

- **Location:** plan:365-367 (R-5d8, "every other method, meets the terminator's 404 through the same `exactPath`/`methodOnly` wrapping"), plan:665-666 (Task 3 criterion: "`POST` gets the same 404"), plan:673 (mutation 3).
- **Defect:** `methodOnly` answers `405` and sets `Allow: <method>` (`cmd/dubplate-server/serve.go:380-389`). So the criterion cannot be met with the mechanism the plan names. If `methodOnly` sits outside the local-peer check, any internet client that sends `POST <prefix>/boundary-probe` gets `405 Allow: GET` instead of the 404 an unknown path gets. That tells a stranger the route exists, which is the property the peer gate exists to hide.
- **Proposed fold:** the probe handler checks the peer first and the method second, and both refusals answer `serve.Terminator()` (`internal/serve/prefix.go:130-134`). `exactPath` stays, and `methodOnly` is not used on this route. Criterion: a non-local `GET`, a non-local `POST`, and a local `POST` each get a 404 byte-equal to an unknown path's.

### M3. "Local" is false for the most common self-hosted topology, a proxy on the same host

- **Location:** plan:279 (the privileged-fields row: "An internet client through the tunnel arrives from the proxy's own address, which is neither"), plan:365-366 and plan:374-375 (R-5d8's reason).
- **Defect:** the route answers any peer that is loopback or a configured bind address. A reverse proxy on the same host (Caddy or nginx `proxy_pass http://127.0.0.1:8080`), or cloudflared run as a host service instead of a container, dials loopback. Every internet request then arrives from `127.0.0.1` and gets the probe's answer. On this estate the claim holds only because cloudflared runs in a container on `musicbox_default` and the cutover binds the bridge gateway. The disclosure is low in value: counts of the caller's own headers that survived the strip, and, through a nonzero group 2 count, whether its peer is a configured trusted proxy. It is still a strip-bypass oracle, and it contradicts a trust-boundary claim that both lenses and the runbook will grade against.
- **Proposed fold:** correct the row and R-5d8's reason to say what holds. The route answers any peer the host routes as local, a same-host proxy's forwarded traffic included. It discloses only header counts, never a value or an address. Record that residual in Task 7's trust-boundary section. Do not add a secret or a second gate, since the counts do not justify the machinery. This has one correct answer: the plan states the property that holds.

### M4. The state-dir check proves the leaf array, never the pre-create that exit gate 2 names

- **Location:** plan:404-417 (R-5d11), plan:751-754 (Task 5 criteria), plan:1014 (exit gate 2: "failing when either leaf is removed from `deploy.sh`").
- **Defect:** R-5d11 has the check read `DUBPLATE_SERVER_LEAVES` and match each unit grant against it. The pre-create is written as a separate loop "for `auth` and `roster`". The shadow leaf keeps its own `mkdir -p` line (`deploy.sh:502-505`), so the loop cannot simply iterate the array. An implementer who writes `for leaf in auth roster` gets a check that stays green when `auth` is deleted from the line that actually runs `install -d`, and every Task 5 seed still passes. Exit gate 2 is then proved against the declaration only. Folding shadow into a `-m 0700` loop is not free either: the box's shadow leaf is `drwxr-xr-x music music` (read 2026-09-25), and `install -d -m 0700` on an existing directory re-modes it.
- **Proposed fold:** the pre-create loop iterates `"${DUBPLATE_SERVER_LEAVES[@]}"` and skips `shadow` by name, with a comment saying its line stays and why. check.sh asserts that the loop reads the array, for example by grepping the `for ... in "${DUBPLATE_SERVER_LEAVES[@]}"` line. Add a seventh seed: a literal leaf list in the loop, or `auth` removed from what the loop creates, fails by name.

### M5. The rendered-doctor test cannot fail as specified, and it dials port 8080 on whatever machine runs the gate

- **Location:** plan:782-795 (Task 6 outcomes and criteria), plan:937-940 (Task 10 step 4).
- **Defect, two parts:**
  1. The template's `DUBPLATE_STATE_DIR` is a literal (`env/dubplate-server.env.template`, `DUBPLATE_STATE_DIR=/var/lib/dubplate`), and R-5d14 adds the roster path as a literal. The test "adds a fixture state directory, a fixture roster". Both `lookupEnviron` (`cmd/dubplate-server/doctor.go:400-411`) and `config.Load`'s `newWall` (`internal/config/config.go:171-181`) are last-wins, so an appended fixture row overrides the template row. The criterion "deleting the `DUBPLATE_ROSTER_FILE` row fails the test naming the magic-link source's missing key" can then never fail, because the appended fixture still supplies the key.
  2. The template carries no port row (R-5d14), so doctor's probe row dials `127.0.0.1:8080`, the registered default (`cmd/dubplate-server/serve.go:97-100`), on every gate run. Anything listening on 8080 on the workstation (a dev server, a container) turns the row into `unexpected answer` and fails the gate, a failure no code change caused. Task 10 step 4 inherits the same dial.
- **Proposed fold:** the parser replaces the value of a template row that exists and never appends one. The criterion states it: "deleting the row leaves no substitute". The fixture environment also sets `DUBPLATE_SERVE_PORT` to a port the test reserved and closed, and the test asserts `boundary probe` reads `no listener`. Name that port as the test's fixture state. Task 10 step 4 does the same, with the scratch state and roster paths substituted and never appended.

### M6. `net.InterfaceAddrs` needs `AF_NETLINK`, which the serve unit's hardening denies

- **Location:** plan:351-359 (R-5d7), plan:823-825 (Task 7's install order: doctor "through `systemd-run` carrying the unit's properties").
- **Defect:** measured with a scratch binary under `systemd-run --user -p RestrictAddressFamilies="AF_INET AF_INET6 AF_UNIX"`, the set R-5d10 inherits: `InterfaceAddrs: 0 route ip+net: netlinkrib: address family not supported by protocol`. The same binary unsandboxed returned 8 addresses. R-5d7 names no state for an `InterfaceAddrs` error. Doctor run the way the runbook specifies would either fail the row with an unnamed error or, if the error is swallowed as an empty list, report every non-loopback bind address as `not on any interface`. That false failure would appear the moment the cutover binds the bridge gateway. Task 10 step 9 passes only `ProtectSystem` and `ReadWritePaths`, so the sitting itself is unaffected.
- **Proposed fold:** R-5d7 adds a fifth state: `InterfaceAddrs` fails, and the row reads `not checked: <error>` and fails, naming `RestrictAddressFamilies` as the likely cause. Task 7's install order runs doctor with the unit's filesystem properties but without `RestrictAddressFamilies`, and says why in one sentence. Do not widen the unit's family set: serve itself never needs netlink.

---

## Minors

**m1. Task 3's mutation 4 survives over the wire.** Location: plan:646-648, plan:674-675. Go's server canonicalizes every incoming field name. Measured: a real client sent `req.Header["Remote_User"]`, and the server saw `Remote-User` and `Remote_user` as two keys. `Header.Set("Remote_User")` stores `Remote_user` (`CanonicalMIMEHeaderKey("Remote_User")` returns `Remote_user`), which is distinct from `Remote-User`, so both spellings still arrive and the count does not change. In process, exact non-canonical keys build a request no server produces, and the strip's `Del` and the route's `Values` lookups, which canonicalize, would miss them. Fold: drop mutation 4 and the "exact map keys" rationale, and send with `Header.Add` over the wire. Replace it with a mutation that drops the underscore half of the list, which the no-strip row catches by its count.

**m2. R-5d9 leaves two dial outcomes unnamed.** Location: plan:377-384. A dial with `LocalAddr` set to an address absent from the host returns `bind: cannot assign requested address` (measured), which is neither refused nor a timeout. A listener that accepts and never answers is folded into `no listener` by "times out", which hides a wedged process. Fold: `no listener` applies only to `ECONNREFUSED` or a dial timeout. The absent-address bind error reads `not probed (address not on any interface)` and adds no failure, since R-5d7 already fails. A timeout after connect reads `unexpected answer: timeout` and fails. State that the 2-second bound covers the dial and the response separately.

**m3. The `client ip` row when `Load` fails.** Location: plan:600-602, plan:340-347. The resolver reads a `*config.Config`, and doctor has none when `Load` fails (`doctor.go:35-36`). R-5d6 names `not checked` for the web row only. Fold: `client ip` reads `not checked` and adds no second reason when `Load` failed. Also name the state R-5d2 omits: flag false with `ADDRS` set is `Load`'s "set without ... enable flag" report (`internal/config/config.go:260-262`).

**m4. `TimeoutStopSec=15` truncates serve's own drain.** Location: plan:396-397. `shutdownGrace` is 25 seconds (`internal/serve/server.go:56`), and its comment says "A supervising unit's own stop timeout (5d's concern, not this package's) is the real backstop". The plan's reason ("serve's shutdown drains in-flight requests") argues for a stop timeout at least as long as the drain. Fold: `TimeoutStopSec=30`, or keep 15 and say that it deliberately cuts the drain short.

**m5. The render claim in Task 5 is false.** Location: plan:756-759. At Task 5 the template is unchanged, and every row resolves today (the box carries ten rows, read 2026-09-25). So `--check` does not abort at the render, and it runs every step. Fold: delete the sentence, and have the criterion read the full `--check` output.

**m6. The clean-web refusal runs after the box has been written.** Location: plan:424-429. "Before `make build-linux`" (`deploy.sh:338`) comes after the apply-mode rsync of compose and config, the unit rsync, and the tmpfiles install (`deploy.sh:228-300`). A dirty `web/` or a failed `npm ci` therefore leaves a partial deploy. Fold: the clean check, `npm ci --prefix web`, the build, and `make build-linux` run right after the render, before the first box write. Name the build command (`npm run build --prefix web`; `make web` runs the gates twice).

**m7. Task 1's fixture names a stub mailer that has no seam.** Location: plan:565-569. `newServeShell` builds its mailer from configuration (`serve.go:278`). No stub is needed: `allowLinkClient` runs before the form parse or any send (`internal/auth/handlers.go:194`), so a POST with an unparseable address consumes the client's bucket. Fold: name that fixture and the `Origin` and `Sec-Fetch-Site` values the origin check needs for the POST. The front-exemption criterion (plan:570-571) needs a spy route registered on the returned shell. That is legal before `Handler()` (`internal/serve/serve.go:88-90`), so say so.

**m8. A forwarded hop that carries a port collapses every client onto the proxy's bucket.** Location: plan:324-330. `netip.ParseAddr("203.0.113.9:1234")` and `ParseAddr("[2001:db8::1]")` both fail (measured). Under R-5d4 such a chain yields the peer, so one limiter key covers everyone behind a proxy that appends ports, and one client can lock the group out of sign-in. The estate's `Cf-Connecting-Ip` is unaffected. Fold: accept `netip.ParseAddrPort` as a second hop form and take `.Addr()`, or record the residual in the runbook. Recommend the first: it costs two lines.

**m9. The failure-ping slug is shared with the incumbent.** Location: plan:395. `OnFailure=dubplate-ping-fail@serve.service` resolves to `DUBPLATE_PING_SERVE_URL`, the same slug `dubplate-serve.service:35` uses. It is inert today. A rung 9 mint would page the incumbent uploader's check for a greenfield failure. Fold: use a distinct slug, such as `server-serve`.

**m10. The web row must not log onto doctor's stdout.** Location: plan:340-344. `newWebRoutes` writes one record through its logger when a branding neutral drops the theme-color meta (`cmd/dubplate-server/web.go:214-216`). Fold: doctor passes a discard or stderr logger, and a test holds the line count with a neutral branding value.

**m11. R-5d12's mutation needs a named way to hide the tool.** Location: plan:763-764. `systemd-analyze` lives in `/usr/bin` beside bash, sed, and git, and `/bin` links to it, so `PATH` cannot simply drop it. Fold: build a shim directory of symlinks to `/usr/bin/*` minus `systemd-analyze`, drop `/usr/bin` and `/bin`, and keep the other `PATH` entries. The expected result is that the gate fails naming the tool, and that it passes with the guard restored.

**m12. The local-peer comparison must use `netip`.** Location: plan:650-651. `SplitBindAddrs` returns strings verbatim, and `net.ParseIP` accepts `::ffff:127.0.0.1` (`internal/serve/bindaddrs.go:46-53`). `RemoteAddr` prints IPv4-mapped peers in dotted form. Fold: compare the unmapped `netip.Addr` of the peer against the bind list and against `IsLoopback`.

---

## The four planning-miss items, per task

| Task | Behavior quoted | States named | Privileged channels | Fixture state |
| --- | --- | --- | --- | --- |
| 1 | Yes (`netip` quotes verified) | Missing: flag false with ADDRS set (m3), port-bearing hop (m8) | Complete | Stub mailer has no seam (m7) |
| 2 | Yes | Missing: `InterfaceAddrs` error (M6), `client ip` with Load failed (m3) | n/a | Yes |
| 3 | `methodOnly` misread (M2); `Header.Set` consequence wrong (m1) | Missing: bind error, hung listener (m2) | "Local" claim false (M3) | Yes |
| 5 | `man` quotes verified | Pre-create proof gap (M4) | Complete | Scratch location unnamed (M1) |
| 6 | `render_template` shape correct (`deploy.sh:183-212`) | Yes | n/a | Override mechanism and port unnamed (M5) |
| 7, 8, 9, 11 | Documents; not graded by this lens beyond M3 and M6's runbook lines | | | |
| 10 | Steps check out (below); step 4 inherits M5 | | | |

---

## Verified sound, to avoid re-litigating

- `man systemd.service` for `Type=exec`: "systemctl start command lines for exec services will report failure when the service's binary cannot be invoked successfully". Quoted correctly.
- `man systemd.exec`: `ReadWritePaths=` "may be prefixed with "-", in which case they will be ignored when they do not exist". Measured: a `-`-prefixed absent path starts cleanly (`status=0/SUCCESS`), which confirms R-5d10's EROFS-instead-of-226 reasoning. The exit table lists "226 EXIT_NAMESPACE Failed to set up mount, UTS, or IPC namespacing".
- `systemd-analyze verify --recursive-errors=no` over the whole `systemd/` set plus a scratch unit built to R-5d10 (the `[Unit]` start limits, `Type=exec`, the shadow block, `[Install]`): clean modulo the existing `check.sh:634` filter.
- The `net/netip` behaviors: `::ffff:198.51.100.7` unmaps to `198.51.100.7`; a CIDR, a hostname, and a leading space all fail to parse; `fe80::1%eth0` parses with zone `eth0`, so R-5d2 must refuse it explicitly, which it does. `CanonicalMIMEHeaderKey("X Bad")` returns the key unmodified, as R-5d3 quotes.
- A Go client sends a `Connection`-listed sentinel and an underscore key unchanged, and a Go server receives both without removing any hop-by-hop header. A `net.Dialer` with `LocalAddr` set to `127.0.0.1` yields `RemoteAddr` `127.0.0.1:<port>`.
- `Load`: an unset or empty flag is off, and an empty Required value is missing (`config.go:192-199`, `:223-233`, `:248-249`). The plan's reading holds.
- `auth.Open` refuses a leaf wider than 0700 (`internal/auth/store.go:53-71`). The roster loader refuses `perm&0o022` (`internal/auth/roster.go:318`). `validateMagicLinkSource` never opens the roster (`cmd/dubplate-server/auth.go:265-325`), so a shadow firing between Task 10's apply and its bootstrap still loads. An absent store is not a doctor failure (`cmd/dubplate-server/store.go:111-114`).
- `member add <address> --admin --name --streamer-username` matches the verb (`cmd/dubplate-server/member.go:68-102`). It opens only the roster (`:120-124`), so step 8's grant of the roster leaf alone is sufficient.
- The strip runs before normalization, the origin check, and routing (`internal/serve/serve.go:164-177`), so the probe route sees exactly what the strip left.
- `web/build/placeholder.txt` survives a build, so `git status --porcelain -- web/` stays clean after a gate run (checked on the working tree).
- Box, read-only, 2026-09-25: systemd 257; `music` is uid 2000; `/var/lib/dubplate` holds only `shadow` (`var_lib_t`, SELinux enforcing); nothing listens on 8080 (only sshd on 22 and docker-proxy on `127.0.0.1:4533`); `/etc/dubplate/server.env` has the ten rows the plan names, with no `DUBPLATE_IDENTITY_SOURCE`; `dubplate-shadow.timer` is active; the incumbent `dubplate-serve.service` is inactive.
- Task order: 1, 2, 3, 5, 6 build as stated. Task 2 creates `doctorprobe.go` with the placeholder row that Task 3 replaces. Task 5 needs only Task 1. Nothing in a later task is consumed earlier.

## Pointer notes (Task 0 catches these; not findings)

- `UNIT_FILES` opens at `deploy.sh:59`, not `:64`. `DUBPLATE_LIB_DIR` is at `:502`, not `:503`.
- `group3KeepSet` is `internal/serve/strip.go:199-203`, not `:201-205`.
- Group 1's names are `identityExtras`, `strip.go:112-135`. The plan's `:11-63` cites the comment block that documents them.
