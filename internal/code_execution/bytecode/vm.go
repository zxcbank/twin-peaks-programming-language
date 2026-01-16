package bytecode

import (
	"fmt"
	"math"
)

type Frame struct {
	locals   []Value
	returnIP int
	prevFP   int
	funcInfo *FunctionInfo
}

// ensureLocalsSize ensures the frame has at least `required` slots in locals.
func (f *Frame) ensureLocalsSize(required int) {
	if required <= len(f.locals) {
		return
	}
	needed := required - len(f.locals)
	f.locals = append(f.locals, make([]Value, needed)...)
}

type VM struct {
	bytecode   *Bytecode
	stack      []Value
	frames     []Frame
	heap       []*Array
	ip         int // Instruction Pointer
	sp         int // Stack Pointer
	fp         int // Frame Pointer (index into frames slice)
	gc         GarbageCollector
	jit        *JITCompiler
	jitEnabled bool
}

func (vm *VM) PrintHeapSize() {
	activeHeapElements := 0
	for _, array := range vm.heap {
		if array != nil {
			activeHeapElements++
		}
	}
	fmt.Println("Heap size:", activeHeapElements)
}

type Array struct {
	size  int
	Array []Value
}

func NewVM(bytecode *Bytecode, jitEnabled bool) *VM {
	return &VM{
		bytecode:   bytecode,
		stack:      make([]Value, 1024*1024*1024),
		heap:       make([]*Array, 0),
		frames:     make([]Frame, 1),
		ip:         bytecode.programStart,
		sp:         -1,
		fp:         0,
		gc:         GarbageCollector{},
		jit:        NewJITCompiler(bytecode),
		jitEnabled: jitEnabled,
	}
}

func (vm *VM) Run() error {
	for vm.ip < len(vm.bytecode.Instructions) {
		instr := vm.bytecode.Instructions[vm.ip]
		vm.ip++
		switch instr.Opcode {
		case OpConst:
			constIndex := instr.Operands[0]
			if constIndex >= len(vm.bytecode.Constants) {
				return fmt.Errorf("constant index out of bounds: %d", constIndex)
			}
			value := vm.bytecode.Constants[constIndex]
			vm.push(Value{Data: value})

		case OpLoad:
			localIndex := instr.Operands[0]
			if vm.fp < 0 || vm.fp >= len(vm.frames) {
				return fmt.Errorf("invalid frame pointer: %d", vm.fp)
			}
			currentFrame := &vm.frames[vm.fp]
			if localIndex >= len(currentFrame.locals) {
				return fmt.Errorf("local index out of bounds: %d", localIndex)
			}
			vm.push(currentFrame.locals[localIndex])

		case OpStore:
			localIndex := instr.Operands[0]
			if vm.sp < 0 {
				return fmt.Errorf("stack underflow")
			}
			value := vm.pop()
			if vm.fp < 0 || vm.fp >= len(vm.frames) {
				return fmt.Errorf("invalid frame pointer: %d", vm.fp)
			}
			currentFrame := &vm.frames[vm.fp]
			currentFrame.ensureLocalsSize(localIndex + 1)
			currentFrame.locals[localIndex] = value

		case OpPop:
			if vm.sp < 0 {
				return fmt.Errorf("stack underflow")
			}
			vm.pop()

		case OpAdd:
			if err := vm.binaryOp(func(a, b Value) Value {
				switch a.Data.(type) {
				case int:
					return Value{Data: a.Data.(int) + b.Data.(int)}
				case float64:
					return Value{Data: a.Data.(float64) + b.Data.(float64)}
				default:
					return Value{Data: 0}
				}
			}); err != nil {
				return err
			}

		case OpSub:
			if err := vm.binaryOp(func(a, b Value) Value {
				switch a.Data.(type) {
				case int:
					return Value{Data: a.Data.(int) - b.Data.(int)}
				case float64:
					return Value{Data: a.Data.(float64) - b.Data.(float64)}
				default:
					return Value{Data: 0}
				}
			}); err != nil {
				return err
			}

		case OpMul:
			if err := vm.binaryOp(func(a, b Value) Value {
				switch a.Data.(type) {
				case int:
					return Value{Data: a.Data.(int) * b.Data.(int)}
				case float64:
					//fmt.Print(a.Data, b.Data)
					return Value{Data: a.Data.(float64) * b.Data.(float64)}
				default:
					return Value{Data: 0}
				}
			}); err != nil {
				return err
			}

		case OpDiv:
			if err := vm.binaryOp(func(a, b Value) Value {
				switch a.Data.(type) {
				case int:
					bInt := b.Data.(int)
					if bInt == 0 {
						return Value{Data: 0}
					}
					return Value{Data: a.Data.(int) / bInt}
				case float64:
					bFloat := b.Data.(float64)
					if bFloat == 0 {
						return Value{Data: 0.0}
					}
					return Value{Data: a.Data.(float64) / bFloat}
				default:
					return Value{Data: 0} // TODO: think about error reporting
				}
			}); err != nil {
				return err
			}
		case OpMod:
			if err := vm.binaryOp(func(a, b Value) Value {
				switch a.Data.(type) {
				case int:
					bInt := b.Data.(int)
					if bInt == 0 {
						return Value{Data: 0}
					}
					return Value{Data: a.Data.(int) % bInt}
				default:
					return Value{Data: 0} // TODO: think about error reporting
				}
			}); err != nil {
				return err
			}
		case OpNeg:
			if vm.sp < 0 {
				return fmt.Errorf("stack underflow")
			}
			val := vm.pop()
			var negated Value
			switch val.Data.(type) {
			case int:
				negated = Value{Data: -val.Data.(int)}
			case float64:
				negated = Value{Data: -val.Data.(float64)}
			default:
				negated = val
			}
			vm.push(negated)
		case OpLt:
			if err := vm.binaryOp(func(a, b Value) Value { return valueLT(a, b) }); err != nil {
				return err
			}
		case OpLe:
			if err := vm.binaryOp(func(a, b Value) Value { return valueLE(a, b) }); err != nil {
				return err
			}
		case OpEq:
			if err := vm.binaryOp(func(a, b Value) Value { return valueEQ(a, b) }); err != nil {
				return err
			}
		case OpNeq:
			if err := vm.binaryOp(func(a, b Value) Value { return valueNEQ(a, b) }); err != nil {
				return err
			}
		case OpGt:
			if err := vm.binaryOp(func(a, b Value) Value { return valueGT(a, b) }); err != nil {
				return err
			}
		case OpGe:
			if err := vm.binaryOp(func(a, b Value) Value { return valueGE(a, b) }); err != nil {
				return err
			}
		case OpAnd:
			err := vm.binaryOp(func(a, b Value) Value {
				return Value{Data: isTruthy(a) && isTruthy(b)}
			})
			if err != nil {
				return err
			}
		case OpOr:
			err := vm.binaryOp(func(a, b Value) Value {
				return Value{Data: isTruthy(a) || isTruthy(b)}
			})
			if err != nil {
				return err
			}
		case OpNot:
			// Logical NOT: pop one value and push its negation
			if vm.sp < 0 {
				return fmt.Errorf("stack underflow")
			}
			val := vm.pop()
			vm.push(Value{Data: !isTruthy(val)})

		case OpPrint:
			if vm.sp < 0 {
				return fmt.Errorf("stack underflow")
			}
			value := vm.pop()
			fmt.Println(value.Data)

		case OpSqrt:
			if vm.sp < 0 {
				return fmt.Errorf("stack underflow")
			}
			value := vm.pop()
			switch v := value.Data.(type) {
			case int:
				result := math.Sqrt(float64(v))
				vm.push(Value{Data: result})
			case float64:
				result := math.Sqrt(v)
				vm.push(Value{Data: result})
			default:
				return fmt.Errorf("SQRT operation requires int or float64")
			}

		case OpHalt:
			return nil

		case OpJmp:
			vm.ip = instr.Operands[0]

		case OpJmpIfFalse:
			if vm.sp < 0 {
				return fmt.Errorf("stack underflow")
			}
			condition := vm.pop()
			if !isTruthy(condition) {
				vm.ip = instr.Operands[0]
			}

		case OpCall:
			funcAddr := instr.Operands[0]

			frame := Frame{
				returnIP: vm.ip,
				prevFP:   vm.fp,
				locals:   make([]Value, 0),
				funcInfo: vm.bytecode.FuncAddresses[funcAddr],
			}

			vm.frames = append(vm.frames, frame)
			vm.fp = len(vm.frames) - 1

			if vm.jitEnabled {
				callParams := make([]Value, frame.funcInfo.ParamCount)
				for i := 0; i < frame.funcInfo.ParamCount; i++ {
					callParams[i] = vm.stack[vm.sp-i]
				}
				if jitResult, newIP := vm.jit.GetCompiledAddress(vm.ip-1, callParams); jitResult == FuncJITCompiled {
					vm.ip = newIP
					break
				}
			}

			vm.ip = funcAddr

		case OpReturn:
			if len(vm.frames) == 0 {
				return fmt.Errorf("no frame to return to")
			}
			frameIndex := len(vm.frames) - 1
			if vm.jitEnabled {
				returnValue := vm.stack[vm.sp]
				info := vm.frames[frameIndex].funcInfo
				vm.jit.NotifyReturn(info.Address, vm.frames[frameIndex].locals[:info.ParamCount], returnValue)
			}

			vm.gc.Collect(vm.heap, vm.frames, frameIndex)

			frame := vm.frames[len(vm.frames)-1]
			vm.frames = vm.frames[:len(vm.frames)-1]

			// Restore previous frame
			vm.ip = frame.returnIP
			vm.fp = frame.prevFP

		case OpReturnVoid:
			if len(vm.frames) == 0 {
				return fmt.Errorf("no frame to return to")
			}

			frameIndex := len(vm.frames) - 1

			if vm.jitEnabled {
				info := vm.frames[frameIndex].funcInfo
				vm.jit.NotifyReturn(info.Address, vm.frames[frameIndex].locals[:info.ParamCount], Value{Type: ValNil})
			}

			vm.gc.Collect(vm.heap, vm.frames, frameIndex)

			frame := vm.frames[len(vm.frames)-1]
			vm.frames = vm.frames[:len(vm.frames)-1]

			// Restore previous frame
			vm.ip = frame.returnIP
			vm.fp = frame.prevFP

		case OpArrayAlloc:
			arrLength, ok := vm.pop().Data.(int)
			if !ok {
				return fmt.Errorf("ARRAY_ALLOC expected int size")
			}
			heapPointer := -1
			newArray := &Array{arrLength, make([]Value, arrLength)}
			for i, v := range vm.heap {
				if v == nil {
					heapPointer = i
					vm.heap[heapPointer] = newArray
				}
			}
			if heapPointer == -1 {
				vm.heap = append(vm.heap, newArray)
				heapPointer = len(vm.heap) - 1
			}

			localIndex := instr.Operands[0]
			if vm.fp < 0 || vm.fp >= len(vm.frames) {
				return fmt.Errorf("invalid frame pointer: %d", vm.fp)
			}

			currentFrame := &vm.frames[vm.fp]
			currentFrame.ensureLocalsSize(localIndex + 1)
			currentFrame.locals[localIndex] = Value{Type: ValHeapPtr, Data: heapPointer}
			vm.push(currentFrame.locals[localIndex])

		case OpArrayStore:
			data := vm.pop()
			arrIndex, ok := vm.pop().Data.(int)
			if !ok {
				return fmt.Errorf("ARRAY_STORE expected intSize")
			}
			localIndex := instr.Operands[0]

			currentFrame := &vm.frames[vm.fp]
			heapPointer, ok := currentFrame.locals[localIndex].Data.(int)
			if !ok {
				return fmt.Errorf("ARRAY_STORE expected intSize")
			}
			if arrIndex >= vm.heap[heapPointer].size {
				return fmt.Errorf("index out of range: %d", arrIndex)
			} else if arrIndex < 0 {
				return fmt.Errorf("negative index: %d", arrIndex)
			}
			vm.heap[heapPointer].Array[arrIndex] = data

		case OpArrayLoad:
			arrIndex, ok := vm.pop().Data.(int)
			if !ok {
				return fmt.Errorf("ARRAY_LOAD expected intSize")
			}
			localIndex := instr.Operands[0]

			currentFrame := &vm.frames[vm.fp]
			heapPointer, ok := currentFrame.locals[localIndex].Data.(int)
			if !ok {
				return fmt.Errorf("ARRAY_LOAD expected intSize")
			}
			if arrIndex >= vm.heap[heapPointer].size {
				return fmt.Errorf("index out of range: %d", arrIndex)
			} else if arrIndex < 0 {
				return fmt.Errorf("negative index: %d", arrIndex)
			}
			data := vm.heap[heapPointer].Array[arrIndex]
			vm.push(data)
		default:
			return fmt.Errorf("unknown opcode in instruction: %s", instr.String())
		}
	}

	return nil
}

func (vm *VM) push(value Value) {
	vm.sp++
	vm.stack[vm.sp] = value
}

func (vm *VM) pop() Value {
	if vm.sp < 0 {
		panic("stack underflow")
	}
	value := vm.stack[vm.sp]
	vm.sp--
	return value
}

func (vm *VM) binaryOp(op func(Value, Value) Value) error {
	if vm.sp < 1 {
		return fmt.Errorf("not enough values on stack for binary operation")
	}
	b := vm.pop()
	a := vm.pop()

	result := op(a, b)

	vm.push(result)
	return nil
}

// Comparison helper functions. Each returns a Value containing a bool result.
func valueLT(a, b Value) Value {
	switch aVal := a.Data.(type) {
	case int:
		if bVal, ok := b.Data.(int); ok {
			return Value{Data: aVal < bVal}
		}
	case float64:
		if bVal, ok := b.Data.(float64); ok {
			return Value{Data: aVal < bVal}
		}
	case bool:
		if bVal, ok := b.Data.(bool); ok {
			// false < true
			return Value{Data: !aVal && bVal}
		}
	}
	return Value{Data: false}
}

func valueLE(a, b Value) Value {
	switch aVal := a.Data.(type) {
	case int:
		if bVal, ok := b.Data.(int); ok {
			return Value{Data: aVal <= bVal}
		}
	case float64:
		if bVal, ok := b.Data.(float64); ok {
			return Value{Data: aVal <= bVal}
		}
	case bool:
		if bVal, ok := b.Data.(bool); ok {
			// a <= b  is true when !a || b
			return Value{Data: !aVal || bVal}
		}
	}
	return Value{Data: false}
}

func valueGT(a, b Value) Value {
	switch aVal := a.Data.(type) {
	case int:
		if bVal, ok := b.Data.(int); ok {
			return Value{Data: aVal > bVal}
		}
	case float64:
		if bVal, ok := b.Data.(float64); ok {
			return Value{Data: aVal > bVal}
		}
	case bool:
		if bVal, ok := b.Data.(bool); ok {
			// true > false
			return Value{Data: aVal && !bVal}
		}
	}
	return Value{Data: false}
}

func valueGE(a, b Value) Value {
	switch aVal := a.Data.(type) {
	case int:
		if bVal, ok := b.Data.(int); ok {
			return Value{Data: aVal >= bVal}
		}
	case float64:
		if bVal, ok := b.Data.(float64); ok {
			return Value{Data: aVal >= bVal}
		}
	case bool:
		if bVal, ok := b.Data.(bool); ok {
			// a >= b is true when a || !b
			return Value{Data: aVal || !bVal}
		}
	}
	return Value{Data: false}
}

func valueEQ(a, b Value) Value {
	switch aVal := a.Data.(type) {
	case int:
		if bVal, ok := b.Data.(int); ok {
			return Value{Data: aVal == bVal}
		}
	case float64:
		if bVal, ok := b.Data.(float64); ok {
			return Value{Data: aVal == bVal}
		}
	case string:
		if bVal, ok := b.Data.(string); ok {
			return Value{Data: aVal == bVal}
		}
	case bool:
		if bVal, ok := b.Data.(bool); ok {
			return Value{Data: aVal == bVal}
		}
	}
	return Value{Data: false}
}

func valueNEQ(a, b Value) Value {
	eq := valueEQ(a, b)
	if bBool, ok := eq.Data.(bool); ok {
		return Value{Data: !bBool}
	}
	return Value{Data: true}
}

func isTruthy(value Value) bool {
	switch v := value.Data.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case float64:
		return v != 0
	case string:
		return v != ""
	default:
		return false
	}
}
