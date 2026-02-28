package bytecode

import (
	"fmt"
	"strings"

	"gitlab.com/bark-lang/barki/ast"
	"gitlab.com/bark-lang/barki/object"
)

// Compiler compiles AST nodes to bytecode
type Compiler struct {
	chunk       *Chunk
	scopeDepth  int
	locals      []Local
	upvalues    []Upvalue
	enclosing   *Compiler // For nested function compilation
	function    *CompiledFunction
	isMemoized  bool
	currentLine int
}

// Local represents a local variable
type Local struct {
	Name       string
	Depth      int
	IsCaptured bool
}

// Upvalue represents a captured variable from an enclosing scope
type Upvalue struct {
	Index   uint8
	IsLocal bool // true if captured from immediate enclosing function's locals
}

// CompileResult holds the result of compilation
type CompileResult struct {
	Function *CompiledFunction
	Errors   []string
}

// New creates a new compiler
func New() *Compiler {
	chunk := NewChunk()
	fn := &CompiledFunction{
		Chunk: chunk,
		Name:  "<script>",
		Arity: 0,
	}
	return &Compiler{
		chunk:    chunk,
		locals:   make([]Local, 0, 256),
		upvalues: make([]Upvalue, 0, 256),
		function: fn,
	}
}

// Compile compiles a program to bytecode
func Compile(program *ast.Program) *CompileResult {
	compiler := New()
	result := &CompileResult{}

	for _, stmt := range program.Statements {
		if err := compiler.compileStatement(stmt); err != nil {
			result.Errors = append(result.Errors, err.Error())
		}
	}

	// Implicit return null at end of script
	compiler.emit(OpNull)
	compiler.emit(OpReturn)

	// Store local variable names for string interpolation
	compiler.function.LocalNames = compiler.getLocalNames()

	result.Function = compiler.function
	return result
}

// Helper methods

func (c *Compiler) emit(op OpCode) {
	c.chunk.WriteOp(op, c.currentLine)
}

func (c *Compiler) emitWithOperand(op OpCode, operand int) int {
	return c.chunk.WriteOpWithOperand(op, operand, c.currentLine)
}

func (c *Compiler) emitWithOperands(op OpCode, operands []int) int {
	return c.chunk.WriteOpWithOperands(op, operands, c.currentLine)
}

func (c *Compiler) emitConstant(obj object.Object) {
	idx := c.chunk.AddConstant(obj)
	c.emitWithOperand(OpConstant, idx)
}

// emitJump emits a jump instruction with a placeholder offset
// Used for control flow compilation (if, match, loops)
var _ = (*Compiler).emitJump // Suppress unused warning - used for control flow

func (c *Compiler) emitJump(op OpCode) int {
	return c.chunk.EmitJump(op, c.currentLine)
}

// patchJump patches a previously emitted jump instruction
var _ = (*Compiler).patchJump // Suppress unused warning - used for control flow

func (c *Compiler) patchJump(offset int) {
	c.chunk.PatchJump(offset)
}

// emitLoop emits a loop instruction that jumps backward
var _ = (*Compiler).emitLoop // Suppress unused warning - used for control flow

func (c *Compiler) emitLoop(loopStart int) {
	c.chunk.EmitLoop(loopStart, c.currentLine)
}

func (c *Compiler) addName(name string) int {
	return c.chunk.AddName(name)
}

// Scope management

func (c *Compiler) beginScope() {
	c.scopeDepth++
}

func (c *Compiler) endScope() {
	c.scopeDepth--

	// Pop locals that are going out of scope
	for len(c.locals) > 0 && c.locals[len(c.locals)-1].Depth > c.scopeDepth {
		local := c.locals[len(c.locals)-1]
		if local.IsCaptured {
			c.emit(OpCloseUpval)
		} else {
			c.emit(OpPop)
		}
		c.locals = c.locals[:len(c.locals)-1]
	}
}

func (c *Compiler) addLocal(name string) {
	c.locals = append(c.locals, Local{
		Name:  name,
		Depth: c.scopeDepth,
	})
}

// getLocalNames returns the names of all local variables by slot index
func (c *Compiler) getLocalNames() []string {
	names := make([]string, len(c.locals))
	for i, local := range c.locals {
		names[i] = local.Name
	}
	return names
}

func (c *Compiler) resolveLocal(name string) int {
	for i := len(c.locals) - 1; i >= 0; i-- {
		if c.locals[i].Name == name {
			return i
		}
	}
	return -1
}

func (c *Compiler) resolveUpvalue(name string) int {
	if c.enclosing == nil {
		return -1
	}

	// Look in enclosing function's locals
	local := c.enclosing.resolveLocal(name)
	if local != -1 {
		c.enclosing.locals[local].IsCaptured = true
		return c.addUpvalue(uint8(local), true)
	}

	// Look in enclosing function's upvalues
	upvalue := c.enclosing.resolveUpvalue(name)
	if upvalue != -1 {
		return c.addUpvalue(uint8(upvalue), false)
	}

	return -1
}

func (c *Compiler) addUpvalue(index uint8, isLocal bool) int {
	// Check if already captured
	for i, uv := range c.upvalues {
		if uv.Index == index && uv.IsLocal == isLocal {
			return i
		}
	}

	c.upvalues = append(c.upvalues, Upvalue{Index: index, IsLocal: isLocal})
	c.function.UpvalueCount = len(c.upvalues)
	return len(c.upvalues) - 1
}

// Statement compilation

func (c *Compiler) compileStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		if err := c.compileExpression(s.Expression); err != nil {
			return err
		}
		c.emit(OpPop)
		return nil

	case *ast.FunctionStatement:
		return c.compileFunctionStatement(s)

	case *ast.MemoizedFunctionStatement:
		return c.compileMemoizedFunctionStatement(s)

	case *ast.BlockStatement:
		return c.compileBlockStatement(s)

	case *ast.ModuleStatement:
		// Module statements are handled at a higher level
		return nil

	case *ast.ImportStatement:
		// Import statements need runtime support
		return c.compileImportStatement(s)

	case *ast.IncludeStatement:
		// Include statements need runtime support
		return c.compileIncludeStatement(s)

	default:
		return fmt.Errorf("unknown statement type: %T", stmt)
	}
}

func (c *Compiler) compileBlockStatement(block *ast.BlockStatement) error {
	c.beginScope()
	for _, stmt := range block.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}
	c.endScope()
	return nil
}

// compileFunctionBody compiles a function body with implicit return of the last expression.
// In bark, if the last statement is an expression, its value is implicitly returned.
func (c *Compiler) compileFunctionBody(body *ast.BlockStatement) error {
	if body == nil || len(body.Statements) == 0 {
		// Empty body - return null
		c.emit(OpNull)
		c.emit(OpReturn)
		return nil
	}

	// Compile all statements except the last
	for i := 0; i < len(body.Statements)-1; i++ {
		if err := c.compileStatement(body.Statements[i]); err != nil {
			return err
		}
	}

	// Handle the last statement specially
	lastStmt := body.Statements[len(body.Statements)-1]

	// Check if last statement is an expression statement
	if exprStmt, ok := lastStmt.(*ast.ExpressionStatement); ok {
		// Compile the expression without popping - its value becomes the return value
		if err := c.compileExpression(exprStmt.Expression); err != nil {
			return err
		}
		c.emit(OpReturn)
		return nil
	}

	// Last statement is not an expression - compile normally and return null
	if err := c.compileStatement(lastStmt); err != nil {
		return err
	}
	c.emit(OpNull)
	c.emit(OpReturn)
	return nil
}

func (c *Compiler) compileFunctionStatement(fs *ast.FunctionStatement) error {
	// Check if this is a constant function (no params, body is just return(literal))
	// If so, we can pre-evaluate and store the constant directly
	if constantValue := c.tryEvaluateConstantFunction(fs.Parameters, fs.Body); constantValue != nil {
		c.emitConstant(constantValue)

		// Store in global or local depending on scope
		if c.scopeDepth == 0 {
			nameIdx := c.addName(fs.Name.Value)
			c.emitWithOperand(OpStoreGlobal, nameIdx)
		} else {
			c.addLocal(fs.Name.Value)
		}
		return nil
	}

	// Create a new compiler for the function body
	fnCompiler := &Compiler{
		chunk:      NewChunk(),
		enclosing:  c,
		scopeDepth: 0,
		locals:     make([]Local, 0, 256),
		upvalues:   make([]Upvalue, 0, 256),
		function: &CompiledFunction{
			Chunk: nil, // Set below
			Name:  fs.Name.Value,
			Arity: len(fs.Parameters),
		},
	}
	fnCompiler.function.Chunk = fnCompiler.chunk

	fnCompiler.beginScope()

	// Reserve slot 0 for the function/closure itself (standard VM calling convention)
	fnCompiler.addLocal("")

	// Add parameters as locals (starting at slot 1)
	for _, param := range fs.Parameters {
		fnCompiler.addLocal(param.Name.Value)
	}

	// Compile function body with implicit return of last expression
	if err := fnCompiler.compileFunctionBody(fs.Body); err != nil {
		return err
	}

	fnCompiler.function.UpvalueCount = len(fnCompiler.upvalues)
	fnCompiler.function.LocalNames = fnCompiler.getLocalNames()

	// Emit closure instruction in the enclosing compiler
	fnIdx := c.chunk.AddConstant(fnCompiler.function)
	c.emitWithOperand(OpClosure, fnIdx)

	// Emit upvalue info
	for _, uv := range fnCompiler.upvalues {
		if uv.IsLocal {
			c.chunk.Write(1, c.currentLine)
		} else {
			c.chunk.Write(0, c.currentLine)
		}
		c.chunk.Write(uv.Index, c.currentLine)
	}

	// Store in global or local depending on scope
	if c.scopeDepth == 0 {
		nameIdx := c.addName(fs.Name.Value)
		c.emitWithOperand(OpStoreGlobal, nameIdx)
	} else {
		c.addLocal(fs.Name.Value)
	}

	return nil
}

func (c *Compiler) compileMemoizedFunctionStatement(mfs *ast.MemoizedFunctionStatement) error {
	// Create a new compiler for the function body
	fnCompiler := &Compiler{
		chunk:      NewChunk(),
		enclosing:  c,
		scopeDepth: 0,
		locals:     make([]Local, 0, 256),
		upvalues:   make([]Upvalue, 0, 256),
		isMemoized: true,
		function: &CompiledFunction{
			Chunk:      nil,
			Name:       mfs.Name.Value,
			Arity:      len(mfs.Parameters),
			IsMemoized: true,
		},
	}
	fnCompiler.function.Chunk = fnCompiler.chunk

	fnCompiler.beginScope()

	// Reserve slot 0 for the function/closure itself (standard VM calling convention)
	fnCompiler.addLocal("")

	// Add parameters as locals (starting at slot 1)
	for _, param := range mfs.Parameters {
		fnCompiler.addLocal(param.Name.Value)
	}

	// Compile function body with implicit return of last expression
	if err := fnCompiler.compileFunctionBody(mfs.Body); err != nil {
		return err
	}

	fnCompiler.function.UpvalueCount = len(fnCompiler.upvalues)
	fnCompiler.function.LocalNames = fnCompiler.getLocalNames()

	// Emit closure instruction in the enclosing compiler
	fnIdx := c.chunk.AddConstant(fnCompiler.function)
	c.emitWithOperand(OpClosure, fnIdx)

	// Emit upvalue info
	for _, uv := range fnCompiler.upvalues {
		if uv.IsLocal {
			c.chunk.Write(1, c.currentLine)
		} else {
			c.chunk.Write(0, c.currentLine)
		}
		c.chunk.Write(uv.Index, c.currentLine)
	}

	// Store in global or local depending on scope
	if c.scopeDepth == 0 {
		nameIdx := c.addName(mfs.Name.Value)
		c.emitWithOperand(OpStoreGlobal, nameIdx)
	} else {
		c.addLocal(mfs.Name.Value)
	}

	return nil
}

func (c *Compiler) compileImportStatement(is *ast.ImportStatement) error {
	// Emit the path as a constant
	c.emitConstant(&object.String{Value: is.Path.Value})

	// Emit the alias name or path basename
	alias := is.Path.Value
	if is.Alias != nil {
		alias = is.Alias.Value
	}
	nameIdx := c.addName(alias)
	c.emitWithOperand(OpStoreGlobal, nameIdx)

	return nil
}

func (c *Compiler) compileIncludeStatement(is *ast.IncludeStatement) error {
	// Emit the path as a constant
	c.emitConstant(&object.String{Value: is.Path.Value})
	// Include needs VM runtime support
	return nil
}

// Expression compilation

func (c *Compiler) compileExpression(expr ast.Expression) error {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		c.emitConstant(&object.Integer{Value: e.Value})
		return nil

	case *ast.FloatLiteral:
		c.emitConstant(&object.Float{Value: e.Value})
		return nil

	case *ast.StringLiteral:
		// Check if the string needs interpolation
		if needsInterpolation(e.Value) {
			// Emit OpInterpolate which will do runtime variable lookup
			idx := c.chunk.AddConstant(&object.String{Value: e.Value})
			c.emitWithOperand(OpInterpolate, idx)
		} else {
			c.emitConstant(&object.String{Value: e.Value})
		}
		return nil

	case *ast.BooleanLiteral:
		if e.Value {
			c.emit(OpTrue)
		} else {
			c.emit(OpFalse)
		}
		return nil

	case *ast.Identifier:
		return c.compileIdentifier(e)

	case *ast.ArrayLiteral:
		return c.compileArrayLiteral(e)

	case *ast.MapLiteral:
		return c.compileMapLiteral(e)

	case *ast.CallExpression:
		return c.compileCallExpression(e)

	case *ast.MemberExpression:
		return c.compileMemberExpression(e)

	case *ast.LinkExpression:
		return c.compileLinkExpression(e)

	case *ast.AnonymousFunction:
		return c.compileAnonymousFunction(e)

	case *ast.TupleExpression:
		return c.compileTupleExpression(e)

	case *ast.TupleDestructure:
		return c.compileTupleDestructure(e)

	case *ast.CaptureExpression:
		return c.compileCaptureExpression(e)

	default:
		return fmt.Errorf("unknown expression type: %T", expr)
	}
}

func (c *Compiler) compileIdentifier(id *ast.Identifier) error {
	// Check special identifiers
	if id.Value == "null" {
		c.emit(OpNull)
		return nil
	}
	if id.Value == "true" {
		c.emit(OpTrue)
		return nil
	}
	if id.Value == "false" {
		c.emit(OpFalse)
		return nil
	}

	// Try local first
	local := c.resolveLocal(id.Value)
	if local != -1 {
		c.emitWithOperand(OpLoadLocal, local)
		return nil
	}

	// Try upvalue
	upvalue := c.resolveUpvalue(id.Value)
	if upvalue != -1 {
		c.emitWithOperand(OpLoadUpval, upvalue)
		return nil
	}

	// Must be global or builtin
	nameIdx := c.addName(id.Value)
	c.emitWithOperand(OpLoadGlobal, nameIdx)
	return nil
}

func (c *Compiler) compileArrayLiteral(al *ast.ArrayLiteral) error {
	for _, elem := range al.Elements {
		if err := c.compileExpression(elem); err != nil {
			return err
		}
	}
	c.emitWithOperand(OpArray, len(al.Elements))
	return nil
}

func (c *Compiler) compileMapLiteral(ml *ast.MapLiteral) error {
	// Use OrderedKeys to maintain insertion order
	for _, key := range ml.OrderedKeys {
		if err := c.compileExpression(key); err != nil {
			return err
		}
		value := ml.Pairs[key]
		if err := c.compileExpression(value); err != nil {
			return err
		}
	}
	c.emitWithOperand(OpMap, len(ml.OrderedKeys))
	return nil
}

func (c *Compiler) compileCallExpression(ce *ast.CallExpression) error {
	// Check if this is a builtin call
	if id, ok := ce.Function.(*ast.Identifier); ok {
		return c.compileBuiltinOrFunctionCall(id.Value, ce.Arguments)
	}

	// Member expression call (module.func())
	if me, ok := ce.Function.(*ast.MemberExpression); ok {
		return c.compileMemberCall(me, ce.Arguments)
	}

	// General case: compile function expression and arguments
	if err := c.compileExpression(ce.Function); err != nil {
		return err
	}

	for _, arg := range ce.Arguments {
		if err := c.compileExpression(arg); err != nil {
			return err
		}
	}

	c.emitWithOperand(OpCall, len(ce.Arguments))
	return nil
}

func (c *Compiler) compileBuiltinOrFunctionCall(name string, args []ast.Expression) error {
	// Compile arguments
	for _, arg := range args {
		if err := c.compileExpression(arg); err != nil {
			return err
		}
	}

	// Emit builtin call
	nameIdx := c.addName(name)
	c.emitWithOperands(OpBuiltin, []int{nameIdx, len(args)})
	return nil
}

func (c *Compiler) compileMemberCall(me *ast.MemberExpression, args []ast.Expression) error {
	// Compile as module.function(args)
	// First, emit the full member name
	if obj, ok := me.Object.(*ast.Identifier); ok {
		fullName := obj.Value + "." + me.Member.Value
		nameIdx := c.addName(fullName)

		// Compile arguments
		for _, arg := range args {
			if err := c.compileExpression(arg); err != nil {
				return err
			}
		}

		c.emitWithOperands(OpBuiltin, []int{nameIdx, len(args)})
		return nil
	}

	// Fallback: compile object, then member access, then call
	if err := c.compileExpression(me.Object); err != nil {
		return err
	}

	nameIdx := c.addName(me.Member.Value)
	c.emitWithOperand(OpMember, nameIdx)

	for _, arg := range args {
		if err := c.compileExpression(arg); err != nil {
			return err
		}
	}

	c.emitWithOperand(OpCall, len(args))
	return nil
}

func (c *Compiler) compileMemberExpression(me *ast.MemberExpression) error {
	if err := c.compileExpression(me.Object); err != nil {
		return err
	}
	nameIdx := c.addName(me.Member.Value)
	c.emitWithOperand(OpMember, nameIdx)
	return nil
}

func (c *Compiler) compileLinkExpression(le *ast.LinkExpression) error {
	// First, compile the left side
	if err := c.compileExpression(le.Left); err != nil {
		return err
	}

	// Handle right side based on type
	switch right := le.Right.(type) {
	case *ast.Identifier:
		// Variable binding: value > varName
		return c.compileLinkBind(right)

	case *ast.CallExpression:
		// Function call: value > func(args)
		return c.compileLinkCall(right)

	case *ast.AnonymousFunction:
		// Anonymous function: value > (x) { ... }
		return c.compileLinkAnonymous(right)

	case *ast.TupleDestructure:
		// Tuple destructure: (a, b) > (x, y)
		return c.compileTupleDestructure(right)

	case *ast.CaptureExpression:
		// Capture: (err, val) > capture(e, v)
		return c.compileCaptureExpression(right)

	case *ast.LinkExpression:
		// Chained link: a > b > c
		return c.compileLinkExpression(right)

	default:
		// General case: evaluate right and call it with left as argument
		if err := c.compileExpression(le.Right); err != nil {
			return err
		}
		c.emitWithOperand(OpLinkCall, 0)
		return nil
	}
}

func (c *Compiler) compileLinkBind(id *ast.Identifier) error {
	// Bind the top of stack to a variable
	if c.scopeDepth == 0 {
		nameIdx := c.addName(id.Value)
		c.emit(OpDup) // Keep value on stack for continuation
		c.emitWithOperand(OpLinkBind, nameIdx)
	} else {
		c.emit(OpDup) // Keep value on stack
		c.addLocal(id.Value)
	}
	return nil
}

func (c *Compiler) compileLinkCall(ce *ast.CallExpression) error {
	// Stack already has the left value
	// Compile additional arguments
	for _, arg := range ce.Arguments {
		if err := c.compileExpression(arg); err != nil {
			return err
		}
	}

	// Compile the function
	// Use OpUnpackBuiltin for link calls - it handles tuple unpacking at runtime
	// The first argument (left value from link operator) may be a tuple that needs unpacking
	if id, ok := ce.Function.(*ast.Identifier); ok {
		nameIdx := c.addName(id.Value)
		// Use OpUnpackBuiltin with extra arg count (not including the left value which may be a tuple)
		c.emitWithOperands(OpUnpackBuiltin, []int{nameIdx, len(ce.Arguments)})
		return nil
	}

	if me, ok := ce.Function.(*ast.MemberExpression); ok {
		if obj, ok := me.Object.(*ast.Identifier); ok {
			fullName := obj.Value + "." + me.Member.Value
			nameIdx := c.addName(fullName)
			c.emitWithOperands(OpUnpackBuiltin, []int{nameIdx, len(ce.Arguments)})
			return nil
		}
	}

	// General case - compile function expression and use OpLinkCall with tuple unpacking
	if err := c.compileExpression(ce.Function); err != nil {
		return err
	}
	c.emitWithOperand(OpLinkCall, len(ce.Arguments))
	return nil
}

func (c *Compiler) compileLinkAnonymous(af *ast.AnonymousFunction) error {
	// Compile the anonymous function
	if err := c.compileAnonymousFunction(af); err != nil {
		return err
	}

	// Stack state: [left_value, closure] where closure is on top
	// OpUnpackCall expects: [left_value, closure] and will pop closure then left_value
	// Use OpUnpackCall to handle tuple unpacking at runtime
	// If the left value is a tuple, its elements will be unpacked as arguments
	c.emitWithOperand(OpUnpackCall, len(af.Parameters))
	return nil
}

func (c *Compiler) compileAnonymousFunction(af *ast.AnonymousFunction) error {
	// Create a new compiler for the function body
	fnCompiler := &Compiler{
		chunk:      NewChunk(),
		enclosing:  c,
		scopeDepth: 0,
		locals:     make([]Local, 0, 256),
		upvalues:   make([]Upvalue, 0, 256),
		function: &CompiledFunction{
			Chunk: nil,
			Name:  "<anonymous>",
			Arity: len(af.Parameters),
		},
	}
	fnCompiler.function.Chunk = fnCompiler.chunk

	fnCompiler.beginScope()

	// Reserve slot 0 for the function/closure itself (standard VM calling convention)
	fnCompiler.addLocal("")

	// Add parameters as locals (starting at slot 1)
	for _, param := range af.Parameters {
		fnCompiler.addLocal(param.Name.Value)
	}

	// Compile function body with implicit return of last expression
	if err := fnCompiler.compileFunctionBody(af.Body); err != nil {
		return err
	}

	fnCompiler.function.UpvalueCount = len(fnCompiler.upvalues)
	fnCompiler.function.LocalNames = fnCompiler.getLocalNames()

	// Emit closure instruction in the enclosing compiler
	fnIdx := c.chunk.AddConstant(fnCompiler.function)
	c.emitWithOperand(OpClosure, fnIdx)

	// Emit upvalue info
	for _, uv := range fnCompiler.upvalues {
		if uv.IsLocal {
			c.chunk.Write(1, c.currentLine)
		} else {
			c.chunk.Write(0, c.currentLine)
		}
		c.chunk.Write(uv.Index, c.currentLine)
	}

	return nil
}

func (c *Compiler) compileTupleExpression(te *ast.TupleExpression) error {
	for _, elem := range te.Elements {
		if err := c.compileExpression(elem); err != nil {
			return err
		}
	}
	c.emitWithOperand(OpTuple, len(te.Elements))
	return nil
}

func (c *Compiler) compileTupleDestructure(td *ast.TupleDestructure) error {
	// The tuple value should already be on the stack
	// We need to destructure it into variables

	// For each identifier, emit code to extract and bind
	for i, id := range td.Identifiers {
		c.emit(OpDup)                                    // Duplicate the tuple
		c.emitConstant(&object.Integer{Value: int64(i)}) // Index
		c.emit(OpIndexGet)                               // Get element

		// Bind to variable
		if c.scopeDepth == 0 {
			nameIdx := c.addName(id.Value)
			c.emitWithOperand(OpStoreGlobal, nameIdx)
		} else {
			c.addLocal(id.Value)
		}
	}

	// Pop the original tuple
	c.emit(OpPop)
	return nil
}

func (c *Compiler) compileCaptureExpression(ce *ast.CaptureExpression) error {
	// The (error, value) tuple should be on the stack
	// capture(errVar, resultVar) binds both and either:
	// - continues with resultVar if no error
	// - stops chain if error present

	errIdx := c.addName(ce.ErrorVar.Value)
	valIdx := c.addName(ce.ResultVar.Value)

	// Emit capture instruction with both variable indices
	c.chunk.WriteOp(OpCapture, c.currentLine)
	c.chunk.Write(byte(errIdx>>8), c.currentLine)
	c.chunk.Write(byte(errIdx), c.currentLine)
	c.chunk.Write(byte(valIdx>>8), c.currentLine)
	c.chunk.Write(byte(valIdx), c.currentLine)

	return nil
}

// tryEvaluateConstantFunction checks if a function is a "constant function" -
// one with no parameters whose body is just return(literal). If so, it returns
// the constant value; otherwise it returns nil.
//
// This optimization avoids function call overhead for simple constant definitions like:
//
//	fn pi() { return(3.14159) }(float)
//
// Instead of creating a closure and calling it each time, we store the value directly.
func (c *Compiler) tryEvaluateConstantFunction(params []*ast.Parameter, body *ast.BlockStatement) object.Object {
	// Must have no parameters
	if len(params) > 0 {
		return nil
	}

	// Body must have exactly one statement
	if body == nil || len(body.Statements) != 1 {
		return nil
	}

	// That statement must be an expression statement
	exprStmt, ok := body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		return nil
	}

	// The expression must be a call to "return"
	callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
	if !ok {
		return nil
	}

	funcIdent, ok := callExpr.Function.(*ast.Identifier)
	if !ok || funcIdent.Value != "return" {
		return nil
	}

	// return() must have exactly one argument
	if len(callExpr.Arguments) != 1 {
		return nil
	}

	// The argument must be a literal value
	return c.tryExtractLiteral(callExpr.Arguments[0])
}

// tryExtractLiteral attempts to extract a constant object from a literal expression.
// Returns nil if the expression is not a simple literal.
func (c *Compiler) tryExtractLiteral(expr ast.Expression) object.Object {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return &object.Integer{Value: e.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: e.Value}
	case *ast.StringLiteral:
		// Only optimize non-interpolated strings
		if !needsInterpolation(e.Value) {
			return &object.String{Value: e.Value}
		}
		return nil
	case *ast.BooleanLiteral:
		return &object.Boolean{Value: e.Value}
	case *ast.Identifier:
		// Handle null, true, false identifiers
		switch e.Value {
		case "null":
			return &object.Null{}
		case "true":
			return &object.Boolean{Value: true}
		case "false":
			return &object.Boolean{Value: false}
		}
		return nil
	case *ast.ArrayLiteral:
		// Only optimize if all elements are literals
		elements := make([]object.Object, len(e.Elements))
		for i, elem := range e.Elements {
			val := c.tryExtractLiteral(elem)
			if val == nil {
				return nil
			}
			elements[i] = val
		}
		return &object.Array{Elements: elements}
	case *ast.MapLiteral:
		// Only optimize if all keys are string literals and all values are literals
		pairs := make(map[string]object.Object)
		keys := make([]string, 0, len(e.OrderedKeys))
		for _, keyExpr := range e.OrderedKeys {
			// Map keys must be string literals for this optimization
			keyLit, ok := keyExpr.(*ast.StringLiteral)
			if !ok {
				return nil
			}
			valExpr := e.Pairs[keyExpr]
			valObj := c.tryExtractLiteral(valExpr)
			if valObj == nil {
				return nil
			}
			keys = append(keys, keyLit.Value)
			pairs[keyLit.Value] = valObj
		}
		return &object.Map{Pairs: pairs, Keys: keys}
	default:
		return nil
	}
}

// needsInterpolation checks if a string contains interpolation placeholders
// or escaped braces that need runtime processing. Returns true if:
// - String contains {identifier} or {identifier.field} patterns (not numeric {0}, {1})
// - String contains escaped braces \{ or \} that need conversion
func needsInterpolation(s string) bool {
	i := 0
	for i < len(s) {
		// Check for escaped braces - these need runtime processing
		if i+1 < len(s) && s[i] == '\\' && (s[i+1] == '{' || s[i+1] == '}') {
			return true
		}

		if s[i] == '{' {
			// Find closing brace
			end := i + 1
			for end < len(s) && s[end] != '}' {
				end++
			}

			if end < len(s) {
				content := s[i+1 : end]
				// Check if content is a valid identifier (not numeric, not empty)
				if content != "" && !isNumeric(content) && isValidInterpolationContent(content) {
					return true
				}
			}
			i = end + 1
		} else {
			i++
		}
	}
	return false
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// isValidInterpolationContent checks if content looks like a valid identifier
// or identifier.field pattern
func isValidInterpolationContent(content string) bool {
	// Check for identifier.field pattern
	if dotIdx := strings.Index(content, "."); dotIdx != -1 {
		identifier := content[:dotIdx]
		field := content[dotIdx+1:]
		// Must have valid identifier and field, no nested dots
		return isValidIdentifier(identifier) && isValidIdentifier(field) && !strings.Contains(field, ".")
	}
	return isValidIdentifier(content)
}

// isValidIdentifier checks if a string is a valid bark identifier
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, ch := range s {
		if i == 0 {
			// First char must be letter or underscore
			if !isLetter(ch) {
				return false
			}
		} else {
			// Subsequent chars can be letter, digit, underscore, or ?
			if !isLetter(ch) && !isDigit(ch) {
				return false
			}
		}
	}
	return true
}

// isLetter checks if a rune is a letter, underscore, or question mark
func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == '?'
}

// isDigit checks if a rune is a digit
func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}
