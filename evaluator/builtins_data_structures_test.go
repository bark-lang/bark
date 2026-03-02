package evaluator

import (
	"testing"

	"gitlab.com/bark-lang/barki/object"
)

// Tests for polymorphic data structure builtins (get, set, first, last, next, prev, head, tail)

func TestGetBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Array access
		{`[1, 2, 3] > get(0)`, 1},
		{`[1, 2, 3] > get(1)`, 2},
		{`[10, 20, 30] > get(2)`, 30},

		// Map access
		{`{"name": "John"} > get("name")`, "John"},
		{`{"age": "30"} > get("age")`, "30"},
		{`{"key": "value"} > get("missing")`, nil}, // NULL for missing keys
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case nil:
			if evaluated != NULL {
				t.Errorf("for input %q: expected NULL, got=%T (%+v)", tt.input, evaluated, evaluated)
			}
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			str, ok := evaluated.(*object.String)
			if !ok {
				t.Errorf("for input %q: object is not String. got=%T (%+v)", tt.input, evaluated, evaluated)
				continue
			}
			if str.Value != expected {
				t.Errorf("for input %q: String has wrong value. got=%q, want=%q", tt.input, str.Value, expected)
			}
		}
	}
}

func TestGetBuiltinErrors(t *testing.T) {
	// Programming errors - these crash the program
	programErrors := []struct {
		input           string
		expectedMessage string
	}{
		{`get([1, 2, 3])`, "wrong number of arguments. got=1, want=2+"},
		{`get([1, 2, 3], "key")`, "array index must be INTEGER, got STRING"},
		{`get({"a": "b"}, 1)`, "map key must be STRING, got INTEGER"},
		{`get(42, 0)`, "cannot index into INTEGER with get()"},
	}

	for _, tt := range programErrors {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("for input %q: no error object returned. got=%T(%+v)", tt.input, evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("for input %q: wrong error message. expected=%q, got=%q",
				tt.input, tt.expectedMessage, errObj.Msg)
		}
	}

	// Execution errors - these are logged and the chain stops, returning NULL
	// These are data-driven errors that don't crash the program
	execErrors := []struct {
		input           string
		expectedMessage string
	}{
		{`get([1, 2, 3], 5)`, "index out of bounds"},
		{`get([1, 2, 3], -1)`, "index out of bounds"},
		{`get({"a": 1}, "missing")`, "key not found"},
	}

	for _, tt := range execErrors {
		evaluated := testEval(tt.input)
		// Execution errors are logged and return NULL (chain stops but program continues)
		if evaluated.Type() != object.NULL_OBJ {
			t.Errorf("for input %q: expected NULL (execution error logged), got=%T(%+v)", tt.input, evaluated, evaluated)
		}
	}
}

func TestSetBuiltin(t *testing.T) {
	tests := []struct {
		input string
		check func(*testing.T, object.Object)
	}{
		{
			`[1, 2, 3] > set(1, 99)`,
			func(t *testing.T, obj object.Object) {
				arr, ok := obj.(*object.Array)
				if !ok {
					t.Errorf("object is not Array. got=%T", obj)
					return
				}
				if len(arr.Elements) != 3 {
					t.Errorf("array has wrong length. got=%d, want=3", len(arr.Elements))
					return
				}
				testIntegerObject(t, arr.Elements[0], 1)
				testIntegerObject(t, arr.Elements[1], 99)
				testIntegerObject(t, arr.Elements[2], 3)
			},
		},
		{
			`{"name": "John"} > set("name", "Jane")`,
			func(t *testing.T, obj object.Object) {
				m, ok := obj.(*object.Map)
				if !ok {
					t.Errorf("object is not Map. got=%T", obj)
					return
				}
				val, exists := m.Pairs["name"]
				if !exists {
					t.Errorf("key 'name' not found in map")
					return
				}
				str, ok := val.(*object.String)
				if !ok {
					t.Errorf("value is not String. got=%T", val)
					return
				}
				if str.Value != "Jane" {
					t.Errorf("value has wrong string. got=%q, want=%q", str.Value, "Jane")
				}
			},
		},
		{
			`{"a": "1"} > set("b", "2")`,
			func(t *testing.T, obj object.Object) {
				m, ok := obj.(*object.Map)
				if !ok {
					t.Errorf("object is not Map. got=%T", obj)
					return
				}
				if len(m.Pairs) != 2 {
					t.Errorf("map has wrong size. got=%d, want=2", len(m.Pairs))
					return
				}
				// Check that new key was added to Keys slice
				if len(m.Keys) != 2 {
					t.Errorf("map Keys has wrong length. got=%d, want=2", len(m.Keys))
				}
			},
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		tt.check(t, evaluated)
	}
}

func TestSetBuiltinErrors(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{`set([1, 2, 3], 0)`, "wrong number of arguments. got=2, want=3"},
		{`set([1, 2, 3], "key", 99)`, "array index must be INTEGER, got STRING"},
		{`set({"a": "b"}, 1, "val")`, "map key must be STRING, got INTEGER"},
		{`set([1, 2, 3], 5, 99)`, "index out of bounds: 5"},
		{`set(42, 0, 1)`, "first argument to `set` must be MAP or ARRAY, got INTEGER"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("for input %q: no error object returned. got=%T(%+v)", tt.input, evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("for input %q: wrong error message. expected=%q, got=%q",
				tt.input, tt.expectedMessage, errObj.Msg)
		}
	}
}

func TestFirstBuiltin(t *testing.T) {
	tests := []struct {
		input string
		check func(*testing.T, object.Object)
	}{
		{
			`[10, 20, 30] > first()`,
			func(t *testing.T, obj object.Object) {
				arr, ok := obj.(*object.Array)
				if !ok {
					t.Errorf("object is not Array. got=%T", obj)
					return
				}
				if len(arr.Elements) != 2 {
					t.Errorf("array has wrong length. got=%d, want=2", len(arr.Elements))
					return
				}
				testIntegerObject(t, arr.Elements[0], 0)  // index
				testIntegerObject(t, arr.Elements[1], 10) // value
			},
		},
		{
			`[] > first()`,
			func(t *testing.T, obj object.Object) {
				arr, ok := obj.(*object.Array)
				if !ok {
					t.Errorf("object is not Array. got=%T", obj)
					return
				}
				testIntegerObject(t, arr.Elements[0], -1) // -1 for empty
			},
		},
		{
			`{"name": "John", "age": "30"} > first()`,
			func(t *testing.T, obj object.Object) {
				arr, ok := obj.(*object.Array)
				if !ok {
					t.Errorf("object is not Array. got=%T", obj)
					return
				}
				if len(arr.Elements) != 2 {
					t.Errorf("array has wrong length. got=%d, want=2", len(arr.Elements))
					return
				}
				// Should return first key (insertion order)
				str, ok := arr.Elements[0].(*object.String)
				if !ok {
					t.Errorf("first element is not String. got=%T", arr.Elements[0])
					return
				}
				if str.Value != "name" {
					t.Errorf("first key is wrong. got=%q, want=%q", str.Value, "name")
				}
			},
		},
		{
			`{} > first()`,
			func(t *testing.T, obj object.Object) {
				arr, ok := obj.(*object.Array)
				if !ok {
					t.Errorf("object is not Array. got=%T", obj)
					return
				}
				str, ok := arr.Elements[0].(*object.String)
				if !ok {
					t.Errorf("first element is not String. got=%T", arr.Elements[0])
					return
				}
				if str.Value != "" {
					t.Errorf("first key should be empty. got=%q", str.Value)
				}
			},
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		tt.check(t, evaluated)
	}
}

func TestLastBuiltin(t *testing.T) {
	tests := []struct {
		input string
		check func(*testing.T, object.Object)
	}{
		{
			`[10, 20, 30] > last()`,
			func(t *testing.T, obj object.Object) {
				arr, ok := obj.(*object.Array)
				if !ok {
					t.Errorf("object is not Array. got=%T", obj)
					return
				}
				if len(arr.Elements) != 2 {
					t.Errorf("array has wrong length. got=%d, want=2", len(arr.Elements))
					return
				}
				testIntegerObject(t, arr.Elements[0], 2)  // index
				testIntegerObject(t, arr.Elements[1], 30) // value
			},
		},
		{
			`[] > last()`,
			func(t *testing.T, obj object.Object) {
				arr, ok := obj.(*object.Array)
				if !ok {
					t.Errorf("object is not Array. got=%T", obj)
					return
				}
				testIntegerObject(t, arr.Elements[0], -1) // -1 for empty
			},
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		tt.check(t, evaluated)
	}
}

func TestNextBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Array next
		{`[10, 20, 30] > next(0)`, 1},
		{`[10, 20, 30] > next(1)`, 2},
		{`[10, 20, 30] > next(2)`, -1}, // at end

		// Map next (based on insertion order)
		{`{"a": "1", "b": "2", "c": "3"} > next("a")`, "b"},
		{`{"a": "1", "b": "2", "c": "3"} > next("b")`, "c"},
		{`{"a": "1", "b": "2", "c": "3"} > next("c")`, ""}, // at end
		{`{"a": "1", "b": "2"} > next("missing")`, ""},     // not found
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			str, ok := evaluated.(*object.String)
			if !ok {
				t.Errorf("for input %q: object is not String. got=%T (%+v)", tt.input, evaluated, evaluated)
				continue
			}
			if str.Value != expected {
				t.Errorf("for input %q: String has wrong value. got=%q, want=%q", tt.input, str.Value, expected)
			}
		}
	}
}

func TestPrevBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Array prev
		{`[10, 20, 30] > prev(2)`, 1},
		{`[10, 20, 30] > prev(1)`, 0},
		{`[10, 20, 30] > prev(0)`, -1}, // at start

		// Map prev
		{`{"a": "1", "b": "2", "c": "3"} > prev("c")`, "b"},
		{`{"a": "1", "b": "2", "c": "3"} > prev("b")`, "a"},
		{`{"a": "1", "b": "2", "c": "3"} > prev("a")`, ""}, // at start
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			str, ok := evaluated.(*object.String)
			if !ok {
				t.Errorf("for input %q: object is not String. got=%T (%+v)", tt.input, evaluated, evaluated)
				continue
			}
			if str.Value != expected {
				t.Errorf("for input %q: String has wrong value. got=%q, want=%q", tt.input, str.Value, expected)
			}
		}
	}
}

func TestHeadBuiltin(t *testing.T) {
	tests := []struct {
		input string
		check func(*testing.T, object.Object)
	}{
		{
			`[10, 20, 30] > head()`,
			func(t *testing.T, obj object.Object) {
				tuple, ok := obj.(*object.Tuple)
				if !ok {
					t.Errorf("object is not Tuple. got=%T", obj)
					return
				}
				if len(tuple.Elements) != 2 {
					t.Errorf("tuple has wrong length. got=%d, want=2", len(tuple.Elements))
					return
				}
				// Second element should be index 0
				testIntegerObject(t, tuple.Elements[1], 0)
			},
		},
		{
			`[] > head()`,
			func(t *testing.T, obj object.Object) {
				tuple, ok := obj.(*object.Tuple)
				if !ok {
					t.Errorf("object is not Tuple. got=%T", obj)
					return
				}
				// Second element should be -1 for empty
				testIntegerObject(t, tuple.Elements[1], -1)
			},
		},
		{
			`{"name": "John"} > head()`,
			func(t *testing.T, obj object.Object) {
				tuple, ok := obj.(*object.Tuple)
				if !ok {
					t.Errorf("object is not Tuple. got=%T", obj)
					return
				}
				// Second element should be first key
				str, ok := tuple.Elements[1].(*object.String)
				if !ok {
					t.Errorf("second element is not String. got=%T", tuple.Elements[1])
					return
				}
				if str.Value != "name" {
					t.Errorf("key is wrong. got=%q, want=%q", str.Value, "name")
				}
			},
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		tt.check(t, evaluated)
	}
}

func TestTailBuiltin(t *testing.T) {
	tests := []struct {
		input string
		check func(*testing.T, object.Object)
	}{
		{
			`[10, 20, 30] > tail()`,
			func(t *testing.T, obj object.Object) {
				tuple, ok := obj.(*object.Tuple)
				if !ok {
					t.Errorf("object is not Tuple. got=%T", obj)
					return
				}
				if len(tuple.Elements) != 2 {
					t.Errorf("tuple has wrong length. got=%d, want=2", len(tuple.Elements))
					return
				}
				// Second element should be last index (2)
				testIntegerObject(t, tuple.Elements[1], 2)
			},
		},
		{
			`[] > tail()`,
			func(t *testing.T, obj object.Object) {
				tuple, ok := obj.(*object.Tuple)
				if !ok {
					t.Errorf("object is not Tuple. got=%T", obj)
					return
				}
				// Second element should be -1 for empty
				testIntegerObject(t, tuple.Elements[1], -1)
			},
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		tt.check(t, evaluated)
	}
}

func TestDataStructureBuiltinErrors(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{`first()`, "wrong number of arguments. got=0, want=1"},
		{`first([1, 2], 3)`, "wrong number of arguments. got=2, want=1"},
		{`first(42)`, "argument to `first` must be MAP or ARRAY, got INTEGER"},

		{`last()`, "wrong number of arguments. got=0, want=1"},
		{`last(42)`, "argument to `last` must be MAP or ARRAY, got INTEGER"},

		{`next([1, 2])`, "wrong number of arguments. got=1, want=2"},
		{`next([1, 2], "key")`, "second argument to `next` must be INTEGER for arrays, got STRING"},
		{`next({"a": "b"}, 1)`, "second argument to `next` must be STRING for maps, got INTEGER"},
		{`next(42, 0)`, "first argument to `next` must be MAP or ARRAY, got INTEGER"},

		{`prev([1, 2])`, "wrong number of arguments. got=1, want=2"},
		{`prev([1, 2], "key")`, "second argument to `prev` must be INTEGER for arrays, got STRING"},
		{`prev({"a": "b"}, 1)`, "second argument to `prev` must be STRING for maps, got INTEGER"},
		{`prev(42, 0)`, "first argument to `prev` must be MAP or ARRAY, got INTEGER"},

		{`head()`, "wrong number of arguments. got=0, want=1"},
		{`head(42)`, "argument to `head` must be MAP or ARRAY, got INTEGER"},

		{`tail()`, "wrong number of arguments. got=0, want=1"},
		{`tail(42)`, "argument to `tail` must be MAP or ARRAY, got INTEGER"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("for input %q: no error object returned. got=%T(%+v)", tt.input, evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("for input %q: wrong error message. expected=%q, got=%q",
				tt.input, tt.expectedMessage, errObj.Msg)
		}
	}
}
