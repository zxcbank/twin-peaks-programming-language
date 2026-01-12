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
	for (x = 0; x < 10; x=x+1) {
	print(x);
if (x == 4) {print(x*2);}
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
	z bool;
z = false;
if (z == false) {
print(100 * x * y);
}`
	ex1 = `
x int;
x = 1 * 2 * 3 * 4 * 5 * 6 * 7 * 8 * 9 * 10 * 11 * 12 * 13 * 14 * 15 * 16 * 17 * 18 * 19 * 20;
print(x);`
	simple_function = `
fn add(a int, b int) {
	return a + b;
}
result int;
result = add(20, 10);
print(result);`
	array_example = `
arr int[10];
arr[0] = 10;print(arr[0]);
arr[0] = 2;
print(arr[0]);
`
	array_function = `
	fn factorial(n int) int {
			if (n <= 1) {
				return 1;
			}
			return n * factorial(n - 1);
	}
	fn f(arr int) {
		x int;
		for (x = 0; x < 20; x=x+1) {
			arr[x] = factorial(x);
		} 	
	}
	arr int[20];
	f(arr);
	x int;
	for (x = 0; x < 20; x=x+1) {	
		print(arr[x]);
	}
	print(arr[100]);
	`

	bubble_sort = `	
	fn bubble_sort(arr int, size int) {
		x int;
		for (x = 0; x < size - 1; x = x + 1) {
			y int;
			for (y = x + 1; y < size; y = y + 1) {
				if (arr[x] > arr[y]) {
					tmp int;
					tmp = arr[x];
					arr[x] = arr[y];
					arr[y] = tmp;
				}
			}
		}
	}
	size_t int;
	size_t = 100000;
	arr int[size_t];
	t int;
	for (t = 0; t < size_t; t = t+1) {
		arr[t] = size_t - t;
	}
	//for (t = 0; t < size_t; t = t+1) {
	//	print(arr[t]);
	//}
	bubble_sort(arr, size_t);
	for (t = 0; t < size_t; t = t+1) {
		print(arr[t]);
	}
	`

	float_expression = `
	print(3.1 + 7.1)`
)
