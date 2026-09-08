---
name: feedback-test-echoes-code
description: A unit test's "want" value must not be built with the same encoding/helper the code under test calls, or the assertion proves nothing.
metadata:
  type: feedback
---

A diff-reviewer finding on dubplate's cmd/harness/capture_test.go: `TestReleaseRequestIsCanonical`
built its `want` string with `url.QueryEscape(maximalReleaseInclude)`, the exact same call
`releaseRequest` makes internally. The test therefore could never fail even if the
production code regressed to joining `inc=` values with a literal `+` instead of the
required wire form `%2B`, so both sides of the comparison would drift together. The one fact
the surrounding fix existed to lock in (join separators must be percent-escaped on the wire,
never literal) was asserted by no test in the repo; a regression would only have surfaced
as a full live capture-derive loop failure days later.

**Why:** the reviewer proved this by grepping the repo for any other test pinning `%2B`
in a release request and finding none.

**How to apply:** when a test's expected value is built by calling the same function,
constant-building expression, or encoding helper the implementation under test also calls,
stop and replace it with a literal string (or an independently-computed value). This applies
most often to encoding/escaping fixes (URL-encoding, JSON marshaling, path canonicalization)
where it is tempting to reuse the "obviously correct" helper on both sides. Add an explicit
negative assertion too when there is a known-wrong prior form (e.g. `strings.Contains(got, "+")`
must be false) so the test names the specific regression it guards against, not just a diff
from itself.
