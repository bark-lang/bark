package bytecode

import (
	"fmt"
	"strings"
)

// Disassemble returns a human-readable representation of the chunk
func Disassemble(chunk *Chunk, name string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("== %s ==\n", name))

	for offset := 0; offset < len(chunk.Code); {
		offset = DisassembleInstruction(chunk, offset, &sb)
	}

	return sb.String()
}

// DisassembleInstruction disassembles a single instruction and returns the next offset
func DisassembleInstruction(chunk *Chunk, offset int, sb *strings.Builder) int {
	_, _ = fmt.Fprintf(sb, "%04d ", offset)

	// Show line number or | for same line
	if offset > 0 && chunk.GetLine(offset) == chunk.GetLine(offset-1) {
		sb.WriteString("   | ")
	} else {
		_, _ = fmt.Fprintf(sb, "%4d ", chunk.GetLine(offset))
	}

	op := OpCode(chunk.Code[offset])
	def, ok := Lookup(op)
	if !ok {
		_, _ = fmt.Fprintf(sb, "Unknown opcode %d\n", op)
		return offset + 1
	}

	switch op {
	case OpConstant:
		return constantInstruction(def.Name, chunk, offset, sb)
	case OpNull, OpTrue, OpFalse:
		return simpleInstruction(def.Name, offset, sb)
	case OpAdd, OpSub, OpMul, OpDiv, OpMod, OpNeg:
		return simpleInstruction(def.Name, offset, sb)
	case OpEq, OpNe, OpLt, OpLe, OpGt, OpGe, OpNot:
		return simpleInstruction(def.Name, offset, sb)
	case OpLoadLocal, OpStoreLocal:
		return byteInstruction(def.Name, chunk, offset, 2, sb)
	case OpLoadGlobal, OpStoreGlobal:
		return nameInstruction(def.Name, chunk, offset, sb)
	case OpLoadUpval, OpStoreUpval:
		return byteInstruction(def.Name, chunk, offset, 1, sb)
	case OpJump, OpJumpIfFalse, OpJumpIfTrue:
		return jumpInstruction(def.Name, 1, chunk, offset, sb)
	case OpLoop:
		return jumpInstruction(def.Name, -1, chunk, offset, sb)
	case OpCall:
		return byteInstruction(def.Name, chunk, offset, 1, sb)
	case OpReturn:
		return simpleInstruction(def.Name, offset, sb)
	case OpClosure:
		return closureInstruction(chunk, offset, sb)
	case OpCloseUpval:
		return simpleInstruction(def.Name, offset, sb)
	case OpUnpackCall:
		return byteInstruction(def.Name, chunk, offset, 1, sb)
	case OpArray, OpMap:
		return byteInstruction(def.Name, chunk, offset, 2, sb)
	case OpIndexGet, OpIndexSet:
		return simpleInstruction(def.Name, offset, sb)
	case OpLinkBind:
		return nameInstruction(def.Name, chunk, offset, sb)
	case OpLinkCall:
		return byteInstruction(def.Name, chunk, offset, 1, sb)
	case OpMember:
		return nameInstruction(def.Name, chunk, offset, sb)
	case OpTuple:
		return byteInstruction(def.Name, chunk, offset, 1, sb)
	case OpCapture:
		return captureInstruction(chunk, offset, sb)
	case OpChainStop:
		return simpleInstruction(def.Name, offset, sb)
	case OpMemoCheck:
		return byteInstruction(def.Name, chunk, offset, 1, sb)
	case OpMemoStore:
		return simpleInstruction(def.Name, offset, sb)
	case OpBuiltin:
		return builtinInstruction("OP_BUILTIN", chunk, offset, sb)
	case OpUnpackBuiltin:
		return builtinInstruction("OP_UNPACK_BUILTIN", chunk, offset, sb)
	case OpPop, OpDup, OpSwap:
		return simpleInstruction(def.Name, offset, sb)
	default:
		_, _ = fmt.Fprintf(sb, "%s\n", def.Name)
		return offset + 1
	}
}

func simpleInstruction(name string, offset int, sb *strings.Builder) int {
	_, _ = fmt.Fprintf(sb, "%s\n", name)
	return offset + 1
}

func constantInstruction(name string, chunk *Chunk, offset int, sb *strings.Builder) int {
	idx := chunk.ReadUint16(offset + 1)
	_, _ = fmt.Fprintf(sb, "%-16s %4d '", name, idx)
	if int(idx) < len(chunk.Constants) {
		sb.WriteString(chunk.Constants[idx].Inspect())
	} else {
		sb.WriteString("???")
	}
	sb.WriteString("'\n")
	return offset + 3
}

func byteInstruction(name string, chunk *Chunk, offset int, width int, sb *strings.Builder) int {
	var value int
	if width == 1 {
		value = int(chunk.ReadUint8(offset + 1))
	} else {
		value = int(chunk.ReadUint16(offset + 1))
	}
	_, _ = fmt.Fprintf(sb, "%-16s %4d\n", name, value)
	return offset + 1 + width
}

func nameInstruction(name string, chunk *Chunk, offset int, sb *strings.Builder) int {
	idx := chunk.ReadUint16(offset + 1)
	_, _ = fmt.Fprintf(sb, "%-16s %4d '", name, idx)
	if int(idx) < len(chunk.Names) {
		sb.WriteString(chunk.Names[idx])
	} else {
		sb.WriteString("???")
	}
	sb.WriteString("'\n")
	return offset + 3
}

func jumpInstruction(name string, sign int, chunk *Chunk, offset int, sb *strings.Builder) int {
	jump := int(chunk.ReadUint16(offset + 1))
	target := offset + 3 + sign*jump
	_, _ = fmt.Fprintf(sb, "%-16s %4d -> %d\n", name, offset, target)
	return offset + 3
}

func closureInstruction(chunk *Chunk, offset int, sb *strings.Builder) int {
	idx := chunk.ReadUint16(offset + 1)
	_, _ = fmt.Fprintf(sb, "%-16s %4d ", "OP_CLOSURE", idx)
	if int(idx) < len(chunk.Constants) {
		sb.WriteString(chunk.Constants[idx].Inspect())
	}
	sb.WriteString("\n")

	// Read upvalue info - each upvalue is 2 bytes: isLocal(1) + index(1)
	offset += 3
	if int(idx) < len(chunk.Constants) {
		// Check if it's a compiled function by checking for UpvalueCount method
		if fn, ok := chunk.Constants[idx].(interface{ GetUpvalueCount() int }); ok {
			for i := 0; i < fn.GetUpvalueCount(); i++ {
				isLocal := chunk.ReadUint8(offset)
				index := chunk.ReadUint8(offset + 1)
				localStr := "upvalue"
				if isLocal == 1 {
					localStr = "local"
				}
				_, _ = fmt.Fprintf(sb, "%04d    |                     %s %d\n", offset, localStr, index)
				offset += 2
			}
		}
	}
	return offset
}

func captureInstruction(chunk *Chunk, offset int, sb *strings.Builder) int {
	errIdx := chunk.ReadUint16(offset + 1)
	valIdx := chunk.ReadUint16(offset + 3)
	errName := "?"
	valName := "?"
	if int(errIdx) < len(chunk.Names) {
		errName = chunk.Names[errIdx]
	}
	if int(valIdx) < len(chunk.Names) {
		valName = chunk.Names[valIdx]
	}
	_, _ = fmt.Fprintf(sb, "%-16s %s, %s\n", "OP_CAPTURE", errName, valName)
	return offset + 5
}

func builtinInstruction(opName string, chunk *Chunk, offset int, sb *strings.Builder) int {
	nameIdx := chunk.ReadUint16(offset + 1)
	argc := chunk.ReadUint8(offset + 3)
	name := "?"
	if int(nameIdx) < len(chunk.Names) {
		name = chunk.Names[nameIdx]
	}
	_, _ = fmt.Fprintf(sb, "%-16s %s (%d args)\n", opName, name, argc)
	return offset + 4
}
