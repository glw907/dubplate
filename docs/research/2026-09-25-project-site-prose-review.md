# Prose and facts review: project site design, revision 2.1

Reviewer: prose-voice-reviewer (Opus 5.5), 2026-09-25, against commit 44f94c9. Register: Google
developer documentation style (`~/.claude/docs/voice/technical-doc-go.md`). tellgrader
(`--register docs`): zero findings, 4491 words, 266 sentences, cadence CV 1.32.

Verdict: reads as a plausible Google-style author with no machine-writing tells. One false
guarantee (B1) and eight mechanical ambiguities (W1-W8).

## Blocker

- B1 (lines 136-139): the credits check reads direct `go.mod` requirements and
  `web/package.json` `dependencies`, but `web/package.json` has no `dependencies` key; Svelte,
  SvelteKit, daisyUI, and Tailwind are `devDependencies` whose code ships in the embedded bundle,
  and indirect Go modules (pflag, modernc libc, x/crypto, uuid, go-humanize) link into the binary.
  "can't ship uncredited" is false. Narrow the claim, or widen the input to linked modules and
  bundled packages. Delete "The web leg has nothing to check until rung 7 adds `tus-js-client`."

## Warnings

- W1 (209-225, 304): "fourth leg" with legs one to three uncounted; the leak check is "the same
  leg" as the guide leg though it covers every exported file. Name four legs, one short paragraph
  each.
- W2 (158-162): "four kinds of file" omits the fonts with `FONTS.md` and OFL texts, and the
  dubplate-owned changelog (fold D8), whose path is never stated.
- W3 (55): open ruling 11 is about the dubplate repo, not `dubplate-site`; say the site repo starts
  private and opens only if the source opens (ruling 11 option b).
- W4 (352): `packaging` runs after `dubplate-organizer` (the active server build, which ends at
  launch); the after-launch work is the TUI and local binary per ROADMAP's header.
- W5 (257-258, 267-269, 214, 349): binary-name paragraph repeated and misplaced; the guide leg's
  "server binary's name" is undefined while the 0.8 name is open. Delete 267-269; the
  planned-names list carries the 0.8 name; line 349 says "0.8's `upgrade` command".
- W6 (141-154): the photo-credits paragraph runs about 215 words and says "no second list" twice;
  split after line 146.
- W7 (91, 96, 101, 116): page titles as bare subjects read as commands ("The Set up a server
  guide opens with...").
- W8 (199, 306-307, 318): the generator placeholder is site-side; line 318's dubplate lane
  declares the class in the manifest and the next sync replaces the site's placeholder.

## Suggestions

- S1 (176): dubplate's `--ink-accent` also serves success (identity spec line 103).
- S2 (33, 189): say ruling 3's release-tag pin binds from Phase B.
- S3 (188): a 0.x tag is a release tag by shape; Phase B names 0.8 explicitly.
- S4 (247): "no exported enumerator" (salvage has unexported `codeToReason`,
  `_salvage/internal/reject/reject.go:158`).
- S5 (131): the "studied for behavior, never copied" rule comes from dubplate's CLAUDE.md and the
  streamer table; the gonic row says nothing about copying.
- S6 (204-205): contrast frame; "the gap is filed as a trial finding for cairn and is never
  patched locally."
- S7 (277): "leave the documented operator path".
- S8 (345): fragment; "The ruling is due before Phase B registers the domain, and Geoff confirms
  the purchase then."
- S9 (341-386): rulings 11 vs 12-13 order options and recommendation differently; pick one.
- S10 (393): ambiguous "which"; split the sentence.
- S11 (160): cut the redundant "exported from the same commit" clause or say why it matters.
