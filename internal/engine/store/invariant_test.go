package store

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
)

// mbForbiddenColumns names the MusicBrainz-shaped columns invariant 4
// refuses outside a foreign key: a generic id column that mixed
// sources, the class of beets bug #604 (spec EVO:575-578).
var mbForbiddenColumns = map[string]bool{
	"release_id": true, "recording_id": true, "artist_id": true,
	"release_group_id": true, "track_id": true, "work_id": true,
}

// invariant6Exempt names every table invariant 6 does not require a
// BLOB primary key named id from. At 6a it holds the ledger alone,
// keyed by version; 6c adds its own singleton and keyed tables.
var invariant6Exempt = map[string]bool{
	ledgerTable: true,
}

// colInfo is one row of pragma_table_info, the columns invariant
// checking reads.
type colInfo struct {
	name    string
	typ     string
	notNull bool
	pk      int
}

// fkInfo is one row of pragma_foreign_key_list, grouped by id for a
// composite key.
type fkInfo struct {
	id       int
	table    string
	from     string
	to       string
	onDelete string
}

// idxInfo is one row of pragma_index_list.
type idxInfo struct {
	name   string
	unique bool
}

// catalogTables lists every user table db holds, sqlite_% tables
// excluded.
func catalogTables(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT name FROM pragma_table_list WHERE schema = 'main' AND type = 'table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("list tables: %w", err)
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// tableIsStrict reports whether table is declared STRICT.
func tableIsStrict(ctx context.Context, db *sql.DB, table string) (bool, error) {
	var strict int
	err := db.QueryRowContext(ctx,
		`SELECT strict FROM pragma_table_list WHERE schema = 'main' AND name = ?`, table,
	).Scan(&strict)
	if err != nil {
		return false, fmt.Errorf("strict flag for %s: %w", table, err)
	}
	return strict != 0, nil
}

// tableColumns reads table's columns through pragma_table_info.
func tableColumns(ctx context.Context, db *sql.DB, table string) ([]colInfo, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`SELECT name, type, "notnull", pk FROM pragma_table_info(%q)`, table))
	if err != nil {
		return nil, fmt.Errorf("columns of %s: %w", table, err)
	}
	defer func() { _ = rows.Close() }()
	var out []colInfo
	for rows.Next() {
		var c colInfo
		var notNull int
		if err := rows.Scan(&c.name, &c.typ, &notNull, &c.pk); err != nil {
			return nil, fmt.Errorf("columns of %s: %w", table, err)
		}
		c.notNull = notNull != 0
		out = append(out, c)
	}
	return out, rows.Err()
}

// tableForeignKeys reads table's foreign keys through
// pragma_foreign_key_list, one row per referencing column, sharing id
// across a composite key's columns.
func tableForeignKeys(ctx context.Context, db *sql.DB, table string) ([]fkInfo, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`SELECT id, "table", "from", "to", on_delete FROM pragma_foreign_key_list(%q)`, table))
	if err != nil {
		return nil, fmt.Errorf("foreign keys of %s: %w", table, err)
	}
	defer func() { _ = rows.Close() }()
	var out []fkInfo
	for rows.Next() {
		var fk fkInfo
		if err := rows.Scan(&fk.id, &fk.table, &fk.from, &fk.to, &fk.onDelete); err != nil {
			return nil, fmt.Errorf("foreign keys of %s: %w", table, err)
		}
		out = append(out, fk)
	}
	return out, rows.Err()
}

// tableIndexes reads table's indexes through pragma_index_list.
func tableIndexes(ctx context.Context, db *sql.DB, table string) ([]idxInfo, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`SELECT name, "unique" FROM pragma_index_list(%q)`, table))
	if err != nil {
		return nil, fmt.Errorf("indexes of %s: %w", table, err)
	}
	defer func() { _ = rows.Close() }()
	var out []idxInfo
	for rows.Next() {
		var idx idxInfo
		var unique int
		if err := rows.Scan(&idx.name, &unique); err != nil {
			return nil, fmt.Errorf("indexes of %s: %w", table, err)
		}
		idx.unique = unique != 0
		out = append(out, idx)
	}
	return out, rows.Err()
}

// indexColumns reads an index's columns through pragma_index_info, in
// index order.
func indexColumns(ctx context.Context, db *sql.DB, index string) ([]string, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`SELECT name FROM pragma_index_info(%q)`, index))
	if err != nil {
		return nil, fmt.Errorf("index columns of %s: %w", index, err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("index columns of %s: %w", index, err)
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// checkInvariants reads db's catalog and reports every violation of
// the seven schema invariants (spec :433-464), one message per
// violation naming the rule and the table. It never parses SQL.
func checkInvariants(ctx context.Context, db *sql.DB) ([]string, error) {
	tables, err := catalogTables(ctx, db)
	if err != nil {
		return nil, err
	}
	sort.Strings(tables)

	columns := make(map[string][]colInfo, len(tables))
	for _, t := range tables {
		cols, err := tableColumns(ctx, db, t)
		if err != nil {
			return nil, err
		}
		columns[t] = cols
	}

	var violations []string
	strictAndPaths, err := checkStrictAndPaths(ctx, db, tables, columns)
	if err != nil {
		return nil, err
	}
	violations = append(violations, strictAndPaths...)
	fk, err := checkInvariant2ForeignKeys(ctx, db, tables, columns)
	if err != nil {
		return nil, err
	}
	violations = append(violations, fk...)
	violations = append(violations, checkKindAt(tables, columns)...)
	violations = append(violations, checkMusicBrainzIDs(ctx, db, tables, columns)...)
	violations = append(violations, checkAttributeBags(ctx, db, tables, columns)...)
	violations = append(violations, checkEntityIDs(tables, columns)...)
	cascade, err := checkNoCascade(ctx, db, tables)
	if err != nil {
		return nil, err
	}
	violations = append(violations, cascade...)
	return violations, nil
}

// checkStrictAndPaths is invariant 1: every table is STRICT, and a
// path column is declared BLOB.
func checkStrictAndPaths(ctx context.Context, db *sql.DB, tables []string, columns map[string][]colInfo) ([]string, error) {
	var out []string
	for _, t := range tables {
		strict, err := tableIsStrict(ctx, db, t)
		if err != nil {
			return nil, err
		}
		if !strict {
			out = append(out, fmt.Sprintf("invariant 1: table %s is not STRICT", t))
		}
		for _, c := range columns[t] {
			if (c.name == "path" || strings.HasSuffix(c.name, "_path")) && c.typ != "BLOB" {
				out = append(out, fmt.Sprintf("invariant 1: table %s column %s is declared %s, want BLOB", t, c.name, c.typ))
			}
		}
	}
	return out, nil
}

// checkInvariant2ForeignKeys is invariant 2: every foreign key resolves to an
// existing table and column, and its parent columns are the primary
// key or carry a unique index.
func checkInvariant2ForeignKeys(ctx context.Context, db *sql.DB, tables []string, columns map[string][]colInfo) ([]string, error) {
	var out []string
	for _, t := range tables {
		fks, err := tableForeignKeys(ctx, db, t)
		if err != nil {
			return nil, err
		}
		byID := map[int][]fkInfo{}
		for _, fk := range fks {
			byID[fk.id] = append(byID[fk.id], fk)
		}
		var ids []int
		for id := range byID {
			ids = append(ids, id)
		}
		sort.Ints(ids)
		for _, id := range ids {
			group := byID[id]
			parent := group[0].table
			parentCols, ok := columns[parent]
			if !ok {
				out = append(out, fmt.Sprintf("invariant 2: table %s foreign key references a missing table %s", t, parent))
				continue
			}
			parentColSet := make(map[string]bool, len(parentCols))
			for _, pc := range parentCols {
				parentColSet[pc.name] = true
			}
			var to []string
			missing := false
			for _, fk := range group {
				to = append(to, fk.to)
				if !parentColSet[fk.to] {
					out = append(out, fmt.Sprintf("invariant 2: table %s foreign key references %s.%s, a missing column", t, parent, fk.to))
					missing = true
				}
			}
			if missing {
				continue
			}
			sort.Strings(to)
			ok, err := parentKeyMatches(ctx, db, parent, parentCols, to)
			if err != nil {
				return nil, err
			}
			if !ok {
				out = append(out, fmt.Sprintf("invariant 2: table %s foreign key parent %s columns %s carry no primary key or unique index", t, parent, strings.Join(to, ",")))
			}
		}
	}
	return out, nil
}

// parentKeyMatches reports whether want, sorted, equals parent's
// primary key columns or some unique index's columns.
func parentKeyMatches(ctx context.Context, db *sql.DB, parent string, parentCols []colInfo, want []string) (bool, error) {
	var pk []string
	for _, c := range parentCols {
		if c.pk > 0 {
			pk = append(pk, c.name)
		}
	}
	sort.Strings(pk)
	if equalStrings(pk, want) {
		return true, nil
	}
	indexes, err := tableIndexes(ctx, db, parent)
	if err != nil {
		return false, err
	}
	for _, idx := range indexes {
		if !idx.unique {
			continue
		}
		cols, err := indexColumns(ctx, db, idx.name)
		if err != nil {
			return false, err
		}
		sort.Strings(cols)
		if equalStrings(cols, want) {
			return true, nil
		}
	}
	return false, nil
}

// checkKindAt is invariant 3: a table with a kind column, or whose
// name ends in _events, _audit, or _journal, carries at and kind, both
// NOT NULL.
func checkKindAt(tables []string, columns map[string][]colInfo) []string {
	var out []string
	for _, t := range tables {
		byName := make(map[string]colInfo, len(columns[t]))
		for _, c := range columns[t] {
			byName[c.name] = c
		}
		_, hasKind := byName["kind"]
		named := strings.HasSuffix(t, "_events") || strings.HasSuffix(t, "_audit") || strings.HasSuffix(t, "_journal")
		if !hasKind && !named {
			continue
		}
		for _, want := range []string{"at", "kind"} {
			if c, ok := byName[want]; !ok || !c.notNull {
				out = append(out, fmt.Sprintf("invariant 3: table %s has no NOT NULL column %s", t, want))
			}
		}
	}
	return out
}

// checkMusicBrainzIDs is invariant 4: a column named release_id,
// recording_id, artist_id, release_group_id, track_id, or work_id that
// is not a foreign key is refused.
func checkMusicBrainzIDs(ctx context.Context, db *sql.DB, tables []string, columns map[string][]colInfo) []string {
	var out []string
	for _, t := range tables {
		fks, err := tableForeignKeys(ctx, db, t)
		if err != nil {
			continue
		}
		fkCols := make(map[string]bool, len(fks))
		for _, fk := range fks {
			fkCols[fk.from] = true
		}
		for _, c := range columns[t] {
			if mbForbiddenColumns[c.name] && !fkCols[c.name] {
				out = append(out, fmt.Sprintf("invariant 4: table %s column %s is a bare MusicBrainz identifier, not a foreign key", t, c.name))
			}
		}
	}
	return out
}

// checkAttributeBags is invariant 5: no table pairs a foreign key
// column with a name/key/attr column and a value column.
func checkAttributeBags(ctx context.Context, db *sql.DB, tables []string, columns map[string][]colInfo) []string {
	var out []string
	for _, t := range tables {
		fks, err := tableForeignKeys(ctx, db, t)
		if err != nil || len(fks) == 0 {
			continue
		}
		byName := make(map[string]bool, len(columns[t]))
		for _, c := range columns[t] {
			byName[c.name] = true
		}
		hasAttrName := byName["name"] || byName["key"] || byName["attr"]
		if hasAttrName && byName["value"] {
			out = append(out, fmt.Sprintf("invariant 5: table %s pairs a foreign key with a name/key/attr and value column, the attribute-bag shape", t))
		}
	}
	return out
}

// checkEntityIDs is invariant 6: a non-exempt table's primary key is a
// BLOB column named id, and no column named id or ending in _id is an
// INTEGER PRIMARY KEY.
func checkEntityIDs(tables []string, columns map[string][]colInfo) []string {
	var out []string
	for _, t := range tables {
		var pk []colInfo
		for _, c := range columns[t] {
			if c.pk > 0 {
				pk = append(pk, c)
			}
		}
		if !invariant6Exempt[t] {
			if len(pk) != 1 || pk[0].name != "id" || pk[0].typ != "BLOB" {
				out = append(out, fmt.Sprintf("invariant 6: table %s primary key is not a single BLOB column named id", t))
			}
		}
		for _, c := range columns[t] {
			if c.pk > 0 && c.typ == "INTEGER" && (c.name == "id" || strings.HasSuffix(c.name, "_id")) {
				out = append(out, fmt.Sprintf("invariant 6: table %s column %s is an INTEGER PRIMARY KEY named id or ending in _id", t, c.name))
			}
		}
	}
	return out
}

// checkNoCascade is invariant 7: no foreign key to users or
// collections declares ON DELETE CASCADE.
func checkNoCascade(ctx context.Context, db *sql.DB, tables []string) ([]string, error) {
	var out []string
	for _, t := range tables {
		fks, err := tableForeignKeys(ctx, db, t)
		if err != nil {
			return nil, err
		}
		for _, fk := range fks {
			if strings.EqualFold(fk.onDelete, "CASCADE") && (fk.table == "users" || fk.table == "collections") {
				out = append(out, fmt.Sprintf("invariant 7: table %s declares ON DELETE CASCADE onto %s", t, fk.table))
			}
		}
	}
	return out, nil
}

// equalStrings reports whether a and b, taken as sorted, hold the same
// elements.
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestSchemaSkeletonSatisfiesInvariants applies every embedded
// migration to an empty database and checks the result against all
// seven invariants: the skeleton itself must hold none of the
// violations the other tests in this file prove each rule can catch.
func TestSchemaSkeletonSatisfiesInvariants(t *testing.T) {
	l, h := newMigrateStore(t)
	ctx := context.Background()
	if err := migrate(ctx, h, l, migrationSource(), defaultOpts()); err != nil {
		t.Fatalf("migrate to head: %v", err)
	}
	violations, err := checkInvariants(ctx, h.write)
	if err != nil {
		t.Fatalf("check invariants: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("skeleton violates its own invariants: %v", violations)
	}
}

// TestSchemaMigrationsCarryPopulatedFixtures asserts every embedded
// migration ships a populated fixture under testdata/fixtures, named
// after it. For each one, the test migrates to that version, applies
// the fixture, migrates to the head, and checks invariants still hold;
// a migration with no fixture fails the subtest naming it.
func TestSchemaMigrationsCarryPopulatedFixtures(t *testing.T) {
	files, err := readMigrations(migrationSource())
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}
	for _, m := range files {
		t.Run(m.name, func(t *testing.T) {
			l, h := newMigrateStore(t)
			ctx := context.Background()
			if err := migrate(ctx, h, l, migrationSourceThrough(t, m.version), defaultOpts()); err != nil {
				t.Fatalf("migrate to version %d: %v", m.version, err)
			}

			fixturePath := filepath.Join("testdata", "fixtures", m.name)
			fixture, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatalf("migration %s has no fixture at %s: %v", m.name, fixturePath, err)
			}
			if _, err := h.write.ExecContext(ctx, string(fixture)); err != nil {
				t.Fatalf("apply fixture %s: %v", fixturePath, err)
			}

			if err := migrate(ctx, h, l, migrationSource(), defaultOpts()); err != nil {
				t.Fatalf("migrate %s's fixture forward to head: %v", m.name, err)
			}
			violations, err := checkInvariants(ctx, h.write)
			if err != nil {
				t.Fatalf("check invariants: %v", err)
			}
			if len(violations) != 0 {
				t.Errorf("invariants after migrating %s's fixture forward: %v", m.name, violations)
			}
		})
	}
}

// migrationSourceThrough returns the embedded migration set's files at
// version 1 through through, so a test can migrate a store to a
// specific version before applying that version's fixture.
func migrationSourceThrough(t *testing.T, through int) fstest.MapFS {
	t.Helper()
	full := migrationSource()
	entries, err := fs.ReadDir(full, ".")
	if err != nil {
		t.Fatalf("read migration source: %v", err)
	}
	out := fstest.MapFS{}
	for _, e := range entries {
		match := migrationName.FindStringSubmatch(e.Name())
		if match == nil {
			continue
		}
		v, err := strconv.Atoi(match[1])
		if err != nil {
			t.Fatalf("parse version of %s: %v", e.Name(), err)
		}
		if v > through {
			continue
		}
		body, err := fs.ReadFile(full, e.Name())
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		out[e.Name()] = &fstest.MapFile{Data: body}
	}
	return out
}

// TestSchemaInvariantViolations is spec criterion 11: each of the nine
// fixtures under testdata/violations proves one invariant can fail for
// its own named fact, over a small standalone schema applied to a
// fresh database, never the skeleton itself.
func TestSchemaInvariantViolations(t *testing.T) {
	tests := []struct {
		file string
		// want is a substring every reported violation set for this
		// fixture must contain: the rule tag and the table it names.
		want string
	}{
		{"rule1a_not_strict.sql", "invariant 1: table bad_table is not STRICT"},
		{"rule1b_path_text.sql", "invariant 1: table bad_table column rel_path is declared TEXT"},
		{"rule2a_missing_column.sql", "invariant 2: table bad_table foreign key references parent.missing_column, a missing column"},
		{"rule2b_nonunique_parent.sql", "invariant 2: table bad_table foreign key parent parent"},
		{"rule3_events_missing_at.sql", "invariant 3: table bad_events has no NOT NULL column at"},
		{"rule4_bare_release_id.sql", "invariant 4: table bad_table column release_id is a bare MusicBrainz identifier"},
		{"rule5_attribute_bag.sql", "invariant 5: table bad_table pairs a foreign key with a name/key/attr and value column"},
		{"rule6_integer_pk.sql", "invariant 6: table bad_table"},
		{"rule7_cascade_users.sql", "invariant 7: table bad_table declares ON DELETE CASCADE onto users"},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			path := filepath.Join("testdata", "violations", tt.file)
			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			// A named memory database keeps every pooled connection on
			// the same in-memory instance; an anonymous ":memory:" DSN
			// gives sql.DB's second connection an empty database of
			// its own even with cache=shared.
			db, err := sql.Open(sqliteDriverName, "file:"+tt.file+"?mode=memory&cache=shared")
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			db.SetMaxOpenConns(1)
			defer func() { _ = db.Close() }()
			if _, err := db.Exec(string(body)); err != nil {
				t.Fatalf("apply %s: %v", path, err)
			}

			violations, err := checkInvariants(context.Background(), db)
			if err != nil {
				t.Fatalf("check invariants: %v", err)
			}
			if len(violations) == 0 {
				t.Fatalf("%s reported no violation, want one containing %q", tt.file, tt.want)
			}
			found := false
			for _, v := range violations {
				if strings.Contains(v, tt.want) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s violations = %v, want one containing %q", tt.file, violations, tt.want)
			}
		})
	}
}
