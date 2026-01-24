package evaluator

import (
	"testing"

	"gitlab.com/bark-lang/bark/lexer"
	"gitlab.com/bark-lang/bark/object"
	"gitlab.com/bark-lang/bark/parser"
)

// =============================================================================
// Builtin Function Benchmarks
// =============================================================================

// BenchmarkAdd benchmarks the add builtin
func BenchmarkAdd(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("add(1, 2)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkAddChained benchmarks chained arithmetic
func BenchmarkAddChained(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("1 > add(2) > add(3) > add(4) > add(5)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkLen benchmarks the len builtin
func BenchmarkLen(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`len("hello world")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkLenArray benchmarks len on arrays
func BenchmarkLenArray(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("len([1, 2, 3, 4, 5, 6, 7, 8, 9, 10])")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Array Operation Benchmarks
// =============================================================================

// BenchmarkArrayGet benchmarks array indexing
func BenchmarkArrayGet(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3, 4, 5] > get(2)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkArrayPush benchmarks array push
func BenchmarkArrayPush(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3] > push(4)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkArrayPushLarge benchmarks push on larger arrays
func BenchmarkArrayPushLarge(b *testing.B) {
	// Build a large array literal
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20] > push(21)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkArraySlice benchmarks array slicing
func BenchmarkArraySlice(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3, 4, 5, 6, 7, 8, 9, 10] > slice(2, 8)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkArrayReverse benchmarks array reversal
func BenchmarkArrayReverse(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3, 4, 5, 6, 7, 8, 9, 10] > reverse()")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkArrayIncludes benchmarks includes? search
func BenchmarkArrayIncludesFound(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3, 4, 5, 6, 7, 8, 9, 10] > includes?(5)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkArrayIncludesNotFound benchmarks includes? worst case
func BenchmarkArrayIncludesNotFound(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3, 4, 5, 6, 7, 8, 9, 10] > includes?(100)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Map Operation Benchmarks
// =============================================================================

// BenchmarkMapGet benchmarks map key lookup
func BenchmarkMapGet(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`{"a": 1, "b": 2, "c": 3} > get("b")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMapSet benchmarks map set operation
func BenchmarkMapSet(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`{"a": 1, "b": 2, "c": 3} > set("d", 4)`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMapSetLarge benchmarks set on larger maps
func BenchmarkMapSetLarge(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6, "g": 7, "h": 8, "i": 9, "j": 10} > set("k", 11)`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMapMerge benchmarks map merge operation
func BenchmarkMapMerge(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`{"a": 1, "b": 2} > merge({"c": 3, "d": 4})`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMapKeys benchmarks keys extraction
func BenchmarkMapKeys(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5} > keys()`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMapKeyPresent benchmarks key_present? lookup
func BenchmarkMapKeyPresent(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5} > key_present?("c")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Comparison Benchmarks
// =============================================================================

// BenchmarkEq benchmarks equality comparison
func BenchmarkEq(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("eq?(5, 5)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkEqString benchmarks string equality
func BenchmarkEqString(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`eq?("hello", "hello")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkGt benchmarks greater than comparison
func BenchmarkGt(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("gt?(10, 5)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Function Call Benchmarks
// =============================================================================

// BenchmarkFunctionCall benchmarks user-defined function calls
func BenchmarkFunctionCall(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`
fn double(x int) { x > mul(2) }(int)
5 > double()
`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkFunctionCallNested benchmarks nested function calls
func BenchmarkFunctionCallNested(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`
fn double(x int) { x > mul(2) }(int)
fn triple(x int) { x > mul(3) }(int)
5 > double() > triple()
`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkAnonymousFunction benchmarks anonymous function calls
func BenchmarkAnonymousFunction(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("5 > fn(x int) { x > mul(2) }(int)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Link Expression Benchmarks
// =============================================================================

// BenchmarkLinkSimple benchmarks simple link expressions
func BenchmarkLinkSimple(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("5 > add(3)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkLinkChain5 benchmarks 5-link chains
func BenchmarkLinkChain5(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("1 > add(1) > add(1) > add(1) > add(1)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkLinkChain10 benchmarks 10-link chains
func BenchmarkLinkChain10(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("1 > add(1) > add(1) > add(1) > add(1) > add(1) > add(1) > add(1) > add(1) > add(1)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkLinkBinding benchmarks variable binding with link
func BenchmarkLinkBinding(b *testing.B) {
	l := lexer.New("5 > x")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Need fresh environment for each iteration to avoid accumulation
		newEnv := object.NewEnvironment()
		Eval(program, newEnv)
	}
}

// =============================================================================
// String Operation Benchmarks
// =============================================================================

// BenchmarkStrUpper benchmarks string uppercase
func BenchmarkStrUpper(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`"hello world" > str.upper()`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkStrSplit benchmarks string split
func BenchmarkStrSplit(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`"a,b,c,d,e" > str.split(",")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkStrReplace benchmarks string replace
func BenchmarkStrReplace(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`"hello world" > str.replace("world", "bark")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// JSON Benchmarks
// =============================================================================

// BenchmarkJSONParse benchmarks JSON parsing
func BenchmarkJSONParse(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`json.parse("{\"a\": 1, \"b\": 2, \"c\": 3}")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkJSONStringify benchmarks JSON stringification
func BenchmarkJSONStringify(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`{"a": 1, "b": 2, "c": 3} > json.stringify()`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Error Handling Benchmarks
// =============================================================================

// BenchmarkErrCreate benchmarks error creation
func BenchmarkErrCreate(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`err("something went wrong")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkErrMessage benchmarks error message extraction
func BenchmarkErrMessage(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`err("test error") > err_msg()`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Parsing Benchmarks (to isolate parsing from evaluation)
// =============================================================================

// BenchmarkParseSimple benchmarks parsing simple expressions
func BenchmarkParseSimple(b *testing.B) {
	input := "add(1, 2)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := parser.New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParseComplex benchmarks parsing complex expressions
func BenchmarkParseComplex(b *testing.B) {
	input := `
fn calculate(x int, y int) {
	x > add(y) > mul(2)
}(int)

10 > calculate(5) > add(100)
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := parser.New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParseMapLiteral benchmarks parsing map literals
func BenchmarkParseMapLiteral(b *testing.B) {
	input := `{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6, "g": 7, "h": 8, "i": 9, "j": 10}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := parser.New(l)
		_ = p.ParseProgram()
	}
}

// =============================================================================
// Real-World Scenario Benchmarks
// =============================================================================

// BenchmarkMapTransform benchmarks a realistic map transformation
func BenchmarkMapTransform(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`
{"name": "Alice", "age": 30} > set("city", "NYC") > set("active", true)
`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkArrayPipeline benchmarks a realistic array pipeline
func BenchmarkArrayPipeline(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`
[1, 2, 3, 4, 5] > push(6, 7, 8) > reverse() > slice(0, 5)
`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Allocation Benchmarks (with memory stats)
// =============================================================================

// BenchmarkAllocAddition benchmarks allocation during integer addition
func BenchmarkAllocAddition(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("1 > add(2) > add(3)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkAllocStringOp benchmarks allocation during string operations
func BenchmarkAllocStringOp(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`"hello world" > str.upper() > str.split(" ")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkAllocArrayOp benchmarks allocation during array operations
func BenchmarkAllocArrayOp(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3] > push(4) > push(5)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkAllocMapOp benchmarks allocation during map operations
func BenchmarkAllocMapOp(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`{"a": 1} > set("b", 2) > set("c", 3)`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkAllocComparison benchmarks allocation during comparison (should use singleton booleans)
func BenchmarkAllocComparison(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("5 > gt?(3)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Numbers Builtins - Complete Coverage
// =============================================================================

// BenchmarkSub benchmarks subtraction
func BenchmarkSub(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("sub(10, 3)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMul benchmarks multiplication
func BenchmarkMul(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("mul(5, 4)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkDiv benchmarks division
func BenchmarkDiv(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("div(20, 4)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Comparison Builtins - Complete Coverage
// =============================================================================

// BenchmarkNe benchmarks not equal
func BenchmarkNe(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("ne?(5, 3)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkGte benchmarks greater than or equal
func BenchmarkGte(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("gte?(10, 10)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkLt benchmarks less than
func BenchmarkLt(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("lt?(3, 7)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkLte benchmarks less than or equal
func BenchmarkLte(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("lte?(5, 5)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkNot benchmarks logical not
func BenchmarkNot(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("not(false)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkPresent benchmarks present?
func BenchmarkPresent(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`present?("hello")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkAbsent benchmarks absent?
func BenchmarkAbsent(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("absent?(null)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Array Builtins - Complete Coverage
// =============================================================================

// BenchmarkArrayPop benchmarks pop
func BenchmarkArrayPop(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3, 4, 5] > pop()")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkArrayShift benchmarks shift
func BenchmarkArrayShift(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3, 4, 5] > shift()")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkArrayUnshift benchmarks unshift
func BenchmarkArrayUnshift(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[2, 3, 4, 5] > unshift(1)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkArrayAppendTo benchmarks append_to
func BenchmarkArrayAppendTo(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("6 > append_to([1, 2, 3, 4, 5])")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkArrayEmpty benchmarks empty?
func BenchmarkArrayEmpty(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3] > empty?()")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkArrayExcludes benchmarks excludes?
func BenchmarkArrayExcludes(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3, 4, 5] > excludes?(10)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Map Builtins - Complete Coverage
// =============================================================================

// BenchmarkMapGetOr benchmarks get_or
func BenchmarkMapGetOr(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`{"a": 1, "b": 2} > get_or("c", 0)`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMapDel benchmarks del
func BenchmarkMapDel(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`{"a": 1, "b": 2, "c": 3} > del("b")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMapValues benchmarks values
func BenchmarkMapValues(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`{"a": 1, "b": 2, "c": 3} > values()`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMapEntries benchmarks entries
func BenchmarkMapEntries(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`{"a": 1, "b": 2, "c": 3} > entries()`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMapKeyAbsent benchmarks key_absent?
func BenchmarkMapKeyAbsent(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`{"a": 1, "b": 2, "c": 3} > key_absent?("d")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Data Structure Builtins - Complete Coverage
// =============================================================================

// BenchmarkFirst benchmarks first
func BenchmarkFirst(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3, 4, 5] > first()")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkLast benchmarks last
func BenchmarkLast(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3, 4, 5] > last()")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkNext benchmarks next
func BenchmarkNext(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3, 4, 5] > next(2)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkPrev benchmarks prev
func BenchmarkPrev(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3, 4, 5] > prev(2)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkHead benchmarks head
func BenchmarkHead(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3, 4, 5] > head()")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkTail benchmarks tail
func BenchmarkTail(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("[1, 2, 3, 4, 5] > tail()")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkSize benchmarks size
func BenchmarkSize(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`{"a": 1, "b": 2, "c": 3} > size()`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Control Flow Builtins
// =============================================================================

// BenchmarkReturnConditional benchmarks return? (conditional early return)
func BenchmarkReturnConditional(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("return?(false, 42)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkContinueConditional benchmarks continue? (conditional chain continuation)
func BenchmarkContinueConditional(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("continue?(true)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkRepeatConditional benchmarks repeat? (conditional recursion)
func BenchmarkRepeatConditional(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("repeat?(false)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkRepeat benchmarks repeat (unconditional recursion marker)
func BenchmarkRepeat(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("repeat()")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkBreakConditional benchmarks break? (conditional exit)
func BenchmarkBreakConditional(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("break?(false, 42)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Core Builtins
// =============================================================================

// BenchmarkToString benchmarks to_string
func BenchmarkToString(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("42 > to_string()")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Tuple and Float Benchmarks
// =============================================================================

// BenchmarkTupleAccess benchmarks tuple creation and access
func BenchmarkTupleAccess(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("(1, 2, 3) > get(1)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkFloatArithmetic benchmarks float arithmetic
func BenchmarkFloatArithmetic(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("3.14 > add(2.86)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Math Module Benchmarks
// =============================================================================

// BenchmarkMathAbs benchmarks math.abs
func BenchmarkMathAbs(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("math.abs(-42)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMathSqrt benchmarks math.sqrt
func BenchmarkMathSqrt(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("math.sqrt(16.0)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMathPow benchmarks math.pow
func BenchmarkMathPow(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("math.pow(2.0, 10.0)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMathSin benchmarks math.sin
func BenchmarkMathSin(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("math.sin(3.14159)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMathFloor benchmarks math.floor
func BenchmarkMathFloor(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("math.floor(3.7)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMathCeil benchmarks math.ceil
func BenchmarkMathCeil(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("math.ceil(3.2)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkMathRound benchmarks math.round
func BenchmarkMathRound(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("math.round(3.5)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Regex Module Benchmarks
// =============================================================================

// BenchmarkRegexMatch benchmarks regex.match
func BenchmarkRegexMatch(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`regex.match("hello world", "world")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkRegexMatchComplex benchmarks regex.match with complex pattern
func BenchmarkRegexMatchComplex(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`regex.match("test@example.com", "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkRegexReplace benchmarks regex.replace
func BenchmarkRegexReplace(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`regex.replace("hello world", "world", "bark")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkRegexSplit benchmarks regex.split
func BenchmarkRegexSplit(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`regex.split("a,b;c:d", "[,;:]")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkRegexFindAll benchmarks regex.find_all
func BenchmarkRegexFindAll(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`regex.find_all("the cat sat on the mat", "\\b\\w{3}\\b")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Time Module Benchmarks
// =============================================================================

// BenchmarkTimeNow benchmarks time.now
func BenchmarkTimeNow(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("time.now()")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkTimeFormat benchmarks time.format
func BenchmarkTimeFormat(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`time.format(1704067200, "2006-01-02")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkTimeParse benchmarks time.parse
func BenchmarkTimeParse(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`time.parse("2024-01-01", "2006-01-02")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkTimeSince benchmarks time.since
func BenchmarkTimeSince(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("time.since(1704067200)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Base64 Module Benchmarks
// =============================================================================

// BenchmarkBase64Encode benchmarks base64.encode
func BenchmarkBase64Encode(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`base64.encode("hello world")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkBase64Decode benchmarks base64.decode
func BenchmarkBase64Decode(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`base64.decode("aGVsbG8gd29ybGQ=")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkBase64URLEncode benchmarks base64.url_encode
func BenchmarkBase64URLEncode(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`base64.url_encode("hello+world/test")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// URL Module Benchmarks
// =============================================================================

// BenchmarkURLParse benchmarks url.parse
func BenchmarkURLParse(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`url.parse("https://example.com:8080/path?query=value#fragment")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkURLEncode benchmarks url.encode
func BenchmarkURLEncode(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`url.encode("hello world & more")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkURLDecode benchmarks url.decode
func BenchmarkURLDecode(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`url.decode("hello%20world%20%26%20more")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkURLBuild benchmarks url.build
func BenchmarkURLBuild(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`url.build({"scheme": "https", "host": "example.com", "path": "/api/v1"})`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Environment Module Benchmarks
// =============================================================================

// BenchmarkEnvGet benchmarks env.get
func BenchmarkEnvGet(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`env.get("PATH")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkEnvGetOr benchmarks env.get_or
func BenchmarkEnvGetOr(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`env.get_or("NONEXISTENT_VAR", "default_value")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Security Module Benchmarks
// =============================================================================

// BenchmarkSecurityEscapeHTML benchmarks security.escape_html
func BenchmarkSecurityEscapeHTML(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`security.escape_html("<script>alert('xss')</script>")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkSecurityValidateEmail benchmarks security.validate_email
func BenchmarkSecurityValidateEmail(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`security.validate_email("test@example.com")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkSecurityValidateURL benchmarks security.validate_url
func BenchmarkSecurityValidateURL(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`security.validate_url("https://example.com/path")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkSecuritySanitize benchmarks security.sanitize
func BenchmarkSecuritySanitize(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`security.sanitize("Hello <script>World</script>!")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Crypto Module Benchmarks
// =============================================================================

// BenchmarkCryptoSHA256 benchmarks crypto.sha256
func BenchmarkCryptoSHA256(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`crypto.sha256("hello world")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkCryptoSHA512 benchmarks crypto.sha512
func BenchmarkCryptoSHA512(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`crypto.sha512("hello world")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkCryptoMD5 benchmarks crypto.md5
func BenchmarkCryptoMD5(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`crypto.md5("hello world")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkCryptoRandomBytes benchmarks crypto.random_bytes
func BenchmarkCryptoRandomBytes(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("crypto.random_bytes(32)")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkCryptoUUID benchmarks crypto.uuid
func BenchmarkCryptoUUID(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New("crypto.uuid()")
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkCryptoHMAC benchmarks crypto.hmac
func BenchmarkCryptoHMAC(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`crypto.hmac("sha256", "secret-key", "message to sign")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// String Module - Additional Benchmarks
// =============================================================================

// BenchmarkStrTrim benchmarks str.trim
func BenchmarkStrTrim(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`"  hello world  " > str.trim()`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkStrLower benchmarks str.lower
func BenchmarkStrLower(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`"HELLO WORLD" > str.lower()`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkStrJoin benchmarks str.join
func BenchmarkStrJoin(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`str.join(["a", "b", "c", "d", "e"], ", ")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkStrPadLeft benchmarks str.pad_left
func BenchmarkStrPadLeft(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`"42" > str.pad_left(5, "0")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkStrStartsWith benchmarks str.starts_with?
func BenchmarkStrStartsWith(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`"hello world" > str.starts_with?("hello")`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkStrSubstring benchmarks str.substring
func BenchmarkStrSubstring(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`"hello world" > str.substring(0, 5)`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkStrCharAt benchmarks str.char_at
func BenchmarkStrCharAt(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`"hello" > str.char_at(2)`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Parallel Processing Benchmarks
// =============================================================================

// BenchmarkParallelMap benchmarks parallel.map
func BenchmarkParallelMap(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`[1, 2, 3, 4, 5] > parallel.map(fn(x int) { x > mul(2) }(int))`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkParallelFilter benchmarks parallel.filter
func BenchmarkParallelFilter(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`[1, 2, 3, 4, 5, 6, 7, 8, 9, 10] > parallel.filter(fn(x int) { gt?(x, 5) }(bool))`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkParallelReduce benchmarks parallel.reduce
func BenchmarkParallelReduce(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`[1, 2, 3, 4, 5] > parallel.reduce(fn(acc int, x int) { acc > add(x) }(int), 0)`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkParallelEach benchmarks parallel.each
func BenchmarkParallelEach(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`[1, 2, 3, 4, 5] > parallel.each(fn(x int) { x > mul(2) }(int))`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkParallelAny benchmarks parallel.any?
func BenchmarkParallelAny(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`[1, 2, 3, 4, 5] > parallel.any?(fn(x int) { gt?(x, 3) }(bool))`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkParallelAll benchmarks parallel.all?
func BenchmarkParallelAll(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`[1, 2, 3, 4, 5] > parallel.all?(fn(x int) { gt?(x, 0) }(bool))`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkParallelFind benchmarks parallel.find
func BenchmarkParallelFind(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`[1, 2, 3, 4, 5] > parallel.find(fn(x int) { eq?(x, 3) }(bool))`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkParallelFlatMap benchmarks parallel.flat_map
func BenchmarkParallelFlatMap(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`[1, 2, 3] > parallel.flat_map(fn(x int) { [x, x > mul(2)] }(array))`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Error Handling - Additional Benchmarks
// =============================================================================

// BenchmarkErrContext benchmarks err_context
func BenchmarkErrContext(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`err("test") > err_add_context("key", "value") > err_context()`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// =============================================================================
// Complex Real-World Scenarios
// =============================================================================

// BenchmarkComplexDataTransform benchmarks complex data transformation
func BenchmarkComplexDataTransform(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`
{"users": [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}]}
> get("users")
> parallel.map(fn(u map) { u > set("active", true) }(map))
`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkStringProcessing benchmarks string processing pipeline
func BenchmarkStringProcessing(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`
"  Hello, World!  " > str.trim() > str.lower() > str.replace(",", "") > str.split(" ")
`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}

// BenchmarkNumericComputation benchmarks numeric computation
func BenchmarkNumericComputation(b *testing.B) {
	env := object.NewEnvironment()
	l := lexer.New(`
100 > add(50) > mul(2) > div(3) > sub(10) > math.abs() > math.floor()
`)
	p := parser.New(l)
	program := p.ParseProgram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(program, env)
	}
}
