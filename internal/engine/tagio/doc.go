// Package tagio reads and writes FLAC tags. It knows nothing about
// matching: no consumer's type appears in its signatures, it imports no
// package under internal/harness, and it reads no clock. Every path and
// mode this package touches reaches it as a parameter from its caller.
//
// It sits beside internal/engine/tag, which computes the applied field
// and touches no file.
//
// Read and Write sniff the file first and refuse one that is not FLAC;
// Sniff and SupportedFormats name the formats tagio handles, and FLAC
// is the only one. Parse is FLAC's own entry point and returns a FLAC
// Container, which carries the pictures Read's TagSet does not.
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
//     emits.
//
//  2. The date triple is composed as a unit. mediafile's own
//     MediaFile.update performs a read-modify-write over each date
//     component it is given (mediafile:365-378), so setting only a
//     track's month leaves its existing year and day untouched. Write
//     takes a whole TagSet rather than a partial update, so it has no
//     surface for that partial case: the date codec always composes
//     DATE (or ORIGINALDATE) from the year, month and day it is handed
//     in one call, treating a nil month or day as falsy the same as
//     zero. A caller wanting a partial update reads the current TagSet
//     first and writes the merged result.
//
//  3. An empty list is indistinguishable from absent on read-back. A
//     list field holding no values writes no comment at all, so both
//     states leave identical bytes on disk, and both project to a nil
//     slice on read. The ambiguity is in the format itself, not a
//     choice this package's read path makes.
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
//  6. A picture write discards incoming geometry. Width, height, colour
//     depth and indexed-colour count are always written as zero,
//     regardless of what the caller's Picture carried, because
//     mediafile's FLAC image write builds a fresh mutagen.flac.Picture
//     and sets only data, type, mime and desc
//     (mediafile/storage/flac.py:30-37), and Picture.__init__
//     initialises all four to 0 (mutagen/flac.py:577-586). A rewrite
//     therefore never round-trips real dimensions.
//
//  7. APNG discrimination is not reproduced. Pinned filetype tries APNG
//     before PNG and tells them apart by walking the chunk table for an
//     acTL chunk appearing before IDAT or IEND
//     (filetype/types/image.py:51-88); this package matches PNG's own
//     eight-byte signature and reports image/png for both. An animated
//     cover is not a case a group's library realistically holds, and a
//     chunk-table walk over caller-supplied bytes is new attack surface
//     for no benefit.
//
//  8. A CR2-prefixed blob is refused rather than sniffed. Pinned
//     filetype's Tiff.match excludes CR at offset 8 from its own TIFF
//     test and routes that shape to Cr2, reporting image/x-canon-cr2
//     (filetype/types/image.py:193-199, filetype/types/__init__.py:15-32);
//     this package's sniffMediaType applies the same CR exclusion but
//     has no seventh row to route the exclusion to, so a CR2-prefixed
//     blob matches nothing and the write is refused. A cover
//     realistically never carries a camera raw file, so the refusal
//     costs no format a library needs.
//
//  9. Every declared block length is trusted rather than re-derived
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
