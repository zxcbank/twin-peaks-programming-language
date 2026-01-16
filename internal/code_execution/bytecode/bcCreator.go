package bytecode

import (
	"fmt"
	"strconv"
	"twin-peaks-programming-language/internal/lexer"
	"twin-peaks-programming-language/internal/parser"
)

type Compiler struct {
	bytecode     *Bytecode
	currentScope *Scope
	currentFunc  *FuncContext
	funcTable    map[string]*FunctionInfo
	labels       map[string]int
	labelCounter int
	unresolved   map[string][]int
}

type Scope struct {
	variables map[string]int // имя -> индекс локальной переменной
}

func NewCompiler() *Compiler {
	return &Compiler{
		bytecode: &Bytecode{
			Instructions:  []Instruction{},
			Constants:     []interface{}{},
			FuncAddresses: make(map[int]*FunctionInfo),
		},
		currentScope: &Scope{variables: make(map[string]int)},
		funcTable:    make(map[string]*FunctionInfo),
		labels:       make(map[string]int),
		unresolved:   make(map[string][]int),
		labelCounter: 0,
	}
}

func (c *Compiler) Compile(ast *parser.ASTNode) (*Bytecode, error) {
	if ast.Type != parser.NodeProgram {
		return nil, fmt.Errorf("expected Program node")
	}

	for _, child := range ast.Children {
		if err := c.compileNode(child); err != nil {
			return nil, err
		}
	}

	c.emit(OpHalt)

	if err := c.resolveUnresolvedLabels(); err != nil {
		return nil, err
	}

	for _, info := range c.funcTable {
		c.bytecode.FuncAddresses[info.Address] = info
	}

	return c.bytecode, nil
}

func (c *Compiler) compileNode(node *parser.ASTNode) error {
	switch node.Type {
	case parser.NodeVarDecl:
		return c.compileVarDecl(node)
	case parser.NodeArrayDecl:
		return c.compileArrayDecl(node)
	case parser.NodeArrayAccess: // неувязочка, имеется в виду что ArrayStore вызывается так и так в =, а вот NodeArrayAccess (x=array[i]) может быть вызван
		return c.compileArrayLoad(node)
	case parser.NodeBinaryOp:
		return c.compileBinaryOp(node)
	case parser.NodeUnaryOp:
		return c.compileUnaryOp(node)
	case parser.NodeIdentifier:
		return c.compileIdentifier(node)
	case parser.NodeLiteral:
		return c.compileLiteral(node)
	case parser.NodeCall:
		return c.compileCall(node)
	case parser.NodeIf:
		return c.compileIf(node)
	case parser.NodeFor:
		return c.compileFor(node)
	case parser.NodeReturn:
		return c.compileReturn(node)
	case parser.NodeFuncDecl:
		return c.compileFuncDecl(node)
	case parser.NodeBlock:
		return c.compileBlock(node)

	default:
		return fmt.Errorf("unsupported node type: \n%s", node.String())
	}
}

func (c *Compiler) compileVarDecl(node *parser.ASTNode) error {
	if len(node.Children) < 2 {
		return fmt.Errorf("invalid VarDecl node")
	}

	varName := node.Children[0].Value.(string)

	localIndex := c.allocateLocal(varName)

	if len(node.Children) > 2 {
		if err := c.compileNode(node.Children[2]); err != nil {
			return err
		}
		c.emit(OpStore, localIndex)
	}

	return nil
}

func (c *Compiler) compileArrayDecl(node *parser.ASTNode) error {
	if len(node.Children) < 3 {
		return fmt.Errorf("invalid ArrayDeclNode node")
	}

	arrayName := node.Children[0].Value.(string)

	arrayIndex := c.allocateLocal(arrayName)
	err := c.compileNode(node.Children[1])
	if err != nil {
		return err
	}

	c.emit(OpArrayAlloc, arrayIndex)
	c.emit(OpStore, arrayIndex)

	return nil
}

func (c *Compiler) compileArrayLoad(node *parser.ASTNode) error {

	localIndex := c.currentScope.variables[node.Children[0].Value.(string)]
	if err := c.compileNode(node.Children[1]); err != nil {
		return err
	}
	c.emit(OpArrayLoad, localIndex)

	return nil
}

func (c *Compiler) compileArrayStore(l, r *parser.ASTNode) error {
	localIndex := c.currentScope.variables[l.Children[0].Value.(string)]
	if err := c.compileNode(l.Children[1]); err != nil {
		return err
	}
	if err := c.compileNode(r); err != nil {
		return err
	}
	c.emit(OpArrayStore, localIndex)
	return nil
}

func (c *Compiler) compileBinaryOp(node *parser.ASTNode) error {
	if len(node.Children) < 2 {
		return fmt.Errorf("invalid BinaryOp node")
	}

	op := node.Value.(string)

	if op == "=" {
		left := node.Children[0]
		if left.Type == parser.NodeArrayAccess {
			err := c.compileArrayStore(left, node.Children[1])
			if err != nil {
				return err
			}
			return nil
		}

		if left.Type != parser.NodeIdentifier {
			return fmt.Errorf("left side of assignment must be identifier")
		}

		varName := left.Value.(string)
		localIndex, ok := c.currentScope.variables[varName]
		if !ok {
			return fmt.Errorf("variable not declared: %s", varName)
		}

		if err := c.compileNode(node.Children[1]); err != nil {
			return err
		}

		c.emit(OpStore, localIndex)
		return nil
	}

	// Left operand
	if err := c.compileNode(node.Children[0]); err != nil {
		return err
	}

	// Right operand
	if err := c.compileNode(node.Children[1]); err != nil {
		return err
	}

	switch op {
	case "+":
		c.emit(OpAdd)
	case "-":
		c.emit(OpSub)
	case "*":
		c.emit(OpMul)
	case "/":
		c.emit(OpDiv)
	case "%":
		c.emit(OpMod)
	case "==":
		c.emit(OpEq)
	case "!=":
		c.emit(OpNeq)
	case "<":
		c.emit(OpLt)
	case "<=":
		c.emit(OpLe)
	case ">":
		c.emit(OpGt)
	case ">=":
		c.emit(OpGe)
	case "&&":
		c.emit(OpAnd)
	case "||":
		c.emit(OpOr)
	default:
		return fmt.Errorf("unknown binary operator: %s", op)
	}

	return nil
}

func (c *Compiler) compileUnaryOp(node *parser.ASTNode) error {
	if len(node.Children) < 1 {
		return fmt.Errorf("invalid UnaryOp node")
	}

	op := node.Value.(string)

	if err := c.compileNode(node.Children[0]); err != nil {
		return err
	}
	switch op {
	case "!":
		c.emit(OpNot)
		return nil
	case "-":
		c.emit(OpNeg)
		return nil
	default:
		return fmt.Errorf("unknown unary operator: %s", op)
	}
}

func (c *Compiler) compileIdentifier(node *parser.ASTNode) error {
	varName := node.Value.(string)
	localIndex, ok := c.currentScope.variables[varName]
	if !ok {
		return fmt.Errorf("variable not found: %s", varName)
	}
	c.emit(OpLoad, localIndex)
	return nil
}

func (c *Compiler) compileLiteral(node *parser.ASTNode) error {
	value := node.Value.(string)

	switch node.Token.Type {
	case lexer.ConstNum:
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			constIndex := c.addConstant(int(intVal))
			c.emit(OpConst, constIndex)
			return nil
		} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			constIndex := c.addConstant(floatVal)
			c.emit(OpConst, constIndex)
			return nil
		}
		return fmt.Errorf("invalid number literal: %s", value)
	case lexer.ConstText:
		constIndex := c.addConstant(value)
		c.emit(OpConst, constIndex)
		return nil
	case lexer.True:
		constIndex := c.addConstant(true)
		c.emit(OpConst, constIndex)
		return nil
	case lexer.False:
		constIndex := c.addConstant(false)
		c.emit(OpConst, constIndex)
		return nil
	default:
		return fmt.Errorf("unsupported literal token type: %v", node.Token.Type)
	}
}

func (c *Compiler) compileIf(node *parser.ASTNode) error {
	if len(node.Children) < 2 {
		return fmt.Errorf("invalid If node")
	}

	// Condition
	if err := c.compileNode(node.Children[0]); err != nil {
		return err
	}

	elseLabel := c.newLabel("else")
	endLabel := c.newLabel("endif")

	c.emitJump(OpJmpIfFalse, elseLabel)

	// Then-block
	if err := c.compileNode(node.Children[1]); err != nil {
		return err
	}

	c.emitJump(OpJmp, endLabel)

	// Else
	c.placeLabel(elseLabel)

	// Else-block
	if len(node.Children) > 2 {
		if err := c.compileNode(node.Children[2]); err != nil {
			return err
		}
	}

	c.placeLabel(endLabel)

	return nil
}

func (c *Compiler) compileFor(node *parser.ASTNode) error {
	if len(node.Children) < 4 {
		return fmt.Errorf("invalid For node")
	}

	loopStart := c.newLabel("loop_start")
	loopEnd := c.newLabel("loop_end")
	loopContinue := c.newLabel("loop_continue")

	// Init
	if node.Children[0].Type != parser.NodeBlock {
		if err := c.compileNode(node.Children[0]); err != nil {
			return err
		}
	}

	c.placeLabel(loopStart)

	// Condition
	if node.Children[1].Type != parser.NodeLiteral || node.Children[1].Value != "true" {
		if err := c.compileNode(node.Children[1]); err != nil {
			return err
		}
		c.emitJump(OpJmpIfFalse, loopEnd)
	}

	// Loop body
	if err := c.compileNode(node.Children[3]); err != nil {
		return err
	}

	c.placeLabel(loopContinue)

	// Post iteration
	if node.Children[2].Type != parser.NodeBlock {
		if err := c.compileNode(node.Children[2]); err != nil {
			return err
		}
	}

	c.emitJump(OpJmp, loopStart)
	c.placeLabel(loopEnd)

	return nil
}

func (c *Compiler) compileBlock(node *parser.ASTNode) error {
	for _, child := range node.Children {
		if err := c.compileNode(child); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) emit(opcode byte, operands ...int) {
	c.bytecode.Instructions = append(c.bytecode.Instructions, Instruction{
		Opcode:   opcode,
		Operands: operands,
	})
}

func (c *Compiler) emitJump(opcode byte, label string) {
	instr := Instruction{Opcode: opcode, Operands: []int{0}}
	c.bytecode.Instructions = append(c.bytecode.Instructions, instr)
	idx := len(c.bytecode.Instructions) - 1

	addr, ok := c.labels[label]
	if !ok || addr == -1 {
		// label not yet placed - remember to patch later
		c.unresolved[label] = append(c.unresolved[label], idx)
	} else {
		c.bytecode.Instructions[idx].Operands[0] = addr
	}
}

func (c *Compiler) addConstant(value interface{}) int {
	for i, v := range c.bytecode.Constants {
		if v == value {
			return i
		}
	}
	c.bytecode.Constants = append(c.bytecode.Constants, value)
	return len(c.bytecode.Constants) - 1
}

func (c *Compiler) allocateLocal(name string) int {
	index := len(c.currentScope.variables)
	c.currentScope.variables[name] = index
	return index
}

func (c *Compiler) newLabel(prefix string) string {
	labelName := fmt.Sprintf("%s_%d", prefix, c.labelCounter)
	c.labelCounter++
	c.labels[labelName] = -1 // Not yet placed
	return labelName
}

func (c *Compiler) placeLabel(label string) {
	addr := len(c.bytecode.Instructions)
	c.labels[label] = addr

	// Patch all unresolved jumps to this label
	if list, ok := c.unresolved[label]; ok {
		for _, idx := range list {
			if idx >= 0 && idx < len(c.bytecode.Instructions) {
				c.bytecode.Instructions[idx].Operands[0] = addr
			}
		}
		delete(c.unresolved, label)
	}
}

func (c *Compiler) resolveUnresolvedLabels() error {
	if len(c.unresolved) == 0 {
		return nil
	}
	for label, idxs := range c.unresolved {
		addr, ok := c.labels[label]
		if !ok || addr == -1 {
			return fmt.Errorf("unresolved label: %s", label)
		}
		for _, idx := range idxs {
			if idx >= 0 && idx < len(c.bytecode.Instructions) {
				c.bytecode.Instructions[idx].Operands[0] = addr
			}
		}
		delete(c.unresolved, label)
	}
	return nil
}

func (c *Compiler) compileCall(node *parser.ASTNode) error {
	funcName := node.Value.(string)

	isBuiltin, err := c.tryCompileBuiltInFunc(node)
	if err != nil {
		return err
	}
	if isBuiltin {
		return nil
	}

	funcInfo, exists := c.funcTable[funcName]
	if !exists {
		return fmt.Errorf("undefined function: %s", funcName)
	}

	if len(node.Children) != funcInfo.ParamCount {
		return fmt.Errorf("function %s expects %d arguments, got %d",
			funcName, funcInfo.ParamCount, len(node.Children))
	}

	// Reverse order of arguments
	for i := len(node.Children) - 1; i >= 0; i-- {
		if err := c.compileNode(node.Children[i]); err != nil {
			return err
		}
	}

	c.emit(OpCall, funcInfo.Address)

	return nil
}

func (c *Compiler) tryCompileBuiltInFunc(node *parser.ASTNode) (bool, error) {
	funcName := node.Value.(string)

	switch funcName {
	case "print":
		for _, arg := range node.Children {
			if err := c.compileNode(arg); err != nil {
				return true, err
			}
		}
		c.emit(OpPrint)
		return true, nil
	case "sqrt":
		if len(node.Children) != 1 {
			return true, fmt.Errorf("function sqrt expects 1 argument, got %d", len(node.Children))
		}
		if err := c.compileNode(node.Children[0]); err != nil {
			return true, err
		}
		c.emit(OpSqrt)
		return true, nil
	}
	return false, nil
}

func (c *Compiler) compileReturn(node *parser.ASTNode) error {
	if c.currentFunc != nil {
		c.currentFunc.HasReturn = true
	}

	if len(node.Children) > 0 {
		if err := c.compileNode(node.Children[0]); err != nil {
			return err
		}
		c.emit(OpReturn)
	} else {
		c.emit(OpReturnVoid)
	}

	return nil
}
