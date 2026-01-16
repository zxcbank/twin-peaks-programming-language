package bytecode

import "fmt"

type Instruction struct {
	Opcode   byte
	Operands []int
	Line     int
}

func (i Instruction) String() string {
	opcodeNames := map[byte]string{
		OpConst:      "CONST",
		OpLoad:       "LOAD",
		OpStore:      "STORE",
		OpPop:        "POP",
		OpAdd:        "ADD",
		OpSub:        "SUB",
		OpMul:        "MUL",
		OpDiv:        "DIV",
		OpMod:        "MOD",
		OpNeg:        "NEG",
		OpEq:         "EQ",
		OpNeq:        "NEQ",
		OpLt:         "LT",
		OpLe:         "LE",
		OpGt:         "GT",
		OpGe:         "GE",
		OpAnd:        "AND",
		OpOr:         "OR",
		OpNot:        "NOT",
		OpJmp:        "JMP",
		OpJmpIfFalse: "JMP_IF_FALSE",
		OpCall:       "CALL",
		OpReturn:     "RETURN",
		OpPrint:      "PRINT",
		OpSqrt:       "SQRT",
		OpHalt:       "HALT",
		OpReturnVoid: "RETURN_VOID",
		OpArrayAlloc: "ARRAY_ALLOC",
		OpArrayLoad:  "ARRAY_LOAD",
		OpArrayStore: "ARRAY_STORE",
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
	return i.Opcode == OpReturn || i.Opcode == OpReturnVoid
}

func (i Instruction) isJump() bool {
	return i.Opcode == OpJmp || i.Opcode == OpJmpIfFalse
}

func (i Instruction) hasSideEffects() bool {
	switch i.Opcode {
	case OpPrint, OpCall, OpArrayStore, OpArrayLoad, OpHalt:
		return true
	default:
		return false
	}
}

type Bytecode struct {
	Instructions  []Instruction
	Constants     []interface{}
	FuncAddresses map[int]*FunctionInfo
	programStart  int
}

// FuncContext is function's compilation context
type FuncContext struct {
	Name        string
	Address     int
	ParamCount  int
	HasReturn   bool
	ReturnLabel int
}

// FunctionInfo runtime information about a function
type FunctionInfo struct {
	Name       string
	Address    int
	ParamCount int
	LocalCount int
	ReturnType string
}
