package modules

import (
	"math"

	"gitlab.com/bark-lang/barki/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/barki/object"
)

// InitMath initializes mathematical operations
func InitMath() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"math.mod": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("wrong number of arguments. got=%d, want=2", len(args))
				}

				if left, ok := args[0].(*object.Integer); ok {
					if right, ok := args[1].(*object.Integer); ok {
						if right.Value == 0 {
							return helpers.NewError("modulo by zero")
						}
						return &object.Integer{Value: left.Value % right.Value}
					}
				}

				return helpers.NewError("arguments to `math.mod` must be INTEGER")
			},
		},

		"math.abs": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				if arg, ok := args[0].(*object.Integer); ok {
					if arg.Value < 0 {
						return &object.Integer{Value: -arg.Value}
					}
					return arg
				}

				if arg, ok := args[0].(*object.Float); ok {
					return &object.Float{Value: math.Abs(arg.Value)}
				}

				return helpers.NewError("argument to `math.abs` must be INTEGER or FLOAT, got %s", args[0].Type())
			},
		},

		"math.ceil": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				if arg, ok := args[0].(*object.Float); ok {
					return &object.Float{Value: math.Ceil(arg.Value)}
				}

				if arg, ok := args[0].(*object.Integer); ok {
					return arg // Already an integer
				}

				return helpers.NewError("argument to `math.ceil` must be FLOAT or INTEGER, got %s", args[0].Type())
			},
		},

		"math.floor": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				if arg, ok := args[0].(*object.Float); ok {
					return &object.Float{Value: math.Floor(arg.Value)}
				}

				if arg, ok := args[0].(*object.Integer); ok {
					return arg // Already an integer
				}

				return helpers.NewError("argument to `math.floor` must be FLOAT or INTEGER, got %s", args[0].Type())
			},
		},

		"math.round": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("wrong number of arguments. got=%d, want=2", len(args))
				}

				var value float64
				if arg, ok := args[0].(*object.Float); ok {
					value = arg.Value
				} else if arg, ok := args[0].(*object.Integer); ok {
					value = float64(arg.Value)
				} else {
					return helpers.NewError("first argument to `math.round` must be FLOAT or INTEGER, got %s", args[0].Type())
				}

				decimals, ok := args[1].(*object.Integer)
				if !ok {
					return helpers.NewError("second argument to `math.round` must be INTEGER, got %s", args[1].Type())
				}

				shift := math.Pow(10, float64(decimals.Value))
				return &object.Float{Value: math.Round(value*shift) / shift}
			},
		},

		"math.min": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) == 0 {
					return helpers.NewError("wrong number of arguments. got=0, want at least 1")
				}

				// Check if all integers
				allIntegers := true
				for _, arg := range args {
					if _, ok := arg.(*object.Integer); !ok {
						allIntegers = false
						break
					}
				}

				if allIntegers {
					min := args[0].(*object.Integer).Value
					for i := 1; i < len(args); i++ {
						val := args[i].(*object.Integer).Value
						if val < min {
							min = val
						}
					}
					return &object.Integer{Value: min}
				}

				// Try floats
				var min float64
				for i, arg := range args {
					var val float64
					if f, ok := arg.(*object.Float); ok {
						val = f.Value
					} else if intVal, ok := arg.(*object.Integer); ok {
						val = float64(intVal.Value)
					} else {
						return helpers.NewError("arguments to `math.min` must be INTEGER or FLOAT")
					}

					if i == 0 || val < min {
						min = val
					}
				}
				return &object.Float{Value: min}
			},
		},

		"math.max": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) == 0 {
					return helpers.NewError("wrong number of arguments. got=0, want at least 1")
				}

				// Check if all integers
				allIntegers := true
				for _, arg := range args {
					if _, ok := arg.(*object.Integer); !ok {
						allIntegers = false
						break
					}
				}

				if allIntegers {
					max := args[0].(*object.Integer).Value
					for i := 1; i < len(args); i++ {
						val := args[i].(*object.Integer).Value
						if val > max {
							max = val
						}
					}
					return &object.Integer{Value: max}
				}

				// Try floats
				var max float64
				for i, arg := range args {
					var val float64
					if f, ok := arg.(*object.Float); ok {
						val = f.Value
					} else if intVal, ok := arg.(*object.Integer); ok {
						val = float64(intVal.Value)
					} else {
						return helpers.NewError("arguments to `math.max` must be INTEGER or FLOAT")
					}

					if i == 0 || val > max {
						max = val
					}
				}
				return &object.Float{Value: max}
			},
		},

		"math.sqrt": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				var val float64
				if arg, ok := args[0].(*object.Float); ok {
					val = arg.Value
				} else if arg, ok := args[0].(*object.Integer); ok {
					val = float64(arg.Value)
				} else {
					return helpers.NewError("argument to `math.sqrt` must be INTEGER or FLOAT, got %s", args[0].Type())
				}

				if val < 0 {
					return helpers.NewError("math.sqrt: cannot take square root of negative number")
				}

				return &object.Float{Value: math.Sqrt(val)}
			},
		},

		"math.pow": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("wrong number of arguments. got=%d, want=2", len(args))
				}

				var base, exponent float64
				if arg, ok := args[0].(*object.Float); ok {
					base = arg.Value
				} else if arg, ok := args[0].(*object.Integer); ok {
					base = float64(arg.Value)
				} else {
					return helpers.NewError("first argument to `math.pow` must be INTEGER or FLOAT, got %s", args[0].Type())
				}

				if arg, ok := args[1].(*object.Float); ok {
					exponent = arg.Value
				} else if arg, ok := args[1].(*object.Integer); ok {
					exponent = float64(arg.Value)
				} else {
					return helpers.NewError("second argument to `math.pow` must be INTEGER or FLOAT, got %s", args[1].Type())
				}

				return &object.Float{Value: math.Pow(base, exponent)}
			},
		},

		"math.exp": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				var val float64
				if arg, ok := args[0].(*object.Float); ok {
					val = arg.Value
				} else if arg, ok := args[0].(*object.Integer); ok {
					val = float64(arg.Value)
				} else {
					return helpers.NewError("argument to `math.exp` must be INTEGER or FLOAT, got %s", args[0].Type())
				}

				return &object.Float{Value: math.Exp(val)}
			},
		},

		"math.log": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				var val float64
				if arg, ok := args[0].(*object.Float); ok {
					val = arg.Value
				} else if arg, ok := args[0].(*object.Integer); ok {
					val = float64(arg.Value)
				} else {
					return helpers.NewError("argument to `math.log` must be INTEGER or FLOAT, got %s", args[0].Type())
				}

				if val <= 0 {
					return helpers.NewError("math.log: argument must be positive")
				}

				return &object.Float{Value: math.Log(val)}
			},
		},

		"math.log10": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				var val float64
				if arg, ok := args[0].(*object.Float); ok {
					val = arg.Value
				} else if arg, ok := args[0].(*object.Integer); ok {
					val = float64(arg.Value)
				} else {
					return helpers.NewError("argument to `math.log10` must be INTEGER or FLOAT, got %s", args[0].Type())
				}

				if val <= 0 {
					return helpers.NewError("math.log10: argument must be positive")
				}

				return &object.Float{Value: math.Log10(val)}
			},
		},

		"math.sin": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				var val float64
				if arg, ok := args[0].(*object.Float); ok {
					val = arg.Value
				} else if arg, ok := args[0].(*object.Integer); ok {
					val = float64(arg.Value)
				} else {
					return helpers.NewError("argument to `math.sin` must be INTEGER or FLOAT, got %s", args[0].Type())
				}

				return &object.Float{Value: math.Sin(val)}
			},
		},

		"math.cos": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				var val float64
				if arg, ok := args[0].(*object.Float); ok {
					val = arg.Value
				} else if arg, ok := args[0].(*object.Integer); ok {
					val = float64(arg.Value)
				} else {
					return helpers.NewError("argument to `math.cos` must be INTEGER or FLOAT, got %s", args[0].Type())
				}

				return &object.Float{Value: math.Cos(val)}
			},
		},

		"math.tan": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				var val float64
				if arg, ok := args[0].(*object.Float); ok {
					val = arg.Value
				} else if arg, ok := args[0].(*object.Integer); ok {
					val = float64(arg.Value)
				} else {
					return helpers.NewError("argument to `math.tan` must be INTEGER or FLOAT, got %s", args[0].Type())
				}

				return &object.Float{Value: math.Tan(val)}
			},
		},

		"math.asin": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				var val float64
				if arg, ok := args[0].(*object.Float); ok {
					val = arg.Value
				} else if arg, ok := args[0].(*object.Integer); ok {
					val = float64(arg.Value)
				} else {
					return helpers.NewError("argument to `math.asin` must be INTEGER or FLOAT, got %s", args[0].Type())
				}

				if val < -1 || val > 1 {
					return helpers.NewError("math.asin: argument must be in range [-1, 1]")
				}

				return &object.Float{Value: math.Asin(val)}
			},
		},

		"math.acos": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				var val float64
				if arg, ok := args[0].(*object.Float); ok {
					val = arg.Value
				} else if arg, ok := args[0].(*object.Integer); ok {
					val = float64(arg.Value)
				} else {
					return helpers.NewError("argument to `math.acos` must be INTEGER or FLOAT, got %s", args[0].Type())
				}

				if val < -1 || val > 1 {
					return helpers.NewError("math.acos: argument must be in range [-1, 1]")
				}

				return &object.Float{Value: math.Acos(val)}
			},
		},

		"math.atan": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				var val float64
				if arg, ok := args[0].(*object.Float); ok {
					val = arg.Value
				} else if arg, ok := args[0].(*object.Integer); ok {
					val = float64(arg.Value)
				} else {
					return helpers.NewError("argument to `math.atan` must be INTEGER or FLOAT, got %s", args[0].Type())
				}

				return &object.Float{Value: math.Atan(val)}
			},
		},

		"math.pi": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 0 {
					return helpers.NewError("wrong number of arguments. got=%d, want=0", len(args))
				}
				return &object.Float{Value: math.Pi}
			},
		},

		"math.e": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 0 {
					return helpers.NewError("wrong number of arguments. got=%d, want=0", len(args))
				}
				return &object.Float{Value: math.E}
			},
		},

		"math.odd?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				if arg, ok := args[0].(*object.Integer); ok {
					return helpers.NativeBoolToBooleanObject(arg.Value%2 != 0)
				}

				return helpers.NewError("argument to `math.odd?` must be INTEGER, got %s", args[0].Type())
			},
		},

		"math.even?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				if arg, ok := args[0].(*object.Integer); ok {
					return helpers.NativeBoolToBooleanObject(arg.Value%2 == 0)
				}

				return helpers.NewError("argument to `math.even?` must be INTEGER, got %s", args[0].Type())
			},
		},

		"math.to_int": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				if arg, ok := args[0].(*object.Float); ok {
					return &object.Integer{Value: int64(arg.Value)}
				}

				if arg, ok := args[0].(*object.Integer); ok {
					return arg // Already an integer
				}

				return helpers.NewError("argument to `math.to_int` must be FLOAT or INTEGER, got %s", args[0].Type())
			},
		},

		"math.to_float": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("wrong number of arguments. got=%d, want=1", len(args))
				}

				if arg, ok := args[0].(*object.Integer); ok {
					return &object.Float{Value: float64(arg.Value)}
				}

				if arg, ok := args[0].(*object.Float); ok {
					return arg // Already a float
				}

				return helpers.NewError("argument to `math.to_float` must be INTEGER or FLOAT, got %s", args[0].Type())
			},
		},
	}
}
