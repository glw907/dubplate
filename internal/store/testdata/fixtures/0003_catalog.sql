-- root_path and rel_path are BLOB columns on a STRICT table, which
-- refuses a TEXT literal even by value, so their hex literals below
-- decode to "group/music" and "Member One/Group Library/01 Track.flac".

INSERT INTO users (id, address, display_name, username, admin, tombstoned_at, created_at, updated_at)
VALUES (x'a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1', 'member1@example.invalid', 'Member One', 'member1', 1, NULL, 1700000000000000, 1700000000000000);

INSERT INTO collections (
	id, name, root_path, owner, visibility, writer, writer_epoch, mirror_generation,
	last_mirror_at, duplicate_collapse, layout_template, tagging_policy,
	admission_format_allowlist, admission_match_threshold
) VALUES (
	x'b1b1b1b1b1b1b1b1b1b1b1b1b1b1b1b1', 'Group Library', x'67726f75702f6d75736963',
	x'a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1', 'group', 'server', 0, 0,
	NULL, 0, '', '', NULL, NULL
);

INSERT INTO albums (id, collection_id, mb_release_id, created_at, updated_at)
VALUES (
	x'c1c1c1c1c1c1c1c1c1c1c1c1c1c1c1c1', x'b1b1b1b1b1b1b1b1b1b1b1b1b1b1b1b1',
	'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', 1700000000000000, 1700000000000000
);

INSERT INTO tracks (id, album_id, collection_id, rel_path, mb_recording_id, created_at, updated_at)
VALUES (
	x'd1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1', x'c1c1c1c1c1c1c1c1c1c1c1c1c1c1c1c1', x'b1b1b1b1b1b1b1b1b1b1b1b1b1b1b1b1',
	x'4d656d626572204f6e652f47726f7570204c6962726172792f303120547261636b2e666c6163', NULL, 1700000000000000, 1700000000000000
);
