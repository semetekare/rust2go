// Package ir предоставляет функции для преобразования AST в IR.
package ir

import (
	"github.com/semetekare/rust2go/internal/ast"
)

// Transformer преобразует AST в промежуточное представление.
type Transformer struct {
	module *Module
}

// NewTransformer создаёт новый трансформер.
func NewTransformer() *Transformer {
	return &Transformer{
		module: &Module{
			Name:        "main",
			PackageName: "main",
			Functions:   []*Function{},
			Structs:     []*Struct{},
		},
	}
}

// Transform преобразует AST-код в IR-модуль.
func (t *Transformer) Transform(crate *ast.Crate) *Module {
	for _, item := range crate.Items {
		switch node := item.(type) {
		case *ast.Function:
			fn := t.transformFunction(node)
			if fn != nil {
				t.module.Functions = append(t.module.Functions, fn)
			}
		case *ast.Struct:
			st := t.transformStruct(node)
			if st != nil {
				t.module.Structs = append(t.module.Structs, st)
			}
		}
	}
	return t.module
}

// transformFunction преобразует AST-функцию в IR-функцию.
func (t *Transformer) transformFunction(fn *ast.Function) *Function {
	if fn.Body == nil {
		return nil
	}

	irFunc := &Function{
		Name:       fn.Name,
		Params:     []*Parameter{},
		ReturnType: t.transformType(fn.ReturnType),
		Body:       []Statement{},
		Pos:        fn.Pos(),
		GoPackage:  "main",
	}

	// Преобразуем параметры
	for _, param := range fn.Params {
		irFunc.Params = append(irFunc.Params, &Parameter{
			Name: param.Name,
			Type: t.transformType(param.Type),
		})
	}

	// Преобразуем тело функции
	for _, stmt := range fn.Body.Stmts {
		irStmt := t.transformStmt(stmt)
		if irStmt != nil {
			irFunc.Body = append(irFunc.Body, irStmt)
		}
	}

	return irFunc
}

// transformStmt преобразует AST-оператор в IR-оператор.
func (t *Transformer) transformStmt(stmt ast.Stmt) Statement {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		return &Declaration{
			Name:      s.Name,
			Type:      t.transformType(s.Type),
			InitValue: t.transformExpr(s.Init),
			Position:  s.Pos(),
		}
	case *ast.ExprStmt:
		return &ExprStmt{
			Expr:     t.transformExpr(s.Expr),
			Position: s.Pos(),
		}
	}
	return nil
}

// transformExpr преобразует AST-выражение в IR-выражение.
func (t *Transformer) transformExpr(expr ast.Expr) Expression {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.Literal:
		return &LiteralExpr{
			Value:    e.Val,
			Kind:     e.Kind,
			TypeInfo: t.getLiteralType(e),
			Position: e.Pos(),
		}
	case *ast.BlockExpr:
		// Пока пропускаем block expressions
		return nil
	case *ast.BinaryExpr:
		left := t.transformExpr(e.Left)
		right := t.transformExpr(e.Right)
		return &BinaryExpr{
			Left:     left,
			Op:       e.Op,
			Right:    right,
			TypeInfo: left.Type(), // Упрощённо берём тип слева
			Position: e.Pos(),
		}
	case *ast.UnaryExpr:
		return &UnaryExpr{
			Op:       e.Op,
			Expr:     t.transformExpr(e.Expr),
			TypeInfo: t.transformExpr(e.Expr).Type(),
			Position: e.Pos(),
		}
	case *ast.CallExpr:
		// Получаем имя функции из литерала
		var funcName string
		if lit, ok := e.Func.(*ast.Literal); ok {
			funcName = lit.Val
		}

		args := []Expression{}
		for _, arg := range e.Args {
			args = append(args, t.transformExpr(arg))
		}

		isMacro := len(funcName) > 0 && funcName[len(funcName)-1] == '!'
		var returnType *Type

		// Определяем возвращаемый тип для макросов
		if isMacro {
			switch funcName {
			case "format!":
				returnType = NewType("string", true)
			default:
				returnType = NewType("()", true)
			}
		} else {
			// Для обычных функций пока возвращаем unit
			returnType = NewType("()", true)
		}

		return &CallExpr{
			FuncName: funcName,
			Args:     args,
			TypeInfo: returnType,
			Position: e.Pos(),
			IsMacro:  isMacro,
		}
	}
	return nil
}

// transformType преобразует AST-тип в IR-тип.
func (t *Transformer) transformType(astType ast.Type) *Type {
	if astType == nil {
		return NewType("", true)
	}

	switch typ := astType.(type) {
	case *ast.PathType:
		typeName := MapRustToGoType(typ.Path)
		return NewType(typeName, true)
	}
	return NewType("interface{}", false)
}

// getLiteralType определяет тип литерала.
func (t *Transformer) getLiteralType(lit *ast.Literal) *Type {
	switch lit.Kind {
	case "INT":
		return NewType("int", true)
	case "FLOAT":
		return NewType("float64", true)
	case "STRING":
		return NewType("string", true)
	case "BOOL":
		return NewType("bool", true)
	case "IDENT":
		// Для идентификаторов - возвращаем тип с именем
		return NewType(lit.Val, false)
	default:
		return NewType("interface{}", false)
	}
}

// transformStruct преобразует AST-структуру в IR-структуру.
func (t *Transformer) transformStruct(st *ast.Struct) *Struct {
	if st == nil {
		return nil
	}

	irStruct := &Struct{
		Name:   st.Name,
		Fields: []*Field{},
		Pos:    st.Pos(),
	}

	for _, field := range st.Fields {
		irStruct.Fields = append(irStruct.Fields, &Field{
			Name: field.Name,
			Type: t.transformType(field.Type),
		})
	}

	return irStruct
}
