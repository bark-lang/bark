package parser

import (
	"fmt"
	"strconv"

	"gitlab.com/bark-lang/bark/ast"
	"gitlab.com/bark-lang/bark/lexer"
	"gitlab.com/bark-lang/bark/token"
)

// Precedence levels for operator parsing
const (
	_ int = iota
	LOWEST
	LINK  // >
	DOT   // .
	CALL  // function()
	INDEX // array[index]
	HIGHEST
)

var precedences = map[token.TokenType]int{
	token.GT:     LINK,
	token.DOT:    DOT,
	token.LPAREN: CALL,
	token.LBRACK: INDEX,
}

// Parser builds an AST from tokens
type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken      token.Token
	peekToken     token.Token
	peekPeekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// New creates a new Parser
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Initialize prefix parse functions
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.LBRACK, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseMapLiteral)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpressionOrAnonymousFunction)
	p.registerPrefix(token.MINUS, p.parseNegativeNumberLiteral)

	// Initialize infix parse functions
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.GT, p.parseLinkExpression)
	p.registerInfix(token.DOT, p.parseMemberExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	// Read three tokens to initialize curToken, peekToken, and peekPeekToken
	p.nextToken()
	p.nextToken()
	p.nextToken()

	return p
}

// Errors returns the parser errors
func (p *Parser) Errors() []string {
	return p.errors
}

// nextToken advances the parser to the next token
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.peekPeekToken
	p.peekPeekToken = p.l.NextToken()

	// Skip comments and newlines (they're not significant for parsing)
	for p.peekToken.Type == token.COMMENT || p.peekToken.Type == token.NEWLINE {
		p.peekToken = p.peekPeekToken
		p.peekPeekToken = p.l.NextToken()
	}
	for p.peekPeekToken.Type == token.COMMENT || p.peekPeekToken.Type == token.NEWLINE {
		p.peekPeekToken = p.l.NextToken()
	}
}

// ParseProgram parses the entire program
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		// Skip newlines and comments at statement level
		if p.curToken.Type == token.NEWLINE || p.curToken.Type == token.COMMENT {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		program.Statements = append(program.Statements, stmt)
		p.nextToken()
	}

	return program
}

// parseStatement parses a statement
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.MODULE:
		return p.parseModuleStatement()
	case token.IMPORT:
		return p.parseImportStatement()
	case token.INCLUDE:
		return p.parseIncludeStatement()
	case token.PUB:
		return p.parseFunctionStatement(true)
	case token.FN:
		return p.parseFunctionStatement(false)
	case token.MFN:
		return p.parseMemoizedFunctionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseModuleStatement parses: module name
func (p *Parser) parseModuleStatement() *ast.ModuleStatement {
	stmt := &ast.ModuleStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return stmt
}

// parseIncludeStatement parses: include "path"
func (p *Parser) parseIncludeStatement() *ast.IncludeStatement {
	stmt := &ast.IncludeStatement{Token: p.curToken}

	if !p.expectPeek(token.STRING) {
		return nil
	}

	stmt.Path = &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}

	return stmt
}

// parseImportStatement parses: import "path" [as alias]
func (p *Parser) parseImportStatement() *ast.ImportStatement {
	stmt := &ast.ImportStatement{Token: p.curToken}

	if !p.expectPeek(token.STRING) {
		return nil
	}

	stmt.Path = &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}

	// Check for optional "as alias"
	if p.peekTokenIs(token.AS) {
		p.nextToken() // consume 'as'
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		stmt.Alias = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	return stmt
}

// parseFunctionStatement parses: [pub] fn name(params) { body }(return_types)
func (p *Parser) parseFunctionStatement(isPublic bool) *ast.FunctionStatement {
	stmt := &ast.FunctionStatement{Public: isPublic}

	if isPublic {
		stmt.Token = p.curToken // 'pub' token
		if !p.expectPeek(token.FN) {
			return nil
		}
	} else {
		stmt.Token = p.curToken // 'fn' token
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	// Check for return type
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		stmt.ReturnType = p.parseTypeList()
		// If type list is empty (nil), we're already at RPAREN, skip expectPeek
		if stmt.ReturnType.Types != nil {
			// Non-empty type list, expect RPAREN next
			if !p.expectPeek(token.RPAREN) {
				return nil
			}
		}
		// Empty type list case: we're already at RPAREN from parseTypeList, no action needed
	}

	return stmt
}

// parseMemoizedFunctionStatement parses: mfn name(params) { body }(return_types)
func (p *Parser) parseMemoizedFunctionStatement() *ast.MemoizedFunctionStatement {
	stmt := &ast.MemoizedFunctionStatement{Token: p.curToken} // 'mfn' token

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	// Check for return type
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		stmt.ReturnType = p.parseTypeList()
		// If type list is empty (nil), we're already at RPAREN, skip expectPeek
		if stmt.ReturnType.Types != nil {
			// Non-empty type list, expect RPAREN next
			if !p.expectPeek(token.RPAREN) {
				return nil
			}
		}
		// Empty type list case: we're already at RPAREN from parseTypeList, no action needed
	}

	return stmt
}

// parseFunctionParameters parses function parameters: (a int, b string)
// Also supports tuple types: (a (int, string), b int)
func (p *Parser) parseFunctionParameters() []*ast.Parameter {
	params := []*ast.Parameter{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return params
	}

	p.nextToken()

	param := &ast.Parameter{
		Token: p.curToken,
		Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}

	// Accept IDENT, ERROR, LPAREN (for tuple types), or FN (for function types) as type
	if !p.peekTokenIs(token.IDENT) && !p.peekTokenIs(token.ERROR) && !p.peekTokenIs(token.LPAREN) && !p.peekTokenIs(token.FN) {
		msg := fmt.Sprintf("expected type, got %s instead", p.peekToken.Type)
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken()

	param.Type = p.parseTypeExpression()
	params = append(params, param)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to identifier

		param := &ast.Parameter{
			Token: p.curToken,
			Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
		}

		// Accept IDENT, ERROR, LPAREN (for tuple types), or FN (for function types) as type
		if !p.peekTokenIs(token.IDENT) && !p.peekTokenIs(token.ERROR) && !p.peekTokenIs(token.LPAREN) && !p.peekTokenIs(token.FN) {
			msg := fmt.Sprintf("expected type, got %s instead", p.peekToken.Type)
			p.errors = append(p.errors, msg)
			return nil
		}
		p.nextToken()

		param.Type = p.parseTypeExpression()
		params = append(params, param)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return params
}

// parseTypeList parses a type list: string, int, bool
// Also supports tuple types in the list: (int, string), error
func (p *Parser) parseTypeList() *ast.TypeList {
	typeList := &ast.TypeList{Token: p.curToken}
	typeList.Types = []*ast.TypeExpression{}

	p.nextToken()

	// Empty type list: () - for functions with no return type
	if p.curTokenIs(token.RPAREN) {
		// Don't consume the RPAREN - let caller handle it
		// But we need to go back one token, which we can't do
		// Instead, we'll handle this in parseFunctionStatement
		// For now, mark that we're at RPAREN by not adding any types
		// Actually, the issue is we've advanced too far
		// Solution: track if empty and handle specially in caller
		typeList.Types = nil // Explicitly mark as empty/none
		return typeList
	}

	typeList.Types = append(typeList.Types, p.parseTypeExpression())

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to type

		typeList.Types = append(typeList.Types, p.parseTypeExpression())
	}

	return typeList
}

// parseTypeExpression parses a type expression, which can be:
// - Simple type: int, string, bool, etc.
// - Tuple type: (int, string)
// - Function type: fn(int, string)(bool)
// - Parameterized array: array[int]
// - Parameterized map: map[string, int]
// - Union type: int | string
func (p *Parser) parseTypeExpression() *ast.TypeExpression {
	firstType := p.parseBaseType()
	if firstType == nil {
		return nil
	}

	// Check for union type: type | type | ...
	if !p.peekTokenIs(token.PIPE) {
		return firstType
	}

	return p.parseUnionType(firstType)
}

// parseBaseType parses a single type (not a union) - used internally
func (p *Parser) parseBaseType() *ast.TypeExpression {
	// Check if this is a tuple type (starts with LPAREN)
	if p.curTokenIs(token.LPAREN) {
		return p.parseTupleType()
	}

	// Check if this is a function type (starts with FN)
	if p.curTokenIs(token.FN) {
		return p.parseFunctionType()
	}

	// Simple type: IDENT or ERROR keyword
	typeExpr := &ast.TypeExpression{Token: p.curToken, Name: p.curToken.Literal}

	// Check for parameterized types: array[type] or map[key, value]
	if p.peekTokenIs(token.LBRACK) {
		switch typeExpr.Name {
		case "array":
			return p.parseArrayType(typeExpr)
		case "map":
			return p.parseMapType(typeExpr)
		}
	}

	return typeExpr
}

// parseUnionType parses a union type: type | type | ...
// Called when we have already parsed the first type and see PIPE
func (p *Parser) parseUnionType(firstType *ast.TypeExpression) *ast.TypeExpression {
	unionType := &ast.TypeExpression{
		Token:      firstType.Token,
		Name:       "union",
		UnionTypes: []*ast.TypeExpression{firstType},
	}

	// Collect all union members
	for p.peekTokenIs(token.PIPE) {
		p.nextToken() // consume the current type token position
		p.nextToken() // move past PIPE to next type

		nextType := p.parseBaseType()
		if nextType == nil {
			return nil
		}
		unionType.UnionTypes = append(unionType.UnionTypes, nextType)
	}

	return unionType
}

// parseTupleType parses a tuple type: (int, string)
// Called when curToken is LPAREN
func (p *Parser) parseTupleType() *ast.TypeExpression {
	startToken := p.curToken
	tupleType := &ast.TypeExpression{
		Token:      startToken,
		Name:       "tuple",
		TupleTypes: []*ast.TypeExpression{},
	}

	p.nextToken() // move past LPAREN

	// Empty tuple type: ()
	if p.curTokenIs(token.RPAREN) {
		return tupleType
	}

	// Parse first type
	elemType := p.parseTypeExpression()
	tupleType.TupleTypes = append(tupleType.TupleTypes, elemType)

	// Parse remaining types
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to next type

		elemType := p.parseTypeExpression()
		tupleType.TupleTypes = append(tupleType.TupleTypes, elemType)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return tupleType
}

// parseFunctionType parses a function type: fn(param_types)(return_types)
// Examples: fn(int, int)(int), fn(string)(bool), fn(int)()
// Called when curToken is FN
func (p *Parser) parseFunctionType() *ast.TypeExpression {
	startToken := p.curToken
	fnType := &ast.TypeExpression{
		Token:       startToken,
		Name:        "fn",
		ParamTypes:  []*ast.TypeExpression{},
		ReturnTypes: []*ast.TypeExpression{},
	}

	// Expect LPAREN for parameter types
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken() // move past LPAREN

	// Parse parameter types (can be empty)
	if !p.curTokenIs(token.RPAREN) {
		paramType := p.parseTypeExpression()
		fnType.ParamTypes = append(fnType.ParamTypes, paramType)

		for p.peekTokenIs(token.COMMA) {
			p.nextToken() // consume comma
			p.nextToken() // move to next type
			paramType := p.parseTypeExpression()
			fnType.ParamTypes = append(fnType.ParamTypes, paramType)
		}

		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	// Expect LPAREN for return types
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken() // move past LPAREN

	// Parse return types (can be empty)
	if !p.curTokenIs(token.RPAREN) {
		returnType := p.parseTypeExpression()
		fnType.ReturnTypes = append(fnType.ReturnTypes, returnType)

		for p.peekTokenIs(token.COMMA) {
			p.nextToken() // consume comma
			p.nextToken() // move to next type
			returnType := p.parseTypeExpression()
			fnType.ReturnTypes = append(fnType.ReturnTypes, returnType)
		}

		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	return fnType
}

// parseArrayType parses a parameterized array type: array[element_type]
// Called when curToken is "array" and peekToken is LBRACK
func (p *Parser) parseArrayType(typeExpr *ast.TypeExpression) *ast.TypeExpression {
	p.nextToken() // move past "array" to LBRACK
	p.nextToken() // move past LBRACK to element type

	typeExpr.ElementType = p.parseTypeExpression()

	if !p.expectPeek(token.RBRACK) {
		return nil
	}

	return typeExpr
}

// parseMapType parses a parameterized map type: map[key_type, value_type]
// Called when curToken is "map" and peekToken is LBRACK
func (p *Parser) parseMapType(typeExpr *ast.TypeExpression) *ast.TypeExpression {
	p.nextToken() // move past "map" to LBRACK
	p.nextToken() // move past LBRACK to key type

	typeExpr.KeyType = p.parseTypeExpression()

	if !p.expectPeek(token.COMMA) {
		return nil
	}

	p.nextToken() // move past COMMA to value type
	typeExpr.ValueType = p.parseTypeExpression()

	if !p.expectPeek(token.RBRACK) {
		return nil
	}

	return typeExpr
}

// parseBlockStatement parses a block: { statements }
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		// Skip newlines and comments
		if p.curToken.Type == token.NEWLINE || p.curToken.Type == token.COMMENT {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		block.Statements = append(block.Statements, stmt)
		p.nextToken()
	}

	return block
}

// parseExpressionStatement parses an expression statement
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	return stmt
}

// parseExpression parses an expression with precedence
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(token.EOF) && precedence < p.peekPrecedence() {
		// Don't continue parsing call expressions if peek token is on a different line
		// This prevents `expr\n(tuple)` from being parsed as `expr(tuple)`
		if p.peekTokenIs(token.LPAREN) && p.curToken.Line < p.peekToken.Line {
			return leftExp
		}

		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

// Prefix parse functions

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	// Remove underscores for parsing
	valueStr := ""
	for _, ch := range p.curToken.Literal {
		if ch != '_' {
			valueStr += string(ch)
		}
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	// Remove underscores for parsing
	valueStr := ""
	for _, ch := range p.curToken.Literal {
		if ch != '_' {
			valueStr += string(ch)
		}
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseNegativeNumberLiteral() ast.Expression {
	// Current token is MINUS, advance to the number
	p.nextToken()

	// Check if the next token is a number
	switch p.curToken.Type {
	case token.INT:
		lit := &ast.IntegerLiteral{Token: p.curToken}

		// Remove underscores for parsing
		valueStr := ""
		for _, ch := range p.curToken.Literal {
			if ch != '_' {
				valueStr += string(ch)
			}
		}

		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
			p.errors = append(p.errors, msg)
			return nil
		}

		lit.Value = -value // Negate the value
		return lit
	case token.FLOAT:
		lit := &ast.FloatLiteral{Token: p.curToken}

		// Remove underscores for parsing
		valueStr := ""
		for _, ch := range p.curToken.Literal {
			if ch != '_' {
				valueStr += string(ch)
			}
		}

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
			p.errors = append(p.errors, msg)
			return nil
		}

		lit.Value = -value // Negate the value
		return lit
	default:
		// If it's not a number, this is an error
		msg := fmt.Sprintf("expected number after '-', got %s", p.curToken.Type)
		p.errors = append(p.errors, msg)
		return nil
	}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACK)
	return array
}

func (p *Parser) parseMapLiteral() ast.Expression {
	mapLit := &ast.MapLiteral{Token: p.curToken}
	mapLit.Pairs = make(map[ast.Expression]ast.Expression)
	mapLit.OrderedKeys = []ast.Expression{}

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		mapLit.Pairs[key] = value
		mapLit.OrderedKeys = append(mapLit.OrderedKeys, key)

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return mapLit
}

func (p *Parser) parseGroupedExpressionOrAnonymousFunction() ast.Expression {
	// Look ahead to determine if this is:
	// 1. Anonymous function: (param type) { ... }(return_type)
	// 2. Tuple expression: (expr1, expr2, ...)
	// 3. Grouped expression: (expr)

	// Save position
	startToken := p.curToken

	// Try to parse as anonymous function parameters
	p.nextToken()

	// Empty parens could be either ()=> or just ()
	if p.curTokenIs(token.RPAREN) {
		// Check if next is { for anonymous function
		if p.peekTokenIs(token.LBRACE) {
			return p.parseAnonymousFunction(startToken, []*ast.Parameter{})
		}
		// Just empty parens, not valid
		p.peekError(token.IDENT)
		return nil
	}

	// If we see identifier followed by a type token, it's likely parameters
	// Type can be: IDENT (int, string), ERROR keyword, LPAREN for tuple types, or FN for function types
	if p.curTokenIs(token.IDENT) && (p.peekTokenIs(token.IDENT) || p.peekTokenIs(token.ERROR) || p.peekTokenIs(token.LPAREN) || p.peekTokenIs(token.FN)) {
		// Parse as anonymous function
		params := p.parseAnonymousFunctionParameters()
		return p.parseAnonymousFunction(startToken, params)
	}

	// Otherwise, parse as grouped expression or tuple
	exp := p.parseExpression(LOWEST)

	// Check if this is a tuple (has comma after first expression)
	if p.peekTokenIs(token.COMMA) {
		return p.parseTupleExpression(startToken, exp)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// parseTupleExpression parses a tuple: (expr1, expr2, ...)
// Called when we've already parsed the first expression and see a comma
func (p *Parser) parseTupleExpression(startToken token.Token, firstElement ast.Expression) ast.Expression {
	tuple := &ast.TupleExpression{
		Token:    startToken,
		Elements: []ast.Expression{firstElement},
	}

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to next expression

		element := p.parseExpression(LOWEST)
		if element == nil {
			return nil
		}
		tuple.Elements = append(tuple.Elements, element)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return tuple
}

// parseAnonymousFunctionParameters also supports tuple types: (a (int, string), b int)
func (p *Parser) parseAnonymousFunctionParameters() []*ast.Parameter {
	params := []*ast.Parameter{}

	param := &ast.Parameter{
		Token: p.curToken,
		Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}

	// Accept IDENT, ERROR, LPAREN (for tuple types), or FN (for function types) as type
	if !p.peekTokenIs(token.IDENT) && !p.peekTokenIs(token.ERROR) && !p.peekTokenIs(token.LPAREN) && !p.peekTokenIs(token.FN) {
		msg := fmt.Sprintf("expected type, got %s instead", p.peekToken.Type)
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken()

	param.Type = p.parseTypeExpression()
	params = append(params, param)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to identifier

		param := &ast.Parameter{
			Token: p.curToken,
			Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
		}

		// Accept IDENT, ERROR, LPAREN (for tuple types), or FN (for function types) as type
		if !p.peekTokenIs(token.IDENT) && !p.peekTokenIs(token.ERROR) && !p.peekTokenIs(token.LPAREN) && !p.peekTokenIs(token.FN) {
			msg := fmt.Sprintf("expected type, got %s instead", p.peekToken.Type)
			p.errors = append(p.errors, msg)
			return nil
		}
		p.nextToken()

		param.Type = p.parseTypeExpression()
		params = append(params, param)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return params
}

func (p *Parser) parseAnonymousFunction(startToken token.Token, params []*ast.Parameter) ast.Expression {
	fn := &ast.AnonymousFunction{Token: startToken, Parameters: params}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	fn.Body = p.parseBlockStatement()

	// Parse return type
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		fn.ReturnType = p.parseTypeList()
		// If type list is empty (nil), we're already at RPAREN
		if fn.ReturnType.Types != nil {
			// Non-empty type list, expect RPAREN next
			if !p.expectPeek(token.RPAREN) {
				return nil
			}
		}
		// For empty type list, we're at RPAREN, no need to advance
	}

	return fn
}

// Infix parse functions

func (p *Parser) parseLinkExpression(left ast.Expression) ast.Expression {
	expression := &ast.LinkExpression{
		Token: p.curToken,
		Left:  left,
	}

	precedence := p.curPrecedence()
	p.nextToken()

	// Check for tuple destructuring: expr > (a, b, c)
	// We need to distinguish between:
	//   Tuple destructuring: > (a, b, c) - just identifiers
	//   Anonymous function: > (param type) { ... } - param with type
	if p.curTokenIs(token.LPAREN) && p.peekTokenIs(token.IDENT) {
		// Use a lookahead to check the pattern
		if p.isTupleDestructuring() {
			return p.parseTupleDestructure(left)
		}
	}

	expression.Right = p.parseExpression(precedence)

	return expression
}

// isTupleDestructuring checks if the current position is at a tuple destructuring pattern
// Assumes curToken is LPAREN and peekToken is IDENT
// Returns true if pattern is (ident, ...) or (ident)
// Returns false if pattern is (ident type, ...) which is an anonymous function
func (p *Parser) isTupleDestructuring() bool {
	// We're at LPAREN, peekToken is IDENT
	// We need to check peekPeekToken:
	// Tuple: (a, b) or (a) -> peekPeekToken is COMMA or RPAREN
	// Anonymous function: (a int) -> peekPeekToken is IDENT (the type)
	return p.peekPeekToken.Type == token.COMMA || p.peekPeekToken.Type == token.RPAREN
}

func (p *Parser) parseTupleDestructure(left ast.Expression) ast.Expression {
	tuple := &ast.TupleDestructure{
		Token:      p.curToken,
		Expression: left,
	}

	tuple.Identifiers = []*ast.Identifier{}

	p.nextToken() // move to first identifier

	tuple.Identifiers = append(tuple.Identifiers, &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	})

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to identifier

		tuple.Identifiers = append(tuple.Identifiers, &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		})
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return tuple
}

func (p *Parser) parseMemberExpression(left ast.Expression) ast.Expression {
	expression := &ast.MemberExpression{
		Token:  p.curToken,
		Object: left,
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	expression.Member = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return expression
}

func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	// Check if this is a capture expression: capture(errVar, resultVar)
	if ident, ok := left.(*ast.Identifier); ok && ident.Value == "capture" {
		return p.parseCaptureExpression(ident.Token)
	}

	exp := &ast.CallExpression{Token: p.curToken, Function: left}
	exp.Arguments = p.parseExpressionList(token.RPAREN)

	// Validate that no arguments are function calls
	// Function calls must be chained, not nested
	for _, arg := range exp.Arguments {
		if !p.isValidFunctionArgument(arg) {
			p.errors = append(p.errors,
				"function calls cannot be used as arguments; use chaining instead (e.g., 'foo() > bar' not 'bar(foo())')",
			)
			return nil
		}
	}

	return exp
}

// isValidFunctionArgument checks if an expression is valid as a function argument.
// Function calls are not allowed as arguments - they must be chained instead.
// Valid arguments: identifiers, literals, arrays, maps, member expressions (for module access)
func (p *Parser) isValidFunctionArgument(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.Identifier:
		return true
	case *ast.IntegerLiteral:
		return true
	case *ast.FloatLiteral:
		return true
	case *ast.StringLiteral:
		return true
	case *ast.BooleanLiteral:
		return true
	case *ast.ArrayLiteral:
		// Check array elements recursively
		for _, el := range e.Elements {
			if !p.isValidFunctionArgument(el) {
				return false
			}
		}
		return true
	case *ast.MapLiteral:
		// Check map values recursively
		for _, val := range e.Pairs {
			if !p.isValidFunctionArgument(val) {
				return false
			}
		}
		return true
	case *ast.MemberExpression:
		// Module.function access is allowed (for passing function references)
		return true
	case *ast.CallExpression:
		// Function calls are NOT allowed as arguments
		return false
	case *ast.LinkExpression:
		// Link expressions are NOT allowed as arguments
		return false
	default:
		// Other expression types are allowed
		return true
	}
}

// parseCaptureExpression parses: capture(errVar, resultVar)
// Called when we've seen "capture" and current token is "("
func (p *Parser) parseCaptureExpression(captureToken token.Token) ast.Expression {
	// Current token is '('
	// We need exactly two identifiers

	// Move past '('
	if !p.expectPeek(token.IDENT) {
		p.errors = append(p.errors, "capture requires error variable as first argument")
		return nil
	}
	errorVar := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Expect comma
	if !p.expectPeek(token.COMMA) {
		p.errors = append(p.errors, "capture requires two arguments: capture(errVar, resultVar)")
		return nil
	}

	// Get result variable
	if !p.expectPeek(token.IDENT) {
		p.errors = append(p.errors, "capture requires result variable as second argument")
		return nil
	}
	resultVar := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Expect closing paren
	if !p.expectPeek(token.RPAREN) {
		p.errors = append(p.errors, "expected ')' after capture arguments")
		return nil
	}

	return &ast.CaptureExpression{
		Token:     captureToken,
		ErrorVar:  errorVar,
		ResultVar: resultVar,
	}
}

// Helper functions

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to next expression
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead at line %d, column %d",
		t, p.peekToken.Type, p.peekToken.Line, p.peekToken.Column)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found at line %d, column %d",
		t, p.curToken.Line, p.curToken.Column)
	p.errors = append(p.errors, msg)
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
