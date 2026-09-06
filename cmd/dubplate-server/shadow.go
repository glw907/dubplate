package main

import (
	"sync"

	"github.com/glw907/dubplate/internal/config"
)

// shadowFloorThreshold is the evidence floor rung 15 measures the
// shadow window against: a configurable duration with a 30-day default
// (build spec ruling 1). Provisional, owner ruling 4 open.
var shadowFloorThreshold = config.Threshold{
	Key:     "DUBPLATE_SHADOW_FLOOR",
	Kind:    config.KindDuration,
	Default: "720h",
	Doc:     "a Go duration, for example 720h",
}

// shadowBudgetThreshold caps one observe pass's MusicBrainz spend. An
// album costs one lookup on the ID-consensus early return
// (internal/engine/tag/tag.go:96-118) and, on the search path, one
// search plus SearchLimit lookups plus one pseudo-release fetch per
// candidate (internal/engine/musicbrainz/client.go:320-333), so the
// real per-album bound is eleven rather than seven. Sixty tokens at two
// seconds a token bounds a pass at two minutes.
// Provisional, owner ruling 5 open.
var shadowBudgetThreshold = config.Threshold{
	Key:     "DUBPLATE_SHADOW_BUDGET",
	Kind:    config.KindInt,
	Default: "60",
	Doc:     "an integer request count per pass",
}

// registerShadowThresholds runs once per process, from newRootCmd
// (Resolution R12): config.Register appends without dedup
// (internal/config/thresholds.go:28-30) and newRootCmd runs more than
// once in a test binary.
var registerShadowThresholds = sync.OnceFunc(func() {
	config.Register(shadowFloorThreshold)
	config.Register(shadowBudgetThreshold)
})
