# Rung 5 identity and auth: adversarial universality review at HEAD

**HEAD:** 5262504. The review read the scoped files and the baseline records the brief named. It edited nothing, used no network, and touched no remote host.

**Exercised as a stranger:** I built the binary twice. One build used the working tree. The other used a clean `git archive HEAD`, so it carried only the placeholder web build. I wrote a stranger's env file by hand: shadow off, no proxy, `smtp.mail.example.net`, its own URL and site name. Under a scratch state directory I then ran `--help`, `doctor`, `check`, `member add/list`, `unidecode install --from <cached wheel>` and `serve`. I also ran `doctor` under `systemd-run --user` with the runbook's `RestrictAddressFamilies`. The netlink claim at step 6 reproduces exactly. Variants tested: an IPv6 `::1` bind with `::1` as the trusted proxy, prefix `/music`, and several `PUBLIC_URL` and bind-address shapes. Request-link POSTs used only unknown addresses or tripped the limiter before sending, so no mail was dialed.

**Estate leak verdict:** none in code. The scoped Go, the three auth pages and the runbook carry no estate hostname, domain, path, user, provider or personal name outside the four allowed homes. The one exception is a comment in `internal/notify/notify.go:20-21` naming Hetzner and "the box" (see the polish tail). The allowed homes do carry estate values, as the brief permits. Two of them break a stranger's install; they are folded into H2 and L2.

## The six focus points

1. **A configured header changing a trust decision: holds in code.** The topology guidance is missing (H1, M3).
   - With `::1` trusted on a dual-stack bind, a spoofed `X-Forwarded-For: 6.6.6.6` from `[::1]` logged `client_ip=6.6.6.6`.
   - The same header from the untrusted `127.0.0.1` logged `client_ip=127.0.0.1`.
   - No trust decision moved.
2. **An identity header surviving ingress: nothing new.** The strip wraps every path and every peer. The mixed-separator gap is the security lens's L1.
3. **doctor green on a configuration serve refuses, or misleading: partial.**
   - doctor reads `client ip direct` and exits 0 in the topology where one anonymous client locks out sign-in for the whole group (H1).
   - It is green when the state directory and the unit's grant disagree (M1), which extends the security lens's M3.
   - It hides the reason that `check` prints (L3).
4. **A grant wider or narrower than the writes: narrower on a stranger's host.** The grant is hard-coded to `/var/lib/dubplate` while the state directory and roster path are configurable (M1). The `User=music` the unit names is never created (H2).
5. **A secret in a log, the repo, or a rendered file: holds for the server.** The magic-link token lands in the stranger's proxy access log (L1). Modes held: the roster is 0600 and the leaves 0700. doctor and `check` print reasons and never values. No secret is in the template or the runbook.
6. **A deploy or install path that breaks, or only works here: breaks.** The manual path has no service-user step, no build step, no front or TLS step, and no sign-in URL (H2). The serve unit's `OnFailure=` depends on the estate's incumbent binary (L2).

## Findings, ranked

No critical findings.

### H1. Behind a same-host proxy in direct mode, one anonymous client locks the whole group out of sign-in, and doctor is green

**Where:**
- `internal/auth/limit.go:130` keys every client on its address, with no loopback case.
- `env/dubplate-server.env.template:124` ships `DUBPLATE_TRUSTED_PROXY_ENABLED=false`.
- `docs/server-runbook.md:308` calls running without a proxy "a chosen state".
- `docs/server-runbook.md:261-264` prices the per-client bound only as "a person with many client addresses".

**What the stranger hits:**
- TLS is mandatory for any non-loopback public URL (`server-runbook.md:360-367`), and the bind defaults to `127.0.0.1`. So the stranger's natural topology is Caddy, nginx, Traefik or cloudflared on the same host.
- The template and runbook lead them to leave the trusted proxy class off.
- Every request then keys on `127.0.0.1`. Twenty requests from anyone exhaust `LINK_REQUESTS_PER_CLIENT_PER_HOUR` for every member, and 30 consumption attempts exhaust the consume bound.
- doctor reads `client ip direct` and `boundary probe 127.0.0.1 refused`, and exits 0.

**Reproduction (run):** start `serve` with the stranger env. Make 21 POSTs to `/dubplate/request-link`, each with a distinct `X-Forwarded-For` and unknown address. The responses read `200 ×20, 429`. The admin's own request then answers `{"outcome":"rate_limited"}` 429, and the log reads `client_ip=127.0.0.1 reason="...link requests per client"`.

**Relation to the baseline:** the estate's direct mode is a recorded residual (plan:567-585), because its cutover pins the tunnel container. A stranger has no cutover. Nothing tells them that direct mode behind a proxy collapses the group onto one key.

**Fix:**
- *Runbook:* state that any same-host front must be configured as trusted, listing `127.0.0.1` and `::1` when the proxy may dial either.
- *serve:* warn once at startup or first sight when, in direct mode, a request from a loopback or private peer carries a group 2 header. That signal is the strip's own `countPresent`.
- *doctor:* make the `client ip direct` row say what it assumes, for example "direct: the peer must be the browser".
- *Interaction:* once trusted, the same-host topology hits the security lens's L2. The probe then reads "not probed: a trusted proxy" and proves nothing. L2's fix (let a trusted peer receive the identity count) is therefore needed for the default stranger topology, not an edge case.

### H2. The install order ends at a server no stranger can reach or sign in to

**Where:**
- `docs/server-runbook.md:16-99` (Install order).
- `systemd/dubplate-server-serve.service:26`.
- `cmd/dubplate-server/web.go:106-108`.

**What the stranger hits, in order:**
1. **Step 1, the binary.** It says "copy the binary" but never says how to make one. A plain `go build` from a clean checkout succeeds, and its `doctor` then reads `web invalid: ... run npm run build in web/`. The real path (Node 24 or later, `make web`, then `make build-linux`) appears nowhere in the runbook. doctor names the remedy, so this costs a detour, not a silent failure.
2. **Step 3, the service user.** `install -d -o music -g music` fails with "invalid user" on a fresh host. No step creates the `music` service user. Nothing says the unit's `User=` and the `systemd-run` lines must change together if the stranger picks another name.
3. **No step installs a front.** HTTPS is required (`:360-367`), and the default bind is loopback-only. Yet the order goes from "start the unit" straight to done:
   - no proxy or TLS step;
   - no `DUBPLATE_SERVE_BIND_ADDRS` guidance for an off-host proxy;
   - no `DUBPLATE_SERVE_PREFIX` mention.
4. **No sign-in URL.** The runbook never names `/dubplate/login`, and `/` answers a bare `404 page not found`. I curled `/`, `/dubplate`, `/dubplate/` and `/login`: all 404, and only `/dubplate/login` answers 200. The admin's add-member page, `/dubplate/members`, is also unnamed. The link page's signed-in state offers no link onward.

**Fix:** add a step 0 for the build and the service user. Add a step between 6 and 7 that installs the TLS front, with one example for a same-host Caddy or nginx and one for cloudflared, each including the trusted proxy rows (H1). End the order with "open `https://<host>/dubplate/login`; admins manage members at `/dubplate/members`." Consider a redirect from `/` to the login page while the fallthrough is empty.

### M1. The grant is pinned to `/var/lib/dubplate` while the paths it must cover are configurable, and doctor cannot see the mismatch

**Where:**
- `systemd/dubplate-server-serve.service:49` and `docs/server-runbook.md:78` hard-code the two leaves.
- `DUBPLATE_STATE_DIR` and `DUBPLATE_ROSTER_FILE` are free values (`env/...template:57,84`).
- `cmd/dubplate-server/serve.go:265` (`auth.Open`) and `:275` (`OpenRoster`).
- The gate's agreement leg checks only the estate's own template.

**What the stranger hits:**
- They put state on a data disk (`DUBPLATE_STATE_DIR=/srv/dubplate`) and follow step 3 as written. `serve` then fails at `auth.Open` with a read-only file system error (`ProtectSystem=strict`), while doctor, run through the runbook's `systemd-run`, exits 0.
- Worse is a roster path outside the roster leaf, for example `/var/lib/dubplate/roster.json`. Both `serve` and `doctor` start green, because `OpenRoster` only reads. The admin's first web "add member" then fails on its atomic rename. `member add` under `systemd-run` fails too, which is the only early signal.
- Nothing at runtime or in doctor relates the configured paths to the granted ones, and the runbook never says they must agree.

**Fix:**
- *doctor:* add the write probe the security lens's M3 proposes: an `auth store` row that checks the leaf exists, its mode and owner, and that it is writable. Add a `roster dir writable` check, run as the effective user under the unit's namespace.
- *serve:* check at startup that `filepath.Dir(roster)` is writable, so the failure comes at start rather than at first use.
- *Runbook:* say the unit's `ReadWritePaths`, step 3 and the `systemd-run` lines must name `$DUBPLATE_STATE_DIR/auth` and the roster file's own directory.

### M2. Mail works only with authenticated PLAIN over TLS, so a LOGIN-only provider or a local relay cannot send

**Where:** `internal/notify/notify.go:117` forces `mail.SMTPAuthPlain`, and `:124` forces `TLSMandatory`. All five SMTP keys are required (`cmd/dubplate-server/auth.go:51-55`).

**The mechanism (quoted from go-mail v0.8.1 `client.go:1516-1519`):** `case SMTPAuthPlain: if !strings.Contains(smtpAuthType, "PLAIN") { return ErrPlainAuthNotSupported }`.

**What the stranger hits:**
- A submission server that advertises only `AUTH LOGIN` fails every send. Microsoft 365's `smtp.office365.com` is the common one; it advertises LOGIN and XOAUTH2 but not PLAIN.
- A stranger relaying through a local MTA on `localhost:25` with no credentials or no STARTTLS certificate cannot configure it at all. Empty user or password rows are refused, and STARTTLS is mandatory.
- doctor never dials (`server-runbook.md:203-204`), so the first evidence is a failed sign-in after the install is otherwise green. The handler answers `failed` with status 500 and logs the reason.

**Fix:**
- Use `mail.SMTPAuthAutoDiscover`, present in the pinned v0.8.1 (`auth.go:162`).
- Decide whether a loopback relay without auth is a supported class. If yes, make the user and password optional together and allow opportunistic TLS only for a loopback host. If no, say so in Mail.
- Add a `doctor --mail` or `notify test` verb that sends one message, so mail is checked before go-live.

### M3. A CDN-fronted origin cannot configure the trusted proxy, and the runbook does not say what to do

**Where:** `cmd/dubplate-server/proxy.go:101-107` refuses networks, and `docs/server-runbook.md:308-315` justifies the refusal.

**What the stranger hits:**
- Behind Cloudflare's proxied DNS (not a tunnel), Fastly or Bunny pointed at a public bind, the peer is one of thousands of edge addresses that publish only as ranges. The stranger can neither list them nor enter a range.
- They stay in direct mode. Every member behind one point of presence shares that edge's bucket with any attacker routed through it. That is H1's lockout, scoped per edge.

**Fix:** keep the address pin, a sound design choice, and document the supported answer: run a local proxy that validates the CDN's published ranges (Caddy `trusted_proxies`, nginx `set_real_ip_from` with `real_ip_header`), then trust only that local proxy. Name the CDN-direct topology as unsupported.

## Low

### L1. The magic-link token lands in the stranger's proxy access log

**Where:** `internal/auth/token.go:181` puts the token in the query (`?t=`).

**What happens:** the server never logs `RawQuery`, but nginx's default `combined` log, Caddy's access log and CDN log pipelines record the full request line. The GET only renders a confirm button (`web/src/routes/login/link/+page.svelte:1-5,86`). The token stays redeemable until the click or expiry, and anyone who reads the proxy log in that window can consume it first.

**Relation to the baseline:** the threat model names the mail scanner and browser history (spec:1990-1994), not the proxy log.

**Fix:** carry the token in the fragment (`#t=`); the client hook already strips it before the router reads the URL. Otherwise, add a runbook line telling operators to drop the query from the front's access log.

### L2. The serve unit's failure alert depends on the estate's incumbent

**Where:** `systemd/dubplate-server-serve.service:12` uses `OnFailure=dubplate-ping-fail@server-serve.service`. That unit runs `/usr/local/bin/dubplate ping-fail` with `/etc/dubplate/env`, both estate-only.

**What happens:** a stranger following step 1 installs neither. A crash loop that trips `StartLimitBurst` therefore alerts nobody and logs a "unit not found" failure.

**Fix:** tell the stranger to delete or replace the `OnFailure=` line. Later, ship a generic alert hook. Also correct the estate-specific `After=docker.service` comment (`:3-7`), which the security lens's L4 noted for the bind claim.

### L3. doctor hides the reason that `check` and `serve` print, and the runbook does not route the stranger to it

**What I saw:**

| Configured value | doctor | `check` / `serve` |
| --- | --- | --- |
| `PUBLIC_URL=https://example.org/music` | `unsatisfied: DUBPLATE_PUBLIC_URL invalid` | `DUBPLATE_PUBLIC_URL: must carry no path; the serve prefix sets where dubplate lives` |
| `DUBPLATE_SESSION_SECRET=short` | `unsatisfied: DUBPLATE_SESSION_SECRET invalid` | `too short, want at least 32 bytes` |

These reasons carry no secret. A stranger who wants a sub-path mount on a shared domain learns only from `check` that the public URL must be a bare origin.

**Fix:** print the non-value reason in the doctor row, or have the runbook say "run `check` for the reason". State in step 2 that `DUBPLATE_PUBLIC_URL` is a bare `https://` origin and that the secrets need at least 32 random bytes.

### L4. A privileged `DUBPLATE_SERVE_PORT` passes doctor and crash-loops serve

**What happens:** the unit sets `CapabilityBoundingSet=` empty, so a port below 1024 (for example 80) fails at bind with permission denied. doctor run before step 7 reads `no listener`, which does not fail.

**Fix:** refuse a port below 1024 in the threshold validator, beside the security lens's L7 port bound, or name the capability in the error.

## Polish tail

- `internal/notify/notify.go:20-21` names Hetzner and "the box" in a library comment. Generalize it to "some hosts block outbound 465".
- The runbook never says how to generate the two HMAC secrets. Give a one-line command, for example `openssl rand -base64 48`.
- `DUBPLATE_ADMIN_EMAIL` (today a MusicBrainz user-agent contact) and `DUBPLATE_LIBRARY_ROOT` (required, read by nothing at rung 5) are unexplained to a stranger.
- nginx defaults that bite this design belong in the H2 example:
  - plain `proxy_pass` forwards a client's `X-Forwarded-For` unchanged, which the runbook warns about only abstractly (`:325-330`);
  - nginx rewrites `Host` to the upstream, which affects only browsers that send no `Sec-Fetch-Site`.
- A dual-stack same-host proxy may dial `::1` or `127.0.0.1`. The runbook's example lists one address family.
- A link-local bind with no zone reads `present` in `bind addresses`, then fails the probe with `bind: invalid argument`. The bind row could refuse a link-local address that has no zone.
- On a systemd-free host (Alpine/OpenRC, a container) there is no documented path. Say so in the runbook's opening scope.
- `systemd-run -p EnvironmentFile=` was accepted on systemd 259. Older stable distributions remain unverified, as the runbook already says.
