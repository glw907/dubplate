# Pass plan: PR checks, and planning documents that describe rounds

Initiative: dubplate-server
Round: cleanup (pass 1 of 5, under the conductor `conductor-dubplate`)
Pass: cleanup-ci

Design: `/var/home/glw907/Projects/initiatives/dubplate-server/specs/2026-11-15-cleanup-round-design.md`, approved by
Geoff in the Approval Queue on 2026-11-16 (document `20261116-000308-88f4d3`: "Use these eight rounds, in this order";
network alert "Warning on every page"). The round start was approved the same day (document `20261116-042624-28c2c2`).
Read in full: "Calls made here", "Scope", "How it fails", pass 1's row under "Passes", "What each merge changes",
"Acceptance" (pass 1 and the two every-pass items), "Standing answers" 13 to 20, "Backlog verdicts", and "The rounds
to 0.8". Where this plan and the design differ, the design wins and the difference is a stop.

Worktree: `/var/home/glw907/Projects/dubplate/.claude/worktrees/cleanup-ci`, branch `cleanup-ci`, cut by the conductor
from `origin/master` at `a5ad875`. The planner commits this plan and its prompt on `cleanup-ci` and pushes them before
the launch. PR base: `master`. Class: `runner` (`.github/`, `scripts/`, `.claude/`). Executor: `dubplate-cleanup-ci`,
`sonnet` at `medium` with the Fable advisor. Design path minutes: 100. This plan's path: 182 (see "Path"), under the
200-minute park line of `round.md` "Parking at planning". The margin is thin; the plan flags it and launches.

## Pre-flight

No step needs Geoff before the launch. The design's "Steps for Geoff" lists none. This pass needs no new secret,
sign-in, setting, hook, or permission rule. The workflow holds no secret. GitHub Actions on `glw907/dubplate` are on
with `allowed_actions: all` and default workflow permissions `read` (`gh api repos/glw907/dubplate/actions/permissions`
and `.../actions/permissions/workflow`, probed 2026-11-15). `origin` is `git@github.com:glw907/dubplate.git`, so a
push goes over SSH, the channel every earlier pass of this repo pushed through; the `gh` token (which carries the
`workflow` scope) serves only `gh pr create`, `gh pr edit`, `gh pr ready`, and the reads. No push of a file under
`.github/workflows/` has been exercised; if the first one is refused, that is a stop.

The T1 Playwright bump needs its matching Chromium in the user cache before any local gate (`run.ts:304-316` refuses
to launch without it): `npx --prefix web playwright install chromium`, run bare. It writes only
`~/.cache/ms-playwright`, which the gate's rendered leg reads; no denial row names it.

Permissions checked: read `~/.local/state/claude-denials.jsonl` on 2026-11-15 against every command in the tasks
below. It holds 75 rows, dated 2026-11-01T00:55 to 2026-11-16T04:50. Row 75 is this planning session's plan-reviewer
(2026-11-16T04:50, Unauthorized Access): a compound `mktemp -d; cd; curl` download of the libxml2 tarball that went on
to run its `configure`. Its answer: the executor reads every checksum without downloading anything (`gh api ...
--jq '.assets[]|[.name,.digest]|@tsv'` for GitHub releases; WebFetch of the published `.sha256sum` for libxml2), and
never runs downloaded code locally; the runner does the downloads. The other 74: Self-Modification 10, refused by the
server-side classifier with no class shown 10, Interfere With Workloads 9, "Classifier unavailable" 9, Destructive
or Unwanted Actions 6, Merge Without Review 3, Auto-Mode Bypass 3, two each of Irreversible Local Destruction, Git
Destructive, Create Public Surface, Unverifiable Deletion Scope, Sensitive-Source Provenance, Security Weaken, and
Undermining Oversight, and one each of Production Deploy, Secret-Store Writes, DNS / Domain / Cert Changes, Bot Check
Circumvention, Code from External, Shared Scratch Sweep, Safety Bypass Flag, Out-of-Place Publication, PII Data
Handling, and Modify Shared Resources. Eight rows touch this pass (row 75 and seven more), each with its answer. A `gh pr checks --watch`
(2026-11-01, Merge Without Review) was refused when chained to a merge; this pass runs it bare as a background Bash
task and merges nothing, with `gh run watch <id> --exit-status` as the fallback. A `gh workflow run` (Auto-Mode
Bypass): this pass never dispatches a workflow, since the check runs on `pull_request`. A python heredoc edit of a
`.claude/skills/.../SKILL.md` (2026-11-10, classifier): the adapter is edited with the Edit tool only. A `go run
actionlint@latest` (Code from External): this pass runs no unpinned tool, and actionlint is not used. `claude stop`
(Interfere With Workloads, several rows): the executor never types it; a stray job is reported to the conductor. An
`npm ci ... | tail` (Destructive or Unwanted Actions): every gate and install runs bare. A `git push origin HEAD:main`
and a BACKLOG commit pushed to the default branch (Modify Shared Resources): every push goes to `cleanup-ci` only, as
`git push origin cleanup-ci`. No row names `npm ci`, `bash scripts/gate.sh`, `go get`, `go mod tidy`, `npm outdated`,
`go list -m -u`, `gh api`, `gh pr create --draft`, `gh pr edit`, `gh pr ready`, `gh pr view`, `gh run view`,
`gh run list`, `playwright install`, `git merge`, `git mv`, or `git revert`. The workflow file is a new public surface in a public repo; the design's "What each merge changes" names it,
and the merge is the conductor's through the queue, so the executor's push to a branch publishes nothing new beyond
the branch itself.
Reviewed: one `standards:plan-reviewer` read (Opus, medium), recorded under "Review record" with each finding's fold.
Evidence: the mechanism table and the recorded searches below. Every row is a probe this planner ran on 2026-11-15
in this worktree at `a5ad875`, or a quoted source line.

Run `npm ci --prefix web` first. The baseline is `a5ad875`; `bash scripts/gate.sh` was green on master at that commit
(the conductor's STATUS commit, docs only, on top of a gated tree). A full local gate averages 326 s on this
workstation (`cairn-run-gate --records /var/home/glw907/Projects/dubplate master`: 6 runs, lock wait 241 s, which
falls when no other dubplate pass runs); a scoped gate over docs-only paths is about a minute, because `check.sh`
`classify_path` (`:362-400`) classes `docs/*`, `.claude/*`, `BACKLOG.md`, and `CLAUDE.md` as `noeffect`.

Who else writes `master` during the run: the conductor (STATUS commits), and other initiatives' sessions, which have
committed to BACKLOG on master before (`acc4089`). A conflict with master stops GitHub from running the check, so rule
4 reads `mergeable` before each watch.

## What the pass changes

A pull request against `master` gains a check that installs the gate's tools at their pinned versions on a hosted
runner and runs `bash scripts/check.sh`. Two throwaway commits prove the check can go red. The dependency sweep, the
round's only one, lands first. Then the planning documents stop describing a 17-rung ladder: BACKLOG's run order
becomes "Rounds to 0.8" with an "Old names" table, every `#rung:` tag becomes a `#round:` or `#pass:` tag, the backlog
verdicts are applied, and `CLAUDE.md`, `docs/standing-rules.md`, and the `dubplate-pass` adapter say "round" and
"pass" where they said "rung". The adapter takes the plan conventions both archived plans repeat, names
`claude-worktree-busy`, names a CI watch, and moves the full-suite home to CI. Four prompt files of landed passes move
to `docs/archive/plans/` unedited.

## Rules that bind every task

1. **Files.** Only pass 1's row: `.github/workflows/check.yml` (new), `.claude/skills/dubplate-pass/SKILL.md`,
   `CLAUDE.md`, `docs/standing-rules.md`, `BACKLOG.md`, the four prompt files (moved only), `go.mod`, `go.sum`,
   `web/package.json`, `web/package-lock.json`, `scripts/check.sh` (standing answer 13 only), T6's two throwaway files
   (changed and reverted, no net change), and this plan and its prompt (moved to `docs/archive/plans/` at the close). A call site forced by a changed signature goes under `## Decisions`; anything
   else is filed with `standards:log-issue` in BACKLOG and the PR body (standing answer 15). Never touch
   `docs/STATUS.md` (the conductor's; the wanted text goes in the PR body under `## STATUS text`), `.claude/settings*`
   (never an executor file; a wanted rule goes in the PR body for the close packet, design call 10), `IDEAS.md`,
   frozen specs, migrations, or anything under the design's "Out" list.
2. **Gates.** The close is `bash scripts/gate.sh --merge`. Every other gate is the fast lane, `bash scripts/gate.sh
   --scope-changed --base <start>`, where `<start>` is the full 40-hex id of the branch head when the task's edits
   began (`git rev-parse HEAD` before the first edit), and no other commit lands between it and the task's commit. A
   merge's gate takes the session head before the merge. This is the G14 rule T4 writes into the adapter: without a
   base, `go.mod` and `check.yml` in the merge-base diff promote every later run to the full form (mechanisms 9 and
   10). A T1 or T2 gate still runs the full form, since its own change set holds machinery or an unmatched path. Each
   gate runs bare from the worktree root with the Bash tool's `timeout: 600000`, re-issued on exit 75 until it prints
   `gate exit:`. Never pipe a gate. Never commit red, except the four T6 commits this plan names.
3. **Commits.** One commit per task, the task id first in the subject (`T3: Retag BACKLOG to rounds`), imperative
   mood, specific files only (`git add <path>...`, never `-A`), ending with the trailer the session's system reminder
   names. A pair half merges with `git merge --no-edit <branch>`. Never commit a secret or an `.env` value; read each
   diff by eye before staging.
4. **Pushes.** Only `git push origin cleanup-ci`. The draft PR opens on T2's push (`gh pr create --draft --base
   master`). After each push, read `gh pr view <n> --json mergeable`, re-reading every 30 seconds while it prints `UNKNOWN`
   (ten reads at most): on `CONFLICTING`, run `git fetch origin` and
   `git merge --no-edit origin/master`, resolve only in this pass's files (a conflict elsewhere is a stop), gate, and
   push again. Then wait until `gh run list --branch cleanup-ci --limit 1 --json headSha` shows the pushed head (if none does
   after five minutes, read `mergeable` once more; still no run is a stop), and
   start the CI watch as a background Bash task (`run_in_background`): `gh pr checks <n> --watch --fail-fast`, bare,
   with `gh run watch <run id> --exit-status` as the fallback. No task waits for CI before the next; a red stops the
   line when it arrives, except a T6 break run, which is meant to go red. A runner-side fetch failure (the
   subsonic.org XSD or the PyPI wheel in `make subsonic-xsd` and `make unidecode-wheel`) gets one `gh run rerun <id>
   --failed`; a second is standing answer 19, and the fix is one `actions/cache` step keyed on the two pins, listed in
   `## Decisions`. Never `gh workflow run`.
5. **Voice.** The adapter and `CLAUDE.md` text follow `~/.claude/docs/voice/agent-facing.md`; load `claude-infra`
   before editing either. The PR's opening paragraph follows `~/.claude/docs/voice/written-for-people.md`. BACKLOG and
   standing-rules text match their files' register. No em dash in a commit or a YAML comment.
6. **Records.** PR notes live at `<git dir>/pr-notes.md` (`git rev-parse --absolute-git-dir`), never committed. The
   first line is `launch-base=<sha>`, the branch head when the executor starts. Each task appends its gate exits, CI
   run links, and anything found outside scope. The PR body is built from them at the close.
7. **Expected values** in every acceptance command are literals from this plan or counts computed independently
   (a `grep -c` over the file, never a figure the edit itself produced).
8. **Smallest thing that meets acceptance.** No abstraction, flag, matrix, cache, or fallback beyond this plan.
9. **Downloads.** Nothing downloaded runs on the workstation. Checksums are read, never computed from a local
   download: `gh api repos/<owner>/<repo>/releases/tags/<tag> --jq '.assets[]|[.name,.digest]|@tsv'` for a GitHub
   release, and WebFetch of `https://download.gnome.org/sources/libxml2/2.15/libxml2-2.15.4.sha256sum` for libxml2.
   A wrong literal fails the CI run closed, which the next fix corrects.
10. **Reds the score counts.** `pass-score`'s `reds` counts local gate reds and CI reds on task commits. T6's two
    break runs are expected and not counted; a T2 fix run that goes red is counted.

## Mechanisms

| # | Mechanism | Source or probe | Result |
| --- | --- | --- | --- |
| 1 | No workflow exists | `ls .github` | "No such file or directory" |
| 2 | Actions are on, token read-only by default | `gh api repos/glw907/dubplate/actions/permissions{,/workflow}` | `enabled: true`, `allowed_actions: all`; `default_workflow_permissions: read` |
| 3 | The push channel | `git remote get-url origin`; `gh auth status` scopes | `git@github.com:glw907/dubplate.git` (SSH, as every earlier pass pushed); the token carries `workflow` and serves only `gh` calls |
| 4 | `check.sh` hard-requires six tools | `scripts/check.sh:218-241` | `go`, `golangci-lint`, `shellcheck`, `ruff`, `gitleaks`, `systemd-analyze`, each `command -v` then exit 1 |
| 5 | Node is required by the photo leg | `scripts/check.sh:762` | `command -v node` then exit 1 |
| 6 | Vale is required | `scripts/vale-comments.sh:35-37` | "vale not installed; install it, do not skip the gate"; styles are committed under `.vale/styles/Survey/` |
| 7 | Leg headers print as `== <leg> ==` | `scripts/check.sh:585-913` | the full form prints 10: make check (Go), shellcheck, ruff, gitleaks, photo pool integrity, vale, process citations, exported identifiers, systemd unit verification, state-dir agreement. make web, npm run build, and svelte-check print only in the scoped and reduced forms; in the full form they run inside `make check`, as does the rendered leg (`Makefile:149`), which has no header of its own |
| 8 | `gate.sh` wraps a workstation-only tool | `scripts/gate.sh` calls `cairn-run-gate` (`~/.local/bin`) | CI runs `bash scripts/check.sh` directly |
| 9 | An unmatched path promotes a scoped run | `check.sh --scope-changed --dry-run` with an untracked `.github/workflows/probe.yml` | `scope=full fallback=".github/workflows/probe.yml is unmatched"` |
| 10 | `go.mod` and `web/package*.json` are gate machinery | `check.sh` `classify_path` | a sweep commit promotes later scoped runs to full |
| 11 | xmllint is inside `go test` | `internal/harness/subsonic/xmlrule/xmllint.go`: `const pinnedLibxml = "21504"`; `newRunner` runs `exec.LookPath("xmllint")` and checks the version; `systemXmllint` (`sync.OnceValues`) is read by `schema_test.go:36` and `check.go:102` | a missing or wrong xmllint fails `go test`, so it is no separable leg |
| 12 | rclone is inside `go test` | `internal/backup/job_test.go:281` | `t.Fatalf` when `rclone` is absent |
| 13 | Two prerequisites need the network | `Makefile:170-180`; `docs/standing-rules.md:92-97` | `make unidecode-wheel` and `make subsonic-xsd` fill the user cache; the gate never fetches |
| 14 | The rendered leg needs Playwright's Chromium | `web/scripts/rendered/run.ts:307` launches `chromium` channel; hint `npx --prefix web playwright install chromium` | Playwright 1.64.0 (`web/package.json:21`) |
| 15 | The systemd leg filters one message only | `scripts/check.sh:883-912`; the plan-reviewer ran `systemd-analyze verify` in an `ubuntu:24.04` container (systemd 255.4) | any other output fails; on 255.4 the only output was the filtered "`/usr/local/bin/dubplate-server` is not executable" line, as on local 259, so standing answer 13 is not expected to fire |
| 16 | The full form reads no history and no `master` ref | `scripts/check.sh:511-526` (change set only in the scoped forms), `:740-755` (history scan only under `--merge`), `:699`, `:783-784` (`git ls-files`) | CI runs the full form at the default checkout depth, never `--merge` |
| 17 | Local tool versions | `rclone version`, `golangci-lint version`, `vale -v`, `gitleaks version`, `xmllint --version`, `systemctl --version`, `shellcheck -V`, `ruff -V`, `node -v`, `go version` | rclone v1.75.2 (Fedora's `/usr/bin/rclone`, outside the repo's control; newest release v1.76.0, held until Fedora ships it, per `docs/research/2026-10-10-dependency-sweep.md:96-101`), golangci-lint 2.15.0, vale 3.27.0, gitleaks 8.30.1, libxml 21504 (2.15.4), systemd 259, shellcheck 0.11.0, ruff 0.17.5, node v26.13.1 (`engines` `>=22.17`), go 1.27.4 (`go.mod` says `go 1.27.0`) |
| 18 | cairn-cms's runner pattern | `~/Projects/cairn-cms/.github/workflows/e2e.yml`, `test.yml` | `actions/setup-node` with node 24 and npm cache; `npx playwright install --with-deps chromium` under a step timeout; `setup-go` pinned `924ae3a1cded613372ab5595356fb5720e22ba16 # v6` |
| 19 | GitHub's hardening guide | `docs.github.com/en/actions/security-for-github-actions/security-guides/security-hardening-for-github-actions`, read 2026-11-15 | "Pin references to actions to a commit SHA in your workflow files."; "When selecting a SHA, you should verify it is from the action's repository and not a repository fork."; "It's good security practice to set the default permission for the `GITHUB_TOKEN` to read access only for repository contents."; "Avoid using the `pull_request_target` workflow trigger if it's not necessary."; untrusted input goes through "an intermediate environment variable" |
| 20 | `ci-green` doesn't fit | `~/.local/bin/ci-green` hardcodes `origin/main`; it needs `.github/ci-green.json` | outside the row; the watch is `gh pr checks` |
| 21 | Sweep surveys | `go list -m -u -json all` (direct only); `npm outdated --prefix web` | Go: kin-openapi v0.149.0→v0.153.0, go-mail v0.8.1→v0.8.2, `golang.org/x/exp` `v0.0.0-20260611194520-c48552f49976`→`v0.0.0-20261113113407-e3970619a1a4`, `golang.org/x/sys` v0.49.0→v0.50.0, modernc sqlite v1.60.1→v1.62.0 (`go list -m -u -f '{{if and (not .Indirect) .Update}}{{.Path}}{{end}}' all` prints these five). npm (`npm outdated --prefix web`, exit 1, run after `npm ci`): nine minors or patches (adapter-static 4.0.0→4.0.1, kit 3.0.1→3.2.0, @tailwindcss/vite 4.3.3→4.3.5, daisyui 5.7.47→5.8.1, playwright 1.64.0→1.65.1, svelte 5.57.2→5.57.3, svelte-check 4.7.6→4.7.8, tailwindcss 4.3.3→4.3.5, vite 8.3.4→8.3.8) and two held majors (typescript 6.0.3→7.0.2, @types/node 24.19.2→26.6.6) |
| 22 | `tools/venv/requirements.txt` is tracked but not in the row | `git ls-files tools/venv`; design pass 1 row | a bump there is filed, not taken |
| 23 | A libxml2 bump touches the goldens | `docs/standing-rules.md:44-47`, `xmllint.go:15` | outside the row; filed, not taken |
| 24 | No commit hook gates commits | `git config core.hooksPath` empty; `.git/hooks` holds samples only; `~/.claude/settings.json` hooks carry no commit gate | the T6 throwaway commits need no bypass flag |
| 25 | `pass-task-clock` is gone | `grep -rl pass-task-clock ~/.claude/skills/pass-core ~/.local/bin` | no match; the clock is `pass-core` `executor.md` "Score" (first execution commit to PR-ready) |
| 26 | The archived conventions | `docs/archive/plans/2026-10-11-admin-api-1.md:87-186`, `2026-10-12-intake-1.md:102-207` | 18 shared headings (G15); `admin-api-1` adds "No migration" |
| 27 | tidy-check diffs the index | `Makefile:43-45` (`git diff --exit-code -- go.mod go.sum`); `scripts/check.sh:23-25` | unstaged `go.mod` or `go.sum` edits fail the gate; stage them before a sweep's gate |
| 28 | Playwright's pin lives in two places | `web/package.json:21`; `docs/standing-rules.md:109` ("a Playwright 1.64.0 script") | a Playwright bump edits both; the second is in T5 |
| 29 | The sweep record | `dependency-upgrade` steps 2 and 6 want a dated record file and a CHANGELOG entry; the repo keeps `docs/research/<date>-dependency-sweep.md` and has no CHANGELOG | both are outside the row; the PR body is this sweep's record, and held majors go in the PR body and BACKLOG |

## Recorded searches

| Search | Hits on `a5ad875` |
| --- | --- |
| `awk '/^## Done/{exit} /#rung:/' BACKLOG.md \| wc -l` | 40 |
| `grep -c '#rung:' BACKLOG.md` | 42 (2 under "Done", lines 612 and 614, which keep their words) |
| `grep -c 'rung' CLAUDE.md docs/STATUS.md .claude/skills/dubplate-pass/SKILL.md` | 8, 2, 3 (STATUS is the conductor's) |
| `grep -c rung docs/standing-rules.md` | 9 (`grep -ci`: 11) |
| `grep -c '^## Path to 0.8' BACKLOG.md` | 1 |
| `grep -rn review-desk docs/plans --exclude='2026-11-15-cleanup-ci*' \| wc -l` | 3 (this plan names the word too, so it is excluded until T7 moves it) |
| `grep -c 'Up to three' .claude/skills/dubplate-pass/SKILL.md` | 1 |
| `grep -c 'worktree-busy' .claude/skills/dubplate-pass/SKILL.md` | 0 |
| `wc -c CLAUDE.md` | 13705 |
| `grep -rn 'Path to 0.8\|#rung:' ~/Projects --include='*.md' --include='*.txt' -l` outside dubplate's BACKLOG, IDEAS, archive, and frozen specs | the adapter (`SKILL.md:40-41`), and the round's planner prompt (`initiatives/dubplate-server/plans/2026-11-15-cleanup-1-planner-prompt.txt:19`, plus its copy in the initiatives worktree `cleanup-record`), which quote the old name as the thing renamed |
| `grep -n 'keyed' cmd/dubplate-server/doctor.go` | `:632` "every refusal validateMagicLinkSource can return is keyed today", still false: `validateSMTPPort` returns plain `fmt.Errorf` (`auth.go:418,421`) |
| `sed -n 224,232p internal/identity/identity.go` | `Valid`'s doc says "a non-empty source" and the code checks `id.Source != ""`: #26's identity comment item is now accurate (stale finding) |

## Tasks

The path is T1, T2, then T3 ∥ T4 as a pair while the session runs T5 and any T2 fix, then the pair's merge, T6,
and T7.

### T1: The dependency sweep (35 min, `runner`)

Outcome: every direct Go module and every `web/` package is at its newest minor or patch, or held with a reason, and
the PR body carries each survey's output and what moved.

Files: `go.mod`, `go.sum`, `web/package.json`, `web/package-lock.json`.

Load `standards:dependency-upgrade` and follow it, with one change (mechanism 29): the PR body is this sweep's
record, in place of the dated file and the CHANGELOG entry, and each held major goes in the PR body and BACKLOG (via
`standards:log-issue`) with its unblock condition. A BACKLOG item T1 files carries `#pass:cleanup-ci` or the
`#round:` tag of the round that takes it, never `#rung:`, so T3's retag table and acceptance 4 still hold. Run `go list -m -u all` and `npm outdated --prefix web` and save
both outputs for the PR body. Survey the pinned binaries too: the newest releases of rclone, golangci-lint, shellcheck,
ruff, gitleaks, vale, Playwright, and libxml2, and the skill's outdated command against `tools/venv/requirements.txt`.

Take every minor and patch in `go.mod` and `web/` (mechanism 21: five Go modules, nine npm packages). A `v0` minor
(kin-openapi 0.149 to 0.153) is a minor under the house rule; read its changelog per the skill, as for kit 3.0 to
3.2, daisyui 5.7 to 5.8, and Playwright 1.64 to 1.65. The majors (typescript 7, @types/node 26) wait for Geoff: take
nothing, record each. The Playwright bump takes its browser: after `npm ci --prefix web` on the new lock, run `npx
--prefix web playwright install chromium`, bare, before the gate. The other binaries are the workstation's, installed
through its tiers outside this repo, so CI pins the version the local gate runs (mechanism 17), and a newer release of
golangci-lint, shellcheck, ruff, gitleaks, vale, or rclone (v1.76.0, held until Fedora ships it) goes in the PR body
under "Found, not done". A newer libxml2 or any `tools/venv` bump is filed (mechanisms 22 and 23). New golangci-lint
findings in Go files after a lint bump are outside the row: hold the lint bump and file it.

Stage before the gate: `git add go.mod go.sum web/package.json web/package-lock.json` (mechanism 27). Then `bash
scripts/gate.sh --scope-changed --base <start>` (the full form, since the change set holds machinery). A bump that
turns the gate red, in the theme, bundle, rendered, or any other leg, is held: revert that one package, file it with
the red line and the version, and gate again. It is never fixed in this pass, since its fix sits outside the row.

Check: `go list -m -u -f '{{if and (not .Indirect) .Update}}{{.Path}}{{end}}' all` prints nothing, or only modules
the PR body records as held; `npm outdated --prefix web` lists only typescript, @types/node, and any package the PR
body records as held; the gate exits 0. Test: the gate.

### T2: The PR check (30 min, `runner`)

Outcome: `.github/workflows/check.yml` runs `bash scripts/check.sh` on every `pull_request` push, with each tool at the
version the local gate runs, and the draft PR is open with its first run started. The run's green is required before
T6, never before the pair: the CI wait sits off the path, and a red is fixed while the pair runs.

Files: `.github/workflows/check.yml`; `scripts/check.sh` only under standing answer 13.

The workflow is one job on `ubuntu-latest` (`ubuntu-24.04` if `latest` moves under it mid-pass, standing answer 19),
with these properties:

- `on: pull_request` only, default activity types. No `push`, no `workflow_dispatch`, never `pull_request_target`.
- Top-level `permissions: contents: read`. No secret, no `secrets.*` or `github.event.*` expression in any `run:`
  line. No `concurrency` key: T6 needs a break run and its revert run to both finish.
- `timeout-minutes: 45` on the job.
- Every `uses:` pinned by a full 40-hex commit id with the tag in a trailing comment (`# v6.0.0`). Resolve each id at
  execution with `gh api repos/<owner>/<action>/git/ref/tags/<tag>` (dereference an annotated tag through
  `git/tags/<sha>`) from the action's own repository, never a fork, never copied from cairn. Actions: `actions/checkout`
  (`persist-credentials: false`, default depth), `actions/setup-go` (`go-version: 1.27.4`, the module cache on),
  `actions/setup-node` (`node-version: 26.13.1`, `cache: npm`, `cache-dependency-path: web/package-lock.json`). No
  other action, except the one cache step rule 4 allows after a second fetch failure.
- Each tool not on the image at the pinned version is installed into `$RUNNER_TEMP/tools/bin` (outside the checkout,
  because the gitleaks leg scans the tree) and added to `$GITHUB_PATH`. Each download is the project's official
  release asset, checked with `sha256sum -c` against a SHA-256 written as a literal in the workflow, read per rule 9
  (the GitHub release asset's `digest` field; shellcheck publishes no checksums file, and the field covers it):
  golangci-lint 2.15.0, shellcheck 0.11.0, ruff 0.17.5, gitleaks 8.30.1, vale 3.27.0, rclone v1.75.2. libxml2 2.15.4
  is built on the runner from `https://download.gnome.org/sources/libxml2/2.15/libxml2-2.15.4.tar.xz`
  (SHA-256 `98087fd181d9070724f3fbc65c7377db03038eb92bd882374daff44940138821`, read by the plan-reviewer from the
  published `.sha256sum`; the executor re-reads it per rule 9), without Python, into `$RUNNER_TEMP/libxml2`, with an
  rpath or `LD_LIBRARY_PATH` so `xmllint` loads `libxml2.so.16`, and its `bin` on the path. Whether the tarball ships
  a generated `configure` or needs meson is unprobed; the first run settles it. A step asserts `xmllint --version`
  prints `21504`. Pins live in one `env:` block at the top of the job, so the adapter's sweep line can name the
  workflow as their home.
- Then: `npm ci --prefix web`; `npx --prefix web playwright install --with-deps chromium` (step `timeout-minutes:
  10`); `make unidecode-wheel`; `make subsonic-xsd`; `bash scripts/check.sh`.
- `systemd-analyze` comes from the image (mechanism 15).

Before the push, `git add .github/workflows/check.yml` and run the fast lane (the full form, mechanism 9). Commit,
push, open the draft PR, and start the CI watch per rule 4. Then go straight to the pair. A red run is fixed in the
workflow (standing answer 19 if the cause is the runner image), one commit per fix with subject `T2: ...`, each gated
by the fast lane (the full form again), committed between the session's T5 commits, never inside one (rule 2's base
rule); each fix and its run link go in the PR notes. One fix cycle is priced (see "Path").

Standing answer 13, if a leg can't run: first try installing its tool. If that fails, `check.sh` gains one input, the
environment variable `DUBPLATE_CHECK_OMIT`, holding a comma-separated list of leg names from a fixed set the script
spells out (`systemd` is the expected one). An unknown name exits 1. An omitted leg prints `== <leg> == omitted by
DUBPLATE_CHECK_OMIT` in place of running. The workflow sets it, the job's `name:` says which leg is missing (`check
(without systemd unit verification)`), the local gate stays that leg's home, and the PR flags acceptance item 1 as
amended. Never a `command -v` skip. Three failures are not legs, because they sit inside `make check` with no header
of their own: xmllint and rclone (inside `go test`, mechanisms 11 and 12) and the rendered leg (`Makefile:149`,
mechanism 7). Each is fixed in its install step (`--with-deps` for Chromium's system libraries); if the advisor finds
no install that works, it is a stop.

Check: the PR's `check` run is green; `gh run view <id> --log | grep -c '== .* ==$'` prints 10 and the ten names
match mechanism 7's full-form list; every `uses:` line of `grep -n 'uses:' .github/workflows/check.yml` carries a
40-hex id; `grep -c 'pull_request_target' .github/workflows/check.yml` prints 0. Test: the run.

### T3 ∥ T4 with T5: the pair, and the session's own task beside it (pair 35 and 30 min, T5 25 min, `runner`)

T3 and T4 have disjoint files, and T5's files are disjoint from both. The session commits T2, records `sha=$(git
rev-parse HEAD)`, appends `dispatch T3+T4 sha=<40-hex>` to the PR notes, and dispatches both as `general-purpose`
subagents with `isolation: "worktree"` on `sonnet` at `medium`. Each gets its task's full text, "Rules that bind every
task", the "Conventions table" below, and the design's path. Each subagent's first act is `git merge-base
--is-ancestor <dispatch sha> HEAD`; on a failure it runs `git merge --ff-only <dispatch sha>` and checks again, and
a second failure is a stop; then `npm ci --prefix web`, the edits, the fast lane with
`--base <dispatch sha>` (a docs-only change set, about a minute), a commit, and no push. While they run, the session
does T5 in its own worktree, then any T2 fix. When both report, the session merges each with `git merge --no-edit
<branch>`, runs the fast lane with `--base <its head before the first merge>`, and pushes.

#### T3: BACKLOG describes rounds (35 min)

Outcome: acceptance items 4, 6, and 11 hold.

Files: `BACKLOG.md`.

1. The header blockquote (lines 3-4) says the round order is "Rounds to 0.8" below.
2. Replace `## Path to 0.8` (lines 6-80) with `## Rounds to 0.8`: one sentence that rounds run one at a time in this
   order and a later design may regroup its passes; the landed list (`foundation`, `corpus`, `matcher`, `shadow`,
   `auth`, `store`, `tag-io`, `streamer-1` to `streamer-3`, `admin-api-1`, `intake-1`); the eight rounds as a numbered
   list, each with what it gains and takes, copied from the design's "The rounds to 0.8" table; the dependency sentence
   (`catalog` needs `intake`'s accepted state; `playback`, `queue`, and `library` need the catalog; `queue` comes before
   streamer slice 7; `engine` depends on nothing after `cleanup`); and that each round's carried work is the items
   tagged `#round:<name>`, and each pass's the items tagged `#pass:<name>`.
3. Under it, `### Old names`: frozen documents keep the old rung numbers and pass ids, and this table maps each to its
   round. Columns: old rung number, old name, landed passes (old ids), round now. Rows: 1 to 6 and 8 landed (round
   "landed"); 7 `intake` → `intake` (`intake-1` landed); 9 `pipeline`, 10 `migration`, 14 `admin-api` → `catalog`
   (`admin-api-1` landed); 11 `builtins` and 13 `queues` → `queue`; 12 `streamer` → `playback` for slices 3 to 6 and
   `queue` for slices 7 and 8 (`streamer-1` to `streamer-3` landed); 15 `integration`, 16 `cutover`, 17 `closeout` →
   `launch`; `library-web-1` → `library`. Then the pass renames: `gate-1` → `cleanup-ci` and `cleanup-gate`;
   `web-shell-1` → `cleanup-web`; `cleanup-engine` → round `engine`.
4. Retag every item above `## Done` per the "Retag table" below.
5. Apply the backlog verdicts per the "Verdict edits" below.

Check: `awk '/^## Done/{exit} /#rung:/' BACKLOG.md | wc -l` prints 0; `grep -c '#rung:' BACKLOG.md` prints 2;
`grep -c '^## Path to 0.8' BACKLOG.md` prints 0; `grep -c '^## Rounds to 0.8' BACKLOG.md` prints 1; `grep -c '^###
Old names' BACKLOG.md` prints 1; `grep -c 'swept 2026-11-15' BACKLOG.md` prints 23 (the keep list under "Verdict edits");
`grep -c '^- \[ \] #78 \|^- \[ \] #30 \|^- \[ \] #33 \|^- \[ \] #35 ' BACKLOG.md` prints 0; `git diff a5ad875 --
IDEAS.md` is empty. Test: the fast lane.

#### T4: The adapter, and the prompt archive (30 min)

Outcome: acceptance items 5 (the adapter's count), 7, and 9 hold.

Files: `.claude/skills/dubplate-pass/SKILL.md`; the four prompt files.

Load `claude-infra` first and read `~/.claude/docs/voice/agent-facing.md`. Edit with the Edit tool only.

1. **One executor per tree** names `claude-worktree-busy <path>` as the first check, before the existing `claude
   agents --json`, `pgrep -f` bracket pattern, and `git status` steps.
2. **Concurrent passes** loses "Up to three dubplate passes run at once" and its ruling citation. New text: passes
   run inside a round, one at a time unless the round design pairs them; each design names its files so two passes
   never edit the same one; with two PRs open, the second to merge first merges master and reruns its close gate.
3. **Plans are just-in-time** says "plan each pass against a tree no other executor is writing", with no "rung".
4. **Carried items** says a round's design reads BACKLOG's items tagged `#round:<its round>` and the "Rounds to 0.8"
   section, and a pass's plan reads the items tagged `#pass:<its pass>`.
5. **Dependency sweep** says the sweep runs once a round, as the first task of the round's first pass (finding G14),
   over `go.mod`, `web/package.json`, `tools/venv/requirements.txt`, and the pinned binaries, whose pins live in
   `.github/workflows/check.yml`'s `env:` block (rclone, golangci-lint, shellcheck, ruff, gitleaks, vale, libxml2) and
   `web/package.json` (Playwright). Keep the minors-land, majors-wait sentence and the standing-rules pointer.
6. **Fast lane** keeps the scoped command and adopts G14's base rule: a pass whose first task moves `go.mod`,
   `go.sum`, or `web/package*.json` passes `--base <task start>`, the 40-hex id of the commit the task started from,
   so later scoped runs exclude the sweep; any other pass passes no `--base`. Keep the full-form fallbacks.
7. **Full-suite home** says CI: the PR check (`.github/workflows/check.yml`) runs `bash scripts/check.sh` on every
   push to a PR. The close still runs `bash scripts/gate.sh --merge` locally, for the branch-history gitleaks scan
   and any leg the check leaves out under `DUBPLATE_CHECK_OMIT`; keep the harness `go vet -tags oracle` line.
   `RUN_GATE_IF_BUSY=defer` is allowed except on a leg the check leaves out and on an `auth-data` task.
8. **CI watch** says: after each push, `gh pr checks <n> --watch --fail-fast` as a background Bash task, run bare;
   the fallback is `gh run watch <run id> --exit-status`.
9. Add `## Conventions every task follows`, after "Rules the pass carries", holding the 18 headings of the
   "Conventions table" below with the reconciled text it gives. A plan cites this section instead of restating it
   (G15).
10. `git mv` the four prompt files, unedited, to `docs/archive/plans/`:
    `docs/plans/2026-10-11-admin-api-1-planning-prompt.txt`, `docs/plans/2026-10-11-admin-api-1-prompt.txt`,
    `docs/plans/2026-10-12-intake-1-planning-prompt.txt`, `docs/plans/2026-10-12-intake-1-prompt.txt` (G19).

Check: `grep -c 'rung' .claude/skills/dubplate-pass/SKILL.md` prints 0; `grep -c 'Up to three'` prints 0; `grep -c
'worktree-busy'` prints at least 1; `grep -n 'CI watch' SKILL.md` names `gh pr checks`; `grep -c 'once a round'`
prints at least 1; each of the 18 headings appears exactly once inside the new section alone, counted over `awk
'/^## Conventions every task follows/{on=1;next} /^## /{on=0} on' .claude/skills/dubplate-pass/SKILL.md` (the
unscoped file already holds "Gate", "Dispatches", and "Mutation proofs are run", so a whole-file count is vacuous);
`git diff -M100% --stat a5ad875 -- docs/plans docs/archive/plans` shows four renames at 100% similarity; `grep -rn
review-desk docs/plans --exclude='2026-11-15-cleanup-ci*' | wc -l` prints 0 (the unexcluded count is acceptance 9's,
read after T7 step 2 moves this plan). Test: the fast lane.

### T5: `CLAUDE.md` and the standing rules (25 min, `runner`, the session's own, beside the pair)

Outcome: acceptance items 5 (the `CLAUDE.md` count), 8, and 10 hold.

Files: `CLAUDE.md`, `docs/standing-rules.md`.

Load `claude-infra`; `CLAUDE.md` follows `agent-facing.md`. Edits to `CLAUDE.md`, each removing only a ladder fact:

- Lines 9-12: the active initiative is the greenfield build, run as rounds, with BACKLOG's "Rounds to 0.8" holding
  the order; drop "on a 17-rung ladder (each rung has a name, ...)".
- "Build for 0.8" (lines 45-48): "The streamer is the built-in OpenSubsonic streamer"; "Plans weigh their surface";
  "built by the pass that first consumes it".
- "Ledgers" (lines 62-63, 67): BACKLOG items are tagged `#round:<name>` or `#pass:<name>` for the round or pass that
  takes each, with the round order under "Rounds to 0.8"; the initiative is "the greenfield build, in rounds, followed
  by packaging".
- "Build process" (line 71): drop the five-milestone ladder; BACKLOG's "Rounds to 0.8" holds the order.
- "Web" (line 111): "The login page was the first page built under this rule."
- "Facts agents get wrong" (line 189): "in each pass's own file".

Edits to `docs/standing-rules.md`, each naming who holds the duty now:

- "Build for 0.8" (lines 8-24): the streamer is the built-in streamer, slices 1 to 8; every other pass plan weighs
  its surface; "a 0.8 forward hook is named in the plan of the pass that first consumes it and built there". Keep
  "a rung 6 schema" as "the store schema" (a landed fact) or drop the clause.
- "Current dependencies" (lines 33-43): the sweep runs once a round, as the round's first pass's first task; a new
  capability is ruled take-now or filed to the round that first leans on it; "a plan that adopts a library names its
  newest version"; the pins live in `.github/workflows/check.yml` and `web/package.json`.
- "Gate legs" (#70): line 108 "Since rung 5c-iii Task 3a" becomes "Since 2026-09-26" or the plain fact without the
  id (and "peak at 3a" loses "at 3a"); line 112 "Rung 6b-i's merge adds" becomes "The `store` work added"; line
  117's parenthesis stays, because it describes what the citation leg matches (`check.sh:781` carries the pattern
  `rung [0-9]`), so "rung" there names a comment form the leg rejects, never a unit of work. Lines 118-120: the
  process-citation leg gates in round `engine` (per #125); pass `cleanup-gate` moves the exported-unused leg to `go
  tool deadcode`, and it stays a warning (design call 14 and pass 2's acceptance item 8: "The leg stays a warning, as
  it is today"). No "comment-cleanup task" and no stale leg time remain.
- Line 109: if T1 took Playwright 1.65.1, "a Playwright 1.64.0 script" names the new version (mechanism 28).
- Add one sentence where the gate's home is described: the PR check (`.github/workflows/check.yml`) runs `bash
  scripts/check.sh` on every pull request, CI is the full-suite home, and the close's `bash scripts/gate.sh --merge`
  runs locally.

Check: `grep -c 'rung' CLAUDE.md` prints 0; `grep -ci 'rung' docs/standing-rules.md` prints 1 (line 117's citation
pattern; today it prints 11); `wc -c < CLAUDE.md`
prints 13705 or less; `grep -c 'check.yml' docs/standing-rules.md` prints at least 1; `grep -n 'exported-unused\|
deadcode' docs/standing-rules.md` says the leg stays a warning. Test: the fast lane with `--base <T5 start>`.

### T6: The check can fail (10 min active, `runner`)

Outcome: acceptance item 2 holds, and the branch tree is unchanged by it.

Files: one Go test file and one `web/src` TypeScript file, each changed and reverted; no net change.

1. Record `pre=$(git rev-parse HEAD^{tree})`.
2. Commit `T6: Throwaway: break a Go test` (change one expected literal in an existing test, for example in
   `internal/reject`), skip the local gate (the one place this plan allows a red commit), push, and wait until `gh run
   list --branch cleanup-ci --limit 1` shows the new run `in_progress`.
3. Commit `T6: Revert the throwaway Go break` with `git revert --no-edit HEAD`, push.
4. Repeat for `T6: Throwaway: break a web/src type` (assign a number to a `string`-typed binding in a `.ts` file
   `svelte-check` reads) and its revert.
5. Check `git rev-parse HEAD^{tree}` equals `$pre`.
6. Wait for the two break runs to end red and the last revert run to end green. Record the two red run URLs and the
   failing line of each log (the Go test name; the `svelte-check` error) in the PR notes.

Check: both break runs' conclusion is `failure` (`gh run view <id> --json conclusion`), each log names the broken
test or type, the tree matches, and the head's run is green. Test: the runs.

### T7: Close (62 min)

Outcome: the PR is ready with every acceptance item's evidence, the review's verdict in the body, and the check green
on its head.

Files: `BACKLOG.md` (the done moves), this plan and its prompt (moved).

1. Move #97 and #70 to `## Done`, checked, citing this PR; #108 stays open with its new text. Commit `T7: ...`.
2. `git mv docs/plans/2026-11-15-cleanup-ci.md docs/archive/plans/` and the prompt beside it. Run the
   retired-form search of "Recorded searches" again; a hit in a file this pass owns is fixed, and a hit elsewhere goes
   in the PR body for the conductor. Read acceptance 9 now: `grep -rn review-desk docs/plans | wc -l` prints 0.
3. `bash scripts/gate.sh --merge` on the close head, bare.
4. One `standards:diff-reviewer` at whole-branch scope (Opus, medium) against the design and this plan, asking for
   correctness gaps ranked by consequence, and naming acceptance items 5, 6, 8, and 10's reviewer checks. Per
   `round-run.md`: on `fix`, one batched fold quoting each finding verbatim, the gate again, then one second read by
   a fresh reviewer. Low findings from the second read are folded once or filed through `standards:log-issue`. A
   verdict still short of `accept` after the second read makes the PR ready with that verdict and its open findings
   named in `Review:`; it is never a stop. On `escalate`, the advisor decides fold or stop.
5. The PR body, per `pass-core` `executor.md` "Score" and `round-run.md` "Evidence": the opening paragraph; `Round:
   cleanup`, `Pass: cleanup-ci`, `Review: <verdict>, open findings <n>`, `Calls: none` or `<n>`; `## Acceptance` with
   every item below verbatim and its evidence; the legs table (CI log headers beside a local full run's headers); the
   two red run links; the sweep outputs and the held majors (mechanism 29); the retag table with old and new tags;
   `## STATUS text`; `## Decisions` (this plan's calls and each new one); `## Planning misses` with miss tags (#26's
   identity item `[stale-fact]`); `## Review` with the verdict in backticks; the score table and `<!-- pass-score
   {...} -->`; the land test of `approval-packet` step 1 and which way it falls; the settings rules for the close
   packet (none known); "Found, not done".
6. Push, wait for the head's check to go green, then `gh pr ready`. End at PR-ready.

Check: every acceptance item has its evidence in the body; `Review:` names the last read's verdict and its open
finding count; the head's check is green. Test: the gate and the review.

## Path

| Step | Minutes |
| --- | --- |
| T1 | 35 |
| T2, to the first push and the watch | 30 |
| The pair (T3 at 35, T4 at 30) beside T5 (25) and one T2 fix (about 11), plus the merge's gate | 40 |
| T6 | 10 |
| T7: done moves 5, `--merge` 7, review 15, fold 10, gate 3, body 10, last CI wait about 12 | 62 |
| Launch checks and the first watch | 5 |
| Path | 182 |

CI waits overlap later tasks and sit off the path, except the close's last green. The margin under `round.md`'s
200-minute park line is 18 minutes. A second T2 fix cycle beyond the priced one, or a second review read, eats most
of it; the clock rule (`running-passes` "Clock") records the overrun under `## Decisions` and the run goes on.

## Acceptance

Two items apply to every pass: the PR's check is green on its head, and `bash scripts/gate.sh --merge` exits 0 on the
close head.

1. The pass's PR shows a check that ran `scripts/check.sh` and passed, and the run's log prints the name of every leg
   it ran. The PR body lists those legs beside the legs of a local full run, and names any leg left out under
   standing answer 13. Today: `ls .github` prints "No such file or directory", and `gh pr list --state open` is empty.
2. The check can fail. The PR body links two runs that ended red, each on a throwaway commit that is reverted before
   the PR is ready: one breaks a Go test and one breaks a type in `web/src`. Today: no run exists.
3. The workflow file pins every action by commit id, declares `permissions: contents: read`, and triggers on
   `pull_request` only. Today: no file.
4. No line above BACKLOG's `## Done` heading carries `#rung:`. Today: 40 lines do. The two entries under "Done" (lines
   612 and 614) keep their words, as do the four source notes in `IDEAS.md` that say "rung".
5. `grep -c 'rung' CLAUDE.md docs/STATUS.md .claude/skills/dubplate-pass/SKILL.md` prints 0 for each, and the closing
   reviewer confirms that none of the three describes a ladder under another word. Today: 8, 3, and 3. The conductor
   makes the STATUS edit.
6. BACKLOG has a "Rounds to 0.8" section holding the eight rounds of this design in order, an "Old names" table, and
   no "Path to 0.8" heading. The PR body carries a table of every retagged item with its old and new tags, and the
   reviewer checks it against the mapping under "The rounds to 0.8". An item with two tags has both mapped. Today:
   `grep -c '^## Path to 0.8' BACKLOG.md` prints 1.
7. The adapter no longer says "Up to three dubplate passes run at once", names `claude-worktree-busy` in its
   one-executor rule, names a CI watch command, and says the dependency sweep runs once a round. It also holds every
   heading under "Conventions every task follows" that appears in both `docs/archive/plans/2026-10-11-admin-api-1.md`
   and `docs/archive/plans/2026-10-12-intake-1.md` (finding G15 counts 18), with its text. Today: `grep -c 'Up to
   three' SKILL.md` prints 1, `grep -c 'worktree-busy'` prints 0, and the CI watch line reads "None".
8. `docs/standing-rules.md` names the round or the pass that now holds each duty it gave a rung, and its text on the
   full-suite home names CI. The reviewer reads each changed line, since a blanket rename would pass a count. Today:
   `grep -c rung docs/standing-rules.md` prints 9.
9. `grep -rn review-desk docs/plans | wc -l` prints 0, because the four prompt files of the landed `admin-api-1` and
   `intake-1` passes moved to `docs/archive/plans/` with `git mv`, unedited. Today: 3 lines, in two of those files.
10. `wc -c CLAUDE.md` prints 13,705 or less, and the reviewer confirms that every fact the file lost was about the
    ladder. Today: 13,705.
11. The backlog verdicts under "Backlog verdicts" are applied. A folded item carries the `#pass:` tag of the pass that
    takes it. A duplicate is removed and the surviving item names its number. A reworded item has the new text. Each
    `keep` item is stamped `*(swept 2026-11-15)*`.
12. The dependency sweep ran: the PR body carries the output of each "what is outdated" command the
    `dependency-upgrade` skill names, and lists what moved or says nothing did.

| Item | Task | Evidence |
| --- | --- | --- |
| Check green on head | T7 | `gh pr checks <n>` on the ready head |
| `--merge` exits 0 | T7 | the `gate exit: 0` line of the close run |
| 1 | T2, T7 | the run link; the log's `== <leg> ==` lines beside a local `bash scripts/gate.sh` run's |
| 2 | T6 | two red run URLs with their failing lines; the tree-equality output |
| 3 | T2 | `grep -n 'uses:\|permissions\|contents:\|^on:\|pull_request' .github/workflows/check.yml` |
| 4 | T3 | the `awk` count (0); `grep -c '#rung:' BACKLOG.md` (2); `git diff a5ad875 -- IDEAS.md` (empty) |
| 5 | T4, T5 | `grep -c 'rung'` over the two files (0, 0); `git show origin/master:docs/STATUS.md \| grep -c rung` (0) after the conductor's edit, read at T7 (the branch never edits STATUS, and it prints 2 today, not the item's 3); the reviewer's line |
| 6 | T3 | the two heading counts; the retag table; the reviewer's line |
| 7 | T4 | the four `grep -c` outputs; the heading loop's output |
| 8 | T5 | `grep -ci rung` (1, the citation pattern); the reviewer's line on each changed line |
| 9 | T4, T7 | `grep -rn review-desk docs/plans \| wc -l` (0) on the close head; the 100% renames |
| 10 | T5 | `wc -c < CLAUDE.md`; the reviewer's line |
| 11 | T3, T7 | the verdict checks of T3; the reviewer's line |
| 12 | T1 | the survey outputs and the moved list |

## Retag table

| Item | Old | New |
| --- | --- | --- |
| #16 | `#rung:streamer` | `#round:playback` (slice 6) |
| #37, #38, #39, #40, #41, #89 | `#rung:cutover` | `#round:launch` |
| #42, #43, #44, #45, #46, #47, #48 | `#rung:streamer` | `#round:playback` (slices 3 to 6) |
| #49, #50 | `#rung:streamer` | `#round:queue` (slices 7 and 8) |
| #52, #74, #75, #81, #90 | `#rung:admin-api` | `#round:catalog` |
| #54 | `#rung:streamer #rung:admin-api` | `#round:playback #round:catalog` |
| #56, #57, #58, #83, #84, #85, #87, #88, #92, #126 | `#rung:intake` | `#round:intake` |
| #59, #60, #86, #91 | `#rung:pipeline` | `#round:catalog` |
| #61 | `#rung:integration` | `#round:launch` |
| #97, #108 | `#pass:gate-1` | `#pass:cleanup-ci` |
| #104, #105 | `#pass:gate-1` | `#pass:cleanup-gate` |
| #106, #107 | `#pass:gate-1` | `#round:launch` |
| #110, #111, #112, #113, #114, #115 | `#pass:cleanup-engine` | `#round:engine` |
| #120 | `#pass:web-shell-1` | `#pass:cleanup-web` |
| #125 | `#pass:cleanup-ops` | `#pass:cleanup-ops #round:engine` |

The `#rung:` words on lines 30 and 32 sit in "Path to 0.8", which T3 replaces. Items tagged `#pass:cleanup-store` or
`#pass:cleanup-ops` keep their tags.

## Verdict edits

Folds add the tag and change no text: #70 `#pass:cleanup-ci`; #34, #67, #69 `#pass:cleanup-gate`; #19
`#pass:cleanup-web`. #97, #104, #110 to #126 get theirs from the retag table.

Duplicates are removed: #78 (#118 already reads "Takes #78's normalizer copies"); #30, #33, #35 (#114 already reads
"takes #30, #33, and #35").

Done: #127 moves to `## Done`, checked, with "answered by Geoff's rulings of 2026-11-15 (the folder-admit command
replaces the migrator, no private deploy, the three candidates sit after 0.8)".

Rewords, each the item's new text in its own words, keeping its source note:

- #108: G14, G15, and G19 are this pass's (`#pass:cleanup-ci`). G18, the settings allow rules for mandated commands,
  rides the round's close packet. G16 and G17 stay with `workstation-refinement` (`#initiative:workstation-refinement`).
- #105: `cleanup-gate` takes the `--reduced` refusal, the duplicate work, the exported-unused leg, and `--fresh`
  through `gate.sh`. The `.golangci.yml` `run: allow-serial-runners: true` clause stays open after it.
- #64: the `check.sh` header's missing photo leg goes to `cleanup-gate` (`#pass:cleanup-gate`). The
  `tools/photos/check.test.mjs` failures and the `git ls-files` read stay open.
- #26: its store parts are #117 and #118, and #124 takes its `cmd` parts. The settings accessor holding its mutex
  across a database read (`internal/settings/accessor.go:88-112`) goes to `cleanup-store` (`#pass:cleanup-store`).
  One comment item stays: doctor's comment at `cmd/dubplate-server/doctor.go:630-634` says every refusal
  `validateMagicLinkSource` returns is keyed, but `validateSMTPPort` returns plain errors (`auth.go:418,421`). The
  design said two comment items stay; the second, identity `Valid`'s "non-empty source" doc, now matches the code
  (`internal/identity/identity.go:224-231`) and is dropped, under `## Planning misses` as `[stale-fact]` (standing
  answer 17).
- #125: the test cuts stay with `cleanup-ops`; making the process-citation leg gate moves to round `engine`, since
  about 125 of the 189 lines it would fail on are in engine packages.
- #66: drop "(5d-ii's doctor twelve-flag OR chain first)", fixed in `4de73e7`. The rest stays.
- #16: `principalCollision` is gone. The roster refuses a username another row holds
  (`internal/auth/roster.go:461-467`, `ErrDuplicateUsername`), but nothing refuses a username equal to another
  member's address. It lands in streamer slice 6 (`#round:playback`).
- #15: `warnForwardedFromLocalPeer` is at `internal/serve/serve.go:177`, its comment at `:159`.
- #57: the favicon paths are named at `cmd/dubplate-server/web.go:305-308` (the icons list) and `:498-506` (the
  favicon loop).

Keeps get `*(swept 2026-11-15)*` at the end of the item, before the date if the item ends in one: #13, #14, #18, #20,
#24, #52, #54, #72, #73, #74, #75, #76, #77, #79, #80, #81, #82, #83, #96, #103, #106, #107, #128. That is 23 items;
with their retags where the table gives one.

## Conventions table

The text the adapter's new section gives each heading. Where an archived text contradicts a rule this pass or
`pass-core` now holds, the reconciled form wins and the PR lists the departure.

| Heading | Reconciled text | Departs from the archived text |
| --- | --- | --- |
| Gate | `bash scripts/gate.sh` (full) and the fast lane, from the worktree root, with `timeout: 600000`, re-issued on exit 75 until `gate exit:`. A worktree session's Bash guard refuses the quoted `cairn-run-gate '<string>'` form and accepts the wrapper. Never fall back to bare `bash scripts/check.sh` for an acceptance; CI runs it bare by design. | names CI |
| The base, and why the fast lane names it | `check.sh` classes `go.mod`, `go.sum`, `web/package*.json`, `scripts/check-state-dir.sh`, and `Makefile` as gate machinery, which promotes a scoped run to full. The pass whose first task is the round's sweep passes `--base <task start>` (a full 40-hex id) on every later fast lane; `check.sh` uses a base only when it resolves to itself and is an ancestor of `HEAD`, so a wrong value costs time, never coverage. Other passes pass no `--base`. | ties the base to the sweep pass; drops the clock-file `base=` line |
| Setup | `npm ci --prefix web` in the session's worktree and in each subagent's before its first gate. | none |
| Test-first | Every `auth-data` task writes its acceptance tests first and records each failing at the task's start commit in the PR notes. | none |
| Mutation proofs are run | Each mutation is applied one at a time, the named test run, and the failure recorded in the PR notes with the test name. A mutation that does not fail is a finding fixed in the same task. None is narrated in a comment. | none |
| Fixture states | Every proof names the fixture state it needs (populated, selected, renamed, removed). A `web/` proof is a named rendered row (`web/scripts/rendered/run.ts`, host `cmd/dubplate-server/rendered_test.go`) saying what it intercepts and what it reaches for real; `web/` has no component test runner. | none |
| Layering | The plan names each import edge its packages must not gain, and a test or the two import-graph tests (`docs/standing-rules.md`, "Gate legs") holds each. | generic; the archived list was `admin-api-1`'s packages |
| No export ahead of a caller | Every new export names its production caller in the task that adds it. | none |
| Config keys | Declared through the registry from the pass's own file, with one `newProductionRegistry` line (`cmd/dubplate-server/main.go:99`); never by editing `internal/config/keys.go`. | generic |
| Skills | Go under `go-conventions`, with "idiomatic, architecturally beautiful Go" an acceptance dimension of every Go task; `web/` under `svelte-conventions`, `ts-conventions`, `daisyui`, and the visual identity spec, which every `web/` task reads. Copy says "the group" and "a member". | none |
| Per-task review | On each `auth-data` task, `diff-reviewer` at task scope on the task's commit or the merged pair's range. `accept` continues; `fix` gets one fix commit quoting each finding verbatim, then one re-read; a second `fix` or any `escalate` is a stop. Each verdict appends `review T<n>=<verdict> <sha>` to the PR notes. | none |
| Commits | One commit per task, the task id first in the subject, imperative mood, specific files only, with the session's trailer. A pair half merges with `git merge --no-edit <branch>`. | none |
| Records | PR notes at `<git dir>/pr-notes.md`, never committed, surviving a restart. The first line is `launch-base=<sha>`, the branch head at the executor's start. Before each dispatch the session appends `dispatch T<a>[+T<b>] sha=<40-hex>`; each task appends its test-first failures, mutation results, review line, and out-of-scope finds. A subagent writes nothing outside its worktree and returns its records in its report. | drops the planning-branch merge |
| Clock | The clock runs from the first execution commit to PR-ready (`pass-core` `executor.md`, "Score"). A task past twice its estimate gets one `## Decisions` line with the estimate and the actual, and the run goes on. | the `pass-task-clock` file is gone (mechanism 25) |
| Dispatches | A pair is two `general-purpose` Agent subagents with `isolation: "worktree"`, dispatched after the session commits, `sonnet` at `medium` unless the plan names `opus` for novel correctness-critical logic. Each gets its task's full text, the plan sections it cites, and this section; its first act is `git merge-base --is-ancestor <dispatch sha> HEAD`, a failure being a stop. It commits, never pushes. The session merges, runs the fast lane, runs any per-task review, and pushes. | generic models; drops the 48 KB paste (cites sections) |
| Stops | A stop is a question `advisor.md` "What always stops" lists, or a pass with no task left. Commit and push the work, open the draft PR if none exists, write the stop under `## Stop` at the end of `docs/STATUS.md` on the branch (the question, what was tried, the choices), push, and end. Never ask Geoff in chat, never run `notify-send`, never work around a refusal with another tool. Background commands end through `TaskStop`. | follows `round-run.md` "Stops"; drops `notify-send` and the wait |
| Resume | The same session, or a fresh one launched with the same prompt, enters the worktree with `EnterWorktree` `path`. It reads its task commits with `git log --no-merges --format='%h %s' <launch-base>..HEAD`, settles the last `dispatch` line by the worktree list (merged, merge now, or a stop for a candidate with no `T<n>:` commit), and resumes at the first task whose acceptance fails, never re-applying a committed task. | drops the planning-branch merge and the clock rewrite |
| Scope | A defect found outside a task goes to BACKLOG through `standards:log-issue` and the PR body's "Found, not done". | names `log-issue` |

## Decisions this plan takes

1. **No `concurrency` key.** A cancel-in-progress group would cancel a T6 break run when its revert lands. Two runs
   per proof cost nothing in a public repo.
2. **Default checkout depth, `persist-credentials: false`.** The full form reads `git ls-files` and no history
   (mechanism 16), so a depth of 1 serves it; credentials left in `.git/config` serve no step.
3. **Tools install into `$RUNNER_TEMP`,** never the checkout, because the gitleaks leg scans the tree.
4. **Checksums are literals in the workflow,** read once from each release's checksums file at execution, so a
   moved release asset fails the run instead of passing silently.
5. **libxml2 builds from source.** Ubuntu's package is 2.9, and the goldens hold 2.15.4's messages (mechanism 11).
6. **The omit input is an environment variable with a fixed name set** (`DUBPLATE_CHECK_OMIT`), only if standing
   answer 13 fires, so the local gate's form is unchanged.
7. **T3 and T4 run as a pair while the session does T5.** Their gates are docs-only under rule 2's `--base` (about
   a minute each), so the heavy lane no longer serializes them, and three disjoint file sets run at once. The
   session keeps T5 because it also owns any T2 fix, and both are short.
8. **G14 lands as a base rule, not a sweep on master.** The sweep stays in the round's first pass (design: "the
   round's only sweep"), and the adapter's fast lane says when to pass `--base`.
9. **#97 and #70 move to Done at the close,** since this pass completes them.
10. **#26's identity comment item is dropped as stale** (standing answer 17).

## Stops

| Stop | Standing answer |
| --- | --- |
| A leg can't run on the hosted runner | 13: install first, else `DUBPLATE_CHECK_OMIT`, named in the job name, item 1 flagged amended |
| xmllint or rclone can't be installed | the advisor's install choice; with none that works, a stop (they sit inside `go test`) |
| The CI check is red for an outside cause | 19: fix the workflow in the branch, list it in `## Decisions` |
| A task needs a file outside the row | 15: a forced call site under `## Decisions`, else file it |
| A finding no longer matches the code | 17: drop it, `[stale-fact]` under `## Planning misses` |
| A major version is available | wait for Geoff; record it with its unblock condition; take nothing |
| A lint bump raises findings in Go files | hold the bump, file it (outside the row) |
| A pass path past twice its estimate | 14: the clock line in `## Decisions`; the run goes on (the park line applied at planning) |
| The PR conflicts with master in this pass's files | rule 4: merge `origin/master`, resolve, gate, push |
| The PR conflicts with master outside this pass's files | a stop |
| The first push of `.github/workflows/` is refused | a stop (mechanism 3) |
| A web bump turns a gate red | T1: revert that package, file it with the red line, gate again |
| A runner fetch (the subsonic.org XSD or the PyPI wheel) fails | rule 4: one `gh run rerun --failed`; a second failure takes one `actions/cache` step keyed on the two pins, under `## Decisions` |
| The rendered leg is red on the runner | T2: fix the install step (`--with-deps`); with no install that works, a stop |
| A checksum read is refused | rule 9: the other named source once (`gh api` digest or WebFetch); refused again, a stop |
| The review stays short of `accept` after its second read | T7 step 4: the PR goes ready with that verdict |
| A red the executor cannot fix in two attempts, or a merge conflict between the pair's branches | a stop, per "Stops" above |

## Review record

One `standards:plan-reviewer` read on Opus at `medium`, 2026-11-16, at `a5ad875`, verdict "not ready": one blocker,
eight major, twelve minor, one over-ceremony. Every finding was checked at its source before the fold; none was
refused.

| # | Finding | Fold |
| --- | --- | --- |
| 1 | Blocker: T1's npm premise is false (nine minors, two majors outdated) | Mechanism 21 corrected; T1 takes the nine, holds typescript 7 and @types/node 26, installs Chromium, holds a bump that turns a gate red; check rewritten; 35 min |
| 2 | The path is under-priced against the 200-minute park line | Re-priced at 182 with the pair beside T5, `--base` on docs gates (22), and T2's CI wait off the path; the margin is stated under "Path" |
| 3 | T7's review bar contradicts `round-run.md` "Evidence" | T7 step 4 follows `round-run.md`: one fold, one fresh second read, ready with the verdict; the check and the prompt match |
| 4 | T5 makes the exported-unused leg gate, against design call 14 | T5 says the leg moves to `go tool deadcode` and stays a warning; a check added |
| 5 | shellcheck publishes no checksums file; no alternative for a refused read | Rule 9: digests from `gh api`, libxml2 from WebFetch; nothing downloaded runs locally; a refused read is a stop row |
| 6 | A newer rclone in CI diverges from the workstation's | CI pins the local versions; rclone v1.76.0 held (mechanism 17) |
| 7 | The skill's sweep record and CHANGELOG sit outside the row | Mechanism 29: the PR body is the record; held majors go to the PR body and BACKLOG |
| 8 | T1 never stages `go.mod` and `go.sum` before its gate | Mechanism 27; T1 stages all four files |
| 9 | No answer for a PR that conflicts with master | Rule 4 reads `mergeable` and merges `origin/master`; Pre-flight names who writes master; two stop rows |
| 10 | The `review-desk` literal is wrong mid-pass | T4 excludes this plan's files; acceptance 9 is read after T7 step 2 |
| 11 | The heading loop is vacuous for three headings | T4's loop counts inside the new section only |
| 12 | T1's Go check prints indirect modules | The direct-only `go list -f` form (mechanism 21) |
| 13 | T2's header check has no literal | 10 headers, named in mechanism 7 and T2's check |
| 14 | Standing answer 13 cannot omit the rendered leg | The rendered leg is fixed in its install step, else a stop; a stop row |
| 15 | Each CI run fetches live from subsonic.org and PyPI | Rule 4: one rerun, then one keyed `actions/cache` step; a stop row |
| 16 | The pair dispatch omits the ancestor check | The pair text carries `git merge-base --is-ancestor` and the dispatch line |
| 17 | Acceptance 5's STATUS literal reads the wrong tree | The evidence row reads `origin/master` after the conductor's edit; the item stays verbatim |
| 18 | T6's expected reds collide with "red stops the line" | Rule 4 exempts T6's break runs; rule 10 says what `reds` counts |
| 19 | Mechanism 3 checks the wrong push channel | Mechanism 3 and Pre-flight say SSH; a refused first workflow push is a stop row |
| 20 | `fetch-depth: 0` has a false justification | Default depth; Decision 2 rewritten |
| 21 | `standing-rules.md:109` pins Playwright too | Mechanism 28; T5 edits it |
| 22 | Over-ceremony: docs gates run the full form | Taken: rule 2's `--base <start>` |

The reviewer's estimate "OWNER FORK" is not a fork: `round.md` "Parking at planning" decides by the number. The
re-priced path is 182, so the plan launches. Its unverified items are carried: the libxml2 build system and rpath
(T2 settles it on the first run, and the `21504` assertion catches a failure), the watch starting before the run
registers (rule 4 waits for the head), and CI wall time against the 45-minute job timeout (unmeasured; a timeout is
standing answer 19).
