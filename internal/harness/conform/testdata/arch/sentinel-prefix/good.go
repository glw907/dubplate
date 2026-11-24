package sentinel

import (
	"errors"
	"fmt"
)

var errBare = errors.New("no record")

func wrap(err error) error { return fmt.Errorf("store: %w", err) }
