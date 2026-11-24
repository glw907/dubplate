package sentinel

import (
	"errors"
	"fmt"
)

var errBare = errors.New("no record")

var errBareFormatted = fmt.Errorf("body exceeds %d MiB", 32)

func wrap(err error) error { return fmt.Errorf("store: %w", err) }
