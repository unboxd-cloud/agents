package adl

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// tokenKind enumerates the lexical categories the parser cares about. Keywords
// are not distinct kinds: they are IDENT tokens matched by literal text at parse
// time, which mirrors Langium's behaviour where a keyword wins in the grammar
// position and keeps the lexer free of a reserved-word table.
type tokenKind int

const (
	tEOF tokenKind = iota
	tIdent
	tString
	tNumber
	tLBrace // {
	tRBrace // }
	tLParen // (
	tRParen // )
	tColon  // :
	tComma  // ,
	tDot    // .
	tArrow  // ->
)

func (k tokenKind) String() string {
	switch k {
	case tEOF:
		return "end of input"
	case tIdent:
		return "identifier"
	case tString:
		return "string"
	case tNumber:
		return "number"
	case tLBrace:
		return "'{'"
	case tRBrace:
		return "'}'"
	case tLParen:
		return "'('"
	case tRParen:
		return "')'"
	case tColon:
		return "':'"
	case tComma:
		return "','"
	case tDot:
		return "'.'"
	case tArrow:
		return "'->'"
	default:
		return "token"
	}
}

type token struct {
	Kind  tokenKind
	Value string // IDENT text, unescaped STRING contents, or NUMBER text
	Pos   Position
}

func (t token) describe() string {
	switch t.Kind {
	case tIdent:
		return fmt.Sprintf("'%s'", t.Value)
	case tString:
		return "string literal"
	case tNumber:
		return fmt.Sprintf("number '%s'", t.Value)
	default:
		return t.Kind.String()
	}
}

// lexer turns ADL source into a token slice, honouring the grammar's hidden
// terminals (whitespace, // line comments and /* */ block comments).
type lexer struct {
	src   string
	pos   int
	line  int
	col   int
	diags []Diagnostic
}

func newLexer(src string) *lexer {
	return &lexer{src: src, line: 1, col: 1}
}

func (l *lexer) here() Position {
	return Position{Line: l.line, Column: l.col, Offset: l.pos}
}

func (l *lexer) advance() rune {
	r, size := utf8.DecodeRuneInString(l.src[l.pos:])
	l.pos += size
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}

func (l *lexer) peekByte() byte {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *lexer) peekByteAt(off int) byte {
	if l.pos+off >= len(l.src) {
		return 0
	}
	return l.src[l.pos+off]
}

// tokenize scans the whole input. Lexical problems (e.g. an unterminated string)
// are recorded as diagnostics rather than aborting, so the editor still gets a
// best-effort token stream.
func (l *lexer) tokenize() ([]token, []Diagnostic) {
	var toks []token
	for {
		l.skipTrivia()
		if l.pos >= len(l.src) {
			toks = append(toks, token{Kind: tEOF, Pos: l.here()})
			return toks, l.diags
		}
		start := l.here()
		c := l.peekByte()
		switch {
		case c == '{':
			l.advance()
			toks = append(toks, token{Kind: tLBrace, Pos: start})
		case c == '}':
			l.advance()
			toks = append(toks, token{Kind: tRBrace, Pos: start})
		case c == '(':
			l.advance()
			toks = append(toks, token{Kind: tLParen, Pos: start})
		case c == ')':
			l.advance()
			toks = append(toks, token{Kind: tRParen, Pos: start})
		case c == ':':
			l.advance()
			toks = append(toks, token{Kind: tColon, Pos: start})
		case c == ',':
			l.advance()
			toks = append(toks, token{Kind: tComma, Pos: start})
		case c == '.':
			l.advance()
			toks = append(toks, token{Kind: tDot, Pos: start})
		case c == '-' && l.peekByteAt(1) == '>':
			l.advance()
			l.advance()
			toks = append(toks, token{Kind: tArrow, Pos: start})
		case c == '"':
			toks = append(toks, l.lexString(start))
		case isDigit(c):
			toks = append(toks, l.lexNumber(start))
		case isIdentStart(c):
			toks = append(toks, l.lexIdent(start))
		default:
			l.diags = append(l.diags, Diagnostic{
				Severity: SeverityError,
				Message:  fmt.Sprintf("unexpected character %q", string(rune(c))),
				Pos:      start,
			})
			l.advance()
		}
	}
}

// skipTrivia consumes whitespace and comments (the hidden terminals).
func (l *lexer) skipTrivia() {
	for l.pos < len(l.src) {
		c := l.peekByte()
		switch {
		case c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\f' || c == '\v':
			l.advance()
		case c == '/' && l.peekByteAt(1) == '/':
			for l.pos < len(l.src) && l.peekByte() != '\n' {
				l.advance()
			}
		case c == '/' && l.peekByteAt(1) == '*':
			startBlock := l.here()
			l.advance()
			l.advance()
			closed := false
			for l.pos < len(l.src) {
				if l.peekByte() == '*' && l.peekByteAt(1) == '/' {
					l.advance()
					l.advance()
					closed = true
					break
				}
				l.advance()
			}
			if !closed {
				l.diags = append(l.diags, Diagnostic{
					Severity: SeverityError,
					Message:  "unterminated block comment",
					Pos:      startBlock,
				})
			}
		default:
			return
		}
	}
}

// lexString matches the STRING terminal: /"[^"\\]*(\\.[^"\\]*)*"/.
func (l *lexer) lexString(start Position) token {
	l.advance() // opening quote
	var b strings.Builder
	for l.pos < len(l.src) {
		c := l.peekByte()
		if c == '"' {
			l.advance()
			return token{Kind: tString, Value: b.String(), Pos: start}
		}
		if c == '\\' {
			l.advance()
			if l.pos >= len(l.src) {
				break
			}
			esc := l.advance()
			switch esc {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			case '/':
				b.WriteByte('/')
			default:
				b.WriteRune(esc)
			}
			continue
		}
		if c == '\n' {
			break // newline before closing quote: unterminated
		}
		b.WriteRune(l.advance())
	}
	l.diags = append(l.diags, Diagnostic{
		Severity: SeverityError,
		Message:  "unterminated string literal",
		Pos:      start,
	})
	return token{Kind: tString, Value: b.String(), Pos: start}
}

// lexNumber matches the NUMBER terminal: /[0-9]+(\.[0-9]+)?/.
func (l *lexer) lexNumber(start Position) token {
	begin := l.pos
	for l.pos < len(l.src) && isDigit(l.peekByte()) {
		l.advance()
	}
	if l.peekByte() == '.' && isDigit(l.peekByteAt(1)) {
		l.advance() // dot
		for l.pos < len(l.src) && isDigit(l.peekByte()) {
			l.advance()
		}
	}
	return token{Kind: tNumber, Value: l.src[begin:l.pos], Pos: start}
}

// lexIdent matches the ID terminal: /[_a-zA-Z][\w_]*/.
func (l *lexer) lexIdent(start Position) token {
	begin := l.pos
	for l.pos < len(l.src) && isIdentPart(l.peekByte()) {
		l.advance()
	}
	return token{Kind: tIdent, Value: l.src[begin:l.pos], Pos: start}
}

func isDigit(c byte) bool      { return c >= '0' && c <= '9' }
func isIdentStart(c byte) bool { return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isIdentPart(c byte) bool  { return isIdentStart(c) || isDigit(c) }
