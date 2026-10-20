# Container needs after rung 8: which audio containers an adopter actually brings

Written 2026-09-07 for the rung 8 plan, against build spec revision 2.1 ladder entry 8, the
streamer 2.0 design spec's F1 ruling and forward hooks 13 and 14, the corpus handoff in
`corpus/README.md`, and Geoff's rung 8 rulings of 2026-09-07. It answers one question: which
containers do the users of a shared, self-hosted music server need tagged and admitted, so the
follow-on passes after rung 8's FLAC writer are ordered by evidence.

Geoff's ruling (1) scopes rung 8 to FLAC with a per-format seam, and hands this document two
decisions to recommend. The first is the order of the follow-on passes. The second is whether any
of them precedes the 1.0 launch. Both recommendations sit in "Recommended pass order" and
"Does any container pass precede 1.0", and both are recommendations for Geoff rather than rulings.

## What this research can and cannot establish

No public, dated survey of self-hosted music library composition turned up. The searches covered
Navidrome, Jellyfin, gonic, Picard, and beets issue trackers and docs, self-hosting community
threads, and store and ripping-tool documentation. Nothing in that record measures what share of
files in a real library is FLAC, MP3, or M4A. Several secondary pages assert format shares with no
method behind them, and none is cited here.

The ranking below therefore rests on proxies, each named and dated:

1. What a mature self-hosted server declares it handles. Its extension list is a record of what
   its users brought it over years of issue reports.
2. What the stores a self-hoster buys from actually sell.
3. What ripping and download tools produce by default, since the default is what a
   non-specialist family member gets.
4. What the tagging libraries in this field support, which bounds what any server can promise.
5. Issue-title frequency across two large trackers, as a weak signal of which formats users hold
   and argue about.

Every claim below carries its source and the date the source was retrieved or published. A claim
that could not be verified against a primary source is marked **(unverified)** in place, not left
to the reader to guess.

## The ranked table

The likelihood column reads as: how likely is it that a stranger who adopts dubplate has files of
this container somewhere in the library they want to bring. It is not a claim about what share of
their files it is.

| Rank | Container (codecs) | Likelihood | Evidence |
|---|---|---|---|
| 1 | MP3 | Near-certain | Bandcamp's default download format is MP3: "The default download format is MP3, and this is probably what you want" ([Bandcamp Help, 2026-06-12](https://get.bandcamp.help/en/articles/15263285-which-audio-format-should-i-download)). Every server, tagger, and library in this study supports it. Highest or second-highest issue-title count in both trackers (see "The issue-title proxy"). Any library assembled before roughly 2010 is MP3 by default. |
| 2 | FLAC | Near-certain for this adopter class | Sold by Bandcamp and by Qobuz at every quality tier, including 16-bit/44.1 kHz ([Qobuz Help, retrieved 2026-09-07](https://help.qobuz.com/en/articles/10167-what-are-the-different-audio-formats-available-for-download)). The recommended target of every CD-ripping guide found **(secondary sources only)**. Highest issue-title count of any format in the Navidrome tracker (27). Navidrome's own MIME table lists `.flac` first among lossless ([`mime_types.yaml`, retrieved 2026-09-07](https://github.com/navidrome/navidrome/blob/master/resources/mime_types.yaml)). The dubplate estate is FLAC-only by config. |
| 3 | MP4 family: M4A (AAC) and M4A (ALAC) | Likely wherever an Apple device has touched the library | Apple's Music app converts CD imports to AAC by default: "The default encoding format is AAC" ([Apple Support, retrieved 2026-09-07](https://support.apple.com/guide/music/import-songs-from-cds-mus2935/mac)), with AIFF, Apple Lossless, MP3, and WAV as the alternatives **(alternatives list is search-derived)**. iTunes Store purchases since 2009 are DRM-free AAC in `.m4a` **(unverified: secondary sources only)**. iOS holds roughly 58% of the US mobile market **(unverified: StatCounter figure reported by secondary sources, 2026)**. Navidrome maps `.m4a`, `.m4b`, and `.alac` to `audio/mp4`. Combined issue-title count for m4a, aac, and alac is 18 in Navidrome. |
| 4 | Ogg Vorbis | Occasional | Offered by Bandcamp ([Bandcamp Help, 2026-06-12](https://get.bandcamp.help/en/articles/15263234-in-which-formats-can-i-download-my-purchases)). Supported by mediafile, Picard, Navidrome, and gonic. Issue-title counts of 5 (Navidrome) and 6 (beets). Arrives mostly from open-source desktop rips and older Linux libraries **(unverified: inference, no source measures origin)**. |
| 5 | Opus | Occasional and rising | No store found sells Opus downloads. It arrives as a transcode or extraction output: `yt-dlp` lists `opus` among its `--audio-format` targets and YouTube serves Opus natively **(search-derived, [yt-dlp manual](https://man.archlinux.org/man/extra/yt-dlp/yt-dlp.1.en))**, and Symfonium users discuss server-side FLAC-to-Opus transcoding ([Symfonium support, retrieved 2026-09-07](https://support.symfonium.app/t/symfonium-transcoding-flac-to-opus/10075)). Issue-title counts of 8 and 8, above Ogg Vorbis in both trackers despite no store selling it. |
| 6 | WAV | Occasional, and worst-tagged | Sold by Bandcamp and Qobuz, both aimed at CD burning. Has no standard tagging format: "the Waveform Audio File Format (WAV) does not have a standard format for tagging, so Picard uses the ID3v2.3 standard, which may not be supported by all software" ([LWN, Joe Brockmeier, 2026-04-14](https://lwn.net/Articles/1066384/)). Picard adds optional RIFF INFO chunks for compatibility ([Picard docs v3.0, retrieved 2026-09-07](https://picard-docs.musicbrainz.org/en/v3.0/faq/faq_file_formats.html)). Common in DJ and production material **(unverified: no source measures this)**. |
| 7 | AIFF | Occasional, Apple-side | Sold by Bandcamp and Qobuz. Offered as a Music app import encoder. Tagged through an ID3 chunk, which mediafile reaches through mutagen's `AIFF` type (`mediafile/__init__.py:189`, mediafile 0.17.0 in `tools/venv`). Issue-title counts of 3 and 5. |
| 8 | WMA (ASF) | Occasional, legacy, shrinking | Windows Media Player's default rip format was WMA at 128 kbps **(unverified: secondary guides only, [Computer Hope](https://www.computerhope.com/issues/ch000093.htm))**, which puts it in any library assembled on Windows before the user changed the setting. Supported by mediafile, Picard, and Navidrome. Issue-title counts of 6 (Navidrome) and 1 (beets), and the gap is consistent with a format people stream but stopped acquiring. |
| 9 | WavPack, Monkey's Audio, Musepack, TAK, OptimFROG | Rare | Navidrome's lossless list carries `.ape`, `.wv`, `.wvp`, and `.tak`; Picard lists all of them plus OptimFROG. mediafile supports `ape`, `wv`, and `mpc`. Issue-title counts are low and lopsided: wavpack 0 (Navidrome) and 5 (beets), ape 1 and 1, musepack 2 and 0. The beets skew fits archive-era collections being imported once. |
| 10 | DSD (DSF, DFF) | Rare, audiophile | Qobuz sells DSD64 through DSD512. Navidrome lists `.dsf` as lossless and mediafile supports `dsf` through a mutagen type that carries ID3 (`mediafile/__init__.py:191`). Symfonium transcodes DSD rather than playing it directly ([Symfonium docs, retrieved 2026-09-07](https://docs.symfonium.app/wiki/other/transcoding-engine/)). Issue-title counts of 5 and 3. |

Two structural facts collapse the table further, and they matter more than the ranks.

**Rank 3 is one writer, not two.** AAC and ALAC are codecs inside the same MP4 container, tagged
through the same iTunes `ilst` atoms. Navidrome maps `.m4a`, `.m4b`, and `.alac` to a single MIME
type. One MP4 pass buys purchased iTunes music, Apple Lossless rips, and the `.m4a` that `yt-dlp`
produces from YouTube.

**Ranks 4 and 5 share a tag standard and not a container.** Ogg Vorbis and Opus both carry Vorbis
comments, so their field mapping is the FLAC mapping. Their framing is different: an Ogg comment
header is a packet inside Ogg pages with a CRC to recompute, and Opus prefixes it with the eight
bytes `OpusTags` ([RFC 7845, 2016-04](https://www.rfc-editor.org/rfc/rfc7845)). The mapping work is
free after rung 8; the container work is not.

## The issue-title proxy

Retrieved 2026-09-07 through the GitHub search API, counting issues and pull requests whose title
contains the term: `search/issues?q=repo:<repo>+<term>+in:title`.

| Term | navidrome/navidrome | beetbox/beets |
|---|---|---|
| flac | 27 | 22 |
| mp3 | 15 | 18 |
| m4a | 11 | 4 |
| opus | 8 | 8 |
| aac | 7 | 6 |
| wma | 6 | 1 |
| ogg | 5 | 6 |
| wav | 5 | 7 |
| dsf | 5 | 3 |
| aiff | 3 | 5 |
| musepack | 2 | 0 |
| ape | 1 | 1 |
| dsd | 1 | 3 |
| alac | 0 | 5 |
| wavpack | 0 | 5 |

Read this as friction, not prevalence. A format that decodes cleanly everywhere generates fewer
issues per file than one that needs transcoding, which biases the count toward lossless and toward
anything a browser refuses to play. The counts also mix questions, bugs, and feature requests. The
proxy is used here only to separate the top three from the tail, which it does consistently across
two trackers with different user populations.

## What each container uses for tags

The mapping column is Picard's, whose tag-mapping appendix covers exactly six tag standards:
"Internal Name", "ID3v2", "Vorbis", "APEv2", "iTunes MP4", "ASF/Windows Media", and "RIFF INFO"
([Picard docs v3.0, retrieved 2026-09-07](https://picard-docs.musicbrainz.org/en/v3.0/appendices/tag_mapping.html)).
The 2.0 spec already adopts those tables as prior art, so the follow-on passes inherit a mapping
rather than deriving one.

| Container | Tag standard | Spec | Cover art | Notes for a writer |
|---|---|---|---|---|
| FLAC | Vorbis comments in a `VORBIS_COMMENT` metadata block | [Xiph VorbisComment](https://wiki.xiph.org/VorbisComment), [FLAC format](https://xiph.org/flac/format.html) | Native `PICTURE` block | The rung 8 scope. The repo already parses both blocks in `internal/engine/shadow/flac.go`. |
| Ogg Vorbis, Ogg Opus | Vorbis comments inside an Ogg packet | [RFC 7845](https://www.rfc-editor.org/rfc/rfc7845) for Opus | Base64 `METADATA_BLOCK_PICTURE` in the comment | Same field mapping as FLAC, different framing. Page CRCs make an in-place rewrite harder than FLAC's block chain. |
| MP3 | ID3v2 | [id3.org, ID3v2.4.0, informal standard dated 2000-11-01](https://id3.org/id3v2.4.0-structure) and [frames](https://id3.org/id3v2.4.0-frames) | `APIC` frame | The frame set is large. Text encodings, unsynchronization, and `TXXX`/`UFID`/`COMM` descriptors carry most of the mapping. |
| MP4: M4A, M4B, ALAC | iTunes `ilst` atoms under `moov/udta/meta` | Picard's "iTunes MP4" column | `covr` atom | Atom sizes are length-prefixed and nested, so a tag rewrite resizes parents. |
| AIFF | ID3v2 in an `ID3 ` chunk | Picard "ID3v2" column | `APIC` frame | mediafile reaches it through mutagen's `AIFF` type. |
| WAV | ID3v2, optionally plus RIFF INFO | Picard "RIFF INFO" column | `APIC` frame | No standard exists; Picard writes ID3v2.3 and adds RIFF INFO for readers that reject ID3 in WAV ([LWN, 2026-04-14](https://lwn.net/Articles/1066384/)). |
| DSF | ID3v2 chunk | Picard "ID3v2" column | `APIC` frame | mediafile routes `dsf` through mutagen's `DSF` type. |
| WavPack, Monkey's Audio, Musepack | APEv2 | Picard "APEv2" column | Binary APEv2 item | One tag standard covers all three. |
| WMA | ASF attributes | Picard "ASF/Windows Media" column | `WM/Picture` | mediafile has a dedicated ASF storage family (`mediafile/storage/afs.py`). |

The whole field reduces to five tag standards behind fourteen containers. mediafile 0.17.0, the
pinned oracle, is organized the same way: `mediafile/storage/` holds exactly `flac.py`, `mp3.py`,
`mp4.py`, `vorbis.py`, `ape.py`, and `afs.py`. A tagio seam keyed by container that dispatches to a
tag-standard implementation matches both the domain and the oracle.

## Go libraries under permissive licenses

Scored on CLAUDE.md's four criteria. Maintenance and adoption figures were retrieved 2026-09-07
from the GitHub API and pkg.go.dev. Bus factor is omitted for the reason the 2.0 spec gives: it is
1 across this entire field and discriminates nothing.

| Library | Role | License | Quality | Maintenance | Adoption | Verdict |
|---|---|---|---|---|---|---|
| `github.com/dhowden/tag` | Read MP3, MP4, Ogg, FLAC, DSF | BSD-2-Clause ([LICENSE](https://github.com/dhowden/tag/blob/master/LICENSE)) | Read-only by design. Open bug #116 (2026-09-04) reports `readPictureBlock` passing an attacker-controlled 4-byte length straight to `make([]byte, dataLen)`, so a 49-byte crafted FLAC commits a ~2 GB allocation. Fix PR #115 has been open since 2026-08-18. | No semantic-version tag has ever been cut; pkg.go.dev serves `v0.0.0-20240417053706-3d75831295e8` (2024-04-17). Last push 2024-07-30. 32 open issues, oldest #7 from 2015-05-20, including #15 "Write Tag Support ?" from 2015-08-25. | 351 importers, 654 stars. The most-adopted tag reader in Go. | **Decline** for the intake read path. dubplate parses uploaded bytes from people it does not fully trust, and the one open bug is precisely that class, with an unmerged fix and an absent maintainer. |
| `github.com/bogem/id3v2/v2` | Read and write ID3v2.3 and v2.4 | MIT ([LICENSE](https://github.com/n10v/id3v2/blob/main/LICENSE)) | The only Go library that writes ID3 with real adoption. Import path is now a redirect: the repo was renamed to `n10v/id3v2`. | Last release v2.1.4 on 2023-02-09; last code commit the same day; a README edit on 2023-05-09. 24 open issues, oldest from 2017-04-30. No fork with adoption. | 72 importers, 370 stars. | **Conditional adopt** for the MP3 pass, pinned, with a hand-roll fallback. See "The two picks F1 made provisionally". |
| `github.com/abema/go-mp4` | MP4 box parse and build, including `ilst` | MIT ([LICENSE](https://github.com/abema/go-mp4/blob/master/LICENSE)) | Low-level and honest about it: "parse or build any MP4 boxes(atoms) directly". Carries `Ilst`, `Meta`, and `IsIlstMetaBoxType`, so the iTunes atoms are typed rather than raw. | v1.7.2 published 2026-09-04, the day this was retrieved. 6 open issues, oldest 2020-07-17. | 72 importers, 547 stars. | **Adopt** as the MP4 pass's box layer. dubplate writes the field-to-atom mapping itself; no Go library ships that mapping. |
| `github.com/Eyevinn/mp4ff` | MP4 boxes for streaming | MIT ([LICENSE](https://github.com/Eyevinn/mp4ff/blob/master/LICENSE)) | Well maintained, wrong shape: "focused on fragmented files as used for streaming in MPEG-DASH, MSS and HLS fMP4". No `ilst` editing documented. | v0.56.0 on 2026-08-22. 4 open issues. | 655 stars; pkg.go.dev reports 0 importers for the root package. | **Decline.** Solves the DASH problem, not the tagging problem. |
| `github.com/gcottom/audiometa` | All-in-one read and write for MP3, MP4, FLAC, Ogg | MIT ([LICENSE.md](https://github.com/gcottom/audiometa/blob/main/LICENSE.md)) | Shape is exactly what rung 8 wants. Its sibling modules `gcottom/mp4meta` and `gcottom/flacmeta` carry no license GitHub recognizes (`NOASSERTION`), and audiometa depends on that family. | v3.0.4 on 2024-12-29, last push the same day. | 11 stars; pkg.go.dev reports 0 importers. | **Decline.** Adoption is effectively zero, and a permissive top-level license over unlicensed dependencies fails the license criterion. |
| `github.com/mewkiz/flac` | FLAC frame decode and encode | Unlicense ([LICENSE](https://github.com/mewkiz/flac/blob/master/LICENSE)) | Idiomatic package split, as the 2.0 spec records. Metadata-block writing is not documented. | v1.0.14 on 2026-08-19. 9 open issues, oldest 2018-04-16. | 96 importers, 361 stars. | **Unchanged.** Adopt for rung 7's integrity check as already ruled. It is not the rung 8 writer. |
| `github.com/go-flac/go-flac/v2` | Split and reassemble FLAC metadata blocks | Apache-2.0 ([LICENSE](https://github.com/go-flac/go-flac/blob/master/LICENSE)) | Does exactly the rung 8 job: parse the block chain, edit `VorbisComment` and `Picture`, `Save()` while preserving audio frames. Siblings `flacvorbis` and `flacpicture` carry the field layer. | v2.0.4 on 2025-08-06; siblings last pushed 2024-08 and 2024-10. 0 open issues. | 8 importers, 44 stars. | **Decline as a dependency, adopt as a design reference.** The repo already owns the reader half in `shadow/flac.go`, F1 rules for extending it, and 8 importers does not justify a dependency for the remaining write half. |
| `github.com/ambeloe/oggv/vorbiscomment` | Read and write Ogg Vorbis comments | Not evaluated | Not evaluated | Not evaluated | Effectively unused | **Decline.** If an Ogg pass ever happens, hand-roll against RFC 7845. |
| `go.senan.xyz/taglib` | Tag read and write through wazero | LGPL-2.1 | n/a | n/a | 11 importers | **Already declined** by F1, and retired as a spike by ruling (2). Recorded here only so the table is complete. |

The field has one shape worth stating plainly. Go has a well-adopted tag **reader** with a security
bug and no maintainer, a lightly-maintained ID3 **writer** with real adoption, an actively
maintained MP4 box library that stops short of the tag mapping, and nothing at all for Ogg. No
library in Go covers reading and writing across containers at the quality bar this repo holds. That
is the same conclusion F1 reached from the license side, arrived at independently from the
maintenance and adoption side.

## The two picks F1 made provisionally

F1 named `dhowden/tag` for reading and `bogem/id3v2` for MP3 writing, with the scoring left to a
later rung. This is that scoring, and it splits.

**`dhowden/tag` should be declined.** Issue #116, opened 2026-09-04, is an unbounded allocation
from an attacker-controlled length field in the FLAC picture path, reproducible from 49 bytes. A
fix PR has sat unmerged since 2026-08-18, against a repo with no commit since 2024-07-30. dubplate's
intake reads bytes uploaded over tus by family members and their friends, which is exactly the
threat model that bug describes. The library is also read-only, so adopting it never removes the
need for per-container writers. Recommendation: sniff container type in-repo from magic bytes,
which the FLAC reader already does with `flacSignature`, and let each container's own reader serve
its pass.

**`bogem/id3v2` should be adopted anyway, pinned, when the MP3 pass runs.** Its maintenance record
is the worst in the table for a library still worth using: no code commit since 2023-02-09, and
open issues reaching back to 2017. The argument for adopting is that the specification is frozen
too. ID3v2.4.0 is an informal standard dated 2000-11-01, and nothing has superseded it in
twenty-five years. A frozen library tracking a frozen format ages differently from a frozen library
tracking a moving one. The library reads and writes both v2.3 and v2.4, has 72 importers, and
carries an MIT license with a clean provenance chain through the rename.

The MP3 pass should hold a stated fallback rather than a hope. If the pass finds a defect in
writing any field the 69-key writable set needs, it hand-rolls the frames it needs against the
id3.org specification instead of forking. The frames dubplate writes are a bounded subset, and the
pass will already own a conformance instrument that proves the subset.

## The oracle constraint, which is specific to this repo

Adding a container costs more here than the Go writer, and the extra cost is invisible from the
public record.

The conformance target for rung 8 is `corpus/analysis/file_tags.json`, twelve subjects over the
69-key writable set, all FLAC, derived through `internal/harness/oracle/derive_file_tags.py`
against the vendored `template.flac`. A container the corpus has never derived has no conformance
target. Producing one is a `make derive` run under the pinned venv, which by the standing policy is
main-tree work rather than worktree work (reverse-edge cell (8,3),
`docs/research/2026-09-04-reverse-edge-completion.md`). Each follow-on container pass
therefore carries a corpus derivation task and a scheduling constraint, not only a Go task.

Two facts bound that cost, and both are favorable.

The oracle can grade every container in the ranked table. mediafile 0.17.0's `TYPES` map covers
`mp3`, `aac`, `alac`, `ogg`, `opus`, `flac`, `ape`, `wv`, `mpc`, `asf`, `aiff`, `dsf`, and `wav`
(`mediafile/constants.py:4`, in `tools/venv`). No container in the table falls outside what pinned
beets can write and read back, so the proof shape rung 8 establishes transfers unchanged.

The template-fixture route transfers too. `gen_template_flac.py` authored the 42-byte silent FLAC
through mutagen alone, and mutagen writes every container in that list. A per-container template
generator is a sibling script, not a new method.

## Recommended pass order

1. **Rung 8, FLAC.** Already ruled. Extends the repo's own reader, has its conformance target
   already derived, and is the estate's only admitted format.
2. **MP3, first follow-on pass.** It is the container a stranger is most likely to hold, and the
   only one with an adoptable Go writer. Its tag standard, ID3v2, is the one that also unlocks
   AIFF, WAV, and DSF later, since all three carry ID3 frames. That makes it the highest-leverage
   second pass regardless of how the remaining ranks move.
3. **MP4, second follow-on pass, covering AAC and ALAC together.** One container, one atom
   mapping, two codecs, and it captures both iTunes purchases and Apple Lossless rips. It is the
   most expensive of the three because no Go library ships the field-to-atom mapping, so the pass
   writes that layer on top of `abema/go-mp4`.
4. **Stop there, and reject the rest at intake with contract wording.** Ogg Vorbis and Opus are the
   only serious candidates for a fourth pass. Their field mapping is free after rung 8 and their
   framing is not. Schedule one only if an adopter asks, which the reject copy makes visible.

Everything at rank 6 and below stays rejected. WAV and AIFF have no tag standard worth committing
to for a server whose product is well-organized metadata, and admitting them means choosing between
ID3-in-RIFF and RIFF INFO on the user's behalf. WMA, APEv2 formats, and DSD are rare enough that a
named reject beats a half-supported write path.

## Does any container pass precede 1.0

**Recommendation: no. Launch FLAC-only, and run the MP3 pass immediately after cutover.**

The reasoning, and the cost, both stated.

Nothing about the launch is blocked. Format acceptance is config by Geoff's 2026-09-03 ruling, the
estate is FLAC-only, and the intake gate is bounded by rung 8's support, so the 1.0 that ships is
the 1.0 the family needs. Adding two container passes to the critical path delays the cutover for a
capability no estate user consumes.

The cost is real and belongs in the record. dubplate's product goal is ease of setup by a stranger,
and MP3 is the container that stranger is most likely to hold. A FLAC-only 1.0 is useful to
lossless-only adopters and turns everyone else away at the gate. That is the price of shipping
sooner, and it is paid back by the first follow-on pass rather than carried indefinitely.

Three things must be inside launch scope so the follow-on passes stay additive. All three are
cheap, and skipping any of them converts a small pass into a rewrite.

1. **A writer seam keyed by container, defined at rung 8 with FLAC as its only implementation.**
   Ruling (1) already requires the seam. This document adds the shape: dispatch by container to a
   tag-standard implementation, since five standards cover fourteen containers.
2. **Rung 7's allowlist reads the supported set from the seam.** The build spec states the
   allowlist is bounded by rung 8's support. Reading that bound from the registry rather than
   restating it as a config default is what makes a later container a registration.
3. **Per-format reject wording in the string authority at rung 7.** The reject copy is contract.
   A message naming the format and saying the server cannot tag it yet is honest at 1.0 and stays
   correct as formats land.

### The conflict this recommendation resolves

The 2.0 design spec's F1 ruling reads "FLAC, MP3, and MP4/AAC at launch". Geoff's ruling (1) of
2026-09-07 scopes rung 8 to FLAC and hands the launch question to this document. The two are
consistent about scope and inconsistent about timing.

Recommended erratum wording, for the dated erratum that ruling (2) already schedules at rung 8's
close:

> F1's "at launch" is superseded. FLAC ships at 1.0. MP3 and MP4/AAC each land as their own
> post-launch pass, in that order, per
> `research/2026-09-07-rung-8-container-needs.md`. F1's other terms stand: pure-Go per-format
> readers and writers under permissive licenses, extending the repo's own FLAC reader, and an
> unusual container rejected at intake rather than mistagged.

## Does lossy admission matter for the estate

Two answers, because the question has two halves.

**For the estate, no, on the current record.** The family's stated inflow is CD rips and lossless
purchases, the config is FLAC-only, and rung 7's gate rejects everything else with contract
wording. Nothing in this research argues for changing that.

One caveat is worth Geoff's attention rather than an assumption. A family member with an Apple
library brings `.m4a`, and a friend who grabbed something with `yt-dlp` brings `.opus` or `.m4a`.
Whether that has ever happened in this family is **(unverified)**; no source consulted here can
know it. If it has, the estate answer is still a reject, and the only question is whether the
reject copy should say "convert it and re-upload" or "this format is coming". That is a copy
decision at rung 7, not a policy change.

**For other adopters, yes, and it is the single largest determinant of who can use dubplate.** A
stranger with an MP3 library cannot use a FLAC-only server at all. This is the whole argument for
the MP3 pass being first and immediate rather than eventual.

## Forward hooks 13 and 14

**Hook 14, the ID3 version ruling, is corroborated.** Geoff's ruling (4) records ID3v2.4 for when
MP3 writing lands. Two independent checks agree. The pinned oracle writes v2.4 by default:
"By default, MP3 files are saved with ID3v2.4 tags" (`mediafile/__init__.py:154`), gated by an
`id3v23` flag that beets leaves off (`config_default.yaml:105`, `id3v23: no`), and the estate's
`config/beets-config.yaml` does not override it. Jellyfin tells its users the same thing: "Files
with ID3v1 tags have a length limit of 30 bytes for most fields. Anything longer will be truncated.
Please upgrade to ID3v2.4"
([Jellyfin music docs, retrieved 2026-09-07](https://jellyfin.org/docs/general/server/media/music/)).
Writing v2.4 keeps the Go writer and the beets oracle byte-comparable, which the corpus proof shape
depends on.

**Hook 13, sort tags at admit, has a per-container name list.** For FLAC and Ogg the keys are the
Vorbis comments `ALBUMARTISTSORT` and `ARTISTSORT`. For MP3 they are the ID3v2 frames `TSO2` and
`TSOP`, and for MP4 the `soaa` and `soar` atoms, per Picard's mapping appendix. Rung 8 writes the
Vorbis pair. Each follow-on pass writes its own equivalents, which is one more reason the seam
dispatches by tag standard rather than by field.

## Open items for Geoff

Only product forks are listed, per the standing rule. Everything else in this document is a
recommendation the conductor rules.

1. **Does the 1.0 ship FLAC-only?** The recommendation is yes, with the MP3 pass scheduled
   immediately after cutover. It changes what a stranger sees on day one, so it is Geoff's.
2. **Has the family ever wanted to upload something that is not FLAC?** A yes changes the rung 7
   reject copy, and possibly the MP3 pass's position relative to the cutover. Only Geoff knows.

## Sources

Retrieved or published on the dates given. Primary sources are listed first within each group.

**Servers and clients**
- Navidrome `resources/mime_types.yaml`, retrieved 2026-09-07: https://github.com/navidrome/navidrome/blob/master/resources/mime_types.yaml
- Navidrome README ("Streams virtually any audio format available"), retrieved 2026-09-07: https://github.com/navidrome/navidrome
- Navidrome discussion #3654, mixed-format libraries, January 2025: https://github.com/navidrome/navidrome/discussions/3654
- gonic README ("using taglib - supports mp3, opus, flac, ape, m4a, wav, etc."), retrieved 2026-09-07: https://github.com/sentriz/gonic
- Jellyfin music documentation, retrieved 2026-09-07: https://jellyfin.org/docs/general/server/media/music/
- Symfonium transcoding engine documentation, retrieved 2026-09-07: https://docs.symfonium.app/wiki/other/transcoding-engine/
- Symfonium support thread on FLAC-to-Opus transcoding, retrieved 2026-09-07: https://support.symfonium.app/t/symfonium-transcoding-flac-to-opus/10075

**Taggers and tag standards**
- MusicBrainz Picard file formats, v3.0 docs, retrieved 2026-09-07: https://picard-docs.musicbrainz.org/en/v3.0/faq/faq_file_formats.html
- MusicBrainz Picard tag mapping appendix, v3.0 docs, retrieved 2026-09-07: https://picard-docs.musicbrainz.org/en/v3.0/appendices/tag_mapping.html
- "Tagging music with MusicBrainz Picard", Joe Brockmeier, LWN, 2026-04-14: https://lwn.net/Articles/1066384/
- ID3v2.4.0 structure, informal standard dated 2000-11-01: https://id3.org/id3v2.4.0-structure
- ID3v2.4.0 native frames, informal standard dated 2000-11-01: https://id3.org/id3v2.4.0-frames
- Xiph VorbisComment specification, retrieved 2026-09-07: https://wiki.xiph.org/VorbisComment
- RFC 7845, Ogg Encapsulation for the Opus Audio Codec, April 2016: https://www.rfc-editor.org/rfc/rfc7845

**Stores and tools**
- Bandcamp Help, "In which formats can I download my purchases?", 2026-06-12: https://get.bandcamp.help/en/articles/15263234-in-which-formats-can-i-download-my-purchases
- Bandcamp Help, "Which audio format should I download?", 2026-06-12: https://get.bandcamp.help/en/articles/15263285-which-audio-format-should-i-download
- Qobuz Help, download formats, retrieved 2026-09-07: https://help.qobuz.com/en/articles/10167-what-are-the-different-audio-formats-available-for-download
- Apple Support, "Import songs from CDs into Music on Mac", retrieved 2026-09-07: https://support.apple.com/guide/music/import-songs-from-cds-mus2935/mac
- yt-dlp manual, audio extraction formats, retrieved 2026-09-07 **(search-derived)**: https://man.archlinux.org/man/extra/yt-dlp/yt-dlp.1.en
- Computer Hope, Windows Media Player rip format **(secondary, unverified)**: https://www.computerhope.com/issues/ch000093.htm
- RIAA 2025 year-end revenue, digital downloads at $221.8 million against $11.54 billion total, a 5.9% year-over-year decline **(reported by Billboard and Music Business Worldwide citing the RIAA report; the RIAA PDF was not fetched)**: https://www.riaa.com/reports/2025-year-end-music-industry-revenue-report-riaa/

**Go libraries**
- pkg.go.dev and the GitHub API, all figures retrieved 2026-09-07. Per-library license URLs appear in the library table above.
- `github.com/dhowden/tag` issue #116, 2026-09-04: https://github.com/dhowden/tag/issues/116
- `github.com/dhowden/tag` pull request #115, 2026-08-18: https://github.com/dhowden/tag/pull/115

**In-repo, read-only**
- `corpus/README.md`, "Handoff to rung 8", from line 2354.
- `docs/specs/2026-09-02-greenfield-server-build.md`, ladder entries 7 and 8.
- `docs/specs/2026-09-07-streamer-2-0-design.md`, forward hooks 13 and 14, ruling F1, prior-art table.
- `docs/research/2026-09-04-reverse-edge-completion.md`, cells (8,3) and (8,4).
- `internal/engine/shadow/flac.go`, the provisional FLAC reader rung 8 absorbs.
- `tools/venv/.venv/.../mediafile/constants.py:4`, `mediafile/__init__.py:154`, `:189`, `:191`, `mediafile/storage/`, mediafile 0.17.0.
- `tools/venv/.venv/.../beets/config_default.yaml:105`.
