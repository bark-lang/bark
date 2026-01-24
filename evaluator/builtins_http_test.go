package evaluator

import (
	"testing"

	"gitlab.com/bark-lang/bark/object"
)

// HTTP module tests
// Note: These are basic error handling tests. Full integration tests would require a mock HTTP server.

func TestHTTPGetErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`http.get()`, "http.get requires 1 argument (url), got=0"},
		{`http.get("url", "extra")`, "http.get requires 1 argument (url), got=2"},

		// Wrong argument type
		{`http.get(123)`, "http.get requires string argument, got=INTEGER"},
		{`http.get(true)`, "http.get requires string argument, got=BOOLEAN"},
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

func TestHTTPPostErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`http.post()`, "http.post requires 2 arguments (url, body), got=0"},
		{`http.post("url")`, "http.post requires 2 arguments (url, body), got=1"},

		// Wrong argument types
		{`http.post(123, "body")`, "http.post requires string url, got=INTEGER"},
		{`http.post("url", 123)`, "http.post requires string body, got=INTEGER"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object for %q, got=%T", tt.input, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message for %q. expected=%q, got=%q", tt.input, tt.expected, errObj.Msg)
		}
	}
}

func TestHTTPPutErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`http.put()`, "http.put requires 2 arguments (url, body), got=0"},
		{`http.put("url")`, "http.put requires 2 arguments (url, body), got=1"},

		// Wrong argument types
		{`http.put(123, "body")`, "http.put requires string url, got=INTEGER"},
		{`http.put("url", 123)`, "http.put requires string body, got=INTEGER"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object for %q, got=%T", tt.input, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message for %q. expected=%q, got=%q", tt.input, tt.expected, errObj.Msg)
		}
	}
}

func TestHTTPDeleteErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`http.delete()`, "http.delete requires 1 argument (url), got=0"},
		{`http.delete("url", "extra")`, "http.delete requires 1 argument (url), got=2"},

		// Wrong argument type
		{`http.delete(123)`, "http.delete requires string argument, got=INTEGER"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object for %q, got=%T", tt.input, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message for %q. expected=%q, got=%q", tt.input, tt.expected, errObj.Msg)
		}
	}
}

func TestHTTPRequestErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Wrong number of arguments
		{`http.request()`, "http.request requires 4 arguments (method, url, headers, body), got=0"},
		{`http.request("GET", "url")`, "http.request requires 4 arguments (method, url, headers, body), got=2"},

		// Wrong argument types
		{`http.request(123, "url", {}, "")`, "http.request requires string method, got=INTEGER"},
		{`http.request("GET", 123, {}, "")`, "http.request requires string url, got=INTEGER"},
		{`http.request("GET", "url", "not-map", "")`, "http.request requires map headers, got=STRING"},
		{`http.request("GET", "url", {}, 123)`, "http.request requires string body, got=INTEGER"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object for %q, got=%T", tt.input, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message for %q. expected=%q, got=%q", tt.input, tt.expected, errObj.Msg)
		}
	}
}

// Note: Actual HTTP call tests would require a mock HTTP server
// These tests verify that the functions exist and have proper error handling
// Integration tests for actual HTTP calls should be done separately
