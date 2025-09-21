```
rust2go/ # корень проекта 
├── cmd/ 
│ └── rust2go/ # CLI: точка входа (main.go) 
├── internal/ 
│ ├── token/ # описание типов токенов и позиций 
│ │ └── token.go 
│ ├── lexer/ # реализация лексера (stateless wrapper + state) 
│ │ ├── lexer.go 
│ │ ├── scanner.go # low-level чтение рун, peekN, позиционирование 
│ │ └── tables.go # списки ключевых слов, операторов, пунктуации 
│ ├── parser/ # синтаксический анализатор (LL/recursive descent / Pratt) 
│ │ ├── parser.go 
│ │ ├── stream.go # TokenStream интерфейс и реализация 
│ │ └── grammar.go # отдельные правила/парсеры для конструкций 
│ ├── ast/ # деревья синтаксиса (узлы AST) 
│ │ ├── nodes.go 
│ │ └── printer.go # pretty-print AST (для отладки) 
│ ├── sema/ # семантический анализ (типизация, проверки) 
│ │ └── checker.go 
│ ├── ir/ # промежуточное представление (необязательно сразу) 
│ │ └── ir.go │ ├── backend/ # генерация кода (Go, WASM, и т.д.) 
│ │ └── go_backend.go 
│ └── util/ # вспомогательные утилиты: errors, positions, testing helpers  
├── go.mod 
├── Makefile 
└── README.md
```