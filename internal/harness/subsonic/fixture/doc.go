// Package fixture builds the streamer harness's test library: a small
// tagged FLAC collection described by a committed manifest.
//
// The manifest declares every file's relative path and every tag value
// as data. Paths are never computed, because the import layout is not
// this package's to decide. Audio and image seeds sit beside the
// manifest, each pinned by its sha256, and a build copies the seeds
// and tags each track through [tagio.Write], so two builds of one
// manifest are byte-identical.
//
// The oracles shape what the library can exercise. gonic answers
// getMusicDirectory without an id, or with id=0, with error 10. Navidrome
// serves a per-track cover that differs from the album cover. gonic splits
// an album held in disc subfolders into two albums, as Borealis does in
// the seed. Neither oracle reports a duration for a track under one
// second, so a fixture that exercises duration needs two seconds or more.
// The seed holds no lyrics, nowPlaying, or radio requests, so the contract
// flags that depend on them wait for a later recording.
package fixture
