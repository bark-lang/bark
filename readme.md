# bark

an experimental programming language that emphasizes simplicity and consistency through function chaining.

---

## origin

Unlike [Slim](https://github.com/slim-template/slim), which had a clear application in mind from the start, bark began as a purely syntactical challenge. I defined behavior rules like the left-to-right process semantics, output of one function as input to the next, etc... and wanted to see if I could make them work.

This idea first came to mind around ten years ago. Over the years, bark would randomly resurface and some unresolved syntax issue would find its resolution. I wasn't close enough to a complete syntax to attempt implementation, even though I did try couple of times with both Go and Rust, but those didn't go far because I'd quickly run into an unresolved issue with the syntax.

Eventually the syntax felt well-defined enough to try again. That's where we are today.

Now that it actually works, I have some thoughts on where bark might be useful. I'm also working through [Rosetta Code](https://rosettacode.org) examples and adding them to bark's examples. This has been a great exercise for finding where bark falls short and where additional built-in functions or some functional tweaks could simplify things.

---

## what it is

bark is an experiment in consistent code construction. everything flows in one direction:

- **left-to-right** through chains
- **top-down** through execution (each line runs unless exited early)

traditional operators like `=` violate this by assigning right-to-left. bark eliminates this inconsistency by replacing operators with builtins:

```bark
// traditional: value flows right-to-left
x = 5

// bark: value flows left-to-right
5 > x
```

```bark
// traditional: operators between operands
x + y == z

// bark: functions chain left-to-right
x > add(y) > eq?(z)
```

### features

- **link operator `>`** - passes output left-to-right, replaces assignment
- **builtins replace operators** - `add()`, `sub()`, `eq?()` instead of `+`, `-`, `==`
- **conditional suffix `?`** - `return?()`, `repeat?()` execute only when input is truthy
- **no global variables** - all state flows through function chains
- **memoized functions** - `mfn` keyword for automatic result caching

---

## tenets

- simplicity
- consistency

---

## getting started

```bash
# start the REPL
go run cmd/repl/main.go

# run a file
go run cmd/bark/main.go examples/01_basic_functions.bark

# run with bytecode VM (faster execution)
go run cmd/bark/main.go -b examples/01_basic_functions.bark
```

---

## the link operator `>`

the link operator passes output from the left as input to the right:

```bark
"hello" > println()
```

the return value on the left becomes the first parameter on the right:

```bark
5 > add(3)    // add(5, 3) = 8
8 > mul(2)    // mul(8, 2) = 16
```

---

## chains

a chain is a sequence of links:

```bark
"hello" > str.upper() > str.concat(" WORLD") > println()
// outputs: HELLO WORLD
```

chains read left-to-right, like a pipeline:

```bark
fn process(name string) {
  name > str.trim() > str.lower() > validate() > save()
}()
```

---

## string interpolation

strings support variable interpolation with `{identifier}` syntax:

```bark
"Alice" > name
30 > age
"Hello, {name}! You are {age} years old." > println()
// outputs: Hello, Alice! You are 30 years old.
```

map field access is also supported:

```bark
{"name": "Bob", "city": "Seattle"} > user
"{user.name} lives in {user.city}" > println()
// outputs: Bob lives in Seattle
```

use `\{` and `\}` for literal braces.

---

## data structures

bark has three data structures:

```bark
[1, 2, 3]              // array - ordered collection
{"name": "Alice"}      // map - key-value pairs
(1, "hello", true)     // tuple - fixed structure
```

### when to use each

| use case | best choice |
|----------|-------------|
| list of same-type items | array |
| named fields, dynamic keys | map |
| fixed structure, positional data | tuple |
| multiple return values | tuple |
| flowing data through `>` | tuple |

### tuple types

tuples can have type annotations to validate structure:

```bark
fn format_point(point (int, int)) {
    point > (x int, y int) {
        println("({x}, {y})")
    }()
}()

(10, 20) > format_point()  // outputs: (10, 20)
```

type mismatches produce clear errors:

```text
type mismatch: parameter 'point' expects (int, int), got STRING
type mismatch: parameter 'point' expects tuple with 2 elements, got 3
```

---

## anonymous functions

anonymous functions provide conditional logic and iteration. they are defined inline and receive values from the chain:

```bark
"hello" > (msg string) {
  msg > println()
}
```

anonymous functions require the same signature syntax as regular functions, including explicit return types when returning values.

---

## repeat and return

`repeat()` and `return()` replace traditional loops and conditionals.

### loops with repeat

`repeat()` re-executes the anonymous function. `repeat?()` does so conditionally:

```bark
// print 1 to 5
1 > (n int) {
  n > println()
  n > add(1) > next
  next > lte?(5) > repeat?(next)
}
```

### early exit with return

`return()` exits the current function. `return?()` does so conditionally:

```bark
// case/switch style
fn day_name(abbr string) {
  abbr > (d string) {
    d > eq?("mon") > return?("monday")
    d > eq?("tue") > return?("tuesday")
    d > eq?("wed") > return?("wednesday")
    return("unknown")
  }(string) > result
  return(result)
}(string)
```

### iterating collections

use `each()` to iterate over arrays and maps:

```bark
// array iteration - fn receives (array, index)
[10, 20, 30] > each((arr array, i int) {
  arr > get(i) > val
  i > add(1) > pos
  println("{0}: {1}", pos, val)
})
// outputs:
// 1: 10
// 2: 20
// 3: 30

// map iteration - fn receives (map, key)
{"a": 1, "b": 2} > each((m map, key string) {
  m > get(key) > val
  println("{0} = {1}", key, val)
})
// outputs:
// a = 1
// b = 2
```

---

## constants

bark has no special constant syntax. constants are simply functions with static return values:

```bark
fn pi() {
  return(3.14159)
}(float)

fn max_users() {
  return(100)
}(int)

pi() > mul(2) > println()  // 6.28318
```

this maintains consistency: everything is a function, everything uses parentheses.

---

## memoized functions

use `mfn` instead of `fn` to define memoized functions. results are cached based on arguments:

```bark
// without memoization: O(2^n) - unusably slow for n > 30
// with memoization: O(n) - instant for any reasonable n
mfn fib(n int) {
  n > lte?(1) > return?(n)
  n > sub(1) > fib() > a
  n > sub(2) > fib() > b
  a > add(b) > return()
}(int)

40 > fib() > println()  // 102334155
```

memoization is beneficial for:

- recursive functions with overlapping subproblems (fibonacci, factorial)
- expensive computations called repeatedly with same inputs
- dynamic programming algorithms

---

## testing

bark has built-in testing support with `assert()` and `assert_error()` functions. test files live in the `tests/` directory:

```bark
// tests/math_test.bark
1 > add(1) > assert(2)
10 > sub(3) > assert(7)
"hello" > assert("hello")
```

assert errors with `assert_error()`:

```bark
err("invalid input") > assert_error("invalid input")
```

run tests with the `bark test` command:

```bash
# run all tests
bark test

# run a specific test file
bark test tests/math_test.bark
```

---

## explore more

- **[examples/](examples/)** - working code examples demonstrating language features
- **[docs/spec.md](docs/spec.md)** - formal language specification
- **[docs/builtins.md](docs/builtins.md)** - complete reference for all built-in functions

---

## credits

bark is built with:

- [Go](https://go.dev/)
- [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto)
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (optional)
- [BurntSushi/toml](https://github.com/BurntSushi/toml)
- [google/uuid](https://github.com/google/uuid)

---

## references

First and foremost, the books by [Thorsten Ball](https://thorstenball.com/):

- [Writing an Interpreter in Go](https://interpreterbook.com/)
- [Writing a Compiler in Go](https://compilerbook.com/)

Projects found via above references (alphabetical order):

- [ghost](https://github.com/ghost-language/ghost)
- [go](https://github.com/golang/go)
- [goby](https://github.com/goby-lang/goby)
- [langur](https://en.langurlang.org/)
- [ludwig](https://github.com/ludwig-lang/ludwig-lang)
- [monkey](https://github.com/skx/monkey)
- [monkey lang](https://github.com/bradford-hamilton/monkey-lang)
