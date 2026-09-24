# Rung 5c plan review: mechanics and feasibility

Target: `docs/superpowers/plans/2026-09-23-rung-5c-pages-member-verb.md` at `dba1679`
(5c-i in full, 5c-ii outlined). Lens: does every mechanism the plan relies on behave
as stated? Toolchain `go1.27.1 linux/amd64`, `golang.org/x/time v0.16.0`. Experiments
ran in a scratch module under `/tmp/claude-1000/rv5c/`. Line numbers below are plan lines.

## Verdict

The plan's mechanisms mostly hold, and I measured most of them. `flock` on a directory
descriptor works, excludes a second descriptor in the same process, and keeps holding
across a child's atomic rename. `encoding/json/v2` refuses case variants and duplicates
and passes `null` through exactly as R-5c4 says. The netip keying table is correct row
by row. `TokensAt(now) >= float64(Burst())` holds with exact equality once a bucket
refills. The shell's origin check does refuse a header-less POST, so Task 9's first row
is sound. Every symbol and line the plan names exists at HEAD. There are no blockers.
Three majors each send an implementer into a contradiction or a silent regression:

- Task 2's refusal set includes `unicode.Zs`, which contains the ASCII space, so it refuses every display name that has a space.
- Task 1 routes `CreateRoster` through the function where Task 6 installs the no-admin guard, so the bootstrap write refuses itself. The plan's own exit gate 11 already forbids that empty-roster write.
- The stat-based reload does not say which `FileInfo` it caches. The wrong order can pin stale content, which quietly brings back review finding 1.

The minors are specification gaps and cost notes: lock deadlines, the actor's source,
the POST encoding, and the ceiling rate. Two are OWNER FORKs.

## Findings

### 1. MAJOR: the refusal set refuses the ASCII space in display names

- **Location:** Task 2, lines 963-970 ("refuses any rune for which `unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zs, unicode.Zl, unicode.Zp)`" and "Display names refuse the same set"). The command-line actor rule at line 979 also uses "the set above".
- **Defect:** `unicode.Zs` contains U+0020. For addresses this changes nothing, because `NormalizeAddress` already refuses `r <= ' '` (`internal/identity/identity.go:50`). For display names it refuses every name that has a space, such as "Alice Example". That breaks the example roster, the existing fixtures, and the product.
- **Evidence:** Measured: `U+0020 control=false Cf=false Zs=true`. Also `U+00A0 Zs=true` and `U+3000 Zs=true`.
- **Proposed fold:** For display names, refuse `IsControl`, `Cf`, `Zl`, `Zp`, and `Zs` except U+0020. Add a Task 2 acceptance row: "`Alice Example` is accepted; U+00A0 inside a display name is refused." Add a matching mutation: "refuse all of Zs, and the plain-space row fails."

### 2. MAJOR: `CreateRoster` writes through the guarded function, and the guard refuses the empty roster it writes

- **Location:** Task 1, lines 807-809 ("One unexported locked read-modify-write function that `AddMember` and `CreateRoster` route through, which Task 6's guard lives in"). Task 6, lines 1358-1361 and table row 1 at line 1370. Exit gate 11 at lines 1781-1782. Task 7, lines 1470-1482.
- **Defect:** Task 6 checks "every roster the store writes names at least one admin" inside the shared function, and `CreateRoster` writes `{"schema_version":1,"members":[]}` (`internal/auth/roster.go` `CreateRoster`). As written, the bootstrap write refuses itself with `ErrNoAdmin`. Exempting it instead contradicts exit gate 11 ("No roster write can leave the roster with no admin"). The two-step bootstrap in Task 7 (create, then add) also leaves an empty roster behind when the add fails after the create, for example on a malformed address that passes the verb's flag checks. That is why Task 7 needs both a pre-check and the "exists, empty" row. Separately, `CreateRoster` now waits on the lock but gains no `ctx` in Task 1's Interfaces (lines 801-809).
- **Evidence:** The plan text quoted above, and `roster.go` `CreateRoster` at HEAD, which writes an empty member list through `atomicfile.WriteFile`.
- **Proposed fold (recommended):** Change the signature to `CreateRoster(ctx, path, first Member, actor string, logger *slog.Logger)`. It writes the first member under the directory lock in one locked write, requires `first.Admin`, audits on `ChannelCommandLine`, and returns `ErrRosterExists` when the file is present. The guard then holds for every write with no exemption, exit gate 11 becomes literally true, and Task 7's absent-roster pre-check shrinks to a call. The "exists, empty" row stays for the hand-edited case. Alternative with less churn: keep the empty create, move it outside the guarded function under the same lock, and reword exit gate 11 to "no mutation of an existing roster". I recommend the first.

### 3. MAJOR: the reload does not say which stat it caches, and a wrong order silently brings back finding 1

- **Location:** R-5c5 (lines 453-459). Task 1 outcomes at lines 839-842 ("After a successful write the stored stat is the written file's").
- **Defect:** A reader that reads the file and then calls `os.Stat(path)` can cache the new inode's `FileInfo` against the old content. The race: a writer in another process renames between the read and the stat. After that, every later `Lookup` sees an unchanged stat and serves the stale roster indefinitely, including a hand-removed member. That is finding 1 again, and none of Task 1's named mutations detects it. A second, smaller gap: an in-place, same-size hand edit within one timestamp tick keeps the inode and size and may keep the mtime. This matters on a kernel with coarse timestamps. The workstation measured nanosecond-distinct mtimes (`mtime equal false 590352067 592036996`), and that does not tell us what the box's kernel does.
- **Evidence:** `go doc os.SameFile`: "SameFile only applies to results returned by this package's Stat." `(*os.File).Stat` qualifies. The order dependence follows from `atomicfile.WriteStream` replacing the path by `os.Rename` (`internal/atomicfile/atomicfile.go`).
- **Proposed fold:** Specify the order. Open the file, take `f.Stat()` on that descriptor, read from the same descriptor, and cache that `FileInfo` with that content. Do the same for the writer's post-write re-read. For the tick gap, add git's racily-clean rule: when the cached mtime is within one second of the moment it was observed, treat the cache as unverified and re-read on the next call. Add a reviewer-verified criterion, since a test cannot force the rename race deterministically.

### 4. MINOR: `encoding/json/v2` sits behind a default-on experiment, not a plain stdlib build

- **Location:** Global Constraints, lines 319-325. R-5c4. Task 0 note 6 (line 765).
- **Defect:** The statement "stdlib in go1.27.1 and builds with no `GOEXPERIMENT`" is true in practice but incomplete. Every file in the package carries `//go:build goexperiment.jsonv2`. The flag is on in go1.27's baseline, so anyone who sets `GOEXPERIMENT=nojsonv2` gets a broken build. An experiment-gated package is not the same stability class as GA stdlib, and the prior-art record should say so. v2 also behaves in ways the plan's table does not list: it rejects invalid UTF-8, it decodes `"admin":null` as false and `"streamer_username":null` as empty, and it rejects trailing data. None of these is harmful.
- **Evidence:** `src/encoding/json/v2/arshal.go:5` reads `//go:build goexperiment.jsonv2`. `src/internal/buildcfg/exp.go:87` sets `JSONv2: true` in the baseline. With `GOEXPERIMENT=nojsonv2 go build` the result is "build constraints exclude all Go files in .../encoding/json/v2". `golangci-lint 2.13.2` (built with go1.27.0) type-checked a v2 importer cleanly in the scratch module. Measured decodes: `{"admin":null}` gives nil error and `Admin:false`; `null`, `{}`, `{"schema_version":1}`, and `"members":null` each give nil error and nil `Members`; `"members":[]` gives non-nil `Members`; `"ADMIN"` and `"Admin"` fail as unknown members; the duplicate `"admin"` fails with `jsontext: duplicate object member name "admin"`.
- **Proposed fold:** Record the experiment gating in R26's note and in Task 1's prior-art line. Keep the adoption. Add the `null`-member-value behavior to Task 1's table as "loads" so nobody writes a refusal test that cannot pass.

### 5. MINOR: the lock wait's deadlines are unstated, and it holds the in-process mutex

- **Location:** Task 1 at lines 828-833. Task 7's row "roster lock held past the verb's deadline" (line 1474). Task 8's `ErrRosterBusy` mapped to 503 (line 1586).
- **Defect:** Neither the verb's deadline nor the API's is named. The API's natural `ctx` is `r.Context()`, which has no deadline. Task 1 has `Lookup`, `List`, and `AddMember` share one in-process mutex. So a serve-side write that waits on a foreign `flock` blocks every login and every gate check in that process for as long as `ctx` allows. `flock` is released at process death, so this bites only when a holder is stopped or wedged, which is a narrow case.
- **Proposed fold:** Name both deadlines as constants, for example a few seconds for the API and somewhat longer for the verb. Also require either that the `flock` is taken before the in-process mutex, or that reads never wait behind a writer's lock wait.

### 6. MINOR: the verb's actor comes from `os/user.Current`, which reads `$USER` in the box's build and reports the real uid

- **Location:** Task 7 at lines 1432 and 1443. R-5c14.
- **Defect:** `make build-linux` sets `CGO_ENABLED=0` (Makefile:15), so the box uses the pure-Go lookup. When `/etc/passwd` lacks the uid, `current()` builds `Username` from `os.Getenv("USER")`, so the audit actor can come from the environment. `Current` also uses `os.Getuid()` (the real uid), while R-5c14's ownership check uses the effective uid. Tests on the workstation run the cgo implementation, so they never exercise this path.
- **Evidence:** `go doc os/user` ("When cgo is available ... cgo-based (libc-backed) code is used"). `src/os/user/lookup_stubs.go` `current()`: `Username: os.Getenv("USER")` after `lookupUserId(uid)` fails, with `currentUID()` returning `os.Getuid()`.
- **Proposed fold:** Build the actor from `user.LookupId(strconv.Itoa(os.Geteuid()))`. Fall back to `uid:<n>`, never to `$USER`.

### 7. MINOR: the roster API's request encoding is unstated

- **Location:** The Goal (line 197, "a JSON API"). Task 8 POST (lines 1564-1570) and the admin-field row (line 1612: `admin=on`, `admin=false`, `admin=`).
- **Defect:** The 5b-ii handlers take a form-encoded body and answer JSON (`internal/auth/handlers.go:321-323`, `readForm`: `http.MaxBytesReader` and then `r.ParseForm`). The admin rows use form syntax, but "JSON API" invites a JSON body. How "present and empty" is detected depends on which it is. The plan also never says whether a query-string `?admin=` counts. `r.PostForm` excludes the query, and `r.Form` includes it.
- **Proposed fold:** State that the POST body is form-encoded through `readForm`, that "present" means `_, ok := r.Form["admin"]` (covering body and query), and that responses stay JSON.

### 8. MINOR: `doctor`'s failing-key names have no stated extraction method

- **Location:** R-5c10 and Task 5 at lines 1289-1293.
- **Defect:** Per-key failures arrive in `statuses`, but relational and format failures arrive as one `errors.Join` from `validateMagicLinkSource`, whose messages follow a "KEY: reason" convention (`cmd/dubplate-server/auth.go:255-285`, and `validatePublicURL` at `:358-377`). Getting names means unwrapping `Unwrap() []error` and cutting at `": "`. A message that breaks the convention would print its leading text as a "key". R-5c10 promises "key names only, never a value".
- **Proposed fold:** Specify the method: unwrap the join, cut at the first `": "`, and print a prefix only when it matches one of the source's Required keys, its Optional keys, or a declared threshold key. Otherwise print a fixed word such as `validation`. Add a mutation: "print the raw prefix, and a non-key message leaks text."

### 9. MINOR: the per-task rate is the optimistic end of 5b-ii's measured range

- **Location:** "Token ceiling", lines 242-259.
- **Defect:** 5b-ii's segments A and B each cost about 1.7M and 1.6M for three tasks, including boundary work (`docs/STATUS.md:12-16`), which is about 0.55M per task. The 0.46M figure matches the later segment (3.3M to 5.6M over five tasks). Step 3 also adds 0.30M for simplifier rounds that step 1's rate already "includes". At 0.55M, nine tasks come to about 4.95M and the pass to about 7.7M, past the 7.5M re-cut trigger at line 283.
- **Proposed fold:** State the range as 7.0M to 7.7M so the Task 6 / Task 7 cut is expected rather than a surprise. Keep the segment A re-derivation as the decision point. Remove the double-counted simplifier line or the "included" clause.

### 10. MINOR, OWNER FORK: the `streamer_username` character set excludes `@`, and it applies at load

- **Location:** R-5c6 (lines 487-495) and Task 2 at lines 971-973 ("at load, in `doctor`'s collision check, and at write time").
- **Defect:** The plan says it "cannot verify what the estate's Navidrome accounts hold", yet it applies a set that refuses `@` (email-style usernames) and spaces, at load as well as on write. If the group's Navidrome usernames fall outside the set, the write path cannot add those members, and rung 6's import cannot match them.
- **Options:** (a) As applied. (b) One read-only check of the estate's Navidrome usernames before Task 2 dispatches, then (a) or a widened set. (c) Apply the set on the write path only, and keep today's ASCII-printable rule at load.
- **Recommendation:** (b), then (a) if the names fit.

### 11. MINOR, OWNER FORK: refusing all of `Cf` in display names refuses ZWJ and ZWNJ

- **Location:** Task 2 at lines 969-970.
- **Defect:** `Cf` includes U+200C (ZWNJ) and U+200D (ZWJ). Persian, many Indic scripts, and emoji sequences need these characters, so refusing them in display names has a universality cost. The bidi overrides (U+202E) and U+200B are the real threat.
- **Evidence:** Measured: `U+200C Cf=true`, `U+200D Cf=true`, `U+00AD Cf=true`, `U+FEFF Cf=true`.
- **Options:** (a) Refuse all of `Cf`, as applied. (b) Allow U+200C and U+200D in display names only, and keep them refused in addresses and actors.
- **Recommendation:** (b).

### 12. MINOR: 5c-ii's runtime-prefix question has concrete evidence and a candidate mechanism the outline omits

- **Location:** The 5c-ii open questions, lines 1957-1968.
- **Defect:** The outline treats the question as open, but the kit source already answers half of it. On hydration the client resolves the route from `location.pathname` with the build-time `base` sliced off. Under a runtime prefix the build never saw, the route resolves to id `null`. The documented universal `reroute` hook runs in `hooks.js`, not a server file, so the gate allows it. It rewrites the URL before route matching and is the obvious candidate. The same docs page also says SPA fallback pages always use absolute paths. Separately, `Shell.Register` refuses a name with `/` (`internal/serve/serve.go` `subtreeNameRefusal`). So `login/link` must be served by the handler registered at `login`, next to the login page, and `_app` needs its own subtree.
- **Evidence:** In `@sveltejs/kit/src/runtime/client/client.js`, `_hydrate` calls `get_navigation_intent(url)`, and `get_url_path` returns `url.pathname.slice(base.length)`. In `utils.js`, `is_external_url` checks `!url.pathname.startsWith(base)`. The SvelteKit configuration docs, `paths.relative`: "Single-page app fallback pages will always use absolute paths, regardless of this setting." The `router.resolution` docs say the manifest is used "along with the `reroute` hook, if it exists", and the hook "allows you to modify the URL before it is used to determine which route to render". `router.type: 'hash'` is another documented option, but it disables prerendering.
- **Proposed fold:** Add these facts to the open question. List `reroute`, with the Go rewrite stamping the prefix into the document, as the first candidate to prove. List the single-segment `Register` constraint under the embed task.

### 13. MINOR: Task 1's file scope misses the second `List` caller

- **Location:** Task 1 Files (lines 793-794, "the request-link handler's store-failure branch only").
- **Defect:** `adminContact` (`internal/auth/handlers.go:238`) also calls `Roster.List()`, whose signature gains an error in Task 1.
- **Proposed fold:** Widen the note to "the store-failure branch and `adminContact`". On a `List` error the kind answer answers `failed`, per R-5c5.

### 14. MINOR: the full-map scan is necessary, and the plan should say so

- **Location:** Task 3 at lines 1079-1081.
- **Defect:** None in the mechanism. One question a reviewer might raise: why not check only the least recently used bucket? Refill time depends on tokens spent, not on recency. A bucket charged five times at T refills after one that was charged once, later. So scanning stops only at the first refilled bucket, and a full scan at capacity is required for correctness. Each scan runs up to `maxLimiterKeys` `TokensAt` calls under the map mutex, and every new key pays that cost under a flood.
- **Proposed fold:** Keep the scan. Have the doc comment say why a least-recently-used-only check would wrongly refuse keys.

## Checked and sound (do not re-litigate)

- **`flock` on a directory** (Task 1). `os.Open(dir)` read-only and then `syscall.Flock(LOCK_EX|LOCK_NB)` succeeds. A second descriptor in the same process gets `EWOULDBLOCK`. The lock still holds after a child `rename`, and it frees when the first descriptor closes. Opening the directory needs read permission on it (`open dir with wx only: permission denied`). `flock` is per open file description, unlike `fcntl`, so the two-`Roster` test models two processes.
- **`http.CrossOriginProtection`.** The quoted text is exact: "The GET, HEAD, and OPTIONS methods are safe methods and are always allowed". The stdlib alone allows a POST with neither header ("Requests without Sec-Fetch-Site or Origin headers ... are allowed"). The shell's `originCheck.allowed` refuses that POST unless a bearer bypass matches (`internal/serve/origin.go:103-115`), so Task 9's first acceptance row holds.
- **netip keying.** Measured: `192.0.2.7` keys as `192.0.2.7`. `::ffff:192.0.2.7` keys as `192.0.2.7`. `2001:db8:1:2:3:4:5:6` keys as `2001:db8:1:2::/64`. `fe80::1%eth0` keys as `fe80::/64`. An empty or garbage value fails to parse. `PrefixFrom` uses `ip.withoutZone()` (`net/netip/netip.go:1299`). `PeerAddress` strips the port with `net.SplitHostPort`, so no bracketed form reaches the parser.
- **`rate.Limiter`.** `TokensAt` and `Burst` exist in v0.16.0 with the quoted docs. A spent bucket reports exactly `Burst` after refill: 5, 3, and 7 all reach `>= Burst` true, and one nanosecond early gives 4.999999999998611, which is false.
- **`url.Error`.** `func (e *Error) Error() string { return fmt.Sprintf("%s %q: %s", e.Op, e.URL, e.Err) }` (`net/url/url.go:40`), as quoted. `validatePublicURL` quotes the value with `%q` today (`cmd/dubplate-server/auth.go:361-377`). `LinkURL` also wraps `url.Parse`'s error with `%w` (`internal/auth/token.go:161-166`), though load-time validation keeps that path unreachable.
- **The symbols and sites exist at HEAD:** `MintSession` at `session.go:102`, `RefreshSession`, `SessionCookie`, `linkPath = "/login/link"`, `LinkURL(publicOrigin, token)`, `TestDoctorCmd_LineCountFollowsRegistry` at `doctor_test.go:163`, `linkRequestsPerAddressCeiling` with its ten-times comment, `identitySourceStatus`, `newMailer` with `strconv.Atoi`, `newRootCmd`'s `AddCommand` block, the three route-name constants, and the "No handler here consumes the seam" comment (`serve.go:244`). `epoch.go` is 288 lines, as the plan says. The only callers of `AddMember` and `ErrRosterAdminChannelForbidden` are `roster.go` and `roster_test.go`. The nonce sweep's read-error branch retains the record with no mtime fallback (`nonce.go` `Sweep`). Removing a mode-0000 record needs only directory write permission.
- **Dependencies and independence marks.** Tasks 2 and 3 touch disjoint files, and so do Tasks 7 and 8. Task 5's nonce half is disjoint from everything else. The sequential chain 1 to 3 to 4 over `handlers.go` is correct. Each segment boundary sits on a task's committed green gate, and no task in the pass is irreversible.
- **Spec pointers spot-checked:** `spec:122-124` (the only path), `spec:1415-1418` (the admin checkbox), `spec:2331` (the 5c row), rung 6 `:1023` (6c's condition), rung 6 `:816-818` (the guard inside 6d's transaction), `Makefile:33`, and `scripts/check.sh:5` (the rejoin wording).
