package modules

import (
	"net/url"

	"gitlab.com/bark-lang/barki/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/barki/object"
)

// InitURL initializes URL encoding operations
func InitURL() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"url.encode": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("url.encode requires 1 argument (value), got=%d", len(args))
				}

				value, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("url.encode requires string argument, got=%s", args[0].Type())
				}

				// URL query encode (uses + for spaces)
				encoded := url.QueryEscape(value.Value)
				return &object.String{Value: encoded}
			},
		},

		"url.decode": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("url.decode requires 1 argument (encoded), got=%d", len(args))
				}

				encoded, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("url.decode requires string argument, got=%s", args[0].Type())
				}

				// URL query decode
				decoded, err := url.QueryUnescape(encoded.Value)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.String{Value: ""},
						},
					}
				}

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						&object.String{Value: decoded},
					},
				}
			},
		},

		"url.parse": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("url.parse requires 1 argument (url), got=%d", len(args))
				}

				urlStr, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("url.parse requires string argument, got=%s", args[0].Type())
				}

				// Parse URL
				parsedURL, err := url.Parse(urlStr.Value)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}

				// Build URL components map
				urlMap := &object.Map{
					Pairs: make(map[string]object.Object),
					Keys:  []string{"scheme", "host", "path", "query", "fragment"},
				}

				urlMap.Pairs["scheme"] = &object.String{Value: parsedURL.Scheme}
				urlMap.Pairs["host"] = &object.String{Value: parsedURL.Host}
				urlMap.Pairs["path"] = &object.String{Value: parsedURL.Path}
				urlMap.Pairs["query"] = &object.String{Value: parsedURL.RawQuery}
				urlMap.Pairs["fragment"] = &object.String{Value: parsedURL.Fragment}

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						urlMap,
					},
				}
			},
		},
	}
}
