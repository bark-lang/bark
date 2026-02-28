package bytecode

import (
	"testing"

	"gitlab.com/bark-lang/barki/lexer"
	"gitlab.com/bark-lang/barki/object"
	"gitlab.com/bark-lang/barki/parser"
)

// TestConstantFunctionOptimization verifies that constant functions are optimized
// to store their value directly instead of creating a closure.
func TestConstantFunctionOptimization(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		shouldOptimize bool // true if we expect no OpClosure
	}{
		{
			name:           "integer constant",
			input:          `fn answer() { return(42) }(int)`,
			shouldOptimize: true,
		},
		{
			name:           "float constant",
			input:          `fn pi() { return(3.14159) }(float)`,
			shouldOptimize: true,
		},
		{
			name:           "string constant",
			input:          `fn greeting() { return("hello") }(string)`,
			shouldOptimize: true,
		},
		{
			name:           "boolean constant",
			input:          `fn yes() { return(true) }(bool)`,
			shouldOptimize: true,
		},
		{
			name:           "null constant",
			input:          `fn nothing() { return(null) }(null)`,
			shouldOptimize: true,
		},
		{
			name:           "array constant",
			input:          `fn days() { return(["mon", "tue", "wed"]) }(array)`,
			shouldOptimize: true,
		},
		{
			name:           "map constant",
			input:          `fn config() { return({"key": "value"}) }(map)`,
			shouldOptimize: true,
		},
		{
			name:           "non-constant - has parameters",
			input:          `fn double(x int) { x > mul(2) }(int)`,
			shouldOptimize: false,
		},
		{
			name:           "non-constant - references builtin",
			input:          `fn getLen() { "test" > len() }(int)`,
			shouldOptimize: false,
		},
		{
			name:           "non-constant - multiple statements",
			input:          `fn multi() { 1 > x return(x) }(int)`,
			shouldOptimize: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasOpClosure := compileAndCheckForClosure(t, tt.input)

			if tt.shouldOptimize && hasOpClosure {
				t.Errorf("expected constant function to be optimized (no OpClosure), but OpClosure was found")
			}
			if !tt.shouldOptimize && !hasOpClosure {
				t.Errorf("expected non-constant function to have OpClosure, but none was found")
			}
		})
	}
}

// compileAndCheckForClosure compiles the input and returns true if OpClosure is found
func compileAndCheckForClosure(t *testing.T, input string) bool {
	t.Helper()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	result := Compile(program)
	if len(result.Errors) != 0 {
		t.Fatalf("compiler errors: %v", result.Errors)
	}

	// Check if OpClosure was emitted
	chunk := result.Function.Chunk
	for i := 0; i < len(chunk.Code); {
		op := OpCode(chunk.Code[i])
		if op == OpClosure {
			return true
		}
		// Advance based on instruction size
		i += opCodeSize(op)
	}
	return false
}

// TestConstantFunctionValueStorage verifies that constant function values
// are stored in the constants pool correctly.
func TestConstantFunctionValueStorage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:     "integer stored as constant",
			input:    `fn answer() { return(42) }(int)`,
			expected: int64(42),
		},
		{
			name:     "float stored as constant",
			input:    `fn pi() { return(3.14159) }(float)`,
			expected: 3.14159,
		},
		{
			name:     "string stored as constant",
			input:    `fn greeting() { return("hello") }(string)`,
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			result := Compile(program)
			if len(result.Errors) != 0 {
				t.Fatalf("compiler errors: %v", result.Errors)
			}

			// The constant value should be in the constants pool
			found := false
			for _, c := range result.Function.Chunk.Constants {
				switch exp := tt.expected.(type) {
				case int64:
					if intObj, ok := c.(*object.Integer); ok && intObj.Value == exp {
						found = true
					}
				case float64:
					if floatObj, ok := c.(*object.Float); ok && floatObj.Value == exp {
						found = true
					}
				case string:
					if strObj, ok := c.(*object.String); ok && strObj.Value == exp {
						found = true
					}
				}
			}

			if !found {
				t.Errorf("expected constant %v to be in constants pool", tt.expected)
			}
		})
	}
}

// opCodeSize returns the total size of an instruction including operands
func opCodeSize(op OpCode) int {
	switch op {
	case OpConstant, OpClosure, OpStoreGlobal, OpLoadGlobal, OpArray, OpMap,
		OpLoadLocal, OpStoreLocal, OpLoadUpval, OpStoreUpval, OpMember,
		OpJump, OpJumpIfFalse, OpLoop, OpTuple, OpLinkBind, OpInterpolate:
		return 3 // opcode + 2-byte operand
	case OpBuiltin, OpUnpackBuiltin:
		return 4 // opcode + 2-byte name index + 1-byte arg count
	case OpCall, OpLinkCall, OpUnpackCall:
		return 2 // opcode + 1-byte arg count
	default:
		return 1 // opcode only
	}
}
