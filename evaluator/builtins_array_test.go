package evaluator

import (
	"testing"

	"gitlab.com/bark-lang/barki/object"
)

func TestArrayDedupe(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`[1, 2, 3, 1, 2, 3] > array.dedupe()`, "[1, 2, 3]"},
		{`[1, 1, 1, 1] > array.dedupe()`, "[1]"},
		{`[1, 2, 3, 4, 5] > array.dedupe()`, "[1, 2, 3, 4, 5]"},
		{`[] > array.dedupe()`, "[]"},
		{`[5, 4, 3, 2, 1, 1, 2, 3, 4, 5] > array.dedupe()`, "[5, 4, 3, 2, 1]"},
		{`["a", "b", "a", "c", "b"] > array.dedupe()`, "[a, b, c]"},
		{`[true, false, true, false] > array.dedupe()`, "[true, false]"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Errorf("object is not Array. got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if arr.Inspect() != tt.expected {
			t.Errorf("wrong result for %q. expected=%s, got=%s", tt.input, tt.expected, arr.Inspect())
		}
	}
}

func TestArrayDedupeErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`array.dedupe()`, "array.dedupe requires 1 argument, got=0"},
		{`"hello" > array.dedupe()`, "array.dedupe requires array argument, got=STRING"},
		{`123 > array.dedupe()`, "array.dedupe requires array argument, got=INTEGER"},
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

func TestArrayPushBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`[1, 2] > array.push(3)`, "[1, 2, 3]"},
		{`[1] > array.push(2, 3, 4)`, "[1, 2, 3, 4]"},
		{`[] > array.push(1)`, "[1]"},
		{`["a", "b"] > array.push("c")`, "[a, b, c]"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Errorf("object is not Array. got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if arr.Inspect() != tt.expected {
			t.Errorf("wrong result for %q. expected=%s, got=%s", tt.input, tt.expected, arr.Inspect())
		}
	}
}

func TestArrayPushErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`array.push()`, "array.push requires at least 2 arguments (array, value...), got=0"},
		{`array.push([])`, "array.push requires at least 2 arguments (array, value...), got=1"},
		{`"not array" > array.push(1)`, "array.push requires array as first argument, got=STRING"},
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

func TestArrayAppendToBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`3 > array.append_to([1, 2])`, "[1, 2, 3]"},
		{`"c" > array.append_to(["a", "b"])`, "[a, b, c]"},
		{`1 > array.append_to([])`, "[1]"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Errorf("object is not Array. got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if arr.Inspect() != tt.expected {
			t.Errorf("wrong result for %q. expected=%s, got=%s", tt.input, tt.expected, arr.Inspect())
		}
	}
}

func TestArrayAppendToErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`array.append_to()`, "array.append_to requires 2 arguments (value, array), got=0"},
		{`array.append_to(1)`, "array.append_to requires 2 arguments (value, array), got=1"},
		{`1 > array.append_to("not array")`, "array.append_to requires array as second argument, got=STRING"},
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

func TestArrayPopBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`[1, 2, 3] > array.pop()`, "[[1, 2], 3]"},
		{`[1] > array.pop()`, "[[], 1]"},
		{`["a", "b"] > array.pop()`, "[[a], b]"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Errorf("object is not Array. got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if arr.Inspect() != tt.expected {
			t.Errorf("wrong result for %q. expected=%s, got=%s", tt.input, tt.expected, arr.Inspect())
		}
	}
}

func TestArrayPopErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`array.pop()`, "array.pop requires 1 argument (array), got=0"},
		{`[] > array.pop()`, "cannot pop from empty array"},
		{`"not array" > array.pop()`, "array.pop requires array argument, got=STRING"},
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

func TestArrayShiftBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`[1, 2, 3] > array.shift()`, "[[2, 3], 1]"},
		{`[1] > array.shift()`, "[[], 1]"},
		{`["a", "b"] > array.shift()`, "[[b], a]"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Errorf("object is not Array. got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if arr.Inspect() != tt.expected {
			t.Errorf("wrong result for %q. expected=%s, got=%s", tt.input, tt.expected, arr.Inspect())
		}
	}
}

func TestArrayShiftErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`array.shift()`, "array.shift requires 1 argument (array), got=0"},
		{`[] > array.shift()`, "cannot shift from empty array"},
		{`"not array" > array.shift()`, "array.shift requires array argument, got=STRING"},
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

func TestArrayUnshiftBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`[2, 3] > array.unshift(1)`, "[1, 2, 3]"},
		{`[3] > array.unshift(1, 2)`, "[1, 2, 3]"},
		{`[] > array.unshift(1)`, "[1]"},
		{`["b", "c"] > array.unshift("a")`, "[a, b, c]"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Errorf("object is not Array. got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if arr.Inspect() != tt.expected {
			t.Errorf("wrong result for %q. expected=%s, got=%s", tt.input, tt.expected, arr.Inspect())
		}
	}
}

func TestArrayUnshiftErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`array.unshift()`, "array.unshift requires at least 2 arguments (array, value...), got=0"},
		{`array.unshift([])`, "array.unshift requires at least 2 arguments (array, value...), got=1"},
		{`"not array" > array.unshift(1)`, "array.unshift requires array as first argument, got=STRING"},
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

func TestArraySliceBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`[1, 2, 3, 4, 5] > array.slice(1, 3)`, "[2, 3]"},
		{`[1, 2, 3, 4, 5] > array.slice(0, 2)`, "[1, 2]"},
		{`[1, 2, 3, 4, 5] > array.slice(3, 5)`, "[4, 5]"},
		{`[1, 2, 3] > array.slice(0, 3)`, "[1, 2, 3]"},
		{`[1, 2, 3] > array.slice(1, 1)`, "[]"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Errorf("object is not Array. got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if arr.Inspect() != tt.expected {
			t.Errorf("wrong result for %q. expected=%s, got=%s", tt.input, tt.expected, arr.Inspect())
		}
	}
}

func TestArraySliceErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`array.slice()`, "array.slice requires 3 arguments (array, start, end), got=0"},
		{`array.slice([], 0)`, "array.slice requires 3 arguments (array, start, end), got=2"},
		{`"not array" > array.slice(0, 1)`, "array.slice requires array as first argument, got=STRING"},
		{`[1, 2, 3] > array.slice("a", 1)`, "array.slice requires integer start index, got=STRING"},
		{`[1, 2, 3] > array.slice(0, "b")`, "array.slice requires integer end index, got=STRING"},
		{`[1, 2, 3] > array.slice(-1, 2)`, "slice start index out of bounds: -1"},
		{`[1, 2, 3] > array.slice(0, 10)`, "slice end index out of bounds: 10"},
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

func TestArrayRangeBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`array.range(1, 5)`, "[1, 2, 3, 4, 5]"},
		{`array.range(0, 3)`, "[0, 1, 2, 3]"},
		{`array.range(5, 5)`, "[5]"},
		{`array.range(-2, 2)`, "[-2, -1, 0, 1, 2]"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Errorf("object is not Array. got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if arr.Inspect() != tt.expected {
			t.Errorf("wrong result for %q. expected=%s, got=%s", tt.input, tt.expected, arr.Inspect())
		}
	}
}

func TestArrayRangeErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`array.range()`, "array.range requires 2 arguments (start, end), got=0"},
		{`array.range(1)`, "array.range requires 2 arguments (start, end), got=1"},
		{`array.range("a", 5)`, "array.range requires integer start, got=STRING"},
		{`array.range(1, "b")`, "array.range requires integer end, got=STRING"},
		{`array.range(5, 1)`, "range end (1) must be >= start (5)"},
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
func TestArrayChaining(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`[1, 2, 2, 3, 3, 3] > array.dedupe() > array.push(4)`, "[1, 2, 3, 4]"},
		{`array.range(1, 3) > array.push(4, 5)`, "[1, 2, 3, 4, 5]"},
		{`[1, 2, 3] > array.slice(0, 2) > array.push(99)`, "[1, 2, 99]"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Errorf("object is not Array. got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if arr.Inspect() != tt.expected {
			t.Errorf("wrong result for %q. expected=%s, got=%s", tt.input, tt.expected, arr.Inspect())
		}
	}
}
