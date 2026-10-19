# Pass 5b-ii Task 0 staleness report

Run 2026-09-12 against HEAD `f75b0d9` at `opus`, fresh context, per the plan's Task 0. Result: 20 OK, 2 NOTE, 0 BLOCKER, item 22 skipped. The conductor's baseline gate (`cairn-run-gate 'bash scripts/check.sh'`) ran green at the same HEAD in 75 seconds. The pass was not launched in this session; Geoff redirected the session to the workstation infra round. This report is the Task 0 evidence for whichever session launches 5b-ii, valid while HEAD is unchanged.

---

# Task 0 staleness report, pass 5b-ii

**HEAD: `f75b0d9`** ("Prune STATUS to the cap by moving the 2026-09-12
process record into HISTORY"). The plan's pointers were resolved at
`b91bae1`; three commits have landed since (`2a81a0c`, `1c5c76d`,
`b003554`, `f2d3a49`, `f75b0d9` — all process, gate, and ledger work,
none under `internal/auth`, `internal/identity`, or `internal/serve`).

**Counts: 20 OK, 2 NOTE, 0 BLOCKER, 1 SKIPPED (item 22).**

The launch is not blocked. Items 1 through 4 all answer yes.

---

## 1. Lane registry and tree freedom — OK

`docs/STATUS.md`'s `## Lane registry` table:

| Lane | Tree | State |
|---|---|---|
| 1 | main tree, master | CLOSED. Awaiting the conductor's next dispatch |
| 2 | worktree removed, branch deleted | CLOSED. 5b-i merged at `1d4ed29` |

Both lanes CLOSED. `git worktree list` shows one entry only,
`/var/home/glw907/Projects/dubplate f75b0d9 [master]`; no lingering
worktree, so the STATUS row's "to be removed by the conductor" is
already discharged. `git branch -a` shows `master` plus the two
origin refs; lane 2's branch is deleted.

`git status --porcelain internal/auth internal/identity internal/serve
cmd/dubplate-server go.mod go.sum` returns nothing. The only untracked
files in the repo are two 2026-09-04 research documents
(`docs/research/2026-09-04-agentic-coding-public-evidence.md`,
`docs/research/2026-09-04-visual-collaborator-outreach.md`),
neither under a contested surface. No uncommitted change I did not
author exists in the pass's territory.

## 2. Any lane writing `internal/serve` since 5b-i's merge — OK

`git log --oneline 1d4ed29..HEAD -- internal/serve` returns an empty
list. `PA:80`'s prohibition holds. `1d4ed29` resolves to
"refactor(auth): give the subdirectory join one home and name two
silences".

## 3. Security review present and unchanged — OK

`docs/research/2026-09-12-rung-5b-i-security-review.md`,
**493 lines** (25908 bytes), matching the drafting-time count exactly.

It is now **tracked**, not untracked as the plan recorded: the
conductor committed it at `f36b10c` ("Add the rung 5b-ii pass plan at
revision 2 with its review record and the 5b-i security read"). It is
clean against that commit (absent from `git status`). Tasks 1 and 2
have their primary input intact.

## 4. The six named findings at HEAD — OK, all six live

| Finding | Expected | At HEAD | Verdict |
|---|---|---|---|
| B1 | `internal/auth/nonce.go:117` writes `nonceRecord(now, s.retention)`; `Insert` at `:104` takes a nonce and a time | `Insert` at `:104`, signature `Insert(nonce [NonceLength]byte, now time.Time) error`; `:117` is `writeAndSync(f, nonceRecord(now, s.retention))` | reproduces, exact |
| M1 | `internal/identity/identity.go:34-43` returns `strings.ToLower(addr.Address)`, no re-parse, no character filter | `NormalizeAddress` at `:34`, closes `:43`; length cap then `mail.ParseAddress` then `strings.ToLower(addr.Address)` at `:42`; no re-parse, no byte filter | reproduces, exact |
| M2 | `internal/auth/epoch.go:80-98` writes `now` without reading the stored value | `Bump` at `:80`; normalizes, resolves the path, truncates to `epoch`, `atomicfile.WriteStream`s `epochRecord(epoch)`, then `e.notify`. No read of the stored epoch, no mutex across read and write, no `max` | reproduces, exact |
| M3 | `internal/auth/token.go:244-258` decodes with `base64.RawURLEncoding.DecodeString` | `splitSigned` at `:244`, the lax decode at `:251` | reproduces, exact |
| M4 | `cmd/dubplate-server/doctor.go:88-112` unwraps one level; `identitySourceStatus` at `:167-177` consults `Required` alone | `rowStatuses` at `:90` (doc comment from `:86`), the single `err.(interface{ Unwrap() []error })` at `:96`, body to `:112`; `identitySourceStatus` at `:167`, loops `src.Required` only, never `src.Optional` | reproduces; the function pointer is two lines low (doc-comment boundary), same subject |
| M5 | `internal/auth/token.go:245` calls `strings.Split` with no bound | `:245` is `parts := strings.Split(value, fieldSeparator)`, no limit argument, the count check after | reproduces, exact |

No later commit has fixed any of them. Every fixing task's test can
still fail before its change.

## 5. `cmd/dubplate-server/auth.go` shapes Task 3 extends — OK

Every pointer exact:

- `secretLengthFloor` — `const secretLengthFloor = 32` at `:68` (doc from `:61`).
- The four thresholds — `var (` block `:75-100`, all four `config.KindDuration`, defaults `15m`, `2160h`, `4320h`, `24h`, keys `DUBPLATE_MAGIC_LINK_EXPIRY`, `DUBPLATE_SESSION_LIFETIME`, `DUBPLATE_SESSION_ABSOLUTE_CAP`, `DUBPLATE_NONCE_RETENTION`.
- `registerAuthConfig` at `:108`.
- `validateMagicLinkSource` at `:124`, closing `:156`.
- `validatePublicURL` at `:171`.
- `isLoopbackHost` at `:189`.
- `resolveAuthConfig` at `:207`.

**`resolveAuthConfig`'s current return shape**, which Task 3 extends
and Task 11 calls:

```
func resolveAuthConfig(environ []string) (magicLinkExpiry, sessionLifetime, sessionAbsoluteCap, nonceRetention time.Duration, err error)
```

Four named `time.Duration` returns plus an error. A fifth or sixth
duration added by Task 3 lengthens an already five-value signature;
the conductor may want to rule on whether it becomes a struct before
Task 3 dispatches, because Task 11 is the caller.

## 6. `internal/config`'s registry family — OK

- `IdentitySource` at `registry.go:349`, struct closing `:377`. Fields, in order: `Name`, `Selector`, `Required []string`, `Optional []string`, and `Validate func(values map[string]string, environ []string) error` at `:376`. Signature exact, including the `environ []string` second parameter the ruling (conductor 2026-09-12, B2) added.
- `Threshold` at `thresholds.go:45`, fields `Key`, `Kind`, `Default`, `Doc`.
- The four kinds at `thresholds.go:16-21`: `KindInt ThresholdKind = iota`, `KindDuration`, `KindFloat`, `KindString`.

Nothing under `internal/config` has changed since `b91bae1`. Task 3
can declare through this family unaltered.

## 7. `internal/serve`'s composition — OK

`Shell.Handler` at `serve.go:150`, closing `:163`. The exact
composition lines, for R-5b27's insertion point:

```
150  func (s *Shell) Handler() http.Handler {
151      s.handlerBuilt = true
152
153      fallthroughHandler := s.fallthroughHandler
154      if fallthroughHandler == nil {
155          fallthroughHandler = http.HandlerFunc(http.NotFound)
156      }
157
158      routed := s.prefix.Route(s.terminated(), s.stripSessionCookie(fallthroughHandler))
159      guarded := refuseEchoMethods(routed)
160      normalized := Normalize(guarded)
161      stripped := s.strip.Wrap(normalized)
162      return s.logRefusal(stripped)
163  }
```

`Register` at `:80`. `SetFallthrough` at `:136`. `stripSessionCookie`
is a method, `func (s *Shell) stripSessionCookie(next http.Handler)
http.Handler` at `:225` (doc from `:217`). All four exact.

## 8. `ClientIP` and the strip — OK

- `func ClientIP(r *http.Request, cfg TrustedProxyConfig) string` at `internal/serve/clientaddr.go:102`.
- `func NewStrip(header, sessionCookie string, proxy TrustedProxyConfig, exemption FrontExemption) Strip` at `internal/serve/strip.go:300`.
- `cmd/dubplate-server/serve.go:158-161`, `newServeShell`, with `:159` passing `auth.SessionCookieName`:
  `strip := serve.NewStrip("", auth.SessionCookieName, serve.TrustedProxyConfig{}, serve.FrontExemption{})`.

All 5a's, unchanged. No task in this pass derives a client address.

## 9. Toolchain, measured — OK

Probe built under its own `go.mod` with `go 1.27.0`, run on the tree's
toolchain, **`go version go1.27.1 linux/amd64`**.

`net/http.CrossOriginProtection` exists, constructed by
`http.NewCrossOriginProtection()`, and all four methods compile and
run:

```
Check -> <nil>
Handler ok: true
SetDenyHandler ok
AddInsecureBypassPattern ok
```

`base64.Encoding.Strict` exists, and
`base64.RawURLEncoding.Strict().DecodeString` refuses a value with
non-zero trailing bits where the lax decoder accepts it:

```
value="AB" lax=(1 bytes, err=<nil>) strict=(0 bytes, err=illegal base64 data at input byte 0)
value="AA" lax=(1 bytes, err=<nil>) strict=(1 bytes, err=<nil>)
```

`go.mod:3` is `go 1.27.0`, as expected. Task 1's D4 and Task 8's D10
rest on measured behavior, not memory. No divergence.

## 10. Redirect-target validation in `internal/auth` — OK, none exists

`grep -rni 'redirect|destination' internal/auth/` finds **three hits,
all of them prose in one doc comment**, none a function, identifier,
or validation:

- `token.go:26` — a comment noting a later field, "rung 12's destination among" them.
- `token.go:115-116` — `MintLink`'s doc, stating there is "no destination parameter" and that a later destination "rides inside the MAC-covered token under a new linkVersion".

No `*_test.go` hit either. `MintLink` at `internal/auth/token.go:118`
is `MintLink(secret []byte, address string, now time.Time, expiry
time.Duration) (string, LinkToken, error)`: **no destination
argument**, as expected.

The `b91bae1` erratum's claim that R33's redirect validation "landed
in `internal/auth` at 5b-i" is false against the tree. R-5b26
disposes it and the close drafts the correction; see item 19.

## 11. Registry arithmetic against `doctor`'s line-count assertions — OK

**The arithmetic.** This pass adds six thresholds, no wall row, no
reserved class, no integration class: **net plus six on the
production registry**, and plus six doctor lines in the threshold
block.

**The two assertion sites, at their current line numbers:**

- The extra-line assertion, `cmd/dubplate-server/doctor_test.go:130`: `t.Errorf("stdout has %d extra line(s) beyond owned+reserved+threshold rows: ...")`, inside `TestDoctorCmd`'s table.
- `TestDoctorCmd_LineCountFollowsRegistry` at `:153` (doc from `:136`), body to `:247`.

**Neither assertion would fail.** Both are derivations, not literal
counts. The extra-line check walks `reg.OwnedWall()`,
`reg.ReservedClasses()`, `reg.IntegrationClasses()` and
`reg.Thresholds()` in order, advancing `li` per row, then compares
`li` to `len(lines)`; six more registered thresholds produce six more
expected rows and six more printed rows. `LineCountFollowsRegistry`
starts from `config.NewRegistry()` and counts each family's length
itself. No hard-coded total exists in either. `TestDoctorCmd_LiteralSpotCheck`
pins two named lines (`DUBPLATE_ADMIN_EMAIL`, `DUBPLATE_IDENTITY_SOURCE`),
not a count. **No blocker for Task 3.**

**Which tests build through which registry.** `productionRegistry()`
is defined at `cmd/dubplate-server/serve_test.go:23` and called at:
`serve_test.go:28`, `:165`, `:238`; `auth_test.go:86`, `:101`,
`:132`, `:341`; `templatekeys_test.go:50`, `:81`, `:154`;
`doctor_test.go:355`. `config.NewRegistry()` is used at
`doctor_test.go:46` (the `TestDoctorCmd` table), `:154`
(`LineCountFollowsRegistry`), `:263`, `:338`, `:387`, `:434`.

Both line-count assertions therefore run on a **fresh** registry, and
the six new thresholds reach neither. They are proved instead by
`LineCountFollowsRegistry`'s per-family scratch additions, and by the
production-registry tests in `auth_test.go` and `templatekeys_test.go`
that load the real declarations. This is the vacuity the plan names:
a family the fresh registry never populates passes both assertions
without evidence. **Task 3 should add its declaration proof against
`productionRegistry()`**, not rely on the two derivations.

**Task 2's M4 reconciled against R10.** R10's statement is at
`doctor_test.go:147-150`, "one per owned wall row, one per reserved
class regardless of its key count, one per integration class, one per
threshold, and one per declared identity source regardless of how
many keys that source requires", tagged `ruled, conductor 2026-09-07,
R10` at `:152`. Its assertion is the identity-source block at
`:231-244`: it adds one source of two required keys, expects
`wantTotal` to rise by exactly one, and fails with "one line per
source, not per key".

**M4 as this plan now states it leaves that meaning intact.**
Flattening the join and consulting `Optional` changes only which
rows `rowStatuses` can report and which keys `identitySourceStatus`
inspects; it adds no printed row, so the per-source arithmetic is
untouched.

**The scratch source is declared but never selected.** It is added
with `Selector: "DUBPLATE_TEST_SCRATCH_SELECTOR"`, and `doctorLines`
(`:325`) runs `newDoctorCmd(baseEnviron(), reg, ...)` where
`baseEnviron()` (`check_test.go:14-21`) sets only
`DUBPLATE_LIBRARY_ROOT`, `DUBPLATE_STATE_DIR`, `DUBPLATE_SITE_NAME`
and `DUBPLATE_ADMIN_EMAIL`. The selector is unset, so the source is
never selected and its row prints "inactive". **The plan's warning is
confirmed:** a per-key expansion scoped to the selected source would
pass R10's assertion without honoring the rule, because this test's
source is never the selected one. M4 must not take that shape, and
the conductor should require Task 2's proof of one-row-per-source to
use a **selected** source.

## 12. No-deploy carry and the env template — OK (a signal, not a blocker)

`env/dubplate-server.env.template:74` is
`DUBPLATE_IDENTITY_SOURCE=magic-link`, exactly as expected.

Grepping the file for each of the nine keys
`magicLinkSource.Required` names (`auth.go:44-54`:
`DUBPLATE_ROSTER_FILE`, `DUBPLATE_MAGIC_LINK_SECRET`,
`DUBPLATE_SESSION_SECRET`, `DUBPLATE_PUBLIC_URL`,
`DUBPLATE_SMTP_HOST`, `DUBPLATE_SMTP_PORT`, `DUBPLATE_SMTP_USER`,
`DUBPLATE_SMTP_PASSWORD`, `DUBPLATE_SMTP_FROM`) returns **zero
assignments for all nine**. The single textual occurrence is prose at
`:72`, inside the comment at `:70-73` that says the source's required
keys "arrive at 5d". Security review m13 stands, the template still
selects a source it cannot satisfy, and the no-deploy carry is live.
This pass writes no template row.

## 13. `go.mod` still cobra-only — OK

```
module github.com/glw907/dubplate

go 1.27.0

require github.com/spf13/cobra v1.10.2

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
)
```

`cobra v1.10.2` at `:5` plus the two indirects at `:7-10`, as
expected.

**Version query, module unmodified.** `go list -m -versions` (no
`GOFLAGS=-mod=mod` needed):

- `github.com/wneessen/go-mail`: newest listed **v0.8.1** (preceded by v0.8.0, v0.7.3, v0.7.2, v0.7.1).
- `golang.org/x/time`: newest listed **v0.16.0** (preceded by v0.15.0, v0.14.0, v0.13.0, v0.12.0).

**Both versions the prior-art table records are still the latest**, so
Tasks 6 and 7 need no note and the 2026-09-08 adoption evidence is
current. `git status --porcelain go.mod go.sum` afterwards is empty:
neither file was touched.

## 14. `_salvage/internal/notify` unchanged — OK

`notify.go` **96 lines**, exactly as expected; `notify_test.go` 324
lines, present.

Exported surface: `type Mailer struct` at `:27`, `func New(cfg
config.SMTPConfig) *Mailer` at `:46`, `func (m *Mailer) Send(ctx
context.Context, to, subject, body string) error` at `:57`. Nothing
else exported. `const implicitTLSPort = 465` at `:24`, read at `:76`.

The two test-only fields each carry a comment saying so: `tlsConfig
*tls.Config` ("New never sets it; this package's own tests do") and
`forceImplicitTLS bool` ("New never sets it. It exists only so this
package's own parity test can exercise that branch"). Both files'
last touch is `9d0dc5f`, the pivot commit; nothing has edited them
since. Task 6 adopts a frozen surface.

## 15. `_salvage/internal/config/contributors.go` fail-open loader — OK

Both cited shapes present, the first two lines lower than the plan
records:

- The missing-file branch at **`:36-40`** (plan said `:33-39`): `func LoadContributors(path string) (Roster, error)` at `:36`, `data, err := os.ReadFile(path)` at `:37`, and `if errors.Is(err, os.ErrNotExist) { return Roster{}, nil }` at `:38-40`. Its doc comment at `:32-35` states the fail-open intent outright: "A missing file returns an empty [Roster] and a nil error". D6 is stated against live code.
- The per-request membership re-check at **`:93`**, exact: `func (r Roster) DirNameFor(email string) (string, bool)` at `:93`, resolving `r.byEmail[strings.ToLower(strings.TrimSpace(email))]` at `:94`. R30's session-path re-check cites a real site.

Off by three lines, same subject, content intact.

## 16. `_salvage/internal/serve/serve.go`'s `requireSameOrigin` — OK

`func requireSameOrigin(next http.Handler) http.Handler` at
`:152-176`. Its ordering, as expected and matching `spec:1206-1210`:

1. `:154-157` — GET and HEAD pass through unchecked.
2. `:159-162` — `Sec-Fetch-Site == "same-origin"` passes, **honored on its own**.
3. `:164-172` — the `Origin` fallback: parse, and refuse with 403 when the parse fails or `!strings.EqualFold(u.Host, r.Host)`; otherwise pass.
4. `:174` — no `Sec-Fetch-Site`, no `Origin`: refuse with 403.

Task 8 adopts the lesson, not the code.

## 17. `internal/atomicfile` entry points and the directory fsync — OK

- `func WriteFile(path string, data []byte, perm os.FileMode) error` at `atomicfile.go:16`.
- `func WriteStream(path string, perm os.FileMode, write func(io.Writer) error) error` at `:55`, closing `:91`.
- The directory open and `Sync` at **`:82-89`**, exact: `dirFile, err := os.Open(dir)` at `:82`, the error return `:83-85`, `defer func() { _ = dirFile.Close() }()` at `:86`, and `if err := dirFile.Sync(); err != nil { return fmt.Errorf("fsync %s: %w", dir, err) }` at `:87-89`.

The full sequence is temp create, write, `tmp.Sync`, close, chmod,
rename, then the containing directory's fsync. D7's roster write rests
on live code.

## 18. HISTORY ranges this plan cites — NOTE

**Every 5b-i range is shifted by exactly plus forty-seven lines.**
Cause identified: HEAD itself, `f75b0d9`, added 47 lines to
`docs/HISTORY.md` (`git diff --numstat`: `47 0`), inserting the
section `### Process changes applied 2026-09-12` at `:7-64`, inside
the 5b-i entry and above every cited subsection. Verified by
regenerating the table of contents from `b91bae1`, where all eight
ranges match the plan exactly.

| Cited | Plan | At HEAD | Subject |
|---|---|---|---|
| 5b-i entry | `:5-544` | `:5-591` | same (`## Rung 5, pass 5b-i …` at `:5`; next `##` at `:592`) |
| exported surface | `:168-184` | `:215-231` | same |
| contract coverage | `:195-207` | `:242-254` | same |
| resolutions | `:208-235` | `:255-282` | same |
| declared gaps | `:368-402` | `:415-449` | same |
| review record | `:403-431` | `:450-478` | same |
| errata list | `:432-495` | `:479-542` | same |
| 5b-i to 5b-ii handoff | `:496-530` | `:543-577` | same |

Every range still resolves to the same subject and the content exists
unchanged; only the offset moved, uniformly. Graded a note rather
than a blocker: nothing was rewritten, and `HISTORY:526-529` (item
11's 5a citation) is in the 5a entry, below the insertion, unaffected.
**The mechanical fix is plus forty-seven on every 5b-i `HISTORY:`
pointer in the plan**, which the close's errata can apply in one
sweep. An implementer handed a stale pointer lands 47 lines high,
inside the neighboring subsection, so the conductor may prefer to
re-key them before Task 1 dispatches.

## 19. The eighteen spec errata — NOTE

`b91bae1` ("Apply the eighteen rung 5 spec errata rung 5b-i's close
owes") adds **162 lines** to
`docs/specs/2026-09-08-rung-5-identity-auth.md`, exactly
as the plan records, all under `## Errata the close applies` at
`:1982`. The 5b-i section is headed at `:2157`, "Pass 5b-i's close,
2026-09-12, adds the corrections below."

**Count of dated erratum blocks: eighteen**, each opening `> Erratum,
2026-09-12 (5b-i close):`. The spec carries 31 such blocks in total,
the other thirteen being 5a's close.

**One contradicts this plan, the one item 10 anticipated.** The R33
erratum (at `:2287-2296`) states that R33's post-login destination
rule "landed in `internal/auth` at 5b-i" and corrects the pass-split
table's 5b row "to include the post-login-destination redirect
validation among 5b-i's built deliverables". Item 10 measured the
opposite: no redirect or destination validation exists anywhere in
`internal/auth`, only two doc comments saying a destination is
deferred to rung 12, and `MintLink` takes no destination argument.
**The erratum overstates what 5b-i built.** R-5b26 already disposes
it and the close drafts the correction, so this is the expected
finding rather than a new one; recorded here as the evidence that
disposal rests on. No other erratum contradicts the plan's text.

## 20. `.gitleaksignore` — OK

Tracked (`git ls-files --error-unmatch` succeeds), **0 lines, 0
bytes**: still the empty file rung 1b left. Its only commits are
`8cd3b7f` (pass 2a) and `3a35558` (the fixture-field rename), both
well before rung 5. **No lane has appended to it.** No fixture
literal quoted.

## 21. The "rejoins at rung 7" sentence, and `deploy.sh` — OK

All four present, none changed:

- `CLAUDE.md:122` — "The `web/` leg rejoins at rung 7, when `web/` is rebuilt against the …".
- `scripts/check.sh:5` — "… the web leg rejoins at rung 7, when …".
- `Makefile:33` — "The web legs rejoin at rung 7, when web/ is rebuilt against the …". Plural subject and present-tense verb rather than "rejoins"; same sentence, same claim. Pass 5c's correction should expect this wording.
- `scripts/deploy.sh:420-426` — still delivering the roster: a `# --- /etc/dubplate/contributors.yaml ---` banner at `:420`, the echo at `:425`, and `deploy_file "$REPO_ROOT/config/contributors.yaml" /etc/dubplate/contributors.yaml 0640 root:music` at `:426`. R37 leaves it alone.

## 22. Gate wall time — SKIPPED

**Skipped by the conductor's instruction**: the gate wall time is
being measured separately. Neither `go test ./...` nor `bash
scripts/check.sh` was run by this task, so no gate ran at any point
during it. 5b-i's 3m20s baseline is unverified here.

## 23. Does `validateMagicLinkSource` validate `DUBPLATE_SMTP_FROM`? — OK

**No.** `cmd/dubplate-server/auth.go:124-156`, key by key, is exactly
three checks:

1. **The four secret lengths** (`:127-135`): for each of `DUBPLATE_MAGIC_LINK_SECRET`, `DUBPLATE_SESSION_SECRET`, `DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS`, `DUBPLATE_SESSION_SECRET_PREVIOUS`, when the key is present and shorter than `secretLengthFloor` (32), it records "too short, want at least 32 bytes".
2. **`DUBPLATE_PUBLIC_URL`** (`:139-143`): when present, delegated to `validatePublicURL`, which checks that it parses, carries a non-empty host, and is https unless the parsed hostname is loopback.
3. **The retention relation** (`:145-153`): parses `DUBPLATE_NONCE_RETENTION` and `DUBPLATE_MAGIC_LINK_EXPIRY` through `config.ParseDuration`, reports either parse error, otherwise refuses `retention < expiry`.

Then `return errors.Join(errs...)`.

**Five of the source's nine required keys are validated by no rule at
all**: `DUBPLATE_ROSTER_FILE`, `DUBPLATE_SMTP_HOST`,
`DUBPLATE_SMTP_PORT`, `DUBPLATE_SMTP_USER`, `DUBPLATE_SMTP_PASSWORD`.
`DUBPLATE_SMTP_FROM` is a Required key at `auth.go:52` and **no
function reads it**: grepping `SMTP_FROM` across `cmd/` and
`internal/` returns three hits, the declaration at `auth.go:52` and
two test-file mentions (`auth_test.go:65`, `check_test.go:42`), and no
production reader. The expectation holds, Task 3's check is a real
addition rather than a duplicate, and the four SMTP transport keys are
the wider unvalidated set the same task may want in view.
