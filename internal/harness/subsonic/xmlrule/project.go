package xmlrule

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

// documentError is a body the projection cannot walk: malformed, or
// written in a form it does not support. ordinal locates it, 0 when no
// element can carry it.
type documentError struct {
	ordinal int
	msg     string
}

func (e *documentError) Error() string { return e.msg }

// projection is a body with its OpenSubsonic additions and residuals
// dropped, ready for xmllint.
type projection struct {
	doc []byte
	// lines maps each line of doc that opens an element to that
	// element's ordinal in the original body: its 1-based position in
	// document order, counting every element, dropped ones included.
	lines     map[int]int
	residuals []Violation
	env       envelope
}

// xmlSpace reports whether b holds only XML whitespace. bytes.TrimSpace
// would also strip Unicode spaces such as U+00A0, which are content.
func xmlSpace(b []byte) bool {
	return len(bytes.Trim(b, " \t\r\n")) == 0
}

// prefixed returns the first namespace prefix an element of body carries,
// or "" when none does. RawToken leaves the prefix in Name.Space, where
// Token would replace it with the namespace URL.
func prefixed(body []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(body))
	for {
		tok, err := dec.RawToken()
		if errors.Is(err, io.EOF) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		if start, ok := tok.(xml.StartElement); ok && start.Name.Space != "" {
			return start.Name.Space, nil
		}
	}
}

// project walks body from its root, resolving each element's type by its
// position: the root by the schema's global elements, every other element
// by its parent's declaration of it. A name the position's type declares
// stays. A name its OpenAPI counterpart adds is dropped. Any other name is
// a residual: reported, then dropped, so that xmllint, which stops at a
// parent's first unexpected child, validates everything after it. A
// dropped element takes its subtree with it.
//
// Each element starts a new line of the projected document, and the line
// table counts the newlines actually written, since end tags and text
// carrying a newline take lines too.
func (s *schema) project(body []byte) (*projection, error) {
	if len(body) == 0 {
		return nil, &documentError{msg: "empty body"}
	}
	prefix, err := prefixed(body)
	if err != nil {
		return nil, &documentError{msg: "malformed XML: " + err.Error()}
	}
	if prefix != "" {
		return nil, &documentError{ordinal: 1,
			msg: fmt.Sprintf("element prefix %q is not supported by the projection", prefix)}
	}

	p := &projection{lines: make(map[int]int)}
	var out bytes.Buffer
	enc := xml.NewEncoder(&out)
	enc.Indent("", "  ")
	newlines, counted := 0, 0
	flush := func() error {
		if err := enc.Flush(); err != nil {
			return err
		}
		newlines += bytes.Count(out.Bytes()[counted:], []byte("\n"))
		counted = out.Len()
		return nil
	}

	// stack holds the type of each open element the projection writes;
	// nil copies an element and its content without scoping, for a root
	// the schema does not declare.
	var stack []*complexType
	ordinal, skip, roots := 0, 0, 0
	dec := xml.NewDecoder(bytes.NewReader(body))
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, &documentError{msg: "malformed XML: " + err.Error()}
		}
		switch t := tok.(type) {
		case xml.StartElement:
			ordinal++
			if skip > 0 {
				skip++
				continue
			}
			var typ *complexType
			switch len(stack) {
			case 0:
				roots++
				if roots > 1 {
					return nil, &documentError{ordinal: ordinal, msg: "malformed XML: a second root element"}
				}
				p.env.status = rootStatus(t.Attr)
				typ = s.root(t.Name.Local)
			case 1:
				p.env.children = append(p.env.children, t.Name.Local)
				fallthrough
			default:
				var keep bool
				if typ, keep = p.place(s, stack[len(stack)-1], ordinal, t.Name.Local); !keep {
					skip = 1
					continue
				}
			}
			start := xml.StartElement{Name: t.Name, Attr: p.scopeAttrs(ordinal, t.Attr, typ)}
			if err := enc.EncodeToken(start); err != nil {
				return nil, fmt.Errorf("write projection: %w", err)
			}
			if err := flush(); err != nil {
				return nil, fmt.Errorf("write projection: %w", err)
			}
			p.lines[newlines+1] = ordinal
			stack = append(stack, typ)
		case xml.EndElement:
			if skip > 0 {
				skip--
				continue
			}
			stack = stack[:len(stack)-1]
			if err := enc.EncodeToken(xml.EndElement{Name: t.Name}); err != nil {
				return nil, fmt.Errorf("write projection: %w", err)
			}
		case xml.CharData:
			if skip > 0 || xmlSpace(t) {
				continue
			}
			if len(stack) == 0 {
				return nil, &documentError{msg: "malformed XML: character data outside the root element"}
			}
			if err := enc.EncodeToken(t); err != nil {
				return nil, fmt.Errorf("write projection: %w", err)
			}
		}
	}
	if roots == 0 {
		return nil, &documentError{msg: "malformed XML: no root element"}
	}
	if err := flush(); err != nil {
		return nil, fmt.Errorf("write projection: %w", err)
	}
	p.doc = out.Bytes()
	return p, nil
}

// place resolves the type of the element name at ordinal under parent.
// keep is false when the element is dropped, as an addition or as a
// residual the projection records.
func (p *projection) place(s *schema, parent *complexType, ordinal int, name string) (typ *complexType, keep bool) {
	if parent == nil {
		return nil, true
	}
	if declared, ok := parent.children[name]; ok {
		return s.resolve(declared), true
	}
	if !parent.additions[name] {
		p.residuals = append(p.residuals, residual(ordinal, "", "element", name, parent))
	}
	return nil, false
}

func rootStatus(attrs []xml.Attr) string {
	for _, a := range attrs {
		if a.Name.Space == "" && a.Name.Local == "status" {
			return a.Value
		}
	}
	return ""
}

// scopeAttrs returns the attributes of the element at ordinal that typ
// declares, dropping its additions and recording its residuals. Namespace
// declarations are dropped; the encoder writes the element's own.
func (p *projection) scopeAttrs(ordinal int, attrs []xml.Attr, typ *complexType) []xml.Attr {
	var kept []xml.Attr
	for _, a := range attrs {
		if a.Name.Space == "xmlns" || (a.Name.Space == "" && a.Name.Local == "xmlns") {
			continue
		}
		switch {
		case typ == nil:
			kept = append(kept, a)
		case a.Name.Space == "" && typ.attrs[a.Name.Local]:
			kept = append(kept, a)
		case a.Name.Space == "" && typ.additions[a.Name.Local]:
		default:
			p.residuals = append(p.residuals, residual(ordinal, a.Name.Local, "attribute", a.Name.Local, typ))
		}
	}
	return kept
}

func residual(ordinal int, attr, kind, name string, typ *complexType) Violation {
	msg := fmt.Sprintf("%s %q is declared by neither schema type %s nor its OpenAPI counterpart", kind, name, typ.name)
	if typ == simpleType {
		msg = fmt.Sprintf("%s %q sits on an element of simple type, which declares none", kind, name)
	}
	return Violation{Ordinal: ordinal, Attr: attr, Message: msg, Source: Residual}
}
