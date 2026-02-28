package modules

import (
	"strings"
	"time"

	"gitlab.com/bark-lang/barki/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/barki/object"
)

// InitTime initializes time operations
func InitTime() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"time.now": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 0 {
					return helpers.NewError("time.now requires 0 arguments, got=%d", len(args))
				}

				// Return current Unix timestamp in seconds
				return &object.Integer{Value: time.Now().UTC().Unix()}
			},
		},

		"time.now_ms": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 0 {
					return helpers.NewError("time.now_ms requires 0 arguments, got=%d", len(args))
				}

				// Return current Unix timestamp in milliseconds
				return &object.Integer{Value: time.Now().UTC().UnixMilli()}
			},
		},

		"time.format": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("time.format requires 2 arguments (timestamp, format), got=%d", len(args))
				}

				timestamp, ok := args[0].(*object.Integer)
				if !ok {
					return helpers.NewError("time.format requires integer timestamp, got=%s", args[0].Type())
				}

				format, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("time.format requires string format, got=%s", args[1].Type())
				}

				// Convert Unix timestamp to time.Time
				t := time.Unix(timestamp.Value, 0).UTC()

				// Convert strftime format to Go format
				goFormat := strftimeToGo(format.Value)

				// Format the time
				result := t.Format(goFormat)

				return &object.String{Value: result}
			},
		},

		"time.parse": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("time.parse requires 2 arguments (time_string, format), got=%d", len(args))
				}

				timeStr, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("time.parse requires string time_string, got=%s", args[0].Type())
				}

				format, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("time.parse requires string format, got=%s", args[1].Type())
				}

				// Convert strftime format to Go format
				goFormat := strftimeToGo(format.Value)

				// Parse the time
				t, err := time.Parse(goFormat, timeStr.Value)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Integer{Value: 0},
						},
					}
				}

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						&object.Integer{Value: t.Unix()},
					},
				}
			},
		},

		"time.format_iso8601": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("time.format_iso8601 requires 1 argument (timestamp), got=%d", len(args))
				}

				timestamp, ok := args[0].(*object.Integer)
				if !ok {
					return helpers.NewError("time.format_iso8601 requires integer timestamp, got=%s", args[0].Type())
				}

				// Convert Unix timestamp to time.Time and format as RFC3339 (ISO 8601)
				t := time.Unix(timestamp.Value, 0).UTC()
				result := t.Format(time.RFC3339)

				return &object.String{Value: result}
			},
		},

		"time.parse_iso8601": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("time.parse_iso8601 requires 1 argument (time_string), got=%d", len(args))
				}

				timeStr, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("time.parse_iso8601 requires string argument, got=%s", args[0].Type())
				}

				// Parse RFC3339 (ISO 8601) format
				t, err := time.Parse(time.RFC3339, timeStr.Value)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							helpers.WrapError(err),
							&object.Integer{Value: 0},
						},
					}
				}

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						&object.Integer{Value: t.Unix()},
					},
				}
			},
		},
	}
}

// strftimeToGo converts strftime format strings to Go time format strings
func strftimeToGo(strftime string) string {
	// Map of strftime codes to Go format strings
	replacements := map[string]string{
		"%Y": "2006",       // 4-digit year
		"%y": "06",         // 2-digit year
		"%m": "01",         // Month (01-12)
		"%B": "January",    // Full month name
		"%b": "Jan",        // Abbreviated month name
		"%d": "02",         // Day of month (01-31)
		"%e": "2",          // Day of month (1-31, space padded)
		"%A": "Monday",     // Full weekday name
		"%a": "Mon",        // Abbreviated weekday
		"%H": "15",         // Hour 24h (00-23)
		"%I": "03",         // Hour 12h (01-12)
		"%M": "04",         // Minute (00-59)
		"%S": "05",         // Second (00-59)
		"%p": "PM",         // AM/PM
		"%Z": "MST",        // Timezone abbreviation
		"%z": "-0700",      // Timezone offset
		"%F": "2006-01-02", // ISO 8601 date (shortcut for %Y-%m-%d)
		"%T": "15:04:05",   // ISO 8601 time (shortcut for %H:%M:%S)
	}

	result := strftime
	for code, goFmt := range replacements {
		result = strings.ReplaceAll(result, code, goFmt)
	}

	return result
}
