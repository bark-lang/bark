package evaluator

import (
	"sort"

	"gitlab.com/bark-lang/barki/object"
)

// InitFunctionalBuiltins creates higher-order array builtins with access to applyFunction
func InitFunctionalBuiltins() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"array.map":     {Fn: arrayMapFunc},
		"array.filter":  {Fn: arrayFilterFunc},
		"array.reduce":  {Fn: arrayReduceFunc},
		"array.sort":    {Fn: arraySortFunc},
		"array.sort_by": {Fn: arraySortByFunc},
	}
}

// array.map transforms each element by calling fn on it, returns new array
// Usage: [1, 2, 3] > array.map((x int) { x > mul(2) }(int))
func arrayMapFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("array.map requires 2 arguments (array, function), got=%d", len(args))
	}

	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("array.map requires array as first argument, got=%s", args[0].Type())
	}

	fn, ok := args[1].(*object.Function)
	if !ok {
		return newError("array.map requires function as second argument, got=%s", args[1].Type())
	}

	result := make([]object.Object, 0, len(arr.Elements))
	for _, elem := range arr.Elements {
		val := applyFunction(fn, []object.Object{elem})
		if isError(val) {
			return val
		}
		if isExecutionError(val) {
			return val
		}
		result = append(result, val)
	}

	return &object.Array{Elements: result}
}

// array.filter keeps elements where fn returns true, returns new array
// Usage: [1, 2, 3, 4, 5] > array.filter((x int) { x > gt?(3) }(bool))
func arrayFilterFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("array.filter requires 2 arguments (array, function), got=%d", len(args))
	}

	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("array.filter requires array as first argument, got=%s", args[0].Type())
	}

	fn, ok := args[1].(*object.Function)
	if !ok {
		return newError("array.filter requires function as second argument, got=%s", args[1].Type())
	}

	result := make([]object.Object, 0)
	for _, elem := range arr.Elements {
		val := applyFunction(fn, []object.Object{elem})
		if isError(val) {
			return val
		}
		if isExecutionError(val) {
			return val
		}
		if isTruthyValue(val) {
			result = append(result, elem)
		}
	}

	return &object.Array{Elements: result}
}

// array.reduce accumulates array into a single value
// Usage: [1, 2, 3, 4] > array.reduce((acc int, x int) { acc > add(x) }(int), 0)
func arrayReduceFunc(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("array.reduce requires 3 arguments (array, function, initial), got=%d", len(args))
	}

	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("array.reduce requires array as first argument, got=%s", args[0].Type())
	}

	fn, ok := args[1].(*object.Function)
	if !ok {
		return newError("array.reduce requires function as second argument, got=%s", args[1].Type())
	}

	acc := args[2]
	for _, elem := range arr.Elements {
		acc = applyFunction(fn, []object.Object{acc, elem})
		if isError(acc) {
			return acc
		}
		if isExecutionError(acc) {
			return acc
		}
	}

	return acc
}

// array.sort sorts an array using natural ordering, returns new array
// Usage: [3, 1, 2] > array.sort()
func arraySortFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("array.sort requires 1 argument (array), got=%d", len(args))
	}

	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("array.sort requires array argument, got=%s", args[0].Type())
	}

	// Copy elements
	sorted := make([]object.Object, len(arr.Elements))
	copy(sorted, arr.Elements)

	var sortErr object.Object
	sort.SliceStable(sorted, func(i, j int) bool {
		if sortErr != nil {
			return false
		}
		cmp, err := compareObjects(sorted[i], sorted[j])
		if err != nil {
			sortErr = err
			return false
		}
		return cmp < 0
	})

	if sortErr != nil {
		return sortErr
	}

	return &object.Array{Elements: sorted}
}

// array.sort_by sorts an array by a derived key, returns new array
// Usage: users > array.sort_by((u map) { u > get("name") }(string))
func arraySortByFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("array.sort_by requires 2 arguments (array, function), got=%d", len(args))
	}

	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("array.sort_by requires array as first argument, got=%s", args[0].Type())
	}

	fn, ok := args[1].(*object.Function)
	if !ok {
		return newError("array.sort_by requires function as second argument, got=%s", args[1].Type())
	}

	// Pre-compute keys
	keys := make([]object.Object, len(arr.Elements))
	for i, elem := range arr.Elements {
		key := applyFunction(fn, []object.Object{elem})
		if isError(key) {
			return key
		}
		if isExecutionError(key) {
			return key
		}
		keys[i] = key
	}

	// Build index array and sort by keys
	indices := make([]int, len(arr.Elements))
	for i := range indices {
		indices[i] = i
	}

	var sortErr object.Object
	sort.SliceStable(indices, func(i, j int) bool {
		if sortErr != nil {
			return false
		}
		cmp, err := compareObjects(keys[indices[i]], keys[indices[j]])
		if err != nil {
			sortErr = err
			return false
		}
		return cmp < 0
	})

	if sortErr != nil {
		return sortErr
	}

	// Build result in sorted order
	result := make([]object.Object, len(arr.Elements))
	for i, idx := range indices {
		result[i] = arr.Elements[idx]
	}

	return &object.Array{Elements: result}
}

// isTruthyValue determines if a value is truthy for filter predicates.
// Matches the semantics of present?() — booleans by value, others by non-emptiness.
func isTruthyValue(obj object.Object) bool {
	switch v := obj.(type) {
	case *object.Boolean:
		return v.Value
	case *object.Null:
		return false
	case *object.Integer:
		return v.Value != 0
	case *object.Float:
		return v.Value != 0.0
	case *object.String:
		return v.Value != ""
	case *object.Array:
		return len(v.Elements) > 0
	case *object.Map:
		return len(v.Pairs) > 0
	default:
		return true
	}
}

// compareObjects compares two objects for sorting.
// Returns -1, 0, or 1. Returns an error for incomparable types.
func compareObjects(a, b object.Object) (int, object.Object) {
	switch av := a.(type) {
	case *object.Integer:
		switch bv := b.(type) {
		case *object.Integer:
			if av.Value < bv.Value {
				return -1, nil
			}
			if av.Value > bv.Value {
				return 1, nil
			}
			return 0, nil
		case *object.Float:
			af := float64(av.Value)
			if af < bv.Value {
				return -1, nil
			}
			if af > bv.Value {
				return 1, nil
			}
			return 0, nil
		}
	case *object.Float:
		switch bv := b.(type) {
		case *object.Float:
			if av.Value < bv.Value {
				return -1, nil
			}
			if av.Value > bv.Value {
				return 1, nil
			}
			return 0, nil
		case *object.Integer:
			bf := float64(bv.Value)
			if av.Value < bf {
				return -1, nil
			}
			if av.Value > bf {
				return 1, nil
			}
			return 0, nil
		}
	case *object.String:
		if bv, ok := b.(*object.String); ok {
			if av.Value < bv.Value {
				return -1, nil
			}
			if av.Value > bv.Value {
				return 1, nil
			}
			return 0, nil
		}
	}

	return 0, newError("cannot compare %s and %s", a.Type(), b.Type())
}
