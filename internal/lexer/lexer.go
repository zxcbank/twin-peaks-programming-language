package lexer

import (
	"fmt"
	"unicode"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	pos          int
	line         int
	column       int
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 1}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
		if l.ch == '\n' {
			l.line++
			l.column = 0
		} else {
			l.column++
		}
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.column = 0
		}
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	l.skipWhitespace()
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readNumber() string {
	start := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch != '.' || !isDigit(l.peekChar()) {
		return l.input[start:l.position]
	}
	l.readChar()
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == 'e' || l.ch == 'E' {
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	return l.input[start:l.position]
}

func (l *Lexer) readString() string {
	start := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			l.readChar()
			break
		}
		if l.ch == '\\' && l.peekChar() == '"' {
			l.readChar()
		}
	}
	return l.input[start : l.position-1]
}

func (l *Lexer) NextToken() (Token, error) {
	var tok Token

	l.skipWhitespace()

	if l.ch == '/' && l.peekChar() == '/' {
		l.skipComment()
		return l.NextToken()
	}

	tok.Pos = l.position
	tok.Line = l.line

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = Eq
			tok.Text = string(ch) + string(l.ch)
		} else {
			tok.Type = Assign
			tok.Text = string(l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = NotEq
			tok.Text = string(ch) + string(l.ch)
		} else {
			tok.Type = Not
			tok.Text = string(l.ch)
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = LtEq
			tok.Text = string(ch) + string(l.ch)
		} else {
			tok.Type = Lt
			tok.Text = string(l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = GtEq
			tok.Text = string(ch) + string(l.ch)
		} else {
			tok.Type = Gt
			tok.Text = string(l.ch)
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok.Type = And
			tok.Text = string(ch) + string(l.ch)
		} else {
			tok.Type = AddressOf
			tok.Text = string(l.ch)
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok.Type = Or
			tok.Text = string(ch) + string(l.ch)
		}
	case '+':
		tok.Type = Plus
		tok.Text = string(l.ch)
	case '-':
		tok.Type = Minus
		tok.Text = string(l.ch)
	case '*':
		tok.Type = Mul
		tok.Text = string(l.ch)
	case '/':
		tok.Type = Div
		tok.Text = string(l.ch)
	case '%':
		tok.Type = Mod
		tok.Text = string(l.ch)
	case '(':
		tok.Type = LParen
		tok.Text = string(l.ch)
	case ')':
		tok.Type = RParen
		tok.Text = string(l.ch)
	case '{':
		tok.Type = LBrace
		tok.Text = string(l.ch)
	case '}':
		tok.Type = RBrace
		tok.Text = string(l.ch)
	case '[':
		tok.Type = LBracket
		tok.Text = string(l.ch)
	case ']':
		tok.Type = RBracket
		tok.Text = string(l.ch)
	case ',':
		tok.Type = Comma
		tok.Text = string(l.ch)
	case ';':
		tok.Type = Semicolon
		tok.Text = string(l.ch)
	case '"':
		tok.Type = ConstText
		tok.Text = l.readString()
		return tok, nil
	case 0:
		tok.Type = Invalid
		tok.Text = ""
	default:
		if isLetter(l.ch) {
			tok.Text = l.readIdentifier()
			tok.Type = lookupIdent(tok.Text)
			return tok, nil
		} else if isDigit(l.ch) {
			tok.Text = l.readNumber()
			tok.Type = ConstNum
			return tok, nil
		} else {
			tok.Type = Invalid
			tok.Text = string(l.ch)
			return tok, fmt.Errorf("invalid character '%v' looking for beginning of value at line %d, column %d", l.ch, l.line, l.column)
		}
	}

	l.readChar()
	return tok, nil
}

func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token

	for {
		tok, err := l.NextToken()

		tokens = append(tokens, tok)
		if err != nil {
			return tokens, err
		}
		if tok.Type == Invalid && tok.Text == "" {
			break
		}
	}

	return tokens, nil
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}

var keywords = map[string]TokenType{
	"int":      Int,
	"uint":     Uint,
	"float":    Float,
	"string":   String,
	"bool":     Bool,
	"fn":       Func,
	"if":       If,
	"else":     Else,
	"for":      For,
	"return":   Return,
	"break":    Break,
	"continue": Continue,
	"true":     True,
	"false":    False,
}

func lookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return Identifier
}

func IsTypeToken(tok Token) bool {
	return tok.Type == Int || tok.Type == Uint || tok.Type == Float || tok.Type == String || tok.Type == Bool
}
