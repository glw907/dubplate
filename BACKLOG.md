# BACKLOG

> Individual work items. Managed by `/log-issue`. Initiatives in `ROADMAP.md`.
> Live-use findings during the Go development land here with
> `#project:prod-polish` so the cleanup pass collects them.

## Open

- [ ] #1 Write the Go-rewrite post-mortem (spend, gates, what the pass
  taught; STATUS has carried "post-mortem owed" since 2026-08-31)
  #infra #project:prod-polish
- [ ] #2 Clear the two stale `musicbox-ping-failure@` failed units on the box
  (`systemctl reset-failed`; cosmetic, pre-fix 404 era; box write needs
  sanction) #infra #project:prod-polish
- [ ] #3 Fix STATUS/runbook wording: musicbox-navidrome was called a "pull"
  check; KV shows all four checks are push (the box's navidrome-check timer
  drives it) #infra #project:prod-polish
- [ ] #4 pings: `remove-check` has no `--force` escape for a check whose
  route record is permanently missing (undeletable record; lives in the
  pings repo) #infra #project:prod-polish
- [ ] #5 provision.sh: `ssh_cidrs_json` exits only its substitution subshell
  on malformed input, so the diagnostic degrades to a bare jq error
  #onboarding #project:prod-polish
- [ ] #6 Verify and close the carried SELinux one-time check on the state
  dir at first deploy (carried from bring-up 2026-08-31; likely moot on the
  running box, needs a stated verdict) #infra #project:prod-polish
- [ ] #7 Create the Navidrome group accounts (owed; per-contributor intake
  is the uploader's job, but listeners still need logins) #onboarding
  #project:prod-polish
- [ ] #8 tus-js-client 5.0 watch item: 5.0.0-pre2 exists under the `next`
  dist-tag; re-check at the prod-polish pass and record the upgrade verdict
  #upload
- [ ] #10 provision.sh refuses to run (even `--check`) when
  `FIREWALL_SSH_CIDRS` is unset, but the SSH allowlist was DECLINED
  2026-08-31 (no stable workstation IP), so a fresh provision is impossible
  without setting a variable whose policy answer is "unrestricted".
  Reconcile with a stated decision; do not drive-by edit the guard
  #onboarding #project:prod-polish
- [ ] #9 Geoff: review the two seed albums parked in
  `/srv/music/review/2026-08-31/geoff/` (Aim "Hinterland (Remastered)",
  illuminati hotties "POWER", plus a TV Girl zine PDF) #pipeline
- [ ] #12 `dubplate-server member add --help` does not mark `--name` and
  `--username` as required, though `MarkFlagRequired` enforces both
  (`cmd/dubplate-server/member.go:100-107`); cobra adds no help annotation, so say
  "(required)" in each flag's usage string. Found by 5d-ii Task 7, 2026-10-05
  #onboarding #project:prod-polish
- [ ] #13 Move the magic-link token from the query (`?t=`) to the URL fragment
  (`#t=`), so a front's access log never records a redeemable token
  (`internal/auth/token.go:181`; the link page's client hook already strips the
  fragment). Changes the link contract the mail, the page, and consumption share.
  The runbook mitigation (drop the query from the access log) landed in 5d-ii.
  From UNI-5d L1, 2026-10-05 #security #project:prod-polish
- [ ] #14 Align `web/package.json`'s `engines` (`>=22.17`) with the Makefile's web
  target, which refuses a Node major below 24; the runbook follows the Makefile.
  From 5d-ii Task 9A, 2026-10-05 #infra #project:prod-polish
- [ ] #15 `serve`'s direct-mode warning (9B: a loopback or private peer sent a
  forwarded-address header) fires on `doctor`'s own boundary probe while serve
  runs, so a healthy doctor makes serve log a false alarm. The warning is
  `warnForwardedFromLocalPeer` (`internal/serve/serve.go:205`; its comment opens
  at `:187`). Have the probe's requests (`internal/serve/probe.go`) carry a marker
  the warning skips, or have the warning ignore the probe route. Found at the 5d-ii Task 10 sitting,
  2026-10-05 #bug #project:prod-polish
- [ ] #16 A username may equal the address of a member who has a username of their own. `principalCollision`
  (`internal/auth/roster.go:830`) already refuses a username equal to an address-principal, but compares
  principals only, so the two principals differ and an audit row can read as that address. Refuse the collision
  on write and at load. Found by the navidrome-removal close,
  2026-10-06 #security #project:prod-polish
- [ ] #17 The members page's `FIELD_FOR_OUTCOME` (`web/src/routes/members/+page.svelte:46`) has no
  `invalid_address` row, so that outcome sets no `aria-invalid` and moves no focus to the address field. Add the
  row. Found by the navidrome-removal close, 2026-10-06 #web #project:prod-polish
- [ ] #18 The rendered gate reaches the add-member username field only through `duplicate_username`
  (`web/scripts/rendered/run.ts`); the client-side pattern refusal is never exercised. Add a row. Found by the
  navidrome-removal close, 2026-10-06 #web #project:prod-polish
- [ ] #19 The add-member hint text uses `opacity-70` on `base-content`
  (`web/src/routes/members/+page.svelte:323`, and the roster line at `:252`), a contrast the theme gate does not
  measure. Measure it or use a token with a gated pair. Found by the navidrome-removal close, 2026-10-06
  #web #project:prod-polish
- [ ] #20 No rendered row asserts the autofill-ignore attributes (`autocomplete=off`, `data-1p-ignore`,
  `data-lpignore`) on the add-member Address and Username fields. Add one assertion. Found by the
  navidrome-removal close, 2026-10-06 #web #project:prod-polish

## Done

- [x] #11 Resolve the docker compose leg's `command -v docker` skip: retired with the leg
  itself (`d1e93e9` removed the compose leg from `scripts/check.sh`), 2026-10-05 #infra
