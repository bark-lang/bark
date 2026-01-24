package modules

import (
	"encoding/base64"

	"gitlab.com/bark-lang/bark/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/bark/object"
)

// InitBase64 initializes base64 encoding operations
func InitBase64() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"base64.encode": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("base64.encode requires 1 argument (data), got=%d", len(args))
				}

				data, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("base64.encode requires string argument, got=%s", args[0].Type())
				}

				// Encode to base64
				encoded := base64.StdEncoding.EncodeToString([]byte(data.Value))
				return &object.String{Value: encoded}
			},
		},

		"base64.decode": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("base64.decode requires 1 argument (encoded), got=%d", len(args))
				}

				encoded, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("base64.decode requires string argument, got=%s", args[0].Type())
				}

				// Decode from base64
				decoded, err := base64.StdEncoding.DecodeString(encoded.Value)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.String{Value: ""},
						},
					}
				}

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						&object.String{Value: string(decoded)},
					},
				}
			},
		},
	}
}
