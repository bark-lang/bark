package evaluator

import (
	"gitlab.com/bark-lang/barki/object"
)

// InitIterationBuiltins creates iteration builtins with access to applyFunction
func InitIterationBuiltins() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"each": {Fn: eachFunc},
	}
}

// each iterates over a collection, calling a function for each element
// For arrays: calls fn(array, index) for each index
// For maps: calls fn(map, key) for each key
// Returns the original collection for chaining
func eachFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("each() requires 2 arguments: collection and function")
	}

	// Get function (second argument)
	fn := args[1]
	if !isCallable(fn) {
		return newError("each() second argument must be function, got %s", args[1].Type())
	}

	// Handle arrays
	if arr, ok := args[0].(*object.Array); ok {
		for i := range arr.Elements {
			// Call fn(array, index)
			index := &object.Integer{Value: int64(i)}
			result := applyFunction(fn, []object.Object{arr, index})

			// Check for errors
			if isError(result) {
				return result
			}
			// Execution errors stop the iteration but don't crash
			if isExecutionError(result) {
				return result
			}
		}
		// Return original collection for chaining
		return arr
	}

	// Handle maps
	if m, ok := args[0].(*object.Map); ok {
		for _, key := range m.Keys {
			// Call fn(map, key)
			keyObj := &object.String{Value: key}
			result := applyFunction(fn, []object.Object{m, keyObj})

			// Check for errors
			if isError(result) {
				return result
			}
			// Execution errors stop the iteration but don't crash
			if isExecutionError(result) {
				return result
			}
		}
		// Return original collection for chaining
		return m
	}

	return newError("each() first argument must be array or map, got %s", args[0].Type())
}
