package bytecode

// OpCode represents a single bytecode instruction
type OpCode byte

const (
	// Constants & Literals
	OpConstant OpCode = iota // Load constant from pool
	OpNull                   // Push NULL singleton
	OpTrue                   // Push TRUE singleton
	OpFalse                  // Push FALSE singleton

	// Arithmetic
	OpAdd // Pop b,a; push a+b
	OpSub // Pop b,a; push a-b
	OpMul // Pop b,a; push a*b
	OpDiv // Pop b,a; push a/b
	OpMod // Pop b,a; push a%b
	OpNeg // Negate top of stack

	// Comparison
	OpEq  // Pop b,a; push a==b
	OpNe  // Pop b,a; push a!=b
	OpLt  // Pop b,a; push a<b
	OpLe  // Pop b,a; push a<=b
	OpGt  // Pop b,a; push a>b
	OpGe  // Pop b,a; push a>=b
	OpNot // Logical not

	// Variables
	OpLoadLocal   // Push locals[slot]
	OpStoreLocal  // Store top to locals[slot]
	OpLoadGlobal  // Push globals[name]
	OpStoreGlobal // Store top to globals[name]
	OpLoadUpval   // Push upvalue[slot]
	OpStoreUpval  // Store top to upvalue[slot]

	// Control Flow
	OpJump        // Unconditional jump
	OpJumpIfFalse // Jump if top is falsy (does not pop)
	OpJumpIfTrue  // Jump if top is truthy (does not pop)
	OpLoop        // Jump backward

	// Functions
	OpCall       // Call function with argc args
	OpReturn     // Return from function
	OpClosure    // Create closure with upvalues
	OpCloseUpval // Close upvalue when scope exits
	OpUnpackCall // Call function, unpacking tuple if needed (for linked anonymous functions)

	// Collections
	OpArray    // Build array from count stack values
	OpMap      // Build map from count*2 stack values
	OpIndexGet // Pop index, coll; push coll[index]
	OpIndexSet // Pop val, idx, coll; set coll[idx]=val; push coll

	// Link Operator
	OpLinkBind // Bind top to variable (with COW share)
	OpLinkCall // Call with prepended argument
	OpMember   // Module member access

	// Tuples & Error Handling
	OpTuple     // Build tuple from count stack values
	OpCapture   // Handle (error, result) tuple
	OpChainStop // Check for chain stop condition

	// Memoization
	OpMemoCheck // Check memoization cache
	OpMemoStore // Store result in cache

	// Builtins
	OpBuiltin       // Direct builtin call by index
	OpUnpackBuiltin // Builtin call with tuple unpacking (for link operator)

	// String Interpolation
	OpInterpolate // Interpolate string constant with variables

	// Stack
	OpPop  // Discard top of stack
	OpDup  // Duplicate top of stack
	OpSwap // Swap top two stack values
)

// OpCodeNames maps opcodes to their string names for debugging
var OpCodeNames = [...]string{
	OpConstant: "OP_CONSTANT",
	OpNull:     "OP_NULL",
	OpTrue:     "OP_TRUE",
	OpFalse:    "OP_FALSE",

	OpAdd: "OP_ADD",
	OpSub: "OP_SUB",
	OpMul: "OP_MUL",
	OpDiv: "OP_DIV",
	OpMod: "OP_MOD",
	OpNeg: "OP_NEG",

	OpEq:  "OP_EQ",
	OpNe:  "OP_NE",
	OpLt:  "OP_LT",
	OpLe:  "OP_LE",
	OpGt:  "OP_GT",
	OpGe:  "OP_GE",
	OpNot: "OP_NOT",

	OpLoadLocal:   "OP_LOAD_LOCAL",
	OpStoreLocal:  "OP_STORE_LOCAL",
	OpLoadGlobal:  "OP_LOAD_GLOBAL",
	OpStoreGlobal: "OP_STORE_GLOBAL",
	OpLoadUpval:   "OP_LOAD_UPVAL",
	OpStoreUpval:  "OP_STORE_UPVAL",

	OpJump:        "OP_JUMP",
	OpJumpIfFalse: "OP_JUMP_IF_FALSE",
	OpJumpIfTrue:  "OP_JUMP_IF_TRUE",
	OpLoop:        "OP_LOOP",

	OpCall:       "OP_CALL",
	OpReturn:     "OP_RETURN",
	OpClosure:    "OP_CLOSURE",
	OpCloseUpval: "OP_CLOSE_UPVAL",
	OpUnpackCall: "OP_UNPACK_CALL",

	OpArray:    "OP_ARRAY",
	OpMap:      "OP_MAP",
	OpIndexGet: "OP_INDEX_GET",
	OpIndexSet: "OP_INDEX_SET",

	OpLinkBind: "OP_LINK_BIND",
	OpLinkCall: "OP_LINK_CALL",
	OpMember:   "OP_MEMBER",

	OpTuple:     "OP_TUPLE",
	OpCapture:   "OP_CAPTURE",
	OpChainStop: "OP_CHAIN_STOP",

	OpMemoCheck: "OP_MEMO_CHECK",
	OpMemoStore: "OP_MEMO_STORE",

	OpBuiltin:       "OP_BUILTIN",
	OpUnpackBuiltin: "OP_UNPACK_BUILTIN",

	OpInterpolate: "OP_INTERPOLATE",

	OpPop:  "OP_POP",
	OpDup:  "OP_DUP",
	OpSwap: "OP_SWAP",
}

// String returns the name of the opcode
func (op OpCode) String() string {
	if int(op) < len(OpCodeNames) {
		return OpCodeNames[op]
	}
	return "OP_UNKNOWN"
}

// Definition describes an opcode's operands
type Definition struct {
	Name          string
	OperandWidths []int // Width of each operand in bytes
}

// Definitions maps opcodes to their definitions
var Definitions = map[OpCode]*Definition{
	// Constants - OpConstant uses 2-byte index
	OpConstant: {"OP_CONSTANT", []int{2}},
	OpNull:     {"OP_NULL", []int{}},
	OpTrue:     {"OP_TRUE", []int{}},
	OpFalse:    {"OP_FALSE", []int{}},

	// Arithmetic - no operands
	OpAdd: {"OP_ADD", []int{}},
	OpSub: {"OP_SUB", []int{}},
	OpMul: {"OP_MUL", []int{}},
	OpDiv: {"OP_DIV", []int{}},
	OpMod: {"OP_MOD", []int{}},
	OpNeg: {"OP_NEG", []int{}},

	// Comparison - no operands
	OpEq:  {"OP_EQ", []int{}},
	OpNe:  {"OP_NE", []int{}},
	OpLt:  {"OP_LT", []int{}},
	OpLe:  {"OP_LE", []int{}},
	OpGt:  {"OP_GT", []int{}},
	OpGe:  {"OP_GE", []int{}},
	OpNot: {"OP_NOT", []int{}},

	// Variables - 2-byte slot/name index
	OpLoadLocal:   {"OP_LOAD_LOCAL", []int{2}},
	OpStoreLocal:  {"OP_STORE_LOCAL", []int{2}},
	OpLoadGlobal:  {"OP_LOAD_GLOBAL", []int{2}},
	OpStoreGlobal: {"OP_STORE_GLOBAL", []int{2}},
	OpLoadUpval:   {"OP_LOAD_UPVAL", []int{1}},
	OpStoreUpval:  {"OP_STORE_UPVAL", []int{1}},

	// Control Flow - 2-byte offset
	OpJump:        {"OP_JUMP", []int{2}},
	OpJumpIfFalse: {"OP_JUMP_IF_FALSE", []int{2}},
	OpJumpIfTrue:  {"OP_JUMP_IF_TRUE", []int{2}},
	OpLoop:        {"OP_LOOP", []int{2}},

	// Functions
	OpCall:       {"OP_CALL", []int{1}},        // 1-byte arg count
	OpReturn:     {"OP_RETURN", []int{}},       // no operands
	OpClosure:    {"OP_CLOSURE", []int{2}},     // 2-byte constant index (upvalue info follows)
	OpCloseUpval: {"OP_CLOSE_UPVAL", []int{}},  // no operands
	OpUnpackCall: {"OP_UNPACK_CALL", []int{1}}, // 1-byte expected arity (for tuple unpacking)

	// Collections - 2-byte count
	OpArray:    {"OP_ARRAY", []int{2}},
	OpMap:      {"OP_MAP", []int{2}},
	OpIndexGet: {"OP_INDEX_GET", []int{}},
	OpIndexSet: {"OP_INDEX_SET", []int{}},

	// Link Operator
	OpLinkBind: {"OP_LINK_BIND", []int{2}}, // 2-byte name index
	OpLinkCall: {"OP_LINK_CALL", []int{1}}, // 1-byte arg count (excluding prepended)
	OpMember:   {"OP_MEMBER", []int{2}},    // 2-byte name index

	// Tuples & Errors
	OpTuple:     {"OP_TUPLE", []int{1}},   // 1-byte count
	OpCapture:   {"OP_CAPTURE", []int{2}}, // 2-byte err var index, result var follows
	OpChainStop: {"OP_CHAIN_STOP", []int{}},

	// Memoization
	OpMemoCheck: {"OP_MEMO_CHECK", []int{1}}, // 1-byte arg count
	OpMemoStore: {"OP_MEMO_STORE", []int{}},

	// Builtins
	OpBuiltin:       {"OP_BUILTIN", []int{2, 1}},        // 2-byte name index, 1-byte arg count
	OpUnpackBuiltin: {"OP_UNPACK_BUILTIN", []int{2, 1}}, // 2-byte name index, 1-byte extra arg count (first arg may be tuple to unpack)

	// String Interpolation
	OpInterpolate: {"OP_INTERPOLATE", []int{2}}, // 2-byte constant index

	// Stack
	OpPop:  {"OP_POP", []int{}},
	OpDup:  {"OP_DUP", []int{}},
	OpSwap: {"OP_SWAP", []int{}},
}

// Lookup returns the definition for an opcode
func Lookup(op OpCode) (*Definition, bool) {
	def, ok := Definitions[op]
	return def, ok
}
