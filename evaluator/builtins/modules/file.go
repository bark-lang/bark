package modules

import (
	"fmt"
	"os"
	"path/filepath"

	"gitlab.com/bark-lang/bark/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/bark/object"
)

// InitFile initializes file I/O operations
func InitFile() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"file.read": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("file.read requires 1 argument (path), got=%d", len(args))
				}

				path, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("file.read requires string argument, got=%s", args[0].Type())
				}

				// Read file
				content, err := os.ReadFile(path.Value)
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
						&object.String{Value: string(content)},
					},
				}
			},
		},

		"file.write": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("file.write requires 2 arguments (path, content), got=%d", len(args))
				}

				path, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("file.write requires string path, got=%s", args[0].Type())
				}

				content, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("file.write requires string content, got=%s", args[1].Type())
				}

				// Create parent directories if they don't exist
				dir := filepath.Dir(path.Value)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return &object.Error{
						Msg:     err.Error(),
						Context: make(map[string]object.Object),
					}
				}

				// Write file (overwrite if exists)
				if err := os.WriteFile(path.Value, []byte(content.Value), 0644); err != nil {
					return &object.Error{
						Msg:     err.Error(),
						Context: make(map[string]object.Object),
					}
				}

				// Return empty map (no error)
				return &object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}
			},
		},

		"file.append": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("file.append requires 2 arguments (path, content), got=%d", len(args))
				}

				path, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("file.append requires string path, got=%s", args[0].Type())
				}

				content, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("file.append requires string content, got=%s", args[1].Type())
				}

				// Open file for appending (create if doesn't exist)
				f, err := os.OpenFile(path.Value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return &object.Error{
						Msg:     err.Error(),
						Context: make(map[string]object.Object),
					}
				}
				defer func() { _ = f.Close() }()

				// Write content
				if _, err := f.WriteString(content.Value); err != nil {
					return &object.Error{
						Msg:     err.Error(),
						Context: make(map[string]object.Object),
					}
				}

				// Return empty map (no error)
				return &object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}
			},
		},

		"file.exists?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("file.exists? requires 1 argument (path), got=%d", len(args))
				}

				path, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("file.exists? requires string argument, got=%s", args[0].Type())
				}

				// Check if file exists
				info, err := os.Stat(path.Value)
				if err != nil {
					// File doesn't exist or permission denied
					return helpers.FALSE
				}

				// Return false for directories (use dir.exists? for directories)
				if info.IsDir() {
					return helpers.FALSE
				}

				return helpers.TRUE
			},
		},

		"file.absent?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("file.absent? requires 1 argument (path), got=%d", len(args))
				}

				path, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("file.absent? requires string argument, got=%s", args[0].Type())
				}

				// Check if file exists
				info, err := os.Stat(path.Value)
				if err != nil {
					// File doesn't exist or permission denied
					return helpers.TRUE
				}

				// Return true for directories (use dir.absent? for directories)
				if info.IsDir() {
					return helpers.TRUE
				}

				return helpers.FALSE
			},
		},

		"file.delete": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("file.delete requires 1 argument (path), got=%d", len(args))
				}

				path, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("file.delete requires string argument, got=%s", args[0].Type())
				}

				// Check if path is a directory
				info, err := os.Stat(path.Value)
				if err == nil && info.IsDir() {
					return &object.Error{
						Msg:     fmt.Sprintf("path is a directory, not a file: %s", path.Value),
						Context: make(map[string]object.Object),
					}
				}

				// Delete file
				if err := os.Remove(path.Value); err != nil {
					return &object.Error{
						Msg:     err.Error(),
						Context: make(map[string]object.Object),
					}
				}

				// Return empty map (no error)
				return &object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}
			},
		},

		"file.info": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("file.info requires 1 argument (path), got=%d", len(args))
				}

				path, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("file.info requires string argument, got=%s", args[0].Type())
				}

				// Get file info
				info, err := os.Stat(path.Value)
				if err != nil {
					// Return error tuple: (error, {})
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     err.Error(),
								Context: make(map[string]object.Object),
							},
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}

				// Build info map
				infoMap := &object.Map{
					Pairs: make(map[string]object.Object),
					Keys:  []string{"size", "modified", "is_dir"},
				}

				infoMap.Pairs["size"] = &object.Integer{Value: info.Size()}
				infoMap.Pairs["modified"] = &object.Integer{Value: info.ModTime().Unix()}
				infoMap.Pairs["is_dir"] = helpers.NativeBoolToBooleanObject(info.IsDir())

				// Return success tuple: ({}, info_map)
				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						infoMap,
					},
				}
			},
		},
	}
}
