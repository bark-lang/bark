package modules

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab.com/bark-lang/bark/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/bark/object"
)

// MaxResponseBodySize is the maximum allowed HTTP response body size (10MB)
const MaxResponseBodySize = 10 * 1024 * 1024

// ErrResponseTooLarge is returned when response body exceeds MaxResponseBodySize
var ErrResponseTooLarge = errors.New("response body exceeds 10MB limit")

// readResponseBody reads the response body with a size limit
func readResponseBody(body io.ReadCloser) ([]byte, error) {
	limitedReader := io.LimitReader(body, MaxResponseBodySize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}
	if len(data) > MaxResponseBodySize {
		return nil, ErrResponseTooLarge
	}
	return data, nil
}

// InitHTTP initializes HTTP client operations
func InitHTTP() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"http.get": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("http.get requires 1 argument (url), got=%d", len(args))
				}

				url, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("http.get requires string argument, got=%s", args[0].Type())
				}

				// Create HTTP client with 30-second timeout
				client := &http.Client{
					Timeout: 30 * time.Second,
				}

				// Make GET request
				resp, err := client.Get(url.Value)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}
				defer func() { _ = resp.Body.Close() }()

				// Read response body with size limit
				body, err := readResponseBody(resp.Body)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}

				// Build response map
				responseMap := buildHTTPResponse(resp, string(body))

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						responseMap,
					},
				}
			},
		},

		"http.post": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("http.post requires 2 arguments (url, body), got=%d", len(args))
				}

				url, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("http.post requires string url, got=%s", args[0].Type())
				}

				bodyStr, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("http.post requires string body, got=%s", args[1].Type())
				}

				// Create HTTP client with 30-second timeout
				client := &http.Client{
					Timeout: 30 * time.Second,
				}

				// Make POST request with text/plain content type
				resp, err := client.Post(url.Value, "text/plain", strings.NewReader(bodyStr.Value))
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}
				defer func() { _ = resp.Body.Close() }()

				// Read response body with size limit
				body, err := readResponseBody(resp.Body)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}

				// Build response map
				responseMap := buildHTTPResponse(resp, string(body))

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						responseMap,
					},
				}
			},
		},

		"http.put": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("http.put requires 2 arguments (url, body), got=%d", len(args))
				}

				url, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("http.put requires string url, got=%s", args[0].Type())
				}

				bodyStr, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("http.put requires string body, got=%s", args[1].Type())
				}

				// Create HTTP client with 30-second timeout
				client := &http.Client{
					Timeout: 30 * time.Second,
				}

				// Create PUT request
				req, err := http.NewRequest("PUT", url.Value, strings.NewReader(bodyStr.Value))
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}

				req.Header.Set("Content-Type", "text/plain")

				// Make request
				resp, err := client.Do(req)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}
				defer func() { _ = resp.Body.Close() }()

				// Read response body with size limit
				body, err := readResponseBody(resp.Body)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}

				// Build response map
				responseMap := buildHTTPResponse(resp, string(body))

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						responseMap,
					},
				}
			},
		},

		"http.delete": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("http.delete requires 1 argument (url), got=%d", len(args))
				}

				url, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("http.delete requires string argument, got=%s", args[0].Type())
				}

				// Create HTTP client with 30-second timeout
				client := &http.Client{
					Timeout: 30 * time.Second,
				}

				// Create DELETE request
				req, err := http.NewRequest("DELETE", url.Value, nil)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}

				// Make request
				resp, err := client.Do(req)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}
				defer func() { _ = resp.Body.Close() }()

				// Read response body with size limit
				body, err := readResponseBody(resp.Body)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}

				// Build response map
				responseMap := buildHTTPResponse(resp, string(body))

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						responseMap,
					},
				}
			},
		},

		"http.request": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 4 {
					return helpers.NewError("http.request requires 4 arguments (method, url, headers, body), got=%d", len(args))
				}

				method, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("http.request requires string method, got=%s", args[0].Type())
				}

				url, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("http.request requires string url, got=%s", args[1].Type())
				}

				headersMap, ok := args[2].(*object.Map)
				if !ok {
					return helpers.NewError("http.request requires map headers, got=%s", args[2].Type())
				}

				bodyStr, ok := args[3].(*object.String)
				if !ok {
					return helpers.NewError("http.request requires string body, got=%s", args[3].Type())
				}

				// Create HTTP client with 30-second timeout
				client := &http.Client{
					Timeout: 30 * time.Second,
				}

				// Create request with body
				var req *http.Request
				var err error
				if bodyStr.Value == "" {
					req, err = http.NewRequest(strings.ToUpper(method.Value), url.Value, nil)
				} else {
					req, err = http.NewRequest(strings.ToUpper(method.Value), url.Value, strings.NewReader(bodyStr.Value))
				}

				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}

				// Set custom headers
				for _, key := range headersMap.Keys {
					if value, ok := headersMap.Pairs[key]; ok {
						if valueStr, ok := value.(*object.String); ok {
							req.Header.Set(key, valueStr.Value)
						}
					}
				}

				// Make request
				resp, err := client.Do(req)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}
				defer func() { _ = resp.Body.Close() }()

				// Read response body with size limit
				body, err := readResponseBody(resp.Body)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						},
					}
				}

				// Build response map
				responseMap := buildHTTPResponse(resp, string(body))

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						responseMap,
					},
				}
			},
		},
	}
}

// buildHTTPResponse creates a bark map from an HTTP response
func buildHTTPResponse(resp *http.Response, body string) *object.Map {
	responseMap := &object.Map{
		Pairs: make(map[string]object.Object),
		Keys:  []string{"status", "headers", "body", "url"},
	}

	// Add status code
	responseMap.Pairs["status"] = &object.Integer{Value: int64(resp.StatusCode)}

	// Add headers as a map
	headersMap := &object.Map{
		Pairs: make(map[string]object.Object),
		Keys:  make([]string, 0, len(resp.Header)),
	}
	for headerName, headerValues := range resp.Header {
		if len(headerValues) > 0 {
			headersMap.Pairs[headerName] = &object.String{Value: headerValues[0]} // Take first value
			headersMap.Keys = append(headersMap.Keys, headerName)
		}
	}

	responseMap.Pairs["headers"] = headersMap

	// Add body
	responseMap.Pairs["body"] = &object.String{Value: body}

	// Add final URL (after redirects)
	responseMap.Pairs["url"] = &object.String{Value: resp.Request.URL.String()}

	return responseMap
}
