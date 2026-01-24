package object

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"strings"
	"sync/atomic"

	"gitlab.com/bark-lang/bark/ast"
)

// ObjectType represents the type of an object
type ObjectType string

const (
	INTEGER_OBJ           = "INTEGER"
	FLOAT_OBJ             = "FLOAT"
	BOOLEAN_OBJ           = "BOOLEAN"
	STRING_OBJ            = "STRING"
	NULL_OBJ              = "NULL"
	RETURN_OBJ            = "RETURN"
	ERROR_OBJ             = "ERROR"
	FUNCTION_OBJ          = "FUNCTION"
	MEMOIZED_FUNCTION_OBJ = "MEMOIZED_FUNCTION"
	BUILTIN_OBJ           = "BUILTIN"
	ARRAY_OBJ             = "ARRAY"
	MAP_OBJ               = "MAP"
	REPEAT_OBJ            = "REPEAT"
	TUPLE_OBJ             = "TUPLE"
	SQL_CONN_OBJ          = "SQL_CONN"
	SQL_TX_OBJ            = "SQL_TX"
)

// Object is the interface that all runtime objects must implement
type Object interface {
	Type() ObjectType
	Inspect() string
}

// Integer represents an integer value
type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

// Float represents a floating-point value
type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string  { return fmt.Sprintf("%g", f.Value) }

// Boolean represents a boolean value
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

// String represents a string value
type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

// Null represents the absence of a value
type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

// ReturnValue wraps a value for early return from functions
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

// RepeatValue signals that the current function should be called again with new arguments
type RepeatValue struct {
	Args []Object
}

func (rv *RepeatValue) Type() ObjectType { return REPEAT_OBJ }
func (rv *RepeatValue) Inspect() string  { return "repeat" }

// Error represents an error value.
// There are two kinds of errors:
// 1. Bark error values - created by err(), can be stored/passed/returned
// 2. Programming errors - created by newError(), stop execution immediately
// The IsProgrammingError flag distinguishes between them.
type Error struct {
	Msg                string
	Context            map[string]Object
	IsProgrammingError bool // true for programming errors (wrong args, type errors)
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string {
	if len(e.Context) > 0 {
		contextStr := []string{}
		for k, v := range e.Context {
			contextStr = append(contextStr, fmt.Sprintf("%s: %s", k, v.Inspect()))
		}
		return fmt.Sprintf("ERROR: %s (context: %s)", e.Msg, strings.Join(contextStr, ", "))
	}
	return fmt.Sprintf("ERROR: %s", e.Msg)
}

// ExecutionError represents a recoverable execution error
// These errors stop the current chain but don't crash the program.
// They are logged to stderr with source location info.
type ExecutionError struct {
	Message    string            // Error message
	Detail     string            // Additional detail (e.g., "index 8 is out of range for array of length 3")
	File       string            // Source file path
	Line       int               // Line number (1-indexed)
	Column     int               // Column number (1-indexed)
	SourceLine string            // The actual source code line
	Context    map[string]Object // Optional context
}

const EXEC_ERROR_OBJ = "EXEC_ERROR"

func (e *ExecutionError) Type() ObjectType { return EXEC_ERROR_OBJ }
func (e *ExecutionError) Inspect() string {
	return fmt.Sprintf("execution error: %s", e.Message)
}

// FormatError returns a formatted error message suitable for stderr output
func (e *ExecutionError) FormatError() string {
	var sb strings.Builder

	// Header: file:line:column: error: message
	if e.File != "" {
		sb.WriteString(fmt.Sprintf("%s:%d:%d: error: %s\n", e.File, e.Line, e.Column, e.Message))
	} else {
		sb.WriteString(fmt.Sprintf("<unknown>:%d:%d: error: %s\n", e.Line, e.Column, e.Message))
	}

	// Source line with pointer
	if e.SourceLine != "" {
		sb.WriteString(fmt.Sprintf("  %s\n", e.SourceLine))
		// Create pointer (spaces + carets)
		pointer := strings.Repeat(" ", e.Column+1) + "^"
		sb.WriteString(fmt.Sprintf("%s\n", pointer))
	}

	// Detail line
	if e.Detail != "" {
		sb.WriteString(fmt.Sprintf("  = %s\n", e.Detail))
	}

	return sb.String()
}

// Function represents a user-defined function
type Function struct {
	Parameters []*ast.Parameter
	Body       *ast.BlockStatement
	Env        *Environment
	ReturnType *ast.TypeList // Return type annotation for validation
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	return fmt.Sprintf("fn(%s) { ... }", strings.Join(params, ", "))
}

// MemoEntry stores a cached result with its original arguments for collision handling
type MemoEntry struct {
	Args   []Object
	Result Object
}

// MemoCache provides fast hash-based caching with collision handling
type MemoCache struct {
	entries map[uint64][]MemoEntry
}

// NewMemoCache creates a new memoization cache
func NewMemoCache() *MemoCache {
	return &MemoCache{
		entries: make(map[uint64][]MemoEntry),
	}
}

// hashArgs computes a fast hash of function arguments
func (mc *MemoCache) hashArgs(args []Object) uint64 {
	h := fnv.New64a()
	for _, arg := range args {
		// Write type discriminator
		_, _ = h.Write([]byte(arg.Type()))
		// Write value based on type
		switch v := arg.(type) {
		case *Integer:
			_ = binary.Write(h, binary.LittleEndian, v.Value)
		case *Float:
			_ = binary.Write(h, binary.LittleEndian, v.Value)
		case *Boolean:
			if v.Value {
				_, _ = h.Write([]byte{1})
			} else {
				_, _ = h.Write([]byte{0})
			}
		case *String:
			_, _ = h.Write([]byte(v.Value))
		case *Null:
			_, _ = h.Write([]byte{0})
		default:
			// Fall back to Inspect() for complex types
			_, _ = h.Write([]byte(arg.Inspect()))
		}
	}
	return h.Sum64()
}

// argsEqual checks if two argument slices are equal
func argsEqual(a, b []Object) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Type() != b[i].Type() {
			return false
		}
		switch va := a[i].(type) {
		case *Integer:
			if vb, ok := b[i].(*Integer); !ok || va.Value != vb.Value {
				return false
			}
		case *Float:
			if vb, ok := b[i].(*Float); !ok || va.Value != vb.Value {
				return false
			}
		case *Boolean:
			if vb, ok := b[i].(*Boolean); !ok || va.Value != vb.Value {
				return false
			}
		case *String:
			if vb, ok := b[i].(*String); !ok || va.Value != vb.Value {
				return false
			}
		case *Null:
			if _, ok := b[i].(*Null); !ok {
				return false
			}
		default:
			// Fall back to Inspect() comparison for complex types
			if a[i].Inspect() != b[i].Inspect() {
				return false
			}
		}
	}
	return true
}

// Get retrieves a cached result for the given arguments
func (mc *MemoCache) Get(args []Object) (Object, bool) {
	hash := mc.hashArgs(args)
	entries, ok := mc.entries[hash]
	if !ok {
		return nil, false
	}
	// Check for exact match (collision handling)
	for _, entry := range entries {
		if argsEqual(entry.Args, args) {
			return entry.Result, true
		}
	}
	return nil, false
}

// Set stores a result for the given arguments
func (mc *MemoCache) Set(args []Object, result Object) {
	hash := mc.hashArgs(args)
	// Copy args to prevent mutation issues
	argsCopy := make([]Object, len(args))
	copy(argsCopy, args)
	mc.entries[hash] = append(mc.entries[hash], MemoEntry{
		Args:   argsCopy,
		Result: result,
	})
}

// MemoizedFunction represents a memoized user-defined function
// Results are cached based on argument values for automatic memoization
type MemoizedFunction struct {
	Parameters []*ast.Parameter
	Body       *ast.BlockStatement
	Env        *Environment
	Cache      *MemoCache
	ReturnType *ast.TypeList // Return type annotation for validation
}

func (mf *MemoizedFunction) Type() ObjectType { return MEMOIZED_FUNCTION_OBJ }
func (mf *MemoizedFunction) Inspect() string {
	params := []string{}
	for _, p := range mf.Parameters {
		params = append(params, p.String())
	}
	return fmt.Sprintf("mfn(%s) { ... }", strings.Join(params, ", "))
}

// BuiltinFunction is a function implemented in Go
type BuiltinFunction func(args ...Object) Object

// Builtin represents a builtin function
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }

// Array represents an array of objects with copy-on-write semantics
type Array struct {
	Elements []Object
	refCount *int32 // Reference count for COW; nil means unshared (refCount=1)
}

func (ao *Array) Type() ObjectType { return ARRAY_OBJ }
func (ao *Array) Inspect() string {
	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

// Share marks the array as shared and returns it (for assignment to multiple vars)
func (ao *Array) Share() *Array {
	if ao.refCount == nil {
		// First share: initialize refCount to 2 (original + new reference)
		rc := int32(2)
		ao.refCount = &rc
	} else {
		atomic.AddInt32(ao.refCount, 1)
	}
	return ao
}

// Unshare decrements the reference count (called when a reference is dropped)
func (ao *Array) Unshare() {
	if ao.refCount != nil {
		atomic.AddInt32(ao.refCount, -1)
	}
}

// IsShared returns true if the array has multiple references
func (ao *Array) IsShared() bool {
	if ao.refCount == nil {
		return false
	}
	return atomic.LoadInt32(ao.refCount) > 1
}

// COWPush appends a value, copying only if shared
func (ao *Array) COWPush(val Object) *Array {
	if !ao.IsShared() {
		// Sole owner: modify in place
		ao.Elements = append(ao.Elements, val)
		return ao
	}
	// Shared: create a copy
	newElements := make([]Object, len(ao.Elements), len(ao.Elements)+1)
	copy(newElements, ao.Elements)
	newElements = append(newElements, val)
	return &Array{Elements: newElements}
}

// COWSet sets a value at index, copying only if shared
func (ao *Array) COWSet(idx int, val Object) *Array {
	if idx < 0 || idx >= len(ao.Elements) {
		return ao // Out of bounds, return unchanged
	}
	if !ao.IsShared() {
		// Sole owner: modify in place
		ao.Elements[idx] = val
		return ao
	}
	// Shared: create a copy
	newElements := make([]Object, len(ao.Elements))
	copy(newElements, ao.Elements)
	newElements[idx] = val
	return &Array{Elements: newElements}
}

// COWSlice returns a slice, sharing underlying data when possible
func (ao *Array) COWSlice(start, end int) *Array {
	if start < 0 {
		start = 0
	}
	if end > len(ao.Elements) {
		end = len(ao.Elements)
	}
	if start >= end {
		return &Array{Elements: []Object{}}
	}
	// Always copy for slice to avoid aliasing issues
	newElements := make([]Object, end-start)
	copy(newElements, ao.Elements[start:end])
	return &Array{Elements: newElements}
}

// COWPop removes and returns the last element, copying only if shared
// Returns (new_array, popped_element)
func (ao *Array) COWPop() (*Array, Object) {
	if len(ao.Elements) == 0 {
		return ao, nil
	}
	lastIdx := len(ao.Elements) - 1
	popped := ao.Elements[lastIdx]

	if !ao.IsShared() {
		// Sole owner: modify in place
		ao.Elements = ao.Elements[:lastIdx]
		return ao, popped
	}
	// Shared: create a copy
	newElements := make([]Object, lastIdx)
	copy(newElements, ao.Elements[:lastIdx])
	return &Array{Elements: newElements}, popped
}

// COWShift removes and returns the first element, copying only if shared
// Returns (new_array, shifted_element)
func (ao *Array) COWShift() (*Array, Object) {
	if len(ao.Elements) == 0 {
		return ao, nil
	}
	first := ao.Elements[0]

	if !ao.IsShared() {
		// Sole owner: modify in place (but slice means we can't avoid allocation entirely)
		ao.Elements = ao.Elements[1:]
		return ao, first
	}
	// Shared: create a copy
	newElements := make([]Object, len(ao.Elements)-1)
	copy(newElements, ao.Elements[1:])
	return &Array{Elements: newElements}, first
}

// COWUnshift prepends values, copying only if shared
func (ao *Array) COWUnshift(vals ...Object) *Array {
	if len(vals) == 0 {
		return ao
	}

	if !ao.IsShared() {
		// Sole owner: prepend in place
		ao.Elements = append(vals, ao.Elements...)
		return ao
	}
	// Shared: create a copy
	newElements := make([]Object, 0, len(ao.Elements)+len(vals))
	newElements = append(newElements, vals...)
	newElements = append(newElements, ao.Elements...)
	return &Array{Elements: newElements}
}

// Map represents a hash map with insertion order tracking and copy-on-write semantics
type Map struct {
	Pairs    map[string]Object
	Keys     []string // Maintains insertion order
	refCount *int32   // Reference count for COW; nil means unshared (refCount=1)
}

func (m *Map) Type() ObjectType { return MAP_OBJ }
func (m *Map) Inspect() string {
	pairs := []string{}
	for _, key := range m.Keys {
		value := m.Pairs[key]
		pairs = append(pairs, fmt.Sprintf("%s: %s", key, value.Inspect()))
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

// Share marks the map as shared and returns it (for assignment to multiple vars)
func (m *Map) Share() *Map {
	if m.refCount == nil {
		// First share: initialize refCount to 2 (original + new reference)
		rc := int32(2)
		m.refCount = &rc
	} else {
		atomic.AddInt32(m.refCount, 1)
	}
	return m
}

// Unshare decrements the reference count (called when a reference is dropped)
func (m *Map) Unshare() {
	if m.refCount != nil {
		atomic.AddInt32(m.refCount, -1)
	}
}

// IsShared returns true if the map has multiple references
func (m *Map) IsShared() bool {
	if m.refCount == nil {
		return false
	}
	return atomic.LoadInt32(m.refCount) > 1
}

// COWSet sets a key-value pair, copying only if shared
func (m *Map) COWSet(key string, val Object) *Map {
	if !m.IsShared() {
		// Sole owner: modify in place
		_, keyExists := m.Pairs[key]
		m.Pairs[key] = val
		if !keyExists {
			m.Keys = append(m.Keys, key)
		}
		return m
	}
	// Shared: create a copy
	newPairs := make(map[string]Object, len(m.Pairs))
	for k, v := range m.Pairs {
		newPairs[k] = v
	}
	newPairs[key] = val

	newKeys := make([]string, len(m.Keys))
	copy(newKeys, m.Keys)
	_, keyExists := m.Pairs[key]
	if !keyExists {
		newKeys = append(newKeys, key)
	}

	return &Map{Pairs: newPairs, Keys: newKeys}
}

// COWDelete removes a key, copying only if shared
func (m *Map) COWDelete(key string) *Map {
	if _, exists := m.Pairs[key]; !exists {
		return m // Key doesn't exist, return unchanged
	}

	if !m.IsShared() {
		// Sole owner: modify in place
		delete(m.Pairs, key)
		// Remove key from Keys slice
		for i, k := range m.Keys {
			if k == key {
				m.Keys = append(m.Keys[:i], m.Keys[i+1:]...)
				break
			}
		}
		return m
	}
	// Shared: create a copy without the key
	newPairs := make(map[string]Object, len(m.Pairs)-1)
	newKeys := make([]string, 0, len(m.Keys)-1)
	for _, k := range m.Keys {
		if k != key {
			newPairs[k] = m.Pairs[k]
			newKeys = append(newKeys, k)
		}
	}
	return &Map{Pairs: newPairs, Keys: newKeys}
}

// Tuple represents an ephemeral tuple of values
// Tuples can only be used to pass multiple values to anonymous functions
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

// CaptureStop signals that a capture expression encountered an error
// and the chain should stop processing. The error has already been bound
// to the error variable in the environment.
const CAPTURE_STOP_OBJ = "CAPTURE_STOP"

type CaptureStop struct {
	Error Object // The error that was captured
}

func (cs *CaptureStop) Type() ObjectType { return CAPTURE_STOP_OBJ }
func (cs *CaptureStop) Inspect() string  { return "capture_stop" }

// ChainStop signals that continue?() returned false and the chain should stop.
// Unlike CaptureStop, this is not an error condition - just early termination.
const CHAIN_STOP_OBJ = "CHAIN_STOP"

type ChainStop struct{}

func (cs *ChainStop) Type() ObjectType { return CHAIN_STOP_OBJ }
func (cs *ChainStop) Inspect() string  { return "null" }

// Environment holds variable bindings
type Environment struct {
	store map[string]Object
	outer *Environment
}

// NewEnvironment creates a new environment
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

// NewEnclosedEnvironment creates a new environment with an outer environment
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get retrieves a value from the environment
// Uses iterative traversal instead of recursion to reduce call overhead
func (e *Environment) Get(name string) (Object, bool) {
	for env := e; env != nil; env = env.outer {
		if obj, ok := env.store[name]; ok {
			return obj, true
		}
	}
	return nil, false
}

// Set stores a value in the environment
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

// SQLConnection represents a database connection
type SQLConnection struct {
	DB     *sql.DB
	Driver string // "sqlite" or "postgres"
	DSN    string // Data source name
}

func (sc *SQLConnection) Type() ObjectType { return SQL_CONN_OBJ }
func (sc *SQLConnection) Inspect() string {
	return fmt.Sprintf("<sql.connection driver=%s>", sc.Driver)
}

// Close closes the database connection
func (sc *SQLConnection) Close() error {
	if sc.DB != nil {
		return sc.DB.Close()
	}
	return nil
}

// SQLTransaction represents an active database transaction
type SQLTransaction struct {
	Tx     *sql.Tx
	Driver string // Inherited from connection
}

func (st *SQLTransaction) Type() ObjectType { return SQL_TX_OBJ }
func (st *SQLTransaction) Inspect() string {
	return fmt.Sprintf("<sql.transaction driver=%s>", st.Driver)
}
