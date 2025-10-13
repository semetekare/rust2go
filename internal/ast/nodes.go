// internal/ast/nodes.go

// Package ast определяет абстрактное синтаксическое дерево (AST) для представления
// синтаксической структуры Rust-подобного языка, транслируемого в Go.
package ast

import (
	"fmt"

	"github.com/semetekare/rust2go/internal/token"
)

// Position — псевдоним для token.Position, представляющий позицию в исходном коде.
type Position = token.Position

// Node — базовый интерфейс для всех узлов AST.
// Любой узел должен знать свою позицию в исходном коде и уметь преобразовываться в строку.
type Node interface {
	// Pos возвращает позицию узла в исходном коде.
	Pos() Position
	// String возвращает человекочитаемое строковое представление узла (в основном для отладки).
	String() string
}

// Crate представляет корень AST — единицу компиляции (crate).
// Соответствует грамматике: Crate ::= InnerAttribute* Item*
type Crate struct {
	pos   Position // Позиция начала crate в исходном коде.
	Items []Item   // Список элементов верхнего уровня (функций, структур и т.д.).
}

// Pos возвращает позицию начала crate.
func (c *Crate) Pos() Position { return c.pos }

// String возвращает строковое представление crate.
func (c *Crate) String() string { return fmt.Sprintf("Crate{Items: %d}", len(c.Items)) }

// NewCrate создаёт новый экземпляр Crate с заданной позицией и списком элементов.
func NewCrate(pos Position, items []Item) *Crate {
	return &Crate{pos: pos, Items: items}
}

// Item — интерфейс для элементов верхнего уровня (items) в crate.
// Примеры: функции, структуры, константы и т.д.
type Item interface {
	Node
	// itemString возвращает строковое представление элемента (для внутреннего использования).
	itemString() string
}

// Function представляет определение функции.
// Соответствует грамматике: Function ::= "fn" IDENTIFIER "(" Param* ")" [ "->" Type ] Block
type Function struct {
	pos        Position // Позиция ключевого слова "fn".
	Name       string   // Имя функции.
	Params     []Param  // Список параметров.
	ReturnType Type     // Возвращаемый тип (может быть nil для unit).
	Body       *Block   // Тело функции.
}

// Pos возвращает позицию начала функции.
func (f *Function) Pos() Position { return f.pos }

// String возвращает строковое представление функции.
func (f *Function) String() string { return fmt.Sprintf("Function{Name: %s}", f.Name) }

// itemString реализует интерфейс Item.
func (f *Function) itemString() string { return f.String() }

// NewFunction создаёт новый узел Function.
func NewFunction(pos Position, name string, params []Param, returnType Type, body *Block) *Function {
	return &Function{pos: pos, Name: name, Params: params, ReturnType: returnType, Body: body}
}

// Struct представляет определение структуры.
// Соответствует грамматике: Struct ::= "struct" IDENTIFIER "{" Field* "}"
type Struct struct {
	pos    Position // Позиция ключевого слова "struct".
	Name   string   // Имя структуры.
	Fields []Field  // Список полей структуры.
}

// Pos возвращает позицию начала структуры.
func (s *Struct) Pos() Position { return s.pos }

// String возвращает строковое представление структуры.
func (s *Struct) String() string { return fmt.Sprintf("Struct{Name: %s}", s.Name) }

// itemString реализует интерфейс Item.
func (s *Struct) itemString() string { return s.String() }

// NewStruct создаёт новый узел Struct.
func NewStruct(pos Position, name string, fields []Field) *Struct {
	return &Struct{pos: pos, Name: name, Fields: fields}
}

// Field представляет поле структуры.
// Соответствует грамматике: Field ::= IDENTIFIER ":" Type
type Field struct {
	pos  Position // Позиция имени поля.
	Name string   // Имя поля.
	Type Type     // Тип поля.
}

// Pos возвращает позицию начала поля.
func (f *Field) Pos() Position { return f.pos }

// String возвращает строковое представление поля.
func (f *Field) String() string { return fmt.Sprintf("Field{Name: %s}", f.Name) }

// NewField создаёт новый узел Field.
func NewField(pos Position, name string, typ Type) *Field {
	return &Field{pos: pos, Name: name, Type: typ}
}

// Stmt — интерфейс для всех видов операторов (statements).
type Stmt interface {
	Node
	// stmtString возвращает строковое представление оператора (для внутреннего использования).
	stmtString() string
}

// LetStmt представляет оператор объявления переменной.
// Соответствует грамматике: "let" IDENTIFIER [":" Type] "=" Expr ";"
// В текущей реализации шаблон (Pattern) упрощён до идентификатора.
type LetStmt struct {
	pos  Position // Позиция ключевого слова "let".
	Name string   // Имя переменной.
	Type Type     // Тип переменной (может быть nil для вывода типа).
	Init Expr     // Выражение инициализации.
}

// Pos возвращает позицию начала оператора let.
func (ls *LetStmt) Pos() Position { return ls.pos }

// String возвращает строковое представление оператора let.
func (ls *LetStmt) String() string { return fmt.Sprintf("LetStmt{Name: %s}", ls.Name) }

// stmtString реализует интерфейс Stmt.
func (ls *LetStmt) stmtString() string { return ls.String() }

// NewLetStmt создаёт новый узел LetStmt.
func NewLetStmt(pos Position, name string, typ Type, init Expr) *LetStmt {
	return &LetStmt{pos: pos, Name: name, Type: typ, Init: init}
}

// ExprStmt представляет выражение, используемое как оператор (например, вызов функции без присваивания).
type ExprStmt struct {
	pos  Position // Позиция выражения.
	Expr Expr     // Выражение.
}

// Pos возвращает позицию выражения-оператора.
func (es *ExprStmt) Pos() Position { return es.pos }

// String возвращает строковое представление выражения-оператора.
func (es *ExprStmt) String() string { return "ExprStmt" }

// stmtString реализует интерфейс Stmt.
func (es *ExprStmt) stmtString() string { return es.String() }

// NewExprStmt создаёт новый узел ExprStmt.
func NewExprStmt(pos Position, expr Expr) *ExprStmt {
	return &ExprStmt{pos: pos, Expr: expr}
}

// Block представляет блок кода, ограниченный фигурными скобками.
// Соответствует грамматике: Block ::= "{" Stmt* "}"
type Block struct {
	pos   Position // Позиция открывающей скобки "{".
	Stmts []Stmt   // Список операторов внутри блока.
}

// Pos возвращает позицию начала блока.
func (b *Block) Pos() Position { return b.pos }

// String возвращает строковое представление блока.
func (b *Block) String() string { return fmt.Sprintf("Block{Stmts: %d}", len(b.Stmts)) }

// stmtString реализует интерфейс Stmt (блок может использоваться как оператор).
func (b *Block) stmtString() string { return b.String() }

// exprString реализует интерфейс Expr (блок может использоваться как выражение).
func (b *Block) exprString() string { return b.String() }

// NewBlock создаёт новый узел Block.
func NewBlock(pos Position, stmts []Stmt) *Block {
	return &Block{pos: pos, Stmts: stmts}
}

// Expr — интерфейс для всех выражений.
type Expr interface {
	Node
	// exprString возвращает строковое представление выражения (для внутреннего использования).
	exprString() string
}

// UnaryExpr представляет унарное выражение (например, `-x`, `!flag`).
type UnaryExpr struct {
	pos  Position // Позиция оператора.
	Op   string   // Оператор (например, "-", "!", "*").
	Expr Expr     // Операнд.
}

// Pos возвращает позицию унарного оператора.
func (ue *UnaryExpr) Pos() Position { return ue.pos }

// String возвращает строковое представление унарного выражения.
func (ue *UnaryExpr) String() string { return fmt.Sprintf("UnaryExpr{%s}", ue.Op) }

// exprString реализует интерфейс Expr.
func (ue *UnaryExpr) exprString() string { return ue.String() }

// NewUnaryExpr создаёт новый узел UnaryExpr.
func NewUnaryExpr(pos Position, op string, expr Expr) *UnaryExpr {
	return &UnaryExpr{pos: pos, Op: op, Expr: expr}
}

// BinaryExpr представляет бинарное выражение (например, `a + b`, `x == y`).
type BinaryExpr struct {
	pos   Position // Позиция оператора.
	Left  Expr     // Левый операнд.
	Op    string   // Бинарный оператор ("+", "-", "==", "<", и т.д.).
	Right Expr     // Правый операнд.
}

// Pos возвращает позицию бинарного оператора.
func (be *BinaryExpr) Pos() Position { return be.pos }

// String возвращает строковое представление бинарного выражения.
func (be *BinaryExpr) String() string { return fmt.Sprintf("BinaryExpr{%s}", be.Op) }

// exprString реализует интерфейс Expr.
func (be *BinaryExpr) exprString() string { return be.String() }

// NewBinaryExpr создаёт новый узел BinaryExpr.
func NewBinaryExpr(pos Position, left Expr, op string, right Expr) *BinaryExpr {
	return &BinaryExpr{pos: pos, Left: left, Op: op, Right: right}
}

// Literal представляет литеральное значение (целое число, строка и т.д.).
type Literal struct {
	pos  Position // Позиция литерала в исходном коде.
	Kind string   // Тип литерала: "INT", "STRING", "BOOL" и т.д.
	Val  string   // Строковое представление значения.
}

// Pos возвращает позицию литерала.
func (l *Literal) Pos() Position { return l.pos }

// String возвращает строковое представление литерала.
func (l *Literal) String() string { return fmt.Sprintf("Literal{%s: %s}", l.Kind, l.Val) }

// exprString реализует интерфейс Expr.
func (l *Literal) exprString() string { return l.String() }

// NewLiteral создаёт новый узел Literal.
func NewLiteral(pos Position, kind string, val string) *Literal {
	return &Literal{pos: pos, Kind: kind, Val: val}
}

// CallExpr представляет вызов функции или метода.
// Соответствует грамматике: CallExpr ::= Expr "(" [Expr ("," Expr)*] ")"
type CallExpr struct {
	pos  Position // Позиция имени вызываемой функции.
	Func Expr     // Вызываемая функция (обычно идентификатор или путь).
	Args []Expr   // Аргументы вызова.
}

// Pos возвращает позицию вызова функции.
func (ce *CallExpr) Pos() Position { return ce.pos }

// String возвращает строковое представление вызова функции.
func (ce *CallExpr) String() string { return fmt.Sprintf("CallExpr{Args: %d}", len(ce.Args)) }

// exprString реализует интерфейс Expr.
func (ce *CallExpr) exprString() string { return ce.String() }

// NewCallExpr создаёт новый узел CallExpr.
func NewCallExpr(pos Position, fn Expr, args []Expr) *CallExpr {
	return &CallExpr{pos: pos, Func: fn, Args: args}
}

// Type — интерфейс для всех типов в языке.
type Type interface {
	Node
	// typeString возвращает строковое представление типа (для внутреннего использования).
	typeString() string
}

// PathType представляет тип, заданный именем (например, `i32`, `String`, `MyStruct`).
type PathType struct {
	pos  Position // Позиция имени типа.
	Path string   // Полное имя типа (в упрощённом виде — строка).
}

// Pos возвращает позицию типа.
func (pt *PathType) Pos() Position { return pt.pos }

// String возвращает строковое представление типа.
func (pt *PathType) String() string { return fmt.Sprintf("Type{%s}", pt.Path) }

// typeString реализует интерфейс Type.
func (pt *PathType) typeString() string { return pt.String() }

// NewPathType создаёт новый узел PathType.
func NewPathType(pos Position, path string) *PathType {
	return &PathType{pos: pos, Path: path}
}

// Param представляет параметр функции.
// Соответствует грамматике: Param ::= IDENTIFIER ":" Type
// В текущей реализации шаблон (Pattern) упрощён до идентификатора.
type Param struct {
	pos  Position // Позиция имени параметра.
	Name string   // Имя параметра.
	Type Type     // Тип параметра.
}

// Pos возвращает позицию параметра.
func (p *Param) Pos() Position { return p.pos }

// String возвращает строковое представление параметра.
func (p *Param) String() string { return fmt.Sprintf("Param{Name: %s}", p.Name) }

// NewParam создаёт новый узел Param.
func NewParam(pos Position, name string, typ Type) *Param {
	return &Param{pos: pos, Name: name, Type: typ}
}

// BlockExpr оборачивает Block, позволяя использовать его как выражение (например, в последнем выражении функции).
type BlockExpr struct {
	pos   Position // Позиция блока.
	Block *Block   // Обёрнутый блок.
}

// Pos возвращает позицию блочного выражения.
func (be *BlockExpr) Pos() Position { return be.pos }

// String возвращает строковое представление блочного выражения.
func (be *BlockExpr) String() string { return "BlockExpr" }

// exprString реализует интерфейс Expr.
func (be *BlockExpr) exprString() string { return be.String() }

// NewBlockExpr создаёт новый узел BlockExpr.
func NewBlockExpr(pos Position, block *Block) *BlockExpr {
	return &BlockExpr{pos: pos, Block: block}
}