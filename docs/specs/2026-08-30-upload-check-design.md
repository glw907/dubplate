# Check-on-upload: the family upload page (musicbox pass 2)

**Status (Geoff, 2026-08-31): input to a brainstorm, not a build-ready
spec.** The uploader effort starts with a fresh brainstorm that weighs the
product framing established after this document was written: the product
essence (a group building a well-organized library for a shared Navidrome
server), the adopter audience and ease-of-setup goal, the design bar below,
and the Go rewrite's internal/gate and internal/reject packages as the
build substrate. This document's mechanisms (stream validation, ZIP
central-directory peeking, the friendly-wording contract) enter that
brainstorm as prior art.

**Design bar (Geoff, 2026-08-31):** the upload experience is the estate's
one user-facing surface and carries the most product weight; it must be
well-polished, not merely functional. The product essence it serves: a group
of family or friends building a well-organized library for a shared
Navidrome server — contribution is the group's verb, and members judge the
system by this screen. The implementing pass invokes the workstation UI
skills (frontend-design for origination, design-refinement toward the
finish) and gates on fresh-context visual verification plus Geoff's own
before/after review prior to production.


**Status: draft for pass 2, which kicks off as soon as pass 1 lands** (Geoff,
2026-08-30). Gets its adversarial review and implementation plan at kickoff; early
family-upload experience feeds the design if any exists by then, but is not a gate.

## What and why

A purpose-built upload page replaces FileBrowser for the family flow: drag an album
(ZIP or files) onto the page, and the server validates it before accepting — wrong
format rejected in the browser at upload time with the same friendly wording as the
pass-1 email, instead of minutes later by mail. The insight from pass 1: the family
never needed a file manager, only a drop target, and no file manager's upload filter
can see inside the ZIPs that are the primary upload unit.

## Shape (to be pressure-tested at kickoff)

- A small Go service on the musicbox VPS (single static binary, systemd unit, Go per
  the workstation go-conventions), published through the existing tunnel at the inbox
  hostname behind the existing Cloudflare Access policy.
- **No app-level login.** Cloudflare Access already authenticates the family; the
  service reads the Access JWT's verified email, maps it through contributors.yaml to
  the uploader's inbox directory, and writes there with the `music` uid. One login
  disappears from the family's life (Access PIN remains).
- Validation at upload: peek the ZIP central directory and each FLAC's STREAMINFO
  header — format, sample rate, bit depth are readable in the first few hundred bytes
  without decoding. Reject per-file with the pass-1 reason wording, rendered in the
  page. The pass-1 pipeline gate stays as the deep check (`flac -t` full decode);
  upload-time checks are the fast 95%, not a replacement.
- **Pluggable validator seam** (Geoff, 2026-08-30): the checks sit behind a small Go
  interface (validator takes an upload's files, returns per-file verdicts with
  human-readable reasons), and the audio rules are its first implementation. The seam
  costs nothing now and is what would let the tool generalize to any validated-intake
  problem (or publish usefully) later; no second validator gets built in this pass.
  The identity check gets the same treatment for the same reason: a verifier
  interface with the Cloudflare Access JWT verifier as the sole implementation
  (trusted-header verifiers for Authelia-style setups are a someday, not a task).
- Accepted uploads land in `/srv/music/inbox/<user>/` exactly as FileBrowser's did;
  the pass-1 pipeline is unchanged.

## Decisions (brainstormed with Geoff, 2026-08-30)

- **FileBrowser Quantum retires.** The upload page fully replaces it for the family,
  removing a single-maintainer dependency, a second login, and the estate's largest
  attack-surface component. (Admin file access over SSH is unaffected.)
- **Uploads are resumable** (tus-style chunking): a dropped residential or iPad
  connection resumes instead of restarting a 700 MB ZIP. Prefer a proven embeddable
  Go tus implementation over hand-rolled chunking; evaluate at kickoff.
- **Success feedback is a stateless confirmation with an ETA** ("your album should
  appear in the library within about 15 minutes"). No upload history is kept.
- **Review status stays email-only** (the pass-1 friendly email); the page never
  tracks import outcomes.
- **The UI is stock Uppy Dashboard, themed to match Navidrome** as closely as easy
  effort allows (Geoff, 2026-08-30): Uppy's dark theme plus CSS-variable overrides
  for Navidrome's palette and type, and page chrome that reads as a sibling of the
  library UI. Theming is CSS-level only — no headless rebuild unless the family
  finds the result jarring. Execution follows the visual-fidelity method: reference
  screenshots of the deployed Navidrome UI anchor the work, and Geoff's eyes gate it
  before family onboarding.

## Kickoff questions (deliberately open)

- Whatever the family's early uploads reveal about dominant rejection reasons.
- Mobile Safari upload ergonomics for the iPad contributor (file picker fallback to
  drag-and-drop).
- Exact tus library choice and its SELinux/systemd hardening posture on the box.

## Out of scope

Browsing, sharing, or managing files in the inbox; any change to the import pipeline,
review flow, or notification emails; music playback.
