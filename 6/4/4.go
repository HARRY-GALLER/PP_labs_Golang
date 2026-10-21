
package main

import (
	"fmt"
	"sync"
)

const (
	goroutinesCount = 10
	incrementsCount = 100000
)

func main() {
	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup

	useMutex := true

	wg.Add(goroutinesCount)

	for i := 0; i < goroutinesCount; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < incrementsCount; j++ {
				if useMutex {
					mu.Lock()
				}

				counter++

				if useMutex {
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	fmt.Println("Итоговый счётчик:", counter)
	fmt.Println("Ожидаемый результат:",
		goroutinesCount*incrementsCount)
}
