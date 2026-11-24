# 7. Refused features

## Status

Accepted

## Context

dubplate is for a group sharing music its members hold, and is built to be easy for that and hard
to turn into a distribution hub. The principles are ratified in
`docs/specs/2026-09-25-group-scale-principles-design.md`, and the refused list first landed in
`CLAUDE.md` at commit `cfd0962`.

## Decision

dubplate never builds, and never accepts a contribution that adds, any of these:

- anything that fetches music: search-and-download, torrent, Usenet, or Soulseek integrations, or
  "fetch missing albums" automation (metadata sources are a separate question);
- federation, or any link between two servers' catalogs;
- a public library: anonymous browsing, public collections, or a link that opens a collection;
- a `download` larger than one album (song or album ids only; other kinds answer error 50);
- a share outside the six bounds: one item, lossy stream only, a clamped expiry, the listener pause,
  an unguessable and unindexed link, and bounds enforced server-side with the owner from login.

## Consequences

Review blocks a change that adds a refused feature.
