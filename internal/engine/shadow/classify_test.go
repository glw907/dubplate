package shadow

import (
	"errors"
	"fmt"
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/musicbrainz"
	"github.com/glw907/dubplate/internal/engine/tag"
)

const (
	relR = "11111111-1111-1111-1111-111111111111"
	relG = "22222222-2222-2222-2222-222222222222"
)

// applyProposal is one album tag.Album applied: a strong recommendation
// and the winning release first in Candidates.
func applyProposal(albumID string) *tag.Proposal {
	return &tag.Proposal{
		Candidates:     []matcher.Candidate{{Info: &matcher.AlbumInfo{AlbumID: albumID}}},
		Recommendation: matcher.RecStrong,
		Action:         matcher.ActionApply,
	}
}

// skipProposal is one album tag.Album declined at rec, carrying the
// candidate it declined. A medium or low recommendation still names a
// best candidate, which is what makes the recommendation and the action
// separable predicates under the estate tunables (mutation 3).
func skipProposal(rec matcher.Recommendation, albumID string) *tag.Proposal {
	return &tag.Proposal{
		Candidates:     []matcher.Candidate{{Info: &matcher.AlbumInfo{AlbumID: albumID}}},
		Recommendation: rec,
		Action:         matcher.ActionSkip,
	}
}

func TestClassifyEveryOutcome(t *testing.T) {
	oracleApplied := OracleSignal{Signal: "apply", ReleaseMBID: relR}
	oracleSkipped := OracleSignal{Signal: "skip"}

	tests := []struct {
		name         string
		oracle       OracleSignal
		proposal     *tag.Proposal
		err          error
		wantClass    string
		wantDeferred string
	}{
		{"same release agrees", oracleApplied, applyProposal(relR), nil, ClassAgreement, ""},
		{"both skip agrees", oracleSkipped, skipProposal(matcher.RecNone, relG), nil, ClassAgreement, ""},
		{"a medium recommendation is a skip, not an apply", oracleSkipped, skipProposal(matcher.RecMedium, relG), nil, ClassAgreement, ""},
		{"no candidates agrees with a skip", oracleSkipped, nil, fmt.Errorf("tag: %w", tag.ErrNoCandidates), ClassAgreement, ""},
		{"go applies where the oracle skipped", oracleSkipped, applyProposal(relG), nil, ClassGoAppliesOracleSkipped, ""},
		{"go skips where the oracle applied", oracleApplied, skipProposal(matcher.RecMedium, relG), nil, ClassGoSkipsOracleApplied, ""},
		{"no candidates where the oracle applied", oracleApplied, nil, fmt.Errorf("tag: %w", tag.ErrNoCandidates), ClassGoSkipsOracleApplied, ""},
		{"different release", oracleApplied, applyProposal(relG), nil, ClassGoAppliesDifferentRelease, ""},
		{"unavailable defers", oracleApplied, nil, fmt.Errorf("tag: looking up %s: %w", relR, musicbrainz.ErrUnavailable), "", DeferUnavailable},
		{"budget exhausted defers", oracleApplied, nil, fmt.Errorf("tag: searching: %w", musicbrainz.ErrBudgetExhausted), "", DeferBudgetExhausted},
		{"limiter closed defers", oracleSkipped, nil, fmt.Errorf("tag: searching: %w", musicbrainz.ErrLimiterClosed), "", DeferLimiterClosed},
		// Contract S6 made mechanical: an unrecognized error must never
		// become a Go skip, because a truncated album written as
		// go-skips-oracle-applied is a fabricated divergence and it
		// consumes exactly the owner adjudication ruling B's stamp
		// exists to save.
		{"an unrecognized error defers rather than becoming a skip", oracleApplied, nil, errors.New("truncated album"), "", DeferUnclassified},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			class, deferred := classify(tt.oracle, tt.proposal, tt.err)
			if class != tt.wantClass || deferred != tt.wantDeferred {
				t.Errorf("classify = (%q, %q), want (%q, %q)", class, deferred, tt.wantClass, tt.wantDeferred)
			}
		})
	}
}

func TestClassifyUsesErrorsIsRatherThanEquality(t *testing.T) {
	// tag.Album wraps every client error in a message, in two shapes:
	// "tag: looking up %s: %w" (tag.go:170) and "tag: searching: %w"
	// (tag.go:128). An equality test passes a bare sentinel and fails
	// every wrapped one, so each sentinel is asserted double-wrapped in
	// both shapes.
	sentinels := []struct {
		name string
		err  error
		want string
	}{
		{"unavailable", musicbrainz.ErrUnavailable, DeferUnavailable},
		{"budget exhausted", musicbrainz.ErrBudgetExhausted, DeferBudgetExhausted},
		{"limiter closed", musicbrainz.ErrLimiterClosed, DeferLimiterClosed},
	}
	wrappings := []struct {
		name string
		wrap func(error) error
	}{
		{"looking up", func(err error) error {
			return fmt.Errorf("tag: looking up %s: %w", relR, fmt.Errorf("musicbrainz: release %s: %w", relR, err))
		}},
		{"searching", func(err error) error {
			return fmt.Errorf("tag: searching: %w", fmt.Errorf("musicbrainz: search: %w", err))
		}},
	}

	for _, s := range sentinels {
		for _, w := range wrappings {
			t.Run(s.name+" "+w.name, func(t *testing.T) {
				class, deferred := classify(OracleSignal{Signal: "apply", ReleaseMBID: relR}, nil, w.wrap(s.err))
				if class != "" || deferred != s.want {
					t.Errorf("classify = (%q, %q), want (\"\", %q)", class, deferred, s.want)
				}
			})
		}
	}
}

func TestClassifyVerdictReadsTheStoredAction(t *testing.T) {
	// The re-score's annotation runs the same table over a stored
	// GoVerdict rather than a live proposal, so the two paths cannot
	// drift on what "applied" means.
	tests := []struct {
		name    string
		oracle  OracleSignal
		verdict GoVerdict
		want    string
	}{
		{"a stored apply of the oracle's own release agrees", OracleSignal{Signal: "apply", ReleaseMBID: relR},
			GoVerdict{Action: string(matcher.ActionApply), AlbumID: relR}, ClassAgreement},
		{"a stored skip against an oracle apply diverges", OracleSignal{Signal: "apply", ReleaseMBID: relR},
			GoVerdict{Action: string(matcher.ActionSkip), AlbumID: relG}, ClassGoSkipsOracleApplied},
		{"a stored skip agrees with an oracle skip", OracleSignal{Signal: "skip"},
			GoVerdict{Action: string(matcher.ActionSkip), AlbumID: relG}, ClassAgreement},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifyVerdict(tt.oracle, &tt.verdict); got != tt.want {
				t.Errorf("classifyVerdict = %q, want %q", got, tt.want)
			}
		})
	}
}
