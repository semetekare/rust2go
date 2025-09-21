// Пакет lexer: основная логика лексирования, реализует Lex(input) ([]token.Token, error).
package lexer

import (
	"fmt"
	"unicode"
	
	"github.com/semetekare/rust2go/internal/token"
)

// lexer — приватная структура, содержащая состояние сканирования.
// Внутренне хранит input как []rune для корректной работы с Unicode.
type lexer struct {
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

// LexerUseCase — интерфейс лексера. Отделяет реализацию от места вызова.
type LexerUseCase interface {
	// Lex принимает входную строку и возвращает слайс токенов или ошибку.
	Lex(input string) ([]token.Token, error)
}

// New создаёт и инициализирует новый лексер.
func New() *lexer {
	return &lexer{
		line: 1,
		col:  0,
		keywords: Keywords,
		operators: Operators,
		punctuations: Punctuations,
	}
}

// Lex запускает разбор входной строки и возвращает слайс токенов.
// Основная точка входа для использования лексера.
func (l *lexer) Lex(input string) ([]token.Token, error) {
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
func (l *lexer) readChar() {
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
func (l *lexer) peek() rune {
	if l.readPos >= l.length {
		return 0
	}
	return l.runes[l.readPos]
}

// peekN возвращает n-ую руну вперед (n >= 1), безопасно при выходе за пределы.
func (l *lexer) peekN(n int) rune {
	idx := l.readPos + n - 1
	if idx >= l.length || idx < 0 {
		return 0
	}
	return l.runes[idx]
}

// skipWhitespace пропускает все пробельные символы (включая новые строки).
func (l *lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		l.readChar()
	}
}

// skipComment пропускает однострочные (//) и блочные (/* ... */) комментарии.
// Блочные комментарии поддерживают вложенность.
func (l *lexer) skipComment() {
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
func (l *lexer) readIdentifier() string {
	start := l.pos
	for unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return string(l.runes[start:l.pos])
}

// readLifetimeOrChar различает lifetime ('a) и char ('a').
// Логика: если после имени идёт закрывающий апостроф — это символьный литерал.
func (l *lexer) readLifetimeOrChar() (string, token.TokenType) {
	// at '\''
	// if pattern is '\'x\'' -> char (single rune possibly escaped)
	// else it's lifetime: '\'name'
	start := l.pos
	l.readChar() // пропускаем открывающий '\''
	// собираем буквы/цифры/подчёркивания (имя lifetime)
	for unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	// если следующий символ — апостроф, то это формат 'x' -> CHAR
	if l.ch == '\'' {
		l.readChar()
		return string(l.runes[start:l.pos]), token.CHAR
	}
	// иначе — lifetime (без завершающего апострофа)
	return string(l.runes[start:l.pos]), token.LIFETIME
}

// readNumber читает целые и дробные литералы, учитывает префиксы 0b/0o/0x,
// экспоненты, подчёркивания для разделения разрядов и суффиксы типов (u32, f64 и т.д.).
func (l *lexer) readNumber() string {
	start := l.pos
	base := 10
	if l.ch == '0' {
		if l.peek() == 'b' || l.peek() == 'o' || l.peek() == 'x' {
			l.readChar()
			switch l.ch {
			case 'b': base=2; l.readChar()
			case 'o': base=8; l.readChar()
			case 'x': base=16; l.readChar()
			}
		}
	}
	for isDigitInBase(l.ch, base) || l.ch == '_' {
		l.readChar()
	}
	// дробная часть (только для base 10)
	if base == 10 && l.ch == '.' && isDigitInBase(l.peek(), 10) {
		l.readChar()
		for unicode.IsDigit(l.ch) || l.ch == '_' { l.readChar() }
	}
	// экспонента (только для base 10)
	if base == 10 && (l.ch == 'e' || l.ch == 'E') {
		l.readChar()
		if l.ch == '+' || l.ch == '-' { l.readChar() }
		for unicode.IsDigit(l.ch) || l.ch == '_' { l.readChar() }
	}
	// суффиксы типа u32, i64, f64 и т.д.
	for unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) {
		l.readChar()
	}
	return string(l.runes[start:l.pos])
}

// readString читает строковые литералы. Параметр prefix указывает на префикс
// (например, "r" для raw, "br" для raw byte), чтобы корректно обрабатывать хеши (#).
func (l *lexer) readString(prefix string) string {
	start := l.pos - len([]rune(prefix))
	hashes := 0
	// обработка raw-строк вида r#"..."# и т.п.
	if prefix == "r" || prefix == "br" {
		for l.ch == '#' { hashes++; l.readChar() }
		if l.ch != '"' { l.err = fmt.Errorf("invalid raw string"); return "" }
	}
	l.readChar() // skip opening '"'
	if prefix == "r" || prefix == "br" {
		for l.ch != 0 {
			if l.ch == '"' {
				l.readChar()
				m := 0
				for m < hashes && l.ch == '#' { m++; l.readChar() }
				if m == hashes { break }
			} else { l.readChar() }
		}
	} else {
		// обычные и byte-строки: учитываем escape-последовательности
		for l.ch != '"' && l.ch != 0 {
			if l.ch == '\\' {
				l.readChar()
				// потребляем тело escape-последовательности (если есть)
				if l.ch != 0 { l.readChar() }
				continue
			}
			l.readChar()
		}
		if l.ch == '"' {
			l.readChar() 
		} else {
			l.err = fmt.Errorf("unterminated string")
		}
	}
	return string(l.runes[start:l.pos])
}

// readAttr читает атрибуты Rust: #[...] и #![...]
// Поддерживает вложенные квадратные скобки внутри атрибута.
func (l *lexer) readAttr() string {
	start := l.pos
	l.readChar() // consume '#'
	if l.ch == '!' { l.readChar() }
	if l.ch != '[' { l.err = fmt.Errorf("invalid attribute"); return "" }
	l.readChar()
	depth := 1
	for l.ch != 0 && depth > 0 {
		if l.ch == '[' {
			depth++ 
		} else if l.ch == ']' { 
			depth--
		}
		l.readChar()
	}
	if depth>0 {
		l.err = fmt.Errorf("unterminated attribute")
	}
	return string(l.runes[start:l.pos])
}

// readOpOrPunct читает операторы и пунктуацию, пытаясь сначала матчить
// трёхсимвольные, затем двухсимвольные, затем односивольные последовательности.
func (l *lexer) readOpOrPunct() string {
	start := l.pos
	b1 := string(l.ch)
	b2 := b1 + string(l.peek())
	b3 := b2 + string(l.peekN(1))
	if l.operators[b3] || l.punctuations[b3] {
		l.readChar(); l.readChar(); l.readChar();
		return string(l.runes[start:l.pos])
	}
	if l.operators[b2] || l.punctuations[b2] {
		l.readChar(); l.readChar();
		return string(l.runes[start:l.pos])
	}
	l.readChar()
	return string(l.runes[start:l.pos])
}

// Вспомогательные предикаты для распознавания операторных и пунктуационных символов.
func isOpChar(ch rune) bool {
	return ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '%' ||
		ch == '=' || ch == '!' || ch == '<' || ch == '>' || ch == '&' || ch == '|' ||
		ch == '^' || ch == '~' || ch == '?'
}
func isPunct(ch rune) bool {
	return ch == '{' || ch == '}' || ch == '(' || ch == ')' || ch == '[' || ch == ']' ||
		ch == ';' || ch == ',' || ch == ':' || ch == '.' || ch == '#' || ch == '@' || ch == '!'
}

// containsDotOrExp проверяет, содержит ли строка точки или показатель экспоненты,
// используется при классификации числа как FLOAT.
func containsDotOrExp(s string) bool { 
	for _, c := range s { 
		if c=='.' || c=='e' || c=='E' {
			return true 
		} 
	} 
	return false 
}

// nextToken — центральная функция, которая анализирует текущую руну и формирует токен.
// Ведёт себя итеративно: пропускает пробелы/комментарии, затем вызывает соответствующие читатели.
func (l *lexer) nextToken() {
	l.skipWhitespace()
	if l.ch=='/' && (l.peek()=='/' || l.peek()=='*') { l.skipComment(); return }
	tok := token.Token{Line: l.line, Col: l.col}

	switch {
	case l.ch==0:
		return
	case l.ch=='\'' :
		lit, t := l.readLifetimeOrChar()
		tok.Literal = lit;
		tok.Type = t
	case unicode.IsLetter(l.ch) || l.ch=='_':
		ident := l.readIdentifier()
		// специальные префиксы для строк: r, br, b
		if ident=="r" && (l.ch=='"'||l.ch=='#') { 
			tok.Literal = l.readString("r"); 
			tok.Type = token.STRING 
		} else if ident=="br" && (l.ch=='"'||l.ch=='#') { 
			tok.Literal = l.readString("br"); 
			tok.Type = token.STRING 
		} else if ident=="b" && l.ch=='"' { 
			tok.Literal = l.readString("b"); 
			tok.Type = token.STRING 
		} else { 
			tok.Literal = ident; 
			if l.keywords[ident] { 
				tok.Type = token.KEYWORD 
			} else {
				tok.Type = token.IDENT 
			} 
		}
	case unicode.IsDigit(l.ch):
		tok.Literal = l.readNumber();
		if containsDotOrExp(tok.Literal) {
			tok.Type = token.FLOAT
		} else {
			tok.Type = token.INT
		}
	case l.ch=='"':
		tok.Literal = l.readString("");
		tok.Type = token.STRING
	case l.ch=='#':
		tok.Literal = l.readAttr();
		tok.Type = token.ATTRIBUTE
	case isOpChar(l.ch) || isPunct(l.ch):
		tok.Literal = l.readOpOrPunct()
		if l.operators[tok.Literal] {
			tok.Type = token.OPERATOR
		} else if l.punctuations[tok.Literal] {
			tok.Type = token.PUNCT
		} else {
			tok.Type = token.ILLEGAL
		}
	default:
		tok.Type = token.ILLEGAL;
		tok.Literal = string(l.ch);
		l.readChar()
	}

	if l.err==nil {
		l.tokens = append(l.tokens, tok)
	}
}
