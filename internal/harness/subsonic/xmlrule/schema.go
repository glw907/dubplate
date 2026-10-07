package xmlrule

import (
	"encoding/xml"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// renames maps an XSD complex type to its OpenAPI schema where the two
// names differ. An XSD type absent here maps to the OpenAPI schema of the
// same name, and one with no such schema has no additions.
var renames = map[string]string{
	"Response":          "SubsonicBaseResponse",
	"Error":             "SubsonicError",
	"AlbumWithSongsID3": "AlbumID3WithSongs",
}

// complexType is one XSD complex type with its extension base folded in.
type complexType struct {
	name  string
	attrs map[string]bool
	// children maps a child element's name to its declared type, the
	// unprefixed name of a complex type or "" for a simple type.
	children map[string]string
	// additions holds the names the OpenAPI counterpart declares and the
	// XSD type does not, as attributes or child elements alike.
	additions map[string]bool
}

// schema is the projection's model of an XSD: its complex types and its
// global elements.
type schema struct {
	types map[string]*complexType
	// roots maps a global element's name to its declared type.
	roots map[string]string
}

// simpleType stands for any element of simple type: it declares no
// attribute, no child, and no addition.
var simpleType = &complexType{}

// resolve returns the complex type named by an element's declared type,
// or simpleType when the name is not a complex type.
func (s *schema) resolve(name string) *complexType {
	if t, ok := s.types[name]; ok {
		return t
	}
	return simpleType
}

// root returns the type of the global element name, or nil when the
// schema declares no such element.
func (s *schema) root(name string) *complexType {
	typ, ok := s.roots[name]
	if !ok {
		return nil
	}
	return s.resolve(typ)
}

// choice returns the payload elements the root element's type admits:
// the members of the envelope's choice.
func (s *schema) choice(rootName string) map[string]bool {
	t := s.root(rootName)
	if t == nil {
		return nil
	}
	names := make(map[string]bool, len(t.children))
	for name := range t.children {
		names[name] = true
	}
	return names
}

// envelopeNames returns the names the root element's type carries
// other than its payload: its attributes and its additions.
func (s *schema) envelopeNames(rootName string) map[string]bool {
	names := make(map[string]bool)
	if t := s.root(rootName); t != nil {
		maps.Copy(names, t.attrs)
		maps.Copy(names, t.additions)
	}
	return names
}

// node is a generic XML element, used to walk the XSD.
type node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Nodes   []node     `xml:",any"`
}

func (n node) attr(name string) string {
	for _, a := range n.Attrs {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

// localType strips a QName's prefix: "sub:Child" names the type Child.
func localType(qname string) string {
	_, local, ok := strings.Cut(qname, ":")
	if !ok {
		return qname
	}
	return local
}

// parseSchema reads the complex types and global elements of xsd. Each
// type's extension base is folded into it, so a type's attributes and
// children are the whole set an instance may carry.
func parseSchema(xsd []byte) (*schema, error) {
	var root node
	if err := xml.Unmarshal(xsd, &root); err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}
	s := &schema{types: make(map[string]*complexType), roots: make(map[string]string)}
	bases := make(map[string]string)
	for _, n := range root.Nodes {
		switch n.XMLName.Local {
		case "element":
			s.roots[n.attr("name")] = localType(n.attr("type"))
		case "complexType":
			t := &complexType{name: n.attr("name"), attrs: make(map[string]bool), children: make(map[string]string)}
			collect(n.Nodes, t, bases)
			s.types[t.name] = t
		}
	}
	for name := range s.types {
		if err := s.fold(name, bases, nil); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// collect gathers the attributes, child elements, and extension base
// declared under a complex type, without entering an element's own
// content.
func collect(nodes []node, t *complexType, bases map[string]string) {
	for _, n := range nodes {
		switch n.XMLName.Local {
		case "attribute":
			t.attrs[n.attr("name")] = true
		case "element":
			t.children[n.attr("name")] = localType(n.attr("type"))
		case "extension":
			bases[t.name] = localType(n.attr("base"))
			collect(n.Nodes, t, bases)
		default:
			collect(n.Nodes, t, bases)
		}
	}
}

// fold merges the extension chain above name into it. seen guards a
// cyclic chain.
func (s *schema) fold(name string, bases map[string]string, seen []string) error {
	base, ok := bases[name]
	if !ok {
		return nil
	}
	if slices.Contains(seen, name) {
		return fmt.Errorf("schema type %s extends itself", name)
	}
	b, ok := s.types[base]
	if !ok {
		return fmt.Errorf("schema type %s extends %s, which the schema does not declare", name, base)
	}
	if err := s.fold(base, bases, append(seen, name)); err != nil {
		return err
	}
	t := s.types[name]
	maps.Copy(t.attrs, b.attrs)
	for child, typ := range b.children {
		if _, ok := t.children[child]; !ok {
			t.children[child] = typ
		}
	}
	delete(bases, name)
	return nil
}

// addAdditions sets each type's additions from doc: the property names of
// its OpenAPI counterpart, following $ref and allOf, minus the names the
// XSD type declares. renames names the counterparts whose names differ,
// and each one must exist in doc.
func (s *schema) addAdditions(doc *openapi3.T, renames map[string]string) error {
	var schemas openapi3.Schemas
	if doc.Components != nil {
		schemas = doc.Components.Schemas
	}
	for _, xsdName := range slices.Sorted(maps.Keys(renames)) {
		if _, ok := schemas[renames[xsdName]]; !ok {
			return fmt.Errorf("openapi document has no schema %s, the counterpart of XSD type %s",
				renames[xsdName], xsdName)
		}
	}
	for name, t := range s.types {
		t.additions = make(map[string]bool)
		ref, ok := schemas[name]
		if alias, renamed := renames[name]; renamed {
			ref, ok = schemas[alias], true
		}
		if !ok || ref == nil || ref.Value == nil {
			continue
		}
		for prop := range properties(ref.Value) {
			if !t.attrs[prop] {
				if _, declared := t.children[prop]; !declared {
					t.additions[prop] = true
				}
			}
		}
	}
	return nil
}

// properties returns the property names of an OpenAPI schema, following
// allOf. kin-openapi has already resolved each $ref into its Value.
func properties(sc *openapi3.Schema) map[string]bool {
	names := make(map[string]bool)
	for name := range sc.Properties {
		names[name] = true
	}
	for _, part := range sc.AllOf {
		if part != nil && part.Value != nil {
			maps.Copy(names, properties(part.Value))
		}
	}
	return names
}
