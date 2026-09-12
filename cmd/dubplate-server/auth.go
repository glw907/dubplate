package main

import "github.com/glw907/dubplate/internal/config"

// identitySourceSelectorRow declares DUBPLATE_IDENTITY_SOURCE as an
// owned wall row with no Pending, retiring the pending row keys.go
// carried from rung 1 (R5): the selector is required now, alongside
// [magicLinkSource] itself, rather than pending a rung that has
// arrived.
var identitySourceSelectorRow = config.WallRow{Name: "DUBPLATE_IDENTITY_SOURCE"}

// magicLinkSource is the shipped default identity source: a login link
// mailed to a roster address, backed by an HMAC-signed session cookie.
// Its nine required names and two optional names retire
// internal/config/keys.go's "auth and identity" reserved class in the
// same commit that declares them (R-5b3): DUBPLATE_ROSTER_FILE and the
// five DUBPLATE_SMTP_* names land here, declared and
// required-when-selected, though their consumers (the roster store,
// the mail seam) arrive at 5b-ii. The two optional names carry a
// rotation-overlap secret each, so a deployment can roll a secret
// without a window where every existing token or session fails at
// once (R-5b6): they are the source's Optional keys rather than
// registered thresholds, because a threshold's KindString default
// must be non-empty and doctor prints a threshold's effective value,
// which would print a secret on an operator's terminal.
//
// Validate is nil here: Task 4 assigns the source's value rules (the
// secret length floor, the public URL's shape, the retention-vs-expiry
// relation) to this field in the same file.
//
// ruled, conductor 2026-09-12, R-5b2, R-5b3
var magicLinkSource = config.IdentitySource{
	Name:     "magic-link",
	Selector: identitySourceSelectorRow.Name,
	Required: []string{
		"DUBPLATE_ROSTER_FILE",
		"DUBPLATE_MAGIC_LINK_SECRET",
		"DUBPLATE_SESSION_SECRET",
		"DUBPLATE_PUBLIC_URL",
		"DUBPLATE_SMTP_HOST",
		"DUBPLATE_SMTP_PORT",
		"DUBPLATE_SMTP_USER",
		"DUBPLATE_SMTP_PASSWORD",
		"DUBPLATE_SMTP_FROM",
	},
	Optional: []string{
		"DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS",
		"DUBPLATE_SESSION_SECRET_PREVIOUS",
	},
}

// registerAuthConfig adds rung 5's twelve names to reg: the selector as
// an owned wall row, and magic-link as a declared identity source. It
// is a plain function on registerServeConfig's own precedent
// (serve.go), called once from the binary's shared registry
// constructor (newProductionRegistry, main.go).
func registerAuthConfig(reg *config.Registry) {
	reg.AddWallRow(identitySourceSelectorRow)
	reg.AddIdentitySource(magicLinkSource)
}
