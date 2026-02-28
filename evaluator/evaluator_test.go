package evaluator

import (
	"testing"

	"gitlab.com/bark-lang/barki/lexer"
	"gitlab.com/bark-lang/barki/object"
	"gitlab.com/bark-lang/barki/parser"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"42", 42},
		{"123", 123},
		{"0", 0},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalFloatExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14", 3.14},
		{"0.5", 0.5},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestEvalStringExpression(t *testing.T) {
	input := `"hello world"`
	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}
	if str.Value != "hello world" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2, 3]"
	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", evaluated, evaluated)
	}
	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d", len(result.Elements))
	}
	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 2)
	testIntegerObject(t, result.Elements[2], 3)
}

func TestMapLiterals(t *testing.T) {
	input := `{"name": "John", "age": "30"}`
	evaluated := testEval(input)
	result, ok := evaluated.(*object.Map)
	if !ok {
		t.Fatalf("object is not Map. got=%T (%+v)", evaluated, evaluated)
	}
	if len(result.Pairs) != 2 {
		t.Fatalf("map has wrong num of pairs. got=%d", len(result.Pairs))
	}
}

func TestVariableBinding(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"42 > x\nx", 42},
		{"5 > a\n10 > b\na", 5},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fn identity(x int) { x > return() }(int)"
	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}
	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v", fn.Parameters)
	}
	if fn.Parameters[0].Name.Value != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0].Name.Value)
	}
}

func TestFunctionCall(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{
			`fn double(x int) { x > mul(2) > return() }(int)
			 double(5)`,
			10,
		},
		{
			`fn add_nums(a int, b int) { a > add(b) > return() }(int)
			 add_nums(3, 7)`,
			10,
		},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestLinkOperatorWithFunction(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5 > add(3)", 8},
		{"10 > sub(3)", 7},
		{"4 > mul(3)", 12},
		{"10 > div(2)", 5},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestChainedLinkOperator(t *testing.T) {
	input := "5 > add(3) > mul(2)"
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 16)
}

func TestComparisonFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"eq?(5, 5)", true},
		{"eq?(5, 3)", false},
		{"gt?(5, 3)", true},
		{"gt?(3, 5)", false},
		{"lt?(3, 5)", true},
		{"lt?(5, 3)", false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("hello")`, 5},
		{`len([1, 2, 3])`, 3},
		{`to_string(42)`, "42"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			str, ok := evaluated.(*object.String)
			if !ok {
				t.Errorf("object is not String. got=%T (%+v)", evaluated, evaluated)
				continue
			}
			if str.Value != expected {
				t.Errorf("String has wrong value. got=%q, want=%q", str.Value, expected)
			}
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return(10)", 10},
		{"return(5)\n9", 5},
		{
			`fn test() {
				return(42)
				return(99)
			}(int)
			test()`,
			42,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"unknown",
			"identifier not found: unknown",
		},
		{
			"add(1)",
			"wrong number of arguments. got=1, want=2",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)", evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tt.expectedMessage, errObj.Msg)
		}
	}
}

func TestTupleToAnonymousFunction(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		// Basic tuple to anonymous function
		{`(5, 1) > (a int, b int) { return(a) }(int)`, 5},
		{`(5, 1) > (a int, b int) { return(b) }(int)`, 1},
		// Tuple with expressions
		{`(3, 2) > (a int, b int) { a > add(b) > return() }(int)`, 5},
		// Three element tuple
		{`(1, 2, 3) > (a int, b int, c int) { a > add(b) > add(c) > return() }(int)`, 6},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestTupleToNamedFunction(t *testing.T) {
	input := `
	fn adder(a int, b int) {
		a > add(b) > return()
	}(int)
	(3, 4) > adder()
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 7)
}

func TestTupleUnpackWithAdditionalArgs(t *testing.T) {
	// Tuple elements + additional args
	input := `
	fn three_args(a int, b int, c int) {
		a > add(b) > add(c) > return()
	}(int)
	(1, 2) > three_args(3)
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 6)
}

func TestTupleErrorCases(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		// Cannot assign tuple to variable
		{`(1, 2) > x`, "cannot assign tuple to variable"},
		// Wrong number of arguments
		{`(1, 2) > (a int) { return(a) }(int)`, "wrong number of arguments: expected 1, got 2"},
		{`(1, 2, 3) > (a int, b int) { return(a) }(int)`, "wrong number of arguments: expected 2, got 3"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned for %q. got=%T(%+v)",
				tt.input, evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("wrong error message for %q. expected=%q, got=%q",
				tt.input, tt.expectedMessage, errObj.Msg)
		}
	}
}

func TestTupleWithRecursion(t *testing.T) {
	// Factorial using tuple initialization
	input := `
	fn factorial(n int) {
		(n, 1) > (num int, acc int) {
			num > lte?(1) > should_return
			should_return > return?(acc)

			num > sub(1) > next_num
			acc > mul(num) > next_acc
			repeat(next_num, next_acc)
		}(int) > result
		return(result)
	}(int)
	factorial(5)
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 120)
}

// Helper functions

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(program, env)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
		return false
	}
	return true
}

func testFloatObject(t *testing.T, obj object.Object, expected float64) bool {
	result, ok := obj.(*object.Float)
	if !ok {
		t.Errorf("object is not Float. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%f, want=%f",
			result.Value, expected)
		return false
	}
	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t",
			result.Value, expected)
		return false
	}
	return true
}

func testStringObject(t *testing.T, obj object.Object, expected string) bool {
	result, ok := obj.(*object.String)
	if !ok {
		t.Errorf("object is not String. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%s, want=%s", result.Value, expected)
		return false
	}
	return true
}

// Capture expression tests

func TestCaptureSuccess(t *testing.T) {
	// Test capture with no error - chain should continue
	input := `
	fn returnSuccess(x int) {
		x > add(10) > result
		return({}, result)
	}(error, int)

	5 > returnSuccess() > capture(e, r) > add(5)
	`
	evaluated := testEval(input)
	// 5 + 10 = 15, then + 5 = 20
	testIntegerObject(t, evaluated, 20)
}

func TestCaptureSuccessBindsVariables(t *testing.T) {
	// Test that capture binds both variables correctly on success
	input := `
	fn returnSuccess(x int) {
		x > add(10) > result
		return({}, result)
	}(error, int)

	5 > returnSuccess() > capture(e, r)
	r
	`
	evaluated := testEval(input)
	// r should be 15 (5 + 10)
	testIntegerObject(t, evaluated, 15)
}

func TestCaptureErrorStopsChain(t *testing.T) {
	// Test capture with error - chain should stop, error bound to variable
	input := `
	fn returnError(x int) {
		err("something went wrong") > e
		return(e, 0)
	}(error, int)

	5 > returnError() > capture(e, r) > add(100)
	e > err_msg()
	`
	evaluated := testEval(input)
	// Chain stops at capture, e has the error
	testStringObject(t, evaluated, "something went wrong")
}

func TestCaptureErrorBindsVariables(t *testing.T) {
	// Test that capture binds error variable on error
	input := `
	fn returnError(x int) {
		err("test error") > e
		return(e, 42)
	}(error, int)

	5 > returnError() > capture(e, r)
	r
	`
	evaluated := testEval(input)
	// r should be 42 (the result value even though there was an error)
	testIntegerObject(t, evaluated, 42)
}

func TestCaptureMultipleInChain(t *testing.T) {
	// Test multiple captures in a chain
	input := `
	fn op1(x int) {
		x > add(1) > result
		return({}, result)
	}(error, int)

	fn op2(x int) {
		x > mul(2) > result
		return({}, result)
	}(error, int)

	5 > op1() > capture(e1, r1) > op2() > capture(e2, r2)
	r2
	`
	evaluated := testEval(input)
	// (5 + 1) * 2 = 12
	testIntegerObject(t, evaluated, 12)
}

func TestCaptureFirstErrorStopsChain(t *testing.T) {
	// Test that first error stops the chain
	input := `
	fn op1(x int) {
		err("first error") > e
		return(e, 0)
	}(error, int)

	fn op2(x int) {
		x > mul(2) > result
		return({}, result)
	}(error, int)

	5 > op1() > capture(e1, r1) > op2() > capture(e2, r2)
	e1 > err_msg()
	`
	evaluated := testEval(input)
	// Chain stops at first capture, e1 has the error
	testStringObject(t, evaluated, "first error")
}

func TestCaptureRequiresTuple(t *testing.T) {
	// Test that capture requires a tuple input
	input := `5 > capture(e, r)`
	evaluated := testEval(input)
	errObj, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("expected Error, got %T", evaluated)
	}
	if errObj.Msg != "capture requires (error, value) tuple, got INTEGER" {
		t.Errorf("wrong error message: %s", errObj.Msg)
	}
}

func TestCaptureWithEmptyMapError(t *testing.T) {
	// Test that empty map {} means no error
	input := `
	fn returnEmptyMap(x int) {
		x > add(10) > result
		return({}, result)
	}(error, int)

	5 > returnEmptyMap() > capture(e, r) > add(5)
	`
	evaluated := testEval(input)
	// Empty map means no error, chain continues: (5 + 10) + 5 = 20
	testIntegerObject(t, evaluated, 20)
}

// Memoized function tests

func TestMemoizedFunctionBasic(t *testing.T) {
	// Test basic memoized function definition and call
	input := `
	mfn double(n int) {
		n > mul(2) > return()
	}(int)
	double(5)
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 10)
}

func TestMemoizedFunctionCaching(t *testing.T) {
	// Test that memoized function returns cached results
	// We use a counter to verify the function body is only executed once per unique args
	input := `
	mfn addTen(n int) {
		n > add(10) > return()
	}(int)

	// Call with same argument multiple times
	5 > addTen() > first
	5 > addTen() > second
	5 > addTen() > third

	// All should return 15
	first > add(second) > add(third)
	`
	evaluated := testEval(input)
	// 15 + 15 + 15 = 45
	testIntegerObject(t, evaluated, 45)
}

func TestMemoizedFunctionDifferentArgs(t *testing.T) {
	// Test that different arguments produce different cached results
	input := `
	mfn square(n int) {
		n > mul(n) > return()
	}(int)

	2 > square() > a
	3 > square() > b
	4 > square() > c

	a > add(b) > add(c)
	`
	evaluated := testEval(input)
	// 4 + 9 + 16 = 29
	testIntegerObject(t, evaluated, 29)
}

func TestMemoizedFunctionRecursive(t *testing.T) {
	// Test recursive memoized function (fibonacci)
	input := `
	mfn fib(n int) {
		n > lte?(1) > return?(n)
		n > sub(1) > fib() > a
		n > sub(2) > fib() > b
		a > add(b) > return()
	}(int)

	10 > fib()
	`
	evaluated := testEval(input)
	// fib(10) = 55
	testIntegerObject(t, evaluated, 55)
}

func TestMemoizedFunctionMultipleArgs(t *testing.T) {
	// Test memoized function with multiple arguments
	input := `
	mfn add_mul(a int, b int) {
		a > add(b) > sum
		a > mul(b) > product
		sum > add(product) > return()
	}(int)

	(3, 4) > add_mul() > first
	(3, 4) > add_mul() > second
	(4, 3) > add_mul() > third

	first > add(second) > add(third)
	`
	evaluated := testEval(input)
	// (3+4) + (3*4) = 7 + 12 = 19
	// first = 19, second = 19 (cached), third = 19 (different order, same result)
	// 19 + 19 + 19 = 57
	testIntegerObject(t, evaluated, 57)
}

func TestMemoizedFunctionWithStrings(t *testing.T) {
	// Test memoized function with string arguments
	input := `
	mfn greet(name string) {
		"Hello, " > str.concat(name) > return()
	}(string)

	"World" > greet()
	`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "Hello, World")
}

func TestMemoizedFunctionInspect(t *testing.T) {
	// Test that memoized function has correct type
	input := `
	mfn test(n int) {
		return(n)
	}(int)
	test
	`
	evaluated := testEval(input)
	mfn, ok := evaluated.(*object.MemoizedFunction)
	if !ok {
		t.Fatalf("object is not MemoizedFunction. got=%T (%+v)", evaluated, evaluated)
	}
	if len(mfn.Parameters) != 1 {
		t.Errorf("wrong number of parameters. got=%d, want=1", len(mfn.Parameters))
	}
}

// String interpolation tests

func TestStringInterpolationSimple(t *testing.T) {
	// Test basic variable interpolation
	input := `
	"Alice" > name
	"Hello, {name}!"
	`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "Hello, Alice!")
}

func TestStringInterpolationMultipleVars(t *testing.T) {
	// Test multiple variables in one string
	input := `
	"Alice" > name
	30 > age
	"{name} is {age} years old"
	`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "Alice is 30 years old")
}

func TestStringInterpolationMapField(t *testing.T) {
	// Test map field access in interpolation
	input := `
	{"name": "Bob", "city": "Seattle"} > user
	"{user.name} lives in {user.city}"
	`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "Bob lives in Seattle")
}

func TestStringInterpolationEscapedBraces(t *testing.T) {
	// Test escaped braces produce literal braces
	input := `"Use \{name\} for interpolation"`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "Use {name} for interpolation")
}

func TestStringInterpolationPositionalPassthrough(t *testing.T) {
	// Test that positional placeholders {0}, {1} are passed through
	input := `"{0} and {1}"`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "{0} and {1}")
}

func TestStringInterpolationMixedWithPositional(t *testing.T) {
	// Test mixing interpolation with positional placeholders
	input := `
	"world" > greeting
	"Hello {greeting}, {0}!"
	`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "Hello world, {0}!")
}

func TestStringInterpolationUndefinedVarPassthrough(t *testing.T) {
	// Test that undefined variables are passed through unchanged
	input := `"{undefined_var}"`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "{undefined_var}")
}

func TestStringInterpolationJSONPassthrough(t *testing.T) {
	// Test that JSON-like content in strings passes through unchanged
	input := `"{\"name\": \"Alice\"}"`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "{\"name\": \"Alice\"}")
}

func TestStringInterpolationEmptyBraces(t *testing.T) {
	// Test that empty braces {} pass through unchanged
	input := `"empty: {}"`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "empty: {}")
}

func TestStringInterpolationWithIntegers(t *testing.T) {
	// Test interpolation with integer values
	input := `
	42 > answer
	"The answer is {answer}"
	`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "The answer is 42")
}

func TestStringInterpolationWithBooleans(t *testing.T) {
	// Test interpolation with boolean values
	input := `
	true > enabled
	"Feature enabled: {enabled}"
	`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "Feature enabled: true")
}

func TestStringInterpolationWithFloats(t *testing.T) {
	// Test interpolation with float values
	input := `
	3.14 > pi
	"Pi is approximately {pi}"
	`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "Pi is approximately 3.14")
}

func TestStringInterpolationWithArray(t *testing.T) {
	// Test interpolation with array values
	input := `
	[1, 2, 3] > nums
	"Numbers: {nums}"
	`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "Numbers: [1, 2, 3]")
}

func TestStringInterpolationNestedMapFieldNotSupported(t *testing.T) {
	// Test that nested field access is passed through unchanged
	input := `
	{"user": {"name": "Alice"}} > data
	"{data.user.name}"
	`
	evaluated := testEval(input)
	// Nested access not supported, should pass through unchanged
	testStringObject(t, evaluated, "{data.user.name}")
}

// ============================================================================
// Type Checking Tests
// ============================================================================

func TestParameterTypeValidation(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		// Type mismatch: string passed to int parameter
		{
			`"hello" > (x int) { return(x) }(int)`,
			"type mismatch: parameter 'x' expects int, got STRING",
		},
		// Type mismatch: int passed to string parameter
		{
			`42 > (s string) { return(s) }(string)`,
			"type mismatch: parameter 's' expects string, got INTEGER",
		},
		// Type mismatch: bool passed to int parameter
		{
			`true > (n int) { return(n) }(int)`,
			"type mismatch: parameter 'n' expects int, got BOOLEAN",
		},
		// Type mismatch: array passed to map parameter
		{
			`[1, 2, 3] > (m map) { return(m) }(map)`,
			"type mismatch: parameter 'm' expects map, got ARRAY",
		},
		// Multiple parameters - second parameter wrong type
		{
			`(1, "two") > (a int, b int) { return(add(a, b)) }(int)`,
			"type mismatch: parameter 'b' expects int, got STRING",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned for %q. got=%T(%+v)",
				tt.input, evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("wrong error message for %q.\nexpected=%q\ngot=%q",
				tt.input, tt.expectedMessage, errObj.Msg)
		}
	}
}

func TestParameterTypeValidationSuccess(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Correct types should pass
		{`42 > (x int) { return(x) }(int)`, int64(42)},
		{`"hello" > (s string) { return(s) }(string)`, "hello"},
		{`true > (b bool) { return(b) }(bool)`, true},
		{`3.14 > (f float) { return(f) }(float)`, 3.14},
		{`[1, 2, 3] > (arr array) { return(arr) }(array)`, "array"},
		{`{"a": 1} > (m map) { return(m) }(map)`, "map"},
		// Multiple parameters with correct types
		{`(1, 2) > (a int, b int) { a > add(b) > return() }(int)`, int64(3)},
		{`("hello", " world") > (a string, b string) { a > str.concat(b) > return() }(string)`, "hello world"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int64:
			testIntegerObject(t, evaluated, expected)
		case string:
			switch expected {
			case "array":
				if _, ok := evaluated.(*object.Array); !ok {
					t.Errorf("expected Array for %q, got=%T", tt.input, evaluated)
				}
			case "map":
				if _, ok := evaluated.(*object.Map); !ok {
					t.Errorf("expected Map for %q, got=%T", tt.input, evaluated)
				}
			default:
				testStringObject(t, evaluated, expected)
			}
		case bool:
			testBooleanObject(t, evaluated, expected)
		case float64:
			testFloatObject(t, evaluated, expected)
		}
	}
}

func TestGenericTypeT(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Generic type t accepts any type
		{`42 > (x t) { return(x) }(t)`, int64(42)},
		{`"hello" > (x t) { return(x) }(t)`, "hello"},
		{`true > (x t) { return(x) }(t)`, true},
		{`[1, 2, 3] > (x t) { return(x) }(t)`, "array"},
		{`{"a": 1} > (x t) { return(x) }(t)`, "map"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		// Should not be an error
		if errObj, ok := evaluated.(*object.Error); ok {
			t.Errorf("unexpected error for %q: %s", tt.input, errObj.Msg)
			continue
		}

		switch expected := tt.expected.(type) {
		case int64:
			testIntegerObject(t, evaluated, expected)
		case string:
			switch expected {
			case "array":
				if _, ok := evaluated.(*object.Array); !ok {
					t.Errorf("expected Array for %q, got=%T", tt.input, evaluated)
				}
			case "map":
				if _, ok := evaluated.(*object.Map); !ok {
					t.Errorf("expected Map for %q, got=%T", tt.input, evaluated)
				}
			default:
				testStringObject(t, evaluated, expected)
			}
		case bool:
			testBooleanObject(t, evaluated, expected)
		}
	}
}

func TestReturnTypeValidation(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		// Return type mismatch: returning string when int expected
		{
			`5 > (x int) { return("wrong") }(int)`,
			"return type mismatch: expected int, got STRING",
		},
		// Return type mismatch: returning int when string expected
		{
			`"hi" > (s string) { return(42) }(string)`,
			"return type mismatch: expected string, got INTEGER",
		},
		// Return type mismatch: returning bool when int expected
		{
			`1 > (x int) { return(true) }(int)`,
			"return type mismatch: expected int, got BOOLEAN",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned for %q. got=%T(%+v)",
				tt.input, evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("wrong error message for %q.\nexpected=%q\ngot=%q",
				tt.input, tt.expectedMessage, errObj.Msg)
		}
	}
}

func TestReturnTypeValidationSuccess(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Correct return types
		{`5 > (x int) { return(x) }(int)`, int64(5)},
		{`"hello" > (s string) { return(s) }(string)`, "hello"},
		{`true > (b bool) { return(b) }(bool)`, true},
		// Generic return type t
		{`42 > (x int) { return(x) }(t)`, int64(42)},
		{`"hi" > (s string) { return(s) }(t)`, "hi"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		// Should not be an error
		if errObj, ok := evaluated.(*object.Error); ok {
			t.Errorf("unexpected error for %q: %s", tt.input, errObj.Msg)
			continue
		}

		switch expected := tt.expected.(type) {
		case int64:
			testIntegerObject(t, evaluated, expected)
		case string:
			testStringObject(t, evaluated, expected)
		case bool:
			testBooleanObject(t, evaluated, expected)
		}
	}
}

func TestNamedFunctionTypeValidation(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		// Named function with type mismatch
		{
			`fn double(x int) { return(mul(x, 2)) }(int)
			"not an int" > double()`,
			"type mismatch: parameter 'x' expects int, got STRING",
		},
		// Named function return type mismatch
		{
			`fn wrong(x int) { return("oops") }(int)
			5 > wrong()`,
			"return type mismatch: expected int, got STRING",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned for %q. got=%T(%+v)",
				tt.input, evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("wrong error message for %q.\nexpected=%q\ngot=%q",
				tt.input, tt.expectedMessage, errObj.Msg)
		}
	}
}

func TestTupleTypeValidation(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		// Tuple type mismatch: wrong element type
		{
			`(1, "hello") > (pair (int, int)) { return(pair) }((int, int))`,
			"type mismatch: parameter 'pair[1]' expects int, got STRING",
		},
		// Tuple type mismatch: not a tuple
		{
			`42 > (pair (int, string)) { return(pair) }((int, string))`,
			"type mismatch: parameter 'pair' expects (int, string), got INTEGER",
		},
		// Tuple type mismatch: wrong number of elements
		{
			`(1, 2, 3) > (pair (int, int)) { return(pair) }((int, int))`,
			"type mismatch: parameter 'pair' expects tuple with 2 elements, got 3",
		},
		// Nested tuple type mismatch
		{
			`((1, 2), "hello") > (nested ((int, string), string)) { return(nested) }(((int, string), string))`,
			"type mismatch: parameter 'nested[0][1]' expects string, got INTEGER",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned for %q. got=%T(%+v)",
				tt.input, evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("wrong error message for %q.\nexpected=%q\ngot=%q",
				tt.input, tt.expectedMessage, errObj.Msg)
		}
	}
}

func TestTupleTypeValidationSuccess(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Correct tuple types - return the tuple itself
		{`(1, "hello") > (pair (int, string)) { return(pair) }((int, string))`, "(1, hello)"},
		{`(true, 3.14) > (pair (bool, float)) { return(pair) }((bool, float))`, "(true, 3.14)"},
		// Generic type t in tuple
		{`(1, "hello") > (pair (int, t)) { return(pair) }((int, t))`, "(1, hello)"},
		// Nested tuples
		{`((1, 2), "outer") > (nested ((int, int), string)) { return(nested) }(((int, int), string))`, "((1, 2), outer)"},
		// Tuple passed to function that processes it
		{`(10, 20) > (pair (int, int)) { pair > (a int, b int) { add(a, b) > return() }(int) > return() }(int)`, "30"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		// Should not be an error
		if errObj, ok := evaluated.(*object.Error); ok {
			t.Errorf("unexpected error for %q: %s", tt.input, errObj.Msg)
			continue
		}

		// Check the string representation
		if evaluated.Inspect() != tt.expected {
			t.Errorf("wrong result for %q. expected=%q, got=%q",
				tt.input, tt.expected, evaluated.Inspect())
		}
	}
}

func TestFunctionTypeValidation(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		// Function type mismatch: passing non-function
		{
			`42 > (f fn(int)(int)) { 5 > f() > return() }(int)`,
			"type mismatch: parameter 'f' expects fn(int)(int), got INTEGER",
		},
		// Function type mismatch: wrong parameter count
		{
			`(x int) { return(x) }(int) > (f fn(int, int)(int)) { (1, 2) > f() > return() }(int)`,
			"type mismatch: parameter 'f' expects function with 2 parameters, got 1",
		},
		// Function type mismatch: wrong parameter type
		{
			`(x string) { return(x) }(string) > (f fn(int)(string)) { 5 > f() > return() }(string)`,
			"type mismatch: parameter 'f' function parameter 0 ('x') expects int, got string",
		},
		// Function type mismatch: wrong return count
		{
			`(x int) { return(x) }(int) > (f fn(int)(int, int)) { 5 > f() > return() }((int, int))`,
			"type mismatch: parameter 'f' expects function with 2 return values, got 1",
		},
		// Function type mismatch: wrong return type
		{
			`(x int) { return(x) }(int) > (f fn(int)(string)) { 5 > f() > return() }(string)`,
			"type mismatch: parameter 'f' function return 0 expects string, got int",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned for %q. got=%T(%+v)",
				tt.input, evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("wrong error message for %q.\nexpected=%q\ngot=%q",
				tt.input, tt.expectedMessage, errObj.Msg)
		}
	}
}

func TestFunctionTypeValidationSuccess(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic function type - int -> int
		{
			`(x int) { mul(x, 2) > return() }(int) > (f fn(int)(int)) { 5 > f() > return() }(int)`,
			"10",
		},
		// Function type with multiple parameters
		{
			`(a int, b int) { add(a, b) > return() }(int) > (f fn(int, int)(int)) { (3, 4) > f() > return() }(int)`,
			"7",
		},
		// Function type with no return values
		{
			`(x int) { println(x) }() > (f fn(int)()) { 42 > f() }()`,
			"null",
		},
		// Generic type t in function type
		{
			`(x int) { return(x) }(int) > (f fn(t)(t)) { 5 > f() > return() }(t)`,
			"5",
		},
		// Function returning tuple
		{
			`(x int) { (x, mul(x, 2)) > return() }((int, int)) > (f fn(int)((int, int))) { 5 > f() > return() }((int, int))`,
			"(5, 10)",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		// Should not be an error
		if errObj, ok := evaluated.(*object.Error); ok {
			t.Errorf("unexpected error for %q: %s", tt.input, errObj.Msg)
			continue
		}

		// Check the string representation
		if evaluated.Inspect() != tt.expected {
			t.Errorf("wrong result for %q. expected=%q, got=%q",
				tt.input, tt.expected, evaluated.Inspect())
		}
	}
}

func TestArrayTypeValidation(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		// Type mismatch: array contains wrong element type
		{
			`[1, 2, "three"] > (arr array[int]) { return(arr) }(array[int])`,
			"type mismatch: parameter 'arr[2]' expects int, got STRING",
		},
		// Type mismatch: array contains wrong element type at first position
		{
			`["one", 2, 3] > (arr array[int]) { return(arr) }(array[int])`,
			"type mismatch: parameter 'arr[0]' expects int, got STRING",
		},
		// Type mismatch: not an array
		{
			`"hello" > (arr array[int]) { return(arr) }(array[int])`,
			"type mismatch: parameter 'arr' expects array[int], got STRING",
		},
		// Type mismatch: nested array with wrong inner type
		{
			`[[1, 2], [3, "four"]] > (arr array[array[int]]) { return(arr) }(array[array[int]])`,
			"type mismatch: parameter 'arr[1][1]' expects int, got STRING",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned for %q. got=%T(%+v)",
				tt.input, evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("wrong error message for %q.\nexpected=%q\ngot=%q",
				tt.input, tt.expectedMessage, errObj.Msg)
		}
	}
}

func TestArrayTypeValidationSuccess(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic array of integers
		{
			`[1, 2, 3] > (arr array[int]) { return(arr) }(array[int])`,
			"[1, 2, 3]",
		},
		// Array of strings
		{
			`["a", "b", "c"] > (arr array[string]) { return(arr) }(array[string])`,
			"[a, b, c]",
		},
		// Empty array is valid for any element type
		{
			`[] > (arr array[int]) { return(arr) }(array[int])`,
			"[]",
		},
		// Nested array
		{
			`[[1, 2], [3, 4]] > (arr array[array[int]]) { return(arr) }(array[array[int]])`,
			"[[1, 2], [3, 4]]",
		},
		// Generic element type accepts any
		{
			`[1, "two", true] > (arr array[t]) { return(arr) }(array[t])`,
			"[1, two, true]",
		},
		// Plain array type still works (no element validation)
		{
			`[1, "two", true] > (arr array) { return(arr) }(array)`,
			"[1, two, true]",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		if errObj, ok := evaluated.(*object.Error); ok {
			t.Errorf("unexpected error for %q: %s", tt.input, errObj.Msg)
			continue
		}

		if evaluated.Inspect() != tt.expected {
			t.Errorf("wrong result for %q. expected=%q, got=%q",
				tt.input, tt.expected, evaluated.Inspect())
		}
	}
}

func TestMapTypeValidation(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		// Type mismatch: map contains wrong value type
		{
			`{"a": 1, "b": "two"} > (m map[string, int]) { return(m) }(map[string, int])`,
			`type mismatch: parameter 'm["b"]' expects int, got STRING`,
		},
		// Type mismatch: not a map
		{
			`[1, 2, 3] > (m map[string, int]) { return(m) }(map[string, int])`,
			"type mismatch: parameter 'm' expects map[string, int], got ARRAY",
		},
		// Type mismatch: non-string key type requested
		{
			`{"a": 1} > (m map[int, int]) { return(m) }(map[int, int])`,
			"type mismatch: map keys must be string type, got int",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned for %q. got=%T(%+v)",
				tt.input, evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("wrong error message for %q.\nexpected=%q\ngot=%q",
				tt.input, tt.expectedMessage, errObj.Msg)
		}
	}
}

func TestMapTypeValidationSuccess(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic map with string keys and int values
		{
			`{"a": 1, "b": 2} > (m map[string, int]) { return(m) }(map[string, int])`,
			"{a: 1, b: 2}",
		},
		// Map with string keys and string values
		{
			`{"name": "alice", "city": "wonderland"} > (m map[string, string]) { return(m) }(map[string, string])`,
			"{name: alice, city: wonderland}",
		},
		// Empty map is valid for any value type
		{
			`{} > (m map[string, int]) { return(m) }(map[string, int])`,
			"{}",
		},
		// Generic value type accepts any
		{
			`{"a": 1, "b": "two"} > (m map[string, t]) { return(m) }(map[string, t])`,
			"{a: 1, b: two}",
		},
		// Generic key type (still only string at runtime)
		{
			`{"a": 1} > (m map[t, int]) { return(m) }(map[t, int])`,
			"{a: 1}",
		},
		// Plain map type still works (no value validation)
		{
			`{"a": 1, "b": "two"} > (m map) { return(m) }(map)`,
			"{a: 1, b: two}",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		if errObj, ok := evaluated.(*object.Error); ok {
			t.Errorf("unexpected error for %q: %s", tt.input, errObj.Msg)
			continue
		}

		if evaluated.Inspect() != tt.expected {
			t.Errorf("wrong result for %q. expected=%q, got=%q",
				tt.input, tt.expected, evaluated.Inspect())
		}
	}
}

func TestUnionTypeValidation(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		// Union type mismatch: bool not in int | string
		{
			`true > (x int | string) { return(x) }(int | string)`,
			"type mismatch: parameter 'x' expects int | string, got BOOLEAN",
		},
		// Union type mismatch: array not in int | string | bool
		{
			`[1, 2] > (x int | string | bool) { return(x) }(int | string | bool)`,
			"type mismatch: parameter 'x' expects int | string | bool, got ARRAY",
		},
		// Union type mismatch: map not in int | string
		{
			`{"a": 1} > (x int | string) { return(x) }(int | string)`,
			"type mismatch: parameter 'x' expects int | string, got MAP",
		},
		// Return type mismatch: returning bool when int | string expected
		{
			`42 > (x int) { return(true) }(int | string)`,
			"return type mismatch: expected int | string, got BOOLEAN",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned for %q. got=%T(%+v)",
				tt.input, evaluated, evaluated)
			continue
		}

		if errObj.Msg != tt.expectedMessage {
			t.Errorf("wrong error message for %q.\nexpected=%q\ngot=%q",
				tt.input, tt.expectedMessage, errObj.Msg)
		}
	}
}

func TestUnionTypeValidationSuccess(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// int matches int | string
		{`42 > (x int | string) { return(x) }(int | string)`, int64(42)},
		// string matches int | string
		{`"hello" > (x int | string) { return(x) }(int | string)`, "hello"},
		// bool matches int | bool
		{`true > (x int | bool) { return(x) }(int | bool)`, true},
		// float matches int | float | string
		{`3.14 > (x int | float | string) { return(x) }(int | float | string)`, 3.14},
		// array matches array | map
		{`[1, 2, 3] > (x array | map) { return(x) }(array | map)`, "array"},
		// map matches array | map
		{`{"a": 1} > (x array | map) { return(x) }(array | map)`, "map"},
		// Multiple parameters with union types
		{`(42, "hi") > (a int | bool, b string | int) { return(a) }(int | bool)`, int64(42)},
		// Union in tuple type
		{`(42, "hi") > (pair (int | string, string | int)) { return(pair) }((int | string, string | int))`, "(42, hi)"},
		// Union in array element type
		{`[1, "two", 3] > (arr array[int | string]) { return(arr) }(array[int | string])`, "[1, two, 3]"},
		// Union in map value type
		{`{"a": 1, "b": "two"} > (m map[string, int | string]) { return(m) }(map[string, int | string])`, "{a: 1, b: two}"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		// Should not be an error
		if errObj, ok := evaluated.(*object.Error); ok {
			t.Errorf("unexpected error for %q: %s", tt.input, errObj.Msg)
			continue
		}

		switch expected := tt.expected.(type) {
		case int64:
			testIntegerObject(t, evaluated, expected)
		case string:
			switch expected {
			case "array":
				if _, ok := evaluated.(*object.Array); !ok {
					t.Errorf("expected Array for %q, got=%T", tt.input, evaluated)
				}
			case "map":
				if _, ok := evaluated.(*object.Map); !ok {
					t.Errorf("expected Map for %q, got=%T", tt.input, evaluated)
				}
			default:
				// Could be a string value or a representation (like tuple or array)
				if strObj, ok := evaluated.(*object.String); ok {
					if strObj.Value != expected {
						t.Errorf("wrong string value for %q. expected=%q, got=%q",
							tt.input, expected, strObj.Value)
					}
				} else if evaluated.Inspect() != expected {
					t.Errorf("wrong result for %q. expected=%q, got=%q",
						tt.input, expected, evaluated.Inspect())
				}
			}
		case bool:
			testBooleanObject(t, evaluated, expected)
		case float64:
			testFloatObject(t, evaluated, expected)
		}
	}
}
