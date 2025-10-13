// Пакет token определяет базовые типы для представления лексем (токенов),
// выделяемых лексическим анализатором (лексером), а также их позиций в исходном коде.
package token

// TokenType — перечисление возможных типов токенов, которые может распознать лексер.
// Каждый тип соответствует определённой категории лексем в языке.
type TokenType int

// String возвращает строковое представление типа токена.
// Метод объявлен, но не реализован — должен быть заменён или удалён в рабочей версии.
func (t TokenType) String() string {
	panic("unimplemented")
}

const (
	// EOF — маркер конца входного потока (end-of-file).
	// Указывает, что лексер достиг конца исходного кода.
	EOF TokenType = iota

	// IDENT — идентификатор: имя переменной, функции, типа и т.д.
	// Примеры: x, my_var, Foo.
	IDENT

	// LIFETIME — lifetime-параметр из Rust (например, 'a, 'static).
	// Используется для управления временем жизни значений.
	LIFETIME

	// KEYWORD — зарезервированное ключевое слово языка.
	// Примеры: fn, let, if, while, struct, impl и т.д.
	KEYWORD

	// TYPE — литерал типа или имя типа.
	// Примеры: i32, String, Vec<T>. Подтип уточняется в поле Subtype.
	TYPE

	// INT — целочисленный литерал.
	// Поддерживает десятичную, двоичную (0b...), восьмеричную (0o...) и шестнадцатеричную (0x...) формы,
	// а также суффиксы типов (например, 42u32).
	INT

	// FLOAT — литерал с плавающей точкой.
	// Примеры: 3.14, 1e-5, 2.0f32.
	FLOAT

	// STRING — строковый литерал.
	// Включает обычные строки ("..."), raw-строки (r#"..."#), байтовые строки (b"...") и их комбинации.
	STRING

	// CHAR — символьный литерал.
	// Примеры: 'a', '\n', b'x' (байтовый символ).
	CHAR

	// OPERATOR — операторы языка.
	// Примеры: +, -, ==, !=, &&, ||, =, += и т.д.
	OPERATOR

	// PUNCT — пунктуационные символы (разделители).
	// Примеры: (, ), {, }, [, ], ,, ., :, :: и т.п.
	PUNCT

	// ATTRIBUTE — атрибуты Rust.
	// Примеры: #[derive(Debug)], #![no_std], #[cfg(...)].
	ATTRIBUTE

	// TERMINATOR — отдельный токен для точки с запятой ';',
	// используемой как завершитель операторов.
	TERMINATOR

	// ILLEGAL — недопустимый или не распознанный токен.
	// Используется для обозначения синтаксических ошибок на этапе лексического анализа.
	ILLEGAL
)

// Position представляет позицию символа в исходном коде.
// Нумерация строк и колонок начинается с 1 (1-based).
type Position struct {
	Line int // Номер строки (начиная с 1).
	Col  int // Номер колонки (начиная с 1).
}

// Token представляет один лексический токен, полученный в результате анализа исходного кода.
type Token struct {
	Type    TokenType // Основной тип токена (см. константы выше).
	Subtype string    // Дополнительная информация о типе (например, "INT", "FLOAT" для TYPE).
	Literal string    // Исходный текст токена, как он встречается в коде.
	Line    int       // Номер строки, в которой находится токен (1-based).
	Col     int       // Номер колонки начала токена (1-based).
}

// Pos возвращает позицию токена в виде структуры Position.
func (t Token) Pos() Position {
	return Position{Line: t.Line, Col: t.Col}
}

// String возвращает человекочитаемое строковое представление токена,
// включая его тип и, при необходимости, подтип.
// Используется в основном для отладки и диагностических сообщений.
func (t Token) String() string {
	switch t.Type {
	case EOF:
		return "EOF"
	case IDENT:
		return "IDENT"
	case LIFETIME:
		return "LIFETIME"
	case KEYWORD:
		return "KEYWORD"
	case TYPE:
		if t.Subtype != "" {
			return "TYPE(" + t.Subtype + ")"
		}
		return "TYPE"
	case INT:
		return "INT"
	case FLOAT:
		return "FLOAT"
	case STRING:
		return "STRING"
	case CHAR:
		return "CHAR"
	case OPERATOR:
		return "OPERATOR"
	case PUNCT:
		return "PUNCT"
	case ATTRIBUTE:
		return "ATTRIBUTE"
	case TERMINATOR:
		return "TERMINATOR"
	case ILLEGAL:
		return "ILLEGAL"
	default:
		return "UNKNOWN"
	}
}