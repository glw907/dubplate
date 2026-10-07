# Rung 12 slice 2a plan review: contract and criteria lens

Target: `docs/superpowers/plans/2026-10-06-rung-12-slice-2-oracle-harness.md` at `f62a151` (draft revision 1, pass
class `tool`). One lens: is every promise testable, does each criterion name its fixture and its failure reason, can
any pass vacuously, and does the ceremony fit the risk. Findings are correctness gaps ranked by consequence, then
over-ceremony ranked by cost. Polish is out of scope.

Verified against the tree: kin-openapi v0.149.0 source in the module cache (`openapi3filter/validate_response.go`,
`openapi3/content.go`, `openapi3filter/options.go`), `go help list` (go1.27.1), `scripts/check.sh:20-60`,
`internal/harness/conform/arch_test.go`, `internal/engine/tagio/write.go`, `cmd/harness/main.go`,
`internal/auth/roster.go:335-345`, the spec's Evidence section (`:258-289`), the re-plan's ruling 5, and
`docs/HISTORY.md`. Two measurements: `go version -m` on a `go build` of `cmd/harness` (stamped
`v0.0.0-20261007023827-87b63807258f+dirty`) and on a `-buildvcs=false` build (stamped `(devel)`), and
`go list -deps ./cmd/dubplate-server | grep -E 'internal/(harness|opensubsonic)|getkin'` (prints nothing today).

Counts: 1 blocker, 6 major, 9 minor correctness findings; 3 minor over-ceremony findings. No owner fork.

## Correctness gaps

### C1 (blocker): the seed-session gate grade cannot pass without corrupting the committed ledger

`plan:16-17`, `:476-481`, `:551-554`, `:560-563`.

**Defect.** Task 8's gate test grades Navidrome against gonic over ten endpoints and asserts zero unledgered
findings. The two servers differ far beyond the folder view. Navidrome emits OpenSubsonic extended fields (sort names,
genre arrays, MBIDs, played and created timestamps, and more) that gonic omits, and they order and group index entries
differently. Every such difference is a finding. None of the four classes absorbs "present on one oracle only."
`volatile` still compares presence (`:477-478`), and `by-construction` means an oracle cannot grade a surface.
The implementer then has two paths. It can misfile dozens of oracle-pair rows into `data/ledger.json`, or it can invent
a class mid-task on the Opus seat, an uncovered decision. Either way, the committed ledger fills with rows about
Navidrome-versus-gonic. The spec's oracle tier diffs dubplate against an oracle (spec `:274-276`), so those rows
describe no dubplate behavior. The rot rule (`:483-484`) then pins every one of them to the seed session, and each
oracle bump re-adjudicates them. The perturbation table has the same flaw (`:560-563`). Its baseline is a real
Navidrome and gonic pair, so "exactly one finding per row" sits on a non-zero baseline. "Zero findings outside the
ledgered classes" is the same circular assertion.

**Fold.**
- Grade the seed pair against a test-only ledger under `grade/testdata/`. It declares the oracle-pair
  disagreements the recording shows. The committed ledger keeps the spec's two declared classes and the dubplate-facing
  `volatile` rows only.
- Assert named findings, not a count. One example: a `by-construction` finding at a named path in the compilation's
  `getMusicDirectory`.
- Pin the full report as a committed golden, so "exactly the disagreements" becomes a byte comparison.
- Run the perturbation table on a same-source pair (one Navidrome body against itself, perturbed once per row). The
  unperturbed baseline is then exactly zero, and "exactly one finding" means something.

### C2 (major): id pairing by position cascades on any ordering difference

`plan:545-546`, `:441-444`.

**Defect.** Ids pair "by the position they occupy in corresponding responses." Navidrome and gonic, and later dubplate
and either oracle, sort artists and albums under different article and diacritic rules. One reordered list then pairs
the wrong ids. Each later request by symbolic key adds an id conflict, so one real ordering difference surfaces as N
spurious conflicts that hide it. The plan already holds the ground truth. Task 5 resolves every symbolic key to a
per-oracle id (`:441-444`), then discards that mapping.

**Fold.**
- Task 4's metadata records each resolved symbolic key with its id per source.
- Task 8 seeds the correspondence from that map, and positional pairing covers only ids no request named.
- Add a mutation: pair by position only, with a session whose two sources order one list differently. The
  reorder row must report one finding, not a cascade.

### C3 (major): Task 7's "no schema" fixture exercises the wrong kin-openapi branch, and silent-nil paths are uncovered

`plan:115`, `:114`, `:511-513`, `:528-529`.

**Defect.** The quotes table cites only the empty-content branch. The source (`openapi3filter/validate_response.go`,
v0.149.0) has more paths that return `nil`, which reads as "valid":
- a `HEAD` request;
- status 301, 304, 307, or 308 ("These status codes will never be validated");
- `responses.Len() == 0`;
- an undeclared status with no `default` response while `IncludeResponseStatus` is false ("By default, status that is
  not documented is allowed");
- empty `content`;
- a media type whose `Schema == nil`.

The plan's no-schema proof uses "a status the file does not declare." With `IncludeResponseStatus` true, that status
returns `ResponseError{Reason: "status is not supported"}`, which the checker reports as a violation, not as no
schema. With the flag false, it returns `nil`, a silent pass. The plan quotes the option but never says which value
the checker sets. So the no-schema test passes or fails depending on an unstated choice. A checker that just calls
`ValidateResponse` still reports "valid" for several of these paths.

**Fold.**
- The checker resolves `Operation.Responses.Status(status)`, then `Default()`, then `Content.Get(mime)`, then
  `.Schema` itself. It returns no-schema or undeclared-status before calling `ValidateResponse`, which then runs only
  when a schema exists.
- State the `IncludeResponseStatus` value and the report for an undeclared status.
- State that 3xx and `HEAD` are never "valid."
- Quote these lines in the quotes table.
- Fixtures: the no-schema proof uses a hand-built minimal document under `conformance/testdata/` with a declared
  status and empty content. The pinned file's 200 entries all carry JSON content (`:118`), so it cannot supply this
  case. Add one proof for an undeclared status.

### C4 (major): the `MultiError` mutation is vacuous against the named fixtures

`plan:526-530`.

**Defect.** Each proof body carries one violation: the `song` array removed, or a numeric `id`. The mutation "drop
`MultiError` (the two-violation body reports one)" needs a body carrying both, and none is named. Dropping
`MultiError` passes every listed proof. Kin wires the option through `openapi3.MultiErrors()` (`validate_response.go`,
the `options.MultiError` branch), and that is exactly what the test must observe.

**Fold.** Add a third proof body carrying both violations. Assert that two violations come back with both JSON paths.

### C5 (major): fixture bytes depend on the building binary's VCS stamp, invisible to the determinism proof

`plan:326-327`, `:334-336`, `:415`.

**Defect.** `tagio.Write` stamps the Vorbis vendor string from `debug.ReadBuildInfo` (`write.go:239-245`):
`"dubplate development"` under `go test` or `(devel)`, and `"dubplate " + info.Main.Version` otherwise. Measured: a
`go build` of `cmd/harness` stamps `v0.0.0-20261007023827-87b63807258f+dirty`, and `-buildvcs=false` stamps
`(devel)`. The vendor block's length is part of the file, so each track's byte size changes with the commit and with
the `+dirty` suffix. Subsonic's `size` field reports that size. The determinism proof always runs under `go test`, so
it always sees `development` and cannot catch this. Consequence: the seed session's `size` values reflect whatever
binary `record-oracle` ran as. Slice 3's dubplate-against-oracle grade, run over a library built anywhere else, then
reports a finding on every song. The session also records only the manifest's sha256 (`:415`), which cannot detect
the drift.

**Fold.**
- Task 1 states the stamp dependence. Its builder test reads each built file's vendor string back and asserts a fixed
  value.
- The recorder computes a library digest over the built tree (per-file sha256 and size) and records it in the session.
- The grader refuses to compare sources recorded over different library digests, naming both.
- Pin the mechanism that fixes the vendor. Recommended: the verbs refuse a binary whose `Main.Version` is not
  `(devel)` or empty, naming `go build -buildvcs=false`. `tagio` stays read-only.

### C6 (major): "no active row can rot" holds only for sessions the gate grades

`plan:483-484`, `:490-493`.

**Defect.** Staleness fires only while grading the session a row names. A row that names a session the gate never
grades, or a session that does not exist, is never checked. Today only the seed session is graded (`:551`). From slice
3 on, every slice records its own sessions (`:205-206`), so the headline promise silently narrows to one session.

**Fold.**
- The ledger loader refuses an `active` row whose session directory is absent under `data/sessions/`. Add this to the
  many-bad states.
- A gate test grades every committed session, discovered by walking `data/sessions/`, so a new session joins
  staleness checking without an edit.

### C7 (major): the Finding and scope contract between Tasks 6 and 8 is unspecified

`plan:470-478`, `:495-499`, `:541-549`.

**Defect.** Task 6 lands first and tests the ledger against "a hand-built findings list," so it defines the Finding
type that Task 8, on Opus, must produce. The plan names neither of two things:
- **The finding kinds.** The `volatile` rule ("presence and type are still compared") and the mutation "let a
  `volatile` row suppress a type change" both require a kind field that separates value, type, and presence findings.
- **The scope grammar.** A "path pattern into the response" must address both the JSON tree and the XML tree, where
  attributes and child elements differ (`:542-544`, `:562`). Without a grammar, a pattern written for JSON may never
  match the XML twin, and an `active` row then goes stale or matches by accident.

This is the plan-staleness defect class: Task 8 reshapes Task 6's type in a fix round.

**Fold.**
- Task 6 names the kind enum: value, type, presence, order, id-conflict, missing-source, and parse.
- Task 6 names the pattern grammar. For example, `endpoint`, `format`, and a `/`-separated path with `*` per segment,
  where XML attributes and elements map to the same segment names as the JSON keys.
- Task 6 names which kinds each class suppresses. Task 8 consumes all of this unchanged.

### C8 (minor): the session's `serverVersion` has a second channel, the committed file

`plan:485-488`, `:428-429`.

**Defect.** The plan says only the recorder sets a session's version. But session metadata is committed JSON, so a
hand edit changes it, and Task 4's testdata sessions are hand-built by design. That is the privileged-field miss: one
channel is unlisted. A hand-edited version would let an oracle bump skip re-adjudication.

**Fold.** The loader cross-checks each response's recorded `serverVersion` against its own body's envelope (the JSON
`subsonic-response.serverVersion` and the XML attribute), and refuses a mismatch. The bodies are verbatim, so this
costs one envelope parse.

### C9 (minor): Task 5's digest and version mutations name no fixture state

`plan:445-446`, `:458-460`.

**Defect.** "Take the digest from the spec" fails a test only if the fake `podman image inspect` returns a digest
different from the digest file. "serverVersion from the envelope" needs the same: a fake envelope version that appears
nowhere else. Neither state is named, so a fake that echoes the spec file's digest lets the mutation pass.

**Fold.** Name both states: the fake inspect returns a digest absent from the digest file, and the fake envelope
carries a unique version string. The test asserts that the session holds those exact values.

### C10 (minor): Task 7's no-link test has no fire-on-fixture proof

`plan:516-519`.

**Defect.** The test passes today: `go list -deps ./cmd/dubplate-server` prints no matching path (measured). Its
parser is deliberately separate from Task 2's, and unlike Task 2 (`:353-355`, `:366`) it gets no canned-violation
proof. A wrong prefix literal passes forever.

**Fold.** Factor the matcher to take a listing, feed it a canned listing carrying `github.com/getkin/kin-openapi/...`
and an `internal/harness` path, and assert that both are named. That mirrors `TestHarnessImportCheckFiresOnItsOwnFixture`.

### C11 (minor): Task 9's auth-parameter grep cannot see the format it guards

`plan:595`, `:416-419`.

**Defect.** Session metadata stores parameters as sorted structured data (`:416-418`), so an auth parameter appears as
a JSON member such as `"t": "…"`, never as `t=`. The grep `[?&](t|s|p|apiKey)=` passes vacuously over exactly the
file that matters.

**Fold.** Replace the grep with the acceptance that every committed session loads through Task 4's loader, whose
many-bad states refuse an auth parameter. That test already exists through Task 5 (`:457-458`).

### C12 (minor): the credential criterion is partly untestable and silent on gonic

`plan:379-380`, `:124`, `:618-620`.

**Defect.** "Generated per run and never written to disk" cannot hold literally, because each oracle persists its
user in its own container database. For gonic, the quoted README documents only the `admin`/`admin` default and no
variable that sets a credential. "Not verified" covers Navidrome's first-admin path but not gonic's.

**Fold.** Reword the criterion: never written to a host file the harness writes, never to a session, never to the
digest file. Add gonic's credential path to "Not verified," and accept the default on a loopback-only, per-run
container if no API changes it.

### C13 (minor): two Task 1 promises have no test

`plan:323-325`.

**Defect.** "Decodable by `ffmpeg`" and "under 512 KiB" are stated but neither appears in the test mandate. A gate
test calling `ffmpeg` would add a workstation prerequisite, and the gate paragraph forbids a skipped check.

**Fold.**
- The size cap becomes a test assertion over the seed directory.
- Decodability is evidenced offline. A seed-session test asserts that both oracles report `duration > 0` for every
  song, a fact the recording already holds.

### C14 (minor): Task 7's "untouched real `getAlbum` body passes" names no source

`plan:526-527`.

**Defect.** The plan expects `contract` rows against the oracles (`:552-553`). If the chosen oracle's `getAlbum` body
itself violates the schema, the baseline proof fails for a reason unrelated to the checker.

**Fold.** Name the body as the seed-session response the Task 8 ledger carries no `contract` row against, chosen
after Task 5's recording. If neither oracle's body qualifies, use a hand-minimized valid body.

### C15 (minor): report determinism is promised and never tested

`plan:259-260`, `:549-550`.

**Defect.** The grader's report is promised byte-identical, but no test runs it twice. Go map iteration order would
break this silently.

**Fold.** Grade the seed session twice in the gate test and compare the bytes. With C1's golden in place, this test
comes for free.

### C16 (minor): Task 0's sweep on the main tree must also wait out a live lane 1 executor

`plan:207-213`, `:286-288`.

**Defect.** The sweep runs on master in the main tree "only while no 6d sweep is in flight." That condition is too
narrow. Any live lane 1 executor in that tree makes the conductor a second writer (repo `CLAUDE.md`, "Never launch a
second executor into the SAME tree"). Lane 1 is idle today (STATUS row 1), but 6d rev 2.3 may launch first.

**Fold.** Gate the sweep on the lane registry showing no live lane 1 executor, not only on 6d's sweep.

## Over-ceremony

### O1 (minor): Task 6's Markdown transcription proof has no 2a caller

`plan:497-498`, `:250`.

A Markdown-to-rows converter and its round trip serve 2b's Task 11 only. That breaks the plan's own "no export ahead
of a caller" constraint, or it builds unexported code with no consumer. Move it to Task 11, where slice 1's actual
table shape is known.

### O2 (minor): the search set-plus-top-N rule is built with no session exercising it

`plan:547-548`.

No search endpoint is recorded before slice 4. The rule is proven only on a fixture pair, and its first real consumer
plans against slice 4's record. Defer it to that consumer's plan, or to 2b, and drop one perturbation surface from the
Opus task.

### O3 (minor): the ceiling is derived from priced ceilings, not from spend

`plan:38-40`.

The removal pass priced 8.0M and spent about 3.7M (`docs/HISTORY.md:191`). Scaling from prices puts 9.5M near 1.5
times the likely spend, so the 7.6M flag will not signal drift. Re-derive from actuals at about 0.5M to 0.6M per task,
with a premium for the Opus Task 8 and the two live tasks: roughly a 7M ceiling with a 5.6M flag. The alternative is
to keep 9.5M and state the variance reason in the header.

## What holds

- **Ruling 1's split is sound.** The 2a/2b seam falls exactly on the slice 1 inputs (`:602-608`), and 2b's 4.0M
  re-derives at its Task 0. The recommendation is correct.
- **Pass class `tool` fits.** No task reaches auth, the store, `web/`, or the box. The raised mutation mandate and the
  close's vacuity read are proportionate for a harness that grades six later slices; C1, C4, and C10 show the risk is
  real.
- **The library quotes check out.** The `go help list` quote and the `.test` suffix handling (`:122`, `:352-355`) are
  correct (measured output `pkg [pkg.test]`). kin's `Content.Get` strips media-type parameters, so a `charset` suffix
  is not a false content-type violation. The `MultiError` and `IncludeResponseStatus` quotes match `options.go:22-26`.
- **The per-task gate fits.** The light lane on `--scope-changed` matches `check.sh:27-40`. Task 2 touches no
  `internal/harness/` file, so its gate also skips the conform suite.
