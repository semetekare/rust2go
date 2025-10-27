// Package backend реализует генератор кода для целевого языка (Go).
package backend

import (
	"fmt"
	"strings"

	"github.com/semetekare/rust2go/internal/ir"
)

// Generator генерирует код на Go из IR.
type Generator struct {
	builder strings.Builder
	indent  int
}

// NewGenerator создаёт новый генератор.
func NewGenerator() *Generator {
	return &Generator{
		indent: 0,
	}
}

// Generate генерирует код Go из IR модуля.
func (g *Generator) Generate(module *ir.Module) string {
	g.builder.Reset()

	// Генерируем заголовок пакета
	g.emit("package %s", module.PackageName)
	g.emit("")
	g.emit("import (")
	g.indent++
	g.emit("\"fmt\"")
	g.emit("// Add more imports as needed")
	g.indent--
	g.emit(")")
	g.emit("")

	// Генерируем структуры
	for _, st := range module.Structs {
		g.generateStruct(st)
		g.emit("")
	}

	// Генерируем функции
	for _, fn := range module.Functions {
		g.generateFunction(fn)
		g.emit("")
	}

	return g.builder.String()
}

// generateStruct генерирует определение структуры на Go.
func (g *Generator) generateStruct(st *ir.Struct) {
	g.emit("type %s struct {", st.Name)
	g.indent++
	for _, field := range st.Fields {
		g.emit("%s %s", capitalize(field.Name), field.Type.String())
	}
	g.indent--
	g.emit("}")
}

// generateFunction генерирует функцию на Go.
func (g *Generator) generateFunction(fn *ir.Function) {
	// Сигнатура функции
	params := g.generateParams(fn.Params)
	var returnType string
	if fn.ReturnType != nil && fn.ReturnType.Name != "" && fn.ReturnType.Name != "()" {
		returnType = fmt.Sprintf(" %s", fn.ReturnType.String())
	}

	g.emit("func %s(%s)%s {", fn.Name, params, returnType)
	g.indent++

	// Проверяем, есть ли явный return
	hasReturn := false
	for _, stmt := range fn.Body {
		if _, ok := stmt.(*ir.Return); ok {
			hasReturn = true
			break
		}
	}

	// Генерируем тело функции
	for i, stmt := range fn.Body {
		// Если это последний ExprStmt в функции с возвращаемым значением,
		// и нет явного return, преобразуем его в return
		isLastStmt := i == len(fn.Body)-1
		if !hasReturn && isLastStmt && fn.ReturnType != nil && fn.ReturnType.Name != "" && fn.ReturnType.Name != "()" {
			if exprStmt, ok := stmt.(*ir.ExprStmt); ok {
				exprStr := g.generateExpression(exprStmt.Expr)
				if exprStr != "" {
					g.emit("return %s", exprStr)
					g.indent--
					g.emit("}")
					return
				}
			}
		}
		g.generateStatement(stmt)
	}

	// Если нет явного return и функция не void, добавляем пустой return
	if fn.ReturnType != nil && fn.ReturnType.Name != "" && fn.ReturnType.Name != "()" && !hasReturn {
		// Проверяем, не добавили ли мы уже return выше
		if len(fn.Body) == 0 || len(fn.Body) > 0 {
			lastStmt := fn.Body[len(fn.Body)-1]
			if _, ok := lastStmt.(*ir.ExprStmt); !ok {
				g.emit("return // TODO: add return value")
			}
		}
	}

	g.indent--
	g.emit("}")
}

// generateParams генерирует список параметров.
func (g *Generator) generateParams(params []*ir.Parameter) string {
	if len(params) == 0 {
		return ""
	}

	parts := []string{}
	for _, param := range params {
		parts = append(parts, fmt.Sprintf("%s %s", param.Name, param.Type.String()))
	}
	return strings.Join(parts, ", ")
}

// generateStatement генерирует оператор Go.
func (g *Generator) generateStatement(stmt ir.Statement) {
	switch s := stmt.(type) {
	case *ir.Declaration:
		// Упрощённая генерация: используем :=
		exprStr := g.generateExpression(s.InitValue)
		if exprStr != "" {
			g.emit("%s := %s", s.Name, exprStr)
		} else if s.Type != nil {
			g.emit("var %s %s", s.Name, s.Type.String())
		}
	case *ir.Assignment:
		g.emit("%s = %s", s.Target, g.generateExpression(s.Value))
	case *ir.Return:
		if s.Value != nil {
			g.emit("return %s", g.generateExpression(s.Value))
		} else {
			g.emit("return")
		}
	case *ir.ExprStmt:
		exprStr := g.generateExpression(s.Expr)
		g.emit("%s", exprStr)
	}
}

// generateExpression генерирует выражение Go.
func (g *Generator) generateExpression(expr ir.Expression) string {
	if expr == nil {
		return ""
	}

	switch e := expr.(type) {
	case *ir.VarExpr:
		return e.Name
	case *ir.LiteralExpr:
		// Для строк добавляем кавычки, но убираем существующие из Value
		if e.Kind == "STRING" {
			val := strings.Trim(e.Value, `"`)
			return fmt.Sprintf(`"%s"`, val)
		}
		return e.Value
	case *ir.BinaryExpr:
		left := g.generateExpression(e.Left)
		right := g.generateExpression(e.Right)
		if left == "" || right == "" {
			return ""
		}
		// Специальная обработка для println! макросов
		if e.Op == "," && isPrintlnMacro(left) {
			args := g.extractPrintlnArgs(e.Right)
			return g.generatePrintlnCall(args)
		}
		return fmt.Sprintf("(%s %s %s)", left, e.Op, right)
	case *ir.UnaryExpr:
		exprStr := g.generateExpression(e.Expr)
		if exprStr == "" {
			return ""
		}
		return fmt.Sprintf("%s%s", e.Op, exprStr)
	case *ir.CallExpr:
		// Обрабатываем макросы
		if e.IsMacro {
			if e.FuncName == "println!" {
				return g.generatePrintlnCall(e.Args)
			}
			if e.FuncName == "format!" {
				return g.generateFormatCall(e.Args)
			}
			// Для других макросов пока возвращаем TODO
			return fmt.Sprintf("// TODO: macro %s", e.FuncName)
		}

		args := []string{}
		for _, arg := range e.Args {
			argStr := g.generateExpression(arg)
			if argStr != "" {
				args = append(args, argStr)
			}
		}
		return fmt.Sprintf("%s(%s)", e.FuncName, strings.Join(args, ", "))
	}
	return ""
}

// generatePrintlnCall генерирует вызов fmt.Println.
func (g *Generator) generatePrintlnCall(args []ir.Expression) string {
	argStrs := []string{}
	for _, arg := range args {
		argStrs = append(argStrs, g.generateExpression(arg))
	}
	return fmt.Sprintf("fmt.Println(%s)", strings.Join(argStrs, ", "))
}

// generateFormatCall генерирует вызов fmt.Sprintf для format! макроса.
func (g *Generator) generateFormatCall(args []ir.Expression) string {
	if len(args) == 0 {
		return `""`
	}

	argStrs := []string{}
	for _, arg := range args {
		argStrs = append(argStrs, g.generateExpression(arg))
	}
	return fmt.Sprintf("fmt.Sprintf(%s)", strings.Join(argStrs, ", "))
}

// isPrintlnMacro проверяет, является ли выражение частью println! макроса.
func isPrintlnMacro(expr string) bool {
	return strings.Contains(expr, "println!") || strings.Contains(expr, "IDENT")
}

// extractPrintlnArgs извлекает аргументы для println! из бинарных операторов.
func (g *Generator) extractPrintlnArgs(expr ir.Expression) []ir.Expression {
	// Упрощённая реализация
	return []ir.Expression{expr}
}

// emit добавляет строку с учётом отступов.
func (g *Generator) emit(format string, args ...interface{}) {
	indent := strings.Repeat("\t", g.indent)
	line := fmt.Sprintf(format, args...)
	g.builder.WriteString(indent + line + "\n")
}

// emitln добавляет пустую строку.
func (g *Generator) emitln() {
	g.builder.WriteString("\n")
}

// capitalize делает первую букву заглавной (для Go).
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
