package object

import (
	"fmt"
	"testing"
)

// BenchmarkEnvGetShallow benchmarks variable lookup in current scope
func BenchmarkEnvGetShallow(b *testing.B) {
	env := NewEnvironment()
	env.Set("x", &Integer{Value: 42})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.Get("x")
	}
}

// BenchmarkEnvGetDepth5 benchmarks variable lookup 5 scopes deep
func BenchmarkEnvGetDepth5(b *testing.B) {
	// Create chain: env0 -> env1 -> env2 -> env3 -> env4
	env := NewEnvironment()
	env.Set("x", &Integer{Value: 42})

	for i := 0; i < 5; i++ {
		env = NewEnclosedEnvironment(env)
		env.Set(fmt.Sprintf("local%d", i), &Integer{Value: int64(i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.Get("x") // Lookup from deepest scope
	}
}

// BenchmarkEnvGetDepth10 benchmarks variable lookup 10 scopes deep
func BenchmarkEnvGetDepth10(b *testing.B) {
	env := NewEnvironment()
	env.Set("x", &Integer{Value: 42})

	for i := 0; i < 10; i++ {
		env = NewEnclosedEnvironment(env)
		env.Set(fmt.Sprintf("local%d", i), &Integer{Value: int64(i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.Get("x")
	}
}

// BenchmarkEnvGetDepth20 benchmarks variable lookup 20 scopes deep
func BenchmarkEnvGetDepth20(b *testing.B) {
	env := NewEnvironment()
	env.Set("x", &Integer{Value: 42})

	for i := 0; i < 20; i++ {
		env = NewEnclosedEnvironment(env)
		env.Set(fmt.Sprintf("local%d", i), &Integer{Value: int64(i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.Get("x")
	}
}

// BenchmarkEnvGetMiss benchmarks lookup that fails (not found)
func BenchmarkEnvGetMiss(b *testing.B) {
	env := NewEnvironment()
	env.Set("x", &Integer{Value: 42})

	for i := 0; i < 10; i++ {
		env = NewEnclosedEnvironment(env)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.Get("nonexistent")
	}
}

// BenchmarkEnvSet benchmarks variable assignment
func BenchmarkEnvSet(b *testing.B) {
	env := NewEnvironment()
	val := &Integer{Value: 42}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.Set("x", val)
	}
}
