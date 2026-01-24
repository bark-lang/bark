package bytecode

import (
	"fmt"

	"gitlab.com/bark-lang/bark/object"
)

// CompiledFunction represents a compiled function
type CompiledFunction struct {
	Chunk        *Chunk
	Name         string
	Arity        int
	UpvalueCount int
	IsMemoized   bool
	Cache        *object.MemoCache // For memoized functions
	LocalNames   []string          // Local variable names by slot index (for string interpolation)
}

// Ensure CompiledFunction implements object.Object
var _ object.Object = (*CompiledFunction)(nil)

// Type returns the object type
func (cf *CompiledFunction) Type() object.ObjectType {
	return "COMPILED_FUNCTION"
}

// Inspect returns a string representation
func (cf *CompiledFunction) Inspect() string {
	if cf.IsMemoized {
		return fmt.Sprintf("<mfn %s>", cf.Name)
	}
	return fmt.Sprintf("<fn %s>", cf.Name)
}

// GetUpvalueCount returns the number of upvalues for this function
// Used by debug.go for disassembly
func (cf *CompiledFunction) GetUpvalueCount() int {
	return cf.UpvalueCount
}
