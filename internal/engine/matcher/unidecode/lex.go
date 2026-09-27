package unidecode

import (
	"fmt"
)

// tokenKind names one lexical token in a Unidecode data module.
type tokenKind int

const (
	tokEOF tokenKind = iota
	tokName
	tokNone
	tokEquals
	tokLParen
	tokRParen
	tokComma
	tokString
)

// token is one lexed token, with its byte offset for error reporting.
type token struct {
	kind tokenKind
	text string // identifier name (tokName) or decoded value (tokString)
	line int
	col  int
}

// lexer scans one Unidecode data module's source. It accepts exactly
// the grammar every shipped module uses: a single `data = ( ... )`
// tuple of string and None literals, one comma per entry including the
// last, and `#` comments outside strings. Anything else is a defect in
// a pinned, checksum-verified wheel, so the lexer refuses it rather
// than skip or guess.
type lexer struct {
	file string
	data []byte
	pos  int
	line int
	col  int
}

func newLexer(file string, data []byte) *lexer {
	return &lexer{file: file, data: data, line: 1, col: 1}
}

func (l *lexer) errf(format string, args ...any) error {
	return fmt.Errorf("%s:%d:%d: "+format, append([]any{l.file, l.line, l.col}, args...)...)
}

func (l *lexer) atEOF() bool { return l.pos >= len(l.data) }

func (l *lexer) peek() byte { return l.data[l.pos] }

// advance consumes and returns one byte, tracking line and column. The
// grammar is ASCII-only with no carriage return, so any other byte is
// itself a refusal.
func (l *lexer) advance() (byte, error) {
	if l.atEOF() {
		return 0, l.errf("unexpected end of file")
	}
	b := l.data[l.pos]
	if b == '\r' {
		return 0, l.errf("unexpected carriage return")
	}
	if b >= 0x80 {
		return 0, l.errf("non-ASCII byte 0x%02x", b)
	}
	l.pos++
	if b == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return b, nil
}

// skipTrivia consumes whitespace and comments between tokens. A `#`
// only starts a comment here, between tokens: lexString never calls
// skipTrivia, so a `#` inside a string is always a character in the
// string, never a comment.
func (l *lexer) skipTrivia() error {
	for !l.atEOF() {
		switch b := l.peek(); b {
		case ' ', '\t', '\n':
			if _, err := l.advance(); err != nil {
				return err
			}
		case '#':
			for !l.atEOF() && l.peek() != '\n' {
				if _, err := l.advance(); err != nil {
					return err
				}
			}
		default:
			return nil
		}
	}
	return nil
}

// punctuationKind maps one of "=(),"'s bytes to its token kind.
func punctuationKind(b byte) tokenKind {
	switch b {
	case '=':
		return tokEquals
	case '(':
		return tokLParen
	case ')':
		return tokRParen
	default:
		return tokComma
	}
}

func isIdentByte(b byte) bool {
	return b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func (l *lexer) next() (token, error) {
	if err := l.skipTrivia(); err != nil {
		return token{}, err
	}
	if l.atEOF() {
		return token{kind: tokEOF, line: l.line, col: l.col}, nil
	}

	line, col := l.line, l.col
	switch b := l.peek(); b {
	case '=', '(', ')', ',':
		if _, err := l.advance(); err != nil {
			return token{}, err
		}
		return token{kind: punctuationKind(b), line: line, col: col}, nil
	case '\'', '"':
		return l.lexString(line, col)
	default:
		if isIdentByte(b) {
			return l.lexIdent(line, col)
		}
		return token{}, l.errf("unexpected byte %q", b)
	}
}

func (l *lexer) lexIdent(line, col int) (token, error) {
	start := l.pos
	for !l.atEOF() && isIdentByte(l.peek()) {
		if _, err := l.advance(); err != nil {
			return token{}, err
		}
	}
	name := string(l.data[start:l.pos])
	if name == "None" {
		return token{kind: tokNone, line: line, col: col}, nil
	}
	return token{kind: tokName, text: name, line: line, col: col}, nil
}

// lexString reads a single- or double-quoted string literal, matching
// the opening quote at close. The grammar carries exactly three
// escapes: \', \\, and \xHH; anything else refuses.
func (l *lexer) lexString(line, col int) (token, error) {
	quote, err := l.advance()
	if err != nil {
		return token{}, err
	}
	var value []byte
	for {
		if l.atEOF() || l.peek() == '\n' {
			return token{}, fmt.Errorf("%s:%d:%d: unterminated string", l.file, line, col)
		}
		b, err := l.advance()
		if err != nil {
			return token{}, err
		}
		if b == quote {
			break
		}
		if b != '\\' {
			value = append(value, b)
			continue
		}
		if l.atEOF() {
			return token{}, fmt.Errorf("%s:%d:%d: unterminated string", l.file, line, col)
		}
		e, err := l.advance()
		if err != nil {
			return token{}, err
		}
		switch e {
		case '\'':
			value = append(value, '\'')
		case '\\':
			value = append(value, '\\')
		case 'x':
			hi, err := l.hexDigit()
			if err != nil {
				return token{}, err
			}
			lo, err := l.hexDigit()
			if err != nil {
				return token{}, err
			}
			value = append(value, hi<<4|lo)
		default:
			return token{}, l.errf("unknown escape \\%c", e)
		}
	}
	return token{kind: tokString, text: string(value), line: line, col: col}, nil
}

func (l *lexer) hexDigit() (byte, error) {
	b, err := l.advance()
	if err != nil {
		return 0, err
	}
	switch {
	case b >= '0' && b <= '9':
		return b - '0', nil
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10, nil
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10, nil
	default:
		return 0, l.errf("invalid hex digit %q", b)
	}
}

// lexModule parses one Unidecode data module's source into its ordered
// entries, nil marking a None literal. file names the source in error
// messages.
func lexModule(file string, data []byte) ([]*string, error) {
	l := newLexer(file, data)

	name, err := l.next()
	if err != nil {
		return nil, err
	}
	if name.kind != tokName || name.text != "data" {
		return nil, fmt.Errorf("%s:%d:%d: expected identifier \"data\"", file, name.line, name.col)
	}
	if err := l.expect(tokEquals, "="); err != nil {
		return nil, err
	}
	if err := l.expect(tokLParen, "("); err != nil {
		return nil, err
	}

	var entries []*string
	for {
		tok, err := l.next()
		if err != nil {
			return nil, err
		}
		if tok.kind == tokRParen {
			break
		}
		var entry *string
		switch tok.kind {
		case tokString:
			s := tok.text
			entry = &s
		case tokNone:
			entry = nil
		default:
			return nil, fmt.Errorf("%s:%d:%d: expected a string or None", file, tok.line, tok.col)
		}
		entries = append(entries, entry)
		if err := l.expect(tokComma, ","); err != nil {
			return nil, err
		}
	}

	tok, err := l.next()
	if err != nil {
		return nil, err
	}
	if tok.kind != tokEOF {
		return nil, fmt.Errorf("%s:%d:%d: unexpected content after closing paren", file, tok.line, tok.col)
	}
	return entries, nil
}

func (l *lexer) expect(kind tokenKind, want string) error {
	tok, err := l.next()
	if err != nil {
		return err
	}
	if tok.kind != kind {
		return fmt.Errorf("%s:%d:%d: expected %q", l.file, tok.line, tok.col, want)
	}
	return nil
}
