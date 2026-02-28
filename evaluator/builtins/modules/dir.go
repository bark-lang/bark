package modules

import (
	"os"

	"gitlab.com/bark-lang/barki/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/barki/object"
)

// InitDir initializes directory operations
func InitDir() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"dir.list": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("dir.list requires 1 argument (path), got=%d", len(args))
				}

				path, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("dir.list requires string argument, got=%s", args[0].Type())
				}

				// Read directory
				entries, err := os.ReadDir(path.Value)
				if err != nil {
					// Return error tuple: (error, [])
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     err.Error(),
								Context: make(map[string]object.Object),
							},
							&object.Array{Elements: []object.Object{}},
						},
					}
				}

				// Build array of filenames
				elements := make([]object.Object, len(entries))
				for i, entry := range entries {
					elements[i] = &object.String{Value: entry.Name()}
				}

				// Return success tuple: ({}, filenames)
				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						&object.Array{Elements: elements},
					},
				}
			},
		},

		"dir.exists?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("dir.exists? requires 1 argument (path), got=%d", len(args))
				}

				path, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("dir.exists? requires string argument, got=%s", args[0].Type())
				}

				// Check if directory exists
				info, err := os.Stat(path.Value)
				if err != nil {
					// Directory doesn't exist or permission denied
					return helpers.FALSE
				}

				// Return true only if it's a directory
				return helpers.NativeBoolToBooleanObject(info.IsDir())
			},
		},

		"dir.absent?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("dir.absent? requires 1 argument (path), got=%d", len(args))
				}

				path, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("dir.absent? requires string argument, got=%s", args[0].Type())
				}

				// Check if directory exists
				info, err := os.Stat(path.Value)
				if err != nil {
					// Directory doesn't exist or permission denied
					return helpers.TRUE
				}

				// Return false if it's a directory (i.e., it exists)
				return helpers.NativeBoolToBooleanObject(!info.IsDir())
			},
		},
	}
}
