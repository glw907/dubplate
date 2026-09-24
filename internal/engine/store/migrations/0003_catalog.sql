CREATE TABLE users (
	id BLOB NOT NULL PRIMARY KEY,
	address TEXT NOT NULL UNIQUE,
	display_name TEXT NOT NULL,
	streamer_username TEXT NOT NULL UNIQUE COLLATE NOCASE,
	admin INTEGER NOT NULL DEFAULT 0 CHECK (admin IN (0, 1)),
	tombstoned_at INTEGER,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL
) STRICT;

CREATE TABLE collections (
	id BLOB NOT NULL PRIMARY KEY,
	name TEXT NOT NULL,
	root_path BLOB NOT NULL UNIQUE CHECK (
		length(root_path) > 0
		AND CAST(substr(root_path, 1, 1) AS BLOB) != X'2F'
		AND instr(root_path, X'00') = 0
		AND instr(CAST(X'2F' || root_path || X'2F' AS BLOB), X'2F2F') = 0
		AND instr(CAST(X'2F' || root_path || X'2F' AS BLOB), X'2F2E2F') = 0
		AND instr(CAST(X'2F' || root_path || X'2F' AS BLOB), X'2F2E2E2F') = 0
	),
	owner BLOB NOT NULL REFERENCES users (id),
	visibility TEXT NOT NULL CHECK (visibility IN ('group', 'private')),
	writer TEXT NOT NULL CHECK (writer IN ('server', 'local')),
	writer_epoch INTEGER NOT NULL DEFAULT 0,
	mirror_generation INTEGER NOT NULL DEFAULT 0,
	last_mirror_at INTEGER,
	duplicate_collapse INTEGER NOT NULL DEFAULT 0 CHECK (duplicate_collapse IN (0, 1)),
	layout_template TEXT NOT NULL DEFAULT '',
	tagging_policy TEXT NOT NULL DEFAULT '',
	admission_format_allowlist TEXT,
	admission_match_threshold REAL
) STRICT;

CREATE TABLE albums (
	id BLOB NOT NULL PRIMARY KEY,
	collection_id BLOB NOT NULL REFERENCES collections (id),
	mb_release_id TEXT,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL
) STRICT;

CREATE UNIQUE INDEX albums_id_collection ON albums (id, collection_id);

CREATE TABLE tracks (
	id BLOB NOT NULL PRIMARY KEY,
	album_id BLOB NOT NULL,
	collection_id BLOB NOT NULL,
	rel_path BLOB NOT NULL CHECK (
		length(rel_path) > 0
		AND CAST(substr(rel_path, 1, 1) AS BLOB) != X'2F'
		AND instr(rel_path, X'00') = 0
		AND instr(CAST(X'2F' || rel_path || X'2F' AS BLOB), X'2F2F') = 0
		AND instr(CAST(X'2F' || rel_path || X'2F' AS BLOB), X'2F2E2F') = 0
		AND instr(CAST(X'2F' || rel_path || X'2F' AS BLOB), X'2F2E2E2F') = 0
	),
	mb_recording_id TEXT,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	FOREIGN KEY (album_id, collection_id) REFERENCES albums (id, collection_id)
) STRICT;

CREATE UNIQUE INDEX tracks_collection_relpath ON tracks (collection_id, rel_path);
