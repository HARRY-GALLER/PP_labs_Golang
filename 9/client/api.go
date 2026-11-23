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

type APIClient struct {
	baseURL string
	http    *http.Client
	token   string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *APIClient) request(method, path string, input, result any) error {
	var body io.Reader
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" && path != "/login" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("не удалось отправить запрос: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr struct {
			Error string `json:"error"`
		}
		if json.NewDecoder(resp.Body).Decode(&apiErr) != nil || apiErr.Error == "" {
			apiErr.Error = http.StatusText(resp.StatusCode)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiErr.Error)
	}
	if result == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("не удалось прочитать ответ сервера: %w", err)
	}
	return nil
}

func (c *APIClient) Login(username, password string) error {
	var result struct {
		Token string `json:"token"`
	}
	// Do not use the old token on a new login attempt.
	c.token = ""
	if err := c.request(http.MethodPost, "/login", map[string]string{
		"username": username,
		"password": password,
	}, &result); err != nil {
		return err
	}
	if result.Token == "" {
		return fmt.Errorf("сервер не вернул токен")
	}
	c.token = result.Token
	return nil
}

func (c *APIClient) ListUsers() (UserList, error) {
	var result UserList
	err := c.request(http.MethodGet, "/users", nil, &result)
	return result, err
}

func (c *APIClient) GetUser(id int64) (User, error) {
	var result User
	err := c.request(http.MethodGet, fmt.Sprintf("/users/%d", id), nil, &result)
	return result, err
}

func (c *APIClient) CreateUser(input UserInput) (User, error) {
	var result User
	err := c.request(http.MethodPost, "/users", input, &result)
	return result, err
}

func (c *APIClient) UpdateUser(id int64, input UserInput) (User, error) {
	var result User
	err := c.request(http.MethodPut, fmt.Sprintf("/users/%d", id), input, &result)
	return result, err
}

func (c *APIClient) DeleteUser(id int64) error {
	return c.request(http.MethodDelete, fmt.Sprintf("/users/%d", id), nil, nil)
}
