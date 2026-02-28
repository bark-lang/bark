package builtins

import (
	"gitlab.com/bark-lang/barki/object"
)

// InitArrays initializes array manipulation operations
// Note: Most array functions have moved to the array module (array.push, array.pop, etc.)
// This file retains polymorphic functions that work on multiple types
func InitArrays() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"len": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}

				switch arg := args[0].(type) {
				case *object.String:
					return &object.Integer{Value: int64(len(arg.Value))}
				case *object.Array:
					return &object.Integer{Value: int64(len(arg.Elements))}
				default:
					return newError("argument to `len` not supported, got %s", args[0].Type())
				}
			},
		},

		"empty?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}

				switch obj := args[0].(type) {
				case *object.Array:
					return nativeBoolToBooleanObject(len(obj.Elements) == 0)
				case *object.Map:
					return nativeBoolToBooleanObject(len(obj.Pairs) == 0)
				case *object.String:
					return nativeBoolToBooleanObject(obj.Value == "")
				default:
					return newError("argument to `empty?` must be ARRAY, MAP, or STRING, got %s", args[0].Type())
				}
			},
		},
	}
}
