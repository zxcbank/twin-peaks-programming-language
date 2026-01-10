package main

import (
	"fmt"
	"twin-peaks-programming-language/internal/lexer"
)

func main() {
	code := `fn mamba(x int, y int){return x+y*x;}; //хуета
f int;
f = 3 + 9;
print(f);`
	nl := lexer.NewLexer(code)
	fmt.Println(nl.Tokenize())
}
