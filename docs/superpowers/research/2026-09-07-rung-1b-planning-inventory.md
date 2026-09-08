# Rung 1b Planning Inventory

Research report, 2026-09-07, prepared for rung 1b's plan authoring,
against HEAD `c98a916`. Scope, quoted from the parallel-approach spec's
Phase 0b paragraph
(`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:54`):

> **0b. Rung 1b: the config wall-row registry** (S band, its own plan,
> the standard implementer/`diff-reviewer`/gate chain). Extend rung 1's
> existing `Register()` threshold-registry pattern to wall rows and
> reserved classes, so each rung declares its keys in its own file and
> no lane ever edits `internal/config/keys.go`. This is the one file the
> whole build touches and the first thing two lanes collide on. It is a
> rung rather than an unplanned chore because it changes rung 1's
> ratified config-core contract, and contract changes get plans. Two
> riders: the work must re-fingerprint `.gitleaksignore`, whose single
> allowed finding is pinned by line number
> (`internal/config/keys_test.go:generic-api-key:66`) so any insertion
> above line 66 reds the gate for every lane at once, and it must note
> that fragility in the file.

The ownership row that governs the surface
(`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:158`):

> | The config wall | the registry (rung 1b); no lane edits `keys.go` |
> every rung declares in its own file |

Method: every claim below is read off the committed tree at `c98a916`,
off the installed toolchain, or off a reproducible offline probe. The
gitleaks probes in section 3 ran against scratch copies under
`/tmp/claude-1000/-var-home-glw907-Projects-dubplate/94aaf30d-ee85-423d-a329-c444340e4973/scratchpad/`,
never against the worktree, and the plan should re-run them rather than
trust them. No file in the repo was modified.

Eight sections follow, each closing with what 1b inherits concretely.
A surprises list closes the report.

---

## 1. `internal/config` as landed, file by file

Three source files and three test files, 1255 lines total: `config.go`
250, `keys.go` 254, `thresholds.go` 96, `config_test.go` 348,
`keys_test.go` 146, `thresholds_test.go` 161.

### 1.1 The five-class contract

The package doc (`internal/config/config.go:1-45`) is the ratified
contract 1b changes. It classifies every `DUBPLATE_*` name into exactly
one of five classes, and states that `Load` enforces the
classification rather than merely reading values.

1. Owned, required now (`config.go:9-17`): four keys plus the identity
   source row, which is owned from rung 1 and has no env key yet.
2. Reserved for a rung (`config.go:18-25`): setting one early is a load
   error naming its class and rung.
3. Adopter-optional integration (`config.go:26-32`): accepted only when
   the class's own `Flag` resolves true.
4. Superseded (`config.go:33-37`): a pre-pivot name, errored with its
   replacement or with the fact that it retired.
5. Unknown (`config.go:38-41`): anything else, refused rather than
   ignored, per evolution lesson U5.

`config.go:43-44` states the separation 1b has to reconcile: the
threshold registry is "a separate, open-ended registry", while the wall
is a closed table. 1b's job is to make the closed table open without
losing what closedness bought.

### 1.2 The wall's data model, every row

**Owned wall** (`keys.go:19-25`), type `WallRow` (`keys.go:8-14`) with
two fields, `Name` and `Pending`.

| Row | Required now | Parser | Default |
|---|---|---|---|
| `DUBPLATE_LIBRARY_ROOT` | yes | `wall.require`, raw string | none |
| `DUBPLATE_STATE_DIR` | yes | `wall.require`, raw string | none |
| `DUBPLATE_SITE_NAME` | yes | `wall.require`, raw string | none |
| `DUBPLATE_ADMIN_EMAIL` | yes | `wall.requireEmail`, RFC 5322 single address, lowercased | none |
| `identity source` | no | none | `Pending: "pending (arrives at rung 5)"` |

The identity-source row is not an env key. It is a table entry whose
`Name` is prose, and `ownedNow` (`keys.go:34-41`) returns false for it
because `Pending` is non-empty. Two consequences 1b must preserve:
`rejectNamespace` never treats it as owned, and `doctor` prints its
`Pending` text where other rows print a status.

`OwnedWall` (`keys.go:28-30`) returns `slices.Clone(ownedWall)`, proved
by `TestOwnedWall_ReturnsCopy` (`keys_test.go:50-57`).

**Reserved classes** (`keys.go:57-142`), type `ReservedClass`
(`keys.go:47-51`) with `Name`, `Rung`, `Keys`. Eight classes, 27 keys.

| Class | Rung | Keys |
|---|---|---|
| auth and identity | 5 | `MAGIC_LINK_SECRET`, `SESSION_SECRET`, `PUBLIC_URL`, `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM` |
| serve | 5 | `SERVE_BIND_ADDRS`, `SERVE_PORT` |
| backup targets | 6 | `R2_ACCOUNT_ID`, `R2_BUCKET`, `R2_REMOTE`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `BACKUP_STAGING` |
| intake | 7 | `UPLOADS_DIR`, `TRIGGER_DIR` |
| pipeline | 9 | `INBOX_DIR`, `STAGING_DIR`, `QUARANTINE_DIR` |
| monitoring | 9 | `PING_IMPORT_URL`, `PING_BACKUP_URL`, `PING_DISK_URL`, `PING_NAVIDROME_URL` |
| decision and review queues | 13 | `REVIEW_DIR` |
| streamer | 12 | `NAVIDROME_URL`, `NAVIDROME_ADMIN_USER`, `NAVIDROME_ADMIN_PASSWORD` |

Table order is not rung order: `decision and review queues` (rung 13)
precedes `streamer` (rung 12). `doctor` prints in table order
(`doctor.go:41-46`), so the ordering is user-visible, and
`TestDoctorCmd` walks `ReservedClasses()` in the same order
(`doctor_test.go:89-98`). A registry that sorts on registration order
changes that output.

`ReservedClass` carries no `Required`/`Optional` split, no per-key
documentation, and no flag. `IntegrationClass` carries all three. That
asymmetry is the first shape question 1b answers.

**Superseded keys** (`keys.go:164-183`), a `map[string]string` from
name to disposition message. Eleven entries: two renames
(`LIBRARY_DIR`, `SERVE_SITE_NAME`), three moves to another component's
config (`ACCESS_TEAM_DOMAIN`, `ACCESS_AUD`, `PUBLIC_NAVIDROME_URL`),
and six outright retirements (`CONTRIBUTORS_FILE`,
`REJECT_EMAIL_TEMPLATE`, `BEETS_CONFIG`, `BEETS_DB`, `NAVIDROME_DB`,
`TUNNEL_TOKEN`). Being a map, it has no order, and `supersessionOf`
(`keys.go:187-190`) is a single lookup. No caller enumerates it, so a
registry may keep the map shape.

**Integration classes** (`keys.go:214-229`), type `IntegrationClass`
(`keys.go:203-209`) with `Name`, `Rung`, `Flag`, `Required`,
`Optional`. One class today, the rung 4 shadow class. The doc comment
at `keys.go:211-213` already names 1b: "Rung 1b's registry absorbs this
table as data rather than re-deriving it (rung 4 spec, risk 6)."

### 1.3 The six `DUBPLATE_SHADOW_*` rows and required-when-enabled

| Key | Role | Semantics |
|---|---|---|
| `DUBPLATE_SHADOW_ENABLED` | `Flag` | `strconv.ParseBool`; unset is off; an unparseable value is an error, never a silent off (`config.go:155-165`) |
| `DUBPLATE_SHADOW_OBSERVE_DIR` | `Required` | required once the flag is set; lands in `ShadowIntegration.ObserveDir` |
| `DUBPLATE_SHADOW_ORACLE_STATE_DIR` | `Required` | as above, `OracleStateDir` |
| `DUBPLATE_SHADOW_REVIEW_DIR` | `Required` | as above, `ReviewDir` |
| `DUBPLATE_SHADOW_FLOOR` | `Optional` | a registered threshold, `KindDuration`, default `720h` (`shadow.go:23-29`) |
| `DUBPLATE_SHADOW_BUDGET` | `Optional` | a registered threshold, `KindInt`, default `60` (`shadow.go:38-42`) |

The `Required`/`Optional` split is a difference in what `Load` does,
never in what `rejectNamespace` accepts (`keys.go:199-202`). `Load`
requires every `Required` key when the flag is set. An `Optional` key
is accepted and never required, because its `Default` stands when it is
unset. Setting any row with the flag unset is a load error naming the
flag (`config.go:190-192`). The flag names itself, so `rejectNamespace`
skips it to avoid emitting two contradictory errors for one typo
(`config.go:185-189`).

`ShadowIntegration` (`config.go:80-88`) holds only the three required
directories. Its doc states the rule 1b must not break: the floor and
the cap are thresholds rather than fields, because `Config` carries no
defaulted value.

`TestLoadShadowIntegration` (`config_test.go:255`) is the proof of all
four behaviors and is the test 1b's migration must keep green
unchanged.

### 1.4 `Load`'s flow

`Load(environ []string) (*Config, error)` (`config.go:212-250`), in
order:

1. `newWall(environ)` (`config.go:102-112`) indexes `os.Environ()` form
   into a map, later entries winning for a duplicate name.
2. Four `require` calls build `Config`, one of them `requireEmail`
   (`config.go:215-220`). Each records `ErrMissing` per unset row
   rather than returning.
3. A loop over `integrationClasses` (`config.go:222-242`): resolve the
   flag, skip when off, `require` every `Required` key into a
   `map[string]string`, then a `switch`-equivalent on `class.Name ==
   "shadow"` that hand-builds `ShadowIntegration`.
4. `rejectNamespace()` (`config.go:173-205`) walks every
   `DUBPLATE_*`-prefixed name in the environment and classifies
   anything not owned-now.
5. `errors.Join(w.errs...)` returns every problem at once
   (`config.go:246-249`).

`wall` (`config.go:95-98`) is the single choke point: `get`
(`config.go:115-121`) owns last-wins, empty-equals-unset, and the
`=`-in-value decision, and every other accessor calls it. That choke
point is worth preserving verbatim; it is what lets `ParseInt` and
`ParseDuration` build their own throwaway `wall` from an
`environ` slice (`thresholds.go:56`, `thresholds.go:80`) with identical
semantics.

The hard-coded `class.Name == "shadow"` branch is the one place `Load`
is not data-driven. A second integration class means a second branch.

### 1.5 The `Threshold` type, its `Kind`, and `Register`

`ThresholdKind` (`thresholds.go:13-18`) has exactly two values,
`KindInt` and `KindDuration`. `KindInt` is `iota`'s zero, deliberately,
so a `Threshold` declared before the field existed keeps its meaning.

`Threshold` (`thresholds.go:24-29`) has four fields: `Key`, `Kind`,
`Default` (a string), `Doc` (the accepted form reported on a parse
error).

`registeredThresholds` (`thresholds.go:34`) is a package-level slice.
`Register` (`thresholds.go:40-42`) appends, with no duplicate check and
no key validation. Its doc states the deliberate choice 1b inherits: a
rung's own package calls `Register` "from an explicit setup call rather
than an init function, so registration order stays visible in the
caller's own code."

`Thresholds` (`thresholds.go:45-47`) returns a clone, proved by
`TestThresholds_ReturnsCopy` (`thresholds_test.go:40-51`).

`ParseInt` (`thresholds.go:55-72`) and `ParseDuration`
(`thresholds.go:79-96`) are the only two parsers. Both distinguish a
bad environment value (an error naming `Key` and `Doc`) from a bad
`Default` (an error naming `Key` alone, called a registration bug).
`TestParseInt_InvalidDefaultIsRegistrationError`
(`thresholds_test.go:109`) and
`TestParseDurationRejectsABadDefaultAsARegistrationBug`
(`thresholds_test.go:156`) hold that line.

Because `Register` appends without dedup, the registry is package-level
mutable state that tests have to save and restore. `withRegistry`
(`thresholds_test.go:12-17`) does exactly that with `t.Cleanup`.

### 1.6 `Register`'s one caller and the `sync.OnceFunc` it forced

There is one `config.Register` call site in the module:
`registerShadowThresholds` in `cmd/dubplate-server/shadow.go:46-52`. It
is wrapped in `sync.OnceFunc`, and its comment names the reason
(Resolution R12): `config.Register` appends without dedup, and
`newRootCmd` runs more than once in a test binary.
`TestNewRootCmd_RegistersShadowThresholdsOnce`
(`cmd/dubplate-server/main_test.go:36-58`) is the regression test for
that exact defect. `newRootCmd` calls it at `main.go:55`, before any
`AddCommand`.

`doctor_test.go:159` calls `registerShadowThresholds()` directly, so
the threshold rows exist for the doctor assertions regardless of test
order.

### 1.7 Every exported symbol and its callers across the module

Callers outside `internal/config`, at HEAD:

| Symbol | Kind | Callers |
|---|---|---|
| `ErrMissing` | var | `doctor.go:99` (`errors.Is` to distinguish "missing" from "invalid") |
| `Config` | type | `doctor.go`, `check.go`, `shadow.go:196,262,299` |
| `ShadowIntegration` | type | reached through `Config.Shadow` in `shadow.go` |
| `Load` | func | `check.go:23`, `doctor.go:31`, `shadow.go:196,262,299`, `check_test.go:107` |
| `WallRow` | type | `doctor.go:34` (range), `doctor_test.go:70` |
| `OwnedWall` | func | `doctor.go:34`, `doctor_test.go:70` |
| `ReservedClass` | type | `doctor.go:41` (range), `doctor_test.go:89` |
| `ReservedClasses` | func | `doctor.go:41`, `doctor_test.go:89` |
| `IntegrationClass` | type | `doctor.go:48` (range), `doctor_test.go:100` |
| `IntegrationClasses` | func | `doctor.go:48`, `doctor_test.go:100` |
| `IntegrationClass.Enabled` | method | `doctor.go:50`, `doctor_test.go:101` |
| `Threshold` | type | `shadow.go:23,38` (two literals), `doctor.go:111` (parameter) |
| `ThresholdKind`, `KindInt`, `KindDuration` | type, consts | `shadow.go:25,40`, `doctor.go:113` |
| `Register` | func | `shadow.go:50-51` |
| `Thresholds` | func | `doctor.go:58`, `doctor_test.go:115`, `main_test.go:44,47` |
| `ParseInt` | func | `doctor.go:121`, `shadow.go:210` |
| `ParseDuration` | func | `doctor.go:114`, `shadow.go:207,266` |

No package under `internal/` imports `internal/config`.
`internal/engine/matcher/tunables.go:49-56` names the four knobs it will
eventually register and states plainly that it "imports nothing from
internal/config".

### What 1b inherits, concretely

A 1255-line package with a five-class contract in its package doc, four
tables of four different shapes (`[]WallRow`, `[]ReservedClass`,
`map[string]string`, `[]IntegrationClass`), one open registry over a
fifth shape (`[]Threshold`), and eighteen exported symbols with
seventeen call sites, all of them in `cmd/dubplate-server`. One
`Register` call site exists, and it is already wrapped in
`sync.OnceFunc` to work around the registry's missing duplicate check.

---

## 2. Every place a rung declares or will declare keys

### 2.1 Today's `Register` callers

Exactly one: `cmd/dubplate-server/shadow.go:46-52`, registering two
thresholds. Rung 4 declared the shadow *class* the other way, by
editing `keys.go:214-229` directly, under the grant recorded at
`docs/superpowers/specs/2026-09-05-rung-4-shadow.md:1650`.

### 2.2 The doctor rows

`newDoctorCmd` (`cmd/dubplate-server/doctor.go:24-72`) prints four
blocks in a fixed order, every row derived from `internal/config`'s own
tables so no rung hard-codes a count:

1. Owned wall (`doctor.go:34-39`): `cmp.Or(row.Pending,
   statuses[row.Name], "ok")`.
2. Reserved classes (`doctor.go:41-46`): `"disabled (arrives at rung
   %d)"`.
3. Integration classes (`doctor.go:48-56`): `"enabled"` or
   `"disabled"`, from `class.Enabled(environ)`, with a parse error
   rendering as `"disabled"`.
4. Registered thresholds (`doctor.go:58-62`): the effective value from
   `thresholdStatus`.

Every line is `fmt.Fprintf(w, "%-32s %s\n", key, status)`
(`doctor.go:74-78`). That 32-column pin is a contract with
`doctor_test.go` and with anything reading the output on the box.

`thresholdStatus` (`doctor.go:111-128`) switches on `Kind` with
`KindDuration` as the only named case and `ParseInt` in `default`. A
third kind added without touching this switch silently parses as an
integer.

`rowStatuses` (`doctor.go:84-106`) type-asserts the joined error to
`interface{ Unwrap() []error }` and splits each message on its first
`": "` (`doctor.go:134-139`). Every registry error message must keep
the `key: reason` shape or the doctor's per-row statuses go blank.

`TestDoctorCmd` (`doctor_test.go:12`) derives its expectations from the
same tables the command reads, which makes it a structural test rather
than a value test. `TestDoctorCmd_LiteralSpotCheck`
(`doctor_test.go:137`) and `TestDoctorCmd_ShadowRows`
(`doctor_test.go:158`) are the hard-coded antidotes, asserting
`DUBPLATE_ADMIN_EMAIL ok`, the identity-source pending text,
`DUBPLATE_SHADOW_FLOOR 720h0m0s`, `DUBPLATE_SHADOW_BUDGET 60`, and the
shadow class's `disabled`/`enabled` transition.

### 2.3 The env template and how `deploy.sh` consumes it

`env/dubplate.env.template` assigns 42 names on their own lines. Its
header (`:1-3`) says outright that it describes the frozen pre-pivot
binary's wall, not `dubplate-server`'s, and that it evolves at rung 4.

Classified against `keys.go` at HEAD:

| Classification | Count | Names |
|---|---|---|
| Owned now | 2 | `ADMIN_EMAIL`, `STATE_DIR` |
| Reserved | 26 | the five SMTP, five R2, three Navidrome, `INBOX_DIR`, `STAGING_DIR`, `REVIEW_DIR`, `QUARANTINE_DIR`, `BACKUP_STAGING`, `TRIGGER_DIR`, `UPLOADS_DIR`, `SERVE_BIND_ADDRS`, `SERVE_PORT`, four PING |
| Superseded | 11 | `CONTRIBUTORS_FILE`, `REJECT_EMAIL_TEMPLATE`, `BEETS_CONFIG`, `BEETS_DB`, `NAVIDROME_DB`, `LIBRARY_DIR`, `ACCESS_TEAM_DOMAIN`, `ACCESS_AUD`, `PUBLIC_NAVIDROME_URL`, `SERVE_SITE_NAME`, `TUNNEL_TOKEN` |
| Not `DUBPLATE_`-prefixed | 3 | `ND_PASSWORDENCRYPTIONKEY`, `ND_SPOTIFY_ID`, `ND_SPOTIFY_SECRET` |

`DUBPLATE_LIBRARY_ROOT` and `DUBPLATE_SITE_NAME`, two of the four
owned-now rows, appear nowhere in the template.

`deploy.sh` consumes it as a manifest of names, never of values
(`scripts/deploy.sh:171-194`). The loop skips blank and comment lines,
takes `name="${line%%=*}"`, resolves `"${!name:-}"` from the shell
(populated by sourcing `~/.local/secrets` at `:93-96`), collects every
unresolved name, and exits nonzero listing all of them at once. The
template path is a single constant at `scripts/deploy.sh:26`. Adding a
name to the template therefore requires minting the secret, which is
CLAUDE.md's stated fact and is confirmed by the code.

A gate-resident test couples the two files.
`TestLoad_TemplateKeysClassified` (`internal/config/config_test.go:196`)
opens `../../env/dubplate.env.template` via `templateKeys`
(`config_test.go:218-247`) and fails any `DUBPLATE_*` name that
classifies as unknown. `internal/config`'s test binary therefore reads
`env/`. Any registry that changes how a name is classified changes what
that test accepts, and rung 4b's planned second template
(`docs/superpowers/specs/2026-09-05-rung-4-shadow.md:1632` names
`env/dubplate-server.env.template`) forces a decision about which file
the test reads.

### 2.4 The salvage tree, read-only

`_salvage/internal/config/` is 1230 lines across four files. It has no
registry of any kind. `Config` (`_salvage/.../config.go:24-50`) is a
flat struct of nested structs, and `Load`
(`_salvage/.../config.go:252-355`) is one long function with local
closures `require`, `requirePort`, and `requireList`. Four ideas are
worth an explicit verdict:

- The `get` choke point (`_salvage/.../config.go:183-190`). Its comment
  states a seam: "the packaging initiative can later overlay a
  non-secret config file by changing get alone, with no call site
  touched." **Adopt the idea, note it is already landed.** The
  greenfield `wall.get` (`config.go:115-121`) is the same choke point
  with stricter semantics. 1b's registry must keep every parser routed
  through it, so the packaging initiative's file overlay stays a
  one-function change.
- The typed optional readers `optInt`, `optInt64`, `optDuration`, and
  `optString` (`_salvage/.../config.go:191-241`). **Decline the shape,
  adopt the coverage gap it exposes.** The greenfield replaced them
  with `Threshold` plus two parsers, which is the better shape because
  the default and its documentation travel with the key. The salvage
  set proves three parsers are wanted that the greenfield lacks:
  `int64`, `string`, and (from `requireList`) a comma-separated list.
- `pingURLsFromEnv` (`_salvage/.../config.go:409-426`), a *pattern* key
  family, `DUBPLATE_PING_<SLUG>_URL`, discovered by scanning the
  environment rather than declared. **Decline for 1b; record it as a
  shape the registry must eventually express.** The greenfield froze
  the same four ping keys as literal reserved names
  (`keys.go:116-125`), which loses the arbitrary-slug capability. A
  registry that only accepts exact names cannot express this family,
  and rung 9 owns the monitoring class.
- `loadCollections` (`_salvage/.../config.go:383-402`), reading
  `DUBPLATE_COLLECTIONS` as a comma-separated list with a
  single-unnamed-collection default. **Decline for 1b, flag for rung
  9.** `DUBPLATE_COLLECTIONS` is in no greenfield class, so an adopter
  who sets it today gets "unknown configuration key". CLAUDE.md's facts
  list still names collection routing as live behavior.

### 2.5 How many declaration sites the registry will actually serve

One exists today (`cmd/dubplate-server/shadow.go`). The projected total,
derived from the tables in section 1.2 rather than estimated, is **ten
places outside `internal/config`**: one per reserved class (eight, and
they land in eight different packages, because rung 5's two classes
split across auth and serve and rung 9's two split across pipeline and
notify), one for the shadow integration class, and one for the four
matcher knobs at `internal/engine/matcher`. Ten is the number the shape
decision in section 5 should be weighed against, and it is small enough
that a single visible wiring function in `cmd/dubplate-server` stays
readable.

### What 1b inherits, concretely

One `Register` call site, one `keys.go` edit already taken under a
grant, a four-block doctor whose column pin and block order are pinned
by tests, a 42-name env template describing a wall that no longer
exists, a gate-resident test that couples `internal/config` to `env/`,
and a salvage package with no registry but three concrete parser gaps
and two key-family shapes the greenfield table cannot express.

---

## 3. The gitleaks fragility

### 3.1 What the finding actually is

`.gitleaksignore` is one line:

```
internal/config/keys_test.go:generic-api-key:66
```

`scripts/check.sh:98-109` runs `gitleaks detect --no-git -s . --redact`
and calls it "the one known false positive (a config-key-name test
fixture that happens to look like a generic API key)"
(`check.sh:104-106`). The gate hard-fails when gitleaks is absent
(`check.sh:35-38`), so the leg genuinely gates.

Installed version: gitleaks 8.30.1 at
`/home/linuxbrew/.linuxbrew/bin/gitleaks`. The repo carries no
`.gitleaks.toml`, so the default rule set applies.

Reproduced with the ignore file redirected to an empty directory, the
finding is exactly one:

```
RuleID:      generic-api-key
File:        internal/config/keys_test.go
StartLine:   66   EndLine: 66   StartColumn: 5   EndColumn: 29
Match:       key: "DUBPLATE_R2_BUCKET"   gitleaks:allow
Secret:      DUBPLATE_R2_BUCKET
Entropy:     3.6143694
Fingerprint: internal/config/keys_test.go:generic-api-key:66
```

Line 66 is the first row of `TestReservedClassOf`'s table
(`keys_test.go:66`):

```go
{key: "DUBPLATE_R2_BUCKET", wantClass: "backup targets", wantRung: 6, wantFound: true}, // gitleaks:allow
```

The trigger is the struct field named `key` sitting on one line beside
a string whose Shannon entropy clears the rule's threshold. The rule
needs both halves.

### 3.2 The trigger is one name, measured

A scratch file was built holding every `DUBPLATE_*` name that appears
in `keys.go`, `env/dubplate.env.template`, or `shadow.go`, 54 names,
each written as `{Key: "NAME"},` on its own line. gitleaks reports
exactly one finding across all 54: `DUBPLATE_R2_BUCKET`.

Two facts follow. The gate's comment overstates the problem, because
no other config key name in the namespace trips the rule. And a 1b
registry that writes `{Key: "DUBPLATE_R2_BUCKET", ...}` on one line in <!-- gitleaks:allow -->
a rung's own declaration file *will* create a new finding at that
file's line, because the `Key:` field name supplies the same keyword
the test's `key:` field supplied.

### 3.3 What the fingerprint format allows

gitleaks 8.30.1 builds the fingerprint as `file:ruleID:startLine` under
`--no-git`, and `commit:file:ruleID:startLine` in git mode. Both forms
carry the line. There is no line-free fingerprint. `.gitleaksignore` is
a list of fingerprints, so a line-pinned ignore is the only thing that
file can express. The line dependence is a property of the mechanism,
never a choice the repo made badly.

The `--baseline-path` flag takes a prior report and suppresses
everything in it. It carries the same line dependence and adds a
generated artifact to the tree, so it is strictly worse here.

### 3.4 The options, each with its measured cost

**Option A. Rename the test's struct field.** Renaming `key` to
`envName` in `TestReservedClassOf`'s table removes the keyword, and the
finding disappears entirely. Verified: a scratch copy with the field
renamed scans clean, and `.gitleaksignore` becomes empty. Cost: one
mechanical rename across `keys_test.go:60-91`, no new config file, and
no suppression to audit. Risk: the fix is invisible, so a later author
can reintroduce the field name and reintroduce the finding with no
explanation nearby. This is answered by a comment, which the rider
already requires.

**Option B. An inline `// gitleaks:allow` comment.** Verified clean on
a scratch copy with the comment appended to line 66. Cost: the
suppression lives on the exact line it suppresses, so it moves with the
line and no fingerprint needs re-computing. It also survives an
insertion anywhere. Risk: it suppresses that line permanently,
including a future real secret written on it, and `--ignore-gitleaks-allow`
is the only way to audit it. This report uses the mechanism on its own
three example lines, which is how it stays green while quoting the
finding verbatim.

**Option C. A repo `.gitleaks.toml` allowlist.** Both a rule-scoped
allowlist and a global `[[allowlists]]` block were verified clean.
This option has a measured trap. With the default `condition`, which is
`OR`, an allowlist combining `paths` and `regexes` allows a finding
matching *either*. A global allowlist with
`regexes = ['''DUBPLATE_[A-Z0-9_]+''']` and no `condition` was measured
to suppress a genuine leak: a file containing
`DUBPLATE_MAGIC_LINK_SECRET=<a real high-entropy value>` scanned clean
under it and flagged without it. Adding `condition = "AND"` alongside
`paths = ['''internal/config/.*_test\.go''']` restores the correct
behavior, verified: the test file is allowed and the leaking file is
still flagged. Cost: a new config file, a subtle keyword whose absence
silently widens the suppression across the whole repo, and a path
regex matched against paths relative to `-s`, which is `.` in the gate.

**Option D. Move the finding to the file's end.** This is the option
the rider's phrasing implies. It reduces the exposure without removing
it, because an insertion below the moved row still shifts it. It also
fights the reason `TestReservedClassOf` puts `DUBPLATE_R2_BUCKET`
first, which is table readability.

**Recommendation for the plan to weigh.** Option A plus a comment
naming why the field is not called `key` is the smallest change that
makes the gate insertion-proof and leaves no suppression to audit. It
also generalizes: whatever shape 1b gives declaration files, the same
rename discipline keeps `Key:` literals off the same line as an
entropy-clearing name. Option C is the only option that also protects
future declaration files automatically, and it is the only option with
a measured way to be silently wrong.

### 3.5 The blast radius, precisely

The parallel spec says an insertion above line 66 "reds the gate for
every lane at once". That is right about the consequence and worth
narrowing: the fingerprint names one file, so only an insertion above
line 66 *of `internal/config/keys_test.go`* breaks it. Every lane feels
it because `scripts/check.sh` is the whole build's one gate.

Rung 4's staleness pre-task already treats this as a standing check:
"Is `.gitleaksignore:1` still
`internal/config/keys_test.go:generic-api-key:66`, and is line 66 still
the row it names?"
(`docs/superpowers/specs/2026-09-05-rung-4-shadow.md:1693-1695`).
Whatever 1b lands should retire that question rather than restate it.

### What 1b inherits, concretely

A one-line ignore file pinned to `keys_test.go:66`, a gate comment that
overstates the finding's generality, four options of which three were
verified clean on scratch copies, one measured over-suppression trap in
the config-file option, and a standing staleness question in rung 4's
spec that a durable fix should retire.

---

## 4. The follow-ons the carries name, and whether 1b lands them

### 4.1 The four-knob wall registration

HISTORY's definition, quoted (`docs/HISTORY.md:1434-1439`):

> 2. **The four-knob wall registration** stays the named follow-on task
>    for the first rung 3 pass that runs after rung 1b closes:
>    `StrongRecThresh`, `MediumRecThresh`, `RecGapThresh`,
>    `SearchLimit`, `internal/config/keys.go` the file no lane edits
>    until then, and `Threshold.Default` is a string with `ParseInt` the
>    only parser, so the three float rec thresholds need a float variant
>    the registry does not yet have.

The four knobs live in `internal/engine/matcher/tunables.go`:
`StrongRecThresh` (`:62-66`, default 0.04), `MediumRecThresh`
(`:68-72`, default 0.25), `RecGapThresh` (`:74-78`, default 0.25),
`SearchLimit` (`:138-147`, default 5). Their defaults are set at
`tunables.go:268-270` and `:309`. The package doc at `tunables.go:49-56`
states the deferral: "This pass names them and defers registration; it
imports nothing from internal/config."

**Verdict: 1b does not land it.** HISTORY calls it the follow-on for
the first rung 3 pass *after* 1b closes, three times
(`:434`, `:1063`, `:1434`), and `docs/STATUS.md:41-43` carries it as an
open item on the same terms. Registering the knobs would also make
`internal/engine/matcher` import `internal/config`, which no
`internal/` package does today.

### 4.2 `KindFloat`

HISTORY's 4a-ii forward carry 3 (`docs/HISTORY.md:434-438`):

> **The four-knob wall registration** stays the named follow-on for the
> first pass after rung 1b, and **rung 4's own `KindFloat` need joins
> it**. `Threshold.Kind` now has two values and the three float rec
> thresholds want a third. `AdminEmail`'s dedicated contact row sits
> beside both.

**Verdict: 1b should land the mechanism, not the registration.** The
kind is a registry capability, and 1b is the registry pass. Landing it
touches `ThresholdKind` (`thresholds.go:13-18`), a new `ParseFloat`
beside the existing two, and `thresholdStatus`'s switch
(`doctor.go:111-128`), whose `default` branch currently makes an
unnamed kind parse as an integer. Landing the *four knobs* is rung 3's
follow-on, per 4.1. The plan should say which half it takes, because
the carries phrase both as one item.

### 4.3 `AdminEmail`'s dedicated contact row

The need, quoted (`docs/HISTORY.md:1470-1475`):

> 7c. **The `AdminEmail` User-Agent contact is interim** (R11).
> `Config.AdminEmail`'s own doc comment says adopters compare it and
> never display it, and the harness precedent this plan cites uses a
> dedicated `DUBPLATE_HARNESS_CONTACT` input instead. A dedicated
> contact row is the post-1b follow-on, beside the four-knob
> registration.

`Config.AdminEmail`'s doc (`config.go:68-72`) does say exactly that.

**Verdict: 1b does not land it, and should not.** It is a new owned
wall row, which means a new required key for every adopter, which is a
product decision rather than a registry one. HISTORY calls it
"post-1b". 1b's contribution is making the row cheap to add later, from
whichever package owns the User-Agent.

### 4.4 `integrationClasses` absorbed as data

HISTORY's carry 4 (`docs/HISTORY.md:439-441`):

> **Rung 1b absorbs `integrationClasses` as data**, not as a shape to
> invent, and the table's `Required`/`Optional` split is the semantics
> it must preserve.

`keys.go:211-213` says the same thing from the code side.

**Verdict: 1b lands it.** The `Required`/`Optional` semantics are
stated at `keys.go:199-202` and proved by `TestLoadShadowIntegration`
(`config_test.go:255`) and
`TestIntegrationClassesCoverEveryShadowKey` (`keys_test.go:125-146`).
Both must stay green with no assertion weakened.

Absorbing the table as data is only half the job. `Load`'s
`class.Name == "shadow"` branch (`config.go:235-241`) is the part that
is *not* data, and a second integration class needs a second branch
unless 1b gives a class a way to receive its own resolved values.

### 4.5 Scope reconciliation

The parallel spec's 0b sentence names two things: "wall rows and
reserved classes". It does not name integration classes, superseded
keys, or thresholds.

| Item | In the 0b sentence | In the carries | Verdict |
|---|---|---|---|
| Wall rows | yes | yes | 1b |
| Reserved classes | yes | yes | 1b |
| Integration classes | no | yes (`HISTORY:439`) and in code (`keys.go:212`) | 1b, by the carries |
| Superseded keys | no | no | 1b by necessity: it is the fourth branch of `rejectNamespace` and a name has to resolve somewhere |
| `KindFloat` mechanism | no | yes, as a post-1b need | 1b should take the mechanism; the plan states so |
| Four knobs registered | no | explicitly the first rung 3 pass after 1b | not 1b |
| `AdminEmail` contact row | no | explicitly post-1b | not 1b |

### What 1b inherits, concretely

Two follow-ons that are unambiguously 1b's (integration classes as
data, and by necessity the superseded map), one that is unambiguously
not (the four knobs), one that is unambiguously not but keeps being
named beside 1b's work (the contact row), and one genuinely split
between the two (`KindFloat`: mechanism now, registration later). The
plan has to state the split, because the carries phrase mechanism and
registration as a single item.

---

## 5. The Go bar and the shape of a compile-time key registry

### 5.1 The bar as written

`docs/superpowers/specs/2026-09-02-greenfield-server-build.md:200-207`:

> The bar is exemplary, not conforming. Stdlib idiom is the default
> vocabulary, exported surfaces read like stdlib APIs, and the ops seam
> is the in-repo shape to imitate. Reviewers are empowered to fail
> awkward-but-working code.

The `go-conventions` rules that bear directly on a registry: no
interface until a second consumer exists, no builder or functional
options, a plain config struct, no package-level `init()`, and
`MixedCaps` throughout. The existing `Register` doc
(`thresholds.go:36-39`) already honors the `init()` rule explicitly and
says why.

### 5.2 How the stdlib does key registries

Four shapes, read off Go 1.27.1 at
`/var/home/linuxbrew/.linuxbrew/Cellar/go/1.27.1/libexec/src`.

**`flag.FlagSet` (`src/flag/flag.go`, `(*FlagSet).Var`).** A *value*
type holds the map. `Var` panics on a redefinition: `panic(msg) //
Happens only if flags are declared with identical names`. The package
also ships a package-level `CommandLine` set and package-level
functions delegating to it, so both composition styles coexist behind
one type. This is the closest analogue to a config wall, because a flag
carries a name, a default, a usage string, and a parser, which is
exactly `Threshold`'s four fields.

**`database/sql.Register` (`src/database/sql/sql.go`).** A package-level
map behind a mutex, populated from driver packages' `init` via blank
import. It panics on a nil driver and panics on a duplicate name:
`panic("sql: Register called twice for driver " + name)`.

**`expvar.Publish` (`src/expvar/expvar.go:309`).** A package-level
`sync.Map` plus a sorted key slice. A duplicate name is
`log.Panicln("Reuse of exported var name:", name)`.

**`image.RegisterFormat` (`src/image/format.go`).** Appends to an
atomic slice with no duplicate check at all. This is the one stdlib
registry shaped like `config.Register` today, and it is shaped that way
because a duplicate format sniffer is harmless: the first matching
magic wins and a second is dead weight.

**`testing/fstest`.** Not a registry. `MapFS` is a test double and
`TestFS` is a conformance checker. It offers nothing to 1b, and the
plan should record that rather than reach for an analogy.

The pattern across the three real registries is unambiguous: a
duplicate key panics, loudly, at registration. `image` is the
exception, and its exception is licensed by a semantics config keys do
not share. Two rungs claiming `DUBPLATE_STATE_DIR` is precisely the
collision 1b exists to prevent, so silence is the wrong answer.

### 5.3 Which shape fits, and the trade-off

The choice is package-level `init` registration versus explicit
composition in `main`.

**Package-level `init` registration**, the `database/sql` shape. Each
rung's package declares its keys in an `init`, and `cmd/dubplate-server`
blank-imports the package. Cost: `go-conventions` forbids `init` for
this repo, `Load`'s result then depends on the import graph rather than
on visible code, and a test that imports `internal/config` alone sees
an empty wall. The repo already rejected this at rung 1
(`thresholds.go:36-39`).

**Explicit composition in `main`**, the `flag.NewFlagSet` shape. Each
rung's package exports its declarations as values, and one visible
function in `cmd/dubplate-server` assembles them. Cost: one wiring
function that every rung edits, which reintroduces a shared file, but a
much smaller and much less conflict-prone one than `keys.go`. A single
`AddCommand`-style block is already the precedent: `main.go:63-68`, and
the rung 4 spec's contested-surface table blesses it as "rung 1's
skeleton; each rung adds its own file and registers through it"
(`docs/superpowers/specs/2026-09-05-rung-4-shadow.md:1647`).

**The half-way point the repo actually built.** `registerShadowThresholds`
(`shadow.go:46-52`) is a package-level `sync.OnceFunc` called from
`newRootCmd` (`main.go:55`). It is neither `init` nor a value passed
explicitly, and the `OnceFunc` exists solely because `Register` cannot
refuse a duplicate. Fixing `Register` to reject a duplicate key would
retire the `OnceFunc`, retire its regression test
(`main_test.go:36-58`), and align the registry with all three stdlib
exemplars in one change.

A second consideration cuts toward a value type. Package-level mutable
state forced `withRegistry` (`thresholds_test.go:12-17`) into the test
file to save and restore the registry around each test. A registry held
in a value, assembled once and passed to `Load`, needs no such helper,
and `Load`'s current signature already takes its other input
(`environ`) as a parameter rather than reading a global.

### 5.4 Prior art from the 2.0 spec

CLAUDE.md requires naming which prior-art entries were consulted, with
adopt or decline and a reason. The 2.0 spec's prior-art table
(`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md:394` onward)
is almost entirely about streaming, tagging, and storage. Per entry:

- **`pressly/goose`.** Declined already, by the 2.0 spec itself, for
  schema migrations. **Not applicable to 1b.** It is a database
  migration runner and touches no configuration surface.
- **`modernc.org/sqlite`.** Adopted conditionally at rung 6. **Not
  applicable to 1b.** Its relevance is FTS5 and WAL, not config.
- **`golang.org/x/time/rate`.** Adopted for 2.0's auth rate limits.
  **Not applicable to 1b**, though its knobs will one day be registered
  thresholds.
- **Navidrome's declined dependency list.** Directly applicable, and
  the one entry that is. The 2.0 spec names `viper` among "Navidrome's
  remaining dependencies ... architecture 2.0 chose against or already
  covers". **Declined, and 1b should restate the decline explicitly**,
  because a config-registry rung is exactly where a configuration
  library gets reached for. The stdlib plus `strconv` plus
  `time.ParseDuration` covers every parser the wall has or wants.
- **OpenSubsonic specification, gonic, Navidrome, taglib, mewkiz/flac,
  wazero, Picard tables, chromaprint, x/image, mewkiz, Melodee.** **Not
  applicable.** None concerns process configuration.

No external library is proposed for 1b. The one adoption worth naming
is internal: `flag`'s panic-on-redefinition semantics, imitated rather
than imported.

### What 1b inherits, concretely

A bar that says exported surfaces read like stdlib APIs, three stdlib
registries that all reject a duplicate key loudly, one that does not
and whose license to be lax does not transfer, an in-repo registry
shaped like the lax one, and a `sync.OnceFunc` plus a regression test
that exist only to compensate. No external prior art applies, and one
config library is worth declining on the record.

---

## 6. The 2.0 forward hooks that touch config

Read from `docs/superpowers/specs/2026-09-07-streamer-2-0-design.md:354`
onward. Fourteen hooks. CLAUDE.md requires each rung plan to state per
task whether it lands a hook or is neutral.

| Hook | Rung | Touches a config registry design? |
|---|---|---|
| 1. Content hash on the catalog row | 6, 7 | No. Schema and gate. |
| 2. Folded sort key | 6, 9 | No. Schema and tagging. |
| 3. Art blob at admit | 6, 9 | No. |
| 4. Play history as timestamped events | 6 | No. |
| 5. A stable identity | 5, 6 | **Weakly.** Rung 5's identity-source selector is the wall's one pending owned row (`keys.go:24`), and the surrogate id lives in the store. The registry must let rung 5 declare the selector plus its required-when-selected block, which is a class shape with a discriminator rather than a boolean flag. |
| 6. The driver seam stays universal | 12 | **Yes, decisively.** Address, proxy, and provisioning are per-driver capabilities "a driver may decline to declare, because an in-process streamer has no address and needs no proxy". The `streamer` reserved class today hard-codes three Navidrome names (`keys.go:133-141`). A universal seam means the *driver*, not the wall, owns its key set, and a driver that declines a capability declares no key for it. A registry keyed on a fixed class table cannot express that; one where a package registers its own declaration can. |
| 7. Schema migrations | 6 | No. |
| 8. Collection ownership | 6 | No, though it interacts with the salvage `DUBPLATE_COLLECTIONS` gap noted in 2.4. |
| 9. A transaction model across catalog events | 6 | No. |
| 10. Live-file backup consistency | 6 | No. |
| 11. ReplayGain and R128 gain | 6, 9 | No. |
| 12. A portable layout | 7 | **Weakly.** The 255-byte component bound and the forbidden-character set are candidate thresholds, and neither is an int or a duration. |
| 13. Sort tags written at admit | 8, 9 | No. |
| 14. An ID3 version ruling | 8 | **Weakly.** "Record whether the library writes ID3v2.3 or v2.4" is a two-valued choice. Expressed as a threshold it needs a string or enumerated kind, which the `KindInt`/`KindDuration` pair cannot represent. |

Hook 6 is the one a 1b design must accommodate outright. Hooks 5, 12,
and 14 all push in the same direction: the kinds a threshold can take
and the shapes a class can take both need room to grow past what rung 4
needed.

### What 1b inherits, concretely

One binding constraint: the streamer's key set must be declarable by a
driver rather than fixed in a wall table, or rung 12's universal seam is
rework. Three softer signals that `ThresholdKind` wants more than a
third value and that a class wants more than a boolean flag.

---

## 7. Contested surfaces

### 7.1 The standing ownership rows

`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:158` gives
the config wall to 1b: "the registry (rung 1b); no lane edits
`keys.go`". The rung 4 spec amends it for the interim
(`docs/superpowers/specs/2026-09-05-rung-4-shadow.md:1650`):

> | `internal/config` (`keys.go`, `thresholds.go`) | rung 1b's registry
> when it lands; **until then rung 4 as the only live code lane**, by
> the grant this spec records | every rung declares its keys |

Rung 4a used that grant. The shadow integration class at
`keys.go:214-229` and the `IntegrationClass` type at `keys.go:203-209`
are both rung 4's edits to the file 1b exists to close.

### 7.2 Does pass 4b write `internal/config`?

`docs/superpowers/plans/2026-09-07-rung-4b-deploy.md` does not exist.
The plans directory ends at
`2026-09-05-rung-4a-shadow-runner.md`, and `docs/STATUS.md:25-27`
confirms: "Pass 4b, the rung 4 deploy, in the main tree. Its plan is
owed just-in-time and does not exist yet."

The rung 4 spec's pass table
(`docs/superpowers/specs/2026-09-05-rung-4-shadow.md:1632`) enumerates
4b's deliverables:

> `systemd/dubplate-shadow.{service,timer}`;
> `env/dubplate-server.env.template` with its literal rows and the
> existing template's header amendment; deploy.sh's six edits, the
> sixth being the two `-ldflags` stamps and the smoke-test assertion;
> `dubplate-server version` printing both stamps; the fresh-context
> unit read; the deploy sitting; the SELinux verifications; the
> recorded deploy date

`internal/config` is absent from that list. 4b touches `systemd/`,
`env/`, `scripts/deploy.sh`, `Makefile`, and `cmd/dubplate-server`'s
version command. The contested-surface row for that set
(`spec:1651`) gives `systemd/`, `scripts/deploy.sh`, and `env/` to rung
1's ownership with rung 4 adding "two units, one template, and six
deploy edits".

**Two real edges survive that clean separation.**

First, `env/`. 4b creates `env/dubplate-server.env.template` and amends
the existing template's header.
`TestLoad_TemplateKeysClassified` (`internal/config/config_test.go:196`)
reads `../../env/dubplate.env.template` by a hard-coded relative path
(`config_test.go:224`). If 1b restructures classification and 4b adds a
second template, the two passes have to agree about which file that test
reads and what it asserts. Whichever lands second owns the
reconciliation. In this window that is 1b.

Second, `cmd/dubplate-server`. 4b edits `main.go` for the version
stamps. 1b, under any registry design that composes explicitly, also
edits `main.go` at or near `newRootCmd` (`main.go:54-70`). The two edits
are in different functions, so a textual conflict is unlikely and a
semantic one is impossible. The row already covering this
(`spec:1647`) reads: "rung 1's skeleton; each rung adds its own file and
registers through it".

### 7.3 The ordering inversion

Rung 4's staleness pre-task
(`docs/superpowers/specs/2026-09-05-rung-4-shadow.md:1687-1690`) opens
with a blocker question:

> **Has rung 1b landed?** If yes, the integration class and the
> duration threshold register through its registry and never by editing
> `keys.go`, and ruling 1's option (a) is re-expressed in that
> registry's own terms.

This window runs 1b *after* 4b's unattended tasks, so that branch never
fires for 4b. The re-expression of ruling 1's option (a) into registry
terms therefore becomes 1b's own migration work rather than 4b's
declaration work. The plan should say so, because the spec places it on
the other side of the boundary.

### 7.4 Lane state at HEAD

`docs/STATUS.md:29-33`:

- lane-1 (main tree): IDLE, pending 4b's plan.
- lane-2 (worktree): BLOCKED until Phase 0b/0c.
- doc lane: IDLE, one candidate queued.

1b *is* Phase 0b, so its landing is what unblocks lane-2, alongside 0c.
The live-executor check is this registry, never `pgrep -f
Projects/dubplate` (CLAUDE.md, and `STATUS.md:21-22`).

### What 1b inherits, concretely

A surface it owns by ratified table, already edited once by rung 4
under a recorded grant. A 4b pass with no plan yet whose named
deliverables exclude `internal/config` but include a second env
template that a test inside `internal/config` will have to be
reconciled with. An inverted ordering that moves one re-expression from
4b to 1b. And the standing fact that 1b's close is what unblocks the
second code lane.

---

## 8. Budget evidence and a proposed ceiling

### 8.1 The measured rate

`docs/HISTORY.md:468-481`, pass 4a-ii:

> **Budget.** 4a-ii spent about 2.7M against its 3M ceiling (Geoff,
> 2026-09-07), an estimate from the agent meters: staleness pre-task
> and chain-args extraction 0.26M, chain 1.32M, the Task 13 and Task 14
> fix and review rounds 0.49M, simplify 0.18M plus its review, the
> close drafter 0.16M, and the conductor's own turns for the pass about
> 0.2M.

The pass ran Tasks 10 through 15, six tasks, in a chain of 16 agents
(`docs/HISTORY.md:10-12`). That gives two usable figures:

- Chain cost per task: 1.32M / 6 = **0.22M**.
- Chain as a share of the pass: 1.32M / 2.7M = **49%**.

The 2.4M flag (80% of 3M) was crossed at the close, after every task
had landed, and nothing was cut.

### 8.2 The S-band precedent

`docs/superpowers/specs/2026-09-05-rung-4-shadow.md:1663-1671` bands 4a
as M and 4b as S, and proposes 5M for 4a and 1.5M for 4b, "each
re-derived from 4a's actual report before 4b's plan is drafted, on the
2c precedent of refining a ceiling from a measurement rather than
asserting one for the whole rung". 4a's actual was 4.99M for 4a-i plus
2.7M for 4a-ii, so the M-band estimate ran roughly 50% over.

The parallel spec bands 1b as S
(`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:54`).

### 8.3 A task-count-based proposal

A plausible 1b task list, from sections 1 through 4:

1. The registry shape: types, the `Register` family, duplicate
   rejection, ordering.
2. Owned wall rows migrated to declarations, `Pending` preserved.
3. Reserved classes migrated, table order preserved.
4. Integration classes migrated as data, `Required`/`Optional`
   preserved, `Load`'s `class.Name == "shadow"` branch retired.
5. `Load` and `doctor` rewired to the registry, the 32-column form and
   block order held.
6. The shadow class and its two thresholds moved into their own
   declaration file, retiring the `sync.OnceFunc` if duplicate
   rejection lands.
7. `KindFloat` and its parser, plus `thresholdStatus`'s switch.
8. The `.gitleaksignore` rider: the durable fix, the note in the file,
   and retiring rung 4's staleness question.

Eight tasks. Documentation of the contract change and the spec
correction rides in the close rather than as a task.

**Arithmetic.** Eight tasks at 4a-ii's 0.22M chain rate is 1.76M. At
4a-ii's 49% chain share that implies 3.6M for the pass, which is M-band
and contradicts the S banding. The rate is the wrong comparator: 4a-ii's
tasks built a shadow runner with FLAC reads, digest determinism, and
budget arithmetic, while 1b's tasks are refactors of one 1255-line
package with its tests already written and its callers already
enumerated in this document. A config task should cost well under half
a shadow task.

**Proposal: 2M, with the 80% flag at 1.6M.** That is the S-band
precedent's 1.5M plus a third, and the delta buys three things 4b does
not carry: a contract change across eighteen exported symbols, a
migration of four tables of four shapes, and the gitleaks rider with its
verification. The plan should re-derive this from its own task count
once the shape decision in section 5 is made, since a value-type
registry is a larger migration than a fixed package-level one.

Cumulative spend to date is about 42.8M (`docs/STATUS.md:66`).

### What 1b inherits, concretely

One measured per-task chain rate (0.22M) from a structurally harder
pass, one measured chain share (49%), an S-band precedent of 1.5M set
for a deploy pass, and a documented habit of re-deriving a ceiling from
a measurement rather than asserting one.

---

## Surprises: where the tree contradicts the spec or the carries

1. **`keys.go` has already been edited by another lane, under a grant
   that the 0b sentence does not mention.** The parallel spec says "no
   lane ever edits `internal/config/keys.go`"
   (`PA:54`), and the rung 4 spec grants rung 4 exactly that until 1b
   lands (`rung-4:1650`). Rung 4a used it: `IntegrationClass`
   (`keys.go:203-209`) and the shadow class (`keys.go:214-229`) are its
   edits. The collision 1b prevents has already happened once, by
   permission.

2. **"Extend rung 1's existing `Register()` pattern" extends a known
   defect.** `Register` (`thresholds.go:40-42`) appends with no
   duplicate check. That forced `sync.OnceFunc` around its only caller
   (`shadow.go:46-52`) and a regression test to hold it
   (`main_test.go:36-58`). The three stdlib registries the Go bar
   points at all panic on a duplicate: `flag.(*FlagSet).Var`,
   `database/sql.Register`, `expvar.Publish`. Extending the pattern
   verbatim to wall rows multiplies the defect across every declaration
   site.

3. **`Load` is not data-driven where the carries assume it is.**
   HISTORY says 1b "absorbs `integrationClasses` as data, not as a
   shape to invent" (`HISTORY:439-441`). The table is already data.
   What is not data is `Load`'s `if class.Name == "shadow"` branch
   (`config.go:235-241`), which hand-builds `ShadowIntegration` from
   the resolved map. Absorbing the table changes nothing there, and the
   shape 1b must actually invent is how a class receives its own
   resolved values.

4. **The gitleaks finding is one name, not a class of names.** The gate
   comment calls it "a config-key-name test fixture that happens to
   look like a generic API key" (`check.sh:104-106`). Measured across
   all 54 `DUBPLATE_*` names in `keys.go`, the env template and
   `shadow.go`, written in the triggering `Key: "NAME"` form, exactly
   one fires: `DUBPLATE_R2_BUCKET`, entropy 3.6143694. The rule needs a
   `key`-like field name on the same line, which is why `keys.go`'s own
   `Keys: []string{...}` block does not fire.

5. **The 0b rider's `.gitleaks.toml` implication has a measured
   over-suppression trap.** A gitleaks `[[allowlists]]` block defaults
   to `condition = "OR"`, so combining `paths` with `regexes` allows
   anything matching either. A global allowlist with
   `regexes = ['''DUBPLATE_[A-Z0-9_]+''']` was measured to suppress a
   genuine `DUBPLATE_MAGIC_LINK_SECRET=<high-entropy value>` finding
   that the same scanner flags without it. `condition = "AND"` fixes
   it, verified. A plan that reaches for a config file without that
   keyword silently widens the suppression across the whole repo.

6. **The blast radius is one file, not the tree.** The fingerprint
   format under `--no-git` is `file:ruleID:startLine`, so only an
   insertion above line 66 of `internal/config/keys_test.go` breaks the
   ignore. The 0b sentence's "reds the gate for every lane at once" is
   right about the consequence and reads as if any insertion anywhere
   does it.

7. **The reserved table is incomplete against its own spec section.**
   The build spec's "The config wall, derived" names an SSO proxy
   capability, a trusted-proxy mode, and a forward-auth front among the
   flag-enabled integrations (`build:167-175`). None has a reserved
   class. The forward-auth front appears only inside two supersession
   messages (`keys.go:167-170`); the trusted-proxy mode and the SSO
   proxy capability appear nowhere in `internal/config`. The same bullet
   says the Navidrome variables "move behind its flag", and they sit in
   a reserved class instead (`keys.go:133-141`), which is the right
   interim shape and not the one the spec describes.

8. **The 0b scope sentence is one class short of the code and the
   carries.** It names "wall rows and reserved classes"
   (`PA:54`). `keys.go:211-213` already says "Rung 1b's registry
   absorbs this table as data" about `integrationClasses`, and
   HISTORY's carry 4 says the same (`HISTORY:439`). The plan should
   record the widening as a spec correction, which the close is already
   required to do for the contract change.

9. **The two class types have incompatible shapes and neither spec says
   which wins.** `ReservedClass` (`keys.go:47-51`) has `Name`, `Rung`,
   `Keys`. `IntegrationClass` (`keys.go:203-209`) has `Name`, `Rung`,
   `Flag`, `Required`, `Optional`. Unifying them means either two
   registry shapes or one shape with fields that are dead for half its
   entries. That question is the pass's central design decision and
   neither the parallel spec nor the build spec states a preference.

10. **The env template describes a wall that no longer exists, and a
    test inside `internal/config` reads it.** Of its 42 assigned names,
    2 are greenfield owned-now rows, 26 reserved, 11 superseded, 3 not
    `DUBPLATE_`-prefixed. Two of the four owned-now rows,
    `DUBPLATE_LIBRARY_ROOT` and `DUBPLATE_SITE_NAME`, are absent. The
    template header admits the mismatch (`:1-3`). The surprise is that
    `TestLoad_TemplateKeysClassified` (`config_test.go:196`) makes the
    mismatch a gate-resident coupling between `internal/config` and
    `env/`, and 4b is about to add a second template
    (`rung-4:1632`).

11. **`KindInt` is the zero value and `thresholdStatus` defaults to
    it.** `ThresholdKind`'s zero is `KindInt` by design
    (`thresholds.go:13-18`), and `doctor.go:120` parses any unnamed kind
    as an integer. Adding `KindFloat` in that shape means a float
    threshold whose `Kind` was never set reports as an integer parse
    failure against the environment, not as the registration bug it is.
    The registry already draws that distinction elsewhere
    (`thresholds.go:60-63`, `thresholds.go:84-87`).

12. **The salvage tree holds two key shapes the greenfield table cannot
    express.** `pingURLsFromEnv` (`_salvage/.../config.go:409-426`)
    discovers a `DUBPLATE_PING_<SLUG>_URL` family by scanning; the
    greenfield froze four literal names (`keys.go:116-125`).
    `loadCollections` (`_salvage/.../config.go:383-402`) reads
    `DUBPLATE_COLLECTIONS`, which is in no greenfield class at all, so
    an adopter setting it today gets "unknown configuration key" while
    CLAUDE.md still lists collection routing as live behavior.

13. **The ordering this window uses inverts what rung 4's spec
    assumed.** Its staleness pre-task asks "Has rung 1b landed?" as a
    blocker (`rung-4:1687-1690`) and rules that a yes moves ruling 1's
    option (a) into registry terms. Running 1b after 4b means that
    branch never fires, and the re-expression becomes 1b's own
    migration rather than 4b's declaration.

14. **The four-knob follow-on is phrased as one item but is two.**
    HISTORY carry 3 (`HISTORY:434-438`) joins "rung 4's own `KindFloat`
    need" to the four-knob registration. The kind is a registry
    capability that belongs to 1b; the registration is explicitly the
    first rung 3 pass after 1b (`HISTORY:1434-1439`). A plan that reads
    the carry as one item lands either both or neither, and both is
    scope creep into rung 3's territory.
