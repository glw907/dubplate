// Command sqlite3 is the e2e harness's stand-in for the real sqlite3
// CLI. It answers three query shapes: the album lookup
// [beets.ResolveLibraryPaths] issues -- "album_id, path FROM items WHERE
// added >= X AND added <= Y ORDER BY album_id, path" -- against the
// plain-text stand-in database the beet stub writes, in the same
// "-noheader -separator \t" output shape the real binary would produce;
// [backup.Runner]'s ".backup '<dest>'" dot-command, which it answers by
// copying the source database file to dest, standing in for a real
// consistent snapshot; and [restore.Runner]'s "PRAGMA integrity_check;",
// which answers "ok" unless the database path is named in
// STUB_SQLITE_CORRUPT_DB, matching real sqlite3's own output shape on
// each outcome.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	boundsRe = regexp.MustCompile(`added >= ([0-9.]+) AND added <= ([0-9.]+)`)
	backupRe = regexp.MustCompile(`^\.backup '(.+)'$`)
)

const integrityCheckQuery = "PRAGMA integrity_check;"

type row struct {
	albumID int
	path    string
}

func main() {
	dbPath, query := parseArgs(os.Args[1:])

	if m := backupRe.FindStringSubmatch(query); m != nil {
		if err := copyFile(dbPath, m[1]); err != nil {
			fmt.Fprintln(os.Stderr, "stub sqlite3:", err)
			os.Exit(1)
		}
		return
	}

	if query == integrityCheckQuery {
		if os.Getenv("STUB_SQLITE_CORRUPT_DB") == dbPath {
			fmt.Println("*** in database main ***\nrow 5 missing from index items_index")
		} else {
			fmt.Println("ok")
		}
		return
	}

	m := boundsRe.FindStringSubmatch(query)
	if m == nil {
		fmt.Fprintln(os.Stderr, "stub sqlite3: query carries no added bounds:", query)
		os.Exit(1)
	}
	since, err1 := strconv.ParseFloat(m[1], 64)
	until, err2 := strconv.ParseFloat(m[2], 64)
	if err1 != nil || err2 != nil {
		fmt.Fprintln(os.Stderr, "stub sqlite3: parse added bounds:", query)
		os.Exit(1)
	}

	rows, err := readRows(dbPath, since, until)
	if err != nil {
		fmt.Fprintln(os.Stderr, "stub sqlite3:", err)
		os.Exit(1)
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].albumID != rows[j].albumID {
			return rows[i].albumID < rows[j].albumID
		}
		return rows[i].path < rows[j].path
	})

	w := bufio.NewWriter(os.Stdout)
	defer func() { _ = w.Flush() }()
	for _, r := range rows {
		fmt.Fprintf(w, "%d\t%s\n", r.albumID, r.path)
	}
}

// parseArgs extracts the database path and the SQL query from a
// "-noheader -separator \t <dbPath> <query>" invocation.
func parseArgs(args []string) (dbPath, query string) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-noheader":
		case "-separator":
			i++ // its value follows; not needed here
		default:
			if dbPath == "" {
				dbPath = args[i]
			} else {
				query = args[i]
			}
		}
	}
	return dbPath, query
}

// copyFile stands in for sqlite3's own consistent-snapshot ".backup":
// this stub's "databases" are plain files, so a byte copy is enough to
// prove [backup.Runner] invoked the dot-command against the right source
// and destination.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source %s: %w", src, err)
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination %s: %w", dst, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return fmt.Errorf("copy %s to %s: %w", src, dst, err)
	}
	return out.Close()
}

func readRows(dbPath string, since, until float64) ([]row, error) {
	f, err := os.Open(dbPath)
	if os.IsNotExist(err) {
		return nil, nil // an empty database: no rows, no output
	}
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	defer func() { _ = f.Close() }()

	var rows []row
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		albumID, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		added, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			continue
		}
		if added < since || added > until {
			continue
		}
		rows = append(rows, row{albumID: albumID, path: parts[1]})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read database: %w", err)
	}
	return rows, nil
}
