package gate

import "errors"

// Unit is what a submission's verdict covers.
type Unit string

const (
	// UnitAlbum judges every file together: one failing track rejects the
	// whole album.
	UnitAlbum Unit = "album"
	// UnitTrack judges each file on its own.
	UnitTrack Unit = "track"
)

// Container is how a submission's files arrive.
type Container string

const (
	// ContainerFiles is loose files, one upload each.
	ContainerFiles Container = "files"
	// ContainerZip is one ZIP archive.
	ContainerZip Container = "zip"
)

// ErrCouldNotRun marks a check the server failed to complete: a read error
// on a staged file, or a tagio error that is no verdict about the file. It
// is never a verdict, because a verdict that blames the file removes its
// bytes. A cancelled context is neither this nor a verdict.
var ErrCouldNotRun = errors.New("the check could not run")
