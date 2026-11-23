package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"userapi/internal/auth"
	"userapi/internal/user"
)

type listResponse struct {
	Data []user.User `json:"data"`
	Meta listMeta    `json:"meta"`
}

type listMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func parseID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, badRequest("invalid_id", "id должен быть положительным целым числом")
	}
	return id, nil
}

func (a *API) listUsers(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	fields := map[string]string{}

	page := intParam(q, "page", 1, 1, 1_000_000, fields)
	limit := intParam(q, "limit", 20, 1, 100, fields)
	age := optIntParam(q, "age", fields)
	minAge := optIntParam(q, "min_age", fields)
	maxAge := optIntParam(q, "max_age", fields)
	if len(fields) == 0 && minAge != nil && maxAge != nil && *minAge > *maxAge {
		fields["min_age"] = "min_age не может быть больше max_age"
	}
	if len(fields) > 0 {
		return validationFailed(fields)
	}

	users, total, err := a.repo.List(r.Context(), user.ListParams{
		Name:   strings.TrimSpace(q.Get("name")),
		Age:    age,
		MinAge: minAge,
		MaxAge: maxAge,
		Limit:  limit,
		Offset: (page - 1) * limit,
	})
	if err != nil {
		return err
	}
	if users == nil {
		users = []user.User{}
	}
	writeJSON(w, http.StatusOK, listResponse{
		Data: users,
		Meta: listMeta{Page: page, Limit: limit, Total: total, TotalPages: (total + limit - 1) / limit},
	})
	return nil
}

func (a *API) getUser(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	u, err := a.repo.Get(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, u)
	return nil
}

func (a *API) createUser(w http.ResponseWriter, r *http.Request) error {
	in, err := readInput(w, r)
	if err != nil {
		return err
	}
	var u user.User
	if in.Password != "" {
		if err := auth.ValidatePassword(in.Password); err != nil {
			return validationFailed(map[string]string{"password": err.Error()})
		}
		u, err = a.auth.Register(r.Context(), in)
	} else {
		// Совместимость с лабораторной №8: без пароля запись существует,
		// но не имеет учетных данных для входа в клиент.
		u, err = a.repo.Create(r.Context(), in)
	}
	if err != nil {
		return err
	}
	w.Header().Set("Location", "/users/"+strconv.FormatInt(u.ID, 10))
	writeJSON(w, http.StatusCreated, u)
	return nil
}

func (a *API) updateUser(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	in, err := readInput(w, r)
	if err != nil {
		return err
	}
	if in.Password != "" {
		return validationFailed(map[string]string{"password": "смена пароля через PUT /users/{id} не поддерживается"})
	}
	u, err := a.repo.Update(r.Context(), id, in)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, u)
	return nil
}

func (a *API) deleteUser(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	if err := a.repo.Delete(r.Context(), id); err != nil {
		return err
	}
	a.auth.RevokeUser(id)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// readInput читает тело, нормализует и валидирует его.
func readInput(w http.ResponseWriter, r *http.Request) (user.Input, error) {
	var in user.Input
	if err := decodeJSON(w, r, &in); err != nil {
		return in, err
	}
	in.Normalize()
	if f := in.Validate(); len(f) > 0 {
		return in, validationFailed(f)
	}
	return in, nil
}

func intParam(q url.Values, key string, def, min, max int, fields map[string]string) int {
	v := q.Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < min || n > max {
		fields[key] = fmt.Sprintf("должно быть целым числом от %d до %d", min, max)
		return def
	}
	return n
}

func optIntParam(q url.Values, key string, fields map[string]string) *int {
	if q.Get(key) == "" {
		return nil
	}
	n := intParam(q, key, 0, 0, 150, fields)
	if _, bad := fields[key]; bad {
		return nil
	}
	return &n
}
