// Command beet is the e2e harness's stand-in for the real beet CLI. It
// walks its target directory (the last argument), treats every path
// named in STUB_BEET_SKIP as beets would treat one it declined to
// auto-tag (left in place, logged to the -l tag log), and "imports"
// everything else: removed from staging (beets' own move:yes) and
// recorded as a row in the plain-text stand-in database at
// STUB_BEETS_DB, which the sqlite3 stub reads back.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	if os.Getenv("STUB_BEET_FAIL") == "1" {
		fmt.Fprintln(os.Stderr, "stub beet: forced failure")
		os.Exit(1)
	}

	// STUB_BEET_SLEEP_MS holds the process open for a controlled window
	// before doing anything else, giving a `dubplate beet` e2e test time
	// to prove a concurrent `dubplate import` sees the lock held.
	if ms := os.Getenv("STUB_BEET_SLEEP_MS"); ms != "" {
		if n, err := strconv.Atoi(ms); err == nil {
			time.Sleep(time.Duration(n) * time.Millisecond)
		}
	}

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "stub beet: no arguments")
		os.Exit(1)
	}
	dir := args[len(args)-1]

	if argsLog := os.Getenv("STUB_BEET_ARGS_LOG"); argsLog != "" {
		if err := os.WriteFile(argsLog, []byte(strings.Join(args, "\n")+"\n"), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "stub beet: write args log:", err)
			os.Exit(1)
		}
	}

	logPath := ""
	for i, a := range args {
		if a == "-l" && i+1 < len(args) {
			logPath = args[i+1]
		}
	}

	skip := make(map[string]bool)
	if raw := os.Getenv("STUB_BEET_SKIP"); raw != "" {
		for _, p := range strings.Split(raw, "\n") {
			if p != "" {
				skip[p] = true
			}
		}
	}

	var logLines []string
	var dbLines []string
	albumIDs := make(map[string]int)
	nextID := 1

	walkErr := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if skip[path] {
			logLines = append(logLines, "skip "+path)
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		segments := strings.Split(filepath.ToSlash(rel), "/")

		// Three or more segments means <user>/<album>/<...track>: group
		// by <user>/<album>, matching one beets album per staged
		// top-level album directory. Fewer segments is a loose file
		// directly under a contributor's staging directory -- a
		// singleton, album_id 0, same as real beets.
		albumID := 0
		if len(segments) >= 3 {
			key := segments[0] + "/" + segments[1]
			id, ok := albumIDs[key]
			if !ok {
				id = nextID
				nextID++
				albumIDs[key] = id
			}
			albumID = id
		}

		added := float64(time.Now().UnixNano()) / 1e9
		dbLines = append(dbLines, fmt.Sprintf("%d\t%s\t%s", albumID, path, strconv.FormatFloat(added, 'f', -1, 64)))
		return os.Remove(path)
	})
	if walkErr != nil {
		fmt.Fprintln(os.Stderr, "stub beet: walk target:", walkErr)
		os.Exit(1)
	}

	if logPath != "" {
		content := strings.Join(logLines, "\n")
		if content != "" {
			content += "\n"
		}
		if err := os.WriteFile(logPath, []byte(content), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "stub beet: write tag log:", err)
			os.Exit(1)
		}
	}

	if dbPath := os.Getenv("STUB_BEETS_DB"); dbPath != "" && len(dbLines) > 0 {
		f, err := os.OpenFile(dbPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "stub beet: open database:", err)
			os.Exit(1)
		}
		w := bufio.NewWriter(f)
		for _, line := range dbLines {
			fmt.Fprintln(w, line)
		}
		if err := w.Flush(); err != nil {
			fmt.Fprintln(os.Stderr, "stub beet: flush database:", err)
			os.Exit(1)
		}
		if err := f.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "stub beet: close database:", err)
			os.Exit(1)
		}
	}
}
