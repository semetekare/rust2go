// CLI точка входа для rust2go.
package main

import (
	"fmt"
	"os"

	"github.com/semetekare/rust2go/internal/lexer"
	"github.com/semetekare/rust2go/internal/token"
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
	for _, t := range toks {
		if t.Type == token.EOF {
			break
		}
		fmt.Printf("%s: %q at %d:%d\n", t.String(), t.Literal, t.Line, t.Col)
	}
}
