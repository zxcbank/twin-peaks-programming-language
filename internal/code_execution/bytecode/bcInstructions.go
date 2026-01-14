package bytecode

import "fmt"

type Instruction struct {
	Opcode   byte
	Operands []int // сомнительно но окей
	Line     int   // Для отладки
}

func (i Instruction) String() string {
	opcodeNames := map[byte]string{
		OP_CONST:        "CONST",
		OP_LOAD:         "LOAD",
		OP_STORE:        "STORE",
		OP_POP:          "POP",
		OP_ADD:          "ADD",
		OP_SUB:          "SUB",
		OP_MUL:          "MUL",
		OP_DIV:          "DIV",
		OP_MOD:          "MOD",
		OP_NEG:          "NEG",
		OP_EQ:           "EQ",
		OP_NEQ:          "NEQ",
		OP_LT:           "LT",
		OP_LE:           "LE",
		OP_GT:           "GT",
		OP_GE:           "GE",
		OP_AND:          "AND",
		OP_OR:           "OR",
		OP_NOT:          "NOT",
		OP_JMP:          "JMP",
		OP_JMP_IF_FALSE: "JMP_IF_FALSE",
		OP_CALL:         "CALL",
		OP_RETURN:       "RETURN",
		OP_PRINT:        "PRINT",
		OP_SQRT:         "SQRT",
		OP_HALT:         "HALT",
		OP_RETURN_VOID:  "RETURN_VOID",
		OP_ARRAY_ALLOC:  "ARRAY_ALLOC",
		OP_ARRAY_LOAD:   "ARRAY_LOAD",
		OP_ARRAY_STORE:  "ARRAY_STORE",
	}

	name := opcodeNames[i.Opcode]
	if name == "" {
		name = fmt.Sprintf("UNKNOWN(%d)", i.Opcode)
	}

	if len(i.Operands) == 0 {
		return fmt.Sprintf("%s", name)
	}
	return fmt.Sprintf("%s %v", name, i.Operands)
}

func (i Instruction) isReturn() bool {
	return i.Opcode == OP_RETURN || i.Opcode == OP_RETURN_VOID
}

func (i Instruction) hasSideEffects() bool {
	switch i.Opcode { // TODO: May JMP have side effects?
	case OP_PRINT, OP_CALL, OP_ARRAY_STORE, OP_ARRAY_LOAD, OP_HALT, OP_RETURN, OP_RETURN_VOID:
		return true
	default:
		return false
	}
}

// Байт-код программы
type Bytecode struct {
	Instructions  []Instruction
	Constants     []interface{}
	FuncAddresses map[int]*FunctionInfo
	programStart  int
}

// Контекст функции
type FuncContext struct {
	Name        string
	Address     int
	ParamCount  int
	HasReturn   bool
	ReturnLabel int
}

// Информация о функции для таблицы
type FunctionInfo struct {
	Name       string
	Address    int
	ParamCount int
	LocalCount int
	ReturnType string
}
