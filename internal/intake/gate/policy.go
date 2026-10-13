package gate

import (
	"path"
	"slices"
	"strings"

	"github.com/glw907/dubplate/internal/engine/tagio"
	"github.com/glw907/dubplate/internal/reject"
)

// Policy is the live limits a check runs under, read once per submission.
// Byte sizes are in bytes.
type Policy struct {
	MaxAlbumBytes    int64
	MaxFiles         int
	MaxPathDepth     int
	MaxTagBytes      int64
	MaxPictureBytes  int64
	MinSampleRate    int
	MinBitsPerSample int

	// Allowlist names the formats this submission admits: the
	// DUBPLATE_INTAKE_FORMAT_ALLOWLIST bound intersected with its
	// collection's admission_format_allowlist when that is set. Every name
	// has a row in the format table.
	Allowlist []string
}

// FormatRow ties one configured format name to the tagio format that reads
// it and the file extensions that name it, lowercase with the dot.
type FormatRow struct {
	Name       string
	Format     tagio.Format
	Extensions []string
}

// formats is the format table, one row per name the allowlist may hold.
var formats = []FormatRow{
	{Name: "flac", Format: tagio.FormatFLAC, Extensions: []string{".flac"}},
}

// LookupFormat returns the row for a configured format name, and false when
// the table holds none.
func LookupFormat(name string) (FormatRow, bool) {
	for _, row := range formats {
		if row.Name == name {
			return row, true
		}
	}
	return FormatRow{}, false
}

// audioExtensions are the extensions the gate takes for music, lowercase
// with the dot. The format table names the formats the library keeps; this
// adds the common ones it does not, so an MP3 answers bad-extension where
// it would otherwise ride along as an extra.
var audioExtensions = []string{".flac", ".mp3", ".m4a", ".ogg", ".opus", ".wav", ".aiff"}

// IsAudio reports whether a member's path names a music file by its
// extension, a format the library keeps or not. A file it refuses is an
// extra, which rides along with an album.
func IsAudio(path string) bool {
	return isAudioExt(extension(path))
}

// isAudioExt reports whether ext names a music format, kept or not.
func isAudioExt(ext string) bool {
	return slices.Contains(audioExtensions, ext) || slices.ContainsFunc(formats, func(row FormatRow) bool {
		return slices.Contains(row.Extensions, ext)
	})
}

// admitted returns the allowlisted format row ext names, and false when no
// row in the policy's allowlist carries it.
func (p Policy) admitted(ext string) (FormatRow, bool) {
	for _, name := range p.Allowlist {
		if row, ok := LookupFormat(name); ok && slices.Contains(row.Extensions, ext) {
			return row, true
		}
	}
	return FormatRow{}, false
}

// extension is a member path's extension, lowercase with the dot.
func extension(name string) string {
	return strings.ToLower(path.Ext(name))
}

// FileReason is one file's refusal or warning: the path as the member sent
// it, in bytes because it need not be valid UTF-8, and the reason it drew.
type FileReason struct {
	Path   []byte
	Reason reject.Reason
}
