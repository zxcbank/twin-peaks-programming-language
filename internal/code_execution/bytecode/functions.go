package bytecode

import (
	"fmt"
	"twin-peaks-programming-language/internal/parser"
)

func (c *Compiler) compileFuncDecl(node *parser.ASTNode) error {
	funcName := node.Value.(string)

	// 1. Запоминаем позицию начала функции
	funcStart := len(c.bytecode.Instructions)

	// 2. Создаем новую область видимости для функции
	prevScope := c.currentScope
	prevFunc := c.currentFunc
	prevLabels := c.labels
	prevLabelCounter := c.labelCounter

	// 3. Инициализируем новую область видимости
	c.currentScope = &Scope{
		variables: make(map[string]int),
		parent:    nil, // Функции имеют свою собственную область видимости
	}

	// 4. Параметры функции (первый ребенок - блок параметров)
	paramsNode := node.Children[0] // NodeBlock с параметрами
	paramCount := len(paramsNode.Children)

	// 5. Регистрируем параметры как локальные переменные
	for i, param := range paramsNode.Children {
		if param.Type != parser.NodeVarDecl || len(param.Children) < 1 {
			return fmt.Errorf("invalid parameter declaration")
		}

		paramName := param.Children[0].Value.(string)
		c.currentScope.variables[paramName] = i // Параметры начинаются с индекса 0
	}

	// 6. Возвращаемый тип (второй ребенок)
	returnTypeNode := node.Children[1]
	_ = returnTypeNode // Можем использовать для проверки типов

	// 7. Тело функции (третий ребенок)
	bodyNode := node.Children[2]

	// 8. Генерируем пролог функции (сохранение параметров из стека в локальные)
	// Когда функция вызывается, аргументы уже лежат на стеке
	for i := paramCount - 1; i >= 0; i-- {
		// LOAD_ARGS - специальная инструкция для загрузки i-го аргумента
		// или можно использовать комбинацию других инструкций
		c.emit(OP_LOAD_ARG, i)
	}

	// 9. Сохраняем текущий контекст функции
	c.currentFunc = &FuncContext{
		Name:       funcName,
		Address:    funcStart,
		ParamCount: paramCount,
		HasReturn:  false,
	}

	// 10. Сбрасываем метки для функции
	c.labels = make(map[string]int)
	c.labelCounter = 0

	// 11. Компилируем тело функции
	if err := c.compileNode(bodyNode); err != nil {
		return err
	}

	// 12. Если функция не заканчивается RETURN, добавляем неявный return
	if !c.currentFunc.HasReturn {
		// Для void функций просто возвращаем nil
		if returnTypeNode.Value == "void" {
			c.emit(OP_RETURN_VOID)
		} else {
			// Для функций, возвращающих значение, возвращаем 0
			c.emit(OP_CONST, c.addConstant(0))
			c.emit(OP_RETURN)
		}
	}

	// 13. Добавляем информацию о функции в таблицу
	c.funcTable[funcName] = &FunctionInfo{
		Name:       funcName,
		Address:    funcStart,
		ParamCount: paramCount,
		LocalCount: len(c.currentScope.variables),
		ReturnType: returnTypeNode.Value.(string),
	}

	// 14. Восстанавливаем предыдущий контекст
	c.currentScope = prevScope
	c.currentFunc = prevFunc
	c.labels = prevLabels
	c.labelCounter = prevLabelCounter

	c.bytecode.programStart = len(c.bytecode.Instructions)

	return nil
}
