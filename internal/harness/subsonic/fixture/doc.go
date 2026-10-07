// Package fixture builds the streamer harness's test library: a small
// tagged FLAC collection described by a committed manifest.
//
// The manifest declares every file's relative path and every tag value
// as data. Paths are never computed, because the import layout is not
// this package's to decide. Audio and image seeds sit beside the
// manifest, each pinned by its sha256, and a build copies the seeds
// and tags each track through [tagio.Write], so two builds of one
// manifest are byte-identical.
package fixture
