
package main

import (
	"fmt"
	"sync"
)

type Request struct {
	A, B   float64
	Op     string
	Result chan string
}

func calculator(requests chan Request, wg *sync.WaitGroup) {
	defer wg.Done()

	for req := range requests {
		var result float64

		switch req.Op {
		case "+":
			result = req.A + req.B
		case "-":
			result = req.A - req.B
		case "*":
			result = req.A * req.B
		case "/":
			if req.B == 0 {
				req.Result <- "Ошибка: деление на ноль"
				continue
			}
			result = req.A / req.B
		default:
			req.Result <- "Ошибка: неизвестная операция"
			continue
		}

		req.Result <- fmt.Sprintf("%.2f", result)
	}
}

func main() {
	requests := make(chan Request)
	var wg sync.WaitGroup

	// Запускаем три обработчика запросов
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go calculator(requests, &wg)
	}

	operations := []Request{
		{A: 10, B: 5, Op: "+"},
		{A: 10, B: 5, Op: "-"},
		{A: 10, B: 5, Op: "*"},
		{A: 10, B: 5, Op: "/"},
		{A: 10, B: 0, Op: "/"},
	}

	for _, op := range operations {
		op.Result = make(chan string, 1)
		requests <- op

		fmt.Printf("%.2f %s %.2f = %s\n", op.A, op.Op, op.B, <-op.Result)
	}

	close(requests)
	wg.Wait()

	fmt.Println("Работа калькулятора завершена")
}
