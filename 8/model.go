package main

import (
	"context"
	"errors"
)

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

type UserInput struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

type UserList struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}

var (
	ErrNotFound    = errors.New("user not found")
	ErrEmailExists = errors.New("email already exists")
)

type Repository interface {
	List(ctx context.Context, name string, age int, page, limit int) (UserList, error)
	Get(ctx context.Context, id int64) (User, error)
	Create(ctx context.Context, input UserInput) (User, error)
	Update(ctx context.Context, id int64, input UserInput) (User, error)
	Delete(ctx context.Context, id int64) error
}
