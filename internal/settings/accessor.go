package settings

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
)

// An Accessor serves typed reads of the settings a registry declares,
// from one store. It caches the store's settings rows and, on every
// read, compares the store's settings generation with the one it
// cached under, reloading when they differ, so a change any process
// commits is seen by the next read. It is safe for concurrent use.
//
// A setting read before a write enters [store.Handle.Update] is the
// value in force for that operation; read it outside the callback.
type Accessor struct {
	h        *store.Handle
	settings map[string]config.Setting

	mu         sync.Mutex
	loaded     bool
	generation int64
	rows       map[string]Row
}

// NewAccessor returns an accessor over h for the settings registered
// on reg.
func NewAccessor(h *store.Handle, reg *config.Registry) *Accessor {
	settings := make(map[string]config.Setting)
	for _, s := range reg.Settings() {
		settings[s.Key] = s
	}
	return &Accessor{h: h, settings: settings}
}

// Int returns the value in force for s, a KindInt setting.
func (a *Accessor) Int(ctx context.Context, s config.Setting) (int, error) {
	return read(ctx, a, s, config.KindInt, strconv.Atoi)
}

// Duration returns the value in force for s, a KindDuration setting.
func (a *Accessor) Duration(ctx context.Context, s config.Setting) (time.Duration, error) {
	return read(ctx, a, s, config.KindDuration, time.ParseDuration)
}

// Float returns the value in force for s, a KindFloat setting.
func (a *Accessor) Float(ctx context.Context, s config.Setting) (float64, error) {
	return read(ctx, a, s, config.KindFloat, func(v string) (float64, error) {
		return strconv.ParseFloat(v, 64)
	})
}

// read returns s's value in force parsed through p, refusing a setting
// a's registry does not declare or declares under another kind.
func read[T any](ctx context.Context, a *Accessor, s config.Setting, kind config.ThresholdKind, p func(string) (T, error)) (T, error) {
	var zero T
	reg, ok := a.settings[s.Key]
	if !ok {
		return zero, fmt.Errorf("%s: not a registered setting", s.Key)
	}
	if reg.Kind != kind {
		return zero, fmt.Errorf("%s: a %v setting read as %v", s.Key, reg.Kind, kind)
	}
	v, err := a.value(ctx, reg)
	if err != nil {
		return zero, err
	}
	return p(v)
}

// value returns s's value in force: its stored row's value, or the
// registry default when the store holds none. A stored value that
// does not parse or falls outside s's bounds is an error, never a
// silent default. Only s's own row is validated, so another key's bad
// row does not fail this read.
func (a *Accessor) value(ctx context.Context, s config.Setting) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// The generation is read before the rows, and the rows are cached
	// under it. Each read-pool call is its own snapshot, so a write
	// landing between the two reads leaves newer rows under an older
	// generation, which the next read reloads; the other order could
	// cache older rows under a newer generation and serve them until
	// some later write.
	var generation int64
	if err := a.h.QueryRow(ctx, `SELECT generation FROM settings_generation WHERE singleton = 1`).Scan(&generation); err != nil {
		return "", fmt.Errorf("%s: read settings_generation: %w", s.Key, err)
	}
	if !a.loaded || generation != a.generation {
		rows, err := readRows(ctx, a.h)
		if err != nil {
			return "", fmt.Errorf("%s: %w", s.Key, err)
		}
		a.rows = make(map[string]Row, len(rows))
		for _, r := range rows {
			a.rows[r.Key] = r
		}
		a.generation, a.loaded = generation, true
	}

	r, ok := a.rows[s.Key]
	if !ok {
		return s.Default, nil
	}
	if err := s.Validate(r.Value); err != nil {
		return "", fmt.Errorf("stored %s value refused: %w", r.Source, err)
	}
	return r.Value, nil
}
