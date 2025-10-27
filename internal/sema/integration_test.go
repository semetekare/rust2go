package sema_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/semetekare/rust2go/internal/ast"
	"github.com/semetekare/rust2go/internal/lexer"
	"github.com/semetekare/rust2go/internal/parser"
	"github.com/semetekare/rust2go/internal/sema"
)

// runTestFileFromTestdata считывает файл из testdata и запускает семантический анализ.
func runTestFileFromTestdata(t *testing.T, filename string) (*ast.Crate, []sema.SemanticError) {
	t.Helper()

	filePath := filepath.Join("..", "..", "testdata", filename)

	b, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filePath, err)
	}

	lx := lexer.NewLexer()
	toks, err := lx.Lex(string(b))
	if err != nil {
		t.Fatalf("Lexing failed for %s: %v", filename, err)
	}

	p := parser.NewParser(toks)
	ast, errs := p.ParseFile()
	if len(errs) > 0 {
		t.Fatalf("Parse errors: %v", errs)
	}

	checker := sema.NewChecker()
	semErrs := checker.Check(ast)
	return ast, semErrs
}

func TestSemaPositiveFiles(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{"Simple Function and Call", "positive/fn_simple.rs"},
		{"Complex Boolean Expression", "positive/expr_complex.rs"},
		{"Multiple Functions", "positive/multiple_functions.rs"},
		{"Nested Expressions", "positive/nested_expressions.rs"},
		{"Comparison Operations", "positive/comparison_ops.rs"},
		{"Logical Operations", "positive/logical_ops.rs"},
		{"Unary Operations", "positive/unary_ops.rs"},
		{"Macro Calls", "positive/macro_calls.rs"},
		{"Type Inference", "positive/type_inference.rs"},
		{"Arithmetic Operations", "positive/arithmetic.rs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := runTestFileFromTestdata(t, tt.file)

			if len(errs) > 0 {
				t.Errorf("Expected no semantic errors, got %d:\n", len(errs))
				for _, err := range errs {
					t.Logf("  %s", err)
				}
			}
		})
	}
}

func TestSemaNegativeFiles(t *testing.T) {
	tests := []struct {
		name           string
		file           string
		expectErrors   bool
		expectedErrors int
	}{
		{"Undefined Variable", "negative/undefined_var.rs", true, 1},
		{"Wrong Argument Count", "negative/wrong_arg_count.rs", true, 1},
		{"Type Mismatch", "negative/type_mismatch.rs", true, 1},
		{"Operator Type Error", "negative/operator_type_error.rs", true, 1},
		{"Logical Type Error", "negative/logical_type_error.rs", true, 1},
		{"Duplicate Function", "negative/duplicate_function.rs", true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := runTestFileFromTestdata(t, tt.file)

			if tt.expectErrors && len(errs) == 0 {
				t.Errorf("Expected semantic errors, got none")
			}

			if !tt.expectErrors && len(errs) > 0 {
				t.Errorf("Expected no semantic errors, got %d:\n", len(errs))
				for _, err := range errs {
					t.Logf("  %s", err)
				}
			}

			if len(errs) < tt.expectedErrors {
				t.Errorf("Expected at least %d errors, got %d", tt.expectedErrors, len(errs))
			}
		})
	}
}

func TestSemaAllPositiveCases(t *testing.T) {
	// Тест, который проходит по всем позитивным файлам и проверяет, что все они проходят семантический анализ
	testDir := filepath.Join("..", "..", "testdata", "positive")
	files, err := os.ReadDir(testDir)
	if err != nil {
		t.Fatalf("Failed to read testdata/positive: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".rs" {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			_, errs := runTestFileFromTestdata(t, "positive/"+file.Name())

			if len(errs) > 0 {
				t.Errorf("Expected no semantic errors for %s, got %d:\n", file.Name(), len(errs))
				for _, err := range errs {
					t.Logf("  %s", err)
				}
			}
		})
	}
}
