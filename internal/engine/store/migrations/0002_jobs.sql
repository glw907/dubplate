CREATE TABLE jobs (
	id BLOB NOT NULL PRIMARY KEY,
	kind TEXT NOT NULL,
	at INTEGER NOT NULL,
	state TEXT NOT NULL CHECK (state IN ('queued', 'running', 'succeeded', 'failed', 'canceled', 'interrupted')),
	owner_instance TEXT NOT NULL,
	actor TEXT NOT NULL,
	cancel_requested INTEGER NOT NULL DEFAULT 0 CHECK (cancel_requested IN (0, 1)),
	progress TEXT,
	failure_code TEXT,
	failure_message TEXT,
	started_at INTEGER,
	updated_at INTEGER NOT NULL,
	finished_at INTEGER
) STRICT;

CREATE INDEX jobs_kind_state_finished ON jobs (kind, state, finished_at);
