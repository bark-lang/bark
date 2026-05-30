package evaluator

import (
	"sort"

	"gitlab.com/bark-lang/barki/object"
)

// InitFunctionalBuiltins creates higher-order array builtins with access to applyFunction
func InitFunctionalBuiltins() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"array.map":       {Fn: arrayMapFunc},
		"array.filter":    {Fn: arrayFilterFunc},
		"array.reduce":    {Fn: arrayReduceFunc},
		"array.sort":      {Fn: arraySortFunc},
		"array.sort_by":   {Fn: arraySortByFunc},
		"array.sum":       {Fn: arraySumFunc},
		"array.min":       {Fn: arrayMinFunc},
		"array.max":       {Fn: arrayMaxFunc},
		"array.min_by":    {Fn: arrayMinByFunc},
		"array.max_by":    {Fn: arrayMaxByFunc},
		"array.flatten":   {Fn: arrayFlattenFunc},
		"array.group_by":  {Fn: arrayGroupByFunc},
		"array.find":      {Fn: arrayFindFunc},
		"array.dedupe_by": {Fn: arrayDedupeByFunc},
		"map.map_keys":    {Fn: mapMapKeysFunc},
		"map.filter_keys": {Fn: mapFilterKeysFunc},
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

	fn := args[1]
	if !isCallable(fn) {
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

	fn := args[1]
	if !isCallable(fn) {
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

	fn := args[1]
	if !isCallable(fn) {
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

	fn := args[1]
	if !isCallable(fn) {
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

// array.sum returns the sum of numeric elements in an array
// Usage: [1, 2, 3] > array.sum()
func arraySumFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("array.sum requires 1 argument (array), got=%d", len(args))
	}

	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("array.sum requires array argument, got=%s", args[0].Type())
	}

	if len(arr.Elements) == 0 {
		return &object.Integer{Value: 0}
	}

	var intSum int64
	var floatSum float64
	isFloat := false

	for _, elem := range arr.Elements {
		switch v := elem.(type) {
		case *object.Integer:
			if isFloat {
				floatSum += float64(v.Value)
			} else {
				intSum += v.Value
			}
		case *object.Float:
			if !isFloat {
				floatSum = float64(intSum)
				isFloat = true
			}
			floatSum += v.Value
		default:
			return newError("array.sum requires numeric elements, got=%s", elem.Type())
		}
	}

	if isFloat {
		return &object.Float{Value: floatSum}
	}
	return &object.Integer{Value: intSum}
}

// array.min returns the minimum element in an array
// Usage: [3, 1, 2] > array.min()
func arrayMinFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("array.min requires 1 argument (array), got=%d", len(args))
	}

	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("array.min requires array argument, got=%s", args[0].Type())
	}

	if len(arr.Elements) == 0 {
		return newError("array.min requires non-empty array")
	}

	min := arr.Elements[0]
	for _, elem := range arr.Elements[1:] {
		cmp, err := compareObjects(elem, min)
		if err != nil {
			return err
		}
		if cmp < 0 {
			min = elem
		}
	}

	return min
}

// array.max returns the maximum element in an array
// Usage: [3, 1, 2] > array.max()
func arrayMaxFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("array.max requires 1 argument (array), got=%d", len(args))
	}

	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("array.max requires array argument, got=%s", args[0].Type())
	}

	if len(arr.Elements) == 0 {
		return newError("array.max requires non-empty array")
	}

	max := arr.Elements[0]
	for _, elem := range arr.Elements[1:] {
		cmp, err := compareObjects(elem, max)
		if err != nil {
			return err
		}
		if cmp > 0 {
			max = elem
		}
	}

	return max
}

// array.min_by returns the element with the minimum derived key
// Usage: users > array.min_by((u map) { u > get("age") }(int))
func arrayMinByFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("array.min_by requires 2 arguments (array, function), got=%d", len(args))
	}

	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("array.min_by requires array as first argument, got=%s", args[0].Type())
	}

	fn := args[1]
	if !isCallable(fn) {
		return newError("array.min_by requires function as second argument, got=%s", args[1].Type())
	}

	if len(arr.Elements) == 0 {
		return newError("array.min_by requires non-empty array")
	}

	minElem := arr.Elements[0]
	minKey := applyFunction(fn, []object.Object{minElem})
	if isError(minKey) {
		return minKey
	}
	if isExecutionError(minKey) {
		return minKey
	}

	for _, elem := range arr.Elements[1:] {
		key := applyFunction(fn, []object.Object{elem})
		if isError(key) {
			return key
		}
		if isExecutionError(key) {
			return key
		}
		cmp, err := compareObjects(key, minKey)
		if err != nil {
			return err
		}
		if cmp < 0 {
			minElem = elem
			minKey = key
		}
	}

	return minElem
}

// array.max_by returns the element with the maximum derived key
// Usage: users > array.max_by((u map) { u > get("age") }(int))
func arrayMaxByFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("array.max_by requires 2 arguments (array, function), got=%d", len(args))
	}

	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("array.max_by requires array as first argument, got=%s", args[0].Type())
	}

	fn := args[1]
	if !isCallable(fn) {
		return newError("array.max_by requires function as second argument, got=%s", args[1].Type())
	}

	if len(arr.Elements) == 0 {
		return newError("array.max_by requires non-empty array")
	}

	maxElem := arr.Elements[0]
	maxKey := applyFunction(fn, []object.Object{maxElem})
	if isError(maxKey) {
		return maxKey
	}
	if isExecutionError(maxKey) {
		return maxKey
	}

	for _, elem := range arr.Elements[1:] {
		key := applyFunction(fn, []object.Object{elem})
		if isError(key) {
			return key
		}
		if isExecutionError(key) {
			return key
		}
		cmp, err := compareObjects(key, maxKey)
		if err != nil {
			return err
		}
		if cmp > 0 {
			maxElem = elem
			maxKey = key
		}
	}

	return maxElem
}

// array.flatten flattens nested arrays. Optional depth argument limits depth.
// Usage: [[1, 2], [3, [4, 5]]] > array.flatten()
// Usage: [[1, [2]], [3, [4, [5]]]] > array.flatten(1)
func arrayFlattenFunc(args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("array.flatten requires 1-2 arguments (array[, depth]), got=%d", len(args))
	}

	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("array.flatten requires array as first argument, got=%s", args[0].Type())
	}

	maxDepth := int64(-1) // -1 means unlimited
	if len(args) == 2 {
		depth, ok := args[1].(*object.Integer)
		if !ok {
			return newError("array.flatten depth must be integer, got=%s", args[1].Type())
		}
		if depth.Value < 0 {
			return newError("array.flatten depth must be non-negative, got=%d", depth.Value)
		}
		maxDepth = depth.Value
	}

	result := flattenArray(arr, maxDepth, 0)
	return &object.Array{Elements: result}
}

func flattenArray(arr *object.Array, maxDepth int64, currentDepth int64) []object.Object {
	result := make([]object.Object, 0, len(arr.Elements))
	for _, elem := range arr.Elements {
		if inner, ok := elem.(*object.Array); ok && (maxDepth == -1 || currentDepth < maxDepth) {
			result = append(result, flattenArray(inner, maxDepth, currentDepth+1)...)
		} else {
			result = append(result, elem)
		}
	}
	return result
}

// array.group_by groups elements by a derived key, returns a map of arrays
// Usage: [1, 2, 3, 4, 5] > array.group_by((x int) { x > mod(2) > to_string() }(string))
func arrayGroupByFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("array.group_by requires 2 arguments (array, function), got=%d", len(args))
	}

	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("array.group_by requires array as first argument, got=%s", args[0].Type())
	}

	fn := args[1]
	if !isCallable(fn) {
		return newError("array.group_by requires function as second argument, got=%s", args[1].Type())
	}

	groups := make(map[string][]object.Object)
	order := make([]string, 0)

	for _, elem := range arr.Elements {
		key := applyFunction(fn, []object.Object{elem})
		if isError(key) {
			return key
		}
		if isExecutionError(key) {
			return key
		}

		keyStr, ok := key.(*object.String)
		if !ok {
			return newError("array.group_by key function must return string, got=%s", key.Type())
		}

		if _, exists := groups[keyStr.Value]; !exists {
			order = append(order, keyStr.Value)
		}
		groups[keyStr.Value] = append(groups[keyStr.Value], elem)
	}

	pairs := make(map[string]object.Object)
	for _, k := range order {
		pairs[k] = &object.Array{Elements: groups[k]}
	}

	return &object.Map{Pairs: pairs, Keys: order}
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

// array.find returns the first element matching a predicate, or absent if none found
// Usage: [1, 2, 3, 4, 5] > array.find((x int) { x > gt?(3) }(bool))
func arrayFindFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("array.find requires 2 arguments (array, function), got=%d", len(args))
	}

	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("array.find requires array as first argument, got=%s", args[0].Type())
	}

	fn := args[1]
	if !isCallable(fn) {
		return newError("array.find requires function as second argument, got=%s", args[1].Type())
	}

	for _, elem := range arr.Elements {
		val := applyFunction(fn, []object.Object{elem})
		if isError(val) {
			return val
		}
		if isExecutionError(val) {
			return val
		}
		if isTruthyValue(val) {
			return elem
		}
	}

	return &object.Null{}
}

// array.dedupe_by deduplicates elements by a derived key function
// Usage: users > array.dedupe_by((u map) { u > get("id") }(int))
func arrayDedupeByFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("array.dedupe_by requires 2 arguments (array, function), got=%d", len(args))
	}

	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("array.dedupe_by requires array as first argument, got=%s", args[0].Type())
	}

	fn := args[1]
	if !isCallable(fn) {
		return newError("array.dedupe_by requires function as second argument, got=%s", args[1].Type())
	}

	seen := make(map[string]bool)
	result := make([]object.Object, 0)

	for _, elem := range arr.Elements {
		key := applyFunction(fn, []object.Object{elem})
		if isError(key) {
			return key
		}
		if isExecutionError(key) {
			return key
		}
		keyStr := key.Inspect()
		if !seen[keyStr] {
			seen[keyStr] = true
			result = append(result, elem)
		}
	}

	return &object.Array{Elements: result}
}

// map.map_keys transforms all keys using a callback function, returns new map
// Usage: {"a": 1, "b": 2} > map.map_keys((k string) { k > str.upper() }(string))
func mapMapKeysFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("map.map_keys requires 2 arguments (map, function), got=%d", len(args))
	}

	m, ok := args[0].(*object.Map)
	if !ok {
		return newError("map.map_keys requires map as first argument, got=%s", args[0].Type())
	}

	fn := args[1]
	if !isCallable(fn) {
		return newError("map.map_keys requires function as second argument, got=%s", args[1].Type())
	}

	newPairs := make(map[string]object.Object)
	newKeys := make([]string, 0, len(m.Keys))
	keyExists := make(map[string]bool)

	for _, k := range m.Keys {
		val := applyFunction(fn, []object.Object{&object.String{Value: k}})
		if isError(val) {
			return val
		}
		if isExecutionError(val) {
			return val
		}

		newKey, ok := val.(*object.String)
		if !ok {
			return newError("map.map_keys callback must return string, got=%s", val.Type())
		}

		newPairs[newKey.Value] = m.Pairs[k]
		if !keyExists[newKey.Value] {
			newKeys = append(newKeys, newKey.Value)
			keyExists[newKey.Value] = true
		}
	}

	return &object.Map{Pairs: newPairs, Keys: newKeys}
}

// map.filter_keys keeps entries where the callback returns truthy for the key
// Usage: {"name": "Alice", "_id": 123} > map.filter_keys((k string) { k > starts_with?("_") > not() }(bool))
func mapFilterKeysFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("map.filter_keys requires 2 arguments (map, function), got=%d", len(args))
	}

	m, ok := args[0].(*object.Map)
	if !ok {
		return newError("map.filter_keys requires map as first argument, got=%s", args[0].Type())
	}

	fn := args[1]
	if !isCallable(fn) {
		return newError("map.filter_keys requires function as second argument, got=%s", args[1].Type())
	}

	newPairs := make(map[string]object.Object)
	newKeys := make([]string, 0)

	for _, k := range m.Keys {
		val := applyFunction(fn, []object.Object{&object.String{Value: k}})
		if isError(val) {
			return val
		}
		if isExecutionError(val) {
			return val
		}
		if isTruthyValue(val) {
			newPairs[k] = m.Pairs[k]
			newKeys = append(newKeys, k)
		}
	}

	return &object.Map{Pairs: newPairs, Keys: newKeys}
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
