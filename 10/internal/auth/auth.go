// Package auth отвечает за пароли и независимые сессии клиентов.
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"userapi/internal/user"
)

// AccountRepository реализуется хранилищем; hash никогда не возвращается в API.
type AccountRepository interface {
	CreateWithPassword(ctx context.Context, in user.Input, passwordHash string) (user.User, error)
	Credentials(ctx context.Context, email string) (userID int64, passwordHash, role string, err error)
}

var ErrInvalidCredentials = errors.New("invalid email or password")

const iterations = 120000
const sessionLifetime = 24 * time.Hour

// ValidatePassword не позволяет устанавливать пустые/слишком длинные пароли.
func ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 128 {
		return errors.New("пароль должен содержать от 8 до 128 байт")
	}
	return nil
}

// HashPassword: PBKDF2-HMAC-SHA256 со случайной солью (только стандартная библиотека).
func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := derive([]byte(password), salt, iterations)
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key)), nil
}

func VerifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	rounds, err := strconv.Atoi(parts[1])
	if err != nil || rounds < 1 || rounds > 1000000 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil || len(salt) != 16 {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(want) != sha256.Size {
		return false
	}
	got := derive([]byte(password), salt, rounds)
	return subtle.ConstantTimeCompare(got, want) == 1
}

// derive вычисляет один 32-байтовый блок PBKDF2 по RFC 8018.
func derive(password, salt []byte, rounds int) []byte {
	mac := hmac.New(sha256.New, password)
	mac.Write(salt)
	mac.Write([]byte{0, 0, 0, 1})
	u := mac.Sum(nil)
	key := append([]byte(nil), u...)
	for i := 1; i < rounds; i++ {
		mac.Reset()
		mac.Write(u)
		u = mac.Sum(u[:0])
		for j := range key {
			key[j] ^= u[j]
		}
	}
	return key
}

type Session struct {
	UserID    int64
	Role      string
	CSRFToken string
	ExpiresAt time.Time
}

// Manager безопасен для нескольких одновременных HTTP-клиентов.
type Manager struct {
	repo     AccountRepository
	mu       sync.RWMutex
	sessions map[string]Session
}

func NewManager(repo AccountRepository) *Manager {
	return &Manager{repo: repo, sessions: make(map[string]Session)}
}

func (m *Manager) Register(ctx context.Context, in user.Input) (user.User, error) {
	hash, err := HashPassword(in.Password)
	if err != nil {
		return user.User{}, err
	}
	return m.repo.CreateWithPassword(ctx, in, hash)
}

func (m *Manager) Login(ctx context.Context, email, password string) (string, Session, error) {
	id, hash, role, err := m.repo.Credentials(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, user.ErrNotFound) {
		return "", Session{}, ErrInvalidCredentials
	}
	if err != nil {
		return "", Session{}, err
	}
	if !VerifyPassword(hash, password) {
		return "", Session{}, ErrInvalidCredentials
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", Session{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	csrfBytes := make([]byte, 32)
	if _, err := rand.Read(csrfBytes); err != nil {
		return "", Session{}, err
	}
	session := Session{UserID: id, Role: role, CSRFToken: base64.RawURLEncoding.EncodeToString(csrfBytes), ExpiresAt: time.Now().Add(sessionLifetime)}
	m.mu.Lock()
	// Очистка просроченных сессий при новом входе.
	for key, old := range m.sessions {
		if time.Now().After(old.ExpiresAt) {
			delete(m.sessions, key)
		}
	}
	m.sessions[token] = session
	m.mu.Unlock()
	return token, session, nil
}

func (m *Manager) Check(token string) (Session, bool) {
	m.mu.RLock()
	s, ok := m.sessions[token]
	m.mu.RUnlock()
	return s, ok && time.Now().Before(s.ExpiresAt)
}

func (m *Manager) Logout(token string) {
	m.mu.Lock()
	delete(m.sessions, token)
	m.mu.Unlock()
}

// RevokeUser завершает сессии при удалении пользователя.
func (m *Manager) RevokeUser(id int64) {
	m.mu.Lock()
	for token, session := range m.sessions {
		if session.UserID == id {
			delete(m.sessions, token)
		}
	}
	m.mu.Unlock()
}
