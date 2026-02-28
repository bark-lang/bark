package evaluator

import (
	"testing"

	"gitlab.com/bark-lang/barki/object"
)

func TestRegexMatchBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Basic matching
		{`regex.match?("hello", "hello")`, true},
		{`regex.match?("hello", "world")`, false},

		// Pattern matching
		{`regex.match?("test123", "\\d+")`, true},
		{`regex.match?("no numbers", "\\d+")`, false},

		// Case sensitivity
		{`regex.match?("Hello", "hello")`, false},
		{`regex.match?("Hello", "(?i)hello")`, true},

		// Partial match
		{`regex.match?("hello world", "world")`, true},
		{`regex.match?("hello world", "goodbye")`, false},

		// Special characters
		{`regex.match?("test@example.com", "\\w+@\\w+\\.\\w+")`, true},
		{`regex.match?("not an email", "\\w+@\\w+\\.\\w+")`, false},

		// Empty string
		{`regex.match?("", "")`, true},
		{`regex.match?("", "test")`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestRegexMatchInvalidPattern(t *testing.T) {
	// Invalid regex pattern returns ExecutionError (logged to stderr, becomes NULL)
	input := `regex.match?("test", "[")`
	evaluated := testEval(input)
	if _, ok := evaluated.(*object.Null); !ok {
		t.Errorf("expected NULL after execution error, got %T", evaluated)
	}
}

func TestRegexMatchErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`regex.match?()`, "regex.match? requires 2 arguments (text, pattern), got=0"},
		{`regex.match?("text")`, "regex.match? requires 2 arguments (text, pattern), got=1"},
		{`regex.match?("a", "b", "c")`, "regex.match? requires 2 arguments (text, pattern), got=3"},

		// Wrong argument types
		{`regex.match?(123, "pattern")`, "regex.match? requires string text, got=INTEGER"},
		{`regex.match?("text", 123)`, "regex.match? requires string pattern, got=INTEGER"},
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

func TestRegexFindBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Find first match
		{`regex.find("test123abc456", "\\d+")`, "123"},
		{`regex.find("hello world", "world")`, "world"},

		// No match returns empty string
		{`regex.find("no numbers", "\\d+")`, ""},

		// Find with groups
		{`regex.find("email@example.com", "\\w+")`, "email"},

		// Empty pattern
		{`regex.find("test", "")`, ""},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestRegexFindInvalidPattern(t *testing.T) {
	// Invalid regex pattern returns ExecutionError (logged to stderr, becomes NULL)
	input := `regex.find("test", "[")`
	evaluated := testEval(input)
	if _, ok := evaluated.(*object.Null); !ok {
		t.Errorf("expected NULL after execution error, got %T", evaluated)
	}
}

func TestRegexFindErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`regex.find()`, "regex.find requires 2 arguments (text, pattern), got=0"},
		{`regex.find("text")`, "regex.find requires 2 arguments (text, pattern), got=1"},
		{`regex.find("a", "b", "c")`, "regex.find requires 2 arguments (text, pattern), got=3"},

		// Wrong argument types
		{`regex.find(123, "pattern")`, "regex.find requires string text, got=INTEGER"},
		{`regex.find("text", 123)`, "regex.find requires string pattern, got=INTEGER"},
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

func TestRegexFindAllBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		// Find all matches
		{`regex.find_all("test123abc456def789", "\\d+")`, []string{"123", "456", "789"}},
		{`regex.find_all("one two three", "\\w+")`, []string{"one", "two", "three"}},

		// No matches returns empty array
		{`regex.find_all("no numbers", "\\d+") > len()`, []string{}}, // Will check length instead

		// Single match
		{`regex.find_all("hello", "hello")`, []string{"hello"}},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		// Special case for length check
		if len(tt.expected) == 0 {
			testIntegerObject(t, evaluated, 0)
			continue
		}

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

func TestRegexFindAllInvalidPattern(t *testing.T) {
	// Invalid regex pattern returns ExecutionError (logged to stderr, becomes NULL)
	input := `regex.find_all("test", "[")`
	evaluated := testEval(input)
	if _, ok := evaluated.(*object.Null); !ok {
		t.Errorf("expected NULL after execution error, got %T", evaluated)
	}
}

func TestRegexFindAllErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`regex.find_all()`, "regex.find_all requires 2 arguments (text, pattern), got=0"},
		{`regex.find_all("text")`, "regex.find_all requires 2 arguments (text, pattern), got=1"},

		// Wrong argument types
		{`regex.find_all(123, "pattern")`, "regex.find_all requires string text, got=INTEGER"},
		{`regex.find_all("text", 123)`, "regex.find_all requires string pattern, got=INTEGER"},
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

func TestRegexReplaceBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic replace
		{`regex.replace("hello world", "world", "universe")`, "hello universe"},
		{`regex.replace("test123", "\\d+", "XXX")`, "testXXX"},

		// Replace all occurrences
		{`regex.replace("one two three", "\\w+", "X")`, "X X X"},

		// No match returns original
		{`regex.replace("hello", "goodbye", "XXX")`, "hello"},

		// Empty replacement
		{`regex.replace("test123", "\\d+", "")`, "test"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestRegexReplaceInvalidPattern(t *testing.T) {
	// Invalid regex pattern returns ExecutionError (logged to stderr, becomes NULL)
	input := `regex.replace("test", "[", "X")`
	evaluated := testEval(input)
	if _, ok := evaluated.(*object.Null); !ok {
		t.Errorf("expected NULL after execution error, got %T", evaluated)
	}
}

func TestRegexReplaceErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`regex.replace()`, "regex.replace requires 3 arguments (text, pattern, replacement), got=0"},
		{`regex.replace("a", "b")`, "regex.replace requires 3 arguments (text, pattern, replacement), got=2"},

		// Wrong argument types
		{`regex.replace(123, "pattern", "repl")`, "regex.replace requires string text, got=INTEGER"},
		{`regex.replace("text", 123, "repl")`, "regex.replace requires string pattern, got=INTEGER"},
		{`regex.replace("text", "pattern", 123)`, "regex.replace requires string replacement, got=INTEGER"},
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

func TestRegexSplitBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		// Basic split
		{`regex.split("one,two,three", ",")`, []string{"one", "two", "three"}},
		{`regex.split("a1b2c3", "\\d")`, []string{"a", "b", "c", ""}},

		// Split on whitespace
		{`regex.split("one two  three", "\\s+")`, []string{"one", "two", "three"}},

		// No match returns original in array
		{`regex.split("hello", "x") > len()`, []string{}}, // Check length
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		// Special case for length check
		if len(tt.expected) == 0 {
			testIntegerObject(t, evaluated, 1) // Original string in array
			continue
		}

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

func TestRegexSplitInvalidPattern(t *testing.T) {
	// Invalid regex pattern returns ExecutionError (logged to stderr, becomes NULL)
	input := `regex.split("test", "[")`
	evaluated := testEval(input)
	if _, ok := evaluated.(*object.Null); !ok {
		t.Errorf("expected NULL after execution error, got %T", evaluated)
	}
}

func TestRegexSplitErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`regex.split()`, "regex.split requires 2 arguments (text, pattern), got=0"},
		{`regex.split("text")`, "regex.split requires 2 arguments (text, pattern), got=1"},

		// Wrong argument types
		{`regex.split(123, "pattern")`, "regex.split requires string text, got=INTEGER"},
		{`regex.split("text", 123)`, "regex.split requires string pattern, got=INTEGER"},
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
