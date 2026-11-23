package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
)

type mockAPI struct {
	mu       sync.Mutex
	next     int64
	sessions map[string]bool
	users    map[int64]User
}

func newMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	m := &mockAPI{next: 1, sessions: map[string]bool{}, users: map[int64]User{}}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond := func(code int, value any) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			if value != nil {
				_ = json.NewEncoder(w).Encode(value)
			}
		}
		problem := func(code int, message string) { respond(code, map[string]string{"error": message}) }
		m.mu.Lock()
		defer m.mu.Unlock()
		if r.URL.Path == "/login" && r.Method == http.MethodPost {
			var input struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				problem(400, "Неверный JSON")
				return
			}
			if input.Username != "admin" || input.Password != "admin123" {
				problem(401, "Неверные данные")
				return
			}
			token := fmt.Sprintf("token-%d", len(m.sessions)+1)
			m.sessions[token] = true
			respond(200, map[string]string{"token": token})
			return
		}
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") || !m.sessions[token] {
			problem(401, "Требуется авторизация")
			return
		}
		if r.URL.Path == "/users" {
			switch r.Method {
			case http.MethodGet:
				list := UserList{Users: make([]User, 0, len(m.users)), Total: len(m.users), Page: 1, Limit: 10}
				for _, u := range m.users {
					list.Users = append(list.Users, u)
				}
				respond(200, list)
			case http.MethodPost:
				var input UserInput
				if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
					problem(400, "Неверный JSON")
					return
				}
				if input.Name == "" || input.Age < 1 || input.Age > 120 || !strings.Contains(input.Email, "@") {
					problem(400, "Неверные данные")
					return
				}
				for _, user := range m.users {
					if user.Email == input.Email {
						problem(409, "Email занят")
						return
					}
				}
				user := User{ID: m.next, Name: input.Name, Age: input.Age, Email: input.Email}
				m.users[user.ID] = user
				m.next++
				respond(201, user)
			default:
				problem(405, "Метод не поддерживается")
			}
			return
		}
		if strings.HasPrefix(r.URL.Path, "/users/") {
			id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/users/"), 10, 64)
			if err != nil || id < 1 {
				problem(400, "Неверный ID")
				return
			}
			user, ok := m.users[id]
			if !ok {
				problem(404, "Пользователь не найден")
				return
			}
			switch r.Method {
			case http.MethodGet:
				respond(200, user)
			case http.MethodPut:
				var input UserInput
				if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
					problem(400, "Неверный JSON")
					return
				}
				if input.Age < 1 || input.Age > 120 {
					problem(400, "Неверный возраст")
					return
				}
				user.Name, user.Age, user.Email = input.Name, input.Age, input.Email
				m.users[id] = user
				respond(200, user)
			case http.MethodDelete:
				delete(m.users, id)
				w.WriteHeader(http.StatusNoContent)
			default:
				problem(405, "Метод не поддерживается")
			}
			return
		}
		problem(404, "Маршрут не найден")
	}))
}

func TestClientCRUDAndResponses(t *testing.T) {
	server := newMockServer(t)
	defer server.Close()
	c := NewAPIClient(server.URL)
	if _, err := c.ListUsers(); err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected HTTP 401 without token, got %v", err)
	}
	if err := c.Login("admin", "bad"); err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected error for wrong password, got %v", err)
	}
	if err := c.Login("admin", "admin123"); err != nil {
		t.Fatal(err)
	}
	if c.token == "" {
		t.Fatal("client did not save token")
	}
	user, err := c.CreateUser(UserInput{Name: "Ivan", Age: 20, Email: "ivan@example.com"})
	if err != nil || user.ID == 0 {
		t.Fatalf("POST: %v %+v", err, user)
	}
	list, err := c.ListUsers()
	if err != nil || list.Total != 1 {
		t.Fatalf("GET list: %v %+v", err, list)
	}
	fetched, err := c.GetUser(user.ID)
	if err != nil || fetched.Name != "Ivan" {
		t.Fatalf("GET by ID: %v %+v", err, fetched)
	}
	updated, err := c.UpdateUser(user.ID, UserInput{Name: "Petr", Age: 21, Email: "petr@example.com"})
	if err != nil || updated.Name != "Petr" {
		t.Fatalf("PUT: %v %+v", err, updated)
	}
	if _, err := c.CreateUser(UserInput{Name: "", Age: 150, Email: "bad"}); err == nil || !strings.Contains(err.Error(), "400") {
		t.Fatalf("expected HTTP 400, got %v", err)
	}
	if err := c.DeleteUser(user.ID); err != nil {
		t.Fatalf("DELETE: %v", err)
	}
	if _, err := c.GetUser(user.ID); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected HTTP 404, got %v", err)
	}
}

func TestMultipleClientsAtOnce(t *testing.T) {
	server := newMockServer(t)
	defer server.Close()
	const numberOfClients = 20
	var wg sync.WaitGroup
	tokens := make(chan string, numberOfClients)
	errors := make(chan error, numberOfClients)
	for i := 0; i < numberOfClients; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c := NewAPIClient(server.URL) // Separate HTTP client and token for each client.
			if err := c.Login("admin", "admin123"); err != nil {
				errors <- err
				return
			}
			tokens <- c.token
			_, err := c.CreateUser(UserInput{Name: "Test", Age: 20, Email: fmt.Sprintf("user%d@example.com", i)})
			if err != nil {
				errors <- err
			}
		}(i)
	}
	wg.Wait()
	close(errors)
	close(tokens)
	for err := range errors {
		t.Error(err)
	}
	seen := map[string]bool{}
	for tok := range tokens {
		if seen[tok] {
			t.Errorf("reused session token %q", tok)
		}
		seen[tok] = true
	}
	if len(seen) != numberOfClients {
		t.Fatalf("expected %d distinct sessions; got %d", numberOfClients, len(seen))
	}
	c := NewAPIClient(server.URL)
	if err := c.Login("admin", "admin123"); err != nil {
		t.Fatal(err)
	}
	list, err := c.ListUsers()
	if err != nil || list.Total != numberOfClients {
		t.Fatalf("expected %d saved users, got %v %+v", numberOfClients, err, list)
	}
}

func TestConsoleMenu(t *testing.T) {
	server := newMockServer(t)
	defer server.Close()
	input := strings.NewReader(strings.Join([]string{
		"admin", "admin123",
		"3", "Ivan", "20", "ivan@example.com",
		"1", "2", "1",
		"4", "1", "Petr", "21", "petr@example.com",
		"5", "1", "0", "",
	}, "\n"))
	var output bytes.Buffer
	if err := run(input, &output, NewAPIClient(server.URL)); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"Авторизация успешна", "Пользователь создан", "Ivan", "Petr", "Пользователь удалён"} {
		if !strings.Contains(output.String(), text) {
			t.Errorf("missing output %q: %s", text, output.String())
		}
	}
}
