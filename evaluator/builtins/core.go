package builtins

import (
	"fmt"
	"os"

	"gitlab.com/bark-lang/bark/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/bark/object"
)

// printFormatted handles printing with optional format support.
// Single arg: prints the value directly.
// Multiple args: first arg is format string with {0}, {1} placeholders.
// Returns (output string, error object) where error is nil on success.
func printFormatted(args []object.Object) (string, object.Object) {
	if len(args) == 0 {
		return "", nil
	}

	if len(args) == 1 {
		// Single argument - print directly
		return args[0].Inspect(), nil
	}

	// Multiple arguments - first is format string
	formatStr, ok := args[0].(*object.String)
	if !ok {
		// Provide helpful hint for common tuple mistake
		hint := fmt.Sprintf(
			"received %d arguments where first must be a format string, got %s instead",
			len(args), args[0].Type(),
		)
		if args[0].Type() == object.ARRAY_OBJ || args[0].Type() == object.MAP_OBJ {
			hint += ". Hint: if passing a tuple result, destructure it first: tuple > (a, b) { println(\"{a}, {b}\") }()"
		}
		return "", helpers.NewExecutionError("print format error", hint)
	}

	result, err := helpers.FormatString(formatStr.Value, args[1:])
	if err != nil {
		return "", helpers.NewExecutionError("print format error", err.Error())
	}
	return result, nil
}

// InitCore initializes core builtin functions (I/O, type conversion, error handling, control flow)
func InitCore() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		// I/O functions
		"print": {
			Fn: func(args ...object.Object) object.Object {
				output, err := printFormatted(args)
				if err != nil {
					return err
				}
				fmt.Print(output)
				return NULL
			},
		},
		"println": {
			Fn: func(args ...object.Object) object.Object {
				output, err := printFormatted(args)
				if err != nil {
					return err
				}
				fmt.Println(output)
				return NULL
			},
		},
		"eprint": {
			Fn: func(args ...object.Object) object.Object {
				output, err := printFormatted(args)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprint(os.Stderr, output)
				return NULL
			},
		},
		"eprintln": {
			Fn: func(args ...object.Object) object.Object {
				output, err := printFormatted(args)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(os.Stderr, output)
				return NULL
			},
		},
		// Conditional print functions - only print when first arg (bool) is true
		"print?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 1 {
					return newError("print? requires at least 1 argument (condition)")
				}
				cond, ok := args[0].(*object.Boolean)
				if !ok {
					return newError("print? first argument must be boolean, got=%s", args[0].Type())
				}
				if !cond.Value {
					return NULL
				}
				output, err := printFormatted(args[1:])
				if err != nil {
					return err
				}
				fmt.Print(output)
				return NULL
			},
		},
		"println?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 1 {
					return newError("println? requires at least 1 argument (condition)")
				}
				cond, ok := args[0].(*object.Boolean)
				if !ok {
					return newError("println? first argument must be boolean, got=%s", args[0].Type())
				}
				if !cond.Value {
					return NULL
				}
				output, err := printFormatted(args[1:])
				if err != nil {
					return err
				}
				fmt.Println(output)
				return NULL
			},
		},
		"eprint?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 1 {
					return newError("eprint? requires at least 1 argument (condition)")
				}
				cond, ok := args[0].(*object.Boolean)
				if !ok {
					return newError("eprint? first argument must be boolean, got=%s", args[0].Type())
				}
				if !cond.Value {
					return NULL
				}
				output, err := printFormatted(args[1:])
				if err != nil {
					return err
				}
				_, _ = fmt.Fprint(os.Stderr, output)
				return NULL
			},
		},
		"eprintln?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 1 {
					return newError("eprintln? requires at least 1 argument (condition)")
				}
				cond, ok := args[0].(*object.Boolean)
				if !ok {
					return newError("eprintln? first argument must be boolean, got=%s", args[0].Type())
				}
				if !cond.Value {
					return NULL
				}
				output, err := printFormatted(args[1:])
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(os.Stderr, output)
				return NULL
			},
		},

		// Type conversion
		"to_string": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}

				return &object.String{Value: args[0].Inspect()}
			},
		},

		// Control flow
		"return": {
			Fn: func(args ...object.Object) object.Object {
				// Wrap in ReturnValue
				if len(args) == 0 {
					return &object.ReturnValue{Value: NULL}
				} else if len(args) == 1 {
					return &object.ReturnValue{Value: args[0]}
				} else {
					// Multiple return values - wrap in tuple
					return &object.ReturnValue{Value: &object.Tuple{Elements: args}}
				}
			},
		},

		// Error handling
		"err": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) == 0 {
					return newError("empty error")
				}

				if len(args) == 1 {
					if msg, ok := args[0].(*object.String); ok {
						return &object.Error{Msg: msg.Value, Context: make(map[string]object.Object)}
					}
					return newError("error message must be string")
				}

				// Two arguments: message and context map
				if msg, ok := args[0].(*object.String); ok {
					if ctx, ok := args[1].(*object.Map); ok {
						return &object.Error{Msg: msg.Value, Context: ctx.Pairs}
					}
					return newError("error context must be map")
				}

				return newError("error message must be string")
			},
		},

		"err_msg": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("err_msg requires exactly 1 argument")
				}

				if err, ok := args[0].(*object.Error); ok {
					return &object.String{Value: err.Msg}
				}

				return newError("err_msg requires error argument")
			},
		},

		"err_context": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("err_context requires exactly 1 argument")
				}

				if err, ok := args[0].(*object.Error); ok {
					keys := make([]string, 0, len(err.Context))
					for k := range err.Context {
						keys = append(keys, k)
					}
					return &object.Map{Pairs: err.Context, Keys: keys}
				}

				return newError("err_context requires error argument")
			},
		},

		"err_add_context": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return newError("err_add_context requires exactly 3 arguments (error, key, value)")
				}

				if err, ok := args[0].(*object.Error); ok {
					if key, ok := args[1].(*object.String); ok {
						// Create a new context map with the existing entries plus the new one
						newContext := make(map[string]object.Object)
						for k, v := range err.Context {
							newContext[k] = v
						}
						newContext[key.Value] = args[2]

						// Return a new error with the updated context
						return &object.Error{
							Msg:     err.Msg,
							Context: newContext,
						}
					}
					return newError("err_add_context key must be string")
				}

				return newError("err_add_context requires error as first argument")
			},
		},
	}
}
