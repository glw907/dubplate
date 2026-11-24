-- Every stored enum value becomes snake_case (record 0017), and each
-- foreign key to users is named user_id. respelled maps each retired
-- spelling to its new one; a column is rewritten only where it holds a
-- retired spelling exactly, so submission_events.reason's free text for
-- an expiry is left alone.
CREATE TEMP TABLE respelled (
	old TEXT NOT NULL PRIMARY KEY,
	new TEXT NOT NULL
) STRICT;

INSERT INTO respelled (old, new) VALUES
	('env-seed', 'env_seed'),
	('command-line', 'command_line'),
	('check-unfinished', 'check_unfinished'),
	('free-download', 'free_download'),
	('own-recording', 'own_recording'),
	('bad-extension', 'bad_extension'),
	('below-cd-quality', 'below_cd_quality'),
	('too-large', 'too_large'),
	('unpack-failed', 'unpack_failed'),
	('unsafe-path', 'unsafe_path'),
	('no-space', 'no_space'),
	('too-many-sessions', 'too_many_sessions'),
	('missing-source', 'missing_source'),
	('uploads-off', 'uploads_off'),
	('over-quota', 'over_quota'),
	('too-many-files', 'too_many_files'),
	('path-too-deep', 'path_too_deep'),
	('no-audio', 'no_audio'),
	('tags-too-large', 'tags_too_large'),
	('picture-too-large', 'picture_too_large'),
	('unsupported-flac', 'unsupported_flac'),
	('source-damaged', 'source_damaged'),
	('empty-track', 'empty_track'),
	('missing-track-number', 'missing_track_number'),
	('duplicate-track-position', 'duplicate_track_position'),
	('incomplete-album', 'incomplete_album'),
	('already-in-library', 'already_in_library'),
	('could-not-check', 'could_not_check'),
	('id3-tags', 'id3_tags'),
	('no-cover-art', 'no_cover_art'),
	('track-total-differs', 'track_total_differs'),
	('extras-dropped', 'extras_dropped'),
	('job-error', 'job_error'),
	('job-invalid', 'job_invalid'),
	('rclone-missing', 'rclone_missing'),
	('remote-invalid', 'remote_invalid'),
	('snapshot-failed', 'snapshot_failed'),
	('reference-failed', 'reference_failed'),
	('library-missing', 'library_missing'),
	('library-invalid', 'library_invalid'),
	('library-unreadable', 'library_unreadable'),
	('library-empty', 'library_empty'),
	('library-short', 'library_short');

UPDATE settings_audit SET channel = (SELECT new FROM respelled WHERE old = channel)
	WHERE channel IN (SELECT old FROM respelled);
UPDATE submissions SET state = (SELECT new FROM respelled WHERE old = state)
	WHERE state IN (SELECT old FROM respelled);
UPDATE submissions SET source = (SELECT new FROM respelled WHERE old = source)
	WHERE source IN (SELECT old FROM respelled);
UPDATE submission_files SET verdict = (SELECT new FROM respelled WHERE old = verdict)
	WHERE verdict IN (SELECT old FROM respelled);
UPDATE submission_events SET kind = (SELECT new FROM respelled WHERE old = kind)
	WHERE kind IN (SELECT old FROM respelled);
UPDATE submission_events SET reason = (SELECT new FROM respelled WHERE old = reason)
	WHERE reason IN (SELECT old FROM respelled);
UPDATE submission_warnings SET code = (SELECT new FROM respelled WHERE old = code)
	WHERE code IN (SELECT old FROM respelled);
UPDATE jobs SET failure_code = (SELECT new FROM respelled WHERE old = failure_code)
	WHERE failure_code IN (SELECT old FROM respelled);

-- A backup that rclone failed records rclone-exit-<status> or
-- rclone-signal-<name>, one code per status or signal, so a pattern
-- rewrites those rather than the map. A signal's name was kill(1)'s
-- uppercase spelling and is now lowercase.
UPDATE jobs SET failure_code = replace(failure_code, '-', '_')
	WHERE failure_code GLOB 'rclone-exit-[0-9]*';
UPDATE jobs SET failure_code = lower(replace(failure_code, '-', '_'))
	WHERE failure_code GLOB 'rclone-signal-[0-9A-Z]*';

-- settings.source carries a CHECK naming the old spelling, so the table
-- is rebuilt. No table references settings, so the rebuild needs no
-- foreign-key pause.
ALTER TABLE settings RENAME TO settings_retired;

CREATE TABLE settings (
	key TEXT NOT NULL PRIMARY KEY,
	value TEXT NOT NULL,
	source TEXT NOT NULL CHECK (source IN ('env_seed', 'admin')),
	seed_value TEXT,
	actor TEXT NOT NULL,
	updated_at INTEGER NOT NULL
) STRICT;

INSERT INTO settings (key, value, source, seed_value, actor, updated_at)
	SELECT key, value, coalesce((SELECT new FROM respelled WHERE old = source), source), seed_value, actor, updated_at
	FROM settings_retired;

DROP TABLE settings_retired;

DROP TABLE respelled;

ALTER TABLE collections RENAME COLUMN owner TO user_id;

ALTER TABLE submissions RENAME COLUMN uploader TO user_id;
