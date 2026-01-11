package parser

import (
	"fmt"
	"strings"
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

func (n *ASTNode) String() string {
	var sb strings.Builder
	n.stringHelper(&sb, 0)
	return sb.String()
}

func (n *ASTNode) stringHelper(sb *strings.Builder, indent int) {
	indentStr := strings.Repeat("  ", indent)

	sb.WriteString(indentStr)
	n.writeName(sb)
	for _, child := range n.Children {
		child.stringHelper(sb, indent+1)
	}
}

func (n *ASTNode) writeName(sb *strings.Builder) {
	switch n.Type {
	case NodeProgram:
		sb.WriteString("Program:\n")
	case NodeVarDecl:
		sb.WriteString("VarDecl:\n")
	case NodeVarType:
		sb.WriteString(fmt.Sprintf("VarType(%s):\n", n.Token.Text))
	case NodeAssignment:
		sb.WriteString(fmt.Sprintf("Assignment(%s):\n", n.Value))
	case NodeBinaryOp:
		sb.WriteString(fmt.Sprintf("BinaryOp(%s):\n", n.Value))
	case NodeUnaryOp:
		sb.WriteString(fmt.Sprintf("UnaryOp(%s):\n", n.Value))
	case NodeLiteral:
		sb.WriteString(fmt.Sprintf("Literal(%v):\n", n.Value))
	case NodeIdentifier:
		sb.WriteString(fmt.Sprintf("Identifier(%s):\n", n.Value))
	case NodeCall:
		sb.WriteString(fmt.Sprintf("Call(%s):\n", n.Value))
	case NodeIf:
		sb.WriteString("If:\n")
	case NodeFor:
		sb.WriteString("For:\n")
	case NodeReturn:
		sb.WriteString("Return:\n")
	case NodeBlock:
		sb.WriteString("Block:\n")
	case NodeArrayDecl:
		sb.WriteString("ArrayDecl:\n")
	case NodeArrayAccess:
		sb.WriteString("ArrayAccess:\n")
	case NodePointerDecl:
		sb.WriteString("PointerDecl:\n")
	case NodeDereference:
		sb.WriteString("Dereference:\n")
	case NodeAddressOf:
		sb.WriteString("AddressOf:\n")
	case NodeFuncDecl:
		sb.WriteString(fmt.Sprintf("FuncDecl(%s):\n", n.Value))
	default:
		sb.WriteString(fmt.Sprintf("Unknown(%d):\n", n.Type))
	}
}

func (n *ASTNode) Name() string {
	var sb strings.Builder
	n.writeName(&sb)
	return sb.String()
}
