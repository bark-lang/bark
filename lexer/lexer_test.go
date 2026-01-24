package lexer

import (
	"testing"

	"gitlab.com/bark-lang/bark/token"
)

func TestNextToken_Operators(t *testing.T) {
	input := `. , > [ ] ( ) { }`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.DOT, "."},
		{token.COMMA, ","},
		{token.GT, ">"},
		{token.LBRACK, "["},
		{token.RBRACK, "]"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Keywords(t *testing.T) {
	input := `fn error include module pub type true false`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.FN, "fn"},
		{token.ERROR, "error"},
		{token.INCLUDE, "include"},
		{token.MODULE, "module"},
		{token.PUB, "pub"},
		{token.TYPE, "type"},
		{token.TRUE, "true"},
		{token.FALSE, "false"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Identifiers(t *testing.T) {
	input := `foo bar_baz has_key? eq? _private`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "foo"},
		{token.IDENT, "bar_baz"},
		{token.IDENT, "has_key?"},
		{token.IDENT, "eq?"},
		{token.IDENT, "_private"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Integers(t *testing.T) {
	input := `0 123 1_000 999_999_999`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INT, "0"},
		{token.INT, "123"},
		{token.INT, "1_000"},
		{token.INT, "999_999_999"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Floats(t *testing.T) {
	input := `3.14 .5 0.0 1.0e10 1.5e-3 2E+5`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.FLOAT, "3.14"},
		{token.FLOAT, ".5"},
		{token.FLOAT, "0.0"},
		{token.FLOAT, "1.0e10"},
		{token.FLOAT, "1.5e-3"},
		{token.FLOAT, "2E+5"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Strings(t *testing.T) {
	input := `"hello" "world\n" "tab\there" ` + "`raw\nstring`"

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.STRING, "hello"},
		{token.STRING, "world\n"},
		{token.STRING, "tab\there"},
		{token.STRING, "raw\nstring"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Comments(t *testing.T) {
	input := `// This is a comment
foo // inline comment
// another comment`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.COMMENT, "// This is a comment"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "foo"},
		{token.COMMENT, "// inline comment"},
		{token.NEWLINE, "\n"},
		{token.COMMENT, "// another comment"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_BasicFunction(t *testing.T) {
	input := `fn hello_world(){
  stderr("hello world")
}`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.FN, "fn"},
		{token.IDENT, "hello_world"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "stderr"},
		{token.LPAREN, "("},
		{token.STRING, "hello world"},
		{token.RPAREN, ")"},
		{token.NEWLINE, "\n"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_FunctionWithReturn(t *testing.T) {
	input := `fn name(){
  return("bark")
}(string)`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.FN, "fn"},
		{token.IDENT, "name"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "return"},
		{token.LPAREN, "("},
		{token.STRING, "bark"},
		{token.RPAREN, ")"},
		{token.NEWLINE, "\n"},
		{token.RBRACE, "}"},
		{token.LPAREN, "("},
		{token.IDENT, "string"},
		{token.RPAREN, ")"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_LinkOperator(t *testing.T) {
	input := `num > add(5) > result`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "num"},
		{token.GT, ">"},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.RPAREN, ")"},
		{token.GT, ">"},
		{token.IDENT, "result"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Array(t *testing.T) {
	input := `[1, 2, 3]`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LBRACK, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.COMMA, ","},
		{token.INT, "3"},
		{token.RBRACK, "]"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Map(t *testing.T) {
	input := `{"name": "bark", "version": "0.0.1"}`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LBRACE, "{"},
		{token.STRING, "name"},
		{token.COLON, ":"},
		{token.STRING, "bark"},
		{token.COMMA, ","},
		{token.STRING, "version"},
		{token.COLON, ":"},
		{token.STRING, "0.0.1"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_LineNumbers(t *testing.T) {
	input := `fn test() {
  return(42)
}`

	// Updated to include NEWLINE tokens
	expectedLines := []int{1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3}
	//                     fn test ( ) { NL ret ( 42 ) NL }

	l := New(input)

	for i, expectedLine := range expectedLines {
		tok := l.NextToken()

		if tok.Line != expectedLine {
			t.Fatalf("tests[%d] - line wrong. expected=%d, got=%d (token: %s)",
				i, expectedLine, tok.Line, tok.Literal)
		}
	}
}

func TestNextToken_AnonymousFunction(t *testing.T) {
	input := `(n int) { return(n) }(int)`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LPAREN, "("},
		{token.IDENT, "n"},
		{token.IDENT, "int"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "return"},
		{token.LPAREN, "("},
		{token.IDENT, "n"},
		{token.RPAREN, ")"},
		{token.RBRACE, "}"},
		{token.LPAREN, "("},
		{token.IDENT, "int"},
		{token.RPAREN, ")"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_EscapeSequences(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello\nworld"`, "hello\nworld"},
		{`"tab\there"`, "tab\there"},
		{`"quote\"test"`, "quote\"test"},
		{`"backslash\\"`, "backslash\\"},
		{`"bell\a"`, "bell\a"},
		{`"backspace\b"`, "backspace\b"},
		{`"formfeed\f"`, "formfeed\f"},
		{`"carriage\r"`, "carriage\r"},
		{`"vertical\v"`, "vertical\v"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != token.STRING {
			t.Fatalf("input %q - tokentype wrong. expected=STRING, got=%q",
				tt.input, tok.Type)
		}

		if tok.Literal != tt.expected {
			t.Fatalf("input %q - literal wrong. expected=%q, got=%q",
				tt.input, tt.expected, tok.Literal)
		}
	}
}

func TestIsBuiltin(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"print", true},
		{"println", true},
		{"eprint", true},
		{"eprintln", true},
		{"return", true},
		{"return?", true},
		{"eq?", true},
		{"gt?", true},
		{"foo", false},
		{"bar_baz", false},
		{"my_func", false},
	}

	for _, tt := range tests {
		result := token.IsBuiltin(tt.name)
		if result != tt.expected {
			t.Fatalf("IsBuiltin(%q) - expected=%t, got=%t",
				tt.name, tt.expected, result)
		}
	}
}

func TestNextToken_Unicode(t *testing.T) {
	// Test Unicode identifiers
	input := `函数 función переменная`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "函数"},
		{token.IDENT, "función"},
		{token.IDENT, "переменная"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_UnicodeStrings(t *testing.T) {
	// Test Unicode in strings
	input := `"Hello 世界" "مرحبا" "Здравствуй"`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.STRING, "Hello 世界"},
		{token.STRING, "مرحبا"},
		{token.STRING, "Здравствуй"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_UnicodeComments(t *testing.T) {
	// Test Unicode in comments
	input := `// This is a comment with Unicode: 你好 مرحبا
foo`

	tests := []struct {
		expectedType token.TokenType
	}{
		{token.COMMENT},
		{token.NEWLINE},
		{token.IDENT},
		{token.EOF},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
	}
}
