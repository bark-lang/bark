package evaluator

import (
	"testing"

	"gitlab.com/bark-lang/bark/object"
)

func TestReturnQuestionBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// return?(false) should return NULL (no early return)
		{`return?(false)`, nil},

		// return?(true) should return early with NULL
		{
			`fn test() {
				return?(true)
				return(99)
			}(int)
			test()`,
			nil,
		},

		// return?(true, value) should return early with value
		{
			`fn test() {
				return?(true, 42)
				return(99)
			}(int)
			test()`,
			42,
		},

		// return?(false, value) should not return early
		{
			`fn test() {
				return?(false, 42)
				return(99)
			}(int)
			test()`,
			99,
		},

		// return?(true, val1, val2) should return tuple of values
		{
			`fn test() {
				return?(true, 1, 2)
				return(99)
			}(int, int)
			test()`,
			[]int64{1, 2},
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case nil:
			if evaluated != NULL {
				t.Errorf("expected NULL, got=%T (%+v)", evaluated, evaluated)
			}
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case []int64:
			tuple, ok := evaluated.(*object.Tuple)
			if !ok {
				t.Errorf("object is not Tuple. got=%T (%+v)", evaluated, evaluated)
				continue
			}
			if len(tuple.Elements) != len(expected) {
				t.Errorf("tuple has wrong length. got=%d, want=%d", len(tuple.Elements), len(expected))
				continue
			}
			for i, exp := range expected {
				testIntegerObject(t, tuple.Elements[i], exp)
			}
		}
	}
}

func TestReturnQuestionErrors(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"return?()",
			"wrong number of arguments. got=0, want at least 1",
		},
		{
			"return?(42)",
			"first argument to `return?` must be BOOLEAN, got INTEGER",
		},
		{
			`return?("true")`,
			"first argument to `return?` must be BOOLEAN, got STRING",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)", evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tt.expectedMessage, errObj.Msg)
		}
	}
}

func TestBreakQuestionBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// break?(false) should return NULL (no early exit)
		{`break?(false)`, nil},

		// break?(true, value) should exit with value
		{
			`fn test() {
				break?(true, 42)
				return(99)
			}(int)
			test()`,
			42,
		},

		// break?(false, value) should not exit early
		{
			`fn test() {
				break?(false, 42)
				return(99)
			}(int)
			test()`,
			99,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case nil:
			if evaluated != NULL {
				t.Errorf("expected NULL, got=%T (%+v)", evaluated, evaluated)
			}
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		}
	}
}

func TestBreakQuestionErrors(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"break?()",
			"wrong number of arguments. got=0, want at least 1",
		},
		{
			"break?(42)",
			"first argument to `break?` must be BOOLEAN, got INTEGER",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)", evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tt.expectedMessage, errObj.Msg)
		}
	}
}

func TestContinueQuestionBuiltin(t *testing.T) {
	// continue?(true) should return Void to allow chain to continue
	// without passing a value to the next function
	evaluated := testEval(`continue?(true)`)
	if _, ok := evaluated.(*object.Void); !ok {
		t.Errorf("continue?(true): expected Void, got=%T (%+v)", evaluated, evaluated)
	}

	// continue?(false) should return ChainStop to halt the chain
	evaluated = testEval(`continue?(false)`)
	if _, ok := evaluated.(*object.ChainStop); !ok {
		t.Errorf("continue?(false): expected ChainStop, got=%T (%+v)", evaluated, evaluated)
	}
}

func TestContinueQuestionChainBehavior(t *testing.T) {
	// When continue?(true), the chain should continue
	// true > continue?() > "reached" should evaluate to "reached"
	evaluated := testEval(`true > continue?() > "reached"`)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("continue?(true) chain: expected String, got=%T (%+v)", evaluated, evaluated)
	}
	if str.Value != "reached" {
		t.Errorf("continue?(true) chain: expected 'reached', got=%q", str.Value)
	}

	// When continue?(false), the chain should stop and not reach subsequent operations
	// false > continue?() > "reached" should return ChainStop (not "reached")
	evaluated = testEval(`false > continue?() > "reached"`)
	if _, ok := evaluated.(*object.ChainStop); !ok {
		t.Errorf("continue?(false) chain: expected ChainStop (chain stopped before 'reached'), got=%T (%+v)", evaluated, evaluated)
	}
}

func TestContinueQuestionErrors(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"continue?()",
			"wrong number of arguments. got=0, want=1",
		},
		{
			"continue?(1, 2)",
			"wrong number of arguments. got=2, want=1",
		},
		{
			"continue?(42)",
			"argument to `continue?` must be BOOLEAN, got INTEGER",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)", evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tt.expectedMessage, errObj.Msg)
		}
	}
}

func TestRepeatQuestionBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// repeat?(false) should return NULL (no repeat)
		{`repeat?(false)`, "NULL"},

		// repeat?(true) should return RepeatValue
		{`repeat?(true)`, "REPEAT_VALUE"},

		// repeat?(true, args...) should return RepeatValue with args
		{`repeat?(true, 1, 2)`, "REPEAT_VALUE"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch tt.expected {
		case "NULL":
			if evaluated != NULL {
				t.Errorf("expected NULL, got=%T (%+v)", evaluated, evaluated)
			}
		case "REPEAT_VALUE":
			_, ok := evaluated.(*object.RepeatValue)
			if !ok {
				t.Errorf("expected RepeatValue, got=%T (%+v)", evaluated, evaluated)
			}
		}
	}
}

func TestRepeatQuestionBuiltinErrors(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		// repeat?() with no args is an error
		{`repeat?()`, "wrong number of arguments. got=0, want at least 1"},

		// repeat?(non-boolean) is an error
		{`repeat?(1, 2, 3)`, "first argument to `repeat?` must be BOOLEAN, got INTEGER"},
		{`repeat?("hello")`, "first argument to `repeat?` must be BOOLEAN, got STRING"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error for input %q, got=%T (%+v)", tt.input, evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("wrong error message for input %q.\nexpected=%q\ngot=%q", tt.input, tt.expectedMessage, errObj.Msg)
		}
	}
}
