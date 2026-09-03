//go:build oracle

// component.go drives derive_component.py, the narrower of this
// package's two beets entry points. Where a derivation needs a stub, a
// rendered profile, and a response set, a component call needs none of
// them, so this driver is one venv invocation and nothing else.
package oracle

import "context"

// DeriveComponentTable runs derive_component.py and returns the table
// it prints, unparsed. The script takes no case, no profile, and no
// MusicBrainz response: every row is a literal argument tuple applied
// to a beets distance primitive under beets' own shipped defaults.
func DeriveComponentTable(ctx context.Context, opts Options) ([]byte, error) {
	return runPython(ctx, opts, "derive_component.py")
}
