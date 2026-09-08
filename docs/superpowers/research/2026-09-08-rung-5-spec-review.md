# Adversarial review: the rung 5 identity and auth spec

Two fresh-context reviews (2026-09-08) of
`specs/2026-09-08-rung-5-identity-auth.md` at revision 1, one through the
web-authentication-security lens and one through the
spec-integrity-and-scope lens, merged here and folded as revision 2 of the
spec the same day. The merged "Verified sound" list below is the
do-not-re-litigate record for every rung 5 pass plan.

Findings are renumbered across both reviews. Every finding was re-verified
before folding: tree claims against the worktree at `be27434`, stdlib
claims against the installed toolchain (`go1.27.1`), and document claims
against the ratified texts they cite. Nothing in the repo was written by
the review. Three findings survive as blockers, twelve as majors, and
fourteen as minors. No finding is rejected outright. One finding is
downgraded from major to minor and four sub-claims inside surviving
findings are corrected, all recorded below with the measurement.

Prior-art figures (module versions, publication dates, importer counts)
were not re-measured. They come from PRA, which read them on 2026-09-08,
and the review takes them as given.

## Downgraded and corrected, with the measurement

**Downgraded to minor: the ingress strip against a forward-auth front.**
The security lens filed as a major that the strip deletes the header a
front placed in front of dubplate would assert, so a deployment selecting
a front as its identity source loses its identity before any handler runs.
The mechanism is real. The severity is not, because no such deployment is
buildable at 1.0: the spec declares no front keys (spec:1054-1056), no
front code exists, and BS:499-503's rule is that the inactive source is
never consulted. There is no shipped defect, only design rework at the
rung that builds a front. It folds as m1, which asks for the carve-out
rule to be written now because rung 5 fixes the strip's shape and PA:80
closes `internal/serve` to other lanes after this rung.

**Corrected: the CSRF scope is stated once, not three ways that
disagree.** The security lens read spec:658-666 as carrying three
conflicting scopes. Re-read, it carries one rule with an enumeration of
today's members and a statement about where the rule lives: "every
cookie-carrying, state-changing surface: the login POST, the consumption
POST, logout, and the proxy path", stated at the shell so a rung 7 or
rung 14 surface inherits it. That is consistent. What survives, and is
the whole of M4, is that R15's justification ("All four protected
surfaces are browser-only by construction", spec:302-303) is false of the
proxy path.

**Corrected: the branding key names are already gated.** The security
lens reported that nothing constrains operator branding values or names.
DL's G8 asserts the generated stylesheet "declares only names drawn from
R2.1's set and emits an empty file when nothing is configured", and the
spec carries that restriction at :941-944. The name half is covered. The
value half and the composed `data-theme` value are not, and m5 is scoped
to those two.

**Corrected: the revocation event's address payload is not a hook 5
violation.** The security lens argued the seam's event carries an address
that "S20 hook 5 says per-user rows must never key on". S20:367-369 rules
that per-user *rows* key on rung 6's surrogate id. An in-process event
payload is not a row, and R4 already rules the epoch keys on the address
at rung 5 by necessity, with the re-key recorded as rung 6's obligation.
The read-versus-publish defect is untouched by the correction and folds
as M8.

**Corrected: the interstitial's no-store rule is stated, and collides
rather than missing.** The security lens implied the consumption
response's cache directive is unstated. Spec:576-577 does state it:
`Cache-Control: no-store` on the GET that renders the interstitial. The
defect is a collision, because R4.5's embed path rules the served
document `no-cache` at :789 and both routes serve the same
client-rendered document under R1.6. M3 names the collision rather than
an omission.

---

## BLOCKERS

**B1. The spec never separates the peer address from the reported client
IP, so the trusted-proxy header, the audit log, and the verify endpoint's
allowlist are all spoofable.** (Security lens.)
Spec:641-642 says limits key on "the real client IP, read from the
trusted-proxy pair when R22's integration is enabled and from
`RemoteAddr` when it is not", carrying BS:478-481 verbatim. R22
(spec:349-353) and the integration class (spec:948-950) name "an enable
flag plus the proxy addresses and the client-address header" and never
state what the proxy addresses are checked against. No clause says the
client-address header is honored only when `RemoteAddr` is one of the
configured proxy addresses, and none says how a multi-hop
`X-Forwarded-For` is parsed. SEC filed exactly this as a blocker: "the
spec must state that the header is only honored when the peer is the
configured trusted proxy, or it becomes a trivially spoofed limit key."
BS dropped that condition when it absorbed the finding.

The silence reaches three surfaces. The limit key becomes
attacker-chosen. The audit requirement at spec:647-648 ("Every auth event
is logged with the real client IP") then logs an attacker-chosen value,
destroying the only audit trail SEC's Access-retirement minor names. And
the forward-auth verify endpoint's class (spec:952-954) is "an enable
flag plus the addresses permitted to make the verify subrequest", with no
statement that the permission check reads `RemoteAddr` rather than the
client-IP-mode result, so with trusted-proxy mode enabled an internet
client can present the proxy's address in the header and reach the
identity oracle R21 itself calls one. Compounding it, the strip set
(spec:380-404) covers only forward-auth identity spellings;
`X-Forwarded-For`, `X-Real-IP`, and `Forwarded` are absent, though SEC's
proxy-hygiene major requires deleting them on ingress.
*Fix:* define two values in the trust-boundary section. The peer address
is `RemoteAddr`, it is the only input to any trust decision, and it is
what the trusted-proxy allowlist and the verify endpoint's
permitted-addresses list compare against. The reported client IP derives
from the configured client-address header only when the peer address
matches a configured proxy address, parses right to left taking the
rightmost untrusted hop, and is used for rate limiting and logging alone.
State that the strip removes `X-Forwarded-For`, `X-Real-IP`, and
`Forwarded` on every path unless trusted-proxy mode is enabled and the
peer matches. Add to 5b's exit gate: a request from a non-allowlisted
peer carrying a forged client-address header is rate-limited and logged
under its peer address, and a verify subrequest from a non-allowlisted
peer carrying a forged client-address header is refused.

**B2. The request-link endpoint carries no per-client-IP limit, so
enumeration and outbound mail are unbounded across distinct addresses.**
(Security lens.)
Fork 3 (spec:146-152) ratifies "five link requests per hour per address,
and thirty consumption attempts per hour per client address", and the
threshold list (spec:930-934) enumerates the same two keys and no third.
The link-request limiter is therefore keyed on the submitted email alone.
An attacker who never repeats an address never trips it. Roster
enumeration is unlimited, and every roster hit sends real mail, so the
SMTP credential's quota and the sender's reputation are the actual limit.
That defeats the mitigation fork 4 rests on, stated at spec:158-159 ("the
rate limits bound what enumeration could probe") and again in the threat
model at spec:1229. It contradicts BS:478-479 ("Limits key on the real
client IP") and the spec's own sentence at :641-642. S20:247-249 records
the settled shape for the 2.0 surface as "per account and per address",
two keys. 5b's exit gate (spec:1289) asserts only the limiter map's bound
under a flood, which is R16's memory property, so a build with no per-IP
link limit passes the pass. B1 and B2 are separate defects on one
mechanism: B1 is the key being spoofable, B2 is the key being absent.
*Fix:* add a per-client-IP link-request threshold beside the per-address
one, with its own default, and state that both are checked and either one
trips. Add a global outbound-send ceiling per interval so an attacker
rotating addresses cannot turn the login form into a mail relay against
the roster. Restate fork 4's recommendation with the pair as its stated
mitigation, since the kind answer is being ratified on the strength of
it. Add to 5b's exit gate: N distinct addresses submitted from one client
address trip the per-IP limit and send no further mail.

**B3. "The wall proves presence" is false against the landed registry. A
registered wall row is neither required by `Load` nor read anywhere, and
`doctor` prints it "ok" while it is unset.** (Integrity lens.)
R5 (spec:214-223) rests the identity-source design on two claims: "The
wall proves presence. The auth package proves the value names a built
source and fails closed on anything else." Neither holds as the tree
stands. `Registry.AddWallRow` (`internal/config/registry.go:138-152`)
appends the row and claims the name, and creates no requirement. `Load`
(`internal/config/config.go:241-279`) requires exactly four hard-coded
rows (`DUBPLATE_LIBRARY_ROOT`, `DUBPLATE_STATE_DIR`,
`DUBPLATE_SITE_NAME`, `DUBPLATE_ADMIN_EMAIL`) and resolves nothing else
except an enabled integration class's `Required` keys. `Config`
(`config.go:66-86`) carries no generic wall-value accessor, and its doc
comment states the gap: "Rung 1 carries the owned wall only; the
identity-source selector joins its fields at its own rung."

The consequence in `doctor` is concrete. `status := cmp.Or(row.Pending,
statuses[row.Name], "ok")` (`cmd/dubplate-server/doctor.go:34`) prints
"ok" for any row `Load` never mentions, and `Load` returns nil, so
`doctor` exits 0 reporting a satisfied identity source on a deployment
where `DUBPLATE_IDENTITY_SOURCE` is unset. The same gap swallows the
eleven keys in "Required when magic-link is the active source"
(spec:897-911): nothing requires or resolves them, while the package
contract says `internal/auth` "reads configuration through values passed
in, never from the environment directly" (spec:861-862) and spec:913-918
declines the integration-class shape, which is the only generic
resolution path the registry ships. A plan built on revision 1 would ship
a security-critical selector that no code requires, no code reads, and
`doctor` reports green.
*Fix:* state that rung 5 edits `internal/config/config.go`: a `Config`
field plus a `w.require` call for the selector, and a named resolution
path for the correlated set. Add `internal/config/config.go` and `Load`'s
owned-row resolution to the contested-surface table beside the existing
`keys.go` removals row, put the edit in 5b's scope, add a 5b exit-gate
assertion that `doctor` exits nonzero and prints the selector as missing
when it is unset, and add a staleness question asking whether `Load`
still hard-codes its four owned requires.

---

## MAJORS

**M1. Link consumption re-checks nothing, so an outstanding link survives
offboarding and mints a session an epoch bump does not refuse.**
(Security lens.)
The consumption section (spec:562-586) fixes the order as verify the MAC,
then the expiry, then insert the nonce, and says nothing about consulting
the roster or the epoch. The token carries the address (spec:553-555), so
the address is trusted from the token. Revocation is defined at
spec:605-607 as refusing a session whose issue time predates the epoch
for its address. Offboarding therefore has a hole: remove the person from
the roster and bump their epoch, and any link minted before the bump
still verifies, still passes expiry, still inserts a fresh nonce, and
mints a session whose issue time is later than the epoch, which the epoch
check then permits for the full session lifetime. The window is the link
expiry, 15 minutes by default under fork 3, and the attacker chooses when
to click. The threat model asserts the opposite at spec:1219-1220 ("an
epoch bump refuses every session older than it"). The related silence is
that the spec never says an established session is re-checked against the
roster on each request, which the salvage design did
(`_salvage/internal/config/contributors.go:93`).
*Fix:* state in the consumption section that the POST re-resolves the
token's address against the loaded roster and fails with the one generic
outcome when it is absent, and that it compares the token's own mint time
against the address's epoch and refuses a token minted before it. State
separately whether the per-request session path re-checks roster
membership, and if it does not, say so in the residuals beside the
streamer-deprovision step. Add to 5b's exit gate: a link minted before an
epoch bump is refused after it, and a link for an address removed from
the roster is refused.

**M2. The spec carries two pruning mechanisms for the nonce store and
does not say which governs, and one of the two readings restores replay.**
(Security lens, evidence re-scoped.)
The state-directory section (spec:983-988) says each nonce record carries
"its expiry" and that "a sweep prunes expired records at startup and on
an interval". The threshold list (spec:930-933) separately names "the
nonce retention window" as a tunable with its own default. Two readings
follow. If the sweep prunes on each record's carried expiry, single use
holds and the retention threshold is redundant. If it prunes on the
retention window, an adopter who raises the link expiry (which the
universality rule at spec:1532-1534 invites) while leaving retention at
its shipped default gets consumed nonces pruned before their tokens
expire, and every one becomes replayable. Single use is the property R11
spends a create-exclusive insert and two fsyncs to buy, and the threat
model claims it unconditionally at spec:1218-1219. A configuration the
config layer accepts without complaint must not be able to defeat it.
*Fix:* pick one mechanism. Either prune on the record's carried expiry
and retire the separate threshold, or keep both and validate at config
load that retention is greater than or equal to the link expiry, failing
the load by name when it is not, on the secret-length-floor precedent.
State the invariant in the state-directory section beside the sweep. Add
to 5b's exit gate: a configuration with retention below the link expiry
fails to load.

**M3. The claim that a token never reaches a logged URL is false, and the
response headers that would bound the leak are unowned.** (Security lens.)
The magic link is a URL carrying the token, so the GET request line
carries it into the server's own access log, the tunnel's log, any
intermediary's log, and browser history. The spec asserts otherwise twice,
at spec:579-580 ("Tokens never appear in a URL that gets logged, and the
GET/POST split is what makes that true") and in the threat model at
spec:1221-1222. The claim originates at BS:437-439, which absorbed SEC's
consumption blocker and kept only half of it. The GET/POST split closes
that blocker's prefetch-burn and login-CSRF halves. It does not close the
third half, which SEC states as "the token rides the URL into cloudflared
logs, Cloudflare's request logs, and browser history". SEC asks for the
mitigation by name, `Referrer-Policy: no-referrer` on the auth pages, and
the spec never mentions `Referrer-Policy`. More broadly, the serve
package's stated responsibilities (spec:865-870) list only the CSP, while
the salvage server set `nosniff`, `Referrer-Policy: no-referrer`, and
`X-Frame-Options: DENY` alongside it
(`_salvage/internal/serve/serve.go:119-128`), and DL R5.6 covers CSP
directives only. A second seam widens the leak: under R1.6 the
interstitial is a client-rendered page, so the token is read from
`window.location` by JavaScript rather than placed in a hidden field by
the server, and the two cache rules collide as the corrections section
records.
*Fix:* replace both claims with what is true and what is done about it.
State that the consumption route logs the path and never `RawQuery`, that
the whole served surface carries `nosniff` and `Referrer-Policy:
no-referrer` as a named serve responsibility, that the consumption
route's document is served `no-store` while every other document is
`no-cache` under R4.5, and that the residual (a token in mail-provider
link scanners and in browser history, bounded by single use and a short
expiry) is recorded in the residuals section. Name in the web section how
the client-rendered interstitial obtains the token and that it clears it
from the address bar after reading it. Write a fourth erratum against
BS:437-439. Add to 5c's exit gate: the consumption route's response
carries `no-store`, and no log record emitted on that route contains the
token.

**M4. The blanket origin check as scoped would refuse non-browser clients
on the proxy path, including the `/rest/*` bypass and rung 14's API.**
(Security lens.)
R15 (spec:300-308) tightens `net/http.CrossOriginProtection` to deny a
state-changing request carrying neither `Origin` nor `Sec-Fetch-Site`,
and justifies it with "All four protected surfaces are browser-only by
construction". The proxy path is not browser-only. It carries `/rest/*`,
which the spec itself describes at :652-654 as an anonymous surface for
native clients whose authentication is the streamer's own, and native
Subsonic clients send neither header. S20:284-285 puts the whole 2.0
OpenSubsonic surface under the same prefix and states that the ingress
strip set still applies to it. Rung 14's non-browser token acquisition
(spec:1117-1119) inherits the same shell rule.

Measured against the installed toolchain, the stdlib's own contract
sharpens the shape of the defect. `CrossOriginProtection` always allows
GET, HEAD, and OPTIONS, so a Subsonic client issuing `star` or `scrobble`
over GET is never checked, while one issuing the same call over POST is
denied. The mechanism for the carve-out exists:
`AddInsecureBypassPattern`, alongside `Check`, `Handler`, and
`SetDenyHandler`.
*Fix:* state the scope once as cookie-authenticated state-changing
requests, and state that a request authenticated by a bearer credential
rather than by the session cookie is exempt, which is the carve-out rung
14 and 2.0 need. Name `AddInsecureBypassPattern` as the mechanism, and
state that a bypassed path may not read the session cookie for
authentication. Say explicitly what the rule does on the fallthrough
before rung 12 exists, so rung 12 inherits a decided contract rather than
a wrapper it must rewrite.

**M5. The redirect-validation rule omits the protocol-relative and
backslash-authority forms, and the spec never says whether a
caller-supplied destination exists at rung 5.** (Security lens.)
The post-login destination section (spec:694-702) states the rule as
"single-decode, relative-only, no scheme, no `..`, and no backslash
variants", carrying BS:509-511 verbatim. SEC lists the bypasses a rule of
that shape admits and names the first one first: `//evil.example`
satisfies relative-only and contains no scheme, and every browser treats
it as an absolute URL. `/\evil.example` and `/\/evil.example` are the
sibling forms. SEC's own recommendation is positive, "accept only a value
matching `^/[^/\\]` after a single decode", which the negative
enumeration does not reproduce. The spec is also silent on whether rung 5
accepts a caller-supplied destination at all. MD:102-104 had consumption
redirect to the originally requested path, the spec's version at :696-698
describes only driver-declared and default destinations, and nothing says
how a requested destination would survive the mail round trip and the
interstitial POST.
*Fix:* state the rule positively. After a single decode, accept only a
value whose first byte is `/` and whose second byte is neither `/` nor a
backslash, and reject any value carrying a scheme, a `..` segment, a
control character, or a percent-encoded separator. State plainly whether
rung 5 accepts a caller-supplied destination. If it does not, say so, so
rung 12 knows it is introducing the parameter. If it does, say where it
travels, and require that it ride inside the MAC-covered token rather
than as a free query parameter on the consumption POST.

**M6. `doctor`'s boundary probe is a compile-time tautology, so B2's
requirement to prove the boundary is unprovable as scoped.** (Security
lens.)
R6 (spec:222-232) lands B2's boundary proof as a `doctor` row whose
"probes run in-process against the assembled handler through `httptest`,
so `doctor` makes no network call and needs no running server, and the
same probe function is the gate-resident test". A probe that constructs
the handler in the process it was compiled into can only report on the
code it was compiled from. It cannot fail for any deployment reason: not
a bind address exposing an unstripped path, not a proxy in front
re-adding a header, not a unit running an older binary. 5d's exit gate
confirms it at spec:1291, where the only stated failure mode is "a
boundary-probe row that fails when the strip is disabled in a test
build". BS:422-425 asks `doctor` to prove "that a spoofed identity header
from a non-trusted source is rejected", and a source is a deployment
concept. The spec's hedge at :230-232 concedes half of this while still
presenting the row as the proof.
*Fix:* make the probe real or say what it is. The real version dials the
configured bind address of the running service over loopback with a
spoofed header set and reports what came back, which fails when a
deployment is wrong. The honest version keeps the in-process check,
labels the row a build-time assertion rather than a boundary proof, and
adds an erratum against BS:422-425 recording that B2's proof is a
gate-resident test plus adopter documentation. Do not leave a row that
reports pass to an operator who reads it as evidence about their
deployment.

**M7. The epoch record is named from an email address with no id
validation, and the naming rule is applied only to the nonce store.**
(Security lens.)
The state-directory section says of the nonce store (spec:983-988) that
"the record name is derived from the nonce, and the store refuses any id
that is empty, `.`, `..`, or carries a separator, on rung 4's `validID`
precedent". The epoch record gets one sentence at :990-991, "One record
per normalized address (R4), written with `atomicfile.WriteFile`", with
no naming rule and no id validation. An address local part may legally
contain `/`, `.`, and `..`, so a record named from a roster address is a
path-traversal shape derived from admin-authored YAML, which is the class
rung 4's `validID` comment calls a loud bug rather than input. A
directory of files named after addresses also puts the roster in
cleartext filenames, and it makes R4's later re-key onto rung 6's
surrogate id a rename of every record rather than a change of one
derivation. The confinement test at spec:999-1002 is stated for the store
generally, and the argument it rests on is the id rule the epoch record
does not have.
*Fix:* state that both record names are derived through one function, the
hex of a digest over the normalized address for the epoch and over the
nonce for the nonce record, and that both pass the same `validID` check.
Say that the digest is what makes R4's re-key a change of one derivation.
Extend the confinement test to the epoch store, so the unit's
`ReadWritePaths` grant rests on the same static argument for both records.

**M8. The revocation seam is publish-only, which cannot satisfy the 2.0
guarantee it is named for.** (Security lens.)
The spec defines the seam at :610-614 as "one observable event carrying
the address, published where 2.0 can subscribe", citing S20:245-246.
S20:245-246 states the guarantee from the other side: "Every app password
carries the user's valid-after epoch, so a roster-level revoke
invalidates all of a person's devices at once and offboarding needs no
runbook step." An in-process event cannot deliver that. A subscriber not
running when the event fires never learns of it, a process restart loses
it, and a credential store that mints an app password after the event has
no way to read the epoch it must stamp on it. The authority is the
durable epoch record, and the seam as specified exposes no way to read it.
*Fix:* specify the seam as a read plus a change notification. The read
returns the current epoch for a resolved identity, which is what a
credential store stamps and what a later check compares against. The
notification is the optional half that lets a subscriber act promptly.
State that the read is the authority and the event is a convenience, and
record in R4 that rung 6's re-key changes the key of both halves together.

**M9. The env-template paragraph covers secrets only, so the required
non-secret rows have no landing state, and the live shadow unit's `Load`
is not considered.** (Integrity lens.)
Spec:967-975 says "Rung 5 adds no name it has not minted. Every rung 5
secret lands as a commented pending row ... and rung 16 mints all three."
Rung 5 makes eight non-secret names required: `DUBPLATE_IDENTITY_SOURCE`,
`DUBPLATE_ROSTER_FILE`, `DUBPLATE_PUBLIC_URL`, and
`DUBPLATE_SMTP_HOST/PORT/USER/FROM` (spec:889-911). The template's header
states three row states, one of which is a literal `NAME=value` line for
a value that will never be minted
(`env/dubplate-server.env.template:8-16`), so these rows have a home the
spec does not assign them. The omission has a production edge:
`dubplate-server shadow` calls `config.Load` with the same registry
(`cmd/dubplate-server/shadow.go:257`, `:324`, `:361`), so once the
selector is genuinely required (B3), the box's shadow verb fails at
`Load` until `/etc/dubplate/server.env` carries the row, and `deploy.sh`
renders that file from this template. 5d deploys to the box.
*Fix:* add a paragraph to "The environment template" naming the row state
each required non-secret key takes at rung 5 and the rung that flips it,
and state the cross-rung consequence: the shadow verb shares the
registry, so a newly required owned row changes what the live unit needs
at its next deploy. Make it a 5d exit-gate line: `deploy.sh --check`
renders without an unresolved name, and `dubplate-server doctor` on the
rendered environment exits 0.

**M10. No composed theme exists at rung 5, so "the standard's default
block" names nothing and 5c's "G1 through G10 passing" is unprovable.**
(Integrity lens.)
Spec:829-831 says "rung 7's roster supersedes the standard's default
block without touching rung 5's markup. Rung 5 authors no theme roster
entry." The design-language standard at revision 2 ships no default
block. It defines the 32 names (R2.1), the composed identity
`dubplate-gruvbox` (DL:310-316), and the gates, and DL section 7 assigns
"the shipped theme roster" to rung 7. A grep of the standard for a values
block returns only the variable-name listing and the two prose mentions
of gruvbox. The sentence traces to RE cell (5,7), whose paste-ready row
assumed "Phase 0c's standard ships one default block", an expectation 0c
did not meet. That same cell names the gap as the cell's substance: "Rung
5's pages render unstyled without at least one block, and rung 5 must not
author a roster entry."

Three of the ten gates rung 5's leg must pass need a theme file to mean
anything. G2 asserts each palette block declares the 21 palette names,
each match block the 11 structural names, and that "the `dubplate` match
block's `--size-field` meets R3.7's floor". G9 computes contrast ratios
"from each composed theme's declared values". G10 reads a computed
structural variable "under a built-in palette, against the composed
identity R3.6 defines". DL's empty-iteration guard is written for G3
through G7 alone; G1, G2, G8, G9, and G10 carry none. With no theme
authored, G2, G9, and G10 either fail or pass on an empty iteration,
which is the defect class DL guards against elsewhere. 5c's exit gate
nonetheless requires "G1 through G10 passing" (spec:1290) and the served
document "carrying the composed `data-theme` value".
*Fix:* decide and state who authors the one composed theme rung 5's pages
render against. Either rung 5 authors a minimal `dubplate` match block
plus one palette block as data (not a roster entry) and says so in 5c's
scope, or the spec defers G2, G9, and G10 to rung 7 with a named reason
and drops them from 5c's exit gate. Correct the "standard's default
block" sentence, and add a staleness question that reads DL for a shipped
default block rather than assuming RE's expectation.

**M11. Fork 1 makes the roster a file the loader fails closed on, and no
rung is assigned to place that file on a deployment.** (Integrity lens,
evidence extended.)
Fork 1 (spec:126-135) and R1 (spec:184-191) make the roster a YAML file
at `DUBPLATE_ROSTER_FILE`, Required (spec:906), and R2 makes a missing
file a load error. The deployment section forecloses the delivery path in
one sentence: "`deploy.sh` gains the unit, the pre-create, and nothing
else" (spec:1039). No other rung takes it. The rung 16 edge lists minting
the three secrets, the ingress flip, and deleting the Access application
(spec:1113-1116), and "What rung 5 does not build" (spec:1048-1067) never
mentions roster delivery. CLAUDE.md's safety rails forbid hand-editing
the box and calling it done.

Re-verified against the tree, a mechanism and an estate roster already
exist, which sharpens the question rather than answering it.
`scripts/deploy.sh:420-426` deploys `config/contributors.yaml` to
`/etc/dubplate/contributors.yaml` at mode 0640, `root:music`, for the
incumbent importer's `DUBPLATE_CONTRIBUTORS_FILE` lookup, and that key is
one of the two R1 rewrites. So the estate's answer could be to point
`DUBPLATE_ROSTER_FILE` at the file `deploy.sh` already places, and the
spec never says so. The universality half is untouched: `config/
contributors.yaml` is a site-specific roster in the repo.
*Fix:* name the rung that delivers a roster file to a deployment and the
mechanism, and add it to the runbook line fork 1 already promises. State
whether rung 5 reuses `deploy.sh`'s existing contributors path or
introduces one, whether the greenfield tree ships an example roster and
where, and where the estate's own roster lives, since a site-specific
roster in the repo would breach the universality rule.

**M12. The universality section claims a plain-SMTP, no-Cloudflare
deployment works without naming the TLS precondition the `__Host-` cookie
imposes.** (Integrity lens.)
Spec:1536-1542 claims "The design makes sense with no Cloudflare account.
Plain SMTP submission plus any streamer, with the tunnel optional and a
forward-auth proxy optional in front." The session cookie is `__Host-`
prefixed with `Secure` (spec:588-596, per BS:441-443), which browsers
refuse over plain HTTP on any origin but localhost. The serve keys the
spec declares are a bind address, a port, and a prefix (spec:920-926),
and nothing says serve terminates TLS or that a TLS front is required. An
adopter running dubplate over HTTP on a LAN gets a login flow that
completes, sets no cookie, and silently never signs anyone in. A grep of
the spec for HTTPS or TLS returns only SMTP transport lines and the
cookie's `Secure` attribute.
*Fix:* state the precondition in the universality section and in the
forward-auth and setup documentation: the deployment is reached over
HTTPS from any terminator, with localhost as the only exception. Add a
`doctor` row or a load-time check that `DUBPLATE_PUBLIC_URL` carries an
https scheme unless its host is localhost, so the failure surfaces at
configuration time rather than as a login that does nothing.

---

## MINORS

**m1. The ingress strip would delete the header a forward-auth front
asserts, and no carve-out rule is written.** (Security lens, downgraded
from major.)
The strip (spec:380-404) removes every inbound identity-header spelling
on every path, including "the trusted-header spellings the forward-auth
field uses, each in both its dash form and its underscore form". The
verify-identity seam (spec:678-681) admits forward-auth fronts as
alternates, and BS:499-503 makes a front an integration with an enable
flag. A front asserts identity in exactly one of the headers the extras
list strips. No front is buildable at 1.0, so there is no shipped defect,
and the cost is design rework at the rung that builds one. PA:80 gives
`internal/serve` to rung 12 with no concurrent serve-writing lane
afterwards, so the strip's shape is fixed here.
*Fix:* state the carve-out rule now. Exactly one header may be exempted,
it is the one named by the active front's configuration, the exemption
applies only when the peer address matches the front's configured
addresses, and every other spelling including the exempted header's
underscore twin is still stripped. Say that the exemption is impossible
while magic-link is the active source, on BS:499-503's rule that the
inactive source is never consulted. Record it as the mechanism a later
front configures rather than as code rung 5 builds.

**m2. The revocation epoch is written with the one `atomicfile` entry
point that does not fsync the directory.** (Security lens.)
R11 (spec:265-274) sends the epoch record to `atomicfile.WriteFile`
"because overwriting is exactly what a revocation does". Measured against
the tree, `WriteFile` creates a temp file, writes, syncs the file,
chmods, and renames (`internal/atomicfile/atomicfile.go:15-42`), and it
does not fsync the containing directory. `WriteStream` does, and its doc
comment says so (`atomicfile.go:45-56`). A rename that has not reached
the directory can be lost in a crash, and for this record the pre-crash
state is the un-revoked one, so the failure mode restores access an
operator believes they removed.
*Fix:* write the epoch through `WriteStream`, or state in R11 why the
weaker durability is accepted for a security record whose loss fails
open. The nonce record already syncs the directory under R11, so the two
records would then agree.

**m3. Path normalization drops two of the four rules SEC asked for, and
the prefix match's case rule is unstated.** (Security lens.)
The composition order at spec:469-472 states normalization as rejecting
"a raw and decoded mismatch and any `..` segment before prefix routing",
and 5a's exit gate asserts those two properties. SEC's path-smuggling
major asks for more: "reject `..` segments, empty segments, and
backslashes outright". Empty segments and backslashes are dropped. The
backslash case matters on the fallthrough, where rung 12 forwards a raw
path to a streamer whose own router may treat a backslash as a separator,
the same smuggling shape one level down that spec:473-478 cites
`mountBasePath` for. The segment-aware prefix match does not say whether
it is case-sensitive, and a case-insensitive match would widen the owned
surface while a case-sensitive one leaves `/Prefix` to the fallthrough.
*Fix:* extend the normalization rule to SEC's full list, adding empty
segments and backslashes, and state the prefix match as case-sensitive
byte comparison on segment boundaries. Add both to 5a's exit-gate
properties.

**m4. The bind-address default is unstated, and `IP_FREEBIND` makes a
wrong one succeed quietly.** (Security lens.)
`DUBPLATE_SERVE_BIND_ADDRS` and `DUBPLATE_SERVE_PORT` are listed at
spec:922-925 as thresholds with defaults, and no default value is named
anywhere in the spec. R20 (spec:337-342) adopts `IP_FREEBIND` so the
listener binds an address not yet present, which by design removes the
error a wrong address would otherwise raise at boot. SEC's
Access-retirement minor names the constraint from the other side: keep
serve bound to the bridge gateway and loopback only, never `0.0.0.0`,
since the login surface is now anonymous-reachable and the box is
production.
*Fix:* name the shipped default and its reason in the serve key table,
with loopback as the safe default and the tunnel or the adopter's proxy
as the documented way to reach it. State the residual that `IP_FREEBIND`
suppresses the boot-time error for a misconfigured address, and give
`doctor` a row reporting the addresses actually bound.

**m5. Config text becomes an attribute value and CSS declarations in the
served document with no stated validation.** (Security lens, scoped by
the correction above.)
R14 (spec:294-298) declares the theme keys, the branding values, and the
operator CSS file path as `KindString`, and the embed path (spec:786-789)
says the Go layer "substitutes a named placeholder token with the composed
`data-theme` value" in the embedded `index.html`. DL's G8 already gates
the branding key names. Nothing constrains the branding values, which
become CSS declarations in a generated stylesheet, and nothing constrains
the composed theme value, which becomes attribute markup. SEC filed the
neighboring version of this defect against the salvage page's
`fmt.Appendf` render.
*Fix:* state that the composed `data-theme` value is validated against
the known match and palette names before substitution and refused
otherwise, which also discharges DL R3.2. State that branding values are
validated against the value grammar their variables accept and escaped
when rendered into the generated stylesheet. Say that the operator CSS
file is read from the configured path only, served with a fixed content
type, and never resolved from any request-supplied component.

**m6. The B12 erratum reassigns proxy hygiene and leaves hop-by-hop
header hygiene owned by neither rung.** (Security lens.)
The erratum at spec:1264-1271 splits B12 into a shell-level half at rung
5 (the strip set, path normalization, the write-deadline policy) and a
rung 12 half ("the `httputil.ReverseProxy` half, its identity injection,
and its cache-control rule"). BS:512-519 lists one more item in the
clause: "session cookie and hop-by-hop hazards stripped". The session
cookie half is claimed by rung 5 at spec:407-410. Hop-by-hop appears in
neither half of the erratum and nowhere else in the spec. SEC's
proxy-hygiene major explains why it matters: a client sending
`Connection: Remote-User` causes the freshly injected identity header to
be deleted before it leaves, a downgrade any attacker can induce.
*Fix:* name hop-by-hop hygiene in the erratum and assign it. Ingress
removal of `Connection`-listed headers belongs with the strip in the
shell, since it must hold for every path, and the injection-time
consequence belongs with rung 12's `Rewrite`. Say which, so the clause is
not lost between two rungs.

**m7. The mail seam's injection guard is scoped to the recipient, while
the subject and every other header-valued field carry the same hazard.**
(Security lens.)
The mail section at spec:719-721 says "The seam owns the header-injection
guard", citing SEC's CRLF finding, and 5b's exit gate asserts only that
"a recipient carrying CRLF is refused by the seam". The seam's contract
at spec:714-717 carries a subject and a body as data from arbitrary
callers, and spec:723-727 says the implementation synthesizes From, Date,
and Message-ID and RFC 2047 encodes the subject. A subject is a header
value, and rung 7's reject mail and rung 9's notifications will pass
values derived from filenames and tag text.
*Fix:* state that the guard rejects control characters in every
header-valued argument, recipient and subject alike, and that the From
value is validated at config load rather than at send. Widen the
exit-gate assertion to cover a subject carrying CRLF.

**m8. Rung 5 ships the build's first secret-shaped fixtures and the spec
never mentions the gitleaks allowlist.** (Security lens.)
`scripts/check.sh` runs gitleaks over the tracked worktree, and rung 5
introduces HMAC secrets, token serialization, and the test fixtures that
exercise them. RE cell (5,2) names this rung specifically: ".gitleaksignore
is the sharpest, because rung 5 is the rung most likely in the whole
build to need an allowed finding", and its paste-ready row requires the
lane that needs a finding to append and to name it in its pass report.
Rung 1b retired the file's one fingerprint by a field rename, and the
tracked file survives at zero bytes. The spec's contested-surface table
carries no `.gitleaksignore` row and the staleness pre-task does not ask
about it.
*Fix:* add RE cell (5,2)'s `.gitleaksignore` row to the contested-surface
table, and state that any allowed finding is named in the pass report.
Prefer fixtures that are obviously not credentials, since an allowlist
entry on the auth package is the one a later reader is least able to
re-audit.

**m9. A fifth in-tree copy of the "web legs rejoin at rung 7" sentence is
neither in the errata nor in the staleness pre-task.** (Integrity lens.)
The spec counts three in-tree copies (`CLAUDE.md:118`,
`scripts/check.sh:5`, `Makefile:33-34`) and writes one erratum against
ladder entry 7, which covers BS:658-659. The build spec states it twice.
BS:332-333 also says rung 1 "removes the web legs until rung 7 rebuilds
`web/`", in the gate-rewiring bullet. After rung 5 closes that sentence
is wrong and no erratum touches it, which is the drift risk 9 names.
*Fix:* add BS:332-333 to the ladder entry 7 erratum, or write a separate
erratum against the gate-rewiring bullet, and add it to the staleness
question that checks the in-tree copies.

**m10. Three `keys.go` supersession-message rewrites are directed without
extending the rung 1b contract that governs edits to that file.**
(Integrity lens.)
R1 rewrites `DUBPLATE_CONTRIBUTORS_FILE`'s message (spec:184-188,
`internal/config/keys.go:177-178`) and spec:958-965 rewrites the two
Access rows' messages (`keys.go:171-176`). Rung 1b's contract, recorded
at `docs/HISTORY.md:494-497`, permits one edit class only: "no lane adds
a name to `internal/config/keys.go`; a lane removes only its own reserved
class from that file, in the same commit that declares the class in its
own package." R5 extends the rule by argument for the pending wall row.
The three supersession rewrites get no equivalent statement, and the
contested-surface row names only the two class removals and the wall row.
*Fix:* state the extension once, the way R5 does, and widen the
contested-surface row to cover the supersession messages rung 5's own
keys replace, so a later lane inherits the rule rather than the
precedent.

**m11. R6's probe rows collide with `doctor`'s gate-resident line-count
assertions, which the spec does not name.** (Integrity lens.)
R6 says the probe rows "do not fit the existing row shape" (spec:224-232)
and the contested table records rung 5 adding "the probe row kind". Rung
1b derived `doctor`'s output from the registry and pinned it with tests.
`cmd/dubplate-server/doctor_test.go:129-130` fails on any "extra line(s)
beyond owned+reserved+threshold rows", and
`TestDoctorCmd_LineCountFollowsRegistry` (`:151-175`) asserts the total
line count equals the registry's own row counts and re-asserts it after
adding one owned wall row. A probe row that is not registry-derived reds
both. The same applies to the per-source satisfied row spec:917-918
promises. The staleness pre-task asks whether `doctor` still derives
every row from the registry, and not whether a test pins the count.
*Fix:* add one sentence to R6 naming both assertions and stating whether
the probe rows register through the registry, keeping the derivation, or
extend the tests. Add the test pin to the staleness question.

**m12. The 0c discharge list omits R3.2, R3.4, and R3.5, and claims R4.2
whole while rung 5 builds no picker.** (Integrity lens, second half
re-scoped.)
Spec:753-757 enumerates the clauses rung 5 discharges as R1.6, R2.1
through R2.7, R3.3, R3.6, R3.7, R4.1 through R4.6, R5.1 through R5.5, and
R5.6. R3.2 (DL:249-260) requires config to accept `dubplate` as the match
value and "reject any other value with an error naming the accepted set",
which is validation on a key rung 5 declares and no other rung reads at
rung 5. R3.4 and R3.5 (composition ordering, fail-loud completeness) are
unmentioned in either direction. R4.2 is claimed in full, and its ruled
client half puts the per-member pick in `localStorage` until rung 6,
which rung 5's two pages do not build.
*Fix:* add R3.2's validation to the config section, which m5 also asks
for, and either claim R3.4 and R3.5 or defer them with rung 7 named.
State which half of R4.2 rung 5 discharges, so the list reads as complete
rather than as a partial enumeration.

**m13. Rung 5's own user-facing copy points at a contract the spec
assigns to rung 7.** (Integrity lens.)
Spec:826-827 says the consumption interstitial's "copy follows the copy
contract's register", and spec:1057-1059 puts "the reject copy contract"
wholly in rung 7's column. The login page's generic outcome,
roster-absent, and rate-limit strings (spec:812-819) are the build's first
user-facing copy and have no named authority, while the build spec's
reuse stance binds reject copy to a ratified string deck with
register-gated changes.
*Fix:* name the authority for rung 5's page copy. Either rung 5 authors
it under the writing-voice editor standard with a register gate at 5c, or
the copy is drawn from the same string authority. Drop the forward
reference to rung 7's contract.

**m14. The gate leg's cost to every other lane is unpriced, and a
worktree carries no `node_modules`.** (Integrity lens.)
R13 makes Node a hard requirement of `scripts/check.sh` (spec:285-292)
and 5c rejoins the leg. `scripts/check.sh` is the one gate every lane
runs at every task close, and PA:150 already measured that whole-repo
legs are "paid by every lane on every task, every re-dispatch, and every
post-rebase re-run". The lane-launch checklist provisions a venv for a
worktree lane precisely because `git worktree add` checks out tracked
files only (PA:226); `web/node_modules` is the same shape and no item
covers it. The contested-surface table lists `scripts/check.sh` with
"every lane" as the consumer, and the ceilings section does not mention
the new per-gate cost.
*Fix:* add a lane-launch line (install web dependencies in any tree that
runs the full gate after 5c) and one sentence in the band-and-ceiling
section acknowledging the added per-gate cost, so the next lane's
sub-ceiling is derived rather than surprised.

---

## Verified sound

### The auth mechanism and its stdlib footing

- R11's read of `atomicfile` is correct. `WriteFile` creates a temp file
  and renames over the target, so it silently overwrites an existing
  record (`internal/atomicfile/atomicfile.go:15-42`), which makes it
  wrong for a first-writer-wins nonce and right for an epoch overwrite.
  `O_CREATE|O_EXCL` plus a file sync is the correct insert.
- The header-canonicalization reasoning at spec:397-404 is correct.
  `textproto.CanonicalMIMEHeaderKey` uppercases the letter after each
  dash and treats the underscore as an ordinary token character, so
  `Remote_User` and `Remote-User` are distinct map keys and deleting one
  leaves the other. That is exactly SEC's underscore finding.
- R15's mechanism is available and behaves as described. Measured on the
  installed `go1.27.1`: `net/http.CrossOriginProtection` exists, its
  documented default is that "Requests without Sec-Fetch-Site or Origin
  headers are currently assumed to be either same-origin or non-browser
  requests, and are allowed", and `go.mod` declares `go 1.27.0`. Wrapping
  it to deny is a genuine tightening. Declining `gorilla/csrf` on
  GO-2025-3884 is sound when the stdlib implements the same check.
- The verification order at spec:562-565, MAC then expiry then nonce
  insert, matches BS:436-439 and prevents an unauthenticated caller from
  growing the nonce directory.
- The session cookie shape at spec:589-596 matches BS:441-448 clause for
  clause: `__Host-` prefix, `Secure`, `HttpOnly`, `SameSite=Lax`,
  `Path=/`, no `Domain`. The prefix also disposes of session fixation
  from a sibling host, so no separate fixation control is owed, and
  same-origin as a precondition of SSO follows correctly.
- The GET-renders, POST-consumes, origin-checked split at spec:575-580
  matches BS:436-439 and closes the prefetch-burn and login-CSRF halves
  of SEC's consumption blocker.
- R2's fail-closed roster is a real correction. The salvage loader
  returns an empty roster and a nil error for a missing file
  (`_salvage/internal/config/contributors.go:32-39`), and a fail-open
  roster presents as a broken login rather than a broken configuration.
- The secrets treatment at spec:566-570 matches BS:472-477: a length
  floor enforced at load, domain-separation and version labels inside the
  MAC input, length-prefixed serialization against claim confusion, and a
  `_PREVIOUS` overlap that verifies and never mints.
- R16's bounded, evicting limiter map with a flood test answers SEC's
  memory-exhaustion primitive, and keying the limiter on the normalized
  lowercased address before any lookup answers the multiplied-bucket
  half.
- R10's 0600 files under a 0700 leaf is the right correction to rung 4's
  0644 and 0755 inheritance for identity state, and stating the modes in
  the store's code rather than relying on the unit's `UMask` is the
  durable form.

### The serve shell and the trust boundary

- The one-owned-prefix treatment at spec:440-455 answers SEC's `/api/*`
  collision blocker and UNI's parallel finding correctly, and the rung 14
  prefix terminator built by rung 5 is the right resolution of RE cell
  (14,12), since rung 5 owns the mux.
- The parallel spec's serve-shell requirements appear in substance. The
  composition order matches PA:79 in sequence (ingress strip, path
  normalization, prefix routing, fallthrough), the extension point is
  declared as a named registration seam plus one root-level fallthrough
  slot, the driver's proxy is placed at the fallthrough with PA:77's
  reasoning reproduced, and the no-two-lanes rule appears in the
  contested-surface table and as a staleness question citing PA:80.
  `internal/serve` does not exist in the tree, so PA:80 is unviolated.
- The deadline reasoning is correct. The salvage server's single
  30-second write timeout (`_salvage/internal/serve/serve.go:601`,
  `:632`) kills an SSE stream and any long audio response, and
  `http.ResponseController.SetWriteDeadline` is the mechanism that lets a
  streaming handler extend its own while the server default protects
  everything else.
- R5's mechanical reading of the pending wall row is correct against the
  tree. `internal/config/keys.go:19-25` carries `{Name: "identity
  source", Pending: ...}`, and `Load` keys its requires on row names, so
  clearing `Pending` alone would make `Load` require a variable literally
  spelled "identity source". The `DUBPLATE_CONTRIBUTORS_FILE`
  supersession message R1 rewrites is live at `keys.go:177-178`.

### The gate, the config surface, and the tree

- R12 and R13's citations hold. `scripts/check.sh:153-168` names
  `systemd/dubplate-shadow.service` literally in the state-dir agreement
  check, and `scripts/check.sh:19-38` already fails by name for go,
  golangci-lint, shellcheck, ruff, and gitleaks, so adding Node to that
  list follows the repo's own gate rule.
- The three live copies of the rung 7 rejoin sentence are present as
  cited: `CLAUDE.md:118`, `scripts/check.sh:5`, and `Makefile:33`. DL
  section 6 names only the first two, so the `Makefile` catch is real.
- The registry contract is followed for the two reserved classes. Each is
  removed from `internal/config/keys.go` in the same commit that declares
  its replacement in rung 5's own file, and the removals are split by
  pass so 5a takes `serve` (`keys.go:79-88`, two keys) and 5b takes `auth
  and identity` (`keys.go:60-72`, eight keys), matching
  `docs/HISTORY.md:494-497`.
- The tree citations rung 5 leans on check out at `be27434`:
  `cmd/dubplate-server/main.go:63-80` is the single `AddCommand` block,
  `cmd/dubplate-server/doctor.go:23-71` derives every row from the
  registry, and `internal/atomicfile` carries both `WriteFile`
  (rename-over) and `WriteStream`, which is what R11's reasoning turns
  on.

### The mail seam and the prior art

- R7's decline of a Cloudflare Email Routing relay is correct on the
  facts. The pings Worker's send_email binding delivers only to verified
  destination addresses, and a magic link goes to whatever address a
  person typed, so that transport cannot serve this flow.
- The prior-art table is complete on CLAUDE.md's four criteria and
  faithful to PRA, which is its cited study. Adoption is named as
  go-mail's weak cell rather than hidden, and each decline carries a
  reason on license, quality, maintenance, or shape. The version and
  importer figures were not re-measured by this review.
- The contested-surface ownership follows PA section 3 and RE.
  `internal/notify` is rung 5's as a general send seam (RE cell (5,9)'s
  paste-ready sentence is reproduced almost verbatim in the mail
  section), the roster loader is rung 5's, and the errata for ladder
  entry 9 and the dependency analysis are recorded rather than left as a
  live contradiction.

### Scope, rulings, and process

- The four given owner rulings are recorded exactly and in their own
  section, and none is reopened or softened elsewhere.
- Every further ruling R1 through R26 states a position rather than a
  question, and the section header applies the provisional mark to all of
  them.
- Product forks are separated correctly and conservatively. All six are
  adopter- or family-visible, and the spec adds fork 6 beyond the
  inventory's five rather than dropping any. No engineering-only ruling
  is escalated: R2, R10, R11, R13, R16, and R20 stay conductor rulings.
- RE cell (5,6)'s two obligations are discharged as the cell asks.
  `internal/identity` is named as the home of the verified identity with
  `internal/ops` forbidden from importing `internal/auth` (R17), and R18
  answers the migration question "no" in writing, which the cell said was
  a legitimate outcome worth recording.
- The pass split, ceilings, and worktree release follow INV's budget
  evidence: four passes at 3M, 4M, 4M, and 3M for 14M total against rung
  3's four-pass actual of about 17.7M, ceilings re-derived per pass, the
  worktree released at each pass close per RE cell (5,3), and every task
  boundary planned as a resumable stopping point on pass 8a's session-cap
  evidence.
- Each pass's exit gate is stated as behavior a test can assert rather
  than as a structure to inspect, and the gates match the risks section
  item for item.
- The 2.0 hook statement is complete against S20's list. Hook 5 is landed
  with three mechanisms named (a single resolution point, a
  required-shaped `streamer_username`, the epoch re-key as rung 6's
  obligation, matching S20:367-369), hook 6 is neutral and kept open with
  the fallthrough empty and the strip set as data, and the remaining
  hooks are genuinely schema, tagging, layout, and migration concerns.
- The CSP directive set at spec:768-777 transcribes DL R5.6 faithfully,
  including the hash-pinned `script-src` with no `'unsafe-inline'`, the
  `style-src` relaxation with its stated reason, and the deliberate
  absence of `font-src`.
- Universality holds in the places it is testable. No hostname, path,
  email, or personal name is defaulted outside config, `DUBPLATE_ROSTER_
  FILE` and the SMTP host ship no default, driver knowledge stays with
  rung 12, every tunable is a threshold with a default, and the mail seam
  offers two worked adopter examples rather than one estate answer. M11
  and M12 are the two gaps in the claim, and neither is a site-specific
  value in the tree.
- The writing-voice standard holds for developer docs. The document
  carries no em dash, no "not X but Y" frames, no connector openers, and
  no scaffold headers, and `tellgrader` scores 0.39 tells per 1000 words
  over 12,885 words with a cadence CV of 1.08. Its five slop-hard hits
  are all the literal word "underscore" describing header spellings.
