package bytecode

import (
	"fmt"
)

type Frame struct {
	locals   []Value
	returnIP int
	basePtr  int
	funcInfo *FunctionInfo
}

// Виртуальная машина
type VM struct {
	bytecode *Bytecode
	stack    []Value
	frames   []Frame
	ip       int // Instruction Pointer
	sp       int // Stack Pointer
	fp       int // Frame Pointer
}

func NewVM(bytecode *Bytecode) *VM {
	return &VM{
		bytecode: bytecode,
		stack:    make([]Value, 1024*1024),
		frames:   make([]Frame, 256*256),
		ip:       bytecode.programStart,
		sp:       -1,
		fp:       0,
	}
}

func (vm *VM) Run() error {
	for vm.ip < len(vm.bytecode.Instructions) {
		instr := vm.bytecode.Instructions[vm.ip]
		vm.ip++

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
			if localIndex >= len(vm.frames[vm.fp].locals) {
				return fmt.Errorf("local variable index out of bounds: %d", localIndex)
			}
			vm.push(vm.frames[vm.fp].locals[localIndex])

		case OP_STORE:
			// Сохраняем значение в локальную переменную
			localIndex := instr.Operands[0]
			if vm.sp < 0 {
				return fmt.Errorf("stack underflow")
			}
			value := vm.pop()
			vm.frames[vm.fp].locals[localIndex] = value

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
					return Value{Data: 0}
				}
			}); err != nil {
				return err
			}

		case OP_PRINT:
			// Вывод значения
			if vm.sp < 0 {
				return fmt.Errorf("stack underflow")
			}
			value := vm.pop()
			fmt.Println(value.Data)

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
				locals:   make([]Value, 256),
			}
			vm.frames = append(vm.frames, frame)
			vm.fp = vm.sp

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
			vm.sp = vm.fp

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
			vm.sp = vm.fp

			// Пушим nil для void функций
			vm.push(Value{Type: ValNil})

		case OP_LOAD_ARG:
			// Загрузка аргумента из текущего фрейма
			argIndex := instr.Operands[0]
			if len(vm.frames) == 0 {
				return fmt.Errorf("no active frame")
			}

			currentFrame := &vm.frames[len(vm.frames)-1]
			if argIndex >= len(currentFrame.locals) {
				return fmt.Errorf("argument index out of bounds: %d", argIndex)
			}

			vm.push(currentFrame.locals[argIndex])

		case OP_ENTER:
			// Инициализация фрейма (может использоваться для выделения локальных)
			localCount := instr.Operands[0]
			if len(vm.frames) > 0 {
				currentFrame := &vm.frames[len(vm.frames)-1]
				currentFrame.locals = make([]Value, localCount)
			}

		case OP_LEAVE:
			// Очистка фрейма
			if len(vm.frames) > 0 {
				vm.frames = vm.frames[:len(vm.frames)-1]
			}

		default:
			return fmt.Errorf("unknown opcode: %d", instr.Opcode)
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
