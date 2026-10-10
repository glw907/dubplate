package session

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/xml"
	"errors"
	"io"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/opensubsonic/client"
)

// credentialLeak is one string of a response body that holds a URL whose
// query carries a credential. It names the parameters and never their
// values, so reporting it does not repeat the secret.
type credentialLeak struct {
	path  string   // the leaf: object members and array indexes, or element and attribute names
	names []string // the credential parameters read from the query, in authParams order
}

func (l credentialLeak) String() string {
	var b strings.Builder
	b.WriteString(l.path)
	b.WriteString(": URL query")
	if len(l.names) > 0 {
		b.WriteString(" carries credential parameter " + strings.Join(l.names, ", "))
	}
	return b.String()
}

// scanCredentials returns one leak per string leaf of body that holds a
// URL query naming a credential parameter, in document order. A body that
// does not decode as format yields the leaks found before the break; the
// envelope check reports the broken body itself.
func scanCredentials(body []byte, format client.Format) []credentialLeak {
	var leaks []credentialLeak
	leaf := func(path, s string) {
		if l, ok := leakIn(s); ok {
			l.path = path
			leaks = append(leaks, l)
		}
	}
	if format == client.XML {
		_ = walkXML(body, leaf)
	} else {
		_ = walkJSON(jsontext.NewDecoder(bytes.NewReader(body)), "", leaf)
	}
	return leaks
}

// leakIn names the credential parameters a leaf carries, when it holds a
// "?". The whole leaf is split on '?', '&' and ';' rather than parsed as one
// query, so prose before a URL and a ';'-joined tail both still name their
// parameters, and a leaf that is not a query names nothing.
func leakIn(leaf string) (credentialLeak, bool) {
	if !strings.Contains(leaf, "?") {
		return credentialLeak{}, false
	}
	var keys []string
	for _, field := range strings.FieldsFunc(leaf, func(r rune) bool { return r == '?' || r == '&' || r == ';' }) {
		key, _, ok := strings.Cut(field, "=")
		if !ok {
			continue
		}
		if unescaped, err := url.QueryUnescape(key); err == nil {
			key = unescaped
		}
		keys = append(keys, key)
	}
	var l credentialLeak
	for _, name := range authParams {
		if slices.Contains(keys, name) {
			l.names = append(l.names, name)
		}
	}
	return l, len(l.names) > 0
}

func walkJSON(dec *jsontext.Decoder, path string, leaf func(path, s string)) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	switch tok.Kind() {
	case '{':
		for dec.PeekKind() != '}' {
			name, err := dec.ReadToken()
			if err != nil {
				return err
			}
			// A token is void after the next read, so the name is copied
			// out before the value is read.
			if err := walkJSON(dec, path+"/"+name.String(), leaf); err != nil {
				return err
			}
		}
		_, err := dec.ReadToken()
		return err
	case '[':
		for i := 0; dec.PeekKind() != ']'; i++ {
			if err := walkJSON(dec, path+"/"+strconv.Itoa(i), leaf); err != nil {
				return err
			}
		}
		_, err := dec.ReadToken()
		return err
	case '"':
		leaf(strings.TrimPrefix(path, "/"), tok.String())
	}
	return nil
}

func walkXML(body []byte, leaf func(path, s string)) error {
	dec := xml.NewDecoder(bytes.NewReader(body))
	var stack []string
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			stack = append(stack, t.Name.Local)
			for _, a := range t.Attr {
				leaf(strings.Join(slices.Concat(stack, []string{a.Name.Local}), "/"), a.Value)
			}
		case xml.CharData:
			leaf(strings.Join(stack, "/"), string(t))
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		}
	}
}
