package evaluator

import (
	"testing"

	"gitlab.com/bark-lang/barki/object"
)

func TestBase64EncodeBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic encoding
		{`base64.encode("hello")`, "aGVsbG8="},
		{`base64.encode("hello world")`, "aGVsbG8gd29ybGQ="},

		// Empty string
		{`base64.encode("")`, ""},

		// Special characters
		{`base64.encode("Hello, World!")`, "SGVsbG8sIFdvcmxkIQ=="},
		{`base64.encode("123456")`, "MTIzNDU2"},

		// Unicode characters
		{`base64.encode("こんにちは")`, "44GT44KT44Gr44Gh44Gv"},
		{`base64.encode("🚀")`, "8J+agA=="},

		// Long text
		{`base64.encode("The quick brown fox jumps over the lazy dog")`, "VGhlIHF1aWNrIGJyb3duIGZveCBqdW1wcyBvdmVyIHRoZSBsYXp5IGRvZw=="},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestBase64EncodeErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`base64.encode()`, "base64.encode requires 1 argument (data), got=0"},
		{`base64.encode("a", "b")`, "base64.encode requires 1 argument (data), got=2"},

		// Wrong argument type
		{`base64.encode(123)`, "base64.encode requires string argument, got=INTEGER"},
		{`base64.encode(true)`, "base64.encode requires string argument, got=BOOLEAN"},
		{`base64.encode([1, 2, 3])`, "base64.encode requires string argument, got=ARRAY"},
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

func TestBase64DecodeBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic decoding - use capture() to extract result from tuple
		{`base64.decode("aGVsbG8=") > capture(e, r) > r`, "hello"},
		{`base64.decode("aGVsbG8gd29ybGQ=") > capture(e, r) > r`, "hello world"},

		// Empty string
		{`base64.decode("") > capture(e, r) > r`, ""},

		// Special characters
		{`base64.decode("SGVsbG8sIFdvcmxkIQ==") > capture(e, r) > r`, "Hello, World!"},
		{`base64.decode("MTIzNDU2") > capture(e, r) > r`, "123456"},

		// Unicode characters
		{`base64.decode("44GT44KT44Gr44Gh44Gv") > capture(e, r) > r`, "こんにちは"},
		{`base64.decode("8J+agA==") > capture(e, r) > r`, "🚀"},

		// Long text
		{`base64.decode("VGhlIHF1aWNrIGJyb3duIGZveCBqdW1wcyBvdmVyIHRoZSBsYXp5IGRvZw==") > capture(e, r) > r`, "The quick brown fox jumps over the lazy dog"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestBase64DecodeSuccess(t *testing.T) {
	// Test that successful decode returns empty map as error
	// After capture(e, r), e is bound to error and r to result
	input := `base64.decode("aGVsbG8=") > capture(e, r)
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

func TestBase64DecodeErrors(t *testing.T) {
	tests := []struct {
		input         string
		shouldHaveErr bool
	}{
		// Invalid base64
		{`base64.decode("not-valid-base64!!!")`, true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		// Should return tuple with error at index 0
		tpl, ok := evaluated.(*object.Tuple)
		if !ok {
			t.Errorf("expected Tuple (error tuple), got=%T (%+v)", evaluated, evaluated)
			continue
		}

		if len(tpl.Elements) != 2 {
			t.Errorf("expected tuple with 2 elements, got=%d", len(tpl.Elements))
			continue
		}

		// First element should be Error object
		errObj, ok := tpl.Elements[0].(*object.Error)
		if !ok && tt.shouldHaveErr {
			t.Errorf("expected Error object at index 0, got=%T (%+v)", tpl.Elements[0], tpl.Elements[0])
			continue
		}

		if tt.shouldHaveErr && errObj == nil {
			t.Errorf("expected error but got none")
		}
	}
}

func TestBase64DecodeWrongArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`base64.decode()`, "base64.decode requires 1 argument (encoded), got=0"},
		{`base64.decode("a", "b")`, "base64.decode requires 1 argument (encoded), got=2"},

		// Wrong argument type
		{`base64.decode(123)`, "base64.decode requires string argument, got=INTEGER"},
		{`base64.decode(true)`, "base64.decode requires string argument, got=BOOLEAN"},
		{`base64.decode([1, 2, 3])`, "base64.decode requires string argument, got=ARRAY"},
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

func TestBase64RoundTrip(t *testing.T) {
	tests := []string{
		"hello world",
		"The quick brown fox jumps over the lazy dog",
		"こんにちは",
		"🚀🌟✨",
		"",
		"123456",
		"Special chars: !@#$%^&*()",
	}

	for _, testStr := range tests {
		// Encode then decode should return original - use capture() to extract result
		input := `"` + testStr + `" > base64.encode() > base64.decode() > capture(e, r) > r`
		evaluated := testEval(input)
		testStringObject(t, evaluated, testStr)
	}
}
