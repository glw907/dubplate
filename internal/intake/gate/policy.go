package gate

import (
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

// FileReason is one file's refusal or warning: the path as the member sent
// it, in bytes because it need not be valid UTF-8, and the reason it drew.
type FileReason struct {
	Path   []byte
	Reason reject.Reason
}
