package settings

import (
	"context"
	"time"

	"github.com/glw907/dubplate/internal/config"
)

// A Reader returns a value in force: one setting's, or one a package
// builds from several, such as a set of limits. Each call reads afresh,
// so a caller that reads at the point of use follows a change without
// a restart. A non-nil error means the value is unknown, and is a
// [*ReadError] naming the setting when an [Accessor] made it.
type Reader[T any] func(ctx context.Context) (T, error)

// A ReadError reports a failed read of the setting named Key. A
// caller that logs a failure may log Key alone, since Err's text
// carries whatever the read met, a stored value among it.
type ReadError struct {
	Key string
	Err error
}

func (e *ReadError) Error() string { return e.Key + ": " + e.Err.Error() }

func (e *ReadError) Unwrap() error { return e.Err }

// IntReader returns the [Reader] of s, a KindInt setting, through a.
func (a *Accessor) IntReader(s config.Setting) Reader[int] {
	return func(ctx context.Context) (int, error) { return a.Int(ctx, s) }
}

// DurationReader returns the [Reader] of s, a KindDuration setting,
// through a.
func (a *Accessor) DurationReader(s config.Setting) Reader[time.Duration] {
	return func(ctx context.Context) (time.Duration, error) { return a.Duration(ctx, s) }
}

// FloatReader returns the [Reader] of s, a KindFloat setting, through a.
func (a *Accessor) FloatReader(s config.Setting) Reader[float64] {
	return func(ctx context.Context) (float64, error) { return a.Float(ctx, s) }
}
