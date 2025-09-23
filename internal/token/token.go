// Пакет token: базовые типы токенов и позиционирования.
package token

// TokenType — перечисление типов токенов, которые лексер может выделять.
type TokenType int

const (
	// EOF — маркер конца входа.
	EOF TokenType = iota
	// IDENT — идентификатор (имена переменных, функций и т.д.).
	IDENT
	// LIFETIME — lifetime в Rust (например, 'a).
	LIFETIME
	// KEYWORD — зарезервированное слово языка (fn, let, break и т.д.).
	KEYWORD
	// TYPE - литерал типа
	TYPE
	// // INT — целочисленный литерал (включая 0b/0o/0x и суффиксы).
	// INT
	// // FLOAT — литерал с плавающей точкой.
	// FLOAT
	// // STRING — строковый литерал (обычный, raw, byte и их варианты).
	// STRING
	// // CHAR — символьный литерал (включая byte char).
	// CHAR
	// OPERATOR — операторы (==, &&, + и т.д.).
	OPERATOR
	// PUNCT — пунктуация (скобки, двоеточие, точки и т.п.).
	PUNCT
	// ATTRIBUTE — атрибуты Rust (например, #[derive(...)] или #![...]).
	ATTRIBUTE
	// TERMINATOR — отдельный токен для ';'
	TERMINATOR
	// ILLEGAL — неизвестный/некорректный токен.
	ILLEGAL
)

// Token — структура, представляющая один токен: тип, текстовое представление и позиция.
type Token struct {
	Type    TokenType // тип токена
	Subtype string
	Literal string    // текстовая форма токена, как встречается в исходнике
	Line    int       // номер строки (1-based)
	Col     int       // номер колонки (1-based)
}

// String helper — возвращает читаемое имя типа токена.
func (t Token) String() string {
	switch t.Type {
		case EOF: return "EOF"
		case IDENT: return "IDENT"
		case LIFETIME: return "LIFETIME"
		case KEYWORD: return "KEYWORD"
		case TYPE: if t.Subtype != "" { 
			return "TYPE(" + t.Subtype + ")" 
			}
		return "TYPE"
		case OPERATOR: return "OPERATOR"
		case PUNCT: return "PUNCT"
		case ATTRIBUTE: return "ATTRIBUTE"
		case ILLEGAL: return "ILLEGAL"
		case TERMINATOR: return "TERMINATOR"
		default: return "UNKNOWN"
	}
}