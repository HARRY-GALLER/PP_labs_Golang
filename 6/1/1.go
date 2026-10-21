
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Расчёт факториала
func factorial(n int) {
	defer wg.Done()

	result := 1
	for i := 1; i <= n; i++ {
		result *= i
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("Факториал", n, "=", result)
}

// Генерация случайных чисел
func randomNumbers() {
	defer wg.Done()

	for i := 0; i < 5; i++ {
		fmt.Println("Случайное число:", rand.Intn(100))
		time.Sleep(300 * time.Millisecond)
	}
}

// Вычисление суммы числового ряда
func calculateSum(n int) {
	defer wg.Done()

	sum := 0
	for i := 1; i <= n; i++ {
		sum += i
		time.Sleep(150 * time.Millisecond)
	}

	fmt.Println("Сумма чисел от 1 до", n, "=", sum)
}

var wg sync.WaitGroup

func main() {

	wg.Add(3)

	go factorial(10)
	go randomNumbers()
	go calculateSum(20)

	fmt.Println("Все три горутины запущены")

	wg.Wait()

	fmt.Println("Все задачи завершены")
}