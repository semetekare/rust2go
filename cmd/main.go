// CLI точка входа для rust2go.
package main

import (
	"fmt"
	"os"

	"github.com/semetekare/rust2go/internal/ast"
	"github.com/semetekare/rust2go/internal/lexer"
	"github.com/semetekare/rust2go/internal/parser"
	"github.com/semetekare/rust2go/internal/sema"
)

// main — пример использования лексера: читает файл и печатает токены.
// CLI: go run main.go example.rs
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: rust2go <file.rs>")
		os.Exit(1)
	}
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("read error: %v\n", err)
		os.Exit(1)
	}
	lx := lexer.NewLexer()
	toks, err := lx.Lex(string(b))
	if err != nil {
		fmt.Printf("lex error: %v\n", err)
		os.Exit(1)
	}
	p := parser.NewParser(toks)
	fileAST, errs := p.ParseFile()
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Println(e)
		}
	} else {
		fmt.Println("Parsing succeeded, AST:")
		fmt.Println(ast.PrettyPrint(fileAST))
		
		// Семантический анализ
		fmt.Println("\n=== Semantic Analysis ===")
		checker := sema.NewChecker()
		semErrs := checker.Check(fileAST)
		if len(semErrs) > 0 {
			fmt.Printf("Found %d semantic error(s):\n", len(semErrs))
			for _, e := range semErrs {
				fmt.Println(e)
			}
		} else {
			fmt.Println("Semantic analysis passed successfully!")
		}
	}
}
