package stack

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"regexp"
)

// ErrNotPinned marks an image reference that carries no sha256 digest.
var ErrNotPinned = errors.New("stack: image reference is not digest-pinned")

// ErrImagesFile marks a digest file that is absent, empty, or malformed.
var ErrImagesFile = errors.New("stack: bad image digest file")

var pinned = regexp.MustCompile(`@sha256:[0-9a-f]{64}$`)

// Images names the oracle images by repository and digest.
type Images struct {
	Navidrome string `json:"navidrome"`
	Gonic     string `json:"gonic"`
}

// LoadImages reads the digest file at path. Every reference must end in
// "@sha256:" and 64 hex digits; a tag is never enough, because a tag can
// move under a recorded session. A file with several bad references
// names each.
func LoadImages(path string) (Images, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Images{}, fmt.Errorf("%w: %s: %w", ErrImagesFile, path, err)
	}
	if len(data) == 0 {
		return Images{}, fmt.Errorf("%w: %s is empty", ErrImagesFile, path)
	}
	var imgs Images
	if err := json.Unmarshal(data, &imgs, json.RejectUnknownMembers(true)); err != nil {
		return Images{}, fmt.Errorf("%w: %s: %w", ErrImagesFile, path, err)
	}
	if err := imgs.check(); err != nil {
		return Images{}, err
	}
	return imgs, nil
}

// check reports every reference that lacks a digest.
func (i Images) check() error {
	var errs []error
	for _, ref := range []string{i.Navidrome, i.Gonic} {
		if !pinned.MatchString(ref) {
			errs = append(errs, fmt.Errorf("%w: %q", ErrNotPinned, ref))
		}
	}
	return errors.Join(errs...)
}
