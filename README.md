# Запуск

Полный pipeline компиляции:
```bash
go run ./cmd/main.go ./example/example.rs
```

Это выполнит полный цикл компиляции:
1. Лексический анализ
2. Синтаксический анализ
3. Семантический анализ
4. Трансформация в IR
5. Генерация Go кода

Результат будет сохранён в `output.go`.

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
│ │ ├── checker.go # реализация семантического анализатора
│ ├── ir/ # промежуточное представление 
│ │ ├── ir.go # структуры IR
│ │ └── transformer.go # преобразование AST -> IR
│ ├── backend/ # генерация кода (Go, WASM, и т.д.) 
│ │ └── go_backend.go # генератор Go кода 
├── example/ 
│ └── example.rs # пример кода на Rust
├── output/ # СГЕНЕРИРОВАННЫЙ GO КОД
├── testdata/               # ДИРЕКТОРИЯ ДЛЯ ТЕСТОВЫХ ФАЙЛОВ
│   ├── positive/           # Корректные конструкции (11 файлов)
│   │   ├── fn_simple.rs
│   │   ├── expr_complex.rs
│   │   ├── struct_def.rs
│   │   ├── multiple_functions.rs
│   │   ├── nested_expressions.rs
│   │   ├── comparison_ops.rs
│   │   ├── logical_ops.rs
│   │   ├── unary_ops.rs
│   │   ├── macro_calls.rs
│   │   ├── type_inference.rs
│   │   └── arithmetic.rs
│   └── negative/           # Синтаксические и семантические ошибки (9 файлов)
│       ├── missing_semi.rs
│       ├── missing_paren.rs
│       ├── bad_operator.rs
│       ├── undefined_var.rs
│       ├── wrong_arg_count.rs
│       ├── type_mismatch.rs
│       ├── operator_type_error.rs
│       ├── logical_type_error.rs
│       └── duplicate_function.rs
├── go.mod 
└── README.md
```