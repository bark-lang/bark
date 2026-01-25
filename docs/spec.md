# Bark Language Specification

## syntax

The syntax is specified using extended backaus-naur format (ebnf)

```ebnf
production  = name "=" [ expr ] "." .
expr        = alt { "|" alt } .
alt         = term { term } .
term        = name | token | group | option | repetition .
group       = "(" expr ")" .
option      = "[" expr "]" .
repetition  = "{" expr "}" .
```

productions are expr constructed from terms and the following operators, in increasing precedence:

```ebnf
|   alternation
()  grouping
[]  option (0 or 1 times)
{}  repetition (0 to n times)
```

an expr specifies the computation of a value by applying operators and functions to operands

```ebnf
expr_list = expr { "," expr } .
block     = "{" { expr_list newline } "}"
```

## source code representation

### character classes

the following terms are used to denote specific Unicode character classes:

```ebnf
newline        = /* code point U+000A */ .
unicode_char   = /* arbitrary code point except newline */ .
unicode_letter = /* code point classified as "letter" */ .
unicode_digit  = /* code point classified as "number, decimal digit" */ .
```

### letters and digits

the underscore character _ (U+005F) is considered a letter.

```ebnf
letter        = unicode_letter | "_" .
decimal_digit = "0" … "9" .
```

The form a … b represents the set of characters from a through b as alternatives. The horizontal ellipsis … is also used elsewhere in the spec to informally denote various enumerations or code snippets that are not further specified. The character … (as opposed to the three characters ...) is not a token.

## lexical elements

### tokens

bark is comprised of tokens. there are four categories of tokens - identifiers, keywords, operators and punctuation, and literals. white space, formed from spaces (U+0020), horizontal tabs (U+0009), carriage returns (U+000D), and newlines (U+000A), is ignored except as it separates tokens that would otherwise combine into a single token.

while breaking the input into tokens, the next token is the longest sequence of characters that form a valid token.

### identifiers

identifiers name code entities such as variables and types. An identifier is a sequence of one or more letters and digits. The first character in an identifier must be a letter.

```ebnf
ident = letter { letter | unicode_digit } .
ident_list = ident { "," ident } .
```

### keywords

the following keywords are reserved and may not be used as identifiers.

```ebnf
as error fn import include mfn module pub type
```

### builtin functions

builtin functions are organized into two categories:

1. **global namespace** - fundamental operations available without module prefix
2. **module namespace** - specialized operations accessed via module.function() syntax

#### global namespace builtins

the following builtin functions are available in the global namespace and may not be redefined:

```ebnf
// arithmetic (4)
add div mul sub

// comparison (9)
absent? eq? gt? gte? lt? lte? ne? not present?

// arrays (2)
empty? len

// iteration (1)
each

// data structures - polymorphic (12)
excludes? first get head includes? last next prev reverse set size tail

// control flow (5)
break? continue? repeat repeat? return?

// error handling (5)
capture err err_add_context err_context err_msg

// core (10)
eprint eprint? eprintln eprintln? print print? println println? return to_string

// parallel processing (8)
all_absent? all_present? first_error parallel parallel_all parallel_limited parallel_race parallel_strict

// testing (2) - only available in test mode (bark test)
assert assert_error
```

#### module namespace builtins

the following builtin modules provide specialized functionality:

**http module** - HTTP operations

```ebnf
http.delete http.get http.post http.put http.request
```

**file module** - file system operations

```ebnf
file.absent? file.append file.delete file.exists? file.info file.read file.write
```

**env module** - environment variable operations

```ebnf
env.absent? env.all env.get env.get_or env.present?
```

**json module** - JSON encoding/decoding

```ebnf
json.parse json.stringify json.stringify_pretty
```

**time module** - date and time operations

```ebnf
time.format time.format_iso8601 time.now time.now_ms time.parse time.parse_iso8601
```

**base64 module** - base64 encoding/decoding

```ebnf
base64.decode base64.encode
```

**url module** - URL encoding/decoding/parsing

```ebnf
url.decode url.encode url.parse
```

**regex module** - regular expression operations

```ebnf
regex.find regex.find_all regex.match? regex.replace regex.split
```

**dir module** - directory operations

```ebnf
dir.absent? dir.exists? dir.list
```

**str module** - string manipulation

```ebnf
str.alphanumeric? str.concat str.ends_with? str.format str.join str.lower str.numeric? str.replace str.split str.starts_with? str.trim str.upper
```

**math module** - mathematical operations

```ebnf
math.abs math.acos math.asin math.atan math.ceil math.cos math.e math.even?
math.exp math.floor math.log math.log10 math.max math.min math.mod math.odd?
math.pi math.pow math.round math.sin math.sqrt math.tan math.to_float math.to_int
```

**array module** - array operations

```ebnf
array.append_to array.dedupe array.pop array.push array.range array.shift array.slice array.unshift
```

**map module** - map operations

```ebnf
map.del map.entries map.get_or map.key_absent? map.key_present? map.keys map.merge map.values
```

**security module** - input validation and sanitization

```ebnf
security.email? security.generate_nonce security.hash_key security.html_escape
security.safe_command? security.sanitize_path security.shell_escape security.sql_escape
security.strip_tags security.url?
```

**crypto module** - cryptographic operations

```ebnf
crypto.aes_decrypt crypto.aes_encrypt crypto.argon2_hash crypto.argon2_verify
crypto.bcrypt_hash crypto.bcrypt_verify crypto.hmac_sha256 crypto.hmac_sha512
crypto.hmac_verify crypto.random_bytes crypto.random_string crypto.sha256 crypto.sha512
```

**sql module** - database operations (requires `-tags sql` build flag)

```ebnf
sql.open sql.close sql.query sql.exec sql.begin sql.commit sql.rollback
```

see [builtins.md](builtins.md) for complete function signatures and descriptions

### operators and punctuation

the following character sequences represent operators and punctuation

```ebnf
. , > [ ] ( ) { }
```

### literals

### numeric literals

```ebnf
int_lit         = "0" | ( "1" … "9" ) [ [ "_" ] decimal_digits ] .
float_lit       = decimal_digits "." [ decimal_digits ] [ decimal_exponent ] |
                  decimal_digits decimal_exponent |
                  "." decimal_digits [ decimal_exponent ] .
decimal_digits  = decimal_digit { [ "_" ] decimal_digit } .

```

### string literals

```ebnf
string_lit             = raw_string_lit | interpreted_string_lit .
raw_string_lit         = "`" { unicode_char | newline } "`" .
interpreted_string_lit = `"` { unicode_value } `"` .

unicode_value    = unicode_char | escaped_char .

escaped_char = `\` ( "a" | "b" | "f" | "n" | "r" | "t" | "v" | `\` | "'" | `"` | "{" | "}" ) .
```

### string interpolation

Interpreted strings support variable interpolation using `{identifier}` syntax:

```bark
"Alice" > name
30 > age
"Hello, {name}! You are {age} years old." > println()
// Output: Hello, Alice! You are 30 years old.
```

Single-level map field access is also supported:

```bark
{"name": "Bob", "city": "Seattle"} > user
"{user.name} lives in {user.city}" > println()
// Output: Bob lives in Seattle
```

**Rules:**

- Only bare identifiers and single-level field access are allowed (no expressions)
- Undefined variables pass through unchanged: `"{undefined}"` outputs `{undefined}`
- Positional placeholders `{0}`, `{1}` are preserved for use with `println()` format arguments
- Use `\{` and `\}` to escape braces: `"Use \{name\} syntax"` outputs `Use {name} syntax`
- Raw strings (backticks) do not support interpolation

## types

### primitive types

bark has the following built-in primitive types:

- `string` - UTF-8 encoded text
- `bool` - Boolean value (true or false)
- `int` - Signed 64-bit integer
- `uint` - Unsigned 64-bit integer
- `float` - IEEE-754 64-bit floating-point number
- `error` - Built-in error type with message and optional context
- `regex` - Compiled regular expression pattern

### Zero Values

Bark has no `nil` value. Every type has a defined zero value:

| type | zero value |
| ------ | ------------ |
| `string` | `""` (empty string) |
| `bool` | `false` |
| `int` | `0` |
| `uint` | `0` |
| `float` | `0.0` |
| `error` | `{}` |
| `regex` | (special) |
| `map` | `{}` (empty map) |
| `array` | `[]` (empty array) |
| `fn` | (special) |

**truthy and falsy values:**

falsy values (evaluate to false in boolean contexts):

- `false`
- `0` (int, uint)
- `0.0` (float)
- `""` (empty string)
- `{}` (empty map)
- `[]` (empty array)

all other values are truthy.

**note:** `error` type is special - its zero value is `{}` (empty map), which represents "no error"

see [implementation/zero_values.md](implementation/zero_values.md) for complete semantics

### type system fundamentals

bark uses **static typing with type inference**:

- **static typing** - all types are determined at compile time
- **type inference** - types are inferred from expressions; explicit annotations rarely needed
- **explicit function signatures** - function parameters and return types must be declared
- **no implicit conversions** - type conversions must be explicit using `to_string()`, `math.to_int()`, etc.
- **generic type `t`** - represents "any type" in function signatures

**type inference examples:**

```bark
fn example() {
  42 > x              // x inferred as int
  "hello" > y         // y inferred as string
  [1, 2, 3] > nums    // nums inferred as array
}
```

**function signatures** must be explicit:

```bark
fn add(a int, b int) {
  a > add(b) > result
  return(result)
}(int)  // Return type required
```

**no implicit conversions:**

```bark
42 > to_string() > s       // Must explicitly convert int to string
"123" > math.to_int() > n  // Must explicitly convert string to int
```

see [docs/implementation/type_system.md](implementation/type_system.md) for complete type system specification

### type syntax

```ebnf
type = union_type .
union_type = base_type { "|" base_type } .
base_type = type_name | type_lit | tuple_type .
type_name = ident | q_ident .
type_lit = list_type | map_type | fn_type .
tuple_type = "(" type { "," type } ")" .

type_list = type { "," type } .
```

#### union types

union types allow a parameter or return value to accept multiple types:

```bark
// parameter accepts either int or string
42 > (x int | string) { return(x) }(int | string)
"hello" > (x int | string) { return(x) }(int | string)

// three-way union
100 > (x int | string | bool) { return(x) }(int | string | bool)

// union in tuple elements
(42, true) > (pair (int | string, bool | int)) { ... }

// union in array elements
[1, "two", 3] > (arr array[int | string]) { ... }

// union in map values
{"a": 1, "b": "two"} > (m map[string, int | string]) { ... }
```

union type validation:

- value must match at least one type in the union
- types are checked in declaration order (left to right)
- error messages show all alternatives: `expects int | string, got BOOLEAN`

#### tuple types

tuple types specify the types of each element in a tuple:

```bark
// parameter accepts tuple of (int, string)
(1, "hello") > (pair (int, string)) { return(pair) }((int, string))

// nested tuple types
((1, 2), "outer") > (nested ((int, int), string)) { return(nested) }(((int, int), string))

// generic types in tuples
(1, "hello") > (pair (int, t)) { return(pair) }((int, t))
```

when a tuple is linked to a function with a single parameter of tuple type, the tuple is passed as-is rather than unpacked.

identifier qualified by a module

```ebnf
q_ident = module_name "." ident .
module_name     = ident .
```

data structure types

```ebnf
list_type = "list" "[" Type "]"
map_type =  "map" "{" "string" "}" Type .
```

**note:** maps only support `string` keys. the key type must always be `string`. users can convert other types to strings using `to_string()` when needed.

**note:** bark does not have user-defined struct or record types. maps serve all data modeling needs, providing direct compatibility with JSON/API data. use validation functions to enforce expected map shapes.

**note:** bark does not have type aliases. use descriptive parameter names and comments to document intent (e.g., `user_id string` instead of `type UserId = string`).

**note:** bark uses a single generic type parameter `t` for "any type". there are no named type parameters (no `fn[K, V]` syntax). the generic `t` is sufficient for all use cases in bark's simple type system.

function type

``` ebnf
fn_type = "fn" "(" [ type_list ] ")" "(" [ type_list ] ")" .
```

function types specify the signature of functions passed as parameters:

```bark
// function type: fn(param_types)(return_types)
fn apply(x int, f fn(int)(int)) {
    x > f() > return()
}(int)

// pass named function
fn double(n int) { mul(n, 2) > return() }(int)
(5, double) > apply() > result  // result = 10

// pass anonymous function
(5, (n int) { mul(n, 2) > return() }(int)) > apply() > result

// function with multiple parameters
fn combine(a int, b int, op fn(int, int)(int)) {
    (a, b) > op() > return()
}(int)

// function with tuple return
fn apply_with_pair(x int, f fn(int)((int, int))) {
    x > f() > return()
}((int, int))

// generic function type using t
fn identity_apply(x t, f fn(t)(t)) {
    x > f() > return()
}(t)
```

function type validation checks:

- parameter count must match
- each parameter type must match
- return type count must match
- each return type must match

## declarations

### function declaration

```ebnf
fn_decl = [ "pub" ] "fn" fn_name params [ fn_body ] fn_return .
fn_name = ident ["?"] .
fn_body = block .
fn_return = "(" [ param_list | type_list ] ")" .
```

**return type matching:**

all execution paths in a function must return values matching the declared return type signature:

- **explicit return required** - all functions with return types must use `return()` to return values
- **early exit** - `return(value)` exits the function and returns the specified value
- **conditional return** - `return?(condition, values...)` exits if condition is truthy
- **all paths must match** - every return path must match the signature

functions without return types complete when execution reaches the end of the function body.

see [docs/implementation/return_semantics.md](implementation/return_semantics.md) for complete specification

### anonymous function

anonymous functions can be used inline, typically for error handlers, collection iteration, recursion, and higher-order functions

```ebnf
anon_fn = "(" param_list ")" fn_body "(" return_types ")" .
anon_fn = "(" param_list ")" ">" expr .
```

**key properties:**

- **explicit parameter types** - parameter types must be specified
- **explicit return types** - return types must be declared (like named functions)
- **storable** - can be stored in variables with type `fn`
- **passable** - can be passed as arguments to functions
- **returnable** - can be returned from functions
- **closures** - capture variables from enclosing scope (lexical scoping)
- **recursion** - can call themselves using `repeat?()` and `repeat()`

**recursion with repeat?() and repeat():**

anonymous functions can recursively call themselves using `repeat?()` (conditional) or `repeat()` (unconditional):

```bark
// Factorial using recursion
n > (num int, acc int) {
  num > lte?(1) > should_return
  should_return > return?(acc)

  num > sub(1) > next_num
  acc > mul(num) > next_acc
  repeat(next_num, next_acc)  // Unconditional recursive call
}(int) > result
```

- `repeat?(condition, params...)` - recursively calls function if condition is true
- `repeat(params...)` - unconditionally recursively calls function
- parameters must match the function signature
- primitives require explicit parameters (pass-by-value)
- collections can omit parameters (pass-by-reference)

see [docs/implementation/anonymous_functions.md](implementation/anonymous_functions.md) for complete specification

### memoized function declaration

memoized functions automatically cache their results based on argument values. once a function is called with specific arguments, subsequent calls with the same arguments return the cached result without re-executing the function body.

```ebnf
mfn_decl = "mfn" fn_name params [ fn_body ] fn_return .
```

**key properties:**

- **automatic caching** - results are cached based on serialized argument values
- **recursive optimization** - recursive calls benefit from memoization automatically
- **same semantics as fn** - follows all the same rules as regular functions
- **cache persistence** - cache persists for the lifetime of the function object

**example - fibonacci with memoization:**

```bark
mfn fib(n int) {
    n > lte?(1) > return?(n)
    n > sub(1) > fib() > a
    n > sub(2) > fib() > b
    a > add(b) > return()
}(int)

50 > fib() > println()  // Fast: each fib(k) computed only once
```

without memoization, `fib(50)` would require billions of recursive calls. with `mfn`, each unique argument is computed only once, making it linear time.

**cache key generation:**

arguments are serialized using their type and string representation:

- `fib(5)` → cache key `"INTEGER:5"`
- `add_mul(3, 4)` → cache key `"INTEGER:3|INTEGER:4"`

**when to use mfn:**

- recursive algorithms (fibonacci, ackermann, etc.)
- pure functions with expensive computations
- functions called repeatedly with the same arguments

**limitations:**

- cache grows unbounded (no automatic eviction)
- only caches successful results (errors are not cached)
- cache is per-function-instance (not global)

### module declaration

modules provide code organization and namespacing. each file can declare its module membership.

```ebnf
module_decl = "module" module_name .
```

files without a module declaration belong to the default `_` module namespace.

module declarations must appear at the top of the file before any other code.

see [docs/implementation/module_system.md](implementation/module_system.md) for complete specification

### import declaration

import brings external modules into the current namespace with qualified access.

```ebnf
import_decl = "import" module_path [ "as" alias ] .
module_path = string_lit .
alias = ident .
```

examples:

```bark
import "utils/strings"
import "utils/math" as m
import "https://modules.bark-lang.org/json@v1.2.3"
```

imported modules are accessed via qualified names: `module_name.function_name()`

### include declaration

include merges another module's code into the current module's namespace.

```ebnf
include_decl = "include" module_path .
```

examples:

```bark
module db
include "db/connection"

pub fn query_users() {
  connect("localhost") > conn  // connect from included module
}()
```

rules:

- can only include modules with same module declaration
- included private functions remain private
- name collisions are compile errors
- circular includes are compile errors

### builtin modules

builtin modules are always available and provide specialized functionality via namespaced functions. no import statement is required.

**available builtin modules:**

```bark
// HTTP operations (both styles work)
http.get("https://api.example.com/users") > (err, response)
"https://api.example.com/users" > http.get() > (err, response)  // via link operator

// File system operations
file.read("/path/to/file.txt") > (err, content)
file.write("/path/to/file.txt", "content") > err
file.exists?("/path/to/file.txt") > bool

// Environment variables
env.get("PORT") > value
env.get_or("PORT", "8080") > value
env.present?("DEBUG") > bool
env.absent?("API_KEY") > bool

// JSON operations
json.parse(json_string) > (err, data)
json.stringify(data) > string
json.stringify_pretty(data) > string

// Time operations
time.now() > timestamp
time.now_ms() > milliseconds
time.parse_iso8601(timestamp_string) > (err, timestamp)

// Base64 encoding
base64.encode("hello") > encoded
base64.decode(encoded) > decoded

// URL encoding
url.encode("hello world") > encoded
url.decode(encoded) > decoded

// Regular expressions
regex.match?("test@example.com", "@") > bool
regex.replace("hello world", "world", "bark") > "hello bark"

// Directory operations
dir.exists?("/path/to/dir") > bool
dir.list("/path/to/dir") > (err, files)
```

**complete example:**

```bark
// Fetch API data and save to file
fn fetch_and_save(url string, path string) {
  // Make HTTP request
  http.get(url) > (e, response)
  e > return?(e)

  // Check status
  response > get("status") > status
  status > ne?(200) > return?(err("HTTP error"))

  // Parse JSON response
  response > get("body") > json.parse() > (e2, data)
  e2 > return?(e2)

  // Convert to pretty JSON and save
  data > json.stringify_pretty() > file.write(path) > e3
  return(e3)
}(error)
```

see [builtins.md](builtins.md) for complete builtin module reference

## expressions

### operator precedence

bark has only one operator: the link operator `>`. all other operations (arithmetic, comparison, logic) are performed via builtin functions (see [builtins.md](builtins.md)).

precedence levels from highest (binds tightest) to lowest (binds loosest):

| precedence | operator | description | associativity |
| ---------- | -------- | ----------- | ------------- |
| 1 | `()` | function call, grouping | n/a |
| 2 | `.` | member access (modules) | left |
| 3 | `>` | link operator | left |
| 4 | `,` | comma (separator in lists, parameters) | left |
| 5 | `:` | colon (map literals only) | n/a |

notes:

- **function calls** (precedence 1) bind tightest: `a > b()` evaluates `b()` first
- **member access** (precedence 2): `module.function()` accesses function in module namespace
- **link operator** (precedence 3): left-associative, `a > b > c` evaluates as `(a > b) > c`
- **comma** (precedence 4): separates items in lists, parameters, and anonymous function processes
- **colon** (precedence 5): used only in map literals `{"key": "value"}`, not an operator

**important**: bark does not have infix arithmetic or comparison operators. use builtin functions instead:

```bark
// arithmetic via builtins
5 > add(3) > mul(2)   // (5 + 3) * 2 = 16

// comparison via builtins
x > eq?(5) > println()   // check if x == 5, output result

// all operations are function calls
value > math.sqrt() > math.round(0) > println()
```

examples:

```bark
// link operator is left-associative
"hello" > str.upper() > println()  // ("hello" > str.upper()) > println()

// member access binds tighter than link
value > math.sqrt()   // value > (math.sqrt())

// function call binds tightest
a > b() > c()        // a > (b()) > (c())
```

### member access expressions

the period `.` operator accesses members (functions, values) from modules

```ebnf
member_access = expr "." ident [ "(" [ arg_list ] ")" ] .
```

example:

```bark
fn user_by_id(conn module, id string) {
  conn.execute("select * from users where id = ?", id) >
    sql_to_user() > user
  return(user)
}(map)
```

### link expressions

the link operator `>` passes the result of the left expression as input to the right expression

```ebnf
link_expr = expr ">" expr .
```

**semantics:**

when the right expression is a function call, the left expression's result value(s) become the first parameter(s) of the function:

```bark
42 > add(10)              // becomes: add(42, 10)
"hello" > add(" world")   // becomes: add("hello", " world")
age > gte?(18) > is_adult // gte?(age, 18) returns bool, which flows to is_adult
```

this applies to all functions, including boolean functions chaining to conditional control flow:

```bark
eq?("a","b") > return?("match")
// eq?("a","b") returns bool
// bool flows as first parameter: return?(bool, "match")
```

see [docs/implementation/boolean_chaining.md](implementation/boolean_chaining.md) for detailed specification

### variable binding

variables are created by linking a value to an identifier

```ebnf
variable_binding = expr ">" ident .
```

the link operator `>` creates a variable binding when the right-hand side is a bare identifier (not followed by parentheses)

example:

```bark
fn example() {
  42 > number           // number is bound to 42
  "hello" > greeting    // greeting is bound to "hello"
  number > add(10) > sum  // sum is bound to 52
}
```

**distinguishing binding from function calls:**

- `value > identifier` → variable binding
- `value > identifier()` → function call (no arguments)
- `value > identifier(args)` → function call (with arguments)

**immutability:**

variables are immutable once bound. attempting to rebind a variable in the same scope is an error

```bark
fn example() {
  10 > x
  20 > x  // ERROR: x is already bound
}
```

**scope:**

variables are scoped to the function body in which they are declared. there is no block-level scoping

**type inference:**

variable types are inferred from the expression they are bound to

**parameter passing:**

bark uses a hybrid parameter passing model for performance and safety:

- **pass-by-value**: primitives (int, uint, float, bool) and error
- **pass-by-reference**: string (immutable), map, array (mutable), fn

see [docs/implementation/memory_model.md](implementation/memory_model.md) for complete parameter passing and memory semantics

see [docs/implementation/variables.md](implementation/variables.md) for variable binding semantics

### tuple destructuring

tuple destructuring splits a tuple-returning expression into named variables

```ebnf
destructure_expr = expr ">" "(" ident_list ")" .
```

example:

```bark
// Function returns (error, string)
name > validateStr() > (err, result)

// err and result are now variables in scope
err > logError()
result > transform()
```

the number of identifiers must match the number of values returned by the expression. destructured variables are scoped to the containing function body

**underscore for ignored values:**

use underscore `_` to ignore values in a tuple

```bark
validate(data) > (e, _)  // Ignore the result, only care about error
e > return?(e)
```

see [docs/implementation/variables.md](implementation/variables.md) for complete destructuring semantics

### tuple expressions (ephemeral tuples)

tuple expressions provide multiple values to functions via the link operator. they are ephemeral - they exist only to pass values and cannot be stored in variables or returned from functions.

```ebnf
tuple_expr = "(" expr "," expr { "," expr } ")" .
```

**syntax distinction:**

- `(expr, expr)` - tuple expression (multiple expressions separated by commas)
- `(expr)` - grouped expression (single expression)
- `(ident type)` - anonymous function parameter list (identifiers with types)

**usage:**

tuples unpack their elements as arguments when used with the link operator:

```bark
// Tuple unpacks to function parameters
(5, 1) > (a int, b int) {
  a > add(b) > return()
}(int)  // Returns 6

// Works with named functions too
fn calculate(x int, y int, z int) {
  x > add(y) > add(z) > return()
}(int)

(1, 2) > calculate(3)  // Unpacks to calculate(1, 2, 3), returns 6
```

**essential for recursive anonymous functions:**

tuples solve the problem of initializing anonymous functions that need multiple parameters:

```bark
fn factorial(n int) {
  // Tuple (n, 1) initializes num=n, acc=1
  (n, 1) > (num int, acc int) {
    num > lte?(1) > should_return
    should_return > return?(acc)

    num > sub(1) > next_num
    acc > mul(num) > next_acc
    repeat?(next_num, next_acc)
  }(int) > result

  return(result)
}(int)
```

**constraints:**

- tuples cannot be assigned to variables: `(1, 2) > x` is an error
- tuples cannot be returned from functions
- the number of tuple elements must match the function's parameter count (for anonymous functions) or be prepended to existing arguments (for function calls)

## error handling

### error type

the `error` type is a built-in opaque type representing errors. errors flow through chains and are caught by error handler functions.

### error handler functions

error handlers are identified by having `error` as the first parameter type. when an error flows through a chain, it is automatically routed to the next error handler function in the chain.

```bark
fn handle_validation_error(err error){
  eprintln("validation failed: {0}", err)
  // return default value or propagate error
}

fn handle_db_error(err error){
  eprintln("database error: {0}", err)
  // handle database-specific errors
}
```

### error propagation in chains

errors propagate through chains and are caught by error handler functions. if a function returns an error, non-error-handler functions are skipped until an error handler is encountered.

```bark
// Basic chain without error handling
data > validate() > save()

// Chain with error handler injection
// if validate() returns error, handle_validation_error() receives it
// if no error, handle_validation_error() is skipped, and save() executes
data > validate() > handle_validation_error() > save()

// Multiple error handlers in a chain
// each handler catches errors from functions before it
data > validate() > handle_validation_error() >
       save() > handle_db_error() >
       notify()
```

### error handler behavior

- error handlers only execute when an error is present in the chain
- if no error exists, error handlers are skipped
- non-error functions are skipped when an error is present, until an error handler processes it
- error handlers can return a value to continue the chain, or propagate the error

### built-in function reference

this section provides formal signatures for all built-in functions. for usage examples, see [docs/builtins.md](builtins.md)

#### error handling functions

- `capture(errVar ident, resultVar ident)` - extracts error and result from `(error, value)` tuple; binds to variables; continues chain with result on success, stops chain on error
- `err(message string)(error)` - creates and returns an error value
- `err(message string, context map)(error)` - creates an error with context
- `err_msg(e error)(string)` - extracts message from error
- `err_context(e error)(map)` - extracts context map from error
- `err_add_context(e error, key string, value t)(error)` - adds context to error and returns updated error

#### error checking functions

- `absent?(err error)(bool)` - returns true if error is absent (empty map `{}`), false otherwise
- `present?(err error)(bool)` - returns true if error is present (non-empty map), false otherwise

#### input/output functions

All print functions support format strings with indexed placeholders when given multiple arguments:

- `print(args...)` - prints to stdout without newline; single arg prints directly, multiple args use first as format string with `{0}`, `{1}` placeholders
- `println(args...)` - prints to stdout with newline; same format support as print
- `eprint(args...)` - prints to stderr without newline; same format support as print
- `eprintln(args...)` - prints to stderr with newline; same format support as print

**Conditional print functions** - only print when the first argument (boolean) is true:

- `print?(cond bool, args...)` - conditional print to stdout without newline
- `println?(cond bool, args...)` - conditional print to stdout with newline
- `eprint?(cond bool, args...)` - conditional print to stderr without newline
- `eprintln?(cond bool, args...)` - conditional print to stderr with newline

**Format examples:**

```bark
println("hello")                      // hello
println("value: {0}", 42)             // value: 42
println("{0} + {1} = {2}", 1, 2, 3)   // 1 + 2 = 3
eprintln("error: {0}", err_msg)   // error: ... (to stderr)

// Conditional printing - useful in chains
err > absent?() > println?("No error occurred")
x > gt?(10) > println?("x is greater than 10")
```

#### type conversion functions

- `to_string(val t)(string)` - converts value to string representation

#### map functions

**Global namespace:**

- `get(m map, key string)(t)` - gets value at key; if key doesn't exist the chain stops, error is output to stderr, and program continues
- `set(m map, key string, val t)(map)` - returns new map with key set to val
- `first(m map)(tuple)` - returns `(key, value)` tuple for first entry, or empty tuple if empty
- `last(m map)(tuple)` - returns `(key, value)` tuple for last entry, or empty tuple if empty
- `next(m map, key string)(string)` - returns next key; empty string if at end or invalid key
- `prev(m map, key string)(string)` - returns previous key; empty string if at beginning or invalid key
- `head(m map)(map, string)` - returns `(map, first_key)` tuple ready for forward iteration, or `(map, "")` if empty
- `tail(m map)(map, string)` - returns `(map, last_key)` tuple ready for reverse iteration, or `(map, "")` if empty
- `absent?(m map)(bool)` - checks if map is empty
- `size(m map)(int)` - returns number of entries in map

**map module:**

- `map.get_or(m map, key string, default t)(t)` - gets value at key or returns default if key doesn't exist
- `map.key_present?(m map, key string)(bool)` - checks if map contains key
- `map.key_absent?(m map, key string)(bool)` - checks if map does not contain key
- `map.del(m map, ...keys string)(map)` - returns new map with keys removed
- `map.keys(m map)(array)` - returns array of all keys
- `map.values(m map)(array)` - returns array of all values
- `map.entries(m map)(array)` - returns array of [key, value] pairs
- `map.merge(m1 map, m2 map)(map)` - returns new map merging m1 and m2 (m2 values override m1)

#### array functions

**Global namespace:**

- `get(arr array, index int)(t)` - gets value at index; if index is out of bounds the chain stops, error is output to stderr, and program continues
- `set(arr array, index int, val t)(array)` - returns new array with value at index
- `len(arr array)(int)` - returns number of elements in array
- `first(arr array)(tuple)` - returns `(index, value)` tuple for first element, or empty tuple if empty
- `last(arr array)(tuple)` - returns `(index, value)` tuple for last element, or empty tuple if empty
- `next(arr array, index int)(int)` - returns next index; `-1` if at end or invalid index
- `prev(arr array, index int)(int)` - returns previous index; `-1` if at beginning or invalid index
- `head(arr array)(array, int)` - returns `(array, 0)` tuple ready for forward iteration, or `(array, -1)` if empty
- `tail(arr array)(array, int)` - returns `(array, last_index)` tuple ready for reverse iteration, or `(array, -1)` if empty
- `includes?(arr array, val t)(bool)` - checks if array contains value (polymorphic with strings)
- `excludes?(arr array, val t)(bool)` - checks if array does not contain value (polymorphic with strings)
- `reverse(arr array)(array)` - returns new array with elements in reverse order (polymorphic with strings)
- `absent?(arr array)(bool)` - checks if array is empty
- `size(arr array)(int)` - returns number of elements in array

**array module:**

- `array.push(arr array, ...vals t)(array)` - returns new array with values appended to end
- `array.pop(arr array)(array, t)` - removes and returns last element, error if empty
- `array.shift(arr array)(array, t)` - removes and returns first element, error if empty
- `array.unshift(arr array, ...vals t)(array)` - returns new array with values prepended to start
- `array.append_to(val t, arr array)(array)` - returns new array with value appended (for use with link operator)
- `array.slice(arr array, start int, end int)(array)` - returns subarray from start (inclusive) to end (exclusive)
- `array.range(start int, end int)(array)` - returns array `[start, start+1, ..., end]` (inclusive)
- `array.dedupe(arr array)(array)` - returns new array with duplicate elements removed

#### iteration functions

- `each(collection t, fn fn)(t)` - calls `fn(collection, index/key)` for each element; returns the collection for chaining

#### string functions

String manipulation functions are in the `str` module:

- `str.concat(s1 string, s2 string, ...)(string)` - concatenates two or more strings
- `str.format(format string, ...args)(string)` - formats string with indexed placeholders `{0}`, `{1}`, etc. Use `{{` and `}}` for literal braces
- `str.upper(s string)(string)` - converts string to uppercase
- `str.lower(s string)(string)` - converts string to lowercase
- `str.starts_with?(s string, prefix string)(bool)` - checks if string starts with prefix
- `str.ends_with?(s string, suffix string)(bool)` - checks if string ends with suffix
- `includes?(s string, substr string)(bool)` - checks if string contains substring (global, polymorphic)
- `excludes?(s string, substr string)(bool)` - checks if string does not contain substring (global, polymorphic)
- `str.replace(s string, old string, new string)(string)` - returns string with all occurrences of old replaced with new
- `str.split(s string, delimiter string)(array)` - splits string by delimiter, returns array of parts
- `str.join(arr array, separator string)(string)` - joins array elements into string with separator
- `str.trim(s string)(string)` - removes leading and trailing whitespace
- `str.numeric?(s string)(bool)` - returns true if string represents a valid number (integer, float, or scientific notation)
- `str.alphanumeric?(s string)(bool)` - returns true if string contains only alphanumeric characters
- `reverse(s string)(string)` - returns string with characters in reverse order (polymorphic with arrays)
- `size(s string)(int)` - returns number of characters in string

#### regex module functions

- `regex.match?(text string, pattern string)(bool)` - returns true if pattern matches anywhere in text
- `regex.find(text string, pattern string)(string)` - returns first match, or empty string if no match
- `regex.find_all(text string, pattern string)(array)` - returns array of all matches, or empty array if no matches
- `regex.replace(text string, pattern string, replacement string)(string)` - replaces all matches with replacement
- `regex.split(text string, pattern string)(array)` - splits text by pattern, returns array of parts

#### global arithmetic functions

- `add(a1 t, a2 t)(t)` - adds two numbers
- `sub(s1 t, s2 t)(t)` - subtracts s2 from s1
- `mul(m1 t, m2 t)(t)` - multiplies two numbers
- `div(d1 t, d2 t)(t)` - divides d1 by d2

#### math module functions

- `math.mod(d1 t, d2 t)(t)` - returns remainder of d1 divided by d2
- `math.abs(n t)(t)` - returns absolute value
- `math.ceil(n t)(t)` - rounds up to nearest integer
- `math.floor(n t)(t)` - rounds down to nearest integer
- `math.round(n t, decimals int)(t)` - rounds to specified decimal places
- `math.min(...nums t)(t)` - returns smallest of provided numbers
- `math.max(...nums t)(t)` - returns largest of provided numbers
- `math.sqrt(n float)(float)` - returns square root
- `math.pow(base float, exp float)(float)` - returns base raised to exp
- `math.exp(n float)(float)` - returns e raised to n
- `math.log(n float)(float)` - returns natural logarithm
- `math.log10(n float)(float)` - returns base-10 logarithm
- `math.sin(n float)(float)` - returns sine
- `math.cos(n float)(float)` - returns cosine
- `math.tan(n float)(float)` - returns tangent
- `math.asin(n float)(float)` - returns arcsine
- `math.acos(n float)(float)` - returns arccosine
- `math.atan(n float)(float)` - returns arctangent
- `math.pi()(float)` - returns pi constant
- `math.e()(float)` - returns e constant
- `math.odd?(n int)(bool)` - returns true if number is odd
- `math.even?(n int)(bool)` - returns true if number is even
- `math.to_int(n float)(int)` - converts float to int (truncates)
- `math.to_float(n int)(float)` - converts int to float

#### comparison functions

- `eq?(a t, b t)(bool)` - returns true if a equals b
- `ne?(a t, b t)(bool)` - returns true if a not equal to b
- `lt?(a t, b t)(bool)` - returns true if a less than b
- `gt?(a t, b t)(bool)` - returns true if a greater than b
- `lte?(a t, b t)(bool)` - returns true if a less than or equal to b
- `gte?(a t, b t)(bool)` - returns true if a greater than or equal to b

#### boolean functions

- `not(b bool)(bool)` - returns logical negation of boolean

#### file module functions

- `file.read(path string)(error, string)` - reads entire file into string
- `file.write(path string, content string)(error)` - writes string to file (creates or overwrites)
- `file.append(path string, content string)(error)` - appends string to end of file
- `file.exists?(path string)(bool)` - checks if regular file exists at path
- `file.absent?(path string)(bool)` - checks if file does not exist at path
- `file.delete(path string)(error)` - deletes file at path
- `file.info(path string)(error, map)` - returns file metadata map: `{"size": int, "modified": int, "is_dir": bool}`

#### json module functions

- `json.parse(json_string string)(error, t)` - parses JSON string into map or array
- `json.stringify(data t)(string)` - serializes map/array to compact JSON string
- `json.stringify_pretty(data t)(string)` - serializes map/array to pretty-printed JSON string

#### http module functions

All HTTP functions return a response map: `{"status": int, "headers": map, "body": string, "url": string}`

- `http.get(url string)(error, map)` - makes HTTP GET request
- `http.post(url string, body string)(error, map)` - makes HTTP POST request with string body
- `http.put(url string, body string)(error, map)` - makes HTTP PUT request with string body
- `http.delete(url string)(error, map)` - makes HTTP DELETE request
- `http.request(method string, url string, headers map, body string)(error, map)` - makes generic HTTP request with full control

#### base64 module functions

- `base64.encode(data string)(string)` - encodes string to base64
- `base64.decode(encoded string)(error, string)` - decodes base64 string

#### url module functions

- `url.encode(value string)(string)` - percent-encodes string for URL query parameters
- `url.decode(encoded string)(error, string)` - decodes percent-encoded URL string
- `url.parse(url_string string)(error, map)` - parses URL into components map

#### time module functions

All timestamps are Unix time (seconds since epoch) in UTC.

- `time.now()(int)` - returns current Unix timestamp in seconds
- `time.now_ms()(int)` - returns current Unix timestamp in milliseconds
- `time.format(timestamp int, format string)(string)` - formats Unix timestamp using Go time format
- `time.parse(time_string string, format string)(error, int)` - parses time string to Unix timestamp
- `time.format_iso8601(timestamp int)(string)` - formats Unix timestamp as ISO 8601 / RFC 3339 string
- `time.parse_iso8601(time_string string)(error, int)` - parses ISO 8601 / RFC 3339 time string

#### env module functions

- `env.get(key string)(string)` - gets environment variable value (returns "" if not set)
- `env.get_or(key string, default string)(string)` - gets environment variable or returns default
- `env.present?(key string)(bool)` - checks if environment variable exists
- `env.absent?(key string)(bool)` - checks if environment variable does not exist
- `env.all()(map)` - returns all environment variables as map

#### dir module functions

- `dir.list(path string)(error, array)` - returns array of filenames in directory
- `dir.exists?(path string)(bool)` - checks if directory exists at path
- `dir.absent?(path string)(bool)` - checks if directory does not exist at path

#### security module functions

- `security.sql_escape(s string)(string)` - escapes string for SQL queries
- `security.shell_escape(s string)(string)` - escapes string for shell commands
- `security.safe_command?(cmd string)(bool)` - checks if command is safe to execute
- `security.sanitize_path(path string)(error, string)` - sanitizes file path to prevent traversal
- `security.email?(s string)(bool)` - validates email format
- `security.url?(s string)(bool)` - validates URL format
- `security.html_escape(s string)(string)` - escapes HTML special characters
- `security.strip_tags(s string)(string)` - removes HTML tags from string
- `security.generate_nonce()(string)` - generates cryptographically secure nonce
- `security.hash_key(key string)(int)` - hashes a key for rate limiting

#### crypto module functions

- `crypto.bcrypt_hash(password string)(error, string)` - hashes password with bcrypt
- `crypto.bcrypt_verify(password string, hash string)(bool)` - verifies bcrypt hash
- `crypto.argon2_hash(password string)(error, string)` - hashes password with Argon2id
- `crypto.argon2_verify(password string, hash string)(bool)` - verifies Argon2id hash
- `crypto.aes_encrypt(plaintext string, key string)(error, string)` - encrypts with AES-256-GCM
- `crypto.aes_decrypt(ciphertext string, key string)(error, string)` - decrypts AES-256-GCM
- `crypto.hmac_sha256(message string, key string)(string)` - creates HMAC-SHA256
- `crypto.hmac_sha512(message string, key string)(string)` - creates HMAC-SHA512
- `crypto.hmac_verify(message string, signature string, key string)(bool)` - verifies HMAC
- `crypto.random_bytes(n int)(error, string)` - generates n random bytes (hex encoded)
- `crypto.random_string(n int)(error, string)` - generates random alphanumeric string of length n
- `crypto.sha256(data string)(string)` - computes SHA-256 hash
- `crypto.sha512(data string)(string)` - computes SHA-512 hash

#### sql module functions

> **Note:** SQL support requires building with `-tags sql`.

- `sql.open(driver string, dsn string)(error, connection)` - opens database connection (supported: "sqlite")
- `sql.close(conn connection)(error, bool)` - closes database connection
- `sql.query(conn connection, query string, params array)(error, array)` - executes SELECT query, returns array of row maps
- `sql.exec(conn connection, query string, params array)(error, int)` - executes INSERT/UPDATE/DELETE, returns rows affected
- `sql.begin(conn connection)(error, transaction)` - begins database transaction
- `sql.commit(tx transaction)(error, bool)` - commits transaction
- `sql.rollback(tx transaction)(error, bool)` - rolls back transaction

#### control flow functions

- `repeat()` - unconditionally recursively calls the current anonymous function; uses current args (pass-by-reference types only)
- `repeat(...params t)` - unconditionally recursively calls the current anonymous function with explicit parameters
- `repeat?(condition bool)` - recursively calls the current anonymous function if condition is true; exits if false; uses current args (pass-by-reference types only)
- `repeat?(condition bool, ...params t)` - recursively calls the current anonymous function with explicit parameters if condition is true; exits if false
- `return()` - unconditionally exits function, returns empty value
- `return(val t)` - unconditionally exits function with specified value
- `return?(condition bool)` - conditionally exits function if condition is true, returns empty value
- `return?(condition bool, val t)` - conditionally exits function with value if condition is true

#### parallel processing functions

- `parallel(arr array, fn fn)(array, array)` - processes array elements concurrently, returns (errors, results) arrays
- `parallel_all(fns array)(array, array)` - executes array of functions concurrently, returns (errors, results) arrays
- `parallel_limited(limit int, arr array, fn fn)(array, array)` - processes array with concurrency limit, returns (errors, results) arrays
- `parallel_race(fns array)(error, t)` - executes functions concurrently, returns first result as (error, result)
- `parallel_strict(arr array, fn fn)(error, array)` - parallel processing that fails fast on first error
- `all_absent?(errors array)(bool)` - returns true if all elements in array are absent/empty errors
- `all_present?(errors array)(bool)` - returns true if all elements in array are present/non-empty errors
- `first_error(errors array)(error)` - returns first non-empty error from array, or empty error if none

#### testing functions

> **Note:** Testing functions are only available in test mode (when running `bark test`).

- `assert(actual t, expected t)(t)` - compares actual to expected, prints error on mismatch, returns actual for chaining
- `assert_error(actual error, expected_msg string)(error)` - compares error message to expected, prints error on mismatch, returns actual for chaining

### handling error tuples with capture()

functions that may fail return `(error, result)` tuples. use `capture()` to extract both values and automatically stop the chain on error:

```bark
fn fetch_user_data(url string) {
  // capture() extracts error and result from tuple
  // On success: binds e={}, data=result, chain continues with data
  // On error: binds e=error, data=empty, chain stops
  url > http.get() > capture(e, response)

  // This line only executes if http.get() succeeded
  response > get("body") > json.parse() > capture(e2, data)

  // This line only executes if json.parse() succeeded
  return(data)
}(map)
```

**capture() behavior:**

- requires an `(error, value)` tuple as input
- binds error to first identifier, result to second identifier
- if error is absent (empty `{}`): chain continues with result value
- if error is present: chain stops, subsequent lines don't execute
- both variables remain in scope for later use

**accessing captured variables after the chain:**

```bark
fn safe_read(path string) {
  path > file.read() > capture(e, content)

  // Can check error after capture
  e > present?() > return?(err("Failed to read file"))

  // Use the result
  return(content)
}(string)
```

### handling multiple chains with tuple destructuring

when a function contains multiple chains that may return errors, use tuple destructuring with `return?()` for explicit error propagation:

```bark
fn validateUser(name string, age int) {
  // First chain - destructure tuple
  name > validateStr() > (e1, validatedName)
  e1 > return?(e1, "", 0)  // If e1 is present (truthy), return immediately

  // Second chain - only executes if e1 was absent
  age > validateInt() > (e2, validatedAge)
  e2 > return?(e2, "", 0)  // If e2 is present (truthy), return immediately

  // Return both validated values
  return({}, validatedName, validatedAge)
}(error, string, int)
```

`return?(condition, values...)` behavior:

- if `condition` is falsy (absent/empty): no operation, execution continues
- if `condition` is truthy (present/non-empty): function returns immediately with the specified values

## expr
