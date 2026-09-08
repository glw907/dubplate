package shadow

import (
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/musicbrainz"
	"github.com/glw907/dubplate/internal/engine/tag"
)

// shadowScoring returns the tunables and compat pair matchAlbum scores
// under: matcher.QuietImportTunables() and matcher.FullCompat() (ruling
// 3). It is its own function, rather than two inline calls
// inside matchAlbum, so a test can hash exactly what matchAlbum used
// against the two constructors' own literal calls: a silent substitute
// (fold, m11's DefaultTunables-with-Quiet-set near miss) moves this
// function's hash without moving any Proposal or Action it produces.
func shadowScoring() (matcher.Tunables, matcher.Compat) {
	return matcher.QuietImportTunables(), matcher.FullCompat()
}

// matchAlbum drives tag.Album once against a capture's items under
// shadowScoring's tunables and compat, through a recording transport,
// spending one token per request from budget.
//
// It returns the proposal, the recorded requests and bodies, and any
// error tag.Album returned, unwrapped classification being the caller's
// (contract S5: classify with errors.Is, never equality, because
// tag.Album wraps every client error).
func matchAlbum(ctx context.Context, cfg Config, budget *musicbrainz.Budget,
	items []matcher.ItemFields, evalDate time.Time) (*tag.Proposal, []RecordedRequest, [][]byte, error) {
	rec := &recordingTransport{base: cmp.Or(cfg.Transport, http.RoundTripper(http.DefaultTransport))}
	client, err := musicbrainz.NewClient(musicbrainz.ClientConfig{
		BaseURL:   cfg.BaseURL,
		UserAgent: cfg.UserAgent,
		Budget:    budget,
		HTTP:      &http.Client{Transport: rec, Timeout: cmp.Or(cfg.Timeout, 30*time.Second)},
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("shadow: building client: %w", err)
	}

	tun, compat := shadowScoring()
	prop, err := tag.Album(ctx, client, items, tun, compat, evalDate, tag.Options{})
	return prop, rec.requests, rec.bodies, err
}

// configHash hashes the Tunables and Compat pair a run scored under.
//
// It is deliberately not named config_hash at the JSON boundary:
// cassette.Expected.ConfigHash is a hash of the Python shim's effective
// config (internal/harness/cassette/cassette.go:80), and conflating the
// two would make a rung 15 comparison silently wrong. The hash runs over
// the JSON rendering of both values, so a field added to either type
// moves the hash without anyone remembering to update this function.
func configHash(tun matcher.Tunables, compat matcher.Compat) string {
	h := sha256.New()
	enc := json.NewEncoder(h)
	// The encoding errors below are unreachable: Tunables and Compat are
	// plain data, no channel, func or cyclic pointer among their fields.
	if err := enc.Encode(tun); err != nil {
		panic(fmt.Sprintf("shadow: encoding Tunables for configHash: %v", err))
	}
	if err := enc.Encode(compat); err != nil {
		panic(fmt.Sprintf("shadow: encoding Compat for configHash: %v", err))
	}
	return hex.EncodeToString(h.Sum(nil))
}
