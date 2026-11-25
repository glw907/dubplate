package settings

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/store"
)

// LocalSnapshotKeep decides [store.ServeOptions.Keep] before serve
// opens the store, from the settings facts f an inspection into report
// read and the registry's own setting, and refuses a start that a migration would otherwise make
// unrecoverable.
//
// A stored keep row in its bounds is the keep, even when the
// environment names another value: that value reaches the keep at the
// seed this start applies after its migration. The outcome's own
// value is the value in force after the seed, so it decides only when
// the row is out of bounds, where it is the healed value or a refusal
// naming the key and the bound. With no keep row, or no facts, the
// registry default is the keep and the open decides whether the state
// itself refuses.
//
// An older store whose evaluation refuses anything, whether an
// administrator's row out of bounds or an environment value that does
// not parse, is refused here, before the migration moves it past the
// binary a rollback would run. An older store is evaluated against
// the facts [SeedFacts] picks, as check and doctor evaluate it, so one
// whose version predates the settings table, or the jobs table and
// with it any facts at all, is evaluated against the empty table its
// first seed meets. A store that does not migrate gets no such check,
// nor does one that carries no facts, and the seed apply after the
// open decides.
func LocalSnapshotKeep(report store.Report, f Facts, environ []string, reg *config.Registry) (int, error) {
	setting := byKey(reg.Settings(), store.LocalSnapshotKeepSetting.Key)
	keep, err := strconv.Atoi(setting.Default)
	if err != nil {
		return 0, fmt.Errorf("%s: default %q: %w", setting.Key, setting.Default, err)
	}
	var facts Facts
	switch {
	case report.State == store.StateOlder:
		facts, _ = SeedFacts(report, f)
	case report.Facts != nil:
		facts = f
	default:
		return keep, nil
	}

	ev := EvaluateFacts(reg, environ, facts)
	if report.State == store.StateOlder && ev.Err != nil {
		return 0, olderStoreRefusal(ev)
	}
	for _, o := range ev.Outcomes {
		if o.Key != setting.Key || o.Stored == nil {
			continue
		}
		if setting.Validate(o.Stored.Value) == nil {
			return strconv.Atoi(o.Stored.Value)
		}
		if o.Err != nil {
			return 0, o.Err
		}
		return strconv.Atoi(o.Value)
	}
	return keep, nil
}

// olderStoreRefusal is ev's refusal of an older store. This binary
// cannot write an older store, so settings reset cannot clear an
// administrator's row it refuses over; when any refusing key holds
// one, the error names the recovery that can: the previous release's
// own settings reset, then the upgrade.
func olderStoreRefusal(ev Evaluation) error {
	var keys []string
	for _, o := range ev.Outcomes {
		if o.Err != nil && o.Stored != nil && o.Stored.Source == SourceAdmin {
			keys = append(keys, o.Key)
		}
	}
	if len(keys) == 0 {
		return ev.Err
	}
	return fmt.Errorf("%w; this store is older than this binary, which cannot write it: redeploy the previous release, run its settings reset for %s, then upgrade",
		ev.Err, strings.Join(keys, ", "))
}

// byKey returns the setting named key out of list, or its zero value
// when list carries none, the shape a defensive caller falls back on
// rather than panics on a key [Evaluate] could not have
// invented.
func byKey(list []config.Setting, key string) config.Setting {
	for _, s := range list {
		if s.Key == key {
			return s
		}
	}
	return config.Setting{}
}
