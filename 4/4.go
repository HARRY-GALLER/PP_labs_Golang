package main

import (
	"fmt"
	"strings"
	"strconv"
)

func add_or_print(karta map[string]int) {
	fmt.Println("Нажмите [1 добавить] [2 вывести записи] [3 удалить]")

	var do int
	fmt.Print(": ")
	fmt.Scan(&do)
	
	if do == 1 {
		var name string
		var age int
		fmt.Print("Введите имя: ")
		fmt.Scan(&name)
		fmt.Print("Введите возраст: ")
		fmt.Scan(&age)

		karta[name] = age
		fmt.Print("Новый человек добавлен!\n\n")

	} else if do == 2 {
		i := 1
		fmt.Println("---------------")
		for key, value := range karta{
			fmt.Printf("%v | %v %v\n", i, key, value)
			i += 1
		}
		fmt.Println("---------------")

	} else if do == 3 {  // ЗАДАНИЕ 3
		var name string
		fmt.Print("Введите имя: ")
		fmt.Scan(&name)
		delete(karta, name)

		fmt.Print("Запись удалена!\n\n")
	}
}

func avg_age(karta map[string]int) float64 {
	var sum_ages float64
	var count_peoples float64

	for _, value := range karta {
		sum_ages += float64(value)
		count_peoples += 1
	}
	return sum_ages / count_peoples
}

func main() {
	fmt.Println("===============================================")

	//

	fmt.Println("ЗАДАНИЕ 1")

	peoples := map[string]int{
		"Семён": 20,
		"Вячеслав": 19,
	}
	add_or_print(peoples)
	add_or_print(peoples)

	//

	fmt.Println("\nЗАДАНИЕ 2")

	fmt.Println(avg_age(peoples))

	
	//

	fmt.Println("\nЗАДАНИЕ 3")

	add_or_print(peoples)
	add_or_print(peoples)

	//

	fmt.Println("\nЗАДАНИЕ 4")

	var str string
	fmt.Print("Введите строку: ")
	fmt.Scan(&str)

	fmt.Println(strings.ToUpper(str))

	//

	fmt.Println("\nЗАДАНИЕ 5")

	var numbers int

	fmt.Println("Чтобы посчитать сумму введите 'c'")
	for {
		var n string
		fmt.Print("Введите число: ")
		fmt.Scan(&n)

		if n == "c" {
			break
		} else {
			n, _ := strconv.Atoi(n)
			numbers += n
		}
	}
	fmt.Printf("\nСумма чисел = %v\n", numbers)

	//

	fmt.Println("\nЗАДАНИЕ 6")

	list := [5]int{1, 2, 3, 4, 5}

	fmt.Println("Исходный массив:", list)

	for i := len(list) - 1; i >= 0; i-- {
		fmt.Printf("%v ", list[i])
	} 


}
