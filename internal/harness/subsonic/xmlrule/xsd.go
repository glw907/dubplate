// Package xmlrule holds the XML tier's inputs for the OpenSubsonic
// harness. The Subsonic 1.16.1 schema ships without a license text, so
// no committed file carries its bytes: it lives in the user cache,
// pinned by size and sha256, and every consumer reads it through
// Cached.
package xmlrule

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// The one pin. No other channel sets the schema's address, size, or
// digest.
const (
	// XSDURL is where the pinned schema is fetched from.
	XSDURL = "https://subsonic.org/pages/inc/api/schema/subsonic-rest-api-1.16.1.xsd"

	// xsdSize is the pinned schema's length in bytes.
	xsdSize = 32775

	// xsdSHA256 is the hex sha256 of the pinned schema.
	xsdSHA256 = "9baee72fa24087190ad22be4f7674d628acc7a3e0fb65160d74115b166df0387"

	xsdFileName = "subsonic-rest-api-1.16.1.xsd"
)

var (
	// ErrNotCached reports that the cache holds no schema file.
	ErrNotCached = errors.New("xmlrule: schema not cached")

	// ErrChecksum reports bytes whose sha256 disagrees with the pin.
	ErrChecksum = errors.New("xmlrule: schema checksum mismatch")

	// ErrSize reports a source longer than the pinned size.
	ErrSize = errors.New("xmlrule: schema larger than its pinned size")
)

// CachePath returns the file the pinned schema caches to, under the
// user's cache directory in a subdirectory named for dubplate.
func CachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("xmlrule: locate cache directory: %w", err)
	}
	return filepath.Join(dir, "dubplate", xsdFileName), nil
}

// Cached returns the cached schema's bytes, and only when their sha256
// matches the pin. An absent file wraps ErrNotCached and names the
// make target that fills it; a file with the wrong sum wraps
// ErrChecksum and names both sums and the path.
func Cached() ([]byte, error) {
	path, err := CachePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w at %s; run make subsonic-xsd", ErrNotCached, path)
	}
	if err != nil {
		return nil, fmt.Errorf("xmlrule: read cached schema: %w", err)
	}
	if got := sum(data); got != xsdSHA256 {
		return nil, fmt.Errorf("%w at %s: got %s, want %s; run make subsonic-xsd",
			ErrChecksum, path, got, xsdSHA256)
	}
	return data, nil
}

// Fetch downloads the pinned schema with client and returns its bytes
// only when they match the pinned size and sha256. Nothing is written.
func Fetch(ctx context.Context, client *http.Client) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, XSDURL, nil)
	if err != nil {
		return nil, fmt.Errorf("xmlrule: build schema request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("xmlrule: fetch schema: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("xmlrule: fetch schema: unexpected status %s", resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, xsdSize+1))
	if err != nil {
		return nil, fmt.Errorf("xmlrule: read schema: %w", err)
	}
	if len(data) > xsdSize {
		return nil, ErrSize
	}
	if got := sum(data); got != xsdSHA256 {
		return nil, fmt.Errorf("%w: got %s, want %s", ErrChecksum, got, xsdSHA256)
	}
	return data, nil
}

func sum(data []byte) string {
	s := sha256.Sum256(data)
	return hex.EncodeToString(s[:])
}
