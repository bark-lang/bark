package evaluator

import (
	builtinspkg "gitlab.com/bark-lang/barki/evaluator/builtins"
	"gitlab.com/bark-lang/barki/object"
)

// builtins contains all builtin functions
var builtins map[string]*object.Builtin

func init() {
	// Get all builtins from the builtins package
	builtins = builtinspkg.GetAll()

	// Add parallel processing builtins (these need access to applyFunction)
	for name, fn := range InitParallelBuiltins() {
		builtins[name] = fn
	}

	// Add iteration builtins (these need access to applyFunction)
	for name, fn := range InitIterationBuiltins() {
		builtins[name] = fn
	}

	// Add functional builtins (these need access to applyFunction)
	for name, fn := range InitFunctionalBuiltins() {
		builtins[name] = fn
	}
}

// SetTestMode enables or disables test mode.
// When enabled, assert() and assert_error() builtins are available.
func SetTestMode(enabled bool) {
	builtinspkg.SetTestMode(enabled)
}

// IsTestMode returns whether test mode is enabled.
func IsTestMode() bool {
	return builtinspkg.IsTestMode()
}
