INSERT INTO jobs (
	id, kind, at, state, owner_instance, actor, cancel_requested,
	progress, failure_code, failure_message, started_at, updated_at, finished_at
) VALUES
	(x'e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1', 'backup', 1700000000000000, 'succeeded', 'inst-1', 'cli', 0,
	 NULL, NULL, NULL, 1700000001000000, 1700000010000000, 1700000010000000),
	(x'e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2', 'backup', 1700000020000000, 'failed', 'inst-1', 'ops/local', 0,
	 NULL, 'disk_full', 'store disk full', 1700000021000000, 1700000030000000, 1700000030000000),
	(x'e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3', 'backup', 1700000040000000, 'queued', 'inst-2', 'cli', 1,
	 NULL, NULL, NULL, NULL, 1700000040000000, NULL);
