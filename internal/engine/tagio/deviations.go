package tagio

// Deviations documents what the twelve-subject corpus does not cover
// and every place this package deliberately differs from pinned
// mediafile 0.17.0's own behavior, so rung 15's launch gate reads a
// statement rather than inferring one from a passing suite. It carries
// no runtime value; its only purpose is to give this note an exported
// anchor `go doc` can print. rules_test.go is the evidence for every
// claim below; this note claims no coverage that file does not have.
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
// never by deriving a new corpus subject (ruling 1 declines that).
//
// # Deliberate divergences from the oracle
//
//  1. The U+007E field-name bound. mutagen's own is_valid_key stops one
//     code point short, at 0x7D; this package refuses only above
//     U+007E, RFC 9639 section 8.6's own bound (C7). A name that
//     mutagen would refuse and this package accepts is a name no real
//     tagger emits, so the corpus never exercises the gap; Task 3's
//     refusal tests and TestRulesU007EFieldNamePreservedThroughRead
//     here are this package's only evidence for it.
//
//  2. The date triple is composed as a unit. mediafile's own
//     MediaFile.update performs a read-modify-write over each date
//     component it is given, so setting only a track's month leaves
//     its existing year and day untouched. tagio's Write takes a whole
//     TagSet rather than a partial update, so it has no surface for
//     that partial case: serializeDate always composes DATE (or
//     ORIGINALDATE) from the year, month and day it is handed in one
//     call, treating a nil month or day as falsy the same as zero.
//     This is a deliberate narrowing of tagio's write contract, not a
//     bug matched to mediafile's read-modify-write; a caller wanting
//     a partial update reads the current TagSet first and writes the
//     merged result.
//
//  3. An empty list is indistinguishable from absent on read-back. A
//     list field holding no values writes no comment at all (C2), so
//     both the empty-list state and the absent state leave identical
//     bytes on disk, and both project to a nil slice on read
//     (TestRulesEmptyListRoundTrip). The ambiguity is on disk, in the
//     format itself, not a choice this package's read path makes.
//
//  4. The PADDING reservation is 8 KiB, against mutagen's 1 KiB
//     (rulings 5 and 6). Every write drops every existing PADDING
//     block, writable or not, and appends one fresh 8 KiB block, last,
//     regardless of what the container being rewritten carried. The
//     wider reservation buys headroom for repeated in-place retagging
//     (rung 9's retag verb) at the cost of eight times mutagen's own
//     default.
//
//  5. The vendor string is fresh. A container that already carries a
//     VORBIS_COMMENT block keeps its loaded vendor exactly (R9); a
//     container that carries none gets a vendor this package invents,
//     "dubplate " followed by the build's module version (or
//     "development" when the toolchain embeds none), never mutagen's
//     own "Lavf" or "mutagen" style string. A byte inspecting a
//     freshly created VORBIS_COMMENT block can tell tagio wrote it.
//
//  6. The sixteen keys the ledger cannot grade, and bpm alongside them
//     (see "What the corpus does not cover" above), are graded by
//     rules_test.go's hand-built rows rather than by any conformance
//     run against the oracle, because the corpus carries no non-empty
//     value for them to compare against.
var Deviations struct{}
