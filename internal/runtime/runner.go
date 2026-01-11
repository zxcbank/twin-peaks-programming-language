package runtime

import (
	"fmt"
	"twin-peaks-programming-language/internal/parser"
)

type Runner struct {
	identifiers map[string]interface{}
	ast         *parser.ASTNode
}

func NewRunner(tree *parser.ASTNode) *Runner {
	return &Runner{
		identifiers: make(map[string]interface{}),
		ast:         tree,
	}
}

func (r *Runner) Run() {
	for _, node := range r.ast.Children {
		r.runNode(node)
	}
}

func (r *Runner) runNode(node *parser.ASTNode) {
	fmt.Printf("Running node %v", node.Name()) // TODO:
	for _, node := range node.Children {
		r.runNode(node)
	}

}
