package store

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/reject"
)

// respellSeed holds, at schema 6, one row for every kebab-case spelling
// a store column can carry, with ids distinct from every fixture's. The
// expiry's reason and the job's message are free text, which keeps its
// hyphens.
const respellSeed = `
INSERT INTO users (id, address, display_name, username, admin, created_at, updated_at)
VALUES (x'71717171717171717171717171717171', 'respell@example.invalid', 'Respell', 'respell', 0, 1700000000000000, 1700000000000000);

INSERT INTO collections (id, name, root_path, owner, visibility, writer)
VALUES (x'72727272727272727272727272727272', 'Respell Library', x'72657370656c6c', x'71717171717171717171717171717171', 'group', 'server');

INSERT INTO settings (key, value, source, seed_value, actor, updated_at) VALUES
	('DUBPLATE_RESPELL_SEEDED', 'a', 'env-seed', 'a', 'system', 1700000000000000),
	('DUBPLATE_RESPELL_WRITTEN', 'b', 'admin', 'a', 'respell@example.invalid', 1700000000000000);

INSERT INTO settings_audit (id, at, kind, actor, channel, key, old_value, new_value) VALUES
	(x'73737373737373737373737373737301', 1700000000000000, 'seed', 'system', 'env-seed', 'DUBPLATE_RESPELL_SEEDED', NULL, 'a'),
	(x'73737373737373737373737373737302', 1700000000000001, 'admin-write', 'respell@example.invalid', 'command-line', 'DUBPLATE_RESPELL_WRITTEN', 'a', 'b'),
	(x'73737373737373737373737373737303', 1700000000000002, 'admin-write', 'respell@example.invalid', 'web', 'DUBPLATE_RESPELL_WRITTEN', 'a', 'b');

INSERT INTO jobs (id, kind, at, state, owner_instance, actor, failure_code, failure_message, started_at, updated_at, finished_at) VALUES
	(x'74747474747474747474747474747401', 'backup', 1700000000000000, 'failed', 'box', 'system', 'rclone-exit-3', 'rclone exited 3: check-the-remote', 1700000000000000, 1700000000000000, 1700000000000000),
	(x'74747474747474747474747474747402', 'backup', 1700000000000001, 'failed', 'box', 'system', 'rclone-exit-130', 'rclone exited 130', 1700000000000000, 1700000000000000, 1700000000000000),
	(x'74747474747474747474747474747403', 'backup', 1700000000000002, 'failed', 'box', 'system', 'library-short', 'too few files', 1700000000000000, 1700000000000000, 1700000000000000),
	(x'74747474747474747474747474747404', 'backup', 1700000000000003, 'failed', 'box', 'system', 'job-error', 'the job failed', 1700000000000000, 1700000000000000, 1700000000000000),
	(x'74747474747474747474747474747405', 'backup', 1700000000000004, 'failed', 'box', 'system', 'rclone-signal-KILL', 'rclone killed', 1700000000000000, 1700000000000000, 1700000000000000),
	(x'74747474747474747474747474747406', 'backup', 1700000000000005, 'failed', 'box', 'system', 'rclone-signal-34', 'rclone signaled', 1700000000000000, 1700000000000000, 1700000000000000);

INSERT INTO submissions (id, uploader, collection_id, verdict_unit, container, state, source, declared_bytes, reserved_bytes, created_at, updated_at) VALUES
	(x'75757575757575757575757575757501', x'71717171717171717171717171717171', x'72727272727272727272727272727272', 'album', 'files', 'check-unfinished', 'free-download', 2048, 2048, 1700000000000000, 1700000000000000),
	(x'75757575757575757575757575757502', x'71717171717171717171717171717171', x'72727272727272727272727272727272', 'track', 'files', 'rejected', 'own-recording', 1024, 0, 1700000000000000, 1700000000000000);

INSERT INTO submission_files (id, submission_id, rel_path, size, role, verdict) VALUES
	(x'76767676767676767676767676767601', x'75757575757575757575757575757501', x'612e666c6163', 1024, 'audio', NULL),
	(x'76767676767676767676767676767602', x'75757575757575757575757575757502', x'612e6d7033', 1024, 'audio', 'below-cd-quality'),
	(x'76767676767676767676767676767603', x'75757575757575757575757575757502', x'622e666c6163', 1024, 'audio', 'missing-track-number');

INSERT INTO submission_events (id, submission_id, seq, at, kind, reason, detail) VALUES
	(x'77777777777777777777777777777701', x'75757575757575757575757575757501', 1, 1700000000000000, 'uploading', NULL, NULL),
	(x'77777777777777777777777777777702', x'75757575757575757575757575757501', 2, 1700000000000001, 'check-unfinished', 'could-not-check', NULL),
	(x'77777777777777777777777777777703', x'75757575757575757575757575757502', 1, 1700000000000000, 'rejected', 'already-in-library', NULL),
	(x'77777777777777777777777777777704', x'75757575757575757575757575757502', 2, 1700000000000001, 'expired', 'staged file went-missing', NULL);

INSERT INTO submission_warnings (id, submission_id, code, subject, at) VALUES
	(x'78787878787878787878787878787801', x'75757575757575757575757575757501', 'id3-tags', NULL, 1700000000000001),
	(x'78787878787878787878787878787802', x'75757575757575757575757575757501', 'track-total-differs', NULL, 1700000000000001);
`

// respelled are the queries a store answers once it migrates from
// schema 6 holding respellSeed, each with the literal rows it must
// read back.
var respelled = []struct {
	query string
	want  []string
}{
	{"SELECT key || ' ' || source FROM settings WHERE key LIKE 'DUBPLATE_RESPELL_%' ORDER BY key", []string{
		"DUBPLATE_RESPELL_SEEDED env_seed",
		"DUBPLATE_RESPELL_WRITTEN admin",
	}},
	{"SELECT kind || ' ' || channel FROM settings_audit WHERE key LIKE 'DUBPLATE_RESPELL_%' ORDER BY id", []string{
		"seed env_seed",
		"admin-write command_line",
		"admin-write web",
	}},
	{"SELECT failure_code || ' / ' || failure_message FROM jobs WHERE id >= x'74747474747474747474747474747401' AND id <= x'74747474747474747474747474747406' ORDER BY id", []string{
		"rclone_exit_3 / rclone exited 3: check-the-remote",
		"rclone_exit_130 / rclone exited 130",
		"library_short / too few files",
		"job_error / the job failed",
		"rclone_signal_kill / rclone killed",
		"rclone_signal_34 / rclone signaled",
	}},
	{"SELECT state || ' ' || source FROM submissions WHERE user_id = x'71717171717171717171717171717171' ORDER BY id", []string{
		"check_unfinished free_download",
		"rejected own_recording",
	}},
	{"SELECT coalesce(verdict, '-') FROM submission_files WHERE id >= x'76767676767676767676767676767601' AND id <= x'76767676767676767676767676767603' ORDER BY id", []string{
		"-",
		"below_cd_quality",
		"missing_track_number",
	}},
	{"SELECT kind || ' ' || coalesce(reason, '-') FROM submission_events WHERE id >= x'77777777777777777777777777777701' AND id <= x'77777777777777777777777777777704' ORDER BY id", []string{
		"uploading -",
		"check_unfinished could_not_check",
		"rejected already_in_library",
		"expired staged file went-missing",
	}},
	{"SELECT code FROM submission_warnings WHERE id >= x'78787878787878787878787878787801' AND id <= x'78787878787878787878787878787802' ORDER BY id", []string{
		"id3_tags",
		"track_total_differs",
	}},
	{"SELECT name FROM collections WHERE user_id = x'71717171717171717171717171717171'", []string{"Respell Library"}},
}

// TestMigrationRespellsStoredEnums migrates a store from schema 6,
// holding every kebab-case spelling a column can carry, to head. Each
// spelling reads back in snake_case, free text keeps its hyphens, the
// owner and uploader columns answer as user_id, every foreign key
// holds, and the settings source check refuses the old spelling.
func TestMigrationRespellsStoredEnums(t *testing.T) {
	l, h := newMigrateStore(t)
	ctx := context.Background()
	if err := migrate(ctx, h, l, readMigrationsT(t, migrationSourceThrough(t, 6)), defaultOpts()); err != nil {
		t.Fatalf("migrate to version 6: %v", err)
	}
	if _, err := h.write.ExecContext(ctx, respellSeed); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := migrate(ctx, h, l, readMigrationsT(t, migrationSource()), defaultOpts()); err != nil {
		t.Fatalf("migrate to head: %v", err)
	}

	var version int
	if err := h.write.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 7 {
		t.Errorf("user_version = %d, want 7", version)
	}
	for _, r := range respelled {
		if got := queryStrings(t, h.write, r.query); !slices.Equal(got, r.want) {
			t.Errorf("%s = %q, want %q", r.query, got, r.want)
		}
	}
	if broken := queryStrings(t, h.write, "SELECT \"table\" || ' ' || parent FROM pragma_foreign_key_check"); len(broken) > 0 {
		t.Errorf("foreign_key_check = %q, want no rows", broken)
	}

	_, err := h.write.ExecContext(ctx,
		"INSERT INTO settings (key, value, source, seed_value, actor, updated_at) VALUES ('DUBPLATE_RESPELL_OLD', 'x', 'env-seed', 'x', 'system', 1)")
	if code := sqliteCode(err); code != sqliteCheckCode {
		t.Errorf("insert source 'env-seed': code %d (%v), want %d", code, err, sqliteCheckCode)
	}
	if _, err := h.write.ExecContext(ctx,
		"INSERT INTO settings (key, value, source, seed_value, actor, updated_at) VALUES ('DUBPLATE_RESPELL_NEW', 'x', 'env_seed', 'x', 'system', 1)"); err != nil {
		t.Errorf("insert source 'env_seed': %v", err)
	}
}

// boxCatalogSum is the sha256 of 0003_catalog.sql as it stood at the
// box's commit, d1e93e9, before the users column was renamed in place.
// It is typed here rather than read from amendments, so the tests below
// cannot drift with the table they check.
const boxCatalogSum = "f962133f7366b5e941e2346fc1d2e3cd4defa743ec02902241408e9173a089c5"

// catalogSum is the sha256 of 0003_catalog.sql as this binary embeds
// it, from sha256sum.
const catalogSum = "42834b8f62df9727bea686bbc02030f1b9af0de2a6c61a59050be72399c7b7fd"

// boxStore returns a store at schema 5 in the shape the box's commit
// left: users carries streamer_username, and the ledger's version 3 row
// records the catalog file's sum from that commit.
func boxStore(t *testing.T) (leaf, *Handle) {
	t.Helper()
	l, h := newMigrateStore(t)
	ctx := context.Background()
	if err := migrate(ctx, h, l, readMigrationsT(t, migrationSourceThrough(t, 5)), defaultOpts()); err != nil {
		t.Fatalf("migrate to version 5: %v", err)
	}
	for _, q := range []string{
		"ALTER TABLE users RENAME COLUMN username TO streamer_username",
		"UPDATE " + ledgerTable + " SET sha256 = '" + boxCatalogSum + "' WHERE version = 3",
		"INSERT INTO users (id, address, display_name, streamer_username, admin, created_at, updated_at) VALUES (x'79797979797979797979797979797979', 'box@example.invalid', 'Box', 'boxer', 1, 1, 1)",
	} {
		if _, err := h.write.ExecContext(ctx, q); err != nil {
			t.Fatalf("%s: %v", strings.Fields(q)[0], err)
		}
	}
	return l, h
}

// TestMigrateAmendsTheBoxCatalog migrates a store the box's commit left
// to head: it migrates, its member keeps the username, and its schema
// matches a fresh store's, so the amendment leaves nothing a later
// migration could trip on.
func TestMigrateAmendsTheBoxCatalog(t *testing.T) {
	l, h := boxStore(t)
	ctx := context.Background()
	files := readMigrationsT(t, migrationSource())
	if err := migrate(ctx, h, l, files, defaultOpts()); err != nil {
		t.Fatalf("migrate the box store to head: %v", err)
	}
	if got := queryStrings(t, h.write, "SELECT username FROM users WHERE id = x'79797979797979797979797979797979'"); !slices.Equal(got, []string{"boxer"}) {
		t.Errorf("username = %q, want [boxer]", got)
	}
	if got := queryStrings(t, h.write, "SELECT sha256 FROM "+ledgerTable+" WHERE version = 3"); !slices.Equal(got, []string{catalogSum}) {
		t.Errorf("ledger version 3 sum = %q, want [%s]", got, catalogSum)
	}

	fl, fresh := newMigrateStore(t)
	if err := migrate(ctx, fresh, fl, files, defaultOpts()); err != nil {
		t.Fatalf("migrate a fresh store: %v", err)
	}
	const schema = "SELECT type || ' ' || name || ' ' || coalesce(sql, '') FROM sqlite_schema ORDER BY type, name"
	got, want := queryStrings(t, h.write, schema), queryStrings(t, fresh.write, schema)
	if !slices.Equal(got, want) {
		t.Errorf("amended schema differs from a fresh store's:\n got %q\nwant %q", got, want)
	}
}

// TestCheckLedgerAcceptsTheBoxCatalogOnlyWithMigrationsPending holds
// the amendment to stores a migration will reach: the box catalog's
// sum passes the ledger check below head and is refused at head, where
// no migration would run to amend it.
func TestCheckLedgerAcceptsTheBoxCatalogOnlyWithMigrationsPending(t *testing.T) {
	_, h := boxStore(t)
	ctx := context.Background()
	files := readMigrationsT(t, migrationSource())
	if err := checkLedger(ctx, h.write, files, 5); err != nil {
		t.Errorf("checkLedger at 5 of %d: %v, want nil", len(files), err)
	}
	err := checkLedger(ctx, h.write, files[:5], 5)
	if err == nil || !strings.Contains(err.Error(), boxCatalogSum) {
		t.Errorf("checkLedger at head 5: %v, want the box catalog's sum refused", err)
	}
}

// backupCodes are the failure codes internal/backup and internal/ops/jobs
// record, typed here because a store test cannot import either package,
// which import the store. rclone's per-status and per-signal codes are
// respellSeed's, since 0007 rewrites them by pattern.
var backupCodes = []string{
	"job_error", "job_invalid", "rclone_missing", "remote_invalid", "snapshot_failed", "reference_failed",
	"library_missing", "library_invalid", "library_unreadable", "library_empty", "library_short",
}

// TestMigrationRespellsEveryCode seeds, at schema 6, every reject code
// in each column that stores one (an event's reason, a file's verdict,
// and a warning's code) and every backup failure code in a job, each
// in its retired hyphenated spelling. After the migration each column
// reads back the code the binary writes. The codes come from the
// reject set and the backup list, never from 0007's map, so a code the
// map leaves out fails here.
func TestMigrationRespellsEveryCode(t *testing.T) {
	l, h := newMigrateStore(t)
	ctx := context.Background()
	if err := migrate(ctx, h, l, readMigrationsT(t, migrationSourceThrough(t, 6)), defaultOpts()); err != nil {
		t.Fatalf("migrate to version 6: %v", err)
	}
	if _, err := h.write.ExecContext(ctx, respellSeed); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var codes []string
	for _, r := range reject.All() {
		codes = append(codes, string(r.Code))
	}
	// Fixture state: the set is the 29 reasons, so a set that came back
	// empty or short would seed too little to prove the map whole.
	if len(codes) != 29 {
		t.Fatalf("reject.All holds %d codes, want 29", len(codes))
	}
	const submission = "x'75757575757575757575757575757501'"
	for i, code := range codes {
		retired := strings.ReplaceAll(code, "_", "-")
		for _, q := range []string{
			"INSERT INTO submission_events (id, submission_id, seq, at, kind, reason) VALUES (randomblob(16), " + submission + ", ?1, 1, 'rejected', ?2)",
			"INSERT INTO submission_files (id, submission_id, rel_path, size, role, verdict) VALUES (randomblob(16), " + submission + ", CAST('code' || ?1 || '.flac' AS BLOB), ?1, 'audio', ?2)",
			"INSERT INTO submission_warnings (id, submission_id, code, at) VALUES (randomblob(16), " + submission + ", ?2, ?1)",
		} {
			if _, err := h.write.ExecContext(ctx, q, 5000+i, retired); err != nil {
				t.Fatalf("seed %s into %s: %v", retired, strings.Fields(q)[2], err)
			}
		}
	}
	for i, code := range backupCodes {
		if _, err := h.write.ExecContext(ctx,
			"INSERT INTO jobs (id, kind, at, state, owner_instance, actor, failure_code, updated_at) VALUES (randomblob(16), 'backup', ?, 'failed', 'box', 'system', ?, 1)",
			5000+i, strings.ReplaceAll(code, "_", "-")); err != nil {
			t.Fatalf("seed job %s: %v", code, err)
		}
	}

	if err := migrate(ctx, h, l, readMigrationsT(t, migrationSource()), defaultOpts()); err != nil {
		t.Fatalf("migrate to head: %v", err)
	}
	for _, c := range []struct {
		query string
		want  []string
	}{
		// respellSeed's own rows sit below 5000 in seq and size and far
		// above it in at.
		{"SELECT reason FROM submission_events WHERE seq >= 5000 ORDER BY seq", codes},
		{"SELECT verdict FROM submission_files WHERE size >= 5000 ORDER BY size", codes},
		{"SELECT code FROM submission_warnings WHERE at < 1000000 ORDER BY at", codes},
		{"SELECT failure_code FROM jobs WHERE at < 1000000 ORDER BY at", backupCodes},
	} {
		if got := queryStrings(t, h.write, c.query); !slices.Equal(got, c.want) {
			t.Errorf("%s = %q, want %q", c.query, got, c.want)
		}
	}
}
