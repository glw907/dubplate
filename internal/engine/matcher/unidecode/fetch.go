package unidecode

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Fetch downloads Unidecode 1.4.0's pinned wheel over client and
// verifies its checksum, refusing a non-200 response, a body over
// MaxWheelSize, and a body whose sha256 disagrees with WheelSHA256.
// The caller's ctx bounds the request; Fetch sets no timeout of its
// own.
func Fetch(ctx context.Context, client *http.Client) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, WheelURL, nil)
	if err != nil {
		return nil, fmt.Errorf("unidecode: build wheel request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unidecode: fetch wheel: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unidecode: fetch wheel: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxWheelSize+1))
	if err != nil {
		return nil, fmt.Errorf("unidecode: read wheel body: %w", err)
	}
	if len(body) > MaxWheelSize {
		return nil, ErrWheelTooLarge
	}

	sum := sha256.Sum256(body)
	got := hex.EncodeToString(sum[:])
	if got != WheelSHA256 {
		return nil, fmt.Errorf("%w: got %s, want %s", ErrWheelChecksum, got, WheelSHA256)
	}
	return body, nil
}

// WheelCachePath returns the workstation file the pinned wheel caches
// to, under the user's cache directory in a subdirectory named for
// this package.
func WheelCachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("unidecode: locate cache directory: %w", err)
	}
	return filepath.Join(dir, "unidecode", "Unidecode-1.4.0-py3-none-any.whl"), nil
}
