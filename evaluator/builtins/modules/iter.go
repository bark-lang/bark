package modules

import (
	"gitlab.com/bark-lang/barki/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/barki/object"
)

// InitIter initializes iterator operations
func InitIter() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"iter.next": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("iter.next requires 1 argument (iterator), got=%d", len(args))
				}

				iter, ok := args[0].(*object.Iterator)
				if !ok {
					return helpers.NewError("iter.next requires iterator argument, got=%s", args[0].Type())
				}

				if iter.IsExhausted() {
					// Return tuple: ({}, null, false)
					return &object.Tuple{
						Elements: []object.Object{
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
							&object.Null{},
							helpers.FALSE,
						},
					}
				}

				value, hasMore, err := iter.Next()
				if err != nil {
					// Return error tuple: (error, null, false)
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Null{},
							helpers.FALSE,
						},
					}
				}

				if !hasMore {
					iter.MarkExhausted()
					// Return tuple: ({}, null, false)
					return &object.Tuple{
						Elements: []object.Object{
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
							&object.Null{},
							helpers.FALSE,
						},
					}
				}

				// Return success tuple: ({}, value, true)
				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						value,
						helpers.TRUE,
					},
				}
			},
		},

		"iter.close": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("iter.close requires 1 argument (iterator), got=%d", len(args))
				}

				iter, ok := args[0].(*object.Iterator)
				if !ok {
					return helpers.NewError("iter.close requires iterator argument, got=%s", args[0].Type())
				}

				if iter.Close != nil {
					if err := iter.Close(); err != nil {
						return helpers.WrapError(err)
					}
				}
				iter.MarkExhausted()

				// Return empty map (no error)
				return &object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}
			},
		},

		"iter.exhausted?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("iter.exhausted? requires 1 argument (iterator), got=%d", len(args))
				}

				iter, ok := args[0].(*object.Iterator)
				if !ok {
					return helpers.NewError("iter.exhausted? requires iterator argument, got=%s", args[0].Type())
				}

				return helpers.NativeBoolToBooleanObject(iter.IsExhausted())
			},
		},
	}
}
