package evaluator

import (
	"context"
	"sync"

	"gitlab.com/bark-lang/barki/object"
)

// InitParallelBuiltins creates parallel processing builtins with access to applyFunction
func InitParallelBuiltins() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"parallel":         {Fn: parallelFunc},
		"parallel_all":     {Fn: parallelAllFunc},
		"parallel_limited": {Fn: parallelLimitedFunc},
		"parallel_race":    {Fn: parallelRaceFunc},
		"parallel_strict":  {Fn: parallelStrictFunc},
		"all_absent?":      {Fn: allAbsentFunc},
		"all_present?":     {Fn: allPresentFunc},
		"first_error":      {Fn: firstErrorFunc},
	}
}

// parallel processes array elements concurrently, applying a function to each element
// Returns (errors array, results array) where order is preserved
func parallelFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("parallel() requires 2 arguments: collection and function")
	}

	// Get collection array
	collection, ok := args[0].(*object.Array)
	if !ok {
		return newError("parallel() first argument must be array, got %s", args[0].Type())
	}

	// Get function
	fn, ok := args[1].(*object.Function)
	if !ok {
		return newError("parallel() second argument must be function, got %s", args[1].Type())
	}

	elements := collection.Elements
	resultCount := len(elements)

	// Pre-allocate result arrays
	errors := make([]object.Object, resultCount)
	results := make([]object.Object, resultCount)

	var wg sync.WaitGroup
	wg.Add(resultCount)

	// Process each element in parallel
	for i, elem := range elements {
		go func(index int, element object.Object) {
			defer wg.Done()

			// Call function with element
			result := applyFunction(fn, []object.Object{element})

			// Function should return a tuple (error, result)
			if tuple, ok := result.(*object.Tuple); ok && len(tuple.Elements) == 2 {
				errors[index] = tuple.Elements[0]
				results[index] = tuple.Elements[1]
			} else if tuple, ok := result.(*object.Array); ok && len(tuple.Elements) == 2 {
				// Also accept arrays for backward compatibility
				errors[index] = tuple.Elements[0]
				results[index] = tuple.Elements[1]
			} else if errObj, ok := result.(*object.Error); ok {
				// If function returned error directly, treat as error
				errors[index] = errObj
				results[index] = &object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}
			} else {
				// Invalid return type
				errors[index] = newError("parallel() function must return (error, result) tuple")
				results[index] = &object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}
			}
		}(i, elem)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Return (errors, results) as array tuple
	return &object.Array{
		Elements: []object.Object{
			&object.Array{Elements: errors},
			&object.Array{Elements: results},
		},
	}
}

// parallelAll executes an array of functions concurrently
// Returns (errors array, results array) where order is preserved
func parallelAllFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("parallel_all() requires 1 argument: array of functions")
	}

	// Get functions array
	functions, ok := args[0].(*object.Array)
	if !ok {
		return newError("parallel_all() argument must be array, got %s", args[0].Type())
	}

	elements := functions.Elements
	resultCount := len(elements)

	// Pre-allocate result arrays
	errors := make([]object.Object, resultCount)
	results := make([]object.Object, resultCount)

	var wg sync.WaitGroup
	wg.Add(resultCount)

	// Execute each function in parallel
	for i, elem := range elements {
		fn, ok := elem.(*object.Function)
		if !ok {
			wg.Done()
			errors[i] = newError("parallel_all() array element %d must be function, got %s", i, elem.Type())
			results[i] = &object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}
			continue
		}

		go func(index int, function *object.Function) {
			defer wg.Done()

			// Call function with a dummy parameter (0) since these are typically closures
			result := applyFunction(function, []object.Object{&object.Integer{Value: 0}})

			// Function should return a tuple (error, result)
			if tuple, ok := result.(*object.Tuple); ok && len(tuple.Elements) == 2 {
				errors[index] = tuple.Elements[0]
				results[index] = tuple.Elements[1]
			} else if tuple, ok := result.(*object.Array); ok && len(tuple.Elements) == 2 {
				// Also accept arrays for backward compatibility
				errors[index] = tuple.Elements[0]
				results[index] = tuple.Elements[1]
			} else if errObj, ok := result.(*object.Error); ok {
				errors[index] = errObj
				results[index] = &object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}
			} else {
				errors[index] = newError("parallel_all() function must return (error, result) tuple")
				results[index] = &object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}
			}
		}(i, fn)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Return (errors, results) as tuple
	return &object.Array{
		Elements: []object.Object{
			&object.Array{Elements: errors},
			&object.Array{Elements: results},
		},
	}
}

// parallelLimited processes collection with maximum concurrent operations (worker pool)
// Returns (errors array, results array) where order is preserved
func parallelLimitedFunc(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("parallel_limited() requires 3 arguments: limit, collection, and function")
	}

	// Get limit
	limitObj, ok := args[0].(*object.Integer)
	if !ok {
		return newError("parallel_limited() first argument must be integer, got %s", args[0].Type())
	}
	limit := int(limitObj.Value)
	if limit < 1 {
		return newError("parallel_limited() limit must be >= 1, got %d", limit)
	}

	// Get collection array
	collection, ok := args[1].(*object.Array)
	if !ok {
		return newError("parallel_limited() second argument must be array, got %s", args[1].Type())
	}

	// Get function
	fn, ok := args[2].(*object.Function)
	if !ok {
		return newError("parallel_limited() third argument must be function, got %s", args[2].Type())
	}

	elements := collection.Elements
	resultCount := len(elements)

	// Pre-allocate result arrays
	errors := make([]object.Object, resultCount)
	results := make([]object.Object, resultCount)

	// Create work channel
	type workItem struct {
		index   int
		element object.Object
	}
	workChan := make(chan workItem, resultCount)

	// Fill work channel
	for i, elem := range elements {
		workChan <- workItem{index: i, element: elem}
	}
	close(workChan)

	// Launch worker pool
	var wg sync.WaitGroup
	wg.Add(limit)

	for w := 0; w < limit; w++ {
		go func() {
			defer wg.Done()

			for work := range workChan {
				// Call function with element
				result := applyFunction(fn, []object.Object{work.element})

				// Function should return a tuple (error, result)
				if tuple, ok := result.(*object.Tuple); ok && len(tuple.Elements) == 2 {
					errors[work.index] = tuple.Elements[0]
					results[work.index] = tuple.Elements[1]
				} else if tuple, ok := result.(*object.Array); ok && len(tuple.Elements) == 2 {
					// Also accept arrays for backward compatibility
					errors[work.index] = tuple.Elements[0]
					results[work.index] = tuple.Elements[1]
				} else if errObj, ok := result.(*object.Error); ok {
					errors[work.index] = errObj
					results[work.index] = &object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}
				} else {
					errors[work.index] = newError("parallel_limited() function must return (error, result) tuple")
					results[work.index] = &object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}
				}
			}
		}()
	}

	// Wait for all workers to complete
	wg.Wait()

	// Return (errors, results) as tuple
	return &object.Array{
		Elements: []object.Object{
			&object.Array{Elements: errors},
			&object.Array{Elements: results},
		},
	}
}

// parallelRace executes functions concurrently and returns first to complete
// Returns (error, result) from the winning function
func parallelRaceFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("parallel_race() requires 1 argument: array of functions")
	}

	// Get functions array
	functions, ok := args[0].(*object.Array)
	if !ok {
		return newError("parallel_race() argument must be array, got %s", args[0].Type())
	}

	if len(functions.Elements) == 0 {
		return newError("parallel_race() requires at least one function")
	}

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to receive first result
	type raceResult struct {
		err    object.Object
		result object.Object
	}
	resultChan := make(chan raceResult, len(functions.Elements))

	// Launch all functions
	for i, elem := range functions.Elements {
		fn, ok := elem.(*object.Function)
		if !ok {
			// Return error immediately if any element is not a function
			return newError("parallel_race() array element %d must be function, got %s", i, elem.Type())
		}

		go func(function *object.Function) {
			// Check if already canceled
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Call function
			result := applyFunction(function, []object.Object{&object.Integer{Value: 0}})

			// Function should return a tuple (error, result)
			if tuple, ok := result.(*object.Tuple); ok && len(tuple.Elements) == 2 {
				select {
				case resultChan <- raceResult{err: tuple.Elements[0], result: tuple.Elements[1]}:
				case <-ctx.Done():
				}
			} else if tuple, ok := result.(*object.Array); ok && len(tuple.Elements) == 2 {
				// Also accept arrays for backward compatibility
				select {
				case resultChan <- raceResult{err: tuple.Elements[0], result: tuple.Elements[1]}:
				case <-ctx.Done():
				}
			} else if errObj, ok := result.(*object.Error); ok {
				select {
				case resultChan <- raceResult{err: errObj, result: &object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}}:
				case <-ctx.Done():
				}
			} else {
				select {
				case resultChan <- raceResult{
					err:    newError("parallel_race() function must return (error, result) tuple"),
					result: &object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
				}:
				case <-ctx.Done():
				}
			}
		}(fn)
	}

	// Wait for first result
	first := <-resultChan
	cancel() // Cancel remaining goroutines

	// Return (error, result) as tuple
	return &object.Array{
		Elements: []object.Object{first.err, first.result},
	}
}

// parallelStrict processes collection with fail-fast behavior
// Returns (error, results) - first error or all results
func parallelStrictFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("parallel_strict() requires 2 arguments: collection and function")
	}

	// Get collection array
	collection, ok := args[0].(*object.Array)
	if !ok {
		return newError("parallel_strict() first argument must be array, got %s", args[0].Type())
	}

	// Get function
	fn, ok := args[1].(*object.Function)
	if !ok {
		return newError("parallel_strict() second argument must be function, got %s", args[1].Type())
	}

	elements := collection.Elements
	resultCount := len(elements)

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Pre-allocate result arrays
	results := make([]object.Object, resultCount)

	// Channel to signal first error
	errorChan := make(chan object.Object, 1)

	var wg sync.WaitGroup
	wg.Add(resultCount)

	// Process each element in parallel
	for i, elem := range elements {
		go func(index int, element object.Object) {
			defer wg.Done()

			// Check if already canceled
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Call function with element
			result := applyFunction(fn, []object.Object{element})

			// Helper to process tuple result
			processTuple := func(errObj, resultObj object.Object) {
				// Check if this is an error (non-empty map with msg field)
				if isbarkError(errObj) {
					// Signal error and cancel
					select {
					case errorChan <- errObj:
						cancel()
					default:
					}
					return
				}

				// Store result
				results[index] = resultObj
			}

			// Function should return a tuple (error, result)
			if tuple, ok := result.(*object.Tuple); ok && len(tuple.Elements) == 2 {
				processTuple(tuple.Elements[0], tuple.Elements[1])
			} else if tuple, ok := result.(*object.Array); ok && len(tuple.Elements) == 2 {
				// Also accept arrays for backward compatibility
				processTuple(tuple.Elements[0], tuple.Elements[1])
			} else if errObj, ok := result.(*object.Error); ok {
				// Direct error object
				select {
				case errorChan <- errObj:
					cancel()
				default:
				}
			} else {
				// Invalid return type
				err := newError("parallel_strict() function must return (error, result) tuple")
				select {
				case errorChan <- err:
					cancel()
				default:
				}
			}
		}(i, elem)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check if any error occurred
	select {
	case err := <-errorChan:
		// Return error and empty results
		return &object.Array{
			Elements: []object.Object{
				err,
				&object.Array{Elements: []object.Object{}},
			},
		}
	default:
		// All succeeded, return empty error and results
		emptyError := &object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}
		return &object.Array{
			Elements: []object.Object{
				emptyError,
				&object.Array{Elements: results},
			},
		}
	}
}

// Helper functions

// allAbsent checks if all errors in array are absent (all operations succeeded)
func allAbsentFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("all_absent?() requires 1 argument: errors array")
	}

	errorsArray, ok := args[0].(*object.Array)
	if !ok {
		return newError("all_absent?() argument must be array, got %s", args[0].Type())
	}

	for _, errObj := range errorsArray.Elements {
		if isbarkError(errObj) {
			return FALSE
		}
	}

	return TRUE
}

// allPresent checks if all errors in array are present (all operations failed)
func allPresentFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("all_present?() requires 1 argument: errors array")
	}

	errorsArray, ok := args[0].(*object.Array)
	if !ok {
		return newError("all_present?() argument must be array, got %s", args[0].Type())
	}

	if len(errorsArray.Elements) == 0 {
		return FALSE
	}

	for _, errObj := range errorsArray.Elements {
		if !isbarkError(errObj) {
			return FALSE
		}
	}

	return TRUE
}

// firstError returns first non-absent error from array, or empty map if all absent
func firstErrorFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("first_error() requires 1 argument: errors array")
	}

	errorsArray, ok := args[0].(*object.Array)
	if !ok {
		return newError("first_error() argument must be array, got %s", args[0].Type())
	}

	for _, errObj := range errorsArray.Elements {
		if isbarkError(errObj) {
			return errObj
		}
	}

	// No errors found, return empty map
	return &object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}
}
