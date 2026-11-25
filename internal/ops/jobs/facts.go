package jobs

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/store"
)

// jobsVersion is the lowest store version holding the jobs table
// (0002_jobs.sql). It equals the version below which [store.Inspect]
// reads no facts at all, and the reader checks it anyway, so the jobs
// package never leans on the store's threshold.
const jobsVersion = 2

// Facts are the job facts a store inspection reads for the doctor
// command.
type Facts struct {
	// LastSuccess maps each job kind that has succeeded to its newest
	// succeeded record. It is nil for a store that predates the jobs
	// table, and for an inspection that never ran the reader.
	LastSuccess map[string]ops.JobSuccess
}

// Reader returns the [store.Reader] that fills f from the jobs table.
// A tie in finish time goes to the higher id.
func (f *Facts) Reader() store.Reader {
	return func(ctx context.Context, q store.Querier, version int) error {
		if version < jobsVersion {
			return nil
		}
		last, err := lastSuccesses(ctx, q)
		if err != nil {
			return fmt.Errorf("read jobs: %w", err)
		}
		f.LastSuccess = last
		return nil
	}
}

func lastSuccesses(ctx context.Context, q store.Querier) (map[string]ops.JobSuccess, error) {
	rows, err := q.QueryContext(ctx, `SELECT kind, finished_at, progress FROM (
	SELECT kind, finished_at, progress,
		row_number() OVER (PARTITION BY kind ORDER BY finished_at DESC, id DESC) AS rank
	FROM jobs WHERE state = 'succeeded' AND finished_at IS NOT NULL
) WHERE rank = 1`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	last := make(map[string]ops.JobSuccess)
	for rows.Next() {
		var (
			kind     string
			finished int64
			progress sql.NullString
		)
		if err := rows.Scan(&kind, &finished, &progress); err != nil {
			return nil, err
		}
		last[kind] = ops.JobSuccess{Finished: store.Micros(finished), Progress: progress.String}
	}
	return last, rows.Err()
}
