package backup

import (
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/settings"
)

// RemoteKey is the integration's one required key: the rclone
// destination a backup syncs to.
const RemoteKey = "DUBPLATE_BACKUP_REMOTE"

// Class is the backup integration: its own enable flag and the one key
// required while it is on, the rclone destination a backup syncs to.
var Class = config.IntegrationClass{
	Name: "backup",
	Flag: config.KeyDecl{
		Key:         "DUBPLATE_BACKUP_ENABLED",
		Description: "turns on backups to the remote",
		Sensitivity: config.Sensitive,
	},
	Required: []config.KeyDecl{
		{Key: RemoteKey, Description: "the rclone destination a backup syncs to", Sensitivity: config.Sensitive},
	},
}

// guardFractionSetting is DUBPLATE_BACKUP_GUARD_FRACTION: the share of
// the catalog's tracks a library tree must hold as regular files
// before a sync may run. Its default, 0.5, carries unchanged the value
// the retired guardFraction constant held. The floor is the protective
// minimum against a value that would turn the short-tree guard off by
// accident.
var guardFractionSetting = config.Setting{
	Key:         "DUBPLATE_BACKUP_GUARD_FRACTION",
	Kind:        config.KindFloat,
	Default:     "0.5",
	Doc:         "a decimal fraction from 0.25 to 1, for example 0.5",
	Apply:       config.ApplyNextOperation,
	Floor:       "0.25",
	Ceiling:     "1",
	Description: "the share of the catalog's tracks the library tree must hold as files before a backup sync runs",
	Sensitivity: config.Sensitive,
}

// snapshotKeepSetting is DUBPLATE_BACKUP_SNAPSHOT_KEEP: how many
// remote snapshots a prune keeps in total, the run's own included. Its
// default, 14, carries unchanged the value the retired snapshotKeep
// constant held. The floor is the protective minimum against a keep
// that would let one run delete every other pre-upgrade remote
// snapshot.
var snapshotKeepSetting = config.Setting{
	Key:         "DUBPLATE_BACKUP_SNAPSHOT_KEEP",
	Kind:        config.KindInt,
	Default:     "14",
	Doc:         "an integer count of remote snapshots to keep, for example 14",
	Apply:       config.ApplyNextOperation,
	Floor:       "3",
	Ceiling:     "1000",
	Description: "how many remote snapshots a backup keeps in total",
	Sensitivity: config.Sensitive,
}

// archiveRetentionSetting is DUBPLATE_BACKUP_ARCHIVE_RETENTION: how
// long an archive directory survives past its stamp. Its default,
// 2160h (90 days), carries unchanged the value the retired
// archiveRetention constant held. The floor is the protective minimum
// against a retention that would purge within a day the files a bad
// sync moved into archive/.
var archiveRetentionSetting = config.Setting{
	Key:         "DUBPLATE_BACKUP_ARCHIVE_RETENTION",
	Kind:        config.KindDuration,
	Default:     "2160h",
	Doc:         "a Go duration string, for example 2160h",
	Apply:       config.ApplyNextOperation,
	Floor:       "168h",
	Ceiling:     "87600h",
	Description: "how long a backup keeps an archived directory after moving it aside",
	Sensitivity: config.Sensitive,
}

// Register adds [Class] and the three backup settings to reg. The
// command calls it after the store's. The six names backup replaced
// (five DUBPLATE_R2_* names and DUBPLATE_BACKUP_STAGING) are
// superseded entries declared in internal/config/keys.go: a
// supersession message is base-wall record-keeping, not backup's own
// surface, so it stays with the package that seeds every registry. No
// reserved class claims those six names, since a name claimed by both
// families panics at registration (internal/config/registry.go,
// claim).
func Register(reg *config.Registry) {
	reg.AddIntegrationClass(Class)
	reg.AddSetting(guardFractionSetting)
	reg.AddSetting(snapshotKeepSetting)
	reg.AddSetting(archiveRetentionSetting)
}

// Readers returns the accessor's typed read of each backup setting,
// one reader per value a [Job] takes: the guard fraction, the snapshot
// count, and the archive retention.
func Readers(acc *settings.Accessor) (
	guardFraction settings.Reader[float64],
	snapshotKeep settings.Reader[int],
	archiveRetention settings.Reader[time.Duration],
) {
	return acc.FloatReader(guardFractionSetting),
		acc.IntReader(snapshotKeepSetting),
		acc.DurationReader(archiveRetentionSetting)
}
