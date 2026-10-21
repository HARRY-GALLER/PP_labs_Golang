
package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

type Task struct {
	ID   int
	Text string
}

type Result struct {
	ID   int
	Text string
}

// Переворачивает строку
func reverse(s string) string {
	runes := []rune(s)

	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// Воркер получает задания и отправляет результаты
func worker(id int, tasks <-chan Task,
	results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range tasks {
		reversed := reverse(task.Text)

		fmt.Printf("Воркер %d обработал строку %d\n",
			id, task.ID)

		results <- Result{
			ID:   task.ID,
			Text: reversed,
		}
	}
}

func main() {
	var workersCount int

	fmt.Print("Введите количество воркеров: ")
	if _, err := fmt.Scan(&workersCount); err != nil ||
		workersCount <= 0 {
		fmt.Println("Ошибка: введите положительное целое число")
		return
	}

	file, err := os.Open("C:/myfiles/LABS/3/PP/6/6/input.txt")
	if err != nil {
		fmt.Println("Ошибка открытия input.txt:", err)
		return
	}

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	scanErr := scanner.Err()
	file.Close()

	if scanErr != nil {
		fmt.Println("Ошибка чтения файла:", scanErr)
		return
	}

	tasks := make(chan Task)
	results := make(chan Result)
	var wg sync.WaitGroup

	for i := 1; i <= workersCount; i++ {
		wg.Add(1)
		go worker(i, tasks, results, &wg)
	}

	// Отправляем задания воркерам
	go func() {
		for i, line := range lines {
			tasks <- Task{ID: i, Text: line}
		}
		close(tasks)
	}()

	// Закрываем канал результатов после завершения воркеров
	go func() {
		wg.Wait()
		close(results)
	}()

	// Сохраняем результаты в исходном порядке
	output := make([]string, len(lines))

	for result := range results {
		output[result.ID] = result.Text
	}

	outFile, err := os.Create("C:/myfiles/LABS/3/PP/6/6/output.txt")
	if err != nil {
		fmt.Println("Ошибка создания output.txt:", err)
		return
	}

	writer := bufio.NewWriter(outFile)

	for _, line := range output {
		fmt.Fprintln(writer, line)
	}

	if err := writer.Flush(); err != nil {
		fmt.Println("Ошибка записи:", err)
		outFile.Close()
		return
	}

	if err := outFile.Close(); err != nil {
		fmt.Println("Ошибка закрытия файла:", err)
		return
	}

	fmt.Println("Обработка завершена. Результат сохранён в output.txt")
}
