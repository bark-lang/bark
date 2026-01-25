package evaluator

import (
	"testing"

	"gitlab.com/bark-lang/bark/object"
)

// SQL Injection Prevention Tests

func TestSecuritySQLEscapeBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic escaping
		{`security.sql_escape("O'Reilly")`, "O''Reilly"},
		{`security.sql_escape("It's working")`, "It''s working"},

		// Multiple quotes
		{`security.sql_escape("''; DROP TABLE users; --")`, "''''; DROP TABLE users; --"},

		// Normal strings
		{`security.sql_escape("normal text")`, "normal text"},
		{`security.sql_escape("")`, ""},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestSecuritySQLEscapeErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`security.sql_escape()`, "security.sql_escape requires 1 argument (input), got=0"},
		{`security.sql_escape("a", "b")`, "security.sql_escape requires 1 argument (input), got=2"},
		{`security.sql_escape(123)`, "security.sql_escape requires string argument, got=INTEGER"},
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

// Shell Escape Tests

func TestSecurityShellEscapeBuiltin(t *testing.T) {
	tests := []struct {
		input      string
		shouldHave string
		desc       string
	}{
		{`security.shell_escape("test;rm -rf /")`, "\\;", "escapes semicolon"},
		{`security.shell_escape("test && malicious")`, "\\&", "escapes ampersand"},
		{`security.shell_escape("test | grep")`, "\\|", "escapes pipe"},
		{`security.shell_escape("test $(whoami)")`, "\\$", "escapes dollar"},
		{`security.shell_escape("test")`, "test", "leaves safe string unchanged"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		str, ok := evaluated.(*object.String)
		if !ok {
			t.Errorf("%s: expected String object, got=%T", tt.desc, evaluated)
			continue
		}
		// Just check that escaped characters are present when expected
		if tt.shouldHave != "" && !containsSubstring(str.Value, tt.shouldHave) {
			t.Errorf("%s: expected result to contain %q, got=%q", tt.desc, tt.shouldHave, str.Value)
		}
	}
}

func containsSubstring(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && hasSubstring(s, substr)
}

func hasSubstring(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Safe Command Tests

func TestSecurityIsSafeCommandBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Safe commands
		{`security.safe_command?("ls -la")`, true},
		{`security.safe_command?("cat file.txt")`, true},
		{`security.safe_command?("echo hello")`, true},

		// Dangerous commands
		{`security.safe_command?("rm -rf /")`, false},
		{`security.safe_command?("curl malicious.com")`, false},
		{`security.safe_command?("wget hack.sh")`, false},
		{`security.safe_command?("shutdown now")`, false},

		// Command chaining
		{`security.safe_command?("ls; rm -rf /")`, false},
		{`security.safe_command?("echo test && whoami")`, false},
		{`security.safe_command?("cat file | grep secret")`, false},
		{`security.safe_command?("echo $(whoami)")`, false},

		// Empty/whitespace
		{`security.safe_command?("")`, true},
		{`security.safe_command?("   ")`, true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

// Path Sanitization Tests

func TestSecuritySanitizePathBuiltin(t *testing.T) {
	// Test basic path cleaning - use capture() to extract result from tuple
	input := `security.sanitize_path("./test/../file.txt") > capture(e, r)
r`
	evaluated := testEval(input)

	str, ok := evaluated.(*object.String)
	if !ok {
		t.Errorf("expected String object, got=%T", evaluated)
		return
	}

	// Should clean the path (exact result depends on filesystem)
	if str.Value == "" {
		t.Errorf("expected non-empty sanitized path")
	}
}

func TestSecuritySanitizePathSuccess(t *testing.T) {
	// Test that successful sanitization returns empty map as error
	// After capture(e, r), e is bound to error and r to result
	input := `security.sanitize_path("test.txt") > capture(e, r)
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

// Input Validation Tests

func TestSecurityIsEmailBuiltin(t *testing.T) {
	// Valid email
	evaluated := testEval(`security.email?("test@example.com")`)
	testBooleanObject(t, evaluated, true)

	// Invalid emails
	evaluated = testEval(`security.email?("not_an_email")`)
	testBooleanObject(t, evaluated, false)

	evaluated = testEval(`security.email?("")`)
	testBooleanObject(t, evaluated, false)
}

func TestSecurityIsURLBuiltin(t *testing.T) {
	// Valid URLs
	evaluated := testEval(`security.url?("https://example.com")`)
	testBooleanObject(t, evaluated, true)

	// Invalid URLs
	evaluated = testEval(`security.url?("not_a_url")`)
	testBooleanObject(t, evaluated, false)
}

// XSS Prevention Tests

func TestSecurityHTMLEscapeBuiltin(t *testing.T) {
	// Test escaping less than/greater than
	evaluated := testEval(`security.html_escape("<script>")`)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("expected String, got=%T", evaluated)
	}
	// Should contain escaped < and >
	if !containsSubstring(str.Value, "&lt;") || !containsSubstring(str.Value, "&gt;") {
		t.Errorf("expected escaped HTML, got=%s", str.Value)
	}

	// Normal text unchanged
	evaluated = testEval(`security.html_escape("normal")`)
	testStringObject(t, evaluated, "normal")
}

func TestSecurityStripTagsBuiltin(t *testing.T) {
	// Strip HTML tags
	evaluated := testEval(`security.strip_tags("<p>Hello</p>")`)
	testStringObject(t, evaluated, "Hello")

	// No tags - unchanged
	evaluated = testEval(`security.strip_tags("plain")`)
	testStringObject(t, evaluated, "plain")
}

// Nonce and Key Generation Tests

func TestSecurityGenerateNonceBuiltin(t *testing.T) {
	// Generate nonce
	evaluated := testEval(`security.generate_nonce()`)

	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("expected String object, got=%T", evaluated)
	}

	// Nonce should be non-empty
	if len(str.Value) == 0 {
		t.Errorf("expected non-empty nonce")
	}
}

func TestSecurityHashKeyBuiltin(t *testing.T) {
	// Hash a key - returns integer hash
	evaluated := testEval(`security.hash_key("testkey")`)

	num, ok := evaluated.(*object.Integer)
	if !ok {
		t.Fatalf("expected Integer object, got=%T", evaluated)
	}

	// Hash should be non-zero
	if num.Value == 0 {
		t.Errorf("expected non-zero hash")
	}
}

// Error Handling Tests

func TestSecurityWrongArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// shell_escape errors
		{`security.shell_escape()`, "security.shell_escape requires 1 argument (input), got=0"},
		{`security.shell_escape(123)`, "security.shell_escape requires string argument, got=INTEGER"},

		// safe_command? errors
		{`security.safe_command?()`, "security.safe_command? requires 1 argument (command), got=0"},
		{`security.safe_command?(123)`, "security.safe_command? requires string argument, got=INTEGER"},

		// email? errors
		{`security.email?()`, "security.email? requires 1 argument (email), got=0"},

		// url? errors
		{`security.url?()`, "security.url? requires 1 argument (url), got=0"},

		// html_escape errors
		{`security.html_escape()`, "security.html_escape requires 1 argument (html), got=0"},

		// strip_tags errors
		{`security.strip_tags()`, "security.strip_tags requires 1 argument (html), got=0"},

		// generate_nonce errors
		{`security.generate_nonce(123)`, "security.generate_nonce requires 0 arguments, got=1"},

		// hash_key errors
		{`security.hash_key()`, "security.hash_key requires 1 argument (input), got=0"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object for %q, got=%T (%+v)", tt.input, evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message for %q. expected=%q, got=%q", tt.input, tt.expected, errObj.Msg)
		}
	}
}
