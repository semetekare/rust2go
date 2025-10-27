// internal/parser/parser.go

// Package parser реализует рекурсивно-нисходящий парсер с базовым восстановлением после ошибок
// для Rust-подобного языка, транслируемого в Go.
package parser

import (
	"fmt"

	"github.com/semetekare/rust2go/internal/ast"
	"github.com/semetekare/rust2go/internal/token"
)

// Parser — основной парсер, управляющий процессом синтаксического анализа.
// Поддерживает сбор ошибок и базовое восстановление после синтаксических ошибок (error recovery).
type Parser struct {
	stream TokenStream  // Поток токенов, полученный от лексического анализатора.
	errors []ParseError // Список накопленных ошибок парсинга.
}

// ParseError представляет ошибку синтаксического анализа.
// Содержит диагностическое сообщение, токен, вызвавший ошибку, и его позицию в исходном коде.
type ParseError struct {
	Msg string         // Описание ошибки.
	Tok token.Token    // Токен, при обработке которого возникла ошибка.
	Pos token.Position // Позиция токена в исходном файле.
}

// String возвращает человекочитаемое строковое представление ошибки парсинга.
func (pe ParseError) String() string {
	return fmt.Sprintf("Parse error at %d:%d: %s (got '%s')", pe.Pos.Line, pe.Pos.Col, pe.Msg, pe.Tok.Literal)
}

// NewParser создаёт новый экземпляр парсера из списка токенов.
// Токены должны быть получены от лексического анализатора (lexer).
func NewParser(tokens []token.Token) *Parser {
	return &Parser{stream: NewTokenStream(tokens)}
}

// ParseFile запускает полный синтаксический анализ входного потока токенов.
// Возвращает корневой узел AST (Crate) и список всех обнаруженных ошибок.
// Даже при наличии ошибок парсер пытается построить частично корректное AST.
func (p *Parser) ParseFile() (*ast.Crate, []ParseError) {
	ast := p.ParseCrate()
	return ast, p.errors
}

// error добавляет новую ошибку в список ошибок парсера.
// Принимает диагностическое сообщение и токен, вызвавший ошибку.
func (p *Parser) error(msg string, tok token.Token) {
	p.errors = append(p.errors, ParseError{Msg: msg, Tok: tok, Pos: tok.Pos()})
}

// recover реализует базовую стратегию восстановления после ошибки (error recovery).
// Пропускает токены до тех пор, пока не встретит один из указанных синхронизирующих токенов
// (например, ";", "}", или другие разделители), чтобы позволить парсеру продолжить работу.
// Возвращает true, если восстановление было выполнено (в том числе при достижении EOF).
// Если ошибок нет, восстановление не требуется и функция возвращает false.
func (p *Parser) recover(syncs ...string) bool {
	// Если ошибок нет — восстановление не нужно
	if len(p.errors) == 0 {
		return false
	}
	for !p.stream.IsEOF() {
		tok := p.stream.Peek()
		// Если текущий токен — один из заданных синхронизирующих литералов,
		// останавливаемся и оставляем его в потоке для последующей обработки
		for _, s := range syncs {
			if tok.Literal == s {
				return true
			}
		}
		// Если встретили явный конец оператора или блока — потребляем токен и завершаем восстановление
		if tok.Type == token.TERMINATOR || (tok.Type == token.PUNCT && (tok.Literal == "}" || tok.Literal == ";")) {
			p.stream.Next()
			return true
		}
		// Иначе пропускаем текущий токен и продолжаем поиск точки синхронизации
		p.stream.Next()
	}
	return true
}
