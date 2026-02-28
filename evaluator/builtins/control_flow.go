package builtins

import (
	"gitlab.com/bark-lang/barki/object"
)

// InitControlFlow initializes control flow operations
func InitControlFlow() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		// return?() - conditional early exit from function
		"return?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) == 0 {
					return newError("wrong number of arguments. got=0, want at least 1")
				}

				// First argument must be a boolean condition
				condition, ok := args[0].(*object.Boolean)
				if !ok {
					return newError("first argument to `return?` must be BOOLEAN, got %s", args[0].Type())
				}

				// If condition is false, return NULL (no early return)
				if !condition.Value {
					return NULL
				}

				// Condition is true, return early
				if len(args) == 1 {
					// No value provided, return zero value
					return &object.ReturnValue{Value: NULL}
				} else if len(args) == 2 {
					// Single return value
					return &object.ReturnValue{Value: args[1]}
				} else {
					// Multiple return values - wrap in tuple
					return &object.ReturnValue{Value: &object.Tuple{Elements: args[1:]}}
				}
			},
		},

		// continue?() - conditional chain continuation
		// Continues if true, stops chain if false
		"continue?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}

				condition, ok := args[0].(*object.Boolean)
				if !ok {
					return newError("argument to `continue?` must be BOOLEAN, got %s", args[0].Type())
				}

				if condition.Value {
					// Continue processing - return Void to allow chain to continue
					// without passing a value to the next function
					return &object.Void{}
				}
				// Stop chain - return ChainStop to halt chain processing
				return &object.ChainStop{}
			},
		},

		// repeat?() - conditional recursive call to current anonymous function
		// Note: This requires special handling in the evaluator to access current function context
		"repeat?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) == 0 {
					return newError("wrong number of arguments. got=0, want at least 1")
				}

				// First argument must be a boolean condition
				condition, ok := args[0].(*object.Boolean)
				if !ok {
					return newError("first argument to `repeat?` must be BOOLEAN, got %s", args[0].Type())
				}

				if !condition.Value {
					// Condition is false, don't repeat
					return NULL
				}

				// Condition is true, repeat with remaining args
				if len(args) == 1 {
					// No new args provided, use current args
					return &object.RepeatValue{Args: nil}
				}
				// Use provided args (skip the condition)
				return &object.RepeatValue{Args: args[1:]}
			},
		},

		// repeat() - unconditional recursive call to current anonymous function
		"repeat": {
			Fn: func(args ...object.Object) object.Object {
				// repeat() unconditionally repeats with the provided args
				// If no args provided, use current args (for pass-by-reference types)
				if len(args) == 0 {
					return &object.RepeatValue{Args: nil}
				}
				return &object.RepeatValue{Args: args}
			},
		},

		// break?() - conditional early exit with values (alias for return? in most contexts)
		"break?": {
			Fn: func(args ...object.Object) object.Object {
				// break?() is essentially the same as return?() for now
				// It's used to exit from recursive anonymous functions
				if len(args) == 0 {
					return newError("wrong number of arguments. got=0, want at least 1")
				}

				condition, ok := args[0].(*object.Boolean)
				if !ok {
					return newError("first argument to `break?` must be BOOLEAN, got %s", args[0].Type())
				}

				if !condition.Value {
					return NULL
				}

				if len(args) == 1 {
					return &object.ReturnValue{Value: NULL}
				} else if len(args) == 2 {
					return &object.ReturnValue{Value: args[1]}
				} else {
					return &object.ReturnValue{Value: &object.Tuple{Elements: args[1:]}}
				}
			},
		},
	}
}
