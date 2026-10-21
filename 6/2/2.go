package main

import (
	"fmt"
	"sync"
	"time"
)

// Генератор чисел Фибоначчи
func fibonacci(ch chan int) {
	a, b := 0, 1

	for i := 0; i < 10; i++ {
		ch <- a
		a, b = b, a+b
		time.Sleep(200 * time.Millisecond)
	}

	close(ch)
}

// Получение чисел из канала
func printNumbers(ch chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for number := range ch {
		fmt.Println("Получено число:", number)
	}
}

var wg sync.WaitGroup	

func main() {
	ch := make(chan int)

	wg.Add(1)

	go fibonacci(ch)
	go printNumbers(ch, &wg)

	wg.Wait()

	fmt.Println("Передача данных завершена")
}
