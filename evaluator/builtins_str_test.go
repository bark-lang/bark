package evaluator

import (
	"testing"

	"gitlab.com/bark-lang/barki/object"
)

func TestStrUpperBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.upper("hello")`, "HELLO"},
		{`str.upper("Hello World")`, "HELLO WORLD"},
		{`str.upper("ALREADY UPPER")`, "ALREADY UPPER"},
		{`str.upper("")`, ""},
		{`str.upper("123abc")`, "123ABC"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestStrUpperErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.upper()`, "str.upper requires 1 argument, got=0"},
		{`str.upper("a", "b")`, "str.upper requires 1 argument, got=2"},
		{`str.upper(123)`, "str.upper requires string argument, got=INTEGER"},
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

func TestStrLowerBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.lower("HELLO")`, "hello"},
		{`str.lower("Hello World")`, "hello world"},
		{`str.lower("already lower")`, "already lower"},
		{`str.lower("")`, ""},
		{`str.lower("123ABC")`, "123abc"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestStrLowerErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.lower()`, "str.lower requires 1 argument, got=0"},
		{`str.lower("a", "b")`, "str.lower requires 1 argument, got=2"},
		{`str.lower(123)`, "str.lower requires string argument, got=INTEGER"},
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

func TestStrTrimBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.trim("  hello  ")`, "hello"},
		{`str.trim("\t\nhello\n\t")`, "hello"},
		{`str.trim("hello")`, "hello"},
		{`str.trim("")`, ""},
		{`str.trim("   ")`, ""},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestStrTrimErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.trim()`, "str.trim requires 1 argument, got=0"},
		{`str.trim("a", "b")`, "str.trim requires 1 argument, got=2"},
		{`str.trim(123)`, "str.trim requires string argument, got=INTEGER"},
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

func TestStrStartsWithBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`str.starts_with?("hello world", "hello")`, true},
		{`str.starts_with?("hello world", "world")`, false},
		{`str.starts_with?("hello", "hello")`, true},
		{`str.starts_with?("hello", "")`, true},
		{`str.starts_with?("", "hello")`, false},
		{`str.starts_with?("", "")`, true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestStrStartsWithErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.starts_with?()`, "str.starts_with? requires 2 arguments (string, prefix), got=0"},
		{`str.starts_with?("a")`, "str.starts_with? requires 2 arguments (string, prefix), got=1"},
		{`str.starts_with?(123, "a")`, "str.starts_with? requires string as first argument, got=INTEGER"},
		{`str.starts_with?("a", 123)`, "str.starts_with? requires string prefix, got=INTEGER"},
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

func TestStrEndsWithBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`str.ends_with?("hello world", "world")`, true},
		{`str.ends_with?("hello world", "hello")`, false},
		{`str.ends_with?("hello", "hello")`, true},
		{`str.ends_with?("hello", "")`, true},
		{`str.ends_with?("", "hello")`, false},
		{`str.ends_with?("", "")`, true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestStrEndsWithErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.ends_with?()`, "str.ends_with? requires 2 arguments (string, suffix), got=0"},
		{`str.ends_with?("a")`, "str.ends_with? requires 2 arguments (string, suffix), got=1"},
		{`str.ends_with?(123, "a")`, "str.ends_with? requires string as first argument, got=INTEGER"},
		{`str.ends_with?("a", 123)`, "str.ends_with? requires string suffix, got=INTEGER"},
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

func TestStrReplaceBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.replace("hello world", "world", "universe")`, "hello universe"},
		{`str.replace("aaa", "a", "b")`, "bbb"},
		{`str.replace("hello", "x", "y")`, "hello"},
		{`str.replace("hello", "hello", "")`, ""},
		{`str.replace("", "a", "b")`, ""},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestStrReplaceErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.replace()`, "str.replace requires 3 arguments (string, old, new), got=0"},
		{`str.replace("a", "b")`, "str.replace requires 3 arguments (string, old, new), got=2"},
		{`str.replace(123, "a", "b")`, "str.replace requires string as first argument, got=INTEGER"},
		{`str.replace("a", 123, "b")`, "str.replace requires string as second argument, got=INTEGER"},
		{`str.replace("a", "b", 123)`, "str.replace requires string as third argument, got=INTEGER"},
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

func TestStrSplitBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{`str.split("a,b,c", ",")`, []string{"a", "b", "c"}},
		{`str.split("hello world", " ")`, []string{"hello", "world"}},
		{`str.split("hello", ",")`, []string{"hello"}},
		{`str.split("", ",")`, []string{""}},
		{`str.split("a::b::c", "::")`, []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Errorf("expected Array object, got=%T (%+v)", evaluated, evaluated)
			continue
		}

		if len(arr.Elements) != len(tt.expected) {
			t.Errorf("wrong array length. expected=%d, got=%d", len(tt.expected), len(arr.Elements))
			continue
		}

		for i, exp := range tt.expected {
			str, ok := arr.Elements[i].(*object.String)
			if !ok {
				t.Errorf("array element %d is not String. got=%T", i, arr.Elements[i])
				continue
			}
			if str.Value != exp {
				t.Errorf("element %d wrong value. expected=%q, got=%q", i, exp, str.Value)
			}
		}
	}
}

func TestStrSplitErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.split()`, "str.split requires 2 arguments (string, delimiter), got=0"},
		{`str.split("a")`, "str.split requires 2 arguments (string, delimiter), got=1"},
		{`str.split(123, ",")`, "str.split requires string as first argument, got=INTEGER"},
		{`str.split("a", 123)`, "str.split requires string delimiter, got=INTEGER"},
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

func TestStrJoinBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.join(["a", "b", "c"], ",")`, "a,b,c"},
		{`str.join(["hello", "world"], " ")`, "hello world"},
		{`str.join(["single"], ",")`, "single"},
		{`str.join([], ",")`, ""},
		{`str.join(["a", "b"], "")`, "ab"},
		// Non-string elements get converted via Inspect
		{`str.join([1, 2, 3], "-")`, "1-2-3"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestStrJoinErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.join()`, "str.join requires 2 arguments (array, separator), got=0"},
		{`str.join(["a"])`, "str.join requires 2 arguments (array, separator), got=1"},
		{`str.join("not array", ",")`, "str.join requires array as first argument, got=STRING"},
		{`str.join(["a"], 123)`, "str.join requires string separator, got=INTEGER"},
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

// Test chaining with link operator
func TestStrChaining(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"  hello  " > str.trim() > str.upper()`, "HELLO"},
		{`"HELLO WORLD" > str.lower() > str.replace("world", "universe")`, "hello universe"},
		{`"a,b,c" > str.split(",") > str.join("-")`, "a-b-c"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestStrConcatBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.concat("hello", " world")`, "hello world"},
		{`str.concat("a", "b", "c")`, "abc"},
		{`str.concat("hello", " ", "world", "!")`, "hello world!"},
		{`str.concat("", "hello")`, "hello"},
		{`str.concat("hello", "")`, "hello"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestStrConcatErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.concat()`, "str.concat requires at least 2 arguments, got=0"},
		{`str.concat("a")`, "str.concat requires at least 2 arguments, got=1"},
		{`str.concat(123, "a")`, "str.concat requires string arguments, argument 1 is INTEGER"},
		{`str.concat("a", 123)`, "str.concat requires string arguments, argument 2 is INTEGER"},
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

func TestStrConcatChaining(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello" > str.concat(" world")`, "hello world"},
		{`"hello" > str.concat(" ", "world")`, "hello world"},
		{`"a" > str.concat("b") > str.concat("c")`, "abc"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestStrFormatBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic single placeholder
		{`str.format("Hello, {0}!", "world")`, "Hello, world!"},
		// Multiple placeholders
		{`str.format("{0} is {1} years old", "Alice", 30)`, "Alice is 30 years old"},
		// Reusing same placeholder
		{`str.format("{0} said {0} likes cats", "Bob")`, "Bob said Bob likes cats"},
		// Reordering arguments
		{`str.format("Age: {1}, Name: {0}", "Alice", 30)`, "Age: 30, Name: Alice"},
		// Multiple uses of same index
		{`str.format("{0}{0}{0}", "a")`, "aaa"},
		// No placeholders (just return format string)
		{`str.format("Hello, world!")`, "Hello, world!"},
		// Empty format string
		{`str.format("")`, ""},
		// Escaped braces
		{`str.format("{{literal}}")`, "{literal}"},
		{`str.format("{{0}}")`, "{0}"},
		{`str.format("Use {{0}} for placeholder")`, "Use {0} for placeholder"},
		{`str.format("Open {{ and close }}")`, "Open { and close }"},
		// Mixed escaped and real placeholders
		{`str.format("{{name}}: {0}", "value")`, "{name}: value"},
		// Different types
		{`str.format("int: {0}, float: {1}, bool: {2}", 42, 3.14, true)`, "int: 42, float: 3.14, bool: true"},
		// Arrays and maps
		{`str.format("arr: {0}", [1, 2, 3])`, "arr: [1, 2, 3]"},
		{`str.format("map: {0}", {"a": 1})`, `map: {a: 1}`},
		// Multi-digit indices
		{`str.format("{0}{1}{2}{3}{4}{5}{6}{7}{8}{9}{10}", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k")`, "abcdefghijk"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestStrFormatErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// No arguments
		{`str.format()`, "str.format requires at least 1 argument (format string), got=0"},
		// Non-string format
		{`str.format(123)`, "str.format requires string as first argument, got=INTEGER"},
		// Empty placeholder
		{`str.format("Hello, {}!")`, "str.format: empty placeholder at position 7: use {0}, {1}, etc"},
		// Non-numeric placeholder
		{`str.format("Hello, {name}!", "world")`, "str.format: invalid placeholder {name}, must be numeric index"},
		// Index out of range
		{`str.format("Hello, {1}!", "world")`, "str.format: placeholder {1} out of range, only 1 arguments provided"},
		{`str.format("{0} and {2}", "a", "b")`, "str.format: placeholder {2} out of range, only 2 arguments provided"},
		// Unclosed placeholder
		{`str.format("Hello, {0")`, "str.format: unclosed placeholder at position 7"},
		{`str.format("{")`, "str.format: unclosed placeholder at position 0"},
		// Unescaped closing brace
		{`str.format("Hello }")`, "str.format: unescaped '}' at position 6, use '}}' for literal"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("for input %q: expected Error object, got=%T (%+v)", tt.input, evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("for input %q: wrong error message.\nexpected=%q\ngot=%q", tt.input, tt.expected, errObj.Msg)
		}
	}
}

func TestStrFormatChaining(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Chaining with link operator
		{`"Hello, {0}!" > str.format("world")`, "Hello, world!"},
		{`"{0} is {1}" > str.format("Alice", 30)`, "Alice is 30"},
		// Chain with other str functions
		{`"hello, {0}!" > str.format("world") > str.upper()`, "HELLO, WORLD!"},
		{`"  {0}  " > str.format("test") > str.trim()`, "test"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestStrNumericBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Valid integers
		{`str.numeric?("123")`, true},
		{`str.numeric?("0")`, true},
		{`str.numeric?("-456")`, true},
		{`str.numeric?("+789")`, true},
		{`str.numeric?("007")`, true},
		{`str.numeric?("-0")`, true},

		// Valid floats
		{`str.numeric?("3.14")`, true},
		{`str.numeric?("-2.718")`, true},
		{`str.numeric?("0.5")`, true},
		{`str.numeric?(".5")`, true},
		{`str.numeric?("123.456")`, true},
		{`str.numeric?("-0.0")`, true},

		// Scientific notation
		{`str.numeric?("1e10")`, true},
		{`str.numeric?("1E10")`, true},
		{`str.numeric?("-1.5e-3")`, true},

		// Whitespace handling (trimmed)
		{`str.numeric?("  42  ")`, true},
		{`str.numeric?("\t123\n")`, true},

		// Invalid inputs
		{`str.numeric?("")`, false},
		{`str.numeric?("hello")`, false},
		{`str.numeric?("12.34.56")`, false},
		{`str.numeric?("-")`, false},
		{`str.numeric?("1a2b")`, false},
		{`str.numeric?(" ")`, false},
		{`str.numeric?("12 34")`, false},
		{`str.numeric?("abc123")`, false},
		{`str.numeric?("123abc")`, false},
		{`str.numeric?(".")`, false},
		{`str.numeric?("+-1")`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestStrNumericErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.numeric?()`, "str.numeric? requires 1 argument, got=0"},
		{`str.numeric?("a", "b")`, "str.numeric? requires 1 argument, got=2"},
		{`str.numeric?(123)`, "str.numeric? requires string argument, got=INTEGER"},
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

func TestStrNumericChaining(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`"123" > str.numeric?()`, true},
		{`"  42  " > str.trim() > str.numeric?()`, true},
		{`"hello" > str.numeric?()`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestStrAlphanumericBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Valid alphanumeric
		{`str.alphanumeric?("abc123")`, true},
		{`str.alphanumeric?("Hello123")`, true},
		{`str.alphanumeric?("123")`, true},
		{`str.alphanumeric?("abc")`, true},

		// Invalid alphanumeric
		{`str.alphanumeric?("hello world")`, false},
		{`str.alphanumeric?("test@example")`, false},
		{`str.alphanumeric?("a_b_c")`, false},
		{`str.alphanumeric?("")`, false},
		{`str.alphanumeric?("test-123")`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestStrAlphanumericErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`str.alphanumeric?()`, "str.alphanumeric? requires 1 argument, got=0"},
		{`str.alphanumeric?(123)`, "str.alphanumeric? requires string argument, got=INTEGER"},
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

func TestStrAlphanumericChaining(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`"abc123" > str.alphanumeric?()`, true},
		{`"hello world" > str.alphanumeric?()`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}
