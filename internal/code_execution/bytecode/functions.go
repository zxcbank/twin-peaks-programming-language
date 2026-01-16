package bytecode

import (
	"fmt"
	"twin-peaks-programming-language/internal/parser"
)

func (c *Compiler) compileFuncDecl(node *parser.ASTNode) error {
	funcName := node.Value.(string)

	funcStart := len(c.bytecode.Instructions)

	prevScope := c.currentScope
	prevFunc := c.currentFunc
	prevLabels := c.labels
	prevLabelCounter := c.labelCounter

	c.currentScope = &Scope{
		variables: make(map[string]int),
	}

	paramsNode := node.Children[0]
	paramCount := len(paramsNode.Children)

	// Parameters as local variables
	for i, param := range paramsNode.Children {
		if param.Type != parser.NodeVarDecl || len(param.Children) < 1 {
			return fmt.Errorf("invalid parameter declaration")
		}

		paramName := param.Children[0].Value.(string)
		c.currentScope.variables[paramName] = i
	}
	for i := 0; i < paramCount; i++ {
		c.emit(OpStore, i)
	}

	c.currentFunc = &FuncContext{
		Name:       funcName,
		Address:    funcStart,
		ParamCount: paramCount,
		HasReturn:  false,
	}

	returnTypeNode := node.Children[1]
	bodyNode := node.Children[2]

	c.funcTable[funcName] = &FunctionInfo{
		Name:       funcName,
		Address:    funcStart,
		ParamCount: paramCount,
		LocalCount: len(c.currentScope.variables),
		ReturnType: returnTypeNode.Value.(string),
	}

	c.labels = make(map[string]int)
	c.labelCounter = 0
	if err := c.compileNode(bodyNode); err != nil {
		return err
	}

	// Add missing return if needed
	if !c.currentFunc.HasReturn {
		if returnTypeNode.Value == "void" {
			c.emit(OpReturnVoid)
		} else {
			c.emit(OpConst, c.addConstant(0))
			c.emit(OpReturn)
		}
	}

	c.currentScope = prevScope
	c.currentFunc = prevFunc
	c.labels = prevLabels
	c.labelCounter = prevLabelCounter

	// Function are stored before the main program
	c.bytecode.programStart = len(c.bytecode.Instructions)

	return nil
}
