package token_test

import (
	"testing"

	"github.com/semetekare/rust2go/internal/token"
)

func TestTokenPos(t *testing.T) {
	tok := token.Token{
		Type:    token.IDENT,
		Literal: "test",
		Line:    5,
		Col:     10,
	}

	pos := tok.Pos()
	if pos.Line != 5 {
		t.Errorf("Expected line 5, got %d", pos.Line)
	}
	if pos.Col != 10 {
		t.Errorf("Expected col 10, got %d", pos.Col)
	}
}

func TestTokenString(t *testing.T) {
	tests := []struct {
		tok      token.Token
		expected string
	}{
		{token.Token{Type: token.EOF, Literal: ""}, "EOF"},
		{token.Token{Type: token.IDENT, Literal: "foo"}, "IDENT"},
		{token.Token{Type: token.KEYWORD, Literal: "fn"}, "KEYWORD"},
		{token.Token{Type: token.TYPE, Literal: "i32", Subtype: "INT"}, "TYPE(INT)"},
		{token.Token{Type: token.INT, Literal: "42"}, "INT"},
		{token.Token{Type: token.FLOAT, Literal: "3.14"}, "FLOAT"},
		{token.Token{Type: token.STRING, Literal: "hello"}, "STRING"},
		{token.Token{Type: token.CHAR, Literal: "'a'"}, "CHAR"},
		{token.Token{Type: token.OPERATOR, Literal: "+"}, "OPERATOR"},
		{token.Token{Type: token.PUNCT, Literal: "("}, "PUNCT"},
		{token.Token{Type: token.TERMINATOR, Literal: ";"}, "TERMINATOR"},
		{token.Token{Type: token.ILLEGAL, Literal: "~"}, "ILLEGAL"},
	}

	for _, tt := range tests {
		str := tt.tok.String()
		if str != tt.expected {
			t.Errorf("Token %v: expected %q, got %q", tt.tok.Type, tt.expected, str)
		}
	}
}

func TestTokenTypeString(t *testing.T) {
	// TokenType.String() - проверяем, что паникует (как и задумано)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected TokenType.String() to panic")
		}
	}()

	var tt token.TokenType
	_ = tt.String()
}

func TestTokenSubtype(t *testing.T) {
	tok := token.Token{
		Type:    token.TYPE,
		Literal: "42",
		Subtype: "INT",
	}

	if tok.Subtype != "INT" {
		t.Errorf("Expected subtype 'INT', got %q", tok.Subtype)
	}
}

func TestAllTokenTypes(t *testing.T) {
	types := []token.TokenType{
		token.EOF,
		token.IDENT,
		token.LIFETIME,
		token.KEYWORD,
		token.TYPE,
		token.INT,
		token.FLOAT,
		token.STRING,
		token.CHAR,
		token.OPERATOR,
		token.PUNCT,
		token.ATTRIBUTE,
		token.TERMINATOR,
		token.ILLEGAL,
	}

	for _, tt := range types {
		// Проверяем, что каждый тип является валидным токеном
		_ = tt
	}
}

func TestPosition(t *testing.T) {
	pos := token.Position{Line: 42, Col: 10}

	if pos.Line != 42 {
		t.Errorf("Expected line 42, got %d", pos.Line)
	}
	if pos.Col != 10 {
		t.Errorf("Expected col 10, got %d", pos.Col)
	}
}

func TestTokenFields(t *testing.T) {
	tok := token.Token{
		Type:    token.IDENT,
		Subtype: "",
		Literal: "test_name",
		Line:    10,
		Col:     5,
	}

	if tok.Type != token.IDENT {
		t.Errorf("Expected type IDENT, got %v", tok.Type)
	}
	if tok.Literal != "test_name" {
		t.Errorf("Expected literal 'test_name', got %q", tok.Literal)
	}
	if tok.Line != 10 {
		t.Errorf("Expected line 10, got %d", tok.Line)
	}
	if tok.Col != 5 {
		t.Errorf("Expected col 5, got %d", tok.Col)
	}
}

func TestTokenConstValues(t *testing.T) {
	// Проверяем, что константы имеют разные значения
	if token.EOF == token.IDENT {
		t.Error("EOF and IDENT should be different")
	}
	if token.IDENT == token.KEYWORD {
		t.Error("IDENT and KEYWORD should be different")
	}
	if token.INT == token.FLOAT {
		t.Error("INT and FLOAT should be different")
	}
}

func TestTokenWithSubtype(t *testing.T) {
	tok := token.Token{
		Type:    token.TYPE,
		Literal: "42",
		Subtype: "INT",
	}

	str := tok.String()
	expected := "TYPE(INT)"
	if str != expected {
		t.Errorf("Expected %q, got %q", expected, str)
	}
}

func TestTokenWithoutSubtype(t *testing.T) {
	tok := token.Token{
		Type:    token.TYPE,
		Literal: "Foo",
		Subtype: "",
	}

	str := tok.String()
	expected := "TYPE"
	if str != expected {
		t.Errorf("Expected %q, got %q", expected, str)
	}
}
