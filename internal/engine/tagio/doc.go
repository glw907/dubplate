// Package tagio reads and writes FLAC tags. It knows nothing about
// matching: no consumer's type appears in its signatures, it imports no
// package under internal/harness, and it reads no clock. Every path and
// mode this package touches reaches it as a parameter from its caller.
//
// It sits beside internal/engine/tag, which computes the applied field
// and touches no file (internal/engine/tag/doc.go:8-11).
//
// Read and Write sniff the file first and refuse one that is not FLAC;
// Sniff and SupportedFormats name the formats tagio handles, and FLAC
// is the only one. Parse is FLAC's own entry point and returns a FLAC
// Container.
//
// Read stays tag-only: it returns a TagSet and nothing else. Pictures
// live on Container, populated by Parse, because a picture is not one
// of the 69 writable keys the tag projection carries. A caller that
// wants both a file's tags and its pictures therefore parses the file
// twice, once through Read and once through Parse, rather than
// reaching for a second read entry point this package does not
// expose. The double parse is the accepted cost of keeping Read's
// contract narrow.
//
// # What the corpus does not cover
//
// Across the twelve corpus subjects, 53 of the 69 fields carry a
// non-empty value on at least one row. The other sixteen are empty,
// null, or the empty list on all twelve: acoustid_fingerprint,
// acoustid_id, comments, encoder, genres, grouping, initial_key,
// lyrics, r128_album_gain, r128_track_gain, remixers, rg_album_gain,
// rg_album_peak, rg_track_gain, rg_track_peak, and subtitle. They are
// exactly the fields a MusicBrainz match never supplies: fingerprints,
// gain analysis, and user-authored text. bpm is graded by the ledger
// (it is BPM=0 on all twelve rows) but keeps a test alongside them,
// because a zero that must survive a write is worth pinning.
//
// The corpus also never reaches six shapes no oracle response happens
// to carry: a zero year, an empty list, catalognum empty, barcode
// empty, an invalid UTF-8 sequence, and a comment field name at the
// U+007E bound. rules_test.go grades all six by hand-built fixture,
// never by deriving a new corpus subject. rules_test.go is the
// evidence for every claim in this section and the next; neither
// claims coverage that file does not have.
//
// # Deliberate divergences from mediafile
//
// This package differs from pinned mediafile 0.17.0's own behavior in
// these places:
//
//  1. The U+007E field-name bound. mutagen's own is_valid_key stops one
//     code point short, at 0x7D; this package refuses only above
//     U+007E, RFC 9639 section 8.6's own bound. A name that mutagen
//     would refuse and this package accepts is a name no real tagger
//     emits, so the corpus never exercises the gap; the container
//     refusal tests and TestRulesU007EFieldNamePreservedThroughRead are
//     this package's only evidence for it.
//
//  2. The date triple is composed as a unit. mediafile's own
//     MediaFile.update performs a read-modify-write over each date
//     component it is given (mediafile:365-378), so setting only a
//     track's month leaves its existing year and day untouched. Write
//     takes a whole TagSet rather than a partial update, so it has no
//     surface for that partial case: the date codec always composes
//     DATE (or ORIGINALDATE) from the year, month and day it is handed
//     in one call, treating a nil month or day as falsy the same as
//     zero. This is a deliberate narrowing of Write's contract, not a
//     bug matched to mediafile's read-modify-write; a caller wanting a
//     partial update reads the current TagSet first and writes the
//     merged result.
//
//  3. An empty list is indistinguishable from absent on read-back. A
//     list field holding no values writes no comment at all, so both
//     the empty-list state and the absent state leave identical bytes
//     on disk, and both project to a nil slice on read
//     (TestRulesEmptyListRoundTrip). The ambiguity is on disk, in the
//     format itself, not a choice this package's read path makes.
//
//  4. The PADDING reservation is 8 KiB, against mutagen's 1 KiB. Every
//     write drops every existing PADDING block, writable or not, and
//     appends one fresh 8 KiB block, last, regardless of what the
//     container being rewritten carried. The wider reservation buys
//     headroom for repeated in-place retagging at the cost of eight
//     times mutagen's own default.
//
//  5. The vendor string is fresh. A container that already carries a
//     VORBIS_COMMENT block keeps its loaded vendor exactly; a container
//     that carries none gets a vendor this package invents, "dubplate "
//     followed by the build's module version (or "development" when
//     the toolchain embeds none), never mutagen's own "Lavf" or
//     "mutagen" style string. A reader inspecting a freshly created
//     VORBIS_COMMENT block can tell tagio wrote it.
//
//  6. The sixteen keys the ledger cannot grade, and bpm alongside them
//     (see "What the corpus does not cover" above), are graded by
//     rules_test.go's hand-built rows rather than by any conformance
//     run against the oracle, because the corpus carries no non-empty
//     value for them to compare against.
//
//  7. A picture write discards incoming geometry. Width, height, colour
//     depth and indexed-colour count are always written as zero,
//     regardless of what the caller's Picture carried, because
//     mediafile's FLAC image write builds a fresh mutagen.flac.Picture
//     and sets only data, type, mime and desc
//     (mediafile/storage/flac.py:30-37), and Picture.__init__
//     initialises all four to 0 (mutagen/flac.py:577-586). A rewrite
//     therefore never round-trips real dimensions;
//     TestWriteGeometryAlwaysZero and TestEncodePictureZerosGeometry
//     are this package's evidence.
//
//  8. APNG discrimination is not reproduced. Pinned filetype tries APNG
//     before PNG and tells them apart by walking the chunk table for an
//     acTL chunk appearing before IDAT or IEND
//     (filetype/types/image.py:51-88); this package matches PNG's own
//     eight-byte signature and reports image/png for both. An animated
//     cover is not a case a group's library realistically holds, and a
//     chunk-table walk over caller-supplied bytes is new attack surface
//     for no benefit. This is a declared coverage gap: no test in this
//     package exercises an APNG blob.
//
//  9. A CR2-prefixed blob is refused rather than sniffed. Pinned
//     filetype's Tiff.match excludes CR at offset 8 from its own TIFF
//     test and routes that shape to Cr2, reporting image/x-canon-cr2
//     (filetype/types/image.py:193-199, filetype/types/__init__.py:15-32);
//     this package's sniffMediaType applies the same CR exclusion but
//     has no seventh row to route the exclusion to, so a CR2-prefixed
//     blob matches nothing and the write is refused. A cover
//     realistically never carries a camera raw file, so the refusal
//     costs no format a library needs.
//
//  10. Every declared block length is trusted rather than re-derived
//     from the file, for both VORBIS_COMMENT and PICTURE. Mutagen sets
//     _distrust_size = True on Picture (mutagen/flac.py:575) and on
//     VCFLACDict (:350), and FLAC.__read_metadata_block therefore
//     ignores the 24-bit block length and parses each block from the
//     file object using its own internal lengths instead (:712-724,
//     whose comment cites quodlibet/mutagen issue 106 for the
//     real-world shape a wrong block length takes). tagio trusts the
//     block header's declared length outright and bounds-checks a
//     block's own internal lengths against it, in vorbisField and
//     decodePicture alike. A library file whose block length disagrees
//     with its own content is one beets can tag and tagio refuses, and
//     the refusal reaches the intake gate as a rejection rather than a
//     hardened re-derivation.
package tagio
