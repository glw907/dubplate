# Phosphor icons

`SiteHeader.svelte` inlines two Phosphor icons as raw SVG markup, written
directly in the component template rather than through an HTML-injecting
directive (no npm dependency either): `sun` and `moon`, for the theme
toggle's two faces.

License: MIT. Source: [phosphor-icons/core](https://github.com/phosphor-icons/core).

**Weight note.** The visual identity spec calls for the bold weight. This
workstation has no network access in this environment (the harness's
"no network" rule), so the exact bold-weight path data and the pinned
release tag could not be fetched or verified against the upstream
repository for this pass. The path data inlined here is Phosphor's
**regular** weight, taken from a locally available, already-vetted copy
in another repo on this workstation
(`aksailingclub-legacy/assets/icons/{sun,moon}.svg`, sourced there under
the same MIT license). A later pass with network access should fetch
the bold weight from a pinned tag and update this file with the tag,
the exact source URLs, and Phosphor's license text verbatim.
