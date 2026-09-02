# Rung 1 (Foundation) Implementation Plan

> **For agentic workers:** execute per this repo's conducting rules: one
> implementer-diff-reviewer-gate chain per task, dispatched with the
> Agent tool (4 tasks, below the workflow threshold). The conductor
> consumes reports and never reads diffs inline.

**Goal:** Pivot the repo to the greenfield tree and land the config
core, the `dubplate-server` skeleton, and a truthful gate, per the
build spec's rung 1.

**Architecture:** One mechanical pivot commit freezes the pre-pivot Go
and web trees under `_salvage/` (excluded from the toolchain and the
gate) and repoints build, gate, and deploy plumbing at
`cmd/dubplate-server`. The greenfield `internal/config` then lands the
derived wall's mechanism: collected required-value errors, unknown- and
superseded-key load errors that name replacements, and typed threshold
parsing. `doctor` and `check` scaffolds make the wall observable.

**Tech Stack:** Go 1.27, cobra, module `github.com/glw907/dubplate`.
No new dependencies.

**Spec:** `docs/superpowers/specs/2026-09-02-greenfield-server-build.md`
(revision 2.1, ratified). Rung 1's entry in "The ladder" and the
sections "Repo pivot and tree shape" and "The config wall, derived"
govern this plan.

**Token ceiling:** 5M (an M rung). **Checkpoint:** write STATUS after
Task 2, and at any split or question.

## Global Constraints

- Invoke the `go-conventions` skill before writing any Go. The bar is
  exemplary, not conforming (spec, "The Go bar"); reviewers fail
  awkward-but-working code.
- Universality: no hostname, path, email, or personal name outside
  `internal/config` (spec, "Universality"). Nothing estate-specific
  enters the greenfield tree.
- `_salvage/` is read-only reference material from the moment it
  exists. Nothing imports from it; no file in it is edited.
- The gate is `bash scripts/check.sh`; every task ends green. Never
  commit red. A check either gates or does not exist—no
  `command -v || skip` guards on new checks.
- No box access of any kind in this rung (spec, "Box safety").
- Table-driven tests under assertion discipline: no silent-success
  fakes, no vacuous assertions; prove each test can fail.
- Prose (ROADMAP, STATUS, HISTORY, comments) follows writing-voice:
  Go Doc Comments for code, the Google developer-doc register for
  docs. Em dashes are banned in code comments.
- Commits: imperative mood, specific files (never `git add -A`),
  `Co-Authored-By: Claude <noreply@anthropic.com>` footer.

---

### Task 1: The pivot commit

**Files:**
- Move (git mv): `cmd/` → `_salvage/cmd/`, `internal/` →
  `_salvage/internal/`, `web/` → `_salvage/web/`, `e2e/` →
  `_salvage/e2e/`
- Delete: `config/filebrowser.yaml` (its integration retired
  2026-09-02; the file is the orphan)
- Create: `cmd/dubplate-server/main.go`
- Modify: `Makefile`, `scripts/check.sh`, `scripts/vale-comments.sh`,
  `go.mod`, `go.sum`
- Test: the gate itself (`bash scripts/check.sh`) plus
  `_salvage`-exclusion probes below

**Interfaces:**
- Consumes: nothing (first task).
- Produces: a module whose only Go package tree is
  `cmd/dubplate-server`; the tag `pre-greenfield`; a `make
  build-linux` that cross-compiles `dubplate-server-linux-amd64` from
  `./cmd/dubplate-server` (rung 4's deploy depends on this name and
  target).

**Steps:**

- [ ] **Tag the pre-pivot state.** Annotated tag `pre-greenfield` on
  the current commit, message: "Last commit with the pre-greenfield
  tree buildable in place; emergency guardian fixes build from here."
  Local only; pushed at Task 4.
- [ ] **Move the four trees** with `git mv` so history follows. Verify
  `go build ./...` now fails only because no Go files remain outside
  `_salvage/`—then create the skeleton before committing.
- [ ] **Create `cmd/dubplate-server/main.go`:** a cobra root command
  named `dubplate-server`, short description from the spec's one-line
  purview ("organizes and serves one shared music library"), and a
  `version` subcommand printing `var version = "dev"` (ldflags wiring
  arrives when releases exist). Package doc comment per Go Doc
  Comments. No other subcommands yet—Task 3 adds `check` and
  `doctor`; the rest arrive with their rungs.
- [ ] **Rewrite `Makefile`:** `BINARY := dubplate-server`; `build`,
  `build-linux`, `install` point at `./cmd/dubplate-server`;
  `build-linux` output name `$(BINARY)-linux-amd64`; delete the four
  `web-*` targets and their comment block; `check: vet lint test
  tidy-check`; `clean` updated; `.PHONY` updated. The web legs rejoin
  at rung 7 per the spec—note this in a one-line comment on the
  `check` target.
- [ ] **Edit `scripts/check.sh`:** remove the npm prerequisite check
  and the web references in the header comment (the gate has no web
  module until rung 7); all other legs (shellcheck, vale, compose
  config, systemd verify) stay exactly as they are.
- [ ] **Edit `scripts/vale-comments.sh`:** exclude `_salvage/` from
  the scanned file list. The script unions `git ls-files`-tracked and
  untracked `.go` files; the Go toolchain ignores underscore
  directories but git does not, so without the exclusion every
  salvage comment re-enters the vale gate. Acceptance probe: run the
  script with a deliberate em dash planted in a `_salvage/` comment
  (unstaged, then reverted) and confirm no finding.
- [ ] **Tidy the module:** `go mod tidy` (dependencies prune to cobra
  and its transitive set); commit the shrunken `go.mod`/`go.sum` so
  `tidy-check` stays green.
- [ ] **Audit the live-ops keep set:** `grep -ri filebrowser scripts/
  systemd/ cloud-init/ compose.yaml config/ env/` must return nothing
  after the delete. If it returns hits, remove them in this task and
  list them in the report (the wind-down claimed they were gone;
  verify rather than trust).
- [ ] **Repoint `scripts/deploy.sh`'s artifact references:** it runs
  `make build-linux` and then handles the built artifact by name.
  Update every reference from the old `dubplate-linux-amd64` artifact
  to `dubplate-server-linux-amd64` so the deploy path builds and
  ships the greenfield binary from rung 1 onward (the spec's pivot
  section requires this for rung 4). The box's existing units and
  their installed binary are untouched until rung 4 and later; only
  the artifact plumbing moves now. No deploy runs in this rung.
- [ ] **Run the gate:** `bash scripts/check.sh` exits 0. Also verify:
  `go build ./...` succeeds; `go vet ./...` reports nothing in
  `_salvage/`; `./dubplate-server version` (after `make build`)
  prints `dev`.
- [ ] **Commit** (one commit, the pivot): moves, delete, skeleton,
  Makefile, both scripts, go.mod/go.sum.

**Acceptance criteria:** gate green; `git tag -l pre-greenfield`
non-empty; no Go package outside `cmd/dubplate-server` visible to
`go list ./...`; `make build-linux` produces
`dubplate-server-linux-amd64`; the filebrowser grep is empty; vale
does not scan `_salvage/`.

**Deliverable count: 4** (tag + move + audit; skeleton; build/gate
rewiring; vale exclusion). At the pass-sizing call-out threshold;
nothing may be added to this task after dispatch.

---

### Task 2: The config core

**Files:**
- Create: `internal/config/config.go`, `internal/config/parse.go`
- Test: `internal/config/config_test.go`, `internal/config/parse_test.go`

**Interfaces:**
- Consumes: nothing from Task 1 beyond the module layout.
- Produces (Task 3 and later rungs build against these exactly):

```go
package config

// Config is the resolved configuration of one dubplate-server
// process. Rung 1 carries the owned wall only; integration blocks
// join with their rungs.
type Config struct {
    LibraryRoot string
    StateDir    string
    SiteName    string
    AdminEmail  string
}

// Load resolves configuration from environ (os.Environ() form,
// "KEY=value"). It reads only DUBPLATE_-prefixed keys and returns an
// error that names every problem at once: each missing required
// value, each unknown key, each superseded key with its replacement,
// and each malformed value.
func Load(environ []string) (*Config, error)
```

- `parse.go` holds the threshold helpers later rungs use, exact
  shapes: `lookup(environ []string, key string) (string, bool)`,
  `parseDuration(environ []string, key string, def time.Duration)
  (time.Duration, error)`, `parseInt(environ []string, key string,
  def int) (int, error)`. Unset keys return the default; malformed
  values return an error naming the key and the accepted form.

**Behavior contract (write the tests from this table first):**

| Case | environ (DUBPLATE_ entries) | Expected |
|---|---|---|
| all present | LIBRARY_ROOT, STATE_DIR, SITE_NAME, ADMIN_EMAIL set | Config populated, nil error |
| one missing | ADMIN_EMAIL absent | error naming `DUBPLATE_ADMIN_EMAIL`, nothing else |
| several missing | only SITE_NAME set | one error naming all three missing keys |
| bad email | ADMIN_EMAIL=`not-an-address` | error naming the key; parsed with `net/mail.ParseAddress` |
| unknown key | + DUBPLATE_FROBNICATE=1 | error: unknown configuration key `DUBPLATE_FROBNICATE` |
| superseded key | + DUBPLATE_ACCESS_TEAM_DOMAIN=x | error naming the key and its replacement ("a forward-auth front is configured at the auth rung") |
| superseded key | + DUBPLATE_PUBLIC_NAVIDROME_URL=x | error naming the key; replacement is streamer driver config (rung 12) |
| non-namespace noise | PATH, HOME, EDITOR present | ignored entirely |
| empty value | DUBPLATE_STATE_DIR="" | treated as missing (required means non-empty) |

The superseded map is a package-level table seeded with
`DUBPLATE_ACCESS_TEAM_DOMAIN`, `DUBPLATE_ACCESS_AUD`, and
`DUBPLATE_PUBLIC_NAVIDROME_URL`; each entry carries the replacement
text used in the error. This is spec requirement U5: unknown or
superseded keys are load errors that name replacements, never a
degraded start.

**Steps:**

- [ ] Write the table-driven tests for `Load` from the contract table;
  run them; confirm every case fails (no implementation yet).
- [ ] Implement `Load` minimally against the table; errors are
  collected (one `errors.Join` or a single multi-line error—pick
  one, document it, and keep it; Task 3's doctor output relies on the
  error being complete rather than first-fail).
- [ ] Write and satisfy the `parse.go` tests: default-on-unset,
  parse-on-set, error-on-malformed for both helpers.
- [ ] Package doc comment: the wall's classification table from the
  spec (owned set; integration classes with their flags, each marked
  "arrives at rung N"), so the config package itself is the wall's
  contract of record.
- [ ] Gate; commit.

**Acceptance criteria:** all contract-table cases pass and are proven
capable of failing; the package doc carries the classification table;
gate green.

**Deliverable count: 3** (Load + errors; parse helpers; wall table
doc).

---

### Task 3: `check` and `doctor` scaffolds

**Files:**
- Create: `cmd/dubplate-server/check.go`, `cmd/dubplate-server/doctor.go`
- Test: `cmd/dubplate-server/check_test.go`, `cmd/dubplate-server/doctor_test.go`
- Modify: `cmd/dubplate-server/main.go` (register the two commands)

**Interfaces:**
- Consumes: `config.Load(environ []string) (*Config, error)` exactly
  as Task 2 defines it.
- Produces: `newCheckCmd(environ []string, out io.Writer)
  *cobra.Command` and `newDoctorCmd(environ []string, out io.Writer)
  *cobra.Command`, the constructor shape every later subcommand
  follows (environ and writer injected for testability; `main` passes
  `os.Environ()` and `os.Stdout`).

**Behavior contract:**

- `check`: machine-facing startup-precondition gate. Loads config;
  on success prints `configuration: ok` and exits 0; on failure
  prints the load error to stderr and exits 1. Later rungs append
  preconditions; rung 1 has exactly this one.
- `doctor`: human-facing diagnostic. Prints one line per owned wall
  key—`DUBPLATE_LIBRARY_ROOT  ok` / `missing` / `invalid: <reason>`
 —then one line per known integration class from the wall table,
  each reading `disabled (arrives at rung N)` in rung 1. Exit 0 when
  the wall is satisfied, 1 otherwise. Doctor grows per rung (the spec
  assigns it collision checks, mode reporting, and preflights later);
  this scaffold fixes its output shape: aligned key, status word,
  optional reason.

**Test cases (in-process command execution, asserting on the injected
writer and the returned error):**

| Command | environ | Assert |
|---|---|---|
| check | full valid wall | out contains `configuration: ok`; RunE returns nil |
| check | missing STATE_DIR | RunE returns error mentioning `DUBPLATE_STATE_DIR` |
| doctor | full valid wall | four `ok` lines, every integration line says `disabled`; nil error |
| doctor | missing two keys | those two lines read `missing`; error non-nil |
| doctor | bad ADMIN_EMAIL | its line reads `invalid:` with a reason; error non-nil |

**Steps:**

- [ ] Tests first from the table; confirm failures.
- [ ] Implement both constructors; register in `main.go`.
- [ ] Gate; commit.

**Acceptance criteria:** the five cases pass; `dubplate-server doctor`
run by hand with an empty environment prints the full missing-wall
report and exits 1; gate green.

**Deliverable count: 3** (check; doctor; constructor pattern).

---

### Task 4: Ledgers, ROADMAP, and reconciliation

**Files:**
- Modify: `ROADMAP.md`, `docs/STATUS.md`, `docs/HISTORY.md`,
  `CLAUDE.md`, `env/dubplate.env.template` (header note only)

**Interfaces:** none; documentation and close-out.

**Steps:**

- [ ] **ROADMAP:** restate the initiative queue for the pivot per the
  spec's rung 1 deliverable: the greenfield server build (spec,
  17-rung ladder) is the active initiative; the organizer-spec
  roadmap revisions travel with it; packaging stays sequenced after.
  Fold the organizer spec's "Roadmap revisions this spec carries"
  list, which has been parked since 2026-09-01.
- [ ] **CLAUDE.md reconciliation, verified claim by claim:** the
  Shell stack-rules block describes a bats suite deleted 2026-08-31—
  remove it; the gate description must list what `scripts/check.sh`
  actually runs after Task 1 (make check without web legs until rung
  7, shellcheck, vale, compose config, systemd verify). The "gate
  runs gitleaks" claim: locate the mechanism (`grep -ri gitleaks
  .claude/ scripts/ Makefile .git/hooks/`); if it exists, name where;
  if it does not, remove the claim and say so in the task report
  rather than silently. Point the salvage-tier language at
  `_salvage/` and the `pre-greenfield` tag.
- [ ] **`env/dubplate.env.template`:** header gains two lines: the
  manifest describes the frozen pre-pivot binary's wall
  (`_salvage/internal/config`); the greenfield wall is
  `internal/config`'s package doc, and this manifest evolves at rung
  4 when the first greenfield deploy needs rendering.
- [ ] **STATUS and HISTORY per the ledger discipline:** HISTORY gains
  the rung 1 entry (what landed, what the gate caught, what a later
  pass must not rediscover—the vale `_salvage` exclusion and the
  tidy prune belong in that last clause); STATUS stays present-tense,
  under 60 lines, next action pointing at rung 2's plan.
- [ ] **Workstation chore:** the age-store item
  `DUBPLATE_SERVE_SITE_NAME` ("907 Music") serves the retired serve
  wall's name. Mint `DUBPLATE_SITE_NAME` with the same value via
  `secret-set.sh`, mark the old name for deletion at the next
  registry sweep (the FILEBROWSER_ADMIN_PASSWORD pattern already in
  STATUS), and record both in the registry.
- [ ] **Close:** run the gate, push branch and the `pre-greenfield`
  tag, write the pass-end STATUS.

**Acceptance criteria:** every CLAUDE.md claim about the gate and
stack is verifiable against the tree; ROADMAP names the 17-rung
initiative; STATUS ≤ 60 lines; tag pushed; gate green.

**Deliverable count: 4** (ROADMAP; CLAUDE.md; ledgers; chores). Docs
only; no code.

---

## Execution model

Four tasks, sequential, one chain each: implementer (general-purpose
agent, model `sonnet`, dispatched with the task text plus Global
Constraints verbatim; it invokes go-conventions before writing Go),
then `diff-reviewer` (model per its definition) against the task's
acceptance criteria, then the gate inside the chain. One re-dispatch
on a `fix` verdict; a second `fix` escalates to the conductor. Domain
reviewer fan-out at pass end is Go-only this pass (no web, no auth
surface yet): `code-simplifier` over the changed Go before the final
commit of each code task, per the repo's commit rules.

Budget: 5M ceiling, spend scored at close. Attended interaction
target: zero between plan approval and pass-end report.
