package main

import (
	"fmt"
	"strings"
	"twin-peaks-programming-language/internal/bytecode"
	"twin-peaks-programming-language/internal/lexer"
	"twin-peaks-programming-language/internal/parser"
	"twin-peaks-programming-language/internal/runtime"
)

const PrintInfo = false

func main() {
	//code, err := io.ReadAll(os.Stdin)
	code := factorial

	l := lexer.NewLexer(string(code))
	tokens, err := l.Tokenize()
	if err != nil {
		fmt.Printf("Lexer error: %v\n", err)
		return
	}

	if PrintInfo {
		fmt.Println("Tokens:")
		tokenStrs := make([]string, len(tokens))
		for i, tok := range tokens {
			tokenStrs[i] = tok.String()
		}
		fmt.Println(strings.Join(tokenStrs, "\n"))
	}

	p := parser.NewParser(tokens)
	ast, err := p.ParseProgram()
	if err != nil {
		fmt.Printf("Parser error: %v\n", err)
		return
	}

	//fmt.Println("\nAST:")
	//fmt.Println(ast.String())

	c := bytecode.NewCompiler()
	bc, err := c.Compile(ast)
	if err != nil {
		fmt.Printf("Compiler error: %v\n", err)
		return
	}

	if PrintInfo {
		fmt.Println("\nBytecode:")
		for i, instr := range bc.Instructions {
			fmt.Printf("%4d: %s\n", i, instr.String())
		}
		fmt.Println("\nConstants:")
		for i, constant := range bc.Constants {
			fmt.Printf("%4d: %v\n", i, constant)
		}
		fmt.Println("\nExecution:")

	}
	virtualMachine := runtime.NewVM(bc, true, PrintInfo)

	if err := virtualMachine.Run(); err != nil {

		fmt.Printf("VM error: %v\n", err)
	}

	if PrintInfo {
		virtualMachine.PrintHeapSize()
	}
}
