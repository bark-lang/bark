package modules

import (
	"gitlab.com/bark-lang/barki/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/barki/object"
)

// InitMap initializes map operations
func InitMap() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"map.get_or": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return helpers.NewError("map.get_or requires 3 arguments (map, key, default), got=%d", len(args))
				}

				m, ok := args[0].(*object.Map)
				if !ok {
					return helpers.NewError("map.get_or requires map as first argument, got=%s", args[0].Type())
				}

				key, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("map.get_or requires string key, got=%s", args[1].Type())
				}

				if val, exists := m.Pairs[key.Value]; exists {
					return val
				}

				return args[2] // Return default value
			},
		},

		"map.del": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 2 {
					return helpers.NewError("map.del requires at least 2 arguments (map, key...), got=%d", len(args))
				}

				m, ok := args[0].(*object.Map)
				if !ok {
					return helpers.NewError("map.del requires map as first argument, got=%s", args[0].Type())
				}

				// Use COW: delete keys one at a time
				result := m
				for i := 1; i < len(args); i++ {
					key, ok := args[i].(*object.String)
					if !ok {
						return helpers.NewError("map.del requires string keys, got=%s", args[i].Type())
					}
					result = result.COWDelete(key.Value)
				}

				return result
			},
		},

		"map.keys": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("map.keys requires 1 argument (map), got=%d", len(args))
				}

				m, ok := args[0].(*object.Map)
				if !ok {
					return helpers.NewError("map.keys requires map argument, got=%s", args[0].Type())
				}

				keys := make([]object.Object, 0, len(m.Keys))
				for _, k := range m.Keys {
					keys = append(keys, &object.String{Value: k})
				}

				return &object.Array{Elements: keys}
			},
		},

		"map.values": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("map.values requires 1 argument (map), got=%d", len(args))
				}

				m, ok := args[0].(*object.Map)
				if !ok {
					return helpers.NewError("map.values requires map argument, got=%s", args[0].Type())
				}

				values := make([]object.Object, 0, len(m.Keys))
				for _, k := range m.Keys {
					values = append(values, m.Pairs[k])
				}

				return &object.Array{Elements: values}
			},
		},

		"map.entries": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("map.entries requires 1 argument (map), got=%d", len(args))
				}

				m, ok := args[0].(*object.Map)
				if !ok {
					return helpers.NewError("map.entries requires map argument, got=%s", args[0].Type())
				}

				entries := make([]object.Object, 0, len(m.Keys))
				for _, k := range m.Keys {
					entry := &object.Array{Elements: []object.Object{
						&object.String{Value: k},
						m.Pairs[k],
					}}
					entries = append(entries, entry)
				}

				return &object.Array{Elements: entries}
			},
		},

		"map.key_present?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("map.key_present? requires 2 arguments (map, key), got=%d", len(args))
				}

				m, ok := args[0].(*object.Map)
				if !ok {
					return helpers.NewError("map.key_present? requires map as first argument, got=%s", args[0].Type())
				}

				key, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("map.key_present? requires string key, got=%s", args[1].Type())
				}

				_, exists := m.Pairs[key.Value]
				return helpers.NativeBoolToBooleanObject(exists)
			},
		},

		"map.key_absent?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("map.key_absent? requires 2 arguments (map, key), got=%d", len(args))
				}

				m, ok := args[0].(*object.Map)
				if !ok {
					return helpers.NewError("map.key_absent? requires map as first argument, got=%s", args[0].Type())
				}

				key, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("map.key_absent? requires string key, got=%s", args[1].Type())
				}

				_, exists := m.Pairs[key.Value]
				return helpers.NativeBoolToBooleanObject(!exists)
			},
		},

		"map.from_entries": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("map.from_entries requires 1 argument (array), got=%d", len(args))
				}

				arr, ok := args[0].(*object.Array)
				if !ok {
					return helpers.NewError("map.from_entries requires array argument, got=%s", args[0].Type())
				}

				pairs := make(map[string]object.Object)
				keys := make([]string, 0, len(arr.Elements))

				for i, elem := range arr.Elements {
					switch entry := elem.(type) {
					case *object.Map:
						// {"key": k, "value": v} format
						keyVal, hasKey := entry.Pairs["key"]
						valVal, hasValue := entry.Pairs["value"]
						if !hasKey || !hasValue {
							return helpers.NewError("map.from_entries: entry %d must have 'key' and 'value' fields", i)
						}
						keyStr, ok := keyVal.(*object.String)
						if !ok {
							return helpers.NewError("map.from_entries: entry %d 'key' must be string, got=%s", i, keyVal.Type())
						}
						if _, exists := pairs[keyStr.Value]; !exists {
							keys = append(keys, keyStr.Value)
						}
						pairs[keyStr.Value] = valVal
					case *object.Array:
						// [key, value] format
						if len(entry.Elements) != 2 {
							return helpers.NewError("map.from_entries: array entry %d must have 2 elements, got=%d", i, len(entry.Elements))
						}
						keyStr, ok := entry.Elements[0].(*object.String)
						if !ok {
							return helpers.NewError("map.from_entries: entry %d key must be string, got=%s", i, entry.Elements[0].Type())
						}
						if _, exists := pairs[keyStr.Value]; !exists {
							keys = append(keys, keyStr.Value)
						}
						pairs[keyStr.Value] = entry.Elements[1]
					default:
						return helpers.NewError("map.from_entries: entry %d must be map or array, got=%s", i, elem.Type())
					}
				}

				return &object.Map{Pairs: pairs, Keys: keys}
			},
		},

		"map.merge": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 2 {
					return helpers.NewError("map.merge requires at least 2 arguments (map, map...), got=%d", len(args))
				}

				// First arg must be a map
				base, ok := args[0].(*object.Map)
				if !ok {
					return helpers.NewError("map.merge requires map arguments, got=%s", args[0].Type())
				}

				// Copy base map
				newPairs := make(map[string]object.Object)
				for k, v := range base.Pairs {
					newPairs[k] = v
				}

				// Start with base keys
				newKeys := make([]string, len(base.Keys))
				copy(newKeys, base.Keys)
				keyExists := make(map[string]bool)
				for _, k := range base.Keys {
					keyExists[k] = true
				}

				// Merge all other maps
				for i := 1; i < len(args); i++ {
					m, ok := args[i].(*object.Map)
					if !ok {
						return helpers.NewError("map.merge requires map arguments, got=%s", args[i].Type())
					}

					for _, k := range m.Keys {
						newPairs[k] = m.Pairs[k]
						if !keyExists[k] {
							newKeys = append(newKeys, k)
							keyExists[k] = true
						}
					}
				}

				return &object.Map{Pairs: newPairs, Keys: newKeys}
			},
		},
	}
}
