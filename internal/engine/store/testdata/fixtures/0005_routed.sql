-- A user and a collection of this fixture's own (ids a3a3... and
-- b7b7..., distinct from 0003's a1a1... and 0004's a6a6...): one
-- migration's fixture applied alone (TestSchemaMigrationsCarryPopulatedFixtures)
-- has no earlier fixture's rows to build on, and every fixture also
-- layers onto every earlier one in fullyMigratedStore, so a shared id or
-- a shared unique value here would collide with an earlier fixture's
-- row there. old_path and new_path decode to "Member One/Old
-- Location/01 Track.flac" and "Member One/New Location/01 Track.flac".
INSERT INTO users (id, address, display_name, streamer_username, admin, tombstoned_at, created_at, updated_at)
VALUES (x'a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7', 'member3@example.invalid', 'Member Three', 'member3', 0, NULL, 1700000000000000, 1700000000000000);

INSERT INTO collections (
	id, name, root_path, owner, visibility, writer, writer_epoch, mirror_generation,
	last_mirror_at, duplicate_collapse, layout_template, tagging_policy,
	admission_format_allowlist, admission_match_threshold
) VALUES (
	x'b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7', 'Routed Fixture Library', x'726f757465642f6d75736963',
	x'a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7', 'group', 'server', 0, 0,
	NULL, 0, '', '', NULL, NULL
);

INSERT INTO member_overrides (user_id, quota_bytes, may_upload, may_create_collections, upload_concurrency_cap, updated_at)
VALUES (x'a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7', 5000000000, 1, 0, 4, 1700000000000000);

INSERT INTO member_preferences (user_id, collapse_duplicates, updated_at)
VALUES (x'a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7', 1, 1700000000000000);

INSERT INTO member_quota (user_id, charged_bytes, updated_at)
VALUES (x'a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7', 123456, 1700000000000000);

INSERT INTO device_credentials (id, credential_hash, user_id, device_id, created_at, last_seen_address, expires_at, revoked_at)
VALUES (
	x'e4e4e4e4e4e4e4e4e4e4e4e4e4e4e4e4', x'11111111111111111111111111111111111111111111111111111111111111',
	x'a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7', x'22222222222222222222222222222222',
	1700000000000000, '203.0.113.5', 1800000000000000, NULL
);

INSERT INTO move_intents (id, at, kind, old_path, new_path, granularity, state)
VALUES (
	x'e5e5e5e5e5e5e5e5e5e5e5e5e5e5e5e5', 1700000000000000, 'rename',
	x'4d656d626572204f6e652f4f6c64204c6f636174696f6e2f303120547261636b2e666c6163',
	x'4d656d626572204f6e652f4e6577204c6f636174696f6e2f303120547261636b2e666c6163',
	'file', 'pending'
);
