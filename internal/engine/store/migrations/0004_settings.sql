CREATE TABLE settings (
	key TEXT NOT NULL PRIMARY KEY,
	value TEXT NOT NULL,
	source TEXT NOT NULL CHECK (source IN ('env-seed', 'admin')),
	seed_value TEXT,
	actor TEXT NOT NULL,
	updated_at INTEGER NOT NULL
) STRICT;

CREATE TABLE settings_audit (
	id BLOB NOT NULL PRIMARY KEY,
	at INTEGER NOT NULL,
	kind TEXT NOT NULL,
	actor TEXT NOT NULL,
	channel TEXT NOT NULL,
	key TEXT NOT NULL,
	old_value TEXT,
	new_value TEXT
) STRICT;

-- settings_generation holds one row, never more: singleton is neither
-- named "id" nor ending "_id" so invariant 6's INTEGER PRIMARY KEY rule
-- does not fire on it, and its CHECK pins the only value a second row
-- may carry, so an insert of any other key is a CHECK refusal and a
-- repeat of the pinned key is a PRIMARY KEY refusal.
CREATE TABLE settings_generation (
	singleton INTEGER NOT NULL PRIMARY KEY CHECK (singleton = 1),
	generation INTEGER NOT NULL
) STRICT;

INSERT INTO settings_generation (singleton, generation) VALUES (1, 0);
