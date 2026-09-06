package shadow

import (
	"net/http"
	"time"

	"github.com/glw907/dubplate/internal/engine/musicbrainz"
)

// DefaultInterval is the shadow's wire pace: twice
// musicbrainz.DefaultInterval, because the box's own beet import paces
// itself independently and shares the box's single egress IP, where
// MusicBrainz's per-IP budget actually applies (conductor ruling,
// docs/superpowers/specs/2026-09-05-rung-4-shadow.md's conductor
// rulings paragraph).
const DefaultInterval = 2 * musicbrainz.DefaultInterval

// Config is one shadow run's inputs. Every value is a plain parameter:
// this package imports internal/config never, on the same grounds
// internal/engine/matcher's tunables (tunables.go:50-54) and
// internal/engine/musicbrainz's identity (client.go:86-92) state, and
// it reads no clock, so now and evalDate arrive as parameters too.
type Config struct {
	// StateDir is the greenfield state directory. The store lives at
	// StateDir/shadow and nothing outside it is ever written.
	StateDir string

	// ObserveDir, OracleStateDir and ReviewDir are the incumbent's inbox,
	// its own state directory, and its review tree. All three are read
	// only, and OracleStateDir's import.lock is opened never.
	ObserveDir     string
	OracleStateDir string
	ReviewDir      string

	// UserAgent is composed in cmd/dubplate-server, because the identity
	// is site-specific (musicbrainz/client.go:86-92). NewClient refuses
	// an empty one, so the shadow cannot run unidentified.
	UserAgent string

	// MatcherCommit stamps every artifact, so a divergence from a
	// superseded matcher revision is re-evaluated mechanically rather
	// than consuming owner adjudication (ratified ruling B). The
	// subcommand refuses to observe when it is empty or "dev".
	MatcherCommit string

	// BudgetLimit caps one pass's MusicBrainz spend. A limit of zero or
	// less means NO CAP (musicbrainz/limiter.go:130-133), so a caller
	// that means "cap" must pass a positive number.
	BudgetLimit int

	// Interval is the wire pace. Zero means DefaultInterval, the value
	// the subcommand passes.
	Interval time.Duration

	// BaseURL is empty for musicbrainz.DefaultBaseURL. A test points it
	// at an httptest.Server.
	BaseURL string

	// Transport is the round tripper the recorder wraps; nil means
	// http.DefaultTransport.
	Transport http.RoundTripper

	// Timeout is the HTTP client's timeout. It is set here rather than
	// on ClientConfig because ClientConfig.Timeout is ignored entirely
	// when HTTP is non-nil (musicbrainz/client.go:137-144), and the
	// shadow always sets HTTP to install its recorder.
	Timeout time.Duration

	// Limiter is the wire pacer. Nil means Run builds one at Interval
	// and closes it when the pass ends. A test supplies a pre-closed
	// limiter so the ErrLimiterClosed branch is reachable at all, which
	// it otherwise is not: Budget.Wait reserves before it waits
	// (limiter.go:372-381), so a spent cap never consults the limiter
	// and the two deferral branches are mutually exclusive within one
	// pass (Resolution R7).
	Limiter *musicbrainz.Limiter
}
