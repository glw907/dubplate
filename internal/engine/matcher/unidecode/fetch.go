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

// Fetch downloads Unidecode 1.4.0's pinned wheel over client,
// refusing a non-200 response and any body [ReadWheel] refuses. The
// caller's ctx bounds the request; Fetch sets no timeout of its own.
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

	return ReadWheel(resp.Body)
}

// ReadWheel reads the pinned wheel from r and returns its bytes. It
// reads at most maxWheelSize+1 bytes, so an oversized source refuses
// before anything is hashed, and refuses bytes whose sha256 disagrees
// with WheelSHA256.
func ReadWheel(r io.Reader) ([]byte, error) {
	wheel, err := io.ReadAll(io.LimitReader(r, maxWheelSize+1))
	if err != nil {
		return nil, fmt.Errorf("unidecode: read wheel: %w", err)
	}
	if len(wheel) > maxWheelSize {
		return nil, errWheelTooLarge
	}
	sum := sha256.Sum256(wheel)
	got := hex.EncodeToString(sum[:])
	if got != WheelSHA256 {
		return nil, fmt.Errorf("%w: got %s, want %s", errWheelChecksum, got, WheelSHA256)
	}
	return wheel, nil
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
