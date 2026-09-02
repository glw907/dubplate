// Command rclone is the e2e harness's stand-in for the real rclone CLI.
// It logs every invocation's arguments to STUB_RCLONE_LOG when set (the
// import flow never calls it at all, so that alone is enough to
// guarantee a real rclone binary is never invoked by accident).
// [backup.Runner]'s suite uses only that logging: sync moves no data in
// this stub. [restore.Runner]'s suite needs real data movement, so
// `copy` and `lsjson` read and write files under STUB_RCLONE_BUCKET_DIR,
// the local directory standing in for the R2 bucket root, matching real
// rclone's S3 behavior that a missing source prefix is an empty result,
// never an error.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if os.Getenv("STUB_RCLONE_FAIL") == "1" {
		fmt.Fprintln(os.Stderr, "stub rclone: forced failure")
		os.Exit(1)
	}

	logArgs()

	args := os.Args[1:]
	if len(args) == 0 {
		return
	}

	var err error
	switch args[0] {
	case "copy":
		err = runCopy(args[1:])
	case "lsjson":
		err = runLsjson(args[1:])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "stub rclone:", err)
		os.Exit(1)
	}
}

func logArgs() {
	logPath := os.Getenv("STUB_RCLONE_LOG")
	if logPath == "" {
		return
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	fmt.Fprintln(f, os.Args[1:])
}

// runCopy copies every file under the bucket-relative source path into
// dst, creating dst if needed. A source the bucket stand-in does not
// have is silently zero files, matching S3's own missing-prefix
// behavior rather than a "no such directory" error.
func runCopy(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("copy: want <src> <dst>, got %v", args)
	}
	src := bucketPath(args[0])
	dst := args[1]

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dst, err)
	}
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		return copyFile(path, filepath.Join(dst, rel))
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// lsjsonEntry mirrors the subset of real rclone's lsjson entry fields
// [restore.Runner] reads.
type lsjsonEntry struct {
	Path  string `json:"Path"`
	IsDir bool   `json:"IsDir"`
}

// runLsjson lists the bucket-relative path named by the last argument,
// honoring --dirs-only and --files-only, and writes the result as JSON
// to stdout. A path the bucket stand-in does not have lists as empty,
// matching runCopy's same missing-prefix leniency.
func runLsjson(args []string) error {
	dirsOnly := false
	filesOnly := false
	var target string
	for _, a := range args {
		switch a {
		case "--dirs-only":
			dirsOnly = true
		case "--files-only":
			filesOnly = true
		case "--recursive":
			// always recursive in this stub
		default:
			if !strings.HasPrefix(a, "-") {
				target = a
			}
		}
	}

	root := bucketPath(target)
	var entries []lsjsonEntry
	if _, err := os.Stat(root); err == nil {
		walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path == root {
				return nil
			}
			if dirsOnly && !d.IsDir() {
				return nil
			}
			if filesOnly && d.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			entries = append(entries, lsjsonEntry{Path: filepath.ToSlash(rel), IsDir: d.IsDir()})
			return nil
		})
		if walkErr != nil {
			return walkErr
		}
	}
	if entries == nil {
		entries = []lsjsonEntry{}
	}

	enc := json.NewEncoder(os.Stdout)
	return enc.Encode(entries)
}

// bucketPath maps a "<remote>:<bucket>/<subpath>" argument to a real
// local path under STUB_RCLONE_BUCKET_DIR, the local stand-in for the R2
// bucket root that [restore.Runner]'s e2e fixtures populate.
func bucketPath(remoteArg string) string {
	bucketDir := os.Getenv("STUB_RCLONE_BUCKET_DIR")
	_, afterColon, ok := strings.Cut(remoteArg, ":")
	if !ok {
		return bucketDir
	}
	_, rel, ok := strings.Cut(afterColon, "/")
	if !ok {
		return bucketDir
	}
	return filepath.Join(bucketDir, rel)
}
