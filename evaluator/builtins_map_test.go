package evaluator

import (
	"testing"

	"gitlab.com/bark-lang/barki/object"
)

func TestMapGetOrBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`{"a": 1, "b": 2} > map.get_or("a", 0)`, int64(1)},
		{`{"a": 1, "b": 2} > map.get_or("c", 99)`, int64(99)},
		{`{"name": "Alice"} > map.get_or("name", "Unknown")`, "Alice"},
		{`{"name": "Alice"} > map.get_or("age", "Unknown")`, "Unknown"},
		{`{} > map.get_or("key", "default")`, "default"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int64:
			testIntegerObject(t, evaluated, expected)
		case string:
			testStringObject(t, evaluated, expected)
		}
	}
}

func TestMapGetOrErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`map.get_or()`, "map.get_or requires 3 arguments (map, key, default), got=0"},
		{`map.get_or({}, "a")`, "map.get_or requires 3 arguments (map, key, default), got=2"},
		{`"not a map" > map.get_or("a", 0)`, "map.get_or requires map as first argument, got=STRING"},
		{`{} > map.get_or(123, 0)`, "map.get_or requires string key, got=INTEGER"},
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

func TestMapDelBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`{"a": 1, "b": 2, "c": 3} > map.del("b")`, `{a: 1, c: 3}`},
		{`{"a": 1, "b": 2, "c": 3} > map.del("a", "c")`, `{b: 2}`},
		{`{"a": 1} > map.del("a")`, `{}`},
		{`{"a": 1} > map.del("nonexistent")`, `{a: 1}`},
		{`{} > map.del("key")`, `{}`},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		m, ok := evaluated.(*object.Map)
		if !ok {
			t.Errorf("object is not Map. got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if m.Inspect() != tt.expected {
			t.Errorf("wrong result for %q. expected=%s, got=%s", tt.input, tt.expected, m.Inspect())
		}
	}
}

func TestMapDelErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`map.del()`, "map.del requires at least 2 arguments (map, key...), got=0"},
		{`map.del({})`, "map.del requires at least 2 arguments (map, key...), got=1"},
		{`"not a map" > map.del("a")`, "map.del requires map as first argument, got=STRING"},
		{`{} > map.del(123)`, "map.del requires string keys, got=INTEGER"},
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

func TestMapKeysBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`{"a": 1, "b": 2, "c": 3} > map.keys()`, `[a, b, c]`},
		{`{"z": 1, "a": 2} > map.keys()`, `[z, a]`},
		{`{} > map.keys()`, `[]`},
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

func TestMapKeysErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`map.keys()`, "map.keys requires 1 argument (map), got=0"},
		{`map.keys({}, {})`, "map.keys requires 1 argument (map), got=2"},
		{`"not a map" > map.keys()`, "map.keys requires map argument, got=STRING"},
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

func TestMapValuesBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`{"a": 1, "b": 2, "c": 3} > map.values()`, `[1, 2, 3]`},
		{`{"name": "Alice", "age": "30"} > map.values()`, `[Alice, 30]`},
		{`{} > map.values()`, `[]`},
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

func TestMapValuesErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`map.values()`, "map.values requires 1 argument (map), got=0"},
		{`map.values({}, {})`, "map.values requires 1 argument (map), got=2"},
		{`"not a map" > map.values()`, "map.values requires map argument, got=STRING"},
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

func TestMapEntriesBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`{"a": 1, "b": 2} > map.entries()`, `[[a, 1], [b, 2]]`},
		{`{"name": "Alice"} > map.entries()`, `[[name, Alice]]`},
		{`{} > map.entries()`, `[]`},
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

func TestMapEntriesErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`map.entries()`, "map.entries requires 1 argument (map), got=0"},
		{`map.entries({}, {})`, "map.entries requires 1 argument (map), got=2"},
		{`"not a map" > map.entries()`, "map.entries requires map argument, got=STRING"},
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

func TestMapKeyPresentBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`{"a": 1, "b": 2} > map.key_present?("a")`, true},
		{`{"a": 1, "b": 2} > map.key_present?("c")`, false},
		{`{} > map.key_present?("key")`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestMapKeyPresentErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`map.key_present?()`, "map.key_present? requires 2 arguments (map, key), got=0"},
		{`map.key_present?({})`, "map.key_present? requires 2 arguments (map, key), got=1"},
		{`"not a map" > map.key_present?("a")`, "map.key_present? requires map as first argument, got=STRING"},
		{`{} > map.key_present?(123)`, "map.key_present? requires string key, got=INTEGER"},
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

func TestMapKeyAbsentBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`{"a": 1, "b": 2} > map.key_absent?("a")`, false},
		{`{"a": 1, "b": 2} > map.key_absent?("c")`, true},
		{`{} > map.key_absent?("key")`, true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestMapKeyAbsentErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`map.key_absent?()`, "map.key_absent? requires 2 arguments (map, key), got=0"},
		{`map.key_absent?({})`, "map.key_absent? requires 2 arguments (map, key), got=1"},
		{`"not a map" > map.key_absent?("a")`, "map.key_absent? requires map as first argument, got=STRING"},
		{`{} > map.key_absent?(123)`, "map.key_absent? requires string key, got=INTEGER"},
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

func TestMapMergeBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`{"a": 1} > map.merge({"b": 2})`, `{a: 1, b: 2}`},
		{`{"a": 1} > map.merge({"a": 99})`, `{a: 99}`},
		{`{"a": 1} > map.merge({"b": 2}, {"c": 3})`, `{a: 1, b: 2, c: 3}`},
		{`{} > map.merge({})`, `{}`},
		{`{} > map.merge({"a": 1})`, `{a: 1}`},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		m, ok := evaluated.(*object.Map)
		if !ok {
			t.Errorf("object is not Map. got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if m.Inspect() != tt.expected {
			t.Errorf("wrong result for %q. expected=%s, got=%s", tt.input, tt.expected, m.Inspect())
		}
	}
}

func TestMapMergeErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`map.merge()`, "map.merge requires at least 2 arguments (map, map...), got=0"},
		{`map.merge({})`, "map.merge requires at least 2 arguments (map, map...), got=1"},
		{`"not a map" > map.merge({})`, "map.merge requires map arguments, got=STRING"},
		{`{} > map.merge("not a map")`, "map.merge requires map arguments, got=STRING"},
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
func TestMapChaining(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Chain map operations
		{`{"a": 1, "b": 2} > map.del("a") > map.merge({"c": 3})`, `{b: 2, c: 3}`},
		{`{"a": 1, "b": 2} > map.keys() > len()`, int64(2)},
		{`{"a": 1, "b": 2} > map.values() > get(0)`, int64(1)},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case string:
			m, ok := evaluated.(*object.Map)
			if !ok {
				t.Errorf("object is not Map. got=%T (%+v)", evaluated, evaluated)
				continue
			}
			if m.Inspect() != expected {
				t.Errorf("wrong result for %q. expected=%s, got=%s", tt.input, expected, m.Inspect())
			}
		case int64:
			testIntegerObject(t, evaluated, expected)
		}
	}
}
