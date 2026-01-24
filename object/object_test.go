package object

import (
	"fmt"
	"testing"
)

// BenchmarkMemoCacheGet benchmarks the fast hash-based cache lookup
func BenchmarkMemoCacheGet(b *testing.B) {
	cache := NewMemoCache()

	// Pre-populate cache with entries
	for i := 0; i < 1000; i++ {
		for j := 0; j < 100; j++ {
			args := []Object{
				&Integer{Value: int64(i)},
				&Integer{Value: int64(j)},
			}
			cache.Set(args, &Integer{Value: int64(i + j)})
		}
	}

	// Benchmark lookups
	args := []Object{
		&Integer{Value: 500},
		&Integer{Value: 50},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(args)
	}
}

// BenchmarkMemoCacheSet benchmarks cache insertion
func BenchmarkMemoCacheSet(b *testing.B) {
	cache := NewMemoCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		args := []Object{
			&Integer{Value: int64(i % 1000)},
			&Integer{Value: int64(i % 100)},
		}
		cache.Set(args, &Integer{Value: int64(i)})
	}
}

// BenchmarkOldStyleKey benchmarks the old string-based key generation
func BenchmarkOldStyleKey(b *testing.B) {
	args := []Object{
		&Integer{Value: 500},
		&Integer{Value: 50},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate old serializeArgs
		key := fmt.Sprintf("%s:%s", args[0].Type(), args[0].Inspect())
		for j := 1; j < len(args); j++ {
			key = fmt.Sprintf("%s|%s:%s", key, args[j].Type(), args[j].Inspect())
		}
		_ = key
	}
}

// BenchmarkNewStyleHash benchmarks the new hash-based key generation
func BenchmarkNewStyleHash(b *testing.B) {
	cache := NewMemoCache()
	args := []Object{
		&Integer{Value: 500},
		&Integer{Value: 50},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.hashArgs(args)
	}
}
