package bytecode

import (
	"fmt"
	"math"
)

type Frame struct {
	locals   []Value
	returnIP int
	basePtr  int // index of previous frame in frames slice
	funcInfo *FunctionInfo
}

// ensureLocalsSize ensures the frame has at least `required` slots in locals.
// It grows the slice using append to avoid unnecessary copying.
func (f *Frame) ensureLocalsSize(required int) {
	if required <= len(f.locals) {
		return
	}
	needed := required - len(f.locals)
	f.locals = append(f.locals, make([]Value, needed)...)
}

// Виртуальная машина
type VM struct {
	bytecode *Bytecode
	stack    []Value
	frames   []Frame
	heap     []Array
	ip       int // Instruction Pointer
	sp       int // Stack Pointer
	fp       int // Frame Pointer (index into frames slice)
}

type Array struct {
	size           int
	Array          []Value
	referenceCount int //  ?? we had them heappas in twin-peaks-pr0gramming-lаn9ua9e
}

func NewVM(bytecode *Bytecode) *VM {
	// Create frames slice with a single base frame
	frames := make([]Frame, 1)
	//frames = append(frames, Frame{
	//	locals: make([]Value, 128), // base frame locals
	//	// returnIP/basePtr/funcInfo are zero values
	//})
	return &VM{
		bytecode: bytecode,
		stack:    make([]Value, 1024*1024*1024),
		frames:   frames,
		ip:       bytecode.programStart,
		sp:       -1,
		fp:       0, // index of current frame
	}
}

func (vm *VM) Run() error {
	for vm.ip < len(vm.bytecode.Instructions) {
		instr := vm.bytecode.Instructions[vm.ip]
		vm.ip++
		//fmt.Printf("Executing instruction %s\n", instr.String())
		switch instr.Opcode {
		case OP_CONST:
			// Загружаем константу на стек
			constIndex := instr.Operands[0]
			if constIndex >= len(vm.bytecode.Constants) {
				return fmt.Errorf("constant index out of bounds: %d", constIndex)
			}
			value := vm.bytecode.Constants[constIndex]
			vm.push(Value{Data: value})

		case OP_LOAD:
			// Загружаем локальную переменную
			localIndex := instr.Operands[0]
			if vm.fp < 0 || vm.fp >= len(vm.frames) {
				return fmt.Errorf("invalid frame pointer: %d", vm.fp)
			}
			// ensure locals capacity
			currentFrame := &vm.frames[vm.fp]
			//currentFrame.ensureLocalsSize(localIndex + 1)
			vm.push(currentFrame.locals[localIndex])

		case OP_STORE:
			// Сохраняем значение в локальную переменную
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

		case OP_ADD:
			if err := vm.binaryOp(func(a, b Value) Value {
				// Сложение
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

		case OP_SUB:
			if err := vm.binaryOp(func(a, b Value) Value {
				// Вычитание
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

		case OP_MUL:
			if err := vm.binaryOp(func(a, b Value) Value {
				// Умножение
				switch a.Data.(type) {
				case int:
					return Value{Data: a.Data.(int) * b.Data.(int)}
				case float64:
					return Value{Data: a.Data.(float64) * b.Data.(float64)}
				default:
					return Value{Data: 0}
				}
			}); err != nil {
				return err
			}

		case OP_DIV:
			if err := vm.binaryOp(func(a, b Value) Value {
				// Деление
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
		case OP_MOD:
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
		case OP_NEG:
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
		case OP_LT:
			if err := vm.binaryOp(func(a, b Value) Value { return valueLT(a, b) }); err != nil {
				return err
			}
		case OP_LE:
			if err := vm.binaryOp(func(a, b Value) Value { return valueLE(a, b) }); err != nil {
				return err
			}
		case OP_EQ:
			if err := vm.binaryOp(func(a, b Value) Value { return valueEQ(a, b) }); err != nil {
				return err
			}
		case OP_NEQ:
			if err := vm.binaryOp(func(a, b Value) Value { return valueNEQ(a, b) }); err != nil {
				return err
			}
		case OP_GT:
			if err := vm.binaryOp(func(a, b Value) Value { return valueGT(a, b) }); err != nil {
				return err
			}
		case OP_GE:
			if err := vm.binaryOp(func(a, b Value) Value { return valueGE(a, b) }); err != nil {
				return err
			}
		case OP_AND:
			err := vm.binaryOp(func(a, b Value) Value {
				return Value{Data: isTruthy(a) && isTruthy(b)}
			})
			if err != nil {
				return err
			}
		case OP_OR:
			err := vm.binaryOp(func(a, b Value) Value {
				return Value{Data: isTruthy(a) || isTruthy(b)}
			})
			if err != nil {
				return err
			}
		case OP_NOT:
			// Logical NOT: pop one value and push its negation
			if vm.sp < 0 {
				return fmt.Errorf("stack underflow")
			}
			val := vm.pop()
			vm.push(Value{Data: !isTruthy(val)})

		case OP_PRINT:
			// Вывод значения
			if vm.sp < 0 {
				return fmt.Errorf("stack underflow")
			}
			value := vm.pop()
			fmt.Println(value.Data)

		case OP_SQRT:
			// Вычисление квадратного корня
			if vm.sp < 0 {
				return fmt.Errorf("stack underflow")
			}
			value := vm.pop()
			switch v := value.Data.(type) {
			case int:
				// Преобразуем к float64 для вычисления корня
				result := math.Sqrt(float64(v))
				vm.push(Value{Data: result})
			case float64:
				result := math.Sqrt(v)
				vm.push(Value{Data: result})
			default:
				return fmt.Errorf("SQRT operation requires int or float64")
			}

		case OP_HALT:
			// Остановка
			return nil

		case OP_JMP:
			// Безусловный переход
			vm.ip = instr.Operands[0]

		case OP_JMP_IF_FALSE:
			// Условный переход
			if vm.sp < 0 {
				return fmt.Errorf("stack underflow")
			}
			condition := vm.pop()
			if !isTruthy(condition) {
				vm.ip = instr.Operands[0]
			}

		case OP_CALL:
			funcAddr := instr.Operands[0]

			// Создаем новый фрейм
			frame := Frame{
				returnIP: vm.ip,
				basePtr:  vm.fp,
				locals:   make([]Value, 0),
			}
			vm.frames = append(vm.frames, frame)
			// FP is index of current frame (last one)
			vm.fp = len(vm.frames) - 1

			// Переходим к функции
			vm.ip = funcAddr

		case OP_RETURN:
			// Возвращаемое значение на вершине стека
			returnValue := vm.pop()

			// Восстанавливаем предыдущий фрейм
			if len(vm.frames) == 0 {
				return fmt.Errorf("no frame to return to")
			}

			frame := vm.frames[len(vm.frames)-1]
			vm.frames = vm.frames[:len(vm.frames)-1]

			// Восстанавливаем состояние
			vm.ip = frame.returnIP
			vm.fp = frame.basePtr
			// DO NOT modify sp here; stack should remain intact

			// Кладем возвращаемое значение на стек
			vm.push(returnValue)

		case OP_RETURN_VOID:
			// Аналогично RETURN, но без значения
			if len(vm.frames) == 0 {
				return fmt.Errorf("no frame to return to")
			}

			frame := vm.frames[len(vm.frames)-1]
			vm.frames = vm.frames[:len(vm.frames)-1]

			vm.ip = frame.returnIP
			vm.fp = frame.basePtr
			// Пушим nil для void функций
			vm.push(Value{Type: ValNil})

		case OP_LOAD_ARG:
			// Загрузка аргумента из текущего фрейма
			argIndex := instr.Operands[0]
			if vm.fp < 0 || vm.fp >= len(vm.frames) {
				return fmt.Errorf("no active frame")
			}

			currentFrame := &vm.frames[vm.fp]
			// grow locals if needed so caller can access args even if OP_ENTER wasn't emitted
			currentFrame.ensureLocalsSize(argIndex + 1)

			vm.push(currentFrame.locals[argIndex])

		case OP_ENTER:
			// Инициализация фрейма (может использоваться для выделения локальных)
			localCount := instr.Operands[0]
			if vm.fp < 0 || vm.fp >= len(vm.frames) {
				return fmt.Errorf("no active frame for enter: %d", vm.fp)
			}
			currentFrame := &vm.frames[vm.fp]
			// Ensure at least localCount slots
			if localCount <= 0 {
				// keep current locals as-is (but ensure non-nil)
				if len(currentFrame.locals) == 0 {
					currentFrame.locals = make([]Value, 0)
				}
			} else {
				currentFrame.locals = make([]Value, localCount)
			}

		case OP_LEAVE:
			// Очистка фрейма
			if len(vm.frames) > 0 {
				vm.frames = vm.frames[:len(vm.frames)-1]
				// adjust fp
				if len(vm.frames) == 0 {
					// restore base frame
					vm.frames = append(vm.frames, Frame{locals: make([]Value, 128)})
					vm.fp = 0
				} else {
					vm.fp = len(vm.frames) - 1
				}
			}
		case OP_ARRAY_ALLOC:
			arrLength, ok := vm.pop().Data.(int)
			if !ok {
				return fmt.Errorf("ARRAY_ALLOC expected intSize")
			}
			neccessaryIndex := len(vm.heap)
			vm.heap = append(vm.heap, Array{arrLength, make([]Value, arrLength), 1})

			localIndex := instr.Operands[0]
			if vm.fp < 0 || vm.fp >= len(vm.frames) {
				return fmt.Errorf("invalid frame pointer: %d", vm.fp)
			}
			// ensure locals capacity
			currentFrame := &vm.frames[vm.fp]
			currentFrame.ensureLocalsSize(localIndex + 1)
			currentFrame.locals[localIndex] = Value{Data: neccessaryIndex}
			vm.push(currentFrame.locals[localIndex])

		case OP_ARRAY_STORE:
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

		case OP_ARRAY_LOAD:
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

// Вспомогательные методы

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

func (vm *VM) pushFrame() {
	// Сохраняем текущий фрейм
	vm.push(Value{Data: vm.fp})
	vm.push(Value{Data: vm.ip})
	vm.fp = vm.sp
}

func (vm *VM) popFrame() {
	// Восстанавливаем предыдущий фрейм
	vm.sp = vm.fp
	vm.ip = vm.pop().Data.(int)
	vm.fp = vm.pop().Data.(int)
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
			return Value{Data: (aVal && !bVal)}
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
	// Just invert EQ
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
