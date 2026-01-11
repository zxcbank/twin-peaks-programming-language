package main

const (
	small = `fn mamba(x int, y int){return x+y*x;} // bullshit
f int;
x int;
x = 1;
y int;
y = 2;
f = mamba(x,y);
print(f);`
	factorial = `
		fn factorial(n int) int {
			if (n <= 1) {
				return 1;
			}
			return n * factorial(n - 1);
		}
		result int;
		result = factorial(20);
		print(result);`
	summ_of_two_numbers = `
		fn number(x int, y int) int {
		return x + y;}
		result int;
		result = number(20, 10);
		print(result);`
	for_example = `x int;
	for (x = 0; x > 10; x=x+1) {
	print(x);
	}`

	math_expression_example = `x int;
		y int;
		y = 10;
		x = 4 - (6 - y * y) / y;
		print(x);`

	if_else_full = `x int;
		x  = 10;
		y int;
		y = 5;
		
		if (x < y) {
		print(x);
		} else {
		print(y);
		}
		print(y);
		if (x > y) {
print(2 * x);}
	`
	ex1 = `
x int;
x = 1 * 2 * 3 * 4 * 5 * 6 * 7 * 8 * 9 * 10 * 11 * 12 * 13 * 14 * 15 * 16 * 17 * 18 * 19 * 20;
print(x);`
)
