-- A member, a collection it owns, and one submission it uploaded, in
-- the schema 0007 leaves (ids a9a9... through f9f9..., distinct from
-- every earlier fixture's): collections and submissions name the member
-- as user_id, and every enum value is snake_case. The setting is seeded
-- under the new env_seed spelling the source check now admits.
INSERT INTO users (id, address, display_name, username, admin, tombstoned_at, created_at, updated_at)
VALUES (x'a9a9a9a9a9a9a9a9a9a9a9a9a9a9a9a9', 'member9@example.invalid', 'Member Nine', 'member9', 0, NULL, 1700000000000000, 1700000000000000);

INSERT INTO collections (id, name, root_path, user_id, visibility, writer)
VALUES (x'b9b9b9b9b9b9b9b9b9b9b9b9b9b9b9b9', 'Snake Case Library', x'736e616b652d6c696272617279', x'a9a9a9a9a9a9a9a9a9a9a9a9a9a9a9a9', 'group', 'server');

INSERT INTO submissions (
	id, user_id, collection_id, verdict_unit, container, state, source,
	declared_bytes, reserved_bytes, created_at, updated_at
) VALUES (
	x'c9c9c9c9c9c9c9c9c9c9c9c9c9c9c9c9', x'a9a9a9a9a9a9a9a9a9a9a9a9a9a9a9a9', x'b9b9b9b9b9b9b9b9b9b9b9b9b9b9b9b9',
	'track', 'files', 'check_unfinished', 'free_download', 1024, 1024, 1700000000000000, 1700000000000000
);

INSERT INTO submission_files (id, submission_id, rel_path, size, role, verdict)
VALUES (x'd7d7d7d7d7d7d7d7d7d7d7d7d7d7d7d7', x'c9c9c9c9c9c9c9c9c9c9c9c9c9c9c9c9', x'30312e666c6163', 1024, 'audio', 'below_cd_quality');

INSERT INTO submission_events (id, submission_id, seq, at, kind, reason, detail)
VALUES (x'e9e9e9e9e9e9e9e9e9e9e9e9e9e9e9e9', x'c9c9c9c9c9c9c9c9c9c9c9c9c9c9c9c9', 1, 1700000000000000, 'check_unfinished', 'could_not_check', NULL);

INSERT INTO submission_warnings (id, submission_id, code, subject, at)
VALUES (x'f9f9f9f9f9f9f9f9f9f9f9f9f9f9f9f9', x'c9c9c9c9c9c9c9c9c9c9c9c9c9c9c9c9', 'no_cover_art', NULL, 1700000000000000);

INSERT INTO settings (key, value, source, seed_value, actor, updated_at)
VALUES ('DUBPLATE_FIXTURE_SNAKE_CASE', 'on', 'env_seed', 'on', 'system', 1700000000000000);
