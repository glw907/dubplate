# Rung 5d adversarial fold: dispositions

The merged record for the two fresh-context lenses over the whole rung 5 subsystem:
`2026-09-25-rung-5d-review-security.md` (SEC-5d, at `6e661a6`: 3 medium, 7 low, a polish tail) and
`2026-09-25-rung-5d-review-universality.md` (UNI-5d, at `5262504`: 2 high, 3 medium, 4 low, a polish
tail). Neither lens found a critical. Conductor rulings, 2026-10-05; the commit and the test that
fails without each fold are filled in as the fold tasks land. This record is the rung's final
do-not-re-litigate list.

The fold needs code in more than six places, so it runs as three tasks at file-disjoint seams
(plan Task 9's split rule): **9A** the deploy script, unit, template, and runbook; **9B**
`internal/serve` and the probe; **9C** `cmd/dubplate-server`'s doctor, ports, and `internal/notify`.
`internal/config` is never touched.

## Folded

| Id | Finding | Fold | Task |
| --- | --- | --- | --- |
| SEC-5d M1 | `deploy.sh` runs `npm ci` and the web build with the sourced workstation secret set | run the local builds before the `source`, or under a scrubbed environment; `npm ci --ignore-scripts` | 9A |
| SEC-5d M2 | the clean-tree refusal misses untracked and ignored build or rsync inputs outside `web/` | refuse any untracked or ignored file under `cmd internal web config compose.yaml go.mod go.sum` (excluding `web/node_modules`, `.svelte-kit`, `build`) | 9A |
| SEC-5d M3, UNI-5d M1 (doctor half) | doctor exits 0 on an auth store or roster directory serve refuses or cannot write | an `auth store` doctor row (leaf mode and owner in serve's words, subdirectories writable) and a roster-directory writability check | 9C |
| UNI-5d M1 (serve half) | the roster directory's writability surfaces only at the first web write | `serve` checks `filepath.Dir(roster)` is writable at start | 9C |
| UNI-5d M1 (runbook half) | nothing says the grant, the leaves, and the configured paths must agree | runbook: the unit's `ReadWritePaths`, step 3, and the `systemd-run` lines name `$DUBPLATE_STATE_DIR/auth` and the roster file's directory | 9A |
| SEC-5d L1 | mixed-separator identity spellings (`X-Forwarded_user`) survive the strip | match on the normalized key, deleting every key whose `_`-to-`-` canonical form is in the dash set; count the same way | 9B |
| SEC-5d L2 | the probe never sends the configured client-address header, and an all-trusted bind exits 0 having proven nothing | the probe request and route use the strip's own forwarded set; a trusted peer receives the identity count; the all-trusted case reads unproven | 9B |
| SEC-5d L3 (runbook half) | the rollback set misses the beets version, the unit files, and `/opt/dubplate`; STATUS still says "rollback by redeploy" | runbook rollback adds those copies and warns that a pre-`c179c99` `deploy.sh` applies by default; the conductor corrects the STATUS row | 9A |
| SEC-5d L4 (comment half) | the serve unit's comment claims a bridge bind the template does not set | correct the comment | 9A |
| SEC-5d L5 | the `unidecode install` sandbox grants all of `/var/lib/dubplate`; the beets install runs as the roster's uid | pre-create the `unidecode` leaf and grant it alone; disposition-table residual 2 names roster and epoch integrity beside secrets | 9A |
| SEC-5d L6 | `CheckProxyHeader` admits address-less forwarded names and `Content-Type` | refuse the rest of the forwarded family and `Content-Type` | 9B |
| SEC-5d L7, UNI-5d L4 | a serve port of 0, or below 1024 under the empty capability set, passes doctor | bound the port to 1024-65535 in the threshold's validator in `cmd/dubplate-server`, naming the capability | 9C |
| UNI-5d H1 (runbook) | a same-host proxy in direct mode keys every member on loopback | runbook: any same-host front is configured as trusted, listing `127.0.0.1` and `::1` | 9A |
| UNI-5d H1 (serve) | nothing signals the misconfiguration | `serve` logs one warning when, in direct mode, a loopback or private peer sends a group 2 header (the strip's own count) | 9B |
| UNI-5d H1 (doctor) | `client ip direct` reads green without its assumption | the row states it: the peer must be the browser | 9C |
| UNI-5d H2 | the install order ends at a server a stranger cannot reach or sign in to | runbook: a step 0 for the build (`make web`, `make build-linux`, Node 24) and the service user; a front and TLS step with a same-host Caddy or nginx example and a cloudflared example, each with the trusted-proxy rows; the order ends at `https://<host>/dubplate/login`, with members at `/dubplate/members` | 9A |
| UNI-5d M2 (auth half) | `SMTPAuthPlain` is forced, so a LOGIN-only provider cannot send | `mail.SMTPAuthAutoDiscover` (quote it from the pinned go-mail source) | 9C |
| UNI-5d M3 | a CDN-fronted origin cannot list its edge addresses | runbook: the supported answer is a local proxy that validates the CDN's published ranges, trusted alone; CDN-direct is unsupported | 9A |
| UNI-5d L1 (mitigation) | the token in `?t=` lands in a front's access log | runbook: drop the query string from the front's access log | 9A |
| UNI-5d L2 | the serve unit's `OnFailure=` depends on the estate's incumbent | runbook: a stranger deletes or replaces the line | 9A |
| UNI-5d L3 | doctor hides the non-secret reason `check` prints | the doctor row prints the reason when it carries no value; runbook step 2 says the public URL is a bare `https://` origin and secrets need 32 random bytes | 9C, 9A |
| Polish, folded | the Hetzner comment in `internal/notify/notify.go:20-21`; how to generate the HMAC secrets; `DUBPLATE_ADMIN_EMAIL` and `DUBPLATE_LIBRARY_ROOT` explained; nginx's `proxy_pass` forwarding a client's header; a dual-stack front dialing either loopback; no systemd-free path, stated in scope | as named | 9C (comment), 9A (runbook) |

## Refused, with reasons

- **SEC-5d L3, automating the rollback.** Having `--apply` write the `.prev` copies and stop the shadow
  timer is new deploy behavior for one attended sitting whose procedure already does both (plan Task
  10). The runbook carries the full set; the sitting reads it.
- **SEC-5d L4, a new serve port.** Task 0's read-only box check found nothing listening on 8080, so the
  collision the finding fears does not exist on this host.
- **UNI-5d M2, a loopback relay without auth.** Whether an unauthenticated local relay is a supported
  mail class is a product question. The runbook's Mail section states that authenticated submission
  over TLS is required (9A); the relay class is filed to `IDEAS.md`.
- **Polish refused:** `auth.Open`'s owner check beyond the 9C row; the bracketed IPv6 hop without a
  port (no front sends one); a link-local bind without a zone (the probe already fails it);
  `Fastly-Client-Ip` client-settable (a documented choice in the runbook's proxy section is enough,
  folded in 9A); a doctor `--mail` send verb (new surface, filed to `IDEAS.md`).

## Carried

- **UNI-5d L1, the token in the URL fragment.** Moving the token from `?t=` to `#t=` changes the link
  format the mail, the link page, and the consumption handler share. The runbook mitigation lands now;
  the fragment move is filed to `BACKLOG.md` for a pass that may change the link contract.
- **UNI-5d H2, a redirect from `/` to the login page.** `/` belongs to the streamer fallthrough; carried
  to rung 12.
- **SEC-5d polish, group 2 in trailers.** Nothing reads trailers before the streamer proxy; carried to
  rung 12.
- **UNI-5d L2, a generic failure-alert hook.** Carried to packaging.

## Landed

The fold ran as `wf_4cf6053e-e09`, every task accepted by its `diff-reviewer` with no fix round:
9A at `fb5411e` (`deploy.sh`, the serve unit, the runbook, disposition residual 2), 9B at `bac992e`
(`internal/serve` and the probe), and 9C at `07f6f82` (doctor, the port bound, mail). Each Go fold
carries a test watched failing before the fix, recorded in the run's mutation ledgers. The `deploy.sh`
folds were proved by `--check 192.0.2.1` runs from a scratch worktree under a scratch `HOME` holding
dummy secret values: before the fix, both npm steps saw all 44 sourced names and an untracked `zz.go`
compiled into the stamped binary; after it, npm sees `HOME PATH` only and both refusals fire. 9D
closes three runbook gaps the 9A review and the 9C report surfaced (the rollback copy set, step 7's
`unidecode install`, and a stale doctor claim); its commit is recorded in HISTORY.

Accepted consequences of the folds, recorded so a later pass does not re-litigate them:

- The widened clean-tree refusal also refuses ignored outputs a gate leaves behind
  (`web/rendered-captures/`, `__pycache__`), so `deploy.sh` runs from a fresh worktree; the runbook
  says so.
- The probe route now answers a peer relayed through a same-host trusted front with one identity
  count, where it used to 404; this is the fold's "a trusted peer receives the identity count".
- Once rung 12 configures a front exemption for a trusted address, doctor's probe of that address will
  count the exempted identity header it sent. Rung 12 owns that interaction alongside `FrontExemption`.
- `AT_EACCESS` in the auth-store row's writability check is quoted from its man page but cannot be
  proved by a test, since the test process always has equal real and effective uids.
