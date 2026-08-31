# ROADMAP

> Strategic initiatives. Managed by `/log-project`. Issues tracked in `BACKLOG.md`.

## Planned

### Check-on-upload page `upload-check`
Pass 2: a Go upload service replacing FileBrowser for the family flow. Resumable
uploads (embedded tus, stock Uppy dashboard themed to Navidrome), format validation
at upload time behind pluggable validator and identity-verifier seams, Cloudflare
Access identity. Spec (decision-complete, review at kickoff):
`~/.dotfiles/docs/superpowers/specs/2026-08-30-upload-check-design.md`. Kicks off
when pass 1 lands.

### Architecture refinement pass `arch-refinement`
Adversarial review of the estate as built — simplification, spec/reality drift,
security posture as one system, altitude, test-suite quality, operational truth —
then a triaged fix wave. Charter (frame ratified, plan at kickoff):
`~/.dotfiles/docs/superpowers/specs/2026-08-30-architecture-refinement-charter.md`.
Sequencing vs `upload-check` is Geoff's call at pass-1 close.

## Someday

### Open-source the pair `oss-pair`
Publish the upload-check tool and the pings Worker together as a paired release of
small self-hosted infrastructure: a validated-intake drop target and a
dead-man's-switch monitor, both proven here first. Both fill verified gaps (no
content-aware upload validation exists in the self-hosted music world; no maintained
push-monitoring Worker exists on Cloudflare). Preconditions: months of family use,
the pass-2 auth/validator seams (already specced), and a clear-eyed decision about
taking on maintainership of internet-facing tools. Ship as documentation-linked
siblings, not a bundle; the musicbox pipeline becomes the reference deployment, not
a product.

First deliverable, independent of any code release: a pattern write-up others can
build from — the validated-intake architecture (authenticated drop target -> upload-
time format checks -> headless gatekeeper -> curated library -> locked backup ->
dead-man's-switch monitoring), what each layer catches that the others cannot, and
the decisions the adversarial reviews forced. Publishable as writing (907.life is a
natural home, in the site's own voice) with zero maintainership cost; the code
release, if it ever happens, links back to it.
