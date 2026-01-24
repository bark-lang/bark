package builtins

import (
	"gitlab.com/bark-lang/bark/object"
)

// InitNumbers initializes basic arithmetic operations
func InitNumbers() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"add": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				// Handle integers
				if left, ok := args[0].(*object.Integer); ok {
					if right, ok := args[1].(*object.Integer); ok {
						return &object.Integer{Value: left.Value + right.Value}
					}
				}

				// Handle floats
				if left, ok := args[0].(*object.Float); ok {
					if right, ok := args[1].(*object.Float); ok {
						return &object.Float{Value: left.Value + right.Value}
					}
				}

				return newError("type mismatch: %s + %s", args[0].Type(), args[1].Type())
			},
		},
		"sub": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				if left, ok := args[0].(*object.Integer); ok {
					if right, ok := args[1].(*object.Integer); ok {
						return &object.Integer{Value: left.Value - right.Value}
					}
				}

				if left, ok := args[0].(*object.Float); ok {
					if right, ok := args[1].(*object.Float); ok {
						return &object.Float{Value: left.Value - right.Value}
					}
				}

				return newError("type mismatch: %s - %s", args[0].Type(), args[1].Type())
			},
		},
		"mul": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				if left, ok := args[0].(*object.Integer); ok {
					if right, ok := args[1].(*object.Integer); ok {
						return &object.Integer{Value: left.Value * right.Value}
					}
				}

				if left, ok := args[0].(*object.Float); ok {
					if right, ok := args[1].(*object.Float); ok {
						return &object.Float{Value: left.Value * right.Value}
					}
				}

				return newError("type mismatch: %s * %s", args[0].Type(), args[1].Type())
			},
		},
		"div": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				if left, ok := args[0].(*object.Integer); ok {
					if right, ok := args[1].(*object.Integer); ok {
						if right.Value == 0 {
							return newError("division by zero")
						}
						return &object.Integer{Value: left.Value / right.Value}
					}
				}

				if left, ok := args[0].(*object.Float); ok {
					if right, ok := args[1].(*object.Float); ok {
						if right.Value == 0.0 {
							return newError("division by zero")
						}
						return &object.Float{Value: left.Value / right.Value}
					}
				}

				return newError("type mismatch: %s / %s", args[0].Type(), args[1].Type())
			},
		},
	}
}
