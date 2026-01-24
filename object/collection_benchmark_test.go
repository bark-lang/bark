package object

import (
	"testing"
)

// =============================================================================
// COW Array Benchmarks (new implementation)
// =============================================================================

// BenchmarkCOWPushUnshared benchmarks push on unshared array (in-place)
func BenchmarkCOWPushUnshared100(b *testing.B) {
	// Pre-create array
	elements := make([]Object, 100)
	for j := range elements {
		elements[j] = &Integer{Value: int64(j)}
	}
	val := &Integer{Value: 999}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset array for each iteration
		arr := &Array{Elements: elements[:100:100]} // Force new backing array
		arr = arr.COWPush(val)
		_ = arr
	}
}

// BenchmarkCOWPushShared benchmarks push on shared array (must copy)
func BenchmarkCOWPushShared100(b *testing.B) {
	// Pre-create array
	elements := make([]Object, 100)
	for j := range elements {
		elements[j] = &Integer{Value: int64(j)}
	}
	val := &Integer{Value: 999}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr := &Array{Elements: elements}
		arr.Share() // Mark as shared

		arr = arr.COWPush(val)
		_ = arr
	}
}

// BenchmarkCOWChainedUnshared benchmarks 5 chained ops on unshared array
func BenchmarkCOWChainedUnshared5(b *testing.B) {
	// Pre-create base elements
	baseElements := make([]Object, 100)
	for j := range baseElements {
		baseElements[j] = &Integer{Value: int64(j)}
	}
	vals := make([]*Integer, 5)
	for j := range vals {
		vals[j] = &Integer{Value: int64(j)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Copy base elements to start fresh
		elements := make([]Object, 100)
		copy(elements, baseElements)
		arr := &Array{Elements: elements}

		// 5 chained operations - all in-place since never shared
		for j := 0; j < 5; j++ {
			arr = arr.COWPush(vals[j])
		}
	}
}

// =============================================================================
// Original Copy Benchmarks (baseline)
// =============================================================================

// BenchmarkArrayCopy benchmarks full array copy (current behavior)
func BenchmarkArrayCopy100(b *testing.B) {
	// Create source array with 100 elements
	elements := make([]Object, 100)
	for i := range elements {
		elements[i] = &Integer{Value: int64(i)}
	}
	arr := &Array{Elements: elements}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate push operation (full copy + append)
		newElements := make([]Object, len(arr.Elements))
		copy(newElements, arr.Elements)
		newElements = append(newElements, &Integer{Value: 999})
		_ = &Array{Elements: newElements}
	}
}

// BenchmarkArrayCopy1000 benchmarks full array copy with 1000 elements
func BenchmarkArrayCopy1000(b *testing.B) {
	elements := make([]Object, 1000)
	for i := range elements {
		elements[i] = &Integer{Value: int64(i)}
	}
	arr := &Array{Elements: elements}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newElements := make([]Object, len(arr.Elements))
		copy(newElements, arr.Elements)
		newElements = append(newElements, &Integer{Value: 999})
		_ = &Array{Elements: newElements}
	}
}

// BenchmarkArrayCopy10000 benchmarks full array copy with 10000 elements
func BenchmarkArrayCopy10000(b *testing.B) {
	elements := make([]Object, 10000)
	for i := range elements {
		elements[i] = &Integer{Value: int64(i)}
	}
	arr := &Array{Elements: elements}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newElements := make([]Object, len(arr.Elements))
		copy(newElements, arr.Elements)
		newElements = append(newElements, &Integer{Value: 999})
		_ = &Array{Elements: newElements}
	}
}

// BenchmarkMapCopy100 benchmarks full map copy with 100 entries
func BenchmarkMapCopy100(b *testing.B) {
	pairs := make(map[string]Object)
	keys := make([]string, 100)
	for i := 0; i < 100; i++ {
		key := string(rune('a' + i%26))
		keys[i] = key
		pairs[key] = &Integer{Value: int64(i)}
	}
	m := &Map{Pairs: pairs, Keys: keys}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate set operation (full copy)
		newPairs := make(map[string]Object)
		for k, v := range m.Pairs {
			newPairs[k] = v
		}
		newKeys := make([]string, len(m.Keys))
		copy(newKeys, m.Keys)
		newPairs["newkey"] = &Integer{Value: 999}
		newKeys = append(newKeys, "newkey")
		_ = &Map{Pairs: newPairs, Keys: newKeys}
	}
}

// BenchmarkChainedArrayOps simulates a chain of 5 array operations
func BenchmarkChainedArrayOps5(b *testing.B) {
	elements := make([]Object, 100)
	for i := range elements {
		elements[i] = &Integer{Value: int64(i)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr := &Array{Elements: elements}

		// 5 chained operations, each copies the full array
		for j := 0; j < 5; j++ {
			newElements := make([]Object, len(arr.Elements))
			copy(newElements, arr.Elements)
			newElements = append(newElements, &Integer{Value: int64(j)})
			arr = &Array{Elements: newElements}
		}
	}
}
