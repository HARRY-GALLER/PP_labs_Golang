package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"
)

type server struct {
	repo Repository
}

func newServer(repo Repository) *server { return &server{repo: repo} }

type apiError struct {
	status  int
	message string
}

func (e *apiError) Error() string { return e.message }

func badRequest(message string) error {
	return &apiError{status: http.StatusBadRequest, message: message}
}

type apiHandler func(http.ResponseWriter, *http.Request) error


func (s *server) wrap(next apiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			status, message := http.StatusInternalServerError, "Внутренняя ошибка сервера"
			var apiErr *apiError
			switch {
			case errors.As(err, &apiErr):
				status, message = apiErr.status, apiErr.message
			case errors.Is(err, ErrNotFound):
				status, message = http.StatusNotFound, "Пользователь не найден"
			case errors.Is(err, ErrEmailExists):
				status, message = http.StatusConflict, "Email уже используется"
			default:
				log.Printf("Внутренняя ошибка: %v", err)
			}
			writeJSON(w, status, map[string]string{"error": message})
		}
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", s.wrap(s.listUsers))
	mux.HandleFunc("GET /users/{id}", s.wrap(s.getUser))
	mux.HandleFunc("POST /users", s.wrap(s.createUser))
	mux.HandleFunc("PUT /users/{id}", s.wrap(s.updateUser))
	mux.HandleFunc("DELETE /users/{id}", s.wrap(s.deleteUser))

	mux.HandleFunc("/users", s.wrap(func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Allow", "GET, POST")
		return &apiError{status: http.StatusMethodNotAllowed, message: "Метод не поддерживается"}
	}))
	mux.HandleFunc("/users/{id}", s.wrap(func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Allow", "GET, PUT, DELETE")
		return &apiError{status: http.StatusMethodNotAllowed, message: "Метод не поддерживается"}
	}))
	mux.HandleFunc("/", s.wrap(func(w http.ResponseWriter, r *http.Request) error {
		return &apiError{status: http.StatusNotFound, message: "Маршрут не найден"}
	}))
	return loggingMiddleware(mux)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s — %v", r.Method, r.URL.RequestURI(), time.Since(start))
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("Ошибка записи JSON: %v", err)
	}
}

func readUserInput(w http.ResponseWriter, r *http.Request) (UserInput, error) {
	var input UserInput
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1МБ
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return input, badRequest("Некорректный JSON: " + err.Error())
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return input, badRequest("В теле запроса должен быть один JSON-объект")
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	if len([]rune(input.Name)) < 2 || len([]rune(input.Name)) > 100 {
		return input, badRequest("Имя должно содержать от 2 до 100 символов")
	}
	if input.Age < 1 || input.Age > 120 {
		return input, badRequest("Возраст должен быть от 1 до 120")
	}
	address, err := mail.ParseAddress(input.Email)
	if err != nil || address.Address != input.Email || len(input.Email) > 254 {
		return input, badRequest("Укажите корректный email")
	}
	return input, nil
}

func parseID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, badRequest("id должен быть положительным целым числом")
	}
	return id, nil
}

func queryInt(r *http.Request, key string, fallback, min, max int) (int, error) {
	values, ok := r.URL.Query()[key]
	if !ok {
		return fallback, nil
	}
	if len(values) != 1 {
		return 0, badRequest(fmt.Sprintf("Параметр %s должен быть указан один раз", key))
	}
	n, err := strconv.Atoi(values[0])
	if err != nil || n < min || n > max {
		return 0, badRequest(fmt.Sprintf("Параметр %s должен быть целым числом от %d до %d", key, min, max))
	}
	return n, nil
}

func (s *server) listUsers(w http.ResponseWriter, r *http.Request) error {
	page, err := queryInt(r, "page", 1, 1, 100000)
	if err != nil {
		return err
	}
	limit, err := queryInt(r, "limit", 10, 1, 100)
	if err != nil {
		return err
	}
	age, err := queryInt(r, "age", 0, 1, 120)
	if err != nil {
		return err
	}
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if len([]rune(name)) > 100 {
		return badRequest("Фильтр name не должен превышать 100 символов")
	}
	result, err := s.repo.List(r.Context(), name, age, page, limit)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, result)
	return nil
}

func (s *server) getUser(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	user, err := s.repo.Get(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, user)
	return nil
}

func (s *server) createUser(w http.ResponseWriter, r *http.Request) error {
	input, err := readUserInput(w, r)
	if err != nil {
		return err
	}
	user, err := s.repo.Create(r.Context(), input)
	if err != nil {
		return err
	}
	w.Header().Set("Location", fmt.Sprintf("/users/%d", user.ID))
	writeJSON(w, http.StatusCreated, user)
	return nil
}

func (s *server) updateUser(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	input, err := readUserInput(w, r)
	if err != nil {
		return err
	}
	user, err := s.repo.Update(r.Context(), id, input)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, user)
	return nil
}

func (s *server) deleteUser(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(r.Context(), id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
