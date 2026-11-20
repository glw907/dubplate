package store

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// A tableOwner names who writes a table: dir, the slash-separated
// directory of the one package that may, or round, the round that
// takes a table nothing writes yet.
type tableOwner struct {
	dir   string
	round string
}

// tableOwners holds every table the migrations create. A package
// writes only the tables listed under its own directory, so a second
// writer is a failing test and a table lands here with its first
// writer. The reader outside the rule is internal/store/inspect.go,
// which reads every table and writes none.
//
// intake writes collections only until the catalog round, which takes
// them.
var tableOwners = map[string]tableOwner{
	"users":              {dir: "internal/auth"},
	"member_preferences": {dir: "internal/auth"},

	"collections":         {dir: "internal/intake"},
	"submissions":         {dir: "internal/intake"},
	"submission_files":    {dir: "internal/intake"},
	"submission_events":   {dir: "internal/intake"},
	"submission_warnings": {dir: "internal/intake"},

	"jobs": {dir: "internal/ops/local"},

	"settings":            {dir: "internal/settings"},
	"settings_audit":      {dir: "internal/settings"},
	"settings_generation": {dir: "internal/settings"},

	"migrations": {dir: "internal/store"},

	"albums":           {round: "catalog"},
	"tracks":           {round: "catalog"},
	"move_intents":     {round: "catalog"},
	"member_quota":     {round: "catalog"},
	"group_quota":      {round: "catalog"},
	"member_overrides": {round: "catalog"},

	"device_credentials": {round: "playback"},
}

// sqlWrite is one SQL write found in non-test Go: the table it names,
// the file and line, and the directory of the package that holds it.
type sqlWrite struct {
	table string
	where string
	dir   string
	file  string
}

// unknownOperand stands in for an operand of a string expression that
// is not a literal or a constant, so the folded text keeps its shape.
const unknownOperand = "\x00"

var (
	// writeSQL matches a write statement and captures its table. The
	// match is case-sensitive, which keeps lowercase log text ("insert
	// job", "delete %s") out.
	writeSQL = regexp.MustCompile(`\b(?:INSERT(?:\s+OR\s+\w+)?\s+INTO|REPLACE\s+INTO|UPDATE|DELETE\s+FROM)\s+(\w+)`)

	// upsertArm matches the text just before the UPDATE of an upsert's
	// DO UPDATE SET arm, which writes the INSERT's table, not one named SET.
	upsertArm = regexp.MustCompile(`\bDO\s+$`)

	// unresolvedWrite matches a string that opens a write statement
	// whose table name is an unknown operand.
	unresolvedWrite = regexp.MustCompile(`^\s*(?:INSERT(?:\s+OR\s+\w+)?\s+INTO|REPLACE\s+INTO|UPDATE|DELETE\s+FROM)\s+\x00`)

	createTable = regexp.MustCompile(`\bCREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`)
)

// foldString returns the text of a string expression: a literal, a
// package or block constant, or a + chain of them. An operand it cannot
// resolve folds to unknownOperand. consts maps a constant's name to
// its expression, and nil marks a name declared twice in the package.
func foldString(e ast.Expr, consts map[string]ast.Expr, depth int) string {
	if depth > 8 {
		return unknownOperand
	}
	switch e := e.(type) {
	case *ast.BasicLit:
		if s, err := strconv.Unquote(e.Value); err == nil && e.Kind == token.STRING {
			return s
		}
	case *ast.ParenExpr:
		return foldString(e.X, consts, depth)
	case *ast.BinaryExpr:
		if e.Op == token.ADD {
			return foldString(e.X, consts, depth) + foldString(e.Y, consts, depth)
		}
	case *ast.Ident:
		if def := consts[e.Name]; def != nil {
			return foldString(def, consts, depth+1)
		}
	}
	return unknownOperand
}

// packageConsts collects the constants every file of one package
// declares, package scope and block scope alike. A name declared more
// than once maps to nil, so a use of it stays unresolved.
func packageConsts(files []*ast.File) map[string]ast.Expr {
	consts := map[string]ast.Expr{}
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			decl, ok := n.(*ast.GenDecl)
			if !ok || decl.Tok != token.CONST {
				return true
			}
			for _, spec := range decl.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok || len(vs.Values) != len(vs.Names) {
					continue
				}
				for i, name := range vs.Names {
					if _, seen := consts[name.Name]; seen {
						consts[name.Name] = nil
						continue
					}
					consts[name.Name] = vs.Values[i]
				}
			}
			return true
		})
	}
	return consts
}

// scanSQLWrites parses every non-test .go file under each of dirs,
// resolved against root, and returns the SQL writes it finds plus the
// positions of write statements whose table name it cannot resolve. A
// testdata subtree is skipped.
func scanSQLWrites(root string, dirs ...string) (writes []sqlWrite, unresolved []string, err error) {
	fset := token.NewFileSet()
	byDir := map[string][]*ast.File{}
	for _, d := range dirs {
		walkErr := filepath.WalkDir(filepath.Join(root, d), func(path string, e fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if e.IsDir() {
				if e.Name() == "testdata" {
					return fs.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			f, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			byDir[filepath.Dir(path)] = append(byDir[filepath.Dir(path)], f)
			return nil
		})
		if walkErr != nil {
			return nil, nil, walkErr
		}
	}

	for dir, files := range byDir {
		consts := packageConsts(files)
		rel, err := filepath.Rel(root, dir)
		if err != nil {
			return nil, nil, err
		}
		for _, f := range files {
			file, err := filepath.Rel(root, fset.Position(f.Pos()).Filename)
			if err != nil {
				return nil, nil, err
			}
			ast.Inspect(f, func(n ast.Node) bool {
				var text string
				switch n := n.(type) {
				case *ast.BasicLit:
					if n.Kind != token.STRING {
						return true
					}
					text = foldString(n, consts, 0)
				case *ast.BinaryExpr:
					if n.Op != token.ADD {
						return true
					}
					text = foldString(n, consts, 0)
				default:
					return true
				}
				where := fset.Position(n.Pos()).String()
				where = strings.TrimPrefix(where, root+string(filepath.Separator))
				if unresolvedWrite.MatchString(text) {
					unresolved = append(unresolved, where)
				}
				for _, m := range writeSQL.FindAllStringSubmatchIndex(text, -1) {
					if upsertArm.MatchString(text[:m[0]]) {
						continue
					}
					writes = append(writes, sqlWrite{
						table: text[m[2]:m[3]],
						where: where,
						dir:   filepath.ToSlash(rel),
						file:  filepath.ToSlash(file),
					})
				}
				// A + chain was folded whole; its operands add nothing.
				_, chain := n.(*ast.BinaryExpr)
				return !chain
			})
		}
	}
	slices.SortFunc(writes, func(a, b sqlWrite) int { return strings.Compare(a.where, b.where) })
	slices.Sort(unresolved)
	return writes, unresolved, nil
}

// migrationTables returns every table the migrations create, sorted.
func migrationTables(t *testing.T) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("migrations", "*.sql"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no migrations found (err %v)", err)
	}
	var tables []string
	for _, p := range paths {
		body, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range createTable.FindAllStringSubmatch(string(body), -1) {
			tables = append(tables, m[1])
		}
	}
	slices.Sort(tables)
	return tables
}

func TestTableOwnersNameEveryTable(t *testing.T) {
	tables := migrationTables(t)
	for _, table := range tables {
		if _, ok := tableOwners[table]; !ok {
			t.Errorf("table %s is created by a migration but has no entry in tableOwners", table)
		}
	}
	for table := range tableOwners {
		if !slices.Contains(tables, table) {
			t.Errorf("tableOwners lists %s, which no migration creates", table)
		}
	}
}

func TestTablesAreWrittenOnlyByTheirOwner(t *testing.T) {
	writes, unresolved, err := scanSQLWrites(filepath.Join("..", ".."), "internal", "cmd")
	if err != nil {
		t.Fatal(err)
	}
	for _, where := range unresolved {
		t.Errorf("%s: a write statement names a table the scan cannot resolve; spell it as a literal or a constant", where)
	}
	for _, w := range writes {
		owner, listed := tableOwners[w.table]
		switch {
		case w.file == "internal/store/inspect.go":
			t.Errorf("%s: the inspector writes %s, and it must write nothing", w.where, w.table)
		case !listed:
			t.Errorf("%s: %s writes %s, which tableOwners does not list", w.where, w.dir, w.table)
		case owner.dir == "":
			t.Errorf("%s: %s writes %s, which no package owns before round %s", w.where, w.dir, w.table, owner.round)
		case owner.dir != w.dir:
			t.Errorf("%s: %s writes %s, which %s owns", w.where, w.dir, w.table, owner.dir)
		}
	}
}

// TestScanSQLWrites proves the scan on a planted package: a literal, a
// package constant, a block constant, an upsert arm, lowercase log
// text, and a write whose table is no constant.
func TestScanSQLWrites(t *testing.T) {
	root := t.TempDir()
	pkg := filepath.Join(root, "internal", "planted")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatal(err)
	}
	const src = `package planted

const pkgTable = "jobs"

func f(tx interface{ Exec(string, ...any) }, dyn string) {
	const blockTable = "users"
	tx.Exec("INSERT INTO settings (k) VALUES (?) ON CONFLICT (k) DO UPDATE SET v = 1")
	tx.Exec("DELETE FROM " + pkgTable)
	tx.Exec("UPDATE " + blockTable + " SET admin = 0")
	tx.Exec("DELETE FROM " + dyn)
	_ = "insert job and delete %s"
}
`
	if err := os.WriteFile(filepath.Join(pkg, "p.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	writes, unresolved, err := scanSQLWrites(root, "internal")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, w := range writes {
		got = append(got, w.dir+" "+w.table)
	}
	want := []string{"internal/planted settings", "internal/planted jobs", "internal/planted users"}
	if !slices.Equal(got, want) {
		t.Errorf("writes = %q, want %q", got, want)
	}
	if len(unresolved) != 1 || !strings.HasPrefix(unresolved[0], filepath.Join("internal", "planted", "p.go")+":10:") {
		t.Errorf("unresolved = %q, want the dyn write at p.go line 10", unresolved)
	}
}
