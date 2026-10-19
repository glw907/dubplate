# Gates field research + iPad test infra (2026-08-31, for the approved uploader plan)

Sonnet dispatch, web-sourced with citations (full source list in the
session transcript; key links inline). Feeds gates 1-3 and U9.

## 1. tus behind Cloudflare - field reports

- The binding constraint is Cloudflare's 100s proxy read window (Free/
  Pro/Business); chunked tus sidesteps it because each PATCH is its own
  short request. Field convergence on chunk size: 5-10MB (10MB showed
  ~1s stalls; 5MB fixed failures at 50MB+); nobody targets large chunks.
  CONFIRMS the plan's 8MB default; warns 16MB risks tail latency on slow
  uplinks. (community.cloudflare.com threads: "TUS stop upload on chunk
  size", "413 for large videos")
- Working retry posture in the wild: tus-js-client
  retryDelays [0, 3000, 5000, 10000, 20000]. (fueled.com tus writeups)
- THREAT, unverified secondary source: a community report says uploads
  through Cloudflare Proxy are buffered such that partial offsets are
  not visible until the request completes - which would degrade
  RESUMABILITY (HEAD offset after interruption). Gate 2 must verify
  offset persistence through the real tunnel, not just chunk timing.
  Also: the free-plan 100MB request-body cap is absolute at the proxy;
  per-chunk sizes are far below it, but confirm the zone's plan limits
  once. (cloudflared#1095 response buffering; community thread
  "Uploads are not resumable if you use Cloudflare Proxy", 403'd on
  direct fetch - verify in gate 2)
- Access cookies do not corrupt tus semantics; the one field bug is
  cookie-domain scoping across subdomains (scope to parent domain).

## 2. SSE through cloudflared - field reports

- cloudflared buffers GET-based SSE (whole stream flushed at close) in
  a documented open issue; POST-based SSE streams correctly
  (cloudflared#1449). Content-Type: text/event-stream signals the edge
  to stream (secondary, confirm).
- Separate hazard: cloudflared's HTTP/2 path shows a ~60s idle-stream
  cut absent keepalive - INSIDE the plan's <=20s heartbeat but the
  proven field interval is 15-30s; 15s gives margin.
- Proven mitigation set: POST-opened stream, text/event-stream +
  no-transform, heartbeat comment every 15s, flush after every write.
  If a spike with all four still misbehaves: ship polling, stop chasing
  (polling is confirmed reliable through the same stack).

## 3. iPad/iOS-Safari testing from a Linux-only shop

- BrowserStack free tier CANNOT test file uploads (file injection is
  paywalled to Team/Enterprise plans) - ruled out for gate 3.
- LambdaTest ~$15/mo Live is the cheapest paid real-device farm; verify
  large-file injection before paying.
- Playwright WebKit is NOT iOS Safari: documented divergence on file
  inputs (playwright#20079), no real memory ceilings or backgrounding.
  Use as CI smoke for upload-form JS only; never as iOS truth.
- RECOMMENDED RIG (free): the household iPad over USB to the Linux
  workstation via usbmuxd + ios-webkit-debug-proxy (or the
  ios-safari-remote-debug-kit fork) -> DevTools at localhost - real
  console/network visibility during a real upload (stalls, memory
  warnings, dropped connections). Rough-edged but functional
  (jade.fyi walkthrough); inspect.dev is the paid fallback if the free
  chain blocks. Geoff's testing role collapses to plugging the iPad in.
