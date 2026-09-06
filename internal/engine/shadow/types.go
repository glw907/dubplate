package shadow

import (
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// Divergence classes and the agreement, the closed vocabulary contract
// S4 fixes. A class outside this set is a bug, never a new row.
const (
	ClassAgreement                 = "agreement"
	ClassGoAppliesOracleSkipped    = "go-applies-oracle-skipped"
	ClassGoSkipsOracleApplied      = "go-skips-oracle-applied"
	ClassGoAppliesDifferentRelease = "go-applies-different-release"
)

// Deferral reasons. Each names a musicbrainz sentinel a caller branches
// on to STOP ASKING rather than to retry (limiter.go:45-53,
// client.go:74-84). DeferUnclassified is the closed enumeration's
// default arm: an error the classifier does not recognize defers and is
// counted, never recorded as a verdict (Resolution R15).
const (
	DeferUnavailable     = "musicbrainz-unavailable"
	DeferBudgetExhausted = "budget-exhausted"
	DeferLimiterClosed   = "limiter-closed"
	DeferUnclassified    = "unclassified-error"
)

// Unreconciled reasons, contract S11's four.
//
// UnreconciledNoReleaseID covers both halves of one honest gap: an
// imported outcome whose LibraryPaths is empty, and one whose named
// library file carries no MUSICBRAINZ_ALBUMID. Either way the oracle
// named no release, and recording an apply against an empty release
// would classify as go-applies-different-release against any Go apply,
// which is a divergence against a release the oracle never named
// (Resolution R22).
const (
	UnreconciledAmbiguousBasename = "ambiguous-basename"
	UnreconciledNoReleaseID       = "imported-with-no-release-id"
	UnreconciledReportSwept       = "report-swept"
	UnreconciledPartialSkip       = "partial-skip"
)

// Exclusion reasons, counted and never scored. Each names a directory
// the shadow saw and deliberately did not observe, so the report prints
// a count with a reason rather than silence.
const (
	ExcludedMultiAlbum   = "multi-album-directory"
	ExcludedGateRejected = "gate-rejected"
	ExcludedNoTracks     = "no-tracks"
)

// A CapturedFile is one file the capture listed, in the digest's own
// terms.
type CapturedFile struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

// A Capture is one album directory as the shadow saw it before the
// incumbent imported it. Digest is the identity (contract S7).
type Capture struct {
	Uploader   string               `json:"uploader"`
	Basename   string               `json:"basename"`
	Digest     string               `json:"digest"`
	Files      []CapturedFile       `json:"files"`
	Items      []matcher.ItemFields `json:"items"`
	CapturedAt time.Time            `json:"captured_at"`
}

// An OracleSignal is the incumbent's own verdict for one album, and the
// artifact it was read from. Signal is "apply" or "skip" and nothing
// else: the oracle is binary (contract S3).
type OracleSignal struct {
	Signal        string `json:"signal"`
	ReleaseMBID   string `json:"release_mbid,omitempty"`
	Uploader      string `json:"uploader"`
	AlbumBasename string `json:"album_basename"`
	ReadFrom      string `json:"read_from"`
}

// A GoVerdict is one scoring mode's answer. Hypothesis marks the
// DefaultCompat re-score, whose verdict is not the product's true
// DefaultCompat verdict wherever the real flow would have widened the
// candidate set (contract S16).
type GoVerdict struct {
	Recommendation string   `json:"recommendation"`
	Action         string   `json:"action"`
	AlbumID        string   `json:"album_id,omitempty"`
	Distance       float64  `json:"distance"`
	NonzeroKeys    []string `json:"nonzero_keys"`
	Hypothesis     bool     `json:"hypothesis,omitempty"`
}

// GoVerdicts is the pair the two scoring modes produce. Default is nil
// when the re-score could not run, which is honest rather than a zero
// verdict a reader would compare against.
type GoVerdicts struct {
	Compat  *GoVerdict `json:"compat"`
	Default *GoVerdict `json:"default,omitempty"`
}

// A BudgetRecord is one pass's MusicBrainz spend, persisted because a
// Budget is in-process state (contract S14).
type BudgetRecord struct {
	Name  string `json:"name"`
	Spent int    `json:"spent"`
}

// An Observation is observations/<obs-id>/observation.json: one album,
// the two verdicts, and the class. Exactly one of Class, Deferred,
// Unreconciled and Excluded is set on a stored record; an observation
// still awaiting its oracle signal carries none of the four.
//
// Excluded is the fourth because one of the three exclusions arrives
// after the record already exists: a multi-album directory and a
// directory with no tracks are excluded at capture and never become
// observations at all, while a gate-rejected album is a single-album
// directory the capture stores and only the run report later reveals as
// quarantined or rejected (fold, B5).
type Observation struct {
	ID                string            `json:"id"`
	Album             string            `json:"album"`
	MatcherCommit     string            `json:"matcher_commit"`
	MatcherConfigHash string            `json:"matcher_config_hash"`
	EvalDate          time.Time         `json:"eval_date"`
	Capture           Capture           `json:"capture"`
	Oracle            *OracleSignal     `json:"oracle,omitempty"`
	Go                *GoVerdicts       `json:"go,omitempty"`
	Class             string            `json:"class,omitempty"`
	Deferred          string            `json:"deferred,omitempty"`
	Unreconciled      string            `json:"unreconciled,omitempty"`
	Excluded          string            `json:"excluded,omitempty"`
	Superseded        bool              `json:"superseded,omitempty"`
	Budget            BudgetRecord      `json:"budget"`
	Requests          []RecordedRequest `json:"requests,omitempty"`
}

// A RecordedRequest is one captured wire request and the file its body
// was stored in, in cassette.Request's own field order so
// cmd/dubplate-server's marshalling is a field copy rather than a
// translation.
type RecordedRequest struct {
	Method    string `json:"method"`
	PathQuery string `json:"path_query"`
	Response  string `json:"response"`
}

// A PassRecord is passes/<pass-id>.json. WireInterval is recorded
// because Budget.Spent alone cannot reconstruct the pace a run used,
// and the pace is a conductor ruling a later reader will want to check.
type PassRecord struct {
	ID            string         `json:"id"`
	StartedAt     time.Time      `json:"started_at"`
	EndedAt       time.Time      `json:"ended_at"`
	MatcherCommit string         `json:"matcher_commit"`
	WireInterval  string         `json:"wire_interval"`
	Captured      int            `json:"captured"`
	PartialReads  int            `json:"partial_reads"`
	Superseded    int            `json:"superseded"`
	Observed      int            `json:"observed"`
	Agreements    int            `json:"agreements"`
	Divergences   map[string]int `json:"divergences"`
	Deferred      map[string]int `json:"deferred"`
	Unreconciled  map[string]int `json:"unreconciled"`
	Excluded      map[string]int `json:"excluded"`
	ReportGaps    int            `json:"report_gaps"`
	Budget        BudgetRecord   `json:"budget"`
}

// A ConsumedReport is one incumbent run report this shadow has already
// reconciled. StartedAt and EndedAt are decoded from the report rather
// than read off the file, because RunID carries no order
// (Resolution R6).
type ConsumedReport struct {
	File      string    `json:"file"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}
