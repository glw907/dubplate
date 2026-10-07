package stack

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"regexp"
)

// errNotPinned marks an image reference that carries no sha256 digest.
var errNotPinned = errors.New("stack: image reference is not digest-pinned")

// errImagesFile marks a digest file that is absent, empty, or malformed.
var errImagesFile = errors.New("stack: bad image digest file")

var pinned = regexp.MustCompile(`@sha256:[0-9a-f]{64}$`)

// images names the oracle images by repository and digest.
type images struct {
	Navidrome string `json:"navidrome"`
	Gonic     string `json:"gonic"`
}

// loadImages reads the digest file at path. Every reference must end in
// "@sha256:" and 64 hex digits; a tag is never enough, because a tag can
// move under a recorded session. A file with several bad references
// names each.
func loadImages(path string) (images, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return images{}, fmt.Errorf("%w: %s: %w", errImagesFile, path, err)
	}
	if len(data) == 0 {
		return images{}, fmt.Errorf("%w: %s is empty", errImagesFile, path)
	}
	var imgs images
	if err := json.Unmarshal(data, &imgs, json.RejectUnknownMembers(true)); err != nil {
		return images{}, fmt.Errorf("%w: %s: %w", errImagesFile, path, err)
	}
	if err := imgs.check(); err != nil {
		return images{}, err
	}
	return imgs, nil
}

// check reports every reference that lacks a digest.
func (i images) check() error {
	var errs []error
	for _, ref := range []string{i.Navidrome, i.Gonic} {
		if !pinned.MatchString(ref) {
			errs = append(errs, fmt.Errorf("%w: %q", errNotPinned, ref))
		}
	}
	return errors.Join(errs...)
}
