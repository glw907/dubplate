-- A route collection is one the server writes for intake: at most one
-- holds each route. SQLite refuses ADD COLUMN ... UNIQUE, so the
-- uniqueness is a separate index, under which any number of NULL rows
-- coexist.
ALTER TABLE collections ADD COLUMN intake_route TEXT CHECK (intake_route IN ('default', 'unlisted'));

CREATE UNIQUE INDEX collections_intake_route ON collections (intake_route);

-- state and source carry no CHECK: Go validates both against closed
-- vocabularies, so a later rung adds a value with no table rebuild.
-- state is written in the transaction that appends the event setting it.
CREATE TABLE submissions (
	id BLOB NOT NULL PRIMARY KEY,
	uploader BLOB NOT NULL REFERENCES users (id),
	collection_id BLOB NOT NULL REFERENCES collections (id),
	verdict_unit TEXT NOT NULL CHECK (verdict_unit IN ('album', 'track')),
	container TEXT NOT NULL CHECK (container IN ('files', 'zip')),
	state TEXT NOT NULL,
	source TEXT NOT NULL,
	source_note TEXT,
	store_url TEXT CHECK (store_url IS NULL OR store_url LIKE 'https://%'),
	declared_bytes INTEGER NOT NULL CHECK (declared_bytes >= 0),
	reserved_bytes INTEGER NOT NULL CHECK (reserved_bytes >= 0),
	mb_release_id TEXT,
	match_candidate TEXT,
	match_evaluated_at INTEGER,
	entered_artist TEXT,
	entered_album TEXT,
	entered_year INTEGER,
	album_key BLOB CHECK (album_key IS NULL OR length(album_key) = 32),
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL
) STRICT;

-- rel_path repeats collections.root_path's relative-path CHECK
-- (0003_catalog.sql): non-empty, no leading slash, no NUL, and no empty,
-- ".", or ".." component. The CHECK admits a backslash, which the Go
-- path validator refuses. role is derived by the server from the
-- extension. audio_key and header_md5 are MD5 digests, file_hash is
-- SHA-256; header_md5 is NULL when the stream header's value is all
-- zero.
CREATE TABLE submission_files (
	id BLOB NOT NULL PRIMARY KEY,
	submission_id BLOB NOT NULL REFERENCES submissions (id),
	rel_path BLOB NOT NULL CHECK (
		length(rel_path) > 0
		AND CAST(substr(rel_path, 1, 1) AS BLOB) != X'2F'
		AND instr(rel_path, X'00') = 0
		AND instr(CAST(X'2F' || rel_path || X'2F' AS BLOB), X'2F2F') = 0
		AND instr(CAST(X'2F' || rel_path || X'2F' AS BLOB), X'2F2E2F') = 0
		AND instr(CAST(X'2F' || rel_path || X'2F' AS BLOB), X'2F2E2E2F') = 0
	),
	size INTEGER NOT NULL CHECK (size >= 0),
	role TEXT NOT NULL CHECK (role IN ('audio', 'extra', 'archive')),
	received_at INTEGER,
	audio_key BLOB CHECK (audio_key IS NULL OR length(audio_key) = 16),
	header_md5 BLOB CHECK (header_md5 IS NULL OR length(header_md5) = 16),
	file_hash BLOB CHECK (file_hash IS NULL OR length(file_hash) = 32),
	sample_rate INTEGER,
	bits_per_sample INTEGER,
	channels INTEGER,
	samples INTEGER,
	disc INTEGER,
	track INTEGER,
	entered_artist TEXT,
	entered_title TEXT,
	verdict TEXT,
	UNIQUE (submission_id, rel_path)
) STRICT;

-- kind is a journey state or an event-only kind, with no CHECK on
-- move_intents.kind's precedent (0005_routed.sql). seq is the
-- submission's largest seq plus one, written inside the transition's
-- transaction: "latest" reads seq, because at ties under a frozen clock,
-- id is random, and a snapshot's VACUUM INTO may renumber rowids. The
-- unique index also refuses a second writer that slipped past a
-- compare-and-set.
CREATE TABLE submission_events (
	id BLOB NOT NULL PRIMARY KEY,
	submission_id BLOB NOT NULL REFERENCES submissions (id),
	seq INTEGER NOT NULL,
	at INTEGER NOT NULL,
	kind TEXT NOT NULL,
	reason TEXT,
	detail TEXT,
	UNIQUE (submission_id, seq)
) STRICT;

-- A warning never blocks a submission. subject is the member name an
-- extras warning names (such a member gets no row), NULL for a
-- completeness warning.
CREATE TABLE submission_warnings (
	id BLOB NOT NULL PRIMARY KEY,
	submission_id BLOB NOT NULL REFERENCES submissions (id),
	code TEXT NOT NULL,
	subject BLOB,
	at INTEGER NOT NULL
) STRICT;
