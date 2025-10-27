// Package ir определяет промежуточное представление (IR) для генерации кода.
// IR упрощает AST и добавляет информацию, необходимую для трансляции в целевой язык.
package ir

import (
	"github.com/semetekare/rust2go/internal/token"
)

// Module представляет IR-модуль, содержащий определения функций и типов.
type Module struct {
	Name        string      // Имя модуля
	Functions   []*Function // Функции модуля
	Structs     []*Struct   // Структуры модуля
	PackageName string      // Имя пакета Go
}

// Function представляет IR-функцию.
type Function struct {
	Name       string         // Имя функции
	Params     []*Parameter   // Параметры функции
	ReturnType *Type          // Возвращаемый тип
	Body       []Statement    // Тело функции (список операторов)
	Pos        token.Position // Позиция в исходном коде
	GoPackage  string         // Пакет Go для экспорта
	GoReceiver string         // Приёмник для методов (если есть)
}

// Parameter представляет параметр функции.
type Parameter struct {
	Name string // Имя параметра
	Type *Type  // Тип параметра
}

// Statement представляет оператор в IR.
type Statement interface {
	stmtNode()
	Pos() token.Position
}

// Declaration представляет объявление переменной.
type Declaration struct {
	Name      string
	Type      *Type
	InitValue Expression
	Position  token.Position
}

func (d *Declaration) stmtNode()           {}
func (d *Declaration) Pos() token.Position { return d.Position }

// Assignment представляет присваивание.
type Assignment struct {
	Target   string
	Value    Expression
	Position token.Position
}

func (a *Assignment) stmtNode()           {}
func (a *Assignment) Pos() token.Position { return a.Position }

// Return представляет возврат значения.
type Return struct {
	Value    Expression
	Position token.Position
}

func (r *Return) stmtNode()           {}
func (r *Return) Pos() token.Position { return r.Position }

// Expression представляет выражение в IR.
type Expression interface {
	exprNode()
	Type() *Type
	Pos() token.Position
}

// VarExpr представляет переменную.
type VarExpr struct {
	Name     string
	TypeInfo *Type
	Position token.Position
}

func (v *VarExpr) exprNode()           {}
func (v *VarExpr) Type() *Type         { return v.TypeInfo }
func (v *VarExpr) Pos() token.Position { return v.Position }

// LiteralExpr представляет литерал.
type LiteralExpr struct {
	Value    string
	Kind     string // "INT", "FLOAT", "STRING", "BOOL"
	TypeInfo *Type
	Position token.Position
}

func (l *LiteralExpr) exprNode()           {}
func (l *LiteralExpr) Type() *Type         { return l.TypeInfo }
func (l *LiteralExpr) Pos() token.Position { return l.Position }

// BinaryExpr представляет бинарное выражение.
type BinaryExpr struct {
	Left     Expression
	Op       string
	Right    Expression
	TypeInfo *Type
	Position token.Position
}

func (b *BinaryExpr) exprNode()           {}
func (b *BinaryExpr) Type() *Type         { return b.TypeInfo }
func (b *BinaryExpr) Pos() token.Position { return b.Position }

// UnaryExpr представляет унарное выражение.
type UnaryExpr struct {
	Op       string
	Expr     Expression
	TypeInfo *Type
	Position token.Position
}

func (u *UnaryExpr) exprNode()           {}
func (u *UnaryExpr) Type() *Type         { return u.TypeInfo }
func (u *UnaryExpr) Pos() token.Position { return u.Position }

// CallExpr представляет вызов функции.
type CallExpr struct {
	FuncName string
	Args     []Expression
	TypeInfo *Type
	Position token.Position
	IsMacro  bool // Является ли это макросом
}

func (c *CallExpr) exprNode()           {}
func (c *CallExpr) Type() *Type         { return c.TypeInfo }
func (c *CallExpr) Pos() token.Position { return c.Position }

// ExprStmt оборачивает выражение как оператор.
type ExprStmt struct {
	Expr     Expression
	Position token.Position
}

func (e *ExprStmt) stmtNode()           {}
func (e *ExprStmt) Pos() token.Position { return e.Position }

// Type представляет тип в IR.
type Type struct {
	Name        string
	IsPrimitive bool
	IsPointer   bool
	IsArray     bool
	ElementType *Type // Для массивов и указателей
}

// Struct представляет определение структуры в IR.
type Struct struct {
	Name   string
	Fields []*Field
	Pos    token.Position
}

// Field представляет поле структуры.
type Field struct {
	Name string
	Type *Type
}

// NewType создаёт новый тип.
func NewType(name string, isPrimitive bool) *Type {
	return &Type{Name: name, IsPrimitive: isPrimitive}
}

// NewArrayType создаёт тип массива.
func NewArrayType(elementType *Type) *Type {
	return &Type{
		Name:        "[]" + elementType.Name,
		IsArray:     true,
		ElementType: elementType,
	}
}

// NewPointerType создаёт тип указателя.
func NewPointerType(elementType *Type) *Type {
	return &Type{
		Name:        "*" + elementType.Name,
		IsPointer:   true,
		ElementType: elementType,
	}
}

// String возвращает строковое представление типа.
func (t *Type) String() string {
	if t.Name != "" {
		return t.Name
	}
	if t.IsArray {
		return "[]" + t.ElementType.String()
	}
	if t.IsPointer {
		return "*" + t.ElementType.String()
	}
	return "unknown"
}

// MapRustToGoType преобразует тип из Rust в Go.
func MapRustToGoType(rustType string) string {
	mapping := map[string]string{
		"i8":     "int8",
		"i16":    "int16",
		"i32":    "int",
		"i64":    "int64",
		"u8":     "uint8",
		"u16":    "uint16",
		"u32":    "uint32",
		"u64":    "uint64",
		"f32":    "float32",
		"f64":    "float64",
		"bool":   "bool",
		"str":    "string",
		"String": "string",
		"()":     "",
	}

	if goType, ok := mapping[rustType]; ok {
		return goType
	}
	// Для неизвестных типов просто возвращаем как есть (могут быть пользовательские типы)
	return rustType
}
