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
	sieve_of_eratosthenes = `
		MAX int;
		MAX = 1000000;
		
		primes int[MAX];
		i int;
		for (i=0; i<MAX; i=i+1) {
			primes[i] = 1;
		}
		limit int;
	    limit = MAX / 2 + 1;
		for (i=2; i<limit; i=i+1) {
			if (primes[i-1]) {
				j int;

				for (j=i*i; j<=MAX; j=j+i) {
					primes[j-1] = 0;
			  	}
			}
		}

		count int;
		count = 0;
		for (i=2; i<=MAX; i=i+1) {
			if (primes[i-1]) {
			  print(i);
			  count=count+1;
			}
		}
		print(count);
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
	size_t = 30000;
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

	quick_sort = `
	fn partition(arr int, low int, high int) {
		pivot int;
		pivot = arr[high]; // Опорный элемент - последний
		i int; 
		i = (low - 1);     // Индекс меньшего элемента
		j int;
		for (j = low; j <= high - 1; j = j + 1) {
			if (arr[j] < pivot) {
				i=i+1;
				tmp int;
				tmp = arr[i];
				arr[i] = arr[j];
				arr[j] = tmp;
			}
		}
		tmp int;
		tmp = arr[i+1];
		arr[i+1] = arr[high];
		arr[high] = tmp;
		return i + 1;
	}

	fn quick_sort(arr int, low int, high int) {
		if (low < high) {
        	pi int; 
			pi = partition(arr, low, high); // Индекс разбиения

        	quick_sort(arr, low, pi - 1);  // Рекурсия для левой части
        	quick_sort(arr, pi + 1, high); // Рекурсия для правой части
    	}
	}
	fn rand_int(a int, seed int, k int) {
		return (a * seed) % k;
	}
	seed int; 
	seed = 1000001;
	size_t int;
	size_t = 30000;
	arr int[size_t];
	t int;
	a int;
	a = 16807;
	m int;
	m = 2147483647;
	for (t = 0; t < size_t; t = t+1) {
		seed = (a * seed) % m;
		arr[t] = seed;
	}
	//for (t = 0; t < size_t; t = t+1) {
	//	print(arr[t]);
	//}
	quick_sort(arr, 0, size_t-1);
	for (t = 0; t < size_t; t = t+1) {
		print(arr[t]);
	}
	`

	float_expression = `
	pi float;
	pi = 3.141592653589793e-10 * 1.0e+10;
	
	print(-pi);
`
	nbody = `
fn setBody(bodies float, idx int, x float, y float, z float, vx float, vy float, vz float, mass float) {
    
	base int;
    base = idx * 7;
    bodies[base+0] = x;
    bodies[base+1] = y;
    bodies[base+2] = z;
    bodies[base+3] = vx;
    bodies[base+4] = vy;
    bodies[base+5] = vz;
    bodies[base+6] = mass;
}

fn offsetMomentum(bodies float, idx int, px float, py float, pz float, SOLAR_MASS float) {
    base int;
    base = idx * 7;
    bodies[base+3] = -px / SOLAR_MASS;
    bodies[base+4] = -py / SOLAR_MASS;
    bodies[base+5] = -pz / SOLAR_MASS;
}

fn energy(bodies float, size int) float {
    e float;
    e = 0.0;
    i int;
    j int;
    for (i = 0; i < size; i = i + 1) {
        base_i int;
        base_i = i * 7;
        mass_i float;
        mass_i = bodies[base_i+6];
        vx float; vy float; vz float;
        vx = bodies[base_i+3];
        vy = bodies[base_i+4];
        vz = bodies[base_i+5];
        e = e + 0.5 * mass_i * (vx*vx + vy*vy + vz*vz);
        for (j = i + 1; j < size; j = j + 1) {
            base_j int;
            base_j = j * 7;
            dx float; dy float; dz float; distance float;
            dx = bodies[base_i+0] - bodies[base_j+0];
            dy = bodies[base_i+1] - bodies[base_j+1];
            dz = bodies[base_i+2] - bodies[base_j+2];
            distance = sqrt(dx*dx + dy*dy + dz*dz);
            e = e - (mass_i * bodies[base_j+6]) / distance;
        }
    }
    return e;
}

fn advance(bodies float, size int, dt float) {
    i int;
    j int;
    for (i = 0; i < size; i = i + 1) {
        base_i int;
        base_i = i * 7;
        mass_i float;
        mass_i = bodies[base_i+6];
        for (j = i + 1; j < size; j = j + 1) {
            base_j int;
            base_j = j * 7;
            dx float; dy float; dz float; distance float; mag float;
            dx = bodies[base_i+0] - bodies[base_j+0];
            dy = bodies[base_i+1] - bodies[base_j+1];
            dz = bodies[base_i+2] - bodies[base_j+2];
            distance = sqrt(dx*dx + dy*dy + dz*dz);
            mag = dt / (distance * distance * distance);
            bodies[base_i+3] = bodies[base_i+3] - dx * bodies[base_j+6] * mag;
            bodies[base_i+4] = bodies[base_i+4] - dy * bodies[base_j+6] * mag;
            bodies[base_i+5] = bodies[base_i+5] - dz * bodies[base_j+6] * mag;
            bodies[base_j+3] = bodies[base_j+3] + dx * mass_i * mag;
            bodies[base_j+4] = bodies[base_j+4] + dy * mass_i * mag;
            bodies[base_j+5] = bodies[base_j+5] + dz * mass_i * mag;
        }
    }
    for (i = 0; i < size; i = i + 1) {
        base_i int;
        base_i = i * 7;
        bodies[base_i+0] = bodies[base_i+0] + dt * bodies[base_i+3];
        bodies[base_i+1] = bodies[base_i+1] + dt * bodies[base_i+4];
        bodies[base_i+2] = bodies[base_i+2] + dt * bodies[base_i+5];
    }
}



fn setJupiter(bodies float, DAYS_PER_YEAR float, SOLAR_MASS float) {
    setBody(bodies, 1,
     4.84143144246472090e+00,
     -1.16032004402742839e+00,
     -1.03622044471123109e-01,
     1.66007664274403694e-03 * DAYS_PER_YEAR,
     7.69901118419740425e-03 * DAYS_PER_YEAR,
     -6.90460016972063023e-05 * DAYS_PER_YEAR,
     9.54791938424326609e-04 * SOLAR_MASS
    );
}

fn setSaturn(bodies float, DAYS_PER_YEAR float, SOLAR_MASS float) {
    setBody(bodies, 2,
        8.34336671824457987e+00,
        4.12479856412430479e+00,
        -4.03523417114321381e-01,
        -2.76742510726862411e-03 * DAYS_PER_YEAR,
        4.99852801234917238e-03 * DAYS_PER_YEAR,
        2.30417297573763929e-05 * DAYS_PER_YEAR,
        2.85885980666130812e-04 * SOLAR_MASS
    );
}

fn setUranus(bodies float, DAYS_PER_YEAR float, SOLAR_MASS float) {
    setBody(bodies, 3,
        1.28943695621391310e+01,
        -1.51111514016986312e+01,
        -2.23307578892655734e-01,
        2.96460137564761618e-03 * DAYS_PER_YEAR,
        2.37847173959480950e-03 * DAYS_PER_YEAR,
        -2.96589568540237556e-05 * DAYS_PER_YEAR,
        4.36624404335156298e-05 * SOLAR_MASS
    );
}

fn setNeptune(bodies float, DAYS_PER_YEAR float, SOLAR_MASS float) {
    setBody(bodies, 4,
        1.53796971148509165e+01,
        -2.59193146099879641e+01,
        1.79258772950371181e-01,
        2.68067772490389322e-03 * DAYS_PER_YEAR,
        1.62824170038242295e-03 * DAYS_PER_YEAR,
        -9.51592254519715870e-05 * DAYS_PER_YEAR,
        5.15138902046611451e-05 * SOLAR_MASS
    );
}

fn setSun(bodies float, SOLAR_MASS float) {
    setBody(bodies, 0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, SOLAR_MASS);
i int;
}
PI float;
SOLAR_MASS float;
DAYS_PER_YEAR float;
PI = 3.141592653589793;
SOLAR_MASS = 4.0 * PI * PI;
DAYS_PER_YEAR = 365.24;
ret float;
ret = 0.0;
n int;
bodies float[35];
for (n = 3; n <= 24; n = n * 2) { // TODO: change limit back to 24
	p int;
	for (p = 0; p < 35; p=p+1) {
		bodies[p] = 0.00;
	}
    setSun(bodies, SOLAR_MASS);
    setJupiter(bodies, DAYS_PER_YEAR, SOLAR_MASS);
    setSaturn(bodies, DAYS_PER_YEAR, SOLAR_MASS);
    setUranus(bodies, DAYS_PER_YEAR, SOLAR_MASS);
    setNeptune(bodies, DAYS_PER_YEAR, SOLAR_MASS);

   px float; py float; pz float;
   px = 0.0; py = 0.0; pz = 0.0;
   size int;
   size = 5;
    i int;
    for (i = 0; i < size; i = i + 1) {
      base int;
      base = i * 7;
      px = px + bodies[base+3] * bodies[base+6];
      py = py + bodies[base+4] * bodies[base+6];
      pz = pz + bodies[base+5] * bodies[base+6];
    }
    offsetMomentum(bodies, 0, px, py, pz, SOLAR_MASS);

    ret = ret + energy(bodies, size);
    max int;
    max = n * 100;
    for (i = 0; i < max; i = i + 1) {
      advance(bodies, size, 0.01);
    }
    ret = ret + energy(bodies, size);
}
expected float;
expected = -1.3524862408537381;
print("EXPECTED:");
print(expected);
print("RETURN:");
print(ret);
`

	simple_float = `
	arr float[1];
	arr[0] = 2;
	print(arr[0]);`

	simple_gc_check = `
	fn array_init() {
		arr int[10];
		i int;
		for (i = 0; i < 2; i = i + 1) {
			arr[i]=i;
		}	
	}	

	array_init();
	array_init();
`
	function_optimization = `
	fn sum_range(n int) int {
		s int;
		s = 0;
		i int;
		for (i = 1; i <= n; i = i + 1) {
			s = s + i;
		}
		return s;
	}
	i int;	
	param int;
	param = 100;
	result int;
	correct int;
	correct = (param * (param + 1)) / 2;
	for (i = 0; i < 3; i = i + 1) {
		result = sum_range(param);
		if (result != correct) {
			print("Error in sum_range (expected vs result):");
			print(correct);
			print(result);
			
		}
	}
	print(result);
`
	fibonacci = `
	fn fibonacci(n int) int {
		if (n <= 1) {
			return n;
		}
		return fibonacci(n - 1) + fibonacci(n - 2);
	}
	result int;
	//i int;
	//for (i = 1; i < 10; i = i + 1) {
	//	result = fibonacci(i);
	//	print(result);
	//}
	result = fibonacci(10);
	result = fibonacci(10);
	print(result);
`
)
