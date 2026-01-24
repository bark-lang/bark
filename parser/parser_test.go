package parser

import (
	"testing"

	"gitlab.com/bark-lang/bark/ast"
	"gitlab.com/bark-lang/bark/lexer"
)

func TestModuleStatement(t *testing.T) {
	input := `module utils`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ModuleStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ModuleStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Name.Value != "utils" {
		t.Errorf("stmt.Name.Value not 'utils'. got=%s", stmt.Name.Value)
	}
}

func TestIncludeStatement(t *testing.T) {
	input := `include "utils/strings"`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IncludeStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.IncludeStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Path.Value != "utils/strings" {
		t.Errorf("stmt.Path.Value not 'utils/strings'. got=%s", stmt.Path.Value)
	}
}

func TestImportStatement(t *testing.T) {
	tests := []struct {
		input         string
		expectedPath  string
		expectedAlias string
		hasAlias      bool
	}{
		{`import "lib/math"`, "lib/math", "", false},
		{`import "lib/utils/strings" as str`, "lib/utils/strings", "str", true},
		{`import "https://modules.bark-lang.org/json@v1.2.3" as json`, "https://modules.bark-lang.org/json@v1.2.3", "json", true},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statement. got=%d",
				len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ImportStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ImportStatement. got=%T",
				program.Statements[0])
		}

		if stmt.Path.Value != tt.expectedPath {
			t.Errorf("stmt.Path.Value not '%s'. got=%s", tt.expectedPath, stmt.Path.Value)
		}

		if tt.hasAlias {
			if stmt.Alias == nil {
				t.Errorf("stmt.Alias is nil, expected '%s'", tt.expectedAlias)
			} else if stmt.Alias.Value != tt.expectedAlias {
				t.Errorf("stmt.Alias.Value not '%s'. got=%s", tt.expectedAlias, stmt.Alias.Value)
			}
		} else {
			if stmt.Alias != nil {
				t.Errorf("stmt.Alias should be nil. got=%s", stmt.Alias.Value)
			}
		}
	}
}

func TestFunctionStatement(t *testing.T) {
	input := `
fn add(a int, b int) {
  a > add(b) > result
  return(result)
}(int)
`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.FunctionStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Name.Value != "add" {
		t.Errorf("stmt.Name.Value not 'add'. got=%s", stmt.Name.Value)
	}

	if len(stmt.Parameters) != 2 {
		t.Fatalf("function has wrong number of parameters. got=%d", len(stmt.Parameters))
	}

	if stmt.Parameters[0].Name.Value != "a" {
		t.Errorf("parameter[0] name wrong. got=%s", stmt.Parameters[0].Name.Value)
	}

	if stmt.Parameters[0].Type.Name != "int" {
		t.Errorf("parameter[0] type wrong. got=%s", stmt.Parameters[0].Type.Name)
	}

	if stmt.ReturnType == nil {
		t.Fatal("stmt.ReturnType is nil")
	}

	if len(stmt.ReturnType.Types) != 1 {
		t.Fatalf("wrong number of return types. got=%d", len(stmt.ReturnType.Types))
	}

	if stmt.ReturnType.Types[0].Name != "int" {
		t.Errorf("return type wrong. got=%s", stmt.ReturnType.Types[0].Name)
	}
}

func TestPublicFunctionStatement(t *testing.T) {
	input := `pub fn hello() { stdout("hello") }`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.FunctionStatement. got=%T",
			program.Statements[0])
	}

	if !stmt.Public {
		t.Error("function should be public")
	}

	if stmt.Name.Value != "hello" {
		t.Errorf("stmt.Name.Value not 'hello'. got=%s", stmt.Name.Value)
	}
}

func TestIntegerLiteral(t *testing.T) {
	input := "42"

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != 42 {
		t.Errorf("literal.Value not %d. got=%d", 42, literal.Value)
	}
}

func TestFloatLiteral(t *testing.T) {
	input := "3.14"

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("exp not *ast.FloatLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != 3.14 {
		t.Errorf("literal.Value not %f. got=%f", 3.14, literal.Value)
	}
}

func TestStringLiteral(t *testing.T) {
	input := `"hello world"`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != "hello world" {
		t.Errorf("literal.Value not %q. got=%q", "hello world", literal.Value)
	}
}

func TestBooleanLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		boolean, ok := stmt.Expression.(*ast.BooleanLiteral)
		if !ok {
			t.Fatalf("exp not *ast.BooleanLiteral. got=%T", stmt.Expression)
		}

		if boolean.Value != tt.expected {
			t.Errorf("boolean.Value not %t. got=%t", tt.expected, boolean.Value)
		}
	}
}

func TestArrayLiteral(t *testing.T) {
	input := `[1, 2, 3]`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("exp not *ast.ArrayLiteral. got=%T", stmt.Expression)
	}

	if len(array.Elements) != 3 {
		t.Fatalf("array has wrong number of elements. got=%d", len(array.Elements))
	}

	testIntegerLiteral(t, array.Elements[0], 1)
	testIntegerLiteral(t, array.Elements[1], 2)
	testIntegerLiteral(t, array.Elements[2], 3)
}

func TestMapLiteral(t *testing.T) {
	input := `{"name": "John", "age": "30"}`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	mapLit, ok := stmt.Expression.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("exp not *ast.MapLiteral. got=%T", stmt.Expression)
	}

	if len(mapLit.Pairs) != 2 {
		t.Fatalf("map has wrong number of pairs. got=%d", len(mapLit.Pairs))
	}
}

func TestCallExpression(t *testing.T) {
	input := `add(1, 2)`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("exp not *ast.CallExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, call.Function, "add") {
		return
	}

	if len(call.Arguments) != 2 {
		t.Fatalf("wrong number of arguments. got=%d", len(call.Arguments))
	}

	testIntegerLiteral(t, call.Arguments[0], 1)
	testIntegerLiteral(t, call.Arguments[1], 2)
}

func TestLinkExpression(t *testing.T) {
	input := `42 > add(10)`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	link, ok := stmt.Expression.(*ast.LinkExpression)
	if !ok {
		t.Fatalf("exp not *ast.LinkExpression. got=%T", stmt.Expression)
	}

	if !testIntegerLiteral(t, link.Left, 42) {
		return
	}

	call, ok := link.Right.(*ast.CallExpression)
	if !ok {
		t.Fatalf("link.Right not *ast.CallExpression. got=%T", link.Right)
	}

	if !testIdentifier(t, call.Function, "add") {
		return
	}
}

func TestChainedLinkExpression(t *testing.T) {
	input := `"hello" > uppercase() > stdout()`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)

	// The result should be: ("hello" > uppercase()) > stdout()
	outerLink, ok := stmt.Expression.(*ast.LinkExpression)
	if !ok {
		t.Fatalf("exp not *ast.LinkExpression. got=%T", stmt.Expression)
	}

	// Left should be another link expression
	innerLink, ok := outerLink.Left.(*ast.LinkExpression)
	if !ok {
		t.Fatalf("outerLink.Left not *ast.LinkExpression. got=%T", outerLink.Left)
	}

	// Innermost left should be "hello"
	if !testStringLiteral(t, innerLink.Left, "hello") {
		return
	}

	// Inner right should be uppercase()
	call1, ok := innerLink.Right.(*ast.CallExpression)
	if !ok {
		t.Fatalf("innerLink.Right not *ast.CallExpression. got=%T", innerLink.Right)
	}
	if !testIdentifier(t, call1.Function, "uppercase") {
		return
	}

	// Outer right should be stdout()
	call2, ok := outerLink.Right.(*ast.CallExpression)
	if !ok {
		t.Fatalf("outerLink.Right not *ast.CallExpression. got=%T", outerLink.Right)
	}
	if !testIdentifier(t, call2.Function, "stdout") {
		return
	}
}

func TestMemberExpression(t *testing.T) {
	input := `math.sqrt`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	member, ok := stmt.Expression.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("exp not *ast.MemberExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, member.Object, "math") {
		return
	}

	if member.Member.Value != "sqrt" {
		t.Errorf("member.Member.Value not 'sqrt'. got=%s", member.Member.Value)
	}
}

func TestMemberExpressionWithCall(t *testing.T) {
	input := `math.sqrt(16)`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("exp not *ast.CallExpression. got=%T", stmt.Expression)
	}

	member, ok := call.Function.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("call.Function not *ast.MemberExpression. got=%T", call.Function)
	}

	if !testIdentifier(t, member.Object, "math") {
		return
	}

	if member.Member.Value != "sqrt" {
		t.Errorf("member.Member.Value not 'sqrt'. got=%s", member.Member.Value)
	}

	if len(call.Arguments) != 1 {
		t.Fatalf("wrong number of arguments. got=%d", len(call.Arguments))
	}

	testIntegerLiteral(t, call.Arguments[0], 16)
}

func TestTupleDestructure(t *testing.T) {
	input := `validate(data) > (err, result)`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	tuple, ok := stmt.Expression.(*ast.TupleDestructure)
	if !ok {
		t.Fatalf("exp not *ast.TupleDestructure. got=%T", stmt.Expression)
	}

	// Expression should be validate(data)
	call, ok := tuple.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("tuple.Expression not *ast.CallExpression. got=%T", tuple.Expression)
	}

	if !testIdentifier(t, call.Function, "validate") {
		return
	}

	// Should have 2 identifiers
	if len(tuple.Identifiers) != 2 {
		t.Fatalf("wrong number of identifiers. got=%d", len(tuple.Identifiers))
	}

	if tuple.Identifiers[0].Value != "err" {
		t.Errorf("identifier[0] not 'err'. got=%s", tuple.Identifiers[0].Value)
	}

	if tuple.Identifiers[1].Value != "result" {
		t.Errorf("identifier[1] not 'result'. got=%s", tuple.Identifiers[1].Value)
	}
}

func TestAnonymousFunction(t *testing.T) {
	input := `(x int) { x > add(1) > return() }(int)`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	fn, ok := stmt.Expression.(*ast.AnonymousFunction)
	if !ok {
		t.Fatalf("exp not *ast.AnonymousFunction. got=%T", stmt.Expression)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("wrong number of parameters. got=%d", len(fn.Parameters))
	}

	if fn.Parameters[0].Name.Value != "x" {
		t.Errorf("parameter name not 'x'. got=%s", fn.Parameters[0].Name.Value)
	}

	if fn.Parameters[0].Type.Name != "int" {
		t.Errorf("parameter type not 'int'. got=%s", fn.Parameters[0].Type.Name)
	}

	if fn.ReturnType == nil {
		t.Fatal("return type is nil")
	}

	if len(fn.ReturnType.Types) != 1 {
		t.Fatalf("wrong number of return types. got=%d", len(fn.ReturnType.Types))
	}

	if fn.ReturnType.Types[0].Name != "int" {
		t.Errorf("return type not 'int'. got=%s", fn.ReturnType.Types[0].Name)
	}
}

func TestNegativeNumberLiterals(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"-5", int64(-5)},
		{"-42", int64(-42)},
		{"-3.14", float64(-3.14)},
		{"-0.5", float64(-0.5)},
		{"-1_000", int64(-1000)},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program has not enough statements. got=%d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		switch expected := tt.expectedValue.(type) {
		case int64:
			literal, ok := stmt.Expression.(*ast.IntegerLiteral)
			if !ok {
				t.Fatalf("exp not *ast.IntegerLiteral for input %s. got=%T", tt.input, stmt.Expression)
			}
			if literal.Value != expected {
				t.Errorf("literal.Value not %d for input %s. got=%d", expected, tt.input, literal.Value)
			}
		case float64:
			literal, ok := stmt.Expression.(*ast.FloatLiteral)
			if !ok {
				t.Fatalf("exp not *ast.FloatLiteral for input %s. got=%T", tt.input, stmt.Expression)
			}
			if literal.Value != expected {
				t.Errorf("literal.Value not %f for input %s. got=%f", expected, tt.input, literal.Value)
			}
		}
	}
}

func TestTupleExpression(t *testing.T) {
	tests := []struct {
		input           string
		expectedCount   int
		expectedStrings []string
	}{
		{"(1, 2)", 2, []string{"1", "2"}},
		{"(x, y, z)", 3, []string{"x", "y", "z"}},
		{"(1, 2, 3, 4)", 4, []string{"1", "2", "3", "4"}},
		{"(a, 1)", 2, []string{"a", "1"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program has wrong number of statements for input %s. got=%d",
				tt.input, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement for input %s. got=%T",
				tt.input, program.Statements[0])
		}

		tuple, ok := stmt.Expression.(*ast.TupleExpression)
		if !ok {
			t.Fatalf("exp not *ast.TupleExpression for input %s. got=%T",
				tt.input, stmt.Expression)
		}

		if len(tuple.Elements) != tt.expectedCount {
			t.Fatalf("tuple has wrong number of elements for input %s. expected=%d, got=%d",
				tt.input, tt.expectedCount, len(tuple.Elements))
		}

		for i, expectedStr := range tt.expectedStrings {
			if tuple.Elements[i].String() != expectedStr {
				t.Errorf("tuple.Elements[%d] wrong for input %s. expected=%s, got=%s",
					i, tt.input, expectedStr, tuple.Elements[i].String())
			}
		}
	}
}

func TestTupleExpressionWithExpressions(t *testing.T) {
	// Test tuple with more complex expressions
	input := "(x, 1)"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	tuple, ok := stmt.Expression.(*ast.TupleExpression)
	if !ok {
		t.Fatalf("exp not *ast.TupleExpression. got=%T", stmt.Expression)
	}

	if len(tuple.Elements) != 2 {
		t.Fatalf("tuple has wrong number of elements. expected=2, got=%d", len(tuple.Elements))
	}

	// First element should be identifier
	if !testIdentifier(t, tuple.Elements[0], "x") {
		return
	}

	// Second element should be integer
	if !testIntegerLiteral(t, tuple.Elements[1], 1) {
		return
	}
}

func TestTupleVsGroupedExpression(t *testing.T) {
	// Single expression in parens should NOT be a tuple
	input := "(5)"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)

	// Should be IntegerLiteral, not TupleExpression
	_, isTuple := stmt.Expression.(*ast.TupleExpression)
	if isTuple {
		t.Fatalf("(5) should NOT parse as TupleExpression")
	}

	intLit, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("(5) should parse as IntegerLiteral. got=%T", stmt.Expression)
	}

	if intLit.Value != 5 {
		t.Errorf("intLit.Value wrong. expected=5, got=%d", intLit.Value)
	}
}

func TestTupleVsAnonymousFunction(t *testing.T) {
	// (param type) { } should be anonymous function, not tuple
	input := "(x int) { return(x) }(int)"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)

	// Should be AnonymousFunction, not TupleExpression
	_, isTuple := stmt.Expression.(*ast.TupleExpression)
	if isTuple {
		t.Fatalf("anonymous function should NOT parse as TupleExpression")
	}

	_, ok := stmt.Expression.(*ast.AnonymousFunction)
	if !ok {
		t.Fatalf("should parse as AnonymousFunction. got=%T", stmt.Expression)
	}
}

func TestTupleString(t *testing.T) {
	input := "(1, 2, 3)"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	tuple := stmt.Expression.(*ast.TupleExpression)

	expected := "(1, 2, 3)"
	if tuple.String() != expected {
		t.Errorf("tuple.String() wrong. expected=%q, got=%q", expected, tuple.String())
	}
}

func TestCaptureExpression(t *testing.T) {
	input := `data > riskyOp() > capture(err, result)`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	// The outer expression is a link: (data > riskyOp()) > capture(err, result)
	link, ok := stmt.Expression.(*ast.LinkExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.LinkExpression. got=%T", stmt.Expression)
	}

	// The right side should be a CaptureExpression
	capture, ok := link.Right.(*ast.CaptureExpression)
	if !ok {
		t.Fatalf("link.Right is not ast.CaptureExpression. got=%T", link.Right)
	}

	if capture.ErrorVar.Value != "err" {
		t.Errorf("capture.ErrorVar.Value not 'err'. got=%s", capture.ErrorVar.Value)
	}

	if capture.ResultVar.Value != "result" {
		t.Errorf("capture.ResultVar.Value not 'result'. got=%s", capture.ResultVar.Value)
	}
}

func TestCaptureExpressionInChain(t *testing.T) {
	input := `data > op1() > capture(e1, r1) > op2() > capture(e2, r2) > final()`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	// Just check that it parses without errors
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Expression == nil {
		t.Fatalf("stmt.Expression is nil")
	}
}

func TestCaptureExpressionString(t *testing.T) {
	capture := &ast.CaptureExpression{
		ErrorVar:  &ast.Identifier{Value: "err"},
		ResultVar: &ast.Identifier{Value: "result"},
	}

	expected := "capture(err, result)"
	if capture.String() != expected {
		t.Errorf("capture.String() wrong. expected=%q, got=%q", expected, capture.String())
	}
}

// Helper functions

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func testIntegerLiteral(t *testing.T, exp ast.Expression, value int64) bool {
	integ, ok := exp.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("exp not *ast.IntegerLiteral. got=%T", exp)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}

	return true
}

func testStringLiteral(t *testing.T, exp ast.Expression, value string) bool {
	str, ok := exp.(*ast.StringLiteral)
	if !ok {
		t.Errorf("exp not *ast.StringLiteral. got=%T", exp)
		return false
	}

	if str.Value != value {
		t.Errorf("str.Value not %q. got=%q", value, str.Value)
		return false
	}

	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}

	return true
}
