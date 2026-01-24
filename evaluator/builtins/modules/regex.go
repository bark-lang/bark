package modules

import (
	"fmt"
	"regexp"

	"gitlab.com/bark-lang/bark/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/bark/object"
)

// regexError creates an ExecutionError for invalid regex patterns.
// The evaluator will enrich this with source location (line, column) before logging.
func regexError(funcName, pattern string, err error) *object.ExecutionError {
	return helpers.NewExecutionError(
		fmt.Sprintf("%s: invalid regex pattern", funcName),
		fmt.Sprintf("pattern %q: %s", pattern, err),
	)
}

// InitRegex initializes regular expression operations
func InitRegex() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"regex.match?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("regex.match? requires 2 arguments (text, pattern), got=%d", len(args))
				}

				text, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("regex.match? requires string text, got=%s", args[0].Type())
				}

				pattern, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("regex.match? requires string pattern, got=%s", args[1].Type())
				}

				// Compile and match regex
				matched, err := regexp.MatchString(pattern.Value, text.Value)
				if err != nil {
					return regexError("regex.match?", pattern.Value, err)
				}

				return helpers.NativeBoolToBooleanObject(matched)
			},
		},

		"regex.find": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("regex.find requires 2 arguments (text, pattern), got=%d", len(args))
				}

				text, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("regex.find requires string text, got=%s", args[0].Type())
				}

				pattern, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("regex.find requires string pattern, got=%s", args[1].Type())
				}

				// Compile regex
				re, err := regexp.Compile(pattern.Value)
				if err != nil {
					return regexError("regex.find", pattern.Value, err)
				}

				// Find first match
				match := re.FindString(text.Value)
				return &object.String{Value: match}
			},
		},

		"regex.find_all": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("regex.find_all requires 2 arguments (text, pattern), got=%d", len(args))
				}

				text, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("regex.find_all requires string text, got=%s", args[0].Type())
				}

				pattern, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("regex.find_all requires string pattern, got=%s", args[1].Type())
				}

				// Compile regex
				re, err := regexp.Compile(pattern.Value)
				if err != nil {
					return regexError("regex.find_all", pattern.Value, err)
				}

				// Find all matches
				matches := re.FindAllString(text.Value, -1)
				elements := make([]object.Object, len(matches))
				for i, match := range matches {
					elements[i] = &object.String{Value: match}
				}

				return &object.Array{Elements: elements}
			},
		},

		"regex.replace": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return helpers.NewError("regex.replace requires 3 arguments (text, pattern, replacement), got=%d", len(args))
				}

				text, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("regex.replace requires string text, got=%s", args[0].Type())
				}

				pattern, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("regex.replace requires string pattern, got=%s", args[1].Type())
				}

				replacement, ok := args[2].(*object.String)
				if !ok {
					return helpers.NewError("regex.replace requires string replacement, got=%s", args[2].Type())
				}

				// Compile regex
				re, err := regexp.Compile(pattern.Value)
				if err != nil {
					return regexError("regex.replace", pattern.Value, err)
				}

				// Replace all occurrences
				result := re.ReplaceAllString(text.Value, replacement.Value)
				return &object.String{Value: result}
			},
		},

		"regex.split": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("regex.split requires 2 arguments (text, pattern), got=%d", len(args))
				}

				text, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("regex.split requires string text, got=%s", args[0].Type())
				}

				pattern, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("regex.split requires string pattern, got=%s", args[1].Type())
				}

				// Compile regex
				re, err := regexp.Compile(pattern.Value)
				if err != nil {
					return regexError("regex.split", pattern.Value, err)
				}

				// Split by regex
				parts := re.Split(text.Value, -1)
				elements := make([]object.Object, len(parts))
				for i, part := range parts {
					elements[i] = &object.String{Value: part}
				}

				return &object.Array{Elements: elements}
			},
		},
	}
}
