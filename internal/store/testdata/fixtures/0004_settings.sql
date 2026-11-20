-- A user and a collection of this fixture's own (ids a6a6... and
-- b6b6..., distinct from 0003's a1a1... and b1b1...) stand alongside the
-- settings rows below, even though no rung 6c table references either:
-- one migration's fixture applied alone (TestSchemaMigrationsCarryPopulatedFixtures)
-- has no earlier fixture's rows to build on, and every fixture also
-- layers onto every earlier one in fullyMigratedStore, so a shared id or
-- a shared unique value here would collide with 0003's row there.
INSERT INTO users (id, address, display_name, username, admin, tombstoned_at, created_at, updated_at)
VALUES (x'a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6', 'member2@example.invalid', 'Member Two', 'member2', 0, NULL, 1700000000000000, 1700000000000000);

INSERT INTO collections (
	id, name, root_path, owner, visibility, writer, writer_epoch, mirror_generation,
	last_mirror_at, duplicate_collapse, layout_template, tagging_policy,
	admission_format_allowlist, admission_match_threshold
) VALUES (
	x'b6b6b6b6b6b6b6b6b6b6b6b6b6b6b6b6', 'Settings Fixture Library', x'73657474696e67732f6d75736963',
	x'a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6a6', 'group', 'server', 0, 0,
	NULL, 0, '', '', NULL, NULL
);

INSERT INTO settings (key, value, source, seed_value, actor, updated_at) VALUES
	('DUBPLATE_JOB_RETENTION', '2160h', 'env-seed', '2160h', 'system', 1700000000000000),
	('DUBPLATE_BACKUP_SNAPSHOT_KEEP', '20', 'admin', '14', 'member2@example.invalid', 1700000001000000);

INSERT INTO settings_audit (id, at, kind, actor, channel, key, old_value, new_value) VALUES
	(x'f1f1f1f1f1f1f1f1f1f1f1f1f1f1f1f1', 1700000000000000, 'seed', 'system', 'env-seed', 'DUBPLATE_JOB_RETENTION', NULL, '2160h'),
	(x'f2f2f2f2f2f2f2f2f2f2f2f2f2f2f2f2', 1700000001000000, 'admin-write', 'member2@example.invalid', 'admin', 'DUBPLATE_BACKUP_SNAPSHOT_KEEP', '14', '20');

UPDATE settings_generation SET generation = 3 WHERE singleton = 1;
