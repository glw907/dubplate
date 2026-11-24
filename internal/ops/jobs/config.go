package jobs

import (
	"context"
	"time"

	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/settings"
)

// jobRetentionSetting is DUBPLATE_JOB_RETENTION: how long a finished
// job record survives past its own kind's newest succeeded record.
// Its default, 2160h (90 days), carries unchanged the value
// this package's retired retention constant held.
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

// BackupConfigFrom builds the backup job's configuration from cfg and
// acc: the flag and the remote through the integration class, the
// library root through its owned wall row, and the three value readers
// through acc.
func BackupConfigFrom(cfg *config.Config, acc *settings.Accessor) BackupConfig {
	guardFraction, snapshotKeep, archiveRetention := backup.Readers(acc)
	return BackupConfig{
		Enabled:          cfg.IntegrationEnabled(backup.Class.Name),
		Remote:           cfg.IntegrationValue(backup.Class.Name, backup.RemoteKey),
		LibraryRoot:      cfg.LibraryRoot,
		GuardFraction:    guardFraction,
		SnapshotKeep:     snapshotKeep,
		ArchiveRetention: archiveRetention,
	}
}
