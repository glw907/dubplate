# Pass: production cleanup and polish (prod-polish)

**STATUS: PLANNED 2026-09-01, ahead of need; adversarially reviewed same day
(opus agent: 3 blockers, 9 majors, 9 minors — all folded, see Review
provenance). Executes after the uploader initiative closes (Pass B through
U13). Geoff's plan approval doubles as the box-write sanction for P4/P5's
named steps.**

**Goal:** Clear the carried production debris, settle the owed rewrite
post-mortem, and create the roster's Navidrome accounts, so the estate is
clean before the organizer (beets replacement) work starts.

**Engine consultation:** no engine asks (dubplate builds no cairn-cms
surface; enumeration run 2026-09-01).

**Scope:** BACKLOG items #1–#8 (`#project:prod-polish` plus the #8 watch
item) and whatever `/log-issue` has collected by pass start. Out of scope:
new features, admin API surface (the organizer initiative owns it), the
beets replacement, packaging, and the `FIREWALL_SSH_CIDRS` provision guard
(#10 — reconciliation needs a policy decision, not a drive-by edit). #9
(seed-album review) is Geoff's own item and rides along untracked.
**Accounts ruling (Geoff, 2026-09-01): roster only** — #7 closes on the
three roster contributors; later listeners are added as they ask.

**Budget:** ceiling 1M tokens, checkpoint every 4 tasks. P1–P4 are
implementer-dispatchable chains via the Agent tool (no workflow); P0, P5,
and P6 are conductor or box work — the workflow-mode question is settled
here. Implementer `general-purpose` pinned `model: sonnet`; reviewer
`diff-reviewer` (opus). Repo gate: `bash scripts/check.sh` (its `scripts/*`
glob shellchecks any new bash script; it does NOT touch Markdown, so it is
a stated no-op for P1/P6 — P6's prose gate is `prose-voice-reviewer`
instead). P3 uses the pings repo's own gate. Unit/script tasks invoke
`vps-conventions`; Go touches invoke `go-conventions`.

---

## P0: Pass-start triage (conductor, inline)

- [ ] Sweep `BACKLOG.md` for items logged since this plan was written; fold
      small fixes into the matching task's dispatch notes, split anything
      larger than a small fix to its own pass. State what was folded and
      what was split.
- [ ] #8: `npm view tus-js-client dist-tags` — record the 5.x verdict (stay
      pinned or schedule the upgrade) in the BACKLOG close line AND as a
      one-line note beside the pin's upgrade note in the web README
      (durable homes, not the session ledger). Close #8.
- [ ] Re-verify the roster's location and shape (Pass B may have replaced
      `config/contributors.yaml` with email-keyed identity); P4 reads the
      roster as it exists at pass start, path resolved here.
- [ ] Confirm no Pass B carry-forwards landed in STATUS that belong here.

## P1: Docs truth sweep (#3)

**Files:** `docs/STATUS.md`, `docs/bringup-runbook.md`,
`env/dubplate.env.template`.

- [ ] Fix the pull-vs-push wording in exactly two places:
      `docs/bringup-runbook.md` ("the pull check's ping:/alert: keys") and
      the STATUS carried item — all four checks are push, the box's
      navidrome-check timer drives the navidrome one. Do NOT grep-sweep
      archived plans/specs; "pull" is legitimate rclone language there.
- [ ] Fix the env-template contradiction: `env/dubplate.env.template` says
      the Navidrome admin credentials back a "native admin API" call while
      `internal/beets/beets.go` implements Subsonic `/rest/startScan`. The
      code is the truth; fix the template comment.
- [ ] Remove the retired STATUS carried item (the pull-framing line).
      Commit (docs-only; repo gate is a no-op here by design).

## P2: provision.sh diagnostic fix (#5)

**Files:** `scripts/provision.sh`.

- [ ] The defect is scoped to ONE call site: `render_firewall_rules` inlines
      `$(ssh_cidrs_json)` into jq's argv, so `exit` inside the substitution
      kills only the subshell and the diagnostic degrades to a bare jq
      error. `firewall_rule_present`'s plain assignment already propagates
      failure under `set -e` and is correct — leave it. Fix shape:
      `ssh_cidrs_json` returns nonzero instead of `exit`, and
      `render_firewall_rules` precomputes into a variable and fails with a
      named, actionable message before jq runs.
- [ ] HARD CONSTRAINT: do not modify the `FIREWALL_SSH_CIDRS` unset-guard
      eight lines above (it blocks fresh provisions while policy declines
      the allowlist — that contradiction is #10, a policy decision, not
      this task).
- [ ] Gate, commit.

## P3: pings orphaned check (#4, cross-repo)

**Repo:** `~/Projects/pings` (its own gate; this task never touches
dubplate). NOTE: pings `docs/STATUS.md` is stale (still lists `musicbox-*`
slugs) — do not trust it; dubplate STATUS records the live slugs are
`dubplate-*`.

- [ ] Identify the orphan first: `wrangler kv key list --binding PINGS_KV
      --remote`, report the full orphan set (nothing on record names the
      slug).
- [ ] PRIMARY FIX (repair, not force): the DO is addressed by the slug in
      the ping URL, and those capability URLs live in the age store as
      `DUBPLATE_PING_*_URL`. Recover the slug, re-`put` the missing
      `route:<name>` KV record, then run the NORMAL `remove-check` path so
      the DO's alarm is actually torn down. A blind `--force` that deletes
      KV around a missing route record would strand an armed, alerting DO
      with no address left — never ship that shape.
- [ ] Tool change (the durable #4 fix): `remove-check` gains an explicit
      `--route-slug <slug>` repair option that re-puts the route record
      and proceeds with normal teardown. Test covers the orphaned-record
      path. No blind force flag.
- [ ] Run `code-simplifier:code-simplifier` over the pings diff INSIDE this
      task (its push happens here, before the pass-end sweep).
- [ ] pings gate, commit, push. Remove the retired STATUS carried item
      (the pings-gap line) in the dubplate repo.

## P4: Roster Navidrome accounts (#7, box task, Geoff-sanctioned)

**Files:** `scripts/create-family-accounts.sh`, `scripts/deploy.sh` (rsync
manifest gains the script — authorized here; manifest edits land with the
file they ship), `docs/bringup-runbook.md` (one section).

**Rulings (Geoff, 2026-09-01):** roster only (three contributors; the admin
account already exists and is skipped). Automated creation, generated
passwords to a root-only box file; Geoff shares through his own channel and
deletes the file afterward (the runbook documents share-then-delete as
Geoff's step; the pass itself deletes nothing).

- [ ] **Spike FIRST, blocking:** Navidrome has no supported user-creation
      API; the working path is the reverse-engineered native API
      (`POST /auth/login` → JWT, `POST /api/user` with
      `x-nd-authorization: Bearer <token>`, token rotates per response).
      Verify against the live box (pinned `0.63.2`) read-only where
      possible; record the exact method, path, body, auth header, and
      version in the task report before the script is written. If the
      spike fails, P4 downgrades to a documented manual runbook procedure
      through the Navidrome UI and #7 closes that way.
- [ ] Script: roster path and admin credentials from the box env — 
      `DUBPLATE_CONTRIBUTORS_FILE`, `DUBPLATE_NAVIDROME_ADMIN_USER`/
      `_PASSWORD` in `/etc/dubplate/env`; `yq` is present on the box for
      the YAML read (the workstation gate cannot exercise that path — say
      so in the script header). Idempotent via list-then-create; a
      `--dry-run` mode prints the create/skip plan and is REQUIRED to run
      clean on the box before the real run (that dry run is the
      idempotency evidence). Passwords never in argv or logs: JSON bodies
      via stdin (`curl -d @-`), generated with real entropy, written only
      to `/root/dubplate-family-accounts.txt` mode 0600.
- [ ] Verification probe: Subsonic `/rest/ping` with the salted-token
      scheme (argv-safe by construction, same scheme as
      `internal/beets`) once per created account.
- [ ] Runbook section: when to run, dry-run first, where the file lands,
      share-then-delete as Geoff's documented final step.
- [ ] Box steps (quiet box): deploy the script via deploy.sh, dry-run,
      real run, probes. Remove the retired STATUS carried item (the
      accounts-owed line). Gate, commit.

## P5: Box tidy and verdicts (#2, #6, box task, Geoff-sanctioned)

- [ ] #2: record `systemctl --failed --no-legend` output in the report
      FIRST, then `systemctl reset-failed musicbox-ping-failure@<unit1>
      <unit2>` naming exactly the two stale units — never the bare
      command, which would sweep evidence of any genuinely failed unit.
      Remove the retired STATUS carried item.
- [ ] #6: the SELinux surface worth checking is the compose `:Z` private
      relabels (`/srv/music/state/navidrome`, `/srv/music/inbox`, and the
      filebrowser state dir if it still exists post-U12b) on paths the
      host dubplate binary also writes: `ls -Z` each, confirm the binary's
      writes still succeed post-relabel (the journal for the last import
      and backup runs is the evidence). "Fine as-is" = relabeled type
      present AND host writes green in the journal; anything else names
      the labeling fix as a new backlog item. `getenforce` alone proves
      nothing new (bring-up already recorded Enforcing as PASS).
- [ ] Update STATUS/BACKLOG, commit.

## P6: Go-rewrite post-mortem (#1, conductor-written, docs-only)

- [ ] Write the REWRITE pass's own HISTORY entry as its post-mortem plus a
      scorecard: spend against the approved ceiling (≈5.0M of 5.5M per
      STATUS), interaction points, what the gates caught, what the
      organizer initiative (which builds on this binary) would be wrong to
      rediscover. Sources: docs/HISTORY.md entries, the archived rewrite
      plan, the model-economy ledger. This is a distinct entry from
      prod-polish's own pass-end HISTORY entry (both get written; see
      checklist).
- [ ] Prose gate: `prose-voice-reviewer` over the entry (the repo gate
      does not touch Markdown).
- [ ] Close #1, commit.

---

## Pass-end checklist

- [ ] `code-simplifier:code-simplifier` over dubplate changed code (P2/P4;
      P3 ran its own inside the task, pre-push)
- [ ] Quality gate: `bash scripts/check.sh` green
- [ ] Review gate (proportionate): `diff-reviewer` already ran per task;
      one skeptical credential read over P4's script — named checks:
      password-generation entropy, the 0600 file's lifecycle, argv/proc/
      log leakage, admin-credential sourcing (dispatch a fresh opus
      reviewer with exactly that brief; `web-auth-security-reviewer`'s
      web-session brief does not fit a provisioning script).
- [ ] Ledgers: STATUS (≤60 lines, each task removed its own retired
      carried item — verify none remain; next starter prompt = organizer
      brainstorm), HISTORY gains PROD-POLISH'S OWN entry (P6's rewrite
      entry is separate), BACKLOG items closed with verdict lines,
      ROADMAP: prod-polish → Done.
- [ ] Archive this plan to `docs/archive/plans/`.
- [ ] Commit, push, score both budgets, prep the context clear.

## Review provenance

Adversarially reviewed at authoring (opus agent, 2026-09-01): B1 the
assumed Navidrome admin API does not exist as a supported surface → P4
spike-first with manual-runbook fallback, env-template contradiction to P1;
B2 deploy.sh never ships `scripts/` → delivery named and authorized; B3 a
blind `--force` would strand an armed alerting DO → repair-first via
ping-URL slug recovery, `--route-slug` replaces force. M1 orphan slug
unnamed → KV-list first, stale pings STATUS flagged; M2 bare reset-failed
sweeps evidence → record-then-reset named units; M3 wrong fix pattern, one
call site already correct → scoped to `render_firewall_rules`; M4 the
FIREWALL_SSH_CIDRS guard contradiction → out of scope as #10 with a hard
do-not-touch line; M5 roster ≠ listeners → Geoff ruled roster-only; M6
roster shape may change in Pass B → P0 re-verifies at pass start; M7 one
HISTORY entry conflated two passes → both entries named; M8 the gate is a
no-op for docs tasks → stated, prose gate added for P6; M9 untestable
idempotency + argv leaks → dry-run evidence, stdin bodies, /rest/ping
salted-token probe. Minors m1–m9 folded as written (scoped grep; per-task
carried-item removals; dispatchability stated; P4 reviewer brief replaced;
P3 simplifier moved in-task; #8 verdict homes named; SELinux check
repointed at the `:Z` relabels; file-fate contradiction reconciled; box
paths pre-extracted).
