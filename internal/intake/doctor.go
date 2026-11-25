package intake

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/doctor"
	"github.com/glw907/dubplate/internal/intake/staging"
	"github.com/glw907/dubplate/internal/store"
	"golang.org/x/sys/unix"
)

// collectionsVersion is the lowest store version holding the
// collections table (0003_catalog.sql).
const collectionsVersion = 3

// Facts are the collection roots a store inspection lists for the
// doctor command, relative to the library root.
type Facts struct {
	roots [][]byte
}

// Reader returns the [store.Reader] that fills f from the collections
// table.
func (f *Facts) Reader() store.Reader {
	return func(ctx context.Context, q store.Querier, version int) error {
		if version < collectionsVersion {
			return nil
		}
		rows, err := q.QueryContext(ctx, `SELECT root_path FROM collections ORDER BY root_path`)
		if err != nil {
			return fmt.Errorf("list collection roots: %w", err)
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var root []byte
			if err := rows.Scan(&root); err != nil {
				return fmt.Errorf("scan a collection root: %w", err)
			}
			f.roots = append(f.roots, root)
		}
		return rows.Err()
	}
}

// volumeRowName is the name of the row reporting the music volume:
// staging, the library root, and the collection roots under it.
const volumeRowName = "music volume"

// DoctorRows renders the music volume row for cfg. It lists the
// collection roots f holds only when report shows f's reader ran, since
// a refused inspection may leave f partly filled.
func DoctorRows(cfg *config.Config, report store.Report, f Facts) []doctor.Row {
	var roots [][]byte
	if report.Facts != nil {
		roots = f.roots
	}
	row := doctor.Row{Name: volumeRowName}
	row.Status, row.Failed = volumeStatus(cfg, roots)
	return []doctor.Row{row}
}

// volumeStatus renders the music volume row: what the check serve runs
// at startup finds of staging and the library root, then where each of
// roots sits.
//
// A directory not created yet is named without failing the row, as on
// a fresh box before deploy.sh runs, and a collection root not created
// yet reads "not created", since the store can list a root before it
// is made. Any other refusal from the startup check fails the row, as
// does a collection root outside the library root's mount. A nil cfg
// ([config.Load] refused the environment) reads "not checked".
func volumeStatus(cfg *config.Config, roots [][]byte) (status string, failed bool) {
	if cfg == nil {
		return "not checked", false
	}
	stagingDir, library := staging.Read(cfg.WallValue(StagingKey)), staging.Read(cfg.LibraryRoot)
	var absent []string
	for _, d := range []struct {
		role string
		dir  staging.Dir
	}{{"staging directory", stagingDir}, {"library root", library}} {
		if errors.Is(d.dir.Err, fs.ErrNotExist) {
			absent = append(absent, d.role+" "+d.dir.Path+" does not exist")
		}
	}
	if len(absent) > 0 {
		return strings.Join(absent, "; "), false
	}
	if err := staging.Check(stagingDir, library); err != nil {
		return err.Error(), true
	}

	parts := []string{"ok: staging and the library root share one mount"}
	for _, root := range roots {
		path := filepath.Join(cfg.LibraryRoot, string(root))
		d := staging.Read(path)
		switch {
		case errors.Is(d.Err, fs.ErrNotExist):
			parts = append(parts, "collection root "+path+" not created")
		case d.Err != nil:
			parts, failed = append(parts, "collection root: "+d.Err.Error()), true
		case d.StatxErr != nil:
			parts, failed = append(parts, "collection root: "+d.StatxErr.Error()), true
		case d.Dev != library.Dev || d.Mask&unix.STATX_MNT_ID == 0 || d.MountID != library.MountID:
			parts, failed = append(parts, "collection root "+path+" sits outside the library root's mount"), true
		default:
			parts = append(parts, "collection root "+path+" on the library root's mount")
		}
	}
	return strings.Join(parts, "; "), failed
}
