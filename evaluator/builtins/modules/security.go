package modules

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gitlab.com/bark-lang/bark/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/bark/object"
)

// InitSecurity initializes security and input validation operations
func InitSecurity() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		// SQL Injection Prevention
		"security.sql_escape": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("security.sql_escape requires 1 argument (input), got=%d", len(args))
				}

				input, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("security.sql_escape requires string argument, got=%s", args[0].Type())
				}

				// Escape single quotes for SQL
				escaped := strings.ReplaceAll(input.Value, "'", "''")
				// Remove null bytes
				escaped = strings.ReplaceAll(escaped, "\x00", "")

				return &object.String{Value: escaped}
			},
		},

		// Command Injection Prevention
		"security.shell_escape": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("security.shell_escape requires 1 argument (input), got=%d", len(args))
				}

				input, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("security.shell_escape requires string argument, got=%s", args[0].Type())
				}

				// Shell metacharacters that need escaping
				dangerous := []string{
					";", "&", "|", "`", "$", "(", ")", "<", ">",
					"*", "?", "[", "]", "{", "}", "~", "!",
					"\n", "\r", "\\", "\"", "'",
				}

				escaped := input.Value
				for _, char := range dangerous {
					escaped = strings.ReplaceAll(escaped, char, "\\"+char)
				}

				return &object.String{Value: escaped}
			},
		},

		// Check for dangerous shell commands
		"security.safe_command?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("security.safe_command? requires 1 argument (command), got=%d", len(args))
				}

				cmd, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("security.safe_command? requires string argument, got=%s", args[0].Type())
				}

				command := strings.ToLower(strings.TrimSpace(cmd.Value))

				// Dangerous commands that should never be allowed
				dangerousCommands := []string{
					"rm -rf", "rm -fr", "rm -r", "rm -f",
					"mkfs", "dd if=", "> /dev/", "chmod 777",
					"chmod -r 777", "chown -r", "curl", "wget",
					"nc -", "netcat", "telnet", "ssh",
					"eval", "exec", "system", "fork", "kill",
					"shutdown", "reboot", "halt", "poweroff",
					"iptables", "ufw", "firewall", "passwd",
					"useradd", "userdel", "groupadd", "crontab",
				}

				for _, dangerous := range dangerousCommands {
					if strings.Contains(command, dangerous) {
						return helpers.FALSE
					}
				}

				// Check for command chaining
				if strings.Contains(command, ";") ||
					strings.Contains(command, "&&") ||
					strings.Contains(command, "||") ||
					strings.Contains(command, "|") ||
					strings.Contains(command, "`") ||
					strings.Contains(command, "$(") {
					return helpers.FALSE
				}

				return helpers.TRUE
			},
		},

		// Path Traversal Prevention
		"security.sanitize_path": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 1 || len(args) > 2 {
					return helpers.NewError("security.sanitize_path requires 1-2 arguments (path, [base_dir]), got=%d", len(args))
				}

				pathStr, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("security.sanitize_path requires string argument, got=%s", args[0].Type())
				}

				basePath := "."
				if len(args) == 2 {
					baseStr, ok := args[1].(*object.String)
					if !ok {
						return helpers.NewError("security.sanitize_path base_dir must be string, got=%s", args[1].Type())
					}
					basePath = baseStr.Value
				}

				// Clean the path
				cleanPath := filepath.Clean(pathStr.Value)

				// Make it absolute relative to base
				absBase, err := filepath.Abs(basePath)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.NewError("failed to resolve base path: %s", err.Error()),
							&object.String{Value: ""},
						},
					}
				}

				absPath := filepath.Join(absBase, cleanPath)

				// Ensure the path doesn't escape the base directory
				relPath, err := filepath.Rel(absBase, absPath)
				if err != nil || strings.HasPrefix(relPath, "..") {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.NewError("path traversal detected: %s", pathStr.Value),
							&object.String{Value: ""},
						},
					}
				}

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}, // No error
						&object.String{Value: absPath},
					},
				}
			},
		},

		// Input Validation - Email
		"security.email?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("security.email? requires 1 argument (email), got=%d", len(args))
				}

				email, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("security.email? requires string argument, got=%s", args[0].Type())
				}

				// Basic email validation
				pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
				matched, _ := regexp.MatchString(pattern, email.Value)

				return helpers.NativeBoolToBooleanObject(matched)
			},
		},

		// Input Validation - URL
		"security.url?": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("security.url? requires 1 argument (url), got=%d", len(args))
				}

				urlStr, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("security.url? requires string argument, got=%s", args[0].Type())
				}

				// Check for valid URL scheme
				pattern := `^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`
				matched, _ := regexp.MatchString(pattern, urlStr.Value)

				return helpers.NativeBoolToBooleanObject(matched)
			},
		},

		// XSS Prevention
		"security.html_escape": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("security.html_escape requires 1 argument (html), got=%d", len(args))
				}

				html, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("security.html_escape requires string argument, got=%s", args[0].Type())
				}

				// Escape HTML special characters
				escaped := html.Value
				escaped = strings.ReplaceAll(escaped, "&", "&amp;")
				escaped = strings.ReplaceAll(escaped, "<", "&lt;")
				escaped = strings.ReplaceAll(escaped, ">", "&gt;")
				escaped = strings.ReplaceAll(escaped, "\"", "&quot;")
				escaped = strings.ReplaceAll(escaped, "'", "&#39;")
				escaped = strings.ReplaceAll(escaped, "/", "&#x2F;")

				return &object.String{Value: escaped}
			},
		},

		// Strip dangerous HTML tags
		"security.strip_tags": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("security.strip_tags requires 1 argument (html), got=%d", len(args))
				}

				html, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("security.strip_tags requires string argument, got=%s", args[0].Type())
				}

				// Remove all HTML tags
				re := regexp.MustCompile(`<[^>]*>`)
				stripped := re.ReplaceAllString(html.Value, "")

				return &object.String{Value: stripped}
			},
		},

		// Content Security Policy Helper
		"security.generate_nonce": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 0 {
					return helpers.NewError("security.generate_nonce requires 0 arguments, got=%d", len(args))
				}

				// Generate a random nonce for CSP
				// Using timestamp + random for uniqueness
				// rand.Seed deprecated in Go 1.20+, global rand is auto-seeded
				nonce := fmt.Sprintf("nonce-%d-%d", time.Now().UnixNano(), rand.Int63())

				return &object.String{Value: nonce}
			},
		},

		// Rate Limiting Helper
		"security.hash_key": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("security.hash_key requires 1 argument (input), got=%d", len(args))
				}

				input, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("security.hash_key requires string argument, got=%s", args[0].Type())
				}

				// Simple hash for rate limiting keys
				hash := int64(0)
				for _, c := range input.Value {
					hash = hash*31 + int64(c)
				}

				return &object.Integer{Value: hash}
			},
		},
	}
}
