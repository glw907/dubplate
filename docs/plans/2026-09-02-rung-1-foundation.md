# Rung 1 (Foundation) Implementation Plan

> **For agentic workers:** this plan runs in workflow mode (owner
> opt-in, 2026-09-02) via the pass-execute chain: one
> implementer-diff-reviewer-gate chain per task. Tasks are sequential
> except where a dependency note says otherwise. The conductor
> consumes reports and never reads diffs inline.

Revision 2, 2026-09-02: folds the adversarial plan review
(`docs/research/2026-09-02-greenfield-spec-reviews/rung-1-plan-review.md`),
all four blockers, eight majors, and ten minors. The review verified
the pivot's toolchain mechanics (underscore exclusion, tidy prune,
vale gap, systemd filter); implementers should not re-litigate those.

**Goal:** Pivot the repo to the greenfield tree and land the config
core, the `dubplate-server` skeleton, and a truthful gate, per the
build spec's rung 1.

**Architecture:** One mechanical pivot commit freezes the pre-pivot Go
and web trees under `_salvage/` and repoints build and gate plumbing
at `cmd/dubplate-server`. A separate judgment task makes the deploy
path additive (new install path beside the untouched incumbent
binary). The greenfield `internal/config` lands the derived wall's
full mechanism: three key classes (owned, reserved-for-a-rung,
unknown), a superseded map naming replacements, a real threshold
registry, and collected errors. `check` and `doctor` scaffolds make
the wall observable.

**Tech Stack:** Go 1.27, cobra, module `github.com/glw907/dubplate`.
No new dependencies.

**Spec:** `docs/specs/2026-09-02-greenfield-server-build.md`
(revision 2.1, ratified). Rung 1's ladder entry and the sections
"Repo pivot and tree shape", "The config wall, derived", "The Go
bar", and "Box safety" govern this plan.

**Token ceiling:** 5M (an M rung). **Checkpoint:** write STATUS after
Task 3, and at any split or question.

## Global Constraints

- Invoke the `go-conventions` skill before writing any Go. The bar is
  exemplary, not conforming (spec, "The Go bar"); reviewers fail
  awkward-but-working code.
- Universality: no hostname, path, email, or personal name outside
  `internal/config`. Nothing estate-specific enters the greenfield
  tree.
- `_salvage/` is read-only reference material from the moment it
  exists. Nothing imports from it; no file in it is edited.
- The gate is `bash scripts/check.sh`; every task ends green and with
  a clean `git status` (no untracked residue). Never commit red.
- No box access of any kind in this rung. `scripts/deploy.sh` is
  edited but never run, not even `--check`.
- Table-driven tests under assertion discipline: no silent-success
  fakes, no vacuous assertions; prove each test can fail (via the
  declared-stub step, not a compile error).
- Prose follows writing-voice: Go Doc Comments for code, the Google
  developer-doc register for docs. Em dashes are banned in code
  comments.
- Commits: imperative mood, specific files (never `git add -A`),
  `Co-Authored-By: Claude <noreply@anthropic.com>` footer.

---

### Task 1: The pivot commit (mechanical)

**Files:**
- Move (git mv): `cmd/` → `_salvage/cmd/`, `internal/` →
  `_salvage/internal/`, `web/` → `_salvage/web/`, `e2e/` →
  `_salvage/e2e/`
- Delete before the move: `web/node_modules/` (139MB of regenerable
  install state; `_salvage/` keeps source, not artifacts)
- Create: `cmd/dubplate-server/main.go`
- Modify: `.gitignore`, `Makefile`, `scripts/check.sh`,
  `scripts/vale-comments.sh`, `go.mod`, `go.sum`

**Interfaces:**
- Consumes: nothing (first task).
- Produces: a module whose only Go package is `cmd/dubplate-server`;
  the annotated tag `pre-greenfield`; `make build-linux` producing
  `dubplate-server-linux-amd64`; the derived subcommand surface as
  `main.go`'s package doc (Tasks 3-4 and rungs 4-14 name verbs from
  it).

**Steps:**

- [ ] **Tag the pre-pivot state.** Annotated tag `pre-greenfield`,
  message: "Last commit with the pre-greenfield tree buildable in
  place; emergency guardian fixes build from here." Local until
  Task 5 pushes it.
- [ ] **Remove `web/node_modules/`** (it is ignored in place today;
  moved, it becomes 5,256 untracked files), then **move the four
  trees** with `git mv`.
- [ ] **Re-anchor `.gitignore`:** the `/web/...` patterns move to
  `/_salvage/web/...` (dist patterns and the placeholder negation
  included); add `/dubplate-server` and `/dubplate-server-linux-amd64`
  beside the old artifact ignores. Acceptance probe: after
  `make build-linux`, `git status --porcelain` is empty.
- [ ] **Create `cmd/dubplate-server/main.go`:** a cobra root command
  named `dubplate-server`, short description "organizes and serves
  one shared music library", and a `version` subcommand printing
  `var version = "dev"`. The package doc carries the re-derived
  subcommand surface (spec, "Repo pivot and tree shape"), derived
  from the organizer spec's list minus `beet` passthrough, each verb
  with its owning rung: `version`, `check`, `doctor` (rung 1, the
  latter two in Task 4); `shadow` (rung 4); `serve` (rung 5);
  `backup`, `restore` (rung 6); `import`, `ping-fail` (rung 9).
  `nudge-consume` is not carried forward: its units were retired in
  the 2026-09-02 wind-down and no greenfield rung claims it. This doc
  comment is the record later rungs derive from; a rung claiming a
  new verb amends it.
- [ ] **Rewrite `Makefile`:** `BINARY := dubplate-server`; `build`,
  `build-linux`, `install` point at `./cmd/dubplate-server`; delete
  `WEB_DIR` and the four `web-*` targets with their comment block;
  `check: vet lint test tidy-check` with a one-line comment that the
  web legs rejoin at rung 7; `clean` and `.PHONY` updated.
- [ ] **Edit `scripts/check.sh`:** remove the npm prerequisite check;
  update the header comment and the `== make check (Go + web) ==`
  echo to say what the gate now runs (Go only until rung 7). Other
  legs stay exactly as they are.
- [ ] **Edit `scripts/vale-comments.sh`:** exclude `_salvage/` from
  the file list (both the tracked and untracked unions; git's `*.go`
  pathspec matches at any depth, so without the exclusion all 76
  salvage Go files re-enter the scan). Verification, one-time and
  recorded in the task report rather than claimed as a test: plant an
  em dash in a `_salvage/` comment unstaged, confirm the script stays
  green, revert.
- [ ] **Tidy the module:** `go mod tidy` prunes to cobra's set;
  commit the shrunken `go.mod`/`go.sum`.
- [ ] **Run the gate** (`bash scripts/check.sh`, exit 0) and build
  probes: `make build && ./dubplate-server version` prints `dev`;
  `make build-linux` leaves `git status --porcelain` empty.
- [ ] **Commit once.** A half-moved tree does not build, so there is
  no green intermediate to commit at; the single pivot commit is the
  deliberate exception to step-wise commits, and its diff is
  dominated by renames. `code-simplifier` runs scoped to
  `cmd/dubplate-server/main.go` only, because `_salvage/` is
  read-only.

**Acceptance criteria:** gate green; `git tag -l pre-greenfield`
non-empty; `go list ./...` shows exactly `cmd/dubplate-server`;
`make build-linux` produces `dubplate-server-linux-amd64` and the
tree stays clean; `main.go`'s package doc carries the full derived
surface with rung owners.

**Deliverable count: 4** (moves + tag + gitignore; skeleton with the
derived surface; Makefile/check.sh rewiring; vale exclusion + tidy).
Nothing may be added after dispatch.

---

### Task 2: Ops audit and the additive deploy path (judgment)

Runs after Task 1. This task carries the pass's production-adjacent
risk and is reviewed on its own.

**Files:**
- Modify: `scripts/deploy.sh`
- Possibly delete, per the audit's classification:
  `config/filebrowser.yaml` and any other *orphaned* configuration

**Interfaces:**
- Consumes: Task 1's `make build-linux` artifact name.
- Produces: a deploy path rung 4 can run additively: it builds and
  installs `dubplate-server-linux-amd64` to
  `/usr/local/bin/dubplate-server`, a new path beside the untouched
  incumbent `/usr/local/bin/dubplate`.

**Steps:**

- [ ] **The ops audit, as classification rather than a grep-empty
  criterion:** run `grep -ri filebrowser scripts/ systemd/
  cloud-init/ compose.yaml config/ env/`; classify every hit as
  orphaned configuration (delete) or historical record (keep). Known
  dispositions the plan review established: `config/filebrowser.yaml`
  is an orphan and goes; the `_comment` in
  `config/cloudflared-tunnel-config.json` records the wind-down and
  the current tunnel state, is the record of what gets PUT to the
  Cloudflare API, and stays. Report both lists.
- [ ] **Make the deploy binary step additive.** In
  `scripts/deploy.sh`: the build step keeps `make build-linux` (now
  producing `dubplate-server-linux-amd64`); the rsync target becomes
  `/usr/local/bin/dubplate-server`; the chmod and the version smoke
  test follow the new path. The incumbent `/usr/local/bin/dubplate`
  is never written, so the four guardian units that exec it are
  unaffected by any future deploy run. Add one comment stating the
  boundary: the incumbent binary is frozen (its source is the
  `pre-greenfield` tag) and retires at rung 16; greenfield units
  arrive from rung 4 and exec the new path.
- [ ] **Do not run deploy.sh.** Verification is shellcheck (in the
  gate) plus the diff review against the boundary comment.
- [ ] Gate; commit.

**Acceptance criteria:** shellcheck green; no rsync, chmod, or smoke
step in `scripts/deploy.sh` references `/usr/local/bin/dubplate` as a
write target (all reference `/usr/local/bin/dubplate-server`); the
audit report lists kept records with reasons;
`config/filebrowser.yaml` gone.

**Deliverable count: 2** (audit with classification; additive deploy
path).

---

### Task 3: The config core

Runs after Task 1; independent of Task 2 (disjoint files).

**Files:**
- Create: `internal/config/config.go`, `internal/config/keys.go`,
  `internal/config/thresholds.go`
- Test: `internal/config/config_test.go`,
  `internal/config/keys_test.go`, `internal/config/thresholds_test.go`

**Interfaces:**
- Consumes: nothing from Tasks 1-2 beyond the module layout.
- Produces (Task 4 and later rungs build against these exactly):

```go
package config

// Config is the resolved configuration of one dubplate-server
// process. Rung 1 carries the owned wall only; the identity-source
// selector and integration blocks join with their rungs.
type Config struct {
    LibraryRoot string
    StateDir    string
    SiteName    string
    AdminEmail  string // addr.Address form, lowercased
}

// Load resolves configuration from environ (os.Environ() form).
// It reads only DUBPLATE_-prefixed keys and returns an error that
// names every problem at once.
func Load(environ []string) (*Config, error)

// Threshold describes one tunable: its key, default, and meaning.
// Rungs register thresholds; doctor reports them with effective
// values.
type Threshold struct {
    Key     string
    Default string
    Doc     string
}

// Thresholds returns the registered tunables in registration order.
func Thresholds() []Threshold
```

- `keys.go` is the wall table as code: the owned set, the reserved
  classes, and the superseded map (below). `doctor` derives its
  output from these tables; nothing hard-codes a key count.
- Internal shape, adopted on merit from the salvage tree
  (`_salvage/internal/config/config.go`, the single `get` choke point
  with an error accumulator, documented there as the seam a config
  file source later plugs into): one lookup function owns the
  duplicate-key (last wins), empty-equals-unset, and `=`-in-value
  decisions; helpers and `Load` all route through it.

**The key classification (the wall's contract; `keys.go` encodes
this and the package doc explains it):**

1. **Owned, required now:** `DUBPLATE_LIBRARY_ROOT`,
   `DUBPLATE_STATE_DIR`, `DUBPLATE_SITE_NAME`,
   `DUBPLATE_ADMIN_EMAIL`. Plus the identity-source row, present in
   the table and the package doc from rung 1: a selector whose
   discriminator and required-when-selected blocks arrive at rung 5
   (spec: "login has one wall shape, owned-with-a-selectable-source").
   Doctor prints its row as `pending (arrives at rung 5)`.
2. **Reserved for a rung** (recognized; setting one before its rung
   is a load error naming the class and rung, never "unknown"):
   auth and identity (`DUBPLATE_MAGIC_LINK_SECRET`,
   `DUBPLATE_SESSION_SECRET`, `DUBPLATE_PUBLIC_URL`,
   `DUBPLATE_SMTP_*`) at rung 5; backup targets (`DUBPLATE_R2_*`) at
   rung 6; intake (`DUBPLATE_UPLOADS_DIR`, `DUBPLATE_TRIGGER_DIR`) at
   rung 7; monitoring (`DUBPLATE_PING_*`) at rung 9; streamer
   (`DUBPLATE_NAVIDROME_*`) at rung 12. The implementer walks every
   `DUBPLATE_*` name in `env/dubplate.env.template` and gives each a
   disposition in these tables; the acceptance criterion below makes
   "none lands in unknown" checkable.
3. **Superseded** (load error naming the replacement):
   `DUBPLATE_LIBRARY_DIR` → `DUBPLATE_LIBRARY_ROOT`;
   `DUBPLATE_SERVE_SITE_NAME` → `DUBPLATE_SITE_NAME`;
   `DUBPLATE_ACCESS_TEAM_DOMAIN` and `DUBPLATE_ACCESS_AUD` → the
   forward-auth front's config (rung 5);
   `DUBPLATE_PUBLIC_NAVIDROME_URL` → streamer driver config
   (rung 12). Template names that are genuinely retired (no
   greenfield successor) join this map with a "retired" note rather
   than falling to unknown.
4. **Unknown:** any other `DUBPLATE_*` key errors as unknown.

**Behavior contract (tests written from this table first, against a
declared stub of `Load` returning a not-implemented sentinel, so each
row fails as an assertion rather than a compile error):**

| Case | environ | Expected |
|---|---|---|
| all present | four owned keys set | Config populated, nil error |
| one missing | ADMIN_EMAIL absent | error naming `DUBPLATE_ADMIN_EMAIL` |
| several missing | only SITE_NAME set | one error naming all three missing |
| bad email | ADMIN_EMAIL=`not-an-address` | error naming the key |
| email normalized | ADMIN_EMAIL=`Geoff <GEOFF@X.ORG>` | AdminEmail == `geoff@x.org` |
| reserved key early | + DUBPLATE_R2_BUCKET=x | error: backup-target config, arrives at rung 6 |
| superseded rename | + DUBPLATE_LIBRARY_DIR=x | error naming `DUBPLATE_LIBRARY_ROOT` |
| superseded rename | + DUBPLATE_SERVE_SITE_NAME=x | error naming `DUBPLATE_SITE_NAME` |
| unknown key | + DUBPLATE_FROBNICATE=1 | error: unknown configuration key |
| non-namespace noise | PATH, HOME present | ignored |
| empty value | DUBPLATE_STATE_DIR="" | treated as missing |
| duplicate key | STATE_DIR set twice | last value wins |
| `=` in value | SITE_NAME=`A=B` | SiteName == `A=B` |

Threshold registry tests: registering a test threshold makes it
appear in `Thresholds()` and parse from environ with its default on
unset; a malformed value errors naming the key and the accepted form.
Rung 1 registers no real thresholds; the mechanism is the
deliverable, stated plainly (this is the registry the spec names, not
a deferral).

**Steps:**

- [ ] Stub `Load`; write the contract-table tests; confirm each row
  fails as an assertion.
- [ ] Implement the lookup choke point, the `keys.go` tables
  (completing the template walk), and `Load` with collected errors
  (`errors.Join`).
- [ ] Implement and test the threshold registry.
- [ ] Package doc: the classification and the wall table, marked as
  the contract of record; each reserved class named with its rung.
- [ ] Gate; commit.

**Acceptance criteria:** every contract row passes and can fail; a
test iterates every `DUBPLATE_*` name in `env/dubplate.env.template`
and asserts none classifies as unknown; the identity-source row
exists in `keys.go` and the doc; gate green.

**Deliverable count: 4** (Load + classes; template disposition walk;
threshold registry; wall doc).

---

### Task 4: `check` and `doctor` scaffolds

Runs after Task 3.

**Files:**
- Create: `cmd/dubplate-server/check.go`, `cmd/dubplate-server/doctor.go`
- Test: `cmd/dubplate-server/check_test.go`,
  `cmd/dubplate-server/doctor_test.go`
- Modify: `cmd/dubplate-server/main.go` (register commands; exit and
  silence wiring)

**Interfaces:**
- Consumes: `config.Load`, the `keys.go` tables, and
  `config.Thresholds()` exactly as Task 3 defines them.
- Produces: `newCheckCmd(environ []string, stdout, stderr io.Writer)
  *cobra.Command` and `newDoctorCmd(environ []string, stdout,
  stderr io.Writer) *cobra.Command`, the constructor shape every
  later subcommand follows. `main` passes `os.Environ()`,
  `os.Stdout`, `os.Stderr`, sets `SilenceUsage` and `SilenceErrors`
  on the root, and exits 1 when `Execute` returns an error.

**Behavior contract:**

- `check`: loads config; success prints `configuration: ok` to
  stdout, exit 0; failure prints the load error to stderr, exit 1.
  No usage text on failure; a failed `check` in a systemd journal
  must be the error alone.
- `doctor`: prints one line per owned wall row (the identity-source
  row included, as `pending (arrives at rung 5)`), then one line per
  reserved class (`disabled (arrives at rung N)`), then one line per
  registered threshold with its effective value. Line format is
  pinned: `fmt.Fprintf(out, "%-32s %s\n", key, status)`, with
  `: reason` appended to `invalid` statuses. Exit 0 when the wall is
  satisfied, 1 otherwise. The row sets derive from Task 3's tables;
  no count is hard-coded.

**Test cases (in-process execution, asserting line-by-line against
the pinned format; expected owned and reserved rows are read from the
same `keys.go` tables the command uses, with a literal-string spot
check on one known line so the test cannot go vacuous):**

| Command | environ | Assert |
|---|---|---|
| check | full valid wall | stdout has `configuration: ok`; nil error; stderr empty |
| check | missing STATE_DIR | stderr mentions `DUBPLATE_STATE_DIR`; error non-nil; no usage text |
| doctor | full valid wall | every owned row `ok` except identity-source `pending (arrives at rung 5)`; every reserved row `disabled`; nil error |
| doctor | missing two keys | those two rows read `missing`; error non-nil |
| doctor | bad ADMIN_EMAIL | its row reads `invalid: ` with a reason; error non-nil |

**Steps:**

- [ ] Tests first from the table against stubs; confirm assertion
  failures.
- [ ] Implement both constructors; wire main (registration, silence
  flags, exit code).
- [ ] Gate; commit.

**Acceptance criteria:** the five cases pass; `dubplate-server check`
with an empty environment prints only the error (no usage block) and
exits 1; gate green.

**Deliverable count: 3** (check; doctor; constructor and exit
pattern).

---

### Task 5: Ledgers, ROADMAP, and reconciliation

Runs last.

**Files:**
- Modify: `ROADMAP.md`, `docs/STATUS.md`, `docs/HISTORY.md`,
  `CLAUDE.md`, `.claude/settings.json`, `env/dubplate.env.template`
  (header note only)

**Steps:**

- [ ] **ROADMAP:** the greenfield server build (17-rung ladder) is
  the active initiative; fold the organizer spec's parked "Roadmap
  revisions this spec carries" list; packaging stays sequenced after.
- [ ] **CLAUDE.md, verified claim by claim:** remove the bats
  sentence (suite deleted 2026-08-31); the gate description lists
  what `scripts/check.sh` runs post-pivot; remove the gitleaks claim
  (the plan review verified no mechanism exists anywhere in the
  repo, so no investigation is needed; report the removal); point
  the salvage-tier language at `_salvage/` and the `pre-greenfield`
  tag.
- [ ] **`.claude/settings.json`:** remove the dead permissions the
  review identified (`Bash(bats tests/*)`, `Bash(npm run build*)`,
  `Bash(npx vite build*)`; the npm pair returns with rung 7).
- [ ] **`env/dubplate.env.template`:** header gains two lines: this
  manifest describes the frozen pre-pivot binary's wall
  (`_salvage/internal/config`); the greenfield wall is
  `internal/config`'s package doc, and this manifest evolves at rung
  4. No key changes. The age-store item `DUBPLATE_SERVE_SITE_NAME`
  stays minted exactly as STATUS records; deploy.sh's render reads
  it from this template today. Minting `DUBPLATE_SITE_NAME` follows
  the wire-first rule and belongs to rung 4, when a greenfield env
  render first consumes it; record that as rung 4 scope, do nothing
  now.
- [ ] **STATUS and HISTORY:** HISTORY gains the rung 1 entry (what
  landed, what the gate caught, what a later pass must not
  rediscover: the vale `_salvage` exclusion, the tidy prune, the
  additive deploy boundary). STATUS stays under 60 lines by moving,
  never deleting; it sits at 59 today and the line to reclaim is the
  stale "Open rulings 1-3 await Geoff" item (all three resolved in
  the ratified spec). Next action points at rung 2's plan.
- [ ] **Close:** gate; push branch and the `pre-greenfield` tag;
  pass-end STATUS.

**Acceptance criteria:** every CLAUDE.md claim about the gate and
stack is verifiable against the tree; settings carry no dead
permissions; ROADMAP names the initiative; STATUS ≤ 60 lines with the
stale rulings line resolved; tag pushed; gate green.

**Deliverable count: 4** (ROADMAP; CLAUDE.md and settings; template
note; ledgers and push). Docs only.

---

## Execution model

Workflow mode (owner opt-in): the pass-execute chain runs each task
through an implementer (model `sonnet`, dispatched with the task text
plus Global Constraints verbatim; invokes go-conventions before
writing Go), then `diff-reviewer` against the task's acceptance
criteria, then the gate. One re-dispatch on `fix`; a second `fix`
escalates to the conductor. Order: Task 1, then Tasks 2 and 3 (3 may
start while 2 is in review; they touch disjoint files), then 4, then
5. `code-simplifier` runs on each code task's changed Go before its
final commit (Task 1: scoped to `main.go` only). The runaway guard is
armed at launch per the workstation rule.

Budget: 5M ceiling, spend scored at close. Attended interaction
target: zero between launch and the pass-end report.
