package evaluator

import (
	"os"
	"testing"

	"gitlab.com/bark-lang/bark/object"
)

func TestEnvGetBuiltin(t *testing.T) {
	// Set test environment variables
	_ = os.Setenv("TEST_VAR_1", "value1")
	_ = os.Setenv("TEST_VAR_2", "hello world")
	_ = os.Setenv("TEST_EMPTY", "")
	defer func() {
		_ = os.Unsetenv("TEST_VAR_1")
		_ = os.Unsetenv("TEST_VAR_2")
		_ = os.Unsetenv("TEST_EMPTY")
	}()

	tests := []struct {
		input    string
		expected string
	}{
		// Get existing variables
		{`env.get("TEST_VAR_1")`, "value1"},
		{`env.get("TEST_VAR_2")`, "hello world"},

		// Get non-existent variable returns empty string
		{`env.get("NONEXISTENT_VAR")`, ""},

		// Get variable set to empty string
		{`env.get("TEST_EMPTY")`, ""},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestEnvGetErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`env.get()`, "env.get requires 1 argument (key), got=0"},
		{`env.get("a", "b")`, "env.get requires 1 argument (key), got=2"},

		// Wrong argument type
		{`env.get(123)`, "env.get requires string argument, got=INTEGER"},
		{`env.get(true)`, "env.get requires string argument, got=BOOLEAN"},
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

func TestEnvGetOrBuiltin(t *testing.T) {
	// Set test environment variables
	_ = os.Setenv("TEST_GET_OR", "actual_value")
	_ = os.Setenv("TEST_EMPTY_GET_OR", "")
	defer func() {
		_ = os.Unsetenv("TEST_GET_OR")
		_ = os.Unsetenv("TEST_EMPTY_GET_OR")
	}()

	tests := []struct {
		input    string
		expected string
	}{
		// Get existing variable
		{`env.get_or("TEST_GET_OR", "default")`, "actual_value"},

		// Get non-existent variable returns default
		{`env.get_or("NONEXISTENT", "default_value")`, "default_value"},

		// Get variable set to empty string returns empty (not default)
		{`env.get_or("TEST_EMPTY_GET_OR", "default")`, ""},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestEnvGetOrErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`env.get_or()`, "env.get_or requires 2 arguments (key, default), got=0"},
		{`env.get_or("key")`, "env.get_or requires 2 arguments (key, default), got=1"},
		{`env.get_or("a", "b", "c")`, "env.get_or requires 2 arguments (key, default), got=3"},

		// Wrong argument types
		{`env.get_or(123, "default")`, "env.get_or requires string key, got=INTEGER"},
		{`env.get_or("key", 123)`, "env.get_or requires string default, got=INTEGER"},
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

func TestEnvPresentBuiltin(t *testing.T) {
	// Set test environment variables
	_ = os.Setenv("TEST_PRESENT", "value")
	_ = os.Setenv("TEST_PRESENT_EMPTY", "")
	defer func() {
		_ = os.Unsetenv("TEST_PRESENT")
		_ = os.Unsetenv("TEST_PRESENT_EMPTY")
	}()

	tests := []struct {
		input    string
		expected bool
	}{
		// Present variables
		{`env.present?("TEST_PRESENT")`, true},
		{`env.present?("TEST_PRESENT_EMPTY")`, true}, // Empty string still present

		// Absent variables
		{`env.present?("NONEXISTENT_VAR")`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestEnvPresentErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`env.present?()`, "env.present? requires 1 argument (key), got=0"},
		{`env.present?("a", "b")`, "env.present? requires 1 argument (key), got=2"},

		// Wrong argument type
		{`env.present?(123)`, "env.present? requires string argument, got=INTEGER"},
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

func TestEnvAbsentBuiltin(t *testing.T) {
	// Set test environment variables
	_ = os.Setenv("TEST_ABSENT", "value")
	_ = os.Setenv("TEST_ABSENT_EMPTY", "")
	defer func() {
		_ = os.Unsetenv("TEST_ABSENT")
		_ = os.Unsetenv("TEST_ABSENT_EMPTY")
	}()

	tests := []struct {
		input    string
		expected bool
	}{
		// Present variables are not absent
		{`env.absent?("TEST_ABSENT")`, false},
		{`env.absent?("TEST_ABSENT_EMPTY")`, false}, // Empty string still present

		// Absent variables
		{`env.absent?("NONEXISTENT_VAR")`, true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestEnvAbsentErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`env.absent?()`, "env.absent? requires 1 argument (key), got=0"},
		{`env.absent?("a", "b")`, "env.absent? requires 1 argument (key), got=2"},

		// Wrong argument type
		{`env.absent?(123)`, "env.absent? requires string argument, got=INTEGER"},
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

func TestEnvAllBuiltin(t *testing.T) {
	// Set some test variables
	_ = os.Setenv("TEST_ALL_1", "value1")
	_ = os.Setenv("TEST_ALL_2", "value2")
	defer func() {
		_ = os.Unsetenv("TEST_ALL_1")
		_ = os.Unsetenv("TEST_ALL_2")
	}()

	// Get all env vars
	evaluated := testEval(`env.all()`)

	mapObj, ok := evaluated.(*object.Map)
	if !ok {
		t.Fatalf("expected Map object, got=%T (%+v)", evaluated, evaluated)
	}

	// Check that our test vars are in the map
	if val, exists := mapObj.Pairs["TEST_ALL_1"]; !exists {
		t.Errorf("TEST_ALL_1 not found in env.all()")
	} else {
		strVal, ok := val.(*object.String)
		if !ok {
			t.Errorf("expected String value, got=%T", val)
		} else if strVal.Value != "value1" {
			t.Errorf("wrong value for TEST_ALL_1. expected=%q, got=%q", "value1", strVal.Value)
		}
	}

	if val, exists := mapObj.Pairs["TEST_ALL_2"]; !exists {
		t.Errorf("TEST_ALL_2 not found in env.all()")
	} else {
		strVal, ok := val.(*object.String)
		if !ok {
			t.Errorf("expected String value, got=%T", val)
		} else if strVal.Value != "value2" {
			t.Errorf("wrong value for TEST_ALL_2. expected=%q, got=%q", "value2", strVal.Value)
		}
	}

	// Check that map has keys (maintains insertion order)
	if len(mapObj.Keys) != len(mapObj.Pairs) {
		t.Errorf("keys length mismatch. keys=%d, pairs=%d", len(mapObj.Keys), len(mapObj.Pairs))
	}
}

func TestEnvAllErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`env.all("arg")`, "env.all requires 0 arguments, got=1"},
		{`env.all("a", "b")`, "env.all requires 0 arguments, got=2"},
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
