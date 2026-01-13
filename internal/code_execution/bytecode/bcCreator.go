package bytecode

import (
	"fmt"
	"strconv"
	"twin-peaks-programming-language/internal/lexer"
	"twin-peaks-programming-language/internal/parser"
)

// Компилятор AST в байт-код
type Compiler struct {
	bytecode     *Bytecode
	currentScope *Scope
	currentFunc  *FuncContext
	funcTable    map[string]*FunctionInfo
	labels       map[string]int
	labelCounter int
	unresolved   map[string][]int
}

// Область видимости
type Scope struct {
	variables map[string]int // имя -> индекс локальной переменной
	parent    *Scope
}

func NewCompiler() *Compiler {
	return &Compiler{
		bytecode: &Bytecode{
			Instructions: []Instruction{},
			Constants:    []interface{}{},
			FuncTable:    make(map[int]*FunctionInfo),
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

	//// Первый проход: собираем информацию о функциях
	//for _, child := range ast.Children {
	//	if child.Type == parser.NodeFuncDecl {
	//		funcName := child.Value.(string)
	//		c.bytecode.FuncTable[funcName] = &FunctionInfo{
	//			Name:       funcName,
	//			Address:    len(c.bytecode.Instructions),
	//			ParamCount: len(child.Children[0].Children),
	//		}
	//	}
	//}

	// Второй проход: компилируем код
	for _, child := range ast.Children {
		if err := c.compileNode(child); err != nil {
			return nil, err
		}
	}

	// Добавляем HALT в конце программы
	c.emit(OP_HALT)

	// Попытка разрешить все отложенные переходы
	if err := c.resolveUnresolvedLabels(); err != nil {
		return nil, err
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

	// Получаем имя переменной
	varName := node.Children[0].Value.(string)

	// Резервируем место для переменной
	localIndex := c.allocateLocal(varName)

	// Если есть инициализация
	if len(node.Children) > 2 {
		// Компилируем выражение инициализации
		if err := c.compileNode(node.Children[2]); err != nil {
			return err
		}
		// Сохраняем результат в переменную
		c.emit(OP_STORE, localIndex)
	}

	return nil
}

func (c *Compiler) compileArrayDecl(node *parser.ASTNode) error {
	if len(node.Children) < 3 {
		return fmt.Errorf("invalid ArrayDeclNode node")
	}

	arrayName := node.Children[0].Value.(string)

	index_of_array := c.allocateLocal(arrayName)
	err := c.compileNode(node.Children[1])
	if err != nil {
		return err
	}

	c.emit(OP_ARRAY_ALLOC, index_of_array)
	c.emit(OP_STORE, index_of_array)

	return nil
}

func (c *Compiler) compileArrayLoad(node *parser.ASTNode) error {

	localIndex := c.currentScope.variables[node.Children[0].Value.(string)]
	if err := c.compileNode(node.Children[1]); err != nil {
		return err
	}
	c.emit(OP_ARRAY_LOAD, localIndex)

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
	c.emit(OP_ARRAY_STORE, localIndex)
	return nil
}

func (c *Compiler) compileBinaryOp(node *parser.ASTNode) error {
	if len(node.Children) < 2 {
		return fmt.Errorf("invalid BinaryOp node")
	}

	op := node.Value.(string)

	// Обработка присваивания
	if op == "=" {
		// Левый операнд должен быть идентификатором
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

		// Компилируем правую часть
		if err := c.compileNode(node.Children[1]); err != nil {
			return err
		}

		// Сохраняем в переменную
		c.emit(OP_STORE, localIndex)
		return nil
	}

	// Компилируем левый операнд
	if err := c.compileNode(node.Children[0]); err != nil {
		return err
	}

	// Компилируем правый операнд
	if err := c.compileNode(node.Children[1]); err != nil {
		return err
	}

	// Генерируем инструкцию операции
	switch op {
	case "+":
		c.emit(OP_ADD)
	case "-":
		c.emit(OP_SUB)
	case "*":
		c.emit(OP_MUL)
	case "/":
		c.emit(OP_DIV)
	case "%":
		c.emit(OP_MOD)
	case "==":
		c.emit(OP_EQ)
	case "!=":
		c.emit(OP_NEQ)
	case "<":
		c.emit(OP_LT)
	case "<=":
		c.emit(OP_LE)
	case ">":
		c.emit(OP_GT)
	case ">=":
		c.emit(OP_GE)
	case "&&":
		c.emit(OP_AND)
	case "||":
		c.emit(OP_OR)
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
		c.emit(OP_NOT)
		return nil
	case "-":
		c.emit(OP_NEG)
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
	c.emit(OP_LOAD, localIndex)
	return nil
}

func (c *Compiler) compileLiteral(node *parser.ASTNode) error {
	value := node.Value.(string)

	switch node.Token.Type {
	case lexer.ConstNum:
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			constIndex := c.addConstant(int(intVal))
			c.emit(OP_CONST, constIndex)
			return nil
		} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			constIndex := c.addConstant(floatVal)
			c.emit(OP_CONST, constIndex)
			return nil
		}
		return fmt.Errorf("invalid number literal: %s", value)
	case lexer.ConstText:
		constIndex := c.addConstant(value)
		c.emit(OP_CONST, constIndex)
		return nil
	case lexer.True:
		constIndex := c.addConstant(true)
		c.emit(OP_CONST, constIndex)
		return nil
	case lexer.False:
		constIndex := c.addConstant(false)
		c.emit(OP_CONST, constIndex)
		return nil
	default:
		return fmt.Errorf("unsupported literal token type: %v", node.Token.Type)
	}
}

//func (c *Compiler) compileCall(node *parser.ASTNode) error {
//	funcName := node.Value.(string)
//
//	// Компилируем аргументы
//	for _, arg := range node.Children {
//		if err := c.compileNode(arg); err != nil {
//			return err
//		}
//	}
//
//	// Встроенные функции
//	switch funcName {
//	case "print":
//		c.emit(OP_PRINT)
//	default:
//		// Проверяем, существует ли функция
//		funcInfo, ok := c.bytecode.FuncTable[funcName]
//		if !ok {
//			return fmt.Errorf("function not found: %s", funcName)
//		}
//		c.emit(OP_CALL, funcInfo.Address)
//	}
//
//	return nil
//}

func (c *Compiler) compileIf(node *parser.ASTNode) error {
	if len(node.Children) < 2 {
		return fmt.Errorf("invalid If node")
	}

	// Компилируем условие
	if err := c.compileNode(node.Children[0]); err != nil {
		return err
	}

	// Создаем метки
	elseLabel := c.newLabel("else")
	endLabel := c.newLabel("endif")

	// Переход если false
	c.emitJump(OP_JMP_IF_FALSE, elseLabel)

	// Компилируем then-блок
	if err := c.compileNode(node.Children[1]); err != nil {
		return err
	}

	// Переход к концу
	c.emitJump(OP_JMP, endLabel)

	// Метка else
	c.placeLabel(elseLabel)

	// Компилируем else-блок если есть
	if len(node.Children) > 2 {
		if err := c.compileNode(node.Children[2]); err != nil {
			return err
		}
	}

	// Метка конца if
	c.placeLabel(endLabel)

	return nil
}

func (c *Compiler) compileFor(node *parser.ASTNode) error {
	if len(node.Children) < 4 {
		return fmt.Errorf("invalid For node")
	}

	// Создаем метки
	loopStart := c.newLabel("loop_start")
	loopEnd := c.newLabel("loop_end")
	loopContinue := c.newLabel("loop_continue")

	// Инициализация
	if node.Children[0].Type != parser.NodeBlock {
		if err := c.compileNode(node.Children[0]); err != nil {
			return err
		}
	}

	// Метка начала цикла
	c.placeLabel(loopStart)

	// Условие
	if node.Children[1].Type != parser.NodeLiteral || node.Children[1].Value != "true" {
		if err := c.compileNode(node.Children[1]); err != nil {
			return err
		}
		c.emitJump(OP_JMP_IF_FALSE, loopEnd)
	}

	// Тело цикла
	if err := c.compileNode(node.Children[3]); err != nil {
		return err
	}

	// Метка для continue
	c.placeLabel(loopContinue)

	// Пост-итерация
	if node.Children[2].Type != parser.NodeBlock {
		if err := c.compileNode(node.Children[2]); err != nil {
			return err
		}
	}

	// Переход к началу
	c.emitJump(OP_JMP, loopStart)

	// Метка конца цикла
	c.placeLabel(loopEnd)

	return nil
}

//func (c *Compiler) compileReturn(node *parser.ASTNode) error {
//	// Компилируем возвращаемое значение если есть
//	if len(node.Children) > 0 {
//		if err := c.compileNode(node.Children[0]); err != nil {
//			return err
//		}
//	}
//	c.emit(OP_RETURN)
//	return nil
//}

//func (c *Compiler) compileFuncDecl(node *parser.ASTNode) error {
//	// Пока пропускаем компиляцию функций (можно реализовать позже)
//	return nil
//}

func (c *Compiler) compileBlock(node *parser.ASTNode) error {
	// Компилируем все statement'ы в блоке
	for _, child := range node.Children {
		if err := c.compileNode(child); err != nil {
			return err
		}
	}
	return nil
}

// Вспомогательные методы

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
	c.labels[labelName] = -1 // -1 означает "еще не размещено"
	return labelName
}

func (c *Compiler) placeLabel(label string) {
	// Устанавливаем адрес метки на текущую позицию инструкций
	addr := len(c.bytecode.Instructions)
	c.labels[label] = addr

	// Патчим все отложенные переходы, ссылающиеся на эту метку
	if list, ok := c.unresolved[label]; ok {
		for _, idx := range list {
			// Защитимся на случай некорректного индекса
			if idx >= 0 && idx < len(c.bytecode.Instructions) {
				c.bytecode.Instructions[idx].Operands[0] = addr
			}
		}
		// Очищаем список
		delete(c.unresolved, label)
	}
}

// resolveUnresolvedLabels проверяет, что все метки были размещены и патчит все отложенные переходы.
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
		// optional: clear map entry
		delete(c.unresolved, label)
	}
	return nil
}

func (c *Compiler) compileCall(node *parser.ASTNode) error {
	funcName := node.Value.(string)

	// Проверяем, встроенная ли это функция
	switch funcName {
	case "print":
		for _, arg := range node.Children {
			if err := c.compileNode(arg); err != nil {
				return err
			}
		}
		c.emit(OP_PRINT)
		return nil
	case "sqrt":
		if len(node.Children) != 1 {
			return fmt.Errorf("function sqrt expects 1 argument, got %d", len(node.Children))
		}
		if err := c.compileNode(node.Children[0]); err != nil {
			return err
		}
		c.emit(OP_SQRT)
		return nil
	}

	// Проверяем, существует ли функция
	funcInfo, exists := c.funcTable[funcName]
	if !exists {
		return fmt.Errorf("undefined function: %s", funcName)
	}

	// Проверяем количество аргументов
	if len(node.Children) != funcInfo.ParamCount {
		return fmt.Errorf("function %s expects %d arguments, got %d",
			funcName, funcInfo.ParamCount, len(node.Children))
	}

	// Компилируем аргументы в обратном порядке
	// (для стековой архитектуры аргументы помещаются в обратном порядке)
	for i := len(node.Children) - 1; i >= 0; i-- {
		if err := c.compileNode(node.Children[i]); err != nil {
			return err
		}
	}

	// Вызов функции
	c.emit(OP_CALL, funcInfo.Address)

	return nil
}

// compileReturn теперь учитывает контекст функции
func (c *Compiler) compileReturn(node *parser.ASTNode) error {
	// Помечаем, что функция имеет return
	if c.currentFunc != nil {
		c.currentFunc.HasReturn = true
	}

	// Если есть возвращаемое значение
	if len(node.Children) > 0 {
		// Компилируем выражение
		if err := c.compileNode(node.Children[0]); err != nil {
			return err
		}
		c.emit(OP_RETURN)
	} else {
		// void return
		c.emit(OP_RETURN_VOID)
	}

	return nil
}
