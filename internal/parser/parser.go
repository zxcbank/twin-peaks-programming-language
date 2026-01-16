package parser

import (
	"fmt"
	"twin-peaks-programming-language/internal/lexer"
)

type Parser struct {
	tokens    []lexer.Token
	position  int
	currToken lexer.Token
}

func NewParser(tokens []lexer.Token) *Parser {
	p := &Parser{tokens: tokens}
	if len(tokens) > 0 {
		p.currToken = tokens[0]
	}
	return p
}

func (p *Parser) advance() {
	p.position++
	if p.position < len(p.tokens) {
		p.currToken = p.tokens[p.position]
	} else {
		p.currToken = lexer.Token{Type: lexer.Invalid}
	}
}

func (p *Parser) peek() lexer.Token {
	if p.position+1 < len(p.tokens) {
		return p.tokens[p.position+1]
	}
	return lexer.Token{Type: lexer.Invalid}
}

func (p *Parser) peekN(offset int) lexer.Token {
	if p.position+offset < len(p.tokens) {
		return p.tokens[p.position+offset]
	}
	return lexer.Token{Type: lexer.Invalid}
}

func (p *Parser) expect(tokenType lexer.TokenType) error {
	if p.currToken.Type != tokenType {
		return fmt.Errorf("expected %v, got %v at line %d", lexer.TokenNames[tokenType], lexer.TokenNames[p.currToken.Type], p.currToken.Line)
	}
	return nil
}

func (p *Parser) check(tokenType lexer.TokenType) bool {
	return p.currToken.Type == tokenType
}

func (p *Parser) nextIsType() bool {
	if p.position+1 >= len(p.tokens) {
		return false
	}
	return p.peek().Type == lexer.Int || p.peek().Type == lexer.Float ||
		p.peek().Type == lexer.String || p.peek().Type == lexer.Bool || p.peek().Type == lexer.Uint
}

func (p *Parser) consume(tokenType lexer.TokenType) error {
	if err := p.expect(tokenType); err != nil {
		return err
	}
	p.advance()
	return nil
}

func (p *Parser) ParseProgram() (*ASTNode, error) {
	program := &ASTNode{
		Type:     NodeProgram,
		Children: []*ASTNode{},
	}

	for !p.check(lexer.Invalid) {
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			program.Children = append(program.Children, stmt)
		}
	}
	return program, nil
}

// ParseStatement -> ParseVarDecl | ParseAssignment | ParseIf | ParseFor | ParseReturn | ParseBlock | ParseExpressionStmt
func (p *Parser) ParseStatement() (*ASTNode, error) {
	switch {
	case p.check(lexer.LBrace):
		return p.ParseBlock()

	case p.check(lexer.Identifier) && (lexer.IsTypeToken(p.peek())):
		if p.peekN(2).Type == lexer.LBracket {
			return p.ParseArrayDecl()
		}
		return p.ParseVarDecl()

	case p.check(lexer.Identifier) && p.peek().Type == lexer.Mul:
		return p.ParsePointerDecl()

	//// Объявление массива
	//case p.check(lexer.Identifier) && p.peek().Type == lexer.LBracket:
	//	return p.ParseArrayDecl()

	case p.check(lexer.Func):
		return p.ParseFuncDecl()

	case p.check(lexer.If):
		return p.ParseIf()

	case p.check(lexer.For):
		return p.ParseFor()

	case p.check(lexer.Return):
		return p.ParseReturn()

	case p.check(lexer.Break) || p.check(lexer.Continue):
		token := p.currToken
		p.advance()
		if err := p.consume(lexer.Semicolon); err != nil {
			return nil, err
		}
		return &ASTNode{
			Type:  NodeIdentifier,
			Value: token.Text,
			Token: token,
		}, nil

	default:
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}

		if expr.Type == NodeBinaryOp && expr.Value == "=" {
			if err := p.consume(lexer.Semicolon); err != nil {
				return nil, err
			}
			return expr, nil
		}

		if err := p.consume(lexer.Semicolon); err != nil {
			return nil, err
		}
		return expr, nil
	}
}

// ParseVarDecl -> identifier type ['=' expression] ';'
func (p *Parser) ParseVarDecl() (*ASTNode, error) {
	identToken := p.currToken
	p.advance()

	typeNode, err := p.ParseType()
	if err != nil {
		return nil, err
	}

	node := &ASTNode{
		Type: NodeVarDecl,
		Children: []*ASTNode{
			{
				Type:  NodeIdentifier,
				Value: identToken.Text,
				Token: identToken,
			},
			typeNode,
		},
	}

	if p.check(lexer.Assign) {
		p.advance()
		value, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, value)
	}

	if err := p.consume(lexer.Semicolon); err != nil {
		return nil, err
	}

	return node, nil
}

// ParseType ->  [*] (Int | Float | String | Bool | Identifier) ['[' expression ']']
func (p *Parser) ParseType() (*ASTNode, error) {
	// Базовый тип
	baseTypeToken := p.currToken
	typeNode := &ASTNode{
		Type:  NodeVarType,
		Value: baseTypeToken.Text,
		Token: baseTypeToken,
	}

	// Массив
	if p.check(lexer.LBracket) {
		p.advance()

		// Размер массива
		size, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}

		if err := p.consume(lexer.RBracket); err != nil {
			return nil, err
		}

		typeNode = &ASTNode{
			Type: NodeArrayDecl,
			Children: []*ASTNode{
				size,
				typeNode,
			},
		}
	}

	// Указатель
	for p.check(lexer.Mul) {
		p.advance()
		typeNode = &ASTNode{
			Type:     NodePointerDecl,
			Children: []*ASTNode{{Type: NodeVarType, Value: p.currToken.Text, Token: p.currToken}},
		}
	}

	baseTypeToken = p.currToken
	if !lexer.IsTypeToken(baseTypeToken) {
		return nil, fmt.Errorf("expected type, got %v", baseTypeToken.String())
	}
	p.advance()

	return typeNode, nil
}

// ParseExpression -> ParseAssignment
func (p *Parser) ParseExpression() (*ASTNode, error) {
	return p.parseAssignment()
}

// parseAssignment -> ParseLogicalOr ['=' parseAssignment]
func (p *Parser) parseAssignment() (*ASTNode, error) {
	left, err := p.parseLogicalOr()
	if err != nil {
		return nil, err
	}

	if p.check(lexer.Assign) {
		token := p.currToken
		p.advance()
		right, err := p.parseAssignment()
		if err != nil {
			return nil, err
		}
		return &ASTNode{
			Type:     NodeBinaryOp,
			Value:    token.Text,
			Token:    token,
			Children: []*ASTNode{left, right},
		}, nil
	}

	return left, nil
}

// parseLogicalOr -> parseLogicalAnd {'||' parseLogicalAnd}
func (p *Parser) parseLogicalOr() (*ASTNode, error) {
	left, err := p.parseLogicalAnd()
	if err != nil {
		return nil, err
	}

	for p.check(lexer.Or) {
		token := p.currToken
		p.advance()
		right, err := p.parseLogicalAnd()
		if err != nil {
			return nil, err
		}
		left = &ASTNode{
			Type:     NodeBinaryOp,
			Value:    token.Text,
			Token:    token,
			Children: []*ASTNode{left, right},
		}
	}

	return left, nil
}

// parseLogicalAnd -> parseEquality {'&&' parseEquality}
func (p *Parser) parseLogicalAnd() (*ASTNode, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}

	for p.check(lexer.And) {
		token := p.currToken
		p.advance()
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = &ASTNode{
			Type:     NodeBinaryOp,
			Value:    token.Text,
			Token:    token,
			Children: []*ASTNode{left, right},
		}
	}

	return left, nil
}

// parseEquality -> parseRelational {('==' | '!=') parseRelational}
func (p *Parser) parseEquality() (*ASTNode, error) {
	left, err := p.parseRelational()
	if err != nil {
		return nil, err
	}

	for p.check(lexer.Eq) || p.check(lexer.NotEq) {
		token := p.currToken
		p.advance()
		right, err := p.parseRelational()
		if err != nil {
			return nil, err
		}
		left = &ASTNode{
			Type:     NodeBinaryOp,
			Value:    token.Text,
			Token:    token,
			Children: []*ASTNode{left, right},
		}
	}

	return left, nil
}

// parseRelational -> parseAdditive {('<' | '<=' | '>' | '>=') parseAdditive}
func (p *Parser) parseRelational() (*ASTNode, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}

	for p.check(lexer.Lt) || p.check(lexer.LtEq) || p.check(lexer.Gt) || p.check(lexer.GtEq) {
		token := p.currToken
		p.advance()
		right, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		left = &ASTNode{
			Type:     NodeBinaryOp,
			Value:    token.Text,
			Token:    token,
			Children: []*ASTNode{left, right},
		}
	}

	return left, nil
}

// parseAdditive -> parseMultiplicative {('+' | '-') parseMultiplicative}
func (p *Parser) parseAdditive() (*ASTNode, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}

	for p.check(lexer.Plus) || p.check(lexer.Minus) {
		token := p.currToken
		p.advance()
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &ASTNode{
			Type:     NodeBinaryOp,
			Value:    token.Text,
			Token:    token,
			Children: []*ASTNode{left, right},
		}
	}

	return left, nil
}

// parseMultiplicative -> parseUnary {('*' | '/' | '%') parseUnary}
func (p *Parser) parseMultiplicative() (*ASTNode, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.check(lexer.Mul) || p.check(lexer.Div) || p.check(lexer.Mod) {
		token := p.currToken
		p.advance()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &ASTNode{
			Type:     NodeBinaryOp,
			Value:    token.Text,
			Token:    token,
			Children: []*ASTNode{left, right},
		}
	}

	return left, nil
}

// parseUnary -> ('!' | '-' | '*' | '&') parseUnary | parsePrimary
func (p *Parser) parseUnary() (*ASTNode, error) {
	if p.check(lexer.Not) || p.check(lexer.Minus) ||
		p.check(lexer.Mul) || p.check(lexer.AddressOf) {
		token := p.currToken
		p.advance()

		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}

		nodeType := NodeUnaryOp
		if token.Type == lexer.Mul {
			nodeType = NodeDereference
		} else if token.Type == lexer.AddressOf {
			nodeType = NodeAddressOf
		}

		return &ASTNode{
			Type:     nodeType,
			Value:    token.Text,
			Token:    token,
			Children: []*ASTNode{operand},
		}, nil
	}

	return p.parsePrimary()
}

// parsePrimary -> identifier | literal | '(' expression ')' | parseCall | parseArrayAccess
func (p *Parser) parsePrimary() (*ASTNode, error) {
	switch p.currToken.Type {
	case lexer.Identifier:
		// Array Access or Function Call
		if p.peek().Type == lexer.LParen {
			return p.parseCall()
		} else if p.peek().Type == lexer.LBracket {
			return p.parseArrayAccess()
		}

		// Regular Identifier
		token := p.currToken
		p.advance()
		return &ASTNode{
			Type:  NodeIdentifier,
			Value: token.Text,
			Token: token,
		}, nil

	case lexer.ConstNum, lexer.ConstText, lexer.True, lexer.False:
		token := p.currToken
		p.advance()
		return &ASTNode{
			Type:  NodeLiteral,
			Value: token.Text,
			Token: token,
		}, nil

	case lexer.LParen:
		p.advance() // skip '('
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		if err := p.consume(lexer.RParen); err != nil {
			return nil, err
		}
		return expr, nil

	default:
		return nil, fmt.Errorf("unexpected token %v at line %d", p.currToken.String(), p.currToken.Line)
	}
}

// ParseCall -> identifier '(' [ParseExpression {',' ParseExpression}] ')'
func (p *Parser) parseCall() (*ASTNode, error) {
	identToken := p.currToken
	p.advance() // skip идентификатор

	if err := p.consume(lexer.LParen); err != nil {
		return nil, err
	}

	node := &ASTNode{
		Type:     NodeCall,
		Value:    identToken.Text,
		Token:    identToken,
		Children: []*ASTNode{},
	}

	// Arguments
	if !p.check(lexer.RParen) {
		for {
			arg, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, arg)

			if !p.check(lexer.Comma) {
				break
			}
			p.advance() // skip comma
		}
	}

	if err := p.consume(lexer.RParen); err != nil {
		return nil, err
	}

	return node, nil
}

// ParseArrayAccess -> identifier '[' expression ']'
func (p *Parser) parseArrayAccess() (*ASTNode, error) {
	identToken := p.currToken
	p.advance() // skip identifier

	if err := p.consume(lexer.LBracket); err != nil {
		return nil, err
	}

	index, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.consume(lexer.RBracket); err != nil {
		return nil, err
	}

	return &ASTNode{
		Type: NodeArrayAccess,
		Children: []*ASTNode{
			{
				Type:  NodeIdentifier,
				Value: identToken.Text,
				Token: identToken,
			},
			index,
		},
	}, nil
}

// ParseReturn -> 'return' [expression] ';'
func (p *Parser) ParseReturn() (*ASTNode, error) {
	returnToken := p.currToken
	p.advance() // skip 'return'

	node := &ASTNode{
		Type:  NodeReturn,
		Token: returnToken,
	}

	if !p.check(lexer.Semicolon) {
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		node.Children = []*ASTNode{expr}
	}

	if err := p.consume(lexer.Semicolon); err != nil {
		return nil, err
	}

	return node, nil
}

// ParseAssignment -> ParseExpression '=' ParseExpression
func (p *Parser) ParseAssignment(left *ASTNode) (*ASTNode, error) {
	assignToken := p.currToken
	p.advance() // skip '='

	right, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.consume(lexer.Semicolon); err != nil {
		return nil, err
	}

	return &ASTNode{
		Type:     NodeAssignment,
		Value:    assignToken.Text,
		Token:    assignToken,
		Children: []*ASTNode{left, right},
	}, nil
}

// ParseIf -> 'if' '(' expression ')' block ['else' (ParseIf | block)]
func (p *Parser) ParseIf() (*ASTNode, error) {
	ifToken := p.currToken
	p.advance() // skip 'if'

	if err := p.consume(lexer.LParen); err != nil {
		return nil, err
	}

	condition, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.consume(lexer.RParen); err != nil {
		return nil, err
	}

	thenBlock, err := p.ParseBlock()
	if err != nil {
		return nil, err
	}

	node := &ASTNode{
		Type:     NodeIf,
		Token:    ifToken,
		Children: []*ASTNode{condition, thenBlock},
	}

	// Else or else if
	if p.check(lexer.Else) {
		p.advance() // skip 'else'

		if p.check(lexer.If) {
			elseIf, err := p.ParseIf()
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, elseIf)
		} else {
			elseBlock, err := p.ParseBlock()
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, elseBlock)
		}
	}

	return node, nil
}

// ParseFor -> 'for' '(' [ParseStatement] ';' [expression] ';' [expression] ')' block
func (p *Parser) ParseFor() (*ASTNode, error) {
	forToken := p.currToken
	p.advance() // skip 'for'

	if err := p.consume(lexer.LParen); err != nil {
		return nil, err
	}

	var init *ASTNode
	if !p.check(lexer.Semicolon) {
		var err error
		init, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
	}

	if err := p.consume(lexer.Semicolon); err != nil {
		return nil, err
	}

	var condition *ASTNode
	if !p.check(lexer.Semicolon) {
		var err error
		condition, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
	}

	if err := p.consume(lexer.Semicolon); err != nil {
		return nil, err
	}

	var post *ASTNode
	if !p.check(lexer.RParen) {
		var err error
		post, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
	}

	if err := p.consume(lexer.RParen); err != nil {
		return nil, err
	}

	body, err := p.ParseBlock()
	if err != nil {
		return nil, err
	}

	var children []*ASTNode
	if init != nil {
		children = append(children, init)
	} else {
		children = append(children, &ASTNode{Type: NodeBlock}) // empty statement
	}

	if condition != nil {
		children = append(children, condition)
	} else {
		children = append(children, &ASTNode{Type: NodeLiteral, Value: "true"}) // infinite loop
	}

	if post != nil {
		children = append(children, post)
	} else {
		children = append(children, &ASTNode{Type: NodeBlock}) // empty statement
	}

	children = append(children, body)

	return &ASTNode{
		Type:     NodeFor,
		Token:    forToken,
		Children: children,
	}, nil
}

// ParsePointerDecl -> identifier '*' type ['=' expression] ';'
func (p *Parser) ParsePointerDecl() (*ASTNode, error) {
	identToken := p.currToken
	p.advance() // skip identifier

	if err := p.consume(lexer.Mul); err != nil {
		return nil, err
	}

	typeNode, err := p.ParseType()
	if err != nil {
		return nil, err
	}

	node := &ASTNode{
		Type: NodePointerDecl,
		Children: []*ASTNode{
			{
				Type:  NodeIdentifier,
				Value: identToken.Text,
				Token: identToken,
			},
			typeNode,
		},
	}

	// Initialization
	if p.check(lexer.Assign) {
		p.advance()
		value, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, value)
	}

	if err := p.consume(lexer.Semicolon); err != nil {
		return nil, err
	}

	return node, nil
}

// ParseArrayDecl -> identifier '[' expression ']' type ['=' '{' expression {',' expression} '}'] ';'
func (p *Parser) ParseArrayDecl() (*ASTNode, error) {
	identToken := p.currToken
	p.advance() // skip identifier

	typeNode, err := p.ParseType()
	if err != nil {
		return nil, err
	}

	if err := p.consume(lexer.LBracket); err != nil {
		return nil, err
	}

	size, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.consume(lexer.RBracket); err != nil {
		return nil, err
	}

	node := &ASTNode{
		Type: NodeArrayDecl,
		Children: []*ASTNode{
			{
				Type:  NodeIdentifier,
				Value: identToken.Text,
				Token: identToken,
			},
			size,
			typeNode,
		},
	}

	if err := p.consume(lexer.Semicolon); err != nil {
		return nil, err
	}

	return node, nil
}

// ParseFuncDecl -> 'fn' identifier '(' [ParseParamList] ')' [ParseType] ParseBlock
func (p *Parser) ParseFuncDecl() (*ASTNode, error) {
	fnToken := p.currToken
	p.advance() // skip 'fn'

	// Function name
	if err := p.expect(lexer.Identifier); err != nil {
		return nil, err
	}
	nameToken := p.currToken
	p.advance()

	if err := p.consume(lexer.LParen); err != nil {
		return nil, err
	}
	params, err := p.ParseParamList()
	if err != nil {
		return nil, err
	}

	if err := p.consume(lexer.RParen); err != nil {
		return nil, err
	}

	var returnType *ASTNode
	if p.check(lexer.LBrace) {
		returnType = &ASTNode{
			Type:  NodeIdentifier,
			Value: "void",
		}
	} else {
		var err error
		returnType, err = p.ParseType()
		if err != nil {
			return nil, err
		}
	}

	body, err := p.ParseBlock()
	if err != nil {
		return nil, err
	}

	funcNode := &ASTNode{
		Type:  NodeFuncDecl,
		Value: nameToken.Text,
		Token: fnToken,
		Children: []*ASTNode{
			{ // params
				Type:     NodeBlock,
				Children: params,
			},
			returnType,
			body,
		},
	}

	return funcNode, nil
}

// ParseParamList -> (ParseParamDecl (',' ParseParamDecl)*)?
func (p *Parser) ParseParamList() ([]*ASTNode, error) {
	var params []*ASTNode

	if p.check(lexer.RParen) {
		return params, nil
	}

	for {
		// identifier type
		if err := p.expect(lexer.Identifier); err != nil {
			return nil, err
		}
		identToken := p.currToken
		p.advance()

		typeNode, err := p.ParseType()
		if err != nil {
			return nil, err
		}

		paramNode := &ASTNode{
			Type: NodeVarDecl,
			Children: []*ASTNode{
				{
					Type:  NodeIdentifier,
					Value: identToken.Text,
					Token: identToken,
				},
				typeNode,
			},
		}
		params = append(params, paramNode)

		if !p.check(lexer.Comma) {
			break
		}
		p.advance() // skip comma
	}

	return params, nil
}

// ParseBlock -> '{' {ParseStatement} '}'
func (p *Parser) ParseBlock() (*ASTNode, error) {
	if err := p.consume(lexer.LBrace); err != nil {
		return nil, fmt.Errorf("expected '{' at line %d, got %v", p.currToken.Line, p.currToken.String())
	}

	block := &ASTNode{
		Type:     NodeBlock,
		Children: []*ASTNode{},
	}

	for !p.check(lexer.RBrace) && !p.check(lexer.Invalid) {
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			block.Children = append(block.Children, stmt)
		}
	}

	if err := p.consume(lexer.RBrace); err != nil {
		return nil, fmt.Errorf("expected '}' at line %d, got %v", p.currToken.Line, p.currToken.String())
	}

	return block, nil
}
