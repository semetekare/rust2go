// internal/parser/parser_test.go
package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/semetekare/rust2go/internal/ast"
	"github.com/semetekare/rust2go/internal/lexer"
	"github.com/semetekare/rust2go/internal/parser"
)

// runTestFile считывает файл, токенизирует его и запускает парсер.
func runTestFile(t *testing.T, filename string) (*ast.Crate, []parser.ParseError) {
	t.Helper()
	
	// Путь относительно корня проекта, предполагаем, что тесты запускаются из корня
	// или используем filepath.Join, если тесты запускаются из `internal/parser`.
	// Для надежности используем относительный путь `../../../testdata/...`
	filePath := filepath.Join("..", "..", "testdata", filename) // <--- ИЗМЕНЕНИЕ ЗДЕСЬ
	
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
	return ast, errs
}

// ====================================================================
// ПОЗИТИВНЫЕ ТЕСТЫ (Корректные конструкции)
// ====================================================================

func TestPositiveSyntax(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{"Simple Function and Call", "positive/fn_simple.rs"},
		{"Complex Boolean Expression", "positive/expr_complex.rs"},
		// Добавьте другие корректные файлы для покрытия всех конструкций
		// {"Struct Definition", "positive/struct_def.rs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, errs := runTestFile(t, tt.file)

			// 1. Проверяем отсутствие ошибок
			if len(errs) > 0 {
				t.Errorf("Expected 0 errors, got %d:\n", len(errs))
				for _, err := range errs {
					t.Logf("  %s", err)
				}
				return
			}

			// 2. Проверяем, что AST не пуст
			if ast == nil {
				t.Errorf("AST should not be nil")
				return
			}
			
			// 3. Проверяем структуру (например, количество top-level элементов)
			if len(ast.Items) < 1 {
				t.Errorf("Expected at least 1 item in AST, got 0")
			}
			
			// 4. Дополнительная проверка на конкретные узлы
			// if tt.name == "Simple Function and Call" && len(ast.Items) != 2 {
			//     t.Errorf("Expected 2 functions (add, main), got %d", len(ast.Items))
			// }

			t.Logf("Parsing successful. AST Root: %s", ast.String())
		})
	}
}

// ====================================================================
// НЕГАТИВНЫЕ ТЕСТЫ (Синтаксические ошибки)
// ====================================================================

func TestNegativeSyntax(t *testing.T) {
	tests := []struct {
		name string
		file string
		expectedErrors int // Ожидаемое минимальное количество ошибок
	}{
		{"Missing Semicolon", "negative/missing_semi.rs", 1},
		{"Missing Closing Parenthesis", "negative/missing_paren.rs", 1},
		{"Bad Binary Operator", "negative/bad_operator.rs", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := runTestFile(t, tt.file)

			// Проверяем, что парсер обнаружил ошибки
			if len(errs) == 0 {
				t.Errorf("Expected at least %d errors, but got 0. Test case failed to detect syntax error.", tt.expectedErrors)
				return
			}

			// Проверяем, что количество ошибок соответствует ожиданиям
			if len(errs) < tt.expectedErrors {
				t.Errorf("Expected at least %d errors, but got only %d", tt.expectedErrors, len(errs))
			}
			
			t.Logf("Successfully detected %d error(s). First error: %s", len(errs), errs[0])
		})
	}
}