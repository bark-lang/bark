# Ephemeral Tuples Implementation

## Overview

Ephemeral tuples are a lightweight language feature designed to solve a specific problem: passing multiple initial values to anonymous functions with multiple parameters. They exist only momentarily during evaluation and cannot be stored in variables.

## Problem Statement

Bark's link operator (`>`) passes a single value from left to right. This creates a limitation when initializing recursive anonymous functions that require multiple parameters:

```bark
// BROKEN: Cannot initialize both num and acc
fn factorial(n int) {
  n > (num int, acc int) {  // ERROR: acc is undefined
    num > lte?(1) > should_return
    should_return > return?(acc)

    num > sub(1) > next_num
    acc > mul(num) > next_acc
    repeat?(next_num, next_acc)
  }(int) > result

  return(result)
}(int)
```

The `repeat?()` builtin can pass multiple values for subsequent iterations, but there's no way to provide multiple initial values via the link operator.

## Solution: Ephemeral Tuples

### Syntax

A tuple is created using parentheses with comma-separated expressions:

```bark
(expr1, expr2, expr3)
```

**Important distinctions:**

- `(start, 1)` - Tuple (expressions without types)
- `(num int, acc int)` - Function parameters (identifiers with types)
- `factorial(5)` - Function call (identifier followed by parentheses)
- `(5 + 3)` - Grouping expression (single expression)

### Usage

Tuples are **only valid** on the left side of `>` when calling an anonymous function:

```bark
fn factorial(n int) {
  n > (start int) {
    (start, 1) > (num int, acc int) {
      num > lte?(1) > should_return
      should_return > return?(acc)

      num > sub(1) > next_num
      acc > mul(num) > next_acc
      repeat?(next_num, next_acc)
    }(int) > return()
  }(int) > result

  return(result)
}(int)
```

### Semantics

1. **Creation**: `(expr1, expr2, ...)` evaluates each expression and creates a temporary tuple object
2. **Unpacking**: When passed via `>` to an anonymous function, tuple values are unpacked into function parameters in order
3. **Ephemeral**: Tuples cannot be stored in variables, returned from functions, or used in any other context
4. **Type checking**: The number of tuple elements must match the number of function parameters

### Constraints

**Valid:**

```bark
(x, y) > (a int, b int) { ... }(int)
(1, 2, 3) > (x int, y int, z int) { ... }()
```

**Invalid:**

```bark
coords = (10, 20)  // ERROR: Cannot assign tuple to variable
fn get_coords() { return((10, 20)) }(tuple)  // ERROR: Cannot return tuple
arr > push((1, 2))  // ERROR: Cannot pass tuple to builtin
```

## Implementation Plan

### 1. Token (lexer/lexer.go)

No new tokens needed. Commas and parentheses already exist.

### 2. AST Node (ast/ast.go)

Add a new `TupleExpression` node:

```go
type TupleExpression struct {
    Token    token.Token // The '(' token
    Elements []Expression
}

func (te *TupleExpression) expressionNode() {}
func (te *TupleExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TupleExpression) String() string {
    elements := []string{}
    for _, el := range te.Elements {
        elements = append(elements, el.String())
    }
    return "(" + strings.Join(elements, ", ") + ")"
}
```

### 3. Parser (parser/parser.go)

Modify `parseGroupedExpression()` to detect tuples:

```go
func (p *Parser) parseGroupedExpression() ast.Expression {
    p.nextToken() // consume '('

    // Parse first expression
    exp := p.parseExpression(LOWEST)

    // Check if this is a tuple (has comma after first expression)
    if p.peekTokenIs(token.COMMA) {
        return p.parseTupleExpression(exp)
    }

    // Regular grouped expression
    if !p.expectPeek(token.RPAREN) {
        return nil
    }

    return exp
}

func (p *Parser) parseTupleExpression(firstElement ast.Expression) ast.Expression {
    tuple := &ast.TupleExpression{
        Token:    p.curToken, // Will be first element's token
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
```

### 4. Object Type (object/object.go)

Add a new `Tuple` object type:

```go
const (
    // ... existing types
    TUPLE_OBJ = "TUPLE"
)

type Tuple struct {
    Elements []Object
}

func (t *Tuple) Type() ObjectType { return TUPLE_OBJ }
func (t *Tuple) Inspect() string {
    elements := []string{}
    for _, el := range t.Elements {
        elements = append(elements, el.Inspect())
    }
    return "(" + strings.Join(elements, ", ") + ")"
}
```

### 5. Evaluator (evaluator/evaluator.go)

#### Evaluate TupleExpression

Add case in `Eval()`:

```go
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
```

#### Handle Tuple in Link Operator

Modify `evalLinkExpression()`:

```go
func evalLinkExpression(node *ast.LinkExpression, env *object.Environment) object.Object {
    left := Eval(node.Left, env)
    if isError(left) {
        return left
    }

    // Check if right side is an anonymous function literal
    if fnLit, ok := node.Right.(*ast.FunctionLiteral); ok {
        fn := Eval(fnLit, env).(*object.Function)

        // Check if left is a tuple
        if tuple, ok := left.(*object.Tuple); ok {
            // Unpack tuple elements as arguments
            return applyFunction(fn, tuple.Elements)
        }

        // Single value
        return applyFunction(fn, []object.Object{left})
    }

    // ... rest of link operator logic
}
```

### 6. Error Handling

Add validation to prevent invalid tuple usage:

```go
// In evalAssignmentExpression - prevent tuple assignment
if _, ok := value.(*object.Tuple); ok {
    return newError("cannot assign tuple to variable")
}

// In unwrapReturnValue - prevent tuple returns
if returnValue, ok := obj.(*object.ReturnValue); ok {
    if _, ok := returnValue.Value.(*object.Tuple); ok {
        return newError("cannot return tuple from function")
    }
    return returnValue.Value
}
```

### 7. Parameter Count Validation

Update `applyFunction()` to validate tuple unpacking:

```go
func applyFunction(fn *object.Function, args []object.Object) object.Object {
    // Validate argument count matches parameter count
    if len(args) != len(fn.Parameters) {
        return newError("wrong number of arguments: expected %d, got %d",
            len(fn.Parameters), len(args))
    }

    // ... rest of function
}
```

## Testing Strategy

### Unit Tests

**Parser Tests (parser/parser_test.go):**

- Parse simple tuple: `(1, 2)`
- Parse tuple with expressions: `(x, y + 1, z > add(3))`
- Parse nested tuples (should fail or parse as grouped expressions)
- Distinguish tuple from grouped expression

**Evaluator Tests (evaluator/evaluator_test.go):**

- Evaluate tuple creation
- Pass tuple to anonymous function
- Validate parameter count matching
- Error on tuple assignment
- Error on tuple return
- Error on tuple passed to builtin

### Integration Tests

**Examples:**

- Fix `examples/02_recursive_functions.bark` to use tuples
- Add tuple examples to `examples/13_anonymous_functions.bark`

## Edge Cases

1. **Empty tuple**: `()` - Should this be allowed? Probably parse as grouped expression with error.
2. **Single element**: `(x,)` - Trailing comma could indicate tuple, but not needed in Bark.
3. **Nested tuples**: `((1, 2), 3)` - Not needed for our use case; disallow or treat as grouped expressions.
4. **Tuple in other contexts**: Already covered by error handling.

## Documentation Updates

After implementation:

- Update `docs/spec.md` with tuple syntax section
- Update `README.md` with tuple examples
- Update `examples/02_recursive_functions.bark` to use correct syntax
- Add tuple section to `examples/13_anonymous_functions.bark`

## Future Considerations

If tuples prove useful, we could later add:

- First-class tuple support (storing in variables)
- Tuple unpacking in assignments: `(x, y) = get_coords()`
- Tuple types in function signatures

However, the current ephemeral implementation should solve the immediate problem while keeping the language simple.
