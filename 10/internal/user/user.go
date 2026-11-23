// Package user содержит доменную модель, валидацию и интерфейс хранилища.
package user

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrNotFound   = errors.New("user not found")
	ErrEmailTaken = errors.New("email already taken")
)

// User — пользователь, как он хранится и возвращается клиенту.
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Input — данные, которые клиент присылает при создании/обновлении.
type Input struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
	// Password указывается только при создании учетной записи (не возвращается в User).
	Password string `json:"password,omitempty"`
}

// Normalize обрезает пробелы и приводит email к нижнему регистру.
func (in *Input) Normalize() {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
}

// Validate возвращает map «поле -> описание ошибки». Пустая map — данные корректны.
func (in Input) Validate() map[string]string {
	errs := map[string]string{}

	if n := utf8.RuneCountInString(in.Name); n < 2 || n > 100 {
		errs["name"] = "длина имени должна быть от 2 до 100 символов"
	}

	if len(in.Email) > 254 {
		errs["email"] = "email слишком длинный"
	} else if addr, err := mail.ParseAddress(in.Email); err != nil || addr.Address != in.Email {
		errs["email"] = "некорректный email"
	}

	if in.Age < 1 || in.Age > 150 {
		errs["age"] = "возраст должен быть от 1 до 150"
	}
	return errs
}

// ListParams — параметры фильтрации и пагинации списка.
type ListParams struct {
	Name   string // подстрока имени, без учёта регистра
	Age    *int   // точный возраст
	MinAge *int
	MaxAge *int
	Limit  int
	Offset int
}

// Repository — контракт хранилища. Реализуется PostgreSQL и тестовой in-memory версией.
type Repository interface {
	List(ctx context.Context, p ListParams) (users []User, total int, err error)
	Get(ctx context.Context, id int64) (User, error)
	Create(ctx context.Context, in Input) (User, error)
	Update(ctx context.Context, id int64, in Input) (User, error)
	Delete(ctx context.Context, id int64) error
}
