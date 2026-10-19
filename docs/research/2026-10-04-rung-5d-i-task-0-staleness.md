# Rung 5d-i Task 0: the staleness pre-task

Plan: `docs/superpowers/plans/2026-09-25-rung-5d-boundary-deployment-close.md` (revision 2.2).
Read against master at `311b74c` on 2026-10-04, 194 commits after the plan's pointer base
`4c766c3`. Box reads were read-only (`ssh -o BatchMode=yes`), names only. No tracked file was edited.

## Verdict

5d-i can launch after two small plan amendments. The code the build tasks lean on survived intact:
`cmd/dubplate-server/serve.go`, `internal/serve/clientaddr.go`, `strip.go`, `serve.go`, `prefix.go`,
and `server.go` are byte-identical to `4c766c3` at every cited line, `web.go`'s `newWebRoutes`
keeps its signature, logger, and read-only behavior, `internal/config` changed (6c) but every shape
R-5d1 to R-5d3 rely on holds, and the box matches "Box facts" except that master's `compose.yaml`
now pins a newer cloudflared. The two 5d-i blockers are process and placement: the implementer's
agent definition now forbids the registry install that every post-4D `deploy.sh --check` evidence
run performs, and the tree prints two families (unidecode, 6c's settings) after the backup rows, so
"after the backup rows" no longer names one position. Three further blockers sit in 5d-ii (the
cloudflared recreate in the sitting, the unidecode row in the pre-deploy local doctor, and two
STATUS-routed runbook rows Task 7 does not take); they need amendment before 5d-ii, not before
5d-i. Everything else is a moved pointer whose content survives.

## Blockers

1. **The implementer may no longer run `npm ci`, which every `deploy.sh --check 192.0.2.1`
   evidence run executes from Task 4D on (5d-i, Tasks 4D, 5, 6).** `bbd96d2` (2026-10-04) added to
   `.claude/agents/dubplate-implementer.md:44-46`: "A dependency bump's registry install and any
   upstream document fetch belong to the conductor: report BLOCKED naming the network step." R-5d13
   puts `npm ci --prefix web` inside `deploy.sh`, and "Global constraints" (plan `:303-307`,
   `:309-317`) runs every evidence command in a fresh detached scratch worktree, which has no
   `node_modules`, so each run is a registry install. The plan's "stated exception, named in every
   dispatch" now contradicts the agent's own hard rule, and CLAUDE.md's "a rule lives where it
   executes" puts the agent definition above a dispatch prompt.
   **Amendment:** land the plan's two named exceptions (a lockfile-pinned `npm ci --prefix web` inside
   a `deploy.sh --check` evidence run, and the unroutable `192.0.2.1` ssh and `rsync -n` attempts) in
   the agent definition's no-network rule before Task 4D dispatches, or have the conductor run each
   task's `--check` evidence itself.

2. **"After the backup rows" no longer names one position (5d-i, R-5d5, Task 2).** The tree now
   prints, after the backup remote row, the unconditional `unidecode` row (`doctor.go:131-138`,
   translit `850cb4c`) and then one row per registered setting plus orphans
   (`doctor.go:140-149`, 6c `2a5a201`), a variable-count family. `TestDoctorCmd` walks fixed rows by
   index through the unidecode row and then asserts no extra lines (`doctor_test.go:204-217`). Plan
   `:409-415`, `:818`, and Task 10 step 13 (`:1370-1376`) each say only "after the backup rows".
   **Amendment:** R-5d5 and Task 2 name the position, recommended after the `unidecode` row and
   before the settings rows, so every fixed-count row precedes the one variable family, and
   `probeRowLines` sits beside `unidecodeRowLines` (`doctor_test.go:43`).

3. **The sitting's apply now recreates the cloudflared container (5d-ii, Task 10 step 6, Owner
   touchpoints).** The dependency sweep `0677934` moved `compose.yaml`'s cloudflared image from
   `2026.8.2` to `2026.9.3`. The box runs `cloudflare/cloudflared:2026.8.2` and its
   `/opt/dubplate/compose.yaml` md5 `ba618f5b...` equals `be27434`'s, while master's is `0699781f...`.
   `git diff --name-only be27434 HEAD -- config/ compose.yaml systemd/ env/` now lists
   `compose.yaml` first. The apply's `docker compose ... up -d` (`deploy.sh:564-566`) will recreate
   the tunnel container, a brief Navidrome outage. The plan lists `compose.yaml` and "compose `up -d`
   (no recreate)" as no-ops (`:1331-1333`) and names "any change to `compose.yaml` or the tunnel" an
   owner touchpoint (`:201-202`).
   **Amendment:** move `compose.yaml` and the cloudflared recreate into Task 10 step 6's expected
   changes, named to Geoff at the sitting as a tunnel restart; the rollback leaves the new image in
   place (it is the repo's intended state) unless Geoff rules otherwise.

4. **The pre-deploy local doctor cannot exit 0 against a scratch state directory (5d-ii, Task 10
   step 3).** Translit added an unconditional `unidecode` row that fails with "missing: run
   dubplate-server unidecode install" when the state directory holds no table
   (`cmd/dubplate-server/unidecode.go:95-108`; exit condition `doctor.go:151`). Step 3
   (`:1291-1300`) substitutes a scratch state directory and expects "Every row read; exit 0".
   **Amendment:** step 3 first runs `<worktree>/dubplate-server-linux-amd64 unidecode install
   --from <the verified wheel cache>` with `DUBPLATE_STATE_DIR` set to the scratch directory
   (offline, `unidecode.go:124-135`), then runs doctor.

5. **Two STATUS-routed 5d items land in no task (5d-ii, Task 7).** STATUS's "No deploy until 5d"
   item routes to 5d "5c-iv's two rows (HISTORY 5c-iv 'To 5d')" and "the transliteration runbook row".
   Task 7 (`:1120-1158`) takes neither of these:
   (a) the operator runbook stating that the kind unknown-address answer reveals the first admin's
   contact and whether an address is a member (`HISTORY.md:516-519`); (b) the transliteration
   runbook row (`HISTORY.md:245-252`): `--check` first and read the staged install's "would run:
   systemd-run ... unidecode install" line, and keep the workstation's verified wheel cache as the
   `--from` source. Its second step, "Roll back by redeploying the previous commit", contradicts Task
   10's own finding (`:1260-1262`) that redeploying an earlier commit is no rollback.
   **Amendment:** add (a) and (b) to Task 7's outcomes, with (b)'s rollback line replaced by Task
   10's `.prev` rollback and the record saying it supersedes translit's.

## Notes

Moved pointers whose content survives, grouped so each amendment is a mechanical re-cite.

### Header, Sources, and Token ceiling

- HISTORY grew by 1715 lines above the older entries. 5a's forward carries `:3857-3885` are now
  `:5572-5600`, and `:4045-4056` are now `:5760-5771`. 5b-ii major 5 and minors 9 and 14 `:2244-2253`
  are now `:3959-3968`, and `:2405-2413` are now `:4120-4128`. 5c-i's runbook carries `:1489` and
  `:1729-1731` are now `:3204` and `:3444-3446`. 5c-ii's declared gaps `:555-608` are now
  `:2270-2323`.
- Token ceiling basis, 5c-i's actual `HISTORY.md:1692-1698`, is now `:3407-3413`.
- The 5c-iii plan's scope wall `:60-65` is now `:58-66`. Its Task 11 handoff `:1768-1777` is now
  `:2075-2097`; the task is now titled "The 5c-iv close" and names the same clean-build row.
- Rung 5 spec section starts are a few lines earlier than cited; the spec is unchanged since
  `4c766c3`, so these predate the plan. "The trust boundary" starts at `:786` (cited `:795`), "The
  state directory" at `:1683` (`:1690`), "Deployment" at `:1732` (`:1737`), and "Process" at `:2791`
  (`:2799`). Every other spot-checked spec, build-spec, rung 6 spec, and streamer-spec pointer resolves.
- Box facts: `dubplate-shadow.service:84-89` (AF_NETLINK dropped) is `:80-86`.
- The heavy gate figure is stale. The latest recorded full gate is 8m15s on the heavy lane at
  `5530cbb` (`HISTORY.md:127`, a recorded figure, not re-measured). The last recorded peak memory is
  about 560 MiB at 5c-iii (`HISTORY.md:1449-1450`); no later entry records memory. 5c-iv added 150
  rendered captures (`HISTORY.md:506-510`). The plan's "about 3-4 minutes" (`:222-224`, `:331-333`)
  understates each gate about twofold, so the +0.30M heavy-lane adder may run light.
- CLAUDE.md now opens every pass with a dependency sweep; the lane registry row says the 2026-10-04
  sweep (`0677934`, `b3c2578`) covers this pass.

### Global constraints and Resolutions

- `deploy.sh:133-137` (`run_remote`) is `:133-141`; `:123-125` and `:74` hold. The `|| true` dry
  runs are `:254-255`, `:267`, and `:373` (was `:341`).
- R-5d1: `main.go:108-117` is `:110-120`. `backup.go:40-54` is `:28-39` (the class) plus `:101-107`
  (its registration).
- `internal/config/registry.go:213` (`AddIntegrationClass`) is `:218`. `keys.go:184-190` holds.
  `config.go:301,408-427` holds at `593d7fd` (an ancestor of HEAD); at HEAD they are `:306` and
  `:419-433`.
- 6c's validator shape (item 5): `config.Setting` carries `Kind` (Int, Duration, or Float; String
  refused at registration), `Default`, `Doc`, `Apply`, `Floor`, `Ceiling`, `Sensitive`, and a
  `Validate(value)` built by `AddSetting` (`internal/config/settings.go:55-90`). It does not fit a
  list-valued integration key, so "Global constraints" `:273-277` stands. `IntegrationClass` still has
  no `Sensitive` field, so Task 1's hook 18 "neutral" holds.
- R-5d5: `doctor_test.go:217-235` (the `storeRowLines` precedent) is now the constants block at
  `:17-43`: `storeRowLines` `:21`, `backupDoctorRowsOff` `:36`, and a new `unidecodeRowLines` `:43`.
  The counts also add `len(reg.Settings())` where the production registry is used (`:362`, `:644`,
  `:917`). `doctor.go:35-36` is `:37-38`.
- R-5d11: the shadow pre-create `deploy.sh:503-505` is now `:319-322`, moved ahead of the binary step
  by translit; it still runs `mkdir -p` plus `chown` of both `/var/lib/dubplate` and the leaf.
  `DUBPLATE_LIB_DIR` `:502` is `:319`.
- R-5d12: the systemd leg `check.sh:631-643` is `:632-655` (the guard `:643`, the skip message
  `:654`); the tool list `:138-157` holds.
- R-5d13: `deploy.sh:240-247` and `:252` hold. `git describe` `:334` is `:354`. The binary step's
  build-under-`--check` precedent `:317-320` is `:337-340`. `web/package.json:8` is `:11` (`"build":
  "vite build"`, unchanged by SvelteKit 3). `Makefile:14-15` holds. After the SvelteKit 3 and
  adapter-static 4 build, `web/build/placeholder.txt` is rewritten in place and `git status
  --porcelain -- web/` is empty. Playwright stays `1.63.0`.
- R-5d16: the argument parse `deploy.sh:81-95` is `:77-92`; the usage `:18-24` holds.

### Task 1

- `backup.go:61-62` (`IntegrationEnabled`, `IntegrationValue`) is `:130-131`.
- All `serve.go` and `internal/serve` pointers hold, and `NewTrustedProxyConfig` still has only test
  callers. STATUS's "To 5d-i: `NewTrustedProxyConfig` has no caller" is covered: the resolver must
  build through it, since the struct's fields are unexported.

### Task 2

- The exit condition `doctor.go:129` is `:151`. `doctorEnviron` `doctor_test.go:908-916` is
  `:932-944` and now installs a unidecode table into its state directory.
- `themeColors` (`web.go:568-590`, changed at `75fde16`) now logs only for a branding neutral that is
  a color function or a four- or eight-digit hex literal; a `#rgb` or `#rrggbb` neutral logs nothing.
  The "neutral branding value ... line count unchanged" row must use a color function to exercise the
  record it means to suppress.

### Task 3

- No pointer moved (`prefix.go:130-134`, `serve.go:380-389`, `serve.go:84-90` hold).

### Task 4D

- The banner `deploy.sh:249` holds. The header comment's list of remote writes must now include
  translit's staged steps: the `.new` rsync and `chmod`, the `systemd-run ... unidecode install`
  (a PyPI download on the box), and the `mv` swap (`:363-440`).
- `make build-linux` currently runs inside the binary step at `:364`; R-5d13 moves the build ahead of
  the repo-config rsync, and the header comment at `:342-345` ("no web/ build step exists yet")
  becomes false.

### Task 5

- The verify filter `check.sh:634` is `:646`; the state-dir leg `:645-658` is `:657-673`.
- No new greenfield unit: `dubplate-shadow.service` is still the only `systemd/` unit whose
  `ExecStart=` names `dubplate-server`.
- Translit's `unidecode/` leaf is created by the verb itself (`unidecodeLeafPerm = 0o755`,
  `unidecode.go:39`) under `deploy.sh`'s `systemd-run -p ReadWritePaths=/var/lib/dubplate`, is not
  pre-created by `deploy.sh`, and is granted to no unit. By item 4's rule it does not join
  `DUBPLATE_SERVER_LEAVES`; the loop's comment can say why.
- The agent definition now requires any local server on an ephemeral loopback port and an `ss -ltnp`
  check at the end (`dubplate-implementer.md:47-50`). The `/dev/shm` evidence's `dubplate-server
  serve` run should set `DUBPLATE_SERVE_PORT` to a reserved free port, never the default 8080.

### Task 6

- `deploy.sh:183-205` holds (`:183-204`); the template's comment block `:38-53` holds.
- The rows `doctor` stats or opens now include the unidecode table under the state directory. The
  rendered-doctor fixture needs `installTestUnidecodeTable` (`unidecode_test.go:43-52`) into its
  state directory, or the row fails exit 1.

### Task 7 and Task 10 (5d-ii)

- Task 7's `HISTORY:884-887` is `:2599-2602`, and `HISTORY:1729-1731` is `:3444-3446`.
- Task 10: the binary swap `deploy.sh:337-343` is now the staged rsync `:363-377` plus the swap
  `:439-440`. The env install `:407-414` is `:458-488`, `TIMER_UNITS` `:526` is `:582`, the compose
  step `:508` is `:564-566`, and the beets step `:530-557` is `:585-616`.
- Step 7 says each `.prev` copy "keeps mode 0600 and root ownership". That holds for both env files.
  The box's `/usr/local/bin/dubplate-server` is mode 0755, owned `1000:1000` (rsync `-a` kept the
  workstation owner), and `cp -p` preserves that.
- Step 6's `--check` does not exercise the box-side PyPI download and table build, so "twice step
  6's wall time" may underestimate the apply.
- The expected set should name the `== /etc/dubplate/contributors.yaml ==` "would install" line. The
  box copy differs from master by the one "family" to "group" comment line. This was true when the
  plan was written.
- The rollback should also remove a leftover `/usr/local/bin/dubplate-server.new` (left when the
  apply aborts at the table install) and list `/var/lib/dubplate/unidecode` among what it leaves
  behind.
- Step 10 should expect the `unidecode` leaf at `music:music 0755` beside the two 0700 leaves.

### STATUS items outside 5d (no amendment to this plan)

- "6c's Ruling 2 (`deploy.sh` halts on `check`'s nonzero exit before restarting `serve`)" is 6c's
  handoff item 5 (`HISTORY.md:859`). The 6d plan takes it (R-6d15, `2026-09-26-rung-6d-...md:303`,
  `:650`). 5d's `deploy.sh` never restarts `serve`.
- "5d's deploy gate checks 6d's two wiring items" (settings apply at serve start; the store through
  `OpenServe`, which has no non-test caller today). The 6d plan's R-6d1 (`:360-368`) records that
  this line contradicts the approved 5d plan and that the conductor rewrites it to "6d lands before
  any deploy that enables the serve unit or routes traffic to it." Task 10 already stops on a "seeds
  at serve's next start" row and leaves the unit disabled, so the line is a STATUS edit, not a plan
  gap.
- STATUS's other 5d items are taken. The nine magic-link keys go to Task 6, `doctor` exiting 0 to
  Tasks 6 and 10, `ReadWritePaths` to Task 5, and the trusted proxy to Task 1. The theme keys' row
  goes to Task 2. The `Referrer-Policy` residual, branding in both schemes, and 6a's and 6b-i's
  backup rows go to Task 7. The clean-`web/` build, with the rendered leg's absence on the box, goes
  to Task 4D under R-5d13.

## Item 4: the translit `--check` lines for Task 10 step 6

Translit has merged (code at `5530cbb`, deploy staging at `2ef4000` and `e8fadec`) and has not
deployed: the box binary is `pre-greenfield-321-gbe27434` (matcher-commit `be274346...`) and
`/var/lib/dubplate` holds only `shadow`. The "if translit merged" block applies. These are the exact
lines `deploy.sh` prints under `--check` for its steps, read from the script with its constants
substituted, never run against the box:

```
== /var/lib/dubplate/shadow ==
  would run: mkdir -p /var/lib/dubplate/shadow && chown music:music /var/lib/dubplate /var/lib/dubplate/shadow
== dubplate-server binary -> root@5.78.148.42:/usr/local/bin/dubplate-server.new ==
  (make build-linux output, then rsync -avzn's listing of dubplate-server-linux-amd64 against
   the LIVE path /usr/local/bin/dubplate-server, deploy.sh:373; no chmod line under --check)
== smoke test: dubplate-server version ==
  would run: if ! output="$(/usr/local/bin/dubplate-server.new version)"; then echo "dubplate-server version failed to execute" >&2; exit 1; fi; printf "%s\n" "$output"; commit=$(printf "%s\n" "$output" | awk "/^matcher-commit/{print \$2}"); if [ -z "$commit" ] || [ "$commit" = "dev" ]; then echo "matcher-commit is dev or empty; aborting deploy" >&2; exit 1; fi
== /usr/local/bin/dubplate-server.new unidecode install ==
  would run: systemd-run --wait --pipe --collect   -p User=music -p Environment=DUBPLATE_STATE_DIR=/var/lib/dubplate   -p NoNewPrivileges=yes -p PrivateTmp=yes -p ProtectSystem=strict   -p ProtectHome=yes -p ProtectKernelTunables=yes   -p ProtectKernelModules=yes -p ProtectControlGroups=yes   -p RestrictNamespaces=yes -p RestrictRealtime=yes   -p RestrictSUIDSGID=yes -p LockPersonality=yes   -p MemoryDenyWriteExecute=yes -p SystemCallFilter=@system-service   -p SystemCallArchitectures=native -p AmbientCapabilities=   -p CapabilityBoundingSet=   -p 'RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX'   -p ReadWritePaths=/var/lib/dubplate   /usr/local/bin/dubplate-server.new unidecode install
== swap /usr/local/bin/dubplate-server.new -> /usr/local/bin/dubplate-server ==
  would run: mv -f /usr/local/bin/dubplate-server.new /usr/local/bin/dubplate-server
```

Corrections to the plan's block (`:1326-1330`): no `deploy.sh` line pre-creates
`/var/lib/dubplate/unidecode`. The verb creates the leaf itself, mode 0755, inside the `systemd-run`
sandbox, and the PyPI download happens inside that same `systemd-run` line. The shadow pre-create
line now prints before the binary step rather than among the late state-directory steps.
Translit's HISTORY records that the box's systemd 257 has not yet accepted this property list
(`HISTORY.md:241-244`); its first acceptance happens at the apply.

R-5d16's cross-lane note resolves: translit has not deployed, so its first deploy is Task 10's
`deploy.sh --apply`.

## Box drift beyond "Box facts"

Read 2026-10-05 UTC, read-only:

- **`compose.yaml`**: the box is at `be27434`'s content (cloudflared `2026.8.2`, running); master
  pins `2026.9.3` (`0677934`). See Blocker 3.
- **`/etc/dubplate/contributors.yaml`**: md5 `a2bccb64...` against master's `6092ec4e...`, the one
  comment line. `deploy_file` will report "would install".
- **The binary's owner**: `/usr/local/bin/dubplate-server` is `1000:1000` mode 0755, not root.
- Matching "Box facts", or matching master: systemd 257 (AlmaLinux 10.2), SELinux Enforcing,
  `music` uid 2000, `/var/lib/dubplate` and `shadow` both `music:music` 0755 `var_lib_t`, nothing on
  8080. `/etc/dubplate/server.env` has the ten rung 4 row names and no `DUBPLATE_IDENTITY_SOURCE`.
  The four `TIMER_UNITS` and `dubplate-shadow.timer` are enabled and active, and
  `dubplate-serve.service` and `dubplate-import-nudge.path` are disabled. Beets is 2.13.1 against
  the pin 2.14.1. `/etc/dubplate/env`'s 42 row names equal master's `env/dubplate.env.template`
  names in order. All 14 units in `/etc/systemd/system/dubplate-*` and
  `/etc/tmpfiles.d/dubplate.conf` are byte-identical to master. Under `/opt/dubplate/config/`,
  `contributors.yaml` and `reject-email.txt` differ, `roster.example.json` is absent, and the other
  four files match.

## Per-item status

1. **Pass.** HEAD `311b74c`. 5c-iv's close is `4f4930b` (fold fix `d618bff`). 6c merged at `e90998a`
   (range `1fc7260..e90998a`, `HISTORY.md:638`), and 6c-fix at `0e3272c..8950de7`. The lane
   registry names 5d-i in lane 1; lane 2 is idle. The only warm tracked change is the conductor's
   lane-registry row in `docs/STATUS.md`. The rest are untracked agent-memory files and two
   2026-09-04 research notes, none in 5d's territory.
2. **Pass, with Blocker 5.** 5c-iv's "To 5d" names the clean-build row, which Task 4D takes, and the
   kind unknown-address runbook row, which no task takes.
3. **Pass.** No 5d file in `serve.go` or `internal/serve` changed at any cited line. `newWebRoutes`
   holds (`web.go:218`, logger parameter, no write). `baseEnviron` and `authEnviron`
   (`check_test.go:119`, `:141`) set no `DUBPLATE_SERVE_PORT`. Doctor pointers moved (Notes).
4. **Done.** Translit added the `unidecode` row and `unidecodeRowLines`. 6b-ii-b added the
   `restore` verb (`main.go:146`) and no doctor row. No new greenfield unit. The unidecode leaf is
   not pre-created or granted. Translit has not deployed. The lines are listed above.
5. **Done, with Blocker 2.** `internal/config` changed (6c: `settings.go`, `registry.go`,
   `config.go`); the validator shape is under Notes. 6c added settings rows after the unidecode row.
   `check.go` gained `printSettingsRows` (`:88`) and no fixture Tasks 1 to 3 or 6 need.
6. **Moved; gate figure stale.** The tool list holds. The systemd and state-dir legs moved (Notes).
   The heavy gate's recorded wall time is 8m15s at `5530cbb`. The last recorded peak memory is about
   560 MiB, at 5c-iii. Both are recorded figures.
7. **Moved.** Every `deploy.sh` pointer is re-cited under Notes. `dubplate-shadow.timer` is not in
   `TIMER_UNITS` (`:582`).
8. **Pass.** The template is unchanged since `4c766c3`: ten rung 4 rows plus
   `DUBPLATE_IDENTITY_SOURCE=magic-link`.
9. **Pass.** In `~/.local/secrets`, the five SMTP names are present, and `DUBPLATE_MAGIC_LINK_SECRET`,
   `DUBPLATE_SESSION_SECRET`, and both `_PREVIOUS` names are absent. In
   `~/.dotfiles/secrets/registry.md`, `DUBPLATE_SMTP_PASSWORD` is documented, the four non-secret
   SMTP names are covered by the `:180` note the plan cites, and neither HMAC name appears. No value
   was printed.
10. **Done, with Blocker 3.** Drift is listed above.
11. **Pass, with moves.** The `internal/auth` pointers hold: `store.go:32,36-39,53-71`,
    `roster.go:318-319`, and `handlers.go:189-197`. So do `pingfail.go:64`, `auth.go:71`,
    `member.go:120-124`, and `secret-set.sh:120`. The shadow `run.go:89-92` and
    `reconcile.go:49,78-81` hold, as do `dubplate-implementer.md:43-44`. The rest moved (Notes).
12. **Pass.** The state-dir check is still hard-coded to `systemd/dubplate-shadow.service`
    (`check.sh:665-666`). `doctor` still derives the five registry families from `reg`
    (`doctor.go:40-74`) and the settings rows from `reg.Settings()`. `.gitleaksignore` is still the
    empty tracked file, last touched at `3a35558`, before rung 5. Only lane 1 has written
    `internal/serve` since `4c766c3` (`75fde16`, `b3c2578`).
