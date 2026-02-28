package bytecode

import (
	"encoding/binary"

	"gitlab.com/bark-lang/barki/object"
)

// Chunk holds bytecode instructions and associated data
type Chunk struct {
	Code      []byte          // Bytecode instructions
	Constants []object.Object // Constant pool
	Lines     []int           // Line numbers for debugging (one per instruction byte)
	Names     []string        // Name pool for globals/members
}

// NewChunk creates a new empty chunk
func NewChunk() *Chunk {
	return &Chunk{
		Code:      make([]byte, 0, 256),
		Constants: make([]object.Object, 0, 64),
		Lines:     make([]int, 0, 256),
		Names:     make([]string, 0, 32),
	}
}

// Write appends a byte to the chunk
func (c *Chunk) Write(b byte, line int) {
	c.Code = append(c.Code, b)
	c.Lines = append(c.Lines, line)
}

// WriteOp writes an opcode to the chunk
func (c *Chunk) WriteOp(op OpCode, line int) {
	c.Write(byte(op), line)
}

// WriteOpWithOperand writes an opcode with a single operand
func (c *Chunk) WriteOpWithOperand(op OpCode, operand int, line int) int {
	def, ok := Lookup(op)
	if !ok {
		return -1
	}

	offset := len(c.Code)
	c.WriteOp(op, line)

	if len(def.OperandWidths) > 0 {
		width := def.OperandWidths[0]
		switch width {
		case 1:
			c.Write(byte(operand), line)
		case 2:
			c.Write(byte(operand>>8), line)
			c.Write(byte(operand), line)
		}
	}

	return offset
}

// WriteOpWithOperands writes an opcode with multiple operands
func (c *Chunk) WriteOpWithOperands(op OpCode, operands []int, line int) int {
	def, ok := Lookup(op)
	if !ok {
		return -1
	}

	offset := len(c.Code)
	c.WriteOp(op, line)

	for i, width := range def.OperandWidths {
		if i >= len(operands) {
			break
		}
		switch width {
		case 1:
			c.Write(byte(operands[i]), line)
		case 2:
			c.Write(byte(operands[i]>>8), line)
			c.Write(byte(operands[i]), line)
		}
	}

	return offset
}

// AddConstant adds a constant to the pool and returns its index
func (c *Chunk) AddConstant(obj object.Object) int {
	// Check for existing constant (deduplication for immutables)
	switch v := obj.(type) {
	case *object.Integer:
		for i, existing := range c.Constants {
			if e, ok := existing.(*object.Integer); ok && e.Value == v.Value {
				return i
			}
		}
	case *object.Float:
		for i, existing := range c.Constants {
			if e, ok := existing.(*object.Float); ok && e.Value == v.Value {
				return i
			}
		}
	case *object.String:
		for i, existing := range c.Constants {
			if e, ok := existing.(*object.String); ok && e.Value == v.Value {
				return i
			}
		}
	}

	c.Constants = append(c.Constants, obj)
	return len(c.Constants) - 1
}

// AddName adds a name to the name pool and returns its index
func (c *Chunk) AddName(name string) int {
	// Check for existing name (deduplication)
	for i, existing := range c.Names {
		if existing == name {
			return i
		}
	}
	c.Names = append(c.Names, name)
	return len(c.Names) - 1
}

// PatchJump patches a jump instruction's offset
func (c *Chunk) PatchJump(offset int) {
	// Calculate jump distance: from after the jump instruction to current position
	jump := len(c.Code) - offset - 3 // -3 for opcode + 2 bytes of offset

	if jump > 65535 {
		panic("jump offset too large")
	}

	c.Code[offset+1] = byte(jump >> 8)
	c.Code[offset+2] = byte(jump)
}

// EmitJump writes a jump instruction with placeholder offset
func (c *Chunk) EmitJump(op OpCode, line int) int {
	c.WriteOp(op, line)
	c.Write(0xff, line) // Placeholder high byte
	c.Write(0xff, line) // Placeholder low byte
	return len(c.Code) - 3
}

// EmitLoop writes a loop instruction that jumps backward
func (c *Chunk) EmitLoop(loopStart int, line int) {
	c.WriteOp(OpLoop, line)

	offset := len(c.Code) - loopStart + 2 // +2 for the two bytes we're about to write
	if offset > 65535 {
		panic("loop body too large")
	}

	c.Write(byte(offset>>8), line)
	c.Write(byte(offset), line)
}

// ReadUint16 reads a 2-byte big-endian integer from the code at offset
func (c *Chunk) ReadUint16(offset int) uint16 {
	return binary.BigEndian.Uint16(c.Code[offset : offset+2])
}

// ReadUint8 reads a 1-byte integer from the code at offset
func (c *Chunk) ReadUint8(offset int) uint8 {
	return c.Code[offset]
}

// GetLine returns the source line for a given instruction offset
func (c *Chunk) GetLine(offset int) int {
	if offset < len(c.Lines) {
		return c.Lines[offset]
	}
	return 0
}

// Len returns the length of the bytecode
func (c *Chunk) Len() int {
	return len(c.Code)
}
