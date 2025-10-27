// CLI точка входа для rust2go.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/semetekare/rust2go/internal/ast"
	"github.com/semetekare/rust2go/internal/backend"
	"github.com/semetekare/rust2go/internal/ir"
	"github.com/semetekare/rust2go/internal/lexer"
	"github.com/semetekare/rust2go/internal/parser"
	"github.com/semetekare/rust2go/internal/sema"
)

// main — точка входа для полного pipeline компиляции.
// CLI: go run ./cmd/main.go example/example.rs
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: rust2go <file.rs>")
		os.Exit(1)
	}
	inputFile := os.Args[1]
	b, err := os.ReadFile(inputFile)
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
		fmt.Println("✓ Parsing succeeded")
		fmt.Println("AST:", ast.PrettyPrint(fileAST))

		// Семантический анализ
		fmt.Println("\n=== Semantic Analysis ===")
		checker := sema.NewChecker()
		semErrs := checker.Check(fileAST)
		if len(semErrs) > 0 {
			fmt.Printf("✗ Found %d semantic error(s):\n", len(semErrs))
			for _, e := range semErrs {
				fmt.Println("  ", e)
			}
			os.Exit(1)
		}
		fmt.Println("✓ Semantic analysis passed")

		// Трансформация в IR
		fmt.Println("\n=== IR Transformation ===")
		transformer := ir.NewTransformer()
		irModule := transformer.Transform(fileAST)
		fmt.Printf("✓ Transformed to IR: %d functions, %d structs\n",
			len(irModule.Functions), len(irModule.Structs))

		// Генерация кода
		fmt.Println("\n=== Code Generation ===")
		gen := backend.NewGenerator()
		goCode := gen.Generate(irModule)

		fmt.Println("Generated Go code:")
		fmt.Println("---")
		fmt.Println(goCode)
		fmt.Println("---")

		// Сохраняем сгенерированный код в output/
		outputDir := "output"
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Warning: could not create output directory: %v\n", err)
		}

		// Имя выходного файла на основе входного
		baseName := filepath.Base(inputFile)
		ext := filepath.Ext(baseName)
		outputFile := filepath.Join(outputDir, baseName[:len(baseName)-len(ext)]+".go")
		if err := os.WriteFile(outputFile, []byte(goCode), 0644); err != nil {
			fmt.Printf("Warning: could not write %s: %v\n", outputFile, err)
		} else {
			fmt.Printf("\n✓ Code written to %s\n", outputFile)
		}
	}
}
