package bytecode

import (
	"fmt"
	"strings"
	"unsafe"

	"gitlab.com/bark-lang/barki/evaluator/builtins"
	"gitlab.com/bark-lang/barki/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/barki/object"
)

const (
	InitialStackSize  = 1024
	InitialFramesSize = 64
	MaxStackSize      = 262144
	MaxFramesSize     = 16384
)

// VM is the bytecode virtual machine
type VM struct {
	stack    []object.Object
	stackTop int

	frames     []CallFrame
	frameCount int

	globals    map[string]object.Object
	builtins   map[string]*object.Builtin
	openUpvals *ObjUpvalue // Linked list of open upvalues

	// Error state
	lastError object.Object
}

// CallFrame represents a function call frame
type CallFrame struct {
	closure *ObjClosure
	ip      int // Instruction pointer
	slots   int // Base slot in stack for this frame
}

// ObjClosure represents a runtime closure
type ObjClosure struct {
	Function *CompiledFunction
	Upvalues []*ObjUpvalue
	Cache    *object.MemoCache // For memoized functions
}

// ObjUpvalue represents a captured variable
type ObjUpvalue struct {
	Location *object.Object // Pointer to stack slot or closed value
	Closed   object.Object  // Value when closed
	Next     *ObjUpvalue    // For linked list of open upvalues
}

// NewVM creates a new virtual machine
func NewVM() *VM {
	vm := &VM{
		stack:    make([]object.Object, InitialStackSize),
		frames:   make([]CallFrame, InitialFramesSize),
		globals:  make(map[string]object.Object),
		builtins: builtins.GetAll(),
	}
	return vm
}

// GetGlobals returns the globals map (for debugging)
func (vm *VM) GetGlobals() map[string]object.Object {
	return vm.globals
}

// Run executes bytecode and returns the result
func (vm *VM) Run(fn *CompiledFunction) (object.Object, error) {
	// Create top-level closure
	closure := &ObjClosure{
		Function: fn,
		Upvalues: make([]*ObjUpvalue, 0),
	}

	// Push closure onto stack
	vm.push(closure)

	// Create initial frame
	vm.frames[0] = CallFrame{
		closure: closure,
		ip:      0,
		slots:   0,
	}
	vm.frameCount = 1

	return vm.execute()
}

func (vm *VM) execute() (object.Object, error) {
	for {
		frame := &vm.frames[vm.frameCount-1]
		chunk := frame.closure.Function.Chunk

		if frame.ip >= len(chunk.Code) {
			return helpers.NULL, nil
		}

		op := OpCode(chunk.Code[frame.ip])
		frame.ip++

		switch op {
		case OpConstant:
			idx := chunk.ReadUint16(frame.ip)
			frame.ip += 2
			vm.push(chunk.Constants[idx])

		case OpNull:
			vm.push(helpers.NULL)

		case OpTrue:
			vm.push(helpers.TRUE)

		case OpFalse:
			vm.push(helpers.FALSE)

		case OpAdd:
			if err := vm.binaryOp(op); err != nil {
				return nil, err
			}

		case OpSub:
			if err := vm.binaryOp(op); err != nil {
				return nil, err
			}

		case OpMul:
			if err := vm.binaryOp(op); err != nil {
				return nil, err
			}

		case OpDiv:
			if err := vm.binaryOp(op); err != nil {
				return nil, err
			}

		case OpMod:
			if err := vm.binaryOp(op); err != nil {
				return nil, err
			}

		case OpNeg:
			val := vm.pop()
			switch v := val.(type) {
			case *object.Integer:
				vm.push(&object.Integer{Value: -v.Value})
			case *object.Float:
				vm.push(&object.Float{Value: -v.Value})
			default:
				return nil, fmt.Errorf("cannot negate %s", val.Type())
			}

		case OpEq:
			b := vm.pop()
			a := vm.pop()
			vm.push(helpers.NativeBoolToBooleanObject(helpers.ObjectsEqual(a, b)))

		case OpNe:
			b := vm.pop()
			a := vm.pop()
			vm.push(helpers.NativeBoolToBooleanObject(!helpers.ObjectsEqual(a, b)))

		case OpLt:
			if err := vm.comparisonOp(op); err != nil {
				return nil, err
			}

		case OpLe:
			if err := vm.comparisonOp(op); err != nil {
				return nil, err
			}

		case OpGt:
			if err := vm.comparisonOp(op); err != nil {
				return nil, err
			}

		case OpGe:
			if err := vm.comparisonOp(op); err != nil {
				return nil, err
			}

		case OpNot:
			val := vm.pop()
			vm.push(helpers.NativeBoolToBooleanObject(!isTruthy(val)))

		case OpLoadLocal:
			slot := int(chunk.ReadUint16(frame.ip))
			frame.ip += 2
			vm.push(vm.stack[frame.slots+slot])

		case OpStoreLocal:
			slot := int(chunk.ReadUint16(frame.ip))
			frame.ip += 2
			vm.stack[frame.slots+slot] = vm.peek(0)

		case OpLoadGlobal:
			nameIdx := chunk.ReadUint16(frame.ip)
			frame.ip += 2
			name := chunk.Names[nameIdx]

			// Check builtins first
			if builtin, ok := vm.builtins[name]; ok {
				vm.push(builtin)
				continue
			}

			// Then check globals
			if val, ok := vm.globals[name]; ok {
				vm.push(val)
				continue
			}

			return nil, fmt.Errorf("undefined variable: %s", name)

		case OpStoreGlobal:
			nameIdx := chunk.ReadUint16(frame.ip)
			frame.ip += 2
			name := chunk.Names[nameIdx]
			vm.globals[name] = vm.peek(0)

		case OpLoadUpval:
			slot := int(chunk.ReadUint8(frame.ip))
			frame.ip++
			vm.push(*frame.closure.Upvalues[slot].Location)

		case OpStoreUpval:
			slot := int(chunk.ReadUint8(frame.ip))
			frame.ip++
			*frame.closure.Upvalues[slot].Location = vm.peek(0)

		case OpJump:
			offset := int(chunk.ReadUint16(frame.ip))
			frame.ip += offset + 2

		case OpJumpIfFalse:
			offset := int(chunk.ReadUint16(frame.ip))
			frame.ip += 2
			if !isTruthy(vm.peek(0)) {
				frame.ip += offset
			}

		case OpJumpIfTrue:
			offset := int(chunk.ReadUint16(frame.ip))
			frame.ip += 2
			if isTruthy(vm.peek(0)) {
				frame.ip += offset
			}

		case OpLoop:
			offset := int(chunk.ReadUint16(frame.ip))
			frame.ip += 2
			frame.ip -= offset

		case OpCall:
			argCount := int(chunk.ReadUint8(frame.ip))
			frame.ip++
			if err := vm.callValue(vm.peek(argCount), argCount); err != nil {
				return nil, err
			}

			// Check if a builtin returned a control flow wrapper
			if vm.stackTop > 0 {
				top := vm.peek(0)
				if returnVal, ok := top.(*object.ReturnValue); ok {
					// Pop the ReturnValue and trigger a return
					vm.pop()

					// Store result in memoization cache if this is a memoized function
					if frame.closure.Cache != nil {
						arity := frame.closure.Function.Arity
						args := make([]object.Object, arity)
						for i := 0; i < arity; i++ {
							args[i] = vm.stack[frame.slots+1+i]
						}
						frame.closure.Cache.Set(args, returnVal.Value)
					}

					// Close upvalues for current frame
					vm.closeUpvalues(&vm.stack[frame.slots])

					vm.frameCount--
					if vm.frameCount == 0 {
						vm.pop() // Pop the script function
						return returnVal.Value, nil
					}

					// Pop all values including the function
					vm.stackTop = frame.slots
					vm.push(returnVal.Value)

					// Update frame pointer after return
					continue
				}

				// Handle RepeatValue for repeat()/repeat?() builtins
				if repeatVal, ok := top.(*object.RepeatValue); ok {
					vm.pop() // Remove the RepeatValue

					// Get the current closure from the frame
					closure := frame.closure

					// Determine new args
					var newArgs []object.Object
					if repeatVal.Args == nil {
						// Use current args (they're already on the stack at frame.slots)
						newArgs = make([]object.Object, closure.Function.Arity)
						for i := 0; i < closure.Function.Arity; i++ {
							newArgs[i] = vm.stack[frame.slots+1+i] // +1 for the closure itself
						}
					} else {
						newArgs = repeatVal.Args
					}

					// Reset IP to start of function
					frame.ip = 0
					// Reset stack to just have the closure and args
					vm.stackTop = frame.slots + 1 + closure.Function.Arity

					// Update args in place
					for i, arg := range newArgs {
						if i < closure.Function.Arity {
							vm.stack[frame.slots+1+i] = arg
						}
					}

					continue
				}
			}

		case OpReturn:
			result := vm.pop()

			// Store result in memoization cache if this is a memoized function
			if frame.closure.Cache != nil {
				// Collect arguments from stack (they're at frame.slots+1 onwards)
				arity := frame.closure.Function.Arity
				args := make([]object.Object, arity)
				for i := 0; i < arity; i++ {
					args[i] = vm.stack[frame.slots+1+i]
				}
				frame.closure.Cache.Set(args, result)
			}

			// Close upvalues
			vm.closeUpvalues(&vm.stack[frame.slots])

			vm.frameCount--
			if vm.frameCount == 0 {
				vm.pop() // Pop the script function
				return result, nil
			}

			// Pop all values including the function
			vm.stackTop = frame.slots
			vm.push(result)

		case OpClosure:
			fnIdx := chunk.ReadUint16(frame.ip)
			frame.ip += 2
			fn := chunk.Constants[fnIdx].(*CompiledFunction)

			closure := &ObjClosure{
				Function: fn,
				Upvalues: make([]*ObjUpvalue, fn.UpvalueCount),
			}

			// Initialize memoization cache for memoized functions
			if fn.IsMemoized {
				closure.Cache = object.NewMemoCache()
			}

			for i := 0; i < fn.UpvalueCount; i++ {
				isLocal := chunk.ReadUint8(frame.ip) == 1
				frame.ip++
				index := int(chunk.ReadUint8(frame.ip))
				frame.ip++

				if isLocal {
					closure.Upvalues[i] = vm.captureUpvalue(&vm.stack[frame.slots+index])
				} else {
					closure.Upvalues[i] = frame.closure.Upvalues[index]
				}
			}

			vm.push(closure)

		case OpCloseUpval:
			vm.closeUpvalues(&vm.stack[vm.stackTop-1])
			vm.pop()

		case OpUnpackCall:
			// Call function, unpacking tuple if the argument is a tuple
			// Stack: [arg, function] where function is on top
			// expectedArity tells us how many args the function expects
			expectedArity := int(chunk.ReadUint8(frame.ip))
			frame.ip++

			// Stack has: [arg, function] (function is on top)
			fn := vm.pop()  // Get the function (on top)
			arg := vm.pop() // Get the argument (left value)

			// Check if arg is a tuple that should be unpacked
			if tuple, ok := arg.(*object.Tuple); ok {
				// Validate that tuple length matches expected arity
				if len(tuple.Elements) != expectedArity {
					return nil, fmt.Errorf("wrong number of arguments: expected %d, got %d",
						expectedArity, len(tuple.Elements))
				}
				// Push function, then all tuple elements as separate args
				vm.push(fn)
				for _, elem := range tuple.Elements {
					vm.push(elem)
				}
				if err := vm.callValue(fn, len(tuple.Elements)); err != nil {
					return nil, err
				}
			} else {
				// Single value - call function with one argument
				if expectedArity != 1 {
					return nil, fmt.Errorf("wrong number of arguments: expected %d, got 1",
						expectedArity)
				}
				vm.push(fn)
				vm.push(arg)
				if err := vm.callValue(fn, 1); err != nil {
					return nil, err
				}
			}

			// Handle control flow wrappers (same as OpCall)
			if vm.stackTop > 0 {
				top := vm.peek(0)
				if returnVal, ok := top.(*object.ReturnValue); ok {
					vm.pop()

					// Store result in memoization cache if this is a memoized function
					if frame.closure.Cache != nil {
						arity := frame.closure.Function.Arity
						args := make([]object.Object, arity)
						for i := 0; i < arity; i++ {
							args[i] = vm.stack[frame.slots+1+i]
						}
						frame.closure.Cache.Set(args, returnVal.Value)
					}

					vm.closeUpvalues(&vm.stack[frame.slots])
					vm.frameCount--
					if vm.frameCount == 0 {
						vm.pop()
						return returnVal.Value, nil
					}
					vm.stackTop = frame.slots
					vm.push(returnVal.Value)
					continue
				}

				if repeatVal, ok := top.(*object.RepeatValue); ok {
					vm.pop()
					closure := frame.closure
					var newArgs []object.Object
					if repeatVal.Args == nil {
						newArgs = make([]object.Object, closure.Function.Arity)
						for i := 0; i < closure.Function.Arity; i++ {
							newArgs[i] = vm.stack[frame.slots+1+i]
						}
					} else {
						newArgs = repeatVal.Args
					}
					frame.ip = 0
					// Reset stack to just have closure and args
					vm.stackTop = frame.slots + 1 + closure.Function.Arity
					for i, arg := range newArgs {
						if i < closure.Function.Arity {
							vm.stack[frame.slots+1+i] = arg
						}
					}
					continue
				}
			}

		case OpArray:
			count := int(chunk.ReadUint16(frame.ip))
			frame.ip += 2
			elements := make([]object.Object, count)
			for i := count - 1; i >= 0; i-- {
				elements[i] = vm.pop()
			}
			vm.push(&object.Array{Elements: elements})

		case OpMap:
			count := int(chunk.ReadUint16(frame.ip))
			frame.ip += 2
			pairs := make(map[string]object.Object)
			keys := make([]string, count)
			for i := count - 1; i >= 0; i-- {
				value := vm.pop()
				key := vm.pop()
				keyStr := key.Inspect()
				pairs[keyStr] = value
				keys[i] = keyStr
			}
			vm.push(&object.Map{Pairs: pairs, Keys: keys})

		case OpIndexGet:
			index := vm.pop()
			collection := vm.pop()
			result, err := vm.indexGet(collection, index)
			if err != nil {
				return nil, err
			}
			vm.push(result)

		case OpIndexSet:
			value := vm.pop()
			index := vm.pop()
			collection := vm.pop()
			result, err := vm.indexSet(collection, index, value)
			if err != nil {
				return nil, err
			}
			vm.push(result)

		case OpLinkBind:
			nameIdx := chunk.ReadUint16(frame.ip)
			frame.ip += 2
			name := chunk.Names[nameIdx]

			// Mark as shared for COW
			val := vm.peek(0)
			if arr, ok := val.(*object.Array); ok {
				arr.Share()
			} else if m, ok := val.(*object.Map); ok {
				m.Share()
			}

			vm.globals[name] = val

		case OpLinkCall:
			extraArgCount := int(chunk.ReadUint8(frame.ip))
			frame.ip++
			// The function is on top, then extra args, then the left value
			// Reorder so left value is first argument, with tuple unpacking
			fn := vm.pop()
			extraArgs := make([]object.Object, extraArgCount)
			for i := extraArgCount - 1; i >= 0; i-- {
				extraArgs[i] = vm.pop()
			}
			left := vm.pop()

			// Build final args: if left is tuple, unpack it; otherwise use single value
			var allArgs []object.Object
			if tuple, ok := left.(*object.Tuple); ok {
				// Unpack tuple elements and append extra args
				allArgs = make([]object.Object, len(tuple.Elements)+extraArgCount)
				copy(allArgs, tuple.Elements)
				copy(allArgs[len(tuple.Elements):], extraArgs)
			} else {
				// Single value prepended to extra args
				allArgs = make([]object.Object, 1+extraArgCount)
				allArgs[0] = left
				copy(allArgs[1:], extraArgs)
			}

			// Push function, then all args
			vm.push(fn)
			for _, arg := range allArgs {
				vm.push(arg)
			}

			if err := vm.callValue(fn, len(allArgs)); err != nil {
				return nil, err
			}

			// Check for control flow wrappers after link call
			if vm.stackTop > 0 {
				top := vm.peek(0)
				if returnVal, ok := top.(*object.ReturnValue); ok {
					vm.pop()

					// Store result in memoization cache if this is a memoized function
					if frame.closure.Cache != nil {
						arity := frame.closure.Function.Arity
						args := make([]object.Object, arity)
						for i := 0; i < arity; i++ {
							args[i] = vm.stack[frame.slots+1+i]
						}
						frame.closure.Cache.Set(args, returnVal.Value)
					}

					vm.closeUpvalues(&vm.stack[frame.slots])
					vm.frameCount--
					if vm.frameCount == 0 {
						vm.pop()
						return returnVal.Value, nil
					}
					vm.stackTop = frame.slots
					vm.push(returnVal.Value)
					continue
				}

				if repeatVal, ok := top.(*object.RepeatValue); ok {
					vm.pop()
					closure := frame.closure
					var newArgs []object.Object
					if repeatVal.Args == nil {
						newArgs = make([]object.Object, closure.Function.Arity)
						for i := 0; i < closure.Function.Arity; i++ {
							newArgs[i] = vm.stack[frame.slots+1+i]
						}
					} else {
						newArgs = repeatVal.Args
					}
					frame.ip = 0
					// Reset stack to just have closure and args
					vm.stackTop = frame.slots + 1 + closure.Function.Arity
					for i, arg := range newArgs {
						if i < closure.Function.Arity {
							vm.stack[frame.slots+1+i] = arg
						}
					}
					continue
				}
			}

		case OpMember:
			nameIdx := chunk.ReadUint16(frame.ip)
			frame.ip += 2
			name := chunk.Names[nameIdx]

			obj := vm.pop()
			if m, ok := obj.(*object.Map); ok {
				if val, ok := m.Pairs[name]; ok {
					vm.push(val)
				} else {
					vm.push(helpers.NULL)
				}
			} else {
				return nil, fmt.Errorf("cannot access member %s of %s", name, obj.Type())
			}

		case OpTuple:
			count := int(chunk.ReadUint8(frame.ip))
			frame.ip++
			elements := make([]object.Object, count)
			for i := count - 1; i >= 0; i-- {
				elements[i] = vm.pop()
			}
			vm.push(&object.Tuple{Elements: elements})

		case OpCapture:
			errIdx := chunk.ReadUint16(frame.ip)
			frame.ip += 2
			valIdx := chunk.ReadUint16(frame.ip)
			frame.ip += 2

			errName := chunk.Names[errIdx]
			valName := chunk.Names[valIdx]

			tuple := vm.pop()
			if t, ok := tuple.(*object.Tuple); ok && len(t.Elements) >= 2 {
				err := t.Elements[0]
				val := t.Elements[1]

				vm.globals[errName] = err
				vm.globals[valName] = val

				// Check if error is present (non-empty map)
				if isError(err) {
					vm.lastError = err
					vm.push(&object.CaptureStop{})
				} else {
					vm.push(val)
				}
			} else {
				return nil, fmt.Errorf("capture expects (error, value) tuple")
			}

		case OpChainStop:
			// Check if last result is a chain stop signal
			val := vm.peek(0)
			if _, ok := val.(*object.ChainStop); ok {
				// Chain should stop
				vm.push(helpers.NULL)
			}
			if _, ok := val.(*object.CaptureStop); ok {
				// Chain should stop
				vm.push(helpers.NULL)
			}

		case OpMemoCheck:
			// Memoization check - for memoized functions
			// This is handled at the function call level
			frame.ip++ // Skip arg count

		case OpMemoStore:
			// Store result in memo cache
			// This is handled at the function return level

		case OpBuiltin:
			nameIdx := chunk.ReadUint16(frame.ip)
			frame.ip += 2
			argCount := int(chunk.ReadUint8(frame.ip))
			frame.ip++
			name := chunk.Names[nameIdx]

			// Get arguments
			args := make([]object.Object, argCount)
			for i := argCount - 1; i >= 0; i-- {
				args[i] = vm.pop()
			}

			// Look up builtin
			if builtin, ok := vm.builtins[name]; ok {
				result := builtin.Fn(args...)

				// Handle control flow wrapper objects
				result = vm.handleControlFlowResult(result)
				vm.push(result)

				// Check for control flow wrappers that need special handling
				if returnVal, ok := result.(*object.ReturnValue); ok {
					vm.pop()

					// Store result in memoization cache if this is a memoized function
					if frame.closure.Cache != nil {
						arity := frame.closure.Function.Arity
						args := make([]object.Object, arity)
						for i := 0; i < arity; i++ {
							args[i] = vm.stack[frame.slots+1+i]
						}
						frame.closure.Cache.Set(args, returnVal.Value)
					}

					vm.closeUpvalues(&vm.stack[frame.slots])
					vm.frameCount--
					if vm.frameCount == 0 {
						vm.pop()
						return returnVal.Value, nil
					}
					vm.stackTop = frame.slots
					vm.push(returnVal.Value)
					continue
				}

				if repeatVal, ok := result.(*object.RepeatValue); ok {
					vm.pop()
					closure := frame.closure
					var newArgs []object.Object
					if repeatVal.Args == nil {
						newArgs = make([]object.Object, closure.Function.Arity)
						for i := 0; i < closure.Function.Arity; i++ {
							newArgs[i] = vm.stack[frame.slots+1+i]
						}
					} else {
						newArgs = repeatVal.Args
					}
					// Reset instruction pointer to start of function
					frame.ip = 0
					// Reset stack to just have the closure and args
					// This cleans up any temporaries from the previous iteration
					vm.stackTop = frame.slots + 1 + closure.Function.Arity
					// Update args with new values
					for i, arg := range newArgs {
						if i < closure.Function.Arity {
							vm.stack[frame.slots+1+i] = arg
						}
					}
					continue
				}
			} else {
				// Try as global function
				if fn, ok := vm.globals[name]; ok {
					// Check if this is a constant value (optimized constant function)
					// If so, just push the value - no function call needed
					if _, isClosure := fn.(*ObjClosure); !isClosure && argCount == 0 {
						vm.push(fn)
						continue
					}

					vm.push(fn)
					for _, arg := range args {
						vm.push(arg)
					}
					if err := vm.callValue(fn, argCount); err != nil {
						return nil, err
					}

					// Check for control flow wrappers after user function call
					if vm.stackTop > 0 {
						top := vm.peek(0)
						if returnVal, ok := top.(*object.ReturnValue); ok {
							vm.pop()

							// Store result in memoization cache if this is a memoized function
							if frame.closure.Cache != nil {
								arity := frame.closure.Function.Arity
								cacheArgs := make([]object.Object, arity)
								for i := 0; i < arity; i++ {
									cacheArgs[i] = vm.stack[frame.slots+1+i]
								}
								frame.closure.Cache.Set(cacheArgs, returnVal.Value)
							}

							vm.closeUpvalues(&vm.stack[frame.slots])
							vm.frameCount--
							if vm.frameCount == 0 {
								vm.pop()
								return returnVal.Value, nil
							}
							vm.stackTop = frame.slots
							vm.push(returnVal.Value)
							continue
						}

						if repeatVal, ok := top.(*object.RepeatValue); ok {
							vm.pop()
							closure := frame.closure
							var newArgs []object.Object
							if repeatVal.Args == nil {
								newArgs = make([]object.Object, closure.Function.Arity)
								for i := 0; i < closure.Function.Arity; i++ {
									newArgs[i] = vm.stack[frame.slots+1+i]
								}
							} else {
								newArgs = repeatVal.Args
							}
							frame.ip = 0
							vm.stackTop = frame.slots + 1 + closure.Function.Arity
							for i, arg := range newArgs {
								if i < closure.Function.Arity {
									vm.stack[frame.slots+1+i] = arg
								}
							}
							continue
						}
					}
				} else {
					return nil, fmt.Errorf("undefined function: %s", name)
				}
			}

		case OpUnpackBuiltin:
			// Similar to OpBuiltin but handles tuple unpacking for link calls
			// Stack: [left_value, extra_args...] where left_value may be a tuple
			nameIdx := chunk.ReadUint16(frame.ip)
			frame.ip += 2
			extraArgCount := int(chunk.ReadUint8(frame.ip))
			frame.ip++
			name := chunk.Names[nameIdx]

			// Pop extra arguments first (they're on top)
			extraArgs := make([]object.Object, extraArgCount)
			for i := extraArgCount - 1; i >= 0; i-- {
				extraArgs[i] = vm.pop()
			}

			// Pop left value (which may be a tuple)
			left := vm.pop()

			// Build final args: if left is tuple, unpack it; otherwise use single value
			var allArgs []object.Object
			if tuple, ok := left.(*object.Tuple); ok {
				// Unpack tuple elements and append extra args
				allArgs = make([]object.Object, len(tuple.Elements)+extraArgCount)
				copy(allArgs, tuple.Elements)
				copy(allArgs[len(tuple.Elements):], extraArgs)
			} else {
				// Single value prepended to extra args
				allArgs = make([]object.Object, 1+extraArgCount)
				allArgs[0] = left
				copy(allArgs[1:], extraArgs)
			}

			// Look up builtin
			if builtin, ok := vm.builtins[name]; ok {
				result := builtin.Fn(allArgs...)

				// Handle control flow wrapper objects
				result = vm.handleControlFlowResult(result)
				vm.push(result)

				// Check for control flow wrappers that need special handling
				if returnVal, ok := result.(*object.ReturnValue); ok {
					vm.pop()

					// Store result in memoization cache if this is a memoized function
					if frame.closure.Cache != nil {
						arity := frame.closure.Function.Arity
						cacheArgs := make([]object.Object, arity)
						for i := 0; i < arity; i++ {
							cacheArgs[i] = vm.stack[frame.slots+1+i]
						}
						frame.closure.Cache.Set(cacheArgs, returnVal.Value)
					}

					vm.closeUpvalues(&vm.stack[frame.slots])
					vm.frameCount--
					if vm.frameCount == 0 {
						vm.pop()
						return returnVal.Value, nil
					}
					vm.stackTop = frame.slots
					vm.push(returnVal.Value)
					continue
				}

				if repeatVal, ok := result.(*object.RepeatValue); ok {
					vm.pop()
					closure := frame.closure
					var newArgs []object.Object
					if repeatVal.Args == nil {
						newArgs = make([]object.Object, closure.Function.Arity)
						for i := 0; i < closure.Function.Arity; i++ {
							newArgs[i] = vm.stack[frame.slots+1+i]
						}
					} else {
						newArgs = repeatVal.Args
					}
					frame.ip = 0
					vm.stackTop = frame.slots + 1 + closure.Function.Arity
					for i, arg := range newArgs {
						if i < closure.Function.Arity {
							vm.stack[frame.slots+1+i] = arg
						}
					}
					continue
				}
			} else {
				// Try as global function (user-defined)
				if fn, ok := vm.globals[name]; ok {
					vm.push(fn)
					for _, arg := range allArgs {
						vm.push(arg)
					}
					if err := vm.callValue(fn, len(allArgs)); err != nil {
						return nil, err
					}

					// Check for control flow wrappers after user function call
					if vm.stackTop > 0 {
						top := vm.peek(0)
						if returnVal, ok := top.(*object.ReturnValue); ok {
							vm.pop()

							// Store result in memoization cache if this is a memoized function
							if frame.closure.Cache != nil {
								arity := frame.closure.Function.Arity
								cacheArgs := make([]object.Object, arity)
								for i := 0; i < arity; i++ {
									cacheArgs[i] = vm.stack[frame.slots+1+i]
								}
								frame.closure.Cache.Set(cacheArgs, returnVal.Value)
							}

							vm.closeUpvalues(&vm.stack[frame.slots])
							vm.frameCount--
							if vm.frameCount == 0 {
								vm.pop()
								return returnVal.Value, nil
							}
							vm.stackTop = frame.slots
							vm.push(returnVal.Value)
							continue
						}

						if repeatVal, ok := top.(*object.RepeatValue); ok {
							vm.pop()
							closure := frame.closure
							var newArgs []object.Object
							if repeatVal.Args == nil {
								newArgs = make([]object.Object, closure.Function.Arity)
								for i := 0; i < closure.Function.Arity; i++ {
									newArgs[i] = vm.stack[frame.slots+1+i]
								}
							} else {
								newArgs = repeatVal.Args
							}
							frame.ip = 0
							vm.stackTop = frame.slots + 1 + closure.Function.Arity
							for i, arg := range newArgs {
								if i < closure.Function.Arity {
									vm.stack[frame.slots+1+i] = arg
								}
							}
							continue
						}
					}
				} else {
					return nil, fmt.Errorf("undefined function: %s", name)
				}
			}

		case OpInterpolate:
			// Get string constant to interpolate
			idx := chunk.ReadUint16(frame.ip)
			frame.ip += 2
			strObj := chunk.Constants[idx].(*object.String)

			// Interpolate the string using current variable scope
			result, err := vm.interpolateString(strObj.Value, frame, chunk)
			if err != nil {
				return nil, err
			}
			vm.push(&object.String{Value: result})

		case OpPop:
			vm.pop()

		case OpDup:
			vm.push(vm.peek(0))

		case OpSwap:
			a := vm.pop()
			b := vm.pop()
			vm.push(a)
			vm.push(b)

		default:
			return nil, fmt.Errorf("unknown opcode: %d", op)
		}
	}
}

// Stack operations

func (vm *VM) push(obj object.Object) {
	if vm.stackTop >= len(vm.stack) {
		// Grow stack
		if len(vm.stack) >= MaxStackSize {
			panic("stack overflow")
		}
		newSize := len(vm.stack) * 2
		if newSize > MaxStackSize {
			newSize = MaxStackSize
		}
		newStack := make([]object.Object, newSize)
		copy(newStack, vm.stack)
		vm.stack = newStack
	}
	vm.stack[vm.stackTop] = obj
	vm.stackTop++
}

func (vm *VM) pop() object.Object {
	if vm.stackTop == 0 {
		panic("stack underflow")
	}
	vm.stackTop--
	return vm.stack[vm.stackTop]
}

func (vm *VM) peek(distance int) object.Object {
	return vm.stack[vm.stackTop-1-distance]
}

// Binary operations

func (vm *VM) binaryOp(op OpCode) error {
	b := vm.pop()
	a := vm.pop()

	// Handle numeric operations
	switch av := a.(type) {
	case *object.Integer:
		switch bv := b.(type) {
		case *object.Integer:
			vm.push(vm.integerOp(op, av.Value, bv.Value))
			return nil
		case *object.Float:
			vm.push(vm.floatOp(op, float64(av.Value), bv.Value))
			return nil
		}
	case *object.Float:
		switch bv := b.(type) {
		case *object.Integer:
			vm.push(vm.floatOp(op, av.Value, float64(bv.Value)))
			return nil
		case *object.Float:
			vm.push(vm.floatOp(op, av.Value, bv.Value))
			return nil
		}
	case *object.String:
		if op == OpAdd {
			if bv, ok := b.(*object.String); ok {
				vm.push(&object.String{Value: av.Value + bv.Value})
				return nil
			}
		}
	}

	return fmt.Errorf("cannot perform %s on %s and %s", op.String(), a.Type(), b.Type())
}

func (vm *VM) integerOp(op OpCode, a, b int64) object.Object {
	switch op {
	case OpAdd:
		return &object.Integer{Value: a + b}
	case OpSub:
		return &object.Integer{Value: a - b}
	case OpMul:
		return &object.Integer{Value: a * b}
	case OpDiv:
		if b == 0 {
			return helpers.NewError("division by zero")
		}
		return &object.Integer{Value: a / b}
	case OpMod:
		if b == 0 {
			return helpers.NewError("modulo by zero")
		}
		return &object.Integer{Value: a % b}
	default:
		return helpers.NULL
	}
}

func (vm *VM) floatOp(op OpCode, a, b float64) object.Object {
	switch op {
	case OpAdd:
		return &object.Float{Value: a + b}
	case OpSub:
		return &object.Float{Value: a - b}
	case OpMul:
		return &object.Float{Value: a * b}
	case OpDiv:
		if b == 0 {
			return helpers.NewError("division by zero")
		}
		return &object.Float{Value: a / b}
	case OpMod:
		return helpers.NewError("modulo not supported for floats")
	default:
		return helpers.NULL
	}
}

// Comparison operations

func (vm *VM) comparisonOp(op OpCode) error {
	b := vm.pop()
	a := vm.pop()

	var result bool

	switch av := a.(type) {
	case *object.Integer:
		switch bv := b.(type) {
		case *object.Integer:
			result = vm.compareInt(op, av.Value, bv.Value)
		case *object.Float:
			result = vm.compareFloat(op, float64(av.Value), bv.Value)
		default:
			return fmt.Errorf("cannot compare %s and %s", a.Type(), b.Type())
		}
	case *object.Float:
		switch bv := b.(type) {
		case *object.Integer:
			result = vm.compareFloat(op, av.Value, float64(bv.Value))
		case *object.Float:
			result = vm.compareFloat(op, av.Value, bv.Value)
		default:
			return fmt.Errorf("cannot compare %s and %s", a.Type(), b.Type())
		}
	case *object.String:
		if bv, ok := b.(*object.String); ok {
			result = vm.compareString(op, av.Value, bv.Value)
		} else {
			return fmt.Errorf("cannot compare %s and %s", a.Type(), b.Type())
		}
	default:
		return fmt.Errorf("cannot compare %s and %s", a.Type(), b.Type())
	}

	vm.push(helpers.NativeBoolToBooleanObject(result))
	return nil
}

func (vm *VM) compareInt(op OpCode, a, b int64) bool {
	switch op {
	case OpLt:
		return a < b
	case OpLe:
		return a <= b
	case OpGt:
		return a > b
	case OpGe:
		return a >= b
	default:
		return false
	}
}

func (vm *VM) compareFloat(op OpCode, a, b float64) bool {
	switch op {
	case OpLt:
		return a < b
	case OpLe:
		return a <= b
	case OpGt:
		return a > b
	case OpGe:
		return a >= b
	default:
		return false
	}
}

func (vm *VM) compareString(op OpCode, a, b string) bool {
	switch op {
	case OpLt:
		return a < b
	case OpLe:
		return a <= b
	case OpGt:
		return a > b
	case OpGe:
		return a >= b
	default:
		return false
	}
}

// Function calls

func (vm *VM) callValue(callee object.Object, argCount int) error {
	switch fn := callee.(type) {
	case *ObjClosure:
		return vm.call(fn, argCount)
	case *object.Builtin:
		args := make([]object.Object, argCount)
		for i := argCount - 1; i >= 0; i-- {
			args[i] = vm.pop()
		}
		vm.pop() // Pop the builtin
		result := fn.Fn(args...)

		// Handle control flow wrapper objects
		result = vm.handleControlFlowResult(result)

		vm.push(result)
		return nil
	default:
		return fmt.Errorf("can only call functions and closures, got %T", callee)
	}
}

// handleControlFlowResult handles control flow builtins that return wrapper objects.
// For builtin calls at the top level (not inside a user function), we need to unwrap
// ReturnValue since there's no function to return from.
// RepeatValue and ChainStop are left as-is for the caller to handle.
func (vm *VM) handleControlFlowResult(result object.Object) object.Object {
	// Check for ReturnValue wrapper from return() or return?() builtins
	if returnVal, ok := result.(*object.ReturnValue); ok {
		// If we're at the top level (only script frame), just unwrap
		// If inside a function call, trigger a return
		if vm.frameCount <= 1 {
			return returnVal.Value
		}
		// Inside a function - we need to signal return to the dispatch loop
		// We'll return the ReturnValue and let the execute loop handle it
		return result
	}
	return result
}

func (vm *VM) call(closure *ObjClosure, argCount int) error {
	if argCount != closure.Function.Arity {
		return fmt.Errorf("expected %d arguments but got %d", closure.Function.Arity, argCount)
	}

	// Check memoization cache before executing
	if closure.Cache != nil {
		// Collect arguments from stack
		args := make([]object.Object, argCount)
		for i := 0; i < argCount; i++ {
			args[i] = vm.stack[vm.stackTop-argCount+i]
		}

		// Check cache
		if cached, ok := closure.Cache.Get(args); ok {
			// Cache hit - pop the function and arguments, push result
			vm.stackTop -= argCount + 1 // Pop args and closure
			vm.push(cached)
			return nil
		}
	}

	if vm.frameCount >= len(vm.frames) {
		// Grow frames
		if len(vm.frames) >= MaxFramesSize {
			return fmt.Errorf("stack overflow")
		}
		newSize := len(vm.frames) * 2
		if newSize > MaxFramesSize {
			newSize = MaxFramesSize
		}
		newFrames := make([]CallFrame, newSize)
		copy(newFrames, vm.frames)
		vm.frames = newFrames
	}

	frame := &vm.frames[vm.frameCount]
	vm.frameCount++
	frame.closure = closure
	frame.ip = 0
	frame.slots = vm.stackTop - argCount - 1

	return nil
}

// Upvalue handling

func (vm *VM) captureUpvalue(local *object.Object) *ObjUpvalue {
	// Search for existing open upvalue
	// We compare stack positions using uintptr since Go doesn't allow > on pointers
	var prevUpvalue *ObjUpvalue
	upvalue := vm.openUpvals
	localAddr := uintptr(unsafe.Pointer(local))

	for upvalue != nil && uintptr(unsafe.Pointer(upvalue.Location)) > localAddr {
		prevUpvalue = upvalue
		upvalue = upvalue.Next
	}

	if upvalue != nil && upvalue.Location == local {
		return upvalue
	}

	// Create new upvalue
	createdUpvalue := &ObjUpvalue{
		Location: local,
		Next:     upvalue,
	}

	if prevUpvalue == nil {
		vm.openUpvals = createdUpvalue
	} else {
		prevUpvalue.Next = createdUpvalue
	}

	return createdUpvalue
}

func (vm *VM) closeUpvalues(last *object.Object) {
	lastAddr := uintptr(unsafe.Pointer(last))
	for vm.openUpvals != nil && uintptr(unsafe.Pointer(vm.openUpvals.Location)) >= lastAddr {
		upvalue := vm.openUpvals
		upvalue.Closed = *upvalue.Location
		upvalue.Location = &upvalue.Closed
		vm.openUpvals = upvalue.Next
	}
}

// Index operations

func (vm *VM) indexGet(collection, index object.Object) (object.Object, error) {
	switch coll := collection.(type) {
	case *object.Array:
		idx, ok := index.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf("array index must be integer, got %s", index.Type())
		}
		if idx.Value < 0 || idx.Value >= int64(len(coll.Elements)) {
			return helpers.NULL, nil
		}
		return coll.Elements[idx.Value], nil

	case *object.Map:
		key := index.Inspect()
		if val, ok := coll.Pairs[key]; ok {
			return val, nil
		}
		return helpers.NULL, nil

	case *object.String:
		idx, ok := index.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf("string index must be integer, got %s", index.Type())
		}
		runes := []rune(coll.Value)
		if idx.Value < 0 || idx.Value >= int64(len(runes)) {
			return helpers.NULL, nil
		}
		return &object.String{Value: string(runes[idx.Value])}, nil

	default:
		return nil, fmt.Errorf("index operation not supported on %s", collection.Type())
	}
}

func (vm *VM) indexSet(collection, index, value object.Object) (object.Object, error) {
	switch coll := collection.(type) {
	case *object.Array:
		idx, ok := index.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf("array index must be integer, got %s", index.Type())
		}
		newArr := coll.COWSet(int(idx.Value), value)
		return newArr, nil

	case *object.Map:
		key := index.Inspect()
		newMap := coll.COWSet(key, value)
		return newMap, nil

	default:
		return nil, fmt.Errorf("index set operation not supported on %s", collection.Type())
	}
}

// Helper functions

func isTruthy(obj object.Object) bool {
	switch v := obj.(type) {
	case *object.Boolean:
		return v.Value
	case *object.Null:
		return false
	case *object.Integer:
		return v.Value != 0
	case *object.Float:
		return v.Value != 0
	case *object.String:
		return len(v.Value) > 0
	case *object.Array:
		return len(v.Elements) > 0
	case *object.Map:
		return len(v.Pairs) > 0
	default:
		return true
	}
}

func isError(obj object.Object) bool {
	if obj == nil {
		return false
	}
	if obj.Type() == object.ERROR_OBJ {
		return true
	}
	// Check if it's a non-empty map (bark error representation)
	if m, ok := obj.(*object.Map); ok {
		return len(m.Pairs) > 0
	}
	return false
}

// Type definitions for object compatibility

// Implement object.Object interface for ObjClosure
func (c *ObjClosure) Type() object.ObjectType { return "CLOSURE" }
func (c *ObjClosure) Inspect() string {
	return fmt.Sprintf("<closure %s>", c.Function.Name)
}

// interpolateString processes string interpolation with {identifier} and {identifier.field} patterns.
// It looks up variables in the current VM scope and replaces the placeholders with their values.
func (vm *VM) interpolateString(input string, frame *CallFrame, chunk *Chunk) (string, error) {
	var result strings.Builder
	i := 0

	for i < len(input) {
		// Check for escaped braces from lexer (\{ and \})
		if i+1 < len(input) && input[i] == '\\' {
			if input[i+1] == '{' {
				result.WriteByte('{')
				i += 2
				continue
			}
			if input[i+1] == '}' {
				result.WriteByte('}')
				i += 2
				continue
			}
		}

		// Check for interpolation placeholder {identifier} or {identifier.field}
		if input[i] == '{' {
			// Find closing brace
			end := i + 1
			for end < len(input) && input[end] != '}' {
				end++
			}

			if end >= len(input) {
				// No closing brace found - pass through unchanged
				result.WriteByte(input[i])
				i++
				continue
			}

			// Get the content between braces
			content := input[i+1 : end]

			// Empty braces {} - pass through unchanged
			if content == "" {
				result.WriteString("{}")
				i = end + 1
				continue
			}

			// Check if it's a numeric index (positional placeholder) - pass through unchanged
			if isNumericVM(content) {
				result.WriteString(input[i : end+1])
				i = end + 1
				continue
			}

			// Parse identifier or identifier.field
			var identifier, field string
			if dotIdx := indexOfDot(content); dotIdx != -1 {
				identifier = content[:dotIdx]
				field = content[dotIdx+1:]
				// If either part is empty or invalid, pass through unchanged
				if identifier == "" || field == "" || !isValidIdentifierVM(identifier) || !isValidIdentifierVM(field) || indexOfDot(field) != -1 {
					result.WriteString(input[i : end+1])
					i = end + 1
					continue
				}
			} else {
				identifier = content
				// If not a valid identifier, pass through unchanged
				if !isValidIdentifierVM(identifier) {
					result.WriteString(input[i : end+1])
					i = end + 1
					continue
				}
			}

			// Look up the variable in the VM scope
			val, found := vm.lookupVariable(identifier, frame, chunk)
			if !found {
				// Variable not found - pass through unchanged
				result.WriteString(input[i : end+1])
				i = end + 1
				continue
			}

			// If field access is requested, look up the field in the map
			if field != "" {
				mapVal, ok := val.(*object.Map)
				if !ok {
					// Not a map - pass through unchanged
					result.WriteString(input[i : end+1])
					i = end + 1
					continue
				}
				fieldVal, ok := mapVal.Pairs[field]
				if !ok {
					// Field not found - pass through unchanged
					result.WriteString(input[i : end+1])
					i = end + 1
					continue
				}
				val = fieldVal
			}

			// Convert value to string
			result.WriteString(objectToStringVM(val))
			i = end + 1
			continue
		}

		// Regular character
		result.WriteByte(input[i])
		i++
	}

	return result.String(), nil
}

// lookupVariable looks up a variable by name in the VM scope
func (vm *VM) lookupVariable(name string, frame *CallFrame, chunk *Chunk) (object.Object, bool) {
	// Check locals first (search all frames from current to top-level)
	for f := vm.frameCount - 1; f >= 0; f-- {
		fr := &vm.frames[f]
		fn := fr.closure.Function

		// Check local variables in this frame
		for slot, localName := range fn.LocalNames {
			if localName == name {
				return vm.stack[fr.slots+slot], true
			}
		}
	}

	// Check globals
	if val, ok := vm.globals[name]; ok {
		return val, true
	}

	return nil, false
}

// Helper functions for VM string interpolation (to avoid conflicts with compiler helpers)

func isNumericVM(s string) bool {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func isValidIdentifierVM(s string) bool {
	if s == "" {
		return false
	}
	for i, ch := range s {
		if i == 0 {
			if !isLetterVM(ch) {
				return false
			}
		} else {
			if !isLetterVM(ch) && !isDigitVM(ch) {
				return false
			}
		}
	}
	return true
}

func isLetterVM(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == '?'
}

func isDigitVM(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func indexOfDot(s string) int {
	for i, ch := range s {
		if ch == '.' {
			return i
		}
	}
	return -1
}

func objectToStringVM(obj object.Object) string {
	switch v := obj.(type) {
	case *object.String:
		return v.Value
	default:
		return obj.Inspect()
	}
}
