# Rung 5 identity and auth: adversarial security review at HEAD (6e661a6)

**Scope read in full:** `internal/serve/{strip,clientaddr,serve,probe,origin,headers,normalize,server,bindaddrs}.go`, `internal/auth/{handlers,admin,token,session,nonce,store,limit}.go` plus the relevant parts of `roster.go` and `epoch.go`, `internal/identity/identity.go`, and `internal/notify/notify.go`. Also `cmd/dubplate-server/{serve,auth,member,doctor,doctorprobe,proxy}.go`, both dubplate-server units, `scripts/deploy.sh`, the server env template, the `/login/link` page with its token capture, and the runbook's install and rollback sections.

**Tests:** `go test -count=1` passes for `./internal/{serve,auth,identity,notify}`.

**Baseline:** findings already recorded as residuals, or marked verified sound in earlier reviews, are not repeated. That covers residue items 1-9, the trusted-proxy direct mode in the estate's template, and `FrontExemption`.

## Verdicts on the six focus points

**1. A configured header changing a trust decision: holds.**
- Every trust decision reads `PeerAddress`, which comes from `RemoteAddr` alone (`clientaddr.go:131`). That covers the strip's group 2/3 carve-out (`strip.go:472`), the probe route's permit rule (`probe.go:92-100`) and `FrontExemption` (`strip.go:360`).
- `ClientIP` reaches only limiter keys and log or audit fields. A grep of every caller finds `handlers.go:191,336,382`, `admin.go:127,208,228` and the `AddMember` audit argument. The admin gate, the origin check and `LinkURL` never read it, and `LinkURL` uses the configured `PublicOrigin` (`token.go:172`).
- The header is read only when `trusts(peer)` is true (`clientaddr.go:173`), after unmapping the peer address (`:196`).
- The hop scan reads every header line, right to left. A malformed rightmost untrusted hop ends the scan and falls back to the peer (`:218-221`), so a planted value never steers the result. Empty hops are skipped, but only an attacker writing on the left side can create one.

**2. An identity header surviving ingress: holds for both stated spellings, with one Low gap.**
- The dash and all-underscore forms are deleted from `Header` and `Trailer` on every request. The trailer is scrubbed again on both `Read` at EOF and `Close` (`strip.go:359-454`).
- The strip wraps the whole mux, so every path and every peer is covered (`serve.go:176`). The production shell passes the zero exemption (`cmd/.../serve.go:259`).
- A `Connection` listing can only delete headers, never keep one.
- Unencrypted HTTP/2 is not enabled, and `OPTIONS *` reaches no handler.
- The gap: a name that mixes `-` and `_` survives (L1).

**3. doctor green while serve would refuse, or while a spoofed header passes: partial.**
- doctor never reports green while the fixed-family group 1 or group 2 headers pass from an untrusted peer. The probe sends both groups plus the `Connection`-listed sentinel through the real listener (`probe.go:46-60`).
- Two gaps:
  - doctor never runs serve's auth-store opening (M3).
  - The probe never exercises the configured client-address header, and exits 0 without proving anything when every bind address is a trusted proxy (L2).

**4. Grant against writes: holds.**
- Every write serve makes lands inside `ReadWritePaths=/var/lib/dubplate/auth /var/lib/dubplate/roster`:
  - the nonce create (`O_EXCL`, `nonce.go:155`), the sweep's `Remove`, and directory fsyncs;
  - the `MkdirAll` of the `nonce` and `epoch` subdirectories (`store.go:62`);
  - the epoch `atomicfile.WriteStream` (`epoch.go:162`);
  - the roster's atomic rename and `flock` on its directory (`roster.go:404`).
- Nothing else is written; settings-store wiring is still unwired (STATUS).
- deploy creates both leaves `music:music 0700`, which `auth.Open` accepts.
- Neither grant carries a `-` prefix, so a missing pre-create fails at namespace setup instead of with EROFS on the first write.
- `UMask=0027` and the explicit modes (0700 directories, 0600 records) agree.
- The only grant wider than its writes is deploy's own (L5).

**5. Secrets: one Medium (M1).**
- `server.env` is installed by piping stdin into `install -D -m 600 -o root -g root` (`deploy.sh:553`). It is never staged on disk.
- `--check` diffs print no content, and the missing-value report names keys only.
- Logs carry the path, never the query (`handlers.go:337`, `document.go:267`).
- Mail errors log go-mail's `Reason` only, never `SendError.Error()`, which would list recipients (`notify.go:137`).
- doctor prints no value of the identity source.

**6. Deploy safety: holds in the script; the rollback is procedural and incomplete (L3).**
- With no flag, deploy only checks (`deploy.sh:102-130`).
- The serve unit is rsynced and daemon-reloaded but never enabled. `TIMER_UNITS` excludes both the serve unit and the shadow timer (`:659`).
- The table install runs on the staged binary before the swap (`:499-517`).
- The swap-before-env hazard (`:517` against `:558-565`) is recorded in plan:1291-1303 and the runbook's rollback copies. The script itself still encodes it.

## Findings, ranked

No Critical or High findings.

### M1. deploy.sh exports the workstation's secret set into npm build code, in check mode too
- **Where:** `scripts/deploy.sh:135-138` sources `~/.local/secrets`. On this workstation, 85 of its 88 lines are `export` (names checked, never values), including `CLOUDFLARE_API_TOKEN`, `HCLOUD_TOKEN` and the SMTP password; from Task 10 it will also hold both HMAC keys.
- `:316-320` then runs `npm ci` and `npm run build` with that environment. Vite, its plugins and any future install script execute third-party code.
- **Attack:** one compromised build dependency reads `process.env`. With the session key it can forge an admin session for any roster address, which is full roster control. With the Cloudflare token it gets the whole account.
- This now happens on every `--check`, including agent runs in non-interactive shells that would otherwise not hold the secrets.
- **Reproduction:** add `console.log(Object.keys(process.env).filter(k=>/SECRET|TOKEN|PASSWORD/.test(k)))` to `vite.config.ts` and run `deploy.sh --check`.
- **Fix:** run the local build before the `source`, or under `env -i PATH="$PATH" HOME="$HOME"`. Add `--ignore-scripts` to `npm ci`; the only install script in the lockfile is `fsevents`, which is macOS-only.

### M2. The clean-tree refusal misses untracked build inputs outside `web/`
- **Where:** `deploy.sh:297-303` passes `--untracked-files=no` everywhere except `web/`.
- Two classes of input reach the box from an unreviewed tree:
  - An untracked `.go` file in any package, or an untracked `internal/engine/store/migrations/*.sql` (embedded at `migrations.go:13`), is compiled into a binary stamped with the clean commit's id.
  - An untracked or gitignored file under `config/` (the root `.gitignore` ignores `*.env`) is shipped by `rsync --delete` to `/opt/dubplate/config` (`:332`). rsync ignores `.gitignore`.
- **Reproduction:** `touch config/local.env; echo 'package main' > cmd/dubplate-server/zz.go; scripts/deploy.sh --check`. The refusal does not fire, the build includes `zz.go`, and the dry run lists `config/local.env`.
- **Fix:** refuse any untracked or ignored file under the build and rsync inputs: `git status --porcelain --ignored -- cmd internal web config compose.yaml go.mod go.sum`, excluding `web/node_modules`, `.svelte-kit` and `build`. A stronger alternative is to build and rsync from a `git archive` of `HEAD`.

### M3. doctor exits 0 on an auth store that serve refuses
- serve refuses at startup in two places doctor never runs:
  - `auth.Open`'s leaf-mode check (`store.go:66-72`) and its `MkdirAll` of the subdirectories;
  - `OpenNonceStore`'s startup sweep, which fails on a directory-listing error or a failed `Remove` (`nonce.go:200-247`).
- Both are reached through `serve.go:265-272`. doctor's failure set (`doctor.go:183-184`) includes neither.
- **Reproduction:** `chmod 0750 /var/lib/dubplate/auth`, or create the leaf as root `0700`. doctor run through the runbook's `systemd-run` exits 0. `serve` exits with `open auth store: leaf … has mode 750`, or a permission error from `MkdirAll`.
- **Test sketch:** in `doctor_test`, give a fully valid environment and admin roster a state directory whose `auth` is `0o750`. Assert doctor fails and names the leaf; it passes today.
- **Fix:** add an `auth store` row that runs read-only checks: stat the leaf (mode, owner equal to the effective uid), list the `nonce` directory, and confirm both subdirectories are writable by the effective uid. A wider leaf mode should fail in the same words serve uses.

### L1. Mixed-separator identity spellings survive the strip
- `identitySpellings` (`strip.go:210-225`) produces the all-dash and all-underscore forms only. A name with three or more segments has mixed forms with their own canonical keys:

  ```
  X-Forwarded_User -> X-Forwarded_user   (survives; verified with go run)
  X_Forwarded-User -> X_forwarded-User
  ```
- A downstream that treats `-` and `_` alike (CGI or WSGI `HTTP_X_FORWARDED_USER`, or nginx with `underscores_in_headers`) reads such a spelling as the identity header.
- Inert at rung 5, which has no fallthrough. Navidrome's `Remote-User` has only two segments, so both of its forms are covered.
- **Test sketch:** set `req.Header["X-Forwarded_user"]`, run `Strip.Wrap`, and assert the key is absent.
- **Fix:** match on the normalized key instead of a spelling list. Delete every key `k` where `CanonicalMIMEHeaderKey(strings.ReplaceAll(k, "_", "-"))` is in the dash set. Count the same way in `identitySpellingPresent` and `countPresent`.

### L2. The probe never exercises the header `ClientIP` actually reads, and can exit 0 having proven nothing
- **The configured header:** when it lies outside the fixed family (for example `X-Client-Ip`), the strip removes it from untrusted peers (`strip.go:301`). The probe request and the probe route still use the fixed `forwardedAddressHeaders` list (`probe.go:54,111`), so doctor never proves that removal.
- **All bind addresses trusted:** with a same-host front configured as trusted on 127.0.0.1, every bind address is skipped. The row reads `not probed: every local address is a trusted proxy` and returns not-failed (`doctorprobe.go:265-267`), although the identity strip applies to trusted peers too.
- **Fix:**
  - Have the probe request and route take the strip's own forwarded set, configured header included.
  - Allow a trusted peer to receive the identity count only.
  - Mark the all-trusted case "unproven" rather than silently passing.

### L3. The rollback is incomplete, and the ledgers disagree on its method
- The runbook's rollback copies (`server-runbook.md:536-555`) cover the binary and the two env files. They do not cover:
  - the beets upgrade that apply runs as `music` (`deploy.sh:690`, pinned 2.14.1), which touches the live import pipeline, although plan:148 promises a restore to 2.13.1;
  - the rsynced unit files (`:341`);
  - `/opt/dubplate` (compose and `config` under `--delete`, followed by `up -d`).
- STATUS's carried transliteration row still says "rollback by redeploy". The runbook (`:527-533`) correctly rejects redeploying as a rollback.
- Redeploying any commit before `c179c99` runs a `deploy.sh` that applies by default, with no flag (verified at `e8fadec`: `check_only=false`). An operator "previewing" a rollback would write the box.
- **Fix:**
  - Add the beets version and `.prev` copies of `/opt/dubplate` and the units to the rollback set.
  - Correct the STATUS row.
  - Better: have `--apply` write the `.prev` copies and stop the shadow timer itself, so the safety does not depend on the sitting's procedure.

### L4. The serve unit's bind claim does not match the template, and a port collision is possible
- `dubplate-server-serve.service:3-7` says serve binds the bridge gateway alongside loopback.
- The template sets neither `DUBPLATE_SERVE_BIND_ADDRS` nor `DUBPLATE_SERVE_PORT`, so serve takes 127.0.0.1:8080.
- The incumbent `dubplate-serve.service` takes its bind address and port from age-store values I cannot see. If it holds 127.0.0.1:8080, Task 10's "start once" crash-loops (`Restart=on-failure`). doctor fails closed in that case (`unexpected answer: 404`), but the cause is not named.
- **Fix:** set an explicit, distinct port in the template and correct the unit comment until rung 16's bridge bind.

### L5. Grants and a uid wider than the writes in deploy
- The `unidecode install` `systemd-run` grants all of `/var/lib/dubplate` (`deploy.sh:510`), which includes `auth/` and `roster/`. The verb writes only `unidecode/`. Pre-create that leaf and grant it alone.
- `runuser -u music uv tool install --force beets…` (`:690`) runs PyPI install code unsandboxed as the uid that owns `roster.json` (0600) and the epoch store. Such code can grant itself the admin flag with no audit record, or delete an epoch record and undo a revocation.
- This is the same class as recorded residual 2 (shared `music` uid, carried to rung 16). The residual's text names secrets only and should also name roster and epoch integrity.

### L6. `CheckProxyHeader` admits names that carry no address or that serve needs
- `clientaddr.go:68-84` refuses `Forwarded` but accepts `X-Forwarded-Host`, `X-Forwarded-Proto` and `X-Forwarded-Prefix`.
  - Every hop in these is malformed, so `ClientIP` always falls back to the proxy and all clients share one limiter bucket.
  - That makes a lockout of the login surface for everyone one request flood away.
- It also accepts `Content-Type`. The strip would then remove that header from every untrusted peer, and the roster POST from a direct peer would always fail at `admin.go:237`.
- **Fix:** refuse the rest of the forwarded family other than the address-bearing names, plus `Content-Type`.

### L7. `DUBPLATE_SERVE_PORT=0` passes both serve and doctor
- `config.ParseInt` accepts 0 (`cmd/.../serve.go:154`). serve then listens on an ephemeral port.
- doctor's probe dials port 0, reads "no listener" and passes.
- **Fix:** bound the port to 1-65535 the way `validateSMTPPort` does.

## Polish tail
- `auth.Open` checks the leaf's mode but not its owner or the subdirectories' modes (`store.go:66-72`).
- A bracketed IPv6 hop without a port (`[2001:db8::1]`) parses as malformed (`clientaddr.go:238-251`), which collapses that client to the proxy's bucket.
- The comment at `strip.go:159-162` cites Fastly adding `Fastly-Client-Ip` only "when a request arrives without one", so a client can set it through Fastly. `CheckProxyHeader` accepts it as a trusted header; document this or refuse it.
- Group 2 is not scrubbed from trailers (`strip.go:460-467`). Nothing reads trailers today; rung 12's proxy should.

**Verdict:** the code holds its six claims: trust decisions read only the peer address, the identity strip covers every path, peer and stated spelling, and the serve grant matches its writes. What needs fixing before Task 10 is in the deploy and doctor tooling: build code runs with the full secret set (M1), unreviewed files can still reach the box (M2), and doctor can pass a store serve refuses (M3).
