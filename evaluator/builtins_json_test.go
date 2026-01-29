package evaluator

import (
	"testing"

	"gitlab.com/bark-lang/bark/object"
)

func TestJSONParseBuiltin(t *testing.T) {
	tests := []struct {
		input   string
		checkFn func(object.Object) bool
		desc    string
	}{
		// Parse simple string - use capture() to extract result
		{
			`json.parse("{\"name\": \"Alice\"}") > capture(e, r)
r > get("name")`,
			func(obj object.Object) bool {
				str, ok := obj.(*object.String)
				return ok && str.Value == "Alice"
			},
			"parse simple object",
		},

		// Parse number
		{
			`json.parse("{\"age\": 30}") > capture(e, r)
r > get("age")`,
			func(obj object.Object) bool {
				num, ok := obj.(*object.Integer)
				return ok && num.Value == 30
			},
			"parse number",
		},

		// Parse boolean
		{
			`json.parse("{\"active\": true}") > capture(e, r)
r > get("active")`,
			func(obj object.Object) bool {
				b, ok := obj.(*object.Boolean)
				return ok && b.Value == true
			},
			"parse boolean",
		},

		// Parse array
		{
			`json.parse("[1, 2, 3]") > capture(e, r)
r > len()`,
			func(obj object.Object) bool {
				num, ok := obj.(*object.Integer)
				return ok && num.Value == 3
			},
			"parse array length",
		},

		// Parse nested object
		{
			`json.parse("{\"user\": {\"name\": \"Bob\"}}") > capture(e, r)
r > get("user") > get("name")`,
			func(obj object.Object) bool {
				str, ok := obj.(*object.String)
				return ok && str.Value == "Bob"
			},
			"parse nested object",
		},

		// Parse empty object
		{
			`json.parse("{}") > capture(e, r)
r > size()`,
			func(obj object.Object) bool {
				num, ok := obj.(*object.Integer)
				return ok && num.Value == 0
			},
			"parse empty object",
		},

		// Parse empty array
		{
			`json.parse("[]") > capture(e, r)
r > len()`,
			func(obj object.Object) bool {
				num, ok := obj.(*object.Integer)
				return ok && num.Value == 0
			},
			"parse empty array",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if !tt.checkFn(evaluated) {
			t.Errorf("%s failed. got=%T (%+v)", tt.desc, evaluated, evaluated)
		}
	}
}

func TestJSONParseSuccess(t *testing.T) {
	// Test that successful parse returns empty map as error
	// After capture(e, r), e is bound to error and r to result
	input := `json.parse("{\"test\": 123}") > capture(e, r)
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

func TestJSONParseErrors(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
		desc        string
	}{
		// Invalid JSON
		{`json.parse("{invalid}")`, true, "invalid JSON object"},
		{`json.parse("[1, 2,]")`, true, "trailing comma"},
		{`json.parse("not json at all")`, true, "not JSON"},
		{`json.parse("")`, true, "empty string"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		// Should return tuple with error at index 0
		tpl, ok := evaluated.(*object.Tuple)
		if !ok {
			t.Errorf("%s: expected Tuple (error tuple), got=%T (%+v)", tt.desc, evaluated, evaluated)
			continue
		}

		if len(tpl.Elements) != 2 {
			t.Errorf("%s: expected tuple with 2 elements, got=%d", tt.desc, len(tpl.Elements))
			continue
		}

		// First element should be Error object if test expects error
		if tt.shouldError {
			_, isErr := tpl.Elements[0].(*object.Error)
			_, isMap := tpl.Elements[0].(*object.Map)
			if !isErr && !isMap {
				t.Errorf("%s: expected Error or Map at index 0, got=%T", tt.desc, tpl.Elements[0])
			}
			if isErr {
				// Successfully got error as expected
				continue
			}
			if isMap {
				t.Errorf("%s: expected error but parse succeeded", tt.desc)
			}
		}
	}
}

func TestJSONParseWrongArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`json.parse()`, "json.parse requires 1 argument (json_string), got=0"},
		{`json.parse("{}", "extra")`, "json.parse requires 1 argument (json_string), got=2"},

		// Wrong argument type
		{`json.parse(123)`, "json.parse requires string argument, got=INTEGER"},
		{`json.parse(true)`, "json.parse requires string argument, got=BOOLEAN"},
		{`json.parse([1, 2])`, "json.parse requires string argument, got=ARRAY"},
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

func TestJSONStringifyBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Stringify simple map
		{`{"name": "Alice"} > json.stringify()`, `{"name":"Alice"}`},

		// Stringify with number
		{`{"age": 30} > json.stringify()`, `{"age":30}`},

		// Stringify with boolean
		{`{"active": true} > json.stringify()`, `{"active":true}`},

		// Stringify array
		{`[1, 2, 3] > json.stringify()`, `[1,2,3]`},

		// Stringify nested structure
		{`{"user": {"name": "Bob"}} > json.stringify()`, `{"user":{"name":"Bob"}}`},

		// Stringify empty map
		{`{} > json.stringify()`, `{}`},

		// Stringify empty array
		{`[] > json.stringify()`, `[]`},

		// Stringify string
		{`"hello" > json.stringify()`, `"hello"`},

		// Stringify number
		{`42 > json.stringify()`, `42`},

		// Stringify boolean
		{`true > json.stringify()`, `true`},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestJSONStringifyErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`json.stringify()`, "json.stringify requires 1 argument (data), got=0"},
		{`json.stringify({}, {})`, "json.stringify requires 1 argument (data), got=2"},
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

func TestJSONStringifyPrettyBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Pretty print object
		{`{"name": "Alice", "age": 30} > json.stringify_pretty()`, "{\n  \"age\": 30,\n  \"name\": \"Alice\"\n}"},

		// Pretty print array
		{`[1, 2, 3] > json.stringify_pretty()`, "[\n  1,\n  2,\n  3\n]"},

		// Pretty print nested
		{
			`{"user": {"name": "Bob"}} > json.stringify_pretty()`,
			"{\n  \"user\": {\n    \"name\": \"Bob\"\n  }\n}",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestJSONStringifyPrettyErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`json.stringify_pretty()`, "json.stringify_pretty requires 1 argument (data), got=0"},
		{`json.stringify_pretty({}, {})`, "json.stringify_pretty requires 1 argument (data), got=2"},
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

func TestJSONParseDepthLimit(t *testing.T) {
	// Generate JSON with 101 levels of nesting (exceeds 100 limit)
	// Use escaped quotes for bark string literal
	deepJSON := ""
	for i := 0; i < 101; i++ {
		deepJSON += `{\"a\":`
	}
	deepJSON += "1"
	for i := 0; i < 101; i++ {
		deepJSON += "}"
	}

	input := `json.parse("` + deepJSON + `")`
	evaluated := testEval(input)

	// Should return tuple with error
	tpl, ok := evaluated.(*object.Tuple)
	if !ok {
		t.Fatalf("expected Tuple, got=%T (%+v)", evaluated, evaluated)
	}

	if len(tpl.Elements) != 2 {
		t.Fatalf("expected tuple with 2 elements, got=%d", len(tpl.Elements))
	}

	// First element should be Error
	errObj, ok := tpl.Elements[0].(*object.Error)
	if !ok {
		t.Fatalf("expected Error at index 0, got=%T", tpl.Elements[0])
	}

	if errObj.Msg != "JSON nesting depth exceeds 100 levels" {
		t.Errorf("wrong error message. expected=%q, got=%q", "JSON nesting depth exceeds 100 levels", errObj.Msg)
	}
}

func TestJSONParseAtDepthLimit(t *testing.T) {
	// Generate JSON with exactly 100 levels of nesting (at limit, should succeed)
	// Use escaped quotes for bark string literal
	deepJSON := ""
	for i := 0; i < 100; i++ {
		deepJSON += `{\"a\":`
	}
	deepJSON += "1"
	for i := 0; i < 100; i++ {
		deepJSON += "}"
	}

	input := `json.parse("` + deepJSON + `") > capture(e, r)
e > size()`
	evaluated := testEval(input)

	// Should succeed - error should be empty map with size 0
	num, ok := evaluated.(*object.Integer)
	if !ok {
		t.Fatalf("expected Integer, got=%T (%+v)", evaluated, evaluated)
	}

	if num.Value != 0 {
		t.Errorf("expected 0 (no error), got=%d", num.Value)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	tests := []struct {
		input   string
		desc    string
		checkFn func(object.Object) bool
	}{
		{
			`{"name": "Alice", "age": 30} > json.stringify() > json.parse() > capture(e, r)
r > get("name")`,
			"round trip object",
			func(obj object.Object) bool {
				str, ok := obj.(*object.String)
				return ok && str.Value == "Alice"
			},
		},
		{
			`[1, 2, 3] > json.stringify() > json.parse() > capture(e, r)
r > get(0)`,
			"round trip array",
			func(obj object.Object) bool {
				num, ok := obj.(*object.Integer)
				return ok && num.Value == 1
			},
		},
		{
			`"hello" > json.stringify() > json.parse() > capture(e, r)
r`,
			"round trip string",
			func(obj object.Object) bool {
				str, ok := obj.(*object.String)
				return ok && str.Value == "hello"
			},
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if !tt.checkFn(evaluated) {
			t.Errorf("%s failed. got=%T (%+v)", tt.desc, evaluated, evaluated)
		}
	}
}
