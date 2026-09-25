# Photo curation spec review: mechanics and feasibility

Target: `docs/superpowers/specs/2026-09-25-photo-curation-design.md` at `67d4adf`. Lens: does each
mechanism behave as stated. Every claim below comes from a probe run on 2026-09-25 in a scratch
directory outside the repo, or from a quoted primary source. Nothing is from memory.

Tools present: `cwebp` 1.6.0 (libsharpyuv 0.4.2), ImageMagick 7.1.2-31, Node LTS via mise, `uv`,
Go 1.27.1, `curl`. Missing: Pillow, `vips`, `ajv` (installed `ajv-cli@5.0.0` into scratch for the
probes).

Summary: 0 blockers, 6 majors, 8 minors. The design is feasible. The majors are places where the
spec names a mechanism that does not exist yet or does not work as written. Each fold is small.

## Majors

### M1 (major): the legal basis given for resizing holds only for CC 4.0, and pre-4.0 attribution needs two fields the record lacks

Location: spec lines 111, 123-126, 203-204.

The spec says "Resizing and re-encoding is a technical modification under CC 4.0, not an
adaptation." For 4.0 this holds verbatim. BY-SA 4.0 §2(a)(4):

> The Licensor authorizes You to exercise the Licensed Rights in all media and formats whether now
> known or hereafter created, and to make technical modifications necessary to do so. [...] For
> purposes of this Public License, simply making modifications authorized by this Section 2(a)(4)
> never produces Adapted Material.

About half the pool is not 4.0. A tally of license mentions across the research files counts
74 CC BY-SA 4.0 and 21 CC BY 4.0, against 50 CC BY 2.0, 47 CC BY-SA 3.0, 25 CC BY-SA 2.0, 8 CC BY
3.0, and ported 3.0 NL and 3.0 DE. The 3.0 and 2.0 texts carry a narrower grant and no safe
harbor. BY-SA 3.0 §3 (2.0 is word-for-word the same):

> The above rights include the right to make such modifications as are technically necessary to
> exercise the rights in other media and formats.

There is no "never produces an Adaptation" sentence. Downscaling a 5760px original to 1600px is
not "technically necessary" to exercise any right. The 3.0 NL port (§3) matches the 3.0 grant:
"technische wijzigingen [...] die noodzakelijk zijn om de rechten met behulp van andere media,
dragers en formats uit te oefenen". The 3.0 "Collection" definition also requires the Work
"included in its entirety in unmodified form".

The outcome mostly survives on a different ground. The CC FAQ says that under US law "mechanical
reproduction of a work into a different format is unlikely to create a separate, new work", so a
resize is very likely not an adaptation in copyright terms. Even in the worst case, share-alike
binds only the resized image, which stays under its own BY-SA license. It never binds the MIT code
it sits beside. So the share-alike worry is fine. What the spec misses is the attribution duties,
because each record is the source a later CREDITS generator reads:

- 4.0 §3(a)(1)(B) requires that You "indicate if You modified the Licensed Material". The CC FAQ
  says "You must also indicate if you have modified the work — for example, if you have taken an
  excerpt, or cropped a photo". The login column crops at render time, and the stored file is
  resized. The spec treats "technical modification" as meaning no notice is owed.
- 3.0 §4(c)(ii) and 2.0 §4(c) require "the title of the Work if supplied". The CC FAQ adds: "For
  versions prior to 4.0, you must also provide the title of the work." The record's `title` is
  "A short caption title", which is dubplate's own caption. It is not the source's title. Nationaal
  Archief, Bundesarchiv, and most Commons files do supply a title.

Proposed fold:

- Replace the sentence at line 124 with the per-version basis: the 4.0 §2(a)(4) safe harbor, and
  for 3.0 and 2.0 the format-shift position (no new originality, per the CC FAQ). Under that
  position the image keeps its own license, and nothing widens past the image.
- Add a `source_title` field, verbatim from the source.
- State that CREDITS marks every entry "resized and converted to WebP", which costs nothing and
  meets 4.0 §3(a)(1)(B) whichever reading holds.
- Line 204 ("unmodified in content") then reads true.

### M2 (major): WebFetch cannot produce a verbatim license quote

Location: spec lines 117, 142-143, 150-151, 158-159.

The `verified` field is "The license text quoted from the source page", and the Opus reader
spot-checks quotes "against the live source pages". The obvious agent tool paraphrases. The
WebFetch tool definition in this session says it "Fetches a URL, converts the page to markdown, and
answers `prompt` against it using a small fast model." Whatever comes back as a "quote" is a small
model's rendition. The spot-check would use the same channel and could agree with a hallucinated
quote.

The pages themselves are easy to fetch raw. The probe results:

- `curl` with its default UA, with a browser UA, or with a descriptive UA gets 200 from the Commons
  file page, `api.php`, and `upload.wikimedia.org`.
- A `python-requests/2.32` UA gets 403 on all three. The Wikimedia User-Agent policy explains why:
  "User-Agent strings that begin with non-descriptive default values, such as python-requests/x,
  may also be blocked [...] Scripts should use an informative User-Agent string with contact
  information, or they may be blocked without notice."
- `index.php?title=File:…&action=raw` returns the license template verbatim, for example
  `{{cc-by-2.0}}`, plus a machine-readable review record for Flickr transfers:
  `{{FlickreviewR|status=passed|author=YardEdge|sourceurl=https://flickr.com/photos/29027546@N07/4546721487|archive=https://web.archive.org/…|reviewdate=2018-08-30 08:15:11|reviewlicense=cc-by-2.0|…}}`.
- A Flickr photo page returns 200 to curl and carries the license as a link
  (`creativecommons.org/licenses/by/2.0`).

Proposed fold: state the evidence channel.

- License evidence is fetched raw with `curl` under a descriptive UA naming dubplate and a contact
  address, never through WebFetch.
- On Commons, `verified` quotes the license template and any review template from `action=raw`.
- On Flickr, it quotes the license link from the page HTML.
- The spot-checker uses the same raw channel.

### M3 (major / OWNER FORK): the 1600px floor would drop close to half of the known candidates, and the spec does not say what happens below it

Location: spec lines 123, 161-162.

A Commons API size query over the 118 research files that have extensions resolved 113 of them:
62 have a long edge of 1600px or more, 51 fall under 1600, and 30 fall under 1000. The period
material is the small part. Examples are Bundesarchiv Bild 183 at 568×608, Tropenmuseum at
700×536, "Googoosh singing, 1969" at 1200×1107, and Havana Tropicana at 916×1455. The spec says an
agent "drops an image [...] too small for a 1600px display". It also says images are "stored as
WebP at a 1600px long edge", which leaves open whether a smaller image is upscaled. Upscaling adds
bytes and invents detail, and it is also a modification outside M1's format-shift reasoning.

Options:

- (a) A hard 1600px floor, as written. This means a thinner period record.
- (b) Store at native size below 1600, never upscale, and set a floor near 1000 to 1200px for the
  login column. Agents also look for the archive's own larger scan: Nationaal Archief and the
  Bundesarchiv often hold bigger files than the Commons copy.
- (c) Upscale to 1600. This is rejected, for the reasons above.

Recommendation: (b). It fits the spec's own scarcity finding for period photos, and "resize down
only" keeps M1's position clean.

### M4 (major): the cross-record rules are described as enforced, but nothing in the pass enforces them

Location: spec lines 36-44, 82-84, 129-130, 133-136.

JSON Schema validates one instance at a time. ajv cannot check anything across records. The spec
nevertheless states several cross-record behaviors as mechanisms:

- "The ID enforces the one-image-per-headliner rule by lookup."
- A landmark appears at most once.
- No region holds more than half the pool.
- The per-family coverage floor.
- "running counts" on the contact sheet and on the sorting page.
- A rejected record stops later rounds from finding the photo again.

Line 134 rules out Go, and no other tool is named. As written, each of these is a conductor doing
it by eye, and the headliner check silently skips every performer with no MBID.

Proposed fold: name one small workstation-only counting script. Node is already a workstation
dependency, and the script sits outside the `./...` gate like the Python derivation tools. It
reads every record and reports:

- the shares
- the coverage floor
- MBID duplicates, plus name duplicates for performers with no MBID
- landmark duplicates
- source-key duplicates (M5)

It runs after each round, and its output feeds the agents' briefs and the sorting page's seed
data.

### M5 (major): no dedup key and no slug ownership, with parallel agents writing records

Location: spec lines 92-104, 128-131, 149-152.

"A passed candidate keeps its record [...] so later rounds do not find it again" works only if a
later agent can recognize the same photo. Nothing names the key. The slug cannot be the key,
because each agent coins its own. The URL is fragile: the same Commons file shows up with spaces or
underscores, with or without percent-encoding, and (observed today) with `utm_*` parameters added
to `upload.wikimedia.org` URLs. The same photo also lives at two sources. For example, Commons
`Culture_Yard_(4546721487).jpg` is Flickr photo 4546721487, and the Flickr copy's bytes differ from
the Commons copy's, so `source_sha256` misses too.

"One file per record keeps parallel search agents from colliding" is also only half true: two
agents can coin the same slug for different photos (say `paradiso-1977`), and the second write
replaces the first without any error.

Proposed fold:

- Add a `source_id` field in canonical form (`commons:<pageid>`, `flickr:<photo id>`,
  `archive:<name>:<inventory number>`) and an `origin_id` for a transfer. Commons records the
  Flickr ID in its FlickreviewR template.
- Agents write candidates into per-agent scratch directories under provisional names.
- The conductor assigns the permanent slug at pick time, because the slug is the website path and
  should be coined once.
- Pin the slug pattern to `^[a-z0-9]+(-[a-z0-9]+)*$` (see M14 for why).

### M6 (major): the sorting page's image delivery is constrained in ways the spec does not account for

Location: spec lines 167-174.

External origins are not loadable in an artifact, so every candidate image has to be uploaded to
it. The Artifact tool's stated limits:

- Published `files`: "at most 255 entries and 64MB per version". A round-one sheet plus later
  rounds will pass 255 images.
- The asset store takes `file_paths` "up to 25 [...] in one call under one approval". That is at
  least 12 approvals per 300 images, which is Geoff's attended time.
- "a file outside the working directory [...] each go in a call of their own". The session
  scratchpad is outside the working directory. The spec's "untracked scratch storage" therefore
  decides between about 12 approvals and 300 or more. The candidates must sit in a gitignored
  directory inside the repo.
- The per-artifact asset quota is not published. The capability doc says only "Quota: per
  artifact (`usage`)". Unverified; check `assets.list().usage` after round one and prune passed
  candidates with `assets.delete`.

An in-repo scratch directory has its own mechanics. `scripts/check.sh` runs gitleaks with
`--no-git` over the tree, and `.gitleaks.toml` already allowlists untracked directories for this
reason (`.claude/worktrees/`, `.superpowers/`). A dot-prefixed directory, such as
`photos/.candidates/`, is also skipped by a `//go:embed photos` directory pattern (M14).

A grid of full 1600px WebPs costs about 45MB per load on a phone. Upload a small thumbnail
(roughly 600px) for the grid and the full file for the Space view, or generate one from the other
in the page.

The database side works as the spec states. The `db` capability stores JSON documents,
`ArtifactData` reads them back "as the person", `list` returns them a page at a time, and choices
persist across devices. Use shared documents keyed by `source_id`, not `data/users/<id>/`, which
the capability doc calls "private even from the owner".

Proposed fold: state these facts in the process section.

- Candidates live in a gitignored, dot-prefixed directory inside the repo, allowlisted in
  `.gitleaks.toml`.
- Thumbnails and full files go up in 25-file batches.
- Choices are stored as shared `db` documents keyed by `source_id`.

## Minors

### M7 (minor): the validation mechanism as written does not run under ajv

Location: spec lines 49-50, 98, 133.

Probes with `ajv-cli@5.0.0` (bundling ajv 8.20.0), in a scratch copy:

- A plain-data `vocab.json` referenced with `"$ref":"vocab.json#/shot"` fails:
  `schema photos/schema.json is invalid / error: strict mode: unknown keyword: "shot"`. A `$ref`
  must land on a schema, so the enums must live under `$defs` as `{"enum":[…]}`.
- A derivation map as a top-level key fails the same way:
  `strict mode: unknown keyword: "scene_family"`. The map has to be parked where strict mode
  allows it (for example, as a `const`), or kept in a separate file.
- `ajv validate -s photos/schema.json -r photos/vocab.json -d 'photos/*.json'` validates
  `schema.json` and `vocab.json` as records and reports both as `invalid` on every run.
- JSON Schema cannot turn a map's keys into an enum. The scene list is either written twice (map
  and enum) or `schema.json` is generated from `vocab.json`.

`ajv` is not installed on the workstation. `ajv-cli` 5.0.0 is the latest release, and its registry
entry was last modified 2023-04-28, which the adoption bar (maintenance) should weigh against ajv
used through a short Node script, a script M4 needs anyway.

Proposed fold:

- Shape `vocab.json` as schema `$defs` with the maps inside, or generate `schema.json` from it.
- Validate records by a pattern that excludes the two meta files (a subdirectory, or a name
  pattern).
- Consider running ajv as a library inside M4's script.

### M8 (minor): the example license identifier is not SPDX, and the license field is not closed

Location: spec line 114.

In the SPDX License List 3.29.0, `CC-BY-SA-3.0-NL` does not exist (only `-AT`, `-DE`, `-IGO`
exist), and `public-domain` is not an SPDX ID. `CC-BY-3.0-NL` does exist. The Free Art License is
`LAL-1.3`. Line 49 closes only the facets, so the license field stays free text and agents will
spell it inconsistently.

Proposed fold: add a closed license list to `vocab.json`. Use SPDX IDs where they exist and
`LicenseRef-`-prefixed IDs otherwise (for example `LicenseRef-CC-BY-SA-3.0-NL`,
`LicenseRef-public-domain`), each mapped to its legal-code URL. This also makes `license_url`
derived rather than typed.

### M9 (minor): the ~150KB estimate depends on an encoder setting the spec does not pin

Location: spec line 123.

The probe used 24 real candidate originals from the research, each with a long edge of at least
1600px, resized with `magick -resize 1600x1600>` and encoded with `cwebp -m 6`. Mean sizes:

| cwebp quality | Mean size |
|---|---|
| `-q 75` (default) | 130KB |
| `-q 80` | 164KB |
| `-q 85` | 212KB |

At `-q 85` the median is 180KB and the maximum 445KB. So 300 images come to about 39MB, 48MB, or
62MB depending on quality. The estimate holds at the default to `-q 80`. Several B&W scans are
stored as 3-channel sRGB.

Proposed fold: pin the encode command (tool version, `-q`, `-m`, resize filter, metadata policy
from M10) in the spec, so `sha256` values are reproducible.

### M10 (minor): cwebp strips embedded copyright notices by default

Location: spec lines 123-126.

`cwebp -longhelp` says: "`-metadata <string>` [...] Valid values: all, none (default), exif, icc,
xmp". Four of the 24 probe originals carry EXIF `Copyright` or `Artist`, for example
`CC-BY-SA_4.0|Thesupermat` and `Jonathan Botkin|@oddbotkin`. BY-SA 3.0 §4(c) requires You to "keep
intact all copyright notices for the Work", and BY-SA 4.0 §3(a)(1)(A)(ii) requires You to retain
"a copyright notice". Dropping ICC can also shift tone on some scans.

Proposed fold: encode with `-metadata exif,icc,xmp`, or state that CREDITS carries the notice
verbatim from the record.

### M11 (minor): parallel agents will trip the MusicBrainz rate limit

Location: spec lines 81-84, 149-150.

MusicBrainz documents the rate limit as follows:

> the rule is as follows: The rate at which your IP address is making requests is measured. If
> that rate is too high, all your requests will be declined (http 503) until the rate drops again.
> Currently that rate is (on average) 1 request per second.

It also says: "Each request sent to MusicBrainz needs to include a User-Agent header, with enough
information in the User-Agent for us [...] to contact the application maintainers."

Several search agents on one workstation share one IP. Wikidata `wbsearchentities` answered fine
(Dischord House returns no item, which is consistent with "where one exists").

Proposed fold: agents record performer names only. One serialized step at no more than 1 request
per second, with a descriptive UA, resolves MBIDs.

### M12 (minor): the rule that a flyer's license must come from its artist conflicts with the Flickr rule

Location: spec lines 141-143, 158-159.

"Rules out trusting a Flickr [...] license field alone" is ambiguous for a Flickr-original photo,
where the Flickr license field is the licensor's own statement. CC grants are irrevocable, so a
live-page spot-check can also flag a correct record when the Flickr owner later switched to
All Rights Reserved. That would trigger "re-verified whole" for no reason.

Proposed fold: for Flickr-origin files, the evidence is the license at grant time. That is the
Commons FlickreviewR record (M2), or an archive.org capture. A later mismatch on the live page is
logged, not treated as a wrong quote.

### M13 (minor): ISO 3166-1 does not cover the countries of many period photos

Location: spec lines 71-73.

`pycountry` returns `None` for `DD`, `SU`, `YU`, `CS`, and `XK` under ISO 3166-1. The historic
countries live in ISO 3166-3 under four-letter codes: DDDE German Democratic Republic, SUHH USSR,
YUCS Yugoslavia. `CS` is ambiguous even there: it is both CSHH Czechoslovakia and CSXX Serbia and
Montenegro. The Bundesarchiv Bild 183 candidates are East Berlin ADN photos.

Proposed fold (one clear answer): record the current ISO 3166-1 code of the place as it is today
(`DE` for East Berlin, `RS` for Belgrade). Add an optional free-text historical country for the
caption. The region map keys on current codes only.

### M14 (minor): go:embed constraints the consuming rung inherits from this pass's layout

Location: spec lines 92-101, 128-131.

`go doc embed` says:

> The patterns are interpreted relative to the package directory containing the source file.
> [...] Patterns may not contain '.' or '..' or empty path elements

and:

> If a pattern names a directory, all files in the subtree rooted at that directory are embedded
> (recursively), except that files with names beginning with '.' or '_' are excluded.

and:

> Patterns must not match files whose names include the special punctuation characters
> " * < > ? ` ' | / \ and :

Consequences:

- The root `photos/` can be embedded only by a package inside `photos/`, the same shape as
  `web/embed.go`.
- A directory pattern would also embed every `rejected` and `pack` JSON record, including the
  `screen` notes. The consumer should embed by pattern, or rejected records should live elsewhere.
- A slug such as `king-tubby's-studio` could not be embedded (the M5 pattern closes this).
- About 45MB lands in every test binary that imports the package, so keep it a leaf package.

This needs no change now beyond the slug pattern. Record it as a note for the consuming rung.

## What holds

- Commons and Flickr pages are fetchable and carry license evidence in raw form.
- The artifact database reads back per the tool contract.
- 150KB per image is right at cwebp's default quality.
- MBID and Wikidata lookups work with a descriptive UA.
- The share-alike conclusion is right: nothing reaches the MIT code.
