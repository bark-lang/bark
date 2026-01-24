package modules

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gitlab.com/bark-lang/bark/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/bark/object"

	// SQLite driver (pure Go, no CGO)
	_ "modernc.org/sqlite"
)

// InitSQL initializes SQL database operations
func InitSQL() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		"sql.open": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("sql.open requires 2 arguments (driver, dsn), got=%d", len(args))
				}

				driver, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("sql.open requires string driver, got=%s", args[0].Type())
				}

				dsn, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("sql.open requires string dsn, got=%s", args[1].Type())
				}

				// Map driver names to database/sql driver names
				var sqlDriver string
				switch driver.Value {
				case "sqlite":
					sqlDriver = "sqlite"
				case "postgres":
					return errorTuple("sql.open: postgres driver not yet implemented")
				default:
					return errorTuple(fmt.Sprintf("sql.open: unknown driver %q (supported: sqlite, postgres)", driver.Value))
				}

				db, err := sql.Open(sqlDriver, dsn.Value)
				if err != nil {
					return errorTuple(fmt.Sprintf("sql.open: %s", err.Error()))
				}

				// Test connection
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := db.PingContext(ctx); err != nil {
					_ = db.Close()
					return errorTuple(fmt.Sprintf("sql.open: connection failed: %s", err.Error()))
				}

				conn := &object.SQLConnection{
					DB:     db,
					Driver: driver.Value,
					DSN:    dsn.Value,
				}

				return successTuple(conn)
			},
		},

		"sql.close": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("sql.close requires 1 argument (connection), got=%d", len(args))
				}

				conn, ok := args[0].(*object.SQLConnection)
				if !ok {
					return helpers.NewError("sql.close requires sql connection, got=%s", args[0].Type())
				}

				if err := conn.Close(); err != nil {
					return errorTuple(fmt.Sprintf("sql.close: %s", err.Error()))
				}

				return successTuple(helpers.TRUE)
			},
		},

		"sql.query": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 2 || len(args) > 3 {
					return helpers.NewError("sql.query requires 2-3 arguments (conn/tx, query, [params]), got=%d", len(args))
				}

				// Get querier (connection or transaction)
				querier, driver, err := getQuerier(args[0])
				if err != nil {
					return helpers.NewError("sql.query: %s", err.Error())
				}

				query, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("sql.query requires string query, got=%s", args[1].Type())
				}

				// Get optional params
				var params []any
				if len(args) == 3 {
					arr, ok := args[2].(*object.Array)
					if !ok {
						return helpers.NewError("sql.query requires array params, got=%s", args[2].Type())
					}
					params = objectArrayToAny(arr)
				}

				// Execute query
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				rows, queryErr := querier.QueryContext(ctx, query.Value, params...)
				if queryErr != nil {
					return errorTuple(fmt.Sprintf("sql.query: %s", queryErr.Error()))
				}
				defer func() { _ = rows.Close() }()

				// Get column names
				columns, colErr := rows.Columns()
				if colErr != nil {
					return errorTuple(fmt.Sprintf("sql.query: %s", colErr.Error()))
				}

				// Read all rows
				var results []object.Object
				for rows.Next() {
					// Create scan destinations
					values := make([]any, len(columns))
					valuePtrs := make([]any, len(columns))
					for i := range values {
						valuePtrs[i] = &values[i]
					}

					if scanErr := rows.Scan(valuePtrs...); scanErr != nil {
						return errorTuple(fmt.Sprintf("sql.query: scan error: %s", scanErr.Error()))
					}

					// Convert to Bark map
					rowMap := &object.Map{
						Pairs: make(map[string]object.Object),
						Keys:  make([]string, len(columns)),
					}
					for i, col := range columns {
						rowMap.Keys[i] = col
						rowMap.Pairs[col] = anyToObject(values[i], driver)
					}
					results = append(results, rowMap)
				}

				if rowsErr := rows.Err(); rowsErr != nil {
					return errorTuple(fmt.Sprintf("sql.query: %s", rowsErr.Error()))
				}

				return successTuple(&object.Array{Elements: results})
			},
		},

		"sql.exec": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 2 || len(args) > 3 {
					return helpers.NewError("sql.exec requires 2-3 arguments (conn/tx, query, [params]), got=%d", len(args))
				}

				// Get execer (connection or transaction)
				execer, _, err := getExecer(args[0])
				if err != nil {
					return helpers.NewError("sql.exec: %s", err.Error())
				}

				query, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("sql.exec requires string query, got=%s", args[1].Type())
				}

				// Get optional params
				var params []any
				if len(args) == 3 {
					arr, ok := args[2].(*object.Array)
					if !ok {
						return helpers.NewError("sql.exec requires array params, got=%s", args[2].Type())
					}
					params = objectArrayToAny(arr)
				}

				// Execute query
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				result, execErr := execer.ExecContext(ctx, query.Value, params...)
				if execErr != nil {
					return errorTuple(fmt.Sprintf("sql.exec: %s", execErr.Error()))
				}

				affected, _ := result.RowsAffected()
				return successTuple(&object.Integer{Value: affected})
			},
		},

		"sql.begin": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("sql.begin requires 1 argument (connection), got=%d", len(args))
				}

				conn, ok := args[0].(*object.SQLConnection)
				if !ok {
					return helpers.NewError("sql.begin requires sql connection, got=%s", args[0].Type())
				}

				// Use background context - transaction lifetime is managed by commit/rollback
				tx, err := conn.DB.BeginTx(context.Background(), nil)
				if err != nil {
					return errorTuple(fmt.Sprintf("sql.begin: %s", err.Error()))
				}

				return successTuple(&object.SQLTransaction{
					Tx:     tx,
					Driver: conn.Driver,
				})
			},
		},

		"sql.commit": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("sql.commit requires 1 argument (transaction), got=%d", len(args))
				}

				tx, ok := args[0].(*object.SQLTransaction)
				if !ok {
					return helpers.NewError("sql.commit requires sql transaction, got=%s", args[0].Type())
				}

				if err := tx.Tx.Commit(); err != nil {
					return errorTuple(fmt.Sprintf("sql.commit: %s", err.Error()))
				}

				return successTuple(helpers.TRUE)
			},
		},

		"sql.rollback": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("sql.rollback requires 1 argument (transaction), got=%d", len(args))
				}

				tx, ok := args[0].(*object.SQLTransaction)
				if !ok {
					return helpers.NewError("sql.rollback requires sql transaction, got=%s", args[0].Type())
				}

				if err := tx.Tx.Rollback(); err != nil {
					return errorTuple(fmt.Sprintf("sql.rollback: %s", err.Error()))
				}

				return successTuple(helpers.TRUE)
			},
		},
	}
}

// Querier interface for query operations
type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// Execer interface for exec operations
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// getQuerier returns a Querier from a connection or transaction
func getQuerier(obj object.Object) (Querier, string, error) {
	switch v := obj.(type) {
	case *object.SQLConnection:
		return v.DB, v.Driver, nil
	case *object.SQLTransaction:
		return v.Tx, v.Driver, nil
	default:
		return nil, "", fmt.Errorf("requires connection or transaction, got=%s", obj.Type())
	}
}

// getExecer returns an Execer from a connection or transaction
func getExecer(obj object.Object) (Execer, string, error) {
	switch v := obj.(type) {
	case *object.SQLConnection:
		return v.DB, v.Driver, nil
	case *object.SQLTransaction:
		return v.Tx, v.Driver, nil
	default:
		return nil, "", fmt.Errorf("requires connection or transaction, got=%s", obj.Type())
	}
}

// errorTuple creates an error tuple (error, null)
func errorTuple(msg string) *object.Tuple {
	return &object.Tuple{
		Elements: []object.Object{
			&object.Error{Msg: msg, Context: make(map[string]object.Object)},
			helpers.NULL,
		},
	}
}

// successTuple creates a success tuple ({}, result)
func successTuple(result object.Object) *object.Tuple {
	return &object.Tuple{
		Elements: []object.Object{
			&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}},
			result,
		},
	}
}

// objectArrayToAny converts a Bark array to []any for SQL params
func objectArrayToAny(arr *object.Array) []any {
	result := make([]any, len(arr.Elements))
	for i, elem := range arr.Elements {
		switch v := elem.(type) {
		case *object.Integer:
			result[i] = v.Value
		case *object.Float:
			result[i] = v.Value
		case *object.String:
			result[i] = v.Value
		case *object.Boolean:
			result[i] = v.Value
		case *object.Null:
			result[i] = nil
		default:
			result[i] = v.Inspect()
		}
	}
	return result
}

// anyToObject converts a SQL value to a Bark object
func anyToObject(val any, _ string) object.Object {
	if val == nil {
		return helpers.NULL
	}

	switch v := val.(type) {
	case int64:
		return &object.Integer{Value: v}
	case float64:
		return &object.Float{Value: v}
	case string:
		return &object.String{Value: v}
	case bool:
		return helpers.NativeBoolToBooleanObject(v)
	case []byte:
		return &object.String{Value: string(v)}
	case time.Time:
		return &object.String{Value: v.Format(time.RFC3339)}
	default:
		return &object.String{Value: fmt.Sprintf("%v", v)}
	}
}
