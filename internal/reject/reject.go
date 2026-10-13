// Package reject holds the reasons intake refuses or flags a submission, one
// enumerable set whose wording the uploader page, the API, and the project
// site's reject page share verbatim. The nine adopted reasons carry
// passb-copy-final.md section 4's text byte for byte; the rest follow its rules.
package reject

import (
	"fmt"
	"slices"
	"strings"
)

// Code is a reason's stable wire identifier: lowercase and hyphenated.
type Code string

// The codes callers name. A blocking reason refuses a submission or a file;
// a non-blocking one is a warning shown beside a verdict.
const (
	BadExtension           Code = "bad-extension"
	Symlink                Code = "symlink"
	Corrupt                Code = "corrupt"
	BelowCDQuality         Code = "below-cd-quality"
	TooLarge               Code = "too-large"
	UnpackFailed           Code = "unpack-failed"
	UnsafePath             Code = "unsafe-path"
	NoSpace                Code = "no-space"
	TooManySessions        Code = "too-many-sessions"
	MissingSource          Code = "missing-source"
	UploadsOff             Code = "uploads-off"
	OverQuota              Code = "over-quota"
	TooManyFiles           Code = "too-many-files"
	PathTooDeep            Code = "path-too-deep"
	NoAudio                Code = "no-audio"
	TagsTooLarge           Code = "tags-too-large"
	PictureTooLarge        Code = "picture-too-large"
	UnsupportedFLAC        Code = "unsupported-flac"
	SourceDamaged          Code = "source-damaged"
	EmptyTrack             Code = "empty-track"
	MissingTrackNumber     Code = "missing-track-number"
	DuplicateTrackPosition Code = "duplicate-track-position"
	IncompleteAlbum        Code = "incomplete-album"
	AlreadyInLibrary       Code = "already-in-library"
	CouldNotCheck          Code = "could-not-check"
	ID3Tags                Code = "id3-tags"
	NoCoverArt             Code = "no-cover-art"
	TrackTotalDiffers      Code = "track-total-differs"
	ExtrasDropped          Code = "extras-dropped"
)

// Reason is one entry of the set. Its JSON form is the wire's reason object:
// the code and the three strings a page shows.
//
// A parameterized reason names its parameter in Param and carries Canonical,
// the wording with that parameter filled generically; both are empty for every
// other reason.
type Reason struct {
	Code      Code   `json:"code"`
	Label     string `json:"label"`
	Friendly  string `json:"friendly"`
	NextStep  string `json:"next_step"`
	Blocking  bool   `json:"-"`
	Param     string `json:"-"`
	Canonical string `json:"-"`
}

// badExtensionWording is bad-extension's friendly sentence once the
// configured formats are not FLAC alone.
const badExtensionWording = "This isn't one of the formats the library keeps (%s)."

var set = []Reason{
	{Code: BadExtension, Label: "not FLAC", Blocking: true,
		Friendly:  "This isn't a FLAC file, and the library keeps lossless FLAC.",
		NextStep:  "Download this album as FLAC and upload that instead.",
		Param:     "formats",
		Canonical: fmt.Sprintf(badExtensionWording, "the configured formats")},
	{Code: Symlink, Label: "shortcut file", Blocking: true,
		Friendly: "This is a shortcut, not the music file itself.",
		NextStep: "Replace it with the real file and upload it again."},
	{Code: Corrupt, Label: "damaged in transit", Blocking: true,
		Friendly: "This file didn't survive the upload intact. Uploading it again usually fixes this.",
		NextStep: "Upload it again from the original file."},
	{Code: BelowCDQuality, Label: "below CD quality", Blocking: true,
		Friendly: "This FLAC is below CD quality.",
		NextStep: "Get a CD-quality or better FLAC and upload that instead."},
	{Code: TooLarge, Label: "too big", Blocking: true,
		Friendly: "This upload is larger than the library allows.",
		NextStep: "Ask the admin to add this one by hand."},
	{Code: UnpackFailed, Label: "wouldn't unpack", Blocking: true,
		Friendly: "This archive wouldn't unpack as a ZIP.",
		NextStep: "Repack it as a standard ZIP file and upload it again."},
	{Code: UnsafePath, Label: "unsafe file path", Blocking: true,
		Friendly: "This archive contains a file path the library can't safely use.",
		NextStep: "Repack the album as a plain ZIP with real files only, then upload it again."},
	{Code: NoSpace, Label: "library is full", Blocking: true,
		Friendly: "The library doesn't have enough free space to accept this upload right now.",
		NextStep: "Try again in a little while, or let the admin know if this keeps happening."},
	{Code: TooManySessions, Label: "too many uploads", Blocking: true,
		Friendly: "You already have the maximum number of uploads in progress.",
		NextStep: "Wait for one of your other uploads to finish, then try again."},

	{Code: MissingSource, Label: "source missing", Blocking: true,
		Friendly: "This upload doesn't say where the music came from.",
		NextStep: "Choose where it came from, then upload it again."},
	{Code: UploadsOff, Label: "uploads paused", Blocking: true,
		Friendly: "Uploads are turned off right now.",
		NextStep: "Check back later, or ask the admin when uploads will open again."},
	{Code: OverQuota, Label: "over your limit", Blocking: true,
		Friendly: "Adding this would put you past the space you're allowed in the library.",
		NextStep: "Ask the admin for more room, or send a smaller upload."},
	{Code: TooManyFiles, Label: "too many files", Blocking: true,
		Friendly: "This upload has more files than the library takes at once.",
		NextStep: "Remove files you don't need, such as extra scans, or ask the admin to add this one by hand."},
	{Code: PathTooDeep, Label: "folders too deep", Blocking: true,
		Friendly: "Some files sit in folders nested deeper than the library allows.",
		NextStep: "Move them closer to the top folder and upload again."},
	{Code: NoAudio, Label: "no music found", Blocking: true,
		Friendly: "This upload has no music files in it.",
		NextStep: "Check that you picked the right folder or ZIP, then upload again."},
	{Code: TagsTooLarge, Label: "details too big", Blocking: true,
		Friendly: "The song details stored inside this file take up more room than the library allows.",
		NextStep: "Trim the file's details in your music app and upload it again, or ask the admin to add it by hand."},
	{Code: PictureTooLarge, Label: "picture too big", Blocking: true,
		Friendly: "The pictures stored inside this file are larger than the library allows.",
		NextStep: "Shrink or remove them and upload the file again."},
	{Code: UnsupportedFLAC, Label: "unsupported bit depth", Blocking: true,
		Friendly: "This FLAC uses a bit depth the library doesn't take.",
		NextStep: "Export it again as 16-bit, 20-bit, or 24-bit FLAC."},
	{Code: SourceDamaged, Label: "damaged at source", Blocking: true,
		Friendly: "This file arrived intact, but it was damaged before it was sent.",
		NextStep: "Get a clean copy from where it came from. Uploading it again won't help."},
	{Code: EmptyTrack, Label: "empty track", Blocking: true,
		Friendly: "This track is zero seconds long.",
		NextStep: "Check the original, then upload a complete copy."},
	{Code: MissingTrackNumber, Label: "no track number", Blocking: true,
		Friendly: "This track has no track number the library can read, and the library needs one to keep an album in order.",
		NextStep: "Add the track numbers in your music app, then upload the album again."},
	{Code: DuplicateTrackPosition, Label: "same track number", Blocking: true,
		Friendly: "Two tracks in this album share one track number.",
		NextStep: "Fix the numbering in your music app, then upload the album again."},
	{Code: IncompleteAlbum, Label: "album incomplete", Blocking: true,
		Friendly: "This album's track numbers skip or don't match its track count, so it looks incomplete.",
		NextStep: "Find the missing tracks and upload the whole album together."},
	{Code: AlreadyInLibrary, Label: "already in library", Blocking: true,
		Friendly: "This music is already in the library.",
		NextStep: "There's no need to upload it again. Search the library to find it."},
	{Code: CouldNotCheck, Label: "check didn't finish", Blocking: true,
		Friendly: "The library couldn't finish checking this. The problem is on the library's side, not something you did.",
		NextStep: "Try again in a little while, or let the admin know if this keeps happening."},
	{Code: ID3Tags, Label: "ID3 tags", Blocking: true,
		Friendly: "This FLAC file carries ID3 tags, which belong on MP3s, and the library doesn't accept them.",
		NextStep: "Run mid3v2 --delete-all on the file to remove them, then upload it again."},

	{Code: NoCoverArt, Label: "no cover",
		Friendly: "This album has no cover picture.",
		NextStep: "It can go in without one, or you can cancel, add a picture to the files, and upload again."},
	{Code: TrackTotalDiffers, Label: "track count differs",
		Friendly: "The files say this album has a different number of tracks than you uploaded.",
		NextStep: "Check that every track is here. If the album is complete, you can leave it as it is."},
	{Code: ExtrasDropped, Label: "extras left out",
		Friendly: "When you upload separate tracks, only the music files are kept, so the other files in the ZIP were left out.",
		NextStep: "If you wanted them kept, upload the whole album instead."},
}

// All returns every reason in the set, adopted reasons first.
func All() []Reason { return slices.Clone(set) }

// Lookup returns the reason for code, and false when the set holds none.
func Lookup(code Code) (Reason, bool) {
	i := slices.IndexFunc(set, func(r Reason) bool { return r.Code == code })
	if i < 0 {
		return Reason{}, false
	}
	return set[i], true
}

// Of returns the reason for a code the caller names as a constant. It panics on
// a code the set lacks, which is a programming error; use [Lookup] for text read
// from storage.
func Of(code Code) Reason {
	r, ok := Lookup(code)
	if !ok {
		panic("reject: unknown code " + string(code))
	}
	return r
}

// WithFormats returns bad-extension rendered for the configured format names.
// FLAC alone renders the adopted strings; any other set names itself and
// suggests converting to lossless FLAC.
func WithFormats(formats []string) Reason {
	r := Of(BadExtension)
	if len(formats) == 1 && formats[0] == "flac" {
		return r
	}
	r.Label = "wrong format"
	r.Friendly = fmt.Sprintf(badExtensionWording, joinFormats(formats))
	r.NextStep = "Get a copy in one of these formats and upload that instead."
	return r
}

// joinFormats upper-cases each name and joins them as English: "A", "A or B",
// "A, B, or C".
func joinFormats(formats []string) string {
	names := make([]string, len(formats))
	for i, f := range formats {
		names[i] = strings.ToUpper(f)
	}
	switch len(names) {
	case 0, 1:
		return strings.Join(names, "")
	case 2:
		return names[0] + " or " + names[1]
	}
	return strings.Join(names[:len(names)-1], ", ") + ", or " + names[len(names)-1]
}
