package evaluator

import (
	"fmt"
	"os"

	"gitlab.com/bark-lang/barki/ast"
	"gitlab.com/bark-lang/barki/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/barki/object"
)

// Re-export helpers for use throughout the evaluator package
var (
	NULL  = helpers.NULL
	TRUE  = helpers.TRUE
	FALSE = helpers.FALSE
	// Global module registry for tracking loaded modules
	moduleRegistry = NewModuleRegistry()
)

// SourceContext holds source code information for error reporting
type SourceContext struct {
	File  string   // File path
	Lines []string // Source lines (for displaying in errors)
}

// Global source context (set when evaluating a file)
var sourceContext *SourceContext

// SetSourceContext sets the source context for error reporting
func SetSourceContext(file string, source string) {
	lines := splitLines(source)
	sourceContext = &SourceContext{
		File:  file,
		Lines: lines,
	}
}

// splitLines splits source code into lines
func splitLines(source string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(source); i++ {
		if source[i] == '\n' {
			lines = append(lines, source[start:i])
			start = i + 1
		}
	}
	if start < len(source) {
		lines = append(lines, source[start:])
	}
	return lines
}

// handleExecutionError logs an execution error and enriches it with source context
func handleExecutionError(execErr *object.ExecutionError) {
	if sourceContext != nil {
		execErr.File = sourceContext.File
		if execErr.Line > 0 && execErr.Line <= len(sourceContext.Lines) {
			execErr.SourceLine = sourceContext.Lines[execErr.Line-1]
		}
	}
	// Log to stderr
	_, _ = fmt.Fprint(os.Stderr, execErr.FormatError())
}

// enrichError adds source location info to an error object
func enrichError(err *object.Error, line, column int) {
	err.Line = line
	err.Column = column
	if sourceContext != nil {
		err.File = sourceContext.File
		if line > 0 && line <= len(sourceContext.Lines) {
			err.SourceLine = sourceContext.Lines[line-1]
		}
	}
}

// newError is a convenience wrapper for helpers.NewError
func newError(format string, a ...interface{}) *object.Error {
	return helpers.NewError(format, a...)
}

// SetCurrentFile sets the current file path in the module registry
// This should be called before evaluating a program to enable relative imports
func SetCurrentFile(path string) {
	moduleRegistry.SetCurrentFile(path)
}

// GetModuleRegistry returns the global module registry
func GetModuleRegistry() *ModuleRegistry {
	return moduleRegistry
}

// Eval evaluates an AST node and returns an object
func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	// Statements
	case *ast.Program:
		return evalProgram(node, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.FunctionStatement:
		fn := &object.Function{
			Parameters: node.Parameters,
			Body:       node.Body,
			Env:        env,
			ReturnType: node.ReturnType,
		}
		env.Set(node.Name.Value, fn)
		return fn

	case *ast.MemoizedFunctionStatement:
		mfn := &object.MemoizedFunction{
			Parameters: node.Parameters,
			Body:       node.Body,
			Env:        env,
			Cache:      object.NewMemoCache(),
			ReturnType: node.ReturnType,
		}
		env.Set(node.Name.Value, mfn)
		return mfn

	case *ast.ModuleStatement:
		return evalModuleStatement(node, env)

	case *ast.ImportStatement:
		return evalImportStatement(node, env)

	case *ast.IncludeStatement:
		return evalIncludeStatement(node, env)

	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}

	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.StringLiteral:
		// Process string interpolation
		interpolated, err := helpers.InterpolateString(node.Value, env)
		if err != nil {
			return newError("string interpolation error: %s", err.Error())
		}
		return &object.String{Value: interpolated}

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	case *ast.MapLiteral:
		return evalMapLiteral(node, env)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.MemberExpression:
		return evalMemberExpression(node, env)

	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		result := applyFunction(function, args)
		// Enrich errors with source location
		switch r := result.(type) {
		case *object.ExecutionError:
			r.Line = node.Token.Line
			r.Column = node.Token.Column
		case *object.Error:
			if r.IsProgrammingError {
				enrichError(r, node.Token.Line, node.Token.Column)
			}
		}
		return result

	case *ast.LinkExpression:
		return evalLinkExpression(node, env)

	case *ast.AnonymousFunction:
		return &object.Function{
			Parameters: node.Parameters,
			Body:       node.Body,
			Env:        env,
			ReturnType: node.ReturnType,
		}

	case *ast.TupleExpression:
		elements := []object.Object{}
		for _, el := range node.Elements {
			evaluated := Eval(el, env)
			if isError(evaluated) {
				return evaluated
			}
			elements = append(elements, evaluated)
		}
		return &object.Tuple{Elements: elements}

	}

	return NULL
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch r := result.(type) {
		case *object.ReturnValue:
			return r.Value
		case *object.Error:
			// Only programming errors stop execution
			// bark error values (from err() builtin) can be stored/passed
			if r.IsProgrammingError {
				return r
			}
		case *object.ExecutionError:
			// Log the execution error and continue to next statement
			handleExecutionError(r)
			result = NULL
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_OBJ || rt == object.REPEAT_OBJ {
				return result
			}
			// Only programming errors stop block execution
			// bark error values can be stored/passed
			if rt == object.ERROR_OBJ {
				if errObj, ok := result.(*object.Error); ok && errObj.IsProgrammingError {
					return result
				}
			}
			// Handle execution errors - log and continue
			if rt == object.EXEC_ERROR_OBJ {
				if execErr, ok := result.(*object.ExecutionError); ok {
					handleExecutionError(execErr)
				}
				result = NULL
			}
			// Handle CaptureStop - chain stopped but execution continues
			// Variables have already been bound, so just continue to next statement
			if rt == object.CAPTURE_STOP_OBJ {
				result = NULL
			}
			// Handle ChainStop - chain stopped via continue?() returning false
			// Just continue to next statement
			if rt == object.CHAIN_STOP_OBJ {
				result = NULL
			}
		}
	}

	return result
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	// Pre-allocate with known capacity to avoid slice growth allocations
	result := make([]object.Object, 0, len(exps))

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func evalMapLiteral(node *ast.MapLiteral, env *object.Environment) object.Object {
	pairs := make(map[string]object.Object)
	keys := make([]string, 0, len(node.OrderedKeys))

	for _, keyNode := range node.OrderedKeys {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		// Keys must be strings
		keyStr, ok := key.(*object.String)
		if !ok {
			return newError("map key must be string, got %s", key.Type())
		}

		valueNode := node.Pairs[keyNode]
		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		pairs[keyStr.Value] = value
		keys = append(keys, keyStr.Value)
	}

	return &object.Map{Pairs: pairs, Keys: keys}
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	// Check environment first - local variables shadow builtins
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	// Fall back to builtin functions
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: %s", node.Value)
}

func evalMemberExpression(node *ast.MemberExpression, env *object.Environment) object.Object {
	// For module.function syntax (e.g., time.now, file.read)
	// Construct the full name and look it up in builtins or user-defined modules
	objectIdent, ok := node.Object.(*ast.Identifier)
	if !ok {
		// If object is not an identifier, evaluate it normally (for future object property access)
		obj := Eval(node.Object, env)
		if isError(obj) {
			return obj
		}
		return newError("member access on non-identifier not yet supported")
	}

	// node.Member is already *ast.Identifier, no need for type assertion
	if node.Member == nil {
		return newError("member is nil")
	}

	moduleName := objectIdent.Value
	functionName := node.Member.Value
	fullName := moduleName + "." + functionName

	// Check if it's a builtin module function
	if builtin, ok := builtins[fullName]; ok {
		return builtin
	}

	// Check for user-defined modules
	// First, resolve the module name (it might be an alias)
	modulePath := moduleRegistry.ResolveAlias(moduleName)

	// Try to get the module from the registry
	if mod, ok := moduleRegistry.GetModule(modulePath); ok {
		// Check if the function is public
		if !mod.IsPublicFunction(functionName) {
			return newError("function '%s' is not exported from module '%s'", functionName, moduleName)
		}

		// Get the function from the module's environment
		if fn, ok := mod.GetFunction(functionName); ok {
			return fn
		}
		return newError("function '%s' not found in module '%s'", functionName, moduleName)
	}

	// Module not found
	return newError("unknown module or module function: %s", fullName)
}

func evalModuleStatement(node *ast.ModuleStatement, env *object.Environment) object.Object {
	// Module declarations are metadata - they don't produce runtime values
	// The module name is tracked in the module registry when the file is loaded
	moduleRegistry.SetCurrentModule(node.Name.Value)
	return NULL
}

func evalImportStatement(node *ast.ImportStatement, env *object.Environment) object.Object {
	// Import statements load external modules and make them available

	// Get the import path from the string literal
	importPath := node.Path.Value

	// 1. Resolve the import path to an absolute file path
	currentFile := moduleRegistry.GetCurrentFile()
	absPath, err := ResolvePathWithRegistry(importPath, currentFile, moduleRegistry)
	if err != nil {
		return newError("failed to resolve import path '%s': %s", importPath, err.Error())
	}

	// 2. Check for circular dependencies before loading
	if moduleRegistry.IsLoading(absPath) {
		return newError("circular dependency detected: %s -> %s",
			moduleRegistry.GetLoadingPath(), absPath)
	}

	// 3. Load and parse the module file (if not already loaded)
	mod, err := LoadModule(absPath, moduleRegistry)
	if err != nil {
		return newError("failed to load module '%s': %s", importPath, err.Error())
	}

	// 4. Evaluate the module in its own environment (if not already evaluated)
	if err := EvaluateModule(mod, moduleRegistry); err != nil {
		return newError("failed to evaluate module '%s': %s", importPath, err.Error())
	}

	// 5. Module is already registered in LoadModule()

	// 6. If there's an alias, register it
	if node.Alias != nil {
		// Register the alias to point to the module path (not module name)
		// This allows us to look up the module later
		moduleRegistry.SetAlias(node.Alias.Value, absPath)
	} else {
		// No alias - register the module name to point to the module path
		moduleRegistry.SetAlias(mod.Name, absPath)
	}

	return NULL
}

func evalIncludeStatement(node *ast.IncludeStatement, env *object.Environment) object.Object {
	// Include statements merge another module's code into the current module
	// Unlike import, include evaluates the code in the CURRENT environment

	// Get the include path from the string literal
	includePath := node.Path.Value

	// 1. Resolve the include path to an absolute file path
	currentFile := moduleRegistry.GetCurrentFile()
	absPath, err := ResolvePathWithRegistry(includePath, currentFile, moduleRegistry)
	if err != nil {
		return newError("failed to resolve include path '%s': %s", includePath, err.Error())
	}

	// 2. Load and parse the included file
	mod, err := LoadModule(absPath, moduleRegistry)
	if err != nil {
		return newError("failed to load include file '%s': %s", includePath, err.Error())
	}

	// 3. Verify it has the same module declaration (optional, but good practice)
	currentModuleName := moduleRegistry.GetCurrentModule()
	if mod.Name != "_" && currentModuleName != "" && mod.Name != currentModuleName {
		return newError("include file '%s' declares different module name '%s' (current module: '%s')",
			includePath, mod.Name, currentModuleName)
	}

	// 4. Evaluate the included code in the current environment (not the module's environment)
	// This is the key difference between include and import
	for _, stmt := range mod.AST.Statements {
		// Skip module statements - we already verified compatibility
		if _, ok := stmt.(*ast.ModuleStatement); ok {
			continue
		}

		result := Eval(stmt, env)
		if isError(result) {
			return newError("error evaluating included file '%s': %s", includePath, result.Inspect())
		}
	}

	// 5. Name collision checking is implicitly handled by the environment
	// If a function is redefined, it will overwrite the previous one

	return NULL
}

func evalLinkExpression(node *ast.LinkExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}
	// Execution errors stop the chain
	if isExecutionError(left) {
		return left
	}
	// CaptureStop signals that capture encountered an error - stop the chain
	if isCaptureStop(left) {
		return left
	}
	// ChainStop signals that continue?() returned false - stop the chain
	if isChainStop(left) {
		return left
	}

	// Use type switch for efficient dispatch on right-hand side
	switch right := node.Right.(type) {
	case *ast.Identifier:
		// Variable binding: value > varName
		// Prevent tuple assignment to variable
		if _, isTuple := left.(*object.Tuple); isTuple {
			return newError("cannot assign tuple to variable")
		}
		// Mark as shared for COW optimization
		markShared(left)
		env.Set(right.Value, left)
		return left

	case *ast.TupleDestructure:
		// For now, we'll handle this in a future iteration
		return newError("tuple destructuring not yet implemented")

	case *ast.CaptureExpression:
		// Capture expression: value > capture(errVar, resultVar)
		return evalCaptureExpression(left, right, env)

	case *ast.AnonymousFunction:
		// Anonymous function - call it with left as first argument
		function := &object.Function{
			Parameters: right.Parameters,
			Body:       right.Body,
			Env:        env,
			ReturnType: right.ReturnType,
		}

		// If left is Void, call with no arguments
		if _, isVoid := left.(*object.Void); isVoid {
			return applyFunction(function, []object.Object{})
		}

		// If left is a tuple, check if it should be unpacked or passed as-is
		if tuple, ok := left.(*object.Tuple); ok {
			// If function has 1 parameter with a tuple type, pass tuple as-is
			if len(function.Parameters) == 1 && function.Parameters[0].Type != nil &&
				function.Parameters[0].Type.TupleTypes != nil {
				return applyFunction(function, []object.Object{left})
			}
			// Otherwise, unpack tuple elements as arguments
			if len(tuple.Elements) != len(function.Parameters) {
				return newError("wrong number of arguments: expected %d, got %d",
					len(function.Parameters), len(tuple.Elements))
			}
			return applyFunction(function, tuple.Elements)
		}

		// Single value - call the function with left value as the argument
		return applyFunction(function, []object.Object{left})

	case *ast.CallExpression:
		// Function call - prepend left as first argument
		function := Eval(right.Function, env)
		if isError(function) {
			return function
		}

		// Evaluate existing arguments
		args := evalExpressions(right.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		// If left is Void, don't prepend it - just use explicit args
		// This allows continue?() to act as a pure gate without passing a value
		var allArgs []object.Object
		if _, isVoid := left.(*object.Void); isVoid {
			allArgs = args
		} else if tuple, ok := left.(*object.Tuple); ok {
			// If left is a tuple, check if it should be unpacked or passed as-is
			// Check if function's first parameter expects a tuple type
			shouldPassAsTuple := false
			if fn, ok := function.(*object.Function); ok {
				// If no explicit args and function has 1 parameter with tuple type
				if len(args) == 0 && len(fn.Parameters) == 1 &&
					fn.Parameters[0].Type != nil && fn.Parameters[0].Type.TupleTypes != nil {
					shouldPassAsTuple = true
				}
			}
			if shouldPassAsTuple {
				allArgs = append([]object.Object{left}, args...)
			} else {
				allArgs = append(tuple.Elements, args...)
			}
		} else {
			// Single value - prepend left value to arguments
			allArgs = append([]object.Object{left}, args...)
		}

		result := applyFunction(function, allArgs)
		// Enrich errors with source location from the call expression
		switch r := result.(type) {
		case *object.ExecutionError:
			r.Line = right.Token.Line
			r.Column = right.Token.Column
		case *object.Error:
			if r.IsProgrammingError {
				enrichError(r, right.Token.Line, right.Token.Column)
			}
		}
		return result

	case *ast.LinkExpression:
		// Nested link expression - evaluate it
		return Eval(right, env)

	default:
		// Otherwise just evaluate the right side
		return Eval(node.Right, env)
	}
}

// evalCaptureExpression handles: value > capture(errVar, resultVar)
// The left value must be a tuple (error, value).
// On success (error is absent): binds both vars, returns resultVar to continue chain
// On error (error is present): binds both vars, returns CaptureStop to stop chain
func evalCaptureExpression(left object.Object, capture *ast.CaptureExpression, env *object.Environment) object.Object {
	// Left must be a tuple with exactly 2 elements: (error, value)
	tuple, ok := left.(*object.Tuple)
	if !ok {
		return newError("capture requires (error, value) tuple, got %s", left.Type())
	}

	if len(tuple.Elements) != 2 {
		return newError("capture requires tuple with exactly 2 elements (error, value), got %d elements", len(tuple.Elements))
	}

	errorVal := tuple.Elements[0]
	resultVal := tuple.Elements[1]

	// Mark as shared for COW optimization
	markShared(errorVal)
	markShared(resultVal)

	// Bind both variables to the environment
	env.Set(capture.ErrorVar.Value, errorVal)
	env.Set(capture.ResultVar.Value, resultVal)

	// Check if error is present (non-empty error)
	// An error is "present" if it's an Error object with a non-empty message,
	// or a non-empty map (for error maps)
	if isErrorPresent(errorVal) {
		// Error occurred - stop the chain
		return &object.CaptureStop{Error: errorVal}
	}

	// No error - continue the chain with the result value
	return resultVal
}

// isErrorPresent checks if an error value indicates an error occurred
// Empty map {} or Error with empty message/context means "no error"
// This is consistent with the present?() builtin
func isErrorPresent(val object.Object) bool {
	switch v := val.(type) {
	case *object.Error:
		// Error object is present if it has a message or context
		return len(v.Msg) > 0 || len(v.Context) > 0
	case *object.Map:
		// Map is present if it has keys (non-empty)
		return len(v.Pairs) > 0
	case *object.Null:
		// Null is not an error
		return false
	default:
		// For other types, assume no error
		return false
	}
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {

	case *object.Function:
		extendedEnv, typeErr := extendFunctionEnv(fn, args)
		if typeErr != nil {
			return typeErr
		}
		// Store the current function in the environment so repeat?() can access it
		extendedEnv.Set("__current_function__", fn)
		extendedEnv.Set("__current_args__", &object.Array{Elements: args})
		evaluated := Eval(fn.Body, extendedEnv)

		// Handle repeat?() - check if result is a RepeatValue
		if repeatVal, ok := evaluated.(*object.RepeatValue); ok {
			// Determine which args to use
			var newArgs []object.Object
			if repeatVal.Args == nil {
				// Use current args
				newArgs = args
			} else {
				// Use provided args
				newArgs = repeatVal.Args
			}
			// Recursively call the function with new args
			return applyFunction(fn, newArgs)
		}

		result := unwrapReturnValue(evaluated)

		// Validate return type if declared
		if fn.ReturnType != nil {
			if retErr := validateReturnType(result, fn.ReturnType); retErr != nil {
				return retErr
			}
		}

		return result

	case *object.MemoizedFunction:
		// Check cache first using fast hash lookup
		if cached, ok := fn.Cache.Get(args); ok {
			return cached
		}

		// Cache miss - evaluate the function
		extendedEnv, typeErr := extendMemoizedFunctionEnv(fn, args)
		if typeErr != nil {
			return typeErr
		}
		extendedEnv.Set("__current_function__", fn)
		extendedEnv.Set("__current_args__", &object.Array{Elements: args})
		evaluated := Eval(fn.Body, extendedEnv)

		// Handle repeat?() - check if result is a RepeatValue
		if repeatVal, ok := evaluated.(*object.RepeatValue); ok {
			var newArgs []object.Object
			if repeatVal.Args == nil {
				newArgs = args
			} else {
				newArgs = repeatVal.Args
			}
			return applyFunction(fn, newArgs)
		}

		result := unwrapReturnValue(evaluated)

		// Validate return type if declared
		if fn.ReturnType != nil {
			if retErr := validateReturnType(result, fn.ReturnType); retErr != nil {
				return retErr
			}
		}

		// Only cache non-error results
		if !isError(result) && !isExecutionError(result) {
			fn.Cache.Set(args, result)
		}

		return result

	case *object.Builtin:
		return fn.Fn(args...)

	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) (*object.Environment, *object.Error) {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		if paramIdx < len(args) {
			arg := args[paramIdx]
			// Validate type if type annotation exists
			if param.Type != nil {
				if err := validateType(arg, param.Type, param.Name.Value); err != nil {
					return nil, err
				}
			}
			// Mark as shared for COW optimization
			markShared(arg)
			env.Set(param.Name.Value, arg)
		}
	}

	return env, nil
}

func extendMemoizedFunctionEnv(fn *object.MemoizedFunction, args []object.Object) (*object.Environment, *object.Error) {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		if paramIdx < len(args) {
			arg := args[paramIdx]
			// Validate type if type annotation exists
			if param.Type != nil {
				if err := validateType(arg, param.Type, param.Name.Value); err != nil {
					return nil, err
				}
			}
			// Mark as shared for COW optimization
			markShared(arg)
			env.Set(param.Name.Value, arg)
		}
	}

	return env, nil
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

// typeMap maps type names to object types for runtime validation
var typeMap = map[string]object.ObjectType{
	"int":      object.INTEGER_OBJ,
	"float":    object.FLOAT_OBJ,
	"string":   object.STRING_OBJ,
	"bool":     object.BOOLEAN_OBJ,
	"array":    object.ARRAY_OBJ,
	"map":      object.MAP_OBJ,
	"fn":       object.FUNCTION_OBJ,
	"null":     object.NULL_OBJ,
	"iterator": object.ITERATOR_OBJ,
	"lazy_map": object.LAZY_MAP_OBJ,
	"t":        "", // Generic type - accepts any type
	"error":    "", // Special type - accepts maps (bark error convention)
}

// validateType checks if an object matches the expected type annotation
func validateType(obj object.Object, typeExpr *ast.TypeExpression, paramName string) *object.Error {
	if typeExpr == nil {
		return nil
	}

	typeName := typeExpr.Name

	// Generic type 't' accepts any type
	if typeName == "t" {
		return nil
	}

	// Handle union type: int | string etc.
	if typeExpr.UnionTypes != nil {
		return validateUnionType(obj, typeExpr, paramName)
	}

	// Handle tuple type: (int, string) etc.
	if typeExpr.TupleTypes != nil {
		return validateTupleType(obj, typeExpr, paramName)
	}

	// Handle function type: fn(param_types)(return_types)
	if typeExpr.ParamTypes != nil {
		return validateFunctionType(obj, typeExpr, paramName)
	}

	// Handle parameterized array type: array[element_type]
	if typeExpr.ElementType != nil {
		return validateArrayType(obj, typeExpr, paramName)
	}

	// Handle parameterized map type: map[key_type, value_type]
	if typeExpr.KeyType != nil && typeExpr.ValueType != nil {
		return validateMapType(obj, typeExpr, paramName)
	}

	// Special handling for 'error' type
	// In bark, errors can be:
	// - object.Error (from err() builtin)
	// - Maps with msg field (bark error convention)
	// - Empty maps {} (representing no error)
	if typeName == "error" {
		if obj.Type() == object.MAP_OBJ || obj.Type() == object.ERROR_OBJ {
			return nil
		}
		return newError("type mismatch: parameter '%s' expects error, got %s",
			paramName, obj.Type())
	}

	expectedType, ok := typeMap[typeName]
	if !ok {
		return newError("unknown type: %s", typeName)
	}

	// Check if the actual type matches the expected type
	actualType := obj.Type()
	if actualType != expectedType {
		return newError("type mismatch: parameter '%s' expects %s, got %s",
			paramName, typeName, actualType)
	}

	return nil
}

// validateTupleType checks if an object matches a tuple type annotation
func validateTupleType(obj object.Object, typeExpr *ast.TypeExpression, paramName string) *object.Error {
	// Must be a tuple object
	tuple, ok := obj.(*object.Tuple)
	if !ok {
		return newError("type mismatch: parameter '%s' expects %s, got %s",
			paramName, typeExpr.String(), obj.Type())
	}

	// Check element count matches
	if len(tuple.Elements) != len(typeExpr.TupleTypes) {
		return newError("type mismatch: parameter '%s' expects tuple with %d elements, got %d",
			paramName, len(typeExpr.TupleTypes), len(tuple.Elements))
	}

	// Check each element type
	for i, elem := range tuple.Elements {
		elemTypeExpr := typeExpr.TupleTypes[i]
		err := validateType(elem, elemTypeExpr, fmt.Sprintf("%s[%d]", paramName, i))
		if err != nil {
			return err
		}
	}

	return nil
}

// validateArrayType checks if an object matches a parameterized array type
// Array type: array[element_type]
func validateArrayType(obj object.Object, typeExpr *ast.TypeExpression, paramName string) *object.Error {
	// Must be an array object
	arr, ok := obj.(*object.Array)
	if !ok {
		return newError("type mismatch: parameter '%s' expects %s, got %s",
			paramName, typeExpr.String(), obj.Type())
	}

	// Validate each element against the element type
	for i, elem := range arr.Elements {
		err := validateType(elem, typeExpr.ElementType, fmt.Sprintf("%s[%d]", paramName, i))
		if err != nil {
			return err
		}
	}

	return nil
}

// validateMapType checks if an object matches a parameterized map type
// Map type: map[key_type, value_type]
func validateMapType(obj object.Object, typeExpr *ast.TypeExpression, paramName string) *object.Error {
	// Must be a map object
	m, ok := obj.(*object.Map)
	if !ok {
		return newError("type mismatch: parameter '%s' expects %s, got %s",
			paramName, typeExpr.String(), obj.Type())
	}

	// Map keys in bark are always strings, so verify the key type is string
	if typeExpr.KeyType.Name != "string" && typeExpr.KeyType.Name != "t" {
		return newError("type mismatch: map keys must be string type, got %s", typeExpr.KeyType.Name)
	}

	// Validate each value against the value type
	for _, key := range m.Keys {
		value := m.Pairs[key]
		err := validateType(value, typeExpr.ValueType, fmt.Sprintf("%s[\"%s\"]", paramName, key))
		if err != nil {
			return err
		}
	}

	return nil
}

// validateUnionType checks if an object matches any type in a union
// Union type: int | string | bool etc.
func validateUnionType(obj object.Object, typeExpr *ast.TypeExpression, paramName string) *object.Error {
	// Try each type in the union - if any matches, validation succeeds
	for _, unionMember := range typeExpr.UnionTypes {
		if err := validateType(obj, unionMember, paramName); err == nil {
			return nil // Found a matching type
		}
	}

	// None matched - return error showing all alternatives
	return newError("type mismatch: parameter '%s' expects %s, got %s",
		paramName, typeExpr.String(), obj.Type())
}

// validateFunctionType checks if an object matches a function type annotation
// Function type: fn(param_types)(return_types)
func validateFunctionType(obj object.Object, typeExpr *ast.TypeExpression, paramName string) *object.Error {
	// Must be a function object
	fn, ok := obj.(*object.Function)
	if !ok {
		return newError("type mismatch: parameter '%s' expects %s, got %s",
			paramName, typeExpr.String(), obj.Type())
	}

	// Check parameter count matches
	if len(fn.Parameters) != len(typeExpr.ParamTypes) {
		return newError("type mismatch: parameter '%s' expects function with %d parameters, got %d",
			paramName, len(typeExpr.ParamTypes), len(fn.Parameters))
	}

	// Check each parameter type matches
	for i, param := range fn.Parameters {
		expectedType := typeExpr.ParamTypes[i]
		if param.Type == nil {
			return newError("type mismatch: parameter '%s' function parameter %d ('%s') has no type annotation",
				paramName, i, param.Name.Value)
		}
		if !typeExpressionsMatch(param.Type, expectedType) {
			return newError("type mismatch: parameter '%s' function parameter %d ('%s') expects %s, got %s",
				paramName, i, param.Name.Value, expectedType.String(), param.Type.String())
		}
	}

	// Check return type count matches
	expectedReturnCount := len(typeExpr.ReturnTypes)
	actualReturnCount := 0
	if fn.ReturnType != nil {
		actualReturnCount = len(fn.ReturnType.Types)
	}
	if actualReturnCount != expectedReturnCount {
		return newError("type mismatch: parameter '%s' expects function with %d return values, got %d",
			paramName, expectedReturnCount, actualReturnCount)
	}

	// Check each return type matches
	if fn.ReturnType != nil {
		for i, returnType := range fn.ReturnType.Types {
			expectedType := typeExpr.ReturnTypes[i]
			if !typeExpressionsMatch(returnType, expectedType) {
				return newError("type mismatch: parameter '%s' function return %d expects %s, got %s",
					paramName, i, expectedType.String(), returnType.String())
			}
		}
	}

	return nil
}

// typeExpressionsMatch checks if two type expressions are equivalent
func typeExpressionsMatch(actual, expected *ast.TypeExpression) bool {
	// Generic type 't' matches anything
	if expected.Name == "t" {
		return true
	}

	// Handle tuple types
	if expected.TupleTypes != nil {
		if actual.TupleTypes == nil || len(actual.TupleTypes) != len(expected.TupleTypes) {
			return false
		}
		for i, expectedElem := range expected.TupleTypes {
			if !typeExpressionsMatch(actual.TupleTypes[i], expectedElem) {
				return false
			}
		}
		return true
	}

	// Handle function types
	if expected.ParamTypes != nil {
		if actual.ParamTypes == nil || len(actual.ParamTypes) != len(expected.ParamTypes) {
			return false
		}
		for i, expectedParam := range expected.ParamTypes {
			if !typeExpressionsMatch(actual.ParamTypes[i], expectedParam) {
				return false
			}
		}
		if len(actual.ReturnTypes) != len(expected.ReturnTypes) {
			return false
		}
		for i, expectedReturn := range expected.ReturnTypes {
			if !typeExpressionsMatch(actual.ReturnTypes[i], expectedReturn) {
				return false
			}
		}
		return true
	}

	// Handle parameterized array types
	if expected.ElementType != nil {
		if actual.ElementType == nil {
			return false
		}
		return typeExpressionsMatch(actual.ElementType, expected.ElementType)
	}

	// Handle parameterized map types
	if expected.KeyType != nil && expected.ValueType != nil {
		if actual.KeyType == nil || actual.ValueType == nil {
			return false
		}
		return typeExpressionsMatch(actual.KeyType, expected.KeyType) &&
			typeExpressionsMatch(actual.ValueType, expected.ValueType)
	}

	// Handle union types
	if expected.UnionTypes != nil {
		// Actual type can match if it matches any member of the expected union
		for _, expectedMember := range expected.UnionTypes {
			if typeExpressionsMatch(actual, expectedMember) {
				return true
			}
		}
		return false
	}
	// If actual is a union and expected is not, check if any member matches
	if actual.UnionTypes != nil {
		for _, actualMember := range actual.UnionTypes {
			if typeExpressionsMatch(actualMember, expected) {
				return true
			}
		}
		return false
	}

	// Simple type comparison
	return actual.Name == expected.Name
}

// validateReturnType checks if a return value matches the declared return type
func validateReturnType(obj object.Object, returnType *ast.TypeList) *object.Error {
	if returnType == nil || len(returnType.Types) == 0 {
		return nil
	}

	// Allow NULL returns for backwards compatibility with return?() without value
	// This supports patterns like: return?(condition) which returns null
	if obj == NULL {
		return nil
	}

	// Skip validation if the result is already a programming error
	// Programming errors should propagate without type checking
	if errObj, ok := obj.(*object.Error); ok && errObj.IsProgrammingError {
		return nil
	}

	// For tuple returns, we need to determine if this is:
	// 1. A single tuple type return: (int, string) -> returning a tuple as one value
	// 2. Multiple value return: int, string -> returning two values as a tuple
	if tuple, ok := obj.(*object.Tuple); ok {
		// If return type has 1 element that is a tuple type, treat the tuple as a single value
		if len(returnType.Types) == 1 && returnType.Types[0].TupleTypes != nil {
			return validateReturnTypeElement(obj, returnType.Types[0], 0)
		}
		// Otherwise, unpack and validate each element
		if len(tuple.Elements) != len(returnType.Types) {
			return newError("return type mismatch: expected %d values, got %d",
				len(returnType.Types), len(tuple.Elements))
		}
		for i, elem := range tuple.Elements {
			typeExpr := returnType.Types[i]
			err := validateReturnTypeElement(elem, typeExpr, i+1)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Single return value
	if len(returnType.Types) != 1 {
		return newError("return type mismatch: expected %d values, got 1",
			len(returnType.Types))
	}

	typeExpr := returnType.Types[0]
	return validateReturnTypeElement(obj, typeExpr, 0)
}

// validateReturnTypeElement validates a single return value element against its type expression
// position is 0 for single returns, 1-indexed for tuple elements
func validateReturnTypeElement(obj object.Object, typeExpr *ast.TypeExpression, position int) *object.Error {
	// Generic type 't' accepts any type
	if typeExpr.Name == "t" {
		return nil
	}

	// Handle union type in return: e.g., returning int | string
	if typeExpr.UnionTypes != nil {
		err := validateUnionType(obj, typeExpr, "return")
		if err != nil {
			if position > 0 {
				return newError("return type mismatch: expected %s at position %d, got %s",
					typeExpr.String(), position, obj.Type())
			}
			return newError("return type mismatch: expected %s, got %s",
				typeExpr.String(), obj.Type())
		}
		return nil
	}

	// Handle tuple type in return: e.g., returning (int, string) as a single value
	if typeExpr.TupleTypes != nil {
		err := validateTupleType(obj, typeExpr, "return")
		if err != nil {
			if position > 0 {
				return newError("return type mismatch: expected %s at position %d, got %s",
					typeExpr.String(), position, obj.Type())
			}
			return newError("return type mismatch: expected %s, got %s",
				typeExpr.String(), obj.Type())
		}
		return nil
	}

	// Special handling for 'error' type
	if typeExpr.Name == "error" {
		if obj.Type() == object.MAP_OBJ || obj.Type() == object.ERROR_OBJ {
			return nil
		}
		if position > 0 {
			return newError("return type mismatch: expected error at position %d, got %s",
				position, obj.Type())
		}
		return newError("return type mismatch: expected error, got %s", obj.Type())
	}

	expectedType, ok := typeMap[typeExpr.Name]
	if !ok {
		return newError("unknown return type: %s", typeExpr.Name)
	}

	if obj.Type() != expectedType {
		if position > 0 {
			return newError("return type mismatch: expected %s at position %d, got %s",
				typeExpr.Name, position, obj.Type())
		}
		return newError("return type mismatch: expected %s, got %s",
			typeExpr.Name, obj.Type())
	}

	return nil
}

// isError checks if an object is a programming error (should stop execution).
// bark error values (from err() builtin) are NOT programming errors.
func isError(obj object.Object) bool {
	if obj != nil && obj.Type() == object.ERROR_OBJ {
		if errObj, ok := obj.(*object.Error); ok {
			return errObj.IsProgrammingError
		}
	}
	return false
}

// isExecutionError checks if an object is a recoverable execution error
func isExecutionError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.EXEC_ERROR_OBJ
	}
	return false
}

// isCaptureStop checks if an object is a CaptureStop signal
func isCaptureStop(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.CAPTURE_STOP_OBJ
	}
	return false
}

// isChainStop checks if an object is a ChainStop signal
func isChainStop(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.CHAIN_STOP_OBJ
	}
	return false
}

// isbarkError checks if an object represents a bark error (non-empty map with msg field or Error object)
func isbarkError(obj object.Object) bool {
	// Check if it's a direct Error object
	if _, isErr := obj.(*object.Error); isErr {
		return true
	}

	// Check if it's a Map with msg field (bark error convention)
	mapObj, ok := obj.(*object.Map)
	if !ok {
		return false
	}

	// Empty map is not an error
	if len(mapObj.Pairs) == 0 {
		return false
	}

	// Check if it has a msg field
	if _, exists := mapObj.Pairs["msg"]; exists {
		return true
	}

	return false
}

// markShared marks an object as shared for copy-on-write optimization.
// Call this when an object is bound to a variable or passed to a function,
// indicating it may have multiple references.
func markShared(obj object.Object) {
	switch v := obj.(type) {
	case *object.Array:
		v.Share()
	case *object.Map:
		v.Share()
	}
}
