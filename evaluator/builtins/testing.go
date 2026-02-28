package builtins

import (
	"fmt"
	"os"

	"gitlab.com/bark-lang/barki/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/barki/object"
)

// testMode indicates whether we're running in test mode (bark test)
var testMode = false

// SetTestMode enables or disables test mode for the builtins package.
// When enabled, assert() and assert_error() builtins are available.
func SetTestMode(enabled bool) {
	testMode = enabled
}

// IsTestMode returns whether test mode is enabled.
func IsTestMode() bool {
	return testMode
}

// InitTesting initializes testing/assertion builtin functions
func InitTesting() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		// assert(expected) - compare piped value to expected value
		// On failure: prints assertion error to stderr
		// Returns: the input value (for chaining) or error on type mismatch
		"assert": {
			Fn: func(args ...object.Object) object.Object {
				if !testMode {
					return newError("assert() can only be used in test files (run with 'bark test')")
				}
				if len(args) != 2 {
					return newError("assert requires exactly 2 arguments (actual, expected)")
				}

				actual := args[0]
				expected := args[1]

				// Check if actual is an error (programming error) - this shouldn't be compared
				if err, ok := actual.(*object.Error); ok {
					if err.IsProgrammingError {
						_, _ = fmt.Fprintf(os.Stderr, "ASSERTION FAILED: unexpected error: %s\n", err.Msg)
						return actual
					}
				}

				// Compare values using ObjectsEqual
				if !helpers.ObjectsEqual(actual, expected) {
					_, _ = fmt.Fprintf(os.Stderr, "ASSERTION FAILED: expected %s, got %s\n",
						expected.Inspect(), actual.Inspect())
				}

				// Return the actual value to allow continued chaining
				return actual
			},
		},

		// assert_error(expected_msg) - compare piped error's msg field to expected message
		// On failure: prints assertion error to stderr
		// Returns: the input value (for chaining)
		"assert_error": {
			Fn: func(args ...object.Object) object.Object {
				if !testMode {
					return newError("assert_error() can only be used in test files (run with 'bark test')")
				}
				if len(args) != 2 {
					return newError("assert_error requires exactly 2 arguments (actual, expected_msg)")
				}

				actual := args[0]
				expectedMsg, ok := args[1].(*object.String)
				if !ok {
					return newError("assert_error: expected_msg must be a string, got %s", args[1].Type())
				}

				// Check if actual is an error
				err, isError := actual.(*object.Error)
				if !isError {
					_, _ = fmt.Fprintf(os.Stderr, "ASSERTION FAILED: expected error with msg %q, got %s: %s\n",
						expectedMsg.Value, actual.Type(), actual.Inspect())
					return actual
				}

				// Compare error messages
				if err.Msg != expectedMsg.Value {
					_, _ = fmt.Fprintf(os.Stderr, "ASSERTION FAILED: expected error msg %q, got %q\n",
						expectedMsg.Value, err.Msg)
				}

				// Return the actual value to allow continued chaining
				return actual
			},
		},
	}
}
