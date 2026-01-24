package object

import (
	"testing"
)

// =============================================================================
// Array COW Correctness Tests
// =============================================================================

func TestArrayCOWPushUnshared(t *testing.T) {
	// Create an unshared array
	arr := &Array{Elements: []Object{
		&Integer{Value: 1},
		&Integer{Value: 2},
	}}

	// Push should modify in place since array is unshared
	result := arr.COWPush(&Integer{Value: 3})

	// Result should be the same pointer (in-place modification)
	if result != arr {
		t.Error("COWPush on unshared array should return same pointer")
	}

	// Should have 3 elements
	if len(result.Elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(result.Elements))
	}
}

func TestArrayCOWPushShared(t *testing.T) {
	// Create a shared array
	arr := &Array{Elements: []Object{
		&Integer{Value: 1},
		&Integer{Value: 2},
	}}
	arr.Share() // Mark as shared

	originalLen := len(arr.Elements)

	// Push should create a new array since array is shared
	result := arr.COWPush(&Integer{Value: 3})

	// Result should be a different pointer (copy was made)
	if result == arr {
		t.Error("COWPush on shared array should return new pointer")
	}

	// Original should be unchanged
	if len(arr.Elements) != originalLen {
		t.Errorf("Original array was modified, expected %d elements, got %d", originalLen, len(arr.Elements))
	}

	// Result should have 3 elements
	if len(result.Elements) != 3 {
		t.Errorf("Expected 3 elements in result, got %d", len(result.Elements))
	}
}

func TestArrayCOWSetUnshared(t *testing.T) {
	arr := &Array{Elements: []Object{
		&Integer{Value: 1},
		&Integer{Value: 2},
		&Integer{Value: 3},
	}}

	result := arr.COWSet(1, &Integer{Value: 99})

	if result != arr {
		t.Error("COWSet on unshared array should return same pointer")
	}

	val := result.Elements[1].(*Integer)
	if val.Value != 99 {
		t.Errorf("Expected 99, got %d", val.Value)
	}
}

func TestArrayCOWSetShared(t *testing.T) {
	arr := &Array{Elements: []Object{
		&Integer{Value: 1},
		&Integer{Value: 2},
		&Integer{Value: 3},
	}}
	arr.Share()

	result := arr.COWSet(1, &Integer{Value: 99})

	if result == arr {
		t.Error("COWSet on shared array should return new pointer")
	}

	// Original unchanged
	origVal := arr.Elements[1].(*Integer)
	if origVal.Value != 2 {
		t.Errorf("Original was modified, expected 2, got %d", origVal.Value)
	}

	// Result has new value
	newVal := result.Elements[1].(*Integer)
	if newVal.Value != 99 {
		t.Errorf("Expected 99, got %d", newVal.Value)
	}
}

func TestArrayCOWPopUnshared(t *testing.T) {
	arr := &Array{Elements: []Object{
		&Integer{Value: 1},
		&Integer{Value: 2},
		&Integer{Value: 3},
	}}

	newArr, popped := arr.COWPop()

	if newArr != arr {
		t.Error("COWPop on unshared array should return same pointer")
	}

	if len(newArr.Elements) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(newArr.Elements))
	}

	poppedVal := popped.(*Integer)
	if poppedVal.Value != 3 {
		t.Errorf("Expected popped value 3, got %d", poppedVal.Value)
	}
}

func TestArrayCOWPopShared(t *testing.T) {
	arr := &Array{Elements: []Object{
		&Integer{Value: 1},
		&Integer{Value: 2},
		&Integer{Value: 3},
	}}
	arr.Share()

	newArr, popped := arr.COWPop()

	if newArr == arr {
		t.Error("COWPop on shared array should return new pointer")
	}

	// Original unchanged
	if len(arr.Elements) != 3 {
		t.Errorf("Original was modified, expected 3 elements, got %d", len(arr.Elements))
	}

	if len(newArr.Elements) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(newArr.Elements))
	}

	poppedVal := popped.(*Integer)
	if poppedVal.Value != 3 {
		t.Errorf("Expected popped value 3, got %d", poppedVal.Value)
	}
}

func TestArrayCOWShiftUnshared(t *testing.T) {
	arr := &Array{Elements: []Object{
		&Integer{Value: 1},
		&Integer{Value: 2},
		&Integer{Value: 3},
	}}

	newArr, shifted := arr.COWShift()

	// For shift, even unshared arrays get a new slice header
	// but we check the result is correct
	if len(newArr.Elements) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(newArr.Elements))
	}

	shiftedVal := shifted.(*Integer)
	if shiftedVal.Value != 1 {
		t.Errorf("Expected shifted value 1, got %d", shiftedVal.Value)
	}

	// First element should now be 2
	firstVal := newArr.Elements[0].(*Integer)
	if firstVal.Value != 2 {
		t.Errorf("Expected first element 2, got %d", firstVal.Value)
	}
}

func TestArrayCOWUnshiftUnshared(t *testing.T) {
	arr := &Array{Elements: []Object{
		&Integer{Value: 2},
		&Integer{Value: 3},
	}}

	result := arr.COWUnshift(&Integer{Value: 1})

	if len(result.Elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(result.Elements))
	}

	firstVal := result.Elements[0].(*Integer)
	if firstVal.Value != 1 {
		t.Errorf("Expected first element 1, got %d", firstVal.Value)
	}
}

func TestArrayCOWUnshiftShared(t *testing.T) {
	arr := &Array{Elements: []Object{
		&Integer{Value: 2},
		&Integer{Value: 3},
	}}
	arr.Share()

	result := arr.COWUnshift(&Integer{Value: 1})

	if result == arr {
		t.Error("COWUnshift on shared array should return new pointer")
	}

	// Original unchanged
	if len(arr.Elements) != 2 {
		t.Errorf("Original was modified, expected 2 elements, got %d", len(arr.Elements))
	}

	if len(result.Elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(result.Elements))
	}
}

func TestArrayShareRefCount(t *testing.T) {
	arr := &Array{Elements: []Object{&Integer{Value: 1}}}

	// Initially not shared
	if arr.IsShared() {
		t.Error("New array should not be shared")
	}

	// First share sets refCount to 2
	arr.Share()
	if !arr.IsShared() {
		t.Error("After Share(), array should be shared")
	}

	// Second share increments to 3
	arr.Share()
	if !arr.IsShared() {
		t.Error("After second Share(), array should still be shared")
	}

	// Unshare decrements to 2
	arr.Unshare()
	if !arr.IsShared() {
		t.Error("After one Unshare(), array should still be shared (refCount=2)")
	}

	// Unshare again decrements to 1
	arr.Unshare()
	if arr.IsShared() {
		t.Error("After two Unshare(), array should not be shared (refCount=1)")
	}
}

// =============================================================================
// Map COW Correctness Tests
// =============================================================================

func TestMapCOWSetUnshared(t *testing.T) {
	m := &Map{
		Pairs: map[string]Object{"a": &Integer{Value: 1}},
		Keys:  []string{"a"},
	}

	result := m.COWSet("b", &Integer{Value: 2})

	if result != m {
		t.Error("COWSet on unshared map should return same pointer")
	}

	if len(result.Pairs) != 2 {
		t.Errorf("Expected 2 pairs, got %d", len(result.Pairs))
	}

	if len(result.Keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(result.Keys))
	}
}

func TestMapCOWSetShared(t *testing.T) {
	m := &Map{
		Pairs: map[string]Object{"a": &Integer{Value: 1}},
		Keys:  []string{"a"},
	}
	m.Share()

	result := m.COWSet("b", &Integer{Value: 2})

	if result == m {
		t.Error("COWSet on shared map should return new pointer")
	}

	// Original unchanged
	if len(m.Pairs) != 1 {
		t.Errorf("Original was modified, expected 1 pair, got %d", len(m.Pairs))
	}

	// Result has new entry
	if len(result.Pairs) != 2 {
		t.Errorf("Expected 2 pairs in result, got %d", len(result.Pairs))
	}
}

func TestMapCOWSetUpdateExisting(t *testing.T) {
	m := &Map{
		Pairs: map[string]Object{"a": &Integer{Value: 1}},
		Keys:  []string{"a"},
	}

	result := m.COWSet("a", &Integer{Value: 99})

	// Key already exists, so Keys slice should not grow
	if len(result.Keys) != 1 {
		t.Errorf("Expected 1 key (no new key added), got %d", len(result.Keys))
	}

	val := result.Pairs["a"].(*Integer)
	if val.Value != 99 {
		t.Errorf("Expected 99, got %d", val.Value)
	}
}

func TestMapCOWDeleteUnshared(t *testing.T) {
	m := &Map{
		Pairs: map[string]Object{
			"a": &Integer{Value: 1},
			"b": &Integer{Value: 2},
		},
		Keys: []string{"a", "b"},
	}

	result := m.COWDelete("a")

	if result != m {
		t.Error("COWDelete on unshared map should return same pointer")
	}

	if len(result.Pairs) != 1 {
		t.Errorf("Expected 1 pair, got %d", len(result.Pairs))
	}

	if _, exists := result.Pairs["a"]; exists {
		t.Error("Key 'a' should have been deleted")
	}

	if len(result.Keys) != 1 || result.Keys[0] != "b" {
		t.Errorf("Expected Keys=['b'], got %v", result.Keys)
	}
}

func TestMapCOWDeleteShared(t *testing.T) {
	m := &Map{
		Pairs: map[string]Object{
			"a": &Integer{Value: 1},
			"b": &Integer{Value: 2},
		},
		Keys: []string{"a", "b"},
	}
	m.Share()

	result := m.COWDelete("a")

	if result == m {
		t.Error("COWDelete on shared map should return new pointer")
	}

	// Original unchanged
	if len(m.Pairs) != 2 {
		t.Errorf("Original was modified, expected 2 pairs, got %d", len(m.Pairs))
	}

	// Result has entry removed
	if len(result.Pairs) != 1 {
		t.Errorf("Expected 1 pair in result, got %d", len(result.Pairs))
	}
}

func TestMapCOWDeleteNonExistent(t *testing.T) {
	m := &Map{
		Pairs: map[string]Object{"a": &Integer{Value: 1}},
		Keys:  []string{"a"},
	}

	result := m.COWDelete("nonexistent")

	// Should return same map unchanged
	if result != m {
		t.Error("COWDelete of non-existent key should return same pointer")
	}

	if len(result.Pairs) != 1 {
		t.Errorf("Expected 1 pair (unchanged), got %d", len(result.Pairs))
	}
}

func TestMapShareRefCount(t *testing.T) {
	m := &Map{
		Pairs: map[string]Object{"a": &Integer{Value: 1}},
		Keys:  []string{"a"},
	}

	if m.IsShared() {
		t.Error("New map should not be shared")
	}

	m.Share()
	if !m.IsShared() {
		t.Error("After Share(), map should be shared")
	}

	m.Share()
	if !m.IsShared() {
		t.Error("After second Share(), map should still be shared")
	}

	m.Unshare()
	if !m.IsShared() {
		t.Error("After one Unshare(), map should still be shared (refCount=2)")
	}

	m.Unshare()
	if m.IsShared() {
		t.Error("After two Unshare(), map should not be shared (refCount=1)")
	}
}

// =============================================================================
// Integration-style Tests
// =============================================================================

func TestArrayChainedOperationsUnshared(t *testing.T) {
	// Simulate: [1] > array.push(2) > array.push(3)
	arr := &Array{Elements: []Object{&Integer{Value: 1}}}

	// Each operation should modify in place since array is never shared
	arr = arr.COWPush(&Integer{Value: 2})
	arr = arr.COWPush(&Integer{Value: 3})

	if len(arr.Elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(arr.Elements))
	}

	// Verify values
	for i, expected := range []int64{1, 2, 3} {
		val := arr.Elements[i].(*Integer)
		if val.Value != expected {
			t.Errorf("Element %d: expected %d, got %d", i, expected, val.Value)
		}
	}
}

func TestMapChainedOperationsUnshared(t *testing.T) {
	// Simulate: {} > set("a", 1) > set("b", 2) > set("c", 3)
	m := &Map{
		Pairs: make(map[string]Object),
		Keys:  []string{},
	}

	m = m.COWSet("a", &Integer{Value: 1})
	m = m.COWSet("b", &Integer{Value: 2})
	m = m.COWSet("c", &Integer{Value: 3})

	if len(m.Pairs) != 3 {
		t.Errorf("Expected 3 pairs, got %d", len(m.Pairs))
	}

	// Verify insertion order preserved
	expectedKeys := []string{"a", "b", "c"}
	for i, expected := range expectedKeys {
		if m.Keys[i] != expected {
			t.Errorf("Key %d: expected %s, got %s", i, expected, m.Keys[i])
		}
	}
}
