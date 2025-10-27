fn main() {
    let result = (1 + 2) * (3 + 4);
    let complex = ((1 + 2) * 3) / (4 - 1);
    let nested = add(subtract(10, 5), multiply(2, 3));
}

fn add(a: i32, b: i32) -> i32 { a + b }
fn subtract(a: i32, b: i32) -> i32 { a - b }
fn multiply(a: i32, b: i32) -> i32 { a * b }

