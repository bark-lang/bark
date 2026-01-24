package evaluator

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"gitlab.com/bark-lang/bark/object"
)

func TestAssertBuiltin(t *testing.T) {
	// Enable test mode for assert builtins
	SetTestMode(true)
	defer SetTestMode(false)

	tests := []struct {
		input          string
		expectedResult interface{}
		expectError    bool
		expectStderr   string // substring expected in stderr output
	}{
		// Basic equality tests - pass silently
		{`1 > assert(1)`, 1, false, ""},
		{`"hello" > assert("hello")`, "hello", false, ""},
		{`true > assert(true)`, true, false, ""},
		{`false > assert(false)`, false, false, ""},

		// Computed value comparison
		{`1 > add(1) > assert(2)`, 2, false, ""},

		// Array comparison
		{`[1, 2, 3] > assert([1, 2, 3])`, nil, false, ""}, // arrays compare by Inspect()

		// Map comparison
		{`{"a": 1} > assert({"a": 1})`, nil, false, ""}, // maps compare by Inspect()

		// Assertion failures - print to stderr but continue
		{`1 > assert(2)`, 1, false, "ASSERTION FAILED: expected 2, got 1"},
		{`"hello" > assert("world")`, "hello", false, "ASSERTION FAILED: expected world, got hello"},
		{`true > assert(false)`, true, false, "ASSERTION FAILED: expected false, got true"},

		// Wrong argument count
		{`assert(1)`, nil, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			evaluated := testEval(tt.input)

			// Restore stderr and read captured output
			_ = w.Close()
			os.Stderr = oldStderr
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			stderrOutput := buf.String()

			if tt.expectError {
				if !isError(evaluated) {
					t.Errorf("expected error but got %T (%+v)", evaluated, evaluated)
				}
				return
			}

			if tt.expectStderr != "" {
				if !strings.Contains(stderrOutput, tt.expectStderr) {
					t.Errorf("expected stderr to contain %q, got %q", tt.expectStderr, stderrOutput)
				}
			}

			if tt.expectedResult != nil {
				if !testObject(t, evaluated, tt.expectedResult) {
					t.Errorf("object mismatch for %s", tt.input)
				}
			}
		})
	}
}

func TestAssertErrorBuiltin(t *testing.T) {
	// Enable test mode for assert builtins
	SetTestMode(true)
	defer SetTestMode(false)

	tests := []struct {
		input          string
		expectStderr   string // substring expected in stderr output
		expectPassThru bool   // whether the error should be passed through
	}{
		// Error with matching message - pass silently
		{`err("invalid input") > assert_error("invalid input")`, "", true},

		// Error with different message - print to stderr
		{
			`err("wrong message") > assert_error("expected message")`,
			`ASSERTION FAILED: expected error msg "expected message", got "wrong message"`,
			true,
		},

		// Non-error value - print to stderr
		{
			`42 > assert_error("some error")`,
			`ASSERTION FAILED: expected error with msg "some error", got INTEGER: 42`,
			false,
		},

		// String value - print to stderr
		{
			`"hello" > assert_error("some error")`,
			`ASSERTION FAILED: expected error with msg "some error", got STRING: hello`,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			evaluated := testEval(tt.input)

			// Restore stderr and read captured output
			_ = w.Close()
			os.Stderr = oldStderr
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			stderrOutput := buf.String()

			if tt.expectStderr != "" {
				if !strings.Contains(stderrOutput, tt.expectStderr) {
					t.Errorf("expected stderr to contain %q, got %q", tt.expectStderr, stderrOutput)
				}
			} else if stderrOutput != "" {
				t.Errorf("expected no stderr output but got %q", stderrOutput)
			}

			// Verify the value is passed through (not an internal error)
			if isError(evaluated) {
				// Only programming errors should cause test failure here
				if err, ok := evaluated.(*object.Error); ok && err.IsProgrammingError {
					if tt.expectPassThru {
						// Programming error when we expected pass-through
						t.Errorf("got programming error: %s", err.Msg)
					}
				}
			}
		})
	}
}

// Helper to test object values
func testObject(t *testing.T, obj object.Object, expected interface{}) bool {
	switch exp := expected.(type) {
	case int:
		return testIntegerObject(t, obj, int64(exp))
	case int64:
		return testIntegerObject(t, obj, exp)
	case string:
		result, ok := obj.(*object.String)
		if !ok {
			t.Errorf("object is not String. got=%T (%+v)", obj, obj)
			return false
		}
		if result.Value != exp {
			t.Errorf("object has wrong value. got=%q, want=%q", result.Value, exp)
			return false
		}
		return true
	case bool:
		return testBooleanObject(t, obj, exp)
	default:
		t.Errorf("unhandled expected type %T", expected)
		return false
	}
}
