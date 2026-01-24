package evaluator

import (
	"math"
	"testing"

	"gitlab.com/bark-lang/bark/object"
)

func TestMathSqrt(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`math.sqrt(4)`, 2.0},
		{`math.sqrt(9)`, 3.0},
		{`math.sqrt(2.0)`, math.Sqrt(2.0)},
		{`math.sqrt(0)`, 0.0},
		{`math.sqrt(1)`, 1.0},
		{`math.sqrt(16.0)`, 4.0},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestMathSqrtErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`math.sqrt()`, "wrong number of arguments. got=0, want=1"},
		{`math.sqrt(1, 2)`, "wrong number of arguments. got=2, want=1"},
		{`math.sqrt("a")`, "argument to `math.sqrt` must be INTEGER or FLOAT, got STRING"},
		{`math.sqrt(-1)`, "math.sqrt: cannot take square root of negative number"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestMathPow(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`math.pow(2, 3)`, 8.0},
		{`math.pow(2, 0)`, 1.0},
		{`math.pow(2.0, 3.0)`, 8.0},
		{`math.pow(10, 2)`, 100.0},
		{`math.pow(2, -1)`, 0.5},
		{`math.pow(9, 0.5)`, 3.0},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestMathPowErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`math.pow()`, "wrong number of arguments. got=0, want=2"},
		{`math.pow(1)`, "wrong number of arguments. got=1, want=2"},
		{`math.pow("a", 2)`, "first argument to `math.pow` must be INTEGER or FLOAT, got STRING"},
		{`math.pow(2, "b")`, "second argument to `math.pow` must be INTEGER or FLOAT, got STRING"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestMathExp(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`math.exp(0)`, 1.0},
		{`math.exp(1)`, math.E},
		{`math.exp(2)`, math.Exp(2)},
		{`math.exp(-1)`, math.Exp(-1)},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestMathExpErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`math.exp()`, "wrong number of arguments. got=0, want=1"},
		{`math.exp(1, 2)`, "wrong number of arguments. got=2, want=1"},
		{`math.exp("a")`, "argument to `math.exp` must be INTEGER or FLOAT, got STRING"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestMathLog(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`math.log(1)`, 0.0},
		{`math.log(2.718281828459045)`, 1.0},
		{`math.log(10)`, math.Log(10)},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestMathLogErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`math.log()`, "wrong number of arguments. got=0, want=1"},
		{`math.log(1, 2)`, "wrong number of arguments. got=2, want=1"},
		{`math.log("a")`, "argument to `math.log` must be INTEGER or FLOAT, got STRING"},
		{`math.log(0)`, "math.log: argument must be positive"},
		{`math.log(-1)`, "math.log: argument must be positive"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestMathLog10(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`math.log10(1)`, 0.0},
		{`math.log10(10)`, 1.0},
		{`math.log10(100)`, 2.0},
		{`math.log10(1000)`, 3.0},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestMathLog10Errors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`math.log10()`, "wrong number of arguments. got=0, want=1"},
		{`math.log10(1, 2)`, "wrong number of arguments. got=2, want=1"},
		{`math.log10("a")`, "argument to `math.log10` must be INTEGER or FLOAT, got STRING"},
		{`math.log10(0)`, "math.log10: argument must be positive"},
		{`math.log10(-1)`, "math.log10: argument must be positive"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestMathSin(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`math.sin(0)`, 0.0},
		{`math.sin(1.5707963267948966)`, 1.0}, // sin(π/2) = 1
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestMathSinErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`math.sin()`, "wrong number of arguments. got=0, want=1"},
		{`math.sin(1, 2)`, "wrong number of arguments. got=2, want=1"},
		{`math.sin("a")`, "argument to `math.sin` must be INTEGER or FLOAT, got STRING"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestMathCos(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`math.cos(0)`, 1.0},
		{`math.cos(3.141592653589793)`, -1.0}, // cos(π) = -1
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestMathCosErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`math.cos()`, "wrong number of arguments. got=0, want=1"},
		{`math.cos(1, 2)`, "wrong number of arguments. got=2, want=1"},
		{`math.cos("a")`, "argument to `math.cos` must be INTEGER or FLOAT, got STRING"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestMathTan(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`math.tan(0)`, 0.0},
		{`math.tan(0.7853981633974483)`, math.Tan(math.Pi / 4)}, // tan(π/4) ≈ 1
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestMathTanErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`math.tan()`, "wrong number of arguments. got=0, want=1"},
		{`math.tan(1, 2)`, "wrong number of arguments. got=2, want=1"},
		{`math.tan("a")`, "argument to `math.tan` must be INTEGER or FLOAT, got STRING"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestMathAsin(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`math.asin(0)`, 0.0},
		{`math.asin(1)`, math.Pi / 2},
		{`math.asin(-1)`, -math.Pi / 2},
		{`math.asin(0.5)`, math.Asin(0.5)},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestMathAsinErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`math.asin()`, "wrong number of arguments. got=0, want=1"},
		{`math.asin(1, 2)`, "wrong number of arguments. got=2, want=1"},
		{`math.asin("a")`, "argument to `math.asin` must be INTEGER or FLOAT, got STRING"},
		{`math.asin(2)`, "math.asin: argument must be in range [-1, 1]"},
		{`math.asin(-2)`, "math.asin: argument must be in range [-1, 1]"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestMathAcos(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`math.acos(1)`, 0.0},
		{`math.acos(0)`, math.Pi / 2},
		{`math.acos(-1)`, math.Pi},
		{`math.acos(0.5)`, math.Acos(0.5)},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestMathAcosErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`math.acos()`, "wrong number of arguments. got=0, want=1"},
		{`math.acos(1, 2)`, "wrong number of arguments. got=2, want=1"},
		{`math.acos("a")`, "argument to `math.acos` must be INTEGER or FLOAT, got STRING"},
		{`math.acos(2)`, "math.acos: argument must be in range [-1, 1]"},
		{`math.acos(-2)`, "math.acos: argument must be in range [-1, 1]"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestMathAtan(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`math.atan(0)`, 0.0},
		{`math.atan(1)`, math.Pi / 4},
		{`math.atan(-1)`, -math.Pi / 4},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestMathAtanErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`math.atan()`, "wrong number of arguments. got=0, want=1"},
		{`math.atan(1, 2)`, "wrong number of arguments. got=2, want=1"},
		{`math.atan("a")`, "argument to `math.atan` must be INTEGER or FLOAT, got STRING"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestMathPi(t *testing.T) {
	evaluated := testEval(`math.pi()`)
	testFloatObject(t, evaluated, math.Pi)
}

func TestMathPiErrors(t *testing.T) {
	evaluated := testEval(`math.pi(1)`)
	errObj, ok := evaluated.(*object.Error)
	if !ok {
		t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
		return
	}
	expected := "wrong number of arguments. got=1, want=0"
	if errObj.Msg != expected {
		t.Errorf("wrong error message. expected=%q, got=%q", expected, errObj.Msg)
	}
}

func TestMathE(t *testing.T) {
	evaluated := testEval(`math.e()`)
	testFloatObject(t, evaluated, math.E)
}

func TestMathEErrors(t *testing.T) {
	evaluated := testEval(`math.e(1)`)
	errObj, ok := evaluated.(*object.Error)
	if !ok {
		t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
		return
	}
	expected := "wrong number of arguments. got=1, want=0"
	if errObj.Msg != expected {
		t.Errorf("wrong error message. expected=%q, got=%q", expected, errObj.Msg)
	}
}

func TestMathOdd(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`math.odd?(1)`, true},
		{`math.odd?(2)`, false},
		{`math.odd?(3)`, true},
		{`math.odd?(0)`, false},
		{`math.odd?(-1)`, true},
		{`math.odd?(-2)`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestMathOddErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`math.odd?()`, "wrong number of arguments. got=0, want=1"},
		{`math.odd?(1, 2)`, "wrong number of arguments. got=2, want=1"},
		{`math.odd?("a")`, "argument to `math.odd?` must be INTEGER, got STRING"},
		{`math.odd?(1.5)`, "argument to `math.odd?` must be INTEGER, got FLOAT"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestMathEven(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`math.even?(2)`, true},
		{`math.even?(1)`, false},
		{`math.even?(4)`, true},
		{`math.even?(0)`, true},
		{`math.even?(-2)`, true},
		{`math.even?(-1)`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestMathEvenErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`math.even?()`, "wrong number of arguments. got=0, want=1"},
		{`math.even?(1, 2)`, "wrong number of arguments. got=2, want=1"},
		{`math.even?("a")`, "argument to `math.even?` must be INTEGER, got STRING"},
		{`math.even?(1.5)`, "argument to `math.even?` must be INTEGER, got FLOAT"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

// Test chaining with link operator
func TestMathChaining(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`4 > math.sqrt()`, 2.0},
		{`2 > math.pow(3)`, 8.0},
		{`0 > math.sin()`, 0.0},
		{`0 > math.cos()`, 1.0},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

// Test existing math functions still work
func TestMathExistingFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`math.abs(-5)`, 5.0},
		{`math.abs(-3.14)`, 3.14},
		{`math.ceil(3.2)`, 4.0},
		{`math.floor(3.8)`, 3.0},
		{`math.round(3.456, 2)`, 3.46},
		{`math.min(1, 2, 3)`, 1.0},
		{`math.max(1, 2, 3)`, 3.0},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		floatObj, ok := evaluated.(*object.Float)
		if ok {
			if floatObj.Value != tt.expected {
				t.Errorf("wrong value for %q. expected=%f, got=%f", tt.input, tt.expected, floatObj.Value)
			}
			continue
		}
		intObj, ok := evaluated.(*object.Integer)
		if ok {
			if float64(intObj.Value) != tt.expected {
				t.Errorf("wrong value for %q. expected=%f, got=%d", tt.input, tt.expected, intObj.Value)
			}
			continue
		}
		t.Errorf("expected Float or Integer object for %q, got=%T (%+v)", tt.input, evaluated, evaluated)
	}
}

func TestMathMod(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`math.mod(10, 3)`, 1},
		{`math.mod(15, 5)`, 0},
		{`math.mod(7, 2)`, 1},
		{`math.mod(-10, 3)`, -1},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestMathToInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`math.to_int(3.7)`, 3},
		{`math.to_int(3.2)`, 3},
		{`math.to_int(-3.7)`, -3},
		{`math.to_int(-3.2)`, -3},
		{`math.to_int(0.0)`, 0},
		{`math.to_int(5)`, 5},     // Integer passthrough
		{`math.to_int(-10)`, -10}, // Integer passthrough
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestMathToIntErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`math.to_int()`, "wrong number of arguments. got=0, want=1"},
		{`math.to_int(1, 2)`, "wrong number of arguments. got=2, want=1"},
		{`math.to_int("a")`, "argument to `math.to_int` must be FLOAT or INTEGER, got STRING"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestMathToFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`math.to_float(5)`, 5.0},
		{`math.to_float(-10)`, -10.0},
		{`math.to_float(0)`, 0.0},
		{`math.to_float(3.14)`, 3.14}, // Float passthrough
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestMathToFloatErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`math.to_float()`, "wrong number of arguments. got=0, want=1"},
		{`math.to_float(1, 2)`, "wrong number of arguments. got=2, want=1"},
		{`math.to_float("a")`, "argument to `math.to_float` must be INTEGER or FLOAT, got STRING"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object, got=%T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Msg)
		}
	}
}

func TestMathToIntChaining(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`9 > math.sqrt() > math.to_int()`, 3},
		{`3.7 > math.to_int()`, 3},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}
