package parser

import (
	"twin-peaks-programming-language/internal/lexer"
)

type NodeType int

const (
	NodeProgram NodeType = iota
	NodeVarDecl
	NodeVarType
	NodeAssignment
	NodeBinaryOp
	NodeUnaryOp
	NodeLiteral
	NodeIdentifier
	NodeCall
	NodeIf
	NodeFor
	NodeReturn
	NodeBlock
	NodeArrayDecl
	NodeArrayAccess
	NodePointerDecl
	NodeDereference
	NodeAddressOf
	NodeFuncDecl
)

type ASTNode struct {
	Type     NodeType
	Children []*ASTNode
	Value    interface{}
	Token    lexer.Token
}
