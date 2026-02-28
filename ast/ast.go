package ast

import (
	"bytes"
	"strings"

	"gitlab.com/bark-lang/barki/token"
)

// Node is the interface that all AST nodes must implement
type Node interface {
	TokenLiteral() string
	String() string
}

// Expression represents any node that produces a value
type Expression interface {
	Node
	expressionNode()
}

// Statement represents any node that performs an action
type Statement interface {
	Node
	statementNode()
}

// ============================================================================
// Program - Root node
// ============================================================================

// Program is the root node of the AST, containing all statements
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// ============================================================================
// Statements
// ============================================================================

// ModuleStatement represents: module name
type ModuleStatement struct {
	Token token.Token // the 'module' token
	Name  *Identifier
}

func (ms *ModuleStatement) statementNode()       {}
func (ms *ModuleStatement) TokenLiteral() string { return ms.Token.Literal }
func (ms *ModuleStatement) String() string {
	return "module " + ms.Name.String()
}

// IncludeStatement represents: include "path"
type IncludeStatement struct {
	Token token.Token // the 'include' token
	Path  *StringLiteral
}

func (is *IncludeStatement) statementNode()       {}
func (is *IncludeStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IncludeStatement) String() string {
	return "include " + is.Path.String()
}

// ImportStatement represents: import "path" OR import "path" as alias
type ImportStatement struct {
	Token token.Token    // the 'import' token
	Path  *StringLiteral // module path
	Alias *Identifier    // optional alias (nil if no 'as' clause)
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string {
	str := "import " + is.Path.String()
	if is.Alias != nil {
		str += " as " + is.Alias.String()
	}
	return str
}

// FunctionStatement represents: fn name(params) { body }(return_types)
type FunctionStatement struct {
	Token      token.Token // the 'fn' token
	Public     bool        // whether function is public (has 'pub' prefix)
	Name       *Identifier
	Parameters []*Parameter
	Body       *BlockStatement
	ReturnType *TypeList // can be nil for functions with no return
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionStatement) String() string {
	var out bytes.Buffer

	if fs.Public {
		out.WriteString("pub ")
	}

	out.WriteString("fn ")
	out.WriteString(fs.Name.String())
	out.WriteString("(")

	params := []string{}
	for _, p := range fs.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")

	if fs.Body != nil {
		out.WriteString(fs.Body.String())
	}

	if fs.ReturnType != nil {
		out.WriteString("(")
		out.WriteString(fs.ReturnType.String())
		out.WriteString(")")
	}

	return out.String()
}

// MemoizedFunctionStatement represents: mfn name(params) { body }(return_types)
// Memoized functions automatically cache results based on argument values
type MemoizedFunctionStatement struct {
	Token      token.Token // the 'mfn' token
	Name       *Identifier
	Parameters []*Parameter
	Body       *BlockStatement
	ReturnType *TypeList // can be nil for functions with no return
}

func (mfs *MemoizedFunctionStatement) statementNode()       {}
func (mfs *MemoizedFunctionStatement) TokenLiteral() string { return mfs.Token.Literal }
func (mfs *MemoizedFunctionStatement) String() string {
	var out bytes.Buffer

	out.WriteString("mfn ")
	out.WriteString(mfs.Name.String())
	out.WriteString("(")

	params := []string{}
	for _, p := range mfs.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")

	if mfs.Body != nil {
		out.WriteString(mfs.Body.String())
	}

	if mfs.ReturnType != nil {
		out.WriteString("(")
		out.WriteString(mfs.ReturnType.String())
		out.WriteString(")")
	}

	return out.String()
}

// ExpressionStatement wraps an expression to make it a statement
type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// BlockStatement represents a block of statements: { stmt1; stmt2; ... }
type BlockStatement struct {
	Token      token.Token // the '{' token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	out.WriteString("{\n")
	for _, s := range bs.Statements {
		out.WriteString("  ")
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	out.WriteString("}")
	return out.String()
}

// ============================================================================
// Expressions
// ============================================================================

// Identifier represents a variable or function name
type Identifier struct {
	Token token.Token // the IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral represents an integer: 42, 1_000_000
type IntegerLiteral struct {
	Token token.Token // the INT token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral represents a floating point number: 3.14, .5, 1.0e10
type FloatLiteral struct {
	Token token.Token // the FLOAT token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral represents a string: "hello" or `raw string`
type StringLiteral struct {
	Token token.Token // the STRING token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

// BooleanLiteral represents true or false
type BooleanLiteral struct {
	Token token.Token // the TRUE or FALSE token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// ArrayLiteral represents an array: [1, 2, 3]
type ArrayLiteral struct {
	Token    token.Token // the '[' token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, e := range al.Elements {
		elements = append(elements, e.String())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// MapLiteral represents a map: {"key": "value", "foo": "bar"}
type MapLiteral struct {
	Token       token.Token // the '{' token
	Pairs       map[Expression]Expression
	OrderedKeys []Expression // Maintains insertion order
}

func (ml *MapLiteral) expressionNode()      {}
func (ml *MapLiteral) TokenLiteral() string { return ml.Token.Literal }
func (ml *MapLiteral) String() string {
	var out bytes.Buffer
	pairs := []string{}
	for _, key := range ml.OrderedKeys {
		value := ml.Pairs[key]
		pairs = append(pairs, key.String()+": "+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

// CallExpression represents a function call: name(arg1, arg2)
type CallExpression struct {
	Token     token.Token // the '(' token
	Function  Expression  // Identifier or MemberExpression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// LinkExpression represents the link operator: a > b
type LinkExpression struct {
	Token token.Token // the '>' token
	Left  Expression
	Right Expression
}

func (le *LinkExpression) expressionNode()      {}
func (le *LinkExpression) TokenLiteral() string { return le.Token.Literal }
func (le *LinkExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(le.Left.String())
	out.WriteString(" > ")
	out.WriteString(le.Right.String())
	out.WriteString(")")
	return out.String()
}

// MemberExpression represents module member access: module.function
type MemberExpression struct {
	Token  token.Token // the '.' token
	Object Expression  // the module or object
	Member *Identifier // the member being accessed
}

func (me *MemberExpression) expressionNode()      {}
func (me *MemberExpression) TokenLiteral() string { return me.Token.Literal }
func (me *MemberExpression) String() string {
	var out bytes.Buffer
	out.WriteString(me.Object.String())
	out.WriteString(".")
	out.WriteString(me.Member.String())
	return out.String()
}

// AnonymousFunction represents: (params) { body }(return_types)
type AnonymousFunction struct {
	Token      token.Token // the '(' token
	Parameters []*Parameter
	Body       *BlockStatement
	ReturnType *TypeList
}

func (af *AnonymousFunction) expressionNode()      {}
func (af *AnonymousFunction) TokenLiteral() string { return af.Token.Literal }
func (af *AnonymousFunction) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	params := []string{}
	for _, p := range af.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")

	if af.Body != nil {
		out.WriteString(af.Body.String())
	}

	if af.ReturnType != nil {
		out.WriteString("(")
		out.WriteString(af.ReturnType.String())
		out.WriteString(")")
	}

	return out.String()
}

// TupleDestructure represents: expr > (a, b, c)
type TupleDestructure struct {
	Token       token.Token // the '(' token
	Expression  Expression
	Identifiers []*Identifier
}

func (td *TupleDestructure) expressionNode()      {}
func (td *TupleDestructure) TokenLiteral() string { return td.Token.Literal }
func (td *TupleDestructure) String() string {
	var out bytes.Buffer
	out.WriteString(td.Expression.String())
	out.WriteString(" > (")

	ids := []string{}
	for _, id := range td.Identifiers {
		ids = append(ids, id.String())
	}
	out.WriteString(strings.Join(ids, ", "))
	out.WriteString(")")

	return out.String()
}

// TupleExpression represents an ephemeral tuple: (expr1, expr2, expr3)
// Tuples are only valid on the left side of > when calling anonymous functions
type TupleExpression struct {
	Token    token.Token // the '(' token
	Elements []Expression
}

func (te *TupleExpression) expressionNode()      {}
func (te *TupleExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TupleExpression) String() string {
	elements := []string{}
	for _, el := range te.Elements {
		elements = append(elements, el.String())
	}
	return "(" + strings.Join(elements, ", ") + ")"
}

// CaptureExpression represents: expr > capture(errVar, resultVar)
// Captures error and result from an (error, value) tuple.
// On success: errVar = {}, resultVar = value, passes resultVar downstream
// On error: errVar = error, resultVar = zero value, chain stops
type CaptureExpression struct {
	Token     token.Token // the 'capture' identifier token
	ErrorVar  *Identifier // variable to bind the error
	ResultVar *Identifier // variable to bind the result
}

func (ce *CaptureExpression) expressionNode()      {}
func (ce *CaptureExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CaptureExpression) String() string {
	return "capture(" + ce.ErrorVar.String() + ", " + ce.ResultVar.String() + ")"
}

// ============================================================================
// Type system nodes
// ============================================================================

// TypeExpression represents a type name or compound type
// For simple types: Name is "int", "string", etc.
// For tuple types: Name is "tuple" and TupleTypes contains element types
// For function types: Name is "fn", ParamTypes contains parameter types, ReturnTypes contains return types
// For parameterized array: Name is "array", ElementType is the element type
// For parameterized map: Name is "map", KeyType and ValueType are the key/value types
// For union types: UnionTypes contains the alternative types (e.g., int | string)
type TypeExpression struct {
	Token       token.Token       // the type token
	Name        string            // type name ("int", "string", "tuple", "fn", "array", "map", etc.)
	TupleTypes  []*TypeExpression // for tuple types: (int, string) -> TupleTypes = [int, string]
	ParamTypes  []*TypeExpression // for function types: fn(int, string) -> ParamTypes = [int, string]
	ReturnTypes []*TypeExpression // for function types: fn()(int, string) -> ReturnTypes = [int, string]
	ElementType *TypeExpression   // for array[type]: ElementType is the element type
	KeyType     *TypeExpression   // for map[key, value]: KeyType is the key type
	ValueType   *TypeExpression   // for map[key, value]: ValueType is the value type
	UnionTypes  []*TypeExpression // for union types: int | string -> UnionTypes = [int, string]
}

func (te *TypeExpression) expressionNode()      {}
func (te *TypeExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TypeExpression) String() string {
	// Union type: int | string
	if te.UnionTypes != nil {
		types := []string{}
		for _, t := range te.UnionTypes {
			types = append(types, t.String())
		}
		return strings.Join(types, " | ")
	}
	if te.TupleTypes != nil {
		types := []string{}
		for _, t := range te.TupleTypes {
			types = append(types, t.String())
		}
		return "(" + strings.Join(types, ", ") + ")"
	}
	// Function type: fn(params)(returns)
	if te.ParamTypes != nil {
		params := []string{}
		for _, t := range te.ParamTypes {
			params = append(params, t.String())
		}
		returns := []string{}
		for _, t := range te.ReturnTypes {
			returns = append(returns, t.String())
		}
		return "fn(" + strings.Join(params, ", ") + ")(" + strings.Join(returns, ", ") + ")"
	}
	// Parameterized array type: array[element_type]
	if te.ElementType != nil {
		return "array[" + te.ElementType.String() + "]"
	}
	// Parameterized map type: map[key_type, value_type]
	if te.KeyType != nil && te.ValueType != nil {
		return "map[" + te.KeyType.String() + ", " + te.ValueType.String() + "]"
	}
	return te.Name
}

// TypeList represents a list of types: (string, int, bool)
type TypeList struct {
	Token token.Token // the '(' token
	Types []*TypeExpression
}

func (tl *TypeList) expressionNode()      {}
func (tl *TypeList) TokenLiteral() string { return tl.Token.Literal }
func (tl *TypeList) String() string {
	types := []string{}
	for _, t := range tl.Types {
		types = append(types, t.String())
	}
	return strings.Join(types, ", ")
}

// Parameter represents a function parameter: name type
type Parameter struct {
	Token token.Token // the identifier token
	Name  *Identifier
	Type  *TypeExpression
}

func (p *Parameter) String() string {
	return p.Name.String() + " " + p.Type.String()
}
