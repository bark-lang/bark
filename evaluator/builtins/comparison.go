package builtins

import (
	"gitlab.com/bark-lang/bark/object"
)

// InitComparison initializes comparison and boolean operations
func InitComparison() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"eq?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				// Compare integers
				if left, ok := args[0].(*object.Integer); ok {
					if right, ok := args[1].(*object.Integer); ok {
						return nativeBoolToBooleanObject(left.Value == right.Value)
					}
				}

				// Compare floats
				if left, ok := args[0].(*object.Float); ok {
					if right, ok := args[1].(*object.Float); ok {
						return nativeBoolToBooleanObject(left.Value == right.Value)
					}
				}

				// Compare strings
				if left, ok := args[0].(*object.String); ok {
					if right, ok := args[1].(*object.String); ok {
						return nativeBoolToBooleanObject(left.Value == right.Value)
					}
				}

				// Compare booleans
				if left, ok := args[0].(*object.Boolean); ok {
					if right, ok := args[1].(*object.Boolean); ok {
						return nativeBoolToBooleanObject(left.Value == right.Value)
					}
				}

				return FALSE
			},
		},

		"gt?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				if left, ok := args[0].(*object.Integer); ok {
					if right, ok := args[1].(*object.Integer); ok {
						return nativeBoolToBooleanObject(left.Value > right.Value)
					}
				}

				if left, ok := args[0].(*object.Float); ok {
					if right, ok := args[1].(*object.Float); ok {
						return nativeBoolToBooleanObject(left.Value > right.Value)
					}
				}

				return newError("type mismatch: %s > %s", args[0].Type(), args[1].Type())
			},
		},

		"lt?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				if left, ok := args[0].(*object.Integer); ok {
					if right, ok := args[1].(*object.Integer); ok {
						return nativeBoolToBooleanObject(left.Value < right.Value)
					}
				}

				if left, ok := args[0].(*object.Float); ok {
					if right, ok := args[1].(*object.Float); ok {
						return nativeBoolToBooleanObject(left.Value < right.Value)
					}
				}

				return newError("type mismatch: %s < %s", args[0].Type(), args[1].Type())
			},
		},

		"neq?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				// Inline equality check to avoid GetAll() allocation
				// Compare integers
				if left, ok := args[0].(*object.Integer); ok {
					if right, ok := args[1].(*object.Integer); ok {
						return nativeBoolToBooleanObject(left.Value != right.Value)
					}
				}

				// Compare floats
				if left, ok := args[0].(*object.Float); ok {
					if right, ok := args[1].(*object.Float); ok {
						return nativeBoolToBooleanObject(left.Value != right.Value)
					}
				}

				// Compare strings
				if left, ok := args[0].(*object.String); ok {
					if right, ok := args[1].(*object.String); ok {
						return nativeBoolToBooleanObject(left.Value != right.Value)
					}
				}

				// Compare booleans
				if left, ok := args[0].(*object.Boolean); ok {
					if right, ok := args[1].(*object.Boolean); ok {
						return nativeBoolToBooleanObject(left.Value != right.Value)
					}
				}

				// Different types are not equal
				return TRUE
			},
		},

		"gte?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				if left, ok := args[0].(*object.Integer); ok {
					if right, ok := args[1].(*object.Integer); ok {
						return nativeBoolToBooleanObject(left.Value >= right.Value)
					}
				}

				if left, ok := args[0].(*object.Float); ok {
					if right, ok := args[1].(*object.Float); ok {
						return nativeBoolToBooleanObject(left.Value >= right.Value)
					}
				}

				return newError("type mismatch: %s >= %s", args[0].Type(), args[1].Type())
			},
		},

		"lte?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				if left, ok := args[0].(*object.Integer); ok {
					if right, ok := args[1].(*object.Integer); ok {
						return nativeBoolToBooleanObject(left.Value <= right.Value)
					}
				}

				if left, ok := args[0].(*object.Float); ok {
					if right, ok := args[1].(*object.Float); ok {
						return nativeBoolToBooleanObject(left.Value <= right.Value)
					}
				}

				return newError("type mismatch: %s <= %s", args[0].Type(), args[1].Type())
			},
		},

		"not": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}

				if arg, ok := args[0].(*object.Boolean); ok {
					if arg.Value {
						return FALSE
					}
					return TRUE
				}

				return newError("argument to `not` must be BOOLEAN, got %s", args[0].Type())
			},
		},

		"present?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}

				// Check for zero values
				switch obj := args[0].(type) {
				case *object.String:
					return nativeBoolToBooleanObject(obj.Value != "")
				case *object.Integer:
					return nativeBoolToBooleanObject(obj.Value != 0)
				case *object.Float:
					return nativeBoolToBooleanObject(obj.Value != 0.0)
				case *object.Array:
					return nativeBoolToBooleanObject(len(obj.Elements) > 0)
				case *object.Map:
					return nativeBoolToBooleanObject(len(obj.Pairs) > 0)
				case *object.Error:
					return nativeBoolToBooleanObject(len(obj.Msg) > 0 || len(obj.Context) > 0)
				case *object.Null:
					return FALSE
				default:
					return TRUE
				}
			},
		},

		"absent?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}

				// Inline present? logic to avoid InitComparison() allocation
				switch obj := args[0].(type) {
				case *object.String:
					return nativeBoolToBooleanObject(obj.Value == "")
				case *object.Integer:
					return nativeBoolToBooleanObject(obj.Value == 0)
				case *object.Float:
					return nativeBoolToBooleanObject(obj.Value == 0.0)
				case *object.Array:
					return nativeBoolToBooleanObject(len(obj.Elements) == 0)
				case *object.Map:
					return nativeBoolToBooleanObject(len(obj.Pairs) == 0)
				case *object.Error:
					return nativeBoolToBooleanObject(len(obj.Msg) == 0 && len(obj.Context) == 0)
				case *object.Null:
					return TRUE
				default:
					return FALSE
				}
			},
		},
	}
}
