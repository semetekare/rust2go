// Пакет lexer: статические таблицы операторов/ключевых слов/пунктуации.
package lexer

var Keywords = map[string]bool{ // common subset
	"as": true, "break": true, "const": true, "continue": true, "crate": true,
	"else": true, "enum": true, "extern": true, "false": true, "fn": true,
	"for": true, "if": true, "impl": true, "in": true, "let": true,
	"loop": true, "match": true, "mod": true, "move": true, "mut": true,
	"pub": true, "ref": true, "return": true, "self": true, "Self": true,
	"static": true, "struct": true, "super": true, "trait": true, "true": true,
	"type": true, "unsafe": true, "use": true, "where": true, "while": true,
	"async": true, "await": true, "dyn": true, "abstract": true, "become": true,
	"box": true, "do": true, "final": true, "macro": true, "override": true,
	"priv": true, "try": true, "typeof": true, "unsized": true, "virtual": true,
	"yield": true,
}

var Operators = map[string]bool{
	"+": true, "-": true, "*": true, "/": true, "%": true,
	"=": true, "==": true, "!=": true, "<": true, ">": true,
	"<=": true, ">=": true, "&&": true, "||": true, "->": true,
}

var Punctuations = map[string]bool{
	"{": true, "}": true, "(": true, ")": true, "[": true, "]": true,
	";": true, ",": true, ":": true, "::": true, ".": true, "..": true,
}

// BuiltinMacros содержит список встроенных макросов Rust (макросы, заканчивающиеся на !).
var BuiltinMacros = map[string]bool{
	"println!": true, "print!": true, "eprintln!": true, "eprint!": true,
	"format!": true, "panic!": true, "assert!": true, "assert_eq!": true,
	"vec!": true, "format_args!": true, "write!": true, "writeln!": true,
	"dbg!": true, "todo!": true, "unimplemented!": true, "unreachable!": true,
}
