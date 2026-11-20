package local

import (
	"context"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/settings"
)

// jobRetentionSetting is DUBPLATE_JOB_RETENTION: how long a finished
// job record survives past its own kind's newest succeeded record.
// Its default, 2160h (90 days), carries unchanged the value
// ops/local's retired retention constant held.
var jobRetentionSetting = config.Setting{
	Key:         "DUBPLATE_JOB_RETENTION",
	Kind:        config.KindDuration,
	Default:     "2160h",
	Doc:         "a Go duration string, for example 2160h",
	Apply:       config.ApplyNextOperation,
	Floor:       "24h",
	Ceiling:     "87600h",
	Description: "how long a finished job record is kept",
	Mark:        config.NotSensitive,
}

// Register adds the job runner's one setting to reg.
func Register(reg *config.Registry) {
	reg.AddSetting(jobRetentionSetting)
}

// JobRetention returns the closure over acc that [Config.Retention]
// reads before every job start: the accessor's typed read of
// DUBPLATE_JOB_RETENTION.
func JobRetention(acc *settings.Accessor) func(context.Context) (time.Duration, error) {
	return func(ctx context.Context) (time.Duration, error) {
		return acc.Duration(ctx, jobRetentionSetting)
	}
}
