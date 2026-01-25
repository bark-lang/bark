# Bark Built-in Functions

This document provides user-friendly documentation for all built-in functions in Bark. For formal syntax specifications, see [spec.md](spec.md).

## Module Organization

Bark organizes built-in functions into two namespaces:

### Global Namespace (58 functions)

Fundamental operations that are always available without a module prefix:

- **Numbers (4)** - `add()`, `sub()`, `mul()`, `div()`
- **Comparison (9)** - `eq?()`, `ne?()`, `gt?()`, `gte?()`, `lt?()`, `lte?()`, `not()`, `present?()`, `absent?()`
- **Arrays (2)** - `len()`, `empty?()`
- **Data Structures - polymorphic (12)** - `get()`, `set()`, `first()`, `last()`, `next()`, `prev()`, `head()`, `tail()`, `includes?()`, `excludes?()`, `reverse()`, `size()`
- **Iteration (1)** - `each()`
- **Control Flow (5)** - `return?()`, `repeat?()`, `repeat()`, `continue?()`, `break?()`
- **Error Handling (5)** - `capture()`, `err()`, `err_msg()`, `err_context()`, `err_add_context()`
- **Parallel Processing (8)** - `parallel()`, `parallel_all()`, `parallel_limited()`, `parallel_race()`, `parallel_strict()`, `all_absent?()`, `all_present?()`, `first_error()`
- **Testing (2)** - `assert()`, `assert_error()`
- **Core (10)** - `print()`, `println()`, `eprint()`, `eprintln()`, `print?()`, `println?()`, `eprint?()`, `eprintln?()`, `return()`, `to_string()`

### Module Namespace (121 functions)

Specialized operations accessed via `module.function()` syntax:

- **http module (5)** - `http.get()`, `http.post()`, `http.put()`, `http.delete()`, `http.request()`
- **file module (7)** - `file.read()`, `file.write()`, `file.append()`, `file.delete()`, `file.exists?()`, `file.absent?()`, `file.info()`
- **env module (5)** - `env.get()`, `env.get_or()`, `env.present?()`, `env.absent?()`, `env.all()`
- **json module (3)** - `json.parse()`, `json.stringify()`, `json.stringify_pretty()`
- **time module (6)** - `time.now()`, `time.now_ms()`, `time.parse()`, `time.parse_iso8601()`, `time.format()`, `time.format_iso8601()`
- **base64 module (2)** - `base64.encode()`, `base64.decode()`
- **url module (3)** - `url.encode()`, `url.decode()`, `url.parse()`
- **regex module (5)** - `regex.match?()`, `regex.find()`, `regex.find_all()`, `regex.replace()`, `regex.split()`
- **dir module (3)** - `dir.list()`, `dir.exists?()`, `dir.absent?()`
- **str module (12)** - `str.upper()`, `str.lower()`, `str.trim()`, `str.replace()`, `str.split()`, `str.join()`, `str.concat()`, `str.format()`, `str.starts_with?()`, `str.ends_with?()`, `str.numeric?()`, `str.alphanumeric?()`
- **math module (24)** - `math.mod()`, `math.abs()`, `math.ceil()`, `math.floor()`, `math.round()`, `math.min()`, `math.max()`, `math.sqrt()`, `math.pow()`, `math.exp()`, `math.log()`, `math.log10()`, `math.sin()`, `math.cos()`, `math.tan()`, `math.asin()`, `math.acos()`, `math.atan()`, `math.pi()`, `math.e()`, `math.odd?()`, `math.even?()`, `math.to_int()`, `math.to_float()`
- **security module (10)** - `security.sql_escape()`, `security.shell_escape()`, `security.safe_command?()`, `security.sanitize_path()`, `security.email?()`, `security.url?()`, `security.html_escape()`, `security.strip_tags()`, `security.generate_nonce()`, `security.hash_key()`
- **crypto module (13)** - `crypto.bcrypt_hash()`, `crypto.bcrypt_verify()`, `crypto.argon2_hash()`, `crypto.argon2_verify()`, `crypto.aes_encrypt()`, `crypto.aes_decrypt()`, `crypto.hmac_sha256()`, `crypto.hmac_sha512()`, `crypto.hmac_verify()`, `crypto.random_bytes()`, `crypto.random_string()`, `crypto.sha256()`, `crypto.sha512()`
- **array module (8)** - `array.push()`, `array.append_to()`, `array.pop()`, `array.shift()`, `array.unshift()`, `array.slice()`, `array.range()`, `array.dedupe()`
- **map module (8)** - `map.get_or()`, `map.del()`, `map.keys()`, `map.values()`, `map.entries()`, `map.key_present?()`, `map.key_absent?()`, `map.merge()`
- **sql module (7)** - `sql.open()`, `sql.close()`, `sql.query()`, `sql.exec()`, `sql.begin()`, `sql.commit()`, `sql.rollback()`

**Total: 179 built-in functions**

## Quick Reference

### Most Commonly Used Functions

**Error Handling (Global):**

- `tuple > capture(e, r)` - Extract error and result from `(error, value)` tuple
- `err(message)` - Create an error
- `err > present?()` - Check if error occurred
- `err > absent?()` - Check if operation succeeded

**Collections (Global + Modules):**

- `map > get(key)` - Get map value (global)
- `map > set(key, value)` - Set map value (global)
- `array > get(index)` - Get array element (global)
- `array > array.push(value)` - Add to array (array module)
- `array > reverse()` - Reverse array or string (global)
- `map > map.keys()` - Get all keys (map module)
- `map > map.get_or(key, default)` - Get with default (map module)

**String Operations (str module):**

- `str > str.concat(other)` - Concatenate strings
- `str > str.format(args...)` - Format string with `{0}`, `{1}` placeholders
- `str > str.upper()` - Convert to uppercase
- `str > str.trim()` - Remove whitespace
- `str > str.split(",")` - Split into array
- `str > str.numeric?()` - Check if string is numeric

**Control Flow (Global):**

- `condition > return?(value)` - Conditional early return
- `condition > return?()` - Conditional loop exit
- `condition > continue?()` - Conditional chain continuation

**I/O (global):**

- `message > print()` - Print to stdout without newline
- `message > println()` - Print to stdout with newline
- `message > eprint()` - Print to stderr without newline
- `message > eprintln()` - Print to stderr with newline
- All print functions support format strings: `println("value: {0}", 42)` outputs `value: 42`
- Conditional variants: `print?`, `println?`, `eprint?`, `eprintln?` - only print when first arg is true
  - Example: `err > absent?() > println?("No error occurred")`

**File I/O (file module):**

- `path > file.read() > (err, content)` - Read file
- `path > file.write(content) > err` - Write file

**HTTP (http module):**

- `url > http.get() > (err, response)` - HTTP GET request
- `url > http.post(body) > (err, response)` - HTTP POST request

**SQL Database (sql module):**

- `sql.open("sqlite", ":memory:") > capture(err, conn)` - Open database connection
- `conn > sql.query("SELECT * FROM users WHERE id = ?", [1]) > capture(err, rows)` - Query database
- `conn > sql.exec("INSERT INTO users (name) VALUES (?)", ["Alice"]) > capture(err, affected)` - Execute statement
- `sql.begin(conn) > capture(err, tx)` - Start transaction
- `tx > sql.commit() > capture(err, ok)` - Commit transaction
- `tx > sql.rollback() > capture(err, ok)` - Rollback transaction
- `conn > sql.close() > capture(err, ok)` - Close connection

**JSON (json module):**

- `string > json.parse() > (err, data)` - Parse JSON string
- `data > json.stringify() > string` - Convert to JSON string

**Environment (env module):**

- `env.get("VAR_NAME") > value` - Get environment variable
- `env.get_or("VAR_NAME", "default") > value` - Get with default
- `env.present?("VAR_NAME") > bool` - Check if variable exists
- `env.absent?("VAR_NAME") > bool` - Check if variable missing

**Math (math module):**

- `n > math.mod(divisor)` - Get remainder
- `n > math.abs()` - Absolute value
- `n > math.sqrt()` - Square root
- `n > math.pow(exp)` - Raise to power
- `n > math.odd?()` - Check if odd
- `n > math.even?()` - Check if even

**Iteration (Global + array module):**

- `array > each(fn)` - Call fn(array, index) for each element
- `map > each(fn)` - Call fn(map, key) for each key
- `array.range(1, 10) > each(fn)` - Iterate over range

**Parallel Processing (Global):**

- `array > parallel(fn)` - Process elements concurrently
- `parallel_limited(4, array, fn)` - Process with concurrency limit
- `errors > all_absent?()` - Check if all operations succeeded

**Testing (Global):**

- `value > assert(expected)` - Assert value equals expected (test files only)
- `error > assert_error(expected_msg)` - Assert error has expected message (test files only)

## Testing

Bark provides built-in assertion functions for testing. Test files live in the `tests/` directory and are run with `bark test`.

**Important:** `assert()` and `assert_error()` can only be used in test files run via `bark test`. Using them in regular Bark programs will produce an error.

### assert(expected)

Compares the piped value to an expected value. On failure, prints an assertion error to stderr.

```bark
// Basic assertions
1 > assert(1)                    // passes silently
"hello" > assert("hello")        // passes silently
true > assert(true)              // passes silently

// Computed value assertions
1 > add(1) > assert(2)           // passes: 1 + 1 = 2
10 > sub(3) > assert(7)          // passes: 10 - 3 = 7

// Failed assertions print to stderr
1 > assert(2)                    // ASSERTION FAILED: expected 2, got 1
```

**Return value:** Returns the input value (allows chaining)

### assert_error(expected_msg)

Asserts that the piped value is an error with the expected message.

```bark
// Assert error message matches
err("invalid input") > assert_error("invalid input")   // passes

// Assert error from operation
parse_int("abc") > assert_error("invalid integer")     // passes if parse_int returns that error

// Failed assertions print to stderr
err("wrong") > assert_error("expected")
// ASSERTION FAILED: expected error msg "expected", got "wrong"

42 > assert_error("some error")
// ASSERTION FAILED: expected error with msg "some error", got INTEGER: 42
```

**Return value:** Returns the input value (allows chaining)

### Running Tests

```bash
# Run all tests in tests/ directory
bark test

# Run a specific test file
bark test tests/my_test.bark

# Run tests in a subdirectory
bark test tests/unit/
```

Test files must be in the `tests/` directory. The test runner:

- Runs each `.bark` file
- Reports PASS/FAIL for each file
- Exits with code 1 if any tests fail
- Detects assertion failures from stderr output

## Error Tuple Pattern with capture()

Many functions that can fail return `(error, result)` tuples. Use `capture()` to extract and handle both values:

```bark
// Basic usage - capture extracts error and result
url > http.get() > capture(e, response)
// e = {} on success, or error map on failure
// response = the HTTP response on success, or empty on failure

// On success: chain continues with response
// On failure: chain stops, but both variables are still bound

// Check the error if needed
e > present?() > return?(err("Request failed"))

// Use the result
response > get("body") > println()
```

**Key behaviors:**

1. `capture()` requires an `(error, result)` tuple as input
2. Both variables are bound regardless of success/failure
3. On success (error is `{}`): chain continues with the result value
4. On failure (error is present): chain stops at that line
5. Variables remain in scope for later use

**Example with chained operations:**

```bark
fn fetch_and_parse(url string) {
  // First capture - HTTP request
  url > http.get() > capture(e1, response)
  // If http.get() fails, execution stops here

  // Second capture - JSON parsing
  response > get("body") > json.parse() > capture(e2, data)
  // Only executes if http.get() succeeded

  // Return the parsed data
  return(data)
}(map)
```

## Notes

- All functions follow Bark's left-to-right chaining convention
- Functions ending with `?` return boolean values
- Error-producing functions return `(error, result_type)` tuples
- Use `capture()` to extract error and result from tuples
- Use `{}` (empty map) to represent "no error" / success
- Use `[]` (empty array) to represent empty collections
- Use `""` (empty string) to represent absent string values
