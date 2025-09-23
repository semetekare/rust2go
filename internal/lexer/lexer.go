// Пакет lexer: основная логика лексирования, реализует Lex(input) ([]token.Token, error).
package lexer

import (
	"fmt"
	"unicode"

	"github.com/semetekare/rust2go/internal/token"
)

// lexer — приватная структура, содержащая состояние сканирования.
// Внутренне хранит input как []rune для корректной работы с Unicode.
type Lexer struct {
	input        string            // исходный текст (как строка)
	runes        []rune            // исходный текст как срез рун (Unicode-aware)
	length       int               // длина s runes
	pos          int               // текущий индекс рун
	readPos      int               // индекс следующей руны
	ch           rune              // текущая просматриваемая руна
	line         int               // текущая строка (1-based)
	col          int               // текущая колонка (1-based)
	tokens       []token.Token           // накопленные токены
	err          error             // первая возникшая ошибка
	keywords     map[string]bool   // таблица ключевых слов
	operators    map[string]bool   // таблица операторов (включая многосимвольные)
	punctuations map[string]bool   // таблица пунктуации (включая многосимвольные)
}

// NewLexer создаёт и инициализирует лексер.
func NewLexer() *Lexer {
	return &Lexer{
		line:         1,
		col:          0,
		keywords:     Keywords,
		operators:    Operators,
		punctuations: Punctuations,
	}
}

// Lex запускает разбор входной строки и возвращает слайс токенов.
// Основная точка входа для использования лексера.
func (l *Lexer) Lex(input string) ([]token.Token, error) {
	l.input = input
	l.runes = []rune(input) // переводим в runes, чтобы корректно работать с UTF-8
	l.length = len(l.runes)
	l.pos = 0
	l.readPos = 0
	l.tokens = nil
	l.err = nil
	l.ch = 0
	l.readChar()

	for l.ch != 0 {
		l.nextToken()
		if l.err != nil {
			return nil, l.err
		}
	}

	// Добавляем EOF токен в конец
	l.tokens = append(l.tokens, token.Token{Type: token.EOF, Line: l.line, Col: l.col})
	return l.tokens, nil
}

// readChar читает следующую руну в поток и обновляет позицию, строку и колонку.
// Реализация работает с индексами рун, чтобы не ломать многобайтовые символы.
func (l *Lexer) readChar() {
	if l.readPos >= l.length {
		l.ch = 0
	} else {
		l.ch = l.runes[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
	if l.ch == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}
}

// peek возвращает следующую руну без продвижения позиции.
// Используется для принятия решений о многосимвольных операторах и префиксах.
func (l *Lexer) peek() rune {
	if l.readPos >= l.length {
		return 0
	}
	return l.runes[l.readPos]
}

// peekN возвращает n-ую руну вперед (n >= 1), безопасно при выходе за пределы.
func (l *Lexer) peekN(n int) rune {
	idx := l.readPos + n - 1
	if idx >= l.length || idx < 0 {
		return 0
	}
	return l.runes[idx]
}

// skipWhitespace пропускает все пробельные символы (включая новые строки).
func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		l.readChar()
	}
}

// skipComment пропускает однострочные (//) и блочные (/* ... */) комментарии.
// Блочные комментарии поддерживают вложенность.
func (l *Lexer) skipComment() {
	if l.ch == '/' && l.peek() == '/' {
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
	} else if l.ch == '/' && l.peek() == '*' {
		l.readChar(); l.readChar()
		nest := 1
		for l.ch != 0 && nest > 0 {
			if l.ch == '/' && l.peek() == '*' {
				l.readChar(); l.readChar(); nest++
			} else if l.ch == '*' && l.peek() == '/' {
				l.readChar(); l.readChar(); nest--
			} else {
				l.readChar()
			}
		}
	}
}

// isDigitInBase проверяет, является ли руна допустимой цифрой для заданного основания.
// Учитывает буквы a-f/A-F для base==16.
func isDigitInBase(ch rune, base int) bool {
	if unicode.IsDigit(ch) {
		d := int(ch - '0')
		return d < base
	}
	if base == 16 {
		if ch >= 'a' && ch <= 'f' { return true }
		if ch >= 'A' && ch <= 'F' { return true }
	}
	return false
}

// readIdentifier читает последовательность символов, образующих идентификатор.
func (l *Lexer) readIdentifier() string {
	start := l.pos
	for unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return string(l.runes[start:l.pos])
}

// readLifetimeOrChar различает lifetime ('a) и char ('a').
// Логика: если после имени идёт закрывающий апостроф — это символьный литерал.
func (l *Lexer) readLifetimeOrChar() (string, token.TokenType, string) {
	// at '\''
	// if pattern is '\'x\'' -> char (single rune possibly escaped)
	// else it's lifetime: '\'name'
	start := l.pos
	l.readChar() // skip '
	// собираем буквы/цифры/подчёркивания (имя lifetime)
	for unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	// если следующий символ — апостроф, то это формат 'x' -> CHAR
	if l.ch == '\'' {
		l.readChar()
		return string(l.runes[start:l.pos]), token.TYPE, "CHAR"
	}
	// иначе — lifetime (без завершающего апострофа)
	return string(l.runes[start:l.pos]), token.LIFETIME, ""
}

// readNumber читает целые и дробные литералы, учитывает префиксы 0b/0o/0x,
// экспоненты, подчёркивания для разделения разрядов и суффиксы типов (u32, f64 и т.д.).
func (l *Lexer) readNumber() (string, string) {
	// возвращаем (literal, subtype) где subtype = "INT" или "FLOAT"
	start := l.pos
	base := 10

	if l.ch == '0' {
		if l.peek() == 'b' || l.peek() == 'o' || l.peek() == 'x' {
			l.readChar()
			switch l.ch {
			case 'b': base=2; l.readChar()
			case 'o': base=8; l.readChar()
			case 'x': base=16; l.readChar()
			default: base=10
			}
		}
	}

	for isDigitInBase(l.ch, base) || l.ch == '_' {
		l.readChar()
	}

	isFloat := false
	if l.ch == '.' && base == 10 && isDigitInBase(l.peek(), 10) {
		isFloat = true
		l.readChar()
		for unicode.IsDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
	}

	if (l.ch == 'e' || l.ch == 'E') && base == 10 {
		isFloat = true
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for unicode.IsDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
	}

	// суффикс
	for unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) {
		l.readChar()
	}

	lit := string(l.runes[start:l.pos])
	if isFloat {
		return lit, "FLOAT"
	}
	return lit, "INT"
}

func (l *Lexer) readString(prefix string) (string, string) {
	// возвращаем (literal, subtype) где subtype == "STRING" (или "CHAR" для byte char handled separately)
	start := l.pos - len([]rune(prefix))
	hashCount := 0

	if prefix == "r" || prefix == "br" {
		for l.ch == '#' {
			hashCount++
			l.readChar()
		}
		if l.ch != '"' {
			l.err = fmt.Errorf("invalid raw string literal at line %d, col %d", l.line, l.col)
			return "", ""
		}
	}

	l.readChar() // Skip opening "

	if prefix == "r" || prefix == "br" {
		for l.ch != 0 {
			if l.ch == '"' {
				l.readChar()
				matched := 0
				for l.ch == '#' && matched < hashCount {
					matched++
					l.readChar()
				}
				if matched == hashCount {
					break
				}
			} else {
				l.readChar()
			}
		}
	} else {
		for l.ch != '"' && l.ch != 0 {
			if l.ch == '\\' {
				l.readChar() // Escape char
				if l.ch == '\n' || l.ch == '\r' {
					if l.ch == '\r' && l.peek() == '\n' {
						l.readChar()
					}
					l.readChar()
					continue
				}
			}
			l.readChar()
		}
		if l.ch == '"' {
			l.readChar()
		} else {
			l.err = fmt.Errorf("unterminated string literal at line %d, col %d", l.line, l.col)
		}
	}

	return string(l.runes[start:l.pos]), "STRING"
}

// readAttr читает атрибуты Rust: #[...] и #![...]
// Поддерживает вложенные квадратные скобки внутри атрибута.
func (l *Lexer) readAttr() string {
	start := l.pos
	l.readChar() // #
	if l.ch == '!' {
		l.readChar() // Consume #!
	}
	if l.ch != '[' {
		l.err = fmt.Errorf("invalid attribute syntax: expected '[' at line %d, col %d", l.line, l.col)
		return ""
	}
	l.readChar() // [
	depth := 1
	for l.ch != 0 && depth > 0 {
		if l.ch == '[' {
			depth++
		} else if l.ch == ']' {
			depth--
		}
		l.readChar()
	}
	if depth > 0 {
		l.err = fmt.Errorf("unterminated attribute at line %d, col %d", l.line, l.col)
	}
	return string(l.runes[start:l.pos])
}

// readOpOrPunct читает операторы и пунктуацию, пытаясь сначала матчить
// трёхсимвольные, затем двухсимвольные, затем односивольные последовательности.
func (l *Lexer) readOpOrPunct() string {
	start := l.pos
	possibleThree := string(l.ch) + string(l.peek()) + string(l.peekN(2))
	possibleTwo := string(l.ch) + string(l.peek())
	if l.operators[possibleThree] || l.punctuations[possibleThree] {
		l.readChar()
		l.readChar()
		l.readChar()
		return string(l.runes[start:l.pos])
	} else if l.operators[possibleTwo] || l.punctuations[possibleTwo] {
		l.readChar()
		l.readChar()
		return string(l.runes[start:l.pos])
	}
	l.readChar()
	return string(l.runes[start:l.pos])
}

// Вспомогательные предикаты для распознавания операторных и пунктуационных символов.
func isOperatorChar(ch rune) bool {
	return ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '%' ||
		ch == '=' || ch == '!' || ch == '<' || ch == '>' || ch == '&' || ch == '|' ||
		ch == '^' || ch == '~' || ch == '?'
}

func isPunctChar(ch rune) bool {
	return ch == '{' || ch == '}' || ch == '(' || ch == ')' || ch == '[' || ch == ']' ||
		ch == ';' || ch == ',' || ch == ':' || ch == '.' || ch == '#' || ch == '@' || ch == '!'
}

// containsDotOrExp проверяет, содержит ли строка точки или показатель экспоненты,
// используется при классификации числа как FLOAT.
func containsDotOrExp(s string) bool {
	for _, c := range s {
		if c == '.' || c == 'e' || c == 'E' {
			return true
		}
	}
	return false
}

// nextToken — центральная функция, которая анализирует текущую руну и формирует токен.
// Ведёт себя итеративно: пропускает пробелы/комментарии, затем вызывает соответствующие читатели.
func (l *Lexer) nextToken() {
	l.skipWhitespace()

	if l.ch == '/' && (l.peek() == '/' || l.peek() == '*') {
		l.skipComment()
		return
	}

	var tok token.Token
	tok.Line = l.line
	tok.Col = l.col

	switch {
	case l.ch == 0:
		return
	case l.ch == '\'' && (unicode.IsLetter(l.peek()) || l.peek() == '_'):
		// need to distinguish lifetime vs char: check next-next char for closing '
		// use helper that returns subtype for CHAR
		lit, ttype, subtype := l.readLifetimeOrChar()
		tok.Literal = lit
		if ttype == token.TYPE {
			tok.Type = token.TYPE
			tok.Subtype = subtype // "CHAR"
		} else {
			tok.Type = token.LIFETIME
		}
	case unicode.IsLetter(l.ch) || l.ch == '_':
		prefix := l.readIdentifier()
		switch {
		case prefix == "r" && (l.ch == '"' || l.ch == '#'):
			lit, subtype := l.readString("r")
			tok.Literal = lit
			tok.Type = token.TYPE
			tok.Subtype = subtype // "STRING"
		case prefix == "br" && (l.ch == '"' || l.ch == '#'):
			lit, subtype := l.readString("br")
			tok.Literal = lit
			tok.Type = token.TYPE
			tok.Subtype = subtype
		case prefix == "b" && l.ch == '"':
			lit, subtype := l.readString("b")
			tok.Literal = lit
			tok.Type = token.TYPE
			tok.Subtype = subtype
		case prefix == "b" && l.ch == '\'':
			// byte char literal
			lit, _ := l.readString("b")
			tok.Literal = lit
			tok.Type = token.TYPE
			tok.Subtype = "CHAR"
		default:
			tok.Literal = prefix
			if l.keywords[tok.Literal] {
				tok.Type = token.KEYWORD
			} else {
				tok.Type = token.IDENT
			}
		}
	case unicode.IsDigit(l.ch):
		lit, subtype := l.readNumber()
		tok.Literal = lit
		tok.Type = token.TYPE
		tok.Subtype = subtype // "INT" or "FLOAT"
	case l.ch == '"':
		lit, subtype := l.readString("")
		tok.Literal = lit
		tok.Type = token.TYPE
		tok.Subtype = subtype // "STRING"
	case l.ch == '\'':
		lit, ttype, subtype := l.readLifetimeOrChar()
		tok.Literal = lit
		if ttype == token.TYPE {
			tok.Type = token.TYPE
			tok.Subtype = subtype
		} else {
			tok.Type = token.LIFETIME
		}
	case l.ch == '#':
		tok.Literal = l.readAttr()
		tok.Type = token.ATTRIBUTE
	default:
		// операторы и пунктуация
		lit := l.readOpOrPunct()
		// отдельный случай: если это точка с запятой — TERMINATOR
		if lit == ";" {
			tok.Type = token.TERMINATOR
			tok.Literal = lit
		} else {
			tok.Literal = lit
			if l.operators[tok.Literal] {
				tok.Type = token.OPERATOR
			} else if l.punctuations[tok.Literal] {
				tok.Type = token.PUNCT
			} else {
				tok.Type = token.ILLEGAL
			}
		}
	}

	if l.err == nil {
		l.tokens = append(l.tokens, tok)
	}
}
