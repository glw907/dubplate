# `cmd/dubplate-server` architecture read (5c-iii merge side)

Read at HEAD `db1e111`, no plan in context, per CLAUDE.md's "Reviewers" (one
`go-architecture-reader` per touched Go package at each pass's merge).

**Verdict: workmanlike.** Density 0.444 against `cmd/gofmt`'s 0.298 (`web.go` alone: 0.40
against `net/http/fs.go`'s 0.373).

## Findings

- **Magic-link key names kept in agreement by hand.** The same key names are typed
  independently across `serve.go:25-37`, `auth.go:46-60/268-285/365`, and `doctor.go:144`. A
  rename in one place silently splits `Load` from `newServeShell`. **Routed:** 5d-i Task 6
  (template rows) or the architecture cleanup pass.
- **`newServeShell`'s discarded `done` channel.** Created at `serve.go:191`, referenced at
  `:243` and `:313`, but nothing waits on it. Either wait on it in the verb or drop the channel.
- **A 10-positional test seam.** `authHandlersConfig` (`serve.go:349-376`) takes ten positional
  fields; a future caller adding an eleventh risks a silent argument-order bug.
- **`rendered_test.go`'s rate-budget check cannot fail.** `linkRequests` and `consumptions`
  (`:44-68,328-353`) are declared but never set, so `checkRateBudget` has no state that could
  make it fail — the row asserts nothing beyond "an unspent budget passes." **Routed:** 5c-iv,
  since Task 8 adds login-POST rows that spend a link; the refresh read decides whether Task 8
  wires these fields or the check is dropped until it can fail.
- **The `web` subtree's naming splits across two files.** `serve.go:39-63` and `web.go:29-39`
  each name part of the same subtree independently.
- **Doctor rows spread over three files.** `backup.go:284-473`, `store.go:152-260`, and
  `doctor.go:95-127` each carry doctor-row logic for their own domain, with no shared shape.
- **`storeBusyTimeout` returns a 3-result tuple.** `store.go:61` — a typed range-error result
  would read better than a bare triple.
- **`identitySourceStatus` re-runs `Validate`.** `doctor.go:342-411` calls `Validate` again
  rather than reusing the result a caller already computed.
- **Ten auth thresholds live in three hand-written lists.** `auth.go:225-237,299-325,514-559` —
  adding a threshold means updating three places by hand.
- **Favicon paths are named twice.** `web.go:295-298` and `:485-488`. **Routed:** 5c-iv (the
  refresh read decides whether Task 8 or 9 touches this route table and can collapse it in
  passing) or the architecture cleanup pass.
- **`exactPath`/`stylesheetRoute` and the 405 logic are each duplicated.**
- **An unreachable branch.** `auth.go:449-451`'s `u.User` check cannot be reached given the
  caller's validation upstream.
- **`webDocuments`'s member-delete logic** (`web.go:327-330`) duplicates a pattern used
  elsewhere in the file.
- **Two shadow-registration functions** (`shadow.go:94-113`) do near-identical work.
  `ShadowIntegration` is exported (`shadow.go:48`) with no caller outside the package found in
  this read.
- **Comments.** About 132 process-citation lines. `main.go:2-19`'s package doc lists verbs that
  do not exist in the current command tree. Minor duplication in `thresholdStatus` and
  `matcherCommit`.

## Best code in the package

- `web.go:419-448` — static file `Content-Type`/`ETag` handling.
- `web.go:642-709` — `readOperatorCSS`'s descriptor checks (the symlink/parent-symlink/FIFO
  discipline security fold M1 hardened).
- `auth.go:431-444` — `validatePublicURL`'s pre-check ordering.

## Routing

`checkRateBudget`'s vacuity and the favicon-table duplication to 5c-iv (Task 8 adds the rows
that would exercise both; the refresh read decides). The magic-link key-name agreement to 5d-i
Task 6 or the architecture cleanup pass. Everything else to the architecture cleanup pass.
