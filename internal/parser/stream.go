// internal/parser/stream.go

// Package parser содержит реализацию парсера, преобразующего последовательность токенов
// в абстрактное синтаксическое дерево (AST).
package parser

import "github.com/semetekare/rust2go/internal/token"

// TokenStream — интерфейс для последовательного чтения токенов во время парсинга.
// Предоставляет методы для безопасного продвижения по потоку токенов,
// просмотра следующего токена без его извлечения и проверки конца входных данных.
type TokenStream interface {
	// Next возвращает следующий токен из потока и перемещает курсор вперёд.
	// Если достигнут конец потока, возвращается токен типа token.EOF.
	Next() token.Token

	// Peek возвращает следующий токен без перемещения курсора («заглядывает» вперёд).
	// При достижении конца возвращается токен типа token.EOF.
	Peek() token.Token

	// IsEOF возвращает true, если следующий токен — это конец файла (EOF).
	IsEOF() bool

	// Pos возвращает позицию следующего токена в исходном коде.
	// Если достигнут конец потока, возвращается позиция токена EOF.
	Pos() token.Position
}

// tokenStreamImpl — конкретная реализация интерфейса TokenStream,
// работающая с заранее сформированным срезом токенов ([]token.Token).
// Используется для парсинга после завершения лексического анализа.
type tokenStreamImpl struct {
	tokens []token.Token // Список токенов, полученных от лексера.
	pos    int           // Текущая позиция курсора в срезе tokens.
}

// NewTokenStream создаёт новый экземпляр TokenStream на основе переданного среза токенов.
// Начальная позиция курсора устанавливается в 0.
func NewTokenStream(tokens []token.Token) TokenStream {
	return &tokenStreamImpl{tokens: tokens, pos: 0}
}

// Next возвращает текущий токен и перемещает курсор на следующую позицию.
// Если курсор выходит за пределы среза, возвращается токен EOF.
func (ts *tokenStreamImpl) Next() token.Token {
	if ts.pos >= len(ts.tokens) {
		return token.Token{Type: token.EOF}
	}
	tok := ts.tokens[ts.pos]
	ts.pos++
	return tok
}

// Peek возвращает токен в текущей позиции без изменения курсора.
// Если позиция выходит за пределы среза, возвращается токен EOF.
func (ts *tokenStreamImpl) Peek() token.Token {
	if ts.pos >= len(ts.tokens) {
		return token.Token{Type: token.EOF}
	}
	return ts.tokens[ts.pos]
}

// IsEOF проверяет, достиг ли курсор конца потока токенов.
// Возвращает true, если следующий токен — EOF.
func (ts *tokenStreamImpl) IsEOF() bool {
	return ts.Peek().Type == token.EOF
}

// Pos возвращает позицию следующего токена (того, который вернёт Peek).
// Если поток исчерпан, возвращается позиция фиктивного токена EOF.
func (ts *tokenStreamImpl) Pos() token.Position {
	return ts.Peek().Pos()
}
