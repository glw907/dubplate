package sentinel

import "fmt"

const limit = 32

var errTooLarge = fmt.Errorf("client: body exceeds %d MiB", limit)
