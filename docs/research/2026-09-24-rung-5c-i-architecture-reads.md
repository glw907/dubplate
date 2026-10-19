# Pass 5c-i architecture reads

On 2026-09-24 three fresh-context `go-architecture-reader` dispatches (Opus) each read one package at HEAD `3228a21`, with no plan in context.
All three packages graded **workmanlike**: the mechanism is strong in each, and the costs are exported surface ahead of a
caller, duplication, and comment density. Every finding routes to the **architecture cleanup pass after
5c-i**, joined with the staged comment-cleanup task, as the 5b-ii reads did
(`2026-09-23-rung-5b-ii-architecture-reads.md`). Nothing here blocks 5c-ii.

## internal/auth

Comment density is 0.709, against 0.352 for `net/http/cookiejar`. It includes 56 process-citation lines.

- Exported surface. `Seam`, `VerifiedIdentity`, `MagicLinkSeam`, and `VerifyIdentity` have no caller;
  `serve.go:247` discards the one construction. `EpochStore.Bump`, `Read`, and `Subscribe`, and
  `ErrEpochBackward`, have no non-test caller. The credential primitives have only in-package callers:
  `MintLink`, `VerifyLink`, `LinkURL`, `MintSession`, `RefreshSession`, `VerifySession`, the cookie
  helpers, `Insert`, `Lookup`, `CheckSession`, `CheckLink`, `AllowConsumption`, and `AllowLinkMail`.
  Sixteen sentinels have no `errors.Is` outside the package. `StoreDir`, `NoncePath`, `EpochPath`, and
  `RecordID` are in-package only.
- Test-only seams:
  - `RunSweep`'s injected `sweep` and `now`.
  - The discarded `NonceSweep` from `OpenNonceStore`.
  - `HandlersConfig.Now`.
  - `keyedLimiter.size`, which belongs in a test file.
- Duplication:
  - `seam.go:119-122` restates `Member.principal`.
  - `CheckSession` and `CheckLink` differ by three values.
  - The refused-rune predicate has three bodies: `identity.go:70`, `roster.go:649`, and `roster.go:674`.
  - `normalizeMemberAddress` is a pass-through, and the package bypasses it.
  - `syncDir` restates `atomicfile`'s directory fsync.
  - The `ParseInt`-then-`time.Unix` pair appears four times.
- File split:
  - Move the signed-value codec from `token.go:236-335` into `signed.go`.
  - Split `roster.go` (862 lines) into read, write, and validate files.
- Comments. Strip the citations, several of which are stale. Drop the branch for the unreachable
  `rand.Read` error; since Go 1.24 it never returns an error.
- Test discipline. `encodeSessionUnderLabel` hand-copies the codec. `foreignLabelFixture` restates
  `sessionPurpose`.
- Best passages: `evictRefilledOrAdmit`, `linkFields` with the MAC computed over the bytes that
  arrived, and `lockRosterDir` locking the directory.

## internal/identity

Comment density is 1.106, against 0.392 for `net/netip`.

- Exported surface. `MaxAddressLength`, `Channel.Valid`, and `ErrInvalidAddress` have no production
  caller outside the package. `Identity.Valid` has no production caller yet. `Principal`,
  `DisplayName`, and `Source` are written and never read.
- Duplication. `addressRefusesRune` is copied in `internal/auth`. The fix is one exported
  predicate owned here.
- Test discipline. Two tests cannot fail: `TestIdentity_PrincipalDistinctFromAddress` and
  `TestIdentity_AddressMatchesNormalizer`.
- Comments. The package doc cites rungs, and the helper docs repeat the category list three times.
- Best passages: the fixed-point re-parse, category-based rune refusal, and `iota + 1` channels.

## cmd/dubplate-server

Comment density is 0.445, against 0.097 for `cmd/go/internal/envcmd`. It includes about 100
process-citation lines.

- Duplication. Doctor rebuilds config resolution in two ways: by cutting strings from `config.Load`'s
  joined error, and through a typed `sourceKeyError`. The fix is a keyed leaf error exported from
  `internal/config`, which makes it a cleanup-pass item, since R-5c17 kept 5c-i out of `internal/config`.
- A false comment. `doctor.go:329-333` says every magic-link refusal carries a key.
  `validateSMTPPort` and `validateSMTPFrom` return unkeyed errors. The output is still correct, because
  the string sweep catches those keys.
- Hand-kept key lists. The magic-link key names are spelled in parallel lists across `serve.go`,
  `auth.go`, and `doctor.go`.
- Test-only seams:
  - `newServeShell` starts a goroutine and returns its `done` channel.
  - The discarded `NewMagicLinkSeam`.
  - The `stdout`/`stderr` parameters on every verb, where `cmd.OutOrStdout()` does the job.
  - `matcherCommit` threaded through as a parameter.
- Cobra silence flags. `SilenceUsage` and `SilenceErrors` are restated on nine subcommands; the root
  already sets both.
- Smaller repetition:
  - The `unstamped` check appears twice.
  - The shadow registration is split into two functions.
  - The `checkDurationBounds` accumulator is awkward to call.
  - There are 16 per-row error checks where a `bufio` sticky writer would need one.
- File split. The key constants belong in `auth.go`, and the shared helpers in `env.go`.
- Comments. The package doc is stale: it omits `member` and lists verbs that do not exist yet. It
  carries line citations that have rotted.
- Test discipline. One assertion in `serve_test.go:537-543` cannot fail on its own.
- Best passages: `validatePublicURL`'s `@` check before parsing, the bootstrap race in `addMember`,
  and `resolveActor` never reading `$USER`.

## Conductor dispositions

- **Refused.** Deleting `Identity.Valid` and `Channel.Valid` is refused. The 5c-ii admin gate is
  planned as their production caller, because it requires `Identity.Valid` to refuse the command-line
  channel (5c-i plan, Task 8's 5c-ii handoff).
- **Weighed with a planned caller in view.** Deleting the seam and `EpochStore.Bump` is a real
  finding under R-5c20. Still, 5c-ii's page is the seam's first consumer, and `Bump` arrives with the
  offboarding verb, so the cleanup pass weighs them against those callers and does not delete blind.
- **Routed as recorded.** Every other finding goes to the cleanup pass as recorded here.
- **Owed to that pass.** The false comment at `doctor.go:329-333` is the one correctness-adjacent
  item: either key the two SMTP refusals or correct the comment.
