package modules

import (
	"encoding/json"

	"gitlab.com/bark-lang/bark/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/bark/object"
)

// InitJSON initializes JSON operations
func InitJSON() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"json.parse": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("json.parse requires 1 argument (json_string), got=%d", len(args))
				}

				jsonStr, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("json.parse requires string argument, got=%s", args[0].Type())
				}

				// Parse JSON into interface{}
				var data interface{}
				if err := json.Unmarshal([]byte(jsonStr.Value), &data); err != nil {
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

				// Convert JSON data to Bark object
				barkObj := jsonToBark(data)

				// Return success tuple: ({}, data)
				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						barkObj,
					},
				}
			},
		},

		"json.stringify": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("json.stringify requires 1 argument (data), got=%d", len(args))
				}

				// Convert Bark object to JSON-compatible interface
				data := barkToJSON(args[0])

				// Marshal to compact JSON
				jsonBytes, err := json.Marshal(data)
				if err != nil {
					return helpers.NewError("json.stringify failed: %s", err.Error())
				}

				return &object.String{Value: string(jsonBytes)}
			},
		},

		"json.stringify_pretty": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("json.stringify_pretty requires 1 argument (data), got=%d", len(args))
				}

				// Convert Bark object to JSON-compatible interface
				data := barkToJSON(args[0])

				// Marshal to pretty-printed JSON with 2-space indentation
				jsonBytes, err := json.MarshalIndent(data, "", "  ")
				if err != nil {
					return helpers.NewError("json.stringify_pretty failed: %s", err.Error())
				}

				return &object.String{Value: string(jsonBytes)}
			},
		},
	}
}

// jsonToBark converts JSON data (from unmarshal) to Bark objects
func jsonToBark(data interface{}) object.Object {
	switch v := data.(type) {
	case map[string]interface{}:
		// JSON object → Bark map
		pairs := make(map[string]object.Object)
		keys := make([]string, 0, len(v))

		// Maintain key order (Go 1.12+ maintains map iteration order for json)
		for key, value := range v {
			pairs[key] = jsonToBark(value)
			keys = append(keys, key)
		}

		return &object.Map{Pairs: pairs, Keys: keys}

	case []interface{}:
		// JSON array → Bark array
		elements := make([]object.Object, len(v))
		for i, item := range v {
			elements[i] = jsonToBark(item)
		}
		return &object.Array{Elements: elements}

	case string:
		return &object.String{Value: v}

	case float64:
		// JSON numbers are always float64
		// Convert to int if whole number, otherwise float
		if v == float64(int64(v)) {
			return &object.Integer{Value: int64(v)}
		}
		// Note: Bark doesn't have a Float type yet, so we convert to int
		// This might lose precision for non-integer numbers
		return &object.Integer{Value: int64(v)}

	case bool:
		return helpers.NativeBoolToBooleanObject(v)

	case nil:
		// JSON null → empty string
		return &object.String{Value: ""}

	default:
		// Unknown type
		return &object.String{Value: ""}
	}
}

// barkToJSON converts Bark objects to JSON-compatible interface{}
func barkToJSON(obj object.Object) interface{} {
	switch v := obj.(type) {
	case *object.Map:
		// Bark map → JSON object (maintain key order)
		result := make(map[string]interface{})
		for _, key := range v.Keys {
			if value, ok := v.Pairs[key]; ok {
				result[key] = barkToJSON(value)
			}
		}
		return result

	case *object.Array:
		// Bark array → JSON array
		result := make([]interface{}, len(v.Elements))
		for i, elem := range v.Elements {
			result[i] = barkToJSON(elem)
		}
		return result

	case *object.String:
		return v.Value

	case *object.Integer:
		return v.Value

	case *object.Boolean:
		return v.Value

	case *object.Null:
		return nil

	default:
		// For unsupported types (Error, Function, etc.), return null
		return nil
	}
}
