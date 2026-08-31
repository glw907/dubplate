# ROADMAP

> Strategic initiatives. Managed by `/log-project`. Issues tracked in `BACKLOG.md`.
> The estate renamed musicbox → dubplate 2026-08-31 (repo/paths converge at the
> rewrite's T14). Product essence: a group of family or friends building a
> well-organized library for a shared Navidrome server.

## Active

### Go rewrite `go-rewrite`
The box-side shell layer becomes one `dubplate` binary (import, backup, beet,
ping-fail, checks, doctor, restore), migrated per-subcommand with rollback,
carrying the estate rename. Fifteen tasks; a fresh session executes. Plan:
`~/.dotfiles/docs/superpowers/plans/2026-08-30-dubplate-go-rewrite.md` — spec
beside it. Starts after the refinement close-out commit lands on master.

## Planned

### Uploader `uploader`
The estate's one user-facing surface: `dubplate serve` + a bespoke Svelte 5 /
DaisyUI page replacing FileBrowser — album-unit validation at upload, an honest
journey to "in your library", a collections picker, kind rejection copy. Spec
(approved 2026-08-31 through three adversarial reviews, plan at kickoff):
`~/.dotfiles/docs/superpowers/specs/2026-08-31-dubplate-uploader-design.md`.
Supersedes the 2026-08-30 upload-check draft. Runs after `go-rewrite`; three
pre-pass gates (SSE spike through the real tunnel, chunk-size timing, real-iPad
picker check).

### Packaging `packaging`
Make the estate easy for a stranger to set up from its ingredients (Navidrome,
beets, the binary, their Cloudflare account, any VPS). Opens with a brainstorm;
context staged in agent memory (distribution model: binary-as-installer vs
clone floor; cloud-init generation; existing-box bootstrap; pre-setup-VPS
option map). Runs after `uploader` so it templates the final intake. The
`oss-pair` write-up below doubles as its demand test.

## Someday

### dubplate as a music library organizer `dubplate-organizer`
Extend dubplate into a batteries-included organizer that runs two ways on
one shared engine: locally as a bubbletea TUI (no tunnel, no Access — the
trust surface deletes itself) or hosted with the web frontend, with
Navidrome-native integration (scan observation, collection routing,
library-aware imports) as the bonus only dubplate can offer. Core thesis:
harvest beets' battle-tested matcher into Go behind a behavioral
conformance harness that tracks upstream releases (autotag churn is
refactors, not retunes — measured 2026-08-31), and ship the empirically
top plugins as built-ins: art fetch/embed, genre, duplicates, tag
hygiene, fingerprinting, ReplayGain (28–90% adoption in a 87-config
survey; scrapers and the sub-10% tail refused permanently). Staged
adoption: beets stays behind the subprocess seam until the harvested
matcher proves conformance, then swaps at that one seam. Two ladder
implications settled early: migration between rungs is backup/restore
(T10b already frames restore as the adopter's path — local-to-VPS is
backup-from-laptop, restore-to-box); and the favored shape is TWO
binaries from one repo and engine (Geoff, 2026-08-31) — `dubplate`
(local: TUI, import, restore; requires only a music dir; no web/tusd/JWT
in the dependency graph) and `dubplate-server` (today's estate, keeping
its full required-no-defaults wall) — so each binary owns its
requirement wall outright and no mode-scoped config matrix exists.
Shared subcommands are shared cobra constructors, not duplicates. The
TUI is remote-capable (Geoff, 2026-08-31): it binds the same operations
in-process locally or over dubplate-server's API remotely
(Access-authenticated via the cloudflared access token flow; admin =
AdminEmail match, no role system) — so the server stays headless, ssh
stops being the admin interface, and review-queue triage becomes the
TUI's first remote job. Admin API surface is this initiative's, NOT the
uploader pass's (its spec's out-of-scope wall holds). A web admin page
stays a deferred option the API spine keeps cheap — it earns its build
only if phone triage (approve from the review email, no laptop) proves
wanted in practice; until then admin verbs stay off the browser origin.
Evidence:
`~/.dotfiles/docs/superpowers/research/2026-08-31-beets-plugin-usage.md`
(+ tally JSON beside it). Preconditions: uploader and packaging shipped;
the packaging story and this mode ladder converge deliberately. Same
maintainership caveat as oss-pair.

### Open-source the pair `oss-pair`
Publish the uploader and the pings Worker as documentation-linked siblings —
a validated-intake drop target and a dead-man's-switch monitor, proven here
first. Preconditions: months of family use, the uploader's auth/validator
seams, and a clear-eyed maintainership decision. First deliverable,
independent of any code release: the validated-intake pattern write-up (what
each layer catches that the others cannot; 907.life is a natural home) —
zero maintainership cost, and the packaging demand test.
