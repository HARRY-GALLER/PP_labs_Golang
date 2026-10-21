
package main

import (
	"fmt"
	"math/rand"
	"time"
)

func generateNumbers(numbers chan int, parity chan string) {
	for i := 0; i < 10; i++ {
		n := rand.Intn(100)

		numbers <- n

		if n%2 == 0 {
			parity <- fmt.Sprintf("%d — чётное число", n)
		} else {
			parity <- fmt.Sprintf("%d — нечётное число", n)
		}

		time.Sleep(200 * time.Millisecond)
	}

	close(numbers)
	close(parity)
}

func main() {

	numbers := make(chan int)
	parity := make(chan string)

	go generateNumbers(numbers, parity)

	numbersOpen := true
	parityOpen := true

	for numbersOpen || parityOpen {
		select {
		case n, ok := <-numbers:
			if !ok {
				numbersOpen = false
			} else {
				fmt.Println("Получено число:", n)
			}

		case message, ok := <-parity:
			if !ok {
				parityOpen = false
			} else {
				fmt.Println("Результат:", message)
			}
		}
	}

	fmt.Println("Обработка каналов завершена")
}
