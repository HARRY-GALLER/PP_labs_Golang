package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
)

// Memory repository: HTTP tests don't need a real database or Docker.
type memoryRepository struct {
	users map[int64]User
	next  int64
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{
		users: map[int64]User{
			1: {ID: 1, Name: "Ivan", Age: 20, Email: "ivan@example.com"},
			2: {ID: 2, Name: "Anna", Age: 21, Email: "anna@example.com"},
			3: {ID: 3, Name: "Irina", Age: 20, Email: "irina@example.com"},
		},
		next: 4,
	}
}

func (m *memoryRepository) List(_ context.Context, name string, age, page, limit int) (UserList, error) {
	result := UserList{Users: []User{}, Page: page, Limit: limit}
	ids := make([]int64, 0, len(m.users))
	for id := range m.users {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	filtered := []User{}
	for _, id := range ids {
		u := m.users[id]
		if name != "" && !strings.Contains(strings.ToLower(u.Name), strings.ToLower(name)) {
			continue
		}
		if age != 0 && u.Age != age {
			continue
		}
		filtered = append(filtered, u)
	}
	result.Total = len(filtered)
	start := (page - 1) * limit
	if start >= len(filtered) {
		return result, nil
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	result.Users = filtered[start:end]
	return result, nil
}

func (m *memoryRepository) Get(_ context.Context, id int64) (User, error) {
	user, ok := m.users[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (m *memoryRepository) emailExists(email string, ignoreID int64) bool {
	for id, u := range m.users {
		if id != ignoreID && u.Email == email {
			return true
		}
	}
	return false
}

func (m *memoryRepository) Create(_ context.Context, input UserInput) (User, error) {
	if m.emailExists(input.Email, 0) {
		return User{}, ErrEmailExists
	}
	u := User{ID: m.next, Name: input.Name, Age: input.Age, Email: input.Email}
	m.users[u.ID] = u
	m.next++
	return u, nil
}

func (m *memoryRepository) Update(_ context.Context, id int64, input UserInput) (User, error) {
	if _, ok := m.users[id]; !ok {
		return User{}, ErrNotFound
	}
	if m.emailExists(input.Email, id) {
		return User{}, ErrEmailExists
	}
	u := User{ID: id, Name: input.Name, Age: input.Age, Email: input.Email}
	m.users[id] = u
	return u, nil
}

func (m *memoryRepository) Delete(_ context.Context, id int64) error {
	if _, ok := m.users[id]; !ok {
		return ErrNotFound
	}
	delete(m.users, id)
	return nil
}

func requestUnauth(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func loginToken(t *testing.T, handler http.Handler) string {
	t.Helper()
	user, password := credentials()
	body, err := json.Marshal(map[string]string{"username": user, "password": password})
	if err != nil {
		t.Fatal(err)
	}
	w := requestUnauth(t, handler, http.MethodPost, "/login", string(body))
	if w.Code != http.StatusOK {
		t.Fatalf("login: %d %s", w.Code, w.Body.String())
	}
	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Token == "" {
		t.Fatal("empty login token")
	}
	return result.Token
}

func requestWithToken(t *testing.T, handler http.Handler, token, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func request(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	return requestWithToken(t, handler, loginToken(t, handler), method, path, body)
}

func assertStatus(t *testing.T, got *httptest.ResponseRecorder, want int) {
	t.Helper()
	if got.Code != want {
		t.Fatalf("status = %d; want %d; body = %s", got.Code, want, got.Body.String())
	}
}

func TestListUsers(t *testing.T) {
	h := newServer(newMemoryRepository()).routes()
	t.Run("all", func(t *testing.T) {
		w := request(t, h, http.MethodGet, "/users", "")
		assertStatus(t, w, http.StatusOK)
		var list UserList
		if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
			t.Fatal(err)
		}
		if list.Total != 3 || len(list.Users) != 3 || list.Page != 1 || list.Limit != 10 {
			t.Fatalf("unexpected list: %+v", list)
		}
	})
	t.Run("filter and pagination", func(t *testing.T) {
		w := request(t, h, http.MethodGet, "/users?name=i&age=20&page=2&limit=1", "")
		assertStatus(t, w, http.StatusOK)
		var list UserList
		if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
			t.Fatal(err)
		}
		if list.Total != 2 || len(list.Users) != 1 || list.Users[0].Name != "Irina" {
			t.Fatalf("unexpected list: %+v", list)
		}
	})
	t.Run("invalid age", func(t *testing.T) {
		assertStatus(t, request(t, h, http.MethodGet, "/users?age=abc", ""), http.StatusBadRequest)
	})
	t.Run("invalid limit", func(t *testing.T) {
		assertStatus(t, request(t, h, http.MethodGet, "/users?limit=0", ""), http.StatusBadRequest)
	})
	t.Run("empty page", func(t *testing.T) {
		w := request(t, h, http.MethodGet, "/users?page=99", "")
		assertStatus(t, w, http.StatusOK)
		if !bytes.Contains(w.Body.Bytes(), []byte(`"users":[]`)) {
			t.Fatalf("want []: %s", w.Body.String())
		}
	})
}

func TestGetUser(t *testing.T) {
	h := newServer(newMemoryRepository()).routes()
	w := request(t, h, http.MethodGet, "/users/1", "")
	assertStatus(t, w, http.StatusOK)
	var u User
	if err := json.Unmarshal(w.Body.Bytes(), &u); err != nil {
		t.Fatal(err)
	}
	if u.Name != "Ivan" {
		t.Fatalf("unexpected user: %+v", u)
	}
	assertStatus(t, request(t, h, http.MethodGet, "/users/999", ""), http.StatusNotFound)
	assertStatus(t, request(t, h, http.MethodGet, "/users/abc", ""), http.StatusBadRequest)
}

func TestCreateUser(t *testing.T) {
	h := newServer(newMemoryRepository()).routes()
	w := request(t, h, http.MethodPost, "/users", `{"name":"Petr","age":19,"email":"petr@example.com"}`)
	assertStatus(t, w, http.StatusCreated)
	if w.Header().Get("Location") != "/users/4" {
		t.Fatal("invalid Location header")
	}
	var u User
	if err := json.Unmarshal(w.Body.Bytes(), &u); err != nil {
		t.Fatal(err)
	}
	if u.ID != 4 || u.Name != "Petr" {
		t.Fatalf("unexpected created user: %+v", u)
	}
	assertStatus(t, request(t, h, http.MethodPost, "/users", `{"name":"A","age":19,"email":"bad"}`), http.StatusBadRequest)
	assertStatus(t, request(t, h, http.MethodPost, "/users", `{"name":"Petr","age":19,"email":"ivan@example.com"}`), http.StatusConflict)
	assertStatus(t, request(t, h, http.MethodPost, "/users", `{bad-json}`), http.StatusBadRequest)
	assertStatus(t, request(t, h, http.MethodPost, "/users", `{"name":"Petr","age":19,"email":"petr2@example.com","extra":1}`), http.StatusBadRequest)
	assertStatus(t, request(t, h, http.MethodPost, "/users", `{"name":"Petr","age":19,"email":"petr2@example.com"} {}`), http.StatusBadRequest)
}

func TestUpdateUser(t *testing.T) {
	h := newServer(newMemoryRepository()).routes()
	w := request(t, h, http.MethodPut, "/users/1", `{"name":"Ivan Petrov","age":22,"email":"ivan2@example.com"}`)
	assertStatus(t, w, http.StatusOK)
	var u User
	if err := json.Unmarshal(w.Body.Bytes(), &u); err != nil {
		t.Fatal(err)
	}
	if u.Name != "Ivan Petrov" || u.Age != 22 {
		t.Fatalf("unexpected updated user: %+v", u)
	}
	assertStatus(t, request(t, h, http.MethodPut, "/users/999", `{"name":"Petr","age":20,"email":"petr@example.com"}`), http.StatusNotFound)
	assertStatus(t, request(t, h, http.MethodPut, "/users/1", `{"name":"Petr","age":20,"email":"anna@example.com"}`), http.StatusConflict)
	assertStatus(t, request(t, h, http.MethodPut, "/users/1", `{"name":"Petr","age":150,"email":"petr@example.com"}`), http.StatusBadRequest)
}

func TestDeleteUser(t *testing.T) {
	h := newServer(newMemoryRepository()).routes()
	w := request(t, h, http.MethodDelete, "/users/1", "")
	assertStatus(t, w, http.StatusNoContent)
	if w.Body.Len() != 0 {
		t.Fatalf("204 should have empty body: %s", w.Body.String())
	}
	assertStatus(t, request(t, h, http.MethodGet, "/users/1", ""), http.StatusNotFound)
	assertStatus(t, request(t, h, http.MethodDelete, "/users/999", ""), http.StatusNotFound)
	assertStatus(t, request(t, h, http.MethodDelete, "/users/0", ""), http.StatusBadRequest)
}

func TestIncorrectMethodAndErrorsJSON(t *testing.T) {
	h := newServer(newMemoryRepository()).routes()
	w := request(t, h, http.MethodPatch, "/users/1", "")
	assertStatus(t, w, http.StatusMethodNotAllowed)
	w = request(t, h, http.MethodGet, "/not-found", "")
	assertStatus(t, w, http.StatusNotFound)
	if !strings.HasPrefix(w.Header().Get("Content-Type"), "application/json") {
		t.Fatal("404 error response should be JSON")
	}
	w = request(t, h, http.MethodGet, "/users/999", "")
	assertStatus(t, w, http.StatusNotFound)
	if !strings.HasPrefix(w.Header().Get("Content-Type"), "application/json") {
		t.Fatal("error response should be JSON")
	}
	body, _ := io.ReadAll(w.Body)
	if !bytes.Contains(body, []byte(`"error"`)) {
		t.Fatalf("no error key: %s", body)
	}
}

func TestLoginAuthorization(t *testing.T) {
	h := newServer(newMemoryRepository()).routes()
	w := requestUnauth(t, h, http.MethodGet, "/users", "")
	assertStatus(t, w, http.StatusUnauthorized)
	w = requestUnauth(t, h, http.MethodPost, "/login", `{"username":"wrong","password":"wrong"}`)
	assertStatus(t, w, http.StatusUnauthorized)
	w = requestUnauth(t, h, http.MethodPost, "/login", `{bad`)
	assertStatus(t, w, http.StatusBadRequest)
	token := loginToken(t, h)
	assertStatus(t, requestWithToken(t, h, token, http.MethodGet, "/users", ""), http.StatusOK)
	assertStatus(t, requestWithToken(t, h, "wrong-token", http.MethodGet, "/users", ""), http.StatusUnauthorized)
	assertStatus(t, requestWithToken(t, h, "", http.MethodDelete, "/users/1", ""), http.StatusUnauthorized)
}

func TestTwoClientsParallel(t *testing.T) {
	// Separate tokens allow two simultaneous clients to use the same server.
	h := newServer(newMemoryRepository()).routes()
	token1, token2 := loginToken(t, h), loginToken(t, h)
	if token1 == token2 {
		t.Fatal("different logins generated the same token")
	}
	const n = 30
	var wg sync.WaitGroup
	failures := make(chan string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			token := token1
			if i%2 == 1 {
				token = token2
			}
			r := httptest.NewRequest(http.MethodGet, "/users/1", nil)
			r.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != http.StatusOK {
				failures <- w.Body.String()
			}
		}(i)
	}
	wg.Wait()
	close(failures)
	for msg := range failures {
		t.Error(msg)
	}
}
