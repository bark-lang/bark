package modules

import (
	"strconv"
	"strings"

	"gitlab.com/bark-lang/bark/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/bark/object"
)

// InitStr initializes string manipulation operations
func InitStr() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"str.upper": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("str.upper requires 1 argument, got=%d", len(args))
				}

				str, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("str.upper requires string argument, got=%s", args[0].Type())
				}

				return &object.String{Value: strings.ToUpper(str.Value)}
			},
		},

		"str.lower": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("str.lower requires 1 argument, got=%d", len(args))
				}

				str, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("str.lower requires string argument, got=%s", args[0].Type())
				}

				return &object.String{Value: strings.ToLower(str.Value)}
			},
		},

		"str.trim": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("str.trim requires 1 argument, got=%d", len(args))
				}

				str, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("str.trim requires string argument, got=%s", args[0].Type())
				}

				return &object.String{Value: strings.TrimSpace(str.Value)}
			},
		},

		"str.starts_with?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("str.starts_with? requires 2 arguments (string, prefix), got=%d", len(args))
				}

				str, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("str.starts_with? requires string as first argument, got=%s", args[0].Type())
				}

				prefix, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("str.starts_with? requires string prefix, got=%s", args[1].Type())
				}

				return helpers.NativeBoolToBooleanObject(strings.HasPrefix(str.Value, prefix.Value))
			},
		},

		"str.ends_with?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("str.ends_with? requires 2 arguments (string, suffix), got=%d", len(args))
				}

				str, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("str.ends_with? requires string as first argument, got=%s", args[0].Type())
				}

				suffix, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("str.ends_with? requires string suffix, got=%s", args[1].Type())
				}

				return helpers.NativeBoolToBooleanObject(strings.HasSuffix(str.Value, suffix.Value))
			},
		},

		"str.replace": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return helpers.NewError("str.replace requires 3 arguments (string, old, new), got=%d", len(args))
				}

				str, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("str.replace requires string as first argument, got=%s", args[0].Type())
				}

				old, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("str.replace requires string as second argument, got=%s", args[1].Type())
				}

				newStr, ok := args[2].(*object.String)
				if !ok {
					return helpers.NewError("str.replace requires string as third argument, got=%s", args[2].Type())
				}

				return &object.String{Value: strings.ReplaceAll(str.Value, old.Value, newStr.Value)}
			},
		},

		"str.split": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("str.split requires 2 arguments (string, delimiter), got=%d", len(args))
				}

				str, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("str.split requires string as first argument, got=%s", args[0].Type())
				}

				delimiter, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("str.split requires string delimiter, got=%s", args[1].Type())
				}

				parts := strings.Split(str.Value, delimiter.Value)
				elements := make([]object.Object, len(parts))
				for i, part := range parts {
					elements[i] = &object.String{Value: part}
				}

				return &object.Array{Elements: elements}
			},
		},

		"str.join": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("str.join requires 2 arguments (array, separator), got=%d", len(args))
				}

				arr, ok := args[0].(*object.Array)
				if !ok {
					return helpers.NewError("str.join requires array as first argument, got=%s", args[0].Type())
				}

				separator, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("str.join requires string separator, got=%s", args[1].Type())
				}

				parts := make([]string, len(arr.Elements))
				for i, elem := range arr.Elements {
					if str, ok := elem.(*object.String); ok {
						parts[i] = str.Value
					} else {
						parts[i] = elem.Inspect()
					}
				}

				return &object.String{Value: strings.Join(parts, separator.Value)}
			},
		},

		"str.concat": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 2 {
					return helpers.NewError("str.concat requires at least 2 arguments, got=%d", len(args))
				}

				var result strings.Builder
				for i, arg := range args {
					str, ok := arg.(*object.String)
					if !ok {
						return helpers.NewError("str.concat requires string arguments, argument %d is %s", i+1, arg.Type())
					}
					result.WriteString(str.Value)
				}

				return &object.String{Value: result.String()}
			},
		},

		"str.format": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 1 {
					return helpers.NewError("str.format requires at least 1 argument (format string), got=%d", len(args))
				}

				formatStr, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("str.format requires string as first argument, got=%s", args[0].Type())
				}

				result, err := helpers.FormatString(formatStr.Value, args[1:])
				if err != nil {
					return helpers.NewError("str.format: %s", err.Error())
				}

				return &object.String{Value: result}
			},
		},

		"str.numeric?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("str.numeric? requires 1 argument, got=%d", len(args))
				}

				str, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("str.numeric? requires string argument, got=%s", args[0].Type())
				}

				trimmed := strings.TrimSpace(str.Value)
				if trimmed == "" {
					return helpers.FALSE
				}

				// Try parsing as integer first
				if _, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
					return helpers.TRUE
				}

				// Try parsing as float
				if _, err := strconv.ParseFloat(trimmed, 64); err == nil {
					return helpers.TRUE
				}

				return helpers.FALSE
			},
		},
	}
}
