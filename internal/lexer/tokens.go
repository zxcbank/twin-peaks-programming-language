package lexer

import (
	"fmt"
)

type TokenType int

const (
	Invalid TokenType = iota
	ConstNum
	ConstText

	Int
	Uint
	Float
	String
	Bool

	Func
	If
	Else
	For
	Return
	Break
	Continue
	True
	False
	Identifier

	Assign
	Eq
	NotEq
	Lt
	LtEq
	Gt
	GtEq
	Not
	And
	Or

	Plus
	Minus
	Mul
	Div
	Mod
	AddressOf

	LParen
	RParen
	LBrace
	RBrace
	LBracket
	RBracket
	Semicolon
	Comma
)

var TokenNames = map[TokenType]string{
	Invalid:    "Invalid",
	ConstNum:   "ConstNum",
	ConstText:  "ConstText",
	Int:        "Int",
	Uint:       "Uint",
	Float:      "Float",
	String:     "String",
	Bool:       "Bool",
	Func:       "Func",
	If:         "If",
	Else:       "Else",
	For:        "For",
	Return:     "Return",
	Break:      "Break",
	Continue:   "Continue",
	True:       "True",
	False:      "False",
	Identifier: "Identifier",
	Assign:     "Assign",
	Eq:         "Eq",
	NotEq:      "NotEq",
	Lt:         "Lt",
	LtEq:       "LtEq",
	Gt:         "Gt",
	GtEq:       "GtEq",
	Not:        "Not",
	And:        "And",
	Or:         "Or",
	Plus:       "Plus",
	Minus:      "Minus",
	Mul:        "Mul",
	Div:        "Div",
	Mod:        "Mod",
	LParen:     "LParen",
	RParen:     "RParen",
	LBrace:     "LBrace",
	RBrace:     "RBrace",
	LBracket:   "LBracket",
	RBracket:   "RBracket",
	Semicolon:  "Semicolon",
	Comma:      "Comma",
	AddressOf:  "AddressOf",
}

type Token struct {
	Type TokenType
	Text string
	Pos  int
	Line int
}

func (t Token) String() string {
	name, ok := TokenNames[t.Type]
	if !ok {
		name = fmt.Sprintf("Unknown(%d)", t.Type)
	}
	return fmt.Sprintf("Token{Type:%s, Text:%q, Pos:%d, Line:%d}", name, t.Text, t.Pos, t.Line)
}
