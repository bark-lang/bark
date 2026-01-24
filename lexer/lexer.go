package lexer

import (
	"unicode"
	"unicode/utf8"

	"gitlab.com/bark-lang/bark/token"
)

// Lexer tokenizes Bark source code
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           rune // current char under examination
	line         int  // current line number
	column       int  // current column number
}

// New creates a new Lexer for the given input
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// readChar advances the lexer to the next character
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // End of file
		l.position = len(l.input)
	} else {
		var size int
		l.ch, size = utf8.DecodeRuneInString(l.input[l.readPosition:])
		l.position = l.readPosition
		l.readPosition += size
	}
	l.column++
}

// peekChar returns the next character without advancing the lexer
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '\n':
		tok = newToken(token.NEWLINE, l.ch, l.line, l.column)
		l.line++
		l.column = 0
	case ',':
		tok = newToken(token.COMMA, l.ch, l.line, l.column)
	case ':':
		tok = newToken(token.COLON, l.ch, l.line, l.column)
	case '>':
		tok = newToken(token.GT, l.ch, l.line, l.column)
	case '-':
		tok = newToken(token.MINUS, l.ch, l.line, l.column)
	case '[':
		tok = newToken(token.LBRACK, l.ch, l.line, l.column)
	case ']':
		tok = newToken(token.RBRACK, l.ch, l.line, l.column)
	case '(':
		tok = newToken(token.LPAREN, l.ch, l.line, l.column)
	case ')':
		tok = newToken(token.RPAREN, l.ch, l.line, l.column)
	case '{':
		tok = newToken(token.LBRACE, l.ch, l.line, l.column)
	case '}':
		tok = newToken(token.RBRACE, l.ch, l.line, l.column)
	case '|':
		tok = newToken(token.PIPE, l.ch, l.line, l.column)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readInterpretedString()
	case '`':
		tok.Type = token.STRING
		tok.Literal = l.readRawString()
	case '/':
		if l.peekChar() == '/' {
			tok.Type = token.COMMENT
			tok.Literal = l.readComment()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch, l.line, l.column)
		}
	case '.':
		// Check if this is a float starting with decimal point (e.g., .5)
		if isDigit(l.peekChar()) {
			return l.readFloatStartingWithDot()
		}
		tok = newToken(token.DOT, l.ch, l.line, l.column)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
		tok.Line = l.line
		tok.Column = l.column
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			return l.readNumber()
		} else {
			tok = newToken(token.ILLEGAL, l.ch, l.line, l.column)
		}
	}

	l.readChar()
	return tok
}

// newToken creates a new token with a single character literal
func newToken(tokenType token.TokenType, ch rune, line, column int) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(ch),
		Line:    line,
		Column:  column,
	}
}

// readIdentifier reads an identifier or keyword
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads an integer or float literal
func (l *Lexer) readNumber() token.Token {
	tok := token.Token{Line: l.line, Column: l.column}
	position := l.position
	isFloat := false

	// Read integer part
	for isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	// Check for decimal point
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar() // consume '.'

		// Read fractional part
		for isDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
	}

	// Check for exponent (e or E)
	if l.ch == 'e' || l.ch == 'E' {
		isFloat = true
		l.readChar() // consume 'e' or 'E'

		// Handle optional sign
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}

		// Read exponent digits
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	tok.Literal = l.input[position:l.position]
	if isFloat {
		tok.Type = token.FLOAT
	} else {
		tok.Type = token.INT
	}

	return tok
}

// readFloatStartingWithDot reads a float literal starting with decimal point (e.g., .5)
func (l *Lexer) readFloatStartingWithDot() token.Token {
	tok := token.Token{Line: l.line, Column: l.column, Type: token.FLOAT}
	position := l.position

	l.readChar() // consume '.'

	// Read fractional part
	for isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	// Check for exponent
	if l.ch == 'e' || l.ch == 'E' {
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	tok.Literal = l.input[position:l.position]
	return tok
}

// readInterpretedString reads a double-quoted string with escape sequences
func (l *Lexer) readInterpretedString() string {
	result := ""

	l.readChar() // consume opening "

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'a':
				result += "\a"
			case 'b':
				result += "\b"
			case 'f':
				result += "\f"
			case 'n':
				result += "\n"
			case 'r':
				result += "\r"
			case 't':
				result += "\t"
			case 'v':
				result += "\v"
			case '\\':
				result += "\\"
			case '\'':
				result += "'"
			case '"':
				result += "\""
			case '{':
				result += "\\{"
			case '}':
				result += "\\}"
			default:
				// Invalid escape sequence - just include the backslash
				result += "\\" + string(l.ch)
			}
			l.readChar()
		} else {
			result += string(l.ch)
			l.readChar()
		}
	}

	// l.ch is now on closing " or 0
	return result
}

// readRawString reads a backtick-quoted raw string
func (l *Lexer) readRawString() string {
	l.readChar() // consume opening `

	position := l.position // start after opening `

	for l.ch != '`' && l.ch != 0 {
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}

	// l.ch is now on closing ` or 0
	// l.position points to the closing `
	return l.input[position:l.position]
}

// readComment reads a single-line comment starting with //
func (l *Lexer) readComment() string {
	position := l.position

	// Skip the two slashes
	l.readChar()
	l.readChar()

	// Read until end of line or end of file
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}

	return l.input[position:l.position]
}

// skipWhitespace skips over whitespace characters (except newlines)
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

// isLetter returns true if the character is a letter or underscore or question mark
func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '?'
}

// isDigit returns true if the character is a decimal digit
func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}
