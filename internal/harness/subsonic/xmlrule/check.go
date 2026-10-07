package xmlrule

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	// ErrUnknownEndpoint reports an endpoint the OpenAPI document
	// declares no path for.
	ErrUnknownEndpoint = errors.New("xmlrule: endpoint not in the openapi document")

	// ErrNotXML reports a binary success the operation declares: it is
	// not an input to the XML tier.
	ErrNotXML = errors.New("xmlrule: not an xml response")
)

// Outcome is the result class of one check. Only [Valid] is a pass.
type Outcome string

const (
	// Valid means the projected body satisfies the schema, and the body
	// fits its endpoint.
	Valid Outcome = "valid"
	// Invalid means the result carries at least one violation.
	Invalid Outcome = "invalid"
	// Unvalidated means the body's payload element is outside the
	// schema's choice, so no authority covers it. It is never a pass.
	Unvalidated Outcome = "declared-unvalidated"
)

// Source names the check a violation came from.
type Source string

const (
	// Document marks a body the projection cannot walk: empty,
	// malformed, or namespace-prefixed.
	Document Source = "document"
	// Binding marks a body that does not fit its endpoint, HTTP status,
	// or content type.
	Binding Source = "binding"
	// Schema marks a validity error xmllint reported.
	Schema Source = "xmllint"
	// Residual marks a name declared by neither the schema type at its
	// position nor that type's OpenAPI counterpart.
	Residual Source = "residual"
)

// sourceRank orders violations sharing an ordinal.
var sourceRank = map[Source]int{Document: 0, Binding: 1, Schema: 2, Residual: 3}

// Violation is one place a body departs from the XML contract.
type Violation struct {
	// Ordinal is the offending element's 1-based position in the
	// original body's document order; 1 is the envelope, and 0 means no
	// element could be located.
	Ordinal int
	// Attr is the attribute the violation names, or "".
	Attr    string
	Message string
	Source  Source
}

// Result is the outcome of one check. Violations is set only for
// [Invalid], sorted by ordinal; Element names the uncovered payload
// element of an [Unvalidated] body.
type Result struct {
	Outcome    Outcome
	Violations []Violation
	Element    string
}

// Validator checks XML responses against the Subsonic XSD, scoped by the
// OpenAPI document's additions, and against each endpoint's binding.
type Validator struct {
	schema     *schema
	schemaPath string
	// choice holds the payload elements the envelope's type admits.
	choice   map[string]bool
	bindings map[string]binding
	xmllint  *runner
}

// New returns a validator over the cached XSD and the OpenAPI document in
// openapi, which the caller has verified. It fails when the XSD is not
// cached or fails its pin, when the document lacks a schema the rename
// table names, and when xmllint is absent or not the pinned version.
func New(openapi []byte) (*Validator, error) {
	xsd, err := Cached()
	if err != nil {
		return nil, err
	}
	path, err := CachePath()
	if err != nil {
		return nil, err
	}
	run, err := systemXmllint()
	if err != nil {
		return nil, err
	}
	return newValidator(xsd, path, openapi, renames, run)
}

func newValidator(xsd []byte, xsdPath string, openapi []byte, renames map[string]string, run *runner) (*Validator, error) {
	s, err := parseSchema(xsd)
	if err != nil {
		return nil, err
	}
	doc, err := openapi3.NewLoader().LoadFromData(openapi)
	if err != nil {
		return nil, fmt.Errorf("load openapi document: %w", err)
	}
	if err := s.addAdditions(doc, renames); err != nil {
		return nil, err
	}
	bindings, err := readBindings(doc, s.envelopeNames(envelopeMember))
	if err != nil {
		return nil, err
	}
	return &Validator{
		schema:     s,
		schemaPath: xsdPath,
		choice:     s.choice(envelopeMember),
		bindings:   bindings,
		xmllint:    run,
	}, nil
}

// Check validates one XML response to endpoint. A body that does not fit
// the endpoint is checked in full besides, so its schema errors are
// reported with the binding's. A failed xmllint run wraps
// [ErrToolFailure].
func (v *Validator) Check(ctx context.Context, endpoint string, status int, contentType string, body []byte) (Result, error) {
	b, ok := v.bindings[endpoint]
	if !ok {
		return Result{}, fmt.Errorf("%w: %s", ErrUnknownEndpoint, endpoint)
	}
	if b.binary && !xmlMedia(contentType) {
		return Result{}, fmt.Errorf("%w: %s answered %q", ErrNotXML, endpoint, contentType)
	}
	p, err := v.schema.project(body)
	if derr, ok := errors.AsType[*documentError](err); ok {
		return invalid([]Violation{{Ordinal: derr.ordinal, Message: derr.msg, Source: Document}}), nil
	}
	if err != nil {
		return Result{}, err
	}

	violations := b.check(endpoint, status, contentType, p.env, v.choice)
	if len(violations) == 0 {
		if elem := b.unvalidated(p.env, v.choice); elem != "" {
			return Result{Outcome: Unvalidated, Element: elem}, nil
		}
	}

	errs, err := v.xmllint.validate(ctx, v.schemaPath, p.doc)
	if err != nil {
		return Result{}, err
	}
	for _, e := range errs {
		violations = append(violations, Violation{Ordinal: p.lines[e.line], Attr: e.attr, Message: e.msg, Source: Schema})
	}
	violations = append(violations, p.residuals...)
	if len(violations) == 0 {
		return Result{Outcome: Valid}, nil
	}
	return invalid(violations), nil
}

func invalid(vs []Violation) Result {
	slices.SortStableFunc(vs, func(a, b Violation) int {
		return cmp.Or(
			cmp.Compare(a.Ordinal, b.Ordinal),
			cmp.Compare(sourceRank[a.Source], sourceRank[b.Source]),
		)
	})
	return Result{Outcome: Invalid, Violations: vs}
}
