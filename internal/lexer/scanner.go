// Пакет lexer: низкоуровневый сканер (работа с runes и позициями).
package lexer

// Scanner — упрощённый ридер по рун-строке. Предоставляет Peek/PeekN и позицию.
type Scanner struct {
	runes   []rune
	length  int
	pos     int // индекс текущей руны
	readPos int // индекс следующей руны
	ch      rune
	Line    int
	Col     int
}

// NewScanner создаёт новый сканер и сразу читает первую руну.
func NewScanner(input string) *Scanner {
	r := []rune(input)
	s := &Scanner{runes: r, length: len(r), pos: 0, readPos: 0, Line: 1, Col: 0}
	s.readChar()
	return s
}

// readChar продвигает сканер на следующую руну.
func (s *Scanner) readChar() {
	if s.readPos >= s.length {
		s.ch = 0
	} else {
		s.ch = s.runes[s.readPos]
	}
	s.pos = s.readPos
	s.readPos++
	if s.ch == '\n' {
		s.Line++
		s.Col = 0
	} else {
		s.Col++
	}
}

// Ch возвращает текущую руну.
func (s *Scanner) Ch() rune { return s.ch }

// Peek возвращает следующую руну без продвижения.
func (s *Scanner) Peek() rune {
	if s.readPos >= s.length {
		return 0
	}
	return s.runes[s.readPos]
}

// PeekN возвращает n-ую руну вперёд (n>=1), безопасно если выходит за пределы.
func (s *Scanner) PeekN(n int) rune {
	idx := s.readPos + n - 1
	if idx >= s.length || idx < 0 {
		return 0
	}
	return s.runes[idx]
}

// Next продвигает сканер и возвращает новую текущую руну.
func (s *Scanner) Next() rune { s.readChar(); return s.ch }

// Pos возвращает текущие координаты (line, col).
func (s *Scanner) Pos() (int, int) { return s.Line, s.Col }

// IsEOF возвращает true, если достигнут конец.
func (s *Scanner) IsEOF() bool { return s.ch == 0 }