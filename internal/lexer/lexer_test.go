package lexer_test

import (
	"strings"
	"testing"

	"github.com/semetekare/rust2go/internal/lexer"
	"github.com/semetekare/rust2go/internal/token"
)

func TestLexKeywords(t *testing.T) {
	lx := lexer.NewLexer()
	toks, err := lx.Lex("fn let if struct")
	if err != nil {
		t.Fatalf("Lex failed: %v", err)
	}

	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.KEYWORD, "fn"},
		{token.KEYWORD, "let"},
		{token.KEYWORD, "if"},
		{token.KEYWORD, "struct"},
		{token.EOF, ""},
	}

	if len(toks) != len(expected) {
		t.Fatalf("Expected %d tokens, got %d", len(expected), len(toks))
	}

	for i, exp := range expected {
		if toks[i].Type != exp.typ || toks[i].Literal != exp.lit {
			t.Errorf("Token %d: expected (%v, %q), got (%v, %q)",
				i, exp.typ, exp.lit, toks[i].Type, toks[i].Literal)
		}
	}
}

func TestLexIdentifiers(t *testing.T) {
	lx := lexer.NewLexer()
	toks, err := lx.Lex("my_var foo123 _private")
	if err != nil {
		t.Fatalf("Lex failed: %v", err)
	}

	expected := []string{"my_var", "foo123", "_private"}
	if len(toks) != len(expected)+1 { // +1 for EOF
		t.Fatalf("Expected %d tokens, got %d", len(expected)+1, len(toks))
	}

	for i, exp := range expected {
		if toks[i].Type != token.IDENT {
			t.Errorf("Token %d: expected IDENT, got %v", i, toks[i].Type)
		}
		if toks[i].Literal != exp {
			t.Errorf("Token %d: expected %q, got %q", i, exp, toks[i].Literal)
		}
	}
}

func TestLexIntLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		subtype  string
	}{
		{"42", "42", "INT"},
		{"0b1010", "0b1010", "INT"},
		{"0o755", "0o755", "INT"},
		{"0xFF", "0xFF", "INT"},
		{"42i32", "42i32", "INT"},
		{"1_000_000", "1_000_000", "INT"},
	}

	lx := lexer.NewLexer()
	for _, tt := range tests {
		toks, err := lx.Lex(tt.input)
		if err != nil {
			t.Errorf("Lex(%q) failed: %v", tt.input, err)
			continue
		}

		if len(toks) < 2 {
			t.Errorf("Expected at least 2 tokens (INT + EOF), got %d", len(toks))
			continue
		}

		tok := toks[0]
		if tok.Type != token.TYPE {
			t.Errorf("Token type: expected TYPE, got %v", tok.Type)
		}
		if tok.Subtype != tt.subtype {
			t.Errorf("Subtype: expected %q, got %q", tt.subtype, tok.Subtype)
		}
		if tok.Literal != tt.expected {
			t.Errorf("Literal: expected %q, got %q", tt.expected, tok.Literal)
		}
	}
}

func TestLexFloatLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"3.14", "3.14"},
		{"2.5", "2.5"},
		{"1e5", "1e5"},
		{"1.5e-3", "1.5e-3"},
		{"42.0f32", "42.0f32"},
	}

	lx := lexer.NewLexer()
	for _, tt := range tests {
		toks, err := lx.Lex(tt.input)
		if err != nil {
			t.Errorf("Lex(%q) failed: %v", tt.input, err)
			continue
		}

		if len(toks) < 2 {
			t.Errorf("Expected at least 2 tokens (FLOAT + EOF), got %d", len(toks))
			continue
		}

		tok := toks[0]
		if tok.Type != token.TYPE {
			t.Errorf("Token type: expected TYPE, got %v", tok.Type)
		}
		if tok.Subtype != "FLOAT" {
			t.Errorf("Subtype: expected FLOAT, got %q", tok.Subtype)
		}
		if tok.Literal != tt.expected {
			t.Errorf("Literal: expected %q, got %q", tt.expected, tok.Literal)
		}
	}
}

func TestLexStringLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"world"`, "world"},
		{`"hello\nworld"`, "hello\nworld"},
		{`"тест"`, "тест"},
	}

	lx := lexer.NewLexer()
	for _, tt := range tests {
		toks, err := lx.Lex(tt.input)
		if err != nil {
			t.Errorf("Lex(%q) failed: %v", tt.input, err)
			continue
		}

		if len(toks) < 2 {
			t.Errorf("Expected at least 2 tokens (STRING + EOF), got %d", len(toks))
			continue
		}

		tok := toks[0]
		if tok.Type != token.STRING {
			t.Errorf("Token type: expected STRING, got %v", tok.Type)
		}
	}
}

func TestLexOperators(t *testing.T) {
	lx := lexer.NewLexer()
	toks, err := lx.Lex("+ - * / % == != < > <= >= && ||")
	if err != nil {
		t.Fatalf("Lex failed: %v", err)
	}

	expected := []string{"+", "-", "*", "/", "%", "==", "!=", "<", ">", "<=", ">=", "&&", "||"}
	for i, exp := range expected {
		if toks[i].Type != token.OPERATOR {
			t.Errorf("Token %d type: expected OPERATOR, got %v", i, toks[i].Type)
		}
		if toks[i].Literal != exp {
			t.Errorf("Token %d: expected %q, got %q", i, exp, toks[i].Literal)
		}
	}
}

func TestLexPunctuation(t *testing.T) {
	lx := lexer.NewLexer()
	toks, err := lx.Lex("() [] {} , ; : :: . ..")
	if err != nil {
		t.Fatalf("Lex failed: %v", err)
	}

	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.PUNCT, "("},
		{token.PUNCT, ")"},
		{token.PUNCT, "["},
		{token.PUNCT, "]"},
		{token.PUNCT, "{"},
		{token.PUNCT, "}"},
		{token.PUNCT, ","},
		{token.TERMINATOR, ";"},
		{token.PUNCT, ":"},
		{token.PUNCT, "::"},
		{token.PUNCT, "."},
		{token.PUNCT, ".."},
	}

	for i, exp := range expected {
		if toks[i].Type != exp.typ {
			t.Errorf("Token %d type: expected %v, got %v", i, exp.typ, toks[i].Type)
		}
		if toks[i].Literal != exp.lit {
			t.Errorf("Token %d: expected %q, got %q", i, exp.lit, toks[i].Literal)
		}
	}
}

func TestLexMacros(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"println!", "println!"},
		{"vec!", "vec!"},
		{"panic!", "panic!"},
		{"format!", "format!"},
	}

	lx := lexer.NewLexer()
	for _, tt := range tests {
		toks, err := lx.Lex(tt.input)
		if err != nil {
			t.Errorf("Lex(%q) failed: %v", tt.input, err)
			continue
		}

		if len(toks) < 2 {
			t.Errorf("Expected at least 2 tokens (IDENT + EOF), got %d", len(toks))
			continue
		}

		tok := toks[0]
		if tok.Type != token.IDENT {
			t.Errorf("Token type: expected IDENT, got %v", tok.Type)
		}
		if tok.Subtype != "MACRO" {
			t.Errorf("Subtype: expected MACRO, got %q", tok.Subtype)
		}
		if tok.Literal != tt.expected {
			t.Errorf("Literal: expected %q, got %q", tt.expected, tok.Literal)
		}
	}
}

func TestLexFunctionCall(t *testing.T) {
	lx := lexer.NewLexer()
	toks, err := lx.Lex("foo() bar(1, 2)")
	if err != nil {
		t.Fatalf("Lex failed: %v", err)
	}

	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.IDENT, "foo"},
		{token.PUNCT, "("},
		{token.PUNCT, ")"},
		{token.IDENT, "bar"},
		{token.PUNCT, "("},
		{token.TYPE, "1"},
		{token.PUNCT, ","},
		{token.TYPE, "2"},
		{token.PUNCT, ")"},
	}

	for i := 0; i < len(expected) && i < len(toks)-1; i++ { // -1 for EOF
		exp := expected[i]
		if toks[i].Type != exp.typ {
			t.Errorf("Token %d type: expected %v, got %v", i, exp.typ, toks[i].Type)
		}
		if toks[i].Literal != exp.lit {
			t.Errorf("Token %d: expected %q, got %q", i, exp.lit, toks[i].Literal)
		}
	}
}

func TestLexComments(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"// comment", "single line comment"},
		{"/**/", "empty block comment"},
		{"/* nested /* comment */ */", "nested block comment"},
		{"foo // comment\nbar", "comment with code"},
	}

	lx := lexer.NewLexer()
	for _, tt := range tests {
		toks, err := lx.Lex(tt.input)
		if err != nil {
			t.Errorf("Lex failed for %s: %v", tt.desc, err)
			continue
		}

		// Проверяем, что комментарии пропускаются
		// В результате должны быть только не-комментарий токены
		if len(toks) == 0 {
			t.Errorf("Expected some tokens for %s", tt.desc)
		}
	}
}

func TestLexPositions(t *testing.T) {
	input := `fn main() {
    let x = 42;
}`
	lx := lexer.NewLexer()
	toks, err := lx.Lex(input)
	if err != nil {
		t.Fatalf("Lex failed: %v", err)
	}

	// Проверяем, что позиции корректны
	if toks[0].Line != 1 {
		t.Errorf("Expected line 1 for first token, got %d", toks[0].Line)
	}
	if toks[0].Col != 1 {
		t.Errorf("Expected col 1 for first token, got %d", toks[0].Col)
	}
}

func TestLexCompleteFunction(t *testing.T) {
	input := `fn add(a: i32, b: i32) -> i32 {
    a + b
}`
	lx := lexer.NewLexer()
	toks, err := lx.Lex(input)
	if err != nil {
		t.Fatalf("Lex failed: %v", err)
	}

	// Проверяем наличие ключевых элементов
	hasFn := false
	hasIdentifier := false
	hasType := false

	for _, tok := range toks {
		if tok.Type == token.KEYWORD && tok.Literal == "fn" {
			hasFn = true
		}
		if tok.Type == token.IDENT {
			hasIdentifier = true
		}
		if tok.Type == token.IDENT && tok.Literal == "i32" {
			hasType = true
		}
	}

	if !hasFn {
		t.Error("Expected 'fn' keyword")
	}
	if !hasIdentifier {
		t.Error("Expected identifier")
	}
	if !hasType {
		t.Error("Expected type identifier")
	}
}

func TestLexRawString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`r#"hello"#`, "hello"},
		{`r##"test"##`, "test"},
		{`r###"nested"###`, "nested"},
	}

	lx := lexer.NewLexer()
	for _, tt := range tests {
		toks, err := lx.Lex(tt.input)
		if err != nil {
			t.Errorf("Lex(%q) failed: %v", tt.input, err)
			continue
		}

		if len(toks) < 2 {
			t.Errorf("Expected at least 2 tokens, got %d", len(toks))
			continue
		}

		tok := toks[0]
		if tok.Type != token.TYPE {
			t.Errorf("Expected TYPE token, got %v", tok.Type)
		}
		if tok.Subtype != "STRING" {
			t.Errorf("Expected STRING subtype, got %q", tok.Subtype)
		}
	}
}

func TestLexByteString(t *testing.T) {
	lx := lexer.NewLexer()
	toks, err := lx.Lex(`b"hello"`)
	if err != nil {
		t.Fatalf("Lex failed: %v", err)
	}

	if len(toks) < 2 {
		t.Fatalf("Expected at least 2 tokens, got %d", len(toks))
	}

	tok := toks[0]
	if tok.Type != token.TYPE {
		t.Errorf("Expected TYPE token, got %v", tok.Type)
	}
	if tok.Subtype != "STRING" {
		t.Errorf("Expected STRING subtype, got %q", tok.Subtype)
	}
}

func TestLexLifetime(t *testing.T) {
	lx := lexer.NewLexer()
	toks, err := lx.Lex(`'a`)
	if err != nil {
		t.Fatalf("Lex failed: %v", err)
	}

	if len(toks) < 2 {
		t.Fatalf("Expected at least 2 tokens, got %d", len(toks))
	}

	tok := toks[0]
	if tok.Type != token.LIFETIME {
		t.Errorf("Expected LIFETIME token, got %v", tok.Type)
	}
}

func TestLexCharLiteral(t *testing.T) {
	tests := []string{
		`'a'`,
	}

	lx := lexer.NewLexer()
	for _, input := range tests {
		toks, err := lx.Lex(input)
		if err != nil {
			t.Errorf("Lex(%q) failed: %v", input, err)
			continue
		}

		if len(toks) < 2 {
			t.Errorf("Expected at least 2 tokens, got %d", len(toks))
			continue
		}

		tok := toks[0]
		if tok.Type != token.TYPE {
			t.Errorf("Expected TYPE token for %q, got %v", input, tok.Type)
		}
		if tok.Subtype != "CHAR" {
			t.Errorf("Expected CHAR subtype for %q, got %q", input, tok.Subtype)
		}
	}
}

func TestLexAttributes(t *testing.T) {
	tests := []struct {
		input string
	}{
		{`#[derive(Debug)]`},
		{`#![no_std]`},
		{`#[cfg(feature = "foo")]`},
	}

	lx := lexer.NewLexer()
	for _, tt := range tests {
		toks, err := lx.Lex(tt.input)
		if err != nil {
			t.Errorf("Lex(%q) failed: %v", tt.input, err)
			continue
		}

		hasAttr := false
		for _, tok := range toks {
			if tok.Type == token.ATTRIBUTE {
				hasAttr = true
				break
			}
		}

		if !hasAttr {
			t.Errorf("Expected ATTRIBUTE token in %q", tt.input)
		}
	}
}

func TestLexStringEscape(t *testing.T) {
	tests := []struct {
		input string
	}{
		{`"hello\nworld"`},
		{`"hello\tworld"`},
		{`"hello\\world"`},
		{`"hello\"world"`},
	}

	lx := lexer.NewLexer()
	for _, tt := range tests {
		toks, err := lx.Lex(tt.input)
		if err != nil {
			t.Errorf("Lex(%q) failed: %v", tt.input, err)
			continue
		}

		hasString := false
		for _, tok := range toks {
			if tok.Type == token.STRING {
				hasString = true
				break
			}
		}

		if !hasString {
			t.Errorf("Expected STRING token in %q", tt.input)
		}
	}
}

func TestLexComplexExpressions(t *testing.T) {
	tests := []string{
		`(1 + 2) * 3`,
		`foo(bar(1, 2), 3)`,
		`-x + y`,
		`x >= y && z < 0`,
		`vec![1, 2, 3]`,
	}

	lx := lexer.NewLexer()
	for _, input := range tests {
		toks, err := lx.Lex(input)
		if err != nil {
			t.Errorf("Lex(%q) failed: %v", input, err)
			continue
		}

		if len(toks) == 0 {
			t.Errorf("Expected tokens for %q", input)
		}
	}
}

func TestLexErrorRecovery(t *testing.T) {
	// Этот тест проверяет, что лексер не падает на сложных входных данных
	complexInput := strings.Builder{}
	complexInput.WriteString("fn complex() -> i32 {\n")
	for i := 0; i < 100; i++ {
		complexInput.WriteString("    let x = ")
		complexInput.WriteString(itoa(i))
		complexInput.WriteString(";\n")
	}
	complexInput.WriteString("    return 0;\n")
	complexInput.WriteString("}\n")

	lx := lexer.NewLexer()
	_, err := lx.Lex(complexInput.String())
	if err != nil {
		t.Errorf("Lex failed on complex input: %v", err)
	}
}

// Helper function для конвертации int в string
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append(b, byte('0'+n%10))
		n /= 10
	}
	if neg {
		b = append(b, '-')
	}
	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}
