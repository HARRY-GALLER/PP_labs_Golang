package main

import (
	"fmt"
	"bufio"
	"os"

	"PP/mathutils"
	"PP/stringutils"
)


func max_len_string(list []string) (string, int) {
	var length int
	var result string

	for _, el := range list {
		if len(el) >= len(result) {
			result = el
			length = len(el)
		}
	}
	return result, length
}


func main() {
	fmt.Println("===============================================")

	//

	fmt.Println("ЗАДАНИЕ 1-2")

	var n int64
	fmt.Print("Введите число n: ")
	fmt.Scanln(&n)

	fmt.Println("Факториал числа n =", mathutils.Factorial(n))

	//

	fmt.Println("\nЗАДАНИЕ 3")

	var str string

	fmt.Print("Введите строку: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	str = scanner.Text()

	fmt.Println(stringutils.Reversed(str))

	//

	fmt.Println("\nЗАДАНИЕ 4")

	list := [5]int{0, 0, 0, 0, 0}

	for i := 0; i <= 4; i++ {
		list[i] = i * 5
	}
	fmt.Println("Массив list:", list)

	//

	fmt.Println("\nЗАДАНИЕ 5")

	slice := list[1:]
	fmt.Printf("Исходный срез: %v\n", slice)

	slice = append(slice, 111)
	fmt.Printf("Добивили элемент: %v\n", slice)

	slice = append(slice[:3], slice[4:]...)
	fmt.Printf("Удалили элемент: %v\n", slice)

	//

	fmt.Println("\nЗАДАНИЕ 6")

	strings := []string{"hello", "qwerty", "iq", "deterministic", "labs"}
	length, word := max_len_string(strings)

	fmt.Printf("Самое длинное слово: %v (%v символов)\n", length, word)

	//

	fmt.Println("===============================================")
}
