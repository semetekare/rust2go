package ast_test

import (
	"strings"
	"testing"

	"github.com/semetekare/rust2go/internal/ast"
	"github.com/semetekare/rust2go/internal/token"
)

func TestNewCrate(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}
	crate := ast.NewCrate(pos, []ast.Item{})

	if crate == nil {
		t.Fatal("Expected crate to be non-nil")
	}
	if crate.Pos().Line != 1 {
		t.Errorf("Expected line 1, got %d", crate.Pos().Line)
	}
	if len(crate.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(crate.Items))
	}
}

func TestNewFunction(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}
	retType := ast.NewPathType(pos, "i32")
	params := []ast.Param{
		*ast.NewParam(pos, "a", ast.NewPathType(pos, "i32")),
		*ast.NewParam(pos, "b", ast.NewPathType(pos, "i32")),
	}
	body := ast.NewBlock(pos, []ast.Stmt{})

	fn := ast.NewFunction(pos, "add", params, retType, body)

	if fn == nil {
		t.Fatal("Expected function to be non-nil")
	}
	if fn.Name != "add" {
		t.Errorf("Expected name 'add', got %q", fn.Name)
	}
	if len(fn.Params) != 2 {
		t.Errorf("Expected 2 params, got %d", len(fn.Params))
	}
}

func TestNewStruct(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}
	fields := []ast.Field{
		*ast.NewField(pos, "x", ast.NewPathType(pos, "i32")),
		*ast.NewField(pos, "y", ast.NewPathType(pos, "i32")),
	}

	st := ast.NewStruct(pos, "Point", fields)

	if st == nil {
		t.Fatal("Expected struct to be non-nil")
	}
	if st.Name != "Point" {
		t.Errorf("Expected name 'Point', got %q", st.Name)
	}
	if len(st.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(st.Fields))
	}
}

func TestNewLetStmt(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}
	typ := ast.NewPathType(pos, "i32")
	init := ast.NewLiteral(pos, "INT", "42")
	stmt := ast.NewLetStmt(pos, "x", typ, init)

	if stmt == nil {
		t.Fatal("Expected let statement to be non-nil")
	}
	if stmt.Name != "x" {
		t.Errorf("Expected name 'x', got %q", stmt.Name)
	}
}

func TestNewExprStmt(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}
	expr := ast.NewLiteral(pos, "INT", "42")
	stmt := ast.NewExprStmt(pos, expr)

	if stmt == nil {
		t.Fatal("Expected expression statement to be non-nil")
	}
	if stmt.Expr == nil {
		t.Error("Expected expression to be non-nil")
	}
}

func TestNewBlock(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}
	block := ast.NewBlock(pos, []ast.Stmt{})

	if block == nil {
		t.Fatal("Expected block to be non-nil")
	}
	if len(block.Stmts) != 0 {
		t.Errorf("Expected 0 statements, got %d", len(block.Stmts))
	}
}

func TestNewLiteral(t *testing.T) {
	tests := []struct {
		kind     string
		val      string
		expected string
	}{
		{"INT", "42", "42"},
		{"STRING", "hello", "hello"},
		{"BOOL", "true", "true"},
	}

	pos := token.Position{Line: 1, Col: 1}
	for _, tt := range tests {
		lit := ast.NewLiteral(pos, tt.kind, tt.val)
		if lit.Val != tt.expected {
			t.Errorf("Expected value %q, got %q", tt.expected, lit.Val)
		}
		if lit.Kind != tt.kind {
			t.Errorf("Expected kind %q, got %q", tt.kind, lit.Kind)
		}
	}
}

func TestNewBinaryExpr(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}
	left := ast.NewLiteral(pos, "INT", "5")
	right := ast.NewLiteral(pos, "INT", "3")

	expr := ast.NewBinaryExpr(pos, left, "+", right)

	if expr == nil {
		t.Fatal("Expected binary expression to be non-nil")
	}
	if expr.Op != "+" {
		t.Errorf("Expected op '+', got %q", expr.Op)
	}
	if expr.Left == nil || expr.Right == nil {
		t.Error("Expected left and right to be non-nil")
	}
}

func TestNewUnaryExpr(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}
	expr := ast.NewLiteral(pos, "INT", "42")
	unary := ast.NewUnaryExpr(pos, "-", expr)

	if unary == nil {
		t.Fatal("Expected unary expression to be non-nil")
	}
	if unary.Op != "-" {
		t.Errorf("Expected op '-', got %q", unary.Op)
	}
}

func TestNewCallExpr(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}
	fn := ast.NewLiteral(pos, "IDENT", "add")
	args := []ast.Expr{
		ast.NewLiteral(pos, "INT", "1"),
		ast.NewLiteral(pos, "INT", "2"),
	}

	call := ast.NewCallExpr(pos, fn, args)

	if call == nil {
		t.Fatal("Expected call expression to be non-nil")
	}
	if len(call.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(call.Args))
	}
}

func TestNewPathType(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}
	typ := ast.NewPathType(pos, "i32")

	if typ == nil {
		t.Fatal("Expected path type to be non-nil")
	}
	if typ.Path != "i32" {
		t.Errorf("Expected path 'i32', got %q", typ.Path)
	}
}

func TestNewParam(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}
	typ := ast.NewPathType(pos, "i32")
	param := ast.NewParam(pos, "x", typ)

	if param == nil {
		t.Fatal("Expected param to be non-nil")
	}
	if param.Name != "x" {
		t.Errorf("Expected name 'x', got %q", param.Name)
	}
}

func TestNewField(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}
	typ := ast.NewPathType(pos, "i32")
	field := ast.NewField(pos, "x", typ)

	if field == nil {
		t.Fatal("Expected field to be non-nil")
	}
	if field.Name != "x" {
		t.Errorf("Expected name 'x', got %q", field.Name)
	}
}

func TestStringMethods(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}

	tests := []struct {
		name     string
		node     ast.Node
		expected string
	}{
		{
			"Crate",
			ast.NewCrate(pos, []ast.Item{}),
			"Crate{Items: 0}",
		},
		{
			"Function",
			ast.NewFunction(pos, "foo", []ast.Param{}, nil, nil),
			"Function{Name: foo}",
		},
		{
			"Struct",
			ast.NewStruct(pos, "Foo", []ast.Field{}),
			"Struct{Name: Foo}",
		},
		{
			"Literal",
			ast.NewLiteral(pos, "INT", "42"),
			"Literal{INT: 42}",
		},
		{
			"Field",
			ast.NewField(pos, "x", nil),
			"Field{Name: x}",
		},
		{
			"Param",
			ast.NewParam(pos, "x", nil),
			"Param{Name: x}",
		},
	}

	for _, tt := range tests {
		str := tt.node.String()
		if !strings.Contains(str, tt.expected) {
			t.Errorf("%s: Expected substring %q in %q", tt.name, tt.expected, str)
		}
	}
}

func TestPrettyPrint(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}
	fn := ast.NewFunction(
		pos,
		"main",
		[]ast.Param{},
		ast.NewPathType(pos, "()"),
		ast.NewBlock(pos, []ast.Stmt{}),
	)
	crate := ast.NewCrate(pos, []ast.Item{fn})

	output := ast.PrettyPrint(crate)
	if !strings.Contains(output, "main") {
		t.Errorf("Expected 'main' in output, got %q", output)
	}
}

func TestInterfaceImplementation(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}

	// Проверяем, что все типы реализуют интерфейсы
	var items []ast.Item
	var stmts []ast.Stmt
	var exprs []ast.Expr
	var types []ast.Type

	fn := ast.NewFunction(pos, "test", []ast.Param{}, nil, nil)
	st := ast.NewStruct(pos, "Test", []ast.Field{})
	ls := ast.NewLetStmt(pos, "x", nil, ast.NewLiteral(pos, "INT", "1"))
	es := ast.NewExprStmt(pos, ast.NewLiteral(pos, "INT", "1"))
	blk := ast.NewBlock(pos, []ast.Stmt{})
	_ = ast.NewBlockExpr(pos, blk)

	items = append(items, fn, st)
	stmts = append(stmts, ls, es, blk)
	exprs = append(exprs, ast.NewLiteral(pos, "INT", "1"), ast.NewBinaryExpr(pos, nil, "+", nil), ast.NewUnaryExpr(pos, "-", nil), ast.NewCallExpr(pos, nil, nil), ast.NewBlockExpr(pos, blk))
	types = append(types, ast.NewPathType(pos, "i32"))

	_ = items
	_ = stmts
	_ = exprs
	_ = types
}

func TestPrettyPrintComplex(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}

	// Создаём сложную структуру AST для полного покрытия prettyPrintNode
	fn := ast.NewFunction(
		pos,
		"complex",
		[]ast.Param{
			*ast.NewParam(pos, "a", ast.NewPathType(pos, "i32")),
			*ast.NewParam(pos, "b", ast.NewPathType(pos, "i32")),
		},
		ast.NewPathType(pos, "i32"),
		ast.NewBlock(pos, []ast.Stmt{
			ast.NewLetStmt(pos, "x", ast.NewPathType(pos, "i32"), ast.NewLiteral(pos, "INT", "5")),
			ast.NewExprStmt(pos, ast.NewBinaryExpr(
				pos,
				ast.NewLiteral(pos, "IDENT", "a"),
				"+",
				ast.NewLiteral(pos, "IDENT", "b"),
			)),
		}),
	)

	st := ast.NewStruct(
		pos,
		"Point",
		[]ast.Field{
			*ast.NewField(pos, "x", ast.NewPathType(pos, "i32")),
			*ast.NewField(pos, "y", ast.NewPathType(pos, "i32")),
		},
	)

	crate := ast.NewCrate(pos, []ast.Item{fn, st})

	output := ast.PrettyPrint(crate)

	// Проверяем, что вывод содержит основные элементы
	if !strings.Contains(output, "complex") {
		t.Error("Expected 'complex' in output")
	}
	if !strings.Contains(output, "Point") {
		t.Error("Expected 'Point' in output")
	}
	if !strings.Contains(output, "a") {
		t.Error("Expected 'a' in output")
	}
}

func TestPrettyPrintUnaryExpr(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}

	unary := ast.NewUnaryExpr(pos, "-", ast.NewLiteral(pos, "INT", "42"))
	crate := ast.NewCrate(pos, []ast.Item{
		ast.NewFunction(pos, "test", []ast.Param{}, nil, ast.NewBlock(pos, []ast.Stmt{
			ast.NewExprStmt(pos, unary),
		})),
	})

	output := ast.PrettyPrint(crate)
	if !strings.Contains(output, "UnaryExpr") {
		t.Error("Expected UnaryExpr in output")
	}
}

func TestPrettyPrintCallExpr(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}

	call := ast.NewCallExpr(
		pos,
		ast.NewLiteral(pos, "IDENT", "foo"),
		[]ast.Expr{
			ast.NewLiteral(pos, "INT", "1"),
			ast.NewLiteral(pos, "INT", "2"),
		},
	)

	crate := ast.NewCrate(pos, []ast.Item{
		ast.NewFunction(pos, "test", []ast.Param{}, nil, ast.NewBlock(pos, []ast.Stmt{
			ast.NewExprStmt(pos, call),
		})),
	})

	output := ast.PrettyPrint(crate)
	if !strings.Contains(output, "CallExpr") {
		t.Error("Expected CallExpr in output")
	}
}

func TestPrettyPrintBlockExpr(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}

	block := ast.NewBlock(pos, []ast.Stmt{
		ast.NewLetStmt(pos, "x", nil, ast.NewLiteral(pos, "INT", "1")),
	})
	_ = ast.NewBlockExpr(pos, block)

	crate := ast.NewCrate(pos, []ast.Item{
		ast.NewFunction(pos, "test", []ast.Param{}, nil, block),
	})

	output := ast.PrettyPrint(crate)
	if !strings.Contains(output, "Block") {
		t.Error("Expected Block in output")
	}
}

func TestPrettyPrintNestedExpressions(t *testing.T) {
	pos := token.Position{Line: 1, Col: 1}

	// Создаём вложенные выражения
	inner := ast.NewBinaryExpr(
		pos,
		ast.NewLiteral(pos, "INT", "1"),
		"+",
		ast.NewLiteral(pos, "INT", "2"),
	)
	outer := ast.NewBinaryExpr(
		pos,
		inner,
		"*",
		ast.NewLiteral(pos, "INT", "3"),
	)

	crate := ast.NewCrate(pos, []ast.Item{
		ast.NewFunction(pos, "test", []ast.Param{}, nil, ast.NewBlock(pos, []ast.Stmt{
			ast.NewExprStmt(pos, outer),
		})),
	})

	output := ast.PrettyPrint(crate)
	if !strings.Contains(output, "BinaryExpr") {
		t.Error("Expected BinaryExpr in output")
	}
}
