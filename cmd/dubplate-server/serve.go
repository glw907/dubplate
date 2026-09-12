package main

import (
	"io"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/spf13/cobra"
)

// serveBindAddrsThreshold, servePortThreshold, and servePrefixThreshold
// are the rung 5a serve shell's own three keys, declared in this file
// rather than internal/config (R-C, following rung 4's
// registerShadowThresholds precedent in shadow.go): the alternative,
// declaring inside internal/serve, would make the shell import
// internal/config, which it has no reason to do.
//
// The three keys replace internal/config/keys.go's own "serve" reserved
// class, deleted in the same commit as this declaration (the removal
// half of the rung 1b registry contract). That class's own comment
// carried one argument forward: a bind address is not an identity
// concern, which is why the wall named it its own class apart from
// "auth and identity" rather than folding it in.
//
// ruled, conductor 2026-09-08, R-A, R-B, R-C

// serveBindAddrsThreshold is the comma-separated list of addresses the
// serve verb binds. Its default, loopback only, is the salvage
// configuration's own default (env/dubplate.env.template:104-113). The
// registry has no list kind (R-B), so this registers as KindString and
// [serve.SplitBindAddrs] splits and validates it in the package that
// binds, keeping the address grammar with the code that uses it.
var serveBindAddrsThreshold = config.Threshold{
	Key:     "DUBPLATE_SERVE_BIND_ADDRS",
	Kind:    config.KindString,
	Default: "127.0.0.1",
	Doc:     "a comma-separated list of bind addresses, for example 127.0.0.1,::1",
}

// servePortThreshold is the TCP port the serve verb binds on every
// configured address. Its default, 8080, is the salvage configuration's
// own default (_salvage/internal/config/config_test.go:44-45).
var servePortThreshold = config.Threshold{
	Key:     "DUBPLATE_SERVE_PORT",
	Kind:    config.KindInt,
	Default: "8080",
	Doc:     "an integer TCP port",
}

// servePrefixThreshold is B3's one owned path prefix: every route rung
// 5 through rung 14 register, and 2.0's OpenSubsonic surface under the
// same 2.0 hook, lives under it, and everything else is the
// fallthrough. Its default, /dubplate, collides with none of
// Navidrome's own surface (/rest, /share, /app, /auth), is one segment
// so the segment-aware match has a boundary to find, and carries no
// estate value.
var servePrefixThreshold = config.Threshold{
	Key:     "DUBPLATE_SERVE_PREFIX",
	Kind:    config.KindString,
	Default: "/dubplate",
	Doc:     "a single path segment beginning with a separator, for example /dubplate",
}

// registerServeConfig adds the serve shell's three thresholds to reg,
// a plain function on registerShadowThresholds' own precedent
// (shadow.go): newRootCmd builds a fresh, independent registry on
// every construction, so registering onto it can never collide with a
// prior process's own registration.
func registerServeConfig(reg *config.Registry) {
	reg.AddThreshold(serveBindAddrsThreshold)
	reg.AddThreshold(servePortThreshold)
	reg.AddThreshold(servePrefixThreshold)
}

// resolveServeConfig reads the three thresholds against environ and
// returns the values [serve.Run] and its construction values take: the
// validated owned prefix, the split and validated bind addresses, and
// the port. It returns before any listener opens, so a caller can
// refuse an invalid prefix or bind list without ever touching the
// network.
func resolveServeConfig(environ []string) (serve.Prefix, []string, int, error) {
	prefixValue, err := config.ParseString(environ, servePrefixThreshold)
	if err != nil {
		return serve.Prefix{}, nil, 0, err
	}
	prefix, err := serve.NewPrefix(prefixValue)
	if err != nil {
		return serve.Prefix{}, nil, 0, err
	}

	bindRaw, err := config.ParseString(environ, serveBindAddrsThreshold)
	if err != nil {
		return serve.Prefix{}, nil, 0, err
	}
	addrs, err := serve.SplitBindAddrs(bindRaw)
	if err != nil {
		return serve.Prefix{}, nil, 0, err
	}

	port, err := config.ParseInt(environ, servePortThreshold)
	if err != nil {
		return serve.Prefix{}, nil, 0, err
	}

	return prefix, addrs, port, nil
}

// newServeCmd builds the serve verb: it validates the owned
// configuration wall through [config.Load] against reg, resolves the
// three thresholds through resolveServeConfig, builds the rung 5a
// shell with no registered subtree and no fallthrough (no driver
// exists yet, so every request outside the owned prefix answers 404),
// and runs it until cmd's context is done. The shell is built by
// newServeShell, which carries the one construction value that exists
// at this rung: the session cookie's name.
//
// Load runs before anything else and before any listener opens: a
// server whose wall is unsatisfied must refuse at startup, in the same
// words doctor reports, rather than accept connections and fail at the
// first request that needs a missing value.
func newServeCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "serve",
		Short:         "Run the HTTP server",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if _, err := config.Load(environ, reg); err != nil {
				return err
			}
			prefix, addrs, port, err := resolveServeConfig(environ)
			if err != nil {
				return err
			}
			shell := newServeShell(prefix)
			return serve.Run(cmd.Context(), shell.Logger(), shell.Handler(), addrs, port)
		},
	}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	return cmd
}

// newServeShell builds the shell the serve verb runs, owning prefix.
// The strip carries auth.SessionCookieName, so the shell removes the
// server's own session cookie from any request it hands the
// fallthrough: no request leaving this server's owned surface carries
// the credential that authenticates it. Rung 5a's coverage gap 4 was
// exactly this name being the empty string, which removes nothing.
//
// The remaining construction values stay empty because nothing supplies
// them yet: no driver exists to name an identity header, no trusted
// proxy is configured, and no front exemption is in force.
//
// The shell is built here rather than inline in RunE so the verb's own
// wiring is what a test exercises, rather than a restatement of it.
func newServeShell(prefix serve.Prefix) *serve.Shell {
	strip := serve.NewStrip("", auth.SessionCookieName, serve.TrustedProxyConfig{}, serve.FrontExemption{})
	return serve.NewShell(prefix, strip, nil)
}
