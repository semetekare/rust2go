// internal/ast/printer.go

// Package ast предоставляет функциональность для печати абстрактного синтаксического дерева (AST)
// в человекочитаемом, отформатированном виде.
package ast

import (
	"strings"
)

// PrettyPrint возвращает красиво отформатированное строковое представление узла AST.
// Результат включает отступы для вложенных узлов, что облегчает визуальный анализ структуры дерева.
// Используется в основном для отладки и логирования.
func PrettyPrint(n Node) string {
	var sb strings.Builder
	prettyPrintNode(&sb, n, 0)
	return sb.String()
}

// prettyPrintNode — рекурсивная вспомогательная функция для печати узла AST с заданным уровнем отступа.
// sb — буфер для накопления результата.
// n — узел AST, который нужно напечатать (может быть nil).
// indent — текущий уровень вложенности (каждый уровень соответствует двум пробелам).
//
// Функция сначала выводит строковое представление узла (через его метод String()),
// а затем рекурсивно обходит все его дочерние узлы (если они есть), увеличивая уровень отступа.
// Узлы, не имеющие потомков (например, Literal, PathType), не требуют дополнительной обработки.
func prettyPrintNode(sb *strings.Builder, n Node, indent int) {
	if n == nil {
		return
	}
	prefix := strings.Repeat("  ", indent)
	sb.WriteString(prefix)
	sb.WriteString(n.String())
	sb.WriteString("\n")

	// Рекурсивный обход дочерних узлов в зависимости от типа текущего узла.
	switch node := n.(type) {
	case *Crate:
		// Печатаем все элементы верхнего уровня (функции, структуры и т.д.).
		for _, item := range node.Items {
			prettyPrintNode(sb, item, indent+1)
		}
	case *Function:
		// Печатаем параметры функции и её тело.
		for _, param := range node.Params {
			prettyPrintNode(sb, &param, indent+1)
		}
		prettyPrintNode(sb, node.Body, indent+1)
	case *Struct:
		// Печатаем поля структуры.
		for _, field := range node.Fields {
			prettyPrintNode(sb, &field, indent+1)
		}
	case *Block:
		// Печатаем все операторы внутри блока.
		for _, stmt := range node.Stmts {
			prettyPrintNode(sb, stmt, indent+1)
		}
	case *LetStmt:
		// Печатаем тип переменной и выражение инициализации.
		prettyPrintNode(sb, node.Type, indent+1)
		prettyPrintNode(sb, node.Init, indent+1)
	case *ExprStmt:
		// Печатаем само выражение.
		prettyPrintNode(sb, node.Expr, indent+1)
	case *BinaryExpr:
		// Печатаем левый и правый операнды.
		prettyPrintNode(sb, node.Left, indent+1)
		prettyPrintNode(sb, node.Right, indent+1)
	case *UnaryExpr:
		// Печатаем операнд унарного выражения.
		prettyPrintNode(sb, node.Expr, indent+1)
	case *CallExpr:
		// Печатаем вызываемую функцию и аргументы.
		prettyPrintNode(sb, node.Func, indent+1)
		for _, arg := range node.Args {
			prettyPrintNode(sb, arg, indent+1)
		}
	case *BlockExpr:
		// Печатаем внутренний блок.
		prettyPrintNode(sb, node.Block, indent+1)
		// Листовые узлы (например, Literal, PathType, Param, Field) не имеют дочерних узлов,
		// поэтому для них отдельные case не требуются.
	}
}

// func prettyPrintItem(sb *strings.Builder, i Item, indent int) {
// 	prettyPrintNode(sb, i, indent)
// 	// Рекурсия для детей, если нужно
// }

// func prettyPrintStmt(sb *strings.Builder, s Stmt, indent int) {
// 	prettyPrintNode(sb, s, indent)
// }

// func prettyPrintExpr(sb *strings.Builder, e Expr, indent int) {
// 	prettyPrintNode(sb, e, indent)
// }
