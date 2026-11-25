package settings

import (
	"context"

	"github.com/glw907/dubplate/internal/store"
)

// settingsVersion is the lowest store version holding the settings
// table (0004_settings.sql).
const settingsVersion = 4

// Facts are the settings rows a store inspection read, for the seed
// rule to evaluate without a connection of its own.
type Facts struct {
	// Rows are the settings rows in key order.
	Rows []Row

	// Absent is set for a store that predates the settings table.
	Absent bool
}

// Reader returns the [store.Reader] that fills f from the settings
// table.
func (f *Facts) Reader() store.Reader {
	return func(ctx context.Context, q store.Querier, version int) error {
		if version < settingsVersion {
			f.Absent = true
			return nil
		}
		rows, err := readRows(ctx, q.QueryContext)
		if err != nil {
			return err
		}
		f.Rows = rows
		return nil
	}
}

// SeedFacts returns the facts the seed rule is evaluated against for a
// store in report, whose inspection ran f's reader, and false when
// there are none to evaluate. An absent store, and an older one whose
// version predates the settings table, holds no settings row: its
// facts are an absent table, the empty row set serve's first seed
// meets, so an environment value that start would refuse is refused
// before it. An older store too old for [store.Inspect] to read any
// facts is evaluated the same way. Any other state evaluates f, and a
// state whose inspection read no facts has nothing to evaluate.
func SeedFacts(report store.Report, f Facts) (Facts, bool) {
	switch {
	case report.State == store.StateAbsent,
		report.State == store.StateOlder && (report.Facts == nil || f.Absent):
		return Facts{Absent: true}, true
	case report.Facts != nil:
		return f, true
	}
	return Facts{}, false
}
