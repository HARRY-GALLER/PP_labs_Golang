package main

import (
	"fmt"
	"time"
)

func sum_and_diff(a float64, b float64) (float64, float64) {
	return a + b, a - b
}

func main() {
	fmt.Println("====================================================")
	
	//

	fmt.Println("ЗАДАНИЕ 1")

	curTime := time.Now() 

	fmt.Println("Текущая дата и время:", curTime.Format("02.01.2006 15:04:05"))

	//

	fmt.Println("\nЗАДАНИЕ 2")

	var intNumber int = 10
	var floatNumber float64 = 10.1
	var str string = "Первая лаба"
	var boolean bool = true

	fmt.Printf("int: %v\nfloat: %v\nstring: %v\nbool: %v\n", intNumber, floatNumber, str, boolean)

	// 
	
	fmt.Println("\nЗАДАНИЕ 3 и 4")

	x := 5
	y := 10
	fmt.Println(x + y)
	fmt.Println(x - y)
	fmt.Println(x * y)
	fmt.Println(y / x)

	//

	fmt.Println("\nЗАДАНИЕ 5")

	fmt.Println(sum_and_diff(5.6, 4.5))

	//

	fmt.Println("\nЗАДАНИЕ 6")

	average_of_3 := (1 + 2 + 3) / 3
	fmt.Println(average_of_3)

	fmt.Println("====================================================")
}