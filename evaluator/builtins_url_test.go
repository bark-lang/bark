package evaluator

import (
	"testing"

	"gitlab.com/bark-lang/bark/object"
)

func TestURLEncodeBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic encoding
		{`url.encode("hello world")`, "hello+world"},
		{`url.encode("hello")`, "hello"},

		// Special characters
		{`url.encode("hello@example.com")`, "hello%40example.com"},
		{`url.encode("key=value&other=123")`, "key%3Dvalue%26other%3D123"},
		{`url.encode("a/b/c")`, "a%2Fb%2Fc"},

		// Empty string
		{`url.encode("")`, ""},

		// Unicode
		{`url.encode("こんにちは")`, "%E3%81%93%E3%82%93%E3%81%AB%E3%81%A1%E3%81%AF"},

		// Space encoding (uses + for query params)
		{`url.encode("one two three")`, "one+two+three"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestURLEncodeErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`url.encode()`, "url.encode requires 1 argument (value), got=0"},
		{`url.encode("a", "b")`, "url.encode requires 1 argument (value), got=2"},

		// Wrong argument type
		{`url.encode(123)`, "url.encode requires string argument, got=INTEGER"},
		{`url.encode(true)`, "url.encode requires string argument, got=BOOLEAN"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestURLDecodeBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic decoding - use capture() to extract result from tuple
		{`url.decode("hello+world") > capture(e, r) > r`, "hello world"},
		{`url.decode("hello") > capture(e, r) > r`, "hello"},

		// Special characters
		{`url.decode("hello%40example.com") > capture(e, r) > r`, "hello@example.com"},
		{`url.decode("key%3Dvalue%26other%3D123") > capture(e, r) > r`, "key=value&other=123"},
		{`url.decode("a%2Fb%2Fc") > capture(e, r) > r`, "a/b/c"},

		// Empty string
		{`url.decode("") > capture(e, r) > r`, ""},

		// Unicode
		{`url.decode("%E3%81%93%E3%82%93%E3%81%AB%E3%81%A1%E3%81%AF") > capture(e, r) > r`, "こんにちは"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestURLDecodeSuccess(t *testing.T) {
	// Test that successful decode returns empty map as error
	// After capture(e, r), e is bound to error and r to result
	input := `url.decode("hello+world") > capture(e, r)
e`
	evaluated := testEval(input)

	mapObj, ok := evaluated.(*object.Map)
	if !ok {
		t.Errorf("expected Map object (empty error), got=%T (%+v)", evaluated, evaluated)
		return
	}

	if len(mapObj.Pairs) != 0 {
		t.Errorf("expected empty map (no error), got map with %d pairs", len(mapObj.Pairs))
	}
}

func TestURLDecodeErrors(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
		desc        string
	}{
		// Invalid percent encoding
		{`url.decode("%")`, true, "incomplete percent encoding"},
		{`url.decode("%GG")`, true, "invalid hex"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		// Should return tuple
		tpl, ok := evaluated.(*object.Tuple)
		if !ok {
			t.Errorf("%s: expected Tuple (error tuple), got=%T (%+v)", tt.desc, evaluated, evaluated)
			continue
		}

		if len(tpl.Elements) != 2 {
			t.Errorf("%s: expected tuple with 2 elements, got=%d", tt.desc, len(tpl.Elements))
			continue
		}

		// Check if first element is Error when expected
		if tt.shouldError {
			_, isErr := tpl.Elements[0].(*object.Error)
			if !isErr {
				t.Errorf("%s: expected Error at index 0, got=%T", tt.desc, tpl.Elements[0])
			}
		}
	}
}

func TestURLDecodeWrongArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`url.decode()`, "url.decode requires 1 argument (encoded), got=0"},
		{`url.decode("a", "b")`, "url.decode requires 1 argument (encoded), got=2"},

		// Wrong argument type
		{`url.decode(123)`, "url.decode requires string argument, got=INTEGER"},
		{`url.decode(true)`, "url.decode requires string argument, got=BOOLEAN"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestURLRoundTrip(t *testing.T) {
	tests := []string{
		"hello world",
		"key=value&other=data",
		"path/to/file",
		"こんにちは",
		"special!@#$%chars",
	}

	for _, testStr := range tests {
		// Encode then decode should return original - use capture() to extract result
		input := `"` + testStr + `" > url.encode() > url.decode() > capture(e, r)
r`
		evaluated := testEval(input)
		testStringObject(t, evaluated, testStr)
	}
}

func TestURLParseBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		key      string
		expected string
	}{
		// Parse complete URL - use capture() to extract result then get field
		{`url.parse("https://example.com/path?key=value#section") > capture(e, r)
r > get("scheme")`, "scheme", "https"},
		{`url.parse("https://example.com/path?key=value#section") > capture(e, r)
r > get("host")`, "host", "example.com"},
		{`url.parse("https://example.com/path?key=value#section") > capture(e, r)
r > get("path")`, "path", "/path"},
		{`url.parse("https://example.com/path?key=value#section") > capture(e, r)
r > get("query")`, "query", "key=value"},
		{`url.parse("https://example.com/path?key=value#section") > capture(e, r)
r > get("fragment")`, "fragment", "section"},

		// Parse URL without query/fragment
		{`url.parse("http://localhost:8080/api") > capture(e, r)
r > get("scheme")`, "scheme", "http"},
		{`url.parse("http://localhost:8080/api") > capture(e, r)
r > get("host")`, "host", "localhost:8080"},
		{`url.parse("http://localhost:8080/api") > capture(e, r)
r > get("path")`, "path", "/api"},

		// Parse relative URL
		{`url.parse("/path/to/resource") > capture(e, r)
r > get("path")`, "path", "/path/to/resource"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestURLParseSuccess(t *testing.T) {
	// Test that successful parse returns empty map as error
	input := `url.parse("https://example.com") > capture(e, r)
e`
	evaluated := testEval(input)

	mapObj, ok := evaluated.(*object.Map)
	if !ok {
		t.Errorf("expected Map object (empty error), got=%T (%+v)", evaluated, evaluated)
		return
	}

	if len(mapObj.Pairs) != 0 {
		t.Errorf("expected empty map (no error), got map with %d pairs", len(mapObj.Pairs))
	}
}

func TestURLParseWrongArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`url.parse()`, "url.parse requires 1 argument (url), got=0"},
		{`url.parse("a", "b")`, "url.parse requires 1 argument (url), got=2"},

		// Wrong argument type
		{`url.parse(123)`, "url.parse requires string argument, got=INTEGER"},
		{`url.parse(true)`, "url.parse requires string argument, got=BOOLEAN"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}
