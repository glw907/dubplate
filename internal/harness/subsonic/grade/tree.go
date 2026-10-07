package grade

import (
	"bytes"
	"cmp"
	"encoding/json/jsontext"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/opensubsonic/client"
)

// nodeKind is the shape of one node of the comparable tree.
type nodeKind uint8

const (
	kindObject nodeKind = iota // a JSON object, or an XML element with attributes or children
	kindArray
	kindString
	kindNumber
	kindBool
	kindNull
	kindText // an XML attribute value or an element's text
)

// xmlForm records whether an XML leaf was an attribute or an element.
// JSON nodes carry formNone.
type xmlForm uint8

const (
	formNone xmlForm = iota
	formAttr
	formElem
)

// node is one value of the comparable tree both formats decode into. A
// scalar keeps its literal: a JSON number is its source text, never a
// float, so 2 and 2.0 differ.
type node struct {
	kind   nodeKind
	form   xmlForm
	lit    string
	fields map[string]*node
	items  []*node
}

// field returns the literal of the scalar member name, or "" when n has
// no such scalar member.
func (n *node) field(name string) string {
	if n == nil || n.kind != kindObject {
		return ""
	}
	f := n.fields[name]
	if f == nil || f.kind == kindObject || f.kind == kindArray {
		return ""
	}
	return f.lit
}

// render is a node's value as a finding prints it.
// title is an entity's display title: its name, or its title when it has
// no name.
func (n *node) title() string { return cmp.Or(n.field("name"), n.field("title")) }

func (n *node) render() string {
	if n == nil {
		return "absent"
	}
	switch n.kind {
	case kindObject:
		if n.form == formElem {
			return "element"
		}
		return "object"
	case kindArray:
		return "array of " + strconv.Itoa(len(n.items))
	case kindString:
		return strconv.Quote(n.lit)
	case kindText:
		if n.form == formAttr {
			return "attribute " + strconv.Quote(n.lit)
		}
		return "element " + strconv.Quote(n.lit)
	}
	return n.lit
}

// canonical is a node's full value as one string, used to pair array
// elements that carry no key of their own.
func (n *node) canonical() string {
	var b strings.Builder
	n.writeCanonical(&b)
	return b.String()
}

func (n *node) writeCanonical(b *strings.Builder) {
	switch n.kind {
	case kindObject:
		b.WriteByte('{')
		for _, k := range slices.Sorted(maps.Keys(n.fields)) {
			b.WriteString(strconv.Quote(k) + ":")
			n.fields[k].writeCanonical(b)
			b.WriteByte(',')
		}
		b.WriteByte('}')
	case kindArray:
		b.WriteByte('[')
		for _, it := range n.items {
			it.writeCanonical(b)
			b.WriteByte(',')
		}
		b.WriteByte(']')
	default:
		b.WriteString(n.render())
	}
}

// decode parses a body in format into its comparable tree.
func decode(body []byte, format client.Format) (*node, error) {
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, errors.New("empty body")
	}
	if format == client.XML {
		return decodeXML(body)
	}
	return decodeJSON(body)
}

func decodeJSON(body []byte) (*node, error) {
	dec := jsontext.NewDecoder(bytes.NewReader(body))
	n, err := readJSON(dec)
	if err != nil {
		return nil, err
	}
	_, err = dec.ReadToken()
	if err == nil {
		return nil, errors.New("data after the top-level value")
	}
	if !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("data after the top-level value: %w", err)
	}
	return n, nil
}

func readJSON(dec *jsontext.Decoder) (*node, error) {
	tok, err := dec.ReadToken()
	if err != nil {
		return nil, err
	}
	switch tok.Kind() {
	case '{':
		n := &node{kind: kindObject, fields: map[string]*node{}}
		for dec.PeekKind() != '}' {
			name, err := dec.ReadToken()
			if err != nil {
				return nil, err
			}
			// A token is void after the next read, so the name is copied
			// out before the value is read.
			key := name.String()
			v, err := readJSON(dec)
			if err != nil {
				return nil, err
			}
			n.fields[key] = v
		}
		_, err := dec.ReadToken()
		return n, err
	case '[':
		n := &node{kind: kindArray}
		for dec.PeekKind() != ']' {
			v, err := readJSON(dec)
			if err != nil {
				return nil, err
			}
			n.items = append(n.items, v)
		}
		_, err := dec.ReadToken()
		return n, err
	case '"':
		return &node{kind: kindString, lit: tok.String()}, nil
	case '0':
		return &node{kind: kindNumber, lit: tok.String()}, nil
	case 't', 'f':
		return &node{kind: kindBool, lit: tok.String()}, nil
	case 'n':
		return &node{kind: kindNull, lit: "null"}, nil
	}
	return nil, fmt.Errorf("unexpected token %v", tok.Kind())
}

// decodeXML maps an XML document onto the tree: the root element becomes
// the one member of a root object, an attribute and a child element both
// become members named by their local name, a name repeated among one
// element's children becomes an array, and an element with neither
// attributes nor children becomes a text leaf. Namespace declarations
// carry no data and are dropped, and whitespace between child elements is
// layout, not text.
func decodeXML(body []byte) (*node, error) {
	dec := xml.NewDecoder(bytes.NewReader(body))
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		if start, ok := tok.(xml.StartElement); ok {
			n, err := readElement(dec, start)
			if err != nil {
				return nil, err
			}
			return &node{kind: kindObject, form: formElem, fields: map[string]*node{start.Name.Local: n}}, nil
		}
	}
}

func readElement(dec *xml.Decoder, start xml.StartElement) (*node, error) {
	fields := map[string]*node{}
	for _, a := range start.Attr {
		if a.Name.Space == "xmlns" || (a.Name.Space == "" && a.Name.Local == "xmlns") {
			continue
		}
		fields[a.Name.Local] = &node{kind: kindText, form: formAttr, lit: a.Value}
	}
	attrs := len(fields)
	var text strings.Builder
	children := 0
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.CharData:
			text.Write(t)
		case xml.StartElement:
			child, err := readElement(dec, t)
			if err != nil {
				return nil, err
			}
			name := t.Name.Local
			children++
			switch prev := fields[name]; {
			case prev == nil:
				fields[name] = child
			case prev.form == formAttr:
				return nil, fmt.Errorf("element %s has an attribute and a child both named %s", start.Name.Local, name)
			case prev.kind == kindArray:
				prev.items = append(prev.items, child)
			default:
				fields[name] = &node{kind: kindArray, form: formElem, items: []*node{prev, child}}
			}
		case xml.EndElement:
			if attrs == 0 && children == 0 {
				return &node{kind: kindText, form: formElem, lit: text.String()}, nil
			}
			if s := strings.TrimSpace(text.String()); s != "" {
				if fields["value"] != nil {
					return nil, fmt.Errorf("element %s has text and a member named value", start.Name.Local)
				}
				fields["value"] = &node{kind: kindText, form: formElem, lit: text.String()}
			}
			return &node{kind: kindObject, form: formElem, fields: fields}, nil
		}
	}
}
