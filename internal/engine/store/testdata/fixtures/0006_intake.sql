-- An admin and a default-route collection of this fixture's own (ids
-- a8a8... and b8b8..., distinct from every earlier fixture's): one
-- migration's fixture applied alone
-- (TestSchemaMigrationsCarryPopulatedFixtures) has no earlier fixture's
-- rows to build on, and every fixture also layers onto every earlier one
-- in fullyMigratedStore, so a shared id or unique value would collide.
-- The collection holds intake_route 'default', which the intake schema
-- tests lean on: a second 'default' is refused. The submission's files
-- decode to "album/01 Track.flac" and "album/cover.jpg", and its two
-- events sit at seq 1 and 2.
INSERT INTO users (id, address, display_name, username, admin, tombstoned_at, created_at, updated_at)
VALUES (x'a8a8a8a8a8a8a8a8a8a8a8a8a8a8a8a8', 'admin4@example.invalid', 'Admin Four', 'admin4', 1, NULL, 1700000000000000, 1700000000000000);

INSERT INTO collections (
	id, name, root_path, owner, visibility, writer, writer_epoch, mirror_generation,
	last_mirror_at, duplicate_collapse, layout_template, tagging_policy,
	admission_format_allowlist, admission_match_threshold, intake_route
) VALUES (
	x'b8b8b8b8b8b8b8b8b8b8b8b8b8b8b8b8', 'Intake Fixture Library', x'696e74616b652d6c696272617279',
	x'a8a8a8a8a8a8a8a8a8a8a8a8a8a8a8a8', 'group', 'server', 0, 0,
	NULL, 0, '', '', NULL, NULL, 'default'
);

INSERT INTO submissions (
	id, uploader, collection_id, verdict_unit, container, state, source, source_note, store_url,
	declared_bytes, reserved_bytes, mb_release_id, match_candidate, match_evaluated_at,
	entered_artist, entered_album, entered_year, album_key, created_at, updated_at
) VALUES (
	x'c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8', x'a8a8a8a8a8a8a8a8a8a8a8a8a8a8a8a8',
	x'b8b8b8b8b8b8b8b8b8b8b8b8b8b8b8b8', 'album', 'files', 'checking', 'bandcamp', 'bought it', 'https://example.invalid/album',
	2048, 2048, NULL, NULL, NULL,
	NULL, NULL, NULL, NULL, 1700000000000000, 1700000000000000
);

INSERT INTO submission_files (
	id, submission_id, rel_path, size, role, received_at, audio_key, header_md5, file_hash,
	sample_rate, bits_per_sample, channels, samples, disc, track, entered_artist, entered_title, verdict
) VALUES
(
	x'd8d8d8d8d8d8d8d8d8d8d8d8d8d8d8d8', x'c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8',
	x'616c62756d2f303120547261636b2e666c6163', 1024, 'audio', 1700000000000000,
	x'01010101010101010101010101010101', x'02020202020202020202020202020202',
	x'0303030303030303030303030303030303030303030303030303030303030303',
	44100, 16, 2, 441000, 1, 1, NULL, NULL, NULL
),
(
	x'd9d9d9d9d9d9d9d9d9d9d9d9d9d9d9d9', x'c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8',
	x'616c62756d2f636f7665722e6a7067', 1024, 'extra', 1700000000000000,
	NULL, NULL, NULL,
	NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL
);

INSERT INTO submission_events (id, submission_id, seq, at, kind, reason, detail) VALUES
(x'e6e6e6e6e6e6e6e6e6e6e6e6e6e6e6e6', x'c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8', 1, 1700000000000000, 'uploading', NULL, NULL),
(x'e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7', x'c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8', 2, 1700000000000001, 'checking', NULL, NULL);

INSERT INTO submission_warnings (id, submission_id, code, subject, at)
VALUES (x'f6f6f6f6f6f6f6f6f6f6f6f6f6f6f6f6', x'c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8', 'no-cover-art', NULL, 1700000000000001);
