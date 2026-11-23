package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Data struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)

		fmt.Printf(
			"Метод: %s | URL: %s | Время: %v\n",
			r.Method,
			r.URL.Path,
			duration,
		)
	})
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"Метод не поддерживается",
			http.StatusMethodNotAllowed,
		)
		return
	}

	fmt.Fprintln(w, "Привет это Go HTTP-сервер.")
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Метод не поддерживается",
			http.StatusMethodNotAllowed,
		)
		return
	}

	var data Data

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(
			w,
			"Некорректный JSON",
			http.StatusBadRequest,
		)
		return
	}

	fmt.Println("Получены данные:")
	fmt.Println("Имя:", data.Name)
	fmt.Println("Сообщение:", data.Message)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"status": "Данные получены",
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/data", dataHandler)

	handler := loggingMiddleware(mux)

	fmt.Println("HTTP-сервер запущен: http://localhost:8080")
	fmt.Println(`Для POST: curl -X POST http://localhost:8080/data -H "Content-Type: application/json" -d "{\"name\":\"Semyon\",\"message\":\"Hello\"}"`)

	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		fmt.Println("Ошибка сервера:", err)
	}
}
