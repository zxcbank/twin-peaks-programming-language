package main

const (
	small = `fn mamba(x int, y int){return x+y*x;}; // bullshit
f int;
f = 3 + 9;
print(f);`
	factorial = `x int;
		x  = 10;
		y *int;
		y = &x;
		arr int[10];
		arr[0] = 1;
		arr[1] = 2;
		arr[2] = 3;
		fn factorial(n uint, k string) int {
			if (n <= 1) {
				return 1;
			}
			return n * factorial(n - 1);
		}
		result int;
		result = factorial(20);
	`

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
	`
	ex1 = `
x int;
x = 2 + 4;
print(x);`
)
