package main

import (
	"io"
	"log/slog"
	"strings"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
)

// The four boundary rows doctor prints after the unidecode row and
// before the settings rows. They are fixed rows rather than registry
// declarations, so reg alone never accounts for them.
const (
	clientIPDoctorRow      = "client ip"
	webDoctorRow           = "web"
	bindAddrsDoctorRow     = "bind addresses"
	boundaryProbeDoctorRow = "boundary probe"
)

// foldRefusal renders err as one status line: its message with every
// newline an [errors.Join] carries folded to "; ".
func foldRefusal(err error) string {
	return strings.ReplaceAll(err.Error(), "\n", "; ")
}

// webRowStatus runs the web resolution serve startup runs,
// [newWebRoutes] with the same inputs, so doctor and serve cannot
// disagree about the theme match, the palette, the branding value, or
// the operator CSS file. It reads the embedded bundle and the operator
// file and writes nothing. The logger discards: the one record
// newWebRoutes writes for a branding neutral that drops the
// theme-color meta is serve's to show, and doctor's stdout is the
// table alone. A refusal reads "invalid: " plus the error and fails
// the row; a nil cfg ([config.Load] refused the environment) reads
// "not checked" and adds no failure of its own.
func webRowStatus(environ []string, cfg *config.Config) (status string, failed bool) {
	if cfg == nil {
		return "not checked", false
	}
	prefixValue, err := config.ParseString(environ, serve.PrefixThreshold)
	if err != nil {
		return "invalid: " + foldRefusal(err), true
	}
	prefix, err := serve.NewPrefix(prefixValue)
	if err != nil {
		return "invalid: " + foldRefusal(err), true
	}
	private := privateFiles{stateDir: cfg.StateDir}
	discard := slog.New(slog.NewTextHandler(io.Discard, nil))
	if _, err := newWebRoutes(environ, prefix, cfg.SiteName, private, discard); err != nil {
		return "invalid: " + foldRefusal(err), true
	}
	return "ok", false
}
