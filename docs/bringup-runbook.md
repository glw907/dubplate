# Bring-up runbook: from reprovisioned box to seeded library

Self-contained: an agent or session with no prior context executes this top to
bottom. State assumptions: server `musicbox` freshly recreated (2026-08-30) from the
corrected cloud-init (all five package fixes); volume `musicbox-data` survived with
Navidrome's DB (admin geoff@907.life) and the /srv/music tree; every secret exists in
`~/.local/secrets` (source it per shell); repo clean at or past `f054b91`. Never
print a secret value; name-and-length checks only. On-box work over SSH as root
(key auth); IP from `hcloud server ip musicbox`.

1. **Gate on first boot.** `cloud-init status` must be `done` (poll; two reboots are
   designed in). Then verify: `getenforce` = Enforcing; docker active; volume
   mounted; `flac ffmpeg sqlite3 rclone bsdtar` all present (flac's package also
   carries `metaflac`); tree ownership 2000:2000 with inbox setgid. Any miss: stop
   and report; do not patch the box. `dubplate doctor` replaces the rest of this
   step's hand verification (config resolution, SMTP/R2/Navidrome reachability,
   ping URL syntax) once the binary is deployed — run it as step 2's smoke test
   rather than checking each dependency by hand.
2. **Deploy.** `bash scripts/deploy.sh --check`, review, then `bash scripts/deploy.sh`.
3. **Acceptance.** Containers healthy with zero restarts after 60s; tunnel healthy
   (read-only GET on cfd_tunnel b6781175-000b-40c3-a54a-05906ab25037); from the
   workstation, music.907.life serves Navidrome (no Access) and inbox.907.life
   challenges via Access; both timers active; a manual `systemctl start
   dubplate-import.service` no-ops cleanly with a success ping; beets loads all
   three plugins (replaygain must load now that ffmpeg exists).
4. **First backup.** `systemctl start dubplate-backup.service`; verify success in
   the journal, then from the workstation (RO token) `rclone lsjson` shows the
   pinned layout (`state/` holding both DB snapshots). Backup ping green.
5. **Seed.** rsync the nine ZIPs from the workstation `~/Downloads/` into
   `/srv/music/inbox/geoff/` (as root, then chown 2000:2000). Wait for the timer (or
   start `dubplate-import.service` once). Track per-album outcomes: imported (visible via
   Subsonic API search on music.907.life with admin creds from env, never printed) vs
   review-queued (with reasons). Rejection or review emails are expected for any
   MusicBrainz misses; that path working is a pass, not a failure.
6. **Post-seed backup.** `systemctl start dubplate-backup.service` again; verify
   `library/` objects appear in R2; note total object count and approximate size.
7. **Monitoring truth.** All three push checks green (backup, import, disk pings
   observed in journals); the `dubplate-navidrome` push check (driven by its box timer) —
   verify the pings Worker's next cron ticks record probe success (read-only
   `wrangler tail` briefly or KV state read of the pull check's ping:/alert: keys by
   name).
8. **Ledgers.** Update docs/STATUS.md (seeded state, per-album table, next actions:
   Task 7 restore drill, Task 8 accounts + onboarding, pings P2c drill) and
   HISTORY.md; commit with the standard footer
   (`Co-Authored-By: Claude <noreply@anthropic.com>`).

Report shape: per-step results; per-album outcome table; anything degraded;
couldNotDo.

## Kept identifiers

`musicbox` (server), `musicbox-data` (volume), and the pings check names
above are live Hetzner/infrastructure identifiers, not stale references —
see README.md's own kept-identifiers note.

## Uploader surface bring-up (U12a, box task, Geoff-sanctioned)

Self-contained like the section above: an agent or session with no prior
context executes this top to bottom once the estate is already live (the
section above has run at least once) and Pass A (U1–U8) is code-complete
in the repo. Same rules: never print a secret value, name-and-length
checks only; on-box work over SSH as root; this is production — check for
a live executor first (`pgrep -f Projects/dubplate`, `git status`) and
read `.claude/instructions/api-access.md` for which token covers which
call. `$CF_ZT_TOKEN` and `$CLOUDFLARE_API_TOKEN` are both sourced from
`~/.local/secrets`; account id `120c269ad6d3dfbe6d63a0bb53758ca0`, zone
`907.life`. Goal: `dubplate serve` reachable through Access at a real
hostname, `/api/*` verified end to end, FileBrowser still up (it retires
in U12b, not here).

**Live estate facts (decided, not operator choices):** the uploader has
no hostname of its own — it lives at `https://music.907.life/upload`,
path-scoped on the estate's existing hostname (`internal/serve`'s own
`basePath` constant owns the `/upload` prefix server-side, since
cloudflared forwards the request path intact rather than stripping it).
`DUBPLATE_SERVE_PORT=8090`,
`DUBPLATE_ACCESS_TEAM_DOMAIN=aksailingclub.cloudflareaccess.com`,
`DUBPLATE_SERVE_BIND_ADDRS=127.0.0.1,172.30.0.1`. The Access application
("dubplate uploader", id `b9d9f27d-bfbf-4c66-a437-5b3074785eee`) already
exists, with its domain being set directly (by Geoff, via the API) to
the path-scoped value `music.907.life/upload` on this existing app —
step 1's recipe below shows that same path-scoped domain for a genuine
rebuild. **No new DNS record is needed**: `music.907.life` already
routes to the tunnel (Task 5's existing CNAME); the path-scoped ingress
rule in step 3 is what makes `/upload` reach serve, nothing at the DNS
layer changes. `DUBPLATE_ACCESS_AUD` is already minted in the age store.
Two more U1-introduced names must also already resolve in
`~/.local/secrets` before `deploy.sh` can render `/etc/dubplate/env` at
all (the renderer fails closed, listing every missing name, if any
template entry is unset — env/dubplate.env.template requires all of
them, not only this pass's new ones): `DUBPLATE_UPLOADS_DIR=
/srv/music/uploads` (the tus store; deploy.sh's own pre-create step
`mkdir -p`s and `chown`s this path, but the env value itself still has
to be minted separately) and `DUBPLATE_TRIGGER_DIR=/run/dubplate` (the
nudge trigger queue's parent; `internal/serve` appends the
`nudge-triggers` subdirectory itself, so this value is the bare
`/run/dubplate`, not the full path the tmpfiles fragment creates).
Every one of these six values reaches the box exactly one way:
`~/.dotfiles/scripts/secrets/secret-set.sh NAME --value ...`, which
regenerates `~/.local/secrets` — the file `scripts/deploy.sh` sources to
render `/etc/dubplate/env`. Never a plain shell `export`: a value only
exported in the current shell is silently overridden the moment
`deploy.sh` sources the regenerated `~/.local/secrets`, so an export
would look like it worked and then vanish at deploy time. If any of
these six values is confirmed missing from `~/.local/secrets`
(name-and-length check only), mint it with `secret-set.sh` before
continuing; do not re-create the Access app if it already exists (step 1
below gives the creation recipe for a genuine rebuild only).

1. **Access application and roster policy (already created; recipe
   below is for a rebuild only).** Self-hosted, domain path-scoped to
   `music.907.life/upload` (NOT a hostname of its own — this app must
   coexist with whatever Access policy, if any, governs bare
   `music.907.life`, since Cloudflare Access matches applications by
   longest path match), restricted to the account's `onetimepin`
   identity provider (`e33c7c2d-facd-446c-96de-8215f64cde76` — confirm
   still current via `GET .../access/identity_providers` before reusing
   it), `auto_redirect_to_identity: true` since it is the only allowed
   IdP. `$CF_ZT_TOKEN`, not `$CLOUDFLARE_API_TOKEN` — the general-purpose
   token 403s reading Access identity providers (HISTORY's own
   rediscovery note).

   ```
   curl -s -X POST \
     "https://api.cloudflare.com/client/v4/accounts/120c269ad6d3dfbe6d63a0bb53758ca0/access/apps" \
     -H "Authorization: Bearer $CF_ZT_TOKEN" -H "Content-Type: application/json" \
     -d '{
       "name": "dubplate uploader",
       "domain": "music.907.life/upload",
       "type": "self_hosted",
       "session_duration": "24h",
       "allowed_idps": ["e33c7c2d-facd-446c-96de-8215f64cde76"],
       "auto_redirect_to_identity": true
     }' | jq -r '.result | {id, aud}'
   ```

   Pipe the response through `jq` as shown — never let the full body
   reach a transcript or log (the same standing rule Task 5's tunnel
   rotation exists because of). On the live app (id
   `b9d9f27d-bfbf-4c66-a437-5b3074785eee`), the roster allow policy is one
   `include` entry per email in `config/contributors.yaml`'s roster (the
   same shape as `inbox.907.life`'s `92425260-87f6-4a3a-b66e-d71bb8d9d6a8`):

   ```
   curl -s -X POST \
     "https://api.cloudflare.com/client/v4/accounts/120c269ad6d3dfbe6d63a0bb53758ca0/access/apps/b9d9f27d-bfbf-4c66-a437-5b3074785eee/policies" \
     -H "Authorization: Bearer $CF_ZT_TOKEN" -H "Content-Type: application/json" \
     -d '{
       "name": "roster allow",
       "decision": "allow",
       "include": [{"email": {"email": "geoff-login@907.life"}}]
     }'
   ```

   Add one `include` object per roster email (comma-separated in the
   array) rather than one policy per person.
2. **`DUBPLATE_ACCESS_AUD` (already minted).** The value is the `aud`
   field from step 1's app-creation response, minted via `secret-set.sh`
   as described above — the env var name was already declared in U1, so
   deploy only ever renders it. Verify by name and length only (64 hex
   chars) against `~/.local/secrets`, never printed.
3. **Tunnel ingress PUT (the remaining actionable step from here down;
   no DNS change accompanies it — see the live-facts note above).**
   `config/cloudflared-tunnel-config.json` already carries the path-scoped
   rule, inserted ABOVE the existing bare `music.907.life` rule (ORDER IS
   LOAD-BEARING — the file's own `_comment` says so): `{"hostname":
   "music.907.life", "path": "^/upload(/|$)", "service":
   "http://host.docker.internal:8090"}` — the regex is anchored to a path
   boundary (`(/|$)`, not a bare `^/upload`) so an unrelated future route
   merely sharing the string prefix (`/uploads-report`, say) can never
   leak to serve instead of Navidrome. cloudflared evaluates ingress
   rules top to bottom and stops at the first match, so this rule must
   precede the bare `music.907.life` → Navidrome rule or that rule would
   swallow every `/upload` request first. `host.docker.internal` resolves
   through `compose.yaml`'s literal `extra_hosts` pin to `172.30.0.1`
   (the pinned bridge gateway, NOT Docker's `host-gateway` magic value,
   which resolves to the *default* bridge's gateway, typically
   `172.17.0.1`, on this multi-network box — using the magic value here
   would have silently pointed cloudflared at the wrong interface). PUT
   the whole config object once this file's contents match the live
   tunnel:

   ```
   curl -s -X PUT \
     "https://api.cloudflare.com/client/v4/accounts/120c269ad6d3dfbe6d63a0bb53758ca0/cfd_tunnel/<tunnel_id>/configurations" \
     -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" -H "Content-Type: application/json" \
     -d @config/cloudflared-tunnel-config.json
   ```

   Verify with a read-only `GET` on the same path: four ingress rules,
   the path-scoped `music.907.life`/`^/upload` rule first, the bare
   `music.907.life` rule second, before the catch-all.
4. **Deploy.** `bash scripts/deploy.sh --check`, review, then `bash
   scripts/deploy.sh`. This rsyncs `dubplate-serve.service` and the
   `dubplate-import-nudge.path`/`.service` pair; installs and activates
   `/etc/tmpfiles.d/dubplate.conf`, which creates `/run/dubplate/
   nudge-triggers` (only that — `/srv/music/uploads` is a separate,
   earlier step in deploy.sh that `mkdir -p`s and `chown`s it directly,
   since it lives on the persistent `/srv/music` volume, not tmpfs); and
   renders the six env values above from the regenerated
   `~/.local/secrets` — but does **not** start or enable serve or the
   nudge path (deploy.sh's own comment says so; that is this runbook's
   next step, deliberately kept separate). Under `--check`, deploy.sh
   also reports (informationally, non-fatal) whether
   `/run/dubplate/nudge-triggers` already exists on the box, so a stale
   or missing tmpfiles state is visible before applying anything.
5. **Enable and start serve.** By hand, on the box:

   ```
   systemctl daemon-reload
   systemctl enable --now dubplate-serve.service dubplate-import-nudge.path
   ```

   `dubplate-serve.service` binds `127.0.0.1` and `172.30.0.1` on port
   `8090`; check `systemctl status` shows `active (running)` with zero
   restarts after 30s.
6. **Doctor.** `dubplate doctor` (as `music`, or read the units'
   `EnvironmentFile`) must report green on: the Access certs endpoint +
   AUD probe, `UploadsDir` writable, `UploadsDir`/`InboxDir` same
   filesystem, ledger appendable, the trigger run-dir present,
   `PublicNavidromeURL` syntactically sane.
7. **End-to-end verify.** Two parts: on the box, against the listener
   directly, then through Access from the workstation.

   On the box first, confirm `mountBasePath` itself answers outside its
   prefix, isolated from Access, Cloudflare, and Navidrome entirely (a
   `curl` through the public hostname proves nothing here: a request to
   `music.907.life/api/config` with no `/upload` prefix matches the
   *bare* `music.907.life` ingress rule from step 3, so it reaches
   Navidrome's own `/api/` surface, not serve — a 401 or an interstitial
   from Navidrome, never evidence about serve's own routing):

   ```
   curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8090/api/config
   ```

   Expect `404` — serve itself answers 404 outside its `/upload` prefix,
   checked directly against the listener, bypassing Cloudflare and
   Navidrome so the result can only be serve's own.

   Then, through Access from the workstation: `curl
   https://music.907.life/upload/api/config` — expect Access's login
   interstitial (HTML, not JSON), proving the app is actually gating the
   route rather than 530ing straight through. Complete the `onetimepin`
   challenge in a real browser as a roster email and confirm
   `/upload/api/config` returns the generated JSON (collections, chunk
   MB, caps, accepted-format text) — that is the "JSON API verified end
   to end" acceptance bar this task exists to clear. Also confirm
   `/upload/api/stream/probe` responds with `text/event-stream` headers
   and at least one heartbeat (gate 1's named target, published by this
   step, not yet exercised by the SSE spike itself) — **it is POST-only**
   (`serve.go` registers `POST /api/stream/probe`; a bare `GET` 405s),
   so drive it explicitly:

   ```
   curl -N -X POST https://music.907.life/upload/api/stream/probe \
     -H "Cf-Access-Jwt-Assertion: <token from the browser session>"
   ```
8. **FileBrowser stays up.** Nothing about its compose service,
   `inbox.907.life`'s Access app, or its env vars changes in this step —
   U12b retires it later, once Pass B ships and Geoff has verified the
   real page on his own devices.
9. **Ledgers.** Update `docs/STATUS.md` (serve deployed and reachable
   through Access at `music.907.life/upload`, port `8090`, JSON API
   verified) and `docs/HISTORY.md` with this step's outcome; commit with
   the standard footer.

**Rollback**, if any step above needs to be undone: `systemctl stop
dubplate-serve.service`, then `systemctl disable --now
dubplate-import-nudge.path` (a rolled-back binary predating this pass has
no `nudge-consume` subcommand, so the path unit has nothing valid to
trigger); PUT the tunnel ingress config back to its pre-step-3 three-rule
form (drop the path-scoped rule, keep the bare `music.907.life` rule
above `inbox.907.life` and the catch-all); the tmpfiles fragment and
`/srv/music/uploads` can stay — an unused empty directory and an inert
tmpfiles rule cost nothing and removing them is optional. The Access
application is an estate fixture, not part of this rollback — it stays
regardless (no DNS record was ever created for this surface to roll
back).
