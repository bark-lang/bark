package builtins

import (
	"fmt"
	"strings"

	"gitlab.com/bark-lang/barki/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/barki/object"
)

// InitDataStructures initializes data structure operations that work on both arrays and maps
func InitDataStructures() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"get": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 2 {
					return newError("wrong number of arguments. got=%d, want=2+", len(args))
				}

				// Walk through each path segment
				current := args[0]
				for _, pathArg := range args[1:] {
					// Handle maps
					if m, ok := current.(*object.Map); ok {
						key, ok := pathArg.(*object.String)
						if !ok {
							return newError("map key must be STRING, got %s", pathArg.Type())
						}

						val, exists := m.Pairs[key.Value]
						if !exists {
							return helpers.NewExecutionError(
								"key not found",
								fmt.Sprintf("key \"%s\" does not exist in map", key.Value),
							)
						}
						current = val
						continue
					}

					// Handle arrays
					if arr, ok := current.(*object.Array); ok {
						index, ok := pathArg.(*object.Integer)
						if !ok {
							return newError("array index must be INTEGER, got %s", pathArg.Type())
						}

						if index.Value < 0 || index.Value >= int64(len(arr.Elements)) {
							return helpers.NewExecutionError(
								"index out of bounds",
								fmt.Sprintf("index %d is out of range for array of length %d", index.Value, len(arr.Elements)),
							)
						}
						current = arr.Elements[index.Value]
						continue
					}

					return newError("cannot index into %s with get()", current.Type())
				}

				return current
			},
		},

		"set": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return newError("wrong number of arguments. got=%d, want=3", len(args))
				}

				// Handle maps
				if m, ok := args[0].(*object.Map); ok {
					key, ok := args[1].(*object.String)
					if !ok {
						return newError("map key must be STRING, got %s", args[1].Type())
					}

					// Use COW: if map is not shared, modify in place
					return m.COWSet(key.Value, args[2])
				}

				// Handle arrays
				if arr, ok := args[0].(*object.Array); ok {
					index, ok := args[1].(*object.Integer)
					if !ok {
						return newError("array index must be INTEGER, got %s", args[1].Type())
					}

					if index.Value < 0 || index.Value >= int64(len(arr.Elements)) {
						return newError("index out of bounds: %d", index.Value)
					}

					// Use COW: if array is not shared, modify in place
					return arr.COWSet(int(index.Value), args[2])
				}

				return newError("first argument to `set` must be MAP or ARRAY, got %s", args[0].Type())
			},
		},

		"first": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}

				// Handle maps
				if m, ok := args[0].(*object.Map); ok {
					if len(m.Keys) == 0 {
						// Return ("", zero_value)
						return &object.Tuple{Elements: []object.Object{
							&object.String{Value: ""},
							&object.Integer{Value: 0},
						}}
					}

					firstKey := m.Keys[0]
					return &object.Tuple{Elements: []object.Object{
						&object.String{Value: firstKey},
						m.Pairs[firstKey],
					}}
				}

				// Handle arrays
				if arr, ok := args[0].(*object.Array); ok {
					if len(arr.Elements) == 0 {
						// Return (-1, zero_value)
						return &object.Array{Elements: []object.Object{
							&object.Integer{Value: -1},
							&object.Integer{Value: 0},
						}}
					}

					// Return (index, value)
					return &object.Array{Elements: []object.Object{
						&object.Integer{Value: 0},
						arr.Elements[0],
					}}
				}

				return newError("argument to `first` must be MAP or ARRAY, got %s", args[0].Type())
			},
		},

		"last": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}

				// Handle maps
				if m, ok := args[0].(*object.Map); ok {
					if len(m.Keys) == 0 {
						// Return ("", zero_value)
						return &object.Tuple{Elements: []object.Object{
							&object.String{Value: ""},
							&object.Integer{Value: 0},
						}}
					}

					lastKey := m.Keys[len(m.Keys)-1]
					return &object.Tuple{Elements: []object.Object{
						&object.String{Value: lastKey},
						m.Pairs[lastKey],
					}}
				}

				// Handle arrays
				if arr, ok := args[0].(*object.Array); ok {
					if len(arr.Elements) == 0 {
						// Return (-1, zero_value)
						return &object.Array{Elements: []object.Object{
							&object.Integer{Value: -1},
							&object.Integer{Value: 0},
						}}
					}

					// Return (index, value)
					lastIdx := len(arr.Elements) - 1
					return &object.Array{Elements: []object.Object{
						&object.Integer{Value: int64(lastIdx)},
						arr.Elements[lastIdx],
					}}
				}

				return newError("argument to `last` must be MAP or ARRAY, got %s", args[0].Type())
			},
		},

		"next": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				// Handle maps
				if m, ok := args[0].(*object.Map); ok {
					key, ok := args[1].(*object.String)
					if !ok {
						return newError("second argument to `next` must be STRING for maps, got %s", args[1].Type())
					}

					// Find the key in the Keys slice
					for i, k := range m.Keys {
						if k == key.Value {
							// Check if there's a next key
							if i < len(m.Keys)-1 {
								return &object.String{Value: m.Keys[i+1]}
							}
							// At end
							return &object.String{Value: ""}
						}
					}

					// Key not found
					return &object.String{Value: ""}
				}

				// Handle arrays
				if arr, ok := args[0].(*object.Array); ok {
					index, ok := args[1].(*object.Integer)
					if !ok {
						return newError("second argument to `next` must be INTEGER for arrays, got %s", args[1].Type())
					}

					// Return -1 if index is invalid or at the end
					if index.Value < 0 || index.Value >= int64(len(arr.Elements)-1) {
						return &object.Integer{Value: -1}
					}

					return &object.Integer{Value: index.Value + 1}
				}

				return newError("first argument to `next` must be MAP or ARRAY, got %s", args[0].Type())
			},
		},

		"prev": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				// Handle maps
				if m, ok := args[0].(*object.Map); ok {
					key, ok := args[1].(*object.String)
					if !ok {
						return newError("second argument to `prev` must be STRING for maps, got %s", args[1].Type())
					}

					// Find the key in the Keys slice
					for i, k := range m.Keys {
						if k == key.Value {
							// Check if there's a previous key
							if i > 0 {
								return &object.String{Value: m.Keys[i-1]}
							}
							// At beginning
							return &object.String{Value: ""}
						}
					}

					// Key not found
					return &object.String{Value: ""}
				}

				// Handle arrays
				if arr, ok := args[0].(*object.Array); ok {
					index, ok := args[1].(*object.Integer)
					if !ok {
						return newError("second argument to `prev` must be INTEGER for arrays, got %s", args[1].Type())
					}

					// Return -1 if index is invalid or at the beginning
					if index.Value <= 0 || index.Value >= int64(len(arr.Elements)) {
						return &object.Integer{Value: -1}
					}

					return &object.Integer{Value: index.Value - 1}
				}

				return newError("first argument to `prev` must be MAP or ARRAY, got %s", args[0].Type())
			},
		},

		"head": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}

				// Handle maps
				if m, ok := args[0].(*object.Map); ok {
					if len(m.Keys) == 0 {
						// Return (map, "") for empty map
						return &object.Tuple{Elements: []object.Object{
							m,
							&object.String{Value: ""},
						}}
					}

					// Return (map, first_key) for non-empty map
					return &object.Tuple{Elements: []object.Object{
						m,
						&object.String{Value: m.Keys[0]},
					}}
				}

				// Handle arrays
				if arr, ok := args[0].(*object.Array); ok {
					if len(arr.Elements) == 0 {
						// Return (array, -1) for empty array
						return &object.Tuple{Elements: []object.Object{
							arr,
							&object.Integer{Value: -1},
						}}
					}

					// Return (array, 0) for non-empty array
					return &object.Tuple{Elements: []object.Object{
						arr,
						&object.Integer{Value: 0},
					}}
				}

				return newError("argument to `head` must be MAP or ARRAY, got %s", args[0].Type())
			},
		},

		"tail": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}

				// Handle maps
				if m, ok := args[0].(*object.Map); ok {
					if len(m.Keys) == 0 {
						// Return (map, "") for empty map
						return &object.Tuple{Elements: []object.Object{
							m,
							&object.String{Value: ""},
						}}
					}

					// Return (map, last_key) for non-empty map
					lastKey := m.Keys[len(m.Keys)-1]
					return &object.Tuple{Elements: []object.Object{
						m,
						&object.String{Value: lastKey},
					}}
				}

				// Handle arrays
				if arr, ok := args[0].(*object.Array); ok {
					if len(arr.Elements) == 0 {
						// Return (array, -1) for empty array
						return &object.Tuple{Elements: []object.Object{
							arr,
							&object.Integer{Value: -1},
						}}
					}

					// Return (array, last_index) for non-empty array
					lastIdx := len(arr.Elements) - 1
					return &object.Tuple{Elements: []object.Object{
						arr,
						&object.Integer{Value: int64(lastIdx)},
					}}
				}

				return newError("argument to `tail` must be MAP or ARRAY, got %s", args[0].Type())
			},
		},

		// Polymorphic containment checks (work on STRING and ARRAY)
		"includes?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				// String contains
				if str, ok := args[0].(*object.String); ok {
					if substr, ok := args[1].(*object.String); ok {
						return nativeBoolToBooleanObject(strings.Contains(str.Value, substr.Value))
					}
				}

				// Array contains - use ObjectsEqual to avoid string allocations
				if arr, ok := args[0].(*object.Array); ok {
					target := args[1]
					for _, elem := range arr.Elements {
						if helpers.ObjectsEqual(elem, target) {
							return TRUE
						}
					}
					return FALSE
				}

				return newError("first argument to `includes?` must be STRING or ARRAY, got %s", args[0].Type())
			},
		},

		"excludes?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				// String contains
				if str, ok := args[0].(*object.String); ok {
					if substr, ok := args[1].(*object.String); ok {
						return nativeBoolToBooleanObject(!strings.Contains(str.Value, substr.Value))
					}
				}

				// Array contains - use ObjectsEqual to avoid string allocations
				if arr, ok := args[0].(*object.Array); ok {
					target := args[1]
					for _, elem := range arr.Elements {
						if helpers.ObjectsEqual(elem, target) {
							return FALSE
						}
					}
					return TRUE
				}

				return newError("first argument to `excludes?` must be STRING or ARRAY, got %s", args[0].Type())
			},
		},

		"size": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}

				switch obj := args[0].(type) {
				case *object.Map:
					return &object.Integer{Value: int64(len(obj.Pairs))}
				case *object.Array:
					return &object.Integer{Value: int64(len(obj.Elements))}
				case *object.String:
					return &object.Integer{Value: int64(len(obj.Value))}
				default:
					return newError("argument to `size` must be MAP, ARRAY, or STRING, got %s", args[0].Type())
				}
			},
		},

		"reverse": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}

				switch obj := args[0].(type) {
				case *object.Array:
					newElements := make([]object.Object, len(obj.Elements))
					for i, elem := range obj.Elements {
						newElements[len(obj.Elements)-1-i] = elem
					}
					return &object.Array{Elements: newElements}
				case *object.String:
					runes := []rune(obj.Value)
					for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
						runes[i], runes[j] = runes[j], runes[i]
					}
					return &object.String{Value: string(runes)}
				default:
					return newError("argument to `reverse` must be ARRAY or STRING, got %s", args[0].Type())
				}
			},
		},
	}
}
