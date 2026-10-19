# Rung 5b-i security review (fresh context, 2026-09-12)

The read R-5b16 requires at the end of the 5b-i half (R24 as corrected).
Graded against the rung 5 spec's threat model
(`docs/specs/2026-09-08-rung-5-identity-auth.md`, "## The
threat model") and its auth-core sections ("## The auth core" through
"### The verify-identity seam"). Scope: `internal/identity/*.go`,
`internal/auth/*.go`, `cmd/dubplate-server/auth.go`, and the 5b-i
changes to `internal/config/{config,registry,keys}.go` and
`cmd/dubplate-server/{doctor,main,serve}.go`, at `1d4ed29`.

Every CONFIRMED finding below was reproduced by running the landed code
in a scratch copy of the tree, not by reading alone. The reproductions
are described inline so the fold can re-run them.

Severity counts and the three 5b-ii constraints are at the end.

---

## Blocker

### B1. A consumed nonce's record can be swept while the link that bore it is still live

CONFIRMED. `internal/auth/nonce.go:117` writes the record with
`nonceRecord(now, s.retention)`, so the record's life is the store's
*currently configured* `DUBPLATE_NONCE_RETENTION` rather than the
*token's own* `ExpiresAt`. The only guard is a load-time relation,
`cmd/dubplate-server/auth.go:151-153`, which compares retention against
the currently configured `DUBPLATE_MAGIC_LINK_EXPIRY`; it says nothing
about a token minted under an earlier, longer expiry.

Attack. An adopter runs `DUBPLATE_MAGIC_LINK_EXPIRY=48h` with
`DUBPLATE_NONCE_RETENTION=48h` (accepted), mails a link, then tightens
to `15m`/`24h` and restarts; the outstanding 48h link is consumed after
the restart, its record is written with a 24h life, and 24 hours later
the sweep removes it while the token still verifies for another day. The
same window opens on a forward clock step past a record's expiry. A
second presentation of the same link URL, which is exactly what a mail
provider's link scanner or a browser-history reader holds, then mints a
second session. This directly contradicts the threat model's claim that
"a nonce cannot be consumed twice across a restart or across any
accepted retention configuration".

Reproduction: open a nonce store with a 1s retention, `Insert` a nonce
(nil), `Insert` again (`ErrNonceReplayed`), `Sweep(now+2s)` (Removed: 1),
`Insert` a third time returns nil.

Fix. Make the record outlive the token that bore it by construction, not
by configuration: take the token's `ExpiresAt` as an `Insert` parameter
and write `max(now.Add(retention), tok.ExpiresAt)` into the record, so
the retention threshold can only ever lengthen a record's life. Do this
before 5b-ii writes the consumption handler against the current
two-argument signature. Keep the load-time relation as a second net.

---

## Major

### M1. `identity.NormalizeAddress` is not idempotent, and the one control the divergence breaks is revocation

CONFIRMED. `internal/identity/identity.go:34-43` returns
`strings.ToLower(addr.Address)`, and `net/mail.ParseAddress` unquotes a
quoted local part. `"a b"@example.com` normalizes to `a b@example.com`,
which `ParseAddress` then refuses. The package doc at
`internal/identity/identity.go:31-33` states the opposite invariant:
"the address inside a token's MAC and the address in a lookup never
diverge."

Attack and failure mode. `MintLink`
(`internal/auth/token.go:118-132`) accepts the address and mails a
working link. Every later step that re-normalizes the token's own
address then fails with a wrapped `ErrInvalidAddress` rather than a
verdict: `EpochStore.CheckLink` (`internal/auth/epoch.go:194-197`),
`EpochStore.CheckSession` (`internal/auth/epoch.go:166-169`),
`MintSession` (`internal/auth/session.go:103-106`), and, decisively,
`EpochStore.Bump` (`internal/auth/epoch.go:81-84`), which cannot write
an epoch for such an address at all. So the system will mint and mail a
credential for an address it can never revoke, and every downstream
refusal arrives as a non-sentinel error. A 5b-ii handler that branches
on `errors.Is(err, ErrLinkRevoked)` and treats anything else as
"not revoked" fails open.

Reproduction: `MintLink(secret, "\"a b\"@example.com", ...)` returns nil
error; `CheckLink` on the verified token returns
`check link epoch: identity: invalid address: mail: no angle-addr`;
`Bump` on the same address returns the same class of error.

Fix. Make normalization a fixed point: after lowercasing, re-parse the
result and refuse the address when the re-parse fails or yields a
different string. Refuse any normalized address carrying a character
outside the unquoted-atom set as well (see m10). Add a table test
asserting `Normalize(Normalize(x)) == Normalize(x)` for the whole corpus.

### M2. The revocation epoch is not monotonic, so a bump can silently un-revoke

CONFIRMED. `EpochStore.Bump` (`internal/auth/epoch.go:80-98`) writes
`now` over whatever the record held, through
`atomicfile.WriteStream`, without reading the existing value.

Attack. Any `Bump` carrying a timestamp earlier than the stored epoch
reinstates every session the earlier bump refused. The reachable
triggers are a backward clock step (an NTP correction, a VM restore or
migration on the box, a host whose clock was wrong at boot) between an
offboarding and any later bump, and a second `Bump` call whose `now` a
caller derived from anything other than the current wall clock. The
threat model claims "an epoch bump refuses every session older than it";
after a backward bump it does not. Revocation is rung 5's only
revocation mechanism, which is what raises this above a curiosity.

Reproduction: mint a session at T, `Bump(addr, T+1m)`, `CheckSession`
returns `auth: session revoked`; then `Bump(addr, T-1h)` and
`CheckSession` returns nil.

Fix. Read the existing epoch inside `Bump` and write
`max(existing, now)`, refusing (or no-op'ing with a logged warning) a
backward move. The read-modify-write needs the store's own mutex held
across it, since `atomicfile.WriteStream`'s rename is atomic but the
read and the write are not; two concurrent bumps otherwise still admit
the older value.

### M3. Token and cookie strings are malleable: one credential has many verifying encodings

CONFIRMED. `splitSigned` (`internal/auth/token.go:244-258`) decodes each
part with `base64.RawURLEncoding.DecodeString`, which does not reject
non-zero trailing bits in the final quantum. Any field whose byte length
is not a multiple of three therefore has several distinct encodings that
decode to identical bytes, and the MAC is computed over the decoded
fields, so every one of them verifies.

Attack. Given one valid link, an attacker derives hundreds of distinct
token strings that all pass `VerifyLink`. The 5b-i single-use property
survives only because `NonceStore.Insert` keys on the decoded nonce
bytes (`internal/auth/nonce.go:105`). Anything 5b-ii keys on the raw
string instead, a per-token rate-limit bucket, a seen-token cache, a
log-line dedup, an audit correlation id, is defeated by mutating one
character. The same applies to the session cookie.

Reproduction: mint a link, mutate only its final character across the
base64url alphabet; three of the substitutions produce different strings
that `VerifyLink` accepts with the identical `LinkToken`.

Fix. Require canonical encoding in `splitSigned`: after decoding each
part, refuse the value unless
`base64.RawURLEncoding.EncodeToString(raw) == part`. Three lines, and it
makes "one credential, one string" true by construction rather than by
every future caller's discipline.

### M4. `doctor` prints an all-green table, including "magic-link active", for a configuration `Load` refuses

CONFIRMED. `identitySourceStatus`
(`cmd/dubplate-server/doctor.go:167-177`) decides "active" from
`statuses`, and `rowStatuses`
(`cmd/dubplate-server/doctor.go:90-112`) unwraps `Load`'s joined error
exactly one level, then splits each member on its first `": "`.
`validateMagicLinkSource` (`cmd/dubplate-server/auth.go:124-156`)
returns its own `errors.Join`, so the whole hook contributes one member
whose `Error()` is newline-joined; only its first key is ever recorded,
and if that key is a threshold or an Optional row, no row consults it.

Attack, or rather the operational failure. With the magic-link source
selected and only `DUBPLATE_NONCE_RETENTION` below the link expiry, the
built binary prints every owned row `ok`, every threshold's value, and
`magic-link active`, then the bare line `configuration is incomplete`.
No row anywhere names the offending key. The same happens when the only
fault is a short `DUBPLATE_SESSION_SECRET_PREVIOUS`. Revision 2 of the
spec recorded this exact defect class as a blocker ("an identity-source
selector that the landed registry requires nowhere and `doctor` reports
green while it is unset"); it has come back in a new place. Exit status
is still 1 and `serve`'s own stderr line names the key correctly, which
is what keeps this out of blocker.

Reproduction: build `cmd/dubplate-server`, run `doctor` under a
satisfied magic-link environment with `DUBPLATE_MAGIC_LINK_EXPIRY=48h`
and `DUBPLATE_NONCE_RETENTION=1h`.

Fix. Flatten nested joins when building `statuses` (recurse through
`Unwrap() []error`, and split a multi-line member per line), have
`identitySourceStatus` consult the source's Optional keys and its own
hook's failures as well as Required, and print one row per unsatisfied
identity-source key so the operator reads a name rather than a verdict.
Add a gate test asserting that no `doctor` run can report a non-failing
table while `Load` returns an error.

### M5. `VerifyLink` allocates sixteen times the input before any structural check

CONFIRMED. `splitSigned` calls `strings.Split(value, fieldSeparator)`
(`internal/auth/token.go:245`) with no bound, then checks the field
count. A value that is nothing but separators becomes one substring
header per byte.

Attack. Measured: an 8 MiB value of `.` characters allocates 134 MB
before `VerifyLink` returns `ErrLinkMalformed`, a 16x amplification. A
handler with a 1 MiB body cap still hands an attacker 16 MB per request,
and the auth endpoints are public by construction and unauthenticated at
the point of the call.

Fix. `strings.SplitN(value, fieldSeparator, count+2)`, which caps the
slice at `count+2` elements and leaves the existing
`len(parts) != count+1` check exact. Keep the body cap as well; this is
the half that does not depend on every caller remembering.

---

## Minor

- **m1. No `Max-Age` on the session cookie.**
  `internal/auth/session.go:146-156` sets `Expires` only; the emitted
  header carries no `Max-Age`, confirmed by printing `Cookie.String()`.
  `Expires` is evaluated against the client's clock, so a browser whose
  clock is slow keeps the cookie past the server's intent (harmless, the
  server enforces its own expiry) and one whose clock is fast drops it
  early. Set `MaxAge` alongside `Expires`; OWASP's Session Management
  cheat sheet treats an explicit age as the primary control.

- **m2. Refresh writes a cookie expiry beyond the absolute cap.**
  `RefreshSession` (`internal/auth/session.go:130-134`) sets
  `ExpiresAt = now + lifetime` with no clamp to `AnchoredAt + cap`.
  Confirmed: a session anchored at T with a 30h cap, refreshed at T+23h
  with a 24h lifetime, carries a cookie expiry at T+47h. The server
  refuses it correctly at the cap (`ErrSessionCapReached`), so this is a
  cosmetic divergence between what the browser is told and what the
  server will honor. Clamp to `min(now+lifetime, anchor+cap)` so the two
  agree and the browser stops sending a value that can only be refused.

- **m3. `auth.Open` does not tighten an existing leaf's permissions.**
  `internal/auth/store.go:52-60` uses `os.MkdirAll(..., 0o700)`, which
  is a no-op on an existing directory whatever its mode. A leaf created
  once at 0755, by a hand-run command or an earlier deploy, stays 0755
  forever and its record *names* (SHA-256 of a nonce, SHA-256 of an
  address) become listable by any local user. The record contents stay
  0600. Stat the leaf after `MkdirAll` and `os.Chmod` it, or refuse to
  start on a leaf whose mode is wider than `dirPerm`.

- **m4. `validatePublicURL` accepts userinfo and a fragment.**
  `cmd/dubplate-server/auth.go:171-184` checks scheme and host only.
  Confirmed: `https://evil.example@dubplate.example` validates and
  `LinkURL` produces
  `https://evil.example@dubplate.example/login/link?t=...`, a link whose
  visible prefix is an attacker-chosen name. A fragment survives too
  (`https://dubplate.example?x=1#frag` yields `...?t=TOK#frag`). Refuse
  a non-nil `u.User`, a non-empty `RawQuery`, and a non-empty `Fragment`.

- **m5. The four duration thresholds have no ceilings or floors.**
  `cmd/dubplate-server/auth.go:75-100` registers them with defaults and
  no bounds, and `config.ParseDuration` imposes none.
  `DUBPLATE_MAGIC_LINK_EXPIRY=8760h` with a matching retention loads
  clean, giving a year-long magic link. The threat model's mail-scanner
  residual is explicitly bounded by "single use, a 15-minute default
  expiry"; nothing enforces the second half. Give the link expiry a
  ceiling (an hour is generous), the session cap a ceiling, and all four
  a positive floor.

- **m6. `Store.Write` is unused and bypasses the first-writer-wins rule.**
  `internal/auth/store.go:94-99` has no caller. It is an
  overwrite-shaped write path sitting next to a store whose whole
  correctness argument is that the nonce record is never overwritten
  (`internal/auth/nonce.go:41-47`). Delete it, or land it with the
  caller that needs it.

- **m7. `recordPath`'s refusal prints the id.**
  `internal/auth/store.go:81-86` formats `%q` of the id. Every current
  caller passes a `RecordID` digest, so nothing secret reaches it today.
  A future caller that skips `RecordID` would print a raw nonce into an
  error string the package doc promises never carries one
  (`internal/auth/doc.go:14-15`). Print the length and the offending
  character class instead of the value.

- **m8. One unreadable record makes the process unstartable.**
  `OpenNonceStore` (`internal/auth/nonce.go:80-91`) returns the initial
  sweep's error, and `Sweep` joins a per-record read failure into it
  (`internal/auth/nonce.go:167-177`). A single record with a bad mode or
  a bad inode takes the login surface down. Fail closed on the store's
  own directory being unreadable, but log and continue past an
  individual record; the record's presence, not its readability, is what
  blocks a replay.

- **m9. There is no cookie-clearing helper, and logout cannot revoke.**
  `internal/auth/session.go` offers `SessionCookie` only. A `__Host-`
  cookie must be cleared with the same `Path`, `Secure`, and absent
  `Domain`, so 5b-ii needs a helper rather than a hand-built
  `http.Cookie`. Note also that these sessions are stateless: clearing
  the cookie does not invalidate the signed value, so a session captured
  before logout keeps working until its expiry or an epoch bump. The
  spec records this in its residuals; 5b-ii's logout copy must not claim
  more than it does.

- **m10. `NormalizeAddress` admits control characters other than CR and LF.**
  `internal/identity/identity.go:34-43` applies no character filter.
  Confirmed: `"a\tb"@example.com` normalizes to `a\tb@example.com`, a
  TAB inside the address. CR and LF are refused by `ParseAddress`, so
  SMTP header injection is not reachable, but the spec's roster rule is
  "no control characters" (spec, "### The roster") and this is the one
  normalizer every caller shares. Reject any normalized address carrying
  a character below 0x20, 0x7F, or a space.

- **m11. No `dubplate-server` unit exists, so the auth leaf has no `ReadWritePaths` grant.**
  `internal/auth/store.go:38-40` puts the leaf at `<state dir>/auth`,
  R-5b11 calls for one static grant, and the only unit naming
  `/var/lib/dubplate` is `systemd/dubplate-shadow.service:40`, which
  grants `/var/lib/dubplate/shadow` alone. `scripts/check.sh:171-185`
  checks that one leaf and nothing else. Under `ProtectSystem=strict`
  the runtime `MkdirAll` will fail at startup. Land the grant, the
  deploy-side pre-create, and a check.sh leg for the auth leaf with the
  serve unit.

- **m12. `RecordID` carries no per-kind domain separator.**
  `internal/auth/store.go:106-109` hashes the bare value for both the
  nonce store and the epoch store. Today the two subdirectories keep
  them apart. A future caller placing two record kinds in one directory
  gets a silent collision. Hash a kind label alongside the value, the
  same discipline `macInput` already applies to the MAC.

- **m13. The env template selects magic-link without shipping its nine required keys.**
  `env/dubplate-server.env.template:74` sets
  `DUBPLATE_IDENTITY_SOURCE=magic-link`, and the nine keys
  `magicLinkSource.Required` names are absent from the file (confirmed
  by grep). `deploy.sh`'s render succeeds, and the failure surfaces only
  when the binary starts. It is fail-closed and the template comment
  schedules the rows for 5d, so this is a note rather than a defect:
  the rows and the secret mints must land before anything selects the
  source on a box.

- **m14. `Bump` notifies subscribers synchronously and inside the call.**
  `internal/auth/epoch.go:142-151` calls each subscriber on the caller's
  goroutine after releasing the lock. A subscriber that panics
  propagates into a revocation, and one that blocks stalls it. The write
  is already durable at that point, which is the right order, but 5b-ii
  and the 2.0 credential-store subscriber should get a recover and a
  deadline around each callback.

---

## Attacked and found sound; the fold need not re-check these

- **Domain separation between the two credentials.** `linkPurpose` and
  `sessionPurpose` (`token.go:36`, `session.go:35`) are distinct, are
  length-prefixed as the first element of the MAC input
  (`token.go:292-298`), and the version labels differ and are checked as
  field 0 in both parsers. A session value cannot verify as a link or
  the reverse, even under one shared secret. The field layout is
  deliberately identical, and the labels are what separate them; that
  works.

- **Field encoding is unambiguous.** Every field is base64url-encoded
  before the `.` join, so no field value can carry the separator, and
  the MAC input length-prefixes each field with a big-endian uint64
  (`token.go:300-303`), so no byte can be shifted across a field
  boundary. The exact-count check in `splitSigned` closes the
  extra-field variant. The only defect here is canonicality (M3), not
  ambiguity.

- **MAC comparison.** `macMatches` (`token.go:270-275`) uses
  `crypto/subtle.ConstantTimeCompare` and returns false for an empty
  secret, so an unconfigured rotation overlap is absent rather than
  permissive. Go's `ConstantTimeCompare` returns 0 on a length mismatch
  without panicking, so there is no length branch to normalize.

- **Verification order.** `VerifyLink` (`token.go:143-155`) is MAC,
  then expiry, then the nonce is returned for the caller to insert.
  `VerifySession` (`session.go:167-182`) is MAC, then the absolute cap,
  then the expiry. Neither acts on an unauthenticated field, and an
  unauthenticated caller cannot reach `Insert`.

- **Expiry boundary arithmetic.** Both use `!now.Before(t)`, so a token
  or session is dead at exactly its expiry. The nonce sweep uses
  `now.After(expiresAt)`, so a record at exactly its expiry is retained.
  The two boundaries lean in the safe direction relative to each other:
  under a single, unchanged configuration the record always outlives
  the token, verified at the zero-margin case (consumed at mint time,
  retention equal to expiry).

- **The nonce insert's atomicity.** `O_WRONLY|O_CREATE|O_EXCL`
  (`nonce.go:110`) is first-writer-wins at the filesystem, with no
  read-then-write window, so two concurrent consumptions of one link
  cannot both succeed. The record is fsynced and the containing
  directory is fsynced after it (`nonce.go:117-126`), so a crash after a
  consumption does not lose the record. A crash *between* the create and
  the write leaves a zero-length file the sweep retains rather than
  deletes (`nonce.go:143-191`), which is the conservative choice.

- **The anchor and the refresh arithmetic.** `RefreshSession`
  (`session.go:130-134`) moves `IssuedAt` and `ExpiresAt` and never
  `AnchoredAt`, and `VerifySession` measures the cap from the anchor, so
  no number of refreshes escapes the cap. Verified by refreshing past
  the cap and getting `ErrSessionCapReached`.

- **The epoch refusal reads the anchor.** `CheckSession`
  (`epoch.go:165-181`) compares `AnchoredAt`, not `IssuedAt`, so a
  session refreshed after a bump stays refused; `CheckLink`
  (`epoch.go:193-209`) compares the link's mint time, not its redemption
  time, so an outstanding link cannot outlive an offboarding. Both use
  `!t.After(epoch)`, refusing at equality. The only defect around the
  epoch is monotonicity (M2) and the normalization error path (M1).

- **Cookie attributes.** `SessionCookie` (`session.go:146-156`) emits
  `Path=/`, `Secure`, `HttpOnly`, `SameSite=Lax`, and no `Domain`,
  which is exactly what the `__Host-` prefix requires; verified by
  printing the rendered header. `Lax` rather than `Strict` is the ruled
  choice, so a link followed from a mail client arrives with the session
  attached. The only gap is `Max-Age` (m1).

- **The session cookie is stripped from the fallthrough.**
  `newServeShell` (`cmd/dubplate-server/serve.go:158-161`) passes
  `auth.SessionCookieName` into `serve.NewStrip`, closing rung 5a's
  coverage gap 4; nothing leaving the owned surface carries the
  credential that authenticates it.

- **Secrets in errors and logs.** `validateMagicLinkSource`
  (`auth.go:124-156`) names keys and never values or lengths; verified
  by running `serve` with a short secret, which prints
  `DUBPLATE_MAGIC_LINK_SECRET: too short, want at least 32 bytes` and
  nothing else. `doctor` prints no identity-source values by
  construction (`doctor.go:167-177`), which is why the rotation-overlap
  secrets are Optional source keys rather than thresholds. The
  secret-length floor is 32 raw bytes, RFC 2104 section 3's own
  recommendation for SHA-256. The env file is installed 0600 root:root
  (`scripts/deploy.sh:402`).

- **The config requirement mechanism.** With the magic-link source
  selected, `Load` requires all nine of its keys and runs the hook even
  when a key is unset, so a missing row and a bad value report together
  (`internal/config/config.go:296-331`). `wall.get` treats an empty
  value as unset, so blanking `DUBPLATE_SESSION_SECRET_PREVIOUS` after a
  rotation is a clean unset rather than a zero-length secret. An unknown
  or reserved key is a load error, and a misconfigured source refuses
  the start rather than degrading it. The mechanism is sound; only
  `doctor`'s *rendering* of its verdict is wrong (M4).

- **Nonce store growth.** A record is written only after a MAC
  verification, so an unauthenticated caller cannot fill the directory.
  The record is 32 bytes and the sweep prunes on the record's own
  carried retention, so lowering the threshold cannot shorten a record
  already on disk.

- **`LinkURL` never reads a request header.**
  `internal/auth/token.go:161-169` takes the configured public origin as
  an argument, and the package imports no `net/http` request type at
  all. Host-header injection is not reachable from here. The token rides
  in the query parameter `t` rather than a path segment, which is R32's
  own mitigation shape.

- **No caller-supplied redirect exists.** `MintLink` takes no
  destination and the token carries no destination field
  (`token.go:178-186`), so R33 holds by construction at this half.

---

## Counts

| Severity | Count |
| --- | --- |
| Blocker | 1 |
| Major | 5 |
| Minor | 14 |

## The three findings 5b-ii's plan must carry as constraints

1. **Fix B1 before the consumption handler is written.** `Insert` must
   take the token's own `ExpiresAt` and write
   `max(now+retention, tok.ExpiresAt)` into the record. A handler built
   against the current two-argument signature will have to be rewritten.
   The handler's own order is fixed and must be stated in the plan: MAC,
   expiry, roster resolution, `CheckLink`, `Insert`, and only then
   `MintSession`, with the session minted only after `Insert` returns
   nil.

2. **Every non-nil error from the auth core denies, not only the named
   sentinels.** M1 makes a non-sentinel error reachable for an address
   the mint path accepted, and `EpochStore.Read` can also fail on an
   unreadable or corrupt record. The plan must require that
   `VerifyLink`, `VerifySession`, `CheckLink`, `CheckSession`, and
   `Insert` be consumed as "nil means permitted, anything else means
   refused", with the sentinel distinction used only to choose the
   server-side log line. One generic outcome reaches the client for all
   of them.

3. **Cap the body before the parser sees it, and key nothing on the raw
   token string.** M5 gives 16x allocation amplification ahead of any
   structural check, and M3 means one credential has many verifying
   encodings. Every auth endpoint caps its body with
   `http.MaxBytesReader` before calling into `internal/auth`, and every
   limiter, cache, audit correlation id, or dedup key derives from the
   decoded nonce or the normalized address, never from the token or
   cookie string. Landing M3's canonical-encoding check in `splitSigned`
   is the durable half of this constraint.

---

Verdict: one blocker (a replay window in the consumed-nonce record's
lifetime), five majors, and fourteen minors; the MAC, the domain
separation, the anchor arithmetic, and the `__Host-` attributes are
sound, and 5b-ii should not be dispatched until B1 and the three
constraints are folded into its plan.
