# Go idiom and architecture read, greenfield tree at 781ddc0

Date: 2026-09-12. Scope: `internal/config`, `internal/identity`, `internal/auth`,
`internal/engine/tagio`, `internal/engine/shadow`, `internal/serve`,
`internal/atomicfile`, `cmd/dubplate-server`. Bar: the `go-conventions` skill and
the standing `idiomatic-beautiful-go` direction, graded against the standard
library rather than against the plans that produced the code.

Method: every non-test file in the eight packages was read. Test files were read
for discipline only. Exported symbols were grepped for callers across the whole
repo before any claim that a surface is unused. Comment density was measured
against four stdlib packages.

## Verdict

The pipeline is producing Go that is correct, well tested, and mechanically
clean, and that is architecturally sound in every place where a real consumer
exists: **sound with nits, where the mechanism is close to exemplary and the
surface discipline is workmanlike.** The good half is not accidental. There is
not one interface declared in the entire tree, not one goroutine outside the
listener fan-out in `internal/serve`, not one `sort.Slice`, not one
`interface{}`, not one defensive nil check on internal state, not one error
string starting with "failed to", and not one em dash. Error handling uses
sentinels where callers branch and `errors.Join` where a caller wants the whole
list. Value types carry behaviour, `ctx` is first, writers and `environ` are
injected at the cobra boundary, and the crypto in `internal/auth` is the work of
someone who understood domain separation and length prefixing before writing it.
Three defects recur in every package and all three have the same origin, which
is that the unit of work is a plan task and the unit of review is that task's
diff. First, the tree exports a large speculative surface with no consumer: the
whole of `internal/identity` beyond one function, sixteen of eighteen `Store`
members in `shadow`, the format registry in `tagio`, `config.OwnedWall`, and
four exported symbols in `serve`. Second, shipped doc comments carry 241 lines of
process citation and run at up to 1.10 comment lines per code line, against a
stdlib band of 0.22 to 0.46. Third, seams exist that only a test uses, and
near-duplicate bodies sit in files that two different tasks wrote. None of this
is fighting the language, and none of it is a correctness risk today. All of it
is invisible to a reviewer reading one diff, which is the process fact the code
is reporting.

## The ten most valuable findings

### 1. A large exported surface has no consumer at all

`internal/identity/identity.go:139-203` defines `Channel`, three channel
constants, `Channel.String`, `Channel.Valid`, `Identity`, and `Identity.Valid`.
Nothing in the repo references any of them. The only used symbol in the package
is `NormalizeAddress`. The same pattern repeats: `config.OwnedWall`
(`internal/config/keys.go:28`) has no caller outside its own test;
`serve.ClientIP`, `serve.NewTrustedProxyConfig`, `serve.NewFrontExemption`, and
`serve.ProxyMode` have no production caller; `tagio.Sniff`,
`tagio.SupportedFormats`, and `tagio.Mapping` have none; sixteen of eighteen
exported `shadow.Store` members are package-internal in practice, two of them
reachable only from tests.

Idiomatic form: unexport until a second package asks. `Identity` and `Channel`
are a rung 6 design sketch, and a sketch costs nothing to keep as a plan
paragraph while it costs a permanent API surface to keep as code. Where the
symbol is used only inside its package, lowercase it; same-package tests still
reach it.

This is the finding with the highest leverage, because a wrong guess here is not
caught by any gate and hardens with every rung that ships around it.

### 2. Doc comments cite the process that produced the code

39 doc comments carry a `ruled, conductor <date>, <id>` trailer. 241 comment
lines in non-test Go cite a rung number, a resolution id, a task number, a
coverage gap, or a lane. Examples: `internal/config/registry.go:33`,
`internal/serve/serve.go:59`, `internal/auth/token.go:181`,
`cmd/dubplate-server/main.go:96`. Worse are the comments that narrate a plan's
own history rather than the code's contract:
`internal/serve/server.go:29-36` explains that a constant "is the constant the
salvage server carried at `_salvage/internal/serve/serve.go:600` that revision 1
of this plan dropped by describing that server as setting a write timeout
alone". `internal/engine/tagio/container.go:66` cites
`plans/2026-09-05-rung-4a-shadow-runner.md:2404-2416` from an exported type's
doc. `internal/engine/shadow/libraryread.go:42` cites a plan file and two line
ranges inside it. `internal/auth/store.go:130` cites
`internal/engine/shadow/store.go:66-70`.

Idiomatic form: state the constraint, cite the symbol, drop the ruling.
`readTimeout` needs one sentence about what it bounds and why 30 seconds. A
caller reading `go doc` cannot resolve "R-V", and a line-number citation into
another package is false the moment that package is edited.

### 3. Comment volume is two to four times the stdlib band

Measured as `//` lines over non-blank non-comment lines, non-test files only:

| package | ratio |
| --- | --- |
| `internal/serve` | 1.10 |
| `internal/identity` | 0.85 |
| `internal/config` | 0.75 |
| `internal/auth` | 0.72 |
| `internal/engine/tagio` | 0.54 |
| `cmd/dubplate-server` | 0.43 |
| `internal/engine/shadow` | 0.41 |
| `internal/atomicfile` | 0.23 |
| `net/http` | 0.46 |
| `os` | 0.34 |
| `archive/zip` | 0.25 |
| `crypto/tls` | 0.22 |

`internal/serve` ships 894 comment lines against 809 code lines. The excess is
not doc comments on exported symbols, which the standard requires. It is
paragraph-length rationale on unexported helpers, rebuttals of alternatives that
were never written, and the process citations of finding 2. The skill's rule is
that a comment earns its place by carrying a reason the code cannot. Most of
these carry a reason the plan already carried.

`internal/atomicfile` is the one package in the stdlib band, and it is also the
easiest package in the tree to read. That is the evidence, not a coincidence.

### 4. `atomicfile` has two copies of one algorithm, and they disagree

`internal/atomicfile/atomicfile.go:16` and `:55` are the same eight steps:
create temp in the target directory, write, sync, close, chmod, rename. The
bodies differ in one line, the write itself. They also differ in durability:
`WriteStream` fsyncs the containing directory after the rename
(`atomicfile.go:82-89`), `WriteFile` does not. Nothing in either doc comment
says the guarantees differ, and `auth.Store.Write` routes records through
`WriteFile` while `EpochStore.Bump` routes through `WriteStream`, so two record
kinds in one leaf have different crash behaviour by accident.

Idiomatic form:

```go
func WriteFile(path string, data []byte, perm os.FileMode) error {
	return WriteStream(path, perm, func(w io.Writer) error {
		_, err := w.Write(data)
		return err
	})
}
```

One implementation, one guarantee, and the asymmetry disappears rather than
being documented.

### 5. `doctor` recovers structure by parsing error text

`cmd/dubplate-server/doctor.go:90` builds the per-key status table from
`config.Load`'s joined error, and `doctor.go:183` recovers the key by
`strings.Cut(err.Error(), ": ")`. The doctor's whole table therefore depends on
the exact prose shape of every error `internal/config` formats. A reworded error
in `config` silently reclassifies a row as "unmentioned and therefore
satisfied", and no test in `config` would notice.

Idiomatic form: `config` returns the structure it already has.

```go
// in config
type KeyError struct {
	Key string
	Err error
}
func (e *KeyError) Error() string { return e.Key + ": " + e.Err.Error() }
func (e *KeyError) Unwrap() error { return e.Err }

// in doctor
var ke *config.KeyError
if errors.As(sub, &ke) { statuses[ke.Key] = statusFor(ke) }
```

`errors.Is(sub, config.ErrMissing)` keeps working unchanged through `Unwrap`.

### 6. `EpochStore` carries a mutex and a subscriber list for a test

`internal/auth/epoch.go:62-63` gives `EpochStore` a `sync.Mutex` and a
`[]func(address string, epoch time.Time)`. `Subscribe` (`epoch.go:136`) has one
caller in the repo, `epoch_test.go:208`. `notify` copies the slice under the
lock on every `Bump`. The package's only concurrency primitive exists to serve a
publish path nothing subscribes to, and the doc comment concedes the point by
calling `Read` the authority and `Subscribe` "the convenience half".

Idiomatic form: delete `Subscribe`, `notify`, `mu`, and `subscribers`. The rung
that needs a notification adds it with its subscriber in the same commit, and
then the mutex is load-bearing. The same call belongs on
`shadow.ReconcileConfig.LibraryReadFn` (`internal/engine/shadow/reconcile.go:35`),
documented as "injected so the test needs no library tree", whose one production
caller always passes `readFLAC`; on `shadow.Store.Resolved`
(`store.go:264`), whose only callers are two test lines; and, more defensibly,
on the `writeTimeout` parameter of `newHTTPServer`
(`internal/serve/server.go:92`), which exists so one test can exhaust a deadline
in milliseconds.

If the seam must stay, the stdlib form is a package-level `var libraryRead =
readFLAC` that a test swaps, not a field every production caller must fill.

### 7. `internal/config` writes six family registrations six times, in C style

`AddReservedClass` (`registry.go:165`), `AddIntegrationClass` (`:213`),
`AddIdentitySource` (`:398`), and `AddThreshold` (`:269`) each re-implement the
same four steps: prefix check per key, within-set duplicate scan, cross-family
scan, claim. `AddIntegrationClass` alone spends 40 lines on it. Separately, six
predicates take the package's own type as their first parameter rather than
being methods on it: `ownedNow` (`keys.go:36`), `reservedClassOf` (`:128`),
`supersessionOf` (`:167`), `integrationClassOf` (`:200`), `thresholdOf`
(`thresholds.go:112`), `identitySourceKey` (`registry.go:442`).

Idiomatic form: one `func (r *Registry) declare(family string, keys ...string)`
carrying the four checks, and six methods `func (r *Registry) ownedNow(name
string) bool` and siblings. The registry pattern itself does pull its weight.
Explicit registration over `init`, a panic on collision on the
`flag`/`database/sql`/`expvar` precedent, and a fresh independent registry per
process are all right calls, and the design is what lets a rung declare its keys
in its own file. It is the six hand-copied bodies that are the cost, not the
idea.

### 8. `tagio` keeps five parallel 69-element lists in sync by hand

The 69 `Field` constants (`tagset.go:16-85`), the 69 `TagSet` fields
(`tagset.go:129-197`), the 69-row `mapping` table (`tagset.go:256-326`), the
69-case read switch (`read.go:86-223`), and the 69-case write switch
(`encode.go:55-186`) must agree. A missing field is caught by a runtime
`panic("tagio: no projection for field %q")`, not by the compiler. Adding one
tag is a five-file edit.

Idiomatic form: the table already keys on `Field`, so give each row the two
functions the switches dispatch to, and let the switches go. Roughly 270 lines
of switch become table rows, and a missed field becomes a nil func a
completeness test catches. The same file's six near-identical
`stringField`/`intField`/`boolField`/`rgGainField`/`rgPeakField`/`r128Field`
bodies (`encode.go:235-317`) collapse into one generic `scalarField[T](row,
*T, func(T) string)`, which is the case generics exist for.

### 9. `shadow` threads a nine-field `Config` where three values would do, and uses out-parameters

`Reevaluate` (`reevaluate.go:77`) takes `Config` and reads one field,
`MatcherCommit`; `cmd/dubplate-server/shadow.go:375` fills a nine-field struct to
pass one string. `reevaluateAlbum` (`reevaluate.go:199-203`) then fabricates a
second synthetic `Config{BaseURL, UserAgent, Transport}` to satisfy
`matchAlbum`, whose signature was frozen by an earlier task. Separately,
`reconcile.go:86,111,154,286` pass `out *[]reconciliation` and a shared `claimed
[]bool` through four helpers, and `reevaluate.go:107-153` selects an output
bucket with `var moved *[]string` pointing at a struct field.

Idiomatic form: `Reevaluate(ctx, s *Store, matcherCommit string, now
time.Time)`; `matchAlbum` takes the four values it reads or the client the
caller built; the reconcile helpers become methods on a small unexported
`reconciler` struct that owns `captures`, `claimed`, and `out`; the bucket
selection names an outcome and appends once at the bottom. Go returns values.

### 10. `cmd/dubplate-server` carries business logic, and exports from `package main`

`shadow.go` holds roughly 300 lines that are not CLI wiring: `caseJSON`
(`:123`), `pendingJSON` (`:158`), `storeArtifacts` (`:192`),
`stageNewlyDiverged` (`:402`), `printShadowReport` (`:437`). The doc comment
explains why, and the reason is sound (the engine may import no harness
package), but the conclusion does not follow: the translation belongs in a small
package both sides may import, not in `main`, where no other binary and no test
outside `package main` can reach it. `shadow.go:48` also exports
`ShadowIntegration` from `package main`, where exporting means nothing.

Smaller, in the same family: `resolveAuthConfig` (`auth.go:207`) returns four
bare `time.Duration` values plus an error and needs named results to stay
readable, where a four-field struct would say what each one is; the nine required
key names in `magicLinkSource` (`auth.go:232-242`) are repeated as string
literals in `validateMagicLinkSource`'s loop (`auth.go:316-321`), so a typo in
one list silently drops a check; `serve.NewStrip(header, sessionCookie string,
...)` (`internal/serve/strip.go:300`) puts two same-typed strings adjacent in a
positional signature; `newCheckCmd` captures `stdout` directly while
`newVersionCmd` uses `cmd.OutOrStdout()`, two sibling files disagreeing about
one convention; `tagio`'s `malformed(path, err.Error())`
(`container.go:213,222,235,248`) stringifies an error to rewrap it and destroys
the chain, where Go 1.20's two-`%w` form preserves it.

## What the pipeline gets right

Three passages, chosen because each shows judgment a linter cannot supply.

**`internal/serve/serve.go:313-371`, `statusCapture`.** A `ResponseWriter`
wrapper that implements `Unwrap` for `http.ResponseController`, passes `Flush`
and `Hijack` through for handlers that type-assert directly, and forwards
`ReadFrom` so a proxied stream still reaches `sendfile(2)`. Wrapping a
`ResponseWriter` correctly is the single most commonly botched thing in Go HTTP
code, and this one is complete. The `ReadFrom` comment names the exact
consequence of omitting it.

**`internal/auth/token.go:414-440`, `sign`, `macInput`, `appendField`.** The MAC
covers a length-prefixed domain-separation label followed by length-prefixed
fields, so no byte can move across a field boundary and carry its signature.
`macMatches` (`:407`) uses `crypto/subtle.ConstantTimeCompare` and says why
`bytes.Equal` is wrong. `VerifyLink` (`:280`) fixes the order as MAC, then
expiry, then nonce, and states that reading the expiry first would act on an
unauthenticated field. The golden vectors in `token_test.go:58` are frozen
literals derived from the written format, not recomputed from the code under
test, and the table's second and third rows are exactly the addresses a naive
join would break.

**`internal/engine/tagio/write.go:302-317`, `copyFrameRegion`.** An
`io.NewSectionReader` into `io.Copy`, so audio frames are never materialized and
a multi-gigabyte rewrite costs a fixed amount of memory. The doc comment says
that and stops.

Runners-up worth naming: `shadow/classify.go:41-78`, a flat switch over a closed
enumeration whose default arm defers and counts rather than inventing a verdict;
`shadow/oraclereport.go:159-174`, nine lines of `strings.SplitSeq` and
`strings.CutPrefix` that read as current-decade Go; `tagio/encode.go:452-496`,
which reimplements Unicode's maximal-subpart rule and explains why
`utf8.DecodeRuneInString` is the wrong tool there.

Test discipline across the tree is genuinely good and should not be disturbed.
Every `t.Skip` found is a conditional root or missing-fixture guard, never a
placeholder. No fake returns `nil` unconditionally from a fallible method;
`shadow`'s `failingCountingTransport` injects per-call errors. No expectation is
computed from the system under test. Goldens are confined to byte-level FLAC and
wire-format fixtures, where the bytes are the value.

## Process changes the code argues for

### 1. Add a per-rung whole-package architecture read, with no plan in context

Every finding above is invisible in the diff that introduced it. The duplicate
`atomicfile` body, the six hand-copied registration bodies, the five parallel
69-element lists, the sixteen over-exported `Store` members, and the zero-caller
`identity.Identity` are all properties of a package, not of a change. The
`diff-reviewer` reads one task's diff against that task's acceptance criteria,
which is the right job and a different job.

Concretely: at each rung's merge, dispatch one Opus reviewer per touched package
with the package's full non-test source, no plan, no spec, and no task list. Grade
on five axes only: exported surface against grepped callers, duplication within
and across the package, whether the file split reads as concerns or as
deliverables, seams that only tests use, and comment density against a named
stdlib package. The absence of plan context is the point, because the plan is
what makes a speculative export look justified.

### 2. Make "no consumer, no export" a gate leg

Verify and enable `unused` with `exported-is-used: false` in `.golangci.yml`,
which works for packages under `internal/`. Where that setting proves too coarse,
a small gate leg is cheap: for every exported identifier in `internal/`, fail if
its only references are its own file and `_test.go` files. Either one would have
caught `identity.Identity`, `identity.Channel`, `config.OwnedWall`,
`tagio.Sniff`, `tagio.SupportedFormats`, `tagio.Mapping`, `serve.ClientIP`,
`serve.NewFrontExemption`, and sixteen `shadow.Store` members before they
shipped. Pair it with a plan rule: a forward hook for 2.0 is named in the plan
and built by the rung that consumes it, never exported ahead of its consumer.

### 3. Fail the gate on process citations, and give the simplifier a mechanical brief

Extend `scripts/vale-comments.sh` or add a grep leg over `//` lines in non-test
Go that fails on `ruled, conductor`, `Resolution R`, `rung \d`, `Task \d`,
`\(C\d+\)`, `plans/.*\.md:\d`, and `_salvage/.*:\d+`. The em dash ban already
works, which is the proof that a mechanical rule holds where an instruction does
not: zero em dashes across 12,000 lines, against 241 citation lines and 39
`ruled, conductor` trailers.

The batched `code-simplifier` round is currently briefed on clarity, and clarity
is exactly what these comments claim to serve. Give it four named patterns to
hunt instead: two function bodies in one package that differ by one statement; a
free function whose first parameter is the package's own type; a function
returning three or more results of the same type; and a struct field, method, or
parameter whose only non-test caller is the package that declares it. All four
are decidable by reading, and all four appear above more than once.
