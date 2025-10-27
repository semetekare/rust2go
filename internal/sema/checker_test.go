package sema_test

import (
	"testing"

	"github.com/semetekare/rust2go/internal/ast"
	"github.com/semetekare/rust2go/internal/lexer"
	"github.com/semetekare/rust2go/internal/parser"
	"github.com/semetekare/rust2go/internal/sema"
)

// Helper function to parse code and get AST
func parseCode(code string, t *testing.T) *ast.Crate {
	t.Helper()
	lx := lexer.NewLexer()
	toks, err := lx.Lex(code)
	if err != nil {
		t.Fatalf("Lex failed: %v", err)
	}

	p := parser.NewParser(toks)
	crate, errs := p.ParseFile()
	if len(errs) > 0 {
		t.Fatalf("Parse errors: %v", errs)
	}

	return crate
}

func TestCheckerFunctionDeclaration(t *testing.T) {
	code := `
fn add(a: i32, b: i32) -> i32 {
    a + b
}

fn main() {
    let result = add(5, 3);
}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d:\n", len(errors))
		for _, err := range errors {
			t.Logf("  %s", err)
		}
	}
}

func TestCheckerTypeMismatch(t *testing.T) {
	code := `
fn add(a: i32, b: i32) -> i32 {
    a + b
}

fn main() {
    let result: bool = add(5, 3);  // Type mismatch
}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) == 0 {
		t.Error("Expected type mismatch error, got none")
	}
}

func TestCheckerUndefinedVariable(t *testing.T) {
	code := `
fn main() {
    let x = undefined_var;
}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) == 0 {
		t.Error("Expected undefined variable error, got none")
	}
}

func TestCheckerFunctionCallArgsCount(t *testing.T) {
	code := `
fn add(a: i32, b: i32) -> i32 {
    a + b
}

fn main() {
    let result = add(5);  // Wrong number of arguments
}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) == 0 {
		t.Error("Expected function call argument count error, got none")
	}
}

func TestCheckerFunctionCallArgTypes(t *testing.T) {
	code := `
fn add(a: i32, b: i32) -> i32 {
    a + b
}

fn main() {
    let result = add("hello", 3);  // Wrong argument type
}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) == 0 {
		t.Error("Expected function argument type error, got none")
	}
}

func TestCheckerArithmeticTypeCheck(t *testing.T) {
	code := `
fn main() {
    let x = "hello" + "world";  // Can't add strings
}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) == 0 {
		t.Error("Expected arithmetic type error for strings, got none")
	}
}

func TestCheckerComparisonTypeCheck(t *testing.T) {
	code := `
fn main() {
    let result = "hello" == 42;  // Can't compare string with int
}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) == 0 {
		t.Error("Expected comparison type error, got none")
	}
}

func TestCheckerLogicalOpTypeCheck(t *testing.T) {
	code := `
fn main() {
    let result = 42 && 10;  // Logical ops need bool
}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) == 0 {
		t.Error("Expected logical operation type error, got none")
	}
}

func TestCheckerUnaryOpTypeCheck(t *testing.T) {
	code := `
fn main() {
    let result = -42;  // Negation of int should work
}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	// Negation of int should be valid, no error expected
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d:\n", len(errors))
		for _, err := range errors {
			t.Logf("  %s", err)
		}
	}
}

func TestCheckerMacroSupport(t *testing.T) {
	code := `
fn main() {
    println!("Hello, World!");
}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) > 0 {
		t.Errorf("Expected no errors with println!, got %d:\n", len(errors))
		for _, err := range errors {
			t.Logf("  %s", err)
		}
	}
}

func TestCheckerTypeInference(t *testing.T) {
	code := `
fn main() {
    let x = 42;      // Infer i32
    let y = 3.14;    // Infer f64 (if supported)
    let z = true;    // Infer bool
}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) > 0 {
		t.Errorf("Expected no errors with type inference, got %d:\n", len(errors))
		for _, err := range errors {
			t.Logf("  %s", err)
		}
	}
}

func TestCheckerDuplicateFunction(t *testing.T) {
	code := `
fn foo() {}
fn foo() {}  // Duplicate
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) == 0 {
		t.Error("Expected duplicate function error, got none")
	}
}

func TestCheckerCorrectExpressions(t *testing.T) {
	code := `
fn main() {
    let a = 5 + 3;        // Correct arithmetic
    let b = 5 < 10;       // Correct comparison
    let c = true && false; // Correct logical
    let d = -42;          // Correct unary
}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) > 0 {
		t.Errorf("Expected no errors with correct expressions, got %d:\n", len(errors))
		for _, err := range errors {
			t.Logf("  %s", err)
		}
	}
}

func TestCheckerNestedExpressions(t *testing.T) {
	code := `
fn main() {
    let result = (1 + 2) * (3 + 4);
}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) > 0 {
		t.Errorf("Expected no errors with nested expressions, got %d:\n", len(errors))
		for _, err := range errors {
			t.Logf("  %s", err)
		}
	}
}

func TestCheckerComplexFunction(t *testing.T) {
	code := `
fn add(a: i32, b: i32) -> i32 {
    a + b
}

fn multiply(a: i32, b: i32) -> i32 {
    a * b
}

fn main() {
    let x = add(1, 2);
    let y = multiply(3, 4);
    let z = add(x, y);
}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) > 0 {
		t.Errorf("Expected no errors with complex function calls, got %d:\n", len(errors))
		for _, err := range errors {
			t.Logf("  %s", err)
		}
	}
}

func TestCheckerPositionTracking(t *testing.T) {
	code := `
fn main() {
    let x = undefined;
}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) == 0 {
		t.Fatal("Expected error")
	}

	// Проверяем, что позиция указана
	err := errors[0]
	if err.Pos.Line == 0 || err.Pos.Col == 0 {
		t.Error("Error position should be valid")
	}
}

func TestCheckerEmptyFunction(t *testing.T) {
	code := `
fn main() {}
`
	ast := parseCode(code, t)
	checker := sema.NewChecker()
	errors := checker.Check(ast)

	if len(errors) > 0 {
		t.Errorf("Expected no errors with empty function, got %d:\n", len(errors))
		for _, err := range errors {
			t.Logf("  %s", err)
		}
	}
}
