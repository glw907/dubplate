# Project site spec review: drift and failure risk

Reviewer lens: data integrity and failure risk, adapted to a website. Target:
`docs/superpowers/specs/2026-09-25-project-site-design.md` at `e91f528`. Line numbers
refer to that commit.

The spec's central promise (line 20-21, 88-106) is that the docs and theme tokens can't drift
between dubplate and the site. The lock file proves one narrow thing: the site's vendor directory
equals what a sync once wrote. It does not prove the pin is current, that the guide matches the
code, that the synced bytes came from the tag, that the file set is complete, or that the tokens
render the same way on the site as in the app. Four of the gaps below break the promise as
stated; the rest are cheaper hardening. Findings are ranked by consequence.

Facts checked in the tree: `web/src/lib/themes/` holds `dubplate.css`, `fonts.css`, `inks.css`,
`label.css`, `label-night.css`, `surface.css`, `system.css`. The font files are not in that
directory; they live in `web/static/fonts/` with `FONTS.md` and a `licenses/` directory, and
`fonts.css` references them by absolute URL (`url('/fonts/...')`). The theme files' comments cite
internal paths (`docs/research/2026-09-24-visual-identity-draft-tokens.md`, spec
line ranges), gate names (G2, G9), and depend on `daisyui@5.7.46` and a Tailwind 4 parsing
behavior. `docs/guide/` does not exist yet.

## Summary

| ID | Severity | Gap |
|---|---|---|
| D1 | major | The pin can go stale with the gate green; a stranger sees guides for a different version than the release |
| D2 | major | The hand-written guide can describe behavior the code at the same tag no longer has |
| D3 | major | The sync's source is undefined: a dirty or wrong checkout, a moved tag, or a working-tree generator run all pass the gate |
| D4 | major | The token files move by hash, but the guarantees they carry (contrast, completeness, font loading) are proven only in dubplate's toolchain |
| D5 | major | A cairn editor commit or a push can deploy past the hash check if the check isn't in the deploy build |
| D6 | major | Vendoring from a private repo can publish internal paths, the production box, or member identities |
| D7 | minor | Added, removed, or half-written vendored files aren't caught by a per-file hash check |
| D8 | minor, OWNER FORK | Release notes and site-owned pages restate vendored facts with no link to their source |
| D9 | minor | The comparison page's re-verify trigger is tied to dubplate releases, but other projects change on their own clocks |
| D10 | minor, OWNER FORK | One site, one version: operators on an older release lose their guide at the next bump |

## Findings

### D1. The pin can go stale with the gate green (major)

Location: lines 102-106, 132-134.

The gate compares vendored files against the lock. It never compares the lock against anything
outside the site. If nobody bumps the pin after 0.8.1 ships, the gate stays green and the site
publishes 0.8.0's guides next to a 0.8.1 download. The only guard is "dubplate's release
checklist gains a step" (line 104-106), a written step in a different repo, which the workstation
rule "a rule lives where it executes" calls a record, never a delivery. The Phase B precondition
"the guides describe the tagged 0.8 release" (line 133) is also unenforced, and Phase A has no
0.8 tag to sync from at all (line 126-127), so the preview must vendor from something the spec
doesn't name.

What a stranger sees today under this design: guides with no version marker, which silently
disagree with the binary they downloaded on a flag, a config key, or an `upgrade` step.

Proposed fold:

1. Every vendored and generated page renders "Describes dubplate `<tag>`" from the lock. This is
   the cheapest fix and it turns silent drift into visible drift.
2. The sync accepts a pre-release tag (for Phase A) and the page banner says "preview of an
   unreleased version" when the tag isn't a release. The Phase B deploy refuses a pre-release
   lock.
3. The bump lives in the tool, not the checklist: the release step that cuts a dubplate tag
   (the `packaging` initiative's release script) runs the site sync, or at minimum prints the
   exact `npm run sync-dubplate -- <tag>` line and exits non-zero until the site lock names that
   tag. Name which of these Phase A builds.

### D2. The guide can disagree with the code at the same tag (major)

Location: lines 55-57, 79-83, 90-92.

The lock proves the site equals the tag. It proves nothing about whether `docs/guide/` matches
the code at that tag. The three generated pages close that gap for reject wording, config keys,
and commands. The hand-written guides ("Set up a server", "Run your server", "For members")
still name commands, flags, keys, and paths in prose, and nothing checks them. A renamed flag
updates the command reference and leaves the setup guide wrong, at the same tag, with every gate
green. This is the drift a stranger actually hits.

Proposed fold: a dubplate gate leg (it lands with the `docs/guide/` lane) that extracts every
backticked `dubplate ...` command and flag and every config key named in `docs/guide/*.md`, and
fails when one is absent from the cobra tree or the key registry. The generators already walk
those structures, so the leg reuses them. State in the spec that the guide is held to the code by
this leg, not by review alone.

### D3. The sync's source is undefined (major)

Location: lines 96-99, 79-82.

"Exports the owned files from a dubplate release tag in a local dubplate checkout" (line 96-97)
admits three readings, and two of them are unsafe:

- If the script copies from the working tree after a `git checkout <tag>`, a dirty tree, an
  untracked file, or a checkout left on another branch is vendored under the tag's name. The lock
  then records the tag and commit next to bytes that commit never contained.
- The generated pages must come from somewhere. If the sync runs the generators, it builds
  dubplate from the working tree, with the same exposure, plus a dependency on the builder's Go
  toolchain.
- A tag can be moved (`git tag -f`) or be lightweight. A re-sync of the same tag name at a
  different commit changes the vendored bytes while the lock's `tag` field stays the same.

Proposed fold:

1. The sync reads only from git objects: resolve the tag to a commit, then `git archive <commit>
   -- <manifest paths>` (or `git show <commit>:<path>`). It never reads the working tree, so a
   dirty or wrong checkout can't leak in.
2. The generated pages are committed in dubplate (for example `docs/guide/reference/`), with a
   dubplate gate leg that regenerates and diffs them. The tag then carries its own generated
   output, and the sync copies it like any other file, never runs Go.
3. The sync refuses a lightweight tag, and refuses when the lock already names the same tag at a
   different commit.

### D4. The tokens' guarantees don't travel with the tokens (major)

Location: lines 90-94, 108-110.

The spec says the token files "move as files, with no translation step" (line 93-94). The bytes
do; what they mean doesn't. Specifically:

- The contrast floor, theme completeness, and stamped values are proven by dubplate's G2 and G9
  against dubplate's built bundle, DaisyUI `5.7.46`, and a specific Tailwind 4 behavior the
  `dubplate.css` header documents (quoted font stacks). The site builds the same files with
  cairn's DaisyUI and Tailwind. A version difference changes the emitted values with the hash
  unchanged.
- `inks.css` exists because the palette's mustard fails 4.5:1 as text, and dubplate's markup
  reaches text color through ink utilities, never `text-warning`. The cairn chassis and the
  site's own markup don't follow that rule, and the component layer that enforces it is the part
  the spec says won't travel (line 109-110). The site can ship text below the floor that the app
  never does.
- The fonts aren't in `web/src/lib/themes/` (line 91 says they are). They live in
  `web/static/fonts/` with their licenses, and `fonts.css` loads them from the absolute path
  `/fonts/`. If the manifest takes `fonts.css` and not the files, or the site serves them
  elsewhere, the font stacks fall back silently.

Proposed fold: the lock also records the DaisyUI and Tailwind versions dubplate's gate proved
the tokens against, and the site gate fails on a mismatch, or the site runs its own contrast and
completeness check over its built CSS (the stronger form, since it also covers the site's own
markup). The manifest names `web/static/fonts/*.woff2`, `FONTS.md`, and `licenses/`, and the site
serves them at `/fonts/`. Correct line 91's location.

### D5. The hash check may not run where the site deploys (major)

Location: lines 40-42, 46-48, 102-103.

Cairn's publish flow commits editor edits straight to main (line 40-42), and a cairn site on
Cloudflare typically builds on push. The spec says "the site's gate fails" (line 102) but not
where the gate runs. If it runs only in the local pre-commit gate, three paths deploy past it:
an editor commit, a sync commit pushed without the gate, and a hand fix pushed from another
machine. The spec also says the editor "can't change" vendored pages (line 47-48) without saying
how; if the vendored guides sit in a cairn content collection so they render, the editor lists
them as editable.

The concurrent case the lens asked about is benign on its own: an editor commit and a sync commit
touch disjoint paths, so the later push fails fast-forward and rebases cleanly. It becomes a
defect only when combined with the gaps above.

Proposed fold: the lock check runs inside the deploy build command (a `prebuild` step in the
command Cloudflare runs), so a mismatched tree never deploys from any path. Vendored pages live
outside every editor collection, and the spec names the mechanism (a read-only content source,
or a route that renders from the vendor directory).

### D6. Private-to-public leakage (major)

Location: lines 90-100.

The source repo is private, and its writing habits reflect that. The token files' comments today
cite `docs/...` paths, spec line ranges, and gate names. The guides will be new, but
written by agents that work in a repo whose CLAUDE.md names the production box
(`5.78.148.42`), the group's hostnames, Fastmail, and members. A setup guide drafted from the
live estate can pick up any of those. Whether the site repo itself is public is also unstated; if
it is, the vendored files' comments publish even when minified CSS drops them.

Proposed fold: the sync runs a deny-list scan over the staged files and refuses to write on a
hit: the box IP and any internal hostname, `docs/{archive,plans,research,specs}`, `_salvage`, `/var/home`, `~/`,
email addresses outside an allow-list, and gitleaks. The same scan runs as a dubplate gate leg
over `docs/guide/`, so the leak is caught at authoring time, not at sync. State whether the site
repo is public.

### D7. File set and interrupted syncs (minor)

Location: lines 97-103.

A per-file hash check catches an edit to a listed file. It misses a file present in the vendor
directory but absent from the lock (a stale page left after dubplate deleted or renamed it), and
depending on implementation, a lock entry whose file is missing. An interrupted sync can leave a
half-written vendor directory with the old lock. Separately, if the list of owned files lives in
the site's sync script, a new guide page or a new palette file added in dubplate is never
vendored and nothing notices.

Proposed fold: the gate asserts set equality in both directions between the vendor directory and
the lock. The sync stages into a temporary directory and swaps it in, writing the lock last. The
owned-file manifest lives in dubplate (whole directories, such as `docs/guide/`, plus a named
theme list), so the tag defines the set.

### D8. Site-owned pages restate vendored facts (minor, OWNER FORK)

Location: lines 52, 61-62.

Home carries "what dubplate is in one paragraph", About carries the license and the stance on
buying music, and release notes describe what changed. Each restates facts whose source is in
dubplate (the is/isn't lines, `LICENSE`, the principles spec, the code). Home and About are small
and change rarely: a fold is to render the is/isn't lines on Home from the vendored file rather
than retyping them.

Release notes are the real fork, since each post describes behavior and will be written from
memory of the release.

- Option A: site-owned posts, as drafted. Editorial voice, editable in cairn, and unchecked
  against the code.
- Option B: dubplate owns a `CHANGELOG.md` that travels in the manifest; the site renders it
  as the release notes page.
- Option C: both. The vendored changelog is the fact record, and each site post is the narrative
  that links to that release's changelog entry.

Recommendation: C. The post keeps its voice, and a factual disagreement is visible on one page.

### D9. The comparison page's re-verify trigger (minor)

Location: lines 72-74.

"Re-verified before each release that touches it" ties re-checking to dubplate's releases, but
other projects change on their own schedules, and the page is site-owned, so no dubplate release
ever "touches" it. The dated citations are the right mechanism with no clock attached.

Proposed fold: the site gate fails when any comparison claim's checked date is older than a
fixed age (for example six months), and the Phase B publish requires every date to be current.

### D10. One version on the site (minor, OWNER FORK)

Location: lines 12, 147-148.

The site describes one release. When 0.9 ships and the pin moves, an operator still on 0.8 loses
the guide for the server they run, and `dubplate upgrade`'s instructions describe the target
version, not the one they're leaving.

- Option A: latest only, with the D1 version banner. Simplest; fits a group-scale audience.
- Option B: latest on the site, plus a link to the guide at each earlier tag in the source repo
  once it's public.
- Option C: versioned guide paths on the site (`/0.8/`, `/0.9/`), one vendor directory per
  version.

Recommendation: B. It costs one link, and it depends on the open-sourcing ruling already due
before Phase B.

## Not findings

- Concurrent editor and sync commits don't corrupt anything on their own (see D5).
- The one-way drift rule (line 103-105) is sound; a site that needs a token adds it in dubplate
  first.
- Keeping the site in its own repo removes any risk of cairn's publish flow touching dubplate's
  lanes.
