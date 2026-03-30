package modules

import (
	"fmt"

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
					entry := &object.Tuple{Elements: []object.Object{
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
					case *object.Tuple:
						// (key, value) format
						if len(entry.Elements) != 2 {
							return helpers.NewError("map.from_entries: tuple entry %d must have 2 elements, got=%d", i, len(entry.Elements))
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
						return helpers.NewError("map.from_entries: entry %d must be map, array, or tuple, got=%s", i, elem.Type())
					}
				}

				return &object.Map{Pairs: pairs, Keys: keys}
			},
		},

		"map.extract": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("map.extract requires 1 argument (map or array), got=%d", len(args))
				}

				switch container := args[0].(type) {
				case *object.Map:
					values := make([]object.Object, 0, len(container.Keys))
					for _, k := range container.Keys {
						values = append(values, container.Pairs[k])
					}
					return &object.Array{Elements: values}
				case *object.Array:
					return container
				default:
					return helpers.NewError("map.extract requires map or array argument, got=%s", args[0].Type())
				}
			},
		},

		"map.get_or_path": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 3 {
					return helpers.NewError("map.get_or_path requires at least 3 arguments (container, default, keys...), got=%d", len(args))
				}

				current := args[0]
				defaultVal := args[1]
				keys := args[2:]

				for _, pathArg := range keys {
					switch container := current.(type) {
					case *object.Map:
						key, ok := pathArg.(*object.String)
						if !ok {
							return defaultVal
						}
						val, exists := container.Pairs[key.Value]
						if !exists {
							return defaultVal
						}
						current = val
					case *object.Array:
						index, ok := pathArg.(*object.Integer)
						if !ok {
							return defaultVal
						}
						if index.Value < 0 || index.Value >= int64(len(container.Elements)) {
							return defaultVal
						}
						current = container.Elements[index.Value]
					default:
						return defaultVal
					}
				}

				return current
			},
		},

		"map.descend": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("map.descend requires 2 arguments (container, field_name), got=%d", len(args))
				}

				fieldName, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("map.descend requires string field_name, got=%s", args[1].Type())
				}

				var results []object.Object
				var descend func(obj object.Object, depth int) object.Object
				descend = func(obj object.Object, depth int) object.Object {
					if depth > MaxJSONDepth {
						return helpers.NewExecutionError(
							"depth exceeded",
							fmt.Sprintf("recursive descent exceeded maximum depth of %d", MaxJSONDepth),
						)
					}

					switch container := obj.(type) {
					case *object.Map:
						for _, k := range container.Keys {
							if k == fieldName.Value {
								results = append(results, container.Pairs[k])
							}
							if err := descend(container.Pairs[k], depth+1); err != nil {
								return err
							}
						}
					case *object.Array:
						for _, elem := range container.Elements {
							if err := descend(elem, depth+1); err != nil {
								return err
							}
						}
					case *object.LazyMap:
						// Materialize lazy map before descending
						barkObj, err := parseJSONDirect([]byte(container.RawJSON))
						if err != nil {
							return helpers.NewExecutionError("materialize failed", err.Error())
						}
						if errObj := descend(barkObj, depth); errObj != nil {
							return errObj
						}
					}
					return nil
				}

				switch args[0].(type) {
				case *object.Map, *object.Array, *object.LazyMap:
					// valid container types
				default:
					return helpers.NewError("map.descend requires map or array argument, got=%s", args[0].Type())
				}

				if errObj := descend(args[0], 0); errObj != nil {
					return errObj
				}

				if results == nil {
					results = []object.Object{}
				}
				return &object.Array{Elements: results}
			},
		},

		"map.del_path": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 2 {
					return helpers.NewError("map.del_path requires at least 2 arguments (container, keys...), got=%d", len(args))
				}

				container, ok := args[0].(*object.Map)
				if !ok {
					return helpers.NewError("map.del_path requires map as first argument, got=%s", args[0].Type())
				}

				keys := args[1:]
				if len(keys) == 1 {
					key, ok := keys[0].(*object.String)
					if !ok {
						return helpers.NewError("map.del_path requires string keys, got=%s", keys[0].Type())
					}
					return container.COWDelete(key.Value)
				}

				// Navigate to parent of final key
				current := object.Object(container)
				parents := make([]object.Object, 0, len(keys)-1)
				parentKeys := make([]object.Object, 0, len(keys)-1)

				for _, pathArg := range keys[:len(keys)-1] {
					parents = append(parents, current)
					parentKeys = append(parentKeys, pathArg)

					switch c := current.(type) {
					case *object.Map:
						key, ok := pathArg.(*object.String)
						if !ok {
							return helpers.NewError("map.del_path requires string keys for maps, got=%s", pathArg.Type())
						}
						val, exists := c.Pairs[key.Value]
						if !exists {
							return helpers.NewExecutionError(
								"key not found",
								fmt.Sprintf("intermediate key \"%s\" does not exist", key.Value),
							)
						}
						current = val
					case *object.Array:
						index, ok := pathArg.(*object.Integer)
						if !ok {
							return helpers.NewError("map.del_path requires integer index for arrays, got=%s", pathArg.Type())
						}
						if index.Value < 0 || index.Value >= int64(len(c.Elements)) {
							return helpers.NewExecutionError(
								"index out of bounds",
								fmt.Sprintf("index %d is out of range for array of length %d", index.Value, len(c.Elements)),
							)
						}
						current = c.Elements[index.Value]
					default:
						return helpers.NewExecutionError(
							"cannot traverse",
							fmt.Sprintf("cannot index into %s", current.Type()),
						)
					}
				}

				// Delete the final key
				finalKey := keys[len(keys)-1]
				switch c := current.(type) {
				case *object.Map:
					key, ok := finalKey.(*object.String)
					if !ok {
						return helpers.NewError("map.del_path requires string keys for maps, got=%s", finalKey.Type())
					}
					current = c.COWDelete(key.Value)
				default:
					return helpers.NewError("map.del_path final target must be a map, got=%s", current.Type())
				}

				// Rebuild path from bottom up
				for i := len(parents) - 1; i >= 0; i-- {
					switch p := parents[i].(type) {
					case *object.Map:
						key := parentKeys[i].(*object.String)
						current = p.COWSet(key.Value, current)
					case *object.Array:
						index := parentKeys[i].(*object.Integer)
						newElements := make([]object.Object, len(p.Elements))
						copy(newElements, p.Elements)
						newElements[index.Value] = current
						current = &object.Array{Elements: newElements}
					}
				}

				return current
			},
		},

		"map.deep_merge": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 2 {
					return helpers.NewError("map.deep_merge requires at least 2 arguments (map, map...), got=%d", len(args))
				}

				base, ok := args[0].(*object.Map)
				if !ok {
					return helpers.NewError("map.deep_merge requires map arguments, got=%s", args[0].Type())
				}

				result := copyMap(base)

				for i := 1; i < len(args); i++ {
					other, ok := args[i].(*object.Map)
					if !ok {
						return helpers.NewError("map.deep_merge requires map arguments, got=%s", args[i].Type())
					}
					result = deepMerge(result, other)
				}

				return result
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

func copyMap(m *object.Map) *object.Map {
	newPairs := make(map[string]object.Object, len(m.Pairs))
	for k, v := range m.Pairs {
		newPairs[k] = v
	}
	newKeys := make([]string, len(m.Keys))
	copy(newKeys, m.Keys)
	return &object.Map{Pairs: newPairs, Keys: newKeys}
}

func deepMerge(base, other *object.Map) *object.Map {
	result := copyMap(base)
	keyExists := make(map[string]bool, len(result.Keys))
	for _, k := range result.Keys {
		keyExists[k] = true
	}

	for _, k := range other.Keys {
		baseVal, baseHas := result.Pairs[k]
		otherVal := other.Pairs[k]

		if baseHas {
			baseMap, baseIsMap := baseVal.(*object.Map)
			otherMap, otherIsMap := otherVal.(*object.Map)
			if baseIsMap && otherIsMap {
				result.Pairs[k] = deepMerge(baseMap, otherMap)
			} else {
				result.Pairs[k] = otherVal
			}
		} else {
			result.Pairs[k] = otherVal
			result.Keys = append(result.Keys, k)
			keyExists[k] = true
		}
	}

	return result
}
