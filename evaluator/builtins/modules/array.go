package modules

import (
	"gitlab.com/bark-lang/barki/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/barki/object"
)

// InitArray initializes array operations
func InitArray() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"array.dedupe": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("array.dedupe requires 1 argument, got=%d", len(args))
				}

				arr, ok := args[0].(*object.Array)
				if !ok {
					return helpers.NewError("array.dedupe requires array argument, got=%s", args[0].Type())
				}

				// Use a map to track seen values
				seen := make(map[string]bool)
				result := []object.Object{}

				for _, elem := range arr.Elements {
					key := elem.Inspect()
					if !seen[key] {
						seen[key] = true
						result = append(result, elem)
					}
				}

				return &object.Array{Elements: result}
			},
		},

		"array.push": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 2 {
					return helpers.NewError("array.push requires at least 2 arguments (array, value...), got=%d", len(args))
				}

				arr, ok := args[0].(*object.Array)
				if !ok {
					return helpers.NewError("array.push requires array as first argument, got=%s", args[0].Type())
				}

				// Use COW: if array is not shared, modify in place
				result := arr
				for i := 1; i < len(args); i++ {
					result = result.COWPush(args[i])
				}
				return result
			},
		},

		"array.append_to": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("array.append_to requires 2 arguments (value, array), got=%d", len(args))
				}

				arr, ok := args[1].(*object.Array)
				if !ok {
					return helpers.NewError("array.append_to requires array as second argument, got=%s", args[1].Type())
				}

				// Use COW: if array is not shared, modify in place
				return arr.COWPush(args[0])
			},
		},

		"array.pop": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("array.pop requires 1 argument (array), got=%d", len(args))
				}

				arr, ok := args[0].(*object.Array)
				if !ok {
					return helpers.NewError("array.pop requires array argument, got=%s", args[0].Type())
				}

				if len(arr.Elements) == 0 {
					return helpers.NewError("cannot pop from empty array")
				}

				// Use COW: if array is not shared, modify in place
				newArr, popped := arr.COWPop()

				// Return (modified_array, element)
				return &object.Array{Elements: []object.Object{newArr, popped}}
			},
		},

		"array.shift": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("array.shift requires 1 argument (array), got=%d", len(args))
				}

				arr, ok := args[0].(*object.Array)
				if !ok {
					return helpers.NewError("array.shift requires array argument, got=%s", args[0].Type())
				}

				if len(arr.Elements) == 0 {
					return helpers.NewError("cannot shift from empty array")
				}

				// Use COW: if array is not shared, modify in place
				newArr, shifted := arr.COWShift()

				// Return (modified_array, element)
				return &object.Array{Elements: []object.Object{newArr, shifted}}
			},
		},

		"array.unshift": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 2 {
					return helpers.NewError("array.unshift requires at least 2 arguments (array, value...), got=%d", len(args))
				}

				arr, ok := args[0].(*object.Array)
				if !ok {
					return helpers.NewError("array.unshift requires array as first argument, got=%s", args[0].Type())
				}

				// Use COW: if array is not shared, modify in place
				return arr.COWUnshift(args[1:]...)
			},
		},

		"array.slice": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return helpers.NewError("array.slice requires 3 arguments (array, start, end), got=%d", len(args))
				}

				arr, ok := args[0].(*object.Array)
				if !ok {
					return helpers.NewError("array.slice requires array as first argument, got=%s", args[0].Type())
				}

				start, ok := args[1].(*object.Integer)
				if !ok {
					return helpers.NewError("array.slice requires integer start index, got=%s", args[1].Type())
				}

				end, ok := args[2].(*object.Integer)
				if !ok {
					return helpers.NewError("array.slice requires integer end index, got=%s", args[2].Type())
				}

				// Bounds checking
				if start.Value < 0 || start.Value > int64(len(arr.Elements)) {
					return helpers.NewError("slice start index out of bounds: %d", start.Value)
				}

				if end.Value < start.Value || end.Value > int64(len(arr.Elements)) {
					return helpers.NewError("slice end index out of bounds: %d", end.Value)
				}

				// Use COWSlice (currently always copies to avoid aliasing)
				return arr.COWSlice(int(start.Value), int(end.Value))
			},
		},

		"array.range": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("array.range requires 2 arguments (start, end), got=%d", len(args))
				}

				start, ok := args[0].(*object.Integer)
				if !ok {
					return helpers.NewError("array.range requires integer start, got=%s", args[0].Type())
				}

				end, ok := args[1].(*object.Integer)
				if !ok {
					return helpers.NewError("array.range requires integer end, got=%s", args[1].Type())
				}

				if end.Value < start.Value {
					return helpers.NewError("range end (%d) must be >= start (%d)", end.Value, start.Value)
				}

				// Create array [start, start+1, ..., end] (inclusive)
				size := end.Value - start.Value + 1
				elements := make([]object.Object, size)
				for i := int64(0); i < size; i++ {
					elements[i] = &object.Integer{Value: start.Value + i}
				}

				return &object.Array{Elements: elements}
			},
		},

		"array.zip": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("array.zip requires 2 arguments (array, array), got=%d", len(args))
				}

				arr1, ok := args[0].(*object.Array)
				if !ok {
					return helpers.NewError("array.zip requires array as first argument, got=%s", args[0].Type())
				}

				arr2, ok := args[1].(*object.Array)
				if !ok {
					return helpers.NewError("array.zip requires array as second argument, got=%s", args[1].Type())
				}

				length := len(arr1.Elements)
				if len(arr2.Elements) < length {
					length = len(arr2.Elements)
				}

				result := make([]object.Object, length)
				for i := 0; i < length; i++ {
					result[i] = &object.Array{Elements: []object.Object{arr1.Elements[i], arr2.Elements[i]}}
				}

				return &object.Array{Elements: result}
			},
		},
	}
}
