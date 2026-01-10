package main

import (
	"fmt"
	"strings"
	"twin-peaks-programming-language/internal/lexer"
	"twin-peaks-programming-language/internal/parser"
)

func main() {
	code := factorial
	l := lexer.NewLexer(code)
	tokens, err := l.Tokenize()
	tokStrs := make([]string, len(tokens))
	for i, tok := range tokens {
		tokStrs[i] = tok.String()
	}
	fmt.Println(strings.Join(tokStrs, "\n"), err)
	if err != nil {
		return
	}

	p := parser.NewParser(tokens)
	ast, err := p.ParseProgram()
	if err != nil {
		fmt.Printf("Parser error: %v\n", err)
		return
	}

	// Вывод AST
	fmt.Println(ast.String())
}
