# Запуск
```go run ./cmd/main.go ./example/example.rs```

---

# Тесты
```go test ./internal/parser```

>Тест позитивных исходов:
>
>```go test -run TestPositiveSyntax ./internal/parser```

>Тест негативных исходов
>
>```go test -run TestNegativeSyntax ./internal/parser```

## Покрытие тестами
```go tool ./... cover -html=coverage.out``` - генерация файла с данными о покрытии

```go tool cover -html=coverage.out``` - создать и открыть отчет о покрытии в браузере

---
# Структура проекта

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
│ │ └── ir.go 
│ ├── backend/ # генерация кода (Go, WASM, и т.д.) 
│ │ └── go_backend.go 
│ └── util/ # вспомогательные утилиты: errors, positions, testing helpers  
├── example/ 
│ ├── example.rs/ # пример кода на Rust 
├── testdata/               # ДИРЕКТОРИЯ ДЛЯ ТЕСТОВЫХ ФАЙЛОВ
│   ├── positive/           # Корректные конструкции
│   │   ├── fn_simple.rs
│   │   ├── let_bind.rs
│   │   └── expr_complex.rs
│   └── negative/           # Синтаксические ошибки
│       ├── missing_semi.rs
│       ├── missing_paren.rs
│       └── bad_operator.rs
├── go.mod 
└── README.md
```