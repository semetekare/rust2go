// Функция main - точка входа в программу
fn main() {
    println!("=== Начало программы ===");
    
    // Вызов нашей дополнительной функции
    let result = add_numbers(5, 3);
    println!("Результат сложения: {}", result);
    
    // Вызов другой функции
    greet_user("Алексей");
    
    println!(hello_user("Данил"));

    // Работа с результатом функции
    let number = 7;
    let is_even_result = is_even(number);
    println!("Число {} чётное: {}", number, is_even_result);
    
    println!("=== Конец программы ===");
}

// Функция для сложения двух чисел
fn add_numbers(a: i32, b: i32) -> i32 {
    a + b  // В Rust последнее выражение без точки с запятой возвращается
}

// Функция для приветствия пользователя
fn greet_user(name: &str) {
    println!("Привет, {}! Добро пожаловать в Rust!", name);
}

fn hello_user(name: &str) -> String {
    format!("Привет {}!", name)
}

// Функция, проверяющая чётность числа
fn is_even(num: i32) -> bool {
    num % 2 == 0
}