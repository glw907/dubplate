# Pass plan: one admin shell, one transport, and one network-failure alert

Initiative: dubplate-server
Round: cleanup (pass 4 of 5, under the conductor `conductor-dubplate`)
Pass: cleanup-web

Design: `/var/home/glw907/Projects/initiatives/dubplate-server/specs/2026-11-15-cleanup-round-design.md`, approved by
Geoff in the Approval Queue on 2026-11-16 (document `20261116-000308-88f4d3`). Read in full: "Scope", "How it fails",
pass 4's row under "Passes" (line 143), "What each merge changes", "Acceptance" (pass 4, and the two every-pass
items), "Standing answers" 13 to 20, "Rulings" (the network-failure alert, and the stacking ruling of document
`20261116-163931-082d9d`), and "Backlog verdicts" (#120, #19). The findings are
`docs/research/2026-11-10-whole-repo-review.md` lines 188-206 (W1 to W8), and its "Leave alone" list is lines 284-288.
The visual identity spec, `docs/specs/2026-09-24-visual-identity-design.md`, binds every task. Where this plan and the
design differ, the design wins and the difference is a stop.

Worktree: `/var/home/glw907/Projects/dubplate/.claude/worktrees/cleanup-web`, branch `cleanup-web`, cut from
`origin/cleanup-store` at `3fe031eb9a499a021520e25eb1946444c8447320`. That is PR #14, whose base is `cleanup-gate`
(PR #13, base `master`); both are open. **The PR's base is `cleanup-store`, not `master`** (D1). Class: `auth-data`
(the design's row). The PR never lands unattended; it waits for the round's close packet with passes 2, 3, and 5.
Review lens: one `standards:plan-reviewer` read of this plan (under "Review record"); a task-scope
`standards:diff-reviewer` read on each `auth-data` task (T1, the merged T2+T3 range, and T4); and at the close the
whole-branch `standards:diff-reviewer`, `standards:svelte-reviewer`, `standards:daisyui-a11y-reviewer`, and
`standards:web-auth-security-reviewer`, each given the identity spec's path. Executor: `dubplate-cleanup-web`,
`sonnet` at `medium` with the Fable advisor; T4 goes to one `opus` subagent. Design path minutes: 100. This plan's
path: 180 (see "Path"), under the 200-minute park line (standing answer 14).

## Pre-flight

No step needs Geoff. The pass adds no secret, no sign-in, no dialog, and no deploy, and it reaches no live account.

Permissions checked: read `~/.local/state/claude-denials.jsonl` on 2026-11-16 for the last 30 days. Fourteen entries
carry a dubplate cwd. Eleven are "Classifier unavailable" refusals, which are classifier outages and no rule. One
"[Unauthorized Access]" is a `curl` download of a libxml2 tarball. One "[Unauthorized Access]" and one "[Destructive
or Unwanted Actions]" are `sed -i` edits into another agent's worktree (`cmd/dubplate-server`,
`internal/engine/store`). No task here downloads anything, and every edit is made with the Edit tool inside this
session's own worktree or a subagent's own worktree. Every command the tasks run (`npm ci --prefix web`, `bash
scripts/gate.sh`, `make web-build web-rendered`, `grep`, `git`, `gh pr checks`, `magick`) has run in passes 1 to 3 or
in this planning session without a denial. `rm -rf web/rendered-captures` removes a gitignored output directory
(`.gitignore:47`) inside the session's worktree. No `.claude/settings*` rule is wanted.

Reviewed: one `standards:plan-reviewer` read; its findings and the fold are under "Review record".

Evidence: the mechanism table and the recorded searches below. Each row was read or run against `3fe031e` on
2026-11-16, by this planner or by the plan reviewer's probes.

## What the pass changes

The members page and the server-settings page render through one admin shell (the head title, the site header,
`main`, and the three gate states), load their list through one `loadList`, and show a failed write's sign-in-again
alert through one component. Both transports (`post.ts` for the auth forms, `v1.ts` for `/api/v1`) send through one
`fetch` site in `lib/api` and keep their own 403 rules. A network failure is `alert-warning` on the sign-in page, as
it already is on the admin pages, and the matrix captures it on both. No button pairs `btn-sm` with `min-h-11`, and
the harness asserts every captured button is at least 44px tall at 320 and 390px. The faded hint text and the 13px
labels move to the tokens the identity spec names (closing BACKLOG #19). `SENT_MOMENT_TEXT` and the review's state
nits in the files this pass edits go. A member sees the alert colour, the button sizes, and the hint text change, and
nothing else.

## Rules that bind every task

1. **Files.** Pass 4's row: `web/src/**`, and `web/scripts/rendered/**` only where an assertion names markup this pass
   changes (change the selector, never weaken what it asserts). Also these, each listed under `## Decisions`:
   `cmd/dubplate-server/rendered_test.go` (forced by the two new capture states, D4), the visual identity spec's
   touch-target bullet at lines 249-253 and its status line (the adapter's Errata rule, D5), `BACKLOG.md` (closing
   #19, rewording #120, filing finds), and this plan and its prompt (moved to `docs/archive/plans/` at the close).
   Anything else is filed through `standards:log-issue` in BACKLOG and the PR body (standing answer 15). Never touch
   `docs/STATUS.md` except to write a `## Stop` (the conductor owns it; wanted text goes in the PR body under
   `## STATUS text`), or `.claude/settings*` ("`.claude/settings*` is never an executor's file: record the rule the
   pass wants for the close packet and go on"). Never touch `web/src/lib/themes/system.css` (the review's "Leave
   alone" list). Server-only SvelteKit files (`+page.server.*`, `+server.*`, `hooks.server.*`, form actions) stay
   forbidden; the gate checks that none exists.
2. **The box.** The round adds no feature, no migration, and no config key, and it deploys nothing. No task reaches
   the box (`5.78.148.42`), by ssh or otherwise.
3. **Gates.** The fast lane is `bash scripts/gate.sh --scope-changed`, with no `--base` (D2). The close is `bash
   scripts/gate.sh --merge`. Each runs bare from the worktree root with the Bash tool's `timeout: 600000`, re-issued on
   exit 75 until it prints `gate exit:`. Never pipe a gate into `tail` or `head` (a hook blocks it). Never commit red.
   `RUN_GATE_IF_BUSY=defer` is never used on an `auth-data` task (T1, T3, T4).
4. **Commits.** One commit per task (a review fix may add one), the task id first in the subject, imperative mood,
   specific files only (never `-A`), ending with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. Never
   commit a secret or a `.env` value; read each diff by eye before staging.
5. **Pushes.** Only `git push origin cleanup-web`. The draft PR opens on T1's push: `gh pr create --draft --base
   cleanup-store`. After each push, watch CI as a background Bash task, bare: `gh pr checks <n> --watch --fail-fast`,
   with `gh run watch <run id> --exit-status` as the fallback. A red stops the line when it arrives. A red outside
   this pass's files is fixed in this branch and listed under `## Decisions` (standing answer 19).
6. **Captures.** The capture leg writes `web/rendered-captures/` and never clears it (mechanism 2), so run `rm -rf
   web/rendered-captures` before any run whose captures are copied, counted, or compared. The session's git dir is
   `/var/home/glw907/Projects/dubplate/.git/worktrees/cleanup-web` (written `<git dir>` below; a subagent uses this
   absolute path, since its own `<git dir>` differs). The base set from T0 lives at `<git dir>/base-captures/` and the
   post-T1 set at `<git dir>/t1-captures/`; neither is committed. Each copy is taken from a fast-lane run made after
   every mutation is reverted. The compare is mechanism 1, with no pixel tolerance. A file that differs only inside
   the B1 columns (x32-34 at 390, x432-434 at 1440, x113-115 on sheets) is listed under item 1's reading and kept. A
   difference that items 2 to 4 do not explain is fixed at its cause, never accepted as the new baseline (standing
   answer 18).
7. **Records.** PR notes at `<git dir>/pr-notes.md`, first line `launch-base=<sha>`. Each task appends its test-first
   failures, mutation results, review line, gate exits, capture compare result, and anything found outside scope.
8. **Expected values** are literals from this plan (44, 210, 105, 42, the state ids), never a figure the change
   itself printed.
9. **Mutation proofs are run.** `make web-rendered` reads the bundle `make web-build` last built and does not rebuild
   it (`Makefile:133-152`; the binary embeds `web/build`, `web/embed.go:20`). So a mutation in `web/src` runs as `make
   web-build web-rendered` from the worktree root; a mutation in `web/scripts/rendered` alone runs as `make
   web-rendered`. Apply each mutation alone, run, record the failing row ids, and revert. One that doesn't fail is a
   finding fixed in the same task.
10. **Smallest thing that meets acceptance.** No abstraction, export, prop, flag, or fallback the design doesn't name.
    Every new export names its production caller in the task that adds it.
11. **Stack idiom.** Load `standards:svelte-conventions`, `standards:ts-conventions`, and `standards:daisyui` before
    the first `web/` edit, and read the identity spec. A DaisyUI component class and its documented modifiers come
    before a utility stack; colours and sizes come only from tokens the spec names; `$derived` comes before an effect
    that mirrors state, and `$effect` is kept for a real side effect with cleanup. A 44px target comes from the
    control's own size class, never `btn-sm` with `min-h-11`. The exemplar is
    `~/Projects/cairn-cms/src/lib/admin-toolkit/`. No em dash in a code comment or commit.
12. **Copy.** No string in `web/src/lib/copy/` changes wording. BACKLOG #83's copy is not this pass's.

## Mechanisms

Each row was read at `3fe031e`.

| # | Mechanism | Source | What a task relies on |
| --- | --- | --- | --- |
| 1 | No committed compare tool exists. `magick compare -metric AE` reads a size change as identical when the extra rows match the background: the plan review measured `0 (0)` at exit 0 for a 10x10 against an 11x10 white image, and `20 (0.166667)` at exit 1 only when the extra rows differ. So the compare first reads `magick identify -format '%w %h' <file>` on both files and counts any mismatch as a difference. Then `magick compare -metric AE <base> <head> null:` prints `N (fraction)` on stderr and exits 0 when identical, 1 when they differ. For a differing file, `magick compare <base> <head> -compose src -highlight-color white -lowlight-color black <diff.png>` then `magick <diff.png> -trim info:` prints the box as `WxH+X+Y` (on an identical pair it warns and prints a 1x1 box at -1-1) | ImageMagick 7.1.2-35 at `/home/linuxbrew/.linuxbrew/bin/magick` | T0, T1, T2+T3, T4, T5. The script lives in `$CLAUDE_JOB_DIR/tmp`, never in the repo |
| 2 | The matrix is 19 states by 5 widths by 2 schemes: 190 cells in 95 files (per state, `390-light.png`, `390-dark.png`, `1440-light.png`, `1440-dark.png`, and `contact-sheet.png` holding the 320, 768, and 2560 cells), from 38 contexts. `captureRows` only runs `mkdir({recursive: true})`, so a stale file survives a run | `web/scripts/rendered/run.ts:5316-5331`, `6088-6152` | T0, T1: two new states make 21, 210 cells, 105 files, and 42 contexts; rule 6 clears the directory |
| 3 | The capture loop reports each cell's property as `unsettled === null && secondary.length === 0 && overflow === null && touch.count > 0 && touch.failures.length === 0`; `overflowReport` runs only at 320 and 390, and a `skipUIFloors` state (`error/not-found`) skips every floor. `touchTargetReport` skips rects with `display:none`, `visibility:hidden`, or zero size; a closed DaisyUI `dialog.modal` is `visibility: hidden` (DaisyUI 5.8.1 `modal.css`), so its buttons are excluded. `TOUCH_FLOOR_PX = 24` gates and `TOUCH_ADVISORY_PX = 44` only advises | `run.ts:5334-5335`, `5377`, `6126-6141` | T1: the 44px check joins this property |
| 4 | DaisyUI 5.8.1 `button.css`: `.btn{--size:calc(var(--size-field,.25rem) * 10);--btn-p:1rem;…height:var(--size);padding-inline:var(--btn-p)}` and `.btn-sm{--fontsize:.75rem;--btn-p:.75rem;--size:calc(var(--size-field,.25rem) * 8)}`. `--size-field: 0.28125rem` in every theme block, and `app.css:18` holds the root at 16px at 320 and 390, so a plain `.btn` is 45px, `.btn-sm` 36px, and `btn-sm min-h-11` 44px. The spec's bullet still tells a small button to carry `min-h-11` | `web/src/lib/themes/dubplate.css:32,47`; `web/src/app.css:18`; identity spec 249-253 | T1: plain `.btn` meets 44 with no `min-h-11`; the spec bullet is edited in place |
| 5 | `OUTCOME_CLASS` has no `network_failure` row, so `outcomeClass` falls to `error` and `outcomeAlertClass` gives `alert-error` on the sign-in page and the interstitial. Its doc says a warning is "something that failed on the server's own end". `v1.ts` `failureAlertClass` gives `warning` for a network failure, `error` for `cross_origin` | `web/src/lib/api/outcome.ts:13-41`; `web/src/lib/api/v1.ts:95-96` | T1 |
| 6 | The harness mirrors the class table in its own `outcomeClass` (it cannot import `outcome.ts`) and derives `wantModifier` from it. `LoginCase.outcome` is a `string`. `auth/login/network-failure` aborts the request and carries `outcome: 'error'`; `wantOutcomeText` picks `NETWORK_FAILURE` by the row's id | `run.ts:2514-2538`, `2651`, `2696`, `2705-2714`, `2774-2786` | T1: the test-first change |
| 7 | `renderedCaptureStates` lists the 19 states with per-scheme `linkRequests` and `consumptions`; the host asserts `n != 38` contexts. Count comments: `rendered_test.go:38-40`, `270`, `302-303`, `654-657` ("the twenty links its ten signed-in states spend … rather than twenty environment variables"), `972`, `975-976`, and `run.ts:5317`, `5536`, `6080-6082`. The parser-table rows at `rendered_test.go:836-840` hold literal "38" strings that test the parser and stay. The fixture's link budget is 200 (`rendered_test.go:62`); the matrix spends 26 today and 28 after T1 | `cmd/dubplate-server/rendered_test.go` | T1 |
| 8 | `members/empty` signs in once per scheme with `mintCaptureLink(endpoints, 'admin')` and `signIn`, sets `visit.expectsAuthErrorStatus`, routes `**/api/v1/list_members`, and waits for the `h1`. `members/unreachable` (a row, not a capture) aborts the first `list_members` and filters `.alert` by `NETWORK_FAILURE`. `loginAddressOutcomeArrange` fills the address, clicks submit, and reads `#outcome`'s text only. `AUTH_ERROR_STATUS_RE` (`run.ts:896`) admits `net::ERR_FAILED` | `run.ts:5563-5590`, `5729-5748`, `3115-3136` | T1: the two new states copy these |
| 9 | The eight `btn-sm` + `min-h-11` sites: members `:340`, `:347` (roster demote and remove), `:424`, `:425` (the reauth notice); settings `:299`, `:300` (the reauth notice), `:340`, `:344` (row buttons). Plain `.btn` with a redundant `min-h-11`: members `:442`, `:443` (the confirm modal), `SiteHeader.svelte:110`. Links keep theirs: `SiteHeader.svelte:104`, members `:316`. Roster and settings rows are `flex flex-wrap` | as cited | T1 |
| 10 | The sent moment's escape hatch is `<button type="button" class="link link-primary">` (login `:120`), the one `<button>` in `web/src` without `.btn`; three rows find it by role `button` and name `SENT_ESCAPE_HATCH` (`run.ts:47`, `2826`, `3097`). The `auth/login/sent` row requires its left edge level with `h1.page-title` and `paddingLeft === '0px'`, and its comment records that a `.btn`'s padding was removed from this control in an earlier review. `.btn-link` keeps `.btn`'s `padding-inline: 1rem`, so `btn btn-link` would fail that row | `web/src/routes/login/+page.svelte:120`; `run.ts:2818-2833` | T1 |
| 11 | `opacity-70` at members `:328` (the roster line), `:414` (the username hint), settings `:317`, `:365`; `text-[13px]` at settings `:313` (a `<label>`), `:361` (a `<span>`). The spec names `base-content` (`#1f1b16` light, `#efe4cc` dark) at lines 73 and 87, gated at 10.6:1 (line 97), and mono's three sizes at line 357 ("11px (badges, film edge), 12px (catalog line), 13px (tracklist numbers and durations)"). `web/src/lib/styles/components.css:529` sets `:where(.label)` to `base-content` | as cited | T1 |
| 12 | `postForm` fetches with a form body, maps a throw to `network_failure`, **every** 403 to `cross_origin`, and a non-JSON or outcome-less body to `failed`. `call` fetches with a JSON body, maps a throw to `network_failure`, and a 403 to `cross_origin` **only** when the body fails to parse; a JSON 403 is a refusal (`reauthentication_required`, `forbidden`). `auth/login/cross-origin` fulfills a JSON 403 `{}`; `settings/reauthenticate` fulfills a JSON 403 `reauthentication_required` and filters `#settings-outcome` by the reauth line | `web/src/lib/api/post.ts:51-76`; `v1.ts:43-66`; `run.ts:2695`, `5900-5925` | T2 |
| 13 | `outcomeClass` (`outcome.ts:29`) is read only at `outcome.ts:34,39`; `refusalText` (`v1.ts:69`) only at `v1.ts:79` | `grep -rn` over `web/src` | T2 |
| 14 | Both pages declare `type Status = 'loading' \| 'ok' \| 'refused' \| 'signedOut' \| 'unreachable'` and `let status = $state<Status>('loading')`, the same head-title expression over `TITLES.notFound`, `TITLES.signedOut`, and their own title, the same `refused`, `signedOut`, and `unreachable` branches (an `h1`, a paragraph or an `alert alert-warning`, and a button), and the same list loader over `call` (`list_members`/`members`, `list_settings`/`settings`). Members: `:76-178`, `:297-313`; settings: `:87-174`, `:269-285` | as cited | T4 |
| 15 | The two reauth notices differ: members renders `<div id="roster-outcome" class="alert {notice.alertClass}">` with `<span id="roster-outcome-text" tabindex="-1">`; settings renders `id="settings-outcome"` with a plain `<span>`, and its `ok` branch puts `{#if refreshFailed}<div class="alert alert-warning">{LIST_REFRESH_FAILED}</div>{/if}` before the notice. Members has `refreshFailed` too (`:96`, `:356`) in its own place. Rows select `#roster-outcome` (`run.ts:3596`, `3623`, `3645`, `3676-3678`, `4197`, `4296`, `4404-4416`); none selects `#roster-outcome-text` | members `:418-430`; settings `:293-305` | T4: the component passes ids through and renders each page's markup byte for byte |
| 16 | W6 sites: login `:33` `const siteName = $derived(siteNameFromMeta \|\| TITLES.login)`; login `:56` `$effect(() => { if (moment !== null) tick().then(() => momentHeading?.focus()); })`; `SiteHeader.svelte:46` `let dark = $state(currentlyDark())`, written at `:68` and read only at `:113` (`checked={dark}`; the checkbox's DOM state carries the toggle); members `:109-111` three `bind:this` refs used by `inputForField`, for fields with ids; members `:120` `$effect(() => { if (pending && confirmDialog && !confirmDialog.open) confirmDialog.showModal(); })`, which opens the dialog after the DOM flush, once `#confirm-question` (inside `{#if pending}`, the target of the dialog's `aria-labelledby`) exists. `run.ts:2821` waits for focus on `#outcome h2`; `members/confirm` (`run.ts:5857`) needs the dialog open. In runes mode, a plain `let` written in a handler and read in markup compiles with a `non_reactive_update` warning that `svelte-check` (no `--fail-on-warnings`, `web/package.json:12`) passes | as cited; Svelte 5.57 compile probe by the plan review | T3, T4 |
| 17 | `SENT_MOMENT_TEXT` (`deck.ts:84`) has no reader; `deck-coverage.ts:140` walks the `deck` namespace generically and names it nowhere | `grep -rn SENT_MOMENT_TEXT web` | T3 |
| 18 | `check.sh --scope-changed --dry-run` printed `scope=full fallback=".gitleaks.toml is machinery" base=7db541b…`: the merge base with `master` spans passes 2 and 3 (76 files), so every fast lane runs the full gate, about 190-225s on pavilion | `bash scripts/check.sh --scope-changed --dry-run` | D2 |
| 19 | `.github/workflows/check.yml` runs on `pull_request` with `permissions: contents: read`; `gh pr checks 14` shows `check pass 7m1s` on a stacked base | as cited | rule 5 |

## Recorded searches

Run at `3fe031e` from the worktree root.

| Search | Hits | Use |
| --- | --- | --- |
| `grep -rn 'btn-sm' web/src \| grep -c 'min-h-11'` | 8 | item 3's start |
| `grep -rn 'opacity-70' web/src \| wc -l` | 4 | item 4's start |
| `grep -rn 'text-\[13px\]' web/src \| wc -l` | 2 | item 4's start |
| `grep -rnE '\bfetch\s*\(' web/src` | 2: `lib/api/post.ts:55`, `lib/api/v1.ts:46` | item 5's start |
| `grep -rnE '\bfetch\b' web/src \| grep -vE '\bfetch\s*\('` | 4, all comments (`deck.ts:122`, `post.ts:36`, settings `:92`, members `:90`) | item 5: no alias today |
| `wc -l web/src/routes/members/+page.svelte web/src/routes/server-settings/+page.svelte` | 450, 371 | item 6's start |
| `grep -rn SENT_MOMENT_TEXT web/src \| wc -l` | 1 | item 7's start |
| `grep -rn '<button' web/src --include='*.svelte' \| grep -v 'class="[^"]*btn'` | 4: three multi-line `btn-sm` buttons (members `:338`, `:345`, settings `:342`) and the escape hatch (login `:120`) | T1: the escape hatch is the one non-`.btn` button |
| `grep -rn "'abort'" web/scripts/rendered/run.ts` | 4: the two mode types, `auth/login/network-failure`, `header/sign-out/network-failure` | T1: no interstitial row aborts |
| `grep -n 'roster-outcome-text' web/scripts/rendered/run.ts` | 0 | T4: no row guards the text id's focus |

## Tasks

Every task's check includes the fast lane at exit 0. The pair's two halves have disjoint files; the session merges
both, runs the fast lane, and pushes.

### T0: Setup and the base set (10 min, session)

Run `npm ci --prefix web` bare, then `rm -rf web/rendered-captures` and the full gate (`bash scripts/gate.sh`) on
`3fe031e` with no change. Copy `web/rendered-captures/` to `<git dir>/base-captures/` (`cp -a`). Check: the copy holds
19 state directories and 95 PNG files (`find <git dir>/base-captures -name '*.png' | wc -l` prints 95), and `gate
exit: 0`. Record the counts and the gate's elapsed seconds in the PR notes. No commit.

### T1: The network-failure alert, the button sizes, and the hint tokens (40 min + 5 review, `auth-data`, session)

Items 2, 3, and 4, and BACKLOG #19. One task, because all three edit `run.ts` and the same pages, and one test-first
run, one fast lane, and one review serve them.

Outcome, item 2: `OUTCOME_CLASS` gains `network_failure: 'warning'`, and its doc says a warning is a failure on the
way to the server or on its end, as opposed to a refusal. The sign-in page and the interstitial then show a network
failure as `alert-warning` (D6). Two capture states join `CAPTURE_STATES`. `login/network-failure` goes after
`login/malformed`: route `**/request-link` to `route.abort()`, run `loginAddressOutcomeArrange` with `ADMIN_ADDRESS`
and `NETWORK_FAILURE`, then read `#outcome`'s class and fail the arrangement unless it holds `alert-warning`.
`members/network-failure` goes after `members/empty`: sign in as `members/empty` does, abort
`**/api/v1/list_members`, go to `members`, and require one `.alert.alert-warning` whose text is `NETWORK_FAILURE`.
`renderedCaptureStates` gains `{id: "login/network-failure"}` and `{id: "members/network-failure", consumptions: 1}`
at the same positions, the context check becomes 42, and the count comments in mechanism 7 say 21 states, 210 cells
or manifest rows, 42 contexts, and twenty-two links for eleven signed-in states.

Outcome, item 3: the eight sites in mechanism 9 drop `btn-sm` and `min-h-11` and keep their other classes (`btn`,
`btn-primary`), and the three plain `.btn` sites drop their redundant `min-h-11`; links keep theirs. The escape hatch
becomes `class="link link-primary inline-flex min-h-11 items-center"`, the link pattern members `:316` and
`SiteHeader.svelte:104` already use (D7). The spec's bullet at lines 249-253 is edited in place to say that a plain
`.btn` measures 45px, that `.btn-sm` (36px) is not used where a touch target is wanted, and that no control pairs it
with `min-h-11`; the status line gains a revision 1.4 entry naming this pass (D5). The harness gains a button-height
check in the capture loop. At 320 and 390px, every visible `button` and `.btn` (the visibility test
`touchTargetReport` uses) must be at least 44px tall, by a named literal constant. Failures go in the row's detail,
and any failure makes the property false. A `skipUIFloors` state skips it, as it skips the other floors, and the
existing `touch.count > 0` term covers a page with no control.

Outcome, item 4: the four `opacity-70` classes go, so the text takes `--color-base-content` (spec lines 73, 87, 97).
The two `text-[13px]` classes become `text-xs`, the 12px catalog-line size (spec line 357, D8).

Files: `web/src/lib/api/outcome.ts`, `web/src/routes/members/+page.svelte`,
`web/src/routes/server-settings/+page.svelte`, `web/src/lib/components/SiteHeader.svelte`,
`web/src/routes/login/+page.svelte` (the escape hatch), `web/scripts/rendered/run.ts`,
`cmd/dubplate-server/rendered_test.go` (D4), `docs/specs/2026-09-24-visual-identity-design.md`, `BACKLOG.md` (#19
closed).

Test-first: make the harness and Go changes first: `run.ts:2696` to `outcome: 'network_failure'`, `network_failure`
in the mirror's warning branch (`run.ts:2525`), the two capture states, the Go entries and counts, and the height
check. Run `make web-rendered` on the unchanged bundle and record three things. `auth/login/network-failure` fails on
its class, since the page shows `alert-error`. The `login/network-failure` capture rows fail on the arrangement's
class check. The height check fails on the escape hatch in `login/sent` at 320 and 390, light and dark, and nowhere
else. That list is the work list: a height failure outside it is fixed by the control's own size class in `web/src`
and listed in the PR (A2).

Mutation, run after the fixes and reverted: give one roster button bare `btn btn-sm` and run `make web-build
web-rendered`. The `members/populated` 320 and 390 capture rows must fail on the height check, which proves the check
reaches a `.btn` as well as the bare `<button>` the test-first run caught.

Check: after the revert, `rm -rf web/rendered-captures` and the fast lane. `make web-rendered`'s output shows 210
capture rows; `find web/rendered-captures -name '*.png' | wc -l` prints 105; `grep -rn 'btn-sm' web/src | grep -c
'min-h-11'` prints 0; both item-4 greps print 0; the existing `overflowReport` stays clean at 320. Compare the 95
base-named files with `<git dir>/base-captures/`. Each differing file must lie in a state that shows a changed control
or token: `login/sent` (the escape hatch), the `members/*` states that render the roster, the add-member hint, or the
confirm modal, and the `settings/*` states. The PR names each file with the item it falls under. Copy
`web/rendered-captures/` to `<git dir>/t1-captures/`. The per-task `diff-reviewer` reads the commit.

### T2 and T3: a pair (35 min, with merge and review)

Each half is a `sonnet` subagent at `medium` (`general-purpose`, `isolation: "worktree"`), dispatched after T1's
commit as the adapter's "Dispatches" sets out. Each half runs `npm ci --prefix web`, its mutation with `make
web-build web-rendered`, and `npm run check --prefix web`, and commits. Neither runs the fast lane: the session runs
it once on the merge (D12). Both are cuts. After the merge, `rm -rf web/rendered-captures`, the fast lane, and the
compare: the 105 capture files are identical to `<git dir>/t1-captures/`, or differ only inside the B1 columns
(listed). One `diff-reviewer` reads the merged range (T3 is `auth-data`).

#### T2: One fetch site in `lib/api` (20 min, ordinary)

Item 5 and W8's two exports. Outcome: one new module in `lib/api` (for example `request.ts`) exports one function that
calls `fetch` once and answers the `Response`, or a marker for a thrown request, which `post.ts` and `v1.ts` (its two
production callers) map to `network_failure`. Each caller keeps its own body, header, and 403 rule (mechanism 12).
`outcomeClass` and `refusalText` lose `export`. `fetch` is never aliased or passed under another name.

Files: `web/src/lib/api/request.ts` (new), `web/src/lib/api/post.ts`, `web/src/lib/api/v1.ts`,
`web/src/lib/api/outcome.ts`.

Tests: the existing rows hold both 403 rules, the part of W2 the review says to keep. Mutations, each run alone and
reverted: (a) make `call` treat every 403 as `cross_origin`; `capture/settings/reauthenticate/*` must fail. (b) Make
`postForm` treat a JSON 403 as a refusal; `auth/login/cross-origin` must fail. The thrown-request mapping takes no
mutation: `auth/login/network-failure` and `members/unreachable` stay green on the merge and guard it.

Check: `grep -rnE '\bfetch\s*\(' web/src` prints one line, in `lib/api`; `grep -rnE '\bfetch\b' web/src | grep -vE
'\bfetch\s*\('` prints only comments; `grep -rn 'export function outcomeClass\|export function refusalText' web/src`
prints nothing.

#### T3: The dead constant and the login and header state nits (15 min, `auth-data`)

Item 7 and W6's login and header sites. Outcome: `SENT_MOMENT_TEXT` is deleted. Login's `siteName` is a plain
`const`. The moment heading's focus moves from the `$effect` into the handler that sets `moment`, after `await
tick()`, and the effect goes. `SiteHeader`'s `dark` becomes `const dark = currentlyDark();` and the write at `:68`
goes, since the checkbox's own DOM state carries the toggle (a plain `let` would compile with a `non_reactive_update`
warning, mechanism 16).

Files: `web/src/lib/copy/deck.ts`, `web/src/routes/login/+page.svelte`, `web/src/lib/components/SiteHeader.svelte`.

Mutation, run and reverted: drop the focus call from the handler; the `auth/login/sent` row, which waits for focus on
`#outcome h2` (`run.ts:2821`), must fail.

Check: `grep -rn SENT_MOMENT_TEXT web/src | wc -l` prints 0; `grep -c '\$effect' web/src/routes/login/+page.svelte`
prints 0; `grep -n 'const siteName = \$derived' web/src/routes/login/+page.svelte` prints nothing; `npm run check
--prefix web` prints no warning for `SiteHeader.svelte`.

### T4: One admin shell (40 min + 5 dispatch + 5 review, `auth-data`, one `opus` subagent)

Item 6 and W1, with W6's members sites. Dispatched after the pair's merge and push, as one `general-purpose` subagent
with `isolation: "worktree"` and `model: opus`, under the adapter's "Dispatches". Outcome:

- `loadList(operation, field)` in `lib/api/v1.ts` answers `{ ok: true, entries }` or `{ ok: false, state, message }`
  exactly as the two page loaders do now (mechanism 14); its production callers are the two pages.
- One component in `lib/components` (for example `AdminPage.svelte`) owns the `Status` type's gate states. It renders
  the head title, `SiteHeader`, `main#main`, and the `refused`, `signedOut`, and `unreachable` branches, and takes the
  page's title, its retry, its unreachable message, and a `children` snippet for `ok`. A route group
  (`(admin)/+layout.svelte`) is not used, since it would move the two route files item 6 names (D9).
- One component renders a write's notice with its sign-in-again link and retry. It takes the notice, the retry, and
  the ids each page uses today (mechanism 15), so `roster-outcome`, `roster-outcome-text` with its `tabindex="-1"`,
  and `settings-outcome` render as they do now, and settings' inline refresh alert stays where it is.
- Members drops the three `bind:this` refs for lookups by id. The `showModal` effect moves into `ask`, which becomes
  async: `pending = …; await tick(); confirmDialog?.showModal();`, so `#confirm-question` exists before focus lands
  (mechanism 16).

Files: `web/src/routes/members/+page.svelte`, `web/src/routes/server-settings/+page.svelte`,
`web/src/lib/api/v1.ts`, `web/src/lib/components/AdminPage.svelte` (new), one notice component (new).

Test-first: `auth-data`, a cut. The existing rows are its tests, and they are green at its start commit by the pair's
merged fast lane; the dispatch cites that run. Mutation, run alone and reverted: drop the `showModal` call from `ask`;
`capture/members/confirm/*` must fail. The gate branches and the notice ids take no mutation: the compare against
`t1-captures` catches any branch or notice that renders differently, and the text id's focus has no row (recorded
search; filed under "Planning misses" for a later pass).

Check, in the subagent: `wc -l` on both route files prints under 450 and under 371; `grep -rn "type Status" web/src`
prints one line; `grep -rn "call<.*list_" web/src` shows only `loadList`; `grep -rn 'roster-outcome-text' web/src`
shows the id passed by members with `tabindex="-1"` kept; `npm run check --prefix web` clean. Check, in the session
after the merge: `rm -rf web/rendered-captures`, the fast lane, and the 105 capture files identical to
`<git dir>/t1-captures/` or differing only inside the B1 columns (listed). The per-task `diff-reviewer` reads the
merged commit.

### T5: The close (40 min, session)

1. Compare all 105 captures from T4's fast lane (the final head's run). The 95 base-named files against
   `<git dir>/base-captures/` (mechanism 1): list every differing file with its bounding box and the item it falls
   under, or, for a box confined to the B1 columns, item 1's reading. Any other difference is fixed at its cause
   (standing answer 18). The ten new files (two states) are new, and the PR states the count: 210 captures, 105
   files.
2. Dispatch, in one message, the whole-branch `standards:diff-reviewer`, `standards:svelte-reviewer`,
   `standards:daisyui-a11y-reviewer`, and `standards:web-auth-security-reviewer`. Give each the design, this plan,
   the identity spec's path, and `git diff 3fe031e...HEAD`, and ask for correctness gaps ranked by consequence and
   conformance to the spec (tokens, inks, the component layer). Ask the Svelte reviewer to confirm that the status
   machine, the load, and the re-authentication alert each exist once (item 6).
3. While they run, render-read every user-path page at 320, 390, 768, 1440, and 2560px from the 105 files (the native
   PNGs and the contact sheets), light and dark, in the main loop. A page that reads wrong is a finding fixed in the
   fold.
4. Fold once, quoting each finding verbatim. Each reviewer that did not return accept then gets one delta read of the
   fold. A delta's Low findings are folded once or filed, and never start another round; a named reviewer short of
   accept after the delta is a stop.
5. `BACKLOG.md`: #19 is closed by T1; #120 is reworded to its residue (W4's three other route files and W5) and
   retagged for a later round (D10); file each find through `standards:log-issue`.
6. `git mv docs/plans/2026-11-15-cleanup-web.md docs/plans/2026-11-15-cleanup-web-prompt.txt docs/archive/plans/`.
7. `bash scripts/gate.sh --merge` on the final head, after steps 4 to 6. Rerun after any later edit. If the fold
   touched `web/src`, rerun step 1's compare on this run's captures (after `rm -rf web/rendered-captures`).
8. The PR body: the opening paragraph per `~/.claude/docs/voice/written-for-people.md`, `Round: cleanup`, `Pass:
   cleanup-web`, the `Review:` and `Calls:` lines, evidence for each acceptance item, `## Decisions`, `## Calls for
   Geoff`, `## Planning misses`, `## STATUS text`, the score, and the land test's result. Mark it ready once the
   head's check is green.

## Standing answers

| # | Stop | Answer |
| --- | --- | --- |
| A1 | A capture differs where items 2 to 4 do not reach | Fix it at the cause (standing answer 18). A difference confined to the B1 columns is listed under item 1's reading and kept. |
| A2 | T1's height check fails a button outside the escape hatch | Give it the control's own size class in `web/src` and list it in the PR. A button outside `web/src` is filed and left. |
| A3 | The pass adds rendered time near `renderedScriptTimeout` (`rendered_test.go:36-42`) | Record the wall time in the PR notes. Raising the bound is a forced edit in the same file, listed under `## Decisions`. |
| A4 | T4's shell cannot keep a capture identical outside the B1 columns | Fix the shell until it does; a cut never changes pixels. |
| A5 | CI red outside the pass's files | Fix in the branch, list under `## Decisions` (standing answer 19). |
| A6 | The fast lane prints `scope=full` | Expected on every run (D2). Record the line once and go on. |
| A7 | The run passes twice a task's estimate | One `## Decisions` line with the estimate and the actual, and the run goes on (the adapter's "Clock"). |

## Decisions for the executor

- D1: The PR's base is `cleanup-store`, stacked on PR #14 and #13 (the conductor note). The planner prompt said
  `master`; the conductor's advisor ruled for `cleanup-store`. CI runs on a stacked base (mechanism 19). If #14 lands
  while the pass runs, keep the base; the conductor retargets it.
- D2: The fast lane names no `--base`, per the planner prompt. The merge base with `master` holds pass 2's machinery,
  so every fast lane runs the full gate, about 4 minutes each; the path counts it.
- D3: The review's "Leave alone" list names "the web markup". The review's own W1 to W8 change web markup, and the
  design's acceptance requires them, so this plan reads the list as the markup outside those findings: no restyling
  beyond items 2 to 4 and no structural change beyond item 6.
- D4: `cmd/dubplate-server/rendered_test.go` is outside the row and pass 5's directory. Item 2's new captures force
  its state table, its context count, and its count comments; the edit is taken and listed (standing answer 15).
- D5: The identity spec's touch-target bullet (249-253) tells a small button to carry `min-h-11`, which item 3
  forbids. The adapter's Errata rule (`.claude/skills/dubplate-pass/SKILL.md:36`) names the visual identity spec a
  living reference "edited in place by the pass that changes what it says, as a task in that pass", so T1 edits it
  and adds a revision line.
- D6: The interstitial (`login/link`) reads the same class table, so its network failure turns `alert-warning` too.
  No row or capture aborts the consume request; Geoff's ruling is "Warning on every page".
- D7: The escape hatch becomes `link link-primary inline-flex min-h-11 items-center`. `btn btn-link` would bring
  `.btn`'s 16px inline padding and its 14px, weight-600 type, and fail the `auth/login/sent` row's zero-padding
  alignment check (mechanism 10), which rule 1 forbids weakening. The link pattern is the one members `:316` and
  `SiteHeader.svelte:104` already use, so the control keeps its look and gains a 44px hit area.
- D8: `text-[13px]` becomes `text-xs` (12px, the catalog line). The spec's 13px is the tracklist's numbers and
  durations, which a setting's key is not; there is no 13px token, and an arbitrary value is what item 4 removes.
- D9: The shell is a component, not a route group, because item 6 names `routes/members/+page.svelte` and
  `routes/server-settings/+page.svelte` and a group would move them.
- D10: W4's three other route files (login, the interstitial, the error page) and W5 (`components.css`'s unused rules,
  which `web/scripts/gates/g1.ts` documents as the one route to a status colour) are left. Neither is in the design's
  acceptance list, and the smallest-pass rule holds. BACKLOG #120 is reworded to that residue rather than closed.
- D11: The two transport vocabularies (`post.ts`'s outcome strings and `v1.ts`'s `ApiResult` kinds) stay. W2 names
  them; item 5 asks for one `fetch` site, and the network-failure pick is T1's.
- D12: Pair halves run their mutations and `svelte-check` but not the fast lane. Every fast lane is a full gate
  (D2), the two would serialize under `cairn-run-gate` with no `defer` (T3 is `auth-data`), and the session's run on
  the merge covers both halves before any push.

## Calls for Geoff

- Item 1's reading (B1). The design says "The 190 rendered captures of the close head are compared with the 190 of
  the pass's base commit, with no pixel tolerance." Pass 2's close (PR #13) found captures that differ between runs
  of the same tree on the alert rule's edge columns. Geoff's pass-2 decide document `20261116-080113-a56d02` picked
  "Spend 30 minutes first", and #13 then picked B1: each file matches the base with no tolerance, or differs only on
  those columns (x32-34 at 390, x432-434 at 1440, x113-115 on sheets), and the PR lists those files. The base set is
  pass 2's output, captured by T0 at `3fe031e` (pass 3's head). This pass uses that reading.
- The settings page's key labels go from 13px to 12px mono (D8).
- The sent moment's "use a different address" link keeps its look and gains a 44px hit area (D7).
- The identity spec's touch-target bullet is rewritten, with a revision 1.4 line (D5).

## Path

| Step | Minutes |
| --- | --- |
| T0 | 10 |
| T1, with its review | 45 |
| T2 and T3, merged, with its review | 35 |
| T4, with its dispatch and review | 50 |
| T5 | 40 |
| Total | 180 `[estimate]` |

Each figure counts its rendered runs at about 4 minutes with the bundle build: T0 one full gate; T1 a test-first run,
one mutation, and a fast lane; the pair two mutations in the longer half and one fast lane on the merge; T4 one
mutation and a fast lane; T5 the `--merge` gate, with the reviewers and the render read running side by side and CI
(about 7 minutes) running under the PR body. The design's figure is 100. This path is under the 200-minute park line,
so the pass runs; the PR records the difference under `## Decisions` with the actual.

## Acceptance

Copied verbatim from the design's pass 4 list, plus the two every-pass items.

| # | Item | Task | Evidence |
| --- | --- | --- | --- |
| 1 | The 190 rendered captures of the close head are compared with the 190 of the pass's base commit, with no pixel tolerance. The PR lists every capture that differs and why, and each difference falls under items 2 to 4. Today: the two sets are identical by definition. | T0, T1, T5 | the compare's per-file list with sizes and bounding boxes |
| 2 | A network failure shows the same alert style on the sign-in page and the admin pages: the style Geoff picks in the design packet, or `alert-warning` if his approval carries no pick. The pass adds a capture of that state on the sign-in page and on the members page, and the PR states the new capture count. Today: `alert-error` on sign-in, `alert-warning` on the admin pages, and no capture of either. | T1 | the two capture states, the test-first failures, "210 captures" |
| 3 | No control pairs `btn-sm` with `min-h-11`, and the rendered harness asserts that every button on every captured page is at least 44 px tall at 320 and 390 px. Today: eight sites pair them, and no such assertion exists. | T1 | the grep at 0, the check's test-first failures, the mutation |
| 4 | `grep -rn 'opacity-70' web/src \| wc -l` and `grep -rn 'text-\[13px\]' web/src \| wc -l` print 0. The PR names the token that replaced each value and the line of the visual identity spec that names it. Today: 4 and 2. | T1 | both greps; `base-content` (spec 73, 87, 97) and `text-xs` (spec 357) |
| 5 | `web/src` calls `fetch` at one site, in `lib/api`, under no other name. Today: 2 sites. | T2 | both `fetch` greps, the two mutations |
| 6 | `routes/members/+page.svelte` and `routes/server-settings/+page.svelte` are both shorter, and the Svelte reviewer confirms that the status machine, the load, and the re-authentication alert each exist once. Today: 450 and 371 lines. | T4, T5 | `wc -l`, the Svelte reviewer's line |
| 7 | `SENT_MOMENT_TEXT` is gone. Today: `lib/copy/deck.ts:84`. | T3 | the grep at 0 |
| 8 | The Svelte and accessibility reviewers and the security reviewer each return accept. | T5 | the three verdicts |
| 9 | The PR's check is green on its head. | T5 | `gh pr checks <n>` |
| 10 | `bash scripts/gate.sh --merge` exits 0 on the close head. | T5 | the `gate exit: 0` line |

Item 1's reading (B1, see "Calls for Geoff"): each capture file matches the base with no tolerance, or differs only on
the alert-rule edge columns (x32-34 at 390, x432-434 at 1440, x113-115 on sheets), and the PR lists those files. The
base set is pass 2's output.

## Planning misses

- `[stale-fact]` BACKLOG #19 cites members `:323` and `:252`; the two sites are now `:414` and `:328`. The finding
  stands; T1 closes it.
- No rendered row asserts focus on `#roster-outcome-text` after a field-less refusal (members `:210`). T4 keeps the id
  and its `tabindex` by grep, and the close files the missing row through `standards:log-issue`.

## Review record

One `standards:plan-reviewer` read (opus, medium) on 2026-11-16 of the first draft, at `3fe031e`. Verdict: not ready,
three blockers, five majors, fourteen minors. Folded:

- B-1 (`make web-rendered` never rebuilds the bundle, so `web/src` mutations ran against the old one): rule 9 and
  every `web/src` mutation now run `make web-build web-rendered`.
- B-2 (the honest path was over 200): the two session tasks merge into T1 (one test-first run, one fast lane, one
  review); T1's class mutation is dropped as a repeat of its test-first failure; the second height mutation, T2's
  thrown-request mutation, and T4's branch and id mutations are dropped, each with its guard named; pair halves skip
  the fast lane (D12); T5 runs the reviewers beside the render read and puts the edits before the merge gate. The
  recount is 180.
- B-3 (`btn btn-link` fails the sent row's zero-padding check): D7 and the call now use the existing link pattern.
- M-1 (a taller image compares identical): mechanism 1 checks dimensions first and quotes the measured results.
- M-2 (cut checks allowed no run-to-run drift): rule 6 and every cut check accept differences inside the B1 columns,
  listed.
- M-3 (T4's id mutation named no row): dropped; the gap is a planning miss filed at the close.
- M-4 (`showModal` before the label exists): `ask` awaits `tick()` first.
- M-5 (a mutation run could overwrite the saved set): copies come from a fast lane after every revert.
- Minors folded: the directory is cleared before counted runs (m-1); the login state checks its class firmly and the
  type note goes (m-2); the link-count comment joins mechanism 7 (m-3); T5's order (m-4); `const dark` (m-5); T3's
  greps (m-6); T2 mutation (c) is dropped, so m-7 lapses; subagent paths and setup (m-8); delta reads (m-9); D5 cites
  the adapter's Errata rule and adds a revision line (m-10, no fork: the rule answers it); the base set's commit
  (m-11); the stop carve-out (m-12); the citations (m-13); the check's edge states (m-14).

No finding changes the approved design. No second read runs: no blocker remains open after the fold.
