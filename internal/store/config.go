package store

import (
	"errors"
	"fmt"
	"time"

	"github.com/glw907/dubplate/internal/config"
)

// busyTimeoutThreshold is DUBPLATE_STORE_BUSY_TIMEOUT: the
// inspect mode's only production reader at this rung, doctor and
// check, bounding how long a read waits on another process's write
// lock inside SQLite. The 15s default is long enough to ride out a
// snapshot or a migration holding the lock, short enough that a stuck
// writer surfaces as a doctor failure rather than a hang.
var busyTimeoutThreshold = config.Threshold{
	Key:         "DUBPLATE_STORE_BUSY_TIMEOUT",
	Kind:        config.KindDuration,
	Default:     "15s",
	Doc:         "a Go duration string, for example 15s",
	Description: "how long a read waits on another process's write lock",
	Mark:        config.NotSensitive,
}

// busyTimeoutFloorMS and busyTimeoutCeilingMS bound
// DUBPLATE_STORE_BUSY_TIMEOUT in milliseconds, the unit SQLite's
// _busy_timeout takes. The floor is 1ms: a positive duration under it
// would truncate to zero, no busy wait at all. The ceiling is the
// largest value a probe measured the driver's DSN parser
// (strconv.ParseInt(..., 10, 64) in sqlite.go) reading back unchanged;
// values above it read back as 0.
const (
	busyTimeoutFloorMS   = 1
	busyTimeoutCeilingMS = 2147483647
)

// busyTimeoutFloor and busyTimeoutCeiling are the same bound
// as [time.Duration], compared against the threshold's parsed value at
// nanosecond precision so a value like 500us fails the floor even
// though it would round to a whole millisecond.
var (
	busyTimeoutFloor   = busyTimeoutFloorMS * time.Millisecond
	busyTimeoutCeiling = busyTimeoutCeilingMS * time.Millisecond
)

// migrationWaitThreshold is DUBPLATE_STORE_MIGRATION_WAIT: how
// long serve waits for the store lock before it exits naming the
// holders. A current store waits only on an exclusive holder, a
// restore, since shared locks coexist; a migrating start waits on
// every shared holder, a backup's sync above all, which on a first run
// lasts hours. The 5m default outlasts every member verb and every
// snapshot copy a group-scale store takes and is far shorter than a
// first sync; the unit's restart policy retries a start that gave up.
var migrationWaitThreshold = config.Threshold{
	Key:         "DUBPLATE_STORE_MIGRATION_WAIT",
	Kind:        config.KindDuration,
	Default:     "5m",
	Doc:         "a Go duration string, for example 5m",
	Description: "how long a start waits for the store lock before it gives up",
	Mark:        config.NotSensitive,
}

// migrationWaitFloor and migrationWaitCeiling bound
// DUBPLATE_STORE_MIGRATION_WAIT.
const (
	migrationWaitFloor   = time.Second
	migrationWaitCeiling = 24 * time.Hour
)

// LocalSnapshotKeepSetting is DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP:
// how many local pre-migration snapshot copies OpenServe's own keep
// bound retains. Its default of 3 is a working value the retired
// package never set: a rollback needs the newest copy, and two older
// copies cover a bad upgrade found a release late, each a full store
// whose size doctor already shows. doctor is its first reader.
var LocalSnapshotKeepSetting = config.Setting{
	Key:         "DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP",
	Kind:        config.KindInt,
	Default:     "3",
	Doc:         "an integer count of local snapshot copies to keep, for example 3",
	Apply:       config.ApplyNextOperation,
	Floor:       "1",
	Ceiling:     "100",
	Description: "how many local pre-migration snapshot copies the server keeps",
	Mark:        config.NotSensitive,
}

// Register registers the store's own thresholds and
// store-homed settings against reg. newProductionRegistry calls it
// after jobs.Register.
func Register(reg *config.Registry) {
	reg.AddThreshold(busyTimeoutThreshold)
	reg.AddThreshold(migrationWaitThreshold)
	reg.AddSetting(LocalSnapshotKeepSetting)
}

// ErrOutOfRange marks a threshold that parsed but fell outside its
// range, as opposed to one that did not parse at all.
var ErrOutOfRange = errors.New("out of range")

// rangeError is an out-of-range refusal whose message is its own
// sentence and whose Unwrap is [ErrOutOfRange].
type rangeError string

func (e rangeError) Error() string { return string(e) }

func (e rangeError) Unwrap() error { return ErrOutOfRange }

// BusyTimeout resolves DUBPLATE_STORE_BUSY_TIMEOUT from environ
// and checks it against its range. A value that parses outside the
// range is refused wrapping ErrOutOfRange and naming the key and the
// bound it crossed; a value that fails to parse is returned as
// config.ParseDuration reports it, a failure doctor's own threshold
// row already names.
func BusyTimeout(environ []string) (time.Duration, error) {
	d, err := config.ParseDuration(environ, busyTimeoutThreshold)
	if err != nil {
		return 0, err
	}
	switch {
	case d < busyTimeoutFloor:
		return 0, rangeError(fmt.Sprintf("%s must be at least %dms", busyTimeoutThreshold.Key, busyTimeoutFloorMS))
	case d > busyTimeoutCeiling:
		return 0, rangeError(fmt.Sprintf("%s must be at most %dms", busyTimeoutThreshold.Key, busyTimeoutCeilingMS))
	}
	return d, nil
}

// MigrationWait resolves DUBPLATE_STORE_MIGRATION_WAIT from
// environ and checks it against its range, naming the key on every
// refusal.
func MigrationWait(environ []string) (time.Duration, error) {
	d, err := config.ParseDuration(environ, migrationWaitThreshold)
	if err != nil {
		return 0, err
	}
	switch {
	case d < migrationWaitFloor:
		return 0, fmt.Errorf("%s must be at least %s", migrationWaitThreshold.Key, migrationWaitFloor)
	case d > migrationWaitCeiling:
		return 0, fmt.Errorf("%s must be at most %s", migrationWaitThreshold.Key, migrationWaitCeiling)
	}
	return d, nil
}
