package helpers

import (
	"fmt"
	"strings"

	"gitlab.com/bark-lang/bark/object"
)

// Singleton boolean objects for efficiency
var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

// NativeBoolToBooleanObject converts a Go boolean to a bark Boolean object.
// Uses singleton objects for efficiency.
func NativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

// NewError creates a new programming error with the given format string and arguments.
// Programming errors (wrong number of args, type errors, etc.) stop execution immediately.
func NewError(format string, a ...interface{}) *object.Error {
	return &object.Error{
		Msg:                fmt.Sprintf(format, a...),
		Context:            make(map[string]object.Object),
		IsProgrammingError: true,
	}
}

// WrapError creates a new programming error from a Go error.
// This is useful for wrapping errors from Go standard library calls.
func WrapError(err error) *object.Error {
	return &object.Error{
		Msg:                err.Error(),
		Context:            make(map[string]object.Object),
		IsProgrammingError: true,
	}
}

// NewExecutionError creates a new execution error (recoverable, stops chain but doesn't crash).
// The evaluator will enrich this with source location info before logging.
func NewExecutionError(message, detail string) *object.ExecutionError {
	return &object.ExecutionError{
		Message: message,
		Detail:  detail,
	}
}

// FormatString formats a string with indexed placeholders {0}, {1}, etc.
// If formatArgs is empty, the format string is returned as-is.
// Escape sequences {{ and }} produce literal { and }.
// Returns the formatted string and any error encountered.
func FormatString(format string, formatArgs []object.Object) (string, error) {
	var result strings.Builder
	input := format
	i := 0

	for i < len(input) {
		// Check for escape sequences {{ and }}
		if i+1 < len(input) {
			if input[i] == '{' && input[i+1] == '{' {
				result.WriteByte('{')
				i += 2
				continue
			}
			if input[i] == '}' && input[i+1] == '}' {
				result.WriteByte('}')
				i += 2
				continue
			}
		}

		// Check for placeholder {n}
		if input[i] == '{' {
			// Find closing brace
			end := i + 1
			for end < len(input) && input[end] != '}' {
				end++
			}

			if end >= len(input) {
				return "", fmt.Errorf("unclosed placeholder at position %d", i)
			}

			// Parse index between braces
			indexStr := input[i+1 : end]
			if indexStr == "" {
				return "", fmt.Errorf("empty placeholder at position %d: use {0}, {1}, etc", i)
			}

			// Parse the index number
			index := 0
			for _, ch := range indexStr {
				if ch < '0' || ch > '9' {
					return "", fmt.Errorf("invalid placeholder {%s}, must be numeric index", indexStr)
				}
				index = index*10 + int(ch-'0')
			}

			if index >= len(formatArgs) {
				return "", fmt.Errorf("placeholder {%d} out of range, only %d arguments provided", index, len(formatArgs))
			}

			// Convert argument to string
			arg := formatArgs[index]
			switch v := arg.(type) {
			case *object.String:
				result.WriteString(v.Value)
			case *object.Integer:
				result.WriteString(v.Inspect())
			case *object.Float:
				result.WriteString(v.Inspect())
			case *object.Boolean:
				result.WriteString(v.Inspect())
			default:
				result.WriteString(arg.Inspect())
			}

			i = end + 1
			continue
		}

		// Check for unescaped closing brace
		if input[i] == '}' {
			return "", fmt.Errorf("unescaped '}' at position %d, use '}}' for literal", i)
		}

		// Regular character
		result.WriteByte(input[i])
		i++
	}

	return result.String(), nil
}

// InterpolateString processes string interpolation with {identifier} and {identifier.field} patterns.
// It looks up variables in the environment and replaces the placeholders with their values.
// Escaped braces \{ and \} are converted to literal { and }.
// Numeric placeholders like {0}, {1} are passed through unchanged for later processing.
// Non-identifier content in braces (like JSON) is passed through unchanged.
// Returns the interpolated string and any error encountered.
func InterpolateString(input string, env *object.Environment) (string, error) {
	var result strings.Builder
	i := 0

	for i < len(input) {
		// Check for escaped braces from lexer (\{ and \})
		if i+1 < len(input) && input[i] == '\\' {
			if input[i+1] == '{' {
				result.WriteByte('{')
				i += 2
				continue
			}
			if input[i+1] == '}' {
				result.WriteByte('}')
				i += 2
				continue
			}
		}

		// Check for interpolation placeholder {identifier} or {identifier.field}
		if input[i] == '{' {
			// Find closing brace
			end := i + 1
			for end < len(input) && input[end] != '}' {
				end++
			}

			if end >= len(input) {
				// No closing brace found - pass through unchanged
				result.WriteByte(input[i])
				i++
				continue
			}

			// Get the content between braces
			content := input[i+1 : end]

			// Empty braces {} - pass through unchanged (used in JSON, etc.)
			if content == "" {
				result.WriteString("{}")
				i = end + 1
				continue
			}

			// Check if it's a numeric index (positional placeholder) - pass through unchanged
			if isNumeric(content) {
				result.WriteString(input[i : end+1])
				i = end + 1
				continue
			}

			// Check if content looks like a valid identifier or identifier.field
			// If not, pass through unchanged (e.g., JSON content like {"name": "value"})
			var identifier, field string
			if dotIdx := strings.Index(content, "."); dotIdx != -1 {
				identifier = content[:dotIdx]
				field = content[dotIdx+1:]
				// If either part is empty or invalid, pass through unchanged
				if identifier == "" || field == "" || !isValidIdentifier(identifier) || !isValidIdentifier(field) || strings.Contains(field, ".") {
					result.WriteString(input[i : end+1])
					i = end + 1
					continue
				}
			} else {
				identifier = content
				// If not a valid identifier, pass through unchanged
				if !isValidIdentifier(identifier) {
					result.WriteString(input[i : end+1])
					i = end + 1
					continue
				}
			}

			// At this point we have a valid identifier pattern - try to look it up
			val, ok := env.Get(identifier)
			if !ok {
				// Variable not found - pass through unchanged (allows {0}, {1} style to work later)
				result.WriteString(input[i : end+1])
				i = end + 1
				continue
			}

			// If field access is requested, look up the field in the map
			if field != "" {
				mapVal, ok := val.(*object.Map)
				if !ok {
					// Not a map - pass through unchanged
					result.WriteString(input[i : end+1])
					i = end + 1
					continue
				}
				fieldVal, ok := mapVal.Pairs[field]
				if !ok {
					// Field not found - pass through unchanged
					result.WriteString(input[i : end+1])
					i = end + 1
					continue
				}
				val = fieldVal
			}

			// Convert value to string
			result.WriteString(objectToString(val))
			i = end + 1
			continue
		}

		// Regular character
		result.WriteByte(input[i])
		i++
	}

	return result.String(), nil
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// isValidIdentifier checks if a string is a valid bark identifier
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, ch := range s {
		if i == 0 {
			// First character must be letter or underscore
			if !isLetter(ch) {
				return false
			}
		} else {
			// Subsequent characters can be letter, digit, underscore, or ?
			if !isLetter(ch) && !isDigit(ch) {
				return false
			}
		}
	}
	return true
}

// isLetter checks if a rune is a letter or underscore or question mark
func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == '?'
}

// isDigit checks if a rune is a digit
func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

// objectToString converts an object to its string representation for interpolation
func objectToString(obj object.Object) string {
	switch v := obj.(type) {
	case *object.String:
		return v.Value
	default:
		return obj.Inspect()
	}
}

// ObjectsEqual compares two bark objects for equality without allocating strings.
// For primitive types (Integer, Float, String, Boolean), it compares values directly.
// For complex types, it falls back to Inspect() comparison.
func ObjectsEqual(a, b object.Object) bool {
	// Fast path: same object reference
	if a == b {
		return true
	}

	// Different types are never equal
	if a.Type() != b.Type() {
		return false
	}

	// Compare by type without string allocation
	switch av := a.(type) {
	case *object.Integer:
		return av.Value == b.(*object.Integer).Value
	case *object.Float:
		return av.Value == b.(*object.Float).Value
	case *object.String:
		return av.Value == b.(*object.String).Value
	case *object.Boolean:
		return av.Value == b.(*object.Boolean).Value
	case *object.Null:
		return true // Both are null
	default:
		// Fall back to Inspect() for complex types (arrays, maps, etc.)
		return a.Inspect() == b.Inspect()
	}
}
