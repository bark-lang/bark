package ast

import (
	"testing"

	"gitlab.com/bark-lang/bark/token"
)

func TestString(t *testing.T) {
	// Test that AST nodes can be converted to strings correctly
	program := &Program{
		Statements: []Statement{
			&FunctionStatement{
				Token: token.Token{Type: token.FN, Literal: "fn"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "add"},
					Value: "add",
				},
				Parameters: []*Parameter{
					{
						Token: token.Token{Type: token.IDENT, Literal: "a"},
						Name: &Identifier{
							Token: token.Token{Type: token.IDENT, Literal: "a"},
							Value: "a",
						},
						Type: &TypeExpression{
							Token: token.Token{Type: token.IDENT, Literal: "int"},
							Name:  "int",
						},
					},
					{
						Token: token.Token{Type: token.IDENT, Literal: "b"},
						Name: &Identifier{
							Token: token.Token{Type: token.IDENT, Literal: "b"},
							Value: "b",
						},
						Type: &TypeExpression{
							Token: token.Token{Type: token.IDENT, Literal: "int"},
							Name:  "int",
						},
					},
				},
				Body: &BlockStatement{
					Token: token.Token{Type: token.LBRACE, Literal: "{"},
					Statements: []Statement{
						&ExpressionStatement{
							Token: token.Token{Type: token.IDENT, Literal: "a"},
							Expression: &LinkExpression{
								Token: token.Token{Type: token.GT, Literal: ">"},
								Left: &Identifier{
									Token: token.Token{Type: token.IDENT, Literal: "a"},
									Value: "a",
								},
								Right: &CallExpression{
									Token: token.Token{Type: token.LPAREN, Literal: "("},
									Function: &Identifier{
										Token: token.Token{Type: token.IDENT, Literal: "add"},
										Value: "add",
									},
									Arguments: []Expression{
										&Identifier{
											Token: token.Token{Type: token.IDENT, Literal: "b"},
											Value: "b",
										},
									},
								},
							},
						},
					},
				},
				ReturnType: &TypeList{
					Token: token.Token{Type: token.LPAREN, Literal: "("},
					Types: []*TypeExpression{
						{
							Token: token.Token{Type: token.IDENT, Literal: "int"},
							Name:  "int",
						},
					},
				},
			},
		},
	}

	// The output should represent a valid bark function (though formatted)
	output := program.String()

	// Check that key elements are present in the output
	if !contains(output, "fn add") {
		t.Errorf("Expected function name 'fn add', got: %s", output)
	}
	if !contains(output, "a int") {
		t.Errorf("Expected parameter 'a int', got: %s", output)
	}
	if !contains(output, "b int") {
		t.Errorf("Expected parameter 'b int', got: %s", output)
	}
	if !contains(output, "(int)") {
		t.Errorf("Expected return type '(int)', got: %s", output)
	}
}

func TestIdentifier(t *testing.T) {
	ident := &Identifier{
		Token: token.Token{Type: token.IDENT, Literal: "myVariable"},
		Value: "myVariable",
	}

	if ident.TokenLiteral() != "myVariable" {
		t.Errorf("TokenLiteral() wrong. got=%q", ident.TokenLiteral())
	}

	if ident.String() != "myVariable" {
		t.Errorf("String() wrong. got=%q", ident.String())
	}
}

func TestIntegerLiteral(t *testing.T) {
	intLit := &IntegerLiteral{
		Token: token.Token{Type: token.INT, Literal: "42"},
		Value: 42,
	}

	if intLit.TokenLiteral() != "42" {
		t.Errorf("TokenLiteral() wrong. got=%q", intLit.TokenLiteral())
	}

	if intLit.String() != "42" {
		t.Errorf("String() wrong. got=%q", intLit.String())
	}
}

func TestFloatLiteral(t *testing.T) {
	floatLit := &FloatLiteral{
		Token: token.Token{Type: token.FLOAT, Literal: "3.14"},
		Value: 3.14,
	}

	if floatLit.TokenLiteral() != "3.14" {
		t.Errorf("TokenLiteral() wrong. got=%q", floatLit.TokenLiteral())
	}

	if floatLit.String() != "3.14" {
		t.Errorf("String() wrong. got=%q", floatLit.String())
	}
}

func TestStringLiteral(t *testing.T) {
	strLit := &StringLiteral{
		Token: token.Token{Type: token.STRING, Literal: "hello"},
		Value: "hello",
	}

	if strLit.TokenLiteral() != "hello" {
		t.Errorf("TokenLiteral() wrong. got=%q", strLit.TokenLiteral())
	}

	if strLit.String() != `"hello"` {
		t.Errorf("String() wrong. got=%q", strLit.String())
	}
}

func TestBooleanLiteral(t *testing.T) {
	tests := []struct {
		token    token.Token
		value    bool
		expected string
	}{
		{token.Token{Type: token.TRUE, Literal: "true"}, true, "true"},
		{token.Token{Type: token.FALSE, Literal: "false"}, false, "false"},
	}

	for _, tt := range tests {
		boolLit := &BooleanLiteral{
			Token: tt.token,
			Value: tt.value,
		}

		if boolLit.String() != tt.expected {
			t.Errorf("String() wrong. got=%q, want=%q", boolLit.String(), tt.expected)
		}
	}
}

func TestArrayLiteral(t *testing.T) {
	arrayLit := &ArrayLiteral{
		Token: token.Token{Type: token.LBRACK, Literal: "["},
		Elements: []Expression{
			&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "1"},
				Value: 1,
			},
			&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "2"},
				Value: 2,
			},
			&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "3"},
				Value: 3,
			},
		},
	}

	expected := "[1, 2, 3]"
	if arrayLit.String() != expected {
		t.Errorf("String() wrong. got=%q, want=%q", arrayLit.String(), expected)
	}
}

func TestMapLiteral(t *testing.T) {
	keyExpr := &StringLiteral{
		Token: token.Token{Type: token.STRING, Literal: "name"},
		Value: "name",
	}
	valueExpr := &StringLiteral{
		Token: token.Token{Type: token.STRING, Literal: "John"},
		Value: "John",
	}

	mapLit := &MapLiteral{
		Token: token.Token{Type: token.LBRACE, Literal: "{"},
		Pairs: map[Expression]Expression{
			keyExpr: valueExpr,
		},
		OrderedKeys: []Expression{keyExpr},
	}

	output := mapLit.String()
	// Map iteration order is now guaranteed by OrderedKeys
	if !contains(output, `"name"`) || !contains(output, `"John"`) {
		t.Errorf("String() wrong. got=%q", output)
	}
}

func TestCallExpression(t *testing.T) {
	call := &CallExpression{
		Token: token.Token{Type: token.LPAREN, Literal: "("},
		Function: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "add"},
			Value: "add",
		},
		Arguments: []Expression{
			&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "1"},
				Value: 1,
			},
			&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "2"},
				Value: 2,
			},
		},
	}

	expected := "add(1, 2)"
	if call.String() != expected {
		t.Errorf("String() wrong. got=%q, want=%q", call.String(), expected)
	}
}

func TestLinkExpression(t *testing.T) {
	link := &LinkExpression{
		Token: token.Token{Type: token.GT, Literal: ">"},
		Left: &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "42"},
			Value: 42,
		},
		Right: &CallExpression{
			Token: token.Token{Type: token.LPAREN, Literal: "("},
			Function: &Identifier{
				Token: token.Token{Type: token.IDENT, Literal: "add"},
				Value: "add",
			},
			Arguments: []Expression{
				&IntegerLiteral{
					Token: token.Token{Type: token.INT, Literal: "10"},
					Value: 10,
				},
			},
		},
	}

	expected := "(42 > add(10))"
	if link.String() != expected {
		t.Errorf("String() wrong. got=%q, want=%q", link.String(), expected)
	}
}

func TestMemberExpression(t *testing.T) {
	member := &MemberExpression{
		Token: token.Token{Type: token.DOT, Literal: "."},
		Object: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "math"},
			Value: "math",
		},
		Member: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "sqrt"},
			Value: "sqrt",
		},
	}

	expected := "math.sqrt"
	if member.String() != expected {
		t.Errorf("String() wrong. got=%q, want=%q", member.String(), expected)
	}
}

func TestModuleStatement(t *testing.T) {
	module := &ModuleStatement{
		Token: token.Token{Type: token.MODULE, Literal: "module"},
		Name: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "utils"},
			Value: "utils",
		},
	}

	expected := "module utils"
	if module.String() != expected {
		t.Errorf("String() wrong. got=%q, want=%q", module.String(), expected)
	}
}

func TestIncludeStatement(t *testing.T) {
	include := &IncludeStatement{
		Token: token.Token{Type: token.INCLUDE, Literal: "include"},
		Path: &StringLiteral{
			Token: token.Token{Type: token.STRING, Literal: "utils/strings"},
			Value: "utils/strings",
		},
	}

	expected := `include "utils/strings"`
	if include.String() != expected {
		t.Errorf("String() wrong. got=%q, want=%q", include.String(), expected)
	}
}

func TestAnonymousFunction(t *testing.T) {
	anonFn := &AnonymousFunction{
		Token: token.Token{Type: token.LPAREN, Literal: "("},
		Parameters: []*Parameter{
			{
				Token: token.Token{Type: token.IDENT, Literal: "x"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "x"},
					Value: "x",
				},
				Type: &TypeExpression{
					Token: token.Token{Type: token.IDENT, Literal: "int"},
					Name:  "int",
				},
			},
		},
		Body: &BlockStatement{
			Token:      token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{},
		},
		ReturnType: &TypeList{
			Token: token.Token{Type: token.LPAREN, Literal: "("},
			Types: []*TypeExpression{
				{
					Token: token.Token{Type: token.IDENT, Literal: "int"},
					Name:  "int",
				},
			},
		},
	}

	output := anonFn.String()
	if !contains(output, "x int") {
		t.Errorf("Expected parameter 'x int', got: %s", output)
	}
	if !contains(output, "(int)") {
		t.Errorf("Expected return type '(int)', got: %s", output)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
