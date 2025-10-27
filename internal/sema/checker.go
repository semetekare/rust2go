// Package sema реализует семантический анализ для Rust-подобного языка.
// Выполняет проверки типов, разрешение символов и другие семантические проверки.
package sema

import (
	"fmt"

	"github.com/semetekare/rust2go/internal/ast"
	"github.com/semetekare/rust2go/internal/token"
)

// Checker представляет семантический анализатор.
// Содержит таблицы символов, информацию о типах и накопленные ошибки.
type Checker struct {
	// Диагностические сообщения о семантических ошибках
	errors []SemanticError

	// Таблица символов: карта имён -> символы
	symbols map[string]*Symbol

	// Текущий контекст для отладки
	currentFunction string
}

// SemanticError представляет семантическую ошибку (например, неопределённая переменная, несовпадение типов).
type SemanticError struct {
	Msg string         // Описание ошибки
	Pos token.Position // Позиция в исходном коде
}

func (e SemanticError) Error() string {
	return fmt.Sprintf("Semantic error at %d:%d: %s", e.Pos.Line, e.Pos.Col, e.Msg)
}

// SymbolKind представляет категорию символа.
type SymbolKind int

const (
	SymbolVariable SymbolKind = iota
	SymbolFunction
	SymbolStruct
)

// Symbol представляет символ в таблице символов (переменная, функция, тип).
type Symbol struct {
	Kind     SymbolKind
	Name     string
	Type     TypeInfo
	Pos      token.Position
	Defined  bool
	Function *ast.Function // Для функций: указатель на определение
}

// TypeInfo представляет информацию о типе.
// В текущей реализации — упрощённая модель.
type TypeInfo struct {
	// Name — имя типа (например, "i32", "String", "()", "infer")
	Name string
	// IsArray — является ли тип массивом или срезом
	IsArray bool
	// IsReference — является ли тип ссылкой (&T)
	IsReference bool
}

// NewChecker создаёт новый семантический анализатор.
func NewChecker() *Checker {
	return &Checker{
		errors:  make([]SemanticError, 0),
		symbols: make(map[string]*Symbol),
	}
}

// Check выполняет семантический анализ над AST.
// Возвращает список обнаруженных семантических ошибок.
func (c *Checker) Check(crate *ast.Crate) []SemanticError {
	// Шаг 1: регистрируем все функции и структуры (декларации)
	c.checkCrateDeclarations(crate)

	// Шаг 2: проверяем тела функций (определения)
	c.checkCrateDefinitions(crate)

	return c.errors
}

// checkCrateDeclarations регистрирует все top-level декларации (функции, структуры).
func (c *Checker) checkCrateDeclarations(crate *ast.Crate) {
	for _, item := range crate.Items {
		switch it := item.(type) {
		case *ast.Function:
			c.registerFunction(it)
		case *ast.Struct:
			c.registerStruct(it)
		}
	}
}

// registerFunction регистрирует функцию в таблице символов.
func (c *Checker) registerFunction(fn *ast.Function) {
	// Проверяем, не объявлена ли функция уже
	if _, exists := c.symbols[fn.Name]; exists {
		c.error(fmt.Sprintf("duplicate function declaration: %s", fn.Name), fn.Pos())
		return
	}

	// Определяем тип возвращаемого значения
	retType := c.extractType(fn.ReturnType)

	// Создаём символ функции
	c.symbols[fn.Name] = &Symbol{
		Kind:     SymbolFunction,
		Name:     fn.Name,
		Type:     retType,
		Pos:      fn.Pos(),
		Defined:  true,
		Function: fn,
	}
}

// registerStruct регистрирует структуру в таблице символов.
func (c *Checker) registerStruct(st *ast.Struct) {
	if _, exists := c.symbols[st.Name]; exists {
		c.error(fmt.Sprintf("duplicate struct declaration: %s", st.Name), st.Pos())
		return
	}

	c.symbols[st.Name] = &Symbol{
		Kind:    SymbolStruct,
		Name:    st.Name,
		Type:    TypeInfo{Name: st.Name},
		Pos:     st.Pos(),
		Defined: true,
	}
}

// checkCrateDefinitions проверяет тела функций на корректность.
func (c *Checker) checkCrateDefinitions(crate *ast.Crate) {
	for _, item := range crate.Items {
		switch it := item.(type) {
		case *ast.Function:
			c.checkFunction(it)
		}
	}
}

// checkFunction выполняет семантическую проверку функции.
func (c *Checker) checkFunction(fn *ast.Function) {
	c.currentFunction = fn.Name

	// Создаём локальную область видимости для параметров
	localScope := make(map[string]*Symbol)

	// Регистрируем параметры как локальные переменные
	for _, param := range fn.Params {
		paramType := c.extractType(param.Type)
		// Преобразуем str в String для согласованности
		if paramType.Name == "str" {
			paramType.Name = "String"
		}
		localScope[param.Name] = &Symbol{
			Kind:    SymbolVariable,
			Name:    param.Name,
			Type:    paramType,
			Pos:     param.Pos(),
			Defined: true,
		}
	}

	// Проверяем тело функции с учётом локальной области
	c.checkBlock(fn.Body, localScope)

	c.currentFunction = ""
}

// checkBlock проверяет блок операторов.
func (c *Checker) checkBlock(block *ast.Block, scope map[string]*Symbol) {
	for _, stmt := range block.Stmts {
		c.checkStmt(stmt, scope)
	}
}

// checkStmt проверяет оператор.
func (c *Checker) checkStmt(stmt ast.Stmt, scope map[string]*Symbol) {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		c.checkLetStmt(s, scope)
	case *ast.ExprStmt:
		c.checkExpr(s.Expr, scope)
	}
}

// checkLetStmt проверяет оператор объявления переменной.
func (c *Checker) checkLetStmt(ls *ast.LetStmt, scope map[string]*Symbol) {
	// Проверяем, не объявлена ли переменная уже
	if _, exists := scope[ls.Name]; exists {
		c.error(fmt.Sprintf("variable %s already declared in this scope", ls.Name), ls.Pos())
		return
	}

	// Тип инициализирующего выражения
	initType := c.checkExpr(ls.Init, scope)

	// Если тип объявлен явно
	if ls.Type != nil {
		declType := c.extractType(ls.Type)

		// Если явный тип — "infer", значит тип должен выводиться из инициализатора
		if declType.Name == "infer" {
			scope[ls.Name] = &Symbol{
				Kind:    SymbolVariable,
				Name:    ls.Name,
				Type:    initType,
				Pos:     ls.Pos(),
				Defined: true,
			}
			return
		}

		// Проверяем совпадение типов
		if !c.typesCompatible(declType, initType) {
			c.error(fmt.Sprintf("type mismatch: expected %s, got %s", declType.Name, initType.Name), ls.Pos())
		}

		// Регистрируем переменную в текущей области
		scope[ls.Name] = &Symbol{
			Kind:    SymbolVariable,
			Name:    ls.Name,
			Type:    declType,
			Pos:     ls.Pos(),
			Defined: true,
		}
	} else {
		// Тип выводится из инициализатора
		if initType.Name == "infer" {
			c.error("cannot infer type for variable without explicit type", ls.Pos())
			return
		}

		scope[ls.Name] = &Symbol{
			Kind:    SymbolVariable,
			Name:    ls.Name,
			Type:    initType,
			Pos:     ls.Pos(),
			Defined: true,
		}
	}
}

// checkExpr проверяет выражение и возвращает его тип.
func (c *Checker) checkExpr(expr ast.Expr, scope map[string]*Symbol) TypeInfo {
	switch e := expr.(type) {
	case *ast.Literal:
		return c.checkLiteral(e, scope)
	case *ast.BinaryExpr:
		return c.checkBinaryExpr(e, scope)
	case *ast.UnaryExpr:
		return c.checkUnaryExpr(e, scope)
	case *ast.CallExpr:
		return c.checkCallExpr(e, scope)
	case *ast.BlockExpr:
		return c.checkBlockExpr(e, scope)
	default:
		c.error("unsupported expression type", expr.Pos())
		return TypeInfo{Name: "()"}
	}
}

// checkLiteral проверяет литеральное значение.
func (c *Checker) checkLiteral(lit *ast.Literal, scope map[string]*Symbol) TypeInfo {
	switch lit.Kind {
	case "INT":
		return TypeInfo{Name: "i32"}
	case "FLOAT":
		return TypeInfo{Name: "f64"}
	case "STRING":
		return TypeInfo{Name: "String"}
	case "BOOL":
		return TypeInfo{Name: "bool"}
	case "IDENT":
		// Идентификатор — нужно разрешить в таблице символов
		return c.resolveIdentifier(lit, scope)
	default:
		return TypeInfo{Name: "()"}
	}
}

// resolveIdentifier разрешает идентификатор (переменную или функцию).
// Использует как глобальную таблицу символов, так и локальную область видимости.
func (c *Checker) resolveIdentifier(lit *ast.Literal, scope map[string]*Symbol) TypeInfo {
	name := lit.Val

	// Проверяем, является ли это макросом (по Subtype)
	// В лексере макросы помечаются как IDENT с Subtype = "MACRO"
	if len(name) > 0 && name[len(name)-1] == '!' {
		// Это встроенный макрос (println!, vec! и т.д.)
		return TypeInfo{Name: "()"}
	}

	// Сначала проверяем локальную область видимости (параметры, локальные переменные)
	if scope != nil {
		if sym, exists := scope[name]; exists {
			return sym.Type
		}
	}

	// Затем проверяем глобальную таблицу символов (функции, структуры)
	sym := c.symbols[name]
	if sym != nil {
		return sym.Type
	}

	c.error(fmt.Sprintf("undefined identifier: %s", name), lit.Pos())
	return TypeInfo{Name: "()"}
}

// checkBinaryExpr проверяет бинарное выражение.
func (c *Checker) checkBinaryExpr(be *ast.BinaryExpr, scope map[string]*Symbol) TypeInfo {
	leftType := c.checkExpr(be.Left, scope)
	rightType := c.checkExpr(be.Right, scope)

	// Проверка арифметических операций
	if c.isArithmeticOp(be.Op) {
		if !c.isNumeric(leftType) || !c.isNumeric(rightType) {
			c.error(fmt.Sprintf("operands of %s must be numeric", be.Op), be.Pos())
			return TypeInfo{Name: "()"}
		}
		return leftType // Результат арифметической операции имеет тот же тип
	}

	// Проверка операций сравнения
	if c.isComparisonOp(be.Op) {
		if !c.typesCompatible(leftType, rightType) {
			c.error(fmt.Sprintf("cannot compare %s with %s", leftType.Name, rightType.Name), be.Pos())
		}
		return TypeInfo{Name: "bool"}
	}

	// Проверка логических операций
	if c.isLogicalOp(be.Op) {
		if !c.isBool(leftType) || !c.isBool(rightType) {
			c.error(fmt.Sprintf("operands of %s must be boolean", be.Op), be.Pos())
		}
		return TypeInfo{Name: "bool"}
	}

	return TypeInfo{Name: "()"}
}

// checkUnaryExpr проверяет унарное выражение.
func (c *Checker) checkUnaryExpr(ue *ast.UnaryExpr, scope map[string]*Symbol) TypeInfo {
	exprType := c.checkExpr(ue.Expr, scope)

	switch ue.Op {
	case "-":
		if !c.isNumeric(exprType) {
			c.error("operand of unary - must be numeric", ue.Pos())
		}
		return exprType
	case "!":
		if !c.isBool(exprType) {
			c.error("operand of unary ! must be boolean", ue.Pos())
		}
		return TypeInfo{Name: "bool"}
	default:
		return TypeInfo{Name: "()"}
	}
}

// checkCallExpr проверяет вызов функции.
func (c *Checker) checkCallExpr(ce *ast.CallExpr, scope map[string]*Symbol) TypeInfo {
	// Получаем функцию из литерала идентификатора
	var fnName string
	switch f := ce.Func.(type) {
	case *ast.Literal:
		if f.Kind == "IDENT" {
			fnName = f.Val
		}
	default:
		c.error("expected function name in call", ce.Pos())
		return TypeInfo{Name: "()"}
	}

	// Проверяем на встроенные макросы (заканчиваются на !)
	if len(fnName) > 0 && fnName[len(fnName)-1] == '!' {
		// Встроенные макросы принимают произвольные аргументы и возвращают ()
		for _, arg := range ce.Args {
			c.checkExpr(arg, scope)
		}
		return TypeInfo{Name: "()"}
	}

	// Ищем функцию в таблице символов
	sym, exists := c.symbols[fnName]
	if !exists {
		c.error(fmt.Sprintf("undefined function: %s", fnName), ce.Pos())
		return TypeInfo{Name: "()"}
	}

	if sym.Kind != SymbolFunction || sym.Function == nil {
		c.error(fmt.Sprintf("%s is not a function", fnName), ce.Pos())
		return TypeInfo{Name: "()"}
	}

	fn := sym.Function

	// Проверяем количество аргументов
	if len(ce.Args) != len(fn.Params) {
		c.error(fmt.Sprintf("function %s expects %d arguments, got %d", fnName, len(fn.Params), len(ce.Args)), ce.Pos())
		return TypeInfo{Name: "()"}
	}

	// Проверяем типы аргументов
	for i, arg := range ce.Args {
		argType := c.checkExpr(arg, scope)
		paramType := c.extractType(fn.Params[i].Type)

		if !c.typesCompatible(paramType, argType) {
			c.error(fmt.Sprintf("argument %d of %s: expected %s, got %s", i+1, fnName, paramType.Name, argType.Name), ce.Pos())
		}
	}

	// Возвращаем тип возвращаемого значения функции
	return c.extractType(fn.ReturnType)
}

// checkBlockExpr проверяет блочное выражение.
func (c *Checker) checkBlockExpr(be *ast.BlockExpr, scope map[string]*Symbol) TypeInfo {
	// Для простоты возвращаем unit тип
	// В полной реализации нужно анализировать последнее выражение блока
	return TypeInfo{Name: "()"}
}

// extractType извлекает информацию о типе из AST типа.
func (c *Checker) extractType(t ast.Type) TypeInfo {
	if t == nil {
		return TypeInfo{Name: "()"}
	}

	switch typ := t.(type) {
	case *ast.PathType:
		return TypeInfo{Name: typ.Path}
	default:
		return TypeInfo{Name: "()"}
	}
}

// typesCompatible проверяет совместимость типов.
func (c *Checker) typesCompatible(t1, t2 TypeInfo) bool {
	// Тип "infer" совместим с любым типом (вывод типа)
	if t1.Name == "infer" || t2.Name == "infer" {
		return true
	}

	// str и &str совместимы с String
	if (t1.Name == "str" && t2.Name == "String") || (t1.Name == "String" && t2.Name == "str") {
		return true
	}

	// В упрощённой реализации считаем, что типы совместимы только если они идентичны
	return t1.Name == t2.Name
}

// isNumeric проверяет, является ли тип числовым.
func (c *Checker) isNumeric(t TypeInfo) bool {
	return t.Name == "i32" || t.Name == "i64" || t.Name == "f32" || t.Name == "f64" || t.Name == "i8" || t.Name == "i16" || t.Name == "u8" || t.Name == "u16" || t.Name == "u32" || t.Name == "u64"
}

// isBool проверяет, является ли тип булевым.
func (c *Checker) isBool(t TypeInfo) bool {
	return t.Name == "bool"
}

// isArithmeticOp проверяет, является ли оператор арифметическим.
func (c *Checker) isArithmeticOp(op string) bool {
	ops := map[string]bool{"+": true, "-": true, "*": true, "/": true, "%": true}
	return ops[op]
}

// isComparisonOp проверяет, является ли оператор оператором сравнения.
func (c *Checker) isComparisonOp(op string) bool {
	ops := map[string]bool{"==": true, "!=": true, "<": true, ">": true, "<=": true, ">=": true}
	return ops[op]
}

// isLogicalOp проверяет, является ли оператор логическим.
func (c *Checker) isLogicalOp(op string) bool {
	ops := map[string]bool{"&&": true, "||": true}
	return ops[op]
}

// error добавляет новую семантическую ошибку.
func (c *Checker) error(msg string, pos token.Position) {
	c.errors = append(c.errors, SemanticError{Msg: msg, Pos: pos})
}
