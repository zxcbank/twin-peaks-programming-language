package main

import (
	"fmt"
	"strings"
	"twin-peaks-programming-language/internal/code_execution/bytecode"
	"twin-peaks-programming-language/internal/lexer"
	"twin-peaks-programming-language/internal/parser"
)

func main() {
	//code, err := io.ReadAll(os.Stdin)
	code := sieve_of_eratosthenes
	// Лексический анализ
	l := lexer.NewLexer(string(code))
	tokens, err := l.Tokenize()
	if err != nil {
		fmt.Printf("Lexer error: %v\n", err)
		return
	}

	// Вывод токенов
	fmt.Println("Tokens:")
	tokenStrs := make([]string, len(tokens))
	for i, tok := range tokens {
		tokenStrs[i] = tok.String()
	}
	fmt.Println(strings.Join(tokenStrs, "\n"))

	// Синтаксический анализ
	p := parser.NewParser(tokens)
	ast, err := p.ParseProgram()
	if err != nil {
		fmt.Printf("Parser error: %v\n", err)
		return
	}

	// Вывод AST
	fmt.Println("\nAST:")
	fmt.Println(ast.String())

	// Компиляция в байт-код
	c := bytecode.NewCompiler()
	bc, err := c.Compile(ast)
	if err != nil {
		fmt.Printf("Compiler error: %v\n", err)
		return
	}
	//
	//// Вывод байт-кода
	fmt.Println("\nBytecode:")
	for i, instr := range bc.Instructions {
		fmt.Printf("%4d: %s\n", i, instr.String())
	}
	//
	//// Вывод констант
	fmt.Println("\nConstants:")
	for i, constant := range bc.Constants {
		fmt.Printf("%4d: %v\n", i, constant)
	}

	//
	//// Выполнение на виртуальной машине
	fmt.Println("\nExecution:")
	virtualMachine := bytecode.NewVM(bc, true)

	if err := virtualMachine.Run(); err != nil {

		fmt.Printf("VM error: %v\n", err)
	}

	virtualMachine.PrintHeapSize()
}
