package modules

import (
	"encoding/csv"
	"strings"

	"gitlab.com/bark-lang/barki/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/barki/object"
)

// InitCSV initializes CSV operations
func InitCSV() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"csv.decode": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("csv.decode requires 1 argument (csv_string), got=%d", len(args))
				}

				csvStr, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("csv.decode requires string argument, got=%s", args[0].Type())
				}

				reader := csv.NewReader(strings.NewReader(csvStr.Value))
				records, err := reader.ReadAll()
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     "csv.decode: " + err.Error(),
								Context: make(map[string]object.Object),
							},
							&object.Array{Elements: []object.Object{}},
						},
					}
				}

				if len(records) < 1 {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
							&object.Array{Elements: []object.Object{}},
						},
					}
				}

				// First row is headers
				headers := records[0]
				rows := make([]object.Object, 0, len(records)-1)

				for _, record := range records[1:] {
					pairs := make(map[string]object.Object)
					keys := make([]string, 0, len(headers))
					for i, header := range headers {
						val := ""
						if i < len(record) {
							val = record[i]
						}
						pairs[header] = &object.String{Value: val}
						keys = append(keys, header)
					}
					rows = append(rows, &object.Map{Pairs: pairs, Keys: keys})
				}

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
						&object.Array{Elements: rows},
					},
				}
			},
		},

		"csv.encode": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("csv.encode requires 1 argument (array_of_maps), got=%d", len(args))
				}

				arr, ok := args[0].(*object.Array)
				if !ok {
					return helpers.NewError("csv.encode requires array argument, got=%s", args[0].Type())
				}

				if len(arr.Elements) == 0 {
					return &object.String{Value: ""}
				}

				// Get headers from first row's keys
				firstRow, ok := arr.Elements[0].(*object.Map)
				if !ok {
					return helpers.NewError("csv.encode requires array of maps, first element is %s", arr.Elements[0].Type())
				}

				headers := firstRow.Keys

				var buf strings.Builder
				writer := csv.NewWriter(&buf)

				// Write headers
				if err := writer.Write(headers); err != nil {
					return helpers.NewError("csv.encode: %s", err.Error())
				}

				// Write data rows
				for i, elem := range arr.Elements {
					row, ok := elem.(*object.Map)
					if !ok {
						return helpers.NewError("csv.encode requires array of maps, element %d is %s", i, elem.Type())
					}

					record := make([]string, len(headers))
					for j, header := range headers {
						if val, exists := row.Pairs[header]; exists {
							record[j] = objectToCSVString(val)
						}
					}

					if err := writer.Write(record); err != nil {
						return helpers.NewError("csv.encode: %s", err.Error())
					}
				}

				writer.Flush()
				if err := writer.Error(); err != nil {
					return helpers.NewError("csv.encode: %s", err.Error())
				}

				return &object.String{Value: buf.String()}
			},
		},
	}
}

func objectToCSVString(obj object.Object) string {
	switch v := obj.(type) {
	case *object.String:
		return v.Value
	case *object.Integer:
		return v.Inspect()
	case *object.Float:
		return v.Inspect()
	case *object.Boolean:
		return v.Inspect()
	default:
		return obj.Inspect()
	}
}
