package modules

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"gitlab.com/bark-lang/barki/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/barki/object"
)

// MaxJSONDepth is the maximum allowed nesting depth for JSON parsing (100 levels)
const MaxJSONDepth = 100

// ErrJSONDepthExceeded is returned when JSON nesting exceeds MaxJSONDepth
var ErrJSONDepthExceeded = errors.New("JSON nesting depth exceeds 100 levels")

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

				// Parse JSON directly to Bark objects (memory-efficient, skips interface{})
				barkObj, err := parseJSONDirect([]byte(jsonStr.Value))
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

				// Convert bark object to JSON-compatible interface
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

				// Convert bark object to JSON-compatible interface
				data := barkToJSON(args[0])

				// Marshal to pretty-printed JSON with 2-space indentation
				jsonBytes, err := json.MarshalIndent(data, "", "  ")
				if err != nil {
					return helpers.NewError("json.stringify_pretty failed: %s", err.Error())
				}

				return &object.String{Value: string(jsonBytes)}
			},
		},

		"json.parse_stream": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("json.parse_stream requires 1 argument (iterator), got=%d", len(args))
				}

				iter, ok := args[0].(*object.Iterator)
				if !ok {
					return helpers.NewError("json.parse_stream requires iterator argument, got=%s", args[0].Type())
				}

				// Collect all chunks from iterator into a buffer
				var buf bytes.Buffer
				for !iter.IsExhausted() {
					chunk, hasMore, err := iter.Next()
					if err != nil {
						return &object.Tuple{
							Elements: []object.Object{
								helpers.WrapError(err),
								&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
							},
						}
					}
					if !hasMore {
						iter.MarkExhausted()
						break
					}
					if str, ok := chunk.(*object.String); ok {
						buf.WriteString(str.Value)
					}
				}

				// Parse directly to Bark objects (memory-efficient, skips interface{})
				barkObj, err := parseJSONFromReader(&buf)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						barkObj,
					},
				}
			},
		},

		"json.parse_lazy": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("json.parse_lazy requires 1 argument (json_string), got=%d", len(args))
				}

				jsonStr, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("json.parse_lazy requires string argument, got=%s", args[0].Type())
				}

				// Validate that the JSON is a valid object (starts with {)
				trimmed := strings.TrimSpace(jsonStr.Value)
				emptyLazyMap := &object.LazyMap{
					RawJSON:     "{}",
					ParsedKeys:  make(map[string]object.Object),
					AllKeys:     []string{},
					FullyParsed: true,
				}

				if len(trimmed) == 0 || trimmed[0] != '{' {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     "json.parse_lazy requires JSON object (starts with {)",
								Context: make(map[string]object.Object),
							},
							emptyLazyMap,
						},
					}
				}

				// Extract top-level keys using json.Decoder token parsing
				keys, err := extractTopLevelKeys(trimmed)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							emptyLazyMap,
						},
					}
				}

				// Create lazy map with raw JSON and empty cache
				lazyMap := &object.LazyMap{
					RawJSON:     trimmed,
					ParsedKeys:  make(map[string]object.Object),
					AllKeys:     keys,
					FullyParsed: false,
				}

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						lazyMap,
					},
				}
			},
		},
	}
}

// InitLazy initializes lazy map operations
func InitLazy() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"lazy.get": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("lazy.get requires 2 arguments (lazy_map, key), got=%d", len(args))
				}

				lazyMap, ok := args[0].(*object.LazyMap)
				if !ok {
					return helpers.NewError("lazy.get requires lazy_map argument, got=%s", args[0].Type())
				}

				key, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("lazy.get requires string key, got=%s", args[1].Type())
				}

				// Check if field is already parsed
				if value, exists := lazyMap.ParsedKeys[key.Value]; exists {
					return value
				}

				// Check if key exists in the JSON
				keyExists := false
				for _, k := range lazyMap.AllKeys {
					if k == key.Value {
						keyExists = true
						break
					}
				}
				if !keyExists {
					return &object.Null{}
				}

				// Parse the specific field on demand
				value, err := parseFieldFromJSON(lazyMap.RawJSON, key.Value)
				if err != nil {
					return helpers.NewError("lazy.get failed to parse field '%s': %s", key.Value, err.Error())
				}

				// Cache the parsed value
				lazyMap.ParsedKeys[key.Value] = value

				// Check if all fields are now parsed
				if len(lazyMap.ParsedKeys) == len(lazyMap.AllKeys) {
					lazyMap.FullyParsed = true
				}

				return value
			},
		},

		"lazy.keys": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("lazy.keys requires 1 argument (lazy_map), got=%d", len(args))
				}

				lazyMap, ok := args[0].(*object.LazyMap)
				if !ok {
					return helpers.NewError("lazy.keys requires lazy_map argument, got=%s", args[0].Type())
				}

				// Return array of all keys
				elements := make([]object.Object, len(lazyMap.AllKeys))
				for i, key := range lazyMap.AllKeys {
					elements[i] = &object.String{Value: key}
				}

				return &object.Array{Elements: elements}
			},
		},

		"lazy.materialize": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("lazy.materialize requires 1 argument (lazy_map), got=%d", len(args))
				}

				lazyMap, ok := args[0].(*object.LazyMap)
				if !ok {
					return helpers.NewError("lazy.materialize requires lazy_map argument, got=%s", args[0].Type())
				}

				// Parse directly to Bark objects (memory-efficient, skips interface{})
				barkObj, err := parseJSONDirect([]byte(lazyMap.RawJSON))
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						barkObj,
					},
				}
			},
		},
	}
}

// parseFieldFromJSON parses a specific field from a JSON object string
func parseFieldFromJSON(jsonStr, fieldName string) (object.Object, error) {
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	interner := newStringInterner()

	// Expect opening brace
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := token.(json.Delim); !ok || delim != '{' {
		return nil, errors.New("expected JSON object")
	}

	// Iterate through key-value pairs
	for decoder.More() {
		// Get key
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		key, ok := keyToken.(string)
		if !ok {
			continue
		}

		if key == fieldName {
			// Found our field - parse directly to Bark object
			return parseJSONTokens(decoder, 0, interner)
		}

		// Skip this value by decoding and discarding
		var skip interface{}
		if err := decoder.Decode(&skip); err != nil {
			return nil, err
		}
	}

	return &object.Null{}, nil
}

// extractTopLevelKeys extracts all top-level keys from a JSON object without fully parsing values
func extractTopLevelKeys(jsonStr string) ([]string, error) {
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	keys := []string{}

	// Expect opening brace
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := token.(json.Delim); !ok || delim != '{' {
		return nil, errors.New("expected JSON object")
	}

	// Read key-value pairs until we hit the closing brace
	for decoder.More() {
		// Read key
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		key, ok := keyToken.(string)
		if !ok {
			return nil, errors.New("expected string key in JSON object")
		}

		keys = append(keys, key)

		// Skip the value using Decode (handles nested objects/arrays correctly)
		var skip interface{}
		if err := decoder.Decode(&skip); err != nil {
			return nil, err
		}
	}

	return keys, nil
}

// stringInterner reuses string allocations for repeated keys
type stringInterner struct {
	cache map[string]string
}

func newStringInterner() *stringInterner {
	return &stringInterner{cache: make(map[string]string)}
}

func (si *stringInterner) intern(s string) string {
	if cached, ok := si.cache[s]; ok {
		return cached
	}
	si.cache[s] = s
	return s
}

// parseJSONTokens parses JSON directly to Bark objects using token-based parsing.
// This avoids the intermediate interface{} representation for better memory efficiency.
func parseJSONTokens(decoder *json.Decoder, depth int, interner *stringInterner) (object.Object, error) {
	if depth > MaxJSONDepth {
		return nil, ErrJSONDepthExceeded
	}

	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}

	switch t := token.(type) {
	case json.Delim:
		switch t {
		case '{':
			// Parse object
			pairs := make(map[string]object.Object)
			keys := make([]string, 0)

			for decoder.More() {
				// Read key
				keyToken, err := decoder.Token()
				if err != nil {
					return nil, err
				}
				key, ok := keyToken.(string)
				if !ok {
					return nil, errors.New("expected string key in JSON object")
				}

				// Intern the key for memory efficiency
				internedKey := interner.intern(key)

				// Parse value recursively
				value, err := parseJSONTokens(decoder, depth+1, interner)
				if err != nil {
					return nil, err
				}

				pairs[internedKey] = value
				keys = append(keys, internedKey)
			}

			// Consume closing brace
			if _, err := decoder.Token(); err != nil {
				return nil, err
			}

			return &object.Map{Pairs: pairs, Keys: keys}, nil

		case '[':
			// Parse array
			elements := make([]object.Object, 0)

			for decoder.More() {
				elem, err := parseJSONTokens(decoder, depth+1, interner)
				if err != nil {
					return nil, err
				}
				elements = append(elements, elem)
			}

			// Consume closing bracket
			if _, err := decoder.Token(); err != nil {
				return nil, err
			}

			return &object.Array{Elements: elements}, nil
		}

	case string:
		return &object.String{Value: t}, nil

	case float64:
		// Convert to int if whole number
		if t == float64(int64(t)) {
			return &object.Integer{Value: int64(t)}, nil
		}
		return &object.Integer{Value: int64(t)}, nil

	case bool:
		return helpers.NativeBoolToBooleanObject(t), nil

	case nil:
		return &object.String{Value: ""}, nil
	}

	return &object.String{Value: ""}, nil
}

// parseJSONDirect parses a JSON string directly to Bark objects without intermediate interface{}.
func parseJSONDirect(jsonData []byte) (object.Object, error) {
	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	interner := newStringInterner()
	return parseJSONTokens(decoder, 0, interner)
}

// parseJSONFromReader parses JSON from an io.Reader directly to Bark objects.
func parseJSONFromReader(reader *bytes.Buffer) (object.Object, error) {
	decoder := json.NewDecoder(reader)
	interner := newStringInterner()
	return parseJSONTokens(decoder, 0, interner)
}

// barkToJSON converts bark objects to JSON-compatible interface{}
func barkToJSON(obj object.Object) interface{} {
	switch v := obj.(type) {
	case *object.Map:
		// bark map → JSON object (maintain key order)
		result := make(map[string]interface{})
		for _, key := range v.Keys {
			if value, ok := v.Pairs[key]; ok {
				result[key] = barkToJSON(value)
			}
		}
		return result

	case *object.Array:
		// bark array → JSON array
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
