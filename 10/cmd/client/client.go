package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// User и UserInput соответствуют JSON-моделям сервера из 8-й работы.
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	Password string `json:"password,omitempty"`
}

type ListResult struct {
	Data []User `json:"data"`
	Meta struct {
		Page       int `json:"page"`
		Limit      int `json:"limit"`
		Total      int `json:"total"`
		TotalPages int `json:"total_pages"`
	} `json:"meta"`
}

type loginResult struct {
	Token     string    `json:"token"`
	CSRFToken string    `json:"csrf_token"`
	Role      string    `json:"role"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type apiError struct {
	Error struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Details map[string]string `json:"details"`
	} `json:"error"`
}

// Client — отдельная сессия с собственным токеном (по одной на процесс).
type Client struct {
	baseURL string
	http    *http.Client
	token   string
	csrf    string
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) request(method, path string, body, result any) error {
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("кодирование JSON: %w", err)
		}
		payload = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, c.baseURL+path, payload)
	if err != nil {
		return fmt.Errorf("создание запроса: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// Все последующие запросы после входа несут токен сессии.
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
		if method != http.MethodGet && method != http.MethodHead {
			req.Header.Set("X-CSRF-Token", c.csrf)
		}
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("соединение с сервером: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("чтение ответа: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr apiError
		if json.Unmarshal(data, &apiErr) == nil && apiErr.Error.Code != "" {
			msg := fmt.Sprintf("HTTP %d [%s]: %s", resp.StatusCode,
				apiErr.Error.Code, apiErr.Error.Message)
			if len(apiErr.Error.Details) > 0 {
				msg += fmt.Sprintf(" (поля: %v)", apiErr.Error.Details)
			}
			return fmt.Errorf("%s", msg)
		}
		return fmt.Errorf("сервер вернул HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	// DELETE и logout возвращают 204 без тела.
	if result == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("ошибка JSON ответа: %w", err)
	}
	return nil
}

func (c *Client) Register(in UserInput) (User, error) {
	var u User
	err := c.request(http.MethodPost, "/auth/register", in, &u)
	return u, err
}

func (c *Client) Login(email, password string) (loginResult, error) {
	var out loginResult
	err := c.request(http.MethodPost, "/auth/login", map[string]string{
		"email": email, "password": password,
	}, &out)
	if err != nil {
		return out, err
	}
	// Сохраненный токен будет использоваться в следующих запросах.
	c.token = out.Token
	c.csrf = out.CSRFToken
	return out, nil
}

func (c *Client) Logout() error {
	if c.token == "" {
		return nil
	}
	err := c.request(http.MethodPost, "/auth/logout", nil, nil)
	if err == nil {
		c.token = ""
		c.csrf = ""
	}
	return err
}

func (c *Client) List(page int) (ListResult, error) {
	var out ListResult
	err := c.request(http.MethodGet, fmt.Sprintf("/users?page=%d&limit=20", page), nil, &out)
	return out, err
}

func (c *Client) Get(id int64) (User, error) {
	var out User
	err := c.request(http.MethodGet, fmt.Sprintf("/users/%d", id), nil, &out)
	return out, err
}

func (c *Client) Create(in UserInput) (User, error) {
	var out User
	err := c.request(http.MethodPost, "/users", in, &out)
	return out, err
}

func (c *Client) Update(id int64, in UserInput) (User, error) {
	var out User
	err := c.request(http.MethodPut, fmt.Sprintf("/users/%d", id), in, &out)
	return out, err
}

func (c *Client) Delete(id int64) error {
	return c.request(http.MethodDelete, fmt.Sprintf("/users/%d", id), nil, nil)
}
