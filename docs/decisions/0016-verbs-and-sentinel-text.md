# 16. Four verbs, and bare sentinel text

## Status

Accepted

## Context

The server used Check, Resolve, Validate, and Verify for overlapping work. `auth.CheckRequest`
verified a session cookie, `serve.CheckProxyHeader` judged a header name's grammar, and
`serve.ResolveConfig` parsed environment strings. A reader could not tell from a name whether a call
read state, touched a credential, or only looked at its argument. Sentinel text followed two
conventions as well: auth, identity, intake, opshttp, and five other packages prefixed their messages
with "pkg: ", and the rest did not. The Go style guides advise against the prefix, because the
caller's wrap supplies the context. The architecture read of 2026-11-23
(`docs/research/2026-11-23-architecture-read.md`) found both, and work `vocabulary` settles them.

## Decision

Each of the four verbs has one meaning in a function or method name.

| Verb | Meaning | Examples |
| --- | --- | --- |
| Parse | Turns text into a value. | `serve.ParseConfig`, `serve.ParseTrustedProxy`, `ops.ParseHandle` |
| Validate | Judges a value against its grammar, with no I/O, and returns an error. | `serve.ValidateProxyHeader`, `intake.ValidateStaging` |
| Check | Probes state outside its arguments (a file, a store, a policy) and reports what it found. | `gate.CheckFile`, `auth.CheckAdmin` |
| Verify | Authenticates a credential. | `auth.VerifyRequest`, `auth.VerifySession`, `Verifier.VerifyIdentity` |

A noun that holds one of these words is not a verb. `intake.CheckWorkerCount` returns the number of
check workers and reads as a noun getter, in the same way `SweepInterval` does.

Sentinel text is a bare lowercase phrase with no package prefix, such as `errors.New("no valid
session")`. The `%w` wrap at the call site supplies the context. A sentinel's text is never the wire
answer: `ops.Error` prints its code, so changing a sentinel's text changes logs, never a client's
response.

## Consequences

`TestSentinelTextIsBare` in `internal/harness/conform/arch_test.go` fails when a non-test Go file
under `cmd` or `internal` calls `errors.New` with text that opens with a lowercase word and a colon.
A new name with one of the four verbs is read against this record in review.
