package evaluator

import (
	"testing"
	"time"

	"gitlab.com/bark-lang/bark/object"
)

func TestTimeNowBuiltin(t *testing.T) {
	// Get current time before and after calling time.now
	before := time.Now().UTC().Unix()
	evaluated := testEval(`time.now()`)
	after := time.Now().UTC().Unix()

	num, ok := evaluated.(*object.Integer)
	if !ok {
		t.Fatalf("expected Integer object, got=%T (%+v)", evaluated, evaluated)
	}

	// Timestamp should be between before and after
	if num.Value < before || num.Value > after {
		t.Errorf("timestamp out of range. expected between %d and %d, got=%d", before, after, num.Value)
	}
}

func TestTimeNowErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`time.now(123)`, "time.now requires 0 arguments, got=1"},
		{`time.now("arg")`, "time.now requires 0 arguments, got=1"},
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

func TestTimeNowMsBuiltin(t *testing.T) {
	// Get current time in ms before and after calling time.now_ms
	before := time.Now().UTC().UnixMilli()
	evaluated := testEval(`time.now_ms()`)
	after := time.Now().UTC().UnixMilli()

	num, ok := evaluated.(*object.Integer)
	if !ok {
		t.Fatalf("expected Integer object, got=%T (%+v)", evaluated, evaluated)
	}

	// Timestamp should be between before and after
	if num.Value < before || num.Value > after {
		t.Errorf("timestamp out of range. expected between %d and %d, got=%d", before, after, num.Value)
	}

	// Should be much larger than Unix timestamp (seconds)
	if num.Value < 1000000000000 {
		t.Errorf("expected millisecond timestamp (>1e12), got=%d", num.Value)
	}
}

func TestTimeNowMsErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`time.now_ms(123)`, "time.now_ms requires 0 arguments, got=1"},
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

func TestTimeFormatBuiltin(t *testing.T) {
	// Known timestamp: 2023-01-15 13:10:45 UTC = 1673788245
	tests := []struct {
		input    string
		expected string
	}{
		// Year formats
		{`1673788245 > time.format("%Y")`, "2023"},
		{`1673788245 > time.format("%y")`, "23"},

		// Month formats
		{`1673788245 > time.format("%m")`, "01"},
		{`1673788245 > time.format("%B")`, "January"},
		{`1673788245 > time.format("%b")`, "Jan"},

		// Day formats
		{`1673788245 > time.format("%d")`, "15"},

		// Time formats
		{`1673788245 > time.format("%H:%M:%S")`, "13:10:45"},
		{`1673788245 > time.format("%H")`, "13"},
		{`1673788245 > time.format("%M")`, "10"},
		{`1673788245 > time.format("%S")`, "45"},

		// Combined formats
		{`1673788245 > time.format("%Y-%m-%d")`, "2023-01-15"},
		{`1673788245 > time.format("%F")`, "2023-01-15"}, // ISO 8601 date shortcut
		{`1673788245 > time.format("%T")`, "13:10:45"},   // ISO 8601 time shortcut

		// Custom format
		{`1673788245 > time.format("%Y/%m/%d %H:%M")`, "2023/01/15 13:10"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestTimeFormatErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`time.format()`, "time.format requires 2 arguments (timestamp, format), got=0"},
		{`time.format(123)`, "time.format requires 2 arguments (timestamp, format), got=1"},
		{`time.format(1, "fmt", "extra")`, "time.format requires 2 arguments (timestamp, format), got=3"},

		// Wrong argument types
		{`time.format("not int", "%Y")`, "time.format requires integer timestamp, got=STRING"},
		{`time.format(123, 456)`, "time.format requires string format, got=INTEGER"},
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

func TestTimeParseBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		// Parse ISO 8601 date - use capture() to extract result from tuple
		{`time.parse("2023-01-15", "%Y-%m-%d") > capture(e, r)
r`, 1673740800}, // 2023-01-15 00:00:00 UTC

		// Parse with time
		{`time.parse("2023-01-15 13:10:45", "%Y-%m-%d %H:%M:%S") > capture(e, r)
r`, 1673788245},

		// Parse with shortcut
		{`time.parse("2023-01-15", "%F") > capture(e, r)
r`, 1673740800},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestTimeParseSuccess(t *testing.T) {
	// Test that successful parse returns empty map as error
	// After capture(e, r), e is bound to error and r to result
	input := `time.parse("2023-01-15", "%Y-%m-%d") > capture(e, r)
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

func TestTimeParseErrors(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
		desc        string
	}{
		// Invalid format
		{`time.parse("2023-01-15", "%Y/%m/%d")`, true, "wrong format separators"},
		{`time.parse("not a date", "%Y-%m-%d")`, true, "invalid date string"},
		{`time.parse("2023-13-45", "%Y-%m-%d")`, true, "invalid month/day"},
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

func TestTimeParseWrongArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`time.parse()`, "time.parse requires 2 arguments (time_string, format), got=0"},
		{`time.parse("date")`, "time.parse requires 2 arguments (time_string, format), got=1"},

		// Wrong argument types
		{`time.parse(123, "%Y")`, "time.parse requires string time_string, got=INTEGER"},
		{`time.parse("2023-01-15", 123)`, "time.parse requires string format, got=INTEGER"},
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

func TestTimeFormatISO8601Builtin(t *testing.T) {
	// Known timestamp: 2023-01-15 13:10:45 UTC = 1673788245
	input := `1673788245 > time.format_iso8601()`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "2023-01-15T13:10:45Z")
}

func TestTimeFormatISO8601Errors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`time.format_iso8601()`, "time.format_iso8601 requires 1 argument (timestamp), got=0"},
		{`time.format_iso8601(1, 2)`, "time.format_iso8601 requires 1 argument (timestamp), got=2"},

		// Wrong argument type
		{`time.format_iso8601("not int")`, "time.format_iso8601 requires integer timestamp, got=STRING"},
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

func TestTimeParseISO8601Builtin(t *testing.T) {
	// Use capture() to extract result from tuple
	input := `time.parse_iso8601("2023-01-15T13:10:45Z") > capture(e, r)
r`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 1673788245)
}

func TestTimeParseISO8601Success(t *testing.T) {
	// Test that successful parse returns empty map as error
	// After capture(e, r), e is bound to error and r to result
	input := `time.parse_iso8601("2023-01-15T13:10:45Z") > capture(e, r)
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

func TestTimeParseISO8601Errors(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
		desc        string
	}{
		// Invalid ISO 8601 format
		{`time.parse_iso8601("2023-01-15")`, true, "missing time component"},
		{`time.parse_iso8601("not a date")`, true, "invalid format"},
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

func TestTimeParseISO8601WrongArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`time.parse_iso8601()`, "time.parse_iso8601 requires 1 argument (time_string), got=0"},
		{`time.parse_iso8601("a", "b")`, "time.parse_iso8601 requires 1 argument (time_string), got=2"},

		// Wrong argument type
		{`time.parse_iso8601(123)`, "time.parse_iso8601 requires string argument, got=INTEGER"},
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

func TestTimeRoundTrip(t *testing.T) {
	// Test format -> parse round trip - use capture() to extract result from tuple
	input := `1673788245 > time.format("%Y-%m-%d %H:%M:%S") > time.parse("%Y-%m-%d %H:%M:%S") > capture(e, r)
r`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 1673788245)

	// Test ISO 8601 round trip
	input2 := `1673788245 > time.format_iso8601() > time.parse_iso8601() > capture(e, r)
r`
	evaluated2 := testEval(input2)
	testIntegerObject(t, evaluated2, 1673788245)
}
