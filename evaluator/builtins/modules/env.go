package modules

import (
	"os"

	"gitlab.com/bark-lang/bark/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/bark/object"
)

// InitEnv initializes environment variable operations
func InitEnv() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"env.get": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("env.get requires 1 argument (key), got=%d", len(args))
				}

				key, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("env.get requires string argument, got=%s", args[0].Type())
				}

				// Get environment variable (returns "" if not set)
				value := os.Getenv(key.Value)
				return &object.String{Value: value}
			},
		},

		"env.get_or": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("env.get_or requires 2 arguments (key, default), got=%d", len(args))
				}

				key, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("env.get_or requires string key, got=%s", args[0].Type())
				}

				defaultVal, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("env.get_or requires string default, got=%s", args[1].Type())
				}

				// Get environment variable, return default if not set
				value := os.Getenv(key.Value)
				if value == "" {
					// Check if variable is actually set to empty string vs not set at all
					_, exists := os.LookupEnv(key.Value)
					if !exists {
						return defaultVal
					}
				}
				return &object.String{Value: value}
			},
		},

		"env.present?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("env.present? requires 1 argument (key), got=%d", len(args))
				}

				key, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("env.present? requires string argument, got=%s", args[0].Type())
				}

				// Check if environment variable exists (even if empty string)
				_, exists := os.LookupEnv(key.Value)
				return helpers.NativeBoolToBooleanObject(exists)
			},
		},

		"env.absent?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("env.absent? requires 1 argument (key), got=%d", len(args))
				}

				key, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("env.absent? requires string argument, got=%s", args[0].Type())
				}

				// Check if environment variable does not exist
				_, exists := os.LookupEnv(key.Value)
				return helpers.NativeBoolToBooleanObject(!exists)
			},
		},

		"env.all": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 0 {
					return helpers.NewError("env.all requires 0 arguments, got=%d", len(args))
				}

				// Get all environment variables
				envVars := os.Environ()

				// Build map of key-value pairs
				envMap := &object.Map{
					Pairs: make(map[string]object.Object),
					Keys:  make([]string, 0, len(envVars)),
				}

				// Parse KEY=VALUE format
				for _, envVar := range envVars {
					// Find first = to split key and value
					for i := 0; i < len(envVar); i++ {
						if envVar[i] == '=' {
							key := envVar[:i]
							value := envVar[i+1:]
							envMap.Pairs[key] = &object.String{Value: value}
							envMap.Keys = append(envMap.Keys, key)
							break
						}
					}
				}

				return envMap
			},
		},
	}
}
