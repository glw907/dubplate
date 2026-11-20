CREATE TABLE member_overrides (
	user_id BLOB NOT NULL PRIMARY KEY REFERENCES users (id),
	quota_bytes INTEGER CHECK (quota_bytes IS NULL OR quota_bytes >= 0),
	may_upload INTEGER CHECK (may_upload IS NULL OR may_upload IN (0, 1)),
	may_create_collections INTEGER CHECK (may_create_collections IS NULL OR may_create_collections IN (0, 1)),
	upload_concurrency_cap INTEGER CHECK (upload_concurrency_cap IS NULL OR upload_concurrency_cap >= 1),
	updated_at INTEGER NOT NULL
) STRICT;

CREATE TABLE member_preferences (
	user_id BLOB NOT NULL PRIMARY KEY REFERENCES users (id),
	collapse_duplicates INTEGER NOT NULL DEFAULT 0 CHECK (collapse_duplicates IN (0, 1)),
	updated_at INTEGER NOT NULL
) STRICT;

CREATE TABLE member_quota (
	user_id BLOB NOT NULL PRIMARY KEY REFERENCES users (id),
	charged_bytes INTEGER NOT NULL DEFAULT 0 CHECK (charged_bytes >= 0),
	updated_at INTEGER NOT NULL
) STRICT;

-- group_quota holds one row, on settings_generation's singleton
-- pattern: singleton is neither named "id" nor ending "_id", so
-- invariant 6's INTEGER PRIMARY KEY rule does not fire on it.
CREATE TABLE group_quota (
	singleton INTEGER NOT NULL PRIMARY KEY CHECK (singleton = 1),
	physical_total_bytes INTEGER NOT NULL DEFAULT 0
) STRICT;

INSERT INTO group_quota (singleton, physical_total_bytes) VALUES (1, 0);

CREATE TABLE device_credentials (
	id BLOB NOT NULL PRIMARY KEY,
	credential_hash BLOB NOT NULL UNIQUE,
	user_id BLOB NOT NULL REFERENCES users (id),
	device_id BLOB NOT NULL UNIQUE,
	created_at INTEGER NOT NULL,
	last_seen_address TEXT,
	expires_at INTEGER NOT NULL,
	revoked_at INTEGER
) STRICT;

-- old_path and new_path repeat collections.root_path's relative-path
-- CHECK (0003_catalog.sql): non-empty, no leading slash, no NUL, and no
-- empty, ".", or ".." component.
CREATE TABLE move_intents (
	id BLOB NOT NULL PRIMARY KEY,
	at INTEGER NOT NULL,
	kind TEXT NOT NULL,
	old_path BLOB NOT NULL CHECK (
		length(old_path) > 0
		AND CAST(substr(old_path, 1, 1) AS BLOB) != X'2F'
		AND instr(old_path, X'00') = 0
		AND instr(CAST(X'2F' || old_path || X'2F' AS BLOB), X'2F2F') = 0
		AND instr(CAST(X'2F' || old_path || X'2F' AS BLOB), X'2F2E2F') = 0
		AND instr(CAST(X'2F' || old_path || X'2F' AS BLOB), X'2F2E2E2F') = 0
	),
	new_path BLOB NOT NULL CHECK (
		length(new_path) > 0
		AND CAST(substr(new_path, 1, 1) AS BLOB) != X'2F'
		AND instr(new_path, X'00') = 0
		AND instr(CAST(X'2F' || new_path || X'2F' AS BLOB), X'2F2F') = 0
		AND instr(CAST(X'2F' || new_path || X'2F' AS BLOB), X'2F2E2F') = 0
		AND instr(CAST(X'2F' || new_path || X'2F' AS BLOB), X'2F2E2E2F') = 0
	),
	granularity TEXT NOT NULL CHECK (granularity IN ('file', 'directory')),
	state TEXT NOT NULL CHECK (state IN ('pending', 'done'))
) STRICT;
