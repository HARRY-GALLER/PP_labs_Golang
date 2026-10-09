package main

import "fmt"

// ------

type Rectangle struct {
	width float64
	height float64
}

func (r Rectangle) area() float64 {
	return r.width * r.height
}

// ------

func main() {
	fmt.Println("====================================================")

	//

	fmt.Println("ЗАДАНИЕ 1")

	var number int
	fmt.Print("Введите число: ")
	fmt.Scan(&number)

	if number % 2 == 0 {
		fmt.Printf("Число %v - четное\n", number)
	} else {
		fmt.Printf("Число %v - нечетное\n", number)
	}

	//

	fmt.Println("\nЗАДАНИЕ 2")

	var isNumber int
	fmt.Print("Введите число: ")
	fmt.Scan(&isNumber)

	pnz := func(x int) string {
		if x > 0 {
			return "Positive"
		} else if x < 0 {
			return "Negative"
		} else {
			return "Zero"
		}
	}

	fmt.Println(pnz(isNumber))

	//

	fmt.Println("\nЗАДАНИЕ 3")

	for i := 1; i <= 10; i++ {
		fmt.Println(i)
	}

	//

	fmt.Println("\nЗАДАНИЕ 4")
	
	var str string
	fmt.Printf("Введите строку: ")
	fmt.Scan(&str)

	str_len := func(str string) int {
		return len(str)
	}
	
	fmt.Printf("Длинна строки - %v символов\n", str_len(str))

	//

	fmt.Println("\nЗАДАНИЕ 5")

	var w float64
	var	h float64

	fmt.Print("Введите ширину: ")
	fmt.Scan(&w)

	fmt.Print("Введите высоту: ")
	fmt.Scan(&h)

	rect := Rectangle{w, h}

	fmt.Println("Площадь:", rect.area())


	//

	fmt.Println("\nЗАДАНИЕ 6")

	var x int
	var y int

	fmt.Print("Введите x: ")
	fmt.Scan(&x)

	fmt.Print("Введите y: ")
	fmt.Scan(&y)

	avg := func(x int, y int) float64 {
		return float64(x + y) / 2
	}

	fmt.Println("Среднее значение:", avg(x, y))

	//

	fmt.Println("====================================================")
}