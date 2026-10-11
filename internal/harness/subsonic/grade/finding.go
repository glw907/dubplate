// Package grade compares what several servers answered to one recorded
// session and applies the divergence ledger to the differences found.
//
// A Finding is one difference, or one failed check, in a session. The
// ledger adjudicates findings: a row says which differences are known and
// why, and a row that matches nothing is itself an error, so the ledger
// cannot outlive the behavior it describes.
//
// The golden grades the grader, not the oracles. A ledger row never names
// a difference that exists between the oracles alone; only a row naming
// dubplate as one side is live.
package grade

import "github.com/glw907/dubplate/internal/opensubsonic/client"

// Kind names what a finding found.
type Kind string

const (
	// KindValue is a leaf whose two values differ.
	KindValue Kind = "value"
	// KindType is a leaf whose two sides differ in type or XML form.
	KindType Kind = "type"
	// KindPresence is a member one side has and the other lacks.
	KindPresence Kind = "presence"
	// KindOrder is an order-significant array whose two orders differ.
	KindOrder Kind = "order"
	// KindIdentity is an id that pairs with two ids, or an entity that
	// pairs with none.
	KindIdentity Kind = "identity"
	// KindMissingSource is a request only one source answered.
	KindMissingSource Kind = "missing-source"
	// KindParse is a body that does not decode.
	KindParse Kind = "parse"
	// KindContract is a body the OpenAPI document does not admit.
	KindContract Kind = "contract"
)

// Finding is one difference between sources, or one source's contract
// violation.
//
// Path is a "/"-separated list of segment names with no leading slash;
// an XML attribute and a child element take the same segment name as the
// JSON key, and an array element is its decimal index. A finding from the
// oracle tier carries the graded pair in Sources and one value per
// source in Values; a [KindContract] finding carries the one source whose
// body failed and that body's value.
type Finding struct {
	// Request is the [session.Request.Key] of the request.
	Request  string
	Endpoint string
	Format   client.Format
	Sources  []string
	Kind     Kind
	Path     string
	Values   []string
}
