# Rung 5d-ii close: architecture reads

Three `go-architecture-reader` dispatches (no plan in context) over `internal/serve`,
`cmd/dubplate-server`, and `internal/notify`, 2026-10-05, after Task 9's folds. The once-per-pass
simplifier ran first over 9B and 9C's Go and found nothing. The readers graded `internal/serve` and
`cmd/dubplate-server` "workmanlike" and `internal/notify` "sound with nits", and found no correctness
defect on any production path. Nothing here is fixed in code. Findings route two ways: forward hooks
that rung 12's Navidrome driver may no longer need go to the brainstorm that re-plans rungs 12 and 16
under the 2026-10-05 ruling (the box serves only dubplate's development; the incumbent stack is
retired); everything else goes to the architecture-cleanup pass, alongside
`2026-10-05-rung-5d-i-architecture-reads.md`.

The condensed reports are the source; line numbers are the readers' at the post-fold tree.

## internal/serve

Density 0.86 comment-to-code (1365/1579 non-test) against `net/http/httputil` at 0.45.

Routed to the rungs 12 and 16 re-plan, because each exists only for a consumer that may not be built:

- `FrontExemption` and `NewFrontExemption` (`strip.go:238-267`) have no production caller
  (`cmd/dubplate-server/serve.go:324` passes the zero value). It matches the proxy set on the raw
  `PeerAddress` string (`strip.go:263`) while `TrustedProxyConfig.trusts` (`clientaddr.go:194-204`)
  unmaps, so a dual-stack `::ffff:a.b.c.d` front never matches. This is 5d-i's finding, still open. The
  reader would remove the type, `NewStrip`'s parameter, `collapseToLast`, and the exempt threading
  until a consumer exists, then rebuild it keyed on the unmapped `netip.Addr`.
- `Shell.SetFallthrough` (`serve.go:147`) has no non-test caller; unexport it and add the
  `handlerBuilt` panic, or delete it with the fallthrough.
- `NewStrip`'s driver `header` parameter (`strip.go:310`, `:208-214`, `serve.go:497`) is `""` at
  every production call.
- `addBearerBypass` (`origin.go:55`, `:61-69`, `:88-91`, `:114-115`) has no non-test caller, so the
  bypass mux runs in production against an empty mux.

Routed to the architecture-cleanup pass:

- Unexport `Listen` (`server.go:113`), `PeerAddress` (`clientaddr.go:137`), `Normalize`
  (`normalize.go:83`), `Prefix.Owns` and `Prefix.Route` (`prefix.go:116`, `:149`), `Strip.Wrap`
  (`strip.go:344`), `RewriteDocument` (`document.go:31`), and `ThemeCookieName` (`document.go:21`);
  the sentinels `ErrInvalidBindAddr` (`bindaddrs.go:13`) and `ErrInvalidPrefix` (`prefix.go:14`) are
  matched only by tests. `Shell.Logger` (`serve.go:240`) round-trips the caller's own logger, and
  `Run` handles nil twice (`server.go:140-142`).
- Duplication: `clientaddr.go:61` recomputes `foldedKey` inline (use `strip.go:401`);
  `clientaddr.go:261-292` hand-rolls `strings.SplitSeq` and `Trim`; `prefix.go:85-95` restates
  `normalize.go:119-129`'s segment walk. Three sites parse and unmap the peer (`clientaddr.go:198-203`,
  `serve.go:216-221`, `probe.go:99-105`); one `peerAddr` helper would make their agreement structural.
  Polish: the parallel `forwarded` and `forwardedCanon` lists (`strip.go:293-297`), `countPresent`
  rebuilding the folded set per request (`serve.go:224`, `probe.go:125-127`), and `strip.go:552-558`
  as `slices.ContainsFunc`.
- File split: `document.go` mixes document serving, static assets, request logging, and build-artifact
  scanning (suggested `document.go`, `assets.go`, `theme.go`, `log.go`, with `statusCapture`,
  `refusalRecord`, `markRefusal`, `logRefusal`, `boundedPath` moved out of `serve.go`).
  `clientaddr.go:1-3` carries a second package comment that contradicts `doc.go`.
- Stale comments: `strip.go:274-276` ("Task 5 composes Strip as the outermost step"),
  `prefix.go:139-142` (calls `Route` the third step; it is step 5 in `doc.go`, and says the mux
  default answers unclaimed owned paths, which `Shell.terminated` now does), `prefix.go:16-18` and
  `server.go:110-111` (rungs and a "2.0 hook" in present tense).
- Process citations: `doc.go:1`, `:33-38`; `serve.go:18`, `:70`, `:290-295`, `:482`; `normalize.go:11`,
  `:38`, `:71`, `:79`; `server.go:29-36`, `:61-63`, `:131-133`; `prefix.go:79-84`, `:99-100`,
  `:108-115`, `:125-130`; `freebind_linux.go:10-16`; `document.go:20`, `:24`, `:296`, `:345`, `:385`;
  `strip.go:26`, `:75`, `:236-237`, `:269`, `:445`; `serve.go:357-359` narrates a past bug. The
  citation lists at `strip.go:26-111` and `:137-173` earn their place as header provenance.
- `newHTTPServer`'s `writeTimeout` parameter (`server.go:77-83`, `:92`) has one production value.

Best passages: `strip.go:441-498` (`trailerScrubbingBody`, holding the request and scrubbing on
`Close` and EOF); `clientaddr.go:206-235` (`rightmostUntrusted`, stopping at the first malformed
untrusted hop); `serve.go:444-499` with `markRefusal` (the refusal record filed by the refusing
wrapper, never inferred from status).

## cmd/dubplate-server

Density 0.428 (1574/3681 non-test) against `cmd/go/internal/modcmd` at 0.141. Read in full:
`doctor.go`, `doctorprobe.go`, `serve.go`, `auth.go`, `web.go`, `main.go`, `proxy.go`, `check.go`,
`store.go`, `member.go`; sampled the rest. All routed to the architecture-cleanup pass, ranked by
the reader:

- Latent defect, `doctor.go:47-202`: each row repeats status, failed, and `printRow`, and line 198
  ORs eleven flags, so a twelfth row that forgets its term exits 0 on failure. One accumulator
  closure plus a sticky-error writer checked at `Flush` removes the class. This is the one finding
  that can turn into a false green; rank it first.
- Key names spelled as literals in `auth.go:46-60`, `:268-273`, `:279`, `:285`, `:369`, `:372` while
  `serve.go:28-40` declares constants and `doctor.go:217` a third local `const rosterKey`; the same
  pattern at `shadow.go:32-35` against `:62-66` and `backup.go:131` against `:424`. This repeats
  5d-i's largest finding.
- `auth.go:297`: serve's port range check (`validateServePort`, `serve.go:122`) runs inside the
  magic-link source's `Validate` hook, so serve config is enforced only while that source is
  selected. It belongs to serve's own load-time check.
- `doctorprobe.go:56-79` rebuilds the trusted-address set `trustedProxy` (`doctorprobe.go:303-317`)
  builds; `doctor.go:345-350` inlines `foldRefusal` (`doctorprobe.go:36-38`). Three owner checks
  (`doctor.go:290-296`, `unidecode.go:272-283`, `member.go:638-652`) read `Stat_t` and compare to the
  euid, and `member.go:647` calls `os.Geteuid` directly, bypassing the `geteuid` seam
  (`unidecode.go:47`); one `fileOwner` helper. `serve.go:388-391` and `doctorprobe.go:103-106` build
  `privateFiles` identically; the ten auth thresholds are listed by hand four times (the var block,
  `auth.go:228-237`, `:303-329`, `:522-561`).
- Test-only seams: `newServeShell` returns a channel production discards (`serve.go:319`,
  `:398-399`, admitted at `:293-297`); `startNonceSweep` (`serve.go:500`) is already tested directly;
  `serve.go:434-461` exists "so a test can read the config it built", but `serve_test.go:654`
  rebuilds it with nil deps and never observes the real one, so a wiring slip at `serve.go:366` would
  pass; the "not registered" branch at `doctorprobe.go:250-255` exists only for other-family test
  registries; `web.go:226-231`'s `buildFS` parameter is acceptable.
- `shadow.go:48` `ShadowIntegration` and its three fields are the only exported identifiers in
  `package main`.
- File split: the auth key constants and four route names in `serve.go:23-66`; the package-wide
  helpers `lookupEnviron`, `printRow`, `splitKeyReason`, `flattenErrors` (`doctor.go:352-410`,
  `:578-603`) belong in `env.go` or `report.go`; `store.go:209-213`'s `backupFactRow` serves settings
  rows (call it `doctorRow`); `checkDirWritable` and `checkRosterDirWritable` (`serve.go:137-167`)
  belong with the filesystem checks.
- Comments: stale line citations at `auth.go:517` (cites `serve.go:81-106`, now `194-225`),
  `serve.go:99`, `serve.go:420-421`; process citations and rebuttals at `auth.go:38-42`, `:174-178`,
  `:218-224`, `:492-498`, `serve.go:68-82`, `:298-303`, `doctor.go:87-107`, `:363-371`, `:413-427`,
  `:474-488`. Navidrome justifications remain at `serve.go:172-174` and `store.go:20-22` though
  Navidrome is retired. `main.go:3-21`'s package doc lists `import` and `ping-fail` verbs that
  `newRootCmd` (`main.go:139-149`) does not register and omits `member`.
- Low: `validateSMTPPort` and `validateSMTPFrom` (`auth.go:351`, `:354`, `:369`, `:372`) return plain
  errors where the hook's refusal shape is `*sourceKeyError`; `authConfig.unknownAddressAnswer` is
  compared to `"generic"` at `serve.go:459` and could resolve once to a bool.

Best passages: `web.go:658-709` (`readOperatorCSS`, descriptor-checked, TOCTOU and FIFO closed, no
path in errors); `auth.go:445-448` (`validatePublicURL`'s `@` pre-check, since `url.Parse`'s error
would leak a password); `doctorprobe.go:407-415` (`probeAnswer` pointer counts, so a missing count
fails loudly).

## internal/notify

Density 0.61 (53/87, `notify.go`) against `net/smtp` at 0.36 and `net/mail` at 0.21. Verdict "sound
with nits". All routed to the architecture-cleanup pass:

- Test-only seam: `forceImplicitTLS` (`notify.go:66-71`, `:127`) is set only by
  `TestSend_ImplicitTLS465` (`notify_test.go:412`), so the real `m.cfg.Port == implicitTLSPort`
  comparison is never tested and a wrong constant would pass. Use `var implicitTLSPort = 465` and
  swap it with `t.Cleanup`.
- Test-only seam: the `tlsConfig` field (`notify.go:60-64`, `:132-134`) is never set by `New`. A
  package-level var set by `TestMain` fits, since the package runs no `t.Parallel`.
- Exported surface: `ErrHeaderInjection`, `ErrInvalidFrom`, `ErrInvalidRecipient`, and `ErrSend`
  (`notify.go:28`, `:33`, `:38`, `:44`) have no reference outside the package; auth's
  `logSendFailure` (`internal/auth/handlers.go:296`) logs `err.Error()` only.
- `From` is validated at config load (`validateSMTPFrom`) and again per `Send` (`notify.go:99-106`
  against `cmd/dubplate-server/auth.go:367-375`), so `ErrInvalidFrom` is unreachable in production.
  Validate once in `New(cfg) (*Mailer, error)`; its sole caller `newMailer` (`serve.go:483`)
  already returns an error.
- Error chain: the fallback at `notify.go:141-146` formats with `%v`, losing `context.DeadlineExceeded`
  from the chain; wrap with `%w`.
- Comments: `notify.go:1-5` names callers that do not exist and build history; `notify.go:117-122`
  pins go-mail v0.8.1 source citations that rot on a bump. Test nits: `notify_test.go:622-623`
  cites "Global Constraints", and `:295-302` takes the mutex twice.

Best passages: `notify.go:141-146` (drops recipient addresses from `SendError`, keeps reason and
code); `notify.go:127-131` with `notify_test.go:491-514` (`TLSMandatory` fails closed, pinned by
test); `notify_test.go:347-352` and `:537-547` (hand-derived encodings, closed-listener guard tests).
