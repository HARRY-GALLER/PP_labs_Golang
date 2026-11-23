package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

type sessionStore struct {
	mu     sync.RWMutex
	tokens map[string]struct{}
}

func newSessionStore() *sessionStore {
	return &sessionStore{tokens: make(map[string]struct{})}
}

func credentials() (string, string) {
	username := os.Getenv("LAB9_USERNAME")
	password := os.Getenv("LAB9_PASSWORD")
	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "admin123"
	}
	return username, password
}

func (s *server) login(w http.ResponseWriter, r *http.Request) error {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
		return badRequest("Некорректный JSON авторизации")
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return badRequest("Ожидается один JSON-объект")
	}
	wantUser, wantPassword := credentials()
	okUser := subtle.ConstantTimeCompare([]byte(input.Username), []byte(wantUser)) == 1
	okPassword := subtle.ConstantTimeCompare([]byte(input.Password), []byte(wantPassword)) == 1
	if !okUser || !okPassword {
		return &apiError{status: http.StatusUnauthorized, message: "Неверный логин или пароль"}
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return err
	}
	token := hex.EncodeToString(raw)
	s.auth.mu.Lock()
	s.auth.tokens[token] = struct{}{}
	s.auth.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
	return nil
}

func (s *server) requireAuth(next apiHandler) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			return &apiError{status: http.StatusUnauthorized, message: "Требуется токен авторизации"}
		}
		token := strings.TrimPrefix(h, "Bearer ")
		s.auth.mu.RLock()
		_, ok := s.auth.tokens[token]
		s.auth.mu.RUnlock()
		if !ok {
			return &apiError{status: http.StatusUnauthorized, message: "Недействительный токен"}
		}
		return next(w, r)
	}
}
