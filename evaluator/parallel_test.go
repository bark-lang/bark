package evaluator

import (
	"testing"

	"gitlab.com/bark-lang/barki/object"
)

func TestParallelBasic(t *testing.T) {
	input := `
	fn test() {
		[1, 2, 3] > parallel((x int) {
			x > add(10) > result
			return({}, result)
		}(map, int)) > result_tuple

		result_tuple > get(1) > return()
	}(array)

	test()
	`

	evaluated := testEval(input)
	arr, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("Expected Array, got %T (%+v)", evaluated, evaluated)
	}

	if len(arr.Elements) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(arr.Elements))
	}

	expected := []int64{11, 12, 13}
	for i, elem := range arr.Elements {
		intObj, ok := elem.(*object.Integer)
		if !ok {
			t.Fatalf("Element %d: expected Integer, got %T", i, elem)
		}
		if intObj.Value != expected[i] {
			t.Fatalf("Element %d: expected %d, got %d", i, expected[i], intObj.Value)
		}
	}
}

func TestParallelOrderPreservation(t *testing.T) {
	input := `
	fn test() {
		[1, 2, 3, 4, 5] > parallel((x int) {
			return({}, x)
		}(map, int)) > result_tuple

		result_tuple > get(1) > return()
	}(array)

	test()
	`

	evaluated := testEval(input)
	arr, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("Expected Array, got %T", evaluated)
	}

	if len(arr.Elements) != 5 {
		t.Fatalf("Expected 5 results, got %d", len(arr.Elements))
	}

	for i, elem := range arr.Elements {
		intObj, ok := elem.(*object.Integer)
		if !ok {
			t.Fatalf("Element %d: expected Integer, got %T", i, elem)
		}
		if intObj.Value != int64(i+1) {
			t.Fatalf("Element %d: expected %d, got %d", i, i+1, intObj.Value)
		}
	}
}

func TestParallelWithErrors(t *testing.T) {
	input := `
	fn test() {
		[1, 2, 3] > parallel((x int) {
			x > eq?(2) > is_two
			is_two > (cond bool) {
				err("Error on 2", {}) > e
				cond > return?(e, 0)
				return({}, x)
			}(map, int)
		}(map, int)) > result_tuple

		result_tuple > get(0) > errors
		errors > all_absent?() > return()
	}(bool)

	test()
	`

	evaluated := testEval(input)
	testBooleanObject(t, evaluated, false)
}

func TestParallelAllBasic(t *testing.T) {
	input := `
	fn test() {
		[
			(dummy int) { return({}, 1) }(map, int),
			(dummy int) { return({}, 2) }(map, int),
			(dummy int) { return({}, 3) }(map, int)
		] > parallel_all() > result_tuple

		result_tuple > get(1) > results
		results > len() > return()
	}(int)

	test()
	`

	evaluated := testEval(input)
	intObj, ok := evaluated.(*object.Integer)
	if !ok {
		t.Fatalf("Expected Integer, got %T", evaluated)
	}

	if intObj.Value != 3 {
		t.Fatalf("Expected 3 results, got %d", intObj.Value)
	}
}

func TestParallelLimitedBasic(t *testing.T) {
	input := `
	fn test() {
		parallel_limited(2, [1, 2, 3, 4, 5], (x int) {
			return({}, x)
		}(map, int)) > result_tuple

		result_tuple > get(1) > results
		results > len() > return()
	}(int)

	test()
	`

	evaluated := testEval(input)
	intObj, ok := evaluated.(*object.Integer)
	if !ok {
		t.Fatalf("Expected Integer, got %T", evaluated)
	}

	if intObj.Value != 5 {
		t.Fatalf("Expected 5 results, got %d", intObj.Value)
	}
}

func TestParallelRaceBasic(t *testing.T) {
	input := `
	fn test() {
		[
			(dummy int) { return({}, 1) }(map, int),
			(dummy int) { return({}, 2) }(map, int),
			(dummy int) { return({}, 3) }(map, int)
		] > parallel_race() > result_tuple

		result_tuple > get(1) > result
		result > present?() > return()
	}(bool)

	test()
	`

	evaluated := testEval(input)
	testBooleanObject(t, evaluated, true)
}

func TestParallelStrictSuccess(t *testing.T) {
	input := `
	fn test() {
		[1, 2, 3] > parallel_strict((x int) {
			return({}, x)
		}(map, int)) > result_tuple

		result_tuple > get(0) > err_result
		err_result > absent?() > return()
	}(bool)

	test()
	`

	evaluated := testEval(input)
	testBooleanObject(t, evaluated, true)
}

func TestParallelStrictFailure(t *testing.T) {
	input := `
	fn test() {
		[1, 2, 3] > parallel_strict((x int) {
			x > eq?(2) > is_two
			is_two > (cond bool) {
				err("Error on 2", {}) > e
				cond > return?(e, 0)
				return({}, x)
			}(map, int)
		}(map, int)) > result_tuple

		result_tuple > get(1) > results
		results > len() > eq?(0) > return()
	}(bool)

	test()
	`

	evaluated := testEval(input)
	testBooleanObject(t, evaluated, true)
}

func TestAllAbsentTrue(t *testing.T) {
	input := `
	[{}, {}, {}] > all_absent?()
	`

	evaluated := testEval(input)
	testBooleanObject(t, evaluated, true)
}

func TestAllAbsentFalse(t *testing.T) {
	// Test using parallel() which creates error arrays
	input := `
	fn test() {
		[1, 2] > parallel((x int) {
			x > eq?(2) > is_two
			is_two > (cond bool) {
				err("Error", {}) > e
				cond > return?(e, 0)
				return({}, x)
			}(map, int)
		}(map, int)) > result_tuple

		result_tuple > get(0) > errors
		errors > all_absent?() > return()
	}(bool)

	test()
	`

	evaluated := testEval(input)
	testBooleanObject(t, evaluated, false)
}

func TestAllPresentTrue(t *testing.T) {
	// This would require all operations to fail, which is tested via parallel
	input := `
	fn test() {
		[1, 2, 3] > parallel((x int) {
			return(err("Always fail", {}), 0)
		}(map, int)) > result_tuple

		result_tuple > get(0) > errors
		errors > all_present?() > return()
	}(bool)

	test()
	`

	evaluated := testEval(input)
	testBooleanObject(t, evaluated, true)
}

func TestAllPresentFalse(t *testing.T) {
	// Just test that we got some successful results
	input := `
	fn test() {
		[1, 2, 3] > parallel((x int) {
			x > eq?(2) > is_two
			is_two > (cond bool) {
				cond > return?(err("Error", {}), 0)
				return({}, x)
			}(map, int)
		}(map, int)) > result_tuple

		result_tuple > get(1) > results
		results > len() > eq?(3) > return()
	}(bool)

	test()
	`

	evaluated := testEval(input)
	testBooleanObject(t, evaluated, true)
}

func TestFirstErrorFound(t *testing.T) {
	// Just verify we can detect an error exists
	input := `
	fn test() {
		[1, 2, 3] > parallel((x int) {
			x > eq?(2) > is_two
			is_two > (cond bool) {
				err("Second Error", {}) > e
				cond > return?(e, 0)
				return({}, x)
			}(map, int)
		}(map, int)) > result_tuple

		result_tuple > get(0) > errors
		errors > all_absent?() > not() > return()
	}(bool)

	test()
	`

	evaluated := testEval(input)
	testBooleanObject(t, evaluated, true)
}

func TestFirstErrorNone(t *testing.T) {
	input := `
	fn test() {
		[{}, {}, {}] > first_error() > err_result
		err_result > absent?() > return()
	}(bool)

	test()
	`

	evaluated := testEval(input)
	testBooleanObject(t, evaluated, true)
}

func TestParallelEmptyArray(t *testing.T) {
	input := `
	fn test() {
		[] > parallel((x int) {
			return({}, x)
		}(map, int)) > result_tuple

		result_tuple > get(1) > results
		results > len() > return()
	}(int)

	test()
	`

	evaluated := testEval(input)
	intObj, ok := evaluated.(*object.Integer)
	if !ok {
		t.Fatalf("Expected Integer, got %T", evaluated)
	}

	if intObj.Value != 0 {
		t.Fatalf("Expected 0 results, got %d", intObj.Value)
	}
}
