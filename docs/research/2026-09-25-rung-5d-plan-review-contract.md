# Rung 5d plan review: contract and criteria lens

Reviewed: `docs/superpowers/plans/2026-09-25-rung-5d-boundary-deployment-close.md` at `af7d941`
(revision 1), against the rung 5 spec (`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md`),
`docs/STATUS.md`'s carried "No deploy until 5d" item, and the tree at `af7d941`. One lens: are the
outcomes testable, does each criterion name its fixture and its failure, can any criterion or named
mutation pass vacuously, is the spec's 5d row covered, and do the four planning-miss items hold on
every task. Findings are ranked by consequence, not polish.

**Counts: 0 blockers, 7 majors, 16 minors.** No finding is an owner fork. Rulings 2 and 3 are correctly
posed as owner questions, and the four self-taken decisions the brief names hold against the spec
(see "Self-taken decisions").

## Coverage of the spec's 5d row and STATUS's carry

Every clause of the pass-split row (`spec:2519`) maps to a task or a recorded deviation:

| 5d row clause | Plan disposition | Verdict |
| --- | --- | --- |
| Forward-auth verify endpoint, class, docs | Ruling 2 (move to rung 12), Task 4V under no, erratum at Task 11 | Recorded as a ruling |
| Trusted-proxy integration class | Task 1 | Covered |
| `doctor`'s probe rows and bound-address row | Tasks 2, 3 | Covered |
| Serve unit, `ReadWritePaths` over both leaves, two pre-creates | Task 5 | Covered (see M3) |
| Example roster file (R37) | Already landed at 5b-ii (`spec:2386-2390` erratum) | Recorded |
| Generalized state-dir check | Task 5, R-5d11 | Covered, but the check does not prove what the exit gate says (M3) |
| Env template's pending rows and "eight literal rows" | Task 6, R-5d14 (six rows per the 5b-i count erratum; SMTP rows bare, secrets bare under Ruling 3's yes) | Recorded as a deviation with reasons |
| SEC and UNI disposition table | Task 8 | Covered |
| Full adversarial cycle | Cycle plus Task 9 | Covered |
| Errata and close | Task 11 | Covered, three errata missing (m10) |
| "Worktree, released at the close" | Superseded by the lane registry, recorded at the close | Recorded |

Exit gate: each of the seven clauses maps in "The exit gate" (`plan:1008-1026`). Clause 3 ("the unit
starting ... failing with a named mount-namespacing error") rests on a user-manager approximation plus
the box start; the plan records that as an erratum, which is honest. Clause 5 is unmet under Ruling 3's
no and the plan does not say so (m6).

STATUS's "No deploy until 5d" carry: the nine keys (roster, two secrets, public URL, five SMTP), `doctor`
exit 0, `ReadWritePaths`, the trusted proxy, the theme keys' doctor row (R-5d6), Ruling 6's
`Referrer-Policy` residual and the branding line (Task 7), the clean `web/` build (R-5d13), and 6a/6b-i's
runbook rows (Task 7) are each assigned. Task 11's acceptance closes or routes each by name. Covered.

## Self-taken decisions

- **Bare SMTP template rows (R-5d14).** Holds. The spec's "none is ever minted" (`spec:1664-1667`) is
  stale against the age store, and the spec's own rule "Rung 5 adds no name it has not minted"
  (`spec:1657`) is satisfied by names already minted. The flat-namespace hazard
  (`env/dubplate-server.env.template:38-53`) resolves in the intended direction here: a bare row reads the
  incumbent's exported value, which is the value wanted. No ruling needed; the Task 11 erratum is the
  right vehicle.
- **`DUBPLATE_PUBLIC_URL=https://music.907.life` literal.** Holds. Not a secret, no path per R-5c9, and
  the template already carries estate literals (`/srv/music/...`). The spec's universality line
  (`spec:2855-2857`) is about Go code and config, and the template is the render manifest.
- **Trusted proxy disabled.** Holds. BS's "chosen, never fallen into" and R22's off-by-default both point
  here, and the estate values correctly wait for the cloudflared address pin. The consequence at cutover
  (every tunnel client shares cloudflared's peer as its per-client key) is carried to rung 16 by name in
  Task 11.
- **The boundary-probe route (R-5d8).** Holds as the only design that observes post-strip state from a
  live dial, and R6 leaves the observation unstated. Two corrections are owed rather than a ruling: the
  spec's closed owned-prefix route list (`spec:916-918`) needs an erratum (m10), and the plan's claim that
  only local peers can reach it is false for a common adopter topology (m1).

## Majors

**M1. The quiet-box check can never pass.** `plan:928-930`, Task 10 step 2. The step requires "the shadow
timer's next firing more than ten minutes away", but `systemd/dubplate-shadow.timer` is
`OnCalendar=*:0/5` with `RandomizedDelaySec=60`: the next firing is never more than about six minutes
away. Taken literally the sitting stops at step 2 every time; taken loosely the conductor improvises on
production, which the plan's own failure rule forbids (`plan:963-965`). **Fold:** replace the condition
with an explicit window: no `dubplate-*` unit activating, then `systemctl stop dubplate-shadow.timer`
(Geoff approves, before step 6), the apply, the step 7 checks, `systemctl start dubplate-shadow.timer`,
and the journal read of its next run. Name the incumbent `dubplate-import.timer` (`*:0/15`) in the same
check, since the apply's compose and env steps touch the stack it drives.

**M2. The apply runs the whole estate deploy, and step 5 reads only the incumbent binary and units.**
`plan:941-944`. `scripts/deploy.sh` in apply mode also rsyncs every incumbent unit and reloads, re-renders
`/etc/dubplate/env`, delivers `contributors.yaml`, runs `docker compose ... up -d` on the live stack,
runs `systemctl enable --now` on the four incumbent timers, and runs `uv tool install --force
beets==$BEETS_VERSION` (`deploy.sh:32`, `:509-540`). Under `--check` the beets step prints "would run"
unconditionally, so `--check` cannot show whether the box's beets changes, and STATUS records beets
2.14.1 as adopted without a record that the box has it. The plan's step 5 criterion ("nothing touches the
incumbent binary or its units") passes while the apply upgrades the incumbent importer's beets or
recreates Navidrome. **Fold:** Task 0 item 10 adds read-only box reads of the installed beets version, the
deployed `compose.yaml` against master's, and the enabled state of every timer in `TIMER_UNITS`; Task 10
step 5 enumerates the expected diff set (the binary, `server.env`, the new unit, the two pre-creates, the
web build) and makes any other diff in the `--check` report, or any Task 0 drift, a stop. A scoped
`deploy.sh` mode would be new scope; the enumeration is the proportionate answer.

**M3. The state-dir check proves array membership, not pre-creation.** `plan:404-417`, `plan:751-754`,
exit gate clause 2. The spec's clause is "failing when either leaf is removed from `deploy.sh`" and R12
says the check "asserts `deploy.sh` pre-creates each leaf" (`spec:459-460`). R-5d11 has `check.sh` read
`DUBPLATE_SERVER_LEAVES`, while the pre-create loop runs over `auth` and `roster` and the shadow leaf keeps
its own line (`deploy.sh:503-505`). Whatever the loop iterates, deleting the loop body, or the shadow
line, leaves the array intact and the check green. None of the six named seeds removes a pre-create;
each edits the array or a unit. **Fold:** one loop over the whole array pre-creates every leaf (retire
the shadow line into it, each leaf's mode stated), and the check asserts both the array's contents and
that the loop iterating `"${DUBPLATE_SERVER_LEAVES[@]}"` is present, with a seventh seed deleting the
loop that fails by name. Quote `install(1)`'s behavior on an existing directory and on a missing parent
(m13) so the loop's order against `/var/lib/dubplate`'s creation is argued, not recalled.

**M4. Every doctor exit-status criterion lacks a named green baseline, so each can pass vacuously.**
`plan:611-616` (Task 2), `plan:661-664` (Task 3), Task 3 mutation 2 (`plan:672`). "`client ip` invalid
... with exit 1", "a missing operator CSS file with exit 1", "the row fails and doctor exits 1", and "no
listener ... doctor's exit status is unchanged" all hold trivially when the fixture environment fails
some other row: an unset secret, a missing roster, an absent admin. Mutation 2 ("report `no listener` as
a failure → exit status changes") only fails if the baseline exits 0. The green-environment fixture
first appears at Task 6, after both tasks. **Fold:** Task 2 builds the green-doctor fixture helper
(every Required key of the selected source with obviously fake values, a state directory, a roster with
one admin at 0600, a reserved and closed port per M5), one test asserts that fixture exits 0 with every
row's status stated, and every exit-1 row is that fixture plus exactly one perturbation. Task 6 reuses the
helper for the rendered-template proof.

**M5. After Task 3, every doctor test dials the host's real `127.0.0.1:8080`.** `plan:649`, `plan:786`.
`doctor` dials each configured bind address on the configured port, and the defaults are `127.0.0.1` and
`8080` (`serve.go:88-100`). `TestDoctorCmd`, `TestDoctorCmd_LineCountFollowsRegistry`, Task 2's row tests,
and Task 6's rendered-doctor test all run with those defaults unless they override them. Any listener on
8080 on the gate host (a dev server, another session's tool) turns `no listener` into `unexpected answer:
404`, a failing row, and a red gate that is no fault of the tree. It also runs from two lanes at once.
**Fold:** the M4 helper sets `DUBPLATE_SERVE_PORT` to a port the test reserved and closed, and asserts
the row reads `no listener` as a precondition. Task 6's rendered environment overrides the port the same
way, stated in the task's fixture list.

**M6. The probe's header spellings rest on a recalled `textproto` behavior, and mutation 4 is vacuous.**
`plan:646-648`, `plan:674-675`. The plan sends "the exact map keys, since `http.Header.Set` would
canonicalize an underscore key". Over the wire this does not matter: the server's
`textproto.Reader.ReadMIMEHeader` canonicalizes every received key, and `CanonicalMIMEHeaderKey` treats
`_` as an ordinary token byte, so `Remote_User` sent verbatim and `Remote_user` sent by `Set` both arrive
as `Remote_user`, a key distinct from `Remote-User`. In-process it is harmful: a request whose map holds
the non-canonical key `Remote_User` is invisible to `Header.Values("Remote_User")`, which canonicalizes
the lookup, so a handler-direct no-strip test would undercount under correct code. Mutation 4 therefore
either passes (the wire path) or measures the test's own construction (the in-process path), never the
product. **Fold:** quote `go doc net/textproto.CanonicalMIMEHeaderKey` and `Reader.ReadMIMEHeader`; build
the probe with `Header.Add` per spelling; run the no-strip proof through `httptest.NewServer` so it crosses
the same parser the live dial does, with both proofs using the one request builder doctor uses. Replace
mutation 4 with a strip mutation the probe must catch: make `stripIdentity` delete only the dash forms
while `identitySpellings` still returns both, and the end-to-end row must read `passed N identity
header(s)`.

**M7. The bind-address row's `net.InterfaceAddrs` fails under the unit's own address-family restriction,
and that state is unnamed.** `plan:351-359`, `plan:823-825`. On Linux `net.InterfaceAddrs` reads the
address table over netlink (`syscall.NetlinkRIB(syscall.RTM_GETADDR, ...)` in
`net/interface_linux.go`), and the greenfield unit inherits `RestrictAddressFamilies=AF_INET AF_INET6
AF_UNIX` with `AF_NETLINK` deliberately dropped (`dubplate-shadow.service:84-89`, R-5d10). The runbook's
install order runs verbs "through `systemd-run` carrying the unit's properties" (Task 7, on the shadow
unit's operator rule). A doctor run that way gets an error from `InterfaceAddrs`, and R-5d7 names four
states, none of which is "interface table unreadable". Task 10 step 9 happens to pass because it carries
only three properties. **Fold:** quote the netlink dependency in R-5d7; name the fifth state (`not
checked: <error>`, failing, since the operator asked for the check); and state in Task 7 exactly which
properties a doctor run carries, excluding `RestrictAddressFamilies`, with the reason.

## Minors

**m1. The probe route's local-peer gate is not local under a same-host proxy.** `plan:279`, `plan:365-375`.
With nginx or Caddy on the same host proxying to `127.0.0.1`, every internet client arrives from loopback
and the route answers it. The answer is counts only, so the exposure is small, but the privileged-fields
table claims otherwise. **Fold:** correct the claim, and record the residual in Task 7's trust-boundary
section and Task 8's residue list. An HMAC under the session secret would close it; that costs more than
a count leak is worth.

**m2. A probe dial from an absent bind address has no named state.** `plan:377-384`. With R-5d7 reporting
an address `not on any interface`, the dialer's `LocalAddr` bind fails with `EADDRNOTAVAIL`, which is
neither `no listener` nor `not probed` (the latter is only for an invalid configuration). **Fold:** such
an address reads `not probed (not on any interface)` and adds no failure beyond R-5d7's.

**m3. Task 5's `--check` criterion has the render order backwards.** `plan:756-759`. At Task 5 the template
carries no new bare row, so the render at `deploy.sh:240-247` passes and every step reports. The abort
arrives at Task 6 under Ruling 3's yes, and because it exits before any step, nothing after Task 6 can
exercise `--check`'s step reports until the Task 10 mint. **Fold:** drop "still aborts at the render" from
Task 5, and note in Task 9 that a fold touching `deploy.sh` cannot run `--check` until Task 10.

**m4. Task 1 mutation 2 names a row that does not exist.** `plan:577`. No `ClientIP` row puts the peer and
the configured set in different textual forms: the resolver unmaps the set, and `RemoteAddr` prints
dotted decimal for a mapped IPv4 peer (`go doc net.IP.String`). A string comparison is equivalent in every
reachable state, so the mutation passes. **Fold:** restate it as "skip `Unmap` in the resolver → the
`::ffff:198.51.100.7` resolver row's second member reads the mapped form and the limiter row keyed from
peer `198.51.100.7` falls back to the peer".

**m5. R-5d12's "mutation" is a proof, and nothing gates the guard's return.** `plan:763-764`. Restoring
the `command -v` guard with the tool hidden makes the leg skip and pass, the opposite of the stated
outcome. **Fold:** call it the proof it is (tool hidden, gate fails naming it), and accept that no leg
detects a re-added guard, or add `systemd-analyze` to the hard-required list at `check.sh:138-157` so the
guard has no job.

**m6. Exit gate clause 5 under Ruling 3's no is unmet and unrecorded.** `plan:136-139`, `plan:1019-1020`.
Under no, Task 6's proof asserts exit 1 naming the two secrets, so "`doctor` exiting 0 against the
rendered environment" is not met by any task. **Fold:** list it under "Deferred, by name" for the no
branch, carried to rung 16 with the live-shadow dependency.

**m7. Task 6's parser re-implements `render_template`, and its mutation mutates test code.** `plan:782-787`,
`plan:804-806`. The Go parser can drift from `deploy.sh`'s, and "make the template parser keep a commented
row" proves the test detects a change to itself. Under Ruling 3's no this re-implementation is the whole
of exit gate clause 5. **Fold:** have the test drive the real function (`bash -c` sourcing only
`render_template`, if `deploy.sh` can expose it without running), or state plainly that Task 10 step 5 is
the render's only real proof and drop the mutation.

**m8. Task 6's fixture overrides are incomplete.** `plan:782-787`. The rendered template carries literal
paths (`/srv/music/library`, the three shadow directories, `/var/lib/musicbox`). If `Load` or a doctor row
stats any of them, the test fails on a workstation. **Fold:** quote which rows `Load` and `doctor` stat,
and override exactly those.

**m9. Task 10 does not require a clean tracked tree.** `plan:926-927`. The binary is stamped with `git
rev-parse HEAD`, so an uncommitted Go edit in the shared main tree ships under a reviewed commit's id
(R-5d13 guards only `web/`). **Fold:** step 1 requires `git status --porcelain --untracked-files=no` empty.

**m10. Three spec errata are missing from Task 11.** `plan:987-995`. The owned-prefix route list
(`spec:916-918`) gains the probe route; R6's "dials the service's own configured bind address over
loopback" (`spec:369-370`) becomes "dials each configured bind address from that address"; and
"Deployment"'s "`deploy.sh` gains the unit and the two pre-creates, and nothing else" (`spec:1759-1760`)
gains R-5d13's clean web build, which STATUS's carry owes.

**m11. The new unit shares the `serve` ping slug with the incumbent.** `plan:395`. Both units would
resolve `DUBPLATE_PING_SERVE_URL` once rung 9 mints it, conflating two services in one check. **Fold:**
a distinct slug (`server-serve`), still inert.

**m12. A hop carrying a port or brackets has no named state.** `plan:324-330`. `203.0.113.9:4711` and
`[2001:db8::1]` fail `netip.ParseAddr`, end the scan, and put every client behind such a proxy on the
proxy's key, which lets one client trip the per-client limit for all. **Fold:** name the state; either
accept `netip.ParseAddrPort` as a second parse (quoted) or refuse and document it in the runbook.

**m13. `install -d`'s behavior is recalled.** `plan:405-407`. Whether GNU `install -d -m -o -g` applies the
mode and owner to an existing directory, and what owner it gives a missing parent, decides whether the
loop fixes a drifted leaf and whether it must run after the shadow line's `chown` of `/var/lib/dubplate`.
**Fold:** quote `install --help` or `info coreutils install` and state the order.

**m14. The group 3 proof rests on an unquoted server behavior.** `plan:361-368`. The sentinel's zero count
is evidence only if `net/http`'s server does not itself remove `Connection`-listed headers; if it did,
the count would be zero with or without the strip. **Fold:** quote the `net/http` source (`readRequest`
removes no hop-by-hop header) in R-5d8; M6's no-strip proof over a real listener then demonstrates it.

**m15. Task 8's "by design" escape and one named carry.** `plan:869-871`. "An explicit 'by design'
pointer" can absorb any finding without a test. **Fold:** a by-design row cites the spec clause that
ratified the design. Name UNI's High "the forward-auth shape ... is never designed" as carried to rung 12
under Ruling 2's yes, since it is the one finding Ruling 2 directly re-routes.

**m16. The privileged-fields table omits `DUBPLATE_PUBLIC_URL`, and the sitting never exercises mail.**
`plan:271-282`, `plan:937-940`. The public URL sets the origin check, every link's base, and HSTS; its one
channel is the template literal through `server.env`. Separately, no step sends a message, so the SMTP
transport's first real use is the cutover. **Fold:** add the table row; record "SMTP transport not
exercised on the box" as a declared gap carried to rung 16.

## Planning-miss items, per task

| Task | Quoted behavior | Four states | Privileged channels | Fixture state |
| --- | --- | --- | --- | --- |
| 1 | Holds (`netip`, `textproto` quoted) | Holds; empty-hop skip implied by today's code; m12 | Holds | Holds |
| 2 | M7 (`InterfaceAddrs`) | M7 fifth state | Holds | M4, M5 |
| 3 | M6, m14 | m2 | m1 | M4, M5 |
| 5 | m13; `Type=exec` and `ReadWritePaths` quoted and hold | Holds for the leg's four states; M3 | Holds | Holds (evidence measured) |
| 6 | m8 | Holds | Holds | m8, M5 |
| 7 | M7 (the doctor invocation's properties) | n/a | m1 residual | n/a |
| 8 | n/a | n/a | n/a | m15 |
| 10 | M1, M2 | M1 | m16 | m9 |
| 11 | n/a | n/a | n/a | m10 |
