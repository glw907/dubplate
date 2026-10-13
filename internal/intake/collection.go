// Package intake is the store side of member uploads: the submissions, the
// two collections the server writes for them, and the rules that move one
// from upload to admission.
package intake

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/glw907/dubplate/internal/ops"
)

// A Route names one of the two collections the server writes for intake.
type Route string

const (
	// RouteDefault is the main route, where a matched album lands.
	RouteDefault Route = "default"
	// RouteUnlisted is the route for music MusicBrainz cannot place, and
	// for separate tracks.
	RouteUnlisted Route = "unlisted"
)

// routeNames is the name each route's collection carries.
var routeNames = map[Route]string{
	RouteDefault:  "Library",
	RouteUnlisted: "Unlisted music",
}

var errNestedRoot = errors.New("collection root nests with an existing root")

// RouteCollection returns the id of the collection holding route, creating
// it in tx when none does. A new collection is group-visible, written by
// the server, and owned by the earliest-created live admin; its root is
// the lowercase hex of its id, so no two collections' roots nest. With no
// live admin it logs and returns an [ops.ErrInternal] error, which cannot
// happen after bootstrap.
func RouteCollection(ctx context.Context, tx *store.Tx, log *slog.Logger, route Route) ([]byte, error) {
	var id []byte
	err := tx.QueryRow(ctx, `SELECT id FROM collections WHERE intake_route = ?`, string(route)).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("find %s collection: %w", route, err)
	}

	var owner []byte
	err = tx.QueryRow(ctx, `SELECT id FROM users WHERE admin = 1 AND tombstoned_at IS NULL
		ORDER BY created_at, id LIMIT 1`).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		log.ErrorContext(ctx, "intake: no live admin to own the collection", "route", route)
		return nil, ops.NewInternalError("no live admin to own the " + routeNames[route] + " collection")
	}
	if err != nil {
		return nil, fmt.Errorf("find collection owner: %w", err)
	}

	id = store.NewID()
	root := []byte(hex.EncodeToString(id))
	if !gate.ValidPath(root) {
		return nil, fmt.Errorf("collection root %q is not a valid path", root)
	}
	if err := checkNoNesting(ctx, tx, root); err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO collections (id, name, root_path, owner, visibility, writer, intake_route)
		VALUES (?, ?, ?, ?, 'group', 'server', ?)`, id, routeNames[route], root, owner, string(route))
	if err != nil {
		return nil, fmt.Errorf("create %s collection: %w", route, err)
	}
	return id, nil
}

// checkNoNesting returns [errNestedRoot] when root equals, sits inside, or
// holds an existing collection's root, comparing whole path components.
func checkNoNesting(ctx context.Context, tx *store.Tx, root []byte) error {
	rows, err := tx.Query(ctx, `SELECT root_path FROM collections`)
	if err != nil {
		return fmt.Errorf("read collection roots: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var existing []byte
		if err := rows.Scan(&existing); err != nil {
			return fmt.Errorf("read collection root: %w", err)
		}
		if within(root, existing) || within(existing, root) {
			return fmt.Errorf("%w: %q and %q", errNestedRoot, root, existing)
		}
	}
	return rows.Close()
}

// within reports whether path equals dir or sits beneath it.
func within(path, dir []byte) bool {
	return bytes.Equal(path, dir) || bytes.HasPrefix(path, append(bytes.Clone(dir), '/'))
}
