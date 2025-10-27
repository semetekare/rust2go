package main

import (
	"fmt"
	// Add more imports as needed
)

func main() {
	fmt.Println("=== Начало программы ===")
	result := add_numbers(5, 3)
	fmt.Println("Результат сложения: {}", result)
	greet_user("Алексей")
	fmt.Println(hello_user("Данил"))
	number := 7
	is_even_result := is_even(number)
	fmt.Println("Число {} чётное: {}", number, is_even_result)
	fmt.Println("=== Конец программы ===")
}

func add_numbers(a int, b int) int {
	return (a + b)
}

func greet_user(name string) {
	fmt.Println("Привет, {}! Добро пожаловать в Rust!", name)
}

func hello_user(name string) string {
	return fmt.Sprintf("Привет {}!", name)
}

func is_even(num int) bool {
	return ((num % 2) == 0)
}

