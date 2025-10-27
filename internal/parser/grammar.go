// internal/parser/grammar.go

// Package parser реализует рекурсивный спуск для парсинга входного потока токенов
// в абстрактное синтаксическое дерево (AST), соответствующее Rust-подобному языку.
package parser

import (
	"fmt"

	"github.com/semetekare/rust2go/internal/ast"
	"github.com/semetekare/rust2go/internal/token"
)

// leftAssoc — флаг, указывающий, что операторы левоассоциативны.
// Используется при построении бинарных выражений.
const leftAssoc = true

// ParseCrate парсит корневой узел AST — единицу компиляции (crate).
// Грамматика: Crate ::= InnerAttribute* Item*
// Метод последовательно парсит все элементы верхнего уровня до конца входного потока.
// В случае ошибки в одном из элементов пытается восстановиться, пропуская токены,
// чтобы избежать зацикливания.
func (p *Parser) ParseCrate() *ast.Crate {
	pos := p.stream.Pos()
	items := []ast.Item{}
	for !p.stream.IsEOF() {
		item := p.ParseItem()
		if item != nil {
			items = append(items, item)
		} else {
			// Если ParseItem вернул nil, значит была ошибка.
			// Пропускаем токен, чтобы избежать бесконечного цикла, если error recovery не справится.
			if p.stream.IsEOF() {
				break
			}
			p.stream.Next()
		}
	}
	return ast.NewCrate(pos, items)
}

// ParseItem парсит элемент верхнего уровня (item): функцию, структуру и т.д.
// Грамматика: Item ::= OuterAttribute* (Function | Struct | ... )?
// Поддерживает пропуск атрибутов (например, #[derive(...)]).
// На данный момент реализованы только "fn" и "struct".
// В случае неизвестного элемента возвращает nil и регистрирует ошибку.
func (p *Parser) ParseItem() ast.Item {
	// Пропускаем все атрибуты перед элементом
	for p.stream.Peek().Type == token.ATTRIBUTE {
		p.stream.Next() // пропускаем атрибут
	}
	tok := p.stream.Peek()
	pos := tok.Pos()
	if tok.Type == token.KEYWORD {
		switch tok.Literal {
		case "fn":
			p.stream.Next() // потребляем "fn"
			nameTok := p.expect(token.IDENT, "", "identifier after fn")
			name := nameTok.Literal
			// Парсим параметры функции
			params := []ast.Param{}
			p.expect(token.PUNCT, "(", "(")
			// Обрабатываем пустой список параметров
			for !p.stream.IsEOF() && !(p.stream.Peek().Type == token.PUNCT && p.stream.Peek().Literal == ")") {
				paramNameTok := p.expect(token.IDENT, "", "param name")
				paramName := paramNameTok.Literal
				p.expect(token.PUNCT, ":", ":")
				paramType := p.ParseType()
				params = append(params, *ast.NewParam(paramNameTok.Pos(), paramName, paramType))
				if p.stream.Peek().Literal == "," {
					p.stream.Next()
					continue
				}
				break
			}
			p.expect(token.PUNCT, ")", ")")
			// Необязательный возвращаемый тип
			var retType ast.Type
			if p.stream.Peek().Literal == "->" {
				p.stream.Next()
				retType = p.ParseType()
			} else {
				retType = ast.NewPathType(pos, "()") // тип по умолчанию — unit
			}
			body := p.ParseBlock()
			return ast.NewFunction(pos, name, params, retType, body)
		case "struct":
			p.stream.Next()
			nameTok := p.expect(token.IDENT, "", "struct name")
			name := nameTok.Literal
			p.expect(token.PUNCT, "{", "{")
			fields := []ast.Field{}
			for !p.stream.IsEOF() && !(p.stream.Peek().Type == token.PUNCT && p.stream.Peek().Literal == "}") {
				fieldNameTok := p.expect(token.IDENT, "", "field name")
				p.expect(token.PUNCT, ":", ":")
				fieldType := p.ParseType()
				fields = append(fields, *ast.NewField(fieldNameTok.Pos(), fieldNameTok.Literal, fieldType))
				if p.stream.Peek().Literal == "," {
					p.stream.Next()
					continue
				}
				break
			}
			p.expect(token.PUNCT, "}", "}")
			return ast.NewStruct(pos, name, fields)
		}
	}
	// Не распознан элемент верхнего уровня
	p.error("expected item (fn, struct, etc.)", tok)
	return nil
}

// ParseExpr парсит выражение с учётом приоритетов операторов.
// Использует рекурсивный спуск и вспомогательный метод parseBinary для обработки
// бинарных операций. Поддерживаемые операторы: сравнения, арифметика, логические.
func (p *Parser) ParseExpr() ast.Expr {
	return p.parseBinary(p.parseUnary, []string{"==", "!=", "<", ">", "+", "-", "*", "/", "%", "&&", "||"}, leftAssoc)
}

// parseBinary — обобщённый метод для парсинга бинарных выражений.
// Принимает:
//   - nextParser: функцию для парсинга подвыражения более высокого приоритета,
//   - ops: список операторов текущего приоритета,
//   - assoc: ассоциативность (в текущей реализации всегда левая).
//
// Возвращает построенное бинарное выражение или nil в случае ошибки.
func (p *Parser) parseBinary(nextParser func() ast.Expr, ops []string, assoc bool) ast.Expr {
	expr := nextParser()
	for {
		if expr == nil {
			return nil
		}
		opTok := p.stream.Peek()
		if !(opTok.Type == token.OPERATOR || opTok.Type == token.PUNCT) {
			break
		}
		op := opTok.Literal
		found := false
		for _, o := range ops {
			if op == o {
				found = true
				break
			}
		}
		if !found {
			break
		}
		p.stream.Next()
		right := nextParser()
		if right == nil {
			p.error("expected expression after operator", p.stream.Peek())
			return nil
		}
		expr = ast.NewBinaryExpr(expr.Pos(), expr, op, right)
	}
	return expr
}

// parseUnary парсит унарные выражения: `-x`, `!flag`, `~bits`.
// Если унарный оператор отсутствует, делегирует парсинг primary-выражениям.
func (p *Parser) parseUnary() ast.Expr {
	tok := p.stream.Peek()
	if tok.Type == token.OPERATOR && (tok.Literal == "-" || tok.Literal == "!" || tok.Literal == "~") {
		p.stream.Next()
		primary := p.parsePrimary()
		if primary == nil {
			return nil
		}
		return ast.NewUnaryExpr(tok.Pos(), tok.Literal, primary)
	}
	return p.parsePrimary()
}

// parsePrimary парсит первичные (атомарные) выражения:
// литералы (числа, строки, булевы), идентификаторы, вызовы функций, блоки и скобочные выражения.
// Поддерживает вызовы вида `foo()` и обработку макросов (например, `println!`), хотя макросы
// пока не обрабатываются отдельно — они трактуются как обычные вызовы.
// В случае ошибки потребляет проблемный токен, чтобы избежать зацикливания.
func (p *Parser) parsePrimary() ast.Expr {
	tok := p.stream.Peek()
	pos := tok.Pos()
	switch tok.Type {
	case token.TYPE: // Для числовых литералов с подтипом (например, INT, FLOAT)
		p.stream.Next()
		return ast.NewLiteral(pos, tok.Subtype, tok.Literal)
	case token.CHAR:
		p.stream.Next()
		return ast.NewLiteral(pos, "CHAR", tok.Literal)
	case token.INT, token.FLOAT:
		p.stream.Next()
		return ast.NewLiteral(pos, tok.Type.String(), tok.Literal)
	case token.STRING:
		p.stream.Next()
		return ast.NewLiteral(pos, "STRING", tok.Literal)
	case token.KEYWORD:
		if tok.Literal == "true" || tok.Literal == "false" {
			p.stream.Next()
			return ast.NewLiteral(pos, "BOOL", tok.Literal)
		}
	case token.IDENT:
		idTok := p.stream.Next()
		isMacro := false
		if p.stream.Peek().Literal == "!" {
			isMacro = true
			p.stream.Next() // потребляем '!'
		}

		// Проверяем, идёт ли после идентификатора '(' — тогда это вызов
		if p.stream.Peek().Type == token.PUNCT && p.stream.Peek().Literal == "(" {
			p.stream.Next() // потребляем '('
			args := []ast.Expr{}

			// Пустой список аргументов
			if p.stream.Peek().Type == token.PUNCT && p.stream.Peek().Literal == ")" {
				p.stream.Next()
				fnLit := ast.NewLiteral(idTok.Pos(), "IDENT", idTok.Literal)
				call := ast.NewCallExpr(idTok.Pos(), fnLit, args)
				_ = isMacro // зарезервировано для будущей обработки макросов
				return call
			}

			// Парсим аргументы
			for {
				arg := p.ParseExpr()
				if arg != nil {
					args = append(args, arg)
				} else {
					// Ошибка в аргументе: восстанавливаемся до ',' или ')'
					for !p.stream.IsEOF() && !(p.stream.Peek().Literal == "," || p.stream.Peek().Literal == ")") {
						p.stream.Next()
					}
					if p.stream.Peek().Literal == "," {
						p.stream.Next()
						continue
					}
				}

				if p.stream.Peek().Literal == "," {
					p.stream.Next()
					continue
				}
				break
			}

			p.expect(token.PUNCT, ")", ")")
			fnLit := ast.NewLiteral(idTok.Pos(), "IDENT", idTok.Literal)
			call := ast.NewCallExpr(idTok.Pos(), fnLit, args)
			_ = isMacro
			return call
		}

		// Иначе — просто переменная или путь
		return ast.NewLiteral(idTok.Pos(), "IDENT", idTok.Literal)
	case token.PUNCT:
		if tok.Literal == "{" {
			block := p.ParseBlock()
			return ast.NewBlockExpr(pos, block)
		}
		if tok.Literal == "(" {
			p.stream.Next()
			inner := p.ParseExpr()
			p.expect(token.PUNCT, ")", ")")
			return inner
		}
	}

	p.error("expected primary expression", tok)
	p.stream.Next() // ВАЖНО: потребляем токен, вызвавший ошибку
	return nil
}

// ParseStmt парсит оператор (statement).
// Поддерживает:
//   - объявления переменных: `let x: i32 = 42;`
//   - выражения с точкой с запятой: `foo();`
//   - tail-выражения в блоках (без ';').
//
// В случае синтаксической ошибки возвращает nil и полагается на восстановление в вызывающем коде.
func (p *Parser) ParseStmt() ast.Stmt {
	tok := p.stream.Peek()
	if tok.Literal == "let" {
		p.stream.Next()
		nameTok := p.expect(token.IDENT, "", "let binding name")
		var typ ast.Type
		if p.stream.Peek().Literal == ":" {
			p.stream.Next()
			typ = p.ParseType()
		}
		if p.expect(token.OPERATOR, "=", "=").Type == token.EOF {
			return nil
		}
		init := p.ParseExpr()
		if init == nil {
			return nil
		}
		if p.expect(token.TERMINATOR, ";", ";").Type == token.EOF {
			return nil
		}

		if typ == nil {
			typ = ast.NewPathType(token.Position{}, "infer") // тип будет выведен позже
		}
		return ast.NewLetStmt(tok.Pos(), nameTok.Literal, typ, init)
	}

	expr := p.ParseExpr()
	if expr == nil {
		return nil
	}

	// Выражение с точкой с запятой
	if p.stream.Peek().Type == token.TERMINATOR {
		p.stream.Next()
		return ast.NewExprStmt(expr.Pos(), expr)
	}

	// Tail-выражение в блоке (например, последнее выражение функции)
	if p.stream.Peek().Literal == "}" {
		return ast.NewExprStmt(expr.Pos(), expr)
	}

	// Нет ни ';', ни '}' — ошибка
	p.error("expected ';' after expression", p.stream.Peek())
	return nil
}

// ParseBlock парсит блок кода, ограниченный фигурными скобками.
// Грамматика: Block ::= "{" Stmt* "}"
// При ошибке в одном из операторов вызывает метод восстановления `recover`,
// чтобы продолжить парсинг последующих операторов.
func (p *Parser) ParseBlock() *ast.Block {
	pos := p.stream.Pos()
	p.expect(token.PUNCT, "{", "{")
	stmts := []ast.Stmt{}

	for !p.stream.IsEOF() && p.stream.Peek().Literal != "}" {
		stmt := p.ParseStmt()
		if stmt != nil {
			stmts = append(stmts, stmt)
		} else {
			// Ошибка в операторе — восстанавливаемся до точки с запятой
			p.recover(";")
		}
	}
	p.expect(token.PUNCT, "}", "}")
	return ast.NewBlock(pos, stmts)
}

// ParseType парсит простой тип по имени (например, `i32`, `String`).
// Поддерживает ссылки (`&T`), но без обработки lifetime'ов.
// Грамматика: Type ::= Path | &Type | ...
// В текущей реализации `&` просто игнорируется, и парсится базовый тип.
func (p *Parser) ParseType() ast.Type {
	if p.stream.Peek().Literal == "&" {
		p.stream.Next() // потребляем '&'
		// TODO: добавить поддержку lifetime'ов, например, &'a T
		return p.ParseType()
	}
	tok := p.expect(token.IDENT, "", "type")
	return ast.NewPathType(tok.Pos(), tok.Literal)
}

// ParseField парсит поле структуры.
// Грамматика: Field ::= IDENTIFIER ":" Type
// Используется при парсинге определения структуры.
func (p *Parser) ParseField() *ast.Field {
	nameTok := p.expect(token.IDENT, "", "field name")
	p.expect(token.PUNCT, ":", ":")
	typ := p.ParseType()
	return ast.NewField(nameTok.Pos(), nameTok.Literal, typ)
}

// expect проверяет, что следующий токен соответствует ожидаемому типу и/или литералу.
// Если нет — регистрирует ошибку и возвращает текущий токен.
// Если да — потребляет токен и возвращает его.
// Параметр `desc` используется в сообщении об ошибке для пояснения контекста.
func (p *Parser) expect(typ token.TokenType, lit string, desc string) token.Token {
	if p.stream.IsEOF() {
		p.error(fmt.Sprintf("expected %s but got EOF", desc), token.Token{Type: token.EOF})
		return token.Token{Type: token.EOF}
	}

	tok := p.stream.Peek()
	match := tok.Type == typ
	if lit != "" {
		match = match && tok.Literal == lit
	}

	if !match {
		if desc == "" {
			desc = lit
		}
		p.error(fmt.Sprintf("expected %s (got '%s')", desc, tok.Literal), tok)
		return tok
	}

	return p.stream.Next()
}
